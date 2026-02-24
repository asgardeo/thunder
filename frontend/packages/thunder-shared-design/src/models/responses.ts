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

import type {Theme} from './theme';
import type {LayoutConfig} from './layout';

/**
 * Theme item in list responses (theme data may be null in list view)
 */
export interface ThemeListItem {
  id: string;
  displayName: string;
  description?: string;
  theme: Theme | null;
}

/**
 * Layout item in list responses (layout data may be null in list view)
 */
export interface LayoutListItem {
  id: string;
  displayName: string;
  layout: LayoutConfig | null;
}

/**
 * Pagination link
 */
export interface Link {
  href: string;
  rel: string;
}

/**
 * Response for listing theme configurations
 */
export interface ThemeListResponse {
  totalResults: number;
  startIndex: number;
  count: number;
  themes: ThemeListItem[];
  links: Link[];
}

/**
 * Response for a single theme configuration
 */
export interface ThemeResponse {
  id: string;
  displayName: string;
  description?: string;
  theme: Theme;
}

/**
 * Response for listing layout configurations
 */
export interface LayoutListResponse {
  totalResults: number;
  startIndex: number;
  count: number;
  layouts: LayoutListItem[];
  links: Link[];
}

/**
 * Response for a single layout configuration
 */
export interface LayoutResponse {
  id: string;
  displayName: string;
  layout: LayoutConfig;
}

/**
 * Response from the design resolve endpoint.
 * Returns the merged theme and layout for a given entity.
 */
export interface DesignResolveResponse {
  theme: Theme;
  layout: LayoutConfig;
}
