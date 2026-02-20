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
	"errors"
	"fmt"
	"slices"

	authncm "github.com/asgardeo/thunder/internal/authn/common"
	authnoauth "github.com/asgardeo/thunder/internal/authn/oauth"
	"github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/internal/flow/core"
	"github.com/asgardeo/thunder/internal/idp"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	systemutils "github.com/asgardeo/thunder/internal/system/utils"
	"github.com/asgardeo/thunder/internal/userprovider"
	"github.com/asgardeo/thunder/internal/userschema"
)

const (
	oAuthLoggerComponentName            = "OAuthExecutor"
	errCannotProvisionUserAutomatically = "user not found and cannot provision automatically"
)

// OAuthTokenResponse represents the response from a OAuth token endpoint.
type OAuthTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token"`
	ExpiresIn    int    `json:"expires_in"`
}

// userInfoSkipAttributes contains the list of user info attributes to skip when mapping to context user.
var userInfoSkipAttributes = []string{"username", "sub", "id"}

// oAuthExecutorInterface defines the interface for OAuth authentication executors.
type oAuthExecutorInterface interface {
	core.ExecutorInterface
	BuildAuthorizeFlow(ctx *core.NodeContext, execResp *common.ExecutorResponse) error
	ProcessAuthFlowResponse(ctx *core.NodeContext, execResp *common.ExecutorResponse) error
	ExchangeCodeForToken(ctx *core.NodeContext, execResp *common.ExecutorResponse,
		code string) (*OAuthTokenResponse, error)
	GetUserInfo(ctx *core.NodeContext, execResp *common.ExecutorResponse,
		accessToken string) (map[string]string, error)
	GetInternalUser(sub string, execResp *common.ExecutorResponse) (*userprovider.User, error)
	ResolveContextUser(ctx *core.NodeContext, execResp *common.ExecutorResponse,
		sub string, internalUser *userprovider.User) (*authncm.AuthenticatedUser, error)
	GetIdpID(ctx *core.NodeContext) (string, error)
}

// oAuthExecutor implements the OAuthExecutorInterface for handling generic OAuth authentication flows.
type oAuthExecutor struct {
	core.ExecutorInterface
	authService       authnoauth.OAuthAuthnCoreServiceInterface
	idpService        idp.IDPServiceInterface
	userSchemaService userschema.UserSchemaServiceInterface
	logger            *log.Logger
}

var _ core.ExecutorInterface = (*oAuthExecutor)(nil)

// newOAuthExecutor creates a new instance of OAuthExecutor.
func newOAuthExecutor(
	name string,
	defaultInputs, prerequisites []common.Input,
	flowFactory core.FlowFactoryInterface,
	idpService idp.IDPServiceInterface,
	userSchemaService userschema.UserSchemaServiceInterface,
	authService authnoauth.OAuthAuthnCoreServiceInterface,
) oAuthExecutorInterface {
	if name == "" {
		name = ExecutorNameOAuth
	}
	if len(defaultInputs) == 0 {
		defaultInputs = []common.Input{
			{
				Identifier: userInputCode,
				Type:       "string",
				Required:   true,
			},
		}
	}
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, oAuthLoggerComponentName),
		log.String(log.LoggerKeyExecutorName, name))

	base := flowFactory.CreateExecutor(name, common.ExecutorTypeAuthentication,
		defaultInputs, prerequisites)

	return &oAuthExecutor{
		ExecutorInterface: base,
		authService:       authService,
		idpService:        idpService,
		userSchemaService: userSchemaService,
		logger:            logger,
	}
}

// Execute executes the OAuth authentication flow.
func (o *oAuthExecutor) Execute(ctx *core.NodeContext) (*common.ExecutorResponse, error) {
	logger := o.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Executing OAuth authentication executor")

	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	if ctx.FlowType != common.FlowTypeAuthentication && ctx.FlowType != common.FlowTypeRegistration {
		logger.Warn("Invalid flow type for OAuth executor. Skipping execution")
		execResp.Status = common.ExecComplete
		return execResp, nil
	}

	if !o.HasRequiredInputs(ctx, execResp) {
		logger.Debug("Required inputs for OAuth authentication executor is not provided")
		err := o.BuildAuthorizeFlow(ctx, execResp)
		if err != nil {
			return nil, err
		}
	} else {
		err := o.ProcessAuthFlowResponse(ctx, execResp)
		if err != nil {
			return nil, err
		}
	}

	logger.Debug("OAuth authentication executor execution completed",
		log.String("status", string(execResp.Status)),
		log.Bool("isAuthenticated", execResp.AuthenticatedUser.IsAuthenticated))

	return execResp, nil
}

// BuildAuthorizeFlow constructs the redirection to the external OAuth provider for user authentication.
func (o *oAuthExecutor) BuildAuthorizeFlow(ctx *core.NodeContext, execResp *common.ExecutorResponse) error {
	logger := o.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Initiating OAuth authentication flow")

	idpID, err := o.GetIdpID(ctx)
	if err != nil {
		return err
	}

	authorizeURL, svcErr := o.authService.BuildAuthorizeURL(idpID)
	if svcErr != nil {
		if svcErr.Type == serviceerror.ClientErrorType {
			execResp.Status = common.ExecFailure
			execResp.FailureReason = svcErr.ErrorDescription
			return nil
		}

		logger.Error("Failed to build authorize URL", log.String("errorCode", svcErr.Code),
			log.String("errorDescription", svcErr.ErrorDescription))
		return errors.New("failed to build authorize URL")
	}

	// Get the idp name for additional data
	idpName, err := o.getIDPName(idpID)
	if err != nil {
		return fmt.Errorf("failed to get idp name: %w", err)
	}

	// Set the response to redirect the user to the authorization URL.
	execResp.Status = common.ExecExternalRedirection
	execResp.RedirectURL = authorizeURL
	execResp.AdditionalData = map[string]string{
		common.DataIDPName: idpName,
	}

	return nil
}

// ProcessAuthFlowResponse processes the response from the OAuth authentication flow and authenticates the user.
func (o *oAuthExecutor) ProcessAuthFlowResponse(ctx *core.NodeContext,
	execResp *common.ExecutorResponse) error {
	logger := o.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Processing OAuth authentication response")

	code, ok := ctx.UserInputs[userInputCode]
	if !ok || code == "" {
		execResp.AuthenticatedUser = authncm.AuthenticatedUser{
			IsAuthenticated: false,
		}
		return nil
	}

	tokenResp, err := o.ExchangeCodeForToken(ctx, execResp, code)
	if err != nil {
		return err
	}
	if execResp.Status == common.ExecFailure {
		return nil
	}

	if tokenResp.Scope == "" {
		logger.Error("Scopes are empty in the token response")
		execResp.Status = common.ExecFailure
		execResp.AuthenticatedUser = authncm.AuthenticatedUser{
			IsAuthenticated: false,
		}
		return nil
	}

	userInfo, err := o.GetUserInfo(ctx, execResp, tokenResp.AccessToken)
	if err != nil {
		return err
	}
	if execResp.Status == common.ExecFailure {
		return nil
	}

	// Extract sub claim from user info
	sub, ok := userInfo[userAttributeSub]
	if !ok || sub == "" {
		execResp.Status = common.ExecFailure
		execResp.FailureReason = "sub claim not found in the response."
		return nil
	}

	internalUser, err := o.GetInternalUser(sub, execResp)
	if err != nil {
		return err
	}
	if execResp.Status == common.ExecFailure {
		return nil
	}

	contextUser, err := o.ResolveContextUser(ctx, execResp, sub, internalUser)
	if err != nil {
		return err
	}
	if execResp.Status == common.ExecFailure {
		return nil
	}
	if contextUser == nil {
		logger.Error("Failed to resolve context user after OAuth authentication")
		return errors.New("unexpected error occurred while resolving user")
	}

	contextUser.Attributes = o.getContextUserAttributes(execResp, userInfo)
	execResp.AuthenticatedUser = *contextUser

	return nil
}

// HasRequiredInputs checks if the required inputs are provided in the context and appends any
// missing inputs to the executor response. Returns true if required inputs are found, otherwise false.
func (o *oAuthExecutor) HasRequiredInputs(ctx *core.NodeContext, execResp *common.ExecutorResponse) bool {
	if code, ok := ctx.UserInputs[userInputCode]; ok && code != "" {
		return true
	}

	return o.ExecutorInterface.HasRequiredInputs(ctx, execResp)
}

// ExchangeCodeForToken exchanges the authorization code for an access token.
func (o *oAuthExecutor) ExchangeCodeForToken(ctx *core.NodeContext, execResp *common.ExecutorResponse,
	code string) (*OAuthTokenResponse, error) {
	logger := o.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Exchanging authorization code for a token")

	idpID, err := o.GetIdpID(ctx)
	if err != nil {
		return nil, err
	}

	tokenResp, svcErr := o.authService.ExchangeCodeForToken(idpID, code, true)
	if svcErr != nil {
		if svcErr.Type == serviceerror.ClientErrorType {
			execResp.Status = common.ExecFailure
			execResp.FailureReason = svcErr.ErrorDescription
			return nil, nil
		}

		logger.Error("Failed to exchange code for a token", log.String("errorCode", svcErr.Code),
			log.String("errorDescription", svcErr.ErrorDescription))
		return nil, errors.New("failed to exchange code for token")
	}

	return &OAuthTokenResponse{
		AccessToken:  tokenResp.AccessToken,
		TokenType:    tokenResp.TokenType,
		Scope:        tokenResp.Scope,
		RefreshToken: tokenResp.RefreshToken,
		IDToken:      tokenResp.IDToken,
		ExpiresIn:    tokenResp.ExpiresIn,
	}, nil
}

// GetUserInfo fetches user information from the OAuth provider using the access token.
func (o *oAuthExecutor) GetUserInfo(ctx *core.NodeContext, execResp *common.ExecutorResponse,
	accessToken string) (map[string]string, error) {
	logger := o.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Fetching user info from OAuth provider")

	idpID, err := o.GetIdpID(ctx)
	if err != nil {
		return nil, err
	}

	userInfo, svcErr := o.authService.FetchUserInfo(idpID, accessToken)
	if svcErr != nil {
		if svcErr.Type == serviceerror.ClientErrorType {
			execResp.Status = common.ExecFailure
			execResp.FailureReason = svcErr.ErrorDescription
			return nil, nil
		}

		logger.Error("Failed to fetch user info", log.String("errorCode", svcErr.Code),
			log.String("errorDescription", svcErr.ErrorDescription))
		return nil, errors.New("failed to fetch user information")
	}

	return systemutils.ConvertInterfaceMapToStringMap(userInfo), nil
}

// GetIdpID retrieves the identity provider ID from the node properties.
func (o *oAuthExecutor) GetIdpID(ctx *core.NodeContext) (string, error) {
	if len(ctx.NodeProperties) > 0 {
		if val, ok := ctx.NodeProperties["idpId"]; ok {
			if idpID, valid := val.(string); valid && idpID != "" {
				return idpID, nil
			}
		}
	}
	return "", errors.New("idpId is not configured in node properties")
}

// getIDPName retrieves the name of the identity provider using its ID.
func (o *oAuthExecutor) getIDPName(idpID string) (string, error) {
	logger := o.logger
	logger.Debug("Retrieving IDP name for the given IDP ID")

	idp, svcErr := o.idpService.GetIdentityProvider(idpID)
	if svcErr != nil {
		if svcErr.Type == serviceerror.ClientErrorType {
			return "", fmt.Errorf("failed to get identity provider: %s", svcErr.ErrorDescription)
		}

		logger.Error("Error while retrieving identity provider", log.String("errorCode", svcErr.Code),
			log.String("errorDescription", svcErr.ErrorDescription))
		return "", errors.New("error while retrieving identity provider")
	}

	return idp.Name, nil
}

func (o *oAuthExecutor) GetInternalUser(sub string, execResp *common.ExecutorResponse) (*userprovider.User, error) {
	logger := o.logger
	logger.Debug("Resolving internal user with the given sub claim")

	user, svcErr := o.authService.GetInternalUser(sub)
	if svcErr != nil {
		if svcErr.Code == authncm.ErrorUserNotFound.Code {
			return nil, nil
		}
		if svcErr.Type == serviceerror.ClientErrorType {
			execResp.Status = common.ExecFailure
			execResp.FailureReason = svcErr.ErrorDescription
			return nil, nil
		}
		logger.Error("Error while retrieving internal user", log.String("errorCode", svcErr.Code),
			log.String("description", svcErr.ErrorDescription))
		return nil, errors.New("error while retrieving internal user")
	}
	if user == nil || user.UserID == "" {
		return nil, nil
	}

	return user, nil
}

// ResolveContextUser resolves the authenticated user in context with the attributes.
func (o *oAuthExecutor) ResolveContextUser(ctx *core.NodeContext,
	execResp *common.ExecutorResponse, sub string, internalUser *userprovider.User) (
	*authncm.AuthenticatedUser, error) {
	if ctx.FlowType == common.FlowTypeAuthentication {
		return o.getContextUserForAuthentication(ctx, execResp, sub, internalUser)
	}
	return o.getContextUserForRegistration(ctx, execResp, sub, internalUser)
}

// getContextUserForAuthentication resolves the authenticated user in context for authentication flows.
func (o *oAuthExecutor) getContextUserForAuthentication(ctx *core.NodeContext,
	execResp *common.ExecutorResponse, sub string, internalUser *userprovider.User) (
	*authncm.AuthenticatedUser, error) {
	logger := o.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))

	// If no local user is found, check if authentication without local user is allowed
	if internalUser == nil {
		allowAuthWithoutLocalUser := false
		if val, ok := ctx.NodeProperties[common.NodePropertyAllowAuthenticationWithoutLocalUser]; ok {
			if boolVal, ok := val.(bool); ok {
				allowAuthWithoutLocalUser = boolVal
			}
		}

		if allowAuthWithoutLocalUser {
			logger.Debug("User not found, but authentication is allowed without a local user")

			err := o.resolveUserTypeForAutoProvisioning(ctx, execResp)
			if err != nil {
				return nil, err
			}
			if execResp.Status == common.ExecFailure {
				return nil, nil
			}

			execResp.Status = common.ExecComplete
			execResp.FailureReason = ""
			execResp.RuntimeData[common.RuntimeKeyUserEligibleForProvisioning] = dataValueTrue
			execResp.RuntimeData[userAttributeSub] = sub

			return &authncm.AuthenticatedUser{
				IsAuthenticated: false,
			}, nil
		}

		execResp.Status = common.ExecFailure
		execResp.FailureReason = "User not found"
		return nil, nil
	}

	// User found, proceed with authentication
	execResp.Status = common.ExecComplete
	authenticatedUser := authncm.AuthenticatedUser{
		IsAuthenticated:    true,
		UserID:             internalUser.UserID,
		OrganizationUnitID: internalUser.OrganizationUnitID,
		UserType:           internalUser.UserType,
	}

	return &authenticatedUser, nil
}

// getContextUserForRegistration resolves the authenticated user in context for registration flows.
func (o *oAuthExecutor) getContextUserForRegistration(ctx *core.NodeContext,
	execResp *common.ExecutorResponse, sub string, internalUser *userprovider.User) (
	*authncm.AuthenticatedUser, error) {
	logger := o.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))

	// If no local user is found, proceed with registration
	if internalUser == nil {
		logger.Debug("User not found for the provided sub claim. Proceeding with registration flow.")
		execResp.Status = common.ExecComplete
		execResp.FailureReason = ""
		execResp.RuntimeData[userAttributeSub] = sub

		return &authncm.AuthenticatedUser{
			IsAuthenticated: false,
		}, nil
	}

	// If a local user is found, check if registration with existing user is allowed
	allowRegistrationWithExistingUser := false
	if val, ok := ctx.NodeProperties[common.NodePropertyAllowRegistrationWithExistingUser]; ok {
		if boolVal, ok := val.(bool); ok {
			allowRegistrationWithExistingUser = boolVal
		}
	}

	if allowRegistrationWithExistingUser {
		logger.Debug("User already exists, but registration flow is allowed to continue")
		execResp.Status = common.ExecComplete
		execResp.FailureReason = ""
		execResp.RuntimeData[common.RuntimeKeySkipProvisioning] = dataValueTrue

		return &authncm.AuthenticatedUser{
			IsAuthenticated:    true,
			UserID:             internalUser.UserID,
			OrganizationUnitID: internalUser.OrganizationUnitID,
			UserType:           internalUser.UserType,
		}, nil
	}

	// Fail the execution as a unique user is found in the system.
	execResp.Status = common.ExecFailure
	execResp.FailureReason = "User already exists with the provided sub claim."
	return nil, nil
}

// resolveUserTypeForAutoProvisioning resolves the user type for auto provisioning in authentication flows.
func (o *oAuthExecutor) resolveUserTypeForAutoProvisioning(ctx *core.NodeContext,
	execResp *common.ExecutorResponse) error {
	logger := o.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Resolving user type for automatic provisioning")

	if len(ctx.Application.AllowedUserTypes) == 0 {
		logger.Debug("No allowed user types configured for the application")
		execResp.Status = common.ExecFailure
		execResp.FailureReason = errCannotProvisionUserAutomatically
		return nil
	}

	// Filter allowed user types to only those with self-registration enabled
	selfRegEnabledSchemas := make([]userschema.UserSchema, 0)
	for _, userType := range ctx.Application.AllowedUserTypes {
		userSchema, svcErr := o.userSchemaService.GetUserSchemaByName(userType)
		if svcErr != nil {
			if svcErr.Type == serviceerror.ClientErrorType {
				execResp.Status = common.ExecFailure
				execResp.FailureReason = svcErr.ErrorDescription
				return nil
			}

			logger.Error("Error while retrieving user schema", log.String("errorCode", svcErr.Code),
				log.String("description", svcErr.ErrorDescription))
			return errors.New("error while retrieving user schema")
		}
		if userSchema.AllowSelfRegistration {
			selfRegEnabledSchemas = append(selfRegEnabledSchemas, *userSchema)
		}
	}

	// Fail if no user types have self-registration enabled
	if len(selfRegEnabledSchemas) == 0 {
		logger.Debug("No user types with self-registration enabled, cannot provision automatically")
		execResp.Status = common.ExecFailure
		execResp.FailureReason = errCannotProvisionUserAutomatically
		return nil
	}

	// Fail if multiple user types have self-registration enabled
	if len(selfRegEnabledSchemas) > 1 {
		logger.Debug("Multiple user types with self-registration enabled, cannot resolve user type automatically")
		execResp.Status = common.ExecFailure
		execResp.FailureReason = errCannotProvisionUserAutomatically
		return nil
	}

	// Proceed with the single resolved user type
	// Add userType and ouID to runtime data
	execResp.RuntimeData[userTypeKey] = selfRegEnabledSchemas[0].Name
	execResp.RuntimeData[defaultOUIDKey] = selfRegEnabledSchemas[0].OrganizationUnitID
	return nil
}

// getContextUserAttributes extracts and returns user attributes from the user info map.
// TODO: Need to convert attributes as per the IDP to local attribute mapping when the support is implemented.
func (o *oAuthExecutor) getContextUserAttributes(execResp *common.ExecutorResponse,
	userInfo map[string]string) map[string]interface{} {
	attributes := make(map[string]interface{})
	for key, value := range userInfo {
		if !slices.Contains(userInfoSkipAttributes, key) {
			attributes[key] = value
		}
	}

	// Append email to runtime data if available.
	if email, ok := attributes[userAttributeEmail]; ok {
		if emailStr, ok := email.(string); ok && emailStr != "" {
			if execResp.RuntimeData == nil {
				execResp.RuntimeData = make(map[string]string)
			}
			execResp.RuntimeData[userAttributeEmail] = emailStr
		}
	}

	return attributes
}
