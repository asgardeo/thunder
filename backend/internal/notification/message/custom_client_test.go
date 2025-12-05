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

package message

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/notification/common"
	"github.com/asgardeo/thunder/internal/system/cmodels"
	"github.com/asgardeo/thunder/internal/system/config"
)

type CustomClientTestSuite struct {
	suite.Suite
}

func TestCustomClientTestSuite(t *testing.T) {
	suite.Run(t, new(CustomClientTestSuite))
}

func (suite *CustomClientTestSuite) SetupSuite() {
	testConfig := &config.Config{
		Crypto: config.CryptoConfig{
			Encrypt: config.EncryptConfig{
				Key: "b735c757d03c6496575f4c1eb5ba708ad3d06a635566d0fe0440802cf750b3f8",
			},
		},
	}
	err := config.InitializeThunderRuntime("", testConfig)
	if err != nil {
		suite.T().Fatalf("Failed to initialize ThunderRuntime: %v", err)
	}
}

func (suite *CustomClientTestSuite) getValidCustomSenderJSON() common.NotificationSenderDTO {
	return common.NotificationSenderDTO{
		Name:     "Test Custom",
		Provider: common.MessageProviderTypeCustom,
		Properties: []cmodels.Property{
			createProperty("url", "https://api.example.com/sms", false),
			createProperty("http_method", "POST", false),
			createProperty("content_type", "JSON", false),
			createProperty("http_headers", "Authorization:Bearer token,X-Api-Key:key123", false),
		},
	}
}

func (suite *CustomClientTestSuite) getValidCustomSenderFORM() common.NotificationSenderDTO {
	return common.NotificationSenderDTO{
		Name:     "Test Custom Form",
		Provider: common.MessageProviderTypeCustom,
		Properties: []cmodels.Property{
			createProperty("url", "https://api.example.com/sms", false),
			createProperty("http_method", "POST", false),
			createProperty("content_type", "FORM", false),
		},
	}
}

func (suite *CustomClientTestSuite) TestNewCustomClient_Success() {
	sender := suite.getValidCustomSenderJSON()

	client, err := NewCustomClient(sender)

	suite.NoError(err)
	suite.NotNil(client)
	suite.Equal("Test Custom", client.GetName())
}

func (suite *CustomClientTestSuite) TestGetName() {
	sender := suite.getValidCustomSenderJSON()
	client, _ := NewCustomClient(sender)

	name := client.GetName()

	suite.Equal("Test Custom", name)
}

func (suite *CustomClientTestSuite) TestSendSMS_JSON_Success() {
	sender := suite.getValidCustomSenderJSON()
	client, _ := NewCustomClient(sender)

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		suite.Equal(http.MethodPost, r.Method)
		suite.Equal("application/json", r.Header.Get("Content-Type"))
		suite.Equal("Bearer token", r.Header.Get("Authorization"))
		suite.Equal("key123", r.Header.Get("X-Api-Key"))

		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"success":true}`)); err != nil {
			suite.T().Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	// Replace the URL with test server URL
	customClient := client.(*CustomClient)
	customClient.url = server.URL

	smsData := common.SMSData{
		To:   "+15559876543",
		Body: `{"message":"Test message"}`,
	}

	err := client.SendSMS(smsData)

	suite.NoError(err)
}

func (suite *CustomClientTestSuite) TestSendSMS_FORM_Success() {
	sender := suite.getValidCustomSenderFORM()
	client, _ := NewCustomClient(sender)

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		suite.Equal(http.MethodPost, r.Method)
		suite.Equal("application/x-www-form-urlencoded", r.Header.Get("Content-Type"))

		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`OK`)); err != nil {
			suite.T().Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	// Replace the URL with test server URL
	customClient := client.(*CustomClient)
	customClient.url = server.URL

	smsData := common.SMSData{
		To:   "+15559876543",
		Body: "to=+15559876543\nmessage=Test message",
	}

	err := client.SendSMS(smsData)

	suite.NoError(err)
}

func (suite *CustomClientTestSuite) TestSendSMS_Error() {
	sender := suite.getValidCustomSenderJSON()
	client, _ := NewCustomClient(sender)

	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		if _, err := w.Write([]byte(`{"error":"Invalid request"}`)); err != nil {
			suite.T().Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	// Replace the URL with test server URL
	customClient := client.(*CustomClient)
	customClient.url = server.URL

	smsData := common.SMSData{
		To:   "+15559876543",
		Body: `{"message":"Test"}`,
	}

	err := client.SendSMS(smsData)

	suite.Error(err)
	suite.Contains(err.Error(), "status code: 400")
}

func (suite *CustomClientTestSuite) TestSendSMS_NetworkError() {
	sender := suite.getValidCustomSenderJSON()
	client, _ := NewCustomClient(sender)

	// Use an invalid URL to force a network error
	customClient := client.(*CustomClient)
	customClient.url = "http://invalid-custom-url.local:99999"

	smsData := common.SMSData{
		To:   "+15559876543",
		Body: `{"message":"Test"}`,
	}

	err := client.SendSMS(smsData)

	suite.Error(err)
}

func (suite *CustomClientTestSuite) TestSendSMS_UnsupportedContentType() {
	sender := common.NotificationSenderDTO{
		Name:     "Test Custom",
		Provider: common.MessageProviderTypeCustom,
		Properties: []cmodels.Property{
			createProperty("url", "https://api.example.com/sms", false),
			createProperty("http_method", "POST", false),
			createProperty("content_type", "XML", false),
		},
	}
	client, _ := NewCustomClient(sender)

	smsData := common.SMSData{
		To:   "+15559876543",
		Body: `<message>Test</message>`,
	}

	err := client.SendSMS(smsData)

	suite.Error(err)
	suite.Contains(err.Error(), "unsupported content type")
}

func (suite *CustomClientTestSuite) TestGetHeadersFromString_Success() {
	sender := suite.getValidCustomSenderJSON()
	client, _ := NewCustomClient(sender)
	customClient := client.(*CustomClient)

	headers, err := customClient.getHeadersFromString("Authorization:Bearer token,X-Api-Key:key123")

	suite.NoError(err)
	suite.Equal(2, len(headers))
	suite.Equal("Bearer token", headers["Authorization"])
	suite.Equal("key123", headers["X-Api-Key"])
}

func (suite *CustomClientTestSuite) TestGetHeadersFromString_InvalidFormat() {
	sender := suite.getValidCustomSenderJSON()
	client, _ := NewCustomClient(sender)
	customClient := client.(*CustomClient)

	headers, err := customClient.getHeadersFromString("InvalidHeader")

	suite.Error(err)
	suite.Nil(headers)
	suite.Contains(err.Error(), "invalid HTTP header format")
}

func (suite *CustomClientTestSuite) TestNewCustomClient_WithUnknownProperty() {
	sender := suite.getValidCustomSenderJSON()
	sender.Properties = append(sender.Properties, createProperty("unknown_prop", "value", false))

	client, err := NewCustomClient(sender)

	// Should succeed and just log a warning for unknown property
	suite.NoError(err)
	suite.NotNil(client)
}

func (suite *CustomClientTestSuite) TestNewCustomClient_InvalidHeaders() {
	sender := common.NotificationSenderDTO{
		Name:     "Test Custom",
		Provider: common.MessageProviderTypeCustom,
		Properties: []cmodels.Property{
			createProperty("url", "https://api.example.com/sms", false),
			createProperty("http_method", "POST", false),
			createProperty("content_type", "JSON", false),
			createProperty("http_headers", "InvalidHeaderFormat", false),
		},
	}

	client, err := NewCustomClient(sender)

	suite.Error(err)
	suite.Nil(client)
	suite.Contains(err.Error(), "invalid HTTP header format")
}
