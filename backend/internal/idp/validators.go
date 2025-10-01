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

package idp

import (
	"fmt"

	"github.com/asgardeo/thunder/internal/system/cmodels"
)

var googleIDPProperties = map[string]bool{
	"client_id":     true,
	"client_secret": true,
}

func validateIDPProperties(idpType IDPType, properties []cmodels.Property) error {
	switch idpType {
	case IDPTypeGoogle:
		return validateProperties(properties, googleIDPProperties)
	default:
		return fmt.Errorf("unknown idp type: %s", idpType)
	}
}

func validateProperties(properties []cmodels.Property, requiredProperties map[string]bool) error {
	providedProperties := make(map[string]bool)
	for _, p := range properties {
		providedProperties[p.Name] = true
	}

	for key := range requiredProperties {
		if !providedProperties[key] {
			return fmt.Errorf("missing required property: %s", key)
		}
	}

	return nil
}
