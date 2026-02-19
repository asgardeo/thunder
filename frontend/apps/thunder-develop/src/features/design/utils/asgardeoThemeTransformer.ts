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

import type {ThemeProviderProps} from '@asgardeo/react';
import type {ThemeConfig} from '@thunder/shared-design';

type AsgardeoTheme = NonNullable<ThemeProviderProps['theme']>;

/**
 * Normalises a border-radius value to a CSS string.
 * Accepts a number (treated as px) or a string (used as-is).
 */
function normaliseBorderRadius(value: string | number | undefined): string | undefined {
  if (value === undefined || value === null) return undefined;
  if (typeof value === 'number') return `${value}px`;
  return value;
}

/**
 * Transforms a Thunder `ThemeConfig` (backend model) into the
 * `RecursivePartial<ThemeConfig>` structure expected by Asgardeo's `ThemeProvider`.
 *
 * The two schemas differ significantly:
 * - Thunder: `colorSchemes.light/dark`, `shape.borderRadius`, `typography.fontFamily`
 * - Asgardeo: flat `colors`, `borderRadius.{small,medium,large}`, `typography.fontFamily`
 *
 * @param thunderTheme - Thunder theme configuration from the API
 * @param activeScheme - Override which color scheme to use for the colors mapping.
 *   Defaults to the theme's `defaultColorScheme` when not provided.
 * @returns Partial Asgardeo theme config suitable for passing to `<ThemeProvider theme={...} />`
 */
export default function toAsgardeoTheme(thunderTheme: ThemeConfig, activeScheme?: 'light' | 'dark'): AsgardeoTheme {
  const activeSchemeKey = activeScheme ?? (thunderTheme.defaultColorScheme === 'dark' ? 'dark' : 'light');
  const colors = thunderTheme.colorSchemes?.[activeSchemeKey]?.colors;

  const br = normaliseBorderRadius(thunderTheme.shape?.borderRadius);

  const result: AsgardeoTheme = {};

  if (thunderTheme.direction === 'ltr' || thunderTheme.direction === 'rtl') {
    result.direction = thunderTheme.direction;
  }

  if (colors) {
    result.colors = {
      primary: {
        main: colors.primary.main,
        contrastText: colors.primary.contrastText,
        ...(colors.primary.dark && {dark: colors.primary.dark}),
      },
      secondary: {
        main: colors.secondary.main,
        contrastText: colors.secondary.contrastText,
        ...(colors.secondary.dark && {dark: colors.secondary.dark}),
      },
      ...(colors.background && {
        background: {
          body: {main: colors.background.default},
          surface: colors.background.paper,
        },
      }),
      ...(colors.text && {
        text: {
          primary: colors.text.primary,
          secondary: colors.text.secondary,
        },
      }),
    };
  }

  if (br) {
    result.borderRadius = {
      small: br,
      medium: br,
      large: br,
    };
  }

  if (thunderTheme.typography?.fontFamily) {
    result.typography = {fontFamily: thunderTheme.typography.fontFamily};
  }

  return result;
}
