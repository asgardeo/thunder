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
import type {ImageConfig} from '../models/theme';

/**
 * Extracts image configuration from branding data based on the active theme.
 *
 * This function analyzes the branding configuration and returns the complete
 * ImageConfig structure for the currently active color scheme (light or dark).
 * It provides access to logos, background images, and custom icons configured
 * for the active theme.
 *
 * @param branding - The complete branding configuration object containing theme preferences
 *
 * @returns The ImageConfig for the active color scheme, containing:
 *   - `logo`: LogoCollection with primary logo and favicon configurations
 *   - `background`: Optional background images including login page backgrounds
 *   - `icons`: Optional custom icons configuration
 *   Returns `undefined` if no branding data or theme is configured
 *
 * @example
 * Basic usage:
 * ```typescript
 * const images = extractImagesFromBranding(brandingData);
 * if (images) {
 *   const logoUrl = images.logo.primary.url;
 *   const faviconUrl = images.logo.favicon.url;
 *   const loginBg = images.background?.login?.url;
 * }
 * ```
 *
 * @example
 * Handling different color schemes:
 * ```typescript
 * // The function automatically uses the active color scheme
 * const images = extractImagesFromBranding(brandingData);
 * // If activeColorScheme is 'dark', returns dark theme images
 * // If activeColorScheme is 'light' or undefined, returns light theme images
 * ```
 *
 * @public
 */
export default function extractImagesFromBranding(branding?: Branding): ImageConfig | undefined {
  if (!branding?.preferences?.theme) {
    return undefined;
  }

  const {theme} = branding.preferences;

  // Determine the active color scheme
  const activeColorScheme = theme.activeColorScheme || 'light';

  // Get the images for the active color scheme
  const activeImages = theme.colorSchemes?.[activeColorScheme]?.images;

  return activeImages;
}
