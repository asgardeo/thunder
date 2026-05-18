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

package core

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/thunder-id/thunderid/internal/flow/common"
)

type InputValidationTestSuite struct {
	suite.Suite
}

func TestInputValidationTestSuite(t *testing.T) {
	suite.Run(t, new(InputValidationTestSuite))
}

func (s *InputValidationTestSuite) TestRegexPass() {
	inputs := []common.Input{
		{
			Identifier: "email",
			Validation: []common.ValidationRule{
				{Type: common.ValidationTypeRegex, Value: "^[^@]+@[^@]+\\.[^@]+$",
					Message: "validation.email.invalid"},
			},
		},
	}
	errs := validateInputValues(inputs, map[string]string{"email": "user@example.com"})
	s.Empty(errs)
}

func (s *InputValidationTestSuite) TestRegexFail() {
	inputs := []common.Input{
		{
			Identifier: "email",
			Validation: []common.ValidationRule{
				{Type: common.ValidationTypeRegex, Value: "^[^@]+@[^@]+\\.[^@]+$",
					Message: "validation.email.invalid"},
			},
		},
	}
	errs := validateInputValues(inputs, map[string]string{"email": "not-an-email"})
	s.Len(errs, 1)
	s.Equal("email", errs[0].Identifier)
	s.Equal("validation.email.invalid", errs[0].Message)
}

func (s *InputValidationTestSuite) TestMinLengthPass() {
	inputs := []common.Input{
		{
			Identifier: "password",
			Validation: []common.ValidationRule{
				{Type: common.ValidationTypeMinLength, Value: float64(8),
					Message: "validation.password.minLength"},
			},
		},
	}
	errs := validateInputValues(inputs, map[string]string{"password": "12345678"})
	s.Empty(errs)
}

func (s *InputValidationTestSuite) TestMinLengthFail() {
	inputs := []common.Input{
		{
			Identifier: "password",
			Validation: []common.ValidationRule{
				{Type: common.ValidationTypeMinLength, Value: float64(8),
					Message: "validation.password.minLength"},
			},
		},
	}
	errs := validateInputValues(inputs, map[string]string{"password": "abc"})
	s.Len(errs, 1)
	s.Equal("password", errs[0].Identifier)
	s.Equal("validation.password.minLength", errs[0].Message)
}

func (s *InputValidationTestSuite) TestMaxLengthPass() {
	inputs := []common.Input{
		{
			Identifier: "username",
			Validation: []common.ValidationRule{
				{Type: common.ValidationTypeMaxLength, Value: float64(10),
					Message: "validation.username.maxLength"},
			},
		},
	}
	errs := validateInputValues(inputs, map[string]string{"username": "ten_chars_"})
	s.Empty(errs)
}

func (s *InputValidationTestSuite) TestMaxLengthFail() {
	inputs := []common.Input{
		{
			Identifier: "username",
			Validation: []common.ValidationRule{
				{Type: common.ValidationTypeMaxLength, Value: float64(5),
					Message: "validation.username.maxLength"},
			},
		},
	}
	errs := validateInputValues(inputs, map[string]string{"username": "way_too_long"})
	s.Len(errs, 1)
	s.Equal("validation.username.maxLength", errs[0].Message)
}

func (s *InputValidationTestSuite) TestMultipleRulesPerFieldProduceSeparateEntries() {
	// Perl-style lookaheads (?=.*[A-Z])(?=.*[0-9]) are not RE2-compatible; use alternation instead.
	inputs := []common.Input{
		{
			Identifier: "password",
			Validation: []common.ValidationRule{
				{Type: common.ValidationTypeMinLength, Value: float64(8),
					Message: "validation.password.minLength"},
				{Type: common.ValidationTypeRegex,
					Value:   "^(?:[^A-Z]*[A-Z][^0-9]*[0-9].*|[^0-9]*[0-9][^A-Z]*[A-Z].*)$",
					Message: "validation.password.complexity"},
			},
		},
	}
	errs := validateInputValues(inputs, map[string]string{"password": "abc"})
	s.Len(errs, 2)
	s.Equal("password", errs[0].Identifier)
	s.Equal("validation.password.minLength", errs[0].Message)
	s.Equal("password", errs[1].Identifier)
	s.Equal("validation.password.complexity", errs[1].Message)
}

func (s *InputValidationTestSuite) TestMultipleFieldsErrorsInSinglePass() {
	inputs := []common.Input{
		{
			Identifier: "username",
			Validation: []common.ValidationRule{
				{Type: common.ValidationTypeRegex, Value: "^[^@]+@[^@]+$",
					Message: "validation.email.invalid"},
			},
		},
		{
			Identifier: "password",
			Validation: []common.ValidationRule{
				{Type: common.ValidationTypeMinLength, Value: float64(8),
					Message: "validation.password.minLength"},
			},
		},
	}
	errs := validateInputValues(inputs, map[string]string{
		"username": "no-at-sign",
		"password": "abc",
	})
	s.Len(errs, 2)
}

func (s *InputValidationTestSuite) TestDefaultMessageUsedWhenAbsent() {
	inputs := []common.Input{
		{
			Identifier: "email",
			Validation: []common.ValidationRule{
				{Type: common.ValidationTypeRegex, Value: "^X+$"},
			},
		},
	}
	errs := validateInputValues(inputs, map[string]string{"email": "abc"})
	s.Len(errs, 1)
	s.Equal(common.DefaultValidationMessageRegex, errs[0].Message)
}

func (s *InputValidationTestSuite) TestMessagePassthroughI18nKey() {
	i18nKey := "{{i18n(validation:email.invalid)}}"
	inputs := []common.Input{
		{
			Identifier: "email",
			Validation: []common.ValidationRule{
				{Type: common.ValidationTypeRegex, Value: "^X+$", Message: i18nKey},
			},
		},
	}
	errs := validateInputValues(inputs, map[string]string{"email": "abc"})
	s.Len(errs, 1)
	s.Equal(i18nKey, errs[0].Message)
}

func (s *InputValidationTestSuite) TestAbsentInputSkipsValidation() {
	inputs := []common.Input{
		{
			Identifier: "email",
			Validation: []common.ValidationRule{
				{Type: common.ValidationTypeRegex, Value: "^X+$",
					Message: "validation.email.invalid"},
			},
		},
	}
	errs := validateInputValues(inputs, map[string]string{})
	s.Empty(errs)
}

func (s *InputValidationTestSuite) TestUnknownRuleTypeIgnored() {
	inputs := []common.Input{
		{
			Identifier: "field",
			Validation: []common.ValidationRule{
				{Type: "unknownRuleType", Value: "anything", Message: "should.not.appear"},
			},
		},
	}
	errs := validateInputValues(inputs, map[string]string{"field": "any value"})
	s.Empty(errs)
}

// Empty pattern has no effective constraint, so the rule passes.
func (s *InputValidationTestSuite) TestRegexEmptyPatternPasses() {
	inputs := []common.Input{
		{
			Identifier: "field",
			Validation: []common.ValidationRule{
				{Type: common.ValidationTypeRegex, Value: "", Message: "should.not.appear"},
			},
		},
	}
	errs := validateInputValues(inputs, map[string]string{"field": "anything"})
	s.Empty(errs)
}

// Non-numeric value (malformed flow definition) fails open rather than failing every submission.
func (s *InputValidationTestSuite) TestMinLengthNonNumericValuePasses() {
	inputs := []common.Input{
		{
			Identifier: "field",
			Validation: []common.ValidationRule{
				{Type: common.ValidationTypeMinLength, Value: "not-a-number", Message: "should.not.appear"},
			},
		},
	}
	errs := validateInputValues(inputs, map[string]string{"field": "abc"})
	s.Empty(errs)
}

func (s *InputValidationTestSuite) TestMaxLengthNonNumericValuePasses() {
	inputs := []common.Input{
		{
			Identifier: "field",
			Validation: []common.ValidationRule{
				{Type: common.ValidationTypeMaxLength, Value: []int{1, 2}, Message: "should.not.appear"},
			},
		},
	}
	errs := validateInputValues(inputs, map[string]string{"field": "abcdefg"})
	s.Empty(errs)
}

func (s *InputValidationTestSuite) TestMinLengthDefaultMessageWhenMessageAbsent() {
	inputs := []common.Input{
		{
			Identifier: "field",
			Validation: []common.ValidationRule{
				{Type: common.ValidationTypeMinLength, Value: float64(8)},
			},
		},
	}
	errs := validateInputValues(inputs, map[string]string{"field": "abc"})
	s.Len(errs, 1)
	s.Equal(common.DefaultValidationMessageMinLength, errs[0].Message)
}

func (s *InputValidationTestSuite) TestMaxLengthDefaultMessageWhenMessageAbsent() {
	inputs := []common.Input{
		{
			Identifier: "field",
			Validation: []common.ValidationRule{
				{Type: common.ValidationTypeMaxLength, Value: float64(3)},
			},
		},
	}
	errs := validateInputValues(inputs, map[string]string{"field": "abcdef"})
	s.Len(errs, 1)
	s.Equal(common.DefaultValidationMessageMaxLength, errs[0].Message)
}

func (s *InputValidationTestSuite) TestNumericRuleValueAcceptsInt() {
	inputs := []common.Input{
		{
			Identifier: "field",
			Validation: []common.ValidationRule{
				{Type: common.ValidationTypeMinLength, Value: 8, Message: "too short"},
			},
		},
	}
	errs := validateInputValues(inputs, map[string]string{"field": "abc"})
	s.Len(errs, 1)
	s.Equal("too short", errs[0].Message)
}

func (s *InputValidationTestSuite) TestNumericRuleValueAcceptsInt64() {
	inputs := []common.Input{
		{
			Identifier: "field",
			Validation: []common.ValidationRule{
				{Type: common.ValidationTypeMaxLength, Value: int64(3), Message: "too long"},
			},
		},
	}
	errs := validateInputValues(inputs, map[string]string{"field": "abcdef"})
	s.Len(errs, 1)
	s.Equal("too long", errs[0].Message)
}

// Multibyte UTF-8 must count by runes (code points), not bytes.
func (s *InputValidationTestSuite) TestMinLengthCountsRunesNotBytes() {
	inputs := []common.Input{
		{
			Identifier: "field",
			Validation: []common.ValidationRule{
				{Type: common.ValidationTypeMinLength, Value: float64(5), Message: "too short"},
			},
		},
	}

	errs := validateInputValues(inputs, map[string]string{"field": "café"})
	s.Len(errs, 1, "café is 4 runes; must fail minLength: 5 even though byte length is 5")

	errs = validateInputValues(inputs, map[string]string{"field": "日本語!!"})
	s.Empty(errs, "日本語!! is 5 runes; must pass minLength: 5")
}

func (s *InputValidationTestSuite) TestMaxLengthCountsRunesNotBytes() {
	inputs := []common.Input{
		{
			Identifier: "field",
			Validation: []common.ValidationRule{
				{Type: common.ValidationTypeMaxLength, Value: float64(3), Message: "too long"},
			},
		},
	}

	errs := validateInputValues(inputs, map[string]string{"field": "abcd"})
	s.Len(errs, 1, "abcd is 4 runes; must fail maxLength: 3")

	errs = validateInputValues(inputs, map[string]string{"field": "日本語"})
	s.Empty(errs, "日本語 is 3 runes; must pass maxLength: 3")

	// "𠮷𠮷": two non-BMP Han Ext B characters, 1 rune each, 4 bytes each. Exercises 4-byte UTF-8.
	errs = validateInputValues(inputs, map[string]string{"field": "𠮷𠮷"})
	s.Empty(errs, "two non-BMP chars are 2 runes (8 bytes); must pass maxLength: 3")
}

// On a multi-prompt node sharing input identifiers, only the selected action's rules must fire.
func (s *InputValidationTestSuite) TestPromptNodeValidatesOnlySelectedActionInputs() {
	node := newPromptNode("prompt-1", map[string]interface{}{}, false, false)
	pn := node.(PromptNodeInterface)
	pn.SetPrompts([]common.Prompt{
		{
			// Prompt A: "credential" must contain '@'.
			Inputs: []common.Input{
				{
					Identifier: "credential",
					Required:   true,
					Validation: []common.ValidationRule{
						{Type: common.ValidationTypeRegex, Value: "@",
							Message: "must.contain.at.sign"},
					},
				},
			},
			Action: &common.Action{Ref: "submit_email", NextNode: "next"},
		},
		{
			// Prompt B: same "credential" must be all digits.
			Inputs: []common.Input{
				{
					Identifier: "credential",
					Required:   true,
					Validation: []common.ValidationRule{
						{Type: common.ValidationTypeRegex, Value: "^[0-9]+$",
							Message: "must.be.digits"},
					},
				},
			},
			Action: &common.Action{Ref: "submit_phone", NextNode: "next"},
		},
	})

	// submit_phone with digits must pass; old dedup behavior would have applied Prompt A's '@' rule.
	ctx := &NodeContext{
		ExecutionID:   "test-flow",
		CurrentAction: "submit_phone",
		UserInputs:    map[string]string{"credential": "5551234"},
	}
	resp, err := node.Execute(ctx)

	s.Nil(err)
	s.NotNil(resp)
	s.Empty(resp.FieldErrors, "phone-style value must not be validated against the email prompt's rule")
}

// Validation-failure response must include Meta when Verbose is true.
func (s *InputValidationTestSuite) TestPromptNodeValidationFailureIncludesMetaWhenVerbose() {
	node := newPromptNode("prompt-1", map[string]interface{}{}, false, false)
	pn := node.(PromptNodeInterface)
	pn.SetPrompts([]common.Prompt{
		{
			Inputs: []common.Input{
				{
					Ref:        "input_password",
					Identifier: "password",
					Required:   true,
					Validation: []common.ValidationRule{
						{Type: common.ValidationTypeMinLength, Value: float64(8),
							Message: "validation.password.minLength"},
					},
				},
			},
			Action: &common.Action{Ref: "submit", NextNode: "next"},
		},
	})
	pn.SetMeta(map[string]interface{}{
		"components": []interface{}{
			map[string]interface{}{
				"id":   "input_password",
				"ref":  "input_password",
				"type": "PASSWORD_INPUT",
			},
			map[string]interface{}{
				"id":   "submit",
				"ref":  "submit",
				"type": "ACTION",
			},
		},
	})

	ctx := &NodeContext{
		ExecutionID:   "test-flow",
		CurrentAction: "submit",
		UserInputs:    map[string]string{"password": "short"},
		Verbose:       true,
	}
	resp, err := node.Execute(ctx)

	s.Nil(err)
	s.NotNil(resp)
	s.Equal(common.NodeStatusIncomplete, resp.Status)
	s.NotEmpty(resp.FieldErrors)
	s.NotNil(resp.Meta, "Meta should be present in the response when verbose mode is enabled")
}

func (s *InputValidationTestSuite) TestInputWithoutValidationRulesNoOp() {
	inputs := []common.Input{
		{Identifier: "any", Required: true},
	}
	errs := validateInputValues(inputs, map[string]string{"any": "value"})
	s.Empty(errs)
}

func (s *InputValidationTestSuite) TestRegexCacheCompilesOnce() {
	pattern := "^cachetest[0-9]+$"
	first, err1 := getCompiledRegex(pattern)
	s.NoError(err1)
	s.NotNil(first)
	second, err2 := getCompiledRegex(pattern)
	s.NoError(err2)
	s.Same(first, second)
}

func (s *InputValidationTestSuite) TestPromptNodeReturnsFieldErrorsOnFailure() {
	node := newPromptNode("prompt-1", map[string]interface{}{}, false, false)
	pn := node.(PromptNodeInterface)
	pn.SetPrompts([]common.Prompt{
		{
			Inputs: []common.Input{
				{
					Identifier: "password",
					Required:   true,
					Validation: []common.ValidationRule{
						{Type: common.ValidationTypeMinLength, Value: float64(8),
							Message: "validation.password.minLength"},
					},
				},
			},
			Action: &common.Action{Ref: "submit", NextNode: "next"},
		},
	})

	ctx := &NodeContext{
		ExecutionID:   "test-flow",
		CurrentAction: "submit",
		UserInputs:    map[string]string{"password": "short"},
	}
	resp, err := node.Execute(ctx)

	s.Nil(err)
	s.NotNil(resp)
	s.Equal(common.NodeStatusIncomplete, resp.Status)
	s.Equal(common.NodeResponseTypeView, resp.Type)
	s.Len(resp.FieldErrors, 1)
	s.Equal("password", resp.FieldErrors[0].Identifier)
	s.Equal("validation.password.minLength", resp.FieldErrors[0].Message)
	_, stillPresent := ctx.UserInputs["password"]
	s.False(stillPresent, "failed input should be cleared from UserInputs")
}

// Validation failure must return the entire prompt (all inputs + actions), not just the failing fields.
func (s *InputValidationTestSuite) TestPromptNodeReturnsAllInputsAndActionsOnValidationFailure() {
	node := newPromptNode("prompt-1", map[string]interface{}{}, false, false)
	pn := node.(PromptNodeInterface)
	pn.SetPrompts([]common.Prompt{
		{
			Inputs: []common.Input{
				{Identifier: "username", Required: true},
				{
					Identifier: "password",
					Required:   true,
					Validation: []common.ValidationRule{
						{Type: common.ValidationTypeMinLength, Value: float64(8),
							Message: "validation.password.minLength"},
					},
				},
			},
			Action: &common.Action{Ref: "submit", NextNode: "next"},
		},
	})

	ctx := &NodeContext{
		ExecutionID:   "test-flow",
		CurrentAction: "submit",
		UserInputs:    map[string]string{"username": "valid_user", "password": "short"},
	}
	resp, err := node.Execute(ctx)

	s.Nil(err)
	s.NotNil(resp)
	s.Equal(common.NodeStatusIncomplete, resp.Status)
	s.Equal(common.NodeResponseTypeView, resp.Type)

	s.Len(resp.FieldErrors, 1)
	s.Equal("password", resp.FieldErrors[0].Identifier)

	s.Len(resp.Inputs, 2)
	identifiers := map[string]bool{}
	for _, in := range resp.Inputs {
		identifiers[in.Identifier] = true
	}
	s.True(identifiers["username"], "username should be in the re-prompted inputs")
	s.True(identifiers["password"], "password should be in the re-prompted inputs")

	s.Len(resp.Actions, 1)
	s.Equal("submit", resp.Actions[0].Ref)

	_, passwordPresent := ctx.UserInputs["password"]
	s.False(passwordPresent, "failed input cleared from UserInputs")
	s.Equal("valid_user", ctx.UserInputs["username"], "passing input retained in UserInputs")
}

func (s *InputValidationTestSuite) TestPromptNodeAdvancesOnValidSubmission() {
	node := newPromptNode("prompt-1", map[string]interface{}{}, false, false)
	pn := node.(PromptNodeInterface)
	pn.SetPrompts([]common.Prompt{
		{
			Inputs: []common.Input{
				{
					Identifier: "password",
					Required:   true,
					Validation: []common.ValidationRule{
						{Type: common.ValidationTypeMinLength, Value: float64(8),
							Message: "validation.password.minLength"},
					},
				},
			},
			Action: &common.Action{Ref: "submit", NextNode: "next"},
		},
	})

	ctx := &NodeContext{
		ExecutionID:   "test-flow",
		CurrentAction: "submit",
		UserInputs:    map[string]string{"password": "longenough"},
	}
	resp, err := node.Execute(ctx)

	s.Nil(err)
	s.NotNil(resp)
	s.Equal(common.NodeStatusComplete, resp.Status)
	s.Empty(resp.FieldErrors)
	s.Equal("next", resp.NextNodeID)
}

// LOGIN_OPTIONS variant must gate on validation rules and not advance with invalid inputs.
func (s *InputValidationTestSuite) TestLoginOptionsVariantReturnsFieldErrorsOnValidationFailure() {
	node := newPromptNode("login-chooser", map[string]interface{}{}, false, false)
	pn := node.(PromptNodeInterface)
	pn.SetVariant(common.NodeVariantLoginOptions)
	pn.SetPrompts([]common.Prompt{
		{
			Inputs: []common.Input{
				{
					Identifier: "password",
					Required:   true,
					Validation: []common.ValidationRule{
						{Type: common.ValidationTypeMinLength, Value: float64(8),
							Message: "validation.password.minLength"},
					},
				},
			},
			Action: &common.Action{Ref: "pwd", NextNode: "pwd-node"},
		},
	})

	ctx := &NodeContext{
		ExecutionID:   "test-flow",
		CurrentAction: "pwd",
		UserInputs:    map[string]string{"password": "short"},
		RuntimeData:   map[string]string{},
	}
	resp, err := node.Execute(ctx)

	s.Nil(err)
	s.NotNil(resp)
	s.Equal(common.NodeStatusIncomplete, resp.Status, "invalid input must not advance the LOGIN_OPTIONS flow")
	s.Len(resp.FieldErrors, 1)
	s.Equal("password", resp.FieldErrors[0].Identifier)
	s.Empty(resp.NextNodeID, "NextNodeID must be empty on validation failure")
}

func (s *InputValidationTestSuite) TestPromptNodeValidationDoesNotRunWhenRequiredInputMissing() {
	node := newPromptNode("prompt-1", map[string]interface{}{}, false, false)
	pn := node.(PromptNodeInterface)
	pn.SetPrompts([]common.Prompt{
		{
			Inputs: []common.Input{
				{
					Identifier: "password",
					Required:   true,
					Validation: []common.ValidationRule{
						{Type: common.ValidationTypeMinLength, Value: float64(8),
							Message: "validation.password.minLength"},
					},
				},
			},
			Action: &common.Action{Ref: "submit", NextNode: "next"},
		},
	})

	ctx := &NodeContext{
		ExecutionID:   "test-flow",
		CurrentAction: "submit",
		UserInputs:    map[string]string{},
	}
	resp, err := node.Execute(ctx)

	s.Nil(err)
	s.NotNil(resp)
	s.Equal(common.NodeStatusIncomplete, resp.Status)
	s.Empty(resp.FieldErrors)
	s.Len(resp.Inputs, 1)
}
