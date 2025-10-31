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
	"errors"
	"time"

	authncm "github.com/asgardeo/thunder/internal/authn/common"
	authngithub "github.com/asgardeo/thunder/internal/authn/github"
	authnoauth "github.com/asgardeo/thunder/internal/authn/oauth"
	"github.com/asgardeo/thunder/internal/executor/oauth"
	"github.com/asgardeo/thunder/internal/executor/oauth/model"
	flowconst "github.com/asgardeo/thunder/internal/flow/common/constants"
	flowmodel "github.com/asgardeo/thunder/internal/flow/common/model"
	"github.com/asgardeo/thunder/internal/idp"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	httpservice "github.com/asgardeo/thunder/internal/system/http"
	"github.com/asgardeo/thunder/internal/system/log"
	systemutils "github.com/asgardeo/thunder/internal/system/utils"
)

const loggerComponentName = "GithubOAuthExecutor"

// GithubOAuthExecutor implements the OAuth authentication executor for GitHub.
type GithubOAuthExecutor struct {
	*oauth.OAuthExecutor
	githubAuthService authngithub.GithubOAuthAuthnServiceInterface
}

var _ flowmodel.ExecutorInterface = (*GithubOAuthExecutor)(nil)

// NewGithubOAuthExecutor creates a new instance of GithubOAuthExecutor with the provided details.
func NewGithubOAuthExecutor(id, name string, properties map[string]string,
	clientID, clientSecret, redirectURI string, scopes []string,
	additionalParams map[string]string) oauth.OAuthExecutorInterface {
	oAuthProps := &model.OAuthExecProperties{
		AuthorizationEndpoint: authngithub.AuthorizeEndpoint,
		TokenEndpoint:         authngithub.TokenEndpoint,
		UserInfoEndpoint:      authngithub.UserInfoEndpoint,
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

	httpClient := httpservice.NewHTTPClientWithTimeout(flowconst.DefaultHTTPTimeout)
	authNService := authnoauth.NewOAuthAuthnService(
		httpClient,
		idp.NewIDPService(),
		endpoints,
	)
	githubAuthService := authngithub.NewGithubOAuthAuthnService(
		authNService,
		httpClient,
	)

	base := oauth.NewOAuthExecutorWithAuthService(id, name, []flowmodel.InputData{},
		properties, oAuthProps, authNService)
	exec, ok := base.(*oauth.OAuthExecutor)
	if !ok {
		panic("failed to cast GithubOAuthExecutor to OAuthExecutor")
	}

	return &GithubOAuthExecutor{
		OAuthExecutor:     exec,
		githubAuthService: githubAuthService,
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
			return err
		}
		if execResp.Status == flowconst.ExecFailure {
			return nil
		}

		if tokenResp.Scope == "" {
			logger.Error("Scopes are empty in the token response")
			execResp.AuthenticatedUser = authncm.AuthenticatedUser{
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

	// Add execution record for successful GitHub OAuth authentication
	execResp.ExecutionRecord = &flowmodel.NodeExecutionRecord{
		ExecutorName: authncm.AuthenticatorGithub,
		ExecutorType: flowconst.ExecutorTypeAuthentication,
		Timestamp:    time.Now().Unix(),
		Status:       flowconst.FlowStatusComplete,
	}

	return nil
}

// GetUserInfo fetches user information from the GitHub OAuth provider using the access token.
func (o *GithubOAuthExecutor) GetUserInfo(ctx *flowmodel.NodeContext, execResp *flowmodel.ExecutorResponse,
	accessToken string) (map[string]string, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName),
		log.String(log.LoggerKeyExecutorID, o.GetID()),
		log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Fetching user info from GitHub OAuth provider",
		log.String("userInfoEndpoint", o.GetUserInfoEndpoint()))

	userInfo, svcErr := o.githubAuthService.FetchUserInfo(o.GetID(), accessToken)
	if svcErr != nil {
		if svcErr.Type == serviceerror.ClientErrorType {
			execResp.Status = flowconst.ExecFailure
			execResp.FailureReason = svcErr.ErrorDescription
			return nil, nil
		}

		logger.Error("Failed to fetch user info", log.String("errorCode", svcErr.Code),
			log.String("errorDescription", svcErr.ErrorDescription))
		return nil, errors.New("failed to fetch user information")
	}

	return systemutils.ConvertInterfaceMapToStringMap(userInfo), nil
}

// getAuthenticatedUserWithAttributes retrieves the authenticated user with attributes from github.
func (o *GithubOAuthExecutor) getAuthenticatedUserWithAttributes(ctx *flowmodel.NodeContext,
	execResp *flowmodel.ExecutorResponse, accessToken string) (*authncm.AuthenticatedUser, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName),
		log.String(log.LoggerKeyExecutorID, o.GetID()),
		log.String(log.LoggerKeyFlowID, ctx.FlowID))

	userInfo, err := o.GetUserInfo(ctx, execResp, accessToken)
	if err != nil {
		return nil, err
	}
	if execResp.Status == flowconst.ExecFailure {
		return nil, nil
	}

	// Resolve user with the sub claim.
	sub, ok := userInfo["sub"]
	if !ok || sub == "" {
		execResp.Status = flowconst.ExecFailure
		execResp.FailureReason = "sub claim not found in the response."
		return nil, nil
	}

	user, svcErr := o.githubAuthService.GetInternalUser(sub)
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

				return &authncm.AuthenticatedUser{
					IsAuthenticated: false,
					Attributes:      getUserAttributes(userInfo, ""),
				}, nil
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
	userID := user.ID

	if execResp.Status == flowconst.ExecFailure {
		return nil, nil
	}

	authenticatedUser := authncm.AuthenticatedUser{
		IsAuthenticated:    true,
		UserID:             userID,
		OrganizationUnitID: user.OrganizationUnit,
		UserType:           user.Type,
		Attributes:         getUserAttributes(userInfo, userID),
	}

	return &authenticatedUser, nil
}

// getUserAttributes extracts user attributes from the user info map, excluding certain keys.
// TODO: Need to convert attributes as per the IDP to local attribute mapping when the support is implemented.
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
