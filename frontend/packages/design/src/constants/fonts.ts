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
 * Common browser-safe / system fonts that do not require loading from Google Fonts.
 * Used by GoogleFontLoader to skip loading and by the theme builder as autocomplete suggestions.
 */
export const BROWSER_SAFE_FONTS: string[] = [
  'Arial',
  'Arial Black',
  'Brush Script MT',
  'Comic Sans MS',
  'Courier New',
  'Georgia',
  'Helvetica',
  'Impact',
  'Lucida Console',
  'Lucida Sans Unicode',
  'Palatino Linotype',
  'system-ui',
  'Tahoma',
  'Times New Roman',
  'Trebuchet MS',
  'Verdana',
];

/** Lowercase set derived from BROWSER_SAFE_FONTS plus generic CSS font families. */
export const SYSTEM_FONTS = new Set([
  ...BROWSER_SAFE_FONTS.map((f) => f.toLowerCase()),
  'sans-serif',
  'serif',
  'monospace',
  'cursive',
  'fantasy',
]);
