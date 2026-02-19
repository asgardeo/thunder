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

import type {PropsWithChildren} from 'react';
import {useState, useMemo, useCallback, useEffect} from 'react';
import type {ThemeConfig} from '@thunder/shared-design';
import {useGetTheme} from '@thunder/shared-design';
import ThemeBuilderContext, {type ThemeBuilderContextType} from './ThemeBuilderContext';
import type {ThemeSection, Viewport} from '../../models/theme-builder';

/**
 * Props for the {@link ThemeBuilderProvider} component.
 *
 * @public
 */
export interface ThemeBuilderProviderProps extends PropsWithChildren {
  /**
   * The ID of the theme being edited
   */
  themeId: string;
}

/**
 * React context provider component that provides theme builder state
 * to all child components.
 *
 * This component manages all the state needed for editing and previewing a theme.
 * It automatically fetches the theme data and maintains a draft copy for live edits.
 *
 * @param props - The component props
 * @param props.themeId - The ID of the theme to edit
 * @param props.children - React children to be wrapped with the theme builder context
 *
 * @returns JSX element that provides theme builder context to children
 *
 * @example
 * ```tsx
 * import ThemeBuilderProvider from './ThemeBuilderProvider';
 * import ThemeBuilderPage from './ThemeBuilderPage';
 *
 * function App() {
 *   return (
 *     <ThemeBuilderProvider themeId="theme-123">
 *       <ThemeBuilderPage />
 *     </ThemeBuilderProvider>
 *   );
 * }
 * ```
 *
 * @public
 */
export default function ThemeBuilderProvider({themeId, children}: ThemeBuilderProviderProps) {
  const {data: themeData, isLoading} = useGetTheme(themeId);

  const [draftTheme, setDraftTheme] = useState<ThemeConfig | null>(null);
  const [isDirty, setIsDirty] = useState<boolean>(false);
  const [activeSection, setActiveSection] = useState<ThemeSection>('colors');
  const [previewColorScheme, setPreviewColorScheme] = useState<'light' | 'dark' | 'system'>('light');
  const [viewport, setViewport] = useState<Viewport>('desktop');
  const [isSaving, setIsSaving] = useState<boolean>(false);

  const displayName = themeData?.displayName ?? null;

  // When theme data loads, sync previewColorScheme with the draft's defaultColorScheme
  useEffect(() => {
    if (themeData?.theme?.defaultColorScheme === 'dark') {
      setPreviewColorScheme('dark');
    }
  }, [themeData]);

  // Initialize draft when theme data loads
  useEffect(() => {
    if (themeData?.theme && !draftTheme) {
      setDraftTheme(themeData.theme);
    }
  }, [themeData, draftTheme]);

  /**
   * Resets the draft to match the original theme
   */
  const resetDraft = useCallback(() => {
    if (themeData?.theme) {
      setDraftTheme(themeData.theme);
      setIsDirty(false);
    }
  }, [themeData]);

  /**
   * Updates a specific path in the draft theme using dot notation
   * @param path - Array of keys representing the path to update
   * @param value - The new value to set
   */
  const updateDraftTheme = useCallback((path: string[], value: unknown) => {
    setDraftTheme((prev) => {
      if (!prev) return prev;

      const newTheme = JSON.parse(JSON.stringify(prev)) as ThemeConfig;
      let current: Record<string, unknown> = newTheme as unknown as Record<string, unknown>;

      // Navigate to the parent of the target property
      for (let i = 0; i < path.length - 1; i += 1) {
        const key = path[i];
        if (!(key in current)) {
          current[key] = {};
        }
        current = current[key] as Record<string, unknown>;
      }

      // Set the value
      const lastKey = path[path.length - 1];
      current[lastKey] = value;

      return newTheme;
    });
    setIsDirty(true);
  }, []);

  const contextValue: ThemeBuilderContextType = useMemo(
    () => ({
      themeId,
      originalTheme: themeData?.theme ?? null,
      displayName,
      draftTheme,
      setDraftTheme,
      isDirty,
      setIsDirty,
      activeSection,
      setActiveSection,
      previewColorScheme,
      setPreviewColorScheme,
      viewport,
      setViewport,
      isSaving,
      setIsSaving,
      resetDraft,
      updateDraftTheme,
    }),
    [
      themeId,
      themeData?.theme,
      displayName,
      draftTheme,
      isDirty,
      activeSection,
      viewport,
      previewColorScheme,
      isSaving,
      resetDraft,
      updateDraftTheme,
    ],
  );

  if (isLoading) {
    return null; // or a loading spinner
  }

  return <ThemeBuilderContext.Provider value={contextValue}>{children}</ThemeBuilderContext.Provider>;
}
