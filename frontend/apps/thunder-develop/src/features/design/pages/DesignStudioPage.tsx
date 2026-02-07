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

import {useState, useEffect, type JSX} from 'react';
import {
  Box,
  Drawer,
  IconButton,
  Paper,
  Stack,
  Toolbar,
  Typography,
  Button,
} from '@wso2/oxygen-ui';
import {ChevronLeft, ChevronRight, Save} from '@wso2/oxygen-ui-icons-react';
import {useTranslation} from 'react-i18next';
import {useLogger} from '@thunder/logger';
import useNotification from '@/features/design-system/hooks/useNotification';
import DeviceSelector from '../components/DeviceSelector';
import ThemeControls from '../components/ThemeControls';
import LayoutControls from '../components/LayoutControls';
import LayoutBuilder from '../components/LayoutBuilder';
import ThemeSaveDialog from '../components/ThemeSaveDialog';
import ThemePreview from '@/features/themes/components/ThemePreview';
import {defaultThemeFormData} from '@/features/themes/schemas/themeSchema';
import {DEVICE_VIEWPORTS, type DeviceType} from '../models/design';
import useGetThemes from '@/features/themes/api/useGetThemes';
import useGetTheme from '@/features/themes/api/useGetTheme';
import useGetLayouts from '@/features/layouts/api/useGetLayouts';
import useGetLayout from '@/features/layouts/api/useGetLayout';
import useCreateTheme from '@/features/themes/api/useCreateTheme';
import useUpdateTheme from '@/features/themes/api/useUpdateTheme';
import type {ThemeConfig} from '@/features/themes/models/theme';
import type {LayoutConfig} from '@/features/layouts/models/layout';

const LEFT_DRAWER_WIDTH = 320;
const RIGHT_DRAWER_WIDTH = 280;

/**
 * Unified Design Studio page for theme customization with layout preview
 */
export default function DesignStudioPage(): JSX.Element {
  const {t} = useTranslation();
  const logger = useLogger('DesignStudioPage');
  const {showSuccess, showError} = useNotification();

  // UI State
  const [leftDrawerOpen, setLeftDrawerOpen] = useState(true);
  const [rightDrawerOpen, setRightDrawerOpen] = useState(true);
  const [selectedDevice, setSelectedDevice] = useState<DeviceType>('desktop');
  const [showLayoutBuilder, setShowLayoutBuilder] = useState(false);
  const [saveDialogOpen, setSaveDialogOpen] = useState(false);

  // Theme State
  const [selectedBaseThemeId, setSelectedBaseThemeId] = useState<string | null>(null);
  const [selectedLayoutId, setSelectedLayoutId] = useState<string | null>(null);

  // API Queries
  const {data: themesData, isLoading: isLoadingThemes} = useGetThemes();
  const {data: themeData, isLoading: isLoadingTheme} = useGetTheme(selectedBaseThemeId || '');
  const {data: layoutsData, isLoading: isLoadingLayouts} = useGetLayouts();
  const {data: layoutData, isLoading: isLoadingLayout} = useGetLayout(selectedLayoutId || '');

  // API Mutations
  const createTheme = useCreateTheme();
  const updateTheme = useUpdateTheme();

  // Derived State
  const [themeConfig, setThemeConfig] = useState<ThemeConfig>(defaultThemeFormData.theme);
  const [baseThemeName, setBaseThemeName] = useState<string | undefined>(undefined);

  const viewport = DEVICE_VIEWPORTS[selectedDevice];

  // Computed values from queries
  const themeConfigFromQuery = selectedBaseThemeId ? themeData?.theme : defaultThemeFormData.theme;
  const layoutConfigFromQuery = selectedLayoutId ? layoutData?.layout : null;

  // Load base theme when query completes
  useEffect(() => {
    if (selectedBaseThemeId && themeData) {
      setThemeConfig(themeData.theme);
      setBaseThemeName(themeData.displayName);
      logger.info('Base theme loaded', {themeId: selectedBaseThemeId, themeName: themeData.displayName});
    } else if (!selectedBaseThemeId) {
      // "Create from scratch" selected
      setThemeConfig(defaultThemeFormData.theme);
      setBaseThemeName(undefined);
    }
  }, [selectedBaseThemeId, themeData, logger]);

  const handleBaseThemeSelect = (themeId: string | null) => {
    setSelectedBaseThemeId(themeId);
  };

  const handleThemeConfigChange = (config: ThemeConfig) => {
    setThemeConfig(config);
  };

  const handleLayoutSelect = (layoutId: string) => {
    setSelectedLayoutId(layoutId);
  };

  const handleCreateLayout = () => {
    setShowLayoutBuilder(true);
  };

  const handleLayoutBuilderClose = () => {
    setShowLayoutBuilder(false);
  };

  const handleLayoutSaved = (layoutId: string) => {
    setShowLayoutBuilder(false);
    setSelectedLayoutId(layoutId);
    showSuccess(t('design:layout.created', {defaultValue: 'Layout created successfully'}));
  };

  const handleSaveClick = () => {
    setSaveDialogOpen(true);
  };

  const handleSaveTheme = async (themeName: string, override: boolean) => {
    try {
      if (override && selectedBaseThemeId) {
        // Override existing theme
        await updateTheme.mutateAsync({
          id: selectedBaseThemeId,
          data: {
            displayName: themeName,
            theme: themeConfig,
          },
        });
        logger.info('Theme updated successfully', {id: selectedBaseThemeId, name: themeName});
        showSuccess(t('design:theme.updated', {defaultValue: `Theme "${themeName}" updated successfully`}));
      } else {
        // Create new theme
        const result = await createTheme.mutateAsync({
          displayName: themeName,
          theme: themeConfig,
        });
        setSelectedBaseThemeId(result.id);
        setBaseThemeName(themeName);
        logger.info('Theme created successfully', {id: result.id, name: themeName});
        showSuccess(t('design:theme.saved', {defaultValue: `Theme "${themeName}" saved successfully`}));
      }
      setSaveDialogOpen(false);
    } catch (error) {
      logger.error('Failed to save theme', error);
      showError(t('design:theme.saveFailed', {defaultValue: 'Failed to save theme. Please try again.'}));
    }
  };

  // Show Layout Builder if in layout creation mode
  if (showLayoutBuilder) {
    return <LayoutBuilder onClose={handleLayoutBuilderClose} onSaved={handleLayoutSaved} />;
  }

  return (
    <Box sx={{display: 'flex', height: '100vh', overflow: 'hidden'}}>
      {/* Left Drawer - Theme Controls */}
      <Drawer
        variant="persistent"
        anchor="left"
        open={leftDrawerOpen}
        sx={{
          width: LEFT_DRAWER_WIDTH,
          flexShrink: 0,
          '& .MuiDrawer-paper': {
            width: LEFT_DRAWER_WIDTH,
            boxSizing: 'border-box',
            position: 'relative',
            borderRight: '1px solid',
            borderColor: 'divider',
          },
        }}
      >
        <Box sx={{height: '100%', display: 'flex', flexDirection: 'column'}}>
          {/* Drawer Header */}
          <Box
            sx={{
              p: 2,
              borderBottom: '1px solid',
              borderColor: 'divider',
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
            }}
          >
            <Typography variant="h6">{t('design:theme.themeEditor', {defaultValue: 'Theme Editor'})}</Typography>
            <IconButton size="small" onClick={() => setLeftDrawerOpen(false)}>
              <ChevronLeft size={18} />
            </IconButton>
          </Box>

          {/* Theme Controls */}
          <ThemeControls
            themes={themesData?.themes || []}
            isLoading={isLoadingThemes}
            selectedBaseThemeId={selectedBaseThemeId}
            onBaseThemeSelect={handleBaseThemeSelect}
            themeConfig={themeConfig}
            onThemeConfigChange={handleThemeConfigChange}
          />
        </Box>
      </Drawer>

      {/* Main Content Area */}
      <Box sx={{flex: 1, display: 'flex', flexDirection: 'column', overflow: 'hidden'}}>
        {/* Toolbar */}
        <Toolbar
          sx={{
            borderBottom: '1px solid',
            borderColor: 'divider',
            bgcolor: 'background.paper',
            gap: 2,
            justifyContent: 'space-between',
          }}
        >
          <Stack direction="row" spacing={2} alignItems="center">
            {!leftDrawerOpen && (
              <IconButton size="small" onClick={() => setLeftDrawerOpen(true)}>
                <ChevronRight size={18} />
              </IconButton>
            )}
            <Typography variant="h6">{t('design:studio.title', {defaultValue: 'Design Studio'})}</Typography>
          </Stack>

          <Stack direction="row" spacing={2} alignItems="center">
            {/* Device Selector */}
            <DeviceSelector selectedDevice={selectedDevice} onDeviceChange={setSelectedDevice} />

            {/* Save Button */}
            <Button
              variant="contained"
              startIcon={<Save size={18} />}
              onClick={handleSaveClick}
              disabled={createTheme.isPending || updateTheme.isPending}
            >
              {createTheme.isPending || updateTheme.isPending
                ? t('common:actions.saving', {defaultValue: 'Saving...'})
                : t('common:actions.save', {defaultValue: 'Save Theme'})}
            </Button>

            {!rightDrawerOpen && (
              <IconButton size="small" onClick={() => setRightDrawerOpen(true)}>
                <ChevronLeft size={18} />
              </IconButton>
            )}
          </Stack>
        </Toolbar>

        {/* Preview Area */}
        <Box
          sx={{
            flex: 1,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            overflow: 'auto',
            bgcolor: 'grey.100',
            p: 4,
          }}
        >
          <Paper
            elevation={4}
            sx={{
              width: viewport.width,
              height: viewport.height,
              overflow: 'hidden',
              transition: 'width 0.3s, height 0.3s',
            }}
          >
            <ThemePreview theme={themeConfig} layout={layoutConfigFromQuery} width={viewport.width} height={viewport.height} />
          </Paper>
        </Box>
      </Box>

      {/* Right Drawer - Layout Controls */}
      <Drawer
        variant="persistent"
        anchor="right"
        open={rightDrawerOpen}
        sx={{
          width: RIGHT_DRAWER_WIDTH,
          flexShrink: 0,
          '& .MuiDrawer-paper': {
            width: RIGHT_DRAWER_WIDTH,
            boxSizing: 'border-box',
            position: 'relative',
            borderLeft: '1px solid',
            borderColor: 'divider',
          },
        }}
      >
        <Box sx={{height: '100%', display: 'flex', flexDirection: 'column'}}>
          {/* Drawer Header */}
          <Box
            sx={{
              p: 2,
              borderBottom: '1px solid',
              borderColor: 'divider',
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
            }}
          >
            <Typography variant="h6">{t('design:layout.preview', {defaultValue: 'Preview Layout'})}</Typography>
            <IconButton size="small" onClick={() => setRightDrawerOpen(false)}>
              <ChevronRight size={18} />
            </IconButton>
          </Box>

          {/* Layout Controls */}
          <LayoutControls
            layouts={layoutsData?.layouts || []}
            isLoading={isLoadingLayouts}
            selectedLayoutId={selectedLayoutId}
            onLayoutSelect={handleLayoutSelect}
            onCreateLayout={handleCreateLayout}
          />
        </Box>
      </Drawer>

      {/* Theme Save Dialog */}
      <ThemeSaveDialog
        open={saveDialogOpen}
        onClose={() => setSaveDialogOpen(false)}
        onSave={handleSaveTheme}
        baseThemeName={baseThemeName}
        currentName={baseThemeName}
        isSaving={createTheme.isPending || updateTheme.isPending}
      />
    </Box>
  );
}
