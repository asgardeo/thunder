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
	authnoidc "github.com/asgardeo/thunder/internal/authn/oidc"
	"github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/internal/flow/core"
	"github.com/asgardeo/thunder/internal/idp"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/usertype"
)

const (
	oidcAuthLoggerComponentName = "OIDCAuthExecutor"
)

// idTokenNonUserAttributes contains the list of non-user attributes that are expected in the ID token.
var idTokenNonUserAttributes = []string{"aud", "exp", "iat", "iss", "at_hash", "azp", "nonce", "sub"}

// oidcAuthExecutorInterface defines the interface for OIDC authentication executors.
type oidcAuthExecutorInterface interface {
	oAuthExecutorInterface
	GetIDTokenClaims(execResp *common.ExecutorResponse, idToken string) (map[string]interface{}, error)
}

// oidcAuthExecutor implements the OIDCAuthExecutorInterface for handling generic OIDC authentication flows.
type oidcAuthExecutor struct {
	oAuthExecutorInterface
	authService authnoidc.OIDCAuthnCoreServiceInterface
	logger      *log.Logger
}

var _ core.ExecutorInterface = (*oidcAuthExecutor)(nil)

// newOIDCAuthExecutor creates a new instance of OIDCAuthExecutor.
func newOIDCAuthExecutor(
	name string,
	defaultInputs, prerequisites []common.Input,
	flowFactory core.FlowFactoryInterface,
	idpService idp.IDPServiceInterface,
	userTypeService usertype.UserTypeServiceInterface,
	authService authnoidc.OIDCAuthnCoreServiceInterface,
) oidcAuthExecutorInterface {
	if name == "" {
		name = ExecutorNameOIDCAuth
	}
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, oidcAuthLoggerComponentName),
		log.String(log.LoggerKeyExecutorName, name))

	oauthSvcCast, ok := authService.(authnoauth.OAuthAuthnCoreServiceInterface)
	if !ok {
		panic("failed to cast OIDCAuthnService to OAuthAuthnCoreServiceInterface")
	}

	base := newOAuthExecutor(name, defaultInputs, prerequisites,
		flowFactory, idpService, userTypeService, oauthSvcCast)

	return &oidcAuthExecutor{
		oAuthExecutorInterface: base,
		authService:            authService,
		logger:                 logger,
	}
}

// Execute executes the OIDC authentication logic.
func (o *oidcAuthExecutor) Execute(ctx *core.NodeContext) (*common.ExecutorResponse, error) {
	logger := o.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Executing OIDC authentication executor")

	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	if !o.HasRequiredInputs(ctx, execResp) {
		logger.Debug("Required inputs for OIDC authentication executor is not provided")
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

	logger.Debug("OIDC authentication executor execution completed",
		log.String("status", string(execResp.Status)),
		log.Bool("isAuthenticated", execResp.AuthenticatedUser.IsAuthenticated))

	return execResp, nil
}

// ProcessAuthFlowResponse processes the response from the OIDC authentication flow and authenticates the user.
func (o *oidcAuthExecutor) ProcessAuthFlowResponse(ctx *core.NodeContext,
	execResp *common.ExecutorResponse) error {
	logger := o.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Processing OIDC authentication response")

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

	idTokenClaims, err := o.GetIDTokenClaims(execResp, tokenResp.IDToken)
	if err != nil {
		return err
	}
	if execResp.Status == common.ExecFailure {
		return nil
	}

	// Validate nonce if configured
	if nonce, ok := ctx.UserInputs[userInputNonce]; ok && nonce != "" {
		if idTokenClaims[userInputNonce] != nonce {
			execResp.Status = common.ExecFailure
			execResp.FailureReason = "Nonce mismatch in ID token claims."
			return nil
		}
	}

	// Extract sub claim from the id token claims
	parsedSub := ""
	sub, ok := idTokenClaims[userAttributeSub]
	if ok && sub != "" {
		if subStr, ok := sub.(string); ok && subStr != "" {
			parsedSub = subStr
		}
	}
	if parsedSub == "" {
		execResp.Status = common.ExecFailure
		execResp.FailureReason = "sub claim not found in the ID token."
		return nil
	}

	internalUser, err := o.GetInternalUser(parsedSub, execResp)
	if err != nil {
		return err
	}
	if execResp.Status == common.ExecFailure {
		return nil
	}

	contextUser, err := o.ResolveContextUser(ctx, execResp, parsedSub, internalUser)
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

	attributes, err := o.getContextUserAttributes(ctx, execResp, idTokenClaims, tokenResp.AccessToken)
	if err != nil {
		return err
	}
	if execResp.Status == common.ExecFailure {
		return nil
	}

	contextUser.Attributes = attributes
	execResp.AuthenticatedUser = *contextUser

	return nil
}

// GetIDTokenClaims extracts the ID token claims from the provided ID token.
func (o *oidcAuthExecutor) GetIDTokenClaims(execResp *common.ExecutorResponse,
	idToken string) (map[string]interface{}, error) {
	logger := o.logger
	logger.Debug("Extracting claims from the ID token")

	claims, svcErr := o.authService.GetIDTokenClaims(idToken)
	if svcErr != nil {
		if svcErr.Type == serviceerror.ClientErrorType {
			execResp.Status = common.ExecFailure
			execResp.FailureReason = svcErr.ErrorDescription
			return nil, nil
		}

		logger.Error("Failed to extract claims from the ID token", log.String("errorCode", svcErr.Code),
			log.String("errorDescription", svcErr.ErrorDescription))
		return nil, errors.New("failed to extract claims from the ID token")
	}

	return claims, nil
}

// getContextUserAttributes retrieves user attributes from the ID token claims and user info endpoint.
// TODO: Need to convert attributes as per the IDP to local attribute mapping when the support is implemented.
func (o *oidcAuthExecutor) getContextUserAttributes(ctx *core.NodeContext, execResp *common.ExecutorResponse,
	idTokenClaims map[string]interface{}, accessToken string) (map[string]interface{}, error) {
	logger := o.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))
	userClaims := make(map[string]interface{})

	// Resolve and add ID token claims
	if len(idTokenClaims) != 0 {
		for attr, val := range idTokenClaims {
			if !slices.Contains(idTokenNonUserAttributes, attr) {
				userClaims[attr] = val
			}
		}
		logger.Debug("Extracted ID token claims", log.Int("noOfClaims", len(idTokenClaims)))
	}

	// Retrieve IDP and check for additional scopes
	idpID, err := o.GetIdpID(ctx)
	if err != nil {
		return nil, err
	}

	oauthConfigs, svcErr := o.authService.GetOAuthClientConfig(idpID)
	if svcErr != nil {
		if svcErr.Type == serviceerror.ClientErrorType {
			execResp.Status = common.ExecFailure
			execResp.FailureReason = fmt.Sprintf("failed to retrieve OAuth client configuration: %s",
				svcErr.ErrorDescription)
			return nil, nil
		}

		logger.Error("Failed to retrieve OAuth client configuration", log.String("errorCode", svcErr.Code),
			log.String("errorDescription", svcErr.ErrorDescription))
		return nil, errors.New("failed to retrieve OAuth client configuration")
	}

	// If additional scopes are configured, retrieve user info
	if len(oauthConfigs.Scopes) > 1 {
		userInfo, err := o.GetUserInfo(ctx, execResp, accessToken)
		if err != nil {
			return nil, err
		}
		if execResp.Status == common.ExecFailure {
			return nil, nil
		}

		for key, value := range userInfo {
			if !slices.Contains(userInfoSkipAttributes, key) {
				userClaims[key] = value
			}
		}
	} else {
		logger.Debug("No additional scopes configured.")
	}

	// Append email to runtime data if available.
	if email, ok := userClaims[userAttributeEmail]; ok {
		if emailStr, ok := email.(string); ok && emailStr != "" {
			if execResp.RuntimeData == nil {
				execResp.RuntimeData = make(map[string]string)
			}
			execResp.RuntimeData[userAttributeEmail] = emailStr
		}
	}

	return userClaims, nil
}
