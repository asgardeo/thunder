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
 * Standard API Error Structure
 *
 * Represents the error response body returned by the Thunder API
 * when an operation fails.
 *
 * @public
 * @remarks
 * This structure is used for error handling in API responses.
 * The `code` field contains a machine-readable error identifier,
 * while `message` and `description` provide human-readable details.
 *
 * @example
 * ```typescript
 * const error: ApiError = {
 *   code: 'OU-60001',
 *   message: 'Organization unit not found',
 *   description: 'No organization unit exists with the given ID'
 * };
 * ```
 */
export interface ApiError {
  /**
   * Machine-readable error code
   * @example 'OU-60001'
   */
  code: string;

  /**
   * Short error message
   * @example 'Organization unit not found'
   */
  message: string;

  /**
   * Detailed error description
   * @example 'No organization unit exists with the given ID'
   */
  description: string;
}
