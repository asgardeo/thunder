/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
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

// Package par implements OAuth 2.0 Pushed Authorization Requests (RFC 9126).
package par

import "errors"

var errInvalidRequestURI = errors.New("invalid request_uri format")

var errRequestURINotFound = errors.New("request_uri not found, expired, or already consumed")

// ErrPARResolutionFailed indicates a server-side failure while resolving a PAR request.
var ErrPARResolutionFailed = errors.New("failed to resolve pushed authorization request")

var errClientIDMismatch = errors.New("client_id does not match the pushed authorization request")
