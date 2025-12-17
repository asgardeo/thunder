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

package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"html"
	"net/http"
	"net/url"
	"strings"
	"unicode"

	"github.com/asgardeo/thunder/internal/system/constants"
	"github.com/asgardeo/thunder/internal/system/error/apierror"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
)

// WriteJSONError writes a JSON error response with the given details.
func WriteJSONError(w http.ResponseWriter, code, desc string, statusCode int, respHeaders []map[string]string) {
	logger := log.GetLogger()
	logger.Error("Error in HTTP response", log.String("error", code), log.String("description", desc))

	// Set the response headers.
	for _, header := range respHeaders {
		for key, value := range header {
			w.Header().Set(key, value)
		}
	}
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(map[string]string{
		"error":             code,
		"error_description": desc,
	})
	if err != nil {
		logger.Error("Failed to write JSON error response", log.Error(err))
		return
	}
}

// ParseURL parses the given URL string and returns a URL object.
func ParseURL(urlStr string) (*url.URL, error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	return parsedURL, nil
}

// IsValidURI checks if the provided URI is valid.
func IsValidURI(uri string) bool {
	if uri == "" {
		return false
	}
	parsed, err := url.Parse(uri)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return false
	}
	return true
}

// GetURIWithQueryParams constructs a URI with the given query parameters.
func GetURIWithQueryParams(uri string, queryParams map[string]string) (string, error) {
	// Parse the URI.
	parsedURL, err := ParseURL(uri)
	if err != nil {
		return "", errors.New("failed to parse the return URI: " + err.Error())
	}

	// Return the URI if there are no query parameters.
	if len(queryParams) == 0 {
		return parsedURL.String(), nil
	}

	// Add the query parameters to the URI.
	query := parsedURL.Query()
	for key, value := range queryParams {
		query.Add(key, value)
	}
	parsedURL.RawQuery = query.Encode()

	// Return the constructed URI.
	return parsedURL.String(), nil
}

// DecodeJSONBody decodes JSON from the request body into any struct type T.
func DecodeJSONBody[T any](r *http.Request) (*T, error) {
	var data T
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		return nil, errors.New("failed to decode JSON: " + err.Error())
	}
	return &data, nil
}

// SanitizeString trims whitespace, removes control characters, and escapes HTML.
func SanitizeString(input string) string {
	if input == "" {
		return input
	}

	// Trim leading and trailing whitespace
	trimmed := strings.TrimSpace(input)

	// Remove non-printable/control characters (except newline and tab)
	cleaned := strings.Map(func(r rune) rune {
		if unicode.IsControl(r) && r != '\n' && r != '\t' {
			return -1
		}
		return r
	}, trimmed)

	// Escape HTML to prevent XSS
	safe := html.EscapeString(cleaned)

	return safe
}

// SanitizeStringMap sanitizes a map of strings.
// This function trim whitespace, removes control characters, and escapes HTML in each map entry.
func SanitizeStringMap(inputs map[string]string) map[string]string {
	if len(inputs) == 0 {
		return inputs
	}

	sanitized := make(map[string]string, len(inputs))
	for key, value := range inputs {
		sanitized[key] = SanitizeString(value)
	}
	return sanitized
}

// ExtractBearerToken extracts the Bearer token from the Authorization header value.
// It validates that the header is not empty, starts with "Bearer" (case-insensitive),
// and contains a non-empty token. Returns the token and an error if validation fails.
func ExtractBearerToken(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("missing Authorization header")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], constants.TokenTypeBearer) {
		return "", errors.New("invalid Authorization header format. Expected: Bearer <token>")
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", errors.New("missing access token")
	}

	return token, nil
}

// WriteSuccessResponse writes a JSON success response with the given status code and data.
func WriteSuccessResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	logger := log.GetLogger()

	if statusCode == http.StatusNoContent {
		w.WriteHeader(statusCode)
		return
	}

	// Encode to buffer first to ensure encoding succeeds before sending headers
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(data); err != nil {
		logger.Error("Failed to encode response", log.Error(err))
		http.Error(w, serviceerror.ErrorEncodingError, http.StatusInternalServerError)
		return
	}

	// Encoding succeeded, now safe to send headers and write response
	w.Header().Set(constants.ContentTypeHeaderName, constants.ContentTypeJSON)
	w.WriteHeader(statusCode)
	_, _ = w.Write(buf.Bytes())
}

// WriteErrorResponse writes a JSON error response with the given status code and error details.
func WriteErrorResponse(w http.ResponseWriter, statusCode int, errorResp apierror.ErrorResponse) {
	logger := log.GetLogger()

	// Encode to buffer first to ensure encoding succeeds before sending headers
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(errorResp); err != nil {
		logger.Error("Failed to encode error response", log.Error(err))
		http.Error(w, serviceerror.ErrorEncodingError, http.StatusInternalServerError)
		return
	}

	// Encoding succeeded, now safe to send headers and write response
	w.Header().Set(constants.ContentTypeHeaderName, constants.ContentTypeJSON)
	w.WriteHeader(statusCode)
	_, _ = w.Write(buf.Bytes())
}
