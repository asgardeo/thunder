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

package usertype_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	declarativeresource "github.com/asgardeo/thunder/internal/system/declarative_resource"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/usertype"
	"github.com/asgardeo/thunder/tests/mocks/usertypemock"
)

// UserTypeExporterTestSuite tests the UserTypeExporter.
type UserTypeExporterTestSuite struct {
	suite.Suite
	mockService *usertypemock.UserTypeServiceInterfaceMock
	exporter    *usertype.UserTypeExporter
	logger      *log.Logger
}

func TestUserTypeExporterTestSuite(t *testing.T) {
	suite.Run(t, new(UserTypeExporterTestSuite))
}

func (s *UserTypeExporterTestSuite) SetupTest() {
	s.mockService = usertypemock.NewUserTypeServiceInterfaceMock(s.T())
	s.exporter = usertype.NewUserTypeExporterForTest(s.mockService)
	s.logger = log.GetLogger()
}

func (s *UserTypeExporterTestSuite) TestNewUserTypeExporter() {
	assert.NotNil(s.T(), s.exporter)
}

func (s *UserTypeExporterTestSuite) TestGetResourceType() {
	assert.Equal(s.T(), "user_type", s.exporter.GetResourceType())
}

func (s *UserTypeExporterTestSuite) TestGetParameterizerType() {
	assert.Equal(s.T(), "UserType", s.exporter.GetParameterizerType())
}

func (s *UserTypeExporterTestSuite) TestGetAllResourceIDs_Success() {
	expectedResponse := &usertype.UserTypeListResponse{
		Types: []usertype.UserTypeListItem{
			{ID: "type1", Name: "Type 1"},
			{ID: "type2", Name: "Type 2"},
		},
	}

	s.mockService.EXPECT().GetUserTypeList(100, 0).Return(expectedResponse, nil)

	ids, err := s.exporter.GetAllResourceIDs()

	assert.Nil(s.T(), err)
	assert.Len(s.T(), ids, 2)
	assert.Equal(s.T(), "schema1", ids[0])
	assert.Equal(s.T(), "schema2", ids[1])
}

func (s *UserTypeExporterTestSuite) TestGetAllResourceIDs_Error() {
	expectedError := &serviceerror.ServiceError{
		Code:  "ERR_CODE",
		Error: "test error",
	}

	s.mockService.EXPECT().GetUserTypeList(100, 0).Return(nil, expectedError)

	ids, err := s.exporter.GetAllResourceIDs()

	assert.Nil(s.T(), ids)
	assert.Equal(s.T(), expectedError, err)
}

func (s *UserTypeExporterTestSuite) TestGetAllResourceIDs_EmptyList() {
	expectedResponse := &usertype.UserTypeListResponse{
		Types: []usertype.UserTypeListItem{},
	}

	s.mockService.EXPECT().GetUserTypeList(100, 0).Return(expectedResponse, nil)

	ids, err := s.exporter.GetAllResourceIDs()

	assert.Nil(s.T(), err)
	assert.Len(s.T(), ids, 0)
}

func (s *UserTypeExporterTestSuite) TestGetResourceByID_Success() {
	expectedSchema := &usertype.UserType{
		ID:   "schema1",
		Name: "Test Schema",
	}

	s.mockService.EXPECT().GetUserType("schema1").Return(expectedSchema, nil)

	resource, name, err := s.exporter.GetResourceByID("schema1")

	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "Test Schema", name)
	assert.Equal(s.T(), expectedSchema, resource)
}

func (s *UserTypeExporterTestSuite) TestGetResourceByID_Error() {
	expectedError := &serviceerror.ServiceError{
		Code:  "ERR_CODE",
		Error: "test error",
	}

	s.mockService.EXPECT().GetUserType("schema1").Return(nil, expectedError)

	resource, name, err := s.exporter.GetResourceByID("schema1")

	assert.Nil(s.T(), resource)
	assert.Empty(s.T(), name)
	assert.Equal(s.T(), expectedError, err)
}

func (s *UserTypeExporterTestSuite) TestValidateResource_Success() {
	schema := &usertype.UserType{
		ID:     "schema1",
		Name:   "Valid Schema",
		Schema: json.RawMessage(`{"field": "value"}`),
	}

	name, err := s.exporter.ValidateResource(schema, "schema1", s.logger)

	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "Valid Schema", name)
}

func (s *UserTypeExporterTestSuite) TestValidateResource_InvalidType() {
	invalidResource := "not a schema"

	name, err := s.exporter.ValidateResource(invalidResource, "schema1", s.logger)

	assert.Empty(s.T(), name)
	assert.NotNil(s.T(), err)
	assert.Equal(s.T(), "user_type", err.ResourceType)
	assert.Equal(s.T(), "schema1", err.ResourceID)
	assert.Equal(s.T(), "INVALID_TYPE", err.Code)
}

func (s *UserTypeExporterTestSuite) TestValidateResource_EmptyName() {
	schema := &usertype.UserType{
		ID:   "schema1",
		Name: "",
	}

	name, err := s.exporter.ValidateResource(schema, "schema1", s.logger)

	assert.Empty(s.T(), name)
	assert.NotNil(s.T(), err)
	assert.Equal(s.T(), "user_type", err.ResourceType)
	assert.Equal(s.T(), "schema1", err.ResourceID)
	assert.Equal(s.T(), "SCHEMA_VALIDATION_ERROR", err.Code)
	assert.Contains(s.T(), err.Error, "name is empty")
}

func (s *UserTypeExporterTestSuite) TestValidateResource_NoSchema() {
	schema := &usertype.UserType{
		ID:     "schema1",
		Name:   "Test Schema",
		Schema: json.RawMessage(`{}`),
	}

	name, err := s.exporter.ValidateResource(schema, "schema1", s.logger)

	// Should still succeed but log a warning
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "Test Schema", name)
}

func (s *UserTypeExporterTestSuite) TestUserTypeExporterImplementsInterface() {
	var _ declarativeresource.ResourceExporter = (*usertype.UserTypeExporter)(nil)
}
