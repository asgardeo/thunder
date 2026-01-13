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
import type {Branding} from '../models/branding';
import type {LayoutConfig} from '../models/layout';

/**
 * Extracts layout configuration from branding data with fallback logic.
 *
 * This function analyzes the branding configuration and returns the layout
 * configuration using a priority-based approach. It first checks for a global
 * layout configuration, then falls back to theme-specific layout settings.
 *
 * @param branding - The complete branding configuration object containing preferences
 *
 * @returns The LayoutConfig object if found, or `undefined` if no layout configuration
 *   is available. The returned config may include properties like:
 *   - `type`: Layout type (e.g., 'LEFT_ALIGNED', 'CENTER_ALIGNED')
 *   - Additional layout-specific properties as defined in LayoutConfig
 *
 * @example
 * Basic usage:
 * ```typescript
 * const layout = extractLayoutFromBranding(brandingData);
 * if (layout) {
 *   console.log('Layout type:', layout.type);
 *   const isLeftAligned = layout.type === 'LEFT_ALIGNED';
 * }
 * ```
 *
 * @example
 * Handling layout priorities:
 * ```typescript
 * // Priority order:
 * // 1. Global layout (branding.preferences.layout)
 * // 2. Theme-specific layout (branding.preferences.theme.layout)
 * // 3. undefined (no layout configuration found)
 *
 * const layout = extractLayoutFromBranding(brandingData);
 * // Returns the highest priority layout configuration available
 * ```
 *
 * @public
 */
export default function extractLayoutFromBranding(branding?: Branding): LayoutConfig | undefined {
  if (!branding?.preferences) {
    return undefined;
  }

  // Check for global layout first
  if (branding.preferences.layout) {
    return branding.preferences.layout;
  }

  // Then check theme-specific layout
  if (branding.preferences.theme?.layout) {
    return branding.preferences.theme.layout;
  }

  return undefined;
}
