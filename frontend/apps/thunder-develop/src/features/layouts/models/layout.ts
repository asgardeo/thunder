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
 * Background configuration for a screen
 */
export interface ScreenBackground {
  type: 'gradient' | 'solid' | 'image';
  value: string;
}

/**
 * Container configuration for a screen
 */
export interface ScreenContainer {
  maxWidth: number;
  padding: number;
  borderRadius: number;
  elevation: number;
  background: string;
}

/**
 * Spacing configuration for a screen
 */
export interface ScreenSpacing {
  componentGap: number;
  sectionGap: number;
}

/**
 * Header configuration for a screen
 */
export interface ScreenHeader {
  enabled: boolean;
  height: number;
  padding: number;
  showLogo?: boolean;
  showLanguageSelector?: boolean;
  showBackButton?: boolean;
}

/**
 * Footer configuration for a screen
 */
export interface ScreenFooter {
  enabled: boolean;
  height: number;
  padding: number;
  showLinks?: boolean;
}

/**
 * Screen configuration
 */
export interface ScreenConfig {
  template: string;
  background?: ScreenBackground;
  container?: ScreenContainer;
  spacing?: ScreenSpacing;
  header?: ScreenHeader;
  footer?: ScreenFooter;
  extends?: string;
}

/**
 * Layout configuration containing all screens
 */
export interface LayoutConfig {
  screens: {
    auth?: ScreenConfig;
    signin?: ScreenConfig;
    signup?: ScreenConfig;
    mfa?: ScreenConfig;
    'forgot-password'?: ScreenConfig;
    recovery?: ScreenConfig;
    'password-reset'?: ScreenConfig;
    [key: string]: ScreenConfig | undefined;
  };
}

/**
 * Full layout object
 */
export interface Layout {
  id: string;
  displayName: string;
  layout: LayoutConfig;
}

/**
 * Basic layout information for listing
 */
export interface BasicLayout {
  id: string;
  displayName: string;
}
