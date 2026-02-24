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
import {useCallback, useEffect, useRef, useState, type JSX, type RefObject} from 'react';
import {useGetTheme, useUpdateTheme} from '@thunder/shared-design';
import type {ColorSchemeColors, ThemeConfig} from '@thunder/shared-design';
import ColorBuilderContent from './themes/ColorBuilderContent';
import useThemeBuilder from '../contexts/ThemeBuilder/useThemeBuilder';
import GeneralBuilderContent from './themes/GeneralBuilderContent';
import ShapeBuilderContent from './themes/ShapeBuilderContent';
import TypographyBuilderContent from './themes/TypographyBuilderContent';
import type {ThemeSection} from '../models/theme-builder';

interface ThemeConfigPanelProps {
  themeId: string | null;
  saveHandlerRef?: RefObject<() => void>;
  /** When set, renders only that section's content (builder mode, no accordions) */
  activeSection?: ThemeSection;
}

// ── Main panel component ────────────────────────────────────────────────────

export default function ThemeConfigPanel({themeId, saveHandlerRef, activeSection}: ThemeConfigPanelProps): JSX.Element {
  // useGetTheme is kept here only to obtain save metadata (displayName, description, id).
  // React Query deduplicates the request — the provider already issues the same call.
  const {data: theme, isLoading} = useGetTheme(themeId ?? '');
  const {mutateAsync} = useUpdateTheme();

  // Draft state lives in the context so ThemePreviewPanel always sees the latest changes.
  const {draftTheme, setDraftTheme, setIsDirty, setPreviewColorScheme} = useThemeBuilder();

  const [colorSchemeTab, setColorSchemeTab] = useState<'light' | 'dark'>('light');

  /**
   * Apply an updater function to a deep-cloned copy of draftTheme and push the
   * result back into the context so the preview panel re-renders automatically.
   */
  const updateDraft = (updater: (d: ThemeConfig) => void): void => {
    setDraftTheme(
      (() => {
        if (!draftTheme) return draftTheme;
        const next = JSON.parse(JSON.stringify(draftTheme)) as ThemeConfig;
        updater(next);
        return next;
      })(),
    );
    setIsDirty(true);
  };

  const updateColorScheme = (scheme: 'light' | 'dark', updater: (c: ColorSchemeColors) => void): void => {
    updateDraft((d) => {
      const colors = d.colorSchemes?.[scheme]?.colors;
      if (colors) updater(colors);
    });
  };

  const handleSave = useCallback(() => {
    if (!draftTheme || !theme) return;
    mutateAsync({
      themeId: theme.id,
      data: {displayName: theme.displayName, description: theme.description, theme: draftTheme},
    })
      .then(() => setIsDirty(false))
      .catch(() => undefined);
  }, [draftTheme, theme, mutateAsync, setIsDirty]);

  const handleSaveLatest = useRef(handleSave);
  handleSaveLatest.current = handleSave;

  useEffect(() => {
    if (saveHandlerRef) {
      saveHandlerRef.current = () => handleSaveLatest.current();
    }
  }, [saveHandlerRef]);

  // ── Early returns ────────────────────────────────────────────────────────

  if (!themeId) {
    return (
      <Box sx={{height: '100%', display: 'flex', alignItems: 'center', justifyContent: 'center', p: 3}}>
        <Typography variant="body2" color="text.secondary" align="center">
          Select a theme to view configuration
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

  if (!theme || !draftTheme) {
    return (
      <Box sx={{p: 2}}>
        <Typography variant="body2" color="text.secondary">
          Failed to load theme configuration.
        </Typography>
      </Box>
    );
  }

  const lightColors = draftTheme.colorSchemes?.light?.colors;
  const darkColors = draftTheme.colorSchemes?.dark?.colors;

  return (
    <Box sx={{height: '100%', display: 'flex', flexDirection: 'column'}}>
      {/* Light / Dark tab strip — only for Colors section */}
      {activeSection === 'colors' && (
        <Box sx={{display: 'flex', borderBottom: '1px solid', borderColor: 'divider', flexShrink: 0}}>
          {(['light', 'dark'] as const).map((scheme) => (
            <Box
              key={scheme}
              onClick={() => {
                setColorSchemeTab(scheme);
                setPreviewColorScheme(scheme);
              }}
              sx={{
                flex: 1,
                py: 1,
                textAlign: 'center',
                cursor: 'pointer',
                borderBottom: '2px solid',
                borderColor: colorSchemeTab === scheme ? 'primary.main' : 'transparent',
                transition: 'border-color 0.15s',
                '&:hover': {bgcolor: 'action.hover'},
              }}
            >
              <Typography
                variant="caption"
                sx={{
                  fontWeight: colorSchemeTab === scheme ? 700 : 400,
                  fontSize: '0.75rem',
                  textTransform: 'capitalize',
                  color: colorSchemeTab === scheme ? 'primary.main' : 'text.secondary',
                }}
              >
                {scheme}
              </Typography>
            </Box>
          ))}
        </Box>
      )}

      <Box sx={{flex: 1, overflowY: 'auto', p: 1.5}}>
        {activeSection === 'colors' && colorSchemeTab === 'light' && lightColors && (
          <ColorBuilderContent colors={lightColors} onUpdate={(up) => updateColorScheme('light', up)} />
        )}
        {activeSection === 'colors' && colorSchemeTab === 'dark' && darkColors && (
          <ColorBuilderContent colors={darkColors} onUpdate={(up) => updateColorScheme('dark', up)} />
        )}
        {activeSection === 'shape' && draftTheme.shape !== undefined && (
          <ShapeBuilderContent draft={draftTheme} onUpdate={updateDraft} />
        )}
        {activeSection === 'typography' && draftTheme.typography && (
          <TypographyBuilderContent draft={draftTheme} onUpdate={updateDraft} />
        )}
        {activeSection === 'general' && <GeneralBuilderContent draft={draftTheme} onUpdate={updateDraft} />}
      </Box>
    </Box>
  );
}
