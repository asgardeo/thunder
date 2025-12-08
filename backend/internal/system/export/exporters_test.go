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

package export

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	appmodel "github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/idp"
	"github.com/asgardeo/thunder/internal/notification/common"
	"github.com/asgardeo/thunder/internal/system/cmodels"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/userschema"
	"github.com/asgardeo/thunder/tests/mocks/applicationmock"
	"github.com/asgardeo/thunder/tests/mocks/idp/idpmock"
	"github.com/asgardeo/thunder/tests/mocks/notification/notificationmock"
	"github.com/asgardeo/thunder/tests/mocks/userschemamock"
)

// ApplicationExporterTestSuite tests the ApplicationExporter.
type ApplicationExporterTestSuite struct {
	suite.Suite
	mockService *applicationmock.ApplicationServiceInterfaceMock
	exporter    *ApplicationExporter
	logger      *log.Logger
}

func TestApplicationExporterTestSuite(t *testing.T) {
	suite.Run(t, new(ApplicationExporterTestSuite))
}

func (s *ApplicationExporterTestSuite) SetupTest() {
	s.mockService = applicationmock.NewApplicationServiceInterfaceMock(s.T())
	s.exporter = NewApplicationExporter(s.mockService)
	s.logger = log.GetLogger()
}

func (s *ApplicationExporterTestSuite) TestNewApplicationExporter() {
	assert.NotNil(s.T(), s.exporter)
	assert.Equal(s.T(), s.mockService, s.exporter.service)
}

func (s *ApplicationExporterTestSuite) TestGetResourceType() {
	assert.Equal(s.T(), resourceTypeApplication, s.exporter.GetResourceType())
}

func (s *ApplicationExporterTestSuite) TestGetParameterizerType() {
	assert.Equal(s.T(), paramTypApplication, s.exporter.GetParameterizerType())
}

func (s *ApplicationExporterTestSuite) TestGetAllResourceIDs_Success() {
	expectedApps := &appmodel.ApplicationListResponse{
		Applications: []appmodel.BasicApplicationResponse{
			{ID: "app1", Name: "App 1"},
			{ID: "app2", Name: "App 2"},
		},
	}

	s.mockService.EXPECT().GetApplicationList().Return(expectedApps, nil)

	ids, err := s.exporter.GetAllResourceIDs()

	assert.Nil(s.T(), err)
	assert.Len(s.T(), ids, 2)
	assert.Equal(s.T(), "app1", ids[0])
	assert.Equal(s.T(), "app2", ids[1])
}

func (s *ApplicationExporterTestSuite) TestGetAllResourceIDs_Error() {
	expectedError := &serviceerror.ServiceError{
		Code:  "ERR_CODE",
		Error: "test error",
	}

	s.mockService.EXPECT().GetApplicationList().Return(nil, expectedError)

	ids, err := s.exporter.GetAllResourceIDs()

	assert.Nil(s.T(), ids)
	assert.Equal(s.T(), expectedError, err)
}

func (s *ApplicationExporterTestSuite) TestGetAllResourceIDs_EmptyList() {
	expectedApps := &appmodel.ApplicationListResponse{
		Applications: []appmodel.BasicApplicationResponse{},
	}

	s.mockService.EXPECT().GetApplicationList().Return(expectedApps, nil)

	ids, err := s.exporter.GetAllResourceIDs()

	assert.Nil(s.T(), err)
	assert.Len(s.T(), ids, 0)
}

func (s *ApplicationExporterTestSuite) TestGetResourceByID_Success() {
	expectedApp := &appmodel.Application{
		ID:   "app1",
		Name: "Test App",
	}

	s.mockService.EXPECT().GetApplication("app1").Return(expectedApp, nil)

	resource, name, err := s.exporter.GetResourceByID("app1")

	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "Test App", name)
	assert.Equal(s.T(), expectedApp, resource)
}

func (s *ApplicationExporterTestSuite) TestGetResourceByID_Error() {
	expectedError := &serviceerror.ServiceError{
		Code:  "ERR_CODE",
		Error: "test error",
	}

	s.mockService.EXPECT().GetApplication("app1").Return(nil, expectedError)

	resource, name, err := s.exporter.GetResourceByID("app1")

	assert.Nil(s.T(), resource)
	assert.Empty(s.T(), name)
	assert.Equal(s.T(), expectedError, err)
}

func (s *ApplicationExporterTestSuite) TestValidateResource_Success() {
	app := &appmodel.Application{
		ID:   "app1",
		Name: "Valid App",
	}

	name, err := s.exporter.ValidateResource(app, "app1", s.logger)

	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "Valid App", name)
}

func (s *ApplicationExporterTestSuite) TestValidateResource_EmptyName() {
	app := &appmodel.Application{
		ID:   "app1",
		Name: "",
	}

	name, err := s.exporter.ValidateResource(app, "app1", s.logger)

	assert.NotNil(s.T(), err)
	assert.Empty(s.T(), name)
	assert.Equal(s.T(), resourceTypeApplication, err.ResourceType)
	assert.Equal(s.T(), "app1", err.ResourceID)
	assert.Equal(s.T(), "APP_VALIDATION_ERROR", err.Code)
	assert.Contains(s.T(), err.Error, "name is empty")
}

func (s *ApplicationExporterTestSuite) TestValidateResource_InvalidType() {
	invalidResource := "not an application"

	name, err := s.exporter.ValidateResource(invalidResource, "app1", s.logger)

	assert.NotNil(s.T(), err)
	assert.Empty(s.T(), name)
	assert.Equal(s.T(), resourceTypeApplication, err.ResourceType)
	assert.Equal(s.T(), "app1", err.ResourceID)
	assert.Equal(s.T(), "INVALID_TYPE", err.Code)
	assert.Equal(s.T(), "Invalid resource type", err.Error)
}

// IDPExporterTestSuite tests the IDPExporter.
type IDPExporterTestSuite struct {
	suite.Suite
	mockService *idpmock.IDPServiceInterfaceMock
	exporter    *IDPExporter
	logger      *log.Logger
}

func TestIDPExporterTestSuite(t *testing.T) {
	suite.Run(t, new(IDPExporterTestSuite))
}

func (s *IDPExporterTestSuite) SetupTest() {
	s.mockService = idpmock.NewIDPServiceInterfaceMock(s.T())
	s.exporter = NewIDPExporter(s.mockService)
	s.logger = log.GetLogger()
}

func (s *IDPExporterTestSuite) TestNewIDPExporter() {
	assert.NotNil(s.T(), s.exporter)
	assert.Equal(s.T(), s.mockService, s.exporter.service)
}

func (s *IDPExporterTestSuite) TestGetResourceType() {
	assert.Equal(s.T(), resourceTypeIdentityProvider, s.exporter.GetResourceType())
}

func (s *IDPExporterTestSuite) TestGetParameterizerType() {
	assert.Equal(s.T(), paramTypIdentityProvider, s.exporter.GetParameterizerType())
}

func (s *IDPExporterTestSuite) TestGetAllResourceIDs_Success() {
	expectedIDPs := []idp.BasicIDPDTO{
		{ID: "idp1", Name: "IDP 1"},
		{ID: "idp2", Name: "IDP 2"},
	}

	s.mockService.EXPECT().GetIdentityProviderList().Return(expectedIDPs, nil)

	ids, err := s.exporter.GetAllResourceIDs()

	assert.Nil(s.T(), err)
	assert.Len(s.T(), ids, 2)
	assert.Equal(s.T(), "idp1", ids[0])
	assert.Equal(s.T(), "idp2", ids[1])
}

func (s *IDPExporterTestSuite) TestGetAllResourceIDs_Error() {
	expectedError := &serviceerror.ServiceError{
		Code:  "ERR_CODE",
		Error: "test error",
	}

	s.mockService.EXPECT().GetIdentityProviderList().Return(nil, expectedError)

	ids, err := s.exporter.GetAllResourceIDs()

	assert.Nil(s.T(), ids)
	assert.Equal(s.T(), expectedError, err)
}

func (s *IDPExporterTestSuite) TestGetAllResourceIDs_EmptyList() {
	expectedIDPs := []idp.BasicIDPDTO{}

	s.mockService.EXPECT().GetIdentityProviderList().Return(expectedIDPs, nil)

	ids, err := s.exporter.GetAllResourceIDs()

	assert.Nil(s.T(), err)
	assert.Len(s.T(), ids, 0)
}

func (s *IDPExporterTestSuite) TestGetResourceByID_Success() {
	prop, _ := cmodels.NewProperty("prop1", "value1", false)
	expectedIDP := &idp.IDPDTO{
		ID:         "idp1",
		Name:       "Test IDP",
		Properties: []cmodels.Property{*prop},
	}

	s.mockService.EXPECT().GetIdentityProvider("idp1").Return(expectedIDP, nil)

	resource, name, err := s.exporter.GetResourceByID("idp1")

	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "Test IDP", name)
	assert.Equal(s.T(), expectedIDP, resource)
}

func (s *IDPExporterTestSuite) TestGetResourceByID_Error() {
	expectedError := &serviceerror.ServiceError{
		Code:  "ERR_CODE",
		Error: "test error",
	}

	s.mockService.EXPECT().GetIdentityProvider("idp1").Return(nil, expectedError)

	resource, name, err := s.exporter.GetResourceByID("idp1")

	assert.Nil(s.T(), resource)
	assert.Empty(s.T(), name)
	assert.Equal(s.T(), expectedError, err)
}

func (s *IDPExporterTestSuite) TestValidateResource_Success() {
	prop, _ := cmodels.NewProperty("prop1", "value1", false)
	idpDTO := &idp.IDPDTO{
		ID:         "idp1",
		Name:       "Valid IDP",
		Properties: []cmodels.Property{*prop},
	}

	name, err := s.exporter.ValidateResource(idpDTO, "idp1", s.logger)

	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "Valid IDP", name)
}

func (s *IDPExporterTestSuite) TestValidateResource_EmptyName() {
	idpDTO := &idp.IDPDTO{
		ID:   "idp1",
		Name: "",
	}

	name, err := s.exporter.ValidateResource(idpDTO, "idp1", s.logger)

	assert.NotNil(s.T(), err)
	assert.Empty(s.T(), name)
	assert.Equal(s.T(), resourceTypeIdentityProvider, err.ResourceType)
	assert.Equal(s.T(), "idp1", err.ResourceID)
	assert.Equal(s.T(), "IDP_VALIDATION_ERROR", err.Code)
	assert.Contains(s.T(), err.Error, "name is empty")
}

func (s *IDPExporterTestSuite) TestValidateResource_NoProperties() {
	// This should succeed but log a warning
	idpDTO := &idp.IDPDTO{
		ID:         "idp1",
		Name:       "IDP No Props",
		Properties: []cmodels.Property{},
	}

	name, err := s.exporter.ValidateResource(idpDTO, "idp1", s.logger)

	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "IDP No Props", name)
}

func (s *IDPExporterTestSuite) TestValidateResource_InvalidType() {
	invalidResource := "not an idp"

	name, err := s.exporter.ValidateResource(invalidResource, "idp1", s.logger)

	assert.NotNil(s.T(), err)
	assert.Empty(s.T(), name)
	assert.Equal(s.T(), resourceTypeIdentityProvider, err.ResourceType)
	assert.Equal(s.T(), "idp1", err.ResourceID)
	assert.Equal(s.T(), "INVALID_TYPE", err.Code)
	assert.Equal(s.T(), "Invalid resource type", err.Error)
}

// NotificationSenderExporterTestSuite tests the NotificationSenderExporter.
type NotificationSenderExporterTestSuite struct {
	suite.Suite
	mockService *notificationmock.NotificationSenderMgtSvcInterfaceMock
	exporter    *NotificationSenderExporter
	logger      *log.Logger
}

func TestNotificationSenderExporterTestSuite(t *testing.T) {
	suite.Run(t, new(NotificationSenderExporterTestSuite))
}

func (s *NotificationSenderExporterTestSuite) SetupTest() {
	s.mockService = notificationmock.NewNotificationSenderMgtSvcInterfaceMock(s.T())
	s.exporter = NewNotificationSenderExporter(s.mockService)
	s.logger = log.GetLogger()
}

func (s *NotificationSenderExporterTestSuite) TestNewNotificationSenderExporter() {
	assert.NotNil(s.T(), s.exporter)
	assert.Equal(s.T(), s.mockService, s.exporter.service)
}

func (s *NotificationSenderExporterTestSuite) TestGetResourceType() {
	assert.Equal(s.T(), resourceTypeNotificationSender, s.exporter.GetResourceType())
}

func (s *NotificationSenderExporterTestSuite) TestGetParameterizerType() {
	assert.Equal(s.T(), paramTypNotificationSender, s.exporter.GetParameterizerType())
}

func (s *NotificationSenderExporterTestSuite) TestGetAllResourceIDs_Success() {
	expectedSenders := []common.NotificationSenderDTO{
		{ID: "sender1", Name: "Sender 1"},
		{ID: "sender2", Name: "Sender 2"},
	}

	s.mockService.EXPECT().ListSenders().Return(expectedSenders, nil)

	ids, err := s.exporter.GetAllResourceIDs()

	assert.Nil(s.T(), err)
	assert.Len(s.T(), ids, 2)
	assert.Equal(s.T(), "sender1", ids[0])
	assert.Equal(s.T(), "sender2", ids[1])
}

func (s *NotificationSenderExporterTestSuite) TestGetAllResourceIDs_Error() {
	expectedError := &serviceerror.ServiceError{
		Code:  "ERR_CODE",
		Error: "test error",
	}

	s.mockService.EXPECT().ListSenders().Return(nil, expectedError)

	ids, err := s.exporter.GetAllResourceIDs()

	assert.Nil(s.T(), ids)
	assert.Equal(s.T(), expectedError, err)
}

func (s *NotificationSenderExporterTestSuite) TestGetAllResourceIDs_EmptyList() {
	expectedSenders := []common.NotificationSenderDTO{}

	s.mockService.EXPECT().ListSenders().Return(expectedSenders, nil)

	ids, err := s.exporter.GetAllResourceIDs()

	assert.Nil(s.T(), err)
	assert.Len(s.T(), ids, 0)
}

func (s *NotificationSenderExporterTestSuite) TestGetResourceByID_Success() {
	prop, _ := cmodels.NewProperty("prop1", "value1", false)
	expectedSender := &common.NotificationSenderDTO{
		ID:         "sender1",
		Name:       "Test Sender",
		Properties: []cmodels.Property{*prop},
	}

	s.mockService.EXPECT().GetSender("sender1").Return(expectedSender, nil)

	resource, name, err := s.exporter.GetResourceByID("sender1")

	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "Test Sender", name)
	assert.Equal(s.T(), expectedSender, resource)
}

func (s *NotificationSenderExporterTestSuite) TestGetResourceByID_Error() {
	expectedError := &serviceerror.ServiceError{
		Code:  "ERR_CODE",
		Error: "test error",
	}

	s.mockService.EXPECT().GetSender("sender1").Return(nil, expectedError)

	resource, name, err := s.exporter.GetResourceByID("sender1")

	assert.Nil(s.T(), resource)
	assert.Empty(s.T(), name)
	assert.Equal(s.T(), expectedError, err)
}

func (s *NotificationSenderExporterTestSuite) TestValidateResource_Success() {
	prop, _ := cmodels.NewProperty("prop1", "value1", false)
	sender := &common.NotificationSenderDTO{
		ID:         "sender1",
		Name:       "Valid Sender",
		Properties: []cmodels.Property{*prop},
	}

	name, err := s.exporter.ValidateResource(sender, "sender1", s.logger)

	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "Valid Sender", name)
}

func (s *NotificationSenderExporterTestSuite) TestValidateResource_EmptyName() {
	sender := &common.NotificationSenderDTO{
		ID:   "sender1",
		Name: "",
	}

	name, err := s.exporter.ValidateResource(sender, "sender1", s.logger)

	assert.NotNil(s.T(), err)
	assert.Empty(s.T(), name)
	assert.Equal(s.T(), resourceTypeNotificationSender, err.ResourceType)
	assert.Equal(s.T(), "sender1", err.ResourceID)
	assert.Equal(s.T(), "SENDER_VALIDATION_ERROR", err.Code)
	assert.Contains(s.T(), err.Error, "name is empty")
}

func (s *NotificationSenderExporterTestSuite) TestValidateResource_NoProperties() {
	// This should succeed but log a warning
	sender := &common.NotificationSenderDTO{
		ID:         "sender1",
		Name:       "Sender No Props",
		Properties: []cmodels.Property{},
	}

	name, err := s.exporter.ValidateResource(sender, "sender1", s.logger)

	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "Sender No Props", name)
}

func (s *NotificationSenderExporterTestSuite) TestValidateResource_InvalidType() {
	invalidResource := "not a sender"

	name, err := s.exporter.ValidateResource(invalidResource, "sender1", s.logger)

	assert.NotNil(s.T(), err)
	assert.Empty(s.T(), name)
	assert.Equal(s.T(), resourceTypeNotificationSender, err.ResourceType)
	assert.Equal(s.T(), "sender1", err.ResourceID)
	assert.Equal(s.T(), "INVALID_TYPE", err.Code)
	assert.Equal(s.T(), "Invalid resource type", err.Error)
}

// UserSchemaExporterTestSuite tests the UserSchemaExporter.
type UserSchemaExporterTestSuite struct {
	suite.Suite
	mockService *userschemamock.UserSchemaServiceInterfaceMock
	exporter    *UserSchemaExporter
	logger      *log.Logger
}

func TestUserSchemaExporterTestSuite(t *testing.T) {
	suite.Run(t, new(UserSchemaExporterTestSuite))
}

func (s *UserSchemaExporterTestSuite) SetupTest() {
	s.mockService = userschemamock.NewUserSchemaServiceInterfaceMock(s.T())
	s.exporter = NewUserSchemaExporter(s.mockService)
	s.logger = log.GetLogger()
}

func (s *UserSchemaExporterTestSuite) TestNewUserSchemaExporter() {
	assert.NotNil(s.T(), s.exporter)
	assert.Equal(s.T(), s.mockService, s.exporter.service)
}

func (s *UserSchemaExporterTestSuite) TestGetResourceType() {
	assert.Equal(s.T(), resourceTypeUserSchema, s.exporter.GetResourceType())
}

func (s *UserSchemaExporterTestSuite) TestGetParameterizerType() {
	assert.Equal(s.T(), paramTypUserSchema, s.exporter.GetParameterizerType())
}

func (s *UserSchemaExporterTestSuite) TestGetAllResourceIDs_Success() {
	expectedSchemas := &userschema.UserSchemaListResponse{
		Schemas: []userschema.UserSchemaListItem{
			{ID: "schema1", Name: "Schema 1"},
			{ID: "schema2", Name: "Schema 2"},
		},
	}

	s.mockService.EXPECT().GetUserSchemaList(0, 1000).Return(expectedSchemas, nil)

	ids, err := s.exporter.GetAllResourceIDs()

	assert.Nil(s.T(), err)
	assert.Len(s.T(), ids, 2)
	assert.Equal(s.T(), "schema1", ids[0])
	assert.Equal(s.T(), "schema2", ids[1])
}

func (s *UserSchemaExporterTestSuite) TestGetAllResourceIDs_Error() {
	expectedError := &serviceerror.ServiceError{
		Code:  "ERR_CODE",
		Error: "test error",
	}

	s.mockService.EXPECT().GetUserSchemaList(0, 1000).Return(nil, expectedError)

	ids, err := s.exporter.GetAllResourceIDs()

	assert.Nil(s.T(), ids)
	assert.Equal(s.T(), expectedError, err)
}

func (s *UserSchemaExporterTestSuite) TestGetAllResourceIDs_EmptyList() {
	expectedSchemas := &userschema.UserSchemaListResponse{
		Schemas: []userschema.UserSchemaListItem{},
	}

	s.mockService.EXPECT().GetUserSchemaList(0, 1000).Return(expectedSchemas, nil)

	ids, err := s.exporter.GetAllResourceIDs()

	assert.Nil(s.T(), err)
	assert.Len(s.T(), ids, 0)
}

func (s *UserSchemaExporterTestSuite) TestGetResourceByID_Success() {
	schemaJSON := json.RawMessage(`{"type": "object"}`)
	expectedSchema := &userschema.UserSchema{
		ID:     "schema1",
		Name:   "Test Schema",
		Schema: schemaJSON,
	}

	s.mockService.EXPECT().GetUserSchema("schema1").Return(expectedSchema, nil)

	resource, name, err := s.exporter.GetResourceByID("schema1")

	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "Test Schema", name)
	assert.Equal(s.T(), expectedSchema, resource)
}

func (s *UserSchemaExporterTestSuite) TestGetResourceByID_Error() {
	expectedError := &serviceerror.ServiceError{
		Code:  "ERR_CODE",
		Error: "test error",
	}

	s.mockService.EXPECT().GetUserSchema("schema1").Return(nil, expectedError)

	resource, name, err := s.exporter.GetResourceByID("schema1")

	assert.Nil(s.T(), resource)
	assert.Empty(s.T(), name)
	assert.Equal(s.T(), expectedError, err)
}

func (s *UserSchemaExporterTestSuite) TestValidateResource_Success() {
	schemaJSON := json.RawMessage(`{"type": "object"}`)
	schema := &userschema.UserSchema{
		ID:     "schema1",
		Name:   "Valid Schema",
		Schema: schemaJSON,
	}

	name, err := s.exporter.ValidateResource(schema, "schema1", s.logger)

	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "Valid Schema", name)
}

func (s *UserSchemaExporterTestSuite) TestValidateResource_EmptyName() {
	schema := &userschema.UserSchema{
		ID:   "schema1",
		Name: "",
	}

	name, err := s.exporter.ValidateResource(schema, "schema1", s.logger)

	assert.NotNil(s.T(), err)
	assert.Empty(s.T(), name)
	assert.Equal(s.T(), resourceTypeUserSchema, err.ResourceType)
	assert.Equal(s.T(), "schema1", err.ResourceID)
	assert.Equal(s.T(), "SCHEMA_VALIDATION_ERROR", err.Code)
	assert.Contains(s.T(), err.Error, "name is empty")
}

func (s *UserSchemaExporterTestSuite) TestValidateResource_NoSchema() {
	// This should succeed but log a warning
	schema := &userschema.UserSchema{
		ID:     "schema1",
		Name:   "Schema No Def",
		Schema: json.RawMessage{},
	}

	name, err := s.exporter.ValidateResource(schema, "schema1", s.logger)

	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "Schema No Def", name)
}

func (s *UserSchemaExporterTestSuite) TestValidateResource_InvalidType() {
	invalidResource := "not a schema"

	name, err := s.exporter.ValidateResource(invalidResource, "schema1", s.logger)

	assert.NotNil(s.T(), err)
	assert.Empty(s.T(), name)
	assert.Equal(s.T(), resourceTypeUserSchema, err.ResourceType)
	assert.Equal(s.T(), "schema1", err.ResourceID)
	assert.Equal(s.T(), "INVALID_TYPE", err.Code)
	assert.Equal(s.T(), "Invalid resource type", err.Error)
}
