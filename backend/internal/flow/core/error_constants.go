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

package core

import (
	"github.com/thunder-id/thunderid/internal/system/error/serviceerror"
	"github.com/thunder-id/thunderid/internal/system/i18n/core"
)

// Define core flow errors

// Executor prerequisite not met error
var ErrExecutorPrerequisiteNotMet = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FLC-1001",
	Error: core.I18nMessage{
		Key:          "error.flow.core.executor_prerequisite_not_met",
		DefaultValue: "A prerequisite for the executor was not met",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "error.flow.core.executor_prerequisite_not_met_description",
		DefaultValue: "One or more prerequisites required for the executor were not satisfied. Please check the inputs and try again.",
	},
}

// Invalid action provided in prompt node error
var ErrInvalidActionProvided = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FLC-1002",
	Error: core.I18nMessage{
		Key:          "error.flow.core.prompt_invalid_action",
		DefaultValue: "Invalid action provided",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "error.flow.core.prompt_invalid_action_description",
		DefaultValue: "The action provided is not valid for the current flow step",
	},
}
