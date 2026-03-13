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

import {useCallback, useRef, useState, type JSX} from 'react';
import {useTranslation} from 'react-i18next';
import {useNavigate} from 'react-router';
import {Box, Button, useColorScheme} from '@wso2/oxygen-ui';
import {Save, Trash} from '@wso2/oxygen-ui-icons-react';
import BuilderLayout from '../../../components/BuilderLayout/BuilderLayout';
import BuilderStaticPanel from '../../../components/BuilderLayout/BuilderStaticPanel';
import ThemePreviewPanel from '../components/ThemePreviewPanel';
import ThemeConfigPanel from '../components/ThemeConfigPanel';
import ThemeBuilderLeftPanel from '../components/themes/ThemeBuilderLeftPanel';
import ThemeDeleteDialog from '../components/themes/ThemeDeleteDialog';
import DesignUIConstants from '../constants/design-ui-constants';
import useThemeBuilder from '../contexts/ThemeBuilder/useThemeBuilder';

export default function ThemeBuilderPage(): JSX.Element {
  const {t} = useTranslation('design');
  const {mode, systemMode} = useColorScheme();
  const navigate = useNavigate();

  const {themeId, displayName, activeSection, setActiveSection, isDirty, draftTheme, setDraftTheme, setIsDirty} =
    useThemeBuilder();

  const saveHandlerRef = useRef<() => void>(() => {});
  const [isPanelOpen, setIsPanelOpen] = useState(true);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);

  const handleTogglePanel = useCallback(() => {
    setIsPanelOpen((prev) => !prev);
  }, []);

  const handleBack = useCallback(() => {
    // eslint-disable-next-line @typescript-eslint/no-floating-promises
    navigate('/design');
  }, [navigate]);

  const handleDeleteSuccess = useCallback(() => {
    // eslint-disable-next-line @typescript-eslint/no-floating-promises
    navigate('/design');
  }, [navigate]);

  const bgColor = (systemMode ?? mode) === 'dark' ? '#141414' : '#f6f7f9';

  const leftPanelContent = (
    <ThemeBuilderLeftPanel
      onBack={handleBack}
      onPanelToggle={handleTogglePanel}
      draftTheme={draftTheme}
      setDraftTheme={setDraftTheme}
      setIsDirty={setIsDirty}
      activeSection={activeSection}
      setActiveSection={setActiveSection}
    />
  );

  return (
    <>
      <Box
        sx={{
          width: '100%',
          height: 'inherit',
          display: 'flex',
          flexDirection: 'column',
          bgcolor: 'var(--flow-builder-background-color)',
          '[data-color-scheme="dark"] &': {
            bgcolor: 'var(--flow-builder-background-color-dark)',
          },
        }}
      >
        {/* ── Three-column builder area ──────────────────────────────────────── */}
        <Box sx={{flex: 1, overflow: 'hidden', p: 1}}>
          <BuilderLayout
            open={isPanelOpen}
            onPanelToggle={handleTogglePanel}
            panelWidth={DesignUIConstants.LEFT_PANEL_WIDTH}
            panelContent={leftPanelContent}
            expandTooltip={t('themes.builder.tooltips.show_sections', 'Show sections')}
            panelPaperSx={{
              overflow: 'hidden',
              display: 'flex',
              flexDirection: 'column',
              borderRight: '1px solid',
              borderColor: 'divider',
            }}
          >
            <Box sx={{display: 'flex', height: '100%', overflow: 'hidden', bgcolor: bgColor}}>
              {/* ── Center: canvas preview ─────────────────────────────────── */}
              <Box
                component="main"
                sx={{
                  flexGrow: 1,
                  height: '100%',
                  overflow: 'hidden',
                  display: 'flex',
                  flexDirection: 'column',
                  borderRadius: 1,
                  mx: 2,
                }}
              >
                <ThemePreviewPanel themeId={themeId ?? null} />
              </Box>

              <Box>
                {/* Save */}
                <Box sx={{mr: 1, mb: 1, p: 2, display: 'flex', justifyContent: 'flex-end', gap: 2}}>
                  <Button
                    variant="text"
                    color="error"
                    startIcon={<Trash size={18} />}
                    onClick={() => setDeleteDialogOpen(true)}
                  >
                    {t('themes.builder.actions.delete.label', 'Delete')}
                  </Button>
                  <Button
                    variant="contained"
                    disabled={!isDirty}
                    startIcon={<Save size={18} />}
                    onClick={() => saveHandlerRef.current()}
                  >
                    {t('themes.builder.actions.save.label', 'Save')}
                  </Button>
                </Box>
                {/* ── Right panel: section config ────────────────────────────── */}
                <BuilderStaticPanel
                  width={DesignUIConstants.RIGHT_PANEL_WIDTH}
                  header={activeSection ? t(`themes.builder.sections.${activeSection}.label`, activeSection) : 'Config'}
                >
                  <ThemeConfigPanel
                    themeId={themeId ?? null}
                    activeSection={activeSection}
                    saveHandlerRef={saveHandlerRef}
                  />
                </BuilderStaticPanel>
              </Box>
            </Box>
          </BuilderLayout>
        </Box>
      </Box>

      <ThemeDeleteDialog
        open={deleteDialogOpen}
        themeId={themeId ?? null}
        themeName={displayName ?? null}
        onClose={() => setDeleteDialogOpen(false)}
        onSuccess={handleDeleteSuccess}
      />
    </>
  );
}
