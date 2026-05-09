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
import LayoutBuilderContext, {type LayoutBuilderContextType} from './LayoutBuilderContext';

/**
 * React hook for accessing layout builder state throughout the application.
 *
 * This hook provides access to all the state needed for editing and previewing layouts.
 * It must be used within a component tree wrapped by `LayoutBuilderProvider`,
 * otherwise it will throw an error.
 *
 * @returns The layout builder context containing state data and utility methods
 *
 * @throws {Error} Throws an error if used outside of LayoutBuilderProvider
 *
 * @example
 * Basic usage:
 * ```tsx
 * import useLayoutBuilder from './useLayoutBuilder';
 *
 * function ScreenSelector() {
 *   const { selectedScreen, setSelectedScreen, getAllScreens, isDirty } = useLayoutBuilder();
 *
 *   const screens = getAllScreens();
 *
 *   return (
 *     <div>
 *       <p>Unsaved changes: {isDirty}</p>
 *       <select value={selectedScreen ?? ''} onChange={(e) => setSelectedScreen(e.target.value)}>
 *         {Object.keys(screens).map((name) => (
 *           <option key={name} value={name}>{name}</option>
 *         ))}
 *       </select>
 *     </div>
 *   );
 * }
 * ```
 *
 * @public
 */
export default function useLayoutBuilder(): LayoutBuilderContextType {
  const context = useContext(LayoutBuilderContext);

  if (context === undefined) {
    throw new Error('useLayoutBuilder must be used within LayoutBuilderProvider');
  }

  return context;
}
