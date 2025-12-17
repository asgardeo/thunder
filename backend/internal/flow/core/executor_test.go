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

package core

import (
	"testing"

	"github.com/stretchr/testify/suite"

	authncm "github.com/asgardeo/thunder/internal/authn/common"
	"github.com/asgardeo/thunder/internal/flow/common"
)

const (
	testExecutorName = "test-executor"
	testInputName    = "username"
	testInputValue   = "testuser"
)

type ExecutorTestSuite struct {
	suite.Suite
}

func TestExecutorTestSuite(t *testing.T) {
	suite.Run(t, new(ExecutorTestSuite))
}

func (s *ExecutorTestSuite) TestNewExecutor() {
	defaultInputs := []common.InputData{{Name: testInputName, Required: true}}
	prerequisites := []common.InputData{{Name: userAttributeUserID, Required: true}}

	exec := newExecutor(testExecutorName, common.ExecutorTypeAuthentication, defaultInputs, prerequisites)

	s.NotNil(exec)
	s.Equal(testExecutorName, exec.GetName())
	s.Equal(common.ExecutorTypeAuthentication, exec.GetType())
	s.Equal(defaultInputs, exec.GetDefaultExecutorInputs())
	s.Equal(prerequisites, exec.GetPrerequisites())
}

func (s *ExecutorTestSuite) TestGetName() {
	exec := newExecutor(testExecutorName, common.ExecutorTypeAuthentication, nil, nil)
	s.Equal(testExecutorName, exec.GetName())
}

func (s *ExecutorTestSuite) TestGetType() {
	exec := newExecutor(testExecutorName, common.ExecutorTypeAuthentication, nil, nil)
	s.Equal(common.ExecutorTypeAuthentication, exec.GetType())
}

func (s *ExecutorTestSuite) TestExecute() {
	exec := newExecutor(testExecutorName, common.ExecutorTypeAuthentication, nil, nil)
	ctx := &NodeContext{FlowID: "test-flow"}

	resp, err := exec.Execute(ctx)

	s.Nil(err)
	s.Nil(resp)
}

func (s *ExecutorTestSuite) TestGetDefaultExecutorInputs() {
	defaultInputs := []common.InputData{
		{Name: testInputName, Required: true},
		{Name: "password", Required: true},
	}
	exec := newExecutor(testExecutorName, common.ExecutorTypeAuthentication, defaultInputs, nil)

	result := exec.GetDefaultExecutorInputs()

	s.Equal(defaultInputs, result)
}

func (s *ExecutorTestSuite) TestGetPrerequisites() {
	prerequisites := []common.InputData{{Name: userAttributeUserID, Required: true}}
	exec := newExecutor(testExecutorName, common.ExecutorTypeAuthentication, nil, prerequisites)

	result := exec.GetPrerequisites()

	s.Equal(prerequisites, result)
}

func (s *ExecutorTestSuite) TestCheckInputData() {
	tests := []struct {
		name              string
		defaultInputs     []common.InputData
		userInputData     map[string]string
		runtimeData       map[string]string
		expectedRequires  bool
		expectedDataCount int
	}{
		{
			"No input data provided",
			[]common.InputData{{Name: testInputName, Required: true}},
			map[string]string{},
			map[string]string{},
			true,
			1,
		},
		{
			"All data in user input",
			[]common.InputData{{Name: testInputName, Required: true}},
			map[string]string{testInputName: testInputValue},
			map[string]string{},
			false,
			0,
		},
		{
			"Data in runtime data",
			[]common.InputData{{Name: testInputName, Required: true}},
			map[string]string{},
			map[string]string{testInputName: testInputValue},
			false,
			0,
		},
		{
			"Partial data in user input",
			[]common.InputData{
				{Name: testInputName, Required: true},
				{Name: "password", Required: true},
			},
			map[string]string{testInputName: testInputValue},
			map[string]string{},
			true,
			1,
		},
		{
			"Empty inputs and empty context",
			[]common.InputData{},
			map[string]string{},
			map[string]string{},
			true,
			0,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			exec := newExecutor(testExecutorName, common.ExecutorTypeAuthentication, tt.defaultInputs, nil)
			ctx := &NodeContext{
				FlowID:        "test-flow",
				UserInputData: tt.userInputData,
				RuntimeData:   tt.runtimeData,
			}
			execResp := &common.ExecutorResponse{}

			result := exec.CheckInputData(ctx, execResp)

			s.Equal(tt.expectedRequires, result)
			s.Len(execResp.RequiredData, tt.expectedDataCount)
		})
	}
}

func (s *ExecutorTestSuite) TestValidatePrerequisites() {
	tests := []struct {
		name               string
		prerequisites      []common.InputData
		authenticatedUser  authncm.AuthenticatedUser
		userInputData      map[string]string
		runtimeData        map[string]string
		expectedValid      bool
		expectedStatus     common.ExecutorStatus
		expectedFailReason string
	}{
		{
			"No prerequisites",
			[]common.InputData{},
			authncm.AuthenticatedUser{},
			map[string]string{},
			map[string]string{},
			true,
			"",
			"",
		},
		{
			"UserID prerequisite met via authenticated user",
			[]common.InputData{{Name: userAttributeUserID, Required: true}},
			authncm.AuthenticatedUser{UserID: "user-123"},
			map[string]string{},
			map[string]string{},
			true,
			"",
			"",
		},
		{
			"UserID prerequisite not met",
			[]common.InputData{{Name: userAttributeUserID, Required: true}},
			authncm.AuthenticatedUser{},
			map[string]string{},
			map[string]string{},
			false,
			common.ExecFailure,
			"Prerequisite not met: userID",
		},
		{
			"Other prerequisite met via user input",
			[]common.InputData{{Name: "email", Required: true}},
			authncm.AuthenticatedUser{},
			map[string]string{"email": "test@example.com"},
			map[string]string{},
			true,
			"",
			"",
		},
		{
			"Other prerequisite met via runtime data",
			[]common.InputData{{Name: "token", Required: true}},
			authncm.AuthenticatedUser{},
			map[string]string{},
			map[string]string{"token": "abc123"},
			true,
			"",
			"",
		},
		{
			"Prerequisite not met",
			[]common.InputData{{Name: "apiKey", Required: true}},
			authncm.AuthenticatedUser{},
			map[string]string{},
			map[string]string{},
			false,
			common.ExecFailure,
			"Prerequisite not met: apiKey",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			exec := newExecutor(testExecutorName, common.ExecutorTypeAuthentication, nil, tt.prerequisites)
			ctx := &NodeContext{
				FlowID:            "test-flow",
				AuthenticatedUser: tt.authenticatedUser,
				UserInputData:     tt.userInputData,
				RuntimeData:       tt.runtimeData,
			}
			execResp := &common.ExecutorResponse{}

			result := exec.ValidatePrerequisites(ctx, execResp)

			s.Equal(tt.expectedValid, result)
			s.Equal(tt.expectedStatus, execResp.Status)
			s.Equal(tt.expectedFailReason, execResp.FailureReason)
		})
	}
}

func (s *ExecutorTestSuite) TestGetUserIDFromContext() {
	tests := []struct {
		name              string
		authenticatedUser authncm.AuthenticatedUser
		runtimeData       map[string]string
		userInputData     map[string]string
		expectedUserID    string
	}{
		{
			"UserID from authenticated user",
			authncm.AuthenticatedUser{UserID: "user-123"},
			map[string]string{},
			map[string]string{},
			"user-123",
		},
		{
			"UserID from runtime data",
			authncm.AuthenticatedUser{},
			map[string]string{userAttributeUserID: "user-456"},
			map[string]string{},
			"user-456",
		},
		{
			"UserID from user input data",
			authncm.AuthenticatedUser{},
			map[string]string{},
			map[string]string{userAttributeUserID: "user-789"},
			"user-789",
		},
		{
			"Priority: authenticated user over runtime data",
			authncm.AuthenticatedUser{UserID: "user-auth"},
			map[string]string{userAttributeUserID: "user-runtime"},
			map[string]string{},
			"user-auth",
		},
		{
			"No userID available",
			authncm.AuthenticatedUser{},
			map[string]string{},
			map[string]string{},
			"",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			exec := newExecutor(testExecutorName, common.ExecutorTypeAuthentication, nil, nil)
			ctx := &NodeContext{
				AuthenticatedUser: tt.authenticatedUser,
				RuntimeData:       tt.runtimeData,
				UserInputData:     tt.userInputData,
			}

			result := exec.GetUserIDFromContext(ctx)

			s.Equal(tt.expectedUserID, result)
		})
	}
}

func (s *ExecutorTestSuite) TestGetRequiredData() {
	tests := []struct {
		name              string
		defaultInputs     []common.InputData
		nodeInputData     []common.InputData
		expectedDataCount int
		expectedContains  []string
	}{
		{
			"No node input, use default only",
			[]common.InputData{{Name: testInputName, Required: true}},
			[]common.InputData{},
			1,
			[]string{testInputName},
		},
		{
			"Node input provided, use node input only",
			[]common.InputData{{Name: testInputName, Required: true}},
			[]common.InputData{{Name: "email", Required: true}},
			1,
			[]string{"email"},
		},
		{
			"Node input with same name as default, use node input",
			[]common.InputData{{Name: testInputName, Required: true}},
			[]common.InputData{{Name: testInputName, Required: true}},
			1,
			[]string{testInputName},
		},
		{
			"No default inputs, use node input",
			[]common.InputData{},
			[]common.InputData{{Name: "custom", Required: false}},
			1,
			[]string{"custom"},
		},
		{
			"Multiple node inputs, use all node inputs",
			[]common.InputData{{Name: testInputName, Required: true}},
			[]common.InputData{{Name: "email", Required: true}, {Name: "phone", Required: true}},
			2,
			[]string{"email", "phone"},
		},
		{
			"No default and no node input, return empty",
			[]common.InputData{},
			[]common.InputData{},
			0,
			[]string{},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			exec := newExecutor(testExecutorName, common.ExecutorTypeAuthentication, tt.defaultInputs, nil)
			ctx := &NodeContext{FlowID: "test-flow", NodeInputData: tt.nodeInputData}

			result := exec.GetRequiredData(ctx)

			s.Len(result, tt.expectedDataCount)
			for _, name := range tt.expectedContains {
				found := false
				for _, input := range result {
					if input.Name == name {
						found = true
						break
					}
				}
				s.True(found, "Expected to find input name: %s", name)
			}
		})
	}
}
