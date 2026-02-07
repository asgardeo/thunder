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
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Accordion,
  AccordionSummary,
  AccordionDetails,
  Typography,
  Stack,
  CircularProgress,
  TextField,
  ToggleButtonGroup,
  ToggleButton,
  Chip,
  Divider,
} from '@wso2/oxygen-ui';
import {ChevronDown, Sun, Moon, Palette} from '@wso2/oxygen-ui-icons-react';
import {useTranslation} from 'react-i18next';
import SimpleColorPicker from '@/features/design-system/components/SimpleColorPicker';
import type {ThemeConfig, Theme} from '@/features/themes/models/theme';

/**
 * Props for ThemeControls component
 */
interface ThemeControlsProps {
  /** Available themes to select as base */
  themes: Theme[];
  /** Whether themes are loading */
  isLoading: boolean;
  /** Currently selected base theme ID */
  selectedBaseThemeId: string | null;
  /** Callback when base theme is selected */
  onBaseThemeSelect: (themeId: string | null) => void;
  /** Current theme configuration being edited */
  themeConfig: ThemeConfig;
  /** Callback when theme config changes */
  onThemeConfigChange: (config: ThemeConfig) => void;
}

/**
 * Theme controls component for customizing theme properties
 */
export default function ThemeControls({
  themes,
  isLoading,
  selectedBaseThemeId,
  onBaseThemeSelect,
  themeConfig,
  onThemeConfigChange,
}: ThemeControlsProps): JSX.Element {
  const {t} = useTranslation();

  // Track which color scheme we're editing (light or dark)
  const [editingScheme, setEditingScheme] = useState<'light' | 'dark'>('dark');

  // Update editing scheme when themeConfig changes (e.g., when loading a new theme)
  useEffect(() => {
    if (themeConfig?.defaultColorScheme) {
      setEditingScheme(themeConfig.defaultColorScheme);
    }
  }, [themeConfig?.defaultColorScheme]);

  const handleColorChange = (path: string, value: string) => {
    const keys = path.split('.');
    const newConfig = JSON.parse(JSON.stringify(themeConfig)); // Deep clone
    let current: any = newConfig;

    for (let i = 0; i < keys.length - 1; i++) {
      if (!current[keys[i]]) {
        current[keys[i]] = {};
      }
      current = current[keys[i]];
    }
    current[keys[keys.length - 1]] = value;

    onThemeConfigChange(newConfig);
  };

  const handleBorderRadiusChange = (value: string) => {
    if (!themeConfig) return;
    const newConfig = JSON.parse(JSON.stringify(themeConfig)); // Deep clone
    if (!newConfig.shape) {
      newConfig.shape = {};
    }
    newConfig.shape.borderRadius = value;
    onThemeConfigChange(newConfig);
  };

  const handleFontFamilyChange = (value: string) => {
    if (!themeConfig) return;
    const newConfig = JSON.parse(JSON.stringify(themeConfig)); // Deep clone
    if (!newConfig.typography) {
      newConfig.typography = {};
    }
    newConfig.typography.fontFamily = value;
    onThemeConfigChange(newConfig);
  };

  const handleDefaultColorSchemeChange = (value: 'light' | 'dark') => {
    if (!themeConfig) return;
    const newConfig = JSON.parse(JSON.stringify(themeConfig)); // Deep clone
    newConfig.defaultColorScheme = value;
    onThemeConfigChange(newConfig);
  };

  // Get color values safely
  const getColor = (path: string): string => {
    if (!themeConfig) return '#000000';
    const keys = path.split('.');
    let current: any = themeConfig;
    for (const key of keys) {
      if (!current?.[key]) return '#000000';
      current = current[key];
    }
    return current as string;
  };

  // Check if themeConfig is valid
  if (!themeConfig) {
    return (
      <Box sx={{height: '100%', display: 'flex', flexDirection: 'column'}}>
        <Box sx={{p: 2, borderBottom: '1px solid', borderColor: 'divider'}}>
          <FormControl fullWidth size="small">
            <InputLabel>{t('design:theme.baseTheme', {defaultValue: 'Base Theme'})}</InputLabel>
            <Select
              value={selectedBaseThemeId || 'scratch'}
              onChange={(e) => onBaseThemeSelect(e.target.value === 'scratch' ? null : e.target.value)}
              label={t('design:theme.baseTheme', {defaultValue: 'Base Theme'})}
              disabled={isLoading}
            >
              <MenuItem value="scratch">
                {t('design:theme.createFromScratch', {defaultValue: 'Create from scratch'})}
              </MenuItem>
              {themes.map((theme) => (
                <MenuItem key={theme.id} value={theme.id}>
                  {theme.displayName}
                </MenuItem>
              ))}
            </Select>
          </FormControl>
        </Box>
        <Box sx={{flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center', p: 4}}>
          <CircularProgress />
        </Box>
      </Box>
    );
  }

  return (
    <Box sx={{height: '100%', display: 'flex', flexDirection: 'column'}}>
      {/* Base Theme Selector */}
      <Box sx={{p: 2, borderBottom: '1px solid', borderColor: 'divider'}}>
        <Stack spacing={2}>
          <FormControl fullWidth size="small">
            <InputLabel>{t('design:theme.baseTheme', {defaultValue: 'Base Theme'})}</InputLabel>
            <Select
              value={selectedBaseThemeId || 'scratch'}
              onChange={(e) => onBaseThemeSelect(e.target.value === 'scratch' ? null : e.target.value)}
              label={t('design:theme.baseTheme', {defaultValue: 'Base Theme'})}
              disabled={isLoading}
            >
              <MenuItem value="scratch">
                {t('design:theme.createFromScratch', {defaultValue: 'Create from scratch'})}
              </MenuItem>
              {themes.map((theme) => (
                <MenuItem key={theme.id} value={theme.id}>
                  {theme.displayName}
                </MenuItem>
              ))}
            </Select>
          </FormControl>

          {selectedBaseThemeId && (
            <Chip
              icon={<Palette size={16} />}
              label={themes.find((t) => t.id === selectedBaseThemeId)?.displayName || 'Unknown'}
              size="small"
              color="primary"
              variant="outlined"
            />
          )}
        </Stack>
      </Box>

      {/* Theme Customization Controls */}
      <Box sx={{flex: 1, overflow: 'auto', p: 2}}>
        {isLoading ? (
          <Box display="flex" justifyContent="center" p={4}>
            <CircularProgress />
          </Box>
        ) : (
          <Stack spacing={2}>
            {/* Color Scheme Toggle */}
            <Box>
              <Typography variant="caption" color="text.secondary" sx={{mb: 1, display: 'block'}}>
                {t('design:theme.editingColorScheme', {defaultValue: 'Editing Color Scheme'})}
              </Typography>
              <ToggleButtonGroup
                value={editingScheme}
                exclusive
                onChange={(_, value) => value && setEditingScheme(value)}
                size="small"
                fullWidth
              >
                <ToggleButton value="light">
                  <Sun size={16} style={{marginRight: 8}} />
                  {t('design:theme.light', {defaultValue: 'Light'})}
                </ToggleButton>
                <ToggleButton value="dark">
                  <Moon size={16} style={{marginRight: 8}} />
                  {t('design:theme.dark', {defaultValue: 'Dark'})}
                </ToggleButton>
              </ToggleButtonGroup>
            </Box>

            <Divider />

            {/* Default Color Scheme */}
            <Accordion defaultExpanded>
              <AccordionSummary expandIcon={<ChevronDown size={20} />}>
                <Typography variant="subtitle2">
                  {t('design:theme.defaultMode', {defaultValue: 'Default Mode'})}
                </Typography>
              </AccordionSummary>
              <AccordionDetails>
                <FormControl fullWidth size="small">
                  <InputLabel>
                    {t('design:theme.defaultColorScheme', {defaultValue: 'Default Color Scheme'})}
                  </InputLabel>
                  <Select
                    value={themeConfig?.defaultColorScheme || 'light'}
                    onChange={(e) => handleDefaultColorSchemeChange(e.target.value)}
                    label={t('design:theme.defaultColorScheme', {defaultValue: 'Default Color Scheme'})}
                  >
                    <MenuItem value="light">
                      <Box sx={{display: 'flex', alignItems: 'center', gap: 1}}>
                        <Sun size={16} />
                        {t('design:theme.light', {defaultValue: 'Light'})}
                      </Box>
                    </MenuItem>
                    <MenuItem value="dark">
                      <Box sx={{display: 'flex', alignItems: 'center', gap: 1}}>
                        <Moon size={16} />
                        {t('design:theme.dark', {defaultValue: 'Dark'})}
                      </Box>
                    </MenuItem>
                  </Select>
                </FormControl>
              </AccordionDetails>
            </Accordion>
            {/* Primary Colors */}
            <Accordion defaultExpanded>
              <AccordionSummary expandIcon={<ChevronDown size={20} />}>
                <Typography variant="subtitle2">
                  {t('design:theme.primaryColors', {defaultValue: 'Primary Colors'})}
                </Typography>
              </AccordionSummary>
              <AccordionDetails>
                <Stack spacing={2}>
                  <SimpleColorPicker
                    label={t('design:theme.primaryMain', {defaultValue: 'Primary Main'})}
                    value={getColor(`colorSchemes.${editingScheme}.colors.primary.main`)}
                    onChange={(value) => handleColorChange(`colorSchemes.${editingScheme}.colors.primary.main`, value)}
                  />
                  <SimpleColorPicker
                    label={t('design:theme.primaryDark', {defaultValue: 'Primary Dark'})}
                    value={getColor(`colorSchemes.${editingScheme}.colors.primary.dark`)}
                    onChange={(value) => handleColorChange(`colorSchemes.${editingScheme}.colors.primary.dark`, value)}
                  />
                  <SimpleColorPicker
                    label={t('design:theme.contrastText', {defaultValue: 'Contrast Text'})}
                    value={getColor(`colorSchemes.${editingScheme}.colors.primary.contrastText`)}
                    onChange={(value) =>
                      handleColorChange(`colorSchemes.${editingScheme}.colors.primary.contrastText`, value)
                    }
                  />
                </Stack>
              </AccordionDetails>
            </Accordion>

            {/* Secondary Colors */}
            <Accordion>
              <AccordionSummary expandIcon={<ChevronDown size={20} />}>
                <Typography variant="subtitle2">
                  {t('design:theme.secondaryColors', {defaultValue: 'Secondary Colors'})}
                </Typography>
              </AccordionSummary>
              <AccordionDetails>
                <Stack spacing={2}>
                  <SimpleColorPicker
                    label={t('design:theme.secondaryMain', {defaultValue: 'Secondary Main'})}
                    value={getColor(`colorSchemes.${editingScheme}.colors.secondary.main`)}
                    onChange={(value) =>
                      handleColorChange(`colorSchemes.${editingScheme}.colors.secondary.main`, value)
                    }
                  />
                  <SimpleColorPicker
                    label={t('design:theme.secondaryDark', {defaultValue: 'Secondary Dark'})}
                    value={getColor(`colorSchemes.${editingScheme}.colors.secondary.dark`)}
                    onChange={(value) =>
                      handleColorChange(`colorSchemes.${editingScheme}.colors.secondary.dark`, value)
                    }
                  />
                  <SimpleColorPicker
                    label={t('design:theme.contrastText', {defaultValue: 'Contrast Text'})}
                    value={getColor(`colorSchemes.${editingScheme}.colors.secondary.contrastText`)}
                    onChange={(value) =>
                      handleColorChange(`colorSchemes.${editingScheme}.colors.secondary.contrastText`, value)
                    }
                  />
                </Stack>
              </AccordionDetails>
            </Accordion>

            {/* Background Colors */}
            <Accordion>
              <AccordionSummary expandIcon={<ChevronDown size={20} />}>
                <Typography variant="subtitle2">
                  {t('design:theme.backgroundColors', {defaultValue: 'Background Colors'})}
                </Typography>
              </AccordionSummary>
              <AccordionDetails>
                <Stack spacing={2}>
                  <SimpleColorPicker
                    label={t('design:theme.backgroundDefault', {defaultValue: 'Background Default'})}
                    value={getColor(`colorSchemes.${editingScheme}.colors.background.default`)}
                    onChange={(value) =>
                      handleColorChange(`colorSchemes.${editingScheme}.colors.background.default`, value)
                    }
                  />
                  <SimpleColorPicker
                    label={t('design:theme.backgroundPaper', {defaultValue: 'Background Paper'})}
                    value={getColor(`colorSchemes.${editingScheme}.colors.background.paper`)}
                    onChange={(value) =>
                      handleColorChange(`colorSchemes.${editingScheme}.colors.background.paper`, value)
                    }
                  />
                </Stack>
              </AccordionDetails>
            </Accordion>

            {/* Text Colors */}
            <Accordion>
              <AccordionSummary expandIcon={<ChevronDown size={20} />}>
                <Typography variant="subtitle2">
                  {t('design:theme.textColors', {defaultValue: 'Text Colors'})}
                </Typography>
              </AccordionSummary>
              <AccordionDetails>
                <Stack spacing={2}>
                  <SimpleColorPicker
                    label={t('design:theme.textPrimary', {defaultValue: 'Text Primary'})}
                    value={getColor(`colorSchemes.${editingScheme}.colors.text.primary`)}
                    onChange={(value) => handleColorChange(`colorSchemes.${editingScheme}.colors.text.primary`, value)}
                  />
                  <SimpleColorPicker
                    label={t('design:theme.textSecondary', {defaultValue: 'Text Secondary'})}
                    value={getColor(`colorSchemes.${editingScheme}.colors.text.secondary`)}
                    onChange={(value) =>
                      handleColorChange(`colorSchemes.${editingScheme}.colors.text.secondary`, value)
                    }
                  />
                </Stack>
              </AccordionDetails>
            </Accordion>

            {/* Shape & Typography */}
            <Accordion>
              <AccordionSummary expandIcon={<ChevronDown size={20} />}>
                <Typography variant="subtitle2">
                  {t('design:theme.shapeAndTypography', {defaultValue: 'Shape & Typography'})}
                </Typography>
              </AccordionSummary>
              <AccordionDetails>
                <Stack spacing={2}>
                  <TextField
                    fullWidth
                    size="small"
                    label={t('design:theme.borderRadius', {defaultValue: 'Border Radius'})}
                    value={themeConfig?.shape?.borderRadius || '8px'}
                    onChange={(e) => handleBorderRadiusChange(e.target.value)}
                    placeholder="8px"
                    helperText={t('design:theme.borderRadiusHelp', {defaultValue: 'Use format: 8px, 12px, etc.'})}
                  />
                  <TextField
                    fullWidth
                    size="small"
                    label={t('design:theme.fontFamily', {defaultValue: 'Font Family'})}
                    value={themeConfig?.typography?.fontFamily || 'Inter, sans-serif'}
                    onChange={(e) => handleFontFamilyChange(e.target.value)}
                    placeholder="Inter, sans-serif"
                    helperText={t('design:theme.fontFamilyHelp', {defaultValue: 'Comma-separated font stack'})}
                  />
                </Stack>
              </AccordionDetails>
            </Accordion>
          </Stack>
        )}
      </Box>
    </Box>
  );
}
