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
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

// Package importer provides functionality for importing resources into the server.
package importer

import (
	layoutmgt "github.com/asgardeo/thunder/internal/design/layout/mgt"
	thememgt "github.com/asgardeo/thunder/internal/design/theme/mgt"
	"github.com/asgardeo/thunder/internal/entitytype"
	"github.com/asgardeo/thunder/internal/idp"
	"github.com/asgardeo/thunder/internal/ou"
	"github.com/asgardeo/thunder/internal/resource"
	"github.com/asgardeo/thunder/internal/role"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/i18n/core"
	"github.com/asgardeo/thunder/internal/user"
	"github.com/asgardeo/thunder/pkg/application"
	flowmgt "github.com/asgardeo/thunder/pkg/flow/mgt"
)

// notFoundErrorCodes is the set of service error codes that represent a resource-not-found condition
// across all domain packages used by the importer. Used to distinguish upsert fallback (create after
// update-not-found) from other update errors.
var notFoundErrorCodes = map[string]struct{}{
	application.ErrorApplicationNotFound.Code: {},
	idp.ErrorIDPNotFound.Code:                 {},
	flowmgt.ErrorFlowNotFound.Code:            {},
	ou.ErrorOrganizationUnitNotFound.Code:     {},
	entitytype.ErrorEntityTypeNotFound.Code:   {},
	role.ErrorRoleNotFound.Code:               {},
	resource.ErrorResourceServerNotFound.Code: {},
	thememgt.ErrorThemeNotFound.Code:          {},
	layoutmgt.ErrorLayoutNotFound.Code:        {},
	user.ErrorUserNotFound.Code:               {},
}

var (
	// ErrorInvalidImportRequest represents malformed import requests.
	ErrorInvalidImportRequest = serviceerror.ServiceError{
		Type:  serviceerror.ClientErrorType,
		Code:  "IMP-1001",
		Error: core.I18nMessage{Key: "error.import.invalidRequest", DefaultValue: "Invalid import request"},
		ErrorDescription: core.I18nMessage{
			Key:          "error.import.invalidRequest.description",
			DefaultValue: "The provided import request is invalid or malformed",
		},
	}

	// ErrorInvalidYAMLContent represents invalid YAML payloads.
	ErrorInvalidYAMLContent = serviceerror.ServiceError{
		Type:  serviceerror.ClientErrorType,
		Code:  "IMP-1002",
		Error: core.I18nMessage{Key: "error.import.invalidYaml", DefaultValue: "Invalid YAML content"},
		ErrorDescription: core.I18nMessage{
			Key:          "error.import.invalidYaml.description",
			DefaultValue: "The provided YAML content cannot be parsed",
		},
	}

	// ErrorTemplateResolutionFailed represents template resolution failures.
	ErrorTemplateResolutionFailed = serviceerror.ServiceError{
		Type: serviceerror.ClientErrorType,
		Code: "IMP-1003",
		Error: core.I18nMessage{
			Key:          "error.import.templateResolutionFailed",
			DefaultValue: "Template resolution failed",
		},
		ErrorDescription: core.I18nMessage{
			Key:          "error.import.templateResolutionFailed.description",
			DefaultValue: "Failed to resolve one or more template variables in YAML content",
		},
	}

	// ErrorAdapterNotConfigured represents missing runtime adapter wiring.
	ErrorAdapterNotConfigured = serviceerror.ServiceError{
		Type:  serviceerror.ClientErrorType,
		Code:  "IMP-1004",
		Error: core.I18nMessage{Key: "error.import.adapterNotConfigured", DefaultValue: "Adapter not configured"},
		ErrorDescription: core.I18nMessage{
			Key:          "error.import.adapterNotConfigured.description",
			DefaultValue: "The required resource adapter is not configured",
		},
	}
)
