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

// Package googleauth provides the Google OIDC authentication executor.
package googleauth

import (
	"errors"
	"fmt"
	"slices"
	"time"

	authncm "github.com/asgardeo/thunder/internal/authn/common"
	authngoogle "github.com/asgardeo/thunder/internal/authn/google"
	authnoauth "github.com/asgardeo/thunder/internal/authn/oauth"
	authnoidc "github.com/asgardeo/thunder/internal/authn/oidc"
	"github.com/asgardeo/thunder/internal/executor/oauth/model"
	"github.com/asgardeo/thunder/internal/executor/oidcauth"
	flowconst "github.com/asgardeo/thunder/internal/flow/common/constants"
	flowmodel "github.com/asgardeo/thunder/internal/flow/common/model"
	"github.com/asgardeo/thunder/internal/idp"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	httpservice "github.com/asgardeo/thunder/internal/system/http"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/user"
)

const loggerComponentName = "GoogleOIDCAuthExecutor"

// GoogleOIDCAuthExecutor implements the OIDC authentication executor for Google.
type GoogleOIDCAuthExecutor struct {
	*oidcauth.OIDCAuthExecutor
	googleAuthService authngoogle.GoogleOIDCAuthnServiceInterface
}

var _ flowmodel.ExecutorInterface = (*GoogleOIDCAuthExecutor)(nil)

// NewGoogleOIDCAuthExecutorFromProps creates a new instance of GoogleOIDCAuthExecutor with the provided properties.
func NewGoogleOIDCAuthExecutorFromProps(execProps flowmodel.ExecutorProperties,
	oAuthProps *model.BasicOAuthExecProperties) oidcauth.OIDCAuthExecutorInterface {
	// Prepare the complete OAuth properties for Google
	compOAuthProps := &model.OAuthExecProperties{
		AuthorizationEndpoint: authngoogle.AuthorizeEndpoint,
		TokenEndpoint:         authngoogle.TokenEndpoint,
		UserInfoEndpoint:      authngoogle.UserInfoEndpoint,
		JwksEndpoint:          authngoogle.JwksEndpoint,
		ClientID:              oAuthProps.ClientID,
		ClientSecret:          oAuthProps.ClientSecret,
		RedirectURI:           oAuthProps.RedirectURI,
		Scopes:                oAuthProps.Scopes,
		AdditionalParams:      oAuthProps.AdditionalParams,
	}
	endpoints := authnoauth.OAuthEndpoints{
		AuthorizationEndpoint: compOAuthProps.AuthorizationEndpoint,
		TokenEndpoint:         compOAuthProps.TokenEndpoint,
		UserInfoEndpoint:      compOAuthProps.UserInfoEndpoint,
	}

	defaultInputs := []flowmodel.InputData{
		{
			Name:     "code",
			Type:     "string",
			Required: true,
		},
		{
			Name:     "nonce",
			Type:     "string",
			Required: false,
		},
	}

	oAuthSvc := authnoauth.NewOAuthAuthnService(
		httpservice.NewHTTPClientWithTimeout(flowconst.DefaultHTTPTimeout),
		idp.NewIDPService(),
		endpoints,
	)
	oidcAuthSvc := authnoidc.NewOIDCAuthnService(oAuthSvc, nil)
	authSvc := authngoogle.NewGoogleOIDCAuthnService(oidcAuthSvc)

	base := oidcauth.NewOIDCAuthExecutor("google_oidc_auth_executor", execProps.Name,
		defaultInputs, execProps.Properties, compOAuthProps)
	exec, ok := base.(*oidcauth.OIDCAuthExecutor)
	if !ok {
		panic("failed to cast GoogleOIDCAuthExecutor to OIDCAuthExecutor")
	}
	return &GoogleOIDCAuthExecutor{
		OIDCAuthExecutor:  exec,
		googleAuthService: authSvc,
	}
}

// NewGoogleOIDCAuthExecutor creates a new instance of GoogleOIDCAuthExecutor with the provided details.
func NewGoogleOIDCAuthExecutor(id, name string, properties map[string]string,
	clientID, clientSecret, redirectURI string, scopes []string,
	additionalParams map[string]string) oidcauth.OIDCAuthExecutorInterface {
	// Prepare the OAuth properties for Google
	oAuthProps := &model.OAuthExecProperties{
		AuthorizationEndpoint: authngoogle.AuthorizeEndpoint,
		TokenEndpoint:         authngoogle.TokenEndpoint,
		UserInfoEndpoint:      authngoogle.UserInfoEndpoint,
		JwksEndpoint:          authngoogle.JwksEndpoint,
		ClientID:              clientID,
		ClientSecret:          clientSecret,
		RedirectURI:           redirectURI,
		Scopes:                scopes,
		AdditionalParams:      additionalParams,
	}
	endpoints := authnoauth.OAuthEndpoints{
		AuthorizationEndpoint: oAuthProps.AuthorizationEndpoint,
		TokenEndpoint:         oAuthProps.TokenEndpoint,
		UserInfoEndpoint:      oAuthProps.UserInfoEndpoint,
	}

	defaultInputs := []flowmodel.InputData{
		{
			Name:     "code",
			Type:     "string",
			Required: true,
		},
		{
			Name:     "nonce",
			Type:     "string",
			Required: false,
		},
	}

	oAuthSvc := authnoauth.NewOAuthAuthnService(
		httpservice.NewHTTPClientWithTimeout(flowconst.DefaultHTTPTimeout),
		idp.NewIDPService(),
		endpoints,
	)
	oidcAuthSvc := authnoidc.NewOIDCAuthnService(oAuthSvc, nil)
	authSvc := authngoogle.NewGoogleOIDCAuthnService(oidcAuthSvc)

	base := oidcauth.NewOIDCAuthExecutor(id, name, defaultInputs, properties, oAuthProps)
	exec, ok := base.(*oidcauth.OIDCAuthExecutor)
	if !ok {
		panic("failed to cast GoogleOIDCAuthExecutor to OIDCAuthExecutor")
	}
	return &GoogleOIDCAuthExecutor{
		OIDCAuthExecutor:  exec,
		googleAuthService: authSvc,
	}
}

// Execute executes the Google OIDC authentication flow.
func (g *GoogleOIDCAuthExecutor) Execute(ctx *flowmodel.NodeContext) (*flowmodel.ExecutorResponse, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Executing Google OIDC auth executor",
		log.String("executorID", g.GetID()), log.String("flowID", ctx.FlowID))

	execResp := &flowmodel.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	if g.CheckInputData(ctx, execResp) {
		logger.Debug("Required input data for Google OIDC auth executor is not provided")
		err := g.BuildAuthorizeFlow(ctx, execResp)
		if err != nil {
			return nil, err
		}

		logger.Debug("Google OIDC auth executor execution completed",
			log.String("status", string(execResp.Status)))
	} else {
		err := g.ProcessAuthFlowResponse(ctx, execResp)
		if err != nil {
			return nil, err
		}

		logger.Debug("Google OIDC auth executor execution completed",
			log.String("status", string(execResp.Status)),
			log.Bool("isAuthenticated", execResp.AuthenticatedUser.IsAuthenticated))
	}

	return execResp, nil
}

// ProcessAuthFlowResponse processes the response from the Google authentication flow and authenticates the user.
// This method has been overridden to handle Google-specific logic in id token validation.
func (g *GoogleOIDCAuthExecutor) ProcessAuthFlowResponse(ctx *flowmodel.NodeContext,
	execResp *flowmodel.ExecutorResponse) error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName),
		log.String(log.LoggerKeyExecutorID, g.GetID()),
		log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Processing Google OIDC authentication response")

	code, ok := ctx.UserInputData["code"]
	if ok && code != "" {
		tokenResp, err := g.ExchangeCodeForToken(ctx, execResp, code)
		if err != nil {
			logger.Error("Failed to exchange code for a token", log.Error(err))
			return fmt.Errorf("failed to exchange code for token: %w", err)
		}
		if execResp.Status == flowconst.ExecFailure {
			return nil
		}

		idTokenClaims, err := g.GetIDTokenClaims(execResp, tokenResp.IDToken)
		if err != nil {
			return errors.New("failed to extract ID token claims: " + err.Error())
		}
		if execResp.Status == flowconst.ExecFailure {
			return nil
		}

		// Validate nonce if configured.
		if nonce, ok := ctx.UserInputData["nonce"]; ok && nonce != "" {
			if idTokenClaims["nonce"] != nonce {
				execResp.Status = flowconst.ExecFailure
				execResp.FailureReason = "Nonce mismatch in ID token claims."
				return nil
			}
		}

		// Resolve user with the sub claim.
		// TODO: For now assume `sub` is the unique identifier for the user always.
		parsedSub := ""
		sub, ok := idTokenClaims["sub"]
		if ok && sub != "" {
			if subStr, ok := sub.(string); ok && subStr != "" {
				parsedSub = subStr
			}
		}
		if parsedSub == "" {
			execResp.Status = flowconst.ExecFailure
			execResp.FailureReason = "sub claim not found in the ID token."
			return nil
		}

		user, err := g.resolveUser(parsedSub, ctx, execResp)
		if err != nil {
			return err
		}
		if execResp.Status == flowconst.ExecFailure {
			return nil
		}

		authenticatedUser, err := g.getAuthenticatedUserWithAttributes(ctx, execResp,
			tokenResp.AccessToken, idTokenClaims, user)
		if err != nil {
			return err
		}
		if execResp.Status == flowconst.ExecFailure || authenticatedUser == nil {
			return nil
		}
		execResp.AuthenticatedUser = *authenticatedUser
	} else {
		execResp.AuthenticatedUser = authncm.AuthenticatedUser{
			IsAuthenticated: false,
		}
	}

	if execResp.AuthenticatedUser.IsAuthenticated {
		execResp.Status = flowconst.ExecComplete
	} else if ctx.FlowType != flowconst.FlowTypeRegistration {
		execResp.Status = flowconst.ExecFailure
		execResp.FailureReason = "Authentication failed. Authorization code not provided or invalid."
		return nil
	}

	// Add execution record for successful Google authentication
	execResp.ExecutionRecord = &flowmodel.NodeExecutionRecord{
		ExecutorName: authncm.AuthenticatorGoogle,
		ExecutorType: flowconst.ExecutorTypeAuthentication,
		Timestamp:    time.Now().Unix(),
		Status:       flowconst.FlowStatusComplete,
	}

	return nil
}

// ExchangeCodeForToken exchanges the authorization code for an access token.
// This method has been overridden to handle Google-specific logic in id token validation.
func (g *GoogleOIDCAuthExecutor) ExchangeCodeForToken(ctx *flowmodel.NodeContext,
	execResp *flowmodel.ExecutorResponse, code string) (*model.TokenResponse, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName),
		log.String(log.LoggerKeyExecutorID, g.GetID()),
		log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Exchanging authorization code for a token", log.String("tokenEndpoint", g.GetTokenEndpoint()))

	tokenResp, svcErr := g.googleAuthService.ExchangeCodeForToken(g.GetID(), code, true)
	if svcErr != nil {
		if svcErr.Type == serviceerror.ClientErrorType {
			execResp.Status = flowconst.ExecFailure
			execResp.FailureReason = svcErr.ErrorDescription
			return nil, nil
		}

		logger.Error("Failed to exchange code for a token", log.String("errorCode", svcErr.Code),
			log.String("errorDescription", svcErr.ErrorDescription))
		return nil, errors.New("failed to exchange code for token")
	}

	return &model.TokenResponse{
		AccessToken:  tokenResp.AccessToken,
		TokenType:    tokenResp.TokenType,
		Scope:        tokenResp.Scope,
		RefreshToken: tokenResp.RefreshToken,
		IDToken:      tokenResp.IDToken,
		ExpiresIn:    tokenResp.ExpiresIn,
	}, nil
}

// ValidateIDToken validates the ID token received from Google.
// This method has been overridden to handle Google-specific logic in id token validation.
func (g *GoogleOIDCAuthExecutor) ValidateIDToken(execResp *flowmodel.ExecutorResponse, idToken string) error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Validating ID token")

	svcErr := g.googleAuthService.ValidateIDToken(g.GetID(), idToken)
	if svcErr != nil {
		if svcErr.Type == serviceerror.ClientErrorType {
			execResp.Status = flowconst.ExecFailure
			execResp.FailureReason = svcErr.ErrorDescription
			return nil
		}

		logger.Error("Failed to validate ID token", log.String("errorCode", svcErr.Code),
			log.String("errorDescription", svcErr.ErrorDescription))
		return errors.New("failed to validate ID token")
	}

	return nil
}

// getAuthenticatedUserWithAttributes constructs the authenticated user object with attributes from the
// ID token and user info.
func (g *GoogleOIDCAuthExecutor) getAuthenticatedUserWithAttributes(ctx *flowmodel.NodeContext,
	execResp *flowmodel.ExecutorResponse, accessToken string, idTokenClaims map[string]interface{},
	user *user.User) (*authncm.AuthenticatedUser, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName),
		log.String(log.LoggerKeyExecutorID, g.GetID()),
		log.String(log.LoggerKeyFlowID, ctx.FlowID))

	userClaims := make(map[string]interface{})
	if len(idTokenClaims) != 0 {
		// Filter non-user claims from the ID token claims.
		for attr, val := range idTokenClaims {
			if !slices.Contains(oidcauth.IDTokenNonUserAttributes, attr) {
				userClaims[attr] = val
			}
		}
		logger.Debug("Extracted ID token claims", log.Int("noOfClaims", len(idTokenClaims)))
	}

	if len(g.GetOAuthProperties().Scopes) == 1 && slices.Contains(g.GetOAuthProperties().Scopes, "openid") {
		logger.Debug("No additional scopes configured.")
	} else {
		// Get user info using the access token
		userInfo, err := g.GetUserInfo(ctx, execResp, accessToken)
		if err != nil {
			return nil, fmt.Errorf("failed to get user info: %w", err)
		}
		if execResp.Status == flowconst.ExecFailure {
			return nil, nil
		}

		for key, value := range userInfo {
			if key != "username" && key != "sub" && key != "id" {
				userClaims[key] = value
			}
		}
	}

	authenticatedUser := authncm.AuthenticatedUser{}
	if ctx.FlowType == flowconst.FlowTypeRegistration {
		authenticatedUser.IsAuthenticated = false
	} else {
		authenticatedUser.IsAuthenticated = true
		authenticatedUser.UserID = user.ID
		authenticatedUser.OrganizationUnitID = user.OrganizationUnit
		authenticatedUser.UserType = user.Type
	}

	// TODO: Need to convert attributes as per the IDP to local attribute mapping
	//  when the support is implemented.
	authenticatedUser.Attributes = userClaims

	return &authenticatedUser, nil
}

// resolveUser resolves the internal user based on the sub claim.
func (g *GoogleOIDCAuthExecutor) resolveUser(sub string, ctx *flowmodel.NodeContext,
	execResp *flowmodel.ExecutorResponse) (*user.User, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName),
		log.String(log.LoggerKeyExecutorID, g.GetID()),
		log.String(log.LoggerKeyFlowID, ctx.FlowID))

	user, svcErr := g.googleAuthService.GetInternalUser(sub)
	if svcErr != nil {
		if svcErr.Code == authncm.ErrorUserNotFound.Code {
			if ctx.FlowType == flowconst.FlowTypeRegistration {
				logger.Debug("User not found for the provided sub claim. Proceeding with registration flow.")
				execResp.Status = flowconst.ExecComplete
				execResp.FailureReason = ""

				if execResp.RuntimeData == nil {
					execResp.RuntimeData = make(map[string]string)
				}
				execResp.RuntimeData["sub"] = sub

				return nil, nil
			} else {
				execResp.Status = flowconst.ExecFailure
				execResp.FailureReason = "User not found"
				return nil, nil
			}
		} else {
			if svcErr.Type == serviceerror.ClientErrorType {
				execResp.Status = flowconst.ExecFailure
				execResp.FailureReason = svcErr.ErrorDescription
				return nil, nil
			}

			logger.Error("Error while retrieving internal user", log.String("errorCode", svcErr.Code),
				log.String("description", svcErr.ErrorDescription))
			return nil, errors.New("error while retrieving internal user")
		}
	}

	if ctx.FlowType == flowconst.FlowTypeRegistration {
		// At this point, a unique user is found in the system. Hence fail the execution.
		execResp.Status = flowconst.ExecFailure
		execResp.FailureReason = "User already exists with the provided sub claim."
		return nil, nil
	}

	if user == nil || user.ID == "" {
		return nil, errors.New("retrieved user is nil or has an empty ID")
	}

	return user, nil
}
