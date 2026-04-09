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

package model

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplicationRequest_UnmarshalJSON_LocalisedFields(t *testing.T) {
	t.Run("parses localized name variants", func(t *testing.T) {
		raw := `{
			"name": "My App",
			"name#fr": "Mon Application",
			"name#de": "Meine Anwendung"
		}`
		var req ApplicationRequest
		require.NoError(t, json.Unmarshal([]byte(raw), &req))

		assert.Equal(t, "My App", req.Name)
		assert.Equal(t, map[string]string{"fr": "Mon Application", "de": "Meine Anwendung"}, req.LocalisedClientName)
	})

	t.Run("parses localized logoUrl, tosUri, policyUri", func(t *testing.T) {
		raw := `{
			"name": "App",
			"logoUrl#fr":   "https://example.com/logo-fr.png",
			"tosUri#fr":    "https://example.com/tos-fr",
			"policyUri#fr": "https://example.com/policy-fr"
		}`
		var req ApplicationRequest
		require.NoError(t, json.Unmarshal([]byte(raw), &req))

		assert.Equal(t, map[string]string{"fr": "https://example.com/logo-fr.png"}, req.LocalisedLogoURL)
		assert.Equal(t, map[string]string{"fr": "https://example.com/tos-fr"}, req.LocalisedTosURI)
		assert.Equal(t, map[string]string{"fr": "https://example.com/policy-fr"}, req.LocalisedPolicyURI)
	})

	t.Run("normalises tag to lowercase", func(t *testing.T) {
		raw := `{"name": "App", "name#FR-CA": "Québec"}`
		var req ApplicationRequest
		require.NoError(t, json.Unmarshal([]byte(raw), &req))

		assert.Equal(t, map[string]string{"fr-ca": "Québec"}, req.LocalisedClientName)
	})

	t.Run("silently ignores tagged variants on non-localisable fields", func(t *testing.T) {
		raw := `{"name": "App", "description#fr": "Un app", "authFlowId#fr": "flow-1"}`
		var req ApplicationRequest
		require.NoError(t, json.Unmarshal([]byte(raw), &req))

		assert.Nil(t, req.LocalisedClientName)
	})

	t.Run("rejects double-hash key on localisable field (AC-08)", func(t *testing.T) {
		raw := `{"name": "App", "name#fr#extra": "value"}`
		var req ApplicationRequest
		err := json.Unmarshal([]byte(raw), &req)
		assert.Error(t, err, "expected error for double-hash key on localisable field")
	})

	t.Run("silently ignores double-hash key on non-localisable field", func(t *testing.T) {
		raw := `{"name": "App", "description#fr#extra": "value"}`
		var req ApplicationRequest
		require.NoError(t, json.Unmarshal([]byte(raw), &req))
		assert.Nil(t, req.LocalisedClientName)
	})

	t.Run("rejects non-string value on localisable field", func(t *testing.T) {
		for _, field := range []string{"name#fr", "logoUrl#fr", "tosUri#fr", "policyUri#fr"} {
			raw := `{"name": "App", "` + field + `": 123}`
			var req ApplicationRequest
			err := json.Unmarshal([]byte(raw), &req)
			assert.Error(t, err, "expected error for non-string value on %s", field)
		}
	})

	t.Run("silently ignores non-string value on non-localisable tagged field", func(t *testing.T) {
		raw := `{"name": "App", "description#fr": 123}`
		var req ApplicationRequest
		require.NoError(t, json.Unmarshal([]byte(raw), &req))
		assert.Nil(t, req.LocalisedClientName)
	})

	t.Run("base fields unmarshalled alongside localized variants", func(t *testing.T) {
		raw := `{"name": "Base", "name#fr": "Base FR", "logoUrl": "https://example.com/logo.png"}`
		var req ApplicationRequest
		require.NoError(t, json.Unmarshal([]byte(raw), &req))

		assert.Equal(t, "Base", req.Name)
		assert.Equal(t, "https://example.com/logo.png", req.LogoURL)
		assert.Equal(t, map[string]string{"fr": "Base FR"}, req.LocalisedClientName)
	})
}

func TestApplicationCompleteResponse_MarshalJSON_LocalisedFields(t *testing.T) {
	t.Run("emits localized fields as field#tag keys", func(t *testing.T) {
		resp := ApplicationCompleteResponse{
			ID:   "app-1",
			Name: "My App",
			LocalisedClientName: map[string]string{
				"fr": "Mon Application",
				"de": "Meine Anwendung",
			},
		}
		data, err := json.Marshal(resp)
		require.NoError(t, err)

		var m map[string]interface{}
		require.NoError(t, json.Unmarshal(data, &m))

		assert.Equal(t, "My App", m["name"])
		assert.Equal(t, "Mon Application", m["name#fr"])
		assert.Equal(t, "Meine Anwendung", m["name#de"])
	})

	t.Run("emits all four localisable fields", func(t *testing.T) {
		resp := ApplicationCompleteResponse{
			ID:                 "app-1",
			LocalisedLogoURL:   map[string]string{"fr": "https://example.com/logo-fr.png"},
			LocalisedTosURI:    map[string]string{"fr": "https://example.com/tos-fr"},
			LocalisedPolicyURI: map[string]string{"fr": "https://example.com/policy-fr"},
		}
		data, err := json.Marshal(resp)
		require.NoError(t, err)

		var m map[string]interface{}
		require.NoError(t, json.Unmarshal(data, &m))

		assert.Equal(t, "https://example.com/logo-fr.png", m["logoUrl#fr"])
		assert.Equal(t, "https://example.com/tos-fr", m["tosUri#fr"])
		assert.Equal(t, "https://example.com/policy-fr", m["policyUri#fr"])
	})

	t.Run("no localized fields produces clean output", func(t *testing.T) {
		resp := ApplicationCompleteResponse{ID: "app-1", Name: "App"}
		data, err := json.Marshal(resp)
		require.NoError(t, err)

		var m map[string]interface{}
		require.NoError(t, json.Unmarshal(data, &m))

		for k := range m {
			assert.NotContains(t, k, "#", "unexpected hash key: %s", k)
		}
	})
}

func TestApplicationGetResponse_MarshalJSON_LocalisedFields(t *testing.T) {
	t.Run("emits localized name as name#tag", func(t *testing.T) {
		resp := ApplicationGetResponse{
			ID:                  "app-1",
			Name:                "My App",
			LocalisedClientName: map[string]string{"fr": "Mon Application"},
		}
		data, err := json.Marshal(resp)
		require.NoError(t, err)

		var m map[string]interface{}
		require.NoError(t, json.Unmarshal(data, &m))

		assert.Equal(t, "My App", m["name"])
		assert.Equal(t, "Mon Application", m["name#fr"])
	})
}
