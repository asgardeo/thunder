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

import {Typography, Stack, Button, Avatar, Tooltip, Radio, Card, CardActionArea, useTheme} from '@wso2/oxygen-ui';
import {Palette, Shuffle} from '@wso2/oxygen-ui-icons-react';
import type {JSX} from 'react';
import {useState, useMemo, useEffect} from 'react';
import {useTranslation} from 'react-i18next';
import {useGetThemes, useGetTheme, type ThemeListItem, type ThemeConfig} from '@thunder/shared-design';
import generateAppLogoSuggestions from '../../utils/generateAppLogoSuggestion';

/**
 * Props for the {@link ConfigureDesign} component.
 *
 * @public
 */
export interface ConfigureDesignProps {
  /**
   * URL of the currently selected application logo
   */
  appLogo: string | null;

  /**
   * The ID of the currently selected theme (from API response)
   */
  themeId?: string | null;

  /**
   * The currently selected theme configuration (UI theme data only, not API response wrapper)
   */
  selectedTheme: ThemeConfig | null;

  /**
   * Callback function when a logo is selected
   */
  onLogoSelect: (logoUrl: string) => void;

  /**
   * Callback function when a theme is selected, receives theme ID and config separately
   */
  onThemeSelect: (themeId: string, themeConfig: ThemeConfig) => void;

  /**
   * Optional callback function when the initial logo is loaded
   */
  onInitialLogoLoad?: (logoUrl: string) => void;

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
 * 1. Selecting a logo from AI-generated avatar suggestions (with shuffle capability)
 * 2. Selecting a theme from available theme configurations
 *
 * The component displays a grid of logo avatars for selection, with the ability to
 * shuffle logos for new suggestions. It fetches available themes and displays them
 * as selectable cards. The step is always ready since default selections are provided.
 *
 * @param props - The component props
 * @param props.appLogo - The currently selected logo URL
 * @param props.themeId - The currently selected theme ID
 * @param props.selectedTheme - The currently selected theme configuration
 * @param props.onLogoSelect - Callback when logo is selected
 * @param props.onThemeSelect - Callback when theme is selected (receives ID and config)
 * @param props.onInitialLogoLoad - Optional callback when initial logo loads
 * @param props.onReadyChange - Optional callback for step readiness
 *
 * @returns JSX element displaying the design customization interface
 *
 * @example
 * ```tsx
 * import ConfigureDesign from './ConfigureDesign';
 *
 * function OnboardingFlow() {
 *   const [logo, setLogo] = useState<string | null>(null);
 *   const [themeId, setThemeId] = useState<string | null>(null);
 *   const [themeConfig, setThemeConfig] = useState<ThemeConfig | null>(null);
 *
 *   return (
 *     <ConfigureDesign
 *       appLogo={logo}
 *       themeId={themeId}
 *       selectedTheme={themeConfig}
 *       onLogoSelect={setLogo}
 *       onThemeSelect={(id, config) => {
 *         setThemeId(id);
 *         setThemeConfig(config);
 *       }}
 *       onInitialLogoLoad={(url) => console.log('Initial logo:', url)}
 *     />
 *   );
 * }
 * ```
 *
 * @public
 */
export default function ConfigureDesign({
  appLogo,
  themeId: externallyProvidedThemeId = null,
  selectedTheme: selectedThemeProp,
  onLogoSelect,
  onThemeSelect,
  onInitialLogoLoad = undefined,
  onReadyChange = undefined,
}: ConfigureDesignProps): JSX.Element {
  const {t} = useTranslation();
  const theme = useTheme();
  const {data: themesData} = useGetThemes({limit: 100});

  const [selectedThemeId, setSelectedThemeId] = useState<string | null>(externallyProvidedThemeId ?? null);
  const {data: selectedThemeDetails} = useGetTheme(selectedThemeId ?? '');

  const hasThemes = Boolean(themesData?.themes?.length);
  const primaryColorLight: string =
    selectedThemeProp?.colorSchemes?.light?.colors?.primary?.main ?? theme.vars?.palette.primary.main ?? '';
  const primaryColorDark: string =
    selectedThemeProp?.colorSchemes?.dark?.colors?.primary?.main ??
    selectedThemeProp?.colorSchemes?.light?.colors?.primary?.main ??
    theme.vars?.palette.primary.main ??
    '';

  const [logoSeed, setLogoSeed] = useState<number>(0);

  // logoSeed is intentionally used as a dependency to trigger logo regeneration on shuffle
  // eslint-disable-next-line react-hooks/exhaustive-deps
  const logoSuggestions: string[] = useMemo((): string[] => generateAppLogoSuggestions(8), [logoSeed]);

  /**
   * Set the first logo as default when component mounts, or when the currently selected
   * logo is no longer available in the shuffled suggestions.
   */
  useEffect((): void => {
    if (logoSuggestions.length > 0 && onInitialLogoLoad) {
      // Only auto-select if there's no current selection, or current selection is not in the new suggestions
      if (!appLogo || !logoSuggestions.includes(appLogo)) {
        onInitialLogoLoad(logoSuggestions[0]);
      }
    }
  }, [logoSuggestions, onInitialLogoLoad, appLogo]);

  /**
   * Auto-select the first theme when themes load and none is selected yet
   */
  useEffect((): void => {
    if (themesData?.themes?.length && !selectedThemeId) {
      setSelectedThemeId(themesData.themes[0].id);
    }
  }, [themesData, selectedThemeId]);

  /**
   * Notify parent when theme details load
   */
  useEffect((): void => {
    if (selectedThemeDetails) {
      onThemeSelect(selectedThemeDetails.id, selectedThemeDetails.theme);
    }
  }, [selectedThemeDetails, onThemeSelect]);

  /**
   * Broadcast readiness - Design step is always ready since it has default values
   */
  useEffect((): void => {
    if (onReadyChange) {
      onReadyChange(true);
    }
  }, [onReadyChange]);

  const handleRotateLogos = (): void => {
    setLogoSeed((prev: number): number => prev + 1);
  };

  const handleLogoSelect = (logoUrl: string): void => {
    onLogoSelect(logoUrl);
  };

  const handleThemeSelect = (themeItem: ThemeListItem): void => {
    setSelectedThemeId(themeItem.id);
  };

  const getAnimalName = (logoUrl: string): string => {
    const match: RegExpExecArray | null = /\/([a-z]+)_lg\.png$/.exec(logoUrl);

    if (match) {
      return match[1].charAt(0).toUpperCase() + match[1].slice(1);
    }

    return t('common:dictionary.unknown');
  };

  return (
    <Stack direction="column" spacing={4}>
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

        {/* Logo Preview and Suggestions - Inline */}
        <Stack direction="row" sx={{flexWrap: 'wrap', gap: 2}}>
          {logoSuggestions.map((logoUrl: string) => {
            const isSelected: boolean = appLogo === logoUrl;

            return (
              <Tooltip key={logoUrl} title={getAnimalName(logoUrl)} placement="top">
                <Avatar
                  src={logoUrl}
                  onClick={(): void => handleLogoSelect(logoUrl)}
                  sx={{
                    width: isSelected ? 70 : 50,
                    height: isSelected ? 70 : 50,
                    cursor: 'pointer',
                    border: isSelected
                      ? `2px solid ${theme.vars?.palette.primary.main}`
                      : `1px solid ${theme.vars?.palette.divider}`,
                    p: 1,
                    '&:hover': {
                      transform: 'scale(1.1)',
                      borderColor: theme.vars?.palette.primary.main,
                    },
                    transition: 'all 0.2s ease-in-out',
                    ...theme.applyStyles('light', {
                      backgroundColor: isSelected ? primaryColorLight : theme.vars?.palette.grey[600],
                    }),
                    ...theme.applyStyles('dark', {
                      backgroundColor: isSelected ? primaryColorDark : theme.vars?.palette.grey[600],
                    }),
                  }}
                />
              </Tooltip>
            );
          })}
        </Stack>
      </Stack>

      {/* Theme Selection */}
      <Stack direction="column" spacing={3}>
        <Stack direction="row" alignItems="center" spacing={1}>
          <Palette size={14} />
          <Typography variant="h6">{t('applications:onboarding.configure.design.theme.title')}</Typography>
        </Stack>

        {hasThemes ? (
          <Stack direction="column" spacing={3}>
            {/* Theme Cards */}
            <Stack direction="column" spacing={1}>
              {themesData?.themes?.map((themeItem: ThemeListItem) => {
                const isSelected: boolean = selectedThemeId === themeItem.id;

                return (
                  <Card
                    key={themeItem.id}
                    data-testid={`theme-card-${themeItem.id}`}
                    variant="outlined"
                    sx={{
                      border: isSelected
                        ? `2px solid ${theme.vars?.palette.primary.main}`
                        : `1px solid ${theme.vars?.palette.divider}`,
                      borderRadius: '8px',
                      transition: 'all 0.15s ease-in-out',
                      '&:hover': {
                        borderColor: theme.vars?.palette.primary.main,
                      },
                    }}
                  >
                    <CardActionArea
                      onClick={(): void => handleThemeSelect(themeItem)}
                      sx={{display: 'flex', alignItems: 'center', gap: 1.5, p: 1.5, justifyContent: 'flex-start'}}
                    >
                      <Radio checked={isSelected} size="small" sx={{p: 0}} />
                      <Stack direction="column" sx={{minWidth: 0}}>
                        <Typography variant="body2" fontWeight={isSelected ? 600 : 400} noWrap>
                          {themeItem.displayName}
                        </Typography>
                        <Typography variant="caption" color="text.secondary" noWrap>
                          {themeItem.description ?? t('applications:onboarding.configure.design.theme.noDescription')}
                        </Typography>
                      </Stack>
                    </CardActionArea>
                  </Card>
                );
              })}
            </Stack>
          </Stack>
        ) : (
          /* No themes configured - empty state */
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
        )}
      </Stack>
    </Stack>
  );
}
