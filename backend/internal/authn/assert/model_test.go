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

package assert

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type AssertModelTestSuite struct {
	suite.Suite
}

func TestAssertModelTestSuite(t *testing.T) {
	suite.Run(t, new(AssertModelTestSuite))
}

func (suite *AssertModelTestSuite) TestLevel() {
	testCases := []struct {
		name     string
		level    AssuranceLevel
		expected int
	}{
		{
			name:     "AAL Level 1",
			level:    AALLevel1,
			expected: 1,
		},
		{
			name:     "IAL Level 1",
			level:    IALLevel1,
			expected: 1,
		},
		{
			name:     "AAL Level 2",
			level:    AALLevel2,
			expected: 2,
		},
		{
			name:     "IAL Level 2",
			level:    IALLevel2,
			expected: 2,
		},
		{
			name:     "AAL Level 3",
			level:    AALLevel3,
			expected: 3,
		},
		{
			name:     "IAL Level 3",
			level:    IALLevel3,
			expected: 3,
		},
		{
			name:     "Unknown Level",
			level:    AssuranceLevel("unknown"),
			expected: 0,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			result := tc.level.Level()
			suite.Equal(tc.expected, result)
		})
	}
}
