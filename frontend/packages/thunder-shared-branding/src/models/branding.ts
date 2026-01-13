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

import type {Theme} from './theme';
import type {LayoutConfig} from './layout';

/**
 * Branding preferences stored as flexible JSON
 * The structure can evolve over time with additional properties
 */
export interface BrandingPreferences {
  /**
   * Theme configuration
   */
  theme?: Theme;

  /**
   * Global layout configuration that applies across all themes
   */
  layout?: LayoutConfig;

  /**
   * Additional properties can be added as needed
   */
  [key: string]: unknown;
}

/**
 * Full branding configuration details
 */
export interface Branding {
  /**
   * Unique identifier for the branding configuration
   * @example "3fa85f64-5717-4562-b3fc-2c963f66afa6"
   */
  id: string;

  /**
   * Display name for the branding configuration
   * @example "Application 1 Branding"
   */
  displayName: string;

  /**
   * Branding preferences containing theme and other customization options
   */
  preferences: BrandingPreferences;
}

/**
 * Enumeration of supported branding resolution types.
 * Used to specify the type of entity for which branding configuration should be resolved.
 */
export const BrandingType = {
  /** Application-level branding */
  APP: 'APP',

  /** Organizational Unit-level branding */
  OU: 'OU',
} as const;

/**
 * Union type representing the possible branding resolution types.
 * @example 'APP' | 'OU'
 */
export type BrandingType = (typeof BrandingType)[keyof typeof BrandingType];
