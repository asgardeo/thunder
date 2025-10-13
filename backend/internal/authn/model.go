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

package authn

import "github.com/asgardeo/thunder/internal/idp"

// IDPAuthInitData represents the data returned when initiating IDP authentication.
type IDPAuthInitData struct {
	RedirectURL  string
	SessionToken string
}

// AuthSessionData represents the data stored in the authentication session token.
type AuthSessionData struct {
	IDPID   string      `json:"idp_id"`
	IDPType idp.IDPType `json:"idp_type"`
}

// AuthenticationResponseDTO represents the data transfer object for the authentication response.
type AuthenticationResponseDTO struct {
	ID               string `json:"id"`
	Type             string `json:"type,omitempty"`
	OrganizationUnit string `json:"organization_unit,omitempty"`
	Assertion        string `json:"assertion,omitempty"`
}

// IDPAuthInitRequestDTO is the request to initiate IDP authentication.
type IDPAuthInitRequestDTO struct {
	IDPID string `json:"idp_id"`
}

// IDPAuthInitResponseDTO is the response after initiating IDP authentication.
type IDPAuthInitResponseDTO struct {
	RedirectURL  string `json:"redirect_url,omitempty"`
	SessionToken string `json:"session_token"`
}

// IDPAuthFinishRequestDTO is the request to complete IDP authentication.
type IDPAuthFinishRequestDTO struct {
	SessionToken  string `json:"session_token"`
	SkipAssertion bool   `json:"skip_assertion"`
	Code          string `json:"code"`
}

// SendOTPAuthRequestDTO is the request to send an OTP for authentication.
type SendOTPAuthRequestDTO struct {
	SenderID  string `json:"sender_id"`
	Recipient string `json:"recipient"`
}

// SendOTPAuthResponseDTO is the response after sending an OTP for authentication.
type SendOTPAuthResponseDTO struct {
	Status       string `json:"status"`
	SessionToken string `json:"session_token"`
}

// VerifyOTPAuthRequestDTO is the request to verify an OTP for authentication.
type VerifyOTPAuthRequestDTO struct {
	SessionToken  string `json:"session_token"`
	SkipAssertion bool   `json:"skip_assertion"`
	OTP           string `json:"otp"`
}
