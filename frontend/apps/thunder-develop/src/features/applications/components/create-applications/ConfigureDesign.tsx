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

import {Box, Typography, Stack, Button, Avatar, Grid, Chip, Tooltip, useTheme} from '@wso2/oxygen-ui';
import {Info, Palette, Plus, Shuffle} from '@wso2/oxygen-ui-icons-react';
import type {JSX} from 'react';
import {useState, useMemo, useEffect} from 'react';
import {useTranslation} from 'react-i18next';
import generateAppLogoSuggestions from '../../utils/generateAppLogoSuggestion';
import generateAppPrimaryColorSuggestions from '../../utils/generateAppPrimaryColorSuggestions';
import useGetBrandings from '../../../branding/api/useGetBrandings';
import useGetBranding from '../../../branding/api/useGetBranding';
import BrandingConstants from '../../constants/branding-contants';

/**
 * Props for the ConfigureDesign component.
 */
export interface ConfigureDesignProps {
  /**
   * URL of the currently selected application logo
   */
  appLogo: string | null;

  /**
   * The currently selected brand color (hex format)
   */
  selectedColor: string;

  /**
   * Optional application name for display purposes
   */
  appName?: string;

  /**
   * Callback function when a logo is selected
   */
  onLogoSelect: (logoUrl: string) => void;

  /**
   * Callback function when a color is selected
   */
  onColorSelect: (color: string) => void;

  /**
   * Optional callback function when the initial logo is loaded
   */
  onInitialLogoLoad?: (logoUrl: string) => void;

  /**
   * Callback function to broadcast whether this step is ready to proceed
   */
  onReadyChange?: (isReady: boolean) => void;

  /**
   * Callback function when branding selection changes
   * Returns true if using DEFAULT branding, false if creating new
   */
  onBrandingSelectionChange?: (useDefaultBranding: boolean, defaultBrandingId?: string) => void;
}

export default function ConfigureDesign({
  appLogo,
  selectedColor,
  appName = undefined,
  onLogoSelect,
  onColorSelect,
  onInitialLogoLoad = undefined,
  onReadyChange = undefined,
  onBrandingSelectionChange = undefined,
}: ConfigureDesignProps): JSX.Element {
  const {t} = useTranslation();
  const theme = useTheme();

  const [logoSeed, setLogoSeed] = useState<number>(0);
  const [customColor, setCustomColor] = useState<string>(theme?.vars?.palette.common.black ?? '#000000');
  const [showCustomColorInput, setShowCustomColorInput] = useState<boolean>(false);
  const [showColorOptions, setShowColorOptions] = useState<boolean>(false);

  // Fetch brandings to check for DEFAULT branding
  const {data: brandingsData} = useGetBrandings({limit: 100});

  // Find the DEFAULT branding
  const defaultBranding = brandingsData?.brandings.find(
    (b) => b.displayName === BrandingConstants.DEFAULT_BRANDING_NAME,
  );

  // Fetch full DEFAULT branding details if it exists
  const {data: defaultBrandingDetails} = useGetBranding(defaultBranding?.id ?? '');

  // logoSeed is intentionally used as a dependency to trigger logo regeneration on shuffle
  // eslint-disable-next-line react-hooks/exhaustive-deps
  const logoSuggestions: string[] = useMemo(() => generateAppLogoSuggestions(8), [logoSeed]);

  const colorOptions: string[] = useMemo(() => generateAppPrimaryColorSuggestions(), []);

  /**
   * Set the first logo as default when component mounts or logos rotate.
   */
  useEffect(() => {
    if (logoSuggestions.length > 0 && onInitialLogoLoad) {
      onInitialLogoLoad(logoSuggestions[0]);
    }
  }, [logoSuggestions, onInitialLogoLoad]);

  /**
   * Apply DEFAULT branding color if it exists
   */
  useEffect(() => {
    if (defaultBrandingDetails?.preferences?.theme?.colorSchemes?.light?.colors?.primary?.main) {
      const defaultColor = defaultBrandingDetails.preferences.theme.colorSchemes.light.colors.primary.main;
      onColorSelect(defaultColor);
    }
  }, [defaultBrandingDetails, onColorSelect]);

  /**
   * Notify parent about branding selection
   */
  useEffect(() => {
    if (onBrandingSelectionChange) {
      // User is using DEFAULT branding if it exists and they haven't opted to pick a different color
      const useDefaultBranding = Boolean(defaultBrandingDetails && !showColorOptions);
      onBrandingSelectionChange(useDefaultBranding, defaultBranding?.id);
    }
  }, [defaultBrandingDetails, showColorOptions, defaultBranding, onBrandingSelectionChange]);

  /**
   * Broadcast readiness - Design step is always ready since it has default values
   */
  useEffect(() => {
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

  const handleColorSelect = (color: string): void => {
    setShowCustomColorInput(false);
    setCustomColor('');
    setShowColorOptions(true); // Keep color options visible once user selects
    onColorSelect(color);
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
        <Stack direction="row" alignItems="center" spacing={1}>
          <Info size={14} />
          <Typography variant="body1" color="text.secondary" sx={{mb: 4}}>
            {t('applications:onboarding.configure.design.subtitle')}
          </Typography>
        </Stack>
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
                      backgroundColor: isSelected ? selectedColor : theme.vars?.palette.grey[600],
                    }),
                    ...theme.applyStyles('dark', {
                      backgroundColor: isSelected ? selectedColor : theme.vars?.palette.grey[600],
                    }),
                  }}
                />
              </Tooltip>
            );
          })}
        </Stack>
      </Stack>

      {/* Color Selection */}
      <Stack direction="column" spacing={3}>
        <Stack direction="row" alignItems="center" spacing={1}>
          <Palette size={14} />
          <Typography variant="h6">{t('applications:onboarding.configure.design.color.title')}</Typography>
        </Stack>

        {/* Show DEFAULT branding color or empty state */}
        {defaultBrandingDetails && !showColorOptions ? (
          <Stack direction="column" spacing={2}>
            <Stack direction="row" alignItems="center" spacing={2}>
              <Box
                sx={{
                  width: 50,
                  height: 50,
                  borderRadius: '8px',
                  bgcolor: selectedColor,
                  border: `2px solid ${theme.vars?.palette.primary.main}`,
                }}
              />
              <Stack direction="column">
                <Typography variant="body2" color="text.secondary">
                  {appName ? (
                    <>
                      <Typography component="strong" fontWeight="bold">
                        {appName}
                      </Typography>{' '}
                      will use the default brand color
                    </>
                  ) : (
                    'Using the default brand color'
                  )}
                </Typography>
                <Typography variant="body1" fontWeight={500}>
                  {selectedColor}
                </Typography>
              </Stack>
            </Stack>
            <Button
              variant="outlined"
              size="medium"
              startIcon={<Palette size={14} />}
              onClick={(): void => setShowColorOptions(true)}
              sx={{alignSelf: 'flex-start'}}
            >
              Pick a different color
            </Button>
          </Stack>
        ) : (
          <>
            <Grid container spacing={1}>
              {colorOptions.map((color: string) => {
                const isSelected: boolean = selectedColor === color && !showCustomColorInput;

                return (
                  <Grid key={color}>
                    <Tooltip title={color} placement="top">
                      <Chip
                        label=""
                        onClick={(): void => handleColorSelect(color)}
                        sx={{
                          bgcolor: color,
                          width: 50,
                          height: 50,
                          borderRadius: '8px',
                          border: isSelected
                            ? `2px solid ${theme.vars?.palette.primary.main}`
                            : `1px solid ${theme.vars?.palette.divider}`,
                          '&:hover': {
                            bgcolor: color,
                            transform: 'scale(1.1)',
                            borderColor: theme.vars?.palette.primary.main,
                          },
                          transition: 'all 0.2s ease-in-out',
                          cursor: 'pointer',
                        }}
                      />
                    </Tooltip>
                  </Grid>
                );
              })}
            </Grid>

            {/* Custom Section */}
            <Stack direction="column" spacing={3}>
              <Typography variant="subtitle2" color="text.secondary">
                {t('applications:onboarding.configure.design.color.customLabel')}
              </Typography>
              <Box sx={{position: 'relative', width: 80, height: 80, display: 'inline-block'}}>
                <Box
                  component="input"
                  type="color"
                  value={customColor}
                  onChange={(e: React.ChangeEvent<HTMLInputElement>): void => {
                    const newColor = e.target.value;
                    setCustomColor(newColor);
                    setShowCustomColorInput(true);
                    onColorSelect(newColor);
                  }}
                  sx={{
                    position: 'absolute',
                    top: 0,
                    left: 0,
                    width: 50,
                    height: 50,
                    borderRadius: '8px',
                    border: `1px solid ${theme.vars?.palette.divider}`,
                    cursor: 'pointer',
                    opacity: customColor ? 1 : 0,
                    zIndex: customColor ? 1 : 0,
                    '&::-webkit-color-swatch-wrapper': {
                      padding: 0,
                    },
                    '&::-webkit-color-swatch': {
                      borderRadius: '7px',
                      border: 'none',
                    },
                    '&::-moz-color-swatch': {
                      borderRadius: '7px',
                      border: 'none',
                    },
                  }}
                />
                <Box
                  onClick={(): void => {
                    const input: HTMLInputElement | null = document.querySelector('input[type="color"]');

                    if (input instanceof HTMLInputElement) {
                      input.click();
                    }
                  }}
                  sx={{
                    position: 'absolute',
                    top: 0,
                    left: 0,
                    width: 50,
                    height: 50,
                    borderRadius: '8px',
                    border: `1px solid ${theme.vars?.palette.divider}`,
                    bgcolor: 'background.paper',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    cursor: 'pointer',
                    fontSize: '1.5rem',
                    fontWeight: 300,
                    color: 'text.secondary',
                    zIndex: 2,
                    '&:hover': {
                      transform: 'scale(1.1)',
                      borderColor: theme.vars?.palette.primary.main,
                    },
                    transition: 'all 0.2s ease-in-out',
                  }}
                >
                  <Plus size={28} />
                </Box>
              </Box>
            </Stack>
          </>
        )}
      </Stack>
    </Stack>
  );
}
