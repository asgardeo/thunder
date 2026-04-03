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

package authn

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/tests/mocks/authnprovidermock"
	"github.com/asgardeo/thunder/tests/mocks/consentmock"
	"github.com/asgardeo/thunder/tests/mocks/emailmock"
	"github.com/asgardeo/thunder/tests/mocks/idp/idpmock"
	"github.com/asgardeo/thunder/tests/mocks/jose/jwtmock"
	"github.com/asgardeo/thunder/tests/mocks/notification/notificationmock"
	"github.com/asgardeo/thunder/tests/mocks/templatemock"
	"github.com/asgardeo/thunder/tests/mocks/usermock"
	"github.com/asgardeo/thunder/tests/mocks/userprovidermock"
)

type AuthenticationInitTestSuite struct {
	suite.Suite
}

var (
	initRuntimeMutex sync.Mutex
)

func TestAuthenticationInitTestSuite(t *testing.T) {
	suite.Run(t, new(AuthenticationInitTestSuite))
}

func initializeTestRuntime(root string) error {
	testConfig := &config.Config{
		Server: config.ServerConfig{
			Hostname: "localhost",
			Port:     8090,
		},
		JWT: config.JWTConfig{
			Issuer: "test-issuer",
		},
	}
	return config.InitializeThunderRuntime(root, testConfig)
}

func (suite *AuthenticationInitTestSuite) SetupSuite() {
	initRuntimeMutex.Lock()
	config.ResetThunderRuntime()
	suite.Require().NoError(initializeTestRuntime(suite.T().TempDir()))
}

func (suite *AuthenticationInitTestSuite) TearDownSuite() {
	config.ResetThunderRuntime()
	initRuntimeMutex.Unlock()
}

// TestCreateAuthServiceRegistryWithEmailClient tests magic link service registration when email client is available
func (suite *AuthenticationInitTestSuite) TestCreateAuthServiceRegistryWithEmailClient() {
	mockIDPService := idpmock.NewIDPServiceInterfaceMock(suite.T())
	mockJWTService := jwtmock.NewJWTServiceInterfaceMock(suite.T())
	mockUserService := usermock.NewUserServiceInterfaceMock(suite.T())
	mockUserProvider := userprovidermock.NewUserProviderInterfaceMock(suite.T())
	mockOTPService := notificationmock.NewOTPServiceInterfaceMock(suite.T())
	mockAuthnProvider := authnprovidermock.NewAuthnProviderInterfaceMock(suite.T())
	mockConsentService := consentmock.NewConsentServiceInterfaceMock(suite.T())
	mockEmailClient := emailmock.NewEmailClientInterfaceMock(suite.T())
	mockTemplateService := templatemock.NewTemplateServiceInterfaceMock(suite.T())

	registry := createAuthServiceRegistry(
		mockIDPService,
		mockJWTService,
		mockUserService,
		mockUserProvider,
		mockOTPService,
		mockAuthnProvider,
		mockConsentService,
		mockEmailClient,
		mockTemplateService,
	)

	suite.NotNil(registry)
	suite.NotNil(registry.MagicLinkAuthnService,
		"Magic link service should be initialized when email client is provided")
}

// TestCreateAuthServiceRegistryWithoutEmailClient tests that magic link service is not registered
// when email client is nil
func (suite *AuthenticationInitTestSuite) TestCreateAuthServiceRegistryWithoutEmailClient() {
	mockIDPService := idpmock.NewIDPServiceInterfaceMock(suite.T())
	mockJWTService := jwtmock.NewJWTServiceInterfaceMock(suite.T())
	mockUserService := usermock.NewUserServiceInterfaceMock(suite.T())
	mockUserProvider := userprovidermock.NewUserProviderInterfaceMock(suite.T())
	mockOTPService := notificationmock.NewOTPServiceInterfaceMock(suite.T())
	mockAuthnProvider := authnprovidermock.NewAuthnProviderInterfaceMock(suite.T())
	mockConsentService := consentmock.NewConsentServiceInterfaceMock(suite.T())
	mockTemplateService := templatemock.NewTemplateServiceInterfaceMock(suite.T())

	registry := createAuthServiceRegistry(
		mockIDPService,
		mockJWTService,
		mockUserService,
		mockUserProvider,
		mockOTPService,
		mockAuthnProvider,
		mockConsentService,
		nil, // nil email client
		mockTemplateService,
	)

	suite.NotNil(registry)
	suite.Nil(registry.MagicLinkAuthnService, "Magic link service should be nil when email client is not provided")
}
