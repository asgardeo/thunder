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

import {Box, CircularProgress, Typography} from '@wso2/oxygen-ui';
import {useCallback, useEffect, useRef, type JSX, type RefObject} from 'react';
import {useGetLayout, useUpdateLayout} from '@thunder/shared-design';
import ScreenEditor from './layouts/ScreenEditor';

interface LayoutConfigPanelProps {
  layoutId: string | null;
  selectedScreen: string | null;
  onScreenChange: (screen: string) => void;
  screenDraft: Record<string, unknown> | null;
  onScreenDraftChange: (draft: Record<string, unknown>) => void;
  onDirtyChange?: (dirty: boolean) => void;
  saveHandlerRef?: RefObject<() => void>;
}

// ── Immutable deep-set helper ───────────────────────────────────────────────

function setIn(obj: Record<string, unknown>, path: string[], value: unknown): Record<string, unknown> {
  const [head, ...rest] = path;
  if (rest.length === 0) return {...obj, [head]: value};
  return {
    ...obj,
    [head]: setIn((obj[head] ?? {}) as Record<string, unknown>, rest, value),
  };
}

// ── Main panel component ────────────────────────────────────────────────────

export default function LayoutConfigPanel({
  layoutId,
  selectedScreen,
  onScreenChange,
  screenDraft,
  onScreenDraftChange,
  onDirtyChange,
  saveHandlerRef,
}: LayoutConfigPanelProps): JSX.Element {
  const {data: layout, isLoading} = useGetLayout(layoutId ?? '');
  const {mutateAsync} = useUpdateLayout();

  const screens = (layout?.layout?.screens ?? {}) as Record<string, Record<string, unknown>>;
  const screenNames = Object.keys(screens);

  // Pick first screen when layout loads and none is selected yet
  useEffect(() => {
    if (screenNames.length > 0 && !selectedScreen) {
      onScreenChange(screenNames[0]);
    }
  }, [screenNames.length]); // eslint-disable-line react-hooks/exhaustive-deps

  // Sync draft from server when selected screen or layout data changes
  useEffect(() => {
    if (selectedScreen && screens[selectedScreen]) {
      onScreenDraftChange(JSON.parse(JSON.stringify(screens[selectedScreen])) as Record<string, unknown>);
      onDirtyChange?.(false);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedScreen, layout]);

  const updateField = (path: string[], value: unknown): void => {
    if (!screenDraft) return;
    onScreenDraftChange(setIn(screenDraft, path, value));
    onDirtyChange?.(true);
  };

  const handleSave = useCallback(() => {
    if (!screenDraft || !layout || !selectedScreen) return;
    const updatedScreens = {...screens, [selectedScreen]: screenDraft};
    const updatedLayout = {...layout.layout, screens: updatedScreens};
    mutateAsync({
      layoutId: layout.id,
      data: {displayName: layout.displayName, layout: updatedLayout},
    })
      .then(() => onDirtyChange?.(false))
      .catch(() => undefined);
  }, [screenDraft, layout, selectedScreen, screens, mutateAsync, onDirtyChange]);

  // Keep the parent's save ref pointing to the latest handleSave
  const handleSaveLatest = useRef(handleSave);
  handleSaveLatest.current = handleSave;

  useEffect(() => {
    if (saveHandlerRef) {
      saveHandlerRef.current = () => handleSaveLatest.current();
    }
  }, [saveHandlerRef]);

  // ── Early returns ────────────────────────────────────────────────────────

  if (!layoutId) {
    return (
      <Box sx={{height: '100%', display: 'flex', alignItems: 'center', justifyContent: 'center', p: 3}}>
        <Typography variant="body2" color="text.secondary" align="center">
          Select a layout to view constraints
        </Typography>
      </Box>
    );
  }

  if (isLoading) {
    return (
      <Box sx={{height: '100%', display: 'flex', alignItems: 'center', justifyContent: 'center'}}>
        <CircularProgress size={24} />
      </Box>
    );
  }

  if (!layout) {
    return (
      <Box sx={{p: 2}}>
        <Typography variant="body2" color="text.secondary">
          Failed to load layout configuration.
        </Typography>
      </Box>
    );
  }

  // ── Render ───────────────────────────────────────────────────────────────

  return (
    <Box sx={{overflowY: 'auto', height: '100%'}}>
      {/* Screen config editor */}
      {screenDraft && selectedScreen ? (
        <ScreenEditor screenDraft={screenDraft} onUpdate={updateField} />
      ) : (
        <Box sx={{px: 2, py: 2}}>
          <Typography variant="caption" color="text.secondary">
            No screen selected.
          </Typography>
        </Box>
      )}
    </Box>
  );
}
