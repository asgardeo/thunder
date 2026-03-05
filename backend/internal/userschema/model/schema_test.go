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

	ok, err := schema.Validate(json.RawMessage(`{"email":"user@example.com","age":30}`), s.logger)
	s.Require().NoError(err)
	s.Require().True(ok)
}

func (s *SchemaValidateTestSuite) TestExtraTopLevelAttribute_Rejected() {
	schema, err := CompileUserSchema(json.RawMessage(`{
		"email": {"type": "string", "required": true}
	}`))
	s.Require().NoError(err)

	ok, err := schema.Validate(json.RawMessage(`{"email":"user@example.com","address":"123 Main St"}`), s.logger)
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

	ok, err := schema.Validate(json.RawMessage(`{"address":{"city":"NYC","zip":"10001"}}`), s.logger)
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

	ok, err := schema.Validate(json.RawMessage(`{"email":"a@b.com","age":25,"active":true}`), s.logger)
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

	ok, err := schema.Validate(json.RawMessage(`{"email":"a@b.com"}`), s.logger)
	s.Require().NoError(err)
	s.Require().True(ok)
}

func (s *SchemaValidateTestSuite) TestMultipleExtraAttributes_Rejected() {
	schema, err := CompileUserSchema(json.RawMessage(`{
		"email": {"type": "string"}
	}`))
	s.Require().NoError(err)

	ok, err := schema.Validate(json.RawMessage(`{"email":"a@b.com","foo":"bar","baz":123}`), s.logger)
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
	}`), s.logger)
	s.Require().NoError(err)
	s.Require().False(ok)
}

func (s *SchemaValidateTestSuite) TestEmptyAttributes_Pass() {
	schema, err := CompileUserSchema(json.RawMessage(`{
		"email": {"type": "string"}
	}`))
	s.Require().NoError(err)

	ok, err := schema.Validate(json.RawMessage(`{}`), s.logger)
	s.Require().NoError(err)
	s.Require().True(ok)
}

func (s *SchemaValidateTestSuite) TestNilAttributes_Pass() {
	schema, err := CompileUserSchema(json.RawMessage(`{
		"email": {"type": "string"}
	}`))
	s.Require().NoError(err)

	ok, err := schema.Validate(nil, s.logger)
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

	ok, err := schema.Validate(json.RawMessage(`{"address":{"street":"123 Main","city":"NYC"}}`), s.logger)
	s.Require().NoError(err)
	s.Require().True(ok)
}

type HasAttributeTestSuite struct {
	suite.Suite
}

func TestHasAttributeTestSuite(t *testing.T) {
	suite.Run(t, new(HasAttributeTestSuite))
}

func (s *HasAttributeTestSuite) TestHasAttribute_StringAttribute() {
	schema, err := CompileUserSchema(json.RawMessage(`{
		"email": {"type": "string"}
	}`))
	s.Require().NoError(err)
	s.Require().True(schema.HasAttribute("email"))
}

func (s *HasAttributeTestSuite) TestHasAttribute_NumberAttribute() {
	schema, err := CompileUserSchema(json.RawMessage(`{
		"age": {"type": "number"}
	}`))
	s.Require().NoError(err)
	s.Require().True(schema.HasAttribute("age"))
}

func (s *HasAttributeTestSuite) TestHasAttribute_BooleanAttribute() {
	schema, err := CompileUserSchema(json.RawMessage(`{
		"active": {"type": "boolean"}
	}`))
	s.Require().NoError(err)
	s.Require().True(schema.HasAttribute("active"))
}

func (s *HasAttributeTestSuite) TestHasAttribute_ObjectAttribute() {
	schema, err := CompileUserSchema(json.RawMessage(`{
		"address": {"type": "object", "properties": {"city": {"type": "string"}}}
	}`))
	s.Require().NoError(err)
	s.Require().True(schema.HasAttribute("address"))
}

func (s *HasAttributeTestSuite) TestHasAttribute_ArrayAttribute() {
	schema, err := CompileUserSchema(json.RawMessage(`{
		"tags": {"type": "array", "items": {"type": "string"}}
	}`))
	s.Require().NoError(err)
	s.Require().True(schema.HasAttribute("tags"))
}

func (s *HasAttributeTestSuite) TestHasAttribute_NonExistent() {
	schema, err := CompileUserSchema(json.RawMessage(`{
		"email": {"type": "string"}
	}`))
	s.Require().NoError(err)
	s.Require().False(schema.HasAttribute("username"))
}

func (s *HasAttributeTestSuite) TestHasAttribute_EmptyName() {
	schema, err := CompileUserSchema(json.RawMessage(`{
		"email": {"type": "string"}
	}`))
	s.Require().NoError(err)
	s.Require().False(schema.HasAttribute(""))
}
