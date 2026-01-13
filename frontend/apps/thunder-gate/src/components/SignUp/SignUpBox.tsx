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

/* eslint-disable @typescript-eslint/no-unsafe-assignment */
/* eslint-disable @typescript-eslint/no-explicit-any */
/* eslint-disable @typescript-eslint/no-unsafe-member-access */
/* eslint-disable @typescript-eslint/no-unsafe-call */
/* eslint-disable @typescript-eslint/no-unsafe-return */

import type {JSX} from 'react';
import {
  Box,
  Button,
  FormLabel,
  FormControl,
  Alert,
  TextField,
  Typography,
  styled,
  AlertTitle,
  Paper,
  Stack,
  ColorSchemeImage,
  IconButton,
  InputAdornment,
  Select,
  MenuItem,
  CircularProgress,
  Avatar,
  useTheme,
} from '@wso2/oxygen-ui';
import {EmbeddedFlowComponentType, EmbeddedFlowEventType, SignUp, type EmbeddedFlowComponent} from '@asgardeo/react';
import {Eye, EyeClosed} from '@wso2/oxygen-ui-icons-react';
import {useNavigate, useSearchParams} from 'react-router';
import {useState} from 'react';
import {Trans, useTranslation} from 'react-i18next';
import {useTemplateLiteralResolver} from '@thunder/shared-hooks';
import {mapEmbeddedFlowTextVariant, useBranding} from '@thunder/shared-branding';
import getIntegrationIcon from '../../utils/getIntegrationIcon';
import ROUTES from '../../constants/routes';

interface ComponentWithType {
  type: string;
  [key: string]: unknown;
}

interface ComponentWithRef {
  ref: string;
  [key: string]: unknown;
}

interface ComponentWithLabel {
  label: string;
  [key: string]: unknown;
}

interface ComponentWithId {
  id: string;
  [key: string]: unknown;
}

const StyledPaper = styled(Paper)(({theme}) => ({
  display: 'flex',
  flexDirection: 'column',
  alignSelf: 'center',
  width: '100%',
  padding: theme.spacing(4),
  gap: theme.spacing(2),
  [theme.breakpoints.up('sm')]: {
    width: '450px',
  },
}));

export default function SignUpBox(): JSX.Element {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const {resolve} = useTemplateLiteralResolver();
  const {t} = useTranslation();
  const {images, theme: brandingTheme, isBrandingEnabled} = useBranding();
  const theme = useTheme();

  const [showPasswordMap, setShowPasswordMap] = useState<Record<string, boolean>>({});

  const currentParams = searchParams.toString();
  const signInUrl = currentParams ? `${ROUTES.AUTH.SIGN_IN}?${currentParams}` : ROUTES.AUTH.SIGN_IN;

  const togglePasswordVisibility = (identifier: string): void => {
    setShowPasswordMap((prev: Record<string, boolean>) => ({
      ...prev,
      [identifier]: !prev[identifier],
    }));
  };

  return (
    <Stack gap={2}>
      <ColorSchemeImage
        src={{
          light: images?.logo?.primary?.url ?? `${import.meta.env.BASE_URL}/assets/images/logo.svg`,
          dark: images?.logo?.primary?.url ?? `${import.meta.env.BASE_URL}/assets/images/logo-inverted.svg`,
        }}
        alt={{
          light: images?.logo?.primary?.alt ?? 'Logo (Light)',
          dark: images?.logo?.primary?.alt ?? 'Logo (Dark)',
        }}
        height={images?.logo?.primary?.height ?? 30}
        width={images?.logo?.primary?.width ?? 'auto'}
        sx={{
          display: {xs: 'flex', md: 'none'},
        }}
      />
      {/* Add branded logo above the sign-up box for desktop */}
      {images?.logo?.primary?.url && (
        <Box sx={{display: {xs: 'none', md: 'flex'}, justifyContent: 'center', mb: 2}}>
          <Avatar
            src={images.logo.primary.url}
            alt={images.logo.primary.alt ?? 'Logo'}
            sx={{
              width: 64,
              height: 64,
              p: 1,
              ...theme.applyStyles('light', {
                backgroundColor: brandingTheme?.palette?.primary?.main ?? theme.palette.primary.main,
              }),
              ...theme.applyStyles('dark', {
                backgroundColor: brandingTheme?.palette?.primary?.main ?? theme.palette.primary.main,
              }),
            }}
          />
        </Box>
      )}
      <StyledPaper variant="outlined">
        <SignUp
          shouldRedirectAfterSignUp={false}
          onComplete={() => {
            // eslint-disable-next-line @typescript-eslint/no-floating-promises
            navigate(signInUrl);
          }}
        >
          {({values, fieldErrors, error, touched, handleInputChange, handleSubmit, isLoading, components}: any) => (
            <>
              {!components ? (
                <Box sx={{display: 'flex', justifyContent: 'center', p: 3}}>
                  <CircularProgress />
                </Box>
              ) : (
                <>
                  {error && (
                    <Alert severity="error" sx={{mb: 2}}>
                      <AlertTitle>{t('signup:errors.signup.failed.message')}</AlertTitle>
                      {error.message ?? t('signup:errors.signup.failed.description')}
                    </Alert>
                  )}

                  {/* Handle different flow types */}
                  {(() => {
                    // Handle new component structure with TEXT and BLOCK types
                    if (components && components.length > 0) {
                      return (
                        <Box sx={{display: 'flex', flexDirection: 'column', gap: 2}}>
                          {components.map(
                            (
                              component: ComponentWithType &
                                ComponentWithId &
                                ComponentWithLabel & {
                                  variant?: string;
                                  components?: EmbeddedFlowComponent[];
                                },
                              index: number,
                            ) => {
                              // Handle TEXT components (headings)
                              if ((component.type as EmbeddedFlowComponentType) === EmbeddedFlowComponentType.Text) {
                                return (
                                  <Typography
                                    key={component.id ?? index}
                                    variant={mapEmbeddedFlowTextVariant(component.variant)}
                                    sx={{mb: 1, textAlign: isBrandingEnabled ? 'center' : 'left'}}
                                  >
                                    {t(resolve(component.label)!)}
                                  </Typography>
                                );
                              }

                              // Handle BLOCK components (form blocks or action blocks)
                              if ((component.type as EmbeddedFlowComponentType) === EmbeddedFlowComponentType.Block) {
                                const blockComponents: EmbeddedFlowComponent[] = component.components ?? [];

                                // Check if this block contains a SUBMIT action (form block) or TRIGGER actions (social login block)
                                const hasSubmitAction = blockComponents.some(
                                  (blockComponent: EmbeddedFlowComponent) =>
                                    ((blockComponent.type as EmbeddedFlowComponentType) ===
                                      EmbeddedFlowComponentType.Action &&
                                      blockComponent.eventType === EmbeddedFlowEventType.Submit) ||
                                    (blockComponent.type === 'RESEND' &&
                                      blockComponent.eventType === EmbeddedFlowEventType.Submit),
                                );

                                const hasTriggerAction = blockComponents.some(
                                  (blockComponent: EmbeddedFlowComponent) =>
                                    (blockComponent.type as EmbeddedFlowComponentType) ===
                                      EmbeddedFlowComponentType.Action &&
                                    blockComponent.eventType === EmbeddedFlowEventType.Trigger,
                                );

                                // Render form block with submission logic
                                if (hasSubmitAction) {
                                  return (
                                    <Box
                                      key={component.id ?? index}
                                      component="form"
                                      onSubmit={(event) => {
                                        event.preventDefault();
                                        // Find submit action
                                        const submitAction: EmbeddedFlowComponent = blockComponents.find(
                                          (blockComponent: EmbeddedFlowComponent) =>
                                            (blockComponent.type as EmbeddedFlowComponentType) ===
                                              EmbeddedFlowComponentType.Action &&
                                            blockComponent.eventType === EmbeddedFlowEventType.Submit,
                                        )!;
                                        if (submitAction) {
                                          // Tracker: https://github.com/asgardeo/javascript/issues/222
                                          handleSubmit(submitAction, values).catch(() => {
                                            // Error handled by error prop
                                          });
                                        }
                                      }}
                                      noValidate
                                      sx={{display: 'flex', flexDirection: 'column', width: '100%', gap: 2}}
                                    >
                                      {(component.components ?? []).map(
                                        (_subComponent: EmbeddedFlowComponent, compIndex: number) => {
                                          const subComponent: ComponentWithType &
                                            ComponentWithRef &
                                            ComponentWithLabel &
                                            ComponentWithId & {
                                              placeholder?: string;
                                              required?: boolean;
                                              options?: string[];
                                              hint?: string;
                                              variant?: string;
                                              eventType?: string;
                                            } = _subComponent as unknown as ComponentWithType &
                                            ComponentWithRef &
                                            ComponentWithLabel &
                                            ComponentWithId & {
                                              placeholder?: string;
                                              required?: boolean;
                                              options?: string[];
                                              hint?: string;
                                              variant?: string;
                                              eventType?: string;
                                            };

                                          // Handle TEXT_INPUT components
                                          if (
                                            (subComponent.type as EmbeddedFlowComponentType) ===
                                              EmbeddedFlowComponentType.TextInput &&
                                            subComponent.ref
                                          ) {
                                            const fieldName = subComponent.ref;
                                            const hasError: boolean = touched?.[fieldName] && fieldErrors?.[fieldName];

                                            return (
                                              <FormControl key={subComponent.id ?? compIndex}>
                                                <FormLabel htmlFor={subComponent.ref}>
                                                  {t(resolve(subComponent.label)!)}
                                                </FormLabel>
                                                <TextField
                                                  error={!!hasError}
                                                  helperText={hasError ? fieldErrors?.[fieldName] : undefined}
                                                  id={subComponent.ref}
                                                  name={subComponent.ref}
                                                  type="text"
                                                  placeholder={t(resolve(subComponent.placeholder)!)}
                                                  autoComplete={(() => {
                                                    if (subComponent.ref === 'username') return 'username';
                                                    if (subComponent.ref === 'email') return 'email';
                                                    return 'off';
                                                  })()}
                                                  autoFocus={subComponent.ref === 'firstName'}
                                                  required={subComponent.required}
                                                  fullWidth
                                                  variant="outlined"
                                                  color={hasError ? 'error' : 'primary'}
                                                  disabled={isLoading}
                                                  value={values?.[fieldName] ?? ''}
                                                  onChange={(e) => handleInputChange(fieldName, e.target.value)}
                                                />
                                              </FormControl>
                                            );
                                          }

                                          // Handle PASSWORD_INPUT components
                                          if (
                                            (subComponent.type as EmbeddedFlowComponentType) ===
                                              EmbeddedFlowComponentType.PasswordInput &&
                                            subComponent.ref
                                          ) {
                                            const fieldName = subComponent.ref;
                                            const hasError = touched?.[fieldName] && fieldErrors?.[fieldName];
                                            const showPasswordForField = showPasswordMap[subComponent.ref] ?? false;

                                            return (
                                              <FormControl key={subComponent.id ?? compIndex}>
                                                <FormLabel htmlFor={subComponent.ref}>
                                                  {t(resolve(subComponent.label)!)}
                                                </FormLabel>
                                                <TextField
                                                  error={!!hasError}
                                                  helperText={hasError ? fieldErrors?.[fieldName] : undefined}
                                                  id={subComponent.ref}
                                                  name={subComponent.ref}
                                                  type={showPasswordForField ? 'text' : 'password'}
                                                  placeholder={t(resolve(subComponent.placeholder)!)}
                                                  autoComplete={
                                                    subComponent.ref === 'password' ? 'new-password' : 'off'
                                                  }
                                                  required={subComponent.required}
                                                  fullWidth
                                                  variant="outlined"
                                                  color={hasError ? 'error' : 'primary'}
                                                  disabled={isLoading}
                                                  value={values?.[fieldName] ?? ''}
                                                  onChange={(e) => handleInputChange(fieldName, e.target.value)}
                                                  slotProps={{
                                                    input: {
                                                      endAdornment: (
                                                        <InputAdornment position="end">
                                                          <IconButton
                                                            aria-label="toggle password visibility"
                                                            onClick={() => togglePasswordVisibility(subComponent.ref)}
                                                            edge="end"
                                                            disabled={isLoading}
                                                          >
                                                            {showPasswordForField ? <EyeClosed /> : <Eye />}
                                                          </IconButton>
                                                        </InputAdornment>
                                                      ),
                                                    },
                                                  }}
                                                />
                                              </FormControl>
                                            );
                                          }

                                          // Handle EMAIL_INPUT components
                                          if (subComponent.type === 'EMAIL_INPUT' && subComponent.ref) {
                                            const fieldName = subComponent.ref;
                                            const hasError: boolean = touched?.[fieldName] && fieldErrors?.[fieldName];

                                            return (
                                              <FormControl key={subComponent.id ?? compIndex}>
                                                <FormLabel htmlFor={subComponent.ref}>
                                                  {t(resolve(subComponent.label)!)}
                                                </FormLabel>
                                                <TextField
                                                  error={!!hasError}
                                                  helperText={hasError ? fieldErrors?.[fieldName] : undefined}
                                                  id={subComponent.ref}
                                                  name={subComponent.ref}
                                                  type="email"
                                                  placeholder={t(resolve(subComponent.placeholder)!)}
                                                  autoComplete="email"
                                                  required={subComponent.required}
                                                  fullWidth
                                                  variant="outlined"
                                                  color={hasError ? 'error' : 'primary'}
                                                  disabled={isLoading}
                                                  value={values?.[fieldName] ?? ''}
                                                  onChange={(e) => handleInputChange(fieldName, e.target.value)}
                                                />
                                              </FormControl>
                                            );
                                          }

                                          // Handle SELECT components
                                          if (
                                            subComponent.type === 'SELECT' &&
                                            subComponent.options &&
                                            subComponent.ref
                                          ) {
                                            const fieldName = subComponent.ref;
                                            const hasError = touched?.[fieldName] && fieldErrors?.[fieldName];

                                            return (
                                              <FormControl key={subComponent.id ?? compIndex} fullWidth>
                                                <FormLabel htmlFor={subComponent.ref}>
                                                  {t(resolve(subComponent.label)!)}
                                                </FormLabel>
                                                <Select
                                                  displayEmpty
                                                  size="small"
                                                  id={subComponent.ref}
                                                  name={subComponent.ref}
                                                  required={subComponent.required}
                                                  fullWidth
                                                  disabled={isLoading}
                                                  error={!!hasError}
                                                  value={values?.[fieldName] ?? ''}
                                                  onChange={(e) => handleInputChange(fieldName, e.target.value)}
                                                >
                                                  <MenuItem value="" disabled>
                                                    {subComponent.placeholder ?? 'Select an option'}
                                                  </MenuItem>
                                                  {subComponent.options.map((option: string) => (
                                                    <MenuItem key={option} value={option}>
                                                      {option}
                                                    </MenuItem>
                                                  ))}
                                                </Select>
                                                {hasError && fieldErrors?.[fieldName] && (
                                                  <Typography variant="caption" color="error.main" sx={{mt: 0.5}}>
                                                    {fieldErrors[fieldName]}
                                                  </Typography>
                                                )}
                                                {subComponent.hint && (
                                                  <Typography variant="caption" color="text.secondary">
                                                    {subComponent.hint}
                                                  </Typography>
                                                )}
                                              </FormControl>
                                            );
                                          }

                                          // Handle PHONE_INPUT components
                                          if (
                                            subComponent.type === 'PHONE_INPUT' &&
                                            subComponent.ref &&
                                            typeof subComponent.ref === 'string'
                                          ) {
                                            const fieldName = subComponent.ref;
                                            const hasError = touched?.[fieldName] && fieldErrors?.[fieldName];

                                            return (
                                              <FormControl
                                                key={subComponent.id ?? compIndex}
                                                required={subComponent.required}
                                              >
                                                <FormLabel htmlFor={subComponent.ref}>
                                                  {t(resolve(subComponent.label)!)}
                                                </FormLabel>
                                                <TextField
                                                  fullWidth
                                                  error={!!hasError}
                                                  helperText={hasError ? fieldErrors?.[fieldName] : undefined}
                                                  id={subComponent.ref}
                                                  name={subComponent.ref}
                                                  type="tel"
                                                  placeholder={t(
                                                    resolve(subComponent.placeholder) ?? subComponent.placeholder ?? '',
                                                  )}
                                                  autoComplete="tel"
                                                  required={subComponent.required}
                                                  variant="outlined"
                                                  color={hasError ? 'error' : 'primary'}
                                                  disabled={isLoading}
                                                  value={values?.[fieldName] ?? ''}
                                                  onChange={(e) => handleInputChange(fieldName, e.target.value)}
                                                />
                                              </FormControl>
                                            );
                                          }

                                          // Handle OTP_INPUT components
                                          if (
                                            subComponent.type === 'OTP_INPUT' &&
                                            subComponent.ref &&
                                            typeof subComponent.ref === 'string'
                                          ) {
                                            const fieldName = subComponent.ref;
                                            const hasError = touched?.[fieldName] && fieldErrors?.[fieldName];
                                            const otpLength = 6;
                                            const otpValue = values?.[fieldName] ?? '';
                                            const otpDigits = otpValue
                                              .padEnd(otpLength, ' ')
                                              .split('')
                                              .slice(0, otpLength);

                                            return (
                                              <FormControl
                                                key={subComponent.id ?? compIndex}
                                                required={subComponent.required}
                                              >
                                                <FormLabel htmlFor={subComponent.ref}>
                                                  {t(resolve(subComponent.label)!)}
                                                </FormLabel>
                                                <Box
                                                  sx={{
                                                    display: 'flex',
                                                    gap: 1,
                                                    justifyContent: 'center',
                                                    mt: 1,
                                                  }}
                                                >
                                                  {otpDigits.map((digit: string, idx: number) => (
                                                    <TextField
                                                      key={`otp-${subComponent.ref}`}
                                                      slotProps={{
                                                        htmlInput: {
                                                          maxLength: 1,
                                                          style: {textAlign: 'center', fontSize: '1.5rem'},
                                                          'aria-label': `OTP digit ${idx + 1}`,
                                                        },
                                                      }}
                                                      value={digit.trim()}
                                                      onChange={(e) => {
                                                        const {value} = e.target;
                                                        if (!/^\d*$/.test(value)) return;

                                                        const newOtp = otpDigits.map((d: string, i: number) =>
                                                          i === idx ? value : d.trim(),
                                                        );
                                                        handleInputChange(subComponent.ref, newOtp.join(''));

                                                        // Auto-focus next input
                                                        if (value && idx < otpLength - 1) {
                                                          // eslint-disable-next-line @typescript-eslint/no-unnecessary-type-assertion
                                                          const nextInput = document.querySelector(
                                                            `input[aria-label="OTP digit ${idx + 2}"]`,
                                                          )! as HTMLInputElement;
                                                          nextInput.focus();
                                                        }
                                                      }}
                                                      onKeyDown={(e) => {
                                                        // Handle backspace to move to previous input
                                                        if (
                                                          e.key === 'Backspace' &&
                                                          !otpDigits[idx].trim() &&
                                                          idx > 0
                                                        ) {
                                                          // eslint-disable-next-line @typescript-eslint/no-unnecessary-type-assertion
                                                          const prevInput = document.querySelector(
                                                            `input[aria-label="OTP digit ${idx}"]`,
                                                          )! as HTMLInputElement;
                                                          prevInput.focus();
                                                        }
                                                      }}
                                                      onPaste={(e) => {
                                                        e.preventDefault();
                                                        const pastedData = e.clipboardData.getData('text/plain');
                                                        const digits = pastedData
                                                          .replace(/\D/g, '')
                                                          .slice(0, otpLength);
                                                        handleInputChange(subComponent.ref, digits);

                                                        // Focus last filled input
                                                        const lastIdx = Math.min(digits.length, otpLength - 1);
                                                        // eslint-disable-next-line @typescript-eslint/no-unnecessary-type-assertion
                                                        const lastInput = document.querySelector(
                                                          `input[aria-label="OTP digit ${lastIdx + 1}"]`,
                                                        )! as HTMLInputElement;
                                                        lastInput.focus();
                                                      }}
                                                      error={!!hasError}
                                                      disabled={isLoading}
                                                      variant="outlined"
                                                      sx={{
                                                        width: 48,
                                                        '& input': {
                                                          padding: '12px 8px',
                                                        },
                                                      }}
                                                    />
                                                  ))}
                                                </Box>
                                                {hasError && fieldErrors?.[fieldName] && (
                                                  <Typography variant="caption" color="error" sx={{mt: 0.5, ml: 1.75}}>
                                                    {fieldErrors[fieldName]}
                                                  </Typography>
                                                )}
                                                )
                                              </FormControl>
                                            );
                                          }

                                          // Handle ACTION components (submit buttons)
                                          if (
                                            (subComponent.type as EmbeddedFlowComponentType) ===
                                              EmbeddedFlowComponentType.Action &&
                                            subComponent.eventType === EmbeddedFlowEventType.Submit
                                          ) {
                                            return (
                                              <Button
                                                key={subComponent.id ?? compIndex}
                                                type="submit"
                                                fullWidth
                                                variant={subComponent.variant === 'PRIMARY' ? 'contained' : 'outlined'}
                                                disabled={isLoading}
                                                sx={{mt: 2}}
                                              >
                                                {isLoading ? t('Creating account...') : t(resolve(subComponent.label)!)}
                                              </Button>
                                            );
                                          }

                                          // Handle RESEND components
                                          if (
                                            subComponent.type === 'RESEND' &&
                                            subComponent.eventType === EmbeddedFlowEventType.Submit
                                          ) {
                                            return (
                                              <Button
                                                key={subComponent.id ?? compIndex}
                                                type="submit"
                                                fullWidth
                                                variant="text"
                                                disabled={isLoading}
                                                sx={{mt: 1}}
                                              >
                                                {t(resolve(subComponent.label)!)}
                                              </Button>
                                            );
                                          }

                                          // Handle TRIGGER action buttons within forms (e.g., OTP verify button)
                                          if (
                                            (subComponent.type as EmbeddedFlowComponentType) ===
                                              EmbeddedFlowComponentType.Action &&
                                            subComponent.eventType === EmbeddedFlowEventType.Trigger
                                          ) {
                                            return (
                                              <Button
                                                key={subComponent.id ?? compIndex}
                                                fullWidth
                                                variant={subComponent.variant === 'PRIMARY' ? 'contained' : 'outlined'}
                                                disabled={isLoading}
                                                onClick={() => {
                                                  // Tracker: https://github.com/asgardeo/javascript/issues/222
                                                  handleSubmit(subComponent, values).catch(() => {
                                                    // Error handled by error prop
                                                  });
                                                }}
                                                sx={{mt: 2}}
                                              >
                                                {t(resolve(subComponent.label)!)}
                                              </Button>
                                            );
                                          }

                                          return null;
                                        },
                                      )}
                                    </Box>
                                  );
                                }

                                // Render action block for social login buttons
                                if (hasTriggerAction) {
                                  return (
                                    <Box
                                      key={component.id ?? index}
                                      sx={{display: 'flex', flexDirection: 'column', width: '100%', gap: 2, mt: 2}}
                                    >
                                      {blockComponents.map(
                                        (_actionComponent: EmbeddedFlowComponent, actionIndex: number) => {
                                          const actionComponent = _actionComponent as ComponentWithType &
                                            ComponentWithId &
                                            ComponentWithLabel & {
                                              eventType?: string;
                                              variant?: string;
                                              image?: string;
                                            };

                                          // Handle TRIGGER action buttons (social login)
                                          if (
                                            (actionComponent.type as EmbeddedFlowComponentType) ===
                                              EmbeddedFlowComponentType.Action &&
                                            actionComponent.eventType === EmbeddedFlowEventType.Trigger
                                          ) {
                                            return (
                                              <Button
                                                key={actionComponent.id ?? actionIndex}
                                                fullWidth
                                                variant="outlined"
                                                disabled={isLoading}
                                                onClick={() => {
                                                  // Tracker: https://github.com/asgardeo/javascript/issues/222
                                                  handleSubmit(actionComponent, values ?? {}).catch(() => {
                                                    // Error handled by error prop
                                                  });
                                                }}
                                                startIcon={getIntegrationIcon(
                                                  actionComponent.label ?? '',
                                                  actionComponent.image ?? '',
                                                )}
                                              >
                                                {t(resolve(actionComponent.label)!)}
                                              </Button>
                                            );
                                          }

                                          return null;
                                        },
                                      )}
                                    </Box>
                                  );
                                }

                                return null;
                              }

                              return null;
                            },
                          )}
                        </Box>
                      );
                    }

                    // SignUpBox fallback error
                    return (
                      <Alert severity="error" sx={{mb: 2}}>
                        <AlertTitle>{t("Oops, that didn't work")}</AlertTitle>
                        {t("We're sorry, we ran into a problem. Please try again!")}
                      </Alert>
                    );
                  })()}
                </>
              )}

              <Typography sx={{textAlign: 'center', mt: 3}}>
                <Trans i18nKey="signup:redirect.to.signin">
                  Already have an account?
                  <Button
                    variant="text"
                    onClick={() => {
                      // eslint-disable-next-line @typescript-eslint/no-floating-promises
                      navigate(signInUrl);
                    }}
                    sx={{
                      p: 0,
                      minWidth: 'auto',
                      textTransform: 'none',
                      color: 'primary.main',
                      textDecoration: 'underline',
                      '&:hover': {
                        textDecoration: 'underline',
                        backgroundColor: 'transparent',
                      },
                    }}
                  >
                    Sign in
                  </Button>
                </Trans>
              </Typography>
            </>
          )}
        </SignUp>
      </StyledPaper>
    </Stack>
  );
}
