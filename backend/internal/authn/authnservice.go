/*
 * Copyright (c) 2025, WSO2 LLC. (http://www.wso2.com).
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

// Package authn provides the implementation for handling authentication requests.
package authn

import (
	"errors"
	"time"

	"github.com/asgardeo/thunder/internal/authn/constants"
	authndto "github.com/asgardeo/thunder/internal/authn/dto"
	"github.com/asgardeo/thunder/internal/authn/model"
	"github.com/asgardeo/thunder/internal/flow"
	flowconst "github.com/asgardeo/thunder/internal/flow/constants"
	flowmodel "github.com/asgardeo/thunder/internal/flow/model"
	"github.com/asgardeo/thunder/internal/oauth/jwt"
	authzutils "github.com/asgardeo/thunder/internal/oauth/oauth2/authz/utils"
	oauthmodel "github.com/asgardeo/thunder/internal/oauth/oauth2/model"
	sessionmodel "github.com/asgardeo/thunder/internal/oauth/session/model"
	sessionstore "github.com/asgardeo/thunder/internal/oauth/session/store"
	sessionutils "github.com/asgardeo/thunder/internal/oauth/session/utils"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	systemutils "github.com/asgardeo/thunder/internal/system/utils"
)

var loggerComponentName = "AuthenticationService"

// AuthenticationServiceInterface defines the interface for authentication service.
type AuthenticationServiceInterface interface {
	Execute(sessionDataKey, flowID, actionID string,
		inputs map[string]string) (*model.AuthNResponse, *serviceerror.ServiceError)
}

// AuthenticationService implements the AuthenticationServiceInterface.
type AuthenticationService struct{}

// NewAuthenticationService creates a new instance of AuthenticationService.
func NewAuthenticationService() AuthenticationServiceInterface {
	return &AuthenticationService{}
}

// Execute executes the authentication flow based on the provided data.
func (s *AuthenticationService) Execute(sessionDataKey, flowID, actionID string,
	inputs map[string]string) (*model.AuthNResponse, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if !isValidFlow(sessionDataKey, flowID) {
		return nil, &constants.ErrorInvalidAuthFlow
	}

	var sessionData *sessionmodel.SessionData
	var flowStep *flowmodel.FlowStep
	var flowErr *serviceerror.ServiceError

	if isNewFlow(sessionDataKey, flowID) {
		var appID string
		var svcErr *serviceerror.ServiceError
		sessionData, appID, svcErr = loadAndUpdateAuthZSession(sessionDataKey)
		if svcErr != nil {
			logger.Error("Failed to load session data", log.Any("serviceError", svcErr))
			return nil, svcErr
		}

		flowStep, flowErr = executeNewFlow(appID, flowID, actionID, inputs, logger)
	} else {
		flowStep, flowErr = executeExistingFlow(flowID, actionID, inputs, logger)
	}

	if flowErr != nil {
		return nil, handleFlowError(flowErr, logger)
	}

	authResp, svcErr := handleFlowResponse(sessionDataKey, flowID, flowStep, sessionData, logger)
	if svcErr != nil {
		logger.Error("Failed to handle flow response", log.Any("serviceError", svcErr))
		return nil, svcErr
	}

	return authResp, nil
}

// isValidFlow checks if the provided data is for a valid authentication flow.
func isValidFlow(sessionDataKey, flowID string) bool {
	if sessionDataKey == "" && flowID == "" {
		return false
	} else if sessionDataKey != "" && flowID != "" {
		return false
	}
	return true
}

// isNewFlow checks if the provided data is for a new flow.
func isNewFlow(sessionDataKey, flowID string) bool {
	return sessionDataKey != "" && flowID == ""
}

// loadAndUpdateAuthZSession loads the session data for the given sessionDataKey
// and updates the session data store accordingly. It returns the session data, application ID,
// and a service error if any.
func loadAndUpdateAuthZSession(sessionDataKey string) (*sessionmodel.SessionData, string,
	*serviceerror.ServiceError) {
	sessionData, svcErr := getSessionData(sessionDataKey)
	if svcErr != nil {
		return nil, "", svcErr
	}

	// Remove the previous session data if it exists.
	sessionDataStore := sessionstore.GetSessionDataStore()
	sessionDataStore.ClearSession(sessionDataKey)

	appID := sessionData.OAuthParameters.AppID
	if appID == "" {
		return nil, "", &constants.ErrorAppIDNotFound
	}

	return sessionData, appID, nil
}

// executeNewFlow executes a new authentication flow with the provided inputs.
func executeNewFlow(appID, flowID, actionID string, inputs map[string]string, logger *log.Logger) (
	*flowmodel.FlowStep, *serviceerror.ServiceError) {
	flowStep, flowErr := flow.GetFlowService().Execute(appID, flowID, actionID,
		flowconst.FlowTypeAuthentication, inputs)
	if flowErr != nil {
		return nil, handleFlowError(flowErr, logger)
	}
	return flowStep, nil
}

// executeExistingFlow executes an existing authentication flow with the provided inputs.
func executeExistingFlow(flowID, actionID string, inputs map[string]string,
	logger *log.Logger) (*flowmodel.FlowStep, *serviceerror.ServiceError) {
	flowStep, flowErr := flow.GetFlowService().Execute("", flowID, actionID, "", inputs)
	if flowErr != nil {
		return nil, handleFlowError(flowErr, logger)
	}
	return flowStep, nil
}

// getSessionData retrieves the session data for the given key from the session data store.
func getSessionData(key string) (*sessionmodel.SessionData, *serviceerror.ServiceError) {
	sessionDataStore := sessionstore.GetSessionDataStore()
	ok, sessionData := sessionDataStore.GetSession(key)
	if !ok {
		return nil, &constants.ErrorSessionNotFound
	}

	return &sessionData, nil
}

// handleFlowError handles errors that occur during flow execution and returns an appropriate authentication error.
func handleFlowError(flowErr *serviceerror.ServiceError, logger *log.Logger) *serviceerror.ServiceError {
	var retErr *serviceerror.ServiceError
	if flowErr.Type == serviceerror.ClientErrorType {
		retErr = &constants.ErrorFlowExecutionClientError
	} else {
		logger.Error("An error occurred while executing the flow", log.Any("error", flowErr))
		retErr = &constants.ErrorFlowExecutionServerError
	}
	retErr.ErrorDescription = flowErr.ErrorDescription
	return retErr
}

// handleFlowResponse processes the flow step and updates the session data store accordingly.
// It returns an AuthNResponse if the flow is complete, or an error if any issues arise during processing.
func handleFlowResponse(sessionDataKey, requestFlowID string, flowStep *flowmodel.FlowStep,
	sessionData *sessionmodel.SessionData, logger *log.Logger) (*model.AuthNResponse, *serviceerror.ServiceError) {
	sessionDataStore := sessionstore.GetSessionDataStore()
	if isNewFlow(sessionDataKey, requestFlowID) {
		if flowStep.Status == flowconst.FlowStatusIncomplete {
			logger.Debug("Flow execution is incomplete, storing session data", log.String("flowID", flowStep.FlowID))
			sessionDataStore.AddSession(flowStep.FlowID, *sessionData)
		}
	} else {
		if flowStep.Status == flowconst.FlowStatusComplete {
			var svcErr *serviceerror.ServiceError
			sessionData, svcErr = getSessionData(flowStep.FlowID)
			if svcErr != nil {
				return nil, svcErr
			}
		}
		if flowStep.Status != flowconst.FlowStatusIncomplete {
			sessionDataStore.ClearSession(flowStep.FlowID)
		}
	}

	if flowStep.Status == flowconst.FlowStatusComplete {
		logger.Debug("Flow execution completed successfully", log.String("flowID", flowStep.FlowID))
		return handleCompletedFlow(flowStep, sessionData)
	}

	return &model.AuthNResponse{
		FlowID:        flowStep.FlowID,
		StepID:        flowStep.StepID,
		FlowStatus:    string(flowStep.Status),
		Type:          string(flowStep.Type),
		Data:          flowStep.Data,
		FailureReason: flowStep.FailureReason,
	}, nil
}

// handleCompletedFlow handles the flow step when the authentication flow is completed.
// It decodes the user attributes from the flow assertion and constructs a new session data object.
func handleCompletedFlow(flowStep *flowmodel.FlowStep, sessionData *sessionmodel.SessionData) (
	*model.AuthNResponse, *serviceerror.ServiceError) {
	userID, userAttributes, err := decodeAttributesFromAssertion(flowStep.Assertion)
	if err != nil {
		svcErr := constants.ErrorWhileDecodingFlowAssertion
		svcErr.ErrorDescription = err.Error()
		return nil, &svcErr
	}

	newSessionDataKey := sessionutils.GenerateNewSessionDataKey()
	newSessionData := &sessionmodel.SessionData{
		OAuthParameters: oauthmodel.OAuthParameters{
			SessionDataKey: newSessionDataKey,
			ClientID:       sessionData.OAuthParameters.ClientID,
			RedirectURI:    sessionData.OAuthParameters.RedirectURI,
			Scopes:         sessionData.OAuthParameters.Scopes,
			State:          sessionData.OAuthParameters.State,
		},
		AuthTime: time.Now(),
		AuthenticatedUser: authndto.AuthenticatedUser{
			IsAuthenticated: true,
			UserID:          userID,
			Attributes:      userAttributes,
		},
	}

	sessionDataStore := sessionstore.GetSessionDataStore()
	sessionDataStore.ClearSession(flowStep.FlowID)
	sessionDataStore.AddSession(newSessionDataKey, *newSessionData)

	redirectURI := authzutils.GetAuthorizationEndpoint()
	queryParams := map[string]string{
		"sessionDataKey": newSessionDataKey,
	}

	redirectURI, err = systemutils.GetURIWithQueryParams(redirectURI, queryParams)
	if err != nil {
		svcErr := constants.ErrorConstructingRedirectURI
		svcErr.ErrorDescription = err.Error()
		return nil, &svcErr
	}

	return &model.AuthNResponse{
		FlowID:     flowStep.FlowID,
		FlowStatus: string(flowStep.Status),
		Data: flowmodel.FlowData{
			RedirectURL: redirectURI,
		},
	}, nil
}

// decodeAttributesFromAssertion decodes user attributes from the flow assertion JWT.
// It returns the user ID, a map of user attributes, and an error if any.
func decodeAttributesFromAssertion(assertion string) (string, map[string]string, error) {
	if assertion == "" {
		return "", nil, errors.New("The assertion could not be found in the flow response")
	}

	_, jwtPayload, err := jwt.DecodeJWT(assertion)
	if err != nil {
		return "", nil, errors.New("Failed to decode the JWT token: " + err.Error())
	}

	userAttributes := make(map[string]string)
	userID := ""
	for key, value := range jwtPayload {
		switch key {
		case "sub":
			if strValue, ok := value.(string); ok {
				userID = strValue
			} else {
				return "", nil, errors.New("JWT 'sub' claim is not a string")
			}
		case "username":
			if strValue, ok := value.(string); ok {
				userAttributes["username"] = strValue
			} else {
				return "", nil, errors.New("JWT 'username' claim is not a string")
			}
		case "email":
			if strValue, ok := value.(string); ok {
				userAttributes["email"] = strValue
			} else {
				return "", nil, errors.New("JWT 'email' claim is not a string")
			}
		case "firstName":
			if strValue, ok := value.(string); ok {
				userAttributes["firstName"] = strValue
			} else {
				return "", nil, errors.New("JWT 'firstName' claim is not a string")
			}
		case "lastName":
			if strValue, ok := value.(string); ok {
				userAttributes["lastName"] = strValue
			} else {
				return "", nil, errors.New("JWT 'lastName' claim is not a string")
			}
		}
	}

	return userID, userAttributes, nil
}
