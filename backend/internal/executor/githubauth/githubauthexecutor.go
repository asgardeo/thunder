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

// Package githubauth provides the GitHub OAuth authentication executor.
package githubauth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"time"

	authndto "github.com/asgardeo/thunder/internal/authn/dto"
	"github.com/asgardeo/thunder/internal/executor/oauth"
	"github.com/asgardeo/thunder/internal/executor/oauth/model"
	flowconst "github.com/asgardeo/thunder/internal/flow/constants"
	flowmodel "github.com/asgardeo/thunder/internal/flow/model"
	"github.com/asgardeo/thunder/internal/system/constants"
	httpservice "github.com/asgardeo/thunder/internal/system/http"
	"github.com/asgardeo/thunder/internal/system/log"
	systemutils "github.com/asgardeo/thunder/internal/system/utils"
)

const loggerComponentName = "GithubOAuthExecutor"

// GithubOAuthExecutor implements the OAuth authentication executor for GitHub.
type GithubOAuthExecutor struct {
	*oauth.OAuthExecutor
}

var _ flowmodel.ExecutorInterface = (*GithubOAuthExecutor)(nil)

// NewGithubOAuthExecutorFromProps creates a new instance of GithubOAuthExecutor with the provided properties.
func NewGithubOAuthExecutorFromProps(execProps flowmodel.ExecutorProperties,
	oAuthProps *model.BasicOAuthExecProperties) oauth.OAuthExecutorInterface {
	// Prepare the complete OAuth properties for GitHub
	compOAuthProps := &model.OAuthExecProperties{
		AuthorizationEndpoint: githubAuthorizeEndpoint,
		TokenEndpoint:         githubTokenEndpoint,
		UserInfoEndpoint:      githubUserInfoEndpoint,
		ClientID:              oAuthProps.ClientID,
		ClientSecret:          oAuthProps.ClientSecret,
		RedirectURI:           oAuthProps.RedirectURI,
		Scopes:                oAuthProps.Scopes,
		AdditionalParams:      oAuthProps.AdditionalParams,
	}

	base := oauth.NewOAuthExecutor("github_oauth_executor", execProps.Name, []flowmodel.InputData{},
		execProps.Properties, compOAuthProps)
	exec, ok := base.(*oauth.OAuthExecutor)
	if !ok {
		panic("failed to cast GithubOAuthExecutor to OAuthExecutor")
	}
	return &GithubOAuthExecutor{
		OAuthExecutor: exec,
	}
}

// NewGithubOAuthExecutor creates a new instance of GithubOAuthExecutor with the provided details.
func NewGithubOAuthExecutor(id, name string, properties map[string]string,
	clientID, clientSecret, redirectURI string, scopes []string,
	additionalParams map[string]string) oauth.OAuthExecutorInterface {
	// Prepare the OAuth properties for GitHub
	oAuthProps := &model.OAuthExecProperties{
		AuthorizationEndpoint: githubAuthorizeEndpoint,
		TokenEndpoint:         githubTokenEndpoint,
		UserInfoEndpoint:      githubUserInfoEndpoint,
		ClientID:              clientID,
		ClientSecret:          clientSecret,
		RedirectURI:           redirectURI,
		Scopes:                scopes,
		AdditionalParams:      additionalParams,
	}

	base := oauth.NewOAuthExecutor(id, name, []flowmodel.InputData{}, properties, oAuthProps)
	exec, ok := base.(*oauth.OAuthExecutor)
	if !ok {
		panic("failed to cast GithubOAuthExecutor to OAuthExecutor")
	}
	return &GithubOAuthExecutor{
		OAuthExecutor: exec,
	}
}

// Execute executes the GitHub OAuth authentication flow.
func (g *GithubOAuthExecutor) Execute(ctx *flowmodel.NodeContext) (*flowmodel.ExecutorResponse, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Executing GitHub OAuth executor",
		log.String("executorID", g.GetID()), log.String("flowID", ctx.FlowID))

	execResp := &flowmodel.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	// Check if the required input data is provided
	if g.CheckInputData(ctx, execResp) {
		// If required input data is not provided, return incomplete status with redirection to github.
		logger.Debug("Required input data for GitHub OAuth executor is not provided")

		err := g.BuildAuthorizeFlow(ctx, execResp)
		if err != nil {
			return nil, err
		}

		logger.Debug("GitHub OAuth executor execution completed",
			log.String("status", string(execResp.Status)))
	} else {
		err := g.ProcessAuthFlowResponse(ctx, execResp)
		if err != nil {
			return nil, err
		}

		logger.Debug("GitHub OAuth executor execution completed",
			log.String("status", string(execResp.Status)),
			log.Bool("isAuthenticated", execResp.AuthenticatedUser.IsAuthenticated))
	}

	return execResp, nil
}

// ProcessAuthFlowResponse processes the response from the GitHub OAuth authentication flow.
func (o *GithubOAuthExecutor) ProcessAuthFlowResponse(ctx *flowmodel.NodeContext,
	execResp *flowmodel.ExecutorResponse) error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName),
		log.String("executorID", o.GetID()), log.String("flowID", ctx.FlowID))
	logger.Debug("Processing GitHub OAuth flow response")

	code, ok := ctx.UserInputData["code"]
	if ok && code != "" {
		tokenResp, err := o.ExchangeCodeForToken(ctx, execResp, code)
		if err != nil {
			logger.Error("Failed to exchange code for a token", log.Error(err))
			return fmt.Errorf("failed to exchange code for a token: %w", err)
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

// GetUserInfo fetches user information from the GitHub OAuth provider using the access token.
func (o *GithubOAuthExecutor) GetUserInfo(ctx *flowmodel.NodeContext, execResp *flowmodel.ExecutorResponse,
	accessToken string) (map[string]string, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Fetching user info from Github OAuth provider",
		log.String("userInfoEndpoint", o.GetUserInfoEndpoint()))

	// Create HTTP request
	req, err := http.NewRequest("GET", o.GetUserInfoEndpoint(), nil)
	if err != nil {
		logger.Error("Failed to create userinfo request", log.Error(err))
		return nil, fmt.Errorf("failed to create userinfo request: %w", err)
	}
	req.Header.Set(constants.AuthorizationHeaderName, constants.TokenTypeBearer+" "+accessToken)
	req.Header.Set(constants.AcceptHeaderName, "application/json")

	// Execute the request
	logger.Debug("Sending userinfo request to GitHub OAuth provider")

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
	logger.Debug("Userinfo response received from GitHub OAuth provider",
		log.Int("statusCode", resp.StatusCode))

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

	// If the user info doesn't contain the email, but scopes contain "user" or "user:email",
	// then fetch the primary email from the email endpoint.
	email := userInfo["email"]
	scopes := o.GetOAuthProperties().Scopes
	if (email == nil || email == "") &&
		(slices.Contains(scopes, userScope) || slices.Contains(scopes, userEmailScope)) {
		logger.Debug("Fetching user email from Github email endpoint",
			log.String("githubUserEmailEndpoint", githubUserEmailEndpoint))

		req, err = http.NewRequest("GET", githubUserEmailEndpoint, nil)
		if err != nil {
			logger.Error("Failed to create user email request: ", log.Error(err))
			return nil, errors.New("failed to create user email request: " + err.Error())
		}
		req.Header.Set(constants.AuthorizationHeaderName, constants.TokenTypeBearer+" "+accessToken)
		req.Header.Set(constants.AcceptHeaderName, "application/json")

		resp, err = client.Do(req)
		if err != nil {
			execResp.Status = flowconst.ExecFailure
			execResp.FailureReason = "Failed to fetch user email: " + err.Error()
			return nil, nil
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				logger.Error("Failed to close response body: ", log.Error(err))
			}
		}()
		logger.Debug("User email response received from GitHub", log.Int("statusCode", resp.StatusCode))

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			execResp.Status = flowconst.ExecFailure
			execResp.FailureReason = fmt.Sprintf("User email request failed with status %s: %s",
				resp.Status, string(body))
			return nil, nil
		}

		var emails []map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
			logger.Error("Failed to decode email response: ", log.Error(err))
			return nil, errors.New("failed to decode email response: " + err.Error())
		}

		// Set the primary email in the user info map.
		for _, emailEntry := range emails {
			if isPrimary, ok := emailEntry["primary"].(bool); ok && isPrimary {
				if primaryEmail, ok := emailEntry["email"].(string); ok {
					userInfo["email"] = primaryEmail
					break
				}
			}
		}
	}

	return systemutils.ConvertInterfaceMapToStringMap(userInfo), nil
}

// validateTokenResponse validates the token response received from the github.
func (o *GithubOAuthExecutor) validateTokenResponse(tokenResp *model.TokenResponse) error {
	if tokenResp == nil {
		return errors.New("token response is nil")
	}
	if tokenResp.AccessToken == "" {
		return errors.New("access token is empty in the token response")
	}
	return nil
}

// getAuthenticatedUserWithAttributes retrieves the authenticated user with attributes from github.
func (o *GithubOAuthExecutor) getAuthenticatedUserWithAttributes(ctx *flowmodel.NodeContext,
	execResp *flowmodel.ExecutorResponse, accessToken string) (*authndto.AuthenticatedUser, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName),
		log.String(log.LoggerKeyExecutorID, o.GetID()),
		log.String(log.LoggerKeyFlowID, ctx.FlowID))

	userInfo, err := o.GetUserInfo(ctx, execResp, accessToken)
	if err != nil {
		logger.Error("Failed to get user info", log.Error(err))
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	if execResp.Status == flowconst.ExecFailure {
		return nil, nil
	}

	// Resolve user with the id claim.
	// TODO: For now assume `id` is the unique identifier for the user always.
	sub, ok := userInfo["id"]
	if !ok || sub == "" {
		execResp.Status = flowconst.ExecFailure
		execResp.FailureReason = "id claim not found in the response."
		return nil, nil
	}

	filters := map[string]interface{}{"sub": sub}
	userID, err := o.IdentifyUser(filters, execResp)
	if err != nil {
		return nil, fmt.Errorf("failed to identify user with id claim: %w", err)
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
		execResp.FailureReason = "User already exists with the provided id claim."
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
		if key != "username" && key != "sub" && key != "id" {
			attributes[key] = value
		}
	}
	if userID != "" {
		attributes["user_id"] = userID
	}

	return attributes
}
