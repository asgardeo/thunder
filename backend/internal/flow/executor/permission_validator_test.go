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
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/internal/flow/core"
	"github.com/asgardeo/thunder/internal/system/security"
	"github.com/asgardeo/thunder/tests/mocks/flow/coremock"
)

type PermissionValidatorTestSuite struct {
	suite.Suite
	mockFlowFactory *coremock.FlowFactoryInterfaceMock
	executor        *permissionValidator
}

func (suite *PermissionValidatorTestSuite) SetupTest() {
	suite.mockFlowFactory = coremock.NewFlowFactoryInterfaceMock(suite.T())
	mockBaseExecutor := coremock.NewExecutorInterfaceMock(suite.T())

	suite.mockFlowFactory.On("CreateExecutor",
		ExecutorNamePermissionValidator,
		common.ExecutorTypeUtility,
		[]common.Input{},
		[]common.Input{}).Return(mockBaseExecutor)

	suite.executor = newPermissionValidator(suite.mockFlowFactory)
}

func (suite *PermissionValidatorTestSuite) TestExecute_DefaultScopeCheck_Success() {
	httpCtx := context.Background()
	authCtx := security.NewSecurityContextForTest(
		"user1", "ou1", "app1", "token",
		map[string]interface{}{"scope": "system other"},
	)
	httpCtx = security.WithSecurityContextTest(httpCtx, authCtx)

	ctx := &core.NodeContext{
		FlowID:      "test-flow",
		HTTPContext: httpCtx,
	}

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), common.ExecComplete, resp.Status)
}

func (suite *PermissionValidatorTestSuite) TestExecute_DefaultScopeCheck_Failure() {
	httpCtx := context.Background()
	authCtx := security.NewSecurityContextForTest(
		"user1", "ou1", "app1", "token",
		map[string]interface{}{"scope": "other"},
	)
	httpCtx = security.WithSecurityContextTest(httpCtx, authCtx)

	ctx := &core.NodeContext{
		FlowID:      "test-flow",
		HTTPContext: httpCtx,
	}

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), common.ExecFailure, resp.Status)
	assert.Equal(suite.T(), "Insufficient permissions", resp.FailureReason)
}

func (suite *PermissionValidatorTestSuite) TestExecute_CustomScopeCheck_Success() {
	httpCtx := context.Background()
	authCtx := security.NewSecurityContextForTest(
		"user1", "ou1", "app1", "token",
		map[string]interface{}{"scope": "invite:create"},
	)
	httpCtx = security.WithSecurityContextTest(httpCtx, authCtx)

	ctx := &core.NodeContext{
		FlowID:      "test-flow",
		HTTPContext: httpCtx,
		NodeProperties: map[string]interface{}{
			"requiredScope": "invite:create",
		},
	}

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), common.ExecComplete, resp.Status)
}

func (suite *PermissionValidatorTestSuite) TestExecute_AuthorizedPermissionsCheck_Success() {
	httpCtx := context.Background()
	authCtx := security.NewSecurityContextForTest(
		"user1", "ou1", "app1", "token",
		map[string]interface{}{"authorized_permissions": "read write admin"},
	)
	httpCtx = security.WithSecurityContextTest(httpCtx, authCtx)

	ctx := &core.NodeContext{
		FlowID:      "test-flow",
		HTTPContext: httpCtx,
		NodeProperties: map[string]interface{}{
			"requiredScope": "admin",
		},
	}

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), common.ExecComplete, resp.Status)
}

func (suite *PermissionValidatorTestSuite) TestExecute_NoHTTPContext() {
	ctx := &core.NodeContext{
		FlowID:      "test-flow",
		HTTPContext: nil,
	}

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), common.ExecFailure, resp.Status)
	assert.Equal(suite.T(), "Insufficient permissions", resp.FailureReason)
}

func (suite *PermissionValidatorTestSuite) TestExecute_EmptyScopes() {
	httpCtx := context.Background()
	authCtx := security.NewSecurityContextForTest(
		"user1", "ou1", "app1", "token",
		map[string]interface{}{"scope": ""},
	)
	httpCtx = security.WithSecurityContextTest(httpCtx, authCtx)

	ctx := &core.NodeContext{
		FlowID:      "test-flow",
		HTTPContext: httpCtx,
	}

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), common.ExecFailure, resp.Status)
	assert.Equal(suite.T(), "Insufficient permissions", resp.FailureReason)
}

func (suite *PermissionValidatorTestSuite) TestExecute_NoScopesInContext() {
	httpCtx := context.Background()
	authCtx := security.NewSecurityContextForTest(
		"user1", "ou1", "app1", "token",
		map[string]interface{}{}, // No scope or authorized_permissions
	)
	httpCtx = security.WithSecurityContextTest(httpCtx, authCtx)

	ctx := &core.NodeContext{
		FlowID:      "test-flow",
		HTTPContext: httpCtx,
	}

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), common.ExecFailure, resp.Status)
	assert.Equal(suite.T(), "Insufficient permissions", resp.FailureReason)
}

func (suite *PermissionValidatorTestSuite) TestExecute_ScopesWithUnexpectedType() {
	httpCtx := context.Background()
	authCtx := security.NewSecurityContextForTest(
		"user1", "ou1", "app1", "token",
		map[string]interface{}{"scope": 123}, // Invalid type (int instead of string)
	)
	httpCtx = security.WithSecurityContextTest(httpCtx, authCtx)

	ctx := &core.NodeContext{
		FlowID:      "test-flow",
		HTTPContext: httpCtx,
	}

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), common.ExecFailure, resp.Status)
	assert.Equal(suite.T(), "Insufficient permissions", resp.FailureReason)
}

func TestPermissionValidatorSuite(t *testing.T) {
	suite.Run(t, new(PermissionValidatorTestSuite))
}
