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

import type {LayoutConfig} from './layout';

/**
 * Logo configuration with URL and metadata
 */
export interface LogoConfig {
  /**
   * URL to the logo image
   * @example "https://example.com/logo-dark.png"
   */
  url: string;

  /**
   * Alternative text for the logo
   * @example "Application Logo Dark"
   */
  alt: string;

  /**
   * Logo width in pixels
   * @example 128
   */
  width: number;

  /**
   * Logo height in pixels
   * @example 64
   */
  height: number;
}

/**
 * Favicon configuration with URL and metadata
 */
export interface FaviconConfig {
  /**
   * URL to the favicon file
   * @example "https://example.com/favicon-dark.ico"
   */
  url: string;

  /**
   * MIME type of the favicon
   * @example "image/x-icon"
   */
  type: string;
}

/**
 * Logo collection for a specific theme
 */
export interface LogoCollection {
  /**
   * Primary application logo
   */
  primary: LogoConfig;

  /**
   * Favicon configuration
   */
  favicon: FaviconConfig;
}

/**
 * Background image configuration
 */
export interface BackgroundImageConfig {
  /**
   * URL to the background image
   * @example "https://example.com/login-bg-dark.jpg"
   */
  url: string;

  /**
   * CSS background-position value
   * @example "center"
   */
  position: string;

  /**
   * CSS background-size value
   * @example "cover"
   */
  size: string;

  /**
   * CSS background-repeat value
   * @example "no-repeat"
   */
  repeat?: string;
}

/**
 * Custom icons configuration
 */
export interface CustomIconsConfig {
  /**
   * Success icon URL
   * @example "https://example.com/icons/success-dark.svg"
   */
  success?: string;

  /**
   * Error icon URL
   * @example "https://example.com/icons/error-dark.svg"
   */
  error?: string;

  /**
   * Additional custom icons
   */
  [iconName: string]: string | undefined;
}

/**
 * Image configuration for a specific color scheme
 */
export interface ImageConfig {
  /**
   * Logo configuration
   */
  logo: LogoCollection;

  /**
   * Background images configuration
   */
  background?: {
    /**
     * Login page background
     */
    login?: BackgroundImageConfig;

    /**
     * Additional background images
     */
    [backgroundName: string]: BackgroundImageConfig | undefined;
  };

  /**
   * Custom icons configuration
   */
  icons?: {
    /**
     * Custom icon set
     */
    custom?: CustomIconsConfig;
  };
}

/**
 * Color configuration for a specific color type (primary, secondary, tertiary)
 */
export interface ThemeColor {
  /**
   * Main color value
   * @example "#1976d2"
   */
  main: string;

  /**
   * Dark variant of the color
   * @example "#0d47a1"
   */
  dark: string;

  /**
   * Text color that contrasts with the main color
   * @example "#ffffff"
   */
  contrastText: string;
}

/**
 * Color palette for a specific color scheme (light or dark)
 */
export interface ColorPalette {
  /**
   * Primary color configuration
   */
  primary: ThemeColor;

  /**
   * Secondary color configuration
   */
  secondary: ThemeColor;

  /**
   * Tertiary color configuration (optional)
   */
  tertiary?: ThemeColor;
}

/**
 * Color scheme configuration for light or dark mode
 */
export interface ColorScheme {
  /**
   * Color palette for this color scheme
   */
  colors: ColorPalette;

  /**
   * Image configuration for this color scheme
   */
  images?: ImageConfig;
}

/**
 * Theme configuration containing color schemes
 */
export interface Theme {
  /**
   * Active color scheme ("light" or "dark")
   */
  activeColorScheme: 'light' | 'dark';

  /**
   * Available color schemes
   */
  colorSchemes: {
    light?: ColorScheme;
    dark?: ColorScheme;
  };

  /**
   * Layout configuration for the theme
   */
  layout?: LayoutConfig;
}
