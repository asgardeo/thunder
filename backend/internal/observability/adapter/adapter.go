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

// Package adapter provides output adapter interfaces and implementations.
package adapter

// OutputAdapter is the interface for writing formatted events to various destinations.
type OutputAdapter interface {
	// Write writes formatted event data to the output destination.
	Write(data []byte) error

	// Flush ensures all buffered data is written.
	Flush() error

	// Close closes the adapter and releases resources.
	Close() error

	// GetName returns the name of this adapter.
	GetName() string
}
