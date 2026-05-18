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
	"regexp"
	"sync"
	"unicode/utf8"

	"github.com/thunder-id/thunderid/internal/flow/common"
)

var (
	regexCache   = make(map[string]*regexp.Regexp)
	regexCacheMu sync.RWMutex
)

// getCompiledRegex returns a cached compiled regex for the given pattern,
// compiling and caching it on first use. Uses Go's RE2 engine, which guarantees
// linear-time matching and prevents ReDoS.
func getCompiledRegex(pattern string) (*regexp.Regexp, error) {
	regexCacheMu.RLock()
	if re, ok := regexCache[pattern]; ok {
		regexCacheMu.RUnlock()
		return re, nil
	}
	regexCacheMu.RUnlock()

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	regexCacheMu.Lock()
	regexCache[pattern] = re
	regexCacheMu.Unlock()
	return re, nil
}

// validateInputValues evaluates the validation rules defined on the given inputs
// against the user-submitted values in userInputs. It returns a FieldError for
// each failing rule. Multiple failing rules on the same field each produce a
// separate entry, preserving the order in which they are declared on the input.
func validateInputValues(inputs []common.Input, userInputs map[string]string) []common.FieldError {
	var fieldErrors []common.FieldError

	for _, input := range inputs {
		if len(input.Validation) == 0 {
			continue
		}
		value, ok := userInputs[input.Identifier]
		if !ok {
			continue
		}

		for _, rule := range input.Validation {
			if !ruleFails(rule, value) {
				continue
			}
			fieldErrors = append(fieldErrors, common.FieldError{
				Identifier: input.Identifier,
				Message:    resolveRuleMessage(rule),
			})
		}
	}

	return fieldErrors
}

// ruleFails returns true when the given rule's constraint is violated by value.
// Unknown rule types are treated as a no-op (return false) — they are ignored
// rather than failing the input, leaving room for forward-compatible additions.
func ruleFails(rule common.ValidationRule, value string) bool {
	switch rule.Type {
	case common.ValidationTypeRegex:
		pattern, ok := rule.Value.(string)
		if !ok || pattern == "" {
			return false
		}
		re, err := getCompiledRegex(pattern)
		if err != nil {
			return true
		}
		return !re.MatchString(value)

	case common.ValidationTypeMinLength:
		minLen, ok := numericRuleValue(rule.Value)
		if !ok {
			return false
		}
		// Count Unicode code points (runes) rather than bytes so multibyte
		// characters such as "café", "日本語", or emoji are counted as the user
		// perceives them rather than as their UTF-8 byte length.
		return utf8.RuneCountInString(value) < minLen

	case common.ValidationTypeMaxLength:
		maxLen, ok := numericRuleValue(rule.Value)
		if !ok {
			return false
		}
		return utf8.RuneCountInString(value) > maxLen
	}
	return false
}

// resolveRuleMessage returns the rule's Message if set, otherwise the default
// i18n key for the rule type. Unknown rule types fall back to the regex default.
func resolveRuleMessage(rule common.ValidationRule) string {
	if rule.Message != "" {
		return rule.Message
	}
	switch rule.Type {
	case common.ValidationTypeMinLength:
		return common.DefaultValidationMessageMinLength
	case common.ValidationTypeMaxLength:
		return common.DefaultValidationMessageMaxLength
	default:
		return common.DefaultValidationMessageRegex
	}
}

// numericRuleValue extracts an int from a JSON-decoded numeric interface{}.
// JSON numbers decode as float64 by default; ints are also accepted to ease testing.
func numericRuleValue(v interface{}) (int, bool) {
	switch n := v.(type) {
	case float64:
		return int(n), true
	case int:
		return n, true
	case int64:
		return int(n), true
	}
	return 0, false
}
