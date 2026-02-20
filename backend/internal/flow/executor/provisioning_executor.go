/*
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
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

package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	authncm "github.com/asgardeo/thunder/internal/authn/common"
	"github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/internal/flow/core"
	"github.com/asgardeo/thunder/internal/group"
	"github.com/asgardeo/thunder/internal/role"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/userprovider"
)

// provisioningExecutor implements the ExecutorInterface for user provisioning in a flow.
type provisioningExecutor struct {
	core.ExecutorInterface
	identifyingExecutorInterface
	userProvider userprovider.UserProviderInterface
	groupService group.GroupServiceInterface
	roleService  role.RoleServiceInterface
	logger       *log.Logger
}

var _ core.ExecutorInterface = (*provisioningExecutor)(nil)
var _ identifyingExecutorInterface = (*provisioningExecutor)(nil)

// newProvisioningExecutor creates a new instance of ProvisioningExecutor.
func newProvisioningExecutor(
	flowFactory core.FlowFactoryInterface,
	groupService group.GroupServiceInterface,
	roleService role.RoleServiceInterface,
	userProvider userprovider.UserProviderInterface,
) *provisioningExecutor {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, ExecutorNameProvisioning),
		log.String(log.LoggerKeyExecutorName, ExecutorNameProvisioning))

	base := flowFactory.CreateExecutor(ExecutorNameProvisioning, common.ExecutorTypeRegistration,
		[]common.Input{}, []common.Input{})

	identifyingExec := newIdentifyingExecutor(ExecutorNameProvisioning,
		[]common.Input{}, []common.Input{}, flowFactory, userProvider)

	return &provisioningExecutor{
		ExecutorInterface:            base,
		identifyingExecutorInterface: identifyingExec,
		userProvider:                 userProvider,
		groupService:                 groupService,
		roleService:                  roleService,
		logger:                       logger,
	}
}

// Execute executes the user provisioning logic based on the inputs provided.
func (p *provisioningExecutor) Execute(ctx *core.NodeContext) (*common.ExecutorResponse, error) {
	logger := p.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Executing user provisioning executor")

	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	// If it's an authentication flow, skip execution if the user is not eligible for provisioning
	if ctx.FlowType == common.FlowTypeAuthentication {
		eligible, ok := ctx.RuntimeData[common.RuntimeKeyUserEligibleForProvisioning]
		if !ok || eligible != dataValueTrue {
			logger.Debug("User is not eligible for provisioning, skipping execution")
			execResp.Status = common.ExecComplete
			return execResp, nil
		}
	}

	if !p.HasRequiredInputs(ctx, execResp) {
		if execResp.Status == common.ExecFailure {
			return execResp, nil
		}

		logger.Debug("Required inputs for provisioning executor is not provided")
		execResp.Status = common.ExecUserInputRequired
		return execResp, nil
	}

	userAttributes := p.getAttributesForProvisioning(ctx)
	if len(userAttributes) == 0 {
		logger.Debug("No user attributes provided for provisioning")
		execResp.Status = common.ExecFailure
		execResp.FailureReason = "No user attributes provided for provisioning"
		return execResp, nil
	}

	userID, err := p.IdentifyUser(userAttributes, execResp)
	if err != nil {
		logger.Error("Failed to identify user", log.Error(err))
		execResp.Status = common.ExecFailure
		execResp.FailureReason = failureReasonFailedToIdentifyUser
		return execResp, nil
	}
	if execResp.Status == common.ExecFailure && execResp.FailureReason != failureReasonUserNotFound {
		return execResp, nil
	}
	if userID != nil && *userID != "" {
		logger.Debug("User already exists", log.String("userID", *userID))

		// If it's a registration flow, check if proceeding with an existing user
		if ctx.FlowType == common.FlowTypeRegistration {
			existing, ok := ctx.RuntimeData[common.RuntimeKeySkipProvisioning]
			if ok && existing == dataValueTrue {
				logger.Debug("Proceeding with an existing user in registration flow, skipping execution")
				execResp.RuntimeData[userAttributeUserID] = *userID
				execResp.Status = common.ExecComplete
				return execResp, nil
			}
		}

		execResp.Status = common.ExecFailure
		execResp.FailureReason = "User already exists"
		return execResp, nil
	}

	// Create the user in the store.
	p.appendNonIdentifyingAttributes(ctx, &userAttributes)
	createdUser, err := p.createUserInStore(ctx, userAttributes)
	if err != nil {
		logger.Error("Failed to create user in the store", log.Error(err))
		execResp.Status = common.ExecFailure
		execResp.FailureReason = "Failed to create user"
		return execResp, nil
	}
	if createdUser == nil || createdUser.UserID == "" {
		logger.Error("Created user is nil or has no ID")
		execResp.Status = common.ExecFailure
		execResp.FailureReason = "Something went wrong while creating the user"
		return execResp, nil
	}

	logger.Debug("User created successfully", log.String("userID", createdUser.UserID))

	// Assign user to groups and roles
	if err := p.assignGroupsAndRoles(ctx, createdUser.UserID); err != nil {
		logger.Error("Failed to assign groups and roles to provisioned user",
			log.String("userID", createdUser.UserID),
			log.Error(err))
		execResp.Status = common.ExecFailure
		execResp.FailureReason = "Failed to assign groups and roles"
		return execResp, nil
	}

	var retAttributes map[string]interface{}
	if err := json.Unmarshal(createdUser.Attributes, &retAttributes); err != nil {
		logger.Error("Failed to unmarshal user attributes", log.Error(err))
		return nil, err
	}

	authenticatedUser := authncm.AuthenticatedUser{
		IsAuthenticated:    true,
		UserID:             createdUser.UserID,
		OrganizationUnitID: createdUser.OrganizationUnitID,
		UserType:           createdUser.UserType,
		Attributes:         retAttributes,
	}
	execResp.AuthenticatedUser = authenticatedUser
	execResp.Status = common.ExecComplete

	// Set user id in runtime data
	execResp.RuntimeData[userAttributeUserID] = createdUser.UserID

	// Set the auto-provisioned flag if it's a user auto provisioning scenario
	if ctx.FlowType == common.FlowTypeAuthentication {
		execResp.RuntimeData[common.RuntimeKeyUserAutoProvisioned] = dataValueTrue
	}

	return execResp, nil
}

// HasRequiredInputs checks if the required inputs are provided in the context and appends any
// missing inputs to the executor response. Returns true if required inputs are found, otherwise false.
func (p *provisioningExecutor) HasRequiredInputs(ctx *core.NodeContext,
	execResp *common.ExecutorResponse) bool {
	logger := p.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Checking inputs for the provisioning executor")

	if p.ExecutorInterface.HasRequiredInputs(ctx, execResp) {
		return true
	}
	if len(execResp.Inputs) == 0 {
		return true
	}

	// Update the executor response with the required inputs retrieved from authenticated user attributes.
	authnUserAttrs := ctx.AuthenticatedUser.Attributes
	if len(authnUserAttrs) > 0 {
		logger.Debug("Authenticated user attributes found, updating executor response required inputs")

		// Clear the required data in the executor response to avoid duplicates.
		missingAttributes := execResp.Inputs
		execResp.Inputs = make([]common.Input, 0)
		if execResp.RuntimeData == nil {
			execResp.RuntimeData = make(map[string]string)
		}

		for _, input := range missingAttributes {
			attribute, exists := authnUserAttrs[input.Identifier]
			if exists {
				attributeStr, ok := attribute.(string)
				if ok {
					logger.Debug("Input exists in authenticated user attributes, adding to runtime data",
						log.String("attributeName", input.Identifier))
					execResp.RuntimeData[input.Identifier] = attributeStr
				}
			} else {
				logger.Debug("Input does not exist in authenticated user attributes, adding to required inputs",
					log.String("attributeName", input.Identifier))
				execResp.Inputs = append(execResp.Inputs, input)
			}
		}

		if len(execResp.Inputs) == 0 {
			logger.Debug("All required inputs are available in authenticated user attributes, " +
				"no further action needed")
			return true
		}
	}

	return false
}

// getAttributesForProvisioning retrieves the input attributes from the context to be stored in user profile.
func (p *provisioningExecutor) getAttributesForProvisioning(ctx *core.NodeContext) map[string]interface{} {
	attributesMap := make(map[string]interface{})
	requiredInputAttrs := p.GetRequiredInputs(ctx)

	// If no input attributes are defined, get all user attributes from the context.
	if len(requiredInputAttrs) == 0 {
		for key, value := range ctx.UserInputs {
			if !slices.Contains(nonUserAttributes, key) {
				attributesMap[key] = value
			}
		}
		for key, value := range ctx.AuthenticatedUser.Attributes {
			if !slices.Contains(nonUserAttributes, key) {
				attributesMap[key] = value
			}
		}
		for key, value := range ctx.RuntimeData {
			if !slices.Contains(nonUserAttributes, key) {
				attributesMap[key] = value
			}
		}
		return attributesMap
	}

	// Otherwise, filter the required input attributes and get their values from the context.
	for _, inputAttr := range requiredInputAttrs {
		if slices.Contains(nonUserAttributes, inputAttr.Identifier) {
			continue
		}

		value, exists := ctx.UserInputs[inputAttr.Identifier]
		if exists {
			attributesMap[inputAttr.Identifier] = value
		} else if runtimeValue, exists := ctx.RuntimeData[inputAttr.Identifier]; exists {
			attributesMap[inputAttr.Identifier] = runtimeValue
		} else if authnValue, exists := ctx.AuthenticatedUser.Attributes[inputAttr.Identifier]; exists {
			attributesMap[inputAttr.Identifier] = authnValue
		}
	}

	return attributesMap
}

// appendNonIdentifyingAttributes appends non-identifying attributes to the provided attributes map.
func (p *provisioningExecutor) appendNonIdentifyingAttributes(ctx *core.NodeContext,
	attributes *map[string]interface{}) {
	if value, exists := ctx.UserInputs[userAttributePassword]; exists {
		(*attributes)[userAttributePassword] = value
	} else if runtimeValue, exists := ctx.RuntimeData[userAttributePassword]; exists {
		(*attributes)[userAttributePassword] = runtimeValue
	}
}

// createUserInStore creates a new user in the user store with the provided attributes.
func (p *provisioningExecutor) createUserInStore(nodeCtx *core.NodeContext,
	userAttributes map[string]interface{}) (*userprovider.User, error) {
	logger := p.logger.With(log.String(log.LoggerKeyFlowID, nodeCtx.FlowID))
	logger.Debug("Creating the user account")

	ouID := p.getOuID(nodeCtx)
	if ouID == "" {
		return nil, fmt.Errorf("organization unit ID not found")
	}
	userType := p.getUserType(nodeCtx)
	if userType == "" {
		return nil, fmt.Errorf("user type not found")
	}

	newUser := userprovider.User{
		OrganizationUnitID: ouID,
		UserType:           userType,
	}

	// Convert the user attributes to JSON.
	attributesJSON, err := json.Marshal(userAttributes)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal user attributes: %w", err)
	}
	newUser.Attributes = attributesJSON

	retUser, svcErr := p.userProvider.CreateUser(&newUser)
	if svcErr != nil {
		return nil, fmt.Errorf("failed to create user in the store: %s", svcErr.Message)
	}
	if retUser != nil && retUser.UserID != "" {
		logger.Debug("User account created successfully", log.String("userID", retUser.UserID))
	}

	return retUser, nil
}

// getOuID retrieves the organization unit ID from runtime data.
func (p *provisioningExecutor) getOuID(ctx *core.NodeContext) string {
	ouID := ""
	// Check for ouId in runtime data
	if val, ok := ctx.RuntimeData[ouIDKey]; ok && val != "" {
		ouID = val
	}
	// If not found, check for defaultOUID in runtime data
	if ouID == "" {
		if val, ok := ctx.RuntimeData[defaultOUIDKey]; ok && val != "" {
			ouID = val
		}
	}

	return ouID
}

// getUserType retrieves the user type from runtime data.
func (p *provisioningExecutor) getUserType(ctx *core.NodeContext) string {
	userType := ""
	if val, ok := ctx.RuntimeData[userTypeKey]; ok && val != "" {
		userType = val
	}

	return userType
}

// assignGroupsAndRoles assigns the newly created user to configured group and role.
// If no group or role is configured, the assignments are skipped.
func (p *provisioningExecutor) assignGroupsAndRoles(
	ctx *core.NodeContext,
	userID string,
) error {
	logger := p.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))

	// Get configured group and role from properties
	groupID := p.getGroupToAssign(ctx)
	roleID := p.getRoleToAssign(ctx)

	// Skip if no group or role configured
	if groupID == "" && roleID == "" {
		logger.Debug("No group or role configured for assignment, skipping")
		return nil
	}

	logger.Debug("Assigning group and role to provisioned user",
		log.String("userID", userID),
		log.String("groupID", groupID),
		log.String("roleID", roleID))

	var groupErr, roleErr error
	// Assign to group
	if groupID != "" {
		if err := p.assignToGroup(context.TODO(), userID, groupID, logger); err != nil {
			groupErr = fmt.Errorf("failed to assign user to group %s: %w", groupID, err)
		}
	}
	// Assign to role
	if roleID != "" {
		if err := p.assignToRole(context.TODO(), userID, roleID, logger); err != nil {
			roleErr = fmt.Errorf("failed to assign user to role %s: %w", roleID, err)
		}
	}
	if groupErr != nil || roleErr != nil {
		if groupErr != nil && roleErr != nil {
			return fmt.Errorf("group assignment error: %w; role assignment error: %s", groupErr, roleErr.Error())
		}
		if groupErr != nil {
			return groupErr
		}
		return roleErr
	}

	logger.Debug("Successfully assigned group and role", log.String("userID", userID))
	return nil
}

// getGroupToAssign retrieves the group ID from node properties.
func (p *provisioningExecutor) getGroupToAssign(ctx *core.NodeContext) string {
	if len(ctx.NodeProperties) == 0 {
		return ""
	}

	groupValue, ok := ctx.NodeProperties[propertyKeyAssignGroup]
	if !ok {
		return ""
	}

	// Handle string value
	if strVal, ok := groupValue.(string); ok {
		return strVal
	}

	return ""
}

// getRoleToAssign retrieves the role ID from node properties.
func (p *provisioningExecutor) getRoleToAssign(ctx *core.NodeContext) string {
	if len(ctx.NodeProperties) == 0 {
		return ""
	}

	roleValue, ok := ctx.NodeProperties[propertyKeyAssignRole]
	if !ok {
		return ""
	}

	// Handle string value
	if strVal, ok := roleValue.(string); ok {
		return strVal
	}

	return ""
}

// assignToGroup adds the user to the specified group.
func (p *provisioningExecutor) assignToGroup(
	ctx context.Context,
	userID string,
	groupID string,
	logger *log.Logger,
) error {
	logger.Debug("Adding user to group",
		log.String("userID", userID),
		log.String("groupID", groupID))

	members := []group.Member{
		{
			ID:   userID,
			Type: group.MemberTypeUser,
		},
	}

	_, svcErr := p.groupService.AddGroupMembers(ctx, groupID, members)
	if svcErr != nil {
		logger.Error("Failed to add user to group",
			log.String("groupID", groupID),
			log.String("userID", userID),
			log.String("error", svcErr.Error))
		return fmt.Errorf("failed to add user to group: %s", svcErr.Error)
	}

	logger.Debug("Successfully added user to group",
		log.String("userID", userID),
		log.String("groupID", groupID))
	return nil
}

// assignToRole adds the user to the specified role.
func (p *provisioningExecutor) assignToRole(
	ctx context.Context, userID string, roleID string, logger *log.Logger) error {
	logger.Debug("Adding user to role",
		log.String("userID", userID),
		log.String("roleID", roleID))

	// AddAssignments appends to existing assignments (doesn't replace)
	assignments := []role.RoleAssignment{
		{
			ID:   userID,
			Type: role.AssigneeTypeUser,
		},
	}

	svcErr := p.roleService.AddAssignments(ctx, roleID, assignments)
	if svcErr != nil {
		logger.Error("Failed to add role assignment",
			log.String("roleID", roleID),
			log.String("userID", userID),
			log.String("error", svcErr.Error))
		return fmt.Errorf("failed to assign role: %s", svcErr.Error)
	}

	logger.Debug("Successfully assigned role",
		log.String("userID", userID),
		log.String("roleID", roleID))
	return nil
}
