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
 * Represents the error response body returned by the API
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
 *   code: 'APP-1020',
 *   message: 'Application already exists',
 *   description: 'An application with the same name already exists'
 * };
 * ```
 */
export interface ApiError {
  /**
   * Machine-readable error code
   * @example 'APP-1020'
   */
  code: string;

  /**
   * Short error message
   * @example 'Application already exists'
   */
  message: string;

  /**
   * Detailed error description
   * @example 'An application with the same name already exists'
   */
  description: string;
}
