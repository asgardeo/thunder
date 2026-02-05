/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

// Package flow provides flow management tool models.
//
//nolint:lll
package flow

import (
	"context"
	"fmt"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	flowCommon "github.com/asgardeo/thunder/internal/flow/common"
	flowmgt "github.com/asgardeo/thunder/internal/flow/mgt"
	"github.com/asgardeo/thunder/internal/mcp/tools/common"
)

// flowTools provides MCP tools for managing Thunder authentication flows.
type flowTools struct {
	flowService flowmgt.FlowMgtServiceInterface
}

// NewFlowTools creates a new flowTools instance.
func NewFlowTools(flowService flowmgt.FlowMgtServiceInterface) *flowTools {
	return &flowTools{
		flowService: flowService,
	}
}

// Schema definitions
var (
	listFlowsInputSchema       *jsonschema.Schema
	getFlowByHandleInputSchema *jsonschema.Schema
	getFlowByIDInputSchema     *jsonschema.Schema
	createFlowInputSchema      *jsonschema.Schema
	updateFlowInputSchema      *jsonschema.Schema
)

func init() {
	listFlowsInputSchema = common.GenerateSchema[listFlowsInput](
		common.WithEnum("", "flow_type", []string{string(flowCommon.FlowTypeAuthentication), string(flowCommon.FlowTypeRegistration)}),
		common.WithDefault("", "limit", 30),
		common.WithDefault("", "offset", 0),
	)

	getFlowByHandleInputSchema = common.GenerateSchema[getFlowByHandleInput](
		common.WithEnum("", "flow_type", []string{string(flowCommon.FlowTypeAuthentication), string(flowCommon.FlowTypeRegistration)}),
		common.WithRequired("", "handle", "flow_type"),
	)

	getFlowByIDInputSchema = common.GenerateSchema[common.IDInput](
		common.WithRequired("", "id"),
	)

	createFlowInputSchema = common.GenerateSchema[flowmgt.FlowDefinition](
		common.WithEnum("", "flowType", []string{string(flowCommon.FlowTypeAuthentication), string(flowCommon.FlowTypeRegistration)}),
	)

	updateFlowInputSchema = common.GenerateSchema[updateFlowInput](
		common.WithRequired("", "id", "name", "nodes"),
	)
}

// RegisterTools registers all flow tools with the MCP server.
func (t *flowTools) RegisterTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "thunder_list_flows",
		Description: `List available flows. Supports optional filtering by flow_type.`,
		InputSchema: listFlowsInputSchema,
		Annotations: &mcp.ToolAnnotations{
			Title:        "List Flows",
			ReadOnlyHint: true,
		},
	}, t.listFlows)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "thunder_get_flow_by_handle",
		Description: `Retrieve a complete definition of a flow by its handle (human-readable identifier).`,
		InputSchema: getFlowByHandleInputSchema,
		Annotations: &mcp.ToolAnnotations{
			Title:        "Get Flow by Handle",
			ReadOnlyHint: true,
		},
	}, t.getFlowByHandle)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "thunder_get_flow_by_id",
		Description: `Retrieve a complete definition of a flow by its unique ID (UUID).`,
		InputSchema: getFlowByIDInputSchema,
		Annotations: &mcp.ToolAnnotations{
			Title:        "Get Flow by ID",
			ReadOnlyHint: true,
		},
	}, t.getFlowByID)

	mcp.AddTool(server, &mcp.Tool{
		Name: "thunder_create_flow",
		Description: `Create a new authentication or registration flow.

Prerequisites:
- For SMS/Email OTP: Must create notification sender first using create_notification_sender if not already available.
- Review existing similar flows by get_flow_by_handle to understand patterns and node structures

Key Requirements:
- Handle: Lowercase alphanumeric, dashes/underscores allowed (not at start/end). Unique per flow type.
- Structure: Must include START and END nodes with at least one functional node in between.
- Node types: START, END, TASK_EXECUTION, PROMPT.
- PROMPT nodes: Require 'meta.components' array for UI rendering.
- Transitions: Use onSuccess/onFailure node IDs to define the path.`,
		InputSchema: createFlowInputSchema,
		Annotations: &mcp.ToolAnnotations{
			Title:          "Create Flow",
			IdempotentHint: false,
		},
	}, t.createFlow)

	mcp.AddTool(server, &mcp.Tool{
		Name: "thunder_update_flow",
		Description: `Update an existing flow definition (full replacement for updateable fields).

Provide the COMPLETE flow object to update the flow. Use get_flow_by_handle first to get current state (including ID).

Workflow:
1. Use get_flow_by_handle to get current flow
2. Modify name and/or nodes as needed
3. Send the complete updated object back

Flow versions are automatically tracked. Updating creates a new version.`,
		InputSchema: updateFlowInputSchema,
		Annotations: &mcp.ToolAnnotations{
			Title:          "Update Flow",
			IdempotentHint: true,
		},
	}, t.updateFlow)
}

// ListFlows handles the list_flows tool call.
func (t *flowTools) listFlows(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input listFlowsInput,
) (*mcp.CallToolResult, flowListOutput, error) {
	limit := input.Limit
	if limit <= 0 {
		limit = 30
	}

	flowType := flowCommon.FlowType(input.FlowType)

	listResponse, svcErr := t.flowService.ListFlows(ctx, limit, input.Offset, flowType)
	if svcErr != nil {
		return nil, flowListOutput{}, fmt.Errorf("failed to list flows: %s", svcErr.ErrorDescription)
	}

	return nil, flowListOutput{
		TotalCount: listResponse.TotalResults,
		Flows:      listResponse.Flows,
	}, nil
}

// GetFlowByHandle handles the get_flow_by_handle tool call.
func (t *flowTools) getFlowByHandle(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input getFlowByHandleInput,
) (*mcp.CallToolResult, *flowmgt.CompleteFlowDefinition, error) {
	flowType := flowCommon.FlowType(input.FlowType)

	flow, svcErr := t.flowService.GetFlowByHandle(ctx, input.Handle, flowType)
	if svcErr != nil {
		return nil, nil, fmt.Errorf("failed to get flow by handle: %s", svcErr.ErrorDescription)
	}

	return nil, flow, nil
}

// GetFlowByID handles the get_flow_by_id tool call.
func (t *flowTools) getFlowByID(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input common.IDInput,
) (*mcp.CallToolResult, *flowmgt.CompleteFlowDefinition, error) {
	flow, svcErr := t.flowService.GetFlow(ctx, input.ID)
	if svcErr != nil {
		return nil, nil, fmt.Errorf("failed to get flow: %s", svcErr.ErrorDescription)
	}

	return nil, flow, nil
}

// CreateFlow handles the create_flow tool call.
func (t *flowTools) createFlow(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input flowmgt.FlowDefinition,
) (*mcp.CallToolResult, *flowmgt.CompleteFlowDefinition, error) {
	createdFlow, svcErr := t.flowService.CreateFlow(ctx, &input)
	if svcErr != nil {
		return nil, nil, fmt.Errorf("failed to create flow: %s", svcErr.ErrorDescription)
	}

	return nil, createdFlow, nil
}

// UpdateFlow handles the update_flow tool call.
func (t *flowTools) updateFlow(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input updateFlowInput,
) (*mcp.CallToolResult, *flowmgt.CompleteFlowDefinition, error) {
	// Get current flow to retrieve immutable fields (handle, flowType)
	currentFlow, svcErr := t.flowService.GetFlow(ctx, input.ID)
	if svcErr != nil {
		return nil, nil, fmt.Errorf("failed to get flow: %s", svcErr.ErrorDescription)
	}

	// Build update definition with immutable fields preserved and input fields replaced
	updateDef := &flowmgt.FlowDefinition{
		Handle:   currentFlow.Handle,
		FlowType: currentFlow.FlowType,
		Name:     input.Name,
		Nodes:    input.Nodes,
	}

	updatedFlow, svcErr := t.flowService.UpdateFlow(ctx, input.ID, updateDef)
	if svcErr != nil {
		return nil, nil, fmt.Errorf("failed to update flow: %s", svcErr.ErrorDescription)
	}

	return nil, updatedFlow, nil
}
