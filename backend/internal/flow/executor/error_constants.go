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
	"fmt"

	"github.com/thunder-id/thunderid/internal/system/error/serviceerror"
	"github.com/thunder-id/thunderid/internal/system/i18n/core"
)

// Identity / user lookup errors (FET-10xxx)

// ErrUserNotFound is returned when the user is not found in the system.
var ErrUserNotFound = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-10001",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.user_not_found",
		DefaultValue: "User not found",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.user_not_found_desc",
		DefaultValue: "The user could not be found in the system",
	},
}

// ErrFailedToIdentifyUser is returned when the user cannot be identified.
var ErrFailedToIdentifyUser = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-10002",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.failed_to_identify_user",
		DefaultValue: "Failed to identify user",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.failed_to_identify_user_desc",
		DefaultValue: "Unable to identify the user with the provided information",
	},
}

// ErrAmbiguousUserIdentity is returned when the user identity is ambiguous.
var ErrAmbiguousUserIdentity = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-10003",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.ambiguous_user_identity",
		DefaultValue: "Ambiguous user identity",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.ambiguous_user_identity_desc",
		DefaultValue: "User identity is ambiguous and cannot be determined",
	},
}

// Authentication — credentials errors (FET-20xxx)

// ErrUserNotAuthenticated is returned when the user is not authenticated.
var ErrUserNotAuthenticated = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-20001",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.user_not_authenticated",
		DefaultValue: "User is not authenticated",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.user_not_authenticated_desc",
		DefaultValue: "The user has not been authenticated in this flow",
	},
}

// ErrInvalidCredentials is returned when the provided credentials are invalid.
var ErrInvalidCredentials = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-20002",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.invalid_credentials",
		DefaultValue: "Invalid credentials provided",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.invalid_credentials_desc",
		DefaultValue: "The credentials provided are invalid",
	},
}

// ErrUserAuthFailed is returned when user authentication fails.
var ErrUserAuthFailed = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-20003",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.user_auth_failed",
		DefaultValue: "User authentication failed",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.user_auth_failed_desc",
		DefaultValue: "An error occurred while authenticating the user",
	},
}

// ErrUserAlreadyExists is returned when the user already exists in the system.
var ErrUserAlreadyExists = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-20004",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.user_already_exists",
		DefaultValue: "User already exists",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.user_already_exists_desc",
		DefaultValue: "A user already exists with the provided attributes",
	},
}

// Authentication — OTP errors (FET-200xx)

// ErrInvalidOTP is returned when the provided OTP is invalid.
var ErrInvalidOTP = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-20010",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.invalid_otp",
		DefaultValue: "Invalid OTP provided",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.invalid_otp_desc",
		DefaultValue: "The one-time password provided is invalid or has expired",
	},
}

// ErrMaxOTPAttemptsReached is returned when the maximum OTP attempts are reached.
var ErrMaxOTPAttemptsReached = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-20011",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.max_otp_attempts_reached",
		DefaultValue: "Maximum OTP attempts reached",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.max_otp_attempts_reached_desc",
		DefaultValue: "The maximum number of OTP verification attempts has been reached",
	},
}

// Authentication — Magic Link errors (FET-200xx)

// ErrInvalidMagicLinkToken is returned when the magic link token is invalid.
var ErrInvalidMagicLinkToken = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-20020",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.invalid_magic_link_token",
		DefaultValue: "Invalid magic link token",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.invalid_magic_link_token_desc",
		DefaultValue: "The magic link token is invalid or has expired",
	},
}

// Authentication — OAuth/OIDC errors (FET-200xx)

// ErrInvalidOAuthState is returned when the OAuth state parameter is invalid.
var ErrInvalidOAuthState = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-20030",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.invalid_oauth_state",
		DefaultValue: "Invalid OAuth state parameter",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.invalid_oauth_state_desc",
		DefaultValue: "The OAuth state parameter is invalid or does not match the expected value",
	},
}

// ErrNonceMismatch is returned when the nonce in the ID token does not match.
var ErrNonceMismatch = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-20031",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.nonce_mismatch",
		DefaultValue: "Nonce mismatch in ID token",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.nonce_mismatch_desc",
		DefaultValue: "The nonce in the ID token claims does not match the expected value",
	},
}

// ErrInvalidOAuthCode is returned when the OAuth authorization code is invalid.
var ErrInvalidOAuthCode = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-20032",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.invalid_oauth_code",
		DefaultValue: "Invalid OAuth authorization code",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.invalid_oauth_code_desc",
		DefaultValue: "The OAuth authorization code is invalid or could not be exchanged for tokens",
	},
}

// ErrInvalidFederatedUser is returned when the federated user information is invalid during authentication.
var ErrInvalidFederatedUser = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-20033",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.invalid_federated_user",
		DefaultValue: "Invalid federated user",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.invalid_federated_user_desc",
		DefaultValue: "The federated user information is invalid or inconsistent",
	},
}

// Authentication — Passkey errors (FET-200xx)

// ErrInvalidPasskey is returned when the passkey credentials are invalid.
var ErrInvalidPasskey = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-20040",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.invalid_passkey",
		DefaultValue: "Invalid passkey credentials",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.invalid_passkey_desc",
		DefaultValue: "The passkey credentials provided are invalid",
	},
}

// ErrNoRegisteredPasskeys is returned when no registered passkeys are found for the user.
var ErrNoRegisteredPasskeys = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-20041",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.no_registered_passkeys",
		DefaultValue: "No registered passkeys found",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.no_registered_passkeys_desc",
		DefaultValue: "No registered passkeys were found for the user",
	},
}

// ErrUserIDRequiredForPasskeyReg is returned when user ID is missing for passkey registration.
var ErrUserIDRequiredForPasskeyReg = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-20042",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.user_id_required_for_passkey_reg",
		DefaultValue: "User ID missing for passkey registration",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.user_id_required_for_passkey_reg_desc",
		DefaultValue: "A user ID is required to register a passkey",
	},
}

// ErrPasskeyRegistrationFailed is returned when passkey registration fails.
var ErrPasskeyRegistrationFailed = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-20043",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.passkey_registration_failed",
		DefaultValue: "Passkey registration failed",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.passkey_registration_failed_desc",
		DefaultValue: "An error occurred while registering the passkey",
	},
}

// ErrPasskeyAuthFailed is returned when passkey authentication fails.	
var ErrPasskeyAuthFailed = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-20044",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.passkey_auth_failed",
		DefaultValue: "Passkey authentication failed",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.passkey_auth_failed_desc",
		DefaultValue: "An error occurred while authenticating with the passkey",
	},
}

// Provisioning / registration errors (FET-30xxx)

// ErrProvisioningUserAttrsMissing is returned when no user attributes are provided for provisioning.
var ErrProvisioningUserAttrsMissing = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-30001",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.provisioning_user_attrs_missing",
		DefaultValue: "No user attributes provided for provisioning",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.provisioning_user_attrs_missing_desc",
		DefaultValue: "User attributes are required to provision a new user",
	},
}

// ErrProvisioningFailed is returned when user provisioning fails.
var ErrProvisioningFailed = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-30002",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.provisioning_failed",
		DefaultValue: "User provisioning failed",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.provisioning_failed_desc",
		DefaultValue: "An error occurred while provisioning the user",
	},
}

// ErrProvisioningAssignmentFailed is returned when group or role assignment fails during provisioning.
var ErrProvisioningAssignmentFailed = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-30003",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.provisioning_assignment_failed",
		DefaultValue: "Failed to assign groups and roles",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.provisioning_assignment_failed_desc",
		DefaultValue: "An error occurred while assigning groups and roles to the provisioned user",
	},
}

// ErrCrossOUProvisioningTargetMissing is returned when target OU is missing for cross-OU provisioning.
var ErrCrossOUProvisioningTargetMissing = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-30004",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.cross_ou_provisioning_target_missing",
		DefaultValue: "Target OU is not set for cross-OU provisioning",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.cross_ou_provisioning_target_missing_desc",
		DefaultValue: "A target organization unit must be specified for cross-OU user provisioning",
	},
}

// ErrUserAlreadyExistsInTargetOU is returned when the user already exists in the target organization.
var ErrUserAlreadyExistsInTargetOU = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-30005",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.user_already_exists_in_target_ou",
		DefaultValue: "User already exists in the target organization",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.user_already_exists_in_target_ou_desc",
		DefaultValue: "A user with the same identity already exists in the target organization unit",
	},
}

// ErrCannotProvisionAutomatically is returned when the user cannot be provisioned automatically.
var ErrCannotProvisionAutomatically = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-30006",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.cannot_provision_automatically",
		DefaultValue: "Cannot provision user automatically",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.cannot_provision_automatically_desc",
		DefaultValue: "The user cannot be provisioned automatically with the provided information",
	},
}

// ErrSelfRegistrationDisabled is returned when self-registration is not enabled.
var ErrSelfRegistrationDisabled = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-30007",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.self_registration_disabled",
		DefaultValue: "Self-registration not enabled",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.self_registration_disabled_desc",
		DefaultValue: "Self-registration is not enabled for this application or user type",
	},
}

// Authorization errors (FET-40xxx)

// ErrInsufficientPermissions is returned when the user lacks required permissions.
var ErrInsufficientPermissions = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-40001",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.insufficient_permissions",
		DefaultValue: "Insufficient permissions",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.insufficient_permissions_desc",
		DefaultValue: "The user does not have sufficient permissions to perform this action",
	},
}

// ErrAuthorizationFailed is returned when authorization validation fails.
var ErrAuthorizationFailed = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-40002",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.authorization_failed",
		DefaultValue: "Authorization validation failed",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.authorization_failed_desc",
		DefaultValue: "Authorization validation failed for the current user",
	},
}

// Organization unit errors (FET-50xxx)

// ErrInvalidOU is returned when the selected organization unit is invalid.
var ErrInvalidOU = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-50001",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.invalid_ou",
		DefaultValue: "Selected organization unit is invalid",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.invalid_ou_desc",
		DefaultValue: "The selected organization unit is not valid for this operation",
	},
}

// ErrOUNotFound is returned when the organization unit is not found.
var ErrOUNotFound = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-50002",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.ou_not_found",
		DefaultValue: "Organization unit not found",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.ou_not_found_desc",
		DefaultValue: "The selected organization unit does not exist",
	},
}

// ErrOUNameConflict is returned when an organization unit with the same name already exists.
var ErrOUNameConflict = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-50003",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.ou_name_conflict",
		DefaultValue: "Organization unit with the same name already exists",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.ou_name_conflict_desc",
		DefaultValue: "An organization unit with the same name already exists in this context",
	},
}

// ErrOUHandleConflict is returned when an organization unit with the same handle already exists.
var ErrOUHandleConflict = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-50004",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.ou_handle_conflict",
		DefaultValue: "Organization unit with the same handle already exists",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.ou_handle_conflict_desc",
		DefaultValue: "An organization unit with the same handle already exists in this context",
	},
}

// ErrOUCreationPrereqFailed is returned when prerequisites validation fails for OU creation.
var ErrOUCreationPrereqFailed = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-50005",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.ou_creation_prereq_failed",
		DefaultValue: "Prerequisites validation failed for OU creation",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.ou_creation_prereq_failed_desc",
		DefaultValue: "The prerequisites for creating an organization unit have not been met",
	},
}

// ErrOUResolutionFailed is returned when the organization unit cannot be resolved.
var ErrOUResolutionFailed = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-50006",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.ou_resolution_failed",
		DefaultValue: "Failed to resolve organization unit",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.ou_resolution_failed_desc",
		DefaultValue: "Unable to resolve the organization unit for the current context",
	},
}

// ErrOUCreationFailed is returned when organization unit creation fails.
var ErrOUCreationFailed = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-50007",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.ou_creation_failed",
		DefaultValue: "Organization unit creation failed",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.ou_creation_failed_desc",
		DefaultValue: "An error occurred while creating the organization unit",
	},
}

// Notification / communication errors (FET-60xxx)

// ErrEmailSendFailed is returned when the email fails to send.
var ErrEmailSendFailed = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-60001",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.email_send_failed",
		DefaultValue: "Failed to send email",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.email_send_failed_desc",
		DefaultValue: "An error occurred while sending the email",
	},
}

// ErrEmailRecipientMissing is returned when the email recipient is not provided.
var ErrEmailRecipientMissing = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-60002",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.email_recipient_missing",
		DefaultValue: "Email recipient is required",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.email_recipient_missing_desc",
		DefaultValue: "An email recipient must be provided to send the notification",
	},
}

// ErrEmailServiceNotConfigured is returned when the email service is not configured.
var ErrEmailServiceNotConfigured = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-60003",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.email_service_not_configured",
		DefaultValue: "Email service is not configured",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.email_service_not_configured_desc",
		DefaultValue: "The email notification service has not been configured",
	},
}

// ErrSMSRecipientMissing is returned when the SMS recipient is not provided.
var ErrSMSRecipientMissing = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-60004",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.sms_recipient_missing",
		DefaultValue: "SMS recipient is required",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.sms_recipient_missing_desc",
		DefaultValue: "An SMS recipient must be provided to send the notification",
	},
}

// ErrSMSInvalidPhone is returned when the SMS recipient phone number is invalid.
var ErrSMSInvalidPhone = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-60005",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.sms_invalid_phone",
		DefaultValue: "SMS recipient is not a valid phone number",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.sms_invalid_phone_desc",
		DefaultValue: "The provided SMS recipient is not a valid phone number",
	},
}

// ErrSMSTemplateMissing is returned when the SMS template is not provided.
var ErrSMSTemplateMissing = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-60006",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.sms_template_missing",
		DefaultValue: "SMS template is required",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.sms_template_missing_desc",
		DefaultValue: "An SMS template must be provided to send the notification",
	},
}

// ErrSMSProviderNotConfigured is returned when the SMS provider is not configured.
var ErrSMSProviderNotConfigured = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-60007",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.sms_provider_not_configured",
		DefaultValue: "SMS notification provider is not configured",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.sms_provider_not_configured_desc",
		DefaultValue: "The SMS notification provider has not been configured or is errourneous",
	},
}

// Configuration / prerequisites errors (FET-70xxx)

// ErrPrerequisitesFailed is returned when prerequisites validation fails.
var ErrPrerequisitesFailed = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-70001",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.prerequisites_failed",
		DefaultValue: "Prerequisites validation failed",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.prerequisites_failed_desc",
		DefaultValue: "The prerequisites for this operation have not been met",
	},
}

// ErrUserIDMissingInContext is returned when the user ID is not found in the flow context.
var ErrUserIDMissingInContext = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-70002",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.user_id_missing_in_context",
		DefaultValue: "User ID not found",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.user_id_missing_in_context_desc",
		DefaultValue: "The user ID could not be resolved from the current flow context",
	},
}

// ErrCredentialInputMissing is returned when no credential input is configured.
var ErrCredentialInputMissing = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-70003",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.credential_input_missing",
		DefaultValue: "No credential input configured",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.credential_input_missing_desc",
		DefaultValue: "No credential input has been configured for the credential setter",
	},
}

// ErrCredentialInputInvalid is returned when the credential input configuration is invalid.
var ErrCredentialInputInvalid = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-70004",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.credential_input_invalid",
		DefaultValue: "Invalid credential input configuration",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.credential_input_invalid_desc",
		DefaultValue: "The credential input configuration is invalid for the credential setter",
	},
}

// ErrCredentialValueEmpty is returned when the credential value is empty.
var ErrCredentialValueEmpty = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-70005",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.credential_value_empty",
		DefaultValue: "Credential value is empty",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.credential_value_empty_desc",
		DefaultValue: "The credential value must not be empty for the credential setter",
	},
}

// ErrCredentialSetFailed is returned when setting credentials fails.
var ErrCredentialSetFailed = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-70006",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.credential_set_failed",
		DefaultValue: "Failed to set credentials",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.credential_set_failed_desc",
		DefaultValue: "An error occurred while setting the user credentials",
	},
}

// ErrAttributeCollectFailed is returned when updating user attributes fails.
var ErrAttributeCollectFailed = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-70007",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.attribute_collect_failed",
		DefaultValue: "Failed to update user attributes",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.attribute_collect_failed_desc",
		DefaultValue: "An error occurred while updating the user attributes",
	},
}

// ErrHTTPRequestConfigInvalid is returned when the HTTP request executor configuration is invalid.
var ErrHTTPRequestConfigInvalid = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-70008",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.http_request_config_invalid",
		DefaultValue: "Configuration error",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.http_request_config_invalid_desc",
		DefaultValue: "The HTTP request executor configuration is invalid",
	},
}

// ErrAuthNotAvailableForApp is returned when authentication is not available for the application.
var ErrAuthNotAvailableForApp = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-70010",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.auth_not_available_for_app",
		DefaultValue: "Authentication not available for this application",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.auth_not_available_for_app_desc",
		DefaultValue: "The requested authentication method is not available for this application",
	},
}

// ErrSelfRegNotAvailableForApp is returned when self-registration is not available for the application.
var ErrSelfRegNotAvailableForApp = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-70011",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.self_reg_not_available_for_app",
		DefaultValue: "Self-registration not available for this application",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.self_reg_not_available_for_app_desc",
		DefaultValue: "Self-registration is not available for this application",
	},
}

// ErrNoValidUserTypes is returned when no valid user types are available for the flow.
var ErrNoValidUserTypes = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-70012",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.no_valid_user_types",
		DefaultValue: "No valid user types available",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.no_valid_user_types_desc",
		DefaultValue: "There are no valid user types configured for this flow",
	},
}

// ErrUserTypeNotAllowed is returned when the user type is not allowed for the flow.
var ErrUserTypeNotAllowed = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-70013",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.user_type_not_allowed",
		DefaultValue: "User type not allowed for this flow",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.user_type_not_allowed_desc",
		DefaultValue: "The selected user type is not allowed for this flow",
	},
}

// ErrInvalidUserType is returned when the user type is invalid.
var ErrInvalidUserType = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-70014",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.invalid_user_type",
		DefaultValue: "Invalid user type",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.invalid_user_type_desc",
		DefaultValue: "The provided user type is not valid",
	},
}

// ErrNoUserTypesAvailable is returned when no user types are available.
var ErrNoUserTypesAvailable = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-70015",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.no_user_types_available",
		DefaultValue: "No user types available",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.no_user_types_available_desc",
		DefaultValue: "No user types are currently available",
	},
}

// ErrUserTypeRetrievalFailed is returned when user type retrieval fails.
var ErrUserTypeRetrievalFailed = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-70016",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.user_type_retrieval_failed",
		DefaultValue: "Failed to retrieve user types",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.user_type_retrieval_failed_desc",
		DefaultValue: "An error occurred while retrieving available user types",
	},
}

// ErrUserTypeNotValidForOU is returned when the user type is not valid for the selected OU.
var ErrUserTypeNotValidForOU = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-70017",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.user_type_not_valid_for_ou",
		DefaultValue: "User type is not valid",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.user_type_not_valid_for_ou_desc",
		DefaultValue: "The selected user type is not valid for the chosen organization unit",
	},
}

// ErrAttributeNotUnique is returned when an attribute value already exists.
var ErrAttributeNotUnique = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-70020",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.attribute_not_unique",
		DefaultValue: "Attribute value already exists",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.attribute_not_unique_desc",
		DefaultValue: "The provided attribute value already exists and must be unique",
	},
}

// ErrAttributeRetrievalFailed is returned when user attribute retrieval fails.
var ErrAttributeRetrievalFailed = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-70021",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.attribute_retrieval_failed",
		DefaultValue: "Failed to retrieve user attributes",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.attribute_retrieval_failed_desc",
		DefaultValue: "An error occurred while retrieving user attributes",
	},
}

// ErrCredentialProcessingFailed is returned when credential processing fails.
var ErrCredentialProcessingFailed = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-70025",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.credential_processing_failed",
		DefaultValue: "Failed to process credentials",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.credential_processing_failed_desc",
		DefaultValue: "An error occurred while processing the credentials",
	},
}

// ErrInvalidAction is returned when an invalid action is provided to a prompt node.
var ErrInvalidAction = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-70030",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.invalid_action",
		DefaultValue: "Invalid action provided",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.invalid_action_desc",
		DefaultValue: "The action provided is not valid for the current flow step",
	},
}

// Consent errors (FET-80xxx)

// ErrConsentPrereqFailed is returned when prerequisites validation fails for consent.
var ErrConsentPrereqFailed = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-80001",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.consent_prereq_failed",
		DefaultValue: "Prerequisites validation failed for consent",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.consent_prereq_failed_desc",
		DefaultValue: "The prerequisites for the consent executor have not been met",
	},
}

// ErrConsentDenied is returned when the user denies consent.
var ErrConsentDenied = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-80002",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.consent_denied",
		DefaultValue: "User denied consent",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.consent_denied_desc",
		DefaultValue: "The user has denied the required consent",
	},
}

// ErrConsentDecisionsMissing is returned when consent decisions input is missing.
var ErrConsentDecisionsMissing = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-80003",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.consent_decisions_missing",
		DefaultValue: "Consent decisions input is missing or empty",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.consent_decisions_missing_desc",
		DefaultValue: "The consent decisions input is missing or empty",
	},
}

// ErrConsentDecisionsParseFail is returned when consent decisions cannot be parsed.
var ErrConsentDecisionsParseFail = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-80004",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.consent_decisions_parse",
		DefaultValue: "Failed to parse consent decisions",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.consent_decisions_parse_desc",
		DefaultValue: "The consent decisions input could not be parsed",
	},
}

// ErrConsentPromptTimedOut is returned when the consent prompt times out.
var ErrConsentPromptTimedOut = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-80005",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.consent_prompt_timed_out",
		DefaultValue: "Consent prompt has timed out",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.consent_prompt_timed_out_desc",
		DefaultValue: "The consent prompt has timed out without a response",
	},
}

// ErrConsentRecordFailed is returned when recording consent fails.
var ErrConsentRecordFailed = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-80006",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.consent_record_failed",
		DefaultValue: "Failed to record consent",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.consent_record_failed_desc",
		DefaultValue: "An error occurred while recording the user consent",
	},
}

// ErrConsentResolutionFailed is returned when resolving consent fails.
var ErrConsentResolutionFailed = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-80007",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.consent_resolution_failed",
		DefaultValue: "Failed to resolve consent",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.consent_resolution_failed_desc",
		DefaultValue: "An error occurred while resolving the user consent requirements",
	},
}

// External / HTTP errors (FET-90xxx)

// ErrHTTPRequestFailed is returned when the HTTP request executor fails.
var ErrHTTPRequestFailed = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-90001",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.http_request_failed",
		DefaultValue: "HTTP request executor failed",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.http_request_failed_desc",
		DefaultValue: "The HTTP request executor failed to complete the request",
	},
}

// ErrInvalidInviteToken is returned when the invite token is invalid.
var ErrInvalidInviteToken = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-90010",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.invalid_invite_token",
		DefaultValue: "Invalid invite token",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.invalid_invite_token_desc",
		DefaultValue: "The provided invite token is invalid or has expired",
	},
}

// ErrInviteTokenGenerationFailed is returned when generating an invite token fails.
var ErrInviteTokenGenerationFailed = serviceerror.ServiceError{
	Type: serviceerror.ClientErrorType,
	Code: "FET-90011",
	Error: core.I18nMessage{
		Key:          "flows.executor.errors.invite_token_generation_failed",
		DefaultValue: "Failed to generate invite token",
	},
	ErrorDescription: core.I18nMessage{
		Key:          "flows.executor.errors.invite_token_generation_failed_desc",
		DefaultValue: "An error occurred while generating the invite token",
	},
}

// ErrAttributeNotUniqueFor returns a ServiceError for a specific attribute that is not unique.
func ErrAttributeNotUniqueFor(attrName string) *serviceerror.ServiceError {
	e := ErrAttributeNotUnique
	e.ErrorDescription.DefaultValue = fmt.Sprintf("The attribute '%s' already exists and must be unique", attrName)
	return &e
}

// ErrMaxOTPAttemptsReachedFor returns a ServiceError for reaching the maximum OTP attempts.
func ErrMaxOTPAttemptsReachedFor(count int) *serviceerror.ServiceError {
	e := ErrMaxOTPAttemptsReached
	e.ErrorDescription.DefaultValue = fmt.Sprintf("The maximum number of OTP verification attempts (%d) has been reached", count)
	return &e
}

// ErrFailedToRetrieveAttribute returns a ServiceError for failing to retrieve a specific attribute.
func ErrFailedToRetrieveAttribute(attrName string) *serviceerror.ServiceError {
	e := ErrAttributeRetrievalFailed
	e.ErrorDescription.DefaultValue = fmt.Sprintf("An error occurred while retrieving the attribute '%s'", attrName)
	return &e
}
