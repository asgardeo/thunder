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

import {useContext} from 'react';
import ThemeBuilderContext, {type ThemeBuilderContextType} from './ThemeBuilderContext';

/**
 * React hook for accessing theme builder state throughout the application.
 *
 * This hook provides access to all the state needed for editing and previewing themes.
 * It must be used within a component tree wrapped by `ThemeBuilderProvider`,
 * otherwise it will throw an error.
 *
 * @returns The theme builder context containing state data and utility methods
 *
 * @throws {Error} Throws an error if used outside of ThemeBuilderProvider
 *
 * @example
 * Basic usage:
 * ```tsx
 * import useThemeBuilder from './useThemeBuilder';
 *
 * function ColorPicker() {
 *   const { draftTheme, updateDraftTheme, isDirty } = useThemeBuilder();
 *
 *   return (
 *     <div>
 *       <p>Unsaved changes: {isDirty}</p>
 *       <button onClick={() => updateDraftTheme(['colorSchemes', 'light', 'colors', 'primary', 'main'], '#ff0000')}>
 *         Set Red Primary
 *       </button>
 *     </div>
 *   );
 * }
 * ```
 *
 * @public
 */
export default function useThemeBuilder(): ThemeBuilderContextType {
  const context = useContext(ThemeBuilderContext);

  if (context === undefined) {
    throw new Error('useThemeBuilder must be used within ThemeBuilderProvider');
  }

  return context;
}
