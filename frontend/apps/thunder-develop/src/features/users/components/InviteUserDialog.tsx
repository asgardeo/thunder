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

/* eslint-disable @typescript-eslint/no-unsafe-member-access */
/* eslint-disable @typescript-eslint/no-explicit-any */

import {useState, useEffect, useMemo, type JSX} from 'react';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogContentText,
  Box,
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
  Stepper,
  Step,
  StepLabel,
} from '@wso2/oxygen-ui';
import {X, Copy, Check} from '@wso2/oxygen-ui-icons-react';
import {useTranslation} from 'react-i18next';
import {useForm, Controller} from 'react-hook-form';
import {z} from 'zod';
import {zodResolver} from '@hookform/resolvers/zod';
import {
  EmbeddedFlowComponentType,
  EmbeddedFlowEventType,
  InviteUser,
  type EmbeddedFlowComponent,
  type InviteUserRenderProps,
} from '@asgardeo/react';
import {mapEmbeddedFlowTextVariant} from '@thunder/shared-design';
import {useTemplateLiteralResolver} from '@thunder/shared-hooks';
import {useLogger} from '@thunder/logger/react';

/** Typed shape for flow sub-components */
type FlowSubComponent = EmbeddedFlowComponent & {
  placeholder?: string;
  required?: boolean;
  options?: unknown[];
  hint?: string;
  variant?: string;
  eventType?: string;
};

export interface InviteUserDialogProps {
  open: boolean;
  onClose: () => void;
  onSuccess?: (inviteLink: string) => void;
}

interface InviteUserContentProps {
  props: InviteUserRenderProps;
  getOptionValue: (option: unknown) => string;
  getOptionLabel: (option: unknown) => string;
  handleClose: () => void;
  handleCopy: () => void;
  copied: boolean;
  setActiveStep: (step: number) => void;
}

const STEPS = ['User Details', 'Invite Link'];

function InviteUserContent({
  props: {
    values,
    error,
    isLoading,
    components,
    handleInputChange,
    handleSubmit,
    isInviteGenerated,
    inviteLink,
    copyInviteLink,
    inviteLinkCopied,
    resetFlow,
    isValid: propsIsValid,
  },
  getOptionValue,
  getOptionLabel,
  handleClose,
  handleCopy,
  copied,
  setActiveStep,
}: InviteUserContentProps): JSX.Element {
  const {resolve} = useTemplateLiteralResolver();
  const {t} = useTranslation();

  /**
   * Build Zod schema dynamically based on flow components
   */
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
                comp.type === 'SELECT') &&
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

  // Build form schema dynamically from components
  const formSchema = useMemo(() => {
    if (!components?.length) return z.object({}) as z.ZodObject<Record<string, z.ZodString>>;
    return buildFormSchema(components as EmbeddedFlowComponent[]);
  }, [components, buildFormSchema]);

  /**
   * Render form field using react-hook-form Controller
   */
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

    // TEXT_INPUT
    if (String(type) === String(EmbeddedFlowComponentType.TextInput) || type === 'TEXT_INPUT') {
      return (
        <FormControl key={component.id ?? index} required={required}>
          <FormLabel htmlFor={ref}>{t(resolve(labelText) ?? labelText)}</FormLabel>
          <Controller
            name={ref}
            control={formControl}
            rules={{
              required: required ? `${t(resolve(labelText) ?? labelText)} is required` : false,
            }}
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

    // EMAIL_INPUT
    if (type === 'EMAIL_INPUT') {
      return (
        <FormControl key={component.id ?? index} required={required}>
          <FormLabel htmlFor={ref}>{t(resolve(labelText) ?? labelText)}</FormLabel>
          <Controller
            name={ref}
            control={formControl}
            rules={{
              required: required ? `${t(resolve(labelText) ?? labelText)} is required` : false,
              pattern: {
                value: /^[^\s@]+@[^\s@]+\.[^\s@]+$/,
                message: 'Please enter a valid email address',
              },
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

    // SELECT
    if (type === 'SELECT' && options) {
      return (
        <FormControl key={component.id ?? index} fullWidth>
          <FormLabel htmlFor={ref}>{t(resolve(labelText) ?? labelText)}</FormLabel>
          <Controller
            name={ref}
            control={formControl}
            rules={{
              required: required ? `${t(resolve(labelText) ?? labelText)} is required` : false,
            }}
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

  // Initialize react-hook-form
  const {
    control,
    formState: {errors, isValid},
    setValue,
    reset,
  } = useForm({
    resolver: zodResolver(formSchema),
    mode: 'onChange',
    defaultValues: values ?? {},
  });

  // Sync react-hook-form with SDK component values
  useEffect(() => {
    if (values) {
      Object.entries(values).forEach(([key, value]) => {
        setValue(key, String(value));
      });
    }
  }, [values, setValue]);

  // Reset form when flow resets
  useEffect(() => {
    if (!components?.length && Object.keys(values ?? {}).length === 0) {
      reset({});
    }
  }, [components, values, reset]);

  // Loading
  if (isLoading && !components?.length && !isInviteGenerated) {
    return (
      <Box sx={{display: 'flex', justifyContent: 'center', p: 4}}>
        <CircularProgress />
      </Box>
    );
  }

  // Invite generated
  if (isInviteGenerated && inviteLink) {
    return (
      <Box>
        <Alert severity="success" sx={{mb: 3}}>
          <AlertTitle>{t('users:inviteLinkGenerated', 'Invite Link Generated!')}</AlertTitle>
          {t('users:inviteLinkDescription', 'Share this link with the user to complete their registration.')}
        </Alert>
        <Box sx={{mb: 3}}>
          <Typography variant="body2" sx={{mb: 1}}>
            {t('users:inviteLink', 'Invite Link')}
          </Typography>
          <Box sx={{display: 'flex', gap: 1}}>
            <TextField
              fullWidth
              value={inviteLink}
              InputProps={{readOnly: true}}
              size="small"
              sx={{
                '& .MuiInputBase-root': {
                  backgroundColor: 'background.default',
                  fontFamily: 'monospace',
                  fontSize: '0.85rem',
                },
              }}
            />
            <IconButton
              onClick={() => {
                if (copyInviteLink) {
                  copyInviteLink().catch(() => {});
                }
                handleCopy();
              }}
              color={copied || inviteLinkCopied ? 'success' : 'primary'}
              sx={{flexShrink: 0}}
              aria-label={t('users:copyInviteLink', 'Copy invite link')}
            >
              {copied || inviteLinkCopied ? <Check size={18} /> : <Copy size={18} />}
            </IconButton>
          </Box>
        </Box>
        <Box sx={{display: 'flex', gap: 2, justifyContent: 'flex-end'}}>
          <Button variant="outlined" onClick={handleClose}>
            {t('common:actions.close', 'Close')}
          </Button>
          <Button
            variant="contained"
            onClick={() => {
              setActiveStep(0);
              resetFlow();
            }}
          >
            {t('users:inviteAnother', 'Invite Another User')}
          </Button>
        </Box>
      </Box>
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
      {error && (
        <Alert severity="error" sx={{mb: 2}}>
          <AlertTitle>{t('users:errors.failed.title', 'Error')}</AlertTitle>
          {error.message ?? t('users:errors.failed.description', 'An error occurred.')}
        </Alert>
      )}
      <Box sx={{display: 'flex', flexDirection: 'column', gap: 2}}>
        {components.map((component: EmbeddedFlowComponent, index: number) => {
          // TEXT - skip main headings in dialog
          if (String(component.type) === String(EmbeddedFlowComponentType.Text) || component.type === 'TEXT') {
            const variant = typeof component.variant === 'string' ? component.variant : undefined;
            if (variant === 'HEADING_1' || variant === 'HEADING_2') return null;
            const label = typeof component.label === 'string' ? component.label : '';
            return (
              <Typography key={component.id ?? index} variant={mapEmbeddedFlowTextVariant(variant)} sx={{mb: 1}}>
                {t(resolve(label) ?? label)}
              </Typography>
            );
          }

          // BLOCK
          if (String(component.type) === String(EmbeddedFlowComponentType.Block) || component.type === 'BLOCK') {
            const blockComponents = (component.components ?? []) as FlowSubComponent[];
            const submitAction = blockComponents.find(
              (c) =>
                (String(c.type) === String(EmbeddedFlowComponentType.Action) || c.type === 'ACTION') &&
                (String(c.eventType) === String(EmbeddedFlowEventType.Submit) || c.eventType === 'SUBMIT'),
            );

            if (!submitAction) return null;

            // Use propsIsValid from SDK as primary source, fallback to form validation
            const isButtonDisabled = isLoading || (propsIsValid !== undefined ? !propsIsValid : !isValid);

            return (
              <Box
                key={component.id ?? index}
                component="form"
                onSubmit={(e) => {
                  e.preventDefault();
                  if (!isButtonDisabled) {
                    handleSubmit(submitAction, values).catch(() => {});
                  }
                }}
                noValidate
                sx={{display: 'flex', flexDirection: 'column', width: '100%', gap: 2}}
              >
                {blockComponents.map((subComponent, compIndex) => {
                  // Form fields using react-hook-form Controller
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
                      <Box
                        key={subComponent.id ?? compIndex}
                        sx={{display: 'flex', gap: 2, justifyContent: 'flex-end', mt: 2}}
                      >
                        <Button variant="outlined" onClick={handleClose} disabled={isLoading}>
                          {t('common:actions.cancel', 'Cancel')}
                        </Button>
                        <Button
                          type="submit"
                          variant={subComponent.variant === 'PRIMARY' ? 'contained' : 'outlined'}
                          disabled={isButtonDisabled}
                        >
                          {isLoading ? (
                            <CircularProgress size={20} color="inherit" />
                          ) : (
                            t(resolve(subLabel) ?? subLabel)
                          )}
                        </Button>
                      </Box>
                    );
                  }

                  return null;
                })}
              </Box>
            );
          }

          return null;
        })}
      </Box>
    </>
  );
}

export default function InviteUserDialog({open, onClose, onSuccess = undefined}: InviteUserDialogProps): JSX.Element {
  const {t} = useTranslation();
  const logger = useLogger('InviteUserDialog');
  const [copied, setCopied] = useState(false);
  const [activeStep, setActiveStep] = useState(0);

  const handleCopy = () => {
    setCopied(true);
    setTimeout(() => setCopied(false), 3000);
  };

  const handleClose = () => {
    setActiveStep(0);
    setCopied(false);
    onClose();
  };

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

  return (
    <Dialog open={open} onClose={handleClose} maxWidth="sm" fullWidth>
      <DialogTitle sx={{display: 'flex', justifyContent: 'space-between', alignItems: 'center', pb: 1}}>
        <Typography variant="h3" component="h2">
          {t('users:inviteUser', 'Invite User')}
        </Typography>
        <IconButton onClick={handleClose} size="small" aria-label="close">
          <X size={18} />
        </IconButton>
      </DialogTitle>
      <DialogContent>
        <DialogContentText sx={{mb: 3, fontSize: '1rem'}}>
          {t('users:inviteUserDescription', 'Send an invite link to a new user to complete their registration')}
        </DialogContentText>

        <Stepper activeStep={activeStep} sx={{mb: 3}}>
          {STEPS.map((label) => (
            <Step key={label}>
              <StepLabel>{t(`users:invite.steps.${label.toLowerCase().replace(' ', '')}`, label)}</StepLabel>
            </Step>
          ))}
        </Stepper>

        <InviteUser
          onInviteLinkGenerated={(link: string) => {
            setActiveStep(1);
            onSuccess?.(link);
          }}
          onError={(error: Error) => {
            logger.error('User onboarding error', {error});
          }}
        >
          {(props: InviteUserRenderProps) => (
            <InviteUserContent
              props={props}
              getOptionValue={getOptionValue}
              getOptionLabel={getOptionLabel}
              handleClose={handleClose}
              handleCopy={handleCopy}
              copied={copied}
              setActiveStep={setActiveStep}
            />
          )}
        </InviteUser>
      </DialogContent>
    </Dialog>
  );
}
