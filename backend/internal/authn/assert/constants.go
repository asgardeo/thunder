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

const (
	// AALUnknown represents unknown or no authentication factors.
	AALUnknown AssuranceLevel = "UNKNOWN"
	// AALLevel1 represents basic single-factor authentication.
	AALLevel1 AssuranceLevel = "AAL1"
	// AALLevel2 represents two-factor authentication.
	AALLevel2 AssuranceLevel = "AAL2"
	// AALLevel3 represents multi-factor authentication with hardware token.
	AALLevel3 AssuranceLevel = "AAL3"

	// IALUnknown represents unknown or unverified identity.
	IALUnknown AssuranceLevel = "UNKNOWN"
	// IALLevel1 represents self-asserted identity.
	IALLevel1 AssuranceLevel = "IAL1"
	// IALLevel2 represents identity verified by a trusted party.
	IALLevel2 AssuranceLevel = "IAL2"
	// IALLevel3 represents in-person identity proofing.
	IALLevel3 AssuranceLevel = "IAL3"
)
