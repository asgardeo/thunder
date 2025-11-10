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

package adapter

// InitializeConsoleAdapter initializes and returns a console output adapter.
// This is the single entry point for creating console adapters, following the
// initialization pattern used throughout Thunder (similar to authn/credentials/init.go).
//
// The console adapter writes formatted events to stdout.
//
// Returns:
//   - OutputAdapterInterface: The initialized console adapter instance
//
// Example:
//
//	adapter := adapter.InitializeConsoleAdapter()
//	err := adapter.Write(formattedData)
func InitializeConsoleAdapter() OutputAdapterInterface {
	return newConsoleAdapter()
}

// InitializeFileAdapter initializes and returns a file output adapter.
// This is the single entry point for creating file adapters, following the
// initialization pattern used throughout Thunder (similar to authn/credentials/init.go).
//
// The file adapter writes formatted events to a file with optional rotation support.
//
// Parameters:
//   - filePath: The path to the file where events will be written
//
// Returns:
//   - OutputAdapterInterface: The initialized file adapter instance
//   - error: Error if the adapter cannot be created (e.g., invalid path, permission issues)
//
// Example:
//
//	adapter, err := adapter.InitializeFileAdapter("/var/log/observability/events.log")
//	if err != nil {
//	    return err
//	}
//	err = adapter.Write(formattedData)
func InitializeFileAdapter(filePath string) (OutputAdapterInterface, error) {
	return NewFileAdapter(filePath)
}
