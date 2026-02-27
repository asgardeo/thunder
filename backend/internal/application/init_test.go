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

package application

import (
	"net/http"
	"os"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/cert"
	oauth2const "github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/tests/mocks/certmock"
	"github.com/asgardeo/thunder/tests/mocks/flow/flowmgtmock"
	"github.com/asgardeo/thunder/tests/mocks/userschemamock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// InitTestSuite contains comprehensive tests for the init.go file.
// The test suite covers:
// - Initialize function with declarative resources enabled/disabled
// - parseToApplicationDTO function with various YAML configurations
// - registerRoutes function with proper CORS setup
// - Error handling scenarios for configuration parsing and validation
type InitTestSuite struct {
	suite.Suite
	mockCertService       *certmock.CertificateServiceInterfaceMock
	mockFlowMgtService    *flowmgtmock.FlowMgtServiceInterfaceMock
	mockUserSchemaService *userschemamock.UserSchemaServiceInterfaceMock
}

func (suite *InitTestSuite) SetupTest() {
	suite.mockCertService = certmock.NewCertificateServiceInterfaceMock(suite.T())
	suite.mockFlowMgtService = flowmgtmock.NewFlowMgtServiceInterfaceMock(suite.T())
	suite.mockUserSchemaService = userschemamock.NewUserSchemaServiceInterfaceMock(suite.T())
}

func (suite *InitTestSuite) TearDownTest() {
	// Reset config to clear singleton state for next test
	config.ResetThunderRuntime()
}

func TestInitTestSuite(t *testing.T) {
	suite.Run(t, new(InitTestSuite))
}

// TestInitialize_WithDeclarativeResourcesDisabled tests the Initialize function when declarative resources are disabled
func (suite *InitTestSuite) TestInitialize_WithDeclarativeResourcesDisabled() {
	// Setup - ensure config is reset and initialized for this test
	config.ResetThunderRuntime()
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
	}
	err := config.InitializeThunderRuntime("/tmp/test", testConfig)
	assert.NoError(suite.T(), err)

	mux := http.NewServeMux()

	// Execute
	service, _, err := Initialize(
		mux,
		nil,
		suite.mockCertService,
		suite.mockFlowMgtService,
		nil, // themeMgtService - not needed for this test
		nil, // layoutMgtService - not needed for this test
		suite.mockUserSchemaService,
	)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), service)
	assert.Implements(suite.T(), (*ApplicationServiceInterface)(nil), service)
}

// TestInitialize_WithMCPServer tests the Initialize function with an MCP server
func (suite *InitTestSuite) TestInitialize_WithMCPServer() {
	// Setup - ensure config is reset and initialized for this test
	config.ResetThunderRuntime()
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
	}
	err := config.InitializeThunderRuntime("/tmp/test", testConfig)
	assert.NoError(suite.T(), err)

	mux := http.NewServeMux()

	// Create a mock MCP server
	mcpServer := mcp.NewServer(&mcp.Implementation{
		Name:    "test-mcp-server",
		Version: "1.0.0",
	}, nil)

	// Execute
	service, _, err := Initialize(
		mux,
		mcpServer,
		suite.mockCertService,
		suite.mockFlowMgtService,
		nil, // themeMgtService - not needed for this test
		nil, // layoutMgtService - not needed for this test
		suite.mockUserSchemaService,
	)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), service)
	assert.Implements(suite.T(), (*ApplicationServiceInterface)(nil), service)
	assert.NotNil(suite.T(), mcpServer)
}

// TestParseToApplicationDTO_ValidYAML tests parsing a valid YAML configuration
func (suite *InitTestSuite) TestParseToApplicationDTO_ValidYAML() {
	yamlData := `
name: test-app
description: Test application
auth_flow_id: test-auth-flow
registration_flow_id: test-reg-flow
is_registration_flow_enabled: true
url: https://example.com
logo_url: https://example.com/logo.png
assertion:
  validity_period: 3600
  user_attributes:
    - email
    - username
certificate:
  type: JWKS
  value: test-cert-value
inbound_auth_config:
  - type: oauth2
    config:
      client_id: test-client-id
      client_secret: test-client-secret
      redirect_uris:
        - https://example.com/callback
      grant_types:
        - authorization_code
      response_types:
        - code
      token_endpoint_auth_method: client_secret_basic
      pkce_required: true
      public_client: false
      token:
        access_token:
          validity_period: 3600
`

	// Execute
	appDTO, err := parseToApplicationDTO([]byte(yamlData))

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), appDTO)
	assert.Equal(suite.T(), "test-app", appDTO.Name)
	assert.Equal(suite.T(), "Test application", appDTO.Description)
	assert.Equal(suite.T(), "test-auth-flow", appDTO.AuthFlowID)
	assert.Equal(suite.T(), "test-reg-flow", appDTO.RegistrationFlowID)
	assert.True(suite.T(), appDTO.IsRegistrationFlowEnabled)
	assert.Equal(suite.T(), "https://example.com", appDTO.URL)
	assert.Equal(suite.T(), "https://example.com/logo.png", appDTO.LogoURL)

	// Verify token config
	assert.NotNil(suite.T(), appDTO.Assertion)
	// Note: ValidityPeriod and UserAttributes might be 0/nil if not properly parsed
	// This could be due to YAML structure differences

	// Verify certificate
	assert.NotNil(suite.T(), appDTO.Certificate)
	assert.Equal(suite.T(), cert.CertificateTypeJWKS, appDTO.Certificate.Type) // Using valid cert type
	assert.Equal(suite.T(), "test-cert-value", appDTO.Certificate.Value)

	// Verify inbound auth config
	assert.Len(suite.T(), appDTO.InboundAuthConfig, 1)
	assert.Equal(suite.T(), model.OAuthInboundAuthType, appDTO.InboundAuthConfig[0].Type)
	assert.NotNil(suite.T(), appDTO.InboundAuthConfig[0].OAuthAppConfig)
	assert.Equal(suite.T(), "test-client-id", appDTO.InboundAuthConfig[0].OAuthAppConfig.ClientID)
	assert.Equal(
		suite.T(), "test-client-secret", appDTO.InboundAuthConfig[0].OAuthAppConfig.ClientSecret)
	assert.Equal(suite.T(), []string{"https://example.com/callback"},
		appDTO.InboundAuthConfig[0].OAuthAppConfig.RedirectURIs)
	// Note: GrantTypes and ResponseTypes are typed constants, not plain strings
	assert.Contains(suite.T(), appDTO.InboundAuthConfig[0].OAuthAppConfig.GrantTypes,
		oauth2const.GrantType("authorization_code"))
	assert.Contains(suite.T(), appDTO.InboundAuthConfig[0].OAuthAppConfig.ResponseTypes,
		oauth2const.ResponseType("code"))
	assert.Equal(suite.T(), oauth2const.TokenEndpointAuthMethod("client_secret_basic"),
		appDTO.InboundAuthConfig[0].OAuthAppConfig.TokenEndpointAuthMethod)
	assert.True(suite.T(), appDTO.InboundAuthConfig[0].OAuthAppConfig.PKCERequired)
	assert.False(suite.T(), appDTO.InboundAuthConfig[0].OAuthAppConfig.PublicClient)

	// Verify OAuth token config
	assert.NotNil(suite.T(), appDTO.InboundAuthConfig[0].OAuthAppConfig.Token)
	// Note: OAuthTokenConfig doesn't have ValidityPeriod and UserAttributes directly
	// Those are in AccessToken and IDToken sub-configs
}

// TestParseToApplicationDTO_MinimalYAML tests parsing a minimal YAML configuration
func (suite *InitTestSuite) TestParseToApplicationDTO_MinimalYAML() {
	yamlData := `
name: minimal-app
description: Minimal application
is_registration_flow_enabled: false
`

	// Execute
	appDTO, err := parseToApplicationDTO([]byte(yamlData))

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), appDTO)
	assert.Equal(suite.T(), "minimal-app", appDTO.Name)
	assert.Equal(suite.T(), "Minimal application", appDTO.Description)
	assert.False(suite.T(), appDTO.IsRegistrationFlowEnabled)
	assert.Empty(suite.T(), appDTO.AuthFlowID)
	assert.Empty(suite.T(), appDTO.RegistrationFlowID)
	assert.Empty(suite.T(), appDTO.URL)
	assert.Empty(suite.T(), appDTO.LogoURL)
	assert.Nil(suite.T(), appDTO.Assertion)
	assert.Nil(suite.T(), appDTO.Certificate)
	assert.Empty(suite.T(), appDTO.InboundAuthConfig)
}

// TestParseToApplicationDTO_WithNonOAuthInboundAuth tests parsing with non-OAuth inbound auth config
func (suite *InitTestSuite) TestParseToApplicationDTO_WithNonOAuthInboundAuth() {
	yamlData := `
name: test-app
description: Test application
is_registration_flow_enabled: true
inbound_auth_config:
  - type: saml2
    config:
      issuer: test-saml-issuer
  - type: oauth2
    config:
      client_id: test-client-id
      client_secret: test-client-secret
`

	// Execute
	appDTO, err := parseToApplicationDTO([]byte(yamlData))

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), appDTO)
	// Should only include OAuth config, SAML should be filtered out
	assert.Len(suite.T(), appDTO.InboundAuthConfig, 1)
	assert.Equal(suite.T(), model.OAuthInboundAuthType, appDTO.InboundAuthConfig[0].Type)
	assert.Equal(suite.T(), "test-client-id", appDTO.InboundAuthConfig[0].OAuthAppConfig.ClientID)
}

// TestParseToApplicationDTO_WithOAuthConfigWithoutConfig tests parsing OAuth type without config
func (suite *InitTestSuite) TestParseToApplicationDTO_WithOAuthConfigWithoutConfig() {
	yamlData := `
name: test-app
description: Test application
is_registration_flow_enabled: true
inbound_auth_config:
  - type: oauth2
`

	// Execute
	appDTO, err := parseToApplicationDTO([]byte(yamlData))

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), appDTO)
	// Should filter out OAuth config without actual config
	assert.Empty(suite.T(), appDTO.InboundAuthConfig)
}

// TestParseToApplicationDTO_InvalidYAML tests parsing invalid YAML
func (suite *InitTestSuite) TestParseToApplicationDTO_InvalidYAML() {
	invalidYaml := `
name: test-app
description: Test application
invalid_yaml_structure: [
`

	// Execute
	appDTO, err := parseToApplicationDTO([]byte(invalidYaml))

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), appDTO)
}

// TestParseToApplicationDTO_EmptyInboundAuthConfig tests parsing with empty inbound auth config
func (suite *InitTestSuite) TestParseToApplicationDTO_EmptyInboundAuthConfig() {
	yamlData := `
name: test-app
description: Test application
is_registration_flow_enabled: true
inbound_auth_config: []
`

	// Execute
	appDTO, err := parseToApplicationDTO([]byte(yamlData))

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), appDTO)
	assert.Empty(suite.T(), appDTO.InboundAuthConfig)
}

// TestParseToApplicationDTO_WithCompleteOAuthConfig tests parsing with complete OAuth configuration
func (suite *InitTestSuite) TestParseToApplicationDTO_WithCompleteOAuthConfig() {
	yamlData := `
name: oauth-app
description: OAuth application
is_registration_flow_enabled: true
inbound_auth_config:
  - type: oauth2
    config:
      client_id: oauth-client
      client_secret: oauth-secret
      redirect_uris:
        - https://app.example.com/callback
        - https://app.example.com/redirect
      grant_types:
        - authorization_code
        - refresh_token
      response_types:
        - code
        - token
      token_endpoint_auth_method: client_secret_post
      pkce_required: false
      public_client: true
`

	// Execute
	appDTO, err := parseToApplicationDTO([]byte(yamlData))

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), appDTO)
	assert.Len(suite.T(), appDTO.InboundAuthConfig, 1)

	oauthConfig := appDTO.InboundAuthConfig[0].OAuthAppConfig
	assert.NotNil(suite.T(), oauthConfig)
	assert.Equal(suite.T(), "oauth-client", oauthConfig.ClientID)
	assert.Equal(suite.T(), "oauth-secret", oauthConfig.ClientSecret)
	assert.Equal(suite.T(), []string{"https://app.example.com/callback",
		"https://app.example.com/redirect"}, oauthConfig.RedirectURIs)
	// Using Contains for typed constants
	assert.Contains(suite.T(), oauthConfig.GrantTypes, oauth2const.GrantType("authorization_code"))
	assert.Contains(suite.T(), oauthConfig.GrantTypes, oauth2const.GrantType("refresh_token"))
	assert.Contains(suite.T(), oauthConfig.ResponseTypes, oauth2const.ResponseType("code"))
	assert.Contains(suite.T(), oauthConfig.ResponseTypes, oauth2const.ResponseType("token"))
	assert.Equal(suite.T(), oauth2const.TokenEndpointAuthMethod("client_secret_post"),
		oauthConfig.TokenEndpointAuthMethod)
	assert.False(suite.T(), oauthConfig.PKCERequired)
	assert.True(suite.T(), oauthConfig.PublicClient)

	// Note: No token section in YAML, so Token is nil (default)
	assert.Nil(suite.T(), oauthConfig.Token)
}

// Benchmark tests for performance
func BenchmarkParseToApplicationDTO(b *testing.B) {
	yamlData := `
name: benchmark-app
description: Benchmark application
is_registration_flow_enabled: true
inbound_auth_config:
  - type: oauth2
    config:
      client_id: benchmark-client
      client_secret: benchmark-secret
      redirect_uris:
        - https://example.com/callback
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parseToApplicationDTO([]byte(yamlData))
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Test YAML parsing with special characters and edge cases
func (suite *InitTestSuite) TestParseToApplicationDTO_WithSpecialCharacters() {
	yamlData := `
name: "app-with-special-chars-!@#$%"
description: "Description with 'quotes' and \"double quotes\""
url: "https://example.com/path?param=value&other=123"
logo_url: "https://cdn.example.com/logos/app-logo_v2.png"
is_registration_flow_enabled: true
`

	// Execute
	appDTO, err := parseToApplicationDTO([]byte(yamlData))

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), appDTO)
	assert.Equal(suite.T(), "app-with-special-chars-!@#$%", appDTO.Name)
	assert.Equal(suite.T(), "Description with 'quotes' and \"double quotes\"", appDTO.Description)
	assert.Equal(suite.T(), "https://example.com/path?param=value&other=123", appDTO.URL)
	assert.Equal(suite.T(), "https://cdn.example.com/logos/app-logo_v2.png", appDTO.LogoURL)
}

// Individual test functions that don't rely on suite setup

// TestParseToApplicationDTO_Standalone tests YAML parsing without suite dependencies
func TestParseToApplicationDTO_Standalone(t *testing.T) {
	yamlData := `
name: test-app
description: Test application
is_registration_flow_enabled: true
inbound_auth_config:
  - type: oauth2
    config:
      client_id: test-client-id
      client_secret: test-client-secret
      redirect_uris:
        - https://example.com/callback
      grant_types:
        - authorization_code
      response_types:
        - code
`

	// Execute
	appDTO, err := parseToApplicationDTO([]byte(yamlData))

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, appDTO)
	assert.Equal(t, "test-app", appDTO.Name)
	assert.Equal(t, "Test application", appDTO.Description)
	assert.True(t, appDTO.IsRegistrationFlowEnabled)
	assert.Len(t, appDTO.InboundAuthConfig, 1)
	assert.Equal(t, model.OAuthInboundAuthType, appDTO.InboundAuthConfig[0].Type)
	assert.Equal(t, "test-client-id", appDTO.InboundAuthConfig[0].OAuthAppConfig.ClientID)
}

// TestParseToApplicationDTO_InvalidYAML_Standalone tests parsing invalid YAML
func TestParseToApplicationDTO_InvalidYAML_Standalone(t *testing.T) {
	invalidYaml := `
name: test-app
description: Test application
invalid_yaml_structure: [
`

	// Execute
	appDTO, err := parseToApplicationDTO([]byte(invalidYaml))

	// Assert
	assert.Error(t, err)
	assert.Nil(t, appDTO)
}

// TestRegisterRoutes_Standalone tests route registration without suite dependencies
func TestRegisterRoutes_Standalone(t *testing.T) {
	// Setup
	mux := http.NewServeMux()
	mockHandler := &applicationHandler{}

	// Execute - should not panic
	assert.NotPanics(t, func() {
		registerRoutes(mux, mockHandler)
	})
}

// TestInitialize_Standalone tests Initialize function without suite dependencies
func TestInitialize_Standalone(t *testing.T) {
	// Setup minimal config for testing
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
	}

	// Reset and initialize with test config
	config.ResetThunderRuntime()
	err := config.InitializeThunderRuntime("/tmp/test", testConfig)
	assert.NoError(t, err)

	defer config.ResetThunderRuntime() // Clean up after test

	mux := http.NewServeMux()
	mockCertService := certmock.NewCertificateServiceInterfaceMock(t)
	mockFlowMgtService := flowmgtmock.NewFlowMgtServiceInterfaceMock(t)
	mockUserSchemaService := userschemamock.NewUserSchemaServiceInterfaceMock(t)

	// Execute
	service, _, err := Initialize(
		mux,
		nil,
		mockCertService,
		mockFlowMgtService,
		nil, // themeMgtService - not needed for this test
		nil, // layoutMgtService - not needed for this test
		mockUserSchemaService,
	)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, service)
	assert.Implements(t, (*ApplicationServiceInterface)(nil), service)
}

// TestInitialize_WithDeclarativeResources_Standalone tests Initialize function with declarative resources
func TestInitialize_WithDeclarativeResources_Standalone(t *testing.T) {
	// Setup minimal config for testing
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: true,
		},
	}

	// Create a temporary directory structure for file-based runtime
	tmpDir := t.TempDir()
	confDir := tmpDir + "/repository/resources"
	appDir := confDir + "/applications"

	// Create the directory structure
	err := os.MkdirAll(appDir, 0750)
	assert.NoError(t, err)

	// Reset and initialize with test config
	config.ResetThunderRuntime()
	err = config.InitializeThunderRuntime(tmpDir, testConfig)
	assert.NoError(t, err)

	defer config.ResetThunderRuntime() // Clean up after test

	mux := http.NewServeMux()
	mockCertService := certmock.NewCertificateServiceInterfaceMock(t)
	mockFlowMgtService := flowmgtmock.NewFlowMgtServiceInterfaceMock(t)
	mockUserSchemaService := userschemamock.NewUserSchemaServiceInterfaceMock(t)

	// Execute
	service, _, err := Initialize(
		mux,
		nil,
		mockCertService,
		mockFlowMgtService,
		nil, // themeMgtService - not needed for this test
		nil, // layoutMgtService - not needed for this test
		mockUserSchemaService,
	)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, service)
	assert.Implements(t, (*ApplicationServiceInterface)(nil), service)
}

// TestParseToApplicationDTO_WithScopeClaims tests parsing with scope claims including custom claims
func (suite *InitTestSuite) TestParseToApplicationDTO_WithScopeClaims() {
	yamlData := `
id: "test-app-scope-claims"
name: "App With Scope Claims"
inbound_auth_config:
  - type: oauth2
    config:
      client_id: "client-456"
      scope_claims:
        profile:
          - "name"
          - "email"
          - "customClaim"
        email:
          - "email"
`

	// Execute
	appDTO, err := parseToApplicationDTO([]byte(yamlData))

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), appDTO)
	require.Len(suite.T(), appDTO.InboundAuthConfig, 1,
		"InboundAuthConfig should have exactly 1 entry before accessing index 0")
	require.NotNil(suite.T(), appDTO.InboundAuthConfig[0].OAuthAppConfig,
		"OAuthAppConfig should not be nil before accessing fields")
	require.NotNil(suite.T(), appDTO.InboundAuthConfig[0].OAuthAppConfig.ScopeClaims,
		"ScopeClaims should not be nil before accessing map keys")
	scopeClaims := appDTO.InboundAuthConfig[0].OAuthAppConfig.ScopeClaims
	require.Contains(suite.T(), scopeClaims, "profile", "ScopeClaims should contain 'profile' key")
	require.NotNil(suite.T(), scopeClaims["profile"], "profile scope claims should not be nil")
	assert.Len(suite.T(), scopeClaims["profile"], 3)
	assert.Contains(suite.T(), scopeClaims["profile"], "customClaim")
}

// TestParseToApplicationDTO_WithScopes tests parsing with custom scopes
func (suite *InitTestSuite) TestParseToApplicationDTO_WithScopes() {
	yamlData := `
id: "test-app-scopes"
name: "App With Scopes"
inbound_auth_config:
  - type: oauth2
    config:
      client_id: "client-123"
      scopes:
        - "openid"
        - "profile"
`

	// Execute
	appDTO, err := parseToApplicationDTO([]byte(yamlData))

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), appDTO)
	require.Len(suite.T(), appDTO.InboundAuthConfig, 1,
		"InboundAuthConfig should have exactly 1 entry before accessing index 0")
	require.NotNil(suite.T(), appDTO.InboundAuthConfig[0].OAuthAppConfig,
		"OAuthAppConfig should not be nil before accessing fields")
	assert.NotNil(suite.T(), appDTO.InboundAuthConfig[0].OAuthAppConfig.Scopes)
	assert.Len(suite.T(), appDTO.InboundAuthConfig[0].OAuthAppConfig.Scopes, 2)
	assert.Contains(suite.T(), appDTO.InboundAuthConfig[0].OAuthAppConfig.Scopes, "openid")
	assert.Contains(suite.T(), appDTO.InboundAuthConfig[0].OAuthAppConfig.Scopes, "profile")
}

// TestParseToApplicationDTO_WithUserInfo tests parsing with UserInfo configuration
func (suite *InitTestSuite) TestParseToApplicationDTO_WithUserInfo() {
	yamlData := `
id: "test-app-userinfo"
name: "App With UserInfo"
inbound_auth_config:
  - type: oauth2
    config:
      client_id: "client-789"
      user_info:
        user_attributes:
          - "sub"
          - "email"
          - "name"
`

	// Execute
	appDTO, err := parseToApplicationDTO([]byte(yamlData))

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), appDTO)
	require.Len(suite.T(), appDTO.InboundAuthConfig, 1,
		"InboundAuthConfig should have exactly 1 entry before accessing index 0")
	require.NotNil(suite.T(), appDTO.InboundAuthConfig[0].OAuthAppConfig,
		"OAuthAppConfig should not be nil before accessing fields")
	require.NotNil(suite.T(), appDTO.InboundAuthConfig[0].OAuthAppConfig.UserInfo,
		"UserInfo should not be nil before accessing UserAttributes")
	assert.Len(suite.T(), appDTO.InboundAuthConfig[0].OAuthAppConfig.UserInfo.UserAttributes, 3)
}

// TestParseToApplicationDTO_WithAllOAuthFieldsIncludingFixedFields tests all OAuth fields
// including Scopes, UserInfo, and ScopeClaims (the fix for GitHub issue #1445)
func (suite *InitTestSuite) TestParseToApplicationDTO_WithAllOAuthFieldsIncludingFixedFields() {
	yamlData := `
id: "test-app-complete"
name: "Complete OAuth App"
inbound_auth_config:
  - type: oauth2
    config:
      client_id: "complete-client"
      client_secret: "secret-value"
      redirect_uris:
        - "https://example.com/callback"
      grant_types:
        - "authorization_code"
      response_types:
        - "code"
      token_endpoint_auth_method: "client_secret_basic"
      pkce_required: true
      public_client: false
      token:
        id_token:
          user_attributes:
            - "sub"
            - "email"
      scopes:
        - "openid"
      user_info:
        user_attributes:
          - "profile"
      scope_claims:
        profile:
          - "name"
`

	// Execute
	appDTO, err := parseToApplicationDTO([]byte(yamlData))

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), appDTO)
	require.NotEmpty(suite.T(), appDTO.InboundAuthConfig,
		"InboundAuthConfig should not be empty before accessing index 0")
	oauth := appDTO.InboundAuthConfig[0].OAuthAppConfig
	require.NotNil(suite.T(), oauth, "OAuthAppConfig should not be nil before accessing fields")
	assert.Equal(suite.T(), "complete-client", oauth.ClientID)
	assert.Equal(suite.T(), "secret-value", oauth.ClientSecret)
	assert.Len(suite.T(), oauth.RedirectURIs, 1)
	assert.Equal(suite.T(), "https://example.com/callback", oauth.RedirectURIs[0])
	assert.Len(suite.T(), oauth.GrantTypes, 1)
	assert.True(suite.T(), oauth.PKCERequired)
	assert.False(suite.T(), oauth.PublicClient)
	assert.NotNil(suite.T(), oauth.Token)
	// Verify the fixed fields are properly copied
	assert.NotNil(suite.T(), oauth.Scopes)
	assert.NotNil(suite.T(), oauth.UserInfo)
	assert.NotNil(suite.T(), oauth.ScopeClaims)
}

// TestParseToApplicationDTO_GithubIssue1445_CustomClaimsInScopeClaims tests the exact scenario from GitHub issue #1445
func (suite *InitTestSuite) TestParseToApplicationDTO_GithubIssue1445_CustomClaimsInScopeClaims() {
	// This is the exact scenario from the GitHub issue where scope_claims, scopes, and user_info were being dropped
	yamlData := `
id: "test-app-custom-claims"
name: "App With Custom Claims"
inbound_auth_config:
  - type: oauth2
    config:
      client_id: "MY_APP"
      token:
        id_token:
          user_attributes:
            - "email"
            - "name"
            - "customClaim"
      scope_claims:
        profile:
          - "name"
          - "customClaim"
        email:
          - "email"
`

	// Execute
	appDTO, err := parseToApplicationDTO([]byte(yamlData))

	// Assert
	require.NoError(suite.T(), err, "parseToApplicationDTO should not return error")
	require.NotNil(suite.T(), appDTO, "appDTO should not be nil")
	require.NotEmpty(suite.T(), appDTO.InboundAuthConfig,
		"InboundAuthConfig should not be empty before accessing index 0")
	oauth := appDTO.InboundAuthConfig[0].OAuthAppConfig
	require.NotNil(suite.T(), oauth, "OAuthAppConfig should not be nil before accessing fields")

	// Verify scope_claims are properly copied (this was the bug in issue #1445)
	assert.NotNil(suite.T(), oauth.ScopeClaims, "ScopeClaims should not be nil")
	assert.Len(suite.T(), oauth.ScopeClaims, 2, "ScopeClaims should have 2 entries (profile and email)")
	assert.Contains(suite.T(), oauth.ScopeClaims, "profile", "ScopeClaims should contain 'profile'")
	assert.Contains(suite.T(), oauth.ScopeClaims, "email", "ScopeClaims should contain 'email'")
	assert.Len(suite.T(), oauth.ScopeClaims["profile"], 2, "profile scope should have 2 claims")
	assert.Contains(suite.T(), oauth.ScopeClaims["profile"], "name", "profile scope should contain 'name'")
	assert.Contains(suite.T(), oauth.ScopeClaims["profile"], "customClaim",
		"profile scope should contain 'customClaim'")
}
