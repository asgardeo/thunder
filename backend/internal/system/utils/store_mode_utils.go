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
	serverconst "github.com/asgardeo/thunder/internal/system/constants"
)

const (
	// StoreModeComposite represents the composite/hybrid store mode.
	StoreModeComposite = "composite"
)

// IsCompositeModeEnabled checks if the given store mode is composite/hybrid mode.
//
// Parameters:
//   - storeMode: The store mode string (mutable, declarative, or composite)
//
// Returns:
//   - true if storeMode is composite, false otherwise
func IsCompositeModeEnabled(storeMode string) bool {
	return storeMode == StoreModeComposite
}

// ApplyCompositeLimit applies the hard limit for composite store operations.
//
// In composite/hybrid mode, a maximum of 1,000 records can be fetched to prevent
// memory exhaustion when merging results from multiple data sources (database + file-based).
// For larger datasets, users should use search functionality instead.
//
// Parameters:
//   - totalCount: The actual total count of records from both stores
//
// Returns:
//   - effectiveCount: The count capped at MaxCompositeStoreRecords if exceeded
//   - warning: A message to display when the limit is exceeded, empty string otherwise
//   - limitExceeded: true if totalCount exceeds the limit
//
// Example:
//   - totalCount=500 -> effectiveCount=500, warning="", limitExceeded=false
//   - totalCount=1500 -> effectiveCount=1000, warning="Results limited...", limitExceeded=true
func ApplyCompositeLimit(totalCount int) (effectiveCount int, warning string, limitExceeded bool) {
	if totalCount > serverconst.MaxCompositeStoreRecords {
		return serverconst.MaxCompositeStoreRecords,
			serverconst.CompositeStoreLimitWarning,
			true
	}
	return totalCount, "", false
}
