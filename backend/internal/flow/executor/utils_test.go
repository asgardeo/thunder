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

package executor

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	authncm "github.com/asgardeo/thunder/internal/authn/common"
	"github.com/asgardeo/thunder/internal/entityprovider"
	"github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/internal/flow/core"
	"github.com/asgardeo/thunder/tests/mocks/flow/coremock"
)

type UtilsTestSuite struct {
	suite.Suite
}

func TestUtilsTestSuite(t *testing.T) {
	suite.Run(t, new(UtilsTestSuite))
}

func (s *UtilsTestSuite) TestGetAuthnServiceName() {
	tests := []struct {
		name         string
		executorName string
		expectedName string
	}{
		{"BasicAuth executor", ExecutorNameBasicAuth, authncm.AuthenticatorCredentials},
		{"SMS Auth executor", ExecutorNameSMSAuth, authncm.AuthenticatorSMSOTP},
		{"OAuth executor", ExecutorNameOAuth, authncm.AuthenticatorOAuth},
		{"OIDC Auth executor", ExecutorNameOIDCAuth, authncm.AuthenticatorOIDC},
		{"GitHub Auth executor", ExecutorNameGitHubAuth, authncm.AuthenticatorGithub},
		{"Google Auth executor", ExecutorNameGoogleAuth, authncm.AuthenticatorGoogle},
		{"Unknown executor returns empty string", "UnknownExecutor", ""},
		{"Provisioning executor returns empty string", ExecutorNameProvisioning, ""},
		{"AuthAssert executor returns empty string", ExecutorNameAuthAssert, ""},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := getAuthnServiceName(tt.executorName)
			s.Equal(tt.expectedName, result)
		})
	}
}

// createMockAuthExecutor creates a mock executor for OAuth/OIDC authentication.
func createMockAuthExecutor(t *testing.T, executorName string) core.ExecutorInterface {
	mockExec := coremock.NewExecutorInterfaceMock(t)
	mockExec.On("GetName").Return(executorName).Maybe()
	mockExec.On("GetType").Return(common.ExecutorTypeAuthentication).Maybe()
	mockExec.On("GetDefaultInputs").Return([]common.Input{
		{Identifier: "code", Type: "string", Required: true},
	}).Maybe()
	mockExec.On("GetPrerequisites").Return([]common.Input{}).Maybe()
	mockExec.On("HasRequiredInputs", mock.Anything, mock.Anything).Return(
		func(ctx *core.NodeContext, execResp *common.ExecutorResponse) bool {
			if code, ok := ctx.UserInputs["code"]; ok && code != "" {
				return true
			}
			if len(ctx.NodeInputs) == 0 {
				return true
			}
			execResp.Inputs = []common.Input{{Identifier: "code", Type: "string", Required: true}}
			return false
		}).Maybe()
	return mockExec
}

func (s *UtilsTestSuite) TestGetUserAttribute() {
	tests := []struct {
		name         string
		user         *entityprovider.Entity
		attributeKey string
		expectedVal  string
		expectError  bool
	}{
		{
			name: "Success case",
			user: &entityprovider.Entity{
				Attributes: []byte(`{"email":"user@example.com"}`),
			},
			attributeKey: "email",
			expectedVal:  "user@example.com",
			expectError:  false,
		},
		{
			name:         "Nil user",
			user:         nil,
			attributeKey: "email",
			expectError:  true,
		},
		{
			name: "Empty attributes",
			user: &entityprovider.Entity{
				Attributes: []byte(``),
			},
			attributeKey: "email",
			expectError:  true,
		},
		{
			name: "Invalid JSON attributes",
			user: &entityprovider.Entity{
				Attributes: []byte(`invalid-json`),
			},
			attributeKey: "email",
			expectError:  true,
		},
		{
			name: "Attribute not found",
			user: &entityprovider.Entity{
				Attributes: []byte(`{"other":"data"}`),
			},
			attributeKey: "email",
			expectError:  true,
		},
		{
			name: "Non-string attribute value",
			user: &entityprovider.Entity{
				Attributes: []byte(`{"email":123}`),
			},
			attributeKey: "email",
			expectError:  true, // Strictly expecting an error for non-strings now!
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			val, err := GetUserAttribute(tt.user, tt.attributeKey)
			if tt.expectError {
				s.Error(err)
				s.Empty(val)
			} else {
				s.NoError(err)
				s.Equal(tt.expectedVal, val)
			}
		})
	}
}

func (s *UtilsTestSuite) TestFindInputByType() {
	tests := []struct {
		name        string
		inputs      []common.Input
		inputType   string
		expected    common.Input
		expectFound bool
	}{
		{
			name:        "Empty inputs",
			inputs:      []common.Input{},
			inputType:   common.InputTypeEmail,
			expected:    common.Input{},
			expectFound: false,
		},
		{
			name: "Type found",
			inputs: []common.Input{
				{Identifier: "mobile", Type: "phone"},
				{Identifier: "workEmail", Type: common.InputTypeEmail},
			},
			inputType:   common.InputTypeEmail,
			expected:    common.Input{Identifier: "workEmail", Type: common.InputTypeEmail},
			expectFound: true,
		},
		{
			name: "Type not found",
			inputs: []common.Input{
				{Identifier: "mobile", Type: "phone"},
			},
			inputType:   common.InputTypeEmail,
			expected:    common.Input{},
			expectFound: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			res, found := findInputByType(tt.inputs, tt.inputType)
			s.Equal(tt.expectFound, found)
			s.Equal(tt.expected, res)
		})
	}
}

func (s *UtilsTestSuite) TestResolveInputIdentifierByType() {
	tests := []struct {
		name      string
		ctx       *core.NodeContext
		inputType string
		fallback  string
		expected  string
	}{
		{
			name: "Type found in NodeInputs",
			ctx: &core.NodeContext{
				NodeInputs: []common.Input{
					{Identifier: "customEmailIdentifier", Type: common.InputTypeEmail},
				},
			},
			inputType: common.InputTypeEmail,
			fallback:  "defaultEmail",
			expected:  "customEmailIdentifier",
		},
		{
			name: "Type not found, returns fallback",
			ctx: &core.NodeContext{
				NodeInputs: []common.Input{
					{Identifier: "phone", Type: "mobile"},
				},
			},
			inputType: common.InputTypeEmail,
			fallback:  "defaultEmail",
			expected:  "defaultEmail",
		},
		{
			name: "Empty NodeInputs, returns fallback",
			ctx: &core.NodeContext{
				NodeInputs: []common.Input{},
			},
			inputType: common.InputTypeEmail,
			fallback:  "defaultEmail",
			expected:  "defaultEmail",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			res := resolveInputIdentifierByType(tt.ctx, tt.inputType, tt.fallback)
			s.Equal(tt.expected, res)
		})
	}
}
