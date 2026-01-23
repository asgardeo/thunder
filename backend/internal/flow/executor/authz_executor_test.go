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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	authncm "github.com/asgardeo/thunder/internal/authn/common"
	authzsvc "github.com/asgardeo/thunder/internal/authz"
	"github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/internal/flow/core"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/tests/mocks/authzmock"
	"github.com/asgardeo/thunder/tests/mocks/flow/coremock"
)

// createTestAuthzExecutor creates an authorization executor with mocks for testing
func createTestAuthzExecutor(t *testing.T,
	mockAuthzService *authzmock.AuthorizationServiceInterfaceMock) *authorizationExecutor {
	mockFlowFactory := coremock.NewFlowFactoryInterfaceMock(t)

	// Mock the CreateExecutor method to return a base executor
	mockFlowFactory.On("CreateExecutor", ExecutorNameAuthorization, common.ExecutorTypeUtility,
		[]common.Input{}, []common.Input{}).
		Return(createMockExecutor(t, "AuthorizationExecutor", common.ExecutorTypeUtility))

	return newAuthorizationExecutor(mockFlowFactory, mockAuthzService)
}

// createMockExecutor creates a mock executor for testing purposes
func createMockExecutor(t *testing.T, name string, executorType common.ExecutorType) core.ExecutorInterface {
	mockExec := coremock.NewExecutorInterfaceMock(t)
	mockExec.On("GetName").Return(name).Maybe()
	mockExec.On("GetType").Return(executorType).Maybe()
	mockExec.On("GetDefaultInputs").Return([]common.Input{}).Maybe()
	mockExec.On("GetPrerequisites").Return([]common.Input{}).Maybe()
	return mockExec
}

func TestNewAuthorizationExecutor(t *testing.T) {
	mockAuthzService := authzmock.NewAuthorizationServiceInterfaceMock(t)
	executor := createTestAuthzExecutor(t, mockAuthzService)

	assert.NotNil(t, executor)
	assert.Equal(t, "AuthorizationExecutor", executor.GetName())
	prerequisites := executor.GetPrerequisites()
	assert.Empty(t, prerequisites)
}

func TestAuthorizationExecutor_Execute_Success(t *testing.T) {
	// Setup
	mockAuthzService := new(authzmock.AuthorizationServiceInterfaceMock)
	executor := createTestAuthzExecutor(t, mockAuthzService)

	ctx := &core.NodeContext{
		FlowID:   "test-flow",
		FlowType: common.FlowTypeAuthentication,
		AuthenticatedUser: authncm.AuthenticatedUser{
			IsAuthenticated: true,
			UserID:          "user123",
			Attributes: map[string]interface{}{
				"groups": []string{"group1", "group2"},
			},
		},
		RuntimeData: map[string]string{
			requestedPermissionsKey: "read:documents write:documents delete:documents",
		},
	}

	expectedAuthorizedPerms := []string{"read:documents", "write:documents"}
	mockAuthzService.On("GetAuthorizedPermissions",
		mock.MatchedBy(func(req authzsvc.GetAuthorizedPermissionsRequest) bool {
			return req.UserID == "user123" &&
				len(req.GroupIDs) == 2 &&
				req.GroupIDs[0] == "group1" &&
				req.GroupIDs[1] == "group2" &&
				len(req.RequestedPermissions) == 3 &&
				req.RequestedPermissions[0] == "read:documents" &&
				req.RequestedPermissions[1] == "write:documents" &&
				req.RequestedPermissions[2] == "delete:documents"
		})).Return(&authzsvc.GetAuthorizedPermissionsResponse{
		AuthorizedPermissions: expectedAuthorizedPerms,
	}, nil)

	// Execute
	resp, err := executor.Execute(ctx)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, common.ExecComplete, resp.Status)
	assert.Equal(t, "read:documents write:documents", resp.RuntimeData[authorizedPermissionsKey])

	mockAuthzService.AssertExpectations(t)
}

func TestAuthorizationExecutor_Execute_PartialPermissions(t *testing.T) {
	// Setup - user requests multiple permissions but only gets some
	mockAuthzService := new(authzmock.AuthorizationServiceInterfaceMock)
	executor := createTestAuthzExecutor(t, mockAuthzService)

	ctx := &core.NodeContext{
		FlowID:   "test-flow",
		FlowType: common.FlowTypeAuthentication,
		AuthenticatedUser: authncm.AuthenticatedUser{
			IsAuthenticated: true,
			UserID:          "user123",
		},
		RuntimeData: map[string]string{
			requestedPermissionsKey: "read:documents write:documents delete:documents",
		},
	}

	// User only has read permission
	mockAuthzService.On("GetAuthorizedPermissions", mock.Anything).Return(
		&authzsvc.GetAuthorizedPermissionsResponse{
			AuthorizedPermissions: []string{"read:documents"},
		}, nil)

	// Execute
	resp, err := executor.Execute(ctx)

	// Assert - should succeed with partial permissions
	assert.NoError(t, err)
	assert.Equal(t, common.ExecComplete, resp.Status)
	assert.Equal(t, "read:documents", resp.RuntimeData[authorizedPermissionsKey])

	mockAuthzService.AssertExpectations(t)
}

func TestAuthorizationExecutor_Execute_NoPermissions(t *testing.T) {
	// Setup - user has no permissions at all
	mockAuthzService := new(authzmock.AuthorizationServiceInterfaceMock)
	executor := createTestAuthzExecutor(t, mockAuthzService)

	ctx := &core.NodeContext{
		FlowID:   "test-flow",
		FlowType: common.FlowTypeAuthentication,
		AuthenticatedUser: authncm.AuthenticatedUser{
			IsAuthenticated: true,
			UserID:          "user123",
		},
		RuntimeData: map[string]string{
			requestedPermissionsKey: "read:documents write:documents",
		},
	}

	mockAuthzService.On("GetAuthorizedPermissions", mock.Anything).Return(
		&authzsvc.GetAuthorizedPermissionsResponse{
			AuthorizedPermissions: []string{},
		}, nil)

	// Execute
	resp, err := executor.Execute(ctx)

	// Assert - should succeed with empty permissions
	assert.NoError(t, err)
	assert.Equal(t, common.ExecComplete, resp.Status)
	assert.Equal(t, "", resp.RuntimeData[authorizedPermissionsKey])

	mockAuthzService.AssertExpectations(t)
}

func TestAuthorizationExecutor_Execute_NotAuthenticated(t *testing.T) {
	// Setup - user not authenticated
	mockAuthzService := new(authzmock.AuthorizationServiceInterfaceMock)
	executor := createTestAuthzExecutor(t, mockAuthzService)

	ctx := &core.NodeContext{
		FlowID:   "test-flow",
		FlowType: common.FlowTypeAuthentication,
		AuthenticatedUser: authncm.AuthenticatedUser{
			IsAuthenticated: false,
		},
		RuntimeData: make(map[string]string),
	}

	// Execute
	resp, err := executor.Execute(ctx)

	// Assert - should FAIL (changed behavior from original design)
	assert.NoError(t, err)
	assert.Equal(t, common.ExecFailure, resp.Status)
	assert.Equal(t, failureReasonUserNotAuthenticated, resp.FailureReason)

	// Service should NOT be called
	mockAuthzService.AssertNotCalled(t, "GetAuthorizedPermissions")
}

func TestAuthorizationExecutor_Execute_ServiceError(t *testing.T) {
	// Setup - service returns error
	mockAuthzService := new(authzmock.AuthorizationServiceInterfaceMock)
	executor := createTestAuthzExecutor(t, mockAuthzService)

	ctx := &core.NodeContext{
		FlowID:   "test-flow",
		FlowType: common.FlowTypeAuthentication,
		AuthenticatedUser: authncm.AuthenticatedUser{
			IsAuthenticated: true,
			UserID:          "user123",
		},
		RuntimeData: map[string]string{
			requestedPermissionsKey: "read:documents write:documents",
		},
	}

	mockAuthzService.On("GetAuthorizedPermissions", mock.Anything).Return(
		nil, &serviceerror.ServiceError{Error: "service error"})

	// Execute
	resp, err := executor.Execute(ctx)

	// Assert - should fail the flow
	assert.NoError(t, err)
	assert.Equal(t, common.ExecFailure, resp.Status)

	mockAuthzService.AssertExpectations(t)
}

func TestAuthorizationExecutor_Execute_NoRequestedPermissions(t *testing.T) {
	// This test verifies behavior when extractRequestedPermissions returns empty
	// The service should NOT be called, and should return early with ExecComplete

	mockAuthzService := new(authzmock.AuthorizationServiceInterfaceMock)
	executor := createTestAuthzExecutor(t, mockAuthzService)

	ctx := &core.NodeContext{
		FlowID:   "test-flow",
		FlowType: common.FlowTypeAuthentication,
		AuthenticatedUser: authncm.AuthenticatedUser{
			IsAuthenticated: true,
			UserID:          "user123",
		},
		RuntimeData: make(map[string]string), // No requestedPermissionsKey
	}

	// Execute
	resp, err := executor.Execute(ctx)

	// Assert - should return early without calling service
	assert.NoError(t, err)
	assert.Equal(t, common.ExecComplete, resp.Status)
	assert.Empty(t, resp.RuntimeData[authorizedPermissionsKey])

	// Service should NOT be called
	mockAuthzService.AssertNotCalled(t, "GetAuthorizedPermissions")
}

func TestAuthorizationExecutor_ExtractGroupIDs_FromAttributes(t *testing.T) {
	mockAuthzService := authzmock.NewAuthorizationServiceInterfaceMock(t)
	executor := createTestAuthzExecutor(t, mockAuthzService)

	tests := []struct {
		name       string
		attributes map[string]interface{}
		expected   []string
	}{
		{
			name: "Groups as string slice",
			attributes: map[string]interface{}{
				"groups": []string{"group1", "group2", "group3"},
			},
			expected: []string{"group1", "group2", "group3"},
		},
		{
			name: "Groups as interface slice",
			attributes: map[string]interface{}{
				"groups": []interface{}{"group1", "group2"},
			},
			expected: []string{"group1", "group2"},
		},
		{
			name: "Groups as single string",
			attributes: map[string]interface{}{
				"groups": "single-group",
			},
			expected: []string{"single-group"},
		},
		{
			name:       "No groups attribute",
			attributes: map[string]interface{}{},
			expected:   []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &core.NodeContext{
				AuthenticatedUser: authncm.AuthenticatedUser{
					Attributes: tt.attributes,
				},
				RuntimeData: make(map[string]string),
			}

			groupIDs := executor.extractGroupIDs(ctx)
			assert.Equal(t, tt.expected, groupIDs)
		})
	}
}

func TestAuthorizationExecutor_ExtractGroupIDs_FromRuntimeData(t *testing.T) {
	mockAuthzService := authzmock.NewAuthorizationServiceInterfaceMock(t)
	executor := createTestAuthzExecutor(t, mockAuthzService)

	ctx := &core.NodeContext{
		AuthenticatedUser: authncm.AuthenticatedUser{
			Attributes: map[string]interface{}{}, // No groups in attributes
		},
		RuntimeData: map[string]string{
			"groups": "[\"runtime-group1\", \"runtime-group2\"]",
		},
	}

	groupIDs := executor.extractGroupIDs(ctx)
	assert.Equal(t, []string{"runtime-group1", "runtime-group2"}, groupIDs)
}

func TestExtractRequestedPermissions(t *testing.T) {
	tests := []struct {
		name        string
		runtimeData map[string]string
		UserInputs  map[string]string
		expected    []string
	}{
		{
			name: "Space-separated permissions",
			runtimeData: map[string]string{
				requestedPermissionsKey: "read:documents write:documents delete:documents",
			},
			expected: []string{"read:documents", "write:documents", "delete:documents"},
		},
		{
			name: "Single permission",
			runtimeData: map[string]string{
				requestedPermissionsKey: "read:documents",
			},
			expected: []string{"read:documents"},
		},
		{
			name:        "No requested permissions",
			runtimeData: map[string]string{},
			expected:    []string{},
		},
		{
			name: "Empty string",
			runtimeData: map[string]string{
				requestedPermissionsKey: "",
			},
			expected: []string{},
		},
		{
			name: "Permissions from User Inputs",
			UserInputs: map[string]string{
				requestedPermissionsKey: "edit:documents share:documents",
			},
			expected: []string{"edit:documents", "share:documents"},
		},
		{
			name: "Permissions Priority to Runtime Data",
			runtimeData: map[string]string{
				requestedPermissionsKey: "view:documents delete:documents",
			},
			UserInputs: map[string]string{
				requestedPermissionsKey: "edit:documents share:documents",
			},
			expected: []string{"view:documents", "delete:documents"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &core.NodeContext{
				FlowType:    common.FlowTypeAuthentication,
				RuntimeData: tt.runtimeData,
				UserInputs:  tt.UserInputs,
			}

			permissions := extractRequestedPermissions(ctx)
			assert.Equal(t, tt.expected, permissions)
		})
	}
}

func TestAuthorizationExecutor_ExtractGroupIDs_WithNoGroups(t *testing.T) {
	mockAuthzService := authzmock.NewAuthorizationServiceInterfaceMock(t)
	executor := createTestAuthzExecutor(t, mockAuthzService)

	ctx := &core.NodeContext{
		AuthenticatedUser: authncm.AuthenticatedUser{
			IsAuthenticated: true,
			UserID:          "user123",
			Attributes:      map[string]interface{}{}, // No groups
		},
		RuntimeData: make(map[string]string),
	}

	groupIDs := executor.extractGroupIDs(ctx)
	assert.Empty(t, groupIDs)
}

func TestAuthorizationExecutor_Execute_WithMultipleGroups(t *testing.T) {
	mockAuthzService := new(authzmock.AuthorizationServiceInterfaceMock)
	executor := createTestAuthzExecutor(t, mockAuthzService)

	ctx := &core.NodeContext{
		FlowID:   "test-flow",
		FlowType: common.FlowTypeAuthentication,
		AuthenticatedUser: authncm.AuthenticatedUser{
			IsAuthenticated: true,
			UserID:          "user123",
			Attributes: map[string]interface{}{
				"groups": []string{"admin", "editor", "viewer"},
			},
		},
		RuntimeData: map[string]string{
			requestedPermissionsKey: "read:documents write:documents delete:documents",
		},
	}

	mockAuthzService.On("GetAuthorizedPermissions",
		mock.MatchedBy(func(req authzsvc.GetAuthorizedPermissionsRequest) bool {
			return req.UserID == "user123" &&
				len(req.GroupIDs) == 3 &&
				req.GroupIDs[0] == "admin" &&
				req.GroupIDs[1] == "editor" &&
				req.GroupIDs[2] == "viewer"
		})).Return(&authzsvc.GetAuthorizedPermissionsResponse{
		AuthorizedPermissions: []string{"read:documents", "write:documents", "delete:documents"},
	}, nil)

	// Execute
	resp, err := executor.Execute(ctx)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, common.ExecComplete, resp.Status)
	assert.Equal(t, "read:documents write:documents delete:documents", resp.RuntimeData[authorizedPermissionsKey])

	mockAuthzService.AssertExpectations(t)
}

func TestSetAuthorizedPermissions(t *testing.T) {
	tests := []struct {
		name        string
		permissions []string
		expected    string
	}{
		{
			name:        "Multiple permissions",
			permissions: []string{"read:documents", "write:documents", "delete:documents"},
			expected:    "read:documents write:documents delete:documents",
		},
		{
			name:        "Single permission",
			permissions: []string{"read:documents"},
			expected:    "read:documents",
		},
		{
			name:        "Empty permissions",
			permissions: []string{},
			expected:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			execResp := &common.ExecutorResponse{
				RuntimeData: make(map[string]string),
			}

			setAuthorizedPermissions(execResp, tt.permissions)
			assert.Equal(t, tt.expected, execResp.RuntimeData[authorizedPermissionsKey])
		})
	}
}

func TestAuthorizationExecutor_Execute_RegistrationFlow_UnauthenticatedWithoutPermissions(t *testing.T) {
	// Setup - registration flow with unauthenticated user and no requested permissions
	mockAuthzService := new(authzmock.AuthorizationServiceInterfaceMock)
	executor := createTestAuthzExecutor(t, mockAuthzService)

	ctx := &core.NodeContext{
		FlowID:   "test-registration-flow",
		FlowType: common.FlowTypeRegistration,
		AuthenticatedUser: authncm.AuthenticatedUser{
			IsAuthenticated: false,
		},
		RuntimeData: make(map[string]string),
	}

	// Execute
	resp, err := executor.Execute(ctx)

	// Assert - should succeed (bypass authentication check for registration)
	assert.NoError(t, err)
	assert.Equal(t, common.ExecComplete, resp.Status)
	assert.Empty(t, resp.RuntimeData[authorizedPermissionsKey])

	// Service should NOT be called since there are no requested permissions
	mockAuthzService.AssertNotCalled(t, "GetAuthorizedPermissions")
}

func TestAuthorizationExecutor_Execute_RegistrationFlow_UnauthenticatedWithPermissions(t *testing.T) {
	// Setup - registration flow with unauthenticated user but WITH requested permissions
	mockAuthzService := new(authzmock.AuthorizationServiceInterfaceMock)
	executor := createTestAuthzExecutor(t, mockAuthzService)

	ctx := &core.NodeContext{
		FlowID:   "test-registration-flow",
		FlowType: common.FlowTypeRegistration,
		AuthenticatedUser: authncm.AuthenticatedUser{
			IsAuthenticated: false,
			UserID:          "", // No user ID yet in registration
		},
		RuntimeData: map[string]string{
			requestedPermissionsKey: "read:documents write:documents",
		},
	}

	// Mock service call - even though user is not authenticated, registration flow allows it
	mockAuthzService.On("GetAuthorizedPermissions",
		mock.MatchedBy(func(req authzsvc.GetAuthorizedPermissionsRequest) bool {
			return req.UserID == "" &&
				len(req.GroupIDs) == 0 &&
				len(req.RequestedPermissions) == 2 &&
				req.RequestedPermissions[0] == "read:documents" &&
				req.RequestedPermissions[1] == "write:documents"
		})).Return(&authzsvc.GetAuthorizedPermissionsResponse{
		AuthorizedPermissions: []string{}, // New user typically has no permissions yet
	}, nil)

	// Execute
	resp, err := executor.Execute(ctx)

	// Assert - should succeed and call service (registration flow bypasses auth check)
	assert.NoError(t, err)
	assert.Equal(t, common.ExecComplete, resp.Status)
	assert.Equal(t, "", resp.RuntimeData[authorizedPermissionsKey])

	mockAuthzService.AssertExpectations(t)
}

func TestAuthorizationExecutor_Execute_RegistrationFlow_AuthenticatedWithPermissions(t *testing.T) {
	// Setup - registration flow with authenticated user (edge case but possible)
	mockAuthzService := new(authzmock.AuthorizationServiceInterfaceMock)
	executor := createTestAuthzExecutor(t, mockAuthzService)

	ctx := &core.NodeContext{
		FlowID:   "test-registration-flow",
		FlowType: common.FlowTypeRegistration,
		AuthenticatedUser: authncm.AuthenticatedUser{
			IsAuthenticated: true,
			UserID:          "existing-user-123",
			Attributes: map[string]interface{}{
				"groups": []string{"new-users"},
			},
		},
		RuntimeData: map[string]string{
			requestedPermissionsKey: "read:profile write:profile",
		},
	}

	expectedAuthorizedPerms := []string{"read:profile"}
	mockAuthzService.On("GetAuthorizedPermissions",
		mock.MatchedBy(func(req authzsvc.GetAuthorizedPermissionsRequest) bool {
			return req.UserID == "existing-user-123" &&
				len(req.GroupIDs) == 1 &&
				req.GroupIDs[0] == "new-users" &&
				len(req.RequestedPermissions) == 2
		})).Return(&authzsvc.GetAuthorizedPermissionsResponse{
		AuthorizedPermissions: expectedAuthorizedPerms,
	}, nil)

	// Execute
	resp, err := executor.Execute(ctx)

	// Assert - should succeed and call service
	assert.NoError(t, err)
	assert.Equal(t, common.ExecComplete, resp.Status)
	assert.Equal(t, "read:profile", resp.RuntimeData[authorizedPermissionsKey])

	mockAuthzService.AssertExpectations(t)
}

func TestAuthorizationExecutor_Execute_NonRegistrationFlow_UnauthenticatedShouldFail(t *testing.T) {
	// Setup - non-registration flow types should fail if unauthenticated
	mockAuthzService := new(authzmock.AuthorizationServiceInterfaceMock)
	executor := createTestAuthzExecutor(t, mockAuthzService)

	testCases := []struct {
		name     string
		flowType common.FlowType
	}{
		{
			name:     "Authentication flow",
			flowType: common.FlowTypeAuthentication,
		},
		{
			name:     "User onboarding flow",
			flowType: common.FlowTypeUserOnboarding,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := &core.NodeContext{
				FlowID:   "test-flow",
				FlowType: tc.flowType,
				AuthenticatedUser: authncm.AuthenticatedUser{
					IsAuthenticated: false,
				},
				RuntimeData: map[string]string{
					requestedPermissionsKey: "read:documents",
				},
			}

			// Execute
			resp, err := executor.Execute(ctx)

			// Assert - should fail
			assert.NoError(t, err)
			assert.Equal(t, common.ExecFailure, resp.Status)
			assert.Equal(t, failureReasonUserNotAuthenticated, resp.FailureReason)

			// Service should NOT be called
			mockAuthzService.AssertNotCalled(t, "GetAuthorizedPermissions")
		})
	}
}
