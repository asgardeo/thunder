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

import {useContext, useMemo} from 'react';
import type {Theme} from '@wso2/oxygen-ui';
import DesignContext, {DesignContextType} from './DesignContext';
import oxygenUIThemeTransformer from '../../utils/oxygenUIThemeTransformer';

/**
 * React hook for accessing Thunder design configuration throughout the application.
 *
 * This hook provides access to the design data loaded from the server, resolved theme,
 * and layout configuration. It must be used within a component tree wrapped by
 * `DesignProvider`, otherwise it will throw an error.
 *
 * @param baseTheme - Optional base theme to extend from. If provided, overrides the provider's baseTheme
 * @returns The design context containing design data and resolved theme/layout
 *
 * @throws {Error} Throws an error if used outside of DesignProvider
 *
 * @public
 */
export default function useDesign(baseTheme?: Theme): DesignContextType {
  const context = useContext(DesignContext);
  if (context === undefined) {
    throw new Error('useDesign must be used within a DesignProvider');
  }

  // If a baseTheme is provided, override the transformedTheme
  const overriddenTransformedTheme = useMemo(() => {
    if (baseTheme && context.design?.theme) {
      return oxygenUIThemeTransformer(baseTheme, context.design.theme);
    }
    return undefined;
  }, [baseTheme, context.design?.theme]);

  return useMemo(
    () => ({
      ...context,
      transformedTheme: overriddenTransformedTheme ?? context.transformedTheme,
    }),
    [context, overriddenTransformedTheme],
  );
}
