/**
 * Copyright (c) 2025-2026, WSO2 LLC. (https://www.wso2.com).
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

import {useNavigate} from 'react-router';
import {useState, useCallback, useMemo} from 'react';
import {
  Box,
  Stack,
  Button,
  IconButton,
  LinearProgress,
  Breadcrumbs,
  Typography,
  Alert,
  Snackbar,
} from '@wso2/oxygen-ui';
import {X, ChevronRight} from '@wso2/oxygen-ui-icons-react';
import type {JSX} from 'react';
import {useTranslation} from 'react-i18next';
import {useLogger} from '@thunder/logger/react';
import useGetUserSchemas from '../api/useGetUserSchemas';
import useGetUserSchema from '../api/useGetUserSchema';
import useCreateUser from '../api/useCreateUser';
import useUserCreate from '../contexts/UserCreate/useUserCreate';
import {UserCreateFlowStep} from '../models/user-create-flow';
import ConfigureUserType from '../components/create-user/ConfigureUserType';
import ConfigureUserDetails from '../components/create-user/ConfigureUserDetails';

export default function CreateUserPage(): JSX.Element {
  const {t} = useTranslation();
  const navigate = useNavigate();
  const logger = useLogger('CreateUserPage');
  const createUserMutation = useCreateUser();

  const {
    currentStep,
    setCurrentStep,
    selectedSchema,
    setSelectedSchema,
    formValues,
    setFormValues,
    error,
    setError,
  } = useUserCreate();

  const {data: userSchemasData} = useGetUserSchemas();
  const {data: userSchemaDetails, isLoading: isSchemaLoading} = useGetUserSchema(selectedSchema?.id);

  const userSchemas = useMemo(() => userSchemasData?.schemas ?? [], [userSchemasData]);

  const steps: Record<UserCreateFlowStep, {label: string; order: number}> = useMemo(
    () => ({
      USER_TYPE: {label: t('users:createWizard.steps.userType'), order: 1},
      USER_DETAILS: {label: t('users:createWizard.steps.userDetails'), order: 2},
    }),
    [t],
  );

  const [validationError, setValidationError] = useState<string | null>(null);
  const [snackbarOpen, setSnackbarOpen] = useState(false);

  const [stepReady, setStepReady] = useState<Record<UserCreateFlowStep, boolean>>({
    USER_TYPE: false,
    USER_DETAILS: false,
  });

  const handleClose = (): void => {
    if (createUserMutation.isPending) return;
    Promise.resolve(navigate('/users')).catch((_error: unknown) => {
      logger.error('Failed to navigate to users page', {error: _error});
    });
  };

  const handleStepReadyChange = useCallback((step: UserCreateFlowStep, isReady: boolean): void => {
    setStepReady((prev) => ({
      ...prev,
      [step]: isReady,
    }));
  }, []);

  const handleUserTypeStepReadyChange = useCallback(
    (isReady: boolean): void => {
      handleStepReadyChange(UserCreateFlowStep.USER_TYPE, isReady);
    },
    [handleStepReadyChange],
  );

  const handleUserDetailsStepReadyChange = useCallback(
    (isReady: boolean): void => {
      handleStepReadyChange(UserCreateFlowStep.USER_DETAILS, isReady);
    },
    [handleStepReadyChange],
  );

  const handleSchemaChange = useCallback(
    (schema: typeof selectedSchema): void => {
      if (schema?.id !== selectedSchema?.id) {
        setFormValues({});
        setStepReady((prev) => ({...prev, USER_DETAILS: false}));
      }
      setSelectedSchema(schema);
    },
    [selectedSchema, setSelectedSchema, setFormValues],
  );

  const handleSubmit = async (): Promise<void> => {
    setValidationError(null);
    setError(null);

    if (!selectedSchema) {
      setValidationError(t('users:createWizard.validationErrors.userTypeRequired'));
      setSnackbarOpen(true);
      return;
    }

    const trimmedOuId = selectedSchema.ouId?.trim();
    if (!trimmedOuId) {
      setValidationError(t('users:createWizard.validationErrors.ouIdMissing'));
      setSnackbarOpen(true);
      return;
    }

    // Filter out empty/undefined attribute values to avoid sending
    // blank fields that would fail backend schema validation.
    const filteredAttributes = Object.fromEntries(
      Object.entries(formValues).filter(
        ([, v]) => v !== '' && v !== undefined && v !== null,
      ),
    );

    const requestBody = {
      organizationUnit: trimmedOuId,
      type: selectedSchema.name,
      attributes: filteredAttributes,
    };

    try {
      await createUserMutation.mutateAsync(requestBody);
      await navigate('/users');
    } catch (submitError) {
      logger.error('Failed to create user or navigate', {error: submitError});
    }
  };

  const handleNextStep = (): void => {
    switch (currentStep) {
      case UserCreateFlowStep.USER_TYPE:
        setCurrentStep(UserCreateFlowStep.USER_DETAILS);
        break;
      case UserCreateFlowStep.USER_DETAILS:
        handleSubmit().catch(() => {
          // Error handled in handleSubmit
        });
        break;
      default:
        break;
    }
  };

  const handlePrevStep = (): void => {
    switch (currentStep) {
      case UserCreateFlowStep.USER_DETAILS:
        setCurrentStep(UserCreateFlowStep.USER_TYPE);
        break;
      default:
        break;
    }
  };

  const renderStepContent = (): JSX.Element | null => {
    switch (currentStep) {
      case UserCreateFlowStep.USER_TYPE:
        return (
          <ConfigureUserType
            schemas={userSchemas}
            selectedSchema={selectedSchema}
            onSchemaChange={handleSchemaChange}
            onReadyChange={handleUserTypeStepReadyChange}
          />
        );
      case UserCreateFlowStep.USER_DETAILS:
        if (isSchemaLoading) {
          return (
            <Box sx={{textAlign: 'center', py: 4}}>
              <Typography variant="body2" color="text.secondary">
                {t('common:status.loading')}
              </Typography>
            </Box>
          );
        }
        if (!userSchemaDetails) {
          return null;
        }
        return (
          <ConfigureUserDetails
            key={selectedSchema?.id}
            schema={userSchemaDetails}
            defaultValues={formValues}
            onFormValuesChange={setFormValues}
            onReadyChange={handleUserDetailsStepReadyChange}
          />
        );
      default:
        return null;
    }
  };

  const getStepProgress = (): number => {
    const stepNames = Object.keys(steps) as UserCreateFlowStep[];
    return ((stepNames.indexOf(currentStep) + 1) / stepNames.length) * 100;
  };

  const getBreadcrumbSteps = (): UserCreateFlowStep[] => {
    const allSteps: UserCreateFlowStep[] = [
      UserCreateFlowStep.USER_TYPE,
      UserCreateFlowStep.USER_DETAILS,
    ];

    const currentIndex = allSteps.indexOf(currentStep);
    return allSteps.slice(0, currentIndex + 1);
  };

  const handleCloseSnackbar = () => {
    setSnackbarOpen(false);
  };

  const isLastStep = currentStep === UserCreateFlowStep.USER_DETAILS;

  return (
    <Box sx={{minHeight: '100vh', display: 'flex', flexDirection: 'column'}}>
      {/* Progress bar at the very top */}
      <LinearProgress variant="determinate" value={getStepProgress()} sx={{height: 6}} />

      <Box sx={{flex: 1, display: 'flex', flexDirection: 'column'}}>
        {/* Header with close button and breadcrumb */}
        <Box sx={{p: 4, display: 'flex', justifyContent: 'space-between', alignItems: 'center'}}>
          <Stack direction="row" alignItems="center" spacing={2}>
            <IconButton
              aria-label={t('common:actions.close')}
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

                return isLast ? (
                  <Typography key={step} variant="h5" color="text.primary">
                    {steps[step].label}
                  </Typography>
                ) : (
                  <Typography
                    key={step}
                    variant="h5"
                    color="inherit"
                    role="button"
                    tabIndex={0}
                    onClick={() => setCurrentStep(step)}
                    onKeyDown={(e) => {
                      if (e.key === 'Enter' || e.key === ' ') {
                        e.preventDefault();
                        setCurrentStep(step);
                      }
                    }}
                    sx={{cursor: 'pointer', '&:hover': {textDecoration: 'underline'}}}
                  >
                    {steps[step].label}
                  </Typography>
                );
              })}
            </Breadcrumbs>
          </Stack>
        </Box>

        {/* Main content */}
        <Box sx={{flex: 1, display: 'flex', minHeight: 0}}>
          <Box
            sx={{
              flex: 1,
              display: 'flex',
              flexDirection: 'column',
              py: 8,
              px: 20,
              mx: currentStep === UserCreateFlowStep.USER_TYPE ? 'auto' : 0,
              alignItems: 'flex-start',
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
              {/* Error Alerts */}
              {error && (
                <Alert severity="error" sx={{my: 3}} onClose={() => setError(null)}>
                  {error}
                </Alert>
              )}

              {createUserMutation.error && (
                <Alert severity="error" sx={{mb: 3}}>
                  <Typography variant="body2" sx={{fontWeight: 'bold', mb: 0.5}}>
                    {createUserMutation.error.message}
                  </Typography>
                </Alert>
              )}

              {renderStepContent()}

              {/* Navigation buttons */}
              <Stack
                direction="row"
                justifyContent="flex-end"
                alignItems="center"
                spacing={2}
                sx={{mt: 4}}
              >
                {currentStep !== UserCreateFlowStep.USER_TYPE && (
                  <Button variant="text" onClick={handlePrevStep} disabled={createUserMutation.isPending}>
                    {t('common:actions.back')}
                  </Button>
                )}

                <Button
                  variant="contained"
                  disabled={!stepReady[currentStep] || createUserMutation.isPending}
                  sx={{minWidth: 140}}
                  onClick={handleNextStep}
                >
                  {(() => {
                    if (!isLastStep) return t('common:actions.continue');
                    if (createUserMutation.isPending) return t('common:status.saving');
                    return t('users:createUser.title');
                  })()}
                </Button>
              </Stack>
            </Box>
          </Box>
        </Box>
      </Box>

      {/* Validation Error Snackbar */}
      <Snackbar
        open={snackbarOpen}
        autoHideDuration={6000}
        onClose={handleCloseSnackbar}
        anchorOrigin={{vertical: 'top', horizontal: 'right'}}
      >
        <Alert onClose={handleCloseSnackbar} severity="error" sx={{width: '100%'}}>
          {validationError}
        </Alert>
      </Snackbar>
    </Box>
  );
}
