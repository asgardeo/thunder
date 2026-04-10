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
)

// DCRRegistrationRequest represents the RFC 7591 Dynamic Client Registration request.
type DCRRegistrationRequest struct {
	OUID                    string                 `json:"ou_id,omitempty"`
	RedirectURIs            []string               `json:"redirect_uris"`
	GrantTypes              []string               `json:"grant_types,omitempty"`
	ResponseTypes           []string               `json:"response_types,omitempty"`
	ClientName              string                 `json:"client_name,omitempty"`
	ClientURI               string                 `json:"client_uri,omitempty"`
	LogoURI                 string                 `json:"logo_uri,omitempty"`
	TokenEndpointAuthMethod string                 `json:"token_endpoint_auth_method,omitempty"`
	JWKSUri                 string                 `json:"jwks_uri,omitempty"`
	JWKS                    map[string]interface{} `json:"jwks,omitempty"`
	Scope                   string                 `json:"scope,omitempty"`
	Contacts                []string               `json:"contacts,omitempty"`
	TosURI                  string                 `json:"tos_uri,omitempty"`
	PolicyURI               string                 `json:"policy_uri,omitempty"`

	// LocalizedFields holds OIDC language-tagged fields (e.g. "client_name#fr": "Mon Client").
	// These are serialized as top-level #-keyed JSON keys.
	LocalizedFields map[string]string `json:"-"`
}

// MarshalJSON serializes DCRRegistrationRequest, injecting LocalizedFields as top-level #-keyed keys.
func (r DCRRegistrationRequest) MarshalJSON() ([]byte, error) {
	type Alias DCRRegistrationRequest
	base, err := json.Marshal(Alias(r))
	if err != nil {
		return nil, err
	}
	if len(r.LocalizedFields) == 0 {
		return base, nil
	}
	var m map[string]interface{}
	if err := json.Unmarshal(base, &m); err != nil {
		return nil, err
	}
	for k, v := range r.LocalizedFields {
		m[k] = v
	}
	return json.Marshal(m)
}

// DCRRegistrationResponse represents the RFC 7591 Dynamic Client Registration response.
type DCRRegistrationResponse struct {
	ClientID                string                 `json:"client_id"`
	ClientSecret            string                 `json:"client_secret,omitempty"`
	ClientSecretExpiresAt   int64                  `json:"client_secret_expires_at"`
	RedirectURIs            []string               `json:"redirect_uris,omitempty"`
	GrantTypes              []string               `json:"grant_types,omitempty"`
	ResponseTypes           []string               `json:"response_types,omitempty"`
	ClientName              string                 `json:"client_name,omitempty"`
	ClientURI               string                 `json:"client_uri,omitempty"`
	LogoURI                 string                 `json:"logo_uri,omitempty"`
	TokenEndpointAuthMethod string                 `json:"token_endpoint_auth_method,omitempty"`
	JWKSUri                 string                 `json:"jwks_uri,omitempty"`
	JWKS                    map[string]interface{} `json:"jwks,omitempty"`
	Scope                   string                 `json:"scope,omitempty"`
	Contacts                []string               `json:"contacts,omitempty"`
	TosURI                  string                 `json:"tos_uri,omitempty"`
	PolicyURI               string                 `json:"policy_uri,omitempty"`
	AppID                   string                 `json:"app_id,omitempty"`

	// Localized variant maps — populated from #-keyed top-level fields in the response JSON.
	LocalizedClientName map[string]string `json:"-"`
	LocalizedLogoURI    map[string]string `json:"-"`
	LocalizedTosURI     map[string]string `json:"-"`
	LocalizedPolicyURI  map[string]string `json:"-"`
}

// UnmarshalJSON decodes DCRRegistrationResponse, extracting #-keyed localized fields.
func (r *DCRRegistrationResponse) UnmarshalJSON(data []byte) error {
	type Alias DCRRegistrationResponse
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
		var s string
		if err := json.Unmarshal(val, &s); err != nil {
			continue
		}
		switch field {
		case "client_name":
			if r.LocalizedClientName == nil {
				r.LocalizedClientName = make(map[string]string)
			}
			r.LocalizedClientName[tag] = s
		case "logo_uri":
			if r.LocalizedLogoURI == nil {
				r.LocalizedLogoURI = make(map[string]string)
			}
			r.LocalizedLogoURI[tag] = s
		case "tos_uri":
			if r.LocalizedTosURI == nil {
				r.LocalizedTosURI = make(map[string]string)
			}
			r.LocalizedTosURI[tag] = s
		case "policy_uri":
			if r.LocalizedPolicyURI == nil {
				r.LocalizedPolicyURI = make(map[string]string)
			}
			r.LocalizedPolicyURI[tag] = s
		}
	}
	return nil
}

// DCRErrorResponse represents the RFC 7591 Dynamic Client Registration error response.
type DCRErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
}
