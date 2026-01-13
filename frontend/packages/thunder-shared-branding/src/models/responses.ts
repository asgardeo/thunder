/**
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
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

import type {Branding} from './branding';

/**
 * Basic branding information returned in list operations
 */
export interface BrandingListItem {
  /**
   * Unique identifier for the branding configuration
   * @example "3fa85f64-5717-4562-b3fc-2c963f66afa6"
   */
  id: string;

  /**
   * Display name for the branding configuration
   * @example "Application 1 Branding"
   */
  displayName: string;
}

/**
 * Pagination link for navigating through result pages
 */
export interface Link {
  /**
   * URL for the next/previous page
   * @example "branding?offset=20&limit=10"
   */
  href: string;

  /**
   * Relationship type (e.g., "next", "previous")
   * @example "next"
   */
  rel: string;
}

/**
 * Response for listing branding configurations
 */
export interface BrandingListResponse {
  /**
   * Total number of branding configurations that match the query
   * @example 8
   */
  totalResults: number;

  /**
   * Index of the first element of the page (offset + 1)
   * @example 1
   */
  startIndex: number;

  /**
   * Number of elements in the returned page
   * @example 5
   */
  count: number;

  /**
   * Array of branding configurations
   */
  brandings: BrandingListItem[];

  /**
   * Pagination links for navigating to next/previous pages
   */
  links?: Link[];
}

/**
 * Response for a single branding configuration
 * This is the same as the Branding model
 */
export type BrandingResponse = Branding;
