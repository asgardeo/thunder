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

package application

import "github.com/asgardeo/thunder/internal/system/error/serviceerror"

// Client errors for application operations.
var (
	// ErrorApplicationNotFound is the error returned when an application is not found.
	ErrorApplicationNotFound = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "APP-1001",
		Error:            "Application not found",
		ErrorDescription: "The requested application could not be found",
	}
	// ErrorInvalidApplicationID is the error returned when an invalid application ID is provided.
	ErrorInvalidApplicationID = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "APP-1002",
		Error:            "Invalid application ID",
		ErrorDescription: "The provided application ID is invalid or empty",
	}
	// ErrorInvalidClientID is the error returned when an invalid client ID is provided.
	ErrorInvalidClientID = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "APP-1003",
		Error:            "Invalid client ID",
		ErrorDescription: "The provided client ID is invalid or empty",
	}
	// ErrorInvalidApplicationName is the error returned when an invalid application name is provided.
	ErrorInvalidApplicationName = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "APP-1004",
		Error:            "Invalid application name",
		ErrorDescription: "The provided application name is invalid or empty",
	}
	// ErrorInvalidApplicationURL is the error returned when an invalid application URL is provided.
	ErrorInvalidApplicationURL = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "APP-1005",
		Error:            "Invalid application URL",
		ErrorDescription: "The provided application URL is not a valid URI",
	}
	// ErrorInvalidLogoURL is the error returned when an invalid logo URL is provided.
	ErrorInvalidLogoURL = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "APP-1006",
		Error:            "Invalid logo URL",
		ErrorDescription: "The provided logo URL is not a valid URI",
	}
	// ErrorInvalidAuthFlowGraphID is the error returned when an invalid auth flow graph ID is provided.
	ErrorInvalidAuthFlowGraphID = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "APP-1007",
		Error:            "Invalid auth flow graph ID",
		ErrorDescription: "The provided authentication flow graph ID is invalid",
	}
	// ErrorInvalidRegistrationFlowGraphID is the error returned when an invalid registration flow graph ID
	// is provided.
	ErrorInvalidRegistrationFlowGraphID = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "APP-1008",
		Error:            "Invalid registration flow graph ID",
		ErrorDescription: "The provided registration flow graph ID is invalid",
	}
	// ErrorInvalidInboundAuthConfig is the error returned when invalid inbound auth config is provided.
	ErrorInvalidInboundAuthConfig = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "APP-1009",
		Error:            "Invalid inbound authentication configuration",
		ErrorDescription: "The provided inbound authentication configuration is invalid",
	}
	// ErrorInvalidGrantType is the error returned when an invalid grant type is provided.
	ErrorInvalidGrantType = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "APP-1010",
		Error:            "Invalid grant type",
		ErrorDescription: "One or more provided grant types are invalid",
	}
	// ErrorInvalidResponseType is the error returned when an invalid response type is provided.
	ErrorInvalidResponseType = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "APP-1011",
		Error:            "Invalid response type",
		ErrorDescription: "One or more provided response types are invalid",
	}
	// ErrorInvalidRedirectURI is the error returned when an invalid redirect URI is provided.
	ErrorInvalidRedirectURI = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "APP-1012",
		Error:            "Invalid redirect URI",
		ErrorDescription: "One or more provided redirect URIs are not valid URIs",
	}
	// ErrorInvalidTokenEndpointAuthMethod is the error returned when an invalid token endpoint auth method
	// is provided.
	ErrorInvalidTokenEndpointAuthMethod = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "APP-1013",
		Error:            "Invalid token endpoint authentication method",
		ErrorDescription: "The provided token endpoint authentication method is invalid",
	}
	// ErrorInvalidCertificateType is the error returned when an invalid certificate type is provided.
	ErrorInvalidCertificateType = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "APP-1014",
		Error:            "Invalid certificate type",
		ErrorDescription: "The provided certificate type is not supported",
	}
	// ErrorInvalidCertificateValue is the error returned when an invalid certificate value is provided.
	ErrorInvalidCertificateValue = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "APP-1015",
		Error:            "Invalid certificate value",
		ErrorDescription: "The provided certificate value is invalid",
	}
	// ErrorInvalidJWKSURI is the error returned when an invalid JWKS URI is provided.
	ErrorInvalidJWKSURI = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "APP-1016",
		Error:            "Invalid JWKS URI",
		ErrorDescription: "The provided JWKS URI is not a valid URI",
	}
	// ErrorApplicationNil is the error returned when the application object is nil.
	ErrorApplicationNil = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "APP-1017",
		Error:            "Application is nil",
		ErrorDescription: "The provided application object is nil",
	}
	// ErrorInvalidRequestFormat is the error returned when the request format is invalid.
	ErrorInvalidRequestFormat = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "APP-1018",
		Error:            "Invalid request format",
		ErrorDescription: "The request body is malformed or contains invalid data",
	}
	// ErrorCertificateClientError is the error returned when a certificate operation fails due to client error.
	ErrorCertificateClientError = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "APP-1019",
		Error:            "Certificate operation failed",
		ErrorDescription: "An error occurred while processing the application certificate",
	}
	// ErrorApplicationAlreadyExistsWithName is the error returned when an application with the same name
	// already exists.
	ErrorApplicationAlreadyExistsWithName = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "APP-1020",
		Error:            "Application already exists",
		ErrorDescription: "An application with the same name already exists",
	}
	// ErrorApplicationAlreadyExistsWithClientID is the error returned when an application with the same client ID
	// already exists.
	ErrorApplicationAlreadyExistsWithClientID = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "APP-1021",
		Error:            "Application with client ID already exists",
		ErrorDescription: "An application with the same client ID already exists",
	}
	// ErrorJWKSUriNotHTTPS is the error returned when jwks_uri does not use HTTPS.
	ErrorJWKSUriNotHTTPS = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "APP-1022",
		Error:            "Invalid JWKS URI scheme",
		ErrorDescription: "'jwks_uri' must use HTTPS scheme",
	}
	// ErrorInvalidPublicClientConfiguration is the generic error returned for public client configuration issues.
	ErrorInvalidPublicClientConfiguration = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "APP-1023",
		Error:            "Invalid public client configuration",
		ErrorDescription: "The public client configuration is invalid",
	}
	// ErrorInvalidOAuthConfiguration is the generic error returned for OAuth configuration issues.
	ErrorInvalidOAuthConfiguration = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "APP-1024",
		Error:            "Invalid OAuth configuration",
		ErrorDescription: "The OAuth configuration is invalid",
	}
)

// Server errors for application operations.
var (
	// ErrorInternalServerError is the error returned when an internal server error occurs.
	ErrorInternalServerError = serviceerror.ServiceError{
		Type:             serviceerror.ServerErrorType,
		Code:             "APP-5001",
		Error:            "Internal server error",
		ErrorDescription: "An unexpected error occurred while processing the request",
	}
	// ErrorCertificateServerError is the error returned when a certificate operation fails due to server error.
	ErrorCertificateServerError = serviceerror.ServiceError{
		Type:             serviceerror.ServerErrorType,
		Code:             "APP-5002",
		Error:            "Certificate operation failed",
		ErrorDescription: "An error occurred while performing the certificate operation",
	}
)
