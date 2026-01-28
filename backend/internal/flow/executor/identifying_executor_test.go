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
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/user"
	"github.com/asgardeo/thunder/tests/mocks/flow/coremock"
	"github.com/asgardeo/thunder/tests/mocks/usermock"
)

type IdentifyingExecutorTestSuite struct {
	suite.Suite
	mockUserService *usermock.UserServiceInterfaceMock
	mockFlowFactory *coremock.FlowFactoryInterfaceMock
	executor        *identifyingExecutor
}

func TestIdentifyingExecutorSuite(t *testing.T) {
	suite.Run(t, new(IdentifyingExecutorTestSuite))
}

func (suite *IdentifyingExecutorTestSuite) SetupTest() {
	suite.mockUserService = usermock.NewUserServiceInterfaceMock(suite.T())
	suite.mockFlowFactory = coremock.NewFlowFactoryInterfaceMock(suite.T())

	mockExec := createMockExecutor(suite.T(), ExecutorNameIdentifying, common.ExecutorTypeUtility)
	suite.mockFlowFactory.On("CreateExecutor", ExecutorNameIdentifying, common.ExecutorTypeUtility,
		[]common.Input{}, []common.Input{}).Return(mockExec)

	suite.executor = newIdentifyingExecutor(ExecutorNameIdentifying, []common.Input{},
		[]common.Input{}, suite.mockFlowFactory, suite.mockUserService)
}

func (suite *IdentifyingExecutorTestSuite) TestNewIdentifyingExecutor() {
	assert.NotNil(suite.T(), suite.executor)
	assert.NotNil(suite.T(), suite.executor.userService)
}

func (suite *IdentifyingExecutorTestSuite) TestIdentifyUser_Success() {
	filters := map[string]interface{}{"username": "testuser"}
	execResp := &common.ExecutorResponse{
		RuntimeData: make(map[string]string),
	}
	userID := "user-123"

	suite.mockUserService.On("IdentifyUser", mock.Anything, filters).Return(&userID, nil)

	result, err := suite.executor.IdentifyUser(filters, execResp)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "user-123", *result)
	suite.mockUserService.AssertExpectations(suite.T())
}

func (suite *IdentifyingExecutorTestSuite) TestIdentifyUser_UserNotFound() {
	filters := map[string]interface{}{"username": "nonexistent"}
	execResp := &common.ExecutorResponse{
		RuntimeData: make(map[string]string),
	}

	suite.mockUserService.On("IdentifyUser", mock.Anything, filters).
		Return(nil, &user.ErrorUserNotFound)

	result, err := suite.executor.IdentifyUser(filters, execResp)

	assert.NoError(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), common.ExecFailure, execResp.Status)
	assert.Equal(suite.T(), failureReasonUserNotFound, execResp.FailureReason)
	suite.mockUserService.AssertExpectations(suite.T())
}

func (suite *IdentifyingExecutorTestSuite) TestIdentifyUser_ServiceError() {
	filters := map[string]interface{}{"username": "testuser"}
	execResp := &common.ExecutorResponse{
		RuntimeData: make(map[string]string),
	}

	suite.mockUserService.On("IdentifyUser", mock.Anything, filters).
		Return(nil, &serviceerror.ServiceError{Error: "service error"})

	result, err := suite.executor.IdentifyUser(filters, execResp)

	assert.NoError(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), common.ExecFailure, execResp.Status)
	assert.Contains(suite.T(), execResp.FailureReason, "Failed to identify user")
	suite.mockUserService.AssertExpectations(suite.T())
}

func (suite *IdentifyingExecutorTestSuite) TestIdentifyUser_EmptyUserID() {
	filters := map[string]interface{}{"username": "testuser"}
	execResp := &common.ExecutorResponse{
		RuntimeData: make(map[string]string),
	}
	emptyID := ""

	suite.mockUserService.On("IdentifyUser", mock.Anything, filters).Return(&emptyID, nil)

	result, err := suite.executor.IdentifyUser(filters, execResp)

	assert.NoError(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), common.ExecFailure, execResp.Status)
	assert.Equal(suite.T(), failureReasonUserNotFound, execResp.FailureReason)
	suite.mockUserService.AssertExpectations(suite.T())
}

func (suite *IdentifyingExecutorTestSuite) TestIdentifyUser_FilterNonSearchableAttributes() {
	filters := map[string]interface{}{
		"username": "testuser",
		"password": "secret123",
		"code":     "auth-code",
		"nonce":    "nonce-value",
		"otp":      "123456",
	}
	execResp := &common.ExecutorResponse{
		RuntimeData: make(map[string]string),
	}
	userID := "user-123"

	suite.mockUserService.On("IdentifyUser", mock.Anything, map[string]interface{}{
		"username": "testuser",
	}).Return(&userID, nil)

	result, err := suite.executor.IdentifyUser(filters, execResp)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "user-123", *result)
	suite.mockUserService.AssertExpectations(suite.T())
}

func (suite *IdentifyingExecutorTestSuite) TestIdentifyUser_WithEmail() {
	filters := map[string]interface{}{"email": "test@example.com"}
	execResp := &common.ExecutorResponse{
		RuntimeData: make(map[string]string),
	}
	userID := "user-456"

	suite.mockUserService.On("IdentifyUser", mock.Anything, filters).Return(&userID, nil)

	result, err := suite.executor.IdentifyUser(filters, execResp)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "user-456", *result)
	suite.mockUserService.AssertExpectations(suite.T())
}

func (suite *IdentifyingExecutorTestSuite) TestIdentifyUser_WithMobileNumber() {
	filters := map[string]interface{}{"mobileNumber": "+1234567890"}
	execResp := &common.ExecutorResponse{
		RuntimeData: make(map[string]string),
	}
	userID := "user-789"

	suite.mockUserService.On("IdentifyUser", mock.Anything, filters).Return(&userID, nil)

	result, err := suite.executor.IdentifyUser(filters, execResp)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "user-789", *result)
	suite.mockUserService.AssertExpectations(suite.T())
}
