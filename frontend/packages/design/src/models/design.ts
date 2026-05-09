/**
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

/**
 * Enumeration of supported design resolution types.
 * Used to specify the type of entity for which design configuration should be resolved.
 */
export const DesignResolveType = {
  /** Application-level design */
  APP: 'APP',

  /** Organizational Unit-level design */
  OU: 'OU',
} as const;

/**
 * Union type representing the possible design resolution types.
 * @example 'APP' | 'OU'
 */
export type DesignResolveType = (typeof DesignResolveType)[keyof typeof DesignResolveType];
