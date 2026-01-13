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

// Package http provides a centralized HTTP client service for making outbound HTTP requests.
// This package offers an abstraction over the standard http.Client to centralize HTTP operations:
//
//   - NewHTTPClient() - creates a client with default 30s timeout
//   - NewHTTPClientWithTimeout(duration) - creates a client with custom timeout
//
// Usage examples:
//
//	// Default client
//	client := httpservice.NewHTTPClient()
//
//	// Custom timeout
//	client := httpservice.NewHTTPClientWithTimeout(10 * time.Second)
package http

import (
	"crypto/tls"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/asgardeo/thunder/internal/system/config"
)

// HTTPClientInterface defines the interface for HTTP client operations.
type HTTPClientInterface interface {
	// Do executes an HTTP request and returns an HTTP response.
	Do(req *http.Request) (*http.Response, error)
	// Get issues a GET to the specified URL.
	Get(url string) (*http.Response, error)
	// Head issues a HEAD to the specified URL.
	Head(url string) (*http.Response, error)
	// Post issues a POST to the specified URL.
	Post(url, contentType string, body io.Reader) (*http.Response, error)
	// PostForm issues a POST to the specified URL, with data's keys and values URL-encoded as the request body.
	PostForm(url string, data url.Values) (*http.Response, error)
}

// HTTPClient implements HTTPClientInterface and provides a centralized HTTP client.
type HTTPClient struct {
	client *http.Client
}

// NewHTTPClient creates a new HTTPClient with default 30-second timeout.
// This method provides complete abstraction over http.Client references.
func NewHTTPClient() HTTPClientInterface {
	return NewHTTPClientWithTimeout(30 * time.Second)
}

// NewHTTPClientWithTimeout creates a new HTTPClient with a custom timeout.
// This is a convenience method for creating clients with specific timeouts.
func NewHTTPClientWithTimeout(timeout time.Duration) HTTPClientInterface {
	return &HTTPClient{
		client: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				// #nosec G402 -- Min TLS version is TLS 1.2 or higher based on config
				TLSClientConfig: &tls.Config{
					MinVersion: GetTLSVersion(config.GetThunderRuntime().Config),
				},
			},
		},
	}
}

// Do executes an HTTP request and returns an HTTP response.
func (c *HTTPClient) Do(req *http.Request) (*http.Response, error) {
	return c.client.Do(req)
}

// Get issues a GET to the specified URL.
func (c *HTTPClient) Get(url string) (*http.Response, error) {
	return c.client.Get(url)
}

// Head issues a HEAD to the specified URL.
func (c *HTTPClient) Head(url string) (*http.Response, error) {
	return c.client.Head(url)
}

// Post issues a POST to the specified URL.
func (c *HTTPClient) Post(url, contentType string, body io.Reader) (*http.Response, error) {
	return c.client.Post(url, contentType, body)
}

// PostForm issues a POST to the specified URL, with data's keys and values URL-encoded as the request body.
func (c *HTTPClient) PostForm(url string, data url.Values) (*http.Response, error) {
	return c.client.PostForm(url, data)
}
