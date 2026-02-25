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

package model

import (
	"encoding/json"
	"sort"
	"testing"

	"github.com/stretchr/testify/suite"
)

type CredentialTestSuite struct {
	suite.Suite
}

func TestCredentialTestSuite(t *testing.T) {
	suite.Run(t, new(CredentialTestSuite))
}

func (s *CredentialTestSuite) TestIsCredential_StringProperty() {
	schema, err := CompileUserSchema(json.RawMessage(`{
		"password": {"type": "string", "credential": true},
		"email": {"type": "string"}
	}`))
	s.Require().NoError(err)

	s.Require().True(schema.properties["password"].isCredential())
	s.Require().False(schema.properties["email"].isCredential())
}

func (s *CredentialTestSuite) TestIsCredential_NumberProperty() {
	schema, err := CompileUserSchema(json.RawMessage(`{
		"pin": {"type": "number", "credential": true},
		"age": {"type": "number"}
	}`))
	s.Require().NoError(err)

	s.Require().True(schema.properties["pin"].isCredential())
	s.Require().False(schema.properties["age"].isCredential())
}

func (s *CredentialTestSuite) TestIsCredential_BooleanReturnsFalse() {
	schema, err := CompileUserSchema(json.RawMessage(`{
		"active": {"type": "boolean"}
	}`))
	s.Require().NoError(err)

	s.Require().False(schema.properties["active"].isCredential())
}

func (s *CredentialTestSuite) TestIsCredential_ObjectReturnsFalse() {
	schema, err := CompileUserSchema(json.RawMessage(`{
		"address": {"type": "object", "properties": {"city": {"type": "string"}}}
	}`))
	s.Require().NoError(err)

	s.Require().False(schema.properties["address"].isCredential())
}

func (s *CredentialTestSuite) TestIsCredential_ArrayReturnsFalse() {
	schema, err := CompileUserSchema(json.RawMessage(`{
		"tags": {"type": "array", "items": {"type": "string"}}
	}`))
	s.Require().NoError(err)

	s.Require().False(schema.properties["tags"].isCredential())
}

func (s *CredentialTestSuite) TestGetCredentialAttributes_ReturnsOnlyCredentialProperties() {
	schema, err := CompileUserSchema(json.RawMessage(`{
		"password": {"type": "string", "credential": true},
		"apiKey": {"type": "string", "credential": true},
		"email": {"type": "string", "unique": true},
		"age": {"type": "number"},
		"active": {"type": "boolean"}
	}`))
	s.Require().NoError(err)

	fields := schema.GetCredentialAttributes()
	sort.Strings(fields)
	s.Require().Equal([]string{"apiKey", "password"}, fields)
}

func (s *CredentialTestSuite) TestGetCredentialAttributes_EmptyWhenNoCredentials() {
	schema, err := CompileUserSchema(json.RawMessage(`{
		"email": {"type": "string"},
		"age": {"type": "number"}
	}`))
	s.Require().NoError(err)

	fields := schema.GetCredentialAttributes()
	s.Require().Empty(fields)
}

func (s *CredentialTestSuite) TestCredentialFieldDefaultsFalse() {
	schema, err := CompileUserSchema(json.RawMessage(`{
		"name": {"type": "string"}
	}`))
	s.Require().NoError(err)

	s.Require().False(schema.properties["name"].isCredential())
}

func (s *CredentialTestSuite) TestCredentialFieldExplicitFalse() {
	schema, err := CompileUserSchema(json.RawMessage(`{
		"name": {"type": "string", "credential": false}
	}`))
	s.Require().NoError(err)

	s.Require().False(schema.properties["name"].isCredential())
}

func (s *CredentialTestSuite) TestCredentialFieldInvalidValue() {
	_, err := CompileUserSchema(json.RawMessage(`{
		"password": {"type": "string", "credential": "yes"}
	}`))
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "'credential' field must be a boolean")
}

func (s *CredentialTestSuite) TestCredentialFieldNotAllowedOnBoolean() {
	_, err := CompileUserSchema(json.RawMessage(`{
		"active": {"type": "boolean", "credential": true}
	}`))
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "invalid field 'credential'")
}

func (s *CredentialTestSuite) TestCredentialFieldNotAllowedOnObject() {
	_, err := CompileUserSchema(json.RawMessage(`{
		"address": {"type": "object", "credential": true, "properties": {"city": {"type": "string"}}}
	}`))
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "invalid field 'credential'")
}

func (s *CredentialTestSuite) TestCredentialFieldNotAllowedOnArray() {
	_, err := CompileUserSchema(json.RawMessage(`{
		"tags": {"type": "array", "credential": true, "items": {"type": "string"}}
	}`))
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "invalid field 'credential'")
}
