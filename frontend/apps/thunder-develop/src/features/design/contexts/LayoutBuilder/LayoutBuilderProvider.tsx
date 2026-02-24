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
import {useGetLayout} from '@thunder/shared-design';
import LayoutBuilderContext, {type LayoutBuilderContextType, type LayoutConfig} from './LayoutBuilderContext';

/**
 * Props for the {@link LayoutBuilderProvider} component.
 *
 * @public
 */
export interface LayoutBuilderProviderProps extends PropsWithChildren {
  /**
   * The ID of the layout being edited
   */
  layoutId: string;
}

/**
 * React context provider component that provides layout builder state
 * to all child components.
 *
 * This component manages all the state needed for editing and previewing a layout.
 * It automatically fetches the layout data and maintains a draft copy for live edits.
 *
 * @param props - The component props
 * @param props.layoutId - The ID of the layout to edit
 * @param props.children - React children to be wrapped with the layout builder context
 *
 * @returns JSX element that provides layout builder context to children
 *
 * @example
 * ```tsx
 * import LayoutBuilderProvider from './LayoutBuilderProvider';
 * import LayoutBuilderPage from './LayoutBuilderPage';
 *
 * function App() {
 *   return (
 *     <LayoutBuilderProvider layoutId="layout-123">
 *       <LayoutBuilderPage />
 *     </LayoutBuilderProvider>
 *   );
 * }
 * ```
 *
 * @public
 */
export default function LayoutBuilderProvider({layoutId, children}: LayoutBuilderProviderProps) {
  const {data: layoutData, isLoading} = useGetLayout(layoutId);

  const [draftLayout, setDraftLayout] = useState<LayoutConfig | null>(null);
  const [isDirty, setIsDirty] = useState<boolean>(false);
  const [selectedScreen, setSelectedScreen] = useState<string | null>(null);
  const [screenDraft, setScreenDraft] = useState<Record<string, unknown> | null>(null);
  const [extraScreens, setExtraScreens] = useState<Record<string, Record<string, unknown>>>({});
  const [isSaving, setIsSaving] = useState<boolean>(false);

  const displayName = layoutData?.displayName ?? null;

  // Initialize draft when layout data loads
  useEffect(() => {
    if (layoutData?.layout && !draftLayout) {
      setDraftLayout(layoutData.layout as LayoutConfig);
    }
  }, [layoutData, draftLayout]);

  // Auto-select first screen when layout loads
  useEffect(() => {
    if (draftLayout?.screens) {
      const screenNames = Object.keys(draftLayout.screens);
      if (screenNames.length > 0 && !selectedScreen) {
        setSelectedScreen(screenNames[0]);
      }
    }
  }, [draftLayout, selectedScreen]);

  /**
   * Resets the draft to match the original layout
   */
  const resetDraft = useCallback(() => {
    if (layoutData?.layout) {
      setDraftLayout(layoutData.layout as LayoutConfig);
      setExtraScreens({});
      setScreenDraft(null);
      setIsDirty(false);
    }
  }, [layoutData]);

  /**
   * Adds a new screen to the layout
   */
  const addScreen = useCallback((name: string, extendsBase: string) => {
    const newScreen: Record<string, unknown> = {extends: extendsBase};
    setExtraScreens((prev) => ({...prev, [name]: newScreen}));
    setSelectedScreen(name);
    setIsDirty(true);
  }, []);

  /**
   * Updates a specific path in the draft layout using dot notation
   * @param path - Array of keys representing the path to update
   * @param value - The new value to set
   */
  const updateDraftLayout = useCallback((path: string[], value: unknown) => {
    setDraftLayout((prev) => {
      if (!prev) return prev;

      const newLayout = JSON.parse(JSON.stringify(prev)) as LayoutConfig;
      let current: Record<string, unknown> = newLayout as unknown as Record<string, unknown>;

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

      return newLayout;
    });
    setIsDirty(true);
  }, []);

  /**
   * Gets all screens (original + extra)
   */
  const getAllScreens = useCallback(() => {
    const serverScreens = draftLayout?.screens ?? {};
    return {...serverScreens, ...extraScreens};
  }, [draftLayout, extraScreens]);

  /**
   * Gets base screen names (screens without extends property)
   */
  const getBaseScreenNames = useCallback(() => {
    const allScreens = getAllScreens();
    return Object.keys(allScreens).filter((name) => !allScreens[name]?.extends);
  }, [getAllScreens]);

  const contextValue: LayoutBuilderContextType = useMemo(
    () => ({
      layoutId,
      originalLayout: layoutData?.layout as LayoutConfig | null,
      displayName,
      draftLayout,
      setDraftLayout,
      isDirty,
      setIsDirty,
      selectedScreen,
      setSelectedScreen,
      screenDraft,
      setScreenDraft,
      extraScreens,
      setExtraScreens,
      isSaving,
      setIsSaving,
      resetDraft,
      addScreen,
      updateDraftLayout,
      getAllScreens,
      getBaseScreenNames,
    }),
    [
      layoutId,
      layoutData?.layout,
      displayName,
      draftLayout,
      isDirty,
      selectedScreen,
      screenDraft,
      extraScreens,
      isSaving,
      resetDraft,
      addScreen,
      updateDraftLayout,
      getAllScreens,
      getBaseScreenNames,
    ],
  );

  if (isLoading) {
    return null; // or a loading spinner
  }

  return <LayoutBuilderContext.Provider value={contextValue}>{children}</LayoutBuilderContext.Provider>;
}
