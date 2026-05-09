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

import {generateIconSuggestions, ResourceAvatar} from '@thunderid/components';
import {useGetThemes, useGetTheme, type ThemeListItem, type Theme} from '@thunderid/design';
import {
  Typography,
  Stack,
  Button,
  Card,
  Box,
  Grid,
  useTheme,
  Autocomplete,
  TextField,
  CircularProgress,
} from '@wso2/oxygen-ui';
import {Palette, Shuffle, Plus} from '@wso2/oxygen-ui-icons-react';
import type {JSX} from 'react';
import {useState, useMemo, useEffect} from 'react';
import {useTranslation} from 'react-i18next';
import ThemeThumbnail from '../../../design/components/themes/ThemeThumbnail';

/**
 * Props for the {@link ConfigureDesign} component.
 *
 * @public
 */
export interface ConfigureDesignProps {
  /**
   * URL or emoji of the currently selected application logo.
   */
  appLogo: string | null;

  /**
   * The ID of the currently selected theme (from API response)
   */
  themeId?: string | null;

  /**
   * The currently selected theme configuration (UI theme data only, not API response wrapper)
   */
  selectedTheme: Theme | null;

  /**
   * Callback function when a logo is selected
   */
  onLogoSelect: (logoUrl: string) => void;

  /**
   * Callback function when a theme is selected, receives theme ID and config separately
   */
  onThemeSelect: (themeId: string, themeConfig: Theme) => void;

  /**
   * Callback function to broadcast whether this step is ready to proceed
   */
  onReadyChange?: (isReady: boolean) => void;
}

/**
 * React component that renders the design customization step in the
 * application creation onboarding flow.
 *
 * This component allows users to customize their application's visual identity by:
 * 1. Selecting a logo from random emoji suggestions (with shuffle capability) or
 *    opening the full EmojiPicker modal via the "+" button
 * 2. Selecting a theme from available theme configurations
 *
 * @param props - The component props
 * @returns JSX element displaying the design customization interface
 *
 * @public
 */
export default function ConfigureDesign({
  appLogo,
  themeId: externallyProvidedThemeId = null,
  selectedTheme: selectedThemeProp,
  onLogoSelect,
  onThemeSelect,
  onReadyChange = undefined,
}: ConfigureDesignProps): JSX.Element {
  const {t} = useTranslation();
  const theme = useTheme();
  const {data: themesData, isLoading: loadingThemes} = useGetThemes({limit: 100});

  const [selectedThemeId, setSelectedThemeId] = useState<string | null>(externallyProvidedThemeId ?? null);
  const {data: selectedThemeDetails} = useGetTheme(selectedThemeId ?? '');

  const THEME_GRID_THRESHOLD = 8;
  const themeList = themesData?.themes ?? [];
  const hasThemes = Boolean(themeList.length);
  const useAutocomplete = themeList.length > THEME_GRID_THRESHOLD;
  const primaryColorLight: string =
    selectedThemeProp?.colorSchemes?.light?.palette?.primary?.main ?? theme.vars?.palette.primary.main ?? '';
  const primaryColorDark: string =
    selectedThemeProp?.colorSchemes?.dark?.palette?.primary?.main ??
    selectedThemeProp?.colorSchemes?.light?.palette?.primary?.main ??
    theme.vars?.palette.primary.main ??
    '';

  const [logoSeed, setLogoSeed] = useState<number>(0);

  // logoSeed is intentionally used as a dependency to trigger logo regeneration on shuffle

  // eslint-disable-next-line react-hooks/exhaustive-deps
  const logoSuggestions: string[] = useMemo((): string[] => generateIconSuggestions(8), [logoSeed]);

  // Derived: true when appLogo was chosen via the full picker (URL, or an emoji not in the inline suggestions)
  const isCustomLogo: boolean =
    Boolean(appLogo) && (!appLogo!.startsWith('emoji:') || !logoSuggestions.includes(appLogo!.slice(6)));

  /**
   * Auto-select the first logo when component mounts.
   */
  useEffect((): void => {
    if (logoSuggestions.length > 0 && !appLogo) {
      onLogoSelect(`emoji:${logoSuggestions[0]}`);
    }
  }, [logoSuggestions, appLogo, onLogoSelect]);

  /**
   * Auto-select the first theme when themes load and none is selected yet.
   */
  useEffect((): void => {
    if (themesData?.themes?.length && !selectedThemeId) {
      setSelectedThemeId(themesData.themes[0].id);
    }
  }, [themesData, selectedThemeId]);

  /**
   * Notify parent when theme details load.
   */
  useEffect((): void => {
    if (selectedThemeDetails) {
      onThemeSelect(selectedThemeDetails.id, selectedThemeDetails.theme);
    }
  }, [selectedThemeDetails, onThemeSelect]);

  /**
   * Broadcast readiness — Design step is always ready since it has default values.
   */
  useEffect((): void => {
    if (onReadyChange) {
      onReadyChange(true);
    }
  }, [onReadyChange]);

  const handleRotateLogos = (): void => {
    setLogoSeed((prev: number): number => prev + 1);
  };

  const handleLogoSelect = (logoValue: string): void => {
    onLogoSelect(logoValue);
  };

  const handleThemeSelect = (themeItem: ThemeListItem): void => {
    setSelectedThemeId(themeItem.id);
  };

  let themeSelectionContent: JSX.Element;

  if (!hasThemes) {
    themeSelectionContent = (
      <Stack
        direction="column"
        spacing={2}
        alignItems="center"
        sx={{
          p: 4,
          borderRadius: '12px',
          border: `1px dashed ${theme.vars?.palette.divider}`,
        }}
      >
        <Palette size={32} color={theme.vars?.palette.text.secondary} />
        <Typography variant="body1" color="text.secondary" textAlign="center">
          {t('applications:onboarding.configure.design.theme.emptyState')}
        </Typography>
        <Typography variant="caption" color="text.secondary" textAlign="center">
          {t('applications:onboarding.configure.design.theme.emptyStateHint')}
        </Typography>
      </Stack>
    );
  } else if (useAutocomplete) {
    themeSelectionContent = (
      <Autocomplete
        fullWidth
        options={themeList}
        getOptionLabel={(option) => (typeof option === 'string' ? option : option.displayName)}
        value={themeList.find((themeListItem) => themeListItem.id === selectedThemeId) ?? null}
        onChange={(_event, newValue): void => {
          if (newValue) handleThemeSelect(newValue);
        }}
        loading={loadingThemes}
        renderInput={(params) => (
          <TextField
            {...params}
            placeholder={t('applications:onboarding.configure.design.theme.title')}
            slotProps={{
              input: {
                ...params.InputProps,
                endAdornment: (
                  <>
                    {loadingThemes ? <CircularProgress color="inherit" size={20} /> : null}
                    {params.InputProps.endAdornment}
                  </>
                ),
              },
            }}
          />
        )}
      />
    );
  } else {
    themeSelectionContent = (
      <Grid container spacing={2}>
        {themeList.map((themeItem: ThemeListItem) => {
          const isSelected: boolean = selectedThemeId === themeItem.id;
          return (
            <Grid key={themeItem.id} size={{xs: 2, sm: 3, md: 4, lg: 3}}>
              <Card
                data-testid={`theme-card-${themeItem.id}`}
                onClick={(): void => handleThemeSelect(themeItem)}
                sx={{
                  cursor: 'pointer',
                  border: isSelected ? `2px solid ${theme.vars?.palette.primary.main}` : undefined,
                  '&:hover': {
                    borderColor: 'primary.main',
                    boxShadow: '0 4px 20px rgba(0,0,0,0.1)',
                    transform: 'translateY(-2px)',
                  },
                }}
              >
                <Box sx={{aspectRatio: '4/3', overflow: 'hidden', position: 'relative'}}>
                  <ThemeThumbnail theme={themeItem} />
                </Box>
                <Box sx={{px: 1.5, py: 1, borderTop: '1px solid', borderColor: 'divider'}}>
                  <Typography
                    variant="body2"
                    sx={{
                      fontWeight: isSelected ? 600 : 500,
                      fontSize: '0.8125rem',
                      overflow: 'hidden',
                      textOverflow: 'ellipsis',
                      whiteSpace: 'nowrap',
                    }}
                  >
                    {themeItem.displayName}
                  </Typography>
                </Box>
              </Card>
            </Grid>
          );
        })}
      </Grid>
    );
  }

  return (
    <Stack direction="column" spacing={4} data-testid="application-configure-design">
      <Stack direction="column" spacing={1}>
        <Typography variant="h1" gutterBottom>
          {t('applications:onboarding.configure.design.title')}
        </Typography>
        <Typography variant="subtitle1" gutterBottom>
          {t('applications:onboarding.configure.design.subtitle')}
        </Typography>
      </Stack>

      {/* Logo Selection */}
      <Stack direction="column" spacing={4}>
        <Stack direction="row" alignItems="center" justifyContent="space-between">
          <Typography variant="h6">{t('applications:onboarding.configure.design.logo.title')}</Typography>
          <Button
            size="small"
            variant="text"
            startIcon={<Shuffle size={14} />}
            onClick={handleRotateLogos}
            sx={{minWidth: 'auto'}}
          >
            {t('applications:onboarding.configure.design.logo.shuffle')}
          </Button>
        </Stack>

        {/* Inline emoji suggestions + "+" to open full picker */}
        <Stack direction="row" sx={{flexWrap: 'wrap', gap: 2}} alignItems="center">
          {logoSuggestions.map((emoji: string) => {
            const isSelected: boolean = appLogo === `emoji:${emoji}`;

            return (
              <ResourceAvatar
                key={emoji}
                value={emoji}
                size={isSelected ? 70 : 50}
                onClick={(): void => handleLogoSelect(`emoji:${emoji}`)}
                sx={{
                  cursor: 'pointer',
                  border: isSelected
                    ? `2px solid ${theme.vars?.palette.primary.main}`
                    : `1px solid ${theme.vars?.palette.divider}`,
                  p: 1,
                  transition: 'all 0.2s ease-in-out',
                  '&:hover': {
                    transform: 'scale(1.1)',
                    borderColor: theme.vars?.palette.primary.main,
                  },
                  ...theme.applyStyles('light', {
                    backgroundColor: isSelected ? primaryColorLight : theme.vars?.palette.grey[600],
                  }),
                  ...theme.applyStyles('dark', {
                    backgroundColor: isSelected ? primaryColorDark : theme.vars?.palette.grey[600],
                  }),
                }}
              />
            );
          })}

          {/* Button to open full picker — shows selected custom logo when one is active */}
          <ResourceAvatar
            value={isCustomLogo ? appLogo! : undefined}
            fallback={<Plus size={20} />}
            size={isCustomLogo ? 70 : 50}
            onSelect={handleLogoSelect}
            editAriaLabel={t('applications:onboarding.configure.design.logo.chooseLogo')}
            sx={{
              border: isCustomLogo
                ? `2px solid ${theme.vars?.palette.primary.main}`
                : `1px dashed ${theme.vars?.palette.divider}`,
              color: 'text.secondary',
              transition: 'all 0.2s ease-in-out',
              '&:hover': {
                borderColor: theme.vars?.palette.primary.main,
                color: 'primary.main',
                bgcolor: isCustomLogo ? undefined : 'action.hover',
              },
              ...(isCustomLogo
                ? {
                    ...theme.applyStyles('light', {backgroundColor: primaryColorLight}),
                    ...theme.applyStyles('dark', {backgroundColor: primaryColorDark}),
                  }
                : {bgcolor: 'transparent'}),
            }}
          />
        </Stack>
      </Stack>

      {/* Theme Selection */}
      <Stack direction="column" spacing={3}>
        <Stack direction="row" alignItems="center" spacing={1}>
          <Palette size={14} />
          <Typography variant="h6">{t('applications:onboarding.configure.design.theme.title')}</Typography>
        </Stack>

        {themeSelectionContent}
      </Stack>
    </Stack>
  );
}
