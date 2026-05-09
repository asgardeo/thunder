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
 * Common query parameters for list API endpoints.
 *
 * @public
 * @remarks
 * Represents optional filtering and pagination values accepted by list APIs,
 * where `limit` controls page size, `offset` controls the starting index, and
 * `filter` contains service-specific filter expressions.
 */
export interface ApiFilteringParams {
  /**
   * Maximum number of resources to return.
   */
  limit?: number;

  /**
   * Number of resources to skip before collecting results.
   */
  offset?: number;

  /**
   * Filter expression used to narrow the result set.
   */
  filter?: string;
}
