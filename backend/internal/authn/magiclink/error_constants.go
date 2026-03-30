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

package magiclink

import (
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/i18n/core"
)

// Client errors for Magic Link authentication service.
var (
	// ErrorInvalidRecipient is the error returned when the provided recipient is invalid.
	ErrorInvalidRecipient = serviceerror.I18nServiceError{
		Type: serviceerror.ClientErrorType,
		Code: "AUTHN-ML-1001",
		Error: core.I18nMessage{
			Key:          "magiclink.error.invalid_recipient",
			DefaultValue: "Invalid recipient",
		},
		ErrorDescription: core.I18nMessage{
			Key:          "magiclink.error.invalid_recipient_description",
			DefaultValue: "The provided recipient email is invalid or empty",
		},
	}
	// ErrorInvalidToken is the error returned when the provided magic link token is invalid.
	ErrorInvalidToken = serviceerror.I18nServiceError{
		Type: serviceerror.ClientErrorType,
		Code: "AUTHN-ML-1002",
		Error: core.I18nMessage{
			Key:          "magiclink.error.invalid_token",
			DefaultValue: "Invalid token",
		},
		ErrorDescription: core.I18nMessage{
			Key:          "magiclink.error.invalid_token_description",
			DefaultValue: "The provided magic link token is invalid",
		},
	}
	// ErrorExpiredToken is the error returned when the magic link token has expired.
	ErrorExpiredToken = serviceerror.I18nServiceError{
		Type: serviceerror.ClientErrorType,
		Code: "AUTHN-ML-1003",
		Error: core.I18nMessage{
			Key:          "magiclink.error.expired_token",
			DefaultValue: "Expired token",
		},
		ErrorDescription: core.I18nMessage{
			Key:          "magiclink.error.expired_token_description",
			DefaultValue: "The magic link token has expired",
		},
	}
	// ErrorMalformedTokenClaims is the error returned when the token claims are malformed.
	ErrorMalformedTokenClaims = serviceerror.I18nServiceError{
		Type: serviceerror.ClientErrorType,
		Code: "AUTHN-ML-1004",
		Error: core.I18nMessage{
			Key:          "magiclink.error.malformed_token_claims",
			DefaultValue: "Malformed token claims",
		},
		ErrorDescription: core.I18nMessage{
			Key:          "magiclink.error.malformed_token_claims_description",
			DefaultValue: "The magic link token contains invalid or missing claims",
		},
	}
	// ErrorClientErrorWhileResolvingUser is the error returned when there is a client error while resolving the user.
	ErrorClientErrorWhileResolvingUser = serviceerror.I18nServiceError{
		Type: serviceerror.ClientErrorType,
		Code: "AUTHN-ML-1006",
		Error: core.I18nMessage{
			Key:          "magiclink.error.resolving_user",
			DefaultValue: "Error resolving user",
		},
		ErrorDescription: core.I18nMessage{
			Key:          "magiclink.error.resolving_user_description",
			DefaultValue: "An error occurred while resolving the user for the recipient",
		},
	}
)

// Server errors for Magic Link authentication service.
var (
	// ErrorTemplateRenderFailed is the error returned when template rendering fails.
	ErrorTemplateRenderFailed = serviceerror.I18nServiceError{
		Type: serviceerror.ServerErrorType,
		Code: "AUTHN-ML-5001",
		Error: core.I18nMessage{
			Key:          "magiclink.error.template_render_failed",
			DefaultValue: "Template render failed",
		},
		ErrorDescription: core.I18nMessage{
			Key:          "magiclink.error.template_render_failed_description",
			DefaultValue: "Failed to render the magic link email template",
		},
	}
	// ErrorEmailSendFailed is the error returned when sending the magic link email fails.
	ErrorEmailSendFailed = serviceerror.I18nServiceError{
		Type: serviceerror.ServerErrorType,
		Code: "AUTHN-ML-5002",
		Error: core.I18nMessage{
			Key:          "magiclink.error.email_send_failed",
			DefaultValue: "Email send failed",
		},
		ErrorDescription: core.I18nMessage{
			Key:          "magiclink.error.email_send_failed_description",
			DefaultValue: "Failed to send the magic link email",
		},
	}
	// ErrorTokenGenerationFailed is the error returned when token generation fails.
	ErrorTokenGenerationFailed = serviceerror.I18nServiceError{
		Type: serviceerror.ServerErrorType,
		Code: "AUTHN-ML-5003",
		Error: core.I18nMessage{
			Key:          "magiclink.error.token_generation_failed",
			DefaultValue: "Token generation failed",
		},
		ErrorDescription: core.I18nMessage{
			Key:          "magiclink.error.token_generation_failed_description",
			DefaultValue: "Failed to generate the magic link token",
		},
	}
	// ErrorMagicLinkNotConfigured is the error returned when magic link service is not configured.
	ErrorMagicLinkNotConfigured = serviceerror.I18nServiceError{
		Type: serviceerror.ServerErrorType,
		Code: "AUTHN-ML-5004",
		Error: core.I18nMessage{
			Key:          "magiclink.error.not_configured",
			DefaultValue: "Magic link authentication is not configured",
		},
		ErrorDescription: core.I18nMessage{
			Key:          "magiclink.error.not_configured_description",
			DefaultValue: "Magic link service is not available",
		},
	}
)
