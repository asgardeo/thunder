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

package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/flow/common"
)

type MetaBuilderTestSuite struct {
	suite.Suite
}

func TestMetaBuilderSuite(t *testing.T) {
	suite.Run(t, new(MetaBuilderTestSuite))
}

func (suite *MetaBuilderTestSuite) TestNewMetaBuilder() {
	builder := NewMetaBuilder()

	assert.NotNil(suite.T(), builder)
	assert.Equal(suite.T(), "dynamic", builder.idPrefix)
	assert.NotNil(suite.T(), builder.configs)
}

func (suite *MetaBuilderTestSuite) TestMetaBuilder_SingleInput() {
	testCases := []struct {
		name         string
		buildFunc    func() MetaStructure
		validateMeta func(t *testing.T, meta MetaStructure)
	}{
		{
			name: "Single SELECT input with all options",
			buildFunc: func() MetaStructure {
				input := common.Input{
					Ref:        "usertype_input",
					Identifier: "userType",
					Type:       "SELECT",
					Required:   true,
				}
				return NewMetaBuilder().
					WithIDPrefix("usertype").
					WithHeading("{{ t(signup:heading) }}").
					WithInput(input, MetaInputConfig{
						Label:       "{{ t(elements:fields.usertype.label) }}",
						Placeholder: "{{ t(elements:fields.usertype.placeholder) }}",
					}).
					WithSubmitButton("{{ t(elements:buttons.submit.text) }}").
					Build()
			},
			validateMeta: func(t *testing.T, meta MetaStructure) {
				assert.Len(t, meta.Components, 2)

				// First component should be heading
				assert.Equal(t, "TEXT", meta.Components[0].Type)
				assert.Equal(t, "usertype_heading", meta.Components[0].ID)
				assert.Equal(t, "{{ t(signup:heading) }}", meta.Components[0].Label)
				assert.Equal(t, "HEADING_2", meta.Components[0].Variant)

				// Second component should be block
				assert.Equal(t, "BLOCK", meta.Components[1].Type)
				assert.Equal(t, "usertype_block", meta.Components[1].ID)
				assert.Len(t, meta.Components[1].Components, 2) // input + submit

				// Check input component inside block
				inputComp := meta.Components[1].Components[0]
				assert.Equal(t, "SELECT", inputComp.Type)
				assert.Equal(t, "usertype_input", inputComp.ID)
				assert.Equal(t, "userType", inputComp.Ref)
				assert.Equal(t, "{{ t(elements:fields.usertype.label) }}", inputComp.Label)
				assert.Equal(t, "{{ t(elements:fields.usertype.placeholder) }}", inputComp.Placeholder)
				assert.True(t, inputComp.Required)

				// Check submit action
				submitComp := meta.Components[1].Components[1]
				assert.Equal(t, "ACTION", submitComp.Type)
				assert.Equal(t, "usertype_submit", submitComp.ID)
				assert.Equal(t, "PRIMARY", submitComp.Variant)
				assert.Equal(t, "SUBMIT", submitComp.EventType)
			},
		},
		{
			name: "Single TEXT_INPUT without submit action",
			buildFunc: func() MetaStructure {
				input := common.Input{
					Ref:        "username_input",
					Identifier: "username",
					Type:       "TEXT_INPUT",
					Required:   true,
				}
				return NewMetaBuilder().
					WithIDPrefix("login").
					WithInput(input, MetaInputConfig{
						Label:       "Username",
						Placeholder: "Enter username",
					}).
					Build()
			},
			validateMeta: func(t *testing.T, meta MetaStructure) {
				assert.Len(t, meta.Components, 1)

				// Only block component
				assert.Equal(t, "BLOCK", meta.Components[0].Type)
				assert.Len(t, meta.Components[0].Components, 1) // only input, no submit

				inputComp := meta.Components[0].Components[0]
				assert.Equal(t, "TEXT_INPUT", inputComp.Type)
				assert.Equal(t, "username", inputComp.Ref)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			meta := tc.buildFunc()
			tc.validateMeta(suite.T(), meta)
		})
	}
}

func (suite *MetaBuilderTestSuite) TestMetaBuilder_MultipleInputs() {
	testCases := []struct {
		name         string
		buildFunc    func() MetaStructure
		validateMeta func(t *testing.T, meta MetaStructure)
	}{
		{
			name: "Multiple inputs with heading and subtitle",
			buildFunc: func() MetaStructure {
				inputs := []common.Input{
					{
						Ref:        "username_input",
						Identifier: "username",
						Type:       "TEXT_INPUT",
						Required:   true,
					},
					{
						Ref:        "password_input",
						Identifier: "password",
						Type:       "PASSWORD_INPUT",
						Required:   true,
					},
				}
				return NewMetaBuilder().
					WithIDPrefix("login").
					WithHeading("Sign In").
					WithSubtitle("Enter your credentials").
					WithInputs(inputs).
					WithInputConfig("username", MetaInputConfig{Label: "Username", Placeholder: "Enter username"}).
					WithInputConfig("password", MetaInputConfig{Label: "Password", Placeholder: "Enter password"}).
					WithSubmitButton("Submit").
					Build()
			},
			validateMeta: func(t *testing.T, meta MetaStructure) {
				// Should have heading, subtitle, and block
				assert.Len(t, meta.Components, 3)

				// First: heading
				assert.Equal(t, "TEXT", meta.Components[0].Type)
				assert.Equal(t, "login_heading", meta.Components[0].ID)
				assert.Equal(t, "HEADING_2", meta.Components[0].Variant)

				// Second: subtitle
				assert.Equal(t, "TEXT", meta.Components[1].Type)
				assert.Equal(t, "login_subtitle", meta.Components[1].ID)
				assert.Equal(t, "BODY_1", meta.Components[1].Variant)

				// Third: block with 3 components (2 inputs + submit)
				assert.Equal(t, "BLOCK", meta.Components[2].Type)
				assert.Len(t, meta.Components[2].Components, 3)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			meta := tc.buildFunc()
			tc.validateMeta(suite.T(), meta)
		})
	}
}

func (suite *MetaBuilderTestSuite) TestMetaBuilder_DefaultValues() {
	testCases := []struct {
		name         string
		buildFunc    func() MetaStructure
		validateMeta func(t *testing.T, meta MetaStructure)
	}{
		{
			name: "Empty IDPrefix uses default",
			buildFunc: func() MetaStructure {
				input := common.Input{
					Ref:        "input_1",
					Identifier: "field1",
					Type:       "TEXT_INPUT",
					Required:   false,
				}
				return NewMetaBuilder().
					WithInput(input, MetaInputConfig{}).
					WithSubmitButton("{{ t(elements:buttons.submit.text) }}").
					Build()
			},
			validateMeta: func(t *testing.T, meta MetaStructure) {
				assert.Len(t, meta.Components, 1)
				assert.Equal(t, "dynamic_block", meta.Components[0].ID)

				// Check submit action uses default ID
				submitComp := meta.Components[0].Components[1]
				assert.Equal(t, "dynamic_submit", submitComp.ID)
			},
		},
		{
			name: "No InputConfig uses identifier as label",
			buildFunc: func() MetaStructure {
				input := common.Input{
					Ref:        "input_1",
					Identifier: "field1",
					Type:       "TEXT_INPUT",
					Required:   true,
				}
				return NewMetaBuilder().
					WithIDPrefix("test").
					WithInputs([]common.Input{input}).
					Build()
			},
			validateMeta: func(t *testing.T, meta MetaStructure) {
				inputComp := meta.Components[0].Components[0]
				// Should use identifier as label when no config provided
				assert.Equal(t, "field1", inputComp.Label)
				assert.Empty(t, inputComp.Placeholder)
			},
		},
		{
			name: "Custom SubmitActionID",
			buildFunc: func() MetaStructure {
				input := common.Input{
					Ref:        "input_1",
					Identifier: "field1",
					Type:       "TEXT_INPUT",
					Required:   false,
				}
				return NewMetaBuilder().
					WithIDPrefix("test").
					WithInput(input, MetaInputConfig{}).
					WithSubmitButton("Custom Submit").
					WithSubmitButtonID("custom_action_id").
					Build()
			},
			validateMeta: func(t *testing.T, meta MetaStructure) {
				submitComp := meta.Components[0].Components[1]
				assert.Equal(t, "custom_action_id", submitComp.ID)
				assert.Equal(t, "Custom Submit", submitComp.Label)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			meta := tc.buildFunc()
			tc.validateMeta(suite.T(), meta)
		})
	}
}

func (suite *MetaBuilderTestSuite) TestMetaBuilder_EmptyInputs() {
	meta := NewMetaBuilder().
		WithIDPrefix("empty").
		WithHeading("Test Heading").
		WithSubmitButton("Submit").
		Build()

	// Should still create structure with heading and block containing only submit
	assert.Len(suite.T(), meta.Components, 2)
	assert.Equal(suite.T(), "TEXT", meta.Components[0].Type)
	assert.Equal(suite.T(), "BLOCK", meta.Components[1].Type)
	assert.Len(suite.T(), meta.Components[1].Components, 1) // only submit action
}

func (suite *MetaBuilderTestSuite) TestMetaBuilder_FluentChaining() {
	testCases := []struct {
		name         string
		buildFunc    func() MetaStructure
		validateMeta func(t *testing.T, meta MetaStructure)
	}{
		{
			name: "Basic builder with heading and submit",
			buildFunc: func() MetaStructure {
				return NewMetaBuilder().
					WithIDPrefix("test").
					WithHeading("{{ t(test:heading) }}").
					WithSubmitButton("{{ t(test:submit) }}").
					Build()
			},
			validateMeta: func(t *testing.T, meta MetaStructure) {
				assert.Len(t, meta.Components, 2) // heading + block
				assert.Equal(t, "TEXT", meta.Components[0].Type)
				assert.Equal(t, "test_heading", meta.Components[0].ID)
				assert.Equal(t, "{{ t(test:heading) }}", meta.Components[0].Label)

				// Block with only submit
				assert.Equal(t, "BLOCK", meta.Components[1].Type)
				assert.Len(t, meta.Components[1].Components, 1)
				assert.Equal(t, "ACTION", meta.Components[1].Components[0].Type)
			},
		},
		{
			name: "Builder with single input and config",
			buildFunc: func() MetaStructure {
				input := common.Input{
					Ref:        "email_input",
					Identifier: "email",
					Type:       "EMAIL_INPUT",
					Required:   true,
				}
				return NewMetaBuilder().
					WithIDPrefix("signup").
					WithHeading("{{ t(signup:heading) }}").
					WithInput(input, MetaInputConfig{
						Label:       "{{ t(elements:fields.email.label) }}",
						Placeholder: "{{ t(elements:fields.email.placeholder) }}",
					}).
					WithSubmitButton("{{ t(elements:buttons.submit.text) }}").
					Build()
			},
			validateMeta: func(t *testing.T, meta MetaStructure) {
				assert.Len(t, meta.Components, 2) // heading + block

				block := meta.Components[1]
				assert.Equal(t, "BLOCK", block.Type)
				assert.Len(t, block.Components, 2) // input + submit

				emailInput := block.Components[0]
				assert.Equal(t, "EMAIL_INPUT", emailInput.Type)
				assert.Equal(t, "email_input", emailInput.ID)
				assert.Equal(t, "email", emailInput.Ref)
				assert.Equal(t, "{{ t(elements:fields.email.label) }}", emailInput.Label)
				assert.True(t, emailInput.Required)
			},
		},
		{
			name: "Builder with subtitle",
			buildFunc: func() MetaStructure {
				return NewMetaBuilder().
					WithIDPrefix("otp").
					WithHeading("{{ t(otp:heading) }}").
					WithSubtitle("{{ t(otp:subtitle) }}").
					WithSubmitButton("Verify").
					Build()
			},
			validateMeta: func(t *testing.T, meta MetaStructure) {
				assert.Len(t, meta.Components, 3) // heading + subtitle + block

				assert.Equal(t, "TEXT", meta.Components[0].Type)
				assert.Equal(t, "HEADING_2", meta.Components[0].Variant)

				assert.Equal(t, "TEXT", meta.Components[1].Type)
				assert.Equal(t, "otp_subtitle", meta.Components[1].ID)
				assert.Equal(t, "BODY_1", meta.Components[1].Variant)
			},
		},
		{
			name: "Builder without submit button",
			buildFunc: func() MetaStructure {
				input := common.Input{
					Ref:        "test_input",
					Identifier: "test",
					Type:       "TEXT_INPUT",
					Required:   false,
				}
				return NewMetaBuilder().
					WithIDPrefix("nosubmit").
					WithInput(input, MetaInputConfig{Label: "Test"}).
					Build()
			},
			validateMeta: func(t *testing.T, meta MetaStructure) {
				block := meta.Components[0]
				assert.Len(t, block.Components, 1) // only input, no submit
				assert.Equal(t, "TEXT_INPUT", block.Components[0].Type)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			meta := tc.buildFunc()
			tc.validateMeta(suite.T(), meta)
		})
	}
}

func (suite *MetaBuilderTestSuite) TestMetaBuilder_UserTypeSelection() {
	// Test the exact pattern used by UserTypeResolver
	input := common.Input{
		Ref:        "usertype_input",
		Identifier: "userType",
		Type:       "SELECT",
		Required:   true,
		Options:    []string{"Customer", "Employee", "Partner"},
	}

	meta := NewMetaBuilder().
		WithIDPrefix("usertype").
		WithHeading("{{ t(signup:heading) }}").
		WithInput(input, MetaInputConfig{
			Label:       "{{ t(elements:fields.usertype.label) }}",
			Placeholder: "{{ t(elements:fields.usertype.placeholder) }}",
		}).
		WithSubmitButton("{{ t(elements:buttons.submit.text) }}").
		Build()

	// Verify structure
	assert.Len(suite.T(), meta.Components, 2)

	// Heading
	heading := meta.Components[0]
	assert.Equal(suite.T(), "TEXT", heading.Type)
	assert.Equal(suite.T(), "usertype_heading", heading.ID)
	assert.Equal(suite.T(), "{{ t(signup:heading) }}", heading.Label)

	// Block
	block := meta.Components[1]
	assert.Equal(suite.T(), "BLOCK", block.Type)
	assert.Len(suite.T(), block.Components, 2)

	// SELECT input
	selectInput := block.Components[0]
	assert.Equal(suite.T(), "SELECT", selectInput.Type)
	assert.Equal(suite.T(), "usertype_input", selectInput.ID)
	assert.Equal(suite.T(), "userType", selectInput.Ref)
	assert.Equal(suite.T(), "{{ t(elements:fields.usertype.label) }}", selectInput.Label)
	assert.Equal(suite.T(), "{{ t(elements:fields.usertype.placeholder) }}", selectInput.Placeholder)
	assert.True(suite.T(), selectInput.Required)
	// Verify Options field is transferred
	assert.Equal(suite.T(), []string{"Customer", "Employee", "Partner"}, selectInput.Options)

	// Submit action
	submit := block.Components[1]
	assert.Equal(suite.T(), "ACTION", submit.Type)
	assert.Equal(suite.T(), "usertype_submit", submit.ID)
	assert.Equal(suite.T(), "{{ t(elements:buttons.submit.text) }}", submit.Label)
	assert.Equal(suite.T(), "PRIMARY", submit.Variant)
	assert.Equal(suite.T(), "SUBMIT", submit.EventType)
}

func (suite *MetaBuilderTestSuite) TestMetaComponentStructure() {
	// Test that MetaComponent properly marshals with all fields
	component := MetaComponent{
		Type:        "TEXT_INPUT",
		ID:          "test_id",
		Ref:         "test_ref",
		Label:       "Test Label",
		Placeholder: "Test Placeholder",
		Variant:     "PRIMARY",
		Required:    true,
		EventType:   "SUBMIT",
		Components: []MetaComponent{
			{Type: "ACTION", ID: "nested_action"},
		},
	}

	assert.Equal(suite.T(), "TEXT_INPUT", component.Type)
	assert.Equal(suite.T(), "test_id", component.ID)
	assert.Equal(suite.T(), "test_ref", component.Ref)
	assert.Equal(suite.T(), "Test Label", component.Label)
	assert.Equal(suite.T(), "Test Placeholder", component.Placeholder)
	assert.Equal(suite.T(), "PRIMARY", component.Variant)
	assert.True(suite.T(), component.Required)
	assert.Equal(suite.T(), "SUBMIT", component.EventType)
	assert.Len(suite.T(), component.Components, 1)
}
