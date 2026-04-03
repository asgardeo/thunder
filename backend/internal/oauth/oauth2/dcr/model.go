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
	"fmt"
	"strings"

	oauth2const "github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	sysutils "github.com/asgardeo/thunder/internal/system/utils"
)

// Default values for DCR
const (
	ClientSecretExpiresAtNever = 0 // Never expires
)

// DCRRegistrationRequest represents the RFC 7591 Dynamic Client Registration request.
type DCRRegistrationRequest struct {
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

	// Localized variant maps — populated by UnmarshalJSON from field#tag keys; not in standard JSON output.
	LocalisedClientName map[string]string `json:"-"`
	LocalisedLogoURL    map[string]string `json:"-"`
	LocalisedTosURI     map[string]string `json:"-"`
	LocalisedPolicyURI  map[string]string `json:"-"`
}

// dcrRegistrationRequestJSON is an alias used by UnmarshalJSON to avoid infinite recursion.
type dcrRegistrationRequestJSON DCRRegistrationRequest

// localisableDCRFields is the set of DCR field names that accept language-tagged variants.
var localisableDCRFields = map[string]bool{
	"client_name": true, "logo_uri": true, "tos_uri": true, "policy_uri": true,
}

// UnmarshalJSON decodes a DCRRegistrationRequest in a single JSON parse of the input bytes.
// It separates language-tagged keys (field#tag convention, OIDC DCR §2) into the localized
// variant maps, then decodes the remaining standard keys into the typed struct.
// Localisable fields with a non-string tagged value are rejected; tagged variants on
// non-localisable fields are silently ignored.
func (r *DCRRegistrationRequest) UnmarshalJSON(data []byte) error {
	// Single parse into a raw key-value map.
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// Separate tagged (#) keys from standard keys.
	clean := make(map[string]json.RawMessage, len(raw))
	for key, val := range raw {
		field, tag, ok := strings.Cut(key, "#")
		if !ok {
			clean[key] = val
			continue
		}
		// Multiple '#' in a key (e.g. "client_name#en#US") are invalid (AC-08).
		// Reject if the base field is localisable; silently skip otherwise.
		if strings.Contains(tag, "#") {
			if localisableDCRFields[field] {
				return fmt.Errorf("invalid localized field key %q: tag must not contain '#'", key)
			}
			continue
		}
		var s string
		if err := json.Unmarshal(val, &s); err != nil {
			// Non-string value on a recognized localisable field is a client error.
			if localisableDCRFields[field] {
				return err
			}
			continue
		}
		normTag := sysutils.NormaliseBCP47Tag(tag)
		switch field {
		case "client_name":
			if r.LocalisedClientName == nil {
				r.LocalisedClientName = make(map[string]string)
			}
			r.LocalisedClientName[normTag] = s
		case "logo_uri":
			if r.LocalisedLogoURL == nil {
				r.LocalisedLogoURL = make(map[string]string)
			}
			r.LocalisedLogoURL[normTag] = s
		case "tos_uri":
			if r.LocalisedTosURI == nil {
				r.LocalisedTosURI = make(map[string]string)
			}
			r.LocalisedTosURI[normTag] = s
		case "policy_uri":
			if r.LocalisedPolicyURI == nil {
				r.LocalisedPolicyURI = make(map[string]string)
			}
			r.LocalisedPolicyURI[normTag] = s
		}
	}

	// Decode standard (non-tagged) keys using the alias to avoid infinite recursion.
	cleanBytes, err := json.Marshal(clean)
	if err != nil {
		return err
	}
	return json.Unmarshal(cleanBytes, (*dcrRegistrationRequestJSON)(r))
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

	// Localized variant maps — emitted as field#tag keys by MarshalJSON; not decoded from standard JSON.
	LocalisedClientName map[string]string `json:"-"`
	LocalisedLogoURL    map[string]string `json:"-"`
	LocalisedTosURI     map[string]string `json:"-"`
	LocalisedPolicyURI  map[string]string `json:"-"`
}

// dcrRegistrationResponseJSON is an alias used by MarshalJSON to avoid infinite recursion.
type dcrRegistrationResponseJSON DCRRegistrationResponse

// MarshalJSON serializes DCRRegistrationResponse, injecting language-tagged variant keys
// (e.g. "client_name#fr", "logo_uri#fr") as top-level JSON properties.
func (r DCRRegistrationResponse) MarshalJSON() ([]byte, error) {
	base, err := json.Marshal(dcrRegistrationResponseJSON(r))
	if err != nil {
		return nil, err
	}

	if len(r.LocalisedClientName) == 0 && len(r.LocalisedLogoURL) == 0 &&
		len(r.LocalisedTosURI) == 0 && len(r.LocalisedPolicyURI) == 0 {
		return base, nil
	}

	var m map[string]interface{}
	if err := json.Unmarshal(base, &m); err != nil {
		return nil, err
	}

	for tag, val := range r.LocalisedClientName {
		m["client_name#"+tag] = val
	}
	for tag, val := range r.LocalisedLogoURL {
		m["logo_uri#"+tag] = val
	}
	for tag, val := range r.LocalisedTosURI {
		m["tos_uri#"+tag] = val
	}
	for tag, val := range r.LocalisedPolicyURI {
		m["policy_uri#"+tag] = val
	}

	return json.Marshal(m)
}

// DCRErrorResponse represents the RFC 7591 Dynamic Client Registration error response.
type DCRErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
}
