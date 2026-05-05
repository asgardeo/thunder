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

package app

// This file declares the minimum method-sets (ports) that the app builder requires from
// infrastructure dependencies that cannot be moved to pkg/ yet.
//
// Current limitation: some port signatures still reference internal/ types (e.g.,
// event.Event, email.EmailData). Once those value types are promoted to pkg/, external
// consumers will be able to provide their own implementations without any internal/ imports.
// Until then, the With* methods on Builder accept the full internal interface types, and
// these ports serve as documentation of the planned narrow contracts.

import (
	"github.com/asgardeo/thunder/internal/system/email"
	"github.com/asgardeo/thunder/internal/system/observability"
	"github.com/asgardeo/thunder/internal/system/observability/event"
)

// ObservabilityServicePort is the minimum observability capability the builder requires.
// Roadmap: once event.Event moves to pkg/, this becomes the With* parameter type and
// external consumers can implement it without importing internal/.
type ObservabilityServicePort interface {
	IsEnabled() bool
	PublishEvent(evt *event.Event)
	Shutdown()
}

// EmailClientPort is the minimum email capability the builder requires.
// Roadmap: once email.EmailData moves to pkg/, this becomes the With* parameter type.
type EmailClientPort interface {
	Send(emailData email.EmailData) error
}

// compile-time checks: verify that the internal implementations satisfy the port interfaces.
var (
	_ ObservabilityServicePort = (observability.ObservabilityServiceInterface)(nil)
	_ EmailClientPort          = (email.EmailClientInterface)(nil)
)
