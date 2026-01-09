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
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "InviteExecutor"))
	base := flowFactory.CreateExecutor(
		ExecutorNameInviteExecutor,
		common.ExecutorTypeUtility,
		[]common.Input{},
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
	logger.Debug("Executing invite link generator")

	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	// Check if inviteToken is provided by the user
	if inviteTokenInput, exists := ctx.UserInputs["inviteToken"]; exists && inviteTokenInput != "" {
		// User has provided the invite token, validate it against stored token
		storedToken, hasStoredToken := ctx.RuntimeData["inviteToken"]

		if !hasStoredToken {
			logger.Debug("No invite token found in runtime data")
			execResp.Status = common.ExecFailure
			execResp.FailureReason = "Invalid invite token"
			return execResp, nil
		}

		if inviteTokenInput != storedToken {
			logger.Debug("Invite token mismatch",
				log.String("flowId", ctx.FlowID))
			execResp.Status = common.ExecFailure
			execResp.FailureReason = "Invalid invite token"
			return execResp, nil
		}

		logger.Debug("Invite token validated successfully")
		execResp.Status = common.ExecComplete
		return execResp, nil
	}

	// Check if we already generated a token
	var inviteToken string
	if storedToken, exists := ctx.RuntimeData["inviteToken"]; exists && storedToken != "" {
		// Token already generated, reuse it
		inviteToken = storedToken
	} else {
		// Generate new invite token
		newToken, uuidErr := utils.GenerateUUIDv7()
		if uuidErr != nil {
			logger.Debug("Failed to generate invite token", log.Error(uuidErr))
			execResp.Status = common.ExecFailure
			execResp.FailureReason = "Failed to generate invite token"
			return execResp, nil
		}
		inviteToken = newToken
	}

	// Generate invite link using GateClient configuration
	gateConfig := config.GetThunderRuntime().Config.GateClient
	gateAppURL := fmt.Sprintf("%s://%s:%d%s",
		gateConfig.Scheme,
		gateConfig.Hostname,
		gateConfig.Port,
		gateConfig.Path)

	inviteLink := fmt.Sprintf("%s?flowId=%s&inviteToken=%s", gateAppURL, ctx.FlowID, inviteToken)

	// Store the token for validation and return the link
	execResp.RuntimeData["inviteToken"] = inviteToken
	execResp.AdditionalData["inviteLink"] = inviteLink

	execResp.Inputs = []common.Input{
		{
			Identifier: "inviteToken",
			Type:       "HIDDEN",
			Required:   true,
		},
	}
	execResp.Status = common.ExecUserInputRequired
	return execResp, nil
}
