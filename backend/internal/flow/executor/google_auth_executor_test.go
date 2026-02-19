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

	authnoidc "github.com/asgardeo/thunder/internal/authn/oidc"
	"github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/tests/mocks/authn/googlemock"
	"github.com/asgardeo/thunder/tests/mocks/authn/oidcmock"
	"github.com/asgardeo/thunder/tests/mocks/flow/coremock"
	"github.com/asgardeo/thunder/tests/mocks/idp/idpmock"
	"github.com/asgardeo/thunder/tests/mocks/usertypemock"
)

type GoogleAuthExecutorTestSuite struct {
	suite.Suite
	mockFlowFactory       *coremock.FlowFactoryInterfaceMock
	mockIDPService        *idpmock.IDPServiceInterfaceMock
	mockUserTypeService *usertypemock.UserTypeServiceInterfaceMock
	mockGoogleService     *googlemock.GoogleOIDCAuthnServiceInterfaceMock
	mockOIDCService       *oidcmock.OIDCAuthnCoreServiceInterfaceMock
}

func TestGoogleAuthExecutorTestSuite(t *testing.T) {
	suite.Run(t, new(GoogleAuthExecutorTestSuite))
}

func (suite *GoogleAuthExecutorTestSuite) SetupTest() {
	suite.mockFlowFactory = coremock.NewFlowFactoryInterfaceMock(suite.T())
	suite.mockIDPService = idpmock.NewIDPServiceInterfaceMock(suite.T())
	suite.mockUserTypeService = usertypemock.NewUserTypeServiceInterfaceMock(suite.T())
	suite.mockGoogleService = googlemock.NewGoogleOIDCAuthnServiceInterfaceMock(suite.T())
	suite.mockOIDCService = oidcmock.NewOIDCAuthnCoreServiceInterfaceMock(suite.T())
}

func (suite *GoogleAuthExecutorTestSuite) TestNewGoogleOIDCAuthExecutor_Success() {
	defaultInputs := []common.Input{
		{
			Identifier: "code",
			Type:       "string",
			Required:   true,
		},
		{
			Identifier: "nonce",
			Type:       "string",
			Required:   false,
		},
	}
	baseExec := coremock.NewExecutorInterfaceMock(suite.T())
	suite.mockFlowFactory.On("CreateExecutor", ExecutorNameGoogleAuth,
		common.ExecutorTypeAuthentication, defaultInputs, []common.Input{}).
		Return(baseExec).Once()

	mockGoogleSvc := &mockGoogleServiceWithOIDC{
		GoogleOIDCAuthnServiceInterfaceMock: suite.mockGoogleService,
		oidcService:                         suite.mockOIDCService,
	}

	executor := newGoogleOIDCAuthExecutor(suite.mockFlowFactory, suite.mockIDPService,
		suite.mockUserTypeService, mockGoogleSvc)

	suite.NotNil(executor)
	googleExec, ok := executor.(*googleOIDCAuthExecutor)
	suite.True(ok)
	suite.NotNil(googleExec.oidcAuthExecutorInterface)
	suite.Equal(mockGoogleSvc, googleExec.googleAuthService)
}

type mockGoogleServiceWithOIDC struct {
	*googlemock.GoogleOIDCAuthnServiceInterfaceMock
	oidcService authnoidc.OIDCAuthnCoreServiceInterface
}
