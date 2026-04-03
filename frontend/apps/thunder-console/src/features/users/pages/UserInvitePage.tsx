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

/* eslint-disable @typescript-eslint/no-unsafe-member-access */
/* eslint-disable @typescript-eslint/no-explicit-any */

import {
  EmbeddedFlowComponentType,
  EmbeddedFlowEventType,
  InviteUser,
  type EmbeddedFlowComponent,
  type InviteUserRenderProps,
} from '@asgardeo/react';
import {zodResolver} from '@hookform/resolvers/zod';
import {useLogger} from '@thunder/logger/react';
import {useTemplateLiteralResolver} from '@thunder/shared-hooks';
import {
  Box,
  Stack,
  Typography,
  Button,
  Alert,
  AlertTitle,
  CircularProgress,
  TextField,
  IconButton,
  FormControl,
  FormLabel,
  Select,
  MenuItem,
  LinearProgress,
  Breadcrumbs,
} from '@wso2/oxygen-ui';
import {X, Copy, Check, ChevronRight, CheckCircle, UserPlus} from '@wso2/oxygen-ui-icons-react';
import {useState, useEffect, useMemo, useCallback, useRef, type JSX} from 'react';
import {useForm, Controller} from 'react-hook-form';
import {useTranslation} from 'react-i18next';
import {useNavigate} from 'react-router';
import {z} from 'zod';
import OrganizationUnitTreePicker from '../../organization-units/components/OrganizationUnitTreePicker';

/** Typed shape for flow sub-components */
type FlowSubComponent = EmbeddedFlowComponent & {
  placeholder?: string;
  required?: boolean;
  options?: unknown[];
  hint?: string;
  variant?: string;
  eventType?: string;
};

/**
 * Derive the current step label from flow components.
 * The backend sends HEADING_1 text component as step title.
 */
function deriveStepLabel(
  components: EmbeddedFlowComponent[],
  resolve: (key: string) => string | undefined,
  t: ReturnType<typeof useTranslation>['t'],
): string {
  const heading = components.find(
    (comp) =>
      (String(comp.type) === String(EmbeddedFlowComponentType.Text) || comp.type === 'TEXT') &&
      (comp as FlowSubComponent).variant === 'HEADING_1' &&
      typeof comp.label === 'string',
  );

  if (heading && typeof heading.label === 'string') {
    return t(resolve(heading.label) ?? heading.label);
  }

  return '';
}

const getOptionValue = (option: unknown): string => {
  if (typeof option === 'string') return option;
  if (typeof option === 'object' && option !== null && 'value' in option) {
    const {value} = option as {value: unknown};
    if (typeof value === 'string') return value;
    return JSON.stringify(value ?? option);
  }
  return JSON.stringify(option);
};

const getOptionLabel = (option: unknown): string => {
  if (typeof option === 'string') return option;
  if (typeof option === 'object' && option !== null && 'label' in option) {
    const {label} = option as {label: unknown};
    if (typeof label === 'string') return label;
    return JSON.stringify(label ?? option);
  }
  return JSON.stringify(option);
};

/**
 * Inner content component that renders the current flow step's form fields.
 */
function InviteUserStepContent({
  renderProps,
  flowError,
  handleClose,
  handleCopy,
  copied,
  onResetLocalState,
}: {
  renderProps: InviteUserRenderProps;
  flowError: string | null;
  handleClose: () => void;
  handleCopy: () => void;
  copied: boolean;
  onResetLocalState: () => void;
}): JSX.Element {
  const {
    additionalData,
    values,
    error,
    isLoading,
    components,
    handleInputChange,
    handleSubmit,
    isInviteGenerated,
    isEmailSent,
    inviteLink,
    copyInviteLink,
    inviteLinkCopied,
    resetFlow,
    isValid: propsIsValid,
  } = renderProps;
  const {resolve} = useTemplateLiteralResolver();
  const {t} = useTranslation();

  const buildFormSchema = useMemo(
    () =>
      (comps: EmbeddedFlowComponent[]): z.ZodObject<Record<string, z.ZodTypeAny>> => {
        const shape: Record<string, z.ZodTypeAny> = {};

        const processComponents = (compList: EmbeddedFlowComponent[]) => {
          compList.forEach((comp) => {
            if (
              (String(comp.type) === String(EmbeddedFlowComponentType.Block) || comp.type === 'BLOCK') &&
              comp.components
            ) {
              processComponents(comp.components);
            } else if (
              (String(comp.type) === String(EmbeddedFlowComponentType.TextInput) ||
                comp.type === 'TEXT_INPUT' ||
                comp.type === 'EMAIL_INPUT' ||
                comp.type === 'SELECT' ||
                comp.type === 'OU_SELECT') &&
              comp.ref
            ) {
              let fieldSchema: z.ZodTypeAny = z.string();

              if (comp.type === 'EMAIL_INPUT') {
                fieldSchema = z.string().email('Please enter a valid email address');
              }

              const labelText = typeof comp.label === 'string' ? comp.label : comp.ref;
              if (comp.required) {
                fieldSchema = (fieldSchema as z.ZodString).min(
                  1,
                  `${t(resolve(labelText) ?? labelText) ?? comp.ref} is required`,
                );
              } else {
                fieldSchema = (fieldSchema as z.ZodString).optional();
              }

              shape[comp.ref] = fieldSchema;
            }
          });
        };

        processComponents(comps);
        return z.object(shape);
      },
    [t, resolve],
  );

  const formSchema = useMemo(() => {
    if (!components?.length) return z.object({}) as z.ZodObject<Record<string, z.ZodString>>;
    return buildFormSchema(components as EmbeddedFlowComponent[]);
  }, [components, buildFormSchema]);

  const renderFormField = (
    component: FlowSubComponent,
    index: number,
    formControl: ReturnType<typeof useForm>['control'],
    formErrors: ReturnType<typeof useForm>['formState']['errors'],
    isFormLoading: boolean,
    handleInputChangeFn: (field: string, value: string) => void,
  ) => {
    const {type, ref, label, placeholder, required, options, hint} = component;
    if (!ref) return null;

    const labelText = typeof label === 'string' ? label : '';
    const placeholderText = typeof placeholder === 'string' ? placeholder : '';

    if (String(type) === String(EmbeddedFlowComponentType.TextInput) || type === 'TEXT_INPUT') {
      return (
        <FormControl key={component.id ?? index} required={required}>
          <FormLabel htmlFor={ref}>{t(resolve(labelText) ?? labelText)}</FormLabel>
          <Controller
            name={ref}
            control={formControl}
            rules={{required: required ? `${t(resolve(labelText) ?? labelText)} is required` : false}}
            render={({field}) => (
              <TextField
                {...field}
                fullWidth
                size="small"
                id={ref}
                type="text"
                placeholder={t(resolve(placeholderText) ?? placeholderText)}
                autoComplete="off"
                required={required}
                variant="outlined"
                disabled={isFormLoading}
                error={!!formErrors[ref]}
                helperText={formErrors[ref]?.message as string}
                color={formErrors[ref] ? 'error' : 'primary'}
                onChange={(e) => {
                  field.onChange(e);
                  handleInputChangeFn(ref, e.target.value);
                }}
              />
            )}
          />
        </FormControl>
      );
    }

    if (type === 'EMAIL_INPUT') {
      return (
        <FormControl key={component.id ?? index} required={required}>
          <FormLabel htmlFor={ref}>{t(resolve(labelText) ?? labelText)}</FormLabel>
          <Controller
            name={ref}
            control={formControl}
            rules={{
              required: required ? `${t(resolve(labelText) ?? labelText)} is required` : false,
              pattern: {value: /^[^\s@]+@[^\s@]+\.[^\s@]+$/, message: 'Please enter a valid email address'},
            }}
            render={({field}) => (
              <TextField
                {...field}
                fullWidth
                size="small"
                id={ref}
                type="email"
                placeholder={t(resolve(placeholderText) ?? placeholderText)}
                autoComplete="email"
                required={required}
                variant="outlined"
                disabled={isFormLoading}
                error={!!formErrors[ref]}
                helperText={formErrors[ref]?.message as string}
                color={formErrors[ref] ? 'error' : 'primary'}
                onChange={(e) => {
                  field.onChange(e);
                  handleInputChangeFn(ref, e.target.value);
                }}
              />
            )}
          />
        </FormControl>
      );
    }

    if (type === 'OU_SELECT') {
      return (
        <FormControl key={component.id ?? index} fullWidth required={required}>
          <FormLabel htmlFor={ref}>{t(resolve(labelText) ?? labelText)}</FormLabel>
          <Controller
            name={ref}
            control={formControl}
            rules={{required: required ? `${t(resolve(labelText) ?? labelText)} is required` : false}}
            render={({field}) => (
              <OrganizationUnitTreePicker
                value={(field.value as string) ?? ''}
                onChange={(ouId: string) => {
                  field.onChange(ouId);
                  handleInputChangeFn(ref, ouId);
                }}
                rootOuId={additionalData?.rootOuId as string | undefined}
              />
            )}
          />
          {formErrors[ref] && (
            <Typography variant="caption" color="error">
              {formErrors[ref]?.message as string}
            </Typography>
          )}
        </FormControl>
      );
    }

    if (type === 'SELECT' && options) {
      return (
        <FormControl key={component.id ?? index} fullWidth required={required}>
          <FormLabel htmlFor={ref}>{t(resolve(labelText) ?? labelText)}</FormLabel>
          <Controller
            name={ref}
            control={formControl}
            rules={{required: required ? `${t(resolve(labelText) ?? labelText)} is required` : false}}
            render={({field}) => (
              <>
                <Select
                  {...field}
                  value={(field.value as string | undefined) ?? ''}
                  displayEmpty
                  size="small"
                  id={ref}
                  required={required}
                  fullWidth
                  disabled={isFormLoading}
                  error={!!formErrors[ref]}
                  onChange={(e) => {
                    field.onChange(e);
                    handleInputChangeFn(ref, String(e.target.value));
                  }}
                  renderValue={(selected) => {
                    if (!selected || selected === '') {
                      return (
                        <Typography sx={{color: 'text.secondary'}}>
                          {t(resolve(placeholderText) ?? 'Select an option')}
                        </Typography>
                      );
                    }
                    const selectedOption = options.find((opt: unknown) => getOptionValue(opt) === selected);
                    return selectedOption ? getOptionLabel(selectedOption) : String(selected);
                  }}
                >
                  <MenuItem value="" disabled>
                    {t(resolve(placeholderText) ?? 'Select an option')}
                  </MenuItem>
                  {options.map((option: unknown) => (
                    <MenuItem key={getOptionValue(option)} value={getOptionValue(option)}>
                      {getOptionLabel(option)}
                    </MenuItem>
                  ))}
                </Select>
                {formErrors[ref] && (
                  <Typography variant="caption" color="error.main" sx={{mt: 0.5}}>
                    {formErrors[ref]?.message as string}
                  </Typography>
                )}
                {hint && (
                  <Typography variant="caption" color="text.secondary">
                    {hint}
                  </Typography>
                )}
              </>
            )}
          />
        </FormControl>
      );
    }

    return null;
  };

  const {
    control,
    formState: {errors, isValid},
    reset,
    setValue,
  } = useForm({
    resolver: zodResolver(formSchema),
    mode: 'onChange',
    defaultValues: values ?? {},
  });

  useEffect(() => {
    if (!components?.length && Object.keys(values ?? {}).length === 0) {
      reset({});
    }
  }, [components, values, reset]);

  // Pre-select the root OU (user type's OU) when the OU_SELECT step renders.
  useEffect(() => {
    // Key matches BE constant AdditionalDataKeyRootOUID = "rootOuId"
    const rootOuId = additionalData?.rootOuId as string | undefined;
    if (!rootOuId || !components?.length) return;

    const findOuSelectRef = (comps: EmbeddedFlowComponent[]): string | null => {
      for (const comp of comps) {
        if (comp.type === 'OU_SELECT' && comp.ref) return comp.ref;
        if (comp.components) {
          const found = findOuSelectRef(comp.components);
          if (found) return found;
        }
      }
      return null;
    };

    const ouRef = findOuSelectRef(components as EmbeddedFlowComponent[]);
    if (ouRef && !values?.[ouRef]) {
      setValue(ouRef, rootOuId, {shouldValidate: true});
      handleInputChange(ouRef, rootOuId);
    }
  }, [additionalData, components, values, setValue, handleInputChange]);

  // Loading
  if (isLoading && !components?.length && !isInviteGenerated) {
    return (
      <Box sx={{display: 'flex', justifyContent: 'center', p: 4}}>
        <CircularProgress />
      </Box>
    );
  }

  // Email sent successfully
  if (isInviteGenerated && isEmailSent) {
    return (
      <Box>
        <Alert severity="success" sx={{mb: 3}}>
          <AlertTitle>{t('users:inviteEmailSent', 'Invite Email Sent!')}</AlertTitle>
          {t(
            'users:inviteEmailSentDescription',
            'An invite email has been sent to the user to complete their registration.',
          )}
        </Alert>
        <Stack direction="row" spacing={2} justifyContent="flex-end">
          <Button variant="outlined" onClick={handleClose}>
            {t('common:actions.close', 'Close')}
          </Button>
          <Button
            variant="contained"
            onClick={() => {
              resetFlow();
              onResetLocalState();
            }}
          >
            {t('users:inviteAnother', 'Invite Another User')}
          </Button>
        </Stack>
      </Box>
    );
  }

  // Invite link generated but email not sent
  if (isInviteGenerated && inviteLink) {
    return (
      <Stack alignItems="center" spacing={3} sx={{py: 2, flex: 1, justifyContent: 'center'}}>
        <Box
          sx={{
            width: 64,
            height: 64,
            borderRadius: '50%',
            backgroundColor: 'success.main',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
          }}
        >
          <CheckCircle size={32} color="white" />
        </Box>
        <Box sx={{textAlign: 'center'}}>
          <Typography variant="h6" sx={{mb: 0.5}}>
            {t('users:inviteLinkGenerated', 'Invite Link Generated!')}
          </Typography>
          <Typography variant="body2" color="text.secondary">
            {t('users:inviteLinkDescription', 'Share this link with the user to complete their registration.')}
          </Typography>
        </Box>
        <Box sx={{width: '100%'}}>
          <Typography variant="body2" color="text.secondary" sx={{mb: 0.5}}>
            {t('users:inviteLink', 'Invite Link')}
          </Typography>
          <Box
            sx={{
              display: 'flex',
              alignItems: 'center',
              gap: 1,
              p: 1.5,
              borderRadius: 1,
              backgroundColor: 'background.default',
              border: '1px solid',
              borderColor: 'divider',
            }}
          >
            <Typography
              variant="body2"
              sx={{
                flex: 1,
                fontFamily: 'monospace',
                fontSize: '0.85rem',
                overflow: 'hidden',
                textOverflow: 'ellipsis',
                whiteSpace: 'nowrap',
              }}
            >
              {inviteLink}
            </Typography>
            <Button
              variant={copied || inviteLinkCopied ? 'text' : 'outlined'}
              size="small"
              color={copied || inviteLinkCopied ? 'success' : 'primary'}
              startIcon={copied || inviteLinkCopied ? <Check size={16} /> : <Copy size={16} />}
              onClick={() => {
                copyInviteLink().catch(() => undefined);
                handleCopy();
              }}
              aria-label={t('users:copyInviteLink', 'Copy invite link')}
            >
              {copied || inviteLinkCopied ? t('common:actions.copied', 'Copied!') : t('common:actions.copy', 'Copy')}
            </Button>
          </Box>
        </Box>
        <Stack direction="row" spacing={2} sx={{width: '100%', justifyContent: 'flex-end', pt: 1}}>
          <Button variant="outlined" onClick={handleClose}>
            {t('common:actions.close', 'Close')}
          </Button>
          <Button
            variant="contained"
            startIcon={<UserPlus size={16} />}
            onClick={() => {
              resetFlow();
              onResetLocalState();
            }}
          >
            {t('users:inviteAnother', 'Invite Another User')}
          </Button>
        </Stack>
      </Stack>
    );
  }

  // Error without components
  if (error && !components?.length) {
    return (
      <Box>
        <Alert severity="error" sx={{mb: 2}}>
          <AlertTitle>{t('users:errors.failed.title', 'Error')}</AlertTitle>
          {error.message ?? t('users:errors.failed.description', 'An error occurred.')}
        </Alert>
        <Box sx={{display: 'flex', justifyContent: 'flex-end'}}>
          <Button variant="outlined" onClick={handleClose}>
            {t('common:actions.close', 'Close')}
          </Button>
        </Box>
      </Box>
    );
  }

  // Loading components
  if (!components?.length) {
    return (
      <Box sx={{display: 'flex', justifyContent: 'center', p: 4}}>
        <CircularProgress />
      </Box>
    );
  }

  return (
    <>
      {(flowError ?? error) && (
        <Alert severity="error" sx={{mb: 2}}>
          <AlertTitle>{t('users:errors.failed.title', 'Error')}</AlertTitle>
          {flowError ?? error?.message ?? t('users:errors.failed.description', 'An error occurred.')}
        </Alert>
      )}
      <Stack direction="column" spacing={4}>
        {components.map((component: EmbeddedFlowComponent, index: number) => {
          // TEXT - render headings to match user creation wizard design
          if (String(component.type) === String(EmbeddedFlowComponentType.Text) || component.type === 'TEXT') {
            const variant = typeof component.variant === 'string' ? component.variant : undefined;
            const label = typeof component.label === 'string' ? component.label : '';

            if (variant === 'HEADING_1') {
              return (
                <Typography key={component.id ?? index} variant="h1" gutterBottom>
                  {t(resolve(label) ?? label)}
                </Typography>
              );
            }

            // Subtitles and body text
            return (
              <Typography
                key={component.id ?? index}
                variant={variant === 'HEADING_2' ? 'h2' : 'body1'}
                color="text.secondary"
              >
                {t(resolve(label) ?? label)}
              </Typography>
            );
          }

          if (String(component.type) === String(EmbeddedFlowComponentType.Block) || component.type === 'BLOCK') {
            const blockComponents = (component.components ?? []) as FlowSubComponent[];
            const submitAction = blockComponents.find(
              (c) =>
                (String(c.type) === String(EmbeddedFlowComponentType.Action) || c.type === 'ACTION') &&
                (String(c.eventType) === String(EmbeddedFlowEventType.Submit) || c.eventType === 'SUBMIT'),
            );

            if (!submitAction) return null;

            const isButtonDisabled = isLoading || !isValid || (propsIsValid !== undefined && !propsIsValid);

            return (
              <Box
                key={component.id ?? index}
                component="form"
                onSubmit={(e) => {
                  e.preventDefault();
                  if (!isButtonDisabled) {
                    handleSubmit(submitAction, values).catch(() => undefined);
                  }
                }}
                noValidate
                sx={{display: 'flex', flexDirection: 'column', width: '100%', gap: 2}}
              >
                {blockComponents.map((subComponent, compIndex) => {
                  const field = renderFormField(subComponent, compIndex, control, errors, isLoading, handleInputChange);
                  if (field) return field;

                  // Submit button
                  if (
                    (String(subComponent.type) === String(EmbeddedFlowComponentType.Action) ||
                      subComponent.type === 'ACTION') &&
                    (String(subComponent.eventType) === String(EmbeddedFlowEventType.Submit) ||
                      subComponent.eventType === 'SUBMIT')
                  ) {
                    const subLabel = typeof subComponent.label === 'string' ? subComponent.label : '';
                    return (
                      <Stack
                        key={subComponent.id ?? compIndex}
                        direction="row"
                        spacing={2}
                        justifyContent="flex-end"
                        sx={{mt: 4}}
                      >
                        <Button
                          type="submit"
                          variant={subComponent.variant === 'PRIMARY' ? 'contained' : 'outlined'}
                          disabled={isButtonDisabled}
                          sx={{minWidth: 140}}
                        >
                          {isLoading ? (
                            <CircularProgress size={20} color="inherit" />
                          ) : (
                            t(resolve(subLabel) ?? subLabel)
                          )}
                        </Button>
                      </Stack>
                    );
                  }

                  return null;
                })}
              </Box>
            );
          }

          return null;
        })}
      </Stack>
    </>
  );
}

/** Inner component that bridges InviteUser render props with parent state via useEffect */
function InviteUserFlowBridge({
  renderProps,
  flowError,
  handleClose,
  handleCopy,
  copied,
  onStepLabelChange,
  onInviteComplete,
  onOuStepDetected,
  onResetLocalState,
}: {
  renderProps: InviteUserRenderProps;
  flowError: string | null;
  handleClose: () => void;
  handleCopy: () => void;
  copied: boolean;
  onStepLabelChange: (label: string) => void;
  onInviteComplete: () => void;
  onOuStepDetected: () => void;
  onResetLocalState: () => void;
}): JSX.Element {
  const {resolve} = useTemplateLiteralResolver();
  const {t} = useTranslation();
  const {isInviteGenerated} = renderProps;
  const components = renderProps.components as EmbeddedFlowComponent[] | undefined;

  // Derive current step label from the HEADING_1 component
  const currentStepLabel = components?.length ? deriveStepLabel(components, resolve, t) : '';

  // Detect OU step presence to adjust progress calculation
  const currentHasOu =
    components?.some((c) => c.type === 'OU_SELECT' || c.components?.some((sub) => sub.type === 'OU_SELECT')) ?? false;

  // Update breadcrumb trail and OU detection via useEffect to avoid render-time state updates
  useEffect(() => {
    if (currentHasOu) {
      onOuStepDetected();
    }
  }, [currentHasOu, onOuStepDetected]);

  useEffect(() => {
    if (currentStepLabel) {
      onStepLabelChange(currentStepLabel);
    }
  }, [currentStepLabel, onStepLabelChange]);

  useEffect(() => {
    if (isInviteGenerated) {
      onInviteComplete();
    }
  }, [isInviteGenerated, onInviteComplete]);

  return (
    <InviteUserStepContent
      renderProps={renderProps}
      flowError={flowError}
      handleClose={handleClose}
      handleCopy={handleCopy}
      copied={copied}
      onResetLocalState={onResetLocalState}
    />
  );
}

export default function UserInvitePage(): JSX.Element {
  const {t} = useTranslation();
  const navigate = useNavigate();
  const logger = useLogger('UserInvitePage');
  const [copied, setCopied] = useState(false);
  const [flowError, setFlowError] = useState<string | null>(null);

  // Track breadcrumb trail of visited step labels
  const [breadcrumbs, setBreadcrumbs] = useState<string[]>([]);
  const prevStepLabelRef = useRef<string>('');
  const [hasOuStep, setHasOuStep] = useState(false);

  const handleCopy = () => {
    setCopied(true);
    setTimeout(() => setCopied(false), 3000);
  };

  const handleClose = useCallback(() => {
    setCopied(false);
    (async () => {
      await navigate('/users');
    })().catch((err: unknown) => {
      logger.error('Failed to navigate to users page', {error: err});
    });
  }, [navigate, logger]);

  const handleStepLabelChange = useCallback(
    (label: string) => {
      if (label !== prevStepLabelRef.current) {
        prevStepLabelRef.current = label;
        setBreadcrumbs((prev) => {
          const existingIndex = prev.indexOf(label);
          if (existingIndex >= 0) {
            return prev.slice(0, existingIndex + 1);
          }
          return [...prev, label];
        });
      }
    },
    [setBreadcrumbs],
  );

  const handleInviteComplete = useCallback(() => {
    if (prevStepLabelRef.current !== 'complete') {
      prevStepLabelRef.current = 'complete';
      setBreadcrumbs((prev) => [...prev, t('users:invite.steps.complete', 'Complete')]);
    }
  }, [setBreadcrumbs, t]);

  const handleOuStepDetected = useCallback(() => {
    setHasOuStep(true);
  }, []);

  const handleResetLocalState = useCallback(() => {
    setBreadcrumbs([]);
    prevStepLabelRef.current = '';
    setHasOuStep(false);
    setFlowError(null);
    setCopied(false);
  }, []);

  // Compute progress from breadcrumb trail
  // Without OU step: 3 steps (user type, email, user details + credential)
  // With OU step: 4 steps (user type, OU, email, user details + credential)
  const totalSteps = hasOuStep ? 4 : 3;
  const progress = Math.min((breadcrumbs.length / totalSteps) * 100, 100);

  return (
    <Box sx={{minHeight: '100vh', display: 'flex', flexDirection: 'column'}}>
      {/* Progress bar */}
      <LinearProgress variant="determinate" value={progress} sx={{height: 6}} />

      <Box sx={{flex: 1, display: 'flex', flexDirection: 'column'}}>
        {/* Header with close button and breadcrumb */}
        <Box sx={{p: 4, display: 'flex', justifyContent: 'space-between', alignItems: 'center'}}>
          <Stack direction="row" alignItems="center" spacing={2}>
            <IconButton
              aria-label={t('common:actions.close', 'Close')}
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
              {breadcrumbs.map((label, index) => {
                const isLast = index === breadcrumbs.length - 1;
                return (
                  <Typography key={label} variant="h5" color={isLast ? 'text.primary' : 'inherit'}>
                    {label}
                  </Typography>
                );
              })}
              {breadcrumbs.length === 0 && (
                <Typography variant="h5" color="text.primary">
                  {t('users:inviteUser', 'Invite User')}
                </Typography>
              )}
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
              mx: 'auto',
              alignItems: 'center',
            }}
          >
            <Box
              sx={{
                width: '100%',
                maxWidth: 800,
                flex: 1,
                display: 'flex',
                flexDirection: 'column',
              }}
            >
              <InviteUser
                onInviteLinkGenerated={() => {
                  logger.info('Invite link generated');
                }}
                onError={(err: Error) => {
                  logger.error('User onboarding error', {error: err});
                }}
                onFlowChange={(response: any) => {
                  setFlowError((response?.failureReason as string | null) ?? null);
                }}
              >
                {(renderProps: InviteUserRenderProps) => (
                  <InviteUserFlowBridge
                    renderProps={renderProps}
                    flowError={flowError}
                    handleClose={handleClose}
                    handleCopy={handleCopy}
                    copied={copied}
                    onStepLabelChange={handleStepLabelChange}
                    onInviteComplete={handleInviteComplete}
                    onOuStepDetected={handleOuStepDetected}
                    onResetLocalState={handleResetLocalState}
                  />
                )}
              </InviteUser>
            </Box>
          </Box>
        </Box>
      </Box>
    </Box>
  );
}
