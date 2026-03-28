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
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/system/log"
)

type SchemaValidateTestSuite struct {
	suite.Suite
	logger *log.Logger
}

func TestSchemaValidateTestSuite(t *testing.T) {
	suite.Run(t, new(SchemaValidateTestSuite))
}

func (s *SchemaValidateTestSuite) SetupTest() {
	s.logger = log.GetLogger()
}

func (s *SchemaValidateTestSuite) TestValidAttributes_Pass() {
	schema, err := CompileUserSchema(json.RawMessage(`{
		"email": {"type": "string", "required": true},
		"age": {"type": "number"}
	}`))
	s.Require().NoError(err)

	ok, err := schema.Validate(json.RawMessage(`{"email":"user@example.com","age":30}`), s.logger, false)
	s.Require().NoError(err)
	s.Require().True(ok)
}

func (s *SchemaValidateTestSuite) TestExtraTopLevelAttribute_Rejected() {
	schema, err := CompileUserSchema(json.RawMessage(`{
		"email": {"type": "string", "required": true}
	}`))
	s.Require().NoError(err)

	ok, err := schema.Validate(json.RawMessage(`{"email":"user@example.com","address":"123 Main St"}`), s.logger, false)
	s.Require().NoError(err)
	s.Require().False(ok)
}

func (s *SchemaValidateTestSuite) TestExtraNestedObjectAttribute_Rejected() {
	schema, err := CompileUserSchema(json.RawMessage(`{
		"address": {
			"type": "object",
			"properties": {
				"city": {"type": "string"}
			}
		}
	}`))
	s.Require().NoError(err)

	ok, err := schema.Validate(json.RawMessage(`{"address":{"city":"NYC","zip":"10001"}}`), s.logger, false)
	s.Require().NoError(err)
	s.Require().False(ok)
}

func (s *SchemaValidateTestSuite) TestValidOnlyDeclaredAttributes_Pass() {
	schema, err := CompileUserSchema(json.RawMessage(`{
		"email": {"type": "string"},
		"age": {"type": "number"},
		"active": {"type": "boolean"}
	}`))
	s.Require().NoError(err)

	ok, err := schema.Validate(json.RawMessage(`{"email":"a@b.com","age":25,"active":true}`), s.logger, false)
	s.Require().NoError(err)
	s.Require().True(ok)
}

func (s *SchemaValidateTestSuite) TestSubsetOfDeclaredAttributes_Pass() {
	schema, err := CompileUserSchema(json.RawMessage(`{
		"email": {"type": "string"},
		"age": {"type": "number"},
		"active": {"type": "boolean"}
	}`))
	s.Require().NoError(err)

	ok, err := schema.Validate(json.RawMessage(`{"email":"a@b.com"}`), s.logger, false)
	s.Require().NoError(err)
	s.Require().True(ok)
}

func (s *SchemaValidateTestSuite) TestMultipleExtraAttributes_Rejected() {
	schema, err := CompileUserSchema(json.RawMessage(`{
		"email": {"type": "string"}
	}`))
	s.Require().NoError(err)

	ok, err := schema.Validate(json.RawMessage(`{"email":"a@b.com","foo":"bar","baz":123}`), s.logger, false)
	s.Require().NoError(err)
	s.Require().False(ok)
}

func (s *SchemaValidateTestSuite) TestDeeplyNestedExtraAttribute_Rejected() {
	schema, err := CompileUserSchema(json.RawMessage(`{
		"profile": {
			"type": "object",
			"properties": {
				"address": {
					"type": "object",
					"properties": {
						"city": {"type": "string"}
					}
				}
			}
		}
	}`))
	s.Require().NoError(err)

	ok, err := schema.Validate(json.RawMessage(`{
		"profile": {
			"address": {
				"city": "NYC",
				"country": "US"
			}
		}
	}`), s.logger, false)
	s.Require().NoError(err)
	s.Require().False(ok)
}

func (s *SchemaValidateTestSuite) TestEmptyAttributes_Pass() {
	schema, err := CompileUserSchema(json.RawMessage(`{
		"email": {"type": "string"}
	}`))
	s.Require().NoError(err)

	ok, err := schema.Validate(json.RawMessage(`{}`), s.logger, false)
	s.Require().NoError(err)
	s.Require().True(ok)
}

func (s *SchemaValidateTestSuite) TestNilAttributes_Pass() {
	schema, err := CompileUserSchema(json.RawMessage(`{
		"email": {"type": "string"}
	}`))
	s.Require().NoError(err)

	ok, err := schema.Validate(nil, s.logger, false)
	s.Require().NoError(err)
	s.Require().True(ok)
}

func (s *SchemaValidateTestSuite) TestValidNestedObjectAttributes_Pass() {
	schema, err := CompileUserSchema(json.RawMessage(`{
		"address": {
			"type": "object",
			"properties": {
				"street": {"type": "string"},
				"city": {"type": "string"}
			}
		}
	}`))
	s.Require().NoError(err)

	ok, err := schema.Validate(json.RawMessage(`{"address":{"street":"123 Main","city":"NYC"}}`), s.logger, false)
	s.Require().NoError(err)
	s.Require().True(ok)
}

func (s *SchemaValidateTestSuite) TestDisplayNameOnAllPropertyTypes_CompileSuccess() {
	schema, err := CompileUserSchema(json.RawMessage(`{
		"given_name": {"type": "string", "required": true, "displayName": "First Name"},
		"age": {"type": "number", "displayName": "Age"},
		"active": {"type": "boolean", "displayName": "Is Active"},
		"address": {
			"type": "object",
			"displayName": "Home Address",
			"properties": {
				"city": {"type": "string", "displayName": "City"}
			}
		},
		"tags": {
			"type": "array",
			"displayName": "Tags",
			"items": {"type": "string"}
		}
	}`))
	s.Require().NoError(err)
	s.Require().NotNil(schema)

	ok, err := schema.Validate(json.RawMessage(`{
		"given_name": "John",
		"age": 30,
		"active": true,
		"address": {"city": "NYC"},
		"tags": ["admin"]
	}`), s.logger, false)
	s.Require().NoError(err)
	s.Require().True(ok)
}

func (s *SchemaValidateTestSuite) TestDisplayNameWithI18nPattern_CompileSuccess() {
	schema, err := CompileUserSchema(json.RawMessage(`{
		"family_name": {"type": "string", "displayName": "{{t(custom:user.familyName)}}"}
	}`))
	s.Require().NoError(err)
	s.Require().NotNil(schema)
}

func (s *SchemaValidateTestSuite) TestDisplayNameInvalidType_CompileError() {
	_, err := CompileUserSchema(json.RawMessage(`{
		"email": {"type": "string", "displayName": 123}
	}`))
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "'displayName' field must be a string")
}

func (s *SchemaValidateTestSuite) TestSchemaWithoutDisplayName_CompileSuccess() {
	schema, err := CompileUserSchema(json.RawMessage(`{
		"email": {"type": "string", "required": true}
	}`))
	s.Require().NoError(err)
	s.Require().NotNil(schema)
}

func (s *SchemaValidateTestSuite) TestValidateAsDisplayAttribute_RejectsCredential() {
	schema, err := CompileUserSchema(json.RawMessage(`{
		"email": {"type": "string", "required": true},
		"password": {"type": "string", "required": true, "credential": true}
	}`))
	s.Require().NoError(err)

	s.Require().Equal(DisplayAttributeIsCredential, schema.ValidateAsDisplayAttribute("password"))
	s.Require().Equal(DisplayAttributeValid, schema.ValidateAsDisplayAttribute("email"))
}

func (s *SchemaValidateTestSuite) TestValidate_SkipCredentialRequired() {
	emailAndPasswordSchema := json.RawMessage(`{
		"email": {"type": "string", "required": true},
		"password": {"type": "string", "required": true, "credential": true}
	}`)

	schema, err := CompileUserSchema(emailAndPasswordSchema)
	s.Require().NoError(err)

	ok, err := schema.Validate(json.RawMessage(`{"email":"user@example.com"}`), s.logger, true)
	s.Require().NoError(err)
	s.Require().True(ok, "missing credential should pass when skipCredentialRequired=true")

	ok, err = schema.Validate(json.RawMessage(`{}`), s.logger, true)
	s.Require().NoError(err)
	s.Require().False(ok, "missing required non-credential should still fail when skipCredentialRequired=true")

	ok, err = schema.Validate(json.RawMessage(`{"email":"user@example.com"}`), s.logger, false)
	s.Require().NoError(err)
	s.Require().False(ok, "missing required credential should fail when skipCredentialRequired=false")
}
