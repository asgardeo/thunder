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
	"testing"

	"github.com/stretchr/testify/assert"

	serverconst "github.com/asgardeo/thunder/internal/system/constants"
)

func TestIsCompositeModeEnabled(t *testing.T) {
	t.Run("returns true for composite mode", func(t *testing.T) {
		result := IsCompositeModeEnabled(StoreModeComposite)
		assert.True(t, result)
	})

	t.Run("returns false for mutable mode", func(t *testing.T) {
		result := IsCompositeModeEnabled("mutable")
		assert.False(t, result)
	})

	t.Run("returns false for declarative mode", func(t *testing.T) {
		result := IsCompositeModeEnabled("declarative")
		assert.False(t, result)
	})

	t.Run("returns false for empty string", func(t *testing.T) {
		result := IsCompositeModeEnabled("")
		assert.False(t, result)
	})

	t.Run("returns false for invalid mode", func(t *testing.T) {
		result := IsCompositeModeEnabled("invalid")
		assert.False(t, result)
	})
}

func TestApplyCompositeLimit(t *testing.T) {
	t.Run("does not limit when totalCount is below max", func(t *testing.T) {
		totalCount := 500
		effectiveCount, warning, limitExceeded := ApplyCompositeLimit(totalCount)

		assert.Equal(t, 500, effectiveCount)
		assert.Empty(t, warning)
		assert.False(t, limitExceeded)
	})

	t.Run("does not limit when totalCount equals max", func(t *testing.T) {
		totalCount := serverconst.MaxCompositeStoreRecords
		effectiveCount, warning, limitExceeded := ApplyCompositeLimit(totalCount)

		assert.Equal(t, serverconst.MaxCompositeStoreRecords, effectiveCount)
		assert.Empty(t, warning)
		assert.False(t, limitExceeded)
	})

	t.Run("limits when totalCount exceeds max", func(t *testing.T) {
		totalCount := 1500
		effectiveCount, warning, limitExceeded := ApplyCompositeLimit(totalCount)

		assert.Equal(t, serverconst.MaxCompositeStoreRecords, effectiveCount)
		assert.Equal(t, serverconst.CompositeStoreLimitWarning, warning)
		assert.True(t, limitExceeded)
	})

	t.Run("limits when totalCount is exactly max+1", func(t *testing.T) {
		totalCount := serverconst.MaxCompositeStoreRecords + 1
		effectiveCount, warning, limitExceeded := ApplyCompositeLimit(totalCount)

		assert.Equal(t, serverconst.MaxCompositeStoreRecords, effectiveCount)
		assert.Equal(t, serverconst.CompositeStoreLimitWarning, warning)
		assert.True(t, limitExceeded)
	})

	t.Run("handles zero totalCount", func(t *testing.T) {
		totalCount := 0
		effectiveCount, warning, limitExceeded := ApplyCompositeLimit(totalCount)

		assert.Equal(t, 0, effectiveCount)
		assert.Empty(t, warning)
		assert.False(t, limitExceeded)
	})

	t.Run("handles very large totalCount", func(t *testing.T) {
		totalCount := 10000
		effectiveCount, warning, limitExceeded := ApplyCompositeLimit(totalCount)

		assert.Equal(t, serverconst.MaxCompositeStoreRecords, effectiveCount)
		assert.Equal(t, serverconst.CompositeStoreLimitWarning, warning)
		assert.True(t, limitExceeded)
	})
}
