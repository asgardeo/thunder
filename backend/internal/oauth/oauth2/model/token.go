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

// Package model defines the data structures used in the OAuth2 module.
package model

// TokenRequest represents the OAuth2 token request.
type TokenRequest struct {
	GrantType    string `json:"grant_type"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Scope        string `json:"scope,omitempty"`
	Username     string `json:"username,omitempty"`
	Password     string `json:"password,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	CodeVerifier string `json:"code_verifier,omitempty"`
	Code         string `json:"code,omitempty"`
	RedirectURI  string `json:"redirect_uri,omitempty"`
	Resource     string `json:"resource,omitempty"`
}

// TokenResponse represents the OAuth2 token response.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
}

// TokenContext holds context data for the token issuance.
type TokenContext struct {
	TokenAttributes map[string]interface{} `json:"token_attributes,omitempty"`
}

// TokenDTO represents the data transfer object for tokens.
type TokenDTO struct {
	Token          string                 `json:"token"`
	TokenType      string                 `json:"token_type"`
	IssuedAt       int64                  `json:"issued_at"`
	ExpiresIn      int64                  `json:"expires_in"`
	Scopes         []string               `json:"scopes,omitempty"`
	ClientID       string                 `json:"client_id"`
	UserAttributes map[string]interface{} `json:"user_attributes,omitempty"`
}

// TokenResponseDTO represents the data transfer object for token responses.
type TokenResponseDTO struct {
	AccessToken  TokenDTO `json:"access_token"`
	RefreshToken TokenDTO `json:"refresh_token,omitempty"`
	IDToken      TokenDTO `json:"id_token,omitempty"`
}
