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

package dcr

import (
	"encoding/json"
	"strings"

	oauth2const "github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	i18nmgt "github.com/asgardeo/thunder/internal/system/i18n/mgt"
)

// Default values for DCR
const (
	ClientSecretExpiresAtNever = 0 // Never expires
)

// DCRRegistrationRequest represents the RFC 7591 Dynamic Client Registration request.
type DCRRegistrationRequest struct {
	OUID                    string                              `json:"ou_id,omitempty"`
	RedirectURIs            []string                            `json:"redirect_uris"`
	GrantTypes              []oauth2const.GrantType             `json:"grant_types,omitempty"`
	ResponseTypes           []oauth2const.ResponseType          `json:"response_types,omitempty"`
	ClientName              string                              `json:"client_name,omitempty"`
	ClientURI               string                              `json:"client_uri,omitempty"`
	LogoURI                 string                              `json:"logo_uri,omitempty"`
	TokenEndpointAuthMethod oauth2const.TokenEndpointAuthMethod `json:"token_endpoint_auth_method,omitempty"`
	JWKSUri                 string                              `json:"jwks_uri,omitempty"`
	JWKS                    map[string]interface{}              `json:"jwks,omitempty"`
	Scope                   string                              `json:"scope,omitempty"`
	Contacts                []string                            `json:"contacts,omitempty"`
	TosURI                  string                              `json:"tos_uri,omitempty"`
	PolicyURI               string                              `json:"policy_uri,omitempty"`

	// Localized variant maps — populated from #-keyed JSON fields (e.g. "client_name#fr").
	LocalizedClientName map[string]string `json:"-"`
	LocalizedLogoURI    map[string]string `json:"-"`
	LocalizedTosURI     map[string]string `json:"-"`
	LocalizedPolicyURI  map[string]string `json:"-"`
}

// UnmarshalJSON decodes DCRRegistrationRequest from JSON, extracting OIDC language-tagged fields
// (e.g. "client_name#fr") into the localized variant maps.
func (r *DCRRegistrationRequest) UnmarshalJSON(data []byte) error {
	type Alias DCRRegistrationRequest
	if err := json.Unmarshal(data, (*Alias)(r)); err != nil {
		return err
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	for key, val := range raw {
		field, tag, ok := strings.Cut(key, "#")
		if !ok {
			continue
		}
		// Validate the tag eagerly so callers get an error on bad input.
		canonical, valid := i18nmgt.NormaliseBCP47Tag(tag)
		if !valid {
			return &errInvalidBCP47Tag{key: key}
		}
		var s string
		if err := json.Unmarshal(val, &s); err != nil {
			continue
		}
		switch field {
		case "client_name":
			if r.LocalizedClientName == nil {
				r.LocalizedClientName = make(map[string]string)
			}
			r.LocalizedClientName[canonical] = s
		case "logo_uri":
			if r.LocalizedLogoURI == nil {
				r.LocalizedLogoURI = make(map[string]string)
			}
			r.LocalizedLogoURI[canonical] = s
		case "tos_uri":
			if r.LocalizedTosURI == nil {
				r.LocalizedTosURI = make(map[string]string)
			}
			r.LocalizedTosURI[canonical] = s
		case "policy_uri":
			if r.LocalizedPolicyURI == nil {
				r.LocalizedPolicyURI = make(map[string]string)
			}
			r.LocalizedPolicyURI[canonical] = s
		}
	}
	return nil
}

// errInvalidBCP47Tag is returned when a language tag in a DCR request field is not valid BCP 47.
type errInvalidBCP47Tag struct{ key string }

func (e *errInvalidBCP47Tag) Error() string {
	return "invalid BCP 47 language tag in field \"" + e.key + "\""
}

// DCRRegistrationResponse represents the RFC 7591 Dynamic Client Registration response.
type DCRRegistrationResponse struct {
	ClientID                string                              `json:"client_id"`
	ClientSecret            string                              `json:"client_secret,omitempty"`
	ClientSecretExpiresAt   int64                               `json:"client_secret_expires_at"`
	RedirectURIs            []string                            `json:"redirect_uris,omitempty"`
	GrantTypes              []oauth2const.GrantType             `json:"grant_types,omitempty"`
	ResponseTypes           []oauth2const.ResponseType          `json:"response_types,omitempty"`
	ClientName              string                              `json:"client_name,omitempty"`
	ClientURI               string                              `json:"client_uri,omitempty"`
	LogoURI                 string                              `json:"logo_uri,omitempty"`
	TokenEndpointAuthMethod oauth2const.TokenEndpointAuthMethod `json:"token_endpoint_auth_method,omitempty"`
	JWKSUri                 string                              `json:"jwks_uri,omitempty"`
	JWKS                    map[string]interface{}              `json:"jwks,omitempty"`
	Scope                   string                              `json:"scope,omitempty"`
	Contacts                []string                            `json:"contacts,omitempty"`
	TosURI                  string                              `json:"tos_uri,omitempty"`
	PolicyURI               string                              `json:"policy_uri,omitempty"`
	AppID                   string                              `json:"app_id,omitempty"`

	// Localized variant maps — injected as #-keyed top-level fields during serialization.
	LocalizedClientName map[string]string `json:"-"`
	LocalizedLogoURI    map[string]string `json:"-"`
	LocalizedTosURI     map[string]string `json:"-"`
	LocalizedPolicyURI  map[string]string `json:"-"`
}

// MarshalJSON serializes DCRRegistrationResponse to JSON, injecting OIDC language-tagged
// fields (e.g. "client_name#fr") as top-level keys.
func (r DCRRegistrationResponse) MarshalJSON() ([]byte, error) {
	type Alias DCRRegistrationResponse
	base, err := json.Marshal(Alias(r))
	if err != nil {
		return nil, err
	}
	var m map[string]interface{}
	if err := json.Unmarshal(base, &m); err != nil {
		return nil, err
	}
	for tag, val := range r.LocalizedClientName {
		m["client_name#"+tag] = val
	}
	for tag, val := range r.LocalizedLogoURI {
		m["logo_uri#"+tag] = val
	}
	for tag, val := range r.LocalizedTosURI {
		m["tos_uri#"+tag] = val
	}
	for tag, val := range r.LocalizedPolicyURI {
		m["policy_uri#"+tag] = val
	}
	return json.Marshal(m)
}

// DCRErrorResponse represents the RFC 7591 Dynamic Client Registration error response.
type DCRErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
}
