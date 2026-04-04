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

// MatchRedirectURIPattern reports whether incoming matches pattern.
// pattern may contain * (exactly one path segment) or ** (zero or more path segments)
// only in the path component. Scheme, host, and query must match exactly (case-insensitive
// scheme and host per RFC 3986). Returns (false, error) for malformed inputs,
// (false, nil) for no match, (true, nil) for a match.
func MatchRedirectURIPattern(pattern, incoming string) (bool, error) {
	patternURL, err := url.Parse(pattern)
	if err != nil || patternURL.Scheme == "" || patternURL.Host == "" {
		return false, errors.New("invalid pattern URI: missing scheme or host")
	}
	incomingURL, err := url.Parse(incoming)
	if err != nil || incomingURL.Scheme == "" || incomingURL.Host == "" {
		return false, errors.New("invalid incoming URI: missing scheme or host")
	}

	if !strings.EqualFold(patternURL.Scheme, incomingURL.Scheme) {
		return false, nil
	}
	if !strings.EqualFold(patternURL.Host, incomingURL.Host) {
		return false, nil
	}
	if patternURL.RawQuery != incomingURL.RawQuery {
		return false, nil
	}
	if incomingURL.Fragment != "" {
		return false, nil
	}
	return matchPathPattern(patternURL.Path, incomingURL.Path, 0), nil
}

// matchPathPattern reports whether incomingPath matches patternPath.
// Wildcards * (one segment) and ** (zero or more segments) are supported in patternPath.
// depth guards against excessive recursion.
func matchPathPattern(patternPath, incomingPath string, depth int) bool {
	if depth > 32 {
		return false
	}
	patSegs := strings.Split(patternPath, "/")
	incSegs := strings.Split(incomingPath, "/")
	return matchSegs(patSegs, incSegs, 0, 0, depth)
}

func matchSegs(patSegs, incSegs []string, i, j, depth int) bool {
	if depth > 32 {
		return false
	}
	// Both exhausted.
	if i == len(patSegs) && j == len(incSegs) {
		return true
	}
	// Pattern exhausted but incoming still has segments.
	if i == len(patSegs) {
		return false
	}
	// Incoming exhausted but pattern still has segments:
	// only true if all remaining pattern segments are "**".
	if j == len(incSegs) {
		for k := i; k < len(patSegs); k++ {
			if patSegs[k] != "**" {
				return false
			}
		}
		return true
	}

	pSeg := patSegs[i]

	if pSeg == "**" {
		// Try consuming zero incoming segments (advance pattern only).
		if matchSegs(patSegs, incSegs, i+1, j, depth+1) {
			return true
		}
		// Try consuming one incoming segment (keep pattern position).
		return matchSegs(patSegs, incSegs, i, j+1, depth+1)
	}

	if pSeg == "*" {
		// Must match exactly one non-empty segment.
		if incSegs[j] == "" {
			return false
		}
		return matchSegs(patSegs, incSegs, i+1, j+1, depth+1)
	}

	// Literal segment: must match exactly.
	if pSeg != incSegs[j] {
		return false
	}
	return matchSegs(patSegs, incSegs, i+1, j+1, depth+1)
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

// IsValidLogoURI checks if the provided URI is valid for use as a logo URL.
// It enforces a scheme allowlist: http/https require a non-empty host, data/blob/emoji
// schemes are always accepted, and relative paths (no scheme, non-empty path) are accepted.
// All other schemes (e.g. javascript, file) are rejected.
func IsValidLogoURI(uri string) bool {
	if uri == "" {
		return false
	}
	parsed, err := url.Parse(uri)
	if err != nil {
		return false
	}
	switch parsed.Scheme {
	case "http", "https":
		return parsed.Host != ""
	case "data", "blob", "emoji":
		return true
	case "":
		// Accept relative paths (no scheme, but path must start with /)
		return strings.HasPrefix(parsed.Path, "/")
	default:
		return false
	}
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

// DecodeJSONResponse decodes JSON from the response body into any struct type T.
// TODO: Unify DecodeJSONBody and DecodeJSONResponse into a single method.
func DecodeJSONResponse[T any](resp *http.Response) (*T, error) {
	if resp == nil || resp.Body == nil {
		return nil, errors.New("failed to decode JSON response: response or body is nil")
	}
	var data T
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, errors.New("failed to decode JSON response: " + err.Error())
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

// IsBearerAuth checks if the Authorization header uses the Bearer scheme (case-insensitive).
func IsBearerAuth(authHeader string) bool {
	parts := strings.SplitN(authHeader, " ", 2)
	return len(parts) >= 1 && strings.EqualFold(parts[0], constants.TokenTypeBearer)
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

// WriteI18nErrorResponse writes a JSON i18n error response with the given status code and error details.
// This function should be used for services that have been migrated to use I18nErrorResponse.
func WriteI18nErrorResponse(w http.ResponseWriter, statusCode int, errorResp apierror.I18nErrorResponse) {
	logger := log.GetLogger()
	w.Header().Set(constants.ContentTypeHeaderName, constants.ContentTypeJSON)
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(errorResp); err != nil {
		logger.Error("Failed to encode i18n error response", log.Error(err))
		http.Error(w, serviceerror.ErrorEncodingError, http.StatusInternalServerError)
	}
}
