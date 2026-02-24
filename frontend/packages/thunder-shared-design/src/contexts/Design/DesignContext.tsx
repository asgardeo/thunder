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

import {Context, createContext} from 'react';
import type {Theme as OxygenUITheme} from '@wso2/oxygen-ui';
import type {LayoutConfig} from '../../models/layout';
import type {DesignResolveResponse} from '../../models/responses';
import type {Theme} from '../../models/theme';

/**
 * Design context interface that provides access to Thunder design configuration
 * and utility methods for design-related operations.
 *
 * @public
 */
export interface DesignContextType {
  /**
   * The complete design data resolved from the server
   */
  design?: DesignResolveResponse;

  /**
   * Whether design is enabled and loaded
   */
  isDesignEnabled: boolean;

  /**
   * Whether design data is currently being loaded
   */
  isLoading: boolean;

  /**
   * Any error that occurred while loading design data
   */
  error?: Error | null;

  /**
   * The Thunder theme resolved from design data (directly accessible)
   */
  theme?: Theme;

  /**
   * The layout configuration from design data (directly accessible)
   */
  layout?: LayoutConfig;
}

/**
 * React context for accessing Thunder design configuration throughout the application.
 *
 * This context provides access to the design data loaded from the server, resolved theme,
 * and layout configuration. It should be used within a `DesignProvider` component.
 *
 * @public
 */
const DesignContext: Context<DesignContextType | undefined> = createContext<DesignContextType | undefined>(undefined);

export default DesignContext;
