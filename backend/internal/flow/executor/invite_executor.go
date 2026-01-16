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

package executor

import (
	"fmt"

	"github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/internal/flow/core"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/utils"
)

// inviteExecutor generates an invite link for the user to complete registration.
type inviteExecutor struct {
	core.ExecutorInterface
	logger *log.Logger
}

// newInviteExecutor creates a new instance of the invite executor.
func newInviteExecutor(flowFactory core.FlowFactoryInterface) *inviteExecutor {
	defaultInputs := []common.Input{
		{
			Identifier: userInputInviteToken,
			Type:       "HIDDEN",
			Required:   true,
		},
	}
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "InviteExecutor"))
	base := flowFactory.CreateExecutor(
		ExecutorNameInviteExecutor,
		common.ExecutorTypeUtility,
		defaultInputs,
		[]common.Input{},
	)
	return &inviteExecutor{
		ExecutorInterface: base,
		logger:            logger,
	}
}

// Execute generates the invite link and returns it in AdditionalData.
func (e *inviteExecutor) Execute(ctx *core.NodeContext) (*common.ExecutorResponse, error) {
	logger := e.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Executing invite executor")

	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	// Check if inviteToken is provided by the user
	// Check if inviteToken is provided by the user
	if !e.HasRequiredInputs(ctx, execResp) {
		inviteToken, err := e.getOrGenerateToken(ctx)
		if err != nil {
			logger.Debug("Failed to get or generate invite token", log.Error(err))
			execResp.Status = common.ExecFailure
			execResp.FailureReason = "Failed to generate invite token"
			return execResp, nil
		}

		inviteLink := e.generateInviteLink(ctx, inviteToken)

		// Store the token for validation and return the link
		execResp.RuntimeData[runtimeKeyStoredInviteToken] = inviteToken
		execResp.AdditionalData["inviteLink"] = inviteLink

		execResp.Status = common.ExecUserInputRequired
		return execResp, nil
	}

	// User has provided the invite token, validate it against stored token
	inviteTokenInput := ctx.UserInputs[userInputInviteToken]
	storedToken, hasStoredToken := ctx.RuntimeData[runtimeKeyStoredInviteToken]

	if !hasStoredToken {
		logger.Debug("No invite token found in runtime data")
		execResp.Status = common.ExecFailure
		execResp.FailureReason = "Invalid invite token"
		return execResp, nil
	}

	if inviteTokenInput != storedToken {
		logger.Debug("Invite token mismatch", log.String("flowId", ctx.FlowID))
		execResp.Status = common.ExecFailure
		execResp.FailureReason = "Invalid invite token"
		return execResp, nil
	}

	logger.Debug("Invite token validated successfully")
	execResp.Status = common.ExecComplete
	return execResp, nil
}

// getOrGenerateToken retrieves the existing invite token from runtime data or generates a new one.
func (e *inviteExecutor) getOrGenerateToken(ctx *core.NodeContext) (string, error) {
	if storedToken, exists := ctx.RuntimeData[runtimeKeyStoredInviteToken]; exists && storedToken != "" {
		return storedToken, nil
	}

	return utils.GenerateUUIDv7()
}

// generateInviteLink constructs the invite link using the GateClient configuration.
func (e *inviteExecutor) generateInviteLink(ctx *core.NodeContext, inviteToken string) string {
	gateConfig := config.GetThunderRuntime().Config.GateClient
	gateAppURL := fmt.Sprintf("%s://%s:%d%s",
		gateConfig.Scheme,
		gateConfig.Hostname,
		gateConfig.Port,
		gateConfig.Path)

	return fmt.Sprintf("%s?flowId=%s&inviteToken=%s", gateAppURL, ctx.FlowID, inviteToken)
}
