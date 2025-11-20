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
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"regexp"
	"time"
)

var uuidRegex = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// GenerateUUID returns a UUID string in lowercase hexadecimal, following the canonical textual representation
// as specified in RFC 9562.
func GenerateUUID() string {
	var uuid [16]byte
	_, err := rand.Read(uuid[:])
	if err != nil {
		panic(fmt.Errorf("failed to generate random bytes: %w", err))
	}

	uuid[6] = (uuid[6] & 0x0f) | 0x40 // Version 4
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // Variant is 10

	return fmt.Sprintf("%x-%x-%x-%x-%x",
		uuid[0:4],
		uuid[4:6],
		uuid[6:8],
		uuid[8:10],
		uuid[10:],
	)
}

// GenerateUUIDv7 returns a UUID v7 string (time-ordered) in lowercase hexadecimal.
// UUID v7 features a time-ordered value field derived from the widely implemented
// Unix Epoch timestamp source, providing better database index locality and performance.
func GenerateUUIDv7() string {
	var uuid [16]byte

	// Get current Unix timestamp in milliseconds
	now := time.Now()
	unixMillis := uint64(now.UnixMilli()) // #nosec G115 -- safe conversion for UUID timestamp

	// Set timestamp in first 48 bits (6 bytes)
	binary.BigEndian.PutUint32(uuid[0:4], uint32(unixMillis>>16)) // #nosec G115 -- safe conversion for UUID timestamp
	binary.BigEndian.PutUint16(uuid[4:6], uint16(unixMillis))     // #nosec G115 -- safe conversion for UUID timestamp

	// Fill remaining bytes with random data
	_, err := rand.Read(uuid[6:])
	if err != nil {
		panic(fmt.Errorf("failed to generate random bytes: %w", err))
	}

	// Set version 7 in bits 48-51
	uuid[6] = (uuid[6] & 0x0f) | 0x70 // Version 7

	// Set variant bits to 10 in bits 64-65
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // Variant is 10

	return fmt.Sprintf("%x-%x-%x-%x-%x",
		uuid[0:4],
		uuid[4:6],
		uuid[6:8],
		uuid[8:10],
		uuid[10:],
	)
}

// IsValidUUID checks if the input string is a valid UUID.
func IsValidUUID(input string) bool {
	return uuidRegex.MatchString(input)
}
