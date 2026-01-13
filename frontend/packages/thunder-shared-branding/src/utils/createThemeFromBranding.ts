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

import {OxygenTheme, extendTheme} from '@wso2/oxygen-ui';
import type {Theme} from '@wso2/oxygen-ui';
import type {Branding} from '../models/branding';

/* eslint-disable @typescript-eslint/no-unsafe-assignment */
/* eslint-disable @typescript-eslint/no-unsafe-member-access */
/* eslint-disable @typescript-eslint/no-explicit-any */

/**
 * Transforms branding data into an OxygenUI theme configuration
 *
 * @param branding - The branding configuration data
 * @returns Extended OxygenUI theme with branding colors and styles
 */
export default function createThemeFromBranding(branding?: Branding): Theme {
  if (!branding?.preferences?.theme) {
    return OxygenTheme;
  }

  const {theme} = branding.preferences;
  const {colorSchemes} = theme;

  // Extract colors from light and dark schemes
  const lightColors = colorSchemes?.light?.colors;
  const darkColors = colorSchemes?.dark?.colors;

  // Build color schemes for the theme
  const themeColorSchemes: any = {
    ...(OxygenTheme as any).colorSchemes,
  };

  if (lightColors) {
    themeColorSchemes.light = {
      palette: {
        ...((OxygenTheme as any).colorSchemes?.light?.palette ?? {}),
        primary: {
          main: lightColors.primary.main,
          contrastText: lightColors.primary.contrastText,
          ...(lightColors.primary.dark && {dark: lightColors.primary.dark}),
        },
        ...(lightColors.secondary && {
          secondary: {
            main: lightColors.secondary.main,
            contrastText: lightColors.secondary.contrastText,
            ...(lightColors.secondary.dark && {dark: lightColors.secondary.dark}),
          },
        }),
        ...(lightColors.tertiary && {
          tertiary: {
            main: lightColors.tertiary.main,
            contrastText: lightColors.tertiary.contrastText,
            ...(lightColors.tertiary.dark && {dark: lightColors.tertiary.dark}),
          },
        }),
      },
    };
  }

  if (darkColors) {
    themeColorSchemes.dark = {
      palette: {
        ...((OxygenTheme as any).colorSchemes?.dark?.palette ?? {}),
        primary: {
          main: darkColors.primary.main,
          contrastText: darkColors.primary.contrastText,
          ...(darkColors.primary.dark && {dark: darkColors.primary.dark}),
        },
        ...(darkColors.secondary && {
          secondary: {
            main: darkColors.secondary.main,
            contrastText: darkColors.secondary.contrastText,
            ...(darkColors.secondary.dark && {dark: darkColors.secondary.dark}),
          },
        }),
        ...(darkColors.tertiary && {
          tertiary: {
            main: darkColors.tertiary.main,
            contrastText: darkColors.tertiary.contrastText,
            ...(darkColors.tertiary.dark && {dark: darkColors.tertiary.dark}),
          },
        }),
      },
    };
  }

  // Create static colors for primary theme
  const primaryColor = lightColors?.primary?.main ?? darkColors?.primary?.main;

  return extendTheme({
    ...OxygenTheme,
    colorSchemes: themeColorSchemes,
    components: {
      ...OxygenTheme.components,
      ...(primaryColor && {
        MuiButton: {
          styleOverrides: {
            ...(OxygenTheme.components?.MuiButton?.styleOverrides ?? {}),
            containedPrimary: {
              '&:not(:disabled)': {
                backgroundColor: primaryColor,
                color: (lightColors?.primary?.contrastText ?? darkColors?.primary?.contrastText) ?? '#fff',
                '&:hover': {
                  backgroundColor: (lightColors?.primary?.dark ?? darkColors?.primary?.dark) ?? primaryColor,
                  color: (lightColors?.primary?.contrastText ?? darkColors?.primary?.contrastText) ?? '#fff',
                },
              },
            },
          },
        },
      }),
    },
  });
}
