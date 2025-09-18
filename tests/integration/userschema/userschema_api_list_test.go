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

package userschema

import (
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ListUserSchemasTestSuite struct {
	suite.Suite
	client *http.Client
}

func TestListUserSchemasTestSuite(t *testing.T) {
	suite.Run(t, new(ListUserSchemasTestSuite))
}

func (ts *ListUserSchemasTestSuite) SetupSuite() {
	ts.client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
}

// TestListUserSchemas tests GET /user-schemas
func (ts *ListUserSchemasTestSuite) TestListUserSchemas() {
	req, err := http.NewRequest("GET", testServerURL+"/user-schemas", nil)
	if err != nil {
		ts.T().Fatalf("Failed to create request: %v", err)
	}

	resp, err := ts.client.Do(req)
	if err != nil {
		ts.T().Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	ts.Assert().Equal(http.StatusOK, resp.StatusCode, "Should return 200 OK")

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		ts.T().Fatalf("Failed to read response body: %v", err)
	}

	var listResponse UserSchemaListResponse
	err = json.Unmarshal(bodyBytes, &listResponse)
	if err != nil {
		ts.T().Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify response structure according to API spec
	ts.Assert().GreaterOrEqual(listResponse.TotalResults, 0, "TotalResults should be non-negative")
	ts.Assert().Equal(listResponse.Count, len(listResponse.Schemas), "Count should match actual schemas")
	ts.Assert().Equal(1, listResponse.StartIndex, "StartIndex should be 1 (1-based)")
	ts.Assert().NotNil(listResponse.Links, "Should have links array")

	// Verify each schema has required fields for list view
	for _, schema := range listResponse.Schemas {
		ts.Assert().NotEmpty(schema.ID, "Schema should have ID")
		ts.Assert().NotEmpty(schema.Name, "Schema should have name")
	}
}

// TestListUserSchemasWithPagination tests GET /user-schemas with pagination parameters
func (ts *ListUserSchemasTestSuite) TestListUserSchemasWithPagination() {
	// Test with limit parameter
	params := url.Values{}
	params.Add("limit", "5")
	params.Add("offset", "0")

	req, err := http.NewRequest("GET", testServerURL+"/user-schemas?"+params.Encode(), nil)
	if err != nil {
		ts.T().Fatalf("Failed to create request: %v", err)
	}

	resp, err := ts.client.Do(req)
	if err != nil {
		ts.T().Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	ts.Assert().Equal(http.StatusOK, resp.StatusCode, "Should return 200 OK")

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		ts.T().Fatalf("Failed to read response body: %v", err)
	}

	var listResponse UserSchemaListResponse
	err = json.Unmarshal(bodyBytes, &listResponse)
	if err != nil {
		ts.T().Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify pagination
	ts.Assert().LessOrEqual(listResponse.Count, 5, "Should return at most 5 schemas")
	ts.Assert().Equal(1, listResponse.StartIndex, "Start index should be 1 (1-based)")
}

// TestListUserSchemasWithInvalidPagination tests GET /user-schemas with invalid pagination
func (ts *ListUserSchemasTestSuite) TestListUserSchemasWithInvalidPagination() {
	testCases := []struct {
		name   string
		params map[string]string
	}{
		{
			name: "negative limit",
			params: map[string]string{
				"limit": "-1",
			},
		},
		{
			name: "negative offset",
			params: map[string]string{
				"offset": "-5",
			},
		},
		{
			name: "non-numeric limit",
			params: map[string]string{
				"limit": "abc",
			},
		},
		{
			name: "non-numeric offset",
			params: map[string]string{
				"offset": "xyz",
			},
		},
	}

	for _, tc := range testCases {
		ts.T().Run(tc.name, func(t *testing.T) {
			params := url.Values{}
			for key, value := range tc.params {
				params.Add(key, value)
			}

			req, err := http.NewRequest("GET", testServerURL+"/user-schemas?"+params.Encode(), nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			resp, err := ts.client.Do(req)
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}
			defer resp.Body.Close()

			// Should handle invalid parameters gracefully (either 400 or use defaults)
			ts.Assert().True(resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusOK,
				"Should handle invalid pagination parameters appropriately for case: %s", tc.name)
		})
	}
}

// TestListUserSchemasPaginationLinks tests pagination links in response
func (ts *ListUserSchemasTestSuite) TestListUserSchemasPaginationLinks() {
	// Test with small limit to force pagination
	params := url.Values{}
	params.Add("limit", "1")
	params.Add("offset", "0")

	req, err := http.NewRequest("GET", testServerURL+"/user-schemas?"+params.Encode(), nil)
	if err != nil {
		ts.T().Fatalf("Failed to create request: %v", err)
	}

	resp, err := ts.client.Do(req)
	if err != nil {
		ts.T().Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	ts.Assert().Equal(http.StatusOK, resp.StatusCode, "Should return 200 OK")

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		ts.T().Fatalf("Failed to read response body: %v", err)
	}

	var listResponse UserSchemaListResponse
	err = json.Unmarshal(bodyBytes, &listResponse)
	if err != nil {
		ts.T().Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify links structure
	ts.Assert().NotNil(listResponse.Links, "Should have links array")

	// Check if we have pagination links when more results available
	if listResponse.TotalResults > 1 {
		hasNextLink := false
		for _, link := range listResponse.Links {
			if link.Rel == "next" {
				hasNextLink = true
				ts.Assert().Contains(link.Href, "offset=1", "Next link should have correct offset")
				ts.Assert().Contains(link.Href, "limit=1", "Next link should have correct limit")
				break
			}
		}
		ts.Assert().True(hasNextLink, "Should have next link when more results available")
	}
}
