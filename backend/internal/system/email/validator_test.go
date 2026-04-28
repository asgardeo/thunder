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

package email

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type EmailValidationTestSuite struct {
	suite.Suite
}

func TestEmailValidationTestSuite(t *testing.T) {
	suite.Run(t, new(EmailValidationTestSuite))
}

// --- Valid Cases ---

func (suite *EmailValidationTestSuite) TestIsValidEmail_Standard_Success() {
	actual := IsValidEmail("test@example.com")
	suite.True(actual)
}

func (suite *EmailValidationTestSuite) TestIsValidEmail_PlusAddressingAndSubdomains_Success() {
	actual := IsValidEmail("user.name+tag@example.co.uk")
	suite.True(actual)
}

func (suite *EmailValidationTestSuite) TestIsValidEmail_LeadingAndTrailingSpaces_Success() {
	// Should pass because the function uses strings.TrimSpace
	actual := IsValidEmail("  test@example.com  ")
	suite.True(actual)
}

// --- Invalid Cases ---

func (suite *EmailValidationTestSuite) TestIsValidEmail_EmptyString_False() {
	actual := IsValidEmail("")
	suite.False(actual)
}

func (suite *EmailValidationTestSuite) TestIsValidEmail_OnlySpaces_False() {
	actual := IsValidEmail("     ")
	suite.False(actual)
}

func (suite *EmailValidationTestSuite) TestIsValidEmail_MissingDomain_False() {
	actual := IsValidEmail("test@")
	suite.False(actual)
}

func (suite *EmailValidationTestSuite) TestIsValidEmail_MissingUsername_False() {
	actual := IsValidEmail("@example.com")
	suite.False(actual)
}

func (suite *EmailValidationTestSuite) TestIsValidEmail_MissingAtSymbol_False() {
	actual := IsValidEmail("testexample.com")
	suite.False(actual)
}

func (suite *EmailValidationTestSuite) TestIsValidEmail_WithDisplayName_False() {
	// Fails because addr.Address ("john@example.com") != trimmed emailAddr
	actual := IsValidEmail("John Doe <john@example.com>")
	suite.False(actual)
}

// --- Header Injection Prevention Cases (CR/LF) ---

func (suite *EmailValidationTestSuite) TestIsValidEmail_WithCarriageReturn_False() {
	actual := IsValidEmail("test@example.com\r")
	suite.False(actual)
}

func (suite *EmailValidationTestSuite) TestIsValidEmail_WithLineFeed_False() {
	actual := IsValidEmail("test@example\n.com")
	suite.False(actual)
}

func (suite *EmailValidationTestSuite) TestIsValidEmail_WithInlineCRLF_False() {
	actual := IsValidEmail("test@\r\nexample.com")
	suite.False(actual)
}
