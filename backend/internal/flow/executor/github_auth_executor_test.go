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

	"github.com/stretchr/testify/suite"

	authnoauth "github.com/asgardeo/thunder/internal/authn/oauth"
	"github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/tests/mocks/authn/githubmock"
	"github.com/asgardeo/thunder/tests/mocks/authn/oauthmock"
	"github.com/asgardeo/thunder/tests/mocks/flow/coremock"
	"github.com/asgardeo/thunder/tests/mocks/idp/idpmock"
	"github.com/asgardeo/thunder/tests/mocks/usertypemock"
)

type GithubAuthExecutorTestSuite struct {
	suite.Suite
	mockFlowFactory       *coremock.FlowFactoryInterfaceMock
	mockIDPService        *idpmock.IDPServiceInterfaceMock
	mockUserTypeService *usertypemock.UserTypeServiceInterfaceMock
	mockGithubService     *githubmock.GithubOAuthAuthnServiceInterfaceMock
	mockOAuthService      *oauthmock.OAuthAuthnCoreServiceInterfaceMock
}

func TestGithubAuthExecutorTestSuite(t *testing.T) {
	suite.Run(t, new(GithubAuthExecutorTestSuite))
}

func (suite *GithubAuthExecutorTestSuite) SetupTest() {
	suite.mockFlowFactory = coremock.NewFlowFactoryInterfaceMock(suite.T())
	suite.mockIDPService = idpmock.NewIDPServiceInterfaceMock(suite.T())
	suite.mockUserTypeService = usertypemock.NewUserTypeServiceInterfaceMock(suite.T())
	suite.mockGithubService = githubmock.NewGithubOAuthAuthnServiceInterfaceMock(suite.T())
	suite.mockOAuthService = oauthmock.NewOAuthAuthnCoreServiceInterfaceMock(suite.T())
}

func (suite *GithubAuthExecutorTestSuite) TestNewGithubOAuthExecutor_Success() {
	defaultInputs := []common.Input{
		{
			Identifier: "code",
			Type:       "string",
			Required:   true,
		},
	}
	baseExec := coremock.NewExecutorInterfaceMock(suite.T())
	suite.mockFlowFactory.On("CreateExecutor", ExecutorNameGitHubAuth,
		common.ExecutorTypeAuthentication, defaultInputs, []common.Input{}).
		Return(baseExec).Once()

	mockGithubSvc := &mockGithubServiceWithOAuth{
		GithubOAuthAuthnServiceInterfaceMock: suite.mockGithubService,
		oauthService:                         suite.mockOAuthService,
	}

	executor := newGithubOAuthExecutor(suite.mockFlowFactory, suite.mockIDPService,
		suite.mockUserTypeService, mockGithubSvc)

	suite.NotNil(executor)
	githubExec, ok := executor.(*githubOAuthExecutor)
	suite.True(ok)
	suite.NotNil(githubExec.oAuthExecutorInterface)
	suite.Equal(mockGithubSvc, githubExec.githubAuthService)
}

type mockGithubServiceWithOAuth struct {
	*githubmock.GithubOAuthAuthnServiceInterfaceMock
	oauthService authnoauth.OAuthAuthnCoreServiceInterface
}
