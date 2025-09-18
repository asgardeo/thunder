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

// Package oauth provides the OAuth authentication executor for handling OAuth-based authentication flows.
package oauth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	authndto "github.com/asgardeo/thunder/internal/authn/dto"
	"github.com/asgardeo/thunder/internal/executor/identify"
	"github.com/asgardeo/thunder/internal/executor/oauth/model"
	"github.com/asgardeo/thunder/internal/executor/oauth/utils"
	flowconst "github.com/asgardeo/thunder/internal/flow/constants"
	flowmodel "github.com/asgardeo/thunder/internal/flow/model"
	oauth2const "github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/internal/system/constants"
	httpservice "github.com/asgardeo/thunder/internal/system/http"
	"github.com/asgardeo/thunder/internal/system/log"
	systemutils "github.com/asgardeo/thunder/internal/system/utils"
)

const loggerComponentName = "OAuthExecutor"

// OAuthExecutorInterface defines the interface for OAuth authentication executors.
type OAuthExecutorInterface interface {
	flowmodel.ExecutorInterface
	BuildAuthorizeFlow(ctx *flowmodel.NodeContext, execResp *flowmodel.ExecutorResponse) error
	ProcessAuthFlowResponse(ctx *flowmodel.NodeContext, execResp *flowmodel.ExecutorResponse) error
	GetOAuthProperties() model.OAuthExecProperties
	GetCallBackURL() string
	GetAuthorizationEndpoint() string
	GetTokenEndpoint() string
	GetUserInfoEndpoint() string
	GetLogoutEndpoint() string
	GetJWKSEndpoint() string
	ExchangeCodeForToken(ctx *flowmodel.NodeContext, execResp *flowmodel.ExecutorResponse,
		code string) (*model.TokenResponse, error)
	GetUserInfo(ctx *flowmodel.NodeContext, execResp *flowmodel.ExecutorResponse,
		accessToken string) (map[string]string, error)
}

// OAuthExecutor implements the OAuthExecutorInterface for handling generic OAuth authentication flows.
type OAuthExecutor struct {
	*identify.IdentifyingExecutor
	internal        flowmodel.Executor
	oAuthProperties model.OAuthExecProperties
}

var _ flowmodel.ExecutorInterface = (*OAuthExecutor)(nil)

// NewOAuthExecutor creates a new instance of OAuthExecutor.
func NewOAuthExecutor(id, name string, defaultInputs []flowmodel.InputData, properties map[string]string,
	oAuthProps *model.OAuthExecProperties) OAuthExecutorInterface {
	if len(defaultInputs) == 0 {
		defaultInputs = []flowmodel.InputData{
			{
				Name:     "code",
				Type:     "string",
				Required: true,
			},
		}
	}
	return &OAuthExecutor{
		IdentifyingExecutor: identify.NewIdentifyingExecutor(id, name, properties),
		internal:            *flowmodel.NewExecutor(id, name, defaultInputs, []flowmodel.InputData{}, properties),
		oAuthProperties:     *oAuthProps,
	}
}

// GetID returns the ID of the OAuthExecutor.
func (o *OAuthExecutor) GetID() string {
	return o.internal.GetID()
}

// GetName returns the name of the OAuthExecutor.
func (o *OAuthExecutor) GetName() string {
	return o.internal.GetName()
}

// GetProperties returns the properties of the OAuthExecutor.
func (o *OAuthExecutor) GetProperties() flowmodel.ExecutorProperties {
	return o.internal.Properties
}

// GetOAuthProperties returns the OAuth properties of the executor.
func (o *OAuthExecutor) GetOAuthProperties() model.OAuthExecProperties {
	return o.oAuthProperties
}

// GetCallBackURL returns the callback URL for the OAuth authentication.
func (o *OAuthExecutor) GetCallBackURL() string {
	return o.oAuthProperties.RedirectURI
}

// GetAuthorizationEndpoint returns the authorization endpoint of the OAuth authentication.
func (o *OAuthExecutor) GetAuthorizationEndpoint() string {
	return o.oAuthProperties.AuthorizationEndpoint
}

// GetTokenEndpoint returns the token endpoint of the OAuth authentication.
func (o *OAuthExecutor) GetTokenEndpoint() string {
	return o.oAuthProperties.TokenEndpoint
}

// GetUserInfoEndpoint returns the user info endpoint of the OAuth authentication.
func (o *OAuthExecutor) GetUserInfoEndpoint() string {
	return o.oAuthProperties.UserInfoEndpoint
}

// GetLogoutEndpoint returns the logout endpoint of the OAuth authentication.
func (o *OAuthExecutor) GetLogoutEndpoint() string {
	return o.oAuthProperties.LogoutEndpoint
}

// GetJWKSEndpoint returns the JWKs endpoint of the OAuth authentication.
func (o *OAuthExecutor) GetJWKSEndpoint() string {
	return o.oAuthProperties.JwksEndpoint
}

// Execute executes the OAuth authentication flow.
func (o *OAuthExecutor) Execute(ctx *flowmodel.NodeContext) (*flowmodel.ExecutorResponse, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName),
		log.String(log.LoggerKeyExecutorID, o.GetID()),
		log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Executing OAuth authentication executor")

	execResp := &flowmodel.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	// Check if the required input data is provided
	if o.CheckInputData(ctx, execResp) {
		// If required input data is not provided, return incomplete status with redirection.
		logger.Debug("Required input data for OAuth authentication executor is not provided")
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
func (o *OAuthExecutor) BuildAuthorizeFlow(ctx *flowmodel.NodeContext, execResp *flowmodel.ExecutorResponse) error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName),
		log.String(log.LoggerKeyExecutorID, o.GetID()),
		log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Initiating OAuth authentication flow")

	queryParams, err := o.getQueryParams(ctx)
	if err != nil {
		logger.Error("Failed to construct query parameters", log.Error(err))
		return fmt.Errorf("failed to construct query parameters: %w", err)
	}

	// Construct the authorization URL
	authURL, err := systemutils.GetURIWithQueryParams(o.GetAuthorizationEndpoint(), queryParams)
	if err != nil {
		logger.Error("Failed to prepare authorization URL", log.Error(err))
		return fmt.Errorf("failed to prepare authorization URL: %w", err)
	}

	// Set the response to redirect the user to the authorization URL.
	execResp.Status = flowconst.ExecExternalRedirection
	execResp.RedirectURL = authURL
	execResp.AdditionalData = map[string]string{
		flowconst.DataIDPName: o.GetName(),
	}

	return nil
}

// ProcessAuthFlowResponse processes the response from the OAuth authentication flow and authenticates the user.
func (o *OAuthExecutor) ProcessAuthFlowResponse(ctx *flowmodel.NodeContext,
	execResp *flowmodel.ExecutorResponse) error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName),
		log.String(log.LoggerKeyExecutorID, o.GetID()),
		log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Processing OAuth authentication response")

	code, ok := ctx.UserInputData["code"]
	if ok && code != "" {
		tokenResp, err := o.ExchangeCodeForToken(ctx, execResp, code)
		if err != nil {
			logger.Error("Failed to exchange code for a token", log.Error(err))
			return fmt.Errorf("failed to exchange code for token: %w", err)
		}
		if execResp.Status == flowconst.ExecFailure {
			return nil
		}

		err = o.validateTokenResponse(tokenResp)
		if err != nil {
			execResp.Status = flowconst.ExecFailure
			execResp.FailureReason = err.Error()
			return nil
		}

		if tokenResp.Scope == "" {
			logger.Error("Scopes are empty in the token response")
			execResp.AuthenticatedUser = authndto.AuthenticatedUser{
				IsAuthenticated: false,
			}
		} else {
			authenticatedUser, err := o.getAuthenticatedUserWithAttributes(ctx, execResp, tokenResp.AccessToken)
			if err != nil {
				return err
			}
			if authenticatedUser == nil {
				return nil
			}
			execResp.AuthenticatedUser = *authenticatedUser
		}
	} else {
		execResp.AuthenticatedUser = authndto.AuthenticatedUser{
			IsAuthenticated: false,
		}
	}

	if execResp.AuthenticatedUser.IsAuthenticated {
		execResp.Status = flowconst.ExecComplete
	} else if ctx.FlowType != flowconst.FlowTypeRegistration {
		execResp.Status = flowconst.ExecFailure
		execResp.FailureReason = "Authentication failed. Authorization code not provided or invalid."
	}

	return nil
}

// GetDefaultExecutorInputs returns the default required input data for the OAuthExecutor.
func (o *OAuthExecutor) GetDefaultExecutorInputs() []flowmodel.InputData {
	return o.internal.GetDefaultExecutorInputs()
}

// CheckInputData checks if the required input data is provided in the context.
func (o *OAuthExecutor) CheckInputData(ctx *flowmodel.NodeContext, execResp *flowmodel.ExecutorResponse) bool {
	if code, ok := ctx.UserInputData["code"]; ok && code != "" {
		return false
	}
	return o.internal.CheckInputData(ctx, execResp)
}

// GetPrerequisites returns the prerequisites for the OAuthExecutor.
func (o *OAuthExecutor) GetPrerequisites() []flowmodel.InputData {
	return o.internal.GetPrerequisites()
}

// ValidatePrerequisites validates whether the prerequisites for the OAuthExecutor are met.
func (o *OAuthExecutor) ValidatePrerequisites(ctx *flowmodel.NodeContext, execResp *flowmodel.ExecutorResponse) bool {
	return o.internal.ValidatePrerequisites(ctx, execResp)
}

// GetUserIDFromContext retrieves the user ID from the context.
func (o *OAuthExecutor) GetUserIDFromContext(ctx *flowmodel.NodeContext) (string, error) {
	return o.internal.GetUserIDFromContext(ctx)
}

// GetRequiredData returns the required input data for the OAuthExecutor.
func (o *OAuthExecutor) GetRequiredData(ctx *flowmodel.NodeContext) []flowmodel.InputData {
	return o.internal.GetRequiredData(ctx)
}

// ExchangeCodeForToken exchanges the authorization code for an access token.
func (o *OAuthExecutor) ExchangeCodeForToken(ctx *flowmodel.NodeContext, execResp *flowmodel.ExecutorResponse,
	code string) (*model.TokenResponse, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName),
		log.String(log.LoggerKeyExecutorID, o.GetID()),
		log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Exchanging authorization code for a token", log.String("tokenEndpoint", o.GetTokenEndpoint()))

	// Prepare the token request
	data := url.Values{}
	data.Set(oauth2const.RequestParamClientID, o.oAuthProperties.ClientID)
	data.Set(oauth2const.RequestParamClientSecret, o.oAuthProperties.ClientSecret)
	data.Set(oauth2const.RequestParamRedirectURI, o.oAuthProperties.RedirectURI)
	data.Set(oauth2const.RequestParamCode, code)
	data.Set(oauth2const.RequestParamGrantType, string(oauth2const.GrantTypeAuthorizationCode))

	// Create HTTP request
	req, err := http.NewRequest("POST", o.GetTokenEndpoint(), strings.NewReader(data.Encode()))
	if err != nil {
		logger.Error("Failed to create token request", log.Error(err))
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Add(constants.ContentTypeHeaderName, "application/x-www-form-urlencoded")
	req.Header.Add(constants.AcceptHeaderName, "application/json")

	// Execute the request
	logger.Debug("Sending token request to OAuth provider")
	client := httpservice.NewHTTPClientWithTimeout(10 * time.Second)
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Failed to send token request", log.Error(err))
		execResp.Status = flowconst.ExecFailure
		execResp.FailureReason = "Failed to exchange authorization code for token: " + err.Error()
		return nil, nil
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			logger.Error("Failed to close response body", log.Error(closeErr))
		}
	}()
	logger.Debug("Token response received from OAuth provider", log.Int("statusCode", resp.StatusCode))

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		execResp.Status = flowconst.ExecFailure
		execResp.FailureReason = fmt.Sprintf("Token request failed with status %s: %s", resp.Status, string(body))
		return nil, nil
	}

	// Parse the response
	var tokenResp model.TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		logger.Error("Failed to parse token response", log.Error(err))
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	return &tokenResp, nil
}

// GetUserInfo fetches user information from the OAuth provider using the access token.
func (o *OAuthExecutor) GetUserInfo(ctx *flowmodel.NodeContext, execResp *flowmodel.ExecutorResponse,
	accessToken string) (map[string]string, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName),
		log.String(log.LoggerKeyExecutorID, o.GetID()),
		log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Fetching user info from OAuth provider", log.String("userInfoEndpoint", o.GetUserInfoEndpoint()))

	// Create HTTP request
	req, err := http.NewRequest("GET", o.GetUserInfoEndpoint(), nil)
	if err != nil {
		logger.Error("Failed to create userinfo request", log.Error(err))
		return nil, fmt.Errorf("failed to create userinfo request: %w", err)
	}
	req.Header.Set(constants.AuthorizationHeaderName, constants.TokenTypeBearer+" "+accessToken)
	req.Header.Set(constants.AcceptHeaderName, "application/json")

	// Execute the request
	logger.Debug("Sending userinfo request to OAuth provider")

	client := httpservice.NewHTTPClientWithTimeout(10 * time.Second)
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Failed to send userinfo request", log.Error(err))
		execResp.Status = flowconst.ExecFailure
		execResp.FailureReason = "Failed to fetch user information: " + err.Error()
		return nil, nil
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			logger.Error("Failed to close userinfo response body", log.Error(closeErr))
		}
	}()
	logger.Debug("Userinfo response received from OAuth provider", log.Int("statusCode", resp.StatusCode))

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		execResp.Status = flowconst.ExecFailure
		execResp.FailureReason = fmt.Sprintf("Userinfo request failed with status %s: %s", resp.Status, string(body))
		return nil, nil
	}

	// Parse the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Failed to read userinfo response body", log.Error(err))
		return nil, fmt.Errorf("failed to read userinfo response body: %w", err)
	}

	var userInfo map[string]interface{}
	if err := json.Unmarshal(body, &userInfo); err != nil {
		logger.Error("Failed to parse userinfo response", log.Error(err))
		return nil, fmt.Errorf("failed to parse userinfo response: %w", err)
	}

	return systemutils.ConvertInterfaceMapToStringMap(userInfo), nil
}

// getQueryParams constructs the query parameters for the OAuth authorization request.
func (o *OAuthExecutor) getQueryParams(ctx *flowmodel.NodeContext) (map[string]string, error) {
	var queryParams = make(map[string]string)
	queryParams[oauth2const.RequestParamClientID] = o.oAuthProperties.ClientID
	queryParams[oauth2const.RequestParamRedirectURI] = o.oAuthProperties.RedirectURI
	queryParams[oauth2const.RequestParamResponseType] = oauth2const.RequestParamCode
	queryParams[oauth2const.RequestParamScope] = systemutils.StringifyStringArray(o.oAuthProperties.Scopes, " ")

	// append any configured additional parameters.
	additionalParams := o.oAuthProperties.AdditionalParams
	if len(additionalParams) > 0 {
		for key, value := range additionalParams {
			if key != "" && value != "" {
				resolvedValue, err := utils.GetResolvedAdditionalParam(key, value, ctx)
				if err != nil {
					return nil, fmt.Errorf("failed to resolve additional parameter %s: %w", key, err)
				}
				queryParams[key] = resolvedValue
			}
		}
	}

	return queryParams, nil
}

// validateTokenResponse validates the token response received from the OAuth provider.
func (o *OAuthExecutor) validateTokenResponse(tokenResp *model.TokenResponse) error {
	if tokenResp == nil {
		return fmt.Errorf("token response is nil")
	}
	if tokenResp.AccessToken == "" {
		return fmt.Errorf("access token is empty in the token response")
	}
	return nil
}

// getAuthenticatedUserWithAttributes retrieves the authenticated user information with additional attributes
// from the OAuth provider using the access token.
func (o *OAuthExecutor) getAuthenticatedUserWithAttributes(ctx *flowmodel.NodeContext,
	execResp *flowmodel.ExecutorResponse, accessToken string) (*authndto.AuthenticatedUser, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName),
		log.String(log.LoggerKeyExecutorID, o.GetID()),
		log.String(log.LoggerKeyFlowID, ctx.FlowID))

	// Get user info using the access token
	userInfo, err := o.GetUserInfo(ctx, execResp, accessToken)
	if err != nil {
		logger.Error("Failed to get user info", log.Error(err))
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	if execResp.Status == flowconst.ExecFailure {
		return nil, nil
	}

	// Resolve user with the sub claim.
	// TODO: For now assume `sub` is the unique identifier for the user always.
	sub, ok := userInfo["sub"]
	if !ok || sub == "" {
		execResp.Status = flowconst.ExecFailure
		execResp.FailureReason = "sub claim not found in the response."
		return nil, nil
	}

	filters := map[string]interface{}{"sub": sub}
	userID, err := o.IdentifyUser(filters, execResp)
	if err != nil {
		return nil, fmt.Errorf("failed to identify user with sub claim: %w", err)
	}

	// Handle registration flows.
	if ctx.FlowType == flowconst.FlowTypeRegistration {
		if execResp.Status == flowconst.ExecFailure {
			if execResp.FailureReason == "User not found" {
				logger.Debug("User not found for the provided sub claim. Proceeding with registration flow.")
				execResp.Status = flowconst.ExecComplete
				execResp.FailureReason = ""

				if execResp.RuntimeData == nil {
					execResp.RuntimeData = make(map[string]string)
				}
				execResp.RuntimeData["sub"] = sub

				// TODO: Need to convert attributes as per the IDP to local attribute mapping
				//  when the support is implemented.
				return &authndto.AuthenticatedUser{
					IsAuthenticated: false,
					Attributes:      getUserAttributes(userInfo, ""),
				}, nil
			}
			return nil, err
		}

		// At this point, a unique user is found in the system. Hence fail the execution.
		execResp.Status = flowconst.ExecFailure
		execResp.FailureReason = "User already exists with the provided sub claim."
		return nil, nil
	}

	if execResp.Status == flowconst.ExecFailure {
		return nil, nil
	}

	// TODO: Need to convert attributes as per the IDP to local attribute mapping
	//  when the support is implemented.
	authenticatedUser := authndto.AuthenticatedUser{
		IsAuthenticated: true,
		UserID:          *userID,
		Attributes:      getUserAttributes(userInfo, *userID),
	}

	return &authenticatedUser, nil
}

// getUserAttributes extracts user attributes from the user info map, excluding certain keys.
func getUserAttributes(userInfo map[string]string, userID string) map[string]interface{} {
	attributes := make(map[string]interface{})
	for key, value := range userInfo {
		if key != "username" && key != "sub" {
			attributes[key] = value
		}
	}
	if userID != "" {
		attributes["user_id"] = userID
	}

	return attributes
}
