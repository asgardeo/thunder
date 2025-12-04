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

package notification

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/notification/common"
	"github.com/asgardeo/thunder/internal/system/cmodels"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/tests/mocks/jwtmock"
)

type InitTestSuite struct {
	suite.Suite
	mockJWTService *jwtmock.JWTServiceInterfaceMock
	mux            *http.ServeMux
}

func TestInitTestSuite(t *testing.T) {
	suite.Run(t, new(InitTestSuite))
}

func (suite *InitTestSuite) SetupSuite() {
	// Get the current working directory.
	cwd, err := os.Getwd()
	if err != nil {
		suite.T().Fatalf("Failed to get working directory: %v", err)
	}
	suite.T().Logf("Current working directory: %s", cwd)
	cryptoFile := filepath.Join(cwd, "..", "..", "tests", "resources", "testKey")

	if _, err := os.Stat(cryptoFile); os.IsNotExist(err) {
		suite.T().Fatalf("Crypto file not found at expected path: %s", cryptoFile)
	}

	testConfig := &config.Config{
		JWT: config.JWTConfig{
			Issuer:         "test-issuer",
			ValidityPeriod: 3600,
		},
		Security: config.SecurityConfig{
			CryptoFile: cryptoFile,
		},
	}
	err = config.InitializeThunderRuntime("", testConfig)
	if err != nil {
		suite.T().Fatalf("Failed to initialize ThunderRuntime: %v", err)
	}
}

func (suite *InitTestSuite) SetupTest() {
	suite.mockJWTService = jwtmock.NewJWTServiceInterfaceMock(suite.T())
	suite.mux = http.NewServeMux()
}

func (suite *InitTestSuite) TestInitialize() {
	mgtService, otpService := Initialize(suite.mux, suite.mockJWTService)

	suite.NotNil(mgtService)
	suite.NotNil(otpService)
	suite.Implements((*NotificationSenderMgtSvcInterface)(nil), mgtService)
	suite.Implements((*OTPServiceInterface)(nil), otpService)
}

func (suite *InitTestSuite) TestRegisterRoutes_ListEndpoint() {
	Initialize(suite.mux, suite.mockJWTService)

	req := httptest.NewRequest(http.MethodGet, "/notification-senders/message", nil)
	w := httptest.NewRecorder()

	suite.mux.ServeHTTP(w, req)

	suite.NotEqual(http.StatusNotFound, w.Code)
}

func (suite *InitTestSuite) TestRegisterRoutes_CreateEndpoint() {
	Initialize(suite.mux, suite.mockJWTService)

	req := httptest.NewRequest(http.MethodPost, "/notification-senders/message", nil)
	w := httptest.NewRecorder()

	suite.mux.ServeHTTP(w, req)

	suite.NotEqual(http.StatusNotFound, w.Code)
}

func (suite *InitTestSuite) TestRegisterRoutes_GetByIDEndpoint() {
	Initialize(suite.mux, suite.mockJWTService)

	req := httptest.NewRequest(http.MethodGet, "/notification-senders/message/test-id", nil)
	w := httptest.NewRecorder()

	suite.mux.ServeHTTP(w, req)

	suite.NotEqual(http.StatusNotFound, w.Code)
}

func (suite *InitTestSuite) TestRegisterRoutes_UpdateEndpoint() {
	Initialize(suite.mux, suite.mockJWTService)

	req := httptest.NewRequest(http.MethodPut, "/notification-senders/message/test-id", nil)
	w := httptest.NewRecorder()

	suite.mux.ServeHTTP(w, req)

	suite.NotEqual(http.StatusNotFound, w.Code)
}

func (suite *InitTestSuite) TestRegisterRoutes_DeleteEndpoint() {
	Initialize(suite.mux, suite.mockJWTService)

	req := httptest.NewRequest(http.MethodDelete, "/notification-senders/message/test-id", nil)
	w := httptest.NewRecorder()

	suite.mux.ServeHTTP(w, req)

	suite.NotEqual(http.StatusNotFound, w.Code)
}

func (suite *InitTestSuite) TestRegisterRoutes_SendOTPEndpoint() {
	Initialize(suite.mux, suite.mockJWTService)

	req := httptest.NewRequest(http.MethodPost, "/notification-senders/otp/send", nil)
	w := httptest.NewRecorder()

	suite.mux.ServeHTTP(w, req)

	suite.NotEqual(http.StatusNotFound, w.Code)
}

func (suite *InitTestSuite) TestRegisterRoutes_VerifyOTPEndpoint() {
	Initialize(suite.mux, suite.mockJWTService)

	req := httptest.NewRequest(http.MethodPost, "/notification-senders/otp/verify", nil)
	w := httptest.NewRecorder()

	suite.mux.ServeHTTP(w, req)

	suite.NotEqual(http.StatusNotFound, w.Code)
}

func (suite *InitTestSuite) TestRegisterRoutes_CORSPreflight() {
	Initialize(suite.mux, suite.mockJWTService)

	req := httptest.NewRequest(http.MethodOptions, "/notification-senders/message", nil)
	w := httptest.NewRecorder()

	suite.mux.ServeHTTP(w, req)

	suite.NotEqual(http.StatusNotFound, w.Code)
}

// TestParseToNotificationSenderDTO_ValidYAML tests parsing a valid YAML configuration.
func (suite *InitTestSuite) TestParseToNotificationSenderDTO_ValidYAML() {
	yamlData := `
id: "twilio-sender-001"
name: "Twilio SMS Sender"
description: "Production Twilio SMS sender"
provider: "twilio"
properties:
  - name: "account_sid"
    value: "{{.TWILIO_ACCOUNT_SID}}"
    is_secret: false
  - name: "auth_token"
    value: "{{.TWILIO_AUTH_TOKEN}}"
    is_secret: true
  - name: "sender_id"
    value: "{{.TWILIO_FROM_NUMBER}}"
    is_secret: false
`

	sender, err := parseToNotificationSenderDTO([]byte(yamlData))

	suite.NoError(err)
	suite.NotNil(sender)
	suite.Equal("twilio-sender-001", sender.ID)
	suite.Equal("Twilio SMS Sender", sender.Name)
	suite.Equal("Production Twilio SMS sender", sender.Description)
	suite.Equal("twilio", string(sender.Provider))
	suite.Len(sender.Properties, 3)
}

// TestParseToNotificationSenderDTO_InvalidYAML tests parsing invalid YAML.
func (suite *InitTestSuite) TestParseToNotificationSenderDTO_InvalidYAML() {
	yamlData := `
invalid yaml content
  - this is not valid
`

	sender, err := parseToNotificationSenderDTO([]byte(yamlData))

	suite.Error(err)
	suite.Nil(sender)
}

// TestParseToNotificationSenderDTO_MinimalYAML tests parsing minimal YAML configuration.
func (suite *InitTestSuite) TestParseToNotificationSenderDTO_MinimalYAML() {
	yamlData := `
id: "minimal-sender"
name: "Minimal Sender"
provider: "custom"
properties:
  - name: "url"
    value: "https://custom.example.com/sms"
`

	sender, err := parseToNotificationSenderDTO([]byte(yamlData))

	suite.NoError(err)
	suite.NotNil(sender)
	suite.Equal("minimal-sender", sender.ID)
	suite.Equal("Minimal Sender", sender.Name)
	suite.Equal("", sender.Description)
	suite.Equal("custom", string(sender.Provider))
	suite.Len(sender.Properties, 1)
}

// TestParseProviderType_ValidProviders tests parsing valid provider types.
func (suite *InitTestSuite) TestParseProviderType_ValidProviders() {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Twilio lowercase", "twilio", "twilio"},
		{"Twilio uppercase", "TWILIO", "twilio"},
		{"Vonage lowercase", "vonage", "vonage"},
		{"Custom lowercase", "custom", "custom"},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			provider, err := parseProviderType(tt.input)
			suite.NoError(err)
			suite.Equal(tt.expected, string(provider))
		})
	}
}

// TestParseProviderType_InvalidProvider tests parsing invalid provider type.
func (suite *InitTestSuite) TestParseProviderType_InvalidProvider() {
	provider, err := parseProviderType("invalid_provider")

	suite.Error(err)
	suite.Equal("", string(provider))
	suite.Contains(err.Error(), "unsupported provider type")
}

// TestValidateNotificationSenderForInit_ValidSender tests validation with valid sender.
func (suite *InitTestSuite) TestValidateNotificationSenderForInit_ValidSender() {
	properties := []cmodels.Property{}
	prop1, _ := cmodels.NewProperty("account_sid", "test_sid", false)
	properties = append(properties, *prop1)
	prop2, _ := cmodels.NewProperty("auth_token", "test_token", true)
	properties = append(properties, *prop2)
	prop3, _ := cmodels.NewProperty("sender_id", "test_sender", false)
	properties = append(properties, *prop3)

	sender := &common.NotificationSenderDTO{
		ID:          "test-001",
		Name:        "Test Sender",
		Description: "Test notification sender",
		Type:        common.NotificationSenderTypeMessage,
		Provider:    common.MessageProviderTypeTwilio,
		Properties:  properties,
	}

	err := validateNotificationSenderForInit(sender)

	suite.Nil(err)
}

// TestValidateNotificationSenderForInit_NilSender tests validation with nil sender.
func (suite *InitTestSuite) TestValidateNotificationSenderForInit_NilSender() {
	err := validateNotificationSenderForInit(nil)

	suite.NotNil(err)
	suite.Equal("MNS-1003", err.Code)
}

// TestValidateNotificationSenderForInit_EmptyName tests validation with empty name.
func (suite *InitTestSuite) TestValidateNotificationSenderForInit_EmptyName() {
	sender := &common.NotificationSenderDTO{
		Name:     "",
		Provider: common.MessageProviderTypeTwilio,
	}

	err := validateNotificationSenderForInit(sender)

	suite.NotNil(err)
	suite.Equal("MNS-1003", err.Code)
}

// TestValidateNotificationSenderForInit_InvalidProvider tests validation with invalid provider.
func (suite *InitTestSuite) TestValidateNotificationSenderForInit_InvalidProvider() {
	sender := &common.NotificationSenderDTO{
		Name:     "Test Sender",
		Provider: "invalid_provider",
	}

	err := validateNotificationSenderForInit(sender)

	suite.NotNil(err)
	suite.Equal("MNS-1004", err.Code)
}

// TestValidateNotificationSenderForInit_MissingProperties tests validation with missing properties.
func (suite *InitTestSuite) TestValidateNotificationSenderForInit_MissingProperties() {
	sender := &common.NotificationSenderDTO{
		Name:       "Test Sender",
		Provider:   common.MessageProviderTypeTwilio,
		Properties: []cmodels.Property{},
	}

	err := validateNotificationSenderForInit(sender)

	suite.NotNil(err)
	suite.Contains(err.ErrorDescription, "properties cannot be empty")
}

// TestValidateNotificationSenderForInit_MissingRequiredProperty tests validation with missing required property.
func (suite *InitTestSuite) TestValidateNotificationSenderForInit_MissingRequiredProperty() {
	properties := []cmodels.Property{}
	prop1, _ := cmodels.NewProperty("account_sid", "test_sid", false)
	properties = append(properties, *prop1)
	// Missing auth_token and sender_id

	sender := &common.NotificationSenderDTO{
		Name:       "Test Sender",
		Provider:   common.MessageProviderTypeTwilio,
		Properties: properties,
	}

	err := validateNotificationSenderForInit(sender)

	suite.NotNil(err)
	suite.Contains(err.ErrorDescription, "required property")
	suite.Contains(err.ErrorDescription, "missing")
}
