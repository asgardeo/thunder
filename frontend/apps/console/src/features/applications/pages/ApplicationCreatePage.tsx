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

import {useHasMultipleOUs} from '@thunderid/configure-organization-units';
import {useLogger} from '@thunderid/logger/react';
import {
  Box,
  Stack,
  Button,
  IconButton,
  LinearProgress,
  Breadcrumbs,
  Typography,
  Alert,
  CircularProgress,
} from '@wso2/oxygen-ui';
import {X, ChevronRight} from '@wso2/oxygen-ui-icons-react';
import type {JSX} from 'react';
import {useState, useCallback, useMemo} from 'react';
import {useTranslation} from 'react-i18next';
import {useNavigate} from 'react-router';
import useCreateFlow from '../../flows/api/useCreateFlow';
import type {BasicFlowDefinition} from '../../flows/models/responses';
import generateFlowGraph from '../../flows/utils/generateFlowGraph';
import useIdentityProviders from '../../integrations/api/useIdentityProviders';
import {AuthenticatorTypes} from '../../integrations/models/authenticators';
import {IdentityProviderTypes} from '../../integrations/models/identity-provider';
import useGetUserTypes from '../../user-types/api/useGetUserTypes';
import useCreateApplication from '../api/useCreateApplication';
import ConfigureSignInOptions from '../components/create-application/configure-signin-options/ConfigureSignInOptions';
import ConfigureDesign from '../components/create-application/ConfigureDesign';
import ConfigureDetails from '../components/create-application/ConfigureDetails';
import ConfigureExperience from '../components/create-application/ConfigureExperience';
import ConfigureName from '../components/create-application/ConfigureName';
import ConfigureOrganizationUnit from '../components/create-application/ConfigureOrganizationUnit';
import ConfigureStack from '../components/create-application/ConfigureStack';
import ShowClientSecret from '../components/create-application/ShowClientSecret';
import TemplateConstants from '../constants/template-constants';
import useApplicationCreate from '../contexts/ApplicationCreate/useApplicationCreate';
import type {Application} from '../models/application';
import {
  ApplicationCreateFlowConfiguration,
  ApplicationCreateFlowSignInApproach,
  ApplicationCreateFlowStep,
} from '../models/application-create-flow';
import type {OAuth2Config} from '../models/oauth';
import type {CreateApplicationRequest} from '../models/requests';
import getConfigurationTypeFromTemplate from '../utils/getConfigurationTypeFromTemplate';
import GatePreview from '@/components/GatePreview/GatePreview';
import buildPreviewMock from '@/components/GatePreview/mocks/buildPreviewMock';

export default function ApplicationCreatePage(): JSX.Element {
  const {t} = useTranslation();

  const {
    currentStep,
    setCurrentStep,
    appName,
    setAppName,
    ouId,
    setOuId,
    themeId,
    setThemeId,
    selectedTheme,
    setSelectedTheme,
    appLogo,
    setAppLogo,
    integrations,
    toggleIntegration,
    selectedAuthFlow,
    setSelectedAuthFlow,
    signInApproach,
    setSignInApproach,
    selectedTechnology,
    selectedPlatform,
    setHostingUrl,
    callbackUrlFromConfig,
    setCallbackUrlFromConfig,
    relyingPartyId,
    relyingPartyName,
    selectedTemplateConfig,
    error,
    setError,
  } = useApplicationCreate();

  const steps: Record<ApplicationCreateFlowStep, {label: string; order: number}> = useMemo(
    () => ({
      NAME: {label: t('applications:onboarding.steps.name'), order: 1},
      ORGANIZATION_UNIT: {label: t('applications:onboarding.steps.organizationUnit'), order: 2},
      DESIGN: {label: t('applications:onboarding.steps.design'), order: 3},
      OPTIONS: {label: t('applications:onboarding.steps.options'), order: 4},
      EXPERIENCE: {label: t('applications:onboarding.steps.experience'), order: 5},
      STACK: {label: t('applications:onboarding.steps.stack'), order: 6},
      CONFIGURE: {label: t('applications:onboarding.steps.configure'), order: 7},
      COMPLETE: {label: t('applications:onboarding.steps.complete'), order: 8},
    }),
    [t],
  );
  const navigate = useNavigate();
  const logger = useLogger('ApplicationCreatePage');
  const createApplication = useCreateApplication();
  const {data: userTypesData} = useGetUserTypes();
  const {hasMultipleOUs, isLoading: isOuLoading, ouList} = useHasMultipleOUs();

  const [selectedUserTypes, setSelectedUserTypes] = useState<string[]>([]);
  const [createdApplication, setCreatedApplication] = useState<Application | null>(null);

  const createFlow = useCreateFlow();
  const {data: idpData} = useIdentityProviders();

  const [stepReady, setStepReady] = useState<Record<ApplicationCreateFlowStep, boolean>>({
    NAME: false,
    ORGANIZATION_UNIT: false,
    DESIGN: true,
    OPTIONS: true,
    EXPERIENCE: true,
    STACK: true,
    CONFIGURE: true,
    COMPLETE: true,
  });

  const [oauthConfig, setOAuthConfig] = useState<OAuth2Config | null>(null);

  const effectiveOauthConfig = useMemo(
    () =>
      oauthConfig && callbackUrlFromConfig ? {...oauthConfig, redirectUris: [callbackUrlFromConfig]} : oauthConfig,
    [oauthConfig, callbackUrlFromConfig],
  );

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
    setSelectedAuthFlow(null);
  };

  const handleCreateApplication = (skipOAuthConfig = false, overrideFlowId?: string): void => {
    setError(null);

    const authFlowId: string | undefined = overrideFlowId ?? selectedAuthFlow?.id;

    // Validate that we have a valid flow selected
    if (!authFlowId) {
      setError(t('onboarding.configure.SignInOptions.noFlowFound'));

      return;
    }

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

    const effectiveOuId = hasMultipleOUs ? ouId : ouList[0]?.id;

    const applicationData: CreateApplicationRequest = {
      name: appName,
      logoUrl: appLogo ?? undefined,
      authFlowId,
      userAttributes: ['given_name', 'family_name', 'email', 'groups'],
      ...(themeId && {themeId}),
      isRegistrationFlowEnabled: true,
      ...(allowedUserTypes && {allowedUserTypes}),
      ...(effectiveOuId && {ouId: effectiveOuId}),
      // Include template if available, append '-embedded' suffix for CUSTOM approach
      ...(selectedTemplateConfig?.id && {
        template:
          signInApproach === ApplicationCreateFlowSignInApproach.EMBEDDED
            ? `${selectedTemplateConfig.id}${TemplateConstants.EMBEDDED_SUFFIX}`
            : selectedTemplateConfig.id,
      }),
      // Only include OAuth config if not skipping
      ...(!skipOAuthConfig && {
        inboundAuthConfig: [
          {
            type: 'oauth2',
            config: effectiveOauthConfig,
          },
        ],
      }),
    };

    createApplication.mutate(applicationData, {
      onSuccess: (createdApp: Application): void => {
        const hasClientSecret = createdApp.inboundAuthConfig?.some(
          (config) => config.type === 'oauth2' && config.config?.clientSecret,
        );

        if (hasClientSecret) {
          // Store the application and show the COMPLETE step
          setCreatedApplication(createdApp);
          setCurrentStep(ApplicationCreateFlowStep.COMPLETE);
        } else {
          // No client secret, navigate directly to the application details page
          (async () => {
            await navigate(`/applications/${createdApp.id}`);
          })().catch((_error: unknown) => {
            logger.error('Failed to navigate to application details', {error: _error, applicationId: createdApp.id});
          });
        }
      },
      onError: (err: Error) => {
        setError(err.message ?? 'Failed to create application. Please try again.');
      },
    });
  };

  const ensureFlowAndCreateApplication = (skipOAuthConfig = false): void => {
    // If we already have a selected flow, proceed to create application
    if (selectedAuthFlow) {
      handleCreateApplication(skipOAuthConfig);
      return;
    }

    // Check if we need to generate a flow
    const hasEnabledIntegrations = Object.values(integrations).some((v) => v);

    if (hasEnabledIntegrations) {
      const availableIntegrations = idpData ?? [];
      const googleProvider = availableIntegrations.find((idp) => idp.type === IdentityProviderTypes.GOOGLE);
      const githubProvider = availableIntegrations.find((idp) => idp.type === IdentityProviderTypes.GITHUB);

      const generatedFlowRequest = generateFlowGraph({
        hasBasicAuth: integrations[AuthenticatorTypes.BASIC_AUTH] ?? false,
        hasPasskey: integrations[AuthenticatorTypes.PASSKEY] ?? false,
        googleIdpId: integrations[googleProvider?.id ?? ''] ? googleProvider?.id : undefined,
        githubIdpId: integrations[githubProvider?.id ?? ''] ? githubProvider?.id : undefined,
        hasSmsOtp: integrations['sms-otp'] ?? false,
        relyingPartyId: relyingPartyId || window.location.hostname,
        relyingPartyName: relyingPartyName || appName,
      });

      createFlow.mutate(generatedFlowRequest, {
        onSuccess: (savedFlow) => {
          // We cast because BasicFlowDefinition is a subset of FlowDefinitionResponse
          setSelectedAuthFlow(savedFlow as unknown as BasicFlowDefinition);

          // Proceed to create application with the newly generated flow
          handleCreateApplication(skipOAuthConfig, savedFlow.id);
        },
        onError: (err) => {
          setError(err.message ?? 'Failed to generate authentication flow.');
        },
      });
    } else {
      // If no integrations selected, try to create application (will fail validation if flow required)
      handleCreateApplication(skipOAuthConfig);
    }
  };

  const handleNextStep = (): void => {
    switch (currentStep) {
      case ApplicationCreateFlowStep.NAME:
        if (isOuLoading) return;
        if (hasMultipleOUs) {
          setCurrentStep(ApplicationCreateFlowStep.ORGANIZATION_UNIT);
        } else {
          setCurrentStep(ApplicationCreateFlowStep.DESIGN);
        }
        break;
      case ApplicationCreateFlowStep.ORGANIZATION_UNIT:
        setCurrentStep(ApplicationCreateFlowStep.DESIGN);
        break;
      case ApplicationCreateFlowStep.DESIGN:
        setCurrentStep(ApplicationCreateFlowStep.OPTIONS);
        break;
      case ApplicationCreateFlowStep.OPTIONS:
        setCurrentStep(ApplicationCreateFlowStep.EXPERIENCE);
        break;
      case ApplicationCreateFlowStep.EXPERIENCE:
        // Always go to technology selection to set selectedTemplateConfig
        setCurrentStep(ApplicationCreateFlowStep.STACK);
        break;
      case ApplicationCreateFlowStep.STACK: {
        // For INBUILT approach and EMBEDDED approach, check if passkey configuration is needed
        const isPasskeyConfigEnabled: boolean =
          !selectedAuthFlow && (integrations[AuthenticatorTypes.PASSKEY] ?? false);

        // For CUSTOM approach, create app immediately after technology selection, unless passkey config is needed
        if (signInApproach === ApplicationCreateFlowSignInApproach.EMBEDDED) {
          if (isPasskeyConfigEnabled) {
            setCurrentStep(ApplicationCreateFlowStep.CONFIGURE);
          } else {
            ensureFlowAndCreateApplication(true); // Skip OAuth for custom
          }
          break;
        }

        const configurationType: ApplicationCreateFlowConfiguration =
          getConfigurationTypeFromTemplate(selectedTemplateConfig);

        const needsConfiguration: boolean =
          configurationType !== ApplicationCreateFlowConfiguration.NONE || isPasskeyConfigEnabled;

        if (needsConfiguration) {
          setCurrentStep(ApplicationCreateFlowStep.CONFIGURE);
        } else {
          // Skip configure step for technologies/platforms that don't need it
          ensureFlowAndCreateApplication(false);
        }
        break;
      }
      case ApplicationCreateFlowStep.CONFIGURE:
        // Configuration complete, create application
        if (signInApproach === ApplicationCreateFlowSignInApproach.EMBEDDED) {
          ensureFlowAndCreateApplication(true);
        } else {
          ensureFlowAndCreateApplication(false);
        }
        break;
      case ApplicationCreateFlowStep.COMPLETE:
        // Navigate to the application details page
        if (createdApplication) {
          (async () => {
            await navigate(`/applications/${createdApplication.id}`);
          })().catch((_error: unknown) => {
            logger.error('Failed to navigate to application details', {
              error: _error,
              applicationId: createdApplication.id,
            });
          });
        }
        break;
      default:
        break;
    }
  };

  const handlePrevStep = (): void => {
    switch (currentStep) {
      case ApplicationCreateFlowStep.ORGANIZATION_UNIT:
        setCurrentStep(ApplicationCreateFlowStep.NAME);
        break;
      case ApplicationCreateFlowStep.DESIGN:
        if (hasMultipleOUs) {
          setCurrentStep(ApplicationCreateFlowStep.ORGANIZATION_UNIT);
        } else {
          setCurrentStep(ApplicationCreateFlowStep.NAME);
        }
        break;
      case ApplicationCreateFlowStep.OPTIONS:
        setCurrentStep(ApplicationCreateFlowStep.DESIGN);
        break;
      case ApplicationCreateFlowStep.EXPERIENCE:
        setCurrentStep(ApplicationCreateFlowStep.OPTIONS);
        break;
      case ApplicationCreateFlowStep.STACK:
        setCurrentStep(ApplicationCreateFlowStep.EXPERIENCE);
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

  const handleOuStepReadyChange = useCallback(
    (isReady: boolean): void => {
      handleStepReadyChange(ApplicationCreateFlowStep.ORGANIZATION_UNIT, isReady);
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
      handleStepReadyChange(ApplicationCreateFlowStep.EXPERIENCE, isReady);
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

      case ApplicationCreateFlowStep.ORGANIZATION_UNIT:
        return (
          <ConfigureOrganizationUnit
            selectedOuId={ouId}
            onOuIdChange={setOuId}
            onReadyChange={handleOuStepReadyChange}
          />
        );

      case ApplicationCreateFlowStep.DESIGN:
        return (
          <ConfigureDesign
            appLogo={appLogo}
            themeId={themeId}
            selectedTheme={selectedTheme}
            onLogoSelect={handleLogoSelect}
            onThemeSelect={(id, config) => {
              setThemeId(id);
              setSelectedTheme(config);
            }}
            onReadyChange={handleDesignStepReadyChange}
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

      case ApplicationCreateFlowStep.EXPERIENCE:
        return (
          <ConfigureExperience
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

      case ApplicationCreateFlowStep.COMPLETE: {
        if (!createdApplication) {
          return null;
        }

        const oauth2Config = createdApplication.inboundAuthConfig?.find((config) => config.type === 'oauth2');
        const clientSecret = oauth2Config?.config?.clientSecret;

        if (!clientSecret) {
          return null;
        }

        return <ShowClientSecret appName={appName} clientSecret={clientSecret} onContinue={handleNextStep} />;
      }

      default:
        return null;
    }
  };

  const getStepProgress = (): number => {
    const stepNames = Object.keys(steps) as ApplicationCreateFlowStep[];
    return ((stepNames.indexOf(currentStep) + 1) / stepNames.length) * 100;
  };

  const getBreadcrumbSteps = (): ApplicationCreateFlowStep[] => {
    const allSteps: ApplicationCreateFlowStep[] = [ApplicationCreateFlowStep.NAME];

    if (hasMultipleOUs) {
      allSteps.push(ApplicationCreateFlowStep.ORGANIZATION_UNIT);
    }

    allSteps.push(
      ApplicationCreateFlowStep.DESIGN,
      ApplicationCreateFlowStep.OPTIONS,
      ApplicationCreateFlowStep.EXPERIENCE,
    );

    // Only show technology and configure steps for inbuilt approach
    if (signInApproach === ApplicationCreateFlowSignInApproach.INBUILT) {
      allSteps.push(ApplicationCreateFlowStep.STACK);

      // Show configure step if template requires configuration (has empty redirectUris)
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
            flex:
              currentStep === ApplicationCreateFlowStep.NAME ||
              currentStep === ApplicationCreateFlowStep.ORGANIZATION_UNIT ||
              currentStep === ApplicationCreateFlowStep.COMPLETE
                ? 1
                : '0 0 50%',
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
                mx:
                  currentStep === ApplicationCreateFlowStep.NAME ||
                  currentStep === ApplicationCreateFlowStep.ORGANIZATION_UNIT
                    ? 'auto'
                    : 0,
                alignItems: currentStep === ApplicationCreateFlowStep.COMPLETE ? 'center' : 'flex-start',
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
                    justifyContent: currentStep === ApplicationCreateFlowStep.NAME ? 'flex-end' : 'space-between',
                    gap: 2,
                  }}
                >
                  {currentStep !== ApplicationCreateFlowStep.NAME &&
                    currentStep !== ApplicationCreateFlowStep.COMPLETE && (
                      <Button
                        variant="outlined"
                        onClick={handlePrevStep}
                        sx={{minWidth: 100}}
                        disabled={createApplication.isPending}
                      >
                        {t('common:actions.back')}
                      </Button>
                    )}

                  {currentStep !== ApplicationCreateFlowStep.COMPLETE && (
                    <Box sx={{display: 'flex', alignItems: 'center', gap: 2}}>
                      {createFlow.isPending && <CircularProgress size={20} />}
                      <Button
                        data-testid="application-wizard-next-button"
                        variant="contained"
                        disabled={
                          !stepReady[currentStep] ||
                          createFlow.isPending ||
                          (currentStep === ApplicationCreateFlowStep.NAME && isOuLoading)
                        }
                        sx={{minWidth: 100}}
                        onClick={handleNextStep}
                      >
                        {t('common:actions.continue')}
                      </Button>
                    </Box>
                  )}
                </Box>
              </Box>
            </Box>
          </Box>
        </Box>
        {/* Right side - Preview (show from design step onwards, but hide on complete step) */}
        {currentStep !== ApplicationCreateFlowStep.NAME &&
          currentStep !== ApplicationCreateFlowStep.ORGANIZATION_UNIT &&
          currentStep !== ApplicationCreateFlowStep.COMPLETE && (
            <Box sx={{flex: '0 0 50%', display: 'flex', flexDirection: 'column', p: 5}}>
              <GatePreview
                theme={selectedTheme}
                mock={buildPreviewMock(integrations, idpData ?? [], {
                  application: {
                    logoUrl: appLogo!,
                  },
                })}
                displayName={appName ?? undefined}
              />
            </Box>
          )}
      </Box>
    </Box>
  );
}
