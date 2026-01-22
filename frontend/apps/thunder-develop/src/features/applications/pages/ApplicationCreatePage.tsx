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

import {Box, Stack, Button, IconButton, LinearProgress, Breadcrumbs, Typography, Alert} from '@wso2/oxygen-ui';
import {X, ChevronRight} from '@wso2/oxygen-ui-icons-react';
import type {JSX} from 'react';
import {useNavigate} from 'react-router';
import {useState, useCallback, useMemo, useEffect} from 'react';
import {useTranslation} from 'react-i18next';
import {useLogger} from '@thunder/logger/react';
import {
  useCreateBranding,
  useGetBrandings,
  type CreateBrandingRequest,
  type Branding,
  LayoutType,
} from '@thunder/shared-branding';
import ConfigureSignInOptions from '../components/create-application/configure-signin-options/ConfigureSignInOptions';
import ConfigureDesign from '../components/create-application/ConfigureDesign';
import ConfigureName from '../components/create-application/ConfigureName';
import ConfigureApproach from '../components/create-application/ConfigureApproach';
import ConfigureStack from '../components/create-application/ConfigureStack';
import ConfigureDetails from '../components/create-application/ConfigureDetails';
import {getDefaultOAuthConfig} from '../models/oauth';
import Preview from '../components/create-application/Preview';
import useCreateApplication from '../api/useCreateApplication';
import type {CreateApplicationRequest} from '../models/requests';
import type {OAuth2Config} from '../models/oauth';
import type {Application} from '../models/application';
import useApplicationCreate from '../contexts/ApplicationCreate/useApplicationCreate';
import {
  ApplicationCreateFlowConfiguration,
  ApplicationCreateFlowSignInApproach,
  ApplicationCreateFlowStep,
} from '../models/application-create-flow';
import TemplateConstants from '../constants/template-constants';
import getConfigurationTypeFromTemplate from '../utils/getConfigurationTypeFromTemplate';
import useGetUserTypes from '../../user-types/api/useGetUserTypes';

export default function ApplicationCreatePage(): JSX.Element {
  const {t} = useTranslation();

  const {
    currentStep,
    setCurrentStep,
    appName,
    setAppName,
    selectedColor,
    setSelectedColor,
    appLogo,
    setAppLogo,
    integrations,
    toggleIntegration,
    selectedAuthFlow,
    signInApproach,
    setSignInApproach,
    selectedTechnology,
    selectedPlatform,
    setHostingUrl,
    callbackUrlFromConfig,
    setCallbackUrlFromConfig,
    selectedTemplateConfig,
    error,
    setError,
  } = useApplicationCreate();

  const steps: Record<ApplicationCreateFlowStep, {label: string; order: number}> = useMemo(
    () => ({
      NAME: {label: t('applications:onboarding.steps.name'), order: 1},
      DESIGN: {label: t('applications:onboarding.steps.design'), order: 2},
      OPTIONS: {label: t('applications:onboarding.steps.options'), order: 3},
      APPROACH: {label: t('applications:onboarding.steps.experience'), order: 4},
      STACK: {label: t('applications:onboarding.steps.stack'), order: 5},
      CONFIGURE: {label: t('applications:onboarding.steps.configure'), order: 6},
    }),
    [t],
  );
  const navigate = useNavigate();
  const logger = useLogger('ApplicationCreatePage');
  const createApplication = useCreateApplication();
  const createBranding = useCreateBranding();
  const {data: userTypesData} = useGetUserTypes();
  const {data: brandingsData} = useGetBrandings({limit: 100});

  const [selectedUserTypes, setSelectedUserTypes] = useState<string[]>([]);

  const [stepReady, setStepReady] = useState<Record<ApplicationCreateFlowStep, boolean>>({
    NAME: false,
    DESIGN: true,
    OPTIONS: true,
    APPROACH: true,
    STACK: true,
    CONFIGURE: true,
  });
  const [useDefaultBranding, setUseDefaultBranding] = useState<boolean>(false);
  const [defaultBrandingId, setDefaultBrandingId] = useState<string | undefined>(undefined);
  const [oauthConfig, setOAuthConfig] = useState<OAuth2Config | null>(getDefaultOAuthConfig());

  /**
   * Update OAuth config with callback URL from configure step.
   */
  useEffect(() => {
    if (callbackUrlFromConfig) {
      setOAuthConfig((prevConfig) =>
        prevConfig
          ? {
              ...prevConfig,
              redirect_uris: [callbackUrlFromConfig],
            }
          : null,
      );
    }
  }, [callbackUrlFromConfig]);

  const handleClose = (): void => {
    (async () => {
      await navigate('/applications');
    })().catch((_error: unknown) => {
      logger.error('Failed to navigate to applications page', {error: _error});
    });
  };

  const handleLogoSelect = (logoUrl: string): void => {
    setAppLogo(logoUrl);
  };

  const handleIntegrationToggle = (integrationId: string): void => {
    toggleIntegration(integrationId);
  };

  const handleBrandingSelectionChange = (useDefault: boolean, brandingId?: string): void => {
    setUseDefaultBranding(useDefault);
    setDefaultBrandingId(brandingId);
  };

  const handleCreateApplication = (skipOAuthConfig = false): void => {
    setError(null);

    const authFlowId: string | undefined = selectedAuthFlow?.id;

    // Validate that we have a valid flow selected
    if (!authFlowId) {
      setError(t('onboarding.configure.SignInOptions.noFlowFound'));

      return;
    }

    const createApplicationWithBranding = (brandingId: string): void => {
      const userTypes = userTypesData?.schemas ?? [];
      const allowedUserTypes = (() => {
        // If there's exactly 1 user type, automatically include it
        if (userTypes.length === 1) {
          return [userTypes[0].name];
        }

        // If there are multiple user types, use the selected ones
        if (userTypes.length > 1) {
          return selectedUserTypes.length > 0 ? selectedUserTypes : undefined;
        }

        // If there are no user types, don't include the field
        return undefined;
      })();

      const applicationData: CreateApplicationRequest = {
        name: appName,
        logo_url: appLogo ?? undefined,
        auth_flow_id: authFlowId,
        user_attributes: ['given_name', 'family_name', 'email', 'groups'],
        branding_id: brandingId,
        is_registration_flow_enabled: true,
        ...(allowedUserTypes && {allowed_user_types: allowedUserTypes}),
        // Include template if available, append '-embedded' suffix for CUSTOM approach
        ...(selectedTemplateConfig?.id && {
          template:
            signInApproach === ApplicationCreateFlowSignInApproach.EMBEDDED
              ? `${selectedTemplateConfig.id}${TemplateConstants.EMBEDDED_SUFFIX}`
              : selectedTemplateConfig.id,
        }),
        // Only include OAuth config if not skipping
        ...(!skipOAuthConfig && {
          inbound_auth_config: [
            {
              type: 'oauth2',
              config: oauthConfig,
            },
          ],
        }),
      };

      createApplication.mutate(applicationData, {
        onSuccess: (createdApp: Application): void => {
          // Navigate to the application edit page
          (async () => {
            await navigate(`/applications/${createdApp.id}`);
          })().catch((_error: unknown) => {
            logger.error('Failed to navigate to application details', {error: _error, applicationId: createdApp.id});
          });
        },
        onError: (err: Error) => {
          setError(err.message ?? 'Failed to create application. Please try again.');
        },
      });
    };

    // If using DEFAULT branding, use its ID directly
    if (useDefaultBranding && defaultBrandingId) {
      createApplicationWithBranding(defaultBrandingId);
      return;
    }

    // Otherwise, create a new branding with the selected color
    // If there are no brandings in the system, create "Default Theme"
    // If there's at least one branding, create app-specific theme
    const hasNoBrandings = (brandingsData?.brandings?.length ?? 0) === 0;
    const brandingName = hasNoBrandings
      ? t('appearance:theme.defaultTheme')
      : t('appearance:theme.appTheme.displayName', {appName});
    const brandingData: CreateBrandingRequest = {
      displayName: brandingName,
      preferences: {
        layout: {
          type: LayoutType.CENTERED,
        },
        theme: {
          activeColorScheme: 'light',
          colorSchemes: {
            light: {
              colors: {
                primary: {
                  main: selectedColor,
                  dark: selectedColor,
                  contrastText: '#ffffff',
                },
                secondary: {
                  main: selectedColor,
                  dark: selectedColor,
                  contrastText: '#ffffff',
                },
              },
              ...(appLogo && {
                images: {
                  logo: {
                    primary: {
                      url: appLogo,
                      alt: `${appName} Logo`,
                      width: 128,
                      height: 64,
                    },
                    favicon: {
                      url: appLogo,
                      type: 'image/png',
                    },
                  },
                },
              }),
            },
            dark: {
              colors: {
                primary: {
                  main: selectedColor,
                  dark: selectedColor,
                  contrastText: '#ffffff',
                },
                secondary: {
                  main: selectedColor,
                  dark: selectedColor,
                  contrastText: '#ffffff',
                },
              },
              ...(appLogo && {
                images: {
                  logo: {
                    primary: {
                      url: appLogo,
                      alt: `${appName} Logo`,
                      width: 128,
                      height: 64,
                    },
                    favicon: {
                      url: appLogo,
                      type: 'image/png',
                    },
                  },
                },
              }),
            },
          },
        },
      },
    };

    createBranding.mutate(brandingData, {
      onSuccess: (branding: Branding) => {
        createApplicationWithBranding(branding.id);
      },
      onError: (err: Error) => {
        setError(err.message ?? 'Failed to create branding. Please try again.');
      },
    });
  };

  const handleNextStep = (): void => {
    switch (currentStep) {
      case ApplicationCreateFlowStep.NAME:
        setCurrentStep(ApplicationCreateFlowStep.DESIGN);
        break;
      case ApplicationCreateFlowStep.DESIGN:
        setCurrentStep(ApplicationCreateFlowStep.OPTIONS);
        break;
      case ApplicationCreateFlowStep.OPTIONS:
        setCurrentStep(ApplicationCreateFlowStep.APPROACH);
        break;
      case ApplicationCreateFlowStep.APPROACH:
        // Always go to technology selection to set selectedTemplateConfig
        setCurrentStep(ApplicationCreateFlowStep.STACK);
        break;
      case ApplicationCreateFlowStep.STACK: {
        // For CUSTOM approach, create app immediately after technology selection
        if (signInApproach === ApplicationCreateFlowSignInApproach.EMBEDDED) {
          handleCreateApplication(true); // Skip OAuth for custom
          break;
        }

        // For INBUILT approach, check if configuration is needed based on template
        const needsConfiguration: boolean =
          getConfigurationTypeFromTemplate(selectedTemplateConfig) !== ApplicationCreateFlowConfiguration.NONE;

        if (needsConfiguration) {
          setCurrentStep(ApplicationCreateFlowStep.CONFIGURE);
        } else {
          // Skip configure step for technologies/platforms that don't need it
          handleCreateApplication(false);
        }
        break;
      }
      case ApplicationCreateFlowStep.CONFIGURE:
        // Configuration complete, create application with OAuth config
        handleCreateApplication(false);
        break;
      default:
        break;
    }
  };

  const handlePrevStep = (): void => {
    switch (currentStep) {
      case ApplicationCreateFlowStep.DESIGN:
        setCurrentStep(ApplicationCreateFlowStep.NAME);
        break;
      case ApplicationCreateFlowStep.OPTIONS:
        setCurrentStep(ApplicationCreateFlowStep.DESIGN);
        break;
      case ApplicationCreateFlowStep.APPROACH:
        setCurrentStep(ApplicationCreateFlowStep.OPTIONS);
        break;
      case ApplicationCreateFlowStep.STACK:
        setCurrentStep(ApplicationCreateFlowStep.APPROACH);
        break;
      case ApplicationCreateFlowStep.CONFIGURE:
        setCurrentStep(ApplicationCreateFlowStep.STACK);
        break;
      default:
        break;
    }
  };

  const handleStepReadyChange = useCallback((step: ApplicationCreateFlowStep, isReady: boolean): void => {
    setStepReady((prev) => ({
      ...prev,
      [step]: isReady,
    }));
  }, []);

  const handleNameStepReadyChange = useCallback(
    (isReady: boolean): void => {
      handleStepReadyChange(ApplicationCreateFlowStep.NAME, isReady);
    },
    [handleStepReadyChange],
  );

  const handleDesignStepReadyChange = useCallback(
    (isReady: boolean): void => {
      handleStepReadyChange(ApplicationCreateFlowStep.DESIGN, isReady);
    },
    [handleStepReadyChange],
  );

  const handleOptionsStepReadyChange = useCallback(
    (isReady: boolean): void => {
      handleStepReadyChange(ApplicationCreateFlowStep.OPTIONS, isReady);
    },
    [handleStepReadyChange],
  );

  const handleApproachStepReadyChange = useCallback(
    (isReady: boolean): void => {
      handleStepReadyChange(ApplicationCreateFlowStep.APPROACH, isReady);
    },
    [handleStepReadyChange],
  );

  const handleTechnologyStepReadyChange = useCallback(
    (isReady: boolean): void => {
      handleStepReadyChange(ApplicationCreateFlowStep.STACK, isReady);
    },
    [handleStepReadyChange],
  );

  const handleConfigureStepReadyChange = useCallback(
    (isReady: boolean): void => {
      handleStepReadyChange(ApplicationCreateFlowStep.CONFIGURE, isReady);
    },
    [handleStepReadyChange],
  );

  const renderStepContent = (): JSX.Element | null => {
    switch (currentStep) {
      case ApplicationCreateFlowStep.NAME:
        return (
          <ConfigureName appName={appName} onAppNameChange={setAppName} onReadyChange={handleNameStepReadyChange} />
        );

      case ApplicationCreateFlowStep.DESIGN:
        return (
          <ConfigureDesign
            appLogo={appLogo}
            selectedColor={selectedColor}
            appName={appName}
            onLogoSelect={handleLogoSelect}
            onInitialLogoLoad={handleLogoSelect}
            onColorSelect={setSelectedColor}
            onReadyChange={handleDesignStepReadyChange}
            onBrandingSelectionChange={handleBrandingSelectionChange}
          />
        );

      case ApplicationCreateFlowStep.OPTIONS:
        return (
          <ConfigureSignInOptions
            integrations={integrations}
            onIntegrationToggle={handleIntegrationToggle}
            onReadyChange={handleOptionsStepReadyChange}
          />
        );

      case ApplicationCreateFlowStep.APPROACH:
        return (
          <ConfigureApproach
            selectedApproach={signInApproach}
            onApproachChange={setSignInApproach}
            onReadyChange={handleApproachStepReadyChange}
            userTypes={userTypesData?.schemas ?? []}
            selectedUserTypes={selectedUserTypes}
            onUserTypesChange={setSelectedUserTypes}
          />
        );

      case ApplicationCreateFlowStep.STACK:
        return (
          <ConfigureStack
            oauthConfig={oauthConfig}
            onOAuthConfigChange={setOAuthConfig}
            onReadyChange={handleTechnologyStepReadyChange}
            stackTypes={{technology: true, platform: true}}
          />
        );

      case ApplicationCreateFlowStep.CONFIGURE:
        return (
          <ConfigureDetails
            technology={selectedTechnology}
            platform={selectedPlatform}
            onHostingUrlChange={setHostingUrl}
            onCallbackUrlChange={setCallbackUrlFromConfig}
            onReadyChange={handleConfigureStepReadyChange}
          />
        );

      default:
        return null;
    }
  };

  const getStepProgress = (): number => {
    const stepNames = Object.keys(steps) as ApplicationCreateFlowStep[];
    return ((stepNames.indexOf(currentStep) + 1) / stepNames.length) * 100;
  };

  const getBreadcrumbSteps = (): ApplicationCreateFlowStep[] => {
    const allSteps: ApplicationCreateFlowStep[] = [
      ApplicationCreateFlowStep.NAME,
      ApplicationCreateFlowStep.DESIGN,
      ApplicationCreateFlowStep.OPTIONS,
      ApplicationCreateFlowStep.APPROACH,
    ];

    // Only show technology and configure steps for inbuilt approach
    if (signInApproach === ApplicationCreateFlowSignInApproach.INBUILT) {
      allSteps.push(ApplicationCreateFlowStep.STACK);

      // Show configure step if template requires configuration (has empty redirect_uris)
      const needsConfiguration: boolean =
        getConfigurationTypeFromTemplate(selectedTemplateConfig) !== ApplicationCreateFlowConfiguration.NONE;

      if (needsConfiguration) {
        allSteps.push(ApplicationCreateFlowStep.CONFIGURE);
      }
    }

    const currentIndex = allSteps.indexOf(currentStep);
    return allSteps.slice(0, currentIndex + 1);
  };

  return (
    <Box sx={{minHeight: '100vh', display: 'flex', flexDirection: 'column'}}>
      {/* Progress bar at the very top */}
      <LinearProgress variant="determinate" value={getStepProgress()} sx={{height: 6}} />

      <Box sx={{flex: 1, display: 'flex', flexDirection: 'row'}}>
        <Box
          sx={{
            flex: currentStep === ApplicationCreateFlowStep.NAME ? 1 : '0 0 50%',
            display: 'flex',
            flexDirection: 'column',
          }}
        >
          {/* Header with close button and breadcrumb */}
          <Box sx={{p: 4, display: 'flex', justifyContent: 'space-between', alignItems: 'center'}}>
            <Stack direction="row" alignItems="center" spacing={2}>
              <IconButton
                onClick={handleClose}
                sx={{
                  bgcolor: 'background.paper',
                  '&:hover': {bgcolor: 'action.hover'},
                  boxShadow: 1,
                }}
              >
                <X size={24} />
              </IconButton>
              <Breadcrumbs separator={<ChevronRight size={16} />} aria-label="breadcrumb">
                {getBreadcrumbSteps().map((step, index, array) => {
                  const isLast = index === array.length - 1;
                  const isClickable = !isLast;

                  return isClickable ? (
                    <Typography key={step} variant="h5" onClick={() => setCurrentStep(step)} sx={{cursor: 'pointer'}}>
                      {steps[step].label}
                    </Typography>
                  ) : (
                    <Typography key={step} variant="h5" color="text.primary">
                      {steps[step].label}
                    </Typography>
                  );
                })}
              </Breadcrumbs>
            </Stack>
          </Box>

          {/* Main content */}
          <Box sx={{flex: 1, display: 'flex', minHeight: 0}}>
            {/* Left side - Form content */}
            <Box
              sx={{
                flex: 1,
                display: 'flex',
                flexDirection: 'column',
                py: 8,
                px: 20,
                mx: currentStep === ApplicationCreateFlowStep.NAME ? 'auto' : 0,
              }}
            >
              <Box
                sx={{
                  width: '100%',
                  maxWidth: 800,
                  display: 'flex',
                  flexDirection: 'column',
                }}
              >
                {/* Error Alert */}
                {error && (
                  <Alert severity="error" sx={{my: 3}} onClose={() => setError(null)}>
                    {error}
                  </Alert>
                )}

                {renderStepContent()}

                {/* Navigation buttons */}
                <Box
                  sx={{
                    mt: 4,
                    display: 'flex',
                    justifyContent: currentStep === ApplicationCreateFlowStep.NAME ? 'flex-start' : 'space-between',
                    gap: 2,
                  }}
                >
                  {currentStep !== ApplicationCreateFlowStep.NAME && (
                    <Button
                      variant="outlined"
                      onClick={handlePrevStep}
                      sx={{minWidth: 100}}
                      disabled={createApplication.isPending || createBranding.isPending}
                    >
                      {t('common:actions.back')}
                    </Button>
                  )}

                  <Button
                    variant="contained"
                    disabled={!stepReady[currentStep]}
                    sx={{minWidth: 100}}
                    onClick={handleNextStep}
                  >
                    {t('common:actions.continue')}
                  </Button>
                </Box>
              </Box>
            </Box>
          </Box>
        </Box>
        {/* Right side - Preview (show from design step onwards) */}
        {currentStep !== ApplicationCreateFlowStep.NAME && (
          <Box sx={{flex: '0 0 50%', display: 'flex', flexDirection: 'column', p: 5}}>
            <Preview appLogo={appLogo} selectedColor={selectedColor} integrations={integrations} />
          </Box>
        )}
      </Box>
    </Box>
  );
}
