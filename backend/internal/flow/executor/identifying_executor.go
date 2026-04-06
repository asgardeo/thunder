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
	"encoding/json"
	"slices"

	"github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/internal/flow/core"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/userprovider"
)

const (
	idfExecLoggerComponentName = "IdentifyingExecutor"
)

// identifyingExecutorInterface defines the interface for identifying executors.
type identifyingExecutorInterface interface {
	IdentifyUser(filters map[string]interface{},
		execResp *common.ExecutorResponse) (*string, error)
}

// identifyingExecutor implements the ExecutorInterface for identifying users based on provided attributes.
type identifyingExecutor struct {
	core.ExecutorInterface
	userProvider userprovider.UserProviderInterface
	logger       *log.Logger
}

var _ core.ExecutorInterface = (*identifyingExecutor)(nil)
var _ identifyingExecutorInterface = (*identifyingExecutor)(nil)

// newIdentifyingExecutor creates a new instance of IdentifyingExecutor.
func newIdentifyingExecutor(
	name string,
	defaultInputs, prerequisites []common.Input,
	flowFactory core.FlowFactoryInterface,
	userProvider userprovider.UserProviderInterface,
) *identifyingExecutor {
	if name == "" {
		name = ExecutorNameIdentifying
	}
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, idfExecLoggerComponentName),
		log.String(log.LoggerKeyExecutorName, name))

	base := flowFactory.CreateExecutor(ExecutorNameIdentifying, common.ExecutorTypeUtility,
		defaultInputs, prerequisites)
	return &identifyingExecutor{
		ExecutorInterface: base,
		userProvider:      userProvider,
		logger:            logger,
	}
}

// IdentifyUser identifies a user based on the provided attributes.
func (i *identifyingExecutor) IdentifyUser(filters map[string]interface{},
	execResp *common.ExecutorResponse) (*string, error) {
	logger := i.logger
	logger.Debug("Identifying user with filters")

	// filter out non-searchable attributes
	var searchableFilter = make(map[string]interface{})
	for key, value := range filters {
		if !slices.Contains(nonSearchableInputs, key) {
			searchableFilter[key] = value
		}
	}

	userID, err := i.userProvider.IdentifyUser(searchableFilter)
	if err != nil {
		if err.Code == userprovider.ErrorCodeUserNotFound {
			logger.Debug("User not found for the provided filters")
			execResp.Status = common.ExecFailure
			execResp.FailureReason = failureReasonUserNotFound
			return nil, nil
		} else if err.Code == userprovider.ErrorCodeAmbiguousUser {
			logger.Debug("Multiple users found for the provided filters")
			execResp.Status = common.ExecFailure
			execResp.FailureReason = failureReasonAmbiguousUser
			return nil, nil
		} else {
			logger.Debug("Failed to identify user due to error: " + err.Error())
			execResp.Status = common.ExecFailure
			execResp.FailureReason = failureReasonFailedToIdentifyUser
			return nil, nil
		}
	}

	if userID == nil || *userID == "" {
		logger.Debug("User not found for the provided filter")
		execResp.Status = common.ExecFailure
		execResp.FailureReason = failureReasonUserNotFound
		return nil, nil
	}

	return userID, nil
}

// Execute executes the identifying executor logic.
func (i *identifyingExecutor) Execute(ctx *core.NodeContext) (*common.ExecutorResponse, error) {
	logger := i.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Executing identifying executor")

	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	// Check if required inputs are provided
	if !i.HasRequiredInputs(ctx, execResp) {
		logger.Debug("Required inputs for identifying executor are not provided")
		execResp.Status = common.ExecUserInputRequired
		return execResp, nil
	}

	// Branch on executor mode
	if ctx.ExecutorMode == ExecutorModeResolve {
		return i.executeResolve(ctx, execResp)
	}

	userSearchAttributes := map[string]interface{}{}

	for _, inputData := range i.GetRequiredInputs(ctx) {
		if value, ok := ctx.UserInputs[inputData.Identifier]; ok {
			userSearchAttributes[inputData.Identifier] = value
		} else if value, ok := ctx.RuntimeData[inputData.Identifier]; ok {
			// Fallback to RuntimeData if not in UserInputs
			userSearchAttributes[inputData.Identifier] = value
		}
	}

	// Try to identify the user
	userID, err := i.IdentifyUser(userSearchAttributes, execResp)

	if err != nil {
		logger.Debug("Failed to identify user due to error: " + err.Error())
		execResp.Status = common.ExecFailure
		execResp.FailureReason = failureReasonFailedToIdentifyUser
		return execResp, nil
	}

	if userID == nil || *userID == "" {
		// If IdentifyUser already set a failure status (e.g., ambiguous user), preserve it
		if execResp.Status == common.ExecFailure {
			return execResp, nil
		}
		logger.Debug("User not found for the provided attributes")
		execResp.Status = common.ExecFailure
		execResp.FailureReason = failureReasonUserNotFound
		return execResp, nil
	}

	// Store the resolved userID in RuntimeData for subsequent executors
	execResp.RuntimeData[userAttributeUserID] = *userID
	execResp.Status = common.ExecComplete

	logger.Debug("Identifying executor completed successfully",
		log.String("userID", log.MaskString(*userID)))

	return execResp, nil
}

// executeResolve handles the resolve mode for user disambiguation.
func (i *identifyingExecutor) executeResolve(ctx *core.NodeContext,
	execResp *common.ExecutorResponse) (*common.ExecutorResponse, error) {
	logger := i.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Executing identifying executor in resolve mode")

	// Build search attributes from inputs
	userSearchAttributes := map[string]interface{}{}
	for _, inputData := range i.GetRequiredInputs(ctx) {
		if value, ok := ctx.UserInputs[inputData.Identifier]; ok {
			userSearchAttributes[inputData.Identifier] = value
		} else if value, ok := ctx.RuntimeData[inputData.Identifier]; ok {
			userSearchAttributes[inputData.Identifier] = value
		}
	}

	// Check if we have stored candidate users from a previous call
	storedCandidates, hasCandidates := ctx.RuntimeData[runtimeKeyCandidateUsers]

	var candidates []*userprovider.User

	if !hasCandidates {
		// First call: search for users using the provider
		searchableFilters := make(map[string]interface{})
		for key, value := range userSearchAttributes {
			if !slices.Contains(nonSearchableInputs, key) {
				searchableFilters[key] = value
			}
		}

		users, err := i.userProvider.SearchUsers(searchableFilters)
		if err != nil {
			if err.Code == userprovider.ErrorCodeUserNotFound {
				logger.Debug("No users found for the provided filters")
				execResp.Status = common.ExecFailure
				execResp.FailureReason = failureReasonUserNotFound
				return execResp, nil
			}
			logger.Debug("Failed to search users: " + err.Error())
			execResp.Status = common.ExecFailure
			execResp.FailureReason = failureReasonFailedToIdentifyUser
			return execResp, nil
		}

		candidates = users
	} else {
		// Subsequent call: deserialize stored candidates and filter in-memory
		if err := json.Unmarshal([]byte(storedCandidates), &candidates); err != nil {
			logger.Debug("Failed to deserialize candidate users")
			execResp.Status = common.ExecFailure
			execResp.FailureReason = failureReasonFailedToIdentifyUser
			return execResp, nil
		}

		// Filter candidates using all current search attributes
		candidates = filterUsersByAttributes(candidates, userSearchAttributes)
	}

	// Evaluate filtered result
	switch len(candidates) {
	case 0:
		logger.Debug("No matching users after filtering")
		execResp.Status = common.ExecFailure
		execResp.FailureReason = failureReasonUserNotFound
		return execResp, nil
	case 1:
		// Unique user found
		execResp.RuntimeData[userAttributeUserID] = candidates[0].UserID
		execResp.Status = common.ExecComplete
		logger.Debug("User resolved successfully",
			log.String("userID", log.MaskString(candidates[0].UserID)))
		return execResp, nil
	default:
		// Still ambiguous — serialize candidates and request more input
		candidatesJSON, err := json.Marshal(candidates)
		if err != nil {
			logger.Debug("Failed to serialize candidate users")
			execResp.Status = common.ExecFailure
			execResp.FailureReason = failureReasonFailedToIdentifyUser
			return execResp, nil
		}
		execResp.RuntimeData[runtimeKeyCandidateUsers] = string(candidatesJSON)
		execResp.Status = common.ExecUserInputRequired

		// Extract distinct attribute values from candidates and forward as input options.
		// This allows downstream prompt nodes to render dropdowns instead of free-text inputs
		// when the flow definition uses a select-type input.
		execResp.ForwardedData = map[string]interface{}{
			common.ForwardedDataKeyInputs: extractDisambiguationOptions(candidates),
		}

		logger.Debug("Multiple users still match, requesting additional attributes",
			log.Int("candidateCount", len(candidates)))
		return execResp, nil
	}
}

// filterUsersByAttributes filters users by matching their attributes against the provided filters.
func filterUsersByAttributes(users []*userprovider.User, filters map[string]interface{}) []*userprovider.User {
	var matched []*userprovider.User
	for _, u := range users {
		var attrs map[string]interface{}
		if len(u.Attributes) > 0 {
			if err := json.Unmarshal(u.Attributes, &attrs); err != nil {
				continue
			}
		}

		allMatch := true
		for key, expected := range filters {
			if slices.Contains(nonSearchableInputs, key) {
				continue
			}

			expectedStr, eOk := expected.(string)
			if !eOk {
				continue
			}

			// Check top-level User fields first
			switch key {
			case "userType":
				if u.UserType != expectedStr {
					allMatch = false
				}
			case "ouHandle":
				if u.OUHandle != expectedStr {
					allMatch = false
				}
			default:
				// Check in JSON attributes
				if attrs == nil {
					allMatch = false
				} else if value, ok := attrs[key]; !ok {
					allMatch = false
				} else if valueStr, vOk := value.(string); !vOk || valueStr != expectedStr {
					allMatch = false
				}
			}

			if !allMatch {
				break
			}
		}

		if allMatch {
			matched = append(matched, u)
		}
	}
	return matched
}

// extractDisambiguationOptions extracts distinct attribute values from candidate users
// and returns them as []common.Input with Options populated. This allows downstream prompt
// nodes to render dropdowns when enriched via ForwardedData.
func extractDisambiguationOptions(candidates []*userprovider.User) []common.Input {
	// Collect distinct values per attribute key (including top-level fields)
	optionsMap := make(map[string]map[string]struct{})

	for _, u := range candidates {
		// Top-level fields
		if u.UserType != "" {
			if optionsMap["userType"] == nil {
				optionsMap["userType"] = make(map[string]struct{})
			}
			optionsMap["userType"][u.UserType] = struct{}{}
		}
		if u.OUHandle != "" {
			if optionsMap["ouHandle"] == nil {
				optionsMap["ouHandle"] = make(map[string]struct{})
			}
			optionsMap["ouHandle"][u.OUHandle] = struct{}{}
		}

		// JSON attributes
		var attrs map[string]interface{}
		if len(u.Attributes) > 0 {
			if err := json.Unmarshal(u.Attributes, &attrs); err != nil {
				continue
			}
		}
		for key, value := range attrs {
			if slices.Contains(nonSearchableInputs, key) {
				continue
			}
			if valueStr, ok := value.(string); ok && valueStr != "" {
				if optionsMap[key] == nil {
					optionsMap[key] = make(map[string]struct{})
				}
				optionsMap[key][valueStr] = struct{}{}
			}
		}
	}

	// Convert to []common.Input — only include attributes with more than one distinct value
	// (single-value attributes don't help with disambiguation)
	var inputs []common.Input
	for key, valuesSet := range optionsMap {
		if len(valuesSet) <= 1 {
			continue
		}
		options := make([]string, 0, len(valuesSet))
		for v := range valuesSet {
			options = append(options, v)
		}
		inputs = append(inputs, common.Input{
			Identifier: key,
			Options:    options,
		})
	}

	return inputs
}
