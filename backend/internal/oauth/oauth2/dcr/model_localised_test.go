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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDCRRegistrationRequest_UnmarshalJSON_LocalisedFields(t *testing.T) {
	t.Run("parses localized client_name variants", func(t *testing.T) {
		raw := `{
			"redirect_uris": ["https://example.com/cb"],
			"client_name": "My App",
			"client_name#fr": "Mon Application",
			"client_name#de": "Meine Anwendung"
		}`
		var req DCRRegistrationRequest
		require.NoError(t, json.Unmarshal([]byte(raw), &req))

		assert.Equal(t, "My App", req.ClientName)
		assert.Equal(t, map[string]string{"fr": "Mon Application", "de": "Meine Anwendung"}, req.LocalisedClientName)
	})

	t.Run("parses localized logo_uri, tos_uri, policy_uri", func(t *testing.T) {
		raw := `{
			"redirect_uris": ["https://example.com/cb"],
			"logo_uri#fr":    "https://example.com/logo-fr.png",
			"tos_uri#fr":     "https://example.com/tos-fr",
			"policy_uri#fr":  "https://example.com/policy-fr"
		}`
		var req DCRRegistrationRequest
		require.NoError(t, json.Unmarshal([]byte(raw), &req))

		assert.Equal(t, map[string]string{"fr": "https://example.com/logo-fr.png"}, req.LocalisedLogoURL)
		assert.Equal(t, map[string]string{"fr": "https://example.com/tos-fr"}, req.LocalisedTosURI)
		assert.Equal(t, map[string]string{"fr": "https://example.com/policy-fr"}, req.LocalisedPolicyURI)
	})

	t.Run("normalises tag to lowercase", func(t *testing.T) {
		raw := `{"redirect_uris": ["https://example.com/cb"], "client_name#FR-CA": "Québec"}`
		var req DCRRegistrationRequest
		require.NoError(t, json.Unmarshal([]byte(raw), &req))

		assert.Equal(t, map[string]string{"fr-ca": "Québec"}, req.LocalisedClientName)
	})

	t.Run("silently ignores tagged variants on non-localisable fields", func(t *testing.T) {
		raw := `{"redirect_uris": ["https://example.com/cb"], "scope#fr": "openid", "client_uri#fr": "https://x.com"}`
		var req DCRRegistrationRequest
		require.NoError(t, json.Unmarshal([]byte(raw), &req))

		assert.Nil(t, req.LocalisedClientName)
		assert.Nil(t, req.LocalisedLogoURL)
	})

	t.Run("rejects double-hash key on localisable field (AC-08)", func(t *testing.T) {
		raw := `{"redirect_uris": ["https://example.com/cb"], "client_name#fr#extra": "value"}`
		var req DCRRegistrationRequest
		err := json.Unmarshal([]byte(raw), &req)
		assert.Error(t, err, "expected error for double-hash key on localisable field")
	})

	t.Run("silently ignores double-hash key on non-localisable field", func(t *testing.T) {
		raw := `{"redirect_uris": ["https://example.com/cb"], "scope#fr#extra": "value"}`
		var req DCRRegistrationRequest
		require.NoError(t, json.Unmarshal([]byte(raw), &req))
		assert.Nil(t, req.LocalisedClientName)
	})

	t.Run("rejects non-string value on localisable field", func(t *testing.T) {
		for _, field := range []string{"client_name#fr", "logo_uri#fr", "tos_uri#fr", "policy_uri#fr"} {
			raw := `{"redirect_uris": ["https://example.com/cb"], "` + field + `": 123}`
			var req DCRRegistrationRequest
			err := json.Unmarshal([]byte(raw), &req)
			assert.Error(t, err, "expected error for non-string value on %s", field)
		}
	})

	t.Run("silently ignores non-string value on non-localisable tagged field", func(t *testing.T) {
		raw := `{"redirect_uris": ["https://example.com/cb"], "scope#fr": 123}`
		var req DCRRegistrationRequest
		require.NoError(t, json.Unmarshal([]byte(raw), &req))
		assert.Nil(t, req.LocalisedClientName)
	})

	t.Run("base fields unmarshalled correctly alongside localized variants", func(t *testing.T) {
		raw := `{
			"redirect_uris": ["https://example.com/cb"],
			"client_name": "Base App",
			"client_name#fr": "App FR",
			"logo_uri": "https://example.com/logo.png"
		}`
		var req DCRRegistrationRequest
		require.NoError(t, json.Unmarshal([]byte(raw), &req))

		assert.Equal(t, "Base App", req.ClientName)
		assert.Equal(t, "https://example.com/logo.png", req.LogoURI)
		assert.Equal(t, map[string]string{"fr": "App FR"}, req.LocalisedClientName)
		assert.Nil(t, req.LocalisedLogoURL)
	})
}

func TestDCRRegistrationResponse_MarshalJSON_LocalisedFields(t *testing.T) {
	t.Run("emits localized fields as field#tag keys", func(t *testing.T) {
		resp := DCRRegistrationResponse{
			ClientID:   "client-123",
			ClientName: "My App",
			LocalisedClientName: map[string]string{
				"fr": "Mon Application",
				"de": "Meine Anwendung",
			},
		}
		data, err := json.Marshal(resp)
		require.NoError(t, err)

		var m map[string]interface{}
		require.NoError(t, json.Unmarshal(data, &m))

		assert.Equal(t, "My App", m["client_name"])
		assert.Equal(t, "Mon Application", m["client_name#fr"])
		assert.Equal(t, "Meine Anwendung", m["client_name#de"])
	})

	t.Run("emits all four localisable fields", func(t *testing.T) {
		resp := DCRRegistrationResponse{
			ClientID:           "client-123",
			LocalisedLogoURL:   map[string]string{"fr": "https://example.com/logo-fr.png"},
			LocalisedTosURI:    map[string]string{"fr": "https://example.com/tos-fr"},
			LocalisedPolicyURI: map[string]string{"fr": "https://example.com/policy-fr"},
		}
		data, err := json.Marshal(resp)
		require.NoError(t, err)

		var m map[string]interface{}
		require.NoError(t, json.Unmarshal(data, &m))

		assert.Equal(t, "https://example.com/logo-fr.png", m["logo_uri#fr"])
		assert.Equal(t, "https://example.com/tos-fr", m["tos_uri#fr"])
		assert.Equal(t, "https://example.com/policy-fr", m["policy_uri#fr"])
	})

	t.Run("no localized fields produces clean output without hash keys", func(t *testing.T) {
		resp := DCRRegistrationResponse{
			ClientID:   "client-123",
			ClientName: "My App",
		}
		data, err := json.Marshal(resp)
		require.NoError(t, err)

		var m map[string]interface{}
		require.NoError(t, json.Unmarshal(data, &m))

		for k := range m {
			assert.NotContains(t, k, "#", "unexpected hash key: %s", k)
		}
	})

	t.Run("localized fields not present in json output as struct fields", func(t *testing.T) {
		resp := DCRRegistrationResponse{
			ClientID:            "client-123",
			LocalisedClientName: map[string]string{"fr": "App FR"},
		}
		data, err := json.Marshal(resp)
		require.NoError(t, err)

		// The map field itself should not appear under its struct tag name
		assert.NotContains(t, string(data), "localized_client_name")
	})
}
