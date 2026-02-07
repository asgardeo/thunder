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

/**
 * Primary/Secondary color configuration
 */
export interface BrandColorConfig {
  main: string;
  dark: string;
  contrastText: string;
}

/**
 * Background color configuration
 */
export interface BackgroundColorConfig {
  default: string;
  paper: string;
}

/**
 * Text color configuration
 */
export interface TextColorConfig {
  primary: string;
  secondary: string;
}

/**
 * Colors configuration
 */
export interface ColorsConfig {
  primary: BrandColorConfig;
  secondary: BrandColorConfig;
  background: BackgroundColorConfig;
  text: TextColorConfig;
}

/**
 * Color scheme configuration
 */
export interface ColorScheme {
  colors: ColorsConfig;
}

/**
 * Typography configuration
 */
export interface TypographyConfig {
  fontFamily: string;
}

/**
 * Shape configuration
 */
export interface ShapeConfig {
  borderRadius: string;
}

/**
 * Theme configuration
 */
export interface ThemeConfig {
  direction: 'ltr' | 'rtl';
  defaultColorScheme: 'light' | 'dark';
  colorSchemes: {
    light: ColorScheme;
    dark: ColorScheme;
  };
  shape: ShapeConfig;
  typography: TypographyConfig;
}

/**
 * Full theme object
 */
export interface Theme {
  id: string;
  displayName: string;
  theme: ThemeConfig;
}

/**
 * Basic theme information for listing
 */
export interface BasicTheme {
  id: string;
  displayName: string;
}
