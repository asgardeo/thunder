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

import {Box, Typography, Stack, Card, CardActionArea, CardContent, Divider} from '@wso2/oxygen-ui';
import type {JSX} from 'react';
import {useEffect} from 'react';
import {useTranslation} from 'react-i18next';
import {TokenEndpointAuthMethods, type OAuth2Config} from '../../models/oauth';
import type {ApplicationTemplate} from '../../models/application-templates';
import {TechnologyApplicationTemplate, PlatformApplicationTemplate} from '../../models/application-templates';
import TechnologyBasedApplicationTemplateMetadata from '../../config/TechnologyBasedApplicationTemplateMetadata';
import PlatformBasedApplicationTemplateMetadata from '../../config/PlatformBasedApplicationTemplateMetadata';
import inferApplicationTemplateTechnologyFromConfig from '../../utils/inferApplicationTemplateTechnologyFromConfig';
import useApplicationCreate from '../../contexts/ApplicationCreate/useApplicationCreate';
import {ApplicationCreateFlowSignInApproach} from '../../models/application-create-flow';

const TechnologyBasedTemplates: Record<TechnologyApplicationTemplate, ApplicationTemplate> =
  TechnologyBasedApplicationTemplateMetadata.reduce(
    (acc, item) => ({
      ...acc,
      [item.value]: item.template,
    }),
    {} as Record<TechnologyApplicationTemplate, ApplicationTemplate>,
  );

const PlatformBasedTemplates: Record<PlatformApplicationTemplate, ApplicationTemplate> =
  PlatformBasedApplicationTemplateMetadata.reduce(
    (acc, item) => ({
      ...acc,
      [item.value]: item.template,
    }),
    {} as Record<PlatformApplicationTemplate, ApplicationTemplate>,
  );

/**
 * Props for the {@link ConfigureStack} component.
 *
 * @public
 */
export interface ConfigureStackProps {
  /**
   * OAuth configuration
   */
  oauthConfig: OAuth2Config | null;

  /**
   * Callback function when OAuth configuration changes
   */
  onOAuthConfigChange: (config: OAuth2Config | null) => void;

  /**
   * Callback function to broadcast whether this step is ready to proceed
   */
  onReadyChange?: (isReady: boolean) => void;

  /**
   * Configuration for which stack types to show
   */
  stackTypes?: {
    technology?: boolean;
    platform?: boolean;
  };
}

/**
 * React component that renders the technology/platform stack selection step in the
 * application creation onboarding flow.
 *
 * This component allows users to select their application's technology stack by choosing:
 * 1. Technology templates (React, Next.js, or Other) - can be hidden via stackTypes prop
 * 2. Platform templates (Browser, Server, Mobile, Backend) - can be hidden via stackTypes prop
 *
 * The component manages OAuth2 configuration based on the selected template, automatically
 * configuring grant types, PKCE requirements, and other OAuth settings. When the technology
 * section is hidden, it automatically selects the first platform template. Templates can
 * be pre-filled with redirect URIs or left empty to require configuration in the next step.
 *
 * The step is marked as ready once a valid template is selected. For "Other" technology,
 * a platform must also be selected.
 *
 * @param props - The component props
 * @param props.oauthConfig - Current OAuth2 configuration
 * @param props.onOAuthConfigChange - Callback when OAuth config changes based on template selection
 * @param props.onReadyChange - Optional callback to notify parent of step readiness
 * @param props.stackTypes - Configuration to show/hide technology and platform sections
 *
 * @returns JSX element displaying the technology/platform stack selection interface
 *
 * @example
 * ```tsx
 * import ConfigureStack from './ConfigureStack';
 * import { getDefaultOAuthConfig } from '../../models/oauth';
 *
 * function OnboardingFlow() {
 *   const [oauthConfig, setOAuthConfig] = useState(getDefaultOAuthConfig());
 *
 *   return (
 *     <ConfigureStack
 *       oauthConfig={oauthConfig}
 *       onOAuthConfigChange={setOAuthConfig}
 *       onReadyChange={(isReady) => console.log('Ready:', isReady)}
 *       stackTypes={{ technology: false, platform: true }}
 *     />
 *   );
 * }
 * ```
 *
 * @public
 */
export default function ConfigureStack({
  oauthConfig,
  onOAuthConfigChange,
  onReadyChange = undefined,
  stackTypes = {technology: true, platform: true},
}: ConfigureStackProps): JSX.Element {
  const {t} = useTranslation();

  const {
    selectedTechnology,
    setSelectedTechnology,
    selectedPlatform,
    setSelectedPlatform,
    setSelectedTemplateConfig,
    signInApproach,
  } = useApplicationCreate();

  const defaultTechnology: TechnologyApplicationTemplate =
    TechnologyBasedApplicationTemplateMetadata[0]?.value ?? TechnologyApplicationTemplate.REACT;
  const inferredTechnology: TechnologyApplicationTemplate = inferApplicationTemplateTechnologyFromConfig(oauthConfig);
  const defaultPlatform: PlatformApplicationTemplate =
    PlatformBasedApplicationTemplateMetadata[0]?.value ?? PlatformApplicationTemplate.BROWSER;

  const getResolvedTechnology = (): TechnologyApplicationTemplate => {
    if (selectedTechnology) {
      return selectedTechnology;
    }
    if (selectedPlatform) {
      return TechnologyApplicationTemplate.OTHER;
    }
    if (oauthConfig) {
      return inferredTechnology;
    }
    if (stackTypes.technology) {
      return defaultTechnology;
    }
    return TechnologyApplicationTemplate.OTHER;
  };

  const resolvedTechnology: TechnologyApplicationTemplate = getResolvedTechnology();

  const platformForTemplate: PlatformApplicationTemplate =
    selectedPlatform ?? (!stackTypes.technology ? defaultPlatform : (selectedPlatform ?? defaultPlatform));

  const technologyConfig: ApplicationTemplate =
    resolvedTechnology === TechnologyApplicationTemplate.OTHER
      ? PlatformBasedTemplates[platformForTemplate]
      : TechnologyBasedTemplates[resolvedTechnology];

  /**
   * Update OAuth2 configuration when technology or platform changes.
   */
  useEffect((): void => {
    setSelectedTemplateConfig(technologyConfig);

    const oauthInboundConfig: OAuth2Config = technologyConfig.inbound_auth_config?.[0]?.config ?? {
      public_client: false,
      pkce_required: false,
      grant_types: [],
      response_types: [],
      redirect_uris: [],
      token_endpoint_auth_method: TokenEndpointAuthMethods.CLIENT_SECRET_BASIC,
    };

    onOAuthConfigChange({
      public_client: oauthInboundConfig.public_client,
      pkce_required: oauthInboundConfig.pkce_required,
      grant_types: [...oauthInboundConfig.grant_types],
      response_types: [...(oauthInboundConfig.response_types ?? [])],
      redirect_uris: oauthInboundConfig.redirect_uris ? [...oauthInboundConfig.redirect_uris] : [], // Use from template or empty if not present
      token_endpoint_auth_method: oauthInboundConfig.token_endpoint_auth_method,
      scopes: ['openid', 'profile', 'email'],
    });
  }, [resolvedTechnology, platformForTemplate, onOAuthConfigChange, technologyConfig, setSelectedTemplateConfig]);

  /**
   * Notify parent component about readiness to proceed.
   * Always ready once technology is selected (platform required for "other")
   * If technology section is hidden, auto-select first platform
   */
  useEffect((): void => {
    // If technology section is hidden and no platform selected, auto-select first platform
    if (!stackTypes.technology && !selectedPlatform) {
      setSelectedPlatform(defaultPlatform);
    }

    const isReady: boolean = resolvedTechnology !== TechnologyApplicationTemplate.OTHER || selectedPlatform !== null;

    onReadyChange?.(isReady);
  }, [
    resolvedTechnology,
    selectedPlatform,
    onReadyChange,
    stackTypes.technology,
    setSelectedPlatform,
    defaultPlatform,
  ]);

  const handleTechnologyChange = (newTechnology: TechnologyApplicationTemplate): void => {
    setSelectedTechnology(newTechnology);
    setSelectedPlatform(null);
  };

  const handlePlatformChange = (newPlatform: PlatformApplicationTemplate): void => {
    setSelectedTechnology(null);
    setSelectedPlatform(newPlatform);
  };

  return (
    <Stack direction="column" spacing={3}>
      {stackTypes.technology && (
        <>
          <Stack direction="column" spacing={1}>
            <Typography variant="h1" gutterBottom>
              {t('applications:onboarding.configure.stack.technology.title')}
            </Typography>
            <Typography variant="subtitle1" gutterBottom>
              {t('applications:onboarding.configure.stack.technology.subtitle')}
            </Typography>
          </Stack>

          {/* Technology Grid */}
          <Box
            sx={{
              display: 'grid',
              gridTemplateColumns: {
                xs: '1fr',
                sm: '1fr',
                md: '1fr',
                lg: '1fr',
                xl: 'repeat(3, 1fr)',
              },
              gap: 1.5,
            }}
          >
            {TechnologyBasedApplicationTemplateMetadata.map((option) => {
              const isSelected: boolean = resolvedTechnology === option.value;
              const isDisabled: boolean = option.disabled ?? false;

              return (
                <Card
                  key={option.value}
                  variant="outlined"
                  onClick={isDisabled ? undefined : (): void => handleTechnologyChange(option.value)}
                  sx={{position: 'relative'}}
                >
                  {isDisabled && (
                    <Box
                      sx={{
                        position: 'absolute',
                        top: 8,
                        right: 8,
                        bgcolor: 'warning.main',
                        color: 'warning.contrastText',
                        px: 1,
                        py: 0.5,
                        borderRadius: 1,
                        fontSize: '0.75rem',
                        fontWeight: 600,
                        zIndex: 1,
                      }}
                    >
                      Coming Soon
                    </Box>
                  )}
                  <CardActionArea
                    disabled={isDisabled}
                    sx={{
                      height: '100%',
                      cursor: isDisabled ? 'not-allowed' : 'pointer',
                      border: 1,
                      borderColor: isSelected ? 'primary.main' : 'divider',
                      transition: 'all 0.2s ease-in-out',
                      opacity: isDisabled ? 0.5 : 1,
                      '&:hover': {
                        borderColor: isDisabled ? 'divider' : 'primary.main',
                        bgcolor: (() => {
                          if (isDisabled) return 'transparent';
                          if (isSelected) return 'action.selected';
                          return 'action.hover';
                        })(),
                      },
                    }}
                  >
                    <CardContent sx={{py: 2, px: 2}}>
                      <Stack direction="column" spacing={1.5} alignItems="flex-start">
                        <Box
                          sx={{
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'center',
                            width: 40,
                            height: 40,
                          }}
                        >
                          {option.icon}
                        </Box>
                        <Stack direction="column" spacing={0.5}>
                          <Typography variant="subtitle1" sx={{fontWeight: 500}}>
                            {t(option.titleKey)}
                          </Typography>
                          <Typography variant="body2" color="text.secondary">
                            {t(option.descriptionKey)}
                          </Typography>
                        </Stack>
                      </Stack>
                    </CardContent>
                  </CardActionArea>
                </Card>
              );
            })}
          </Box>
        </>
      )}

      {/* Divider between technology and platform sections */}
      {stackTypes.technology &&
        stackTypes.platform &&
        signInApproach !== ApplicationCreateFlowSignInApproach.EMBEDDED && (
          <Divider sx={{my: 2}}>
            <Typography variant="body2" color="text.secondary">
              {t('applications:onboarding.configure.stack.dividerLabel')}
            </Typography>
          </Divider>
        )}

      {/* Platform Selection */}
      {stackTypes.platform && signInApproach !== ApplicationCreateFlowSignInApproach.EMBEDDED && (
        <>
          <Stack direction="column" spacing={1}>
            <Typography variant="h1">{t('applications:onboarding.configure.stack.platform.title')}</Typography>
            <Typography variant="body2" color="text.secondary">
              {t('applications:onboarding.configure.stack.platform.subtitle')}
            </Typography>
          </Stack>

          <Box
            sx={{
              display: 'grid',
              gridTemplateColumns: 'repeat(auto-fill, minmax(240px, 1fr))',
              gap: 2,
            }}
          >
            {PlatformBasedApplicationTemplateMetadata.map((option) => {
              const isSelected: boolean = selectedPlatform === option.value;

              return (
                <Card key={option.value} variant="outlined" onClick={(): void => handlePlatformChange(option.value)}>
                  <CardActionArea
                    sx={{
                      height: '100%',
                      cursor: 'pointer',
                      border: 1,
                      borderColor: isSelected ? 'primary.main' : 'divider',
                      transition: 'all 0.2s ease-in-out',
                      '&:hover': {
                        borderColor: 'primary.main',
                        bgcolor: isSelected ? 'action.selected' : 'action.hover',
                      },
                    }}
                  >
                    <CardContent>
                      <Stack direction="row" spacing={2} alignItems="center">
                        <Box
                          sx={{
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'center',
                            width: 48,
                            height: 48,
                            borderRadius: 1,
                          }}
                        >
                          {option.icon}
                        </Box>
                        <Box sx={{flex: 1}}>
                          <Typography variant="subtitle1">{t(option.titleKey)}</Typography>
                          <Typography variant="body2" color="text.secondary">
                            {t(option.descriptionKey)}
                          </Typography>
                        </Box>
                      </Stack>
                    </CardContent>
                  </CardActionArea>
                </Card>
              );
            })}
          </Box>
        </>
      )}
    </Stack>
  );
}
