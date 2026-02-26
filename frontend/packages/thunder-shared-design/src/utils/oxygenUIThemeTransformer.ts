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

import {extendTheme, type Theme} from '@wso2/oxygen-ui';
import {ColorScheme, ThemeConfig} from '../models/theme';

/**
 * Safely parses a border radius value to a number
 * Falls back to 8px default if parsing fails
 *
 * @param value - Border radius value (number or string like "8px")
 * @returns Parsed number or default value
 */
function parseBorderRadius(value: number | string | undefined): number {
  // If already a number, return it
  if (typeof value === 'number') {
    return value;
  }

  // If not a string or empty, use default
  if (typeof value !== 'string' || value.trim() === '') {
    return 8;
  }

  // Remove 'px' suffix and parse
  const parsed = parseInt(value.replace('px', '').trim(), 10);

  // If parsing resulted in NaN or invalid number, use default
  if (Number.isNaN(parsed) || parsed < 0) {
    return 8;
  }

  return parsed;
}

/**
 * Transforms a Thunder Theme object into an OxygenUI theme configuration using extendTheme
 *
 * @param baseTheme - Base theme to extend from (e.g., OxygenTheme, PaleGrayTheme)
 * @param config - Thunder Theme configuration to apply
 * @returns Extended OxygenUI theme with design colors and styles
 */
export default function oxygenUIThemeTransformer(baseTheme: Theme, config?: ThemeConfig): Theme {
  if (!config) {
    return baseTheme;
  }

  const {colorSchemes, shape, typography, defaultColorScheme, direction} = config;

  // Map ThemeConfig color schemes to OxygenUI palette structure (colors → palette)
  const mappedColorSchemes = Object.fromEntries(
    Object.entries(colorSchemes ?? {})
      .filter((entry): entry is [string, ColorScheme] => entry[1] != null)
      .map(([scheme, value]) => [scheme, {palette: value.colors}]),
  );

  return extendTheme(baseTheme, {
    ...(defaultColorScheme && {defaultColorScheme: defaultColorScheme as 'light' | 'dark'}),
    ...(direction && {direction: direction as 'ltr' | 'rtl'}),
    ...(Object.keys(mappedColorSchemes).length > 0 && {colorSchemes: mappedColorSchemes}),
    ...(shape && {shape: {borderRadius: parseBorderRadius(shape.borderRadius)}}),
    ...(typography && {typography: {fontFamily: typography.fontFamily}}),
  });
}
