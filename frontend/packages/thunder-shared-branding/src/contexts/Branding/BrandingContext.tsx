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

import {Context, createContext} from 'react';
import type {Theme} from '@wso2/oxygen-ui';
import {Branding} from '../../models/branding';
import {LayoutConfig} from '../../models/layout';
import type {ImageConfig} from '../../models/theme';

/**
 * Branding context interface that provides access to Thunder branding configuration
 * and utility methods for branding-related operations.
 *
 * @public
 */
export interface BrandingContextType {
  /**
   * The complete branding data resolved from the server
   */
  branding?: Branding;

  /**
   * Whether branding is enabled and loaded
   */
  isBrandingEnabled: boolean;

  /**
   * Whether branding data is currently being loaded
   */
  isLoading: boolean;

  /**
   * Any error that occurred while loading branding data
   */
  error?: Error | null;

  /**
   * The theme resolved from branding data (directly accessible)
   */
  theme?: Theme;

  /**
   * The images configuration from branding data (directly accessible)
   */
  images?: ImageConfig;

  /**
   * The layout configuration from branding data (directly accessible)
   */
  layout?: LayoutConfig;
}

/**
 * React context for accessing Thunder branding configuration throughout the application.
 *
 * This context provides access to the branding data loaded from the server, resolved theme,
 * layout configuration, images, and utility methods. It should be used within a `BrandingProvider` component.
 *
 * @example
 * ```tsx
 * import BrandingContext from './BrandingContext';
 * import { useContext } from 'react';
 *
 * const MyComponent = () => {
 *   const context = useContext(BrandingContext);
 *   if (!context) {
 *     throw new Error('Component must be used within BrandingProvider');
 *   }
 *
 *   const { theme, isBrandingEnabled, getLayout } = context;
 *   return <div>Branding enabled: {isBrandingEnabled}</div>;
 * };
 * ```
 *
 * @public
 */
const BrandingContext: Context<BrandingContextType | undefined> = createContext<BrandingContextType | undefined>(
  undefined,
);

export default BrandingContext;
