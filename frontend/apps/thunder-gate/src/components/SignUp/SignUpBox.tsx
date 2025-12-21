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
} from '@wso2/oxygen-ui';
import {
  EmbeddedFlowComponentType,
  EmbeddedFlowEventType,
  EmbeddedFlowTextVariant,
  SignUp,
  type EmbeddedFlowComponent,
} from '@asgardeo/react';
import {Eye, EyeClosed} from '@wso2/oxygen-ui-icons-react';
import {useNavigate, useSearchParams} from 'react-router';
import {useState} from 'react';
import {Trans, useTranslation} from 'react-i18next';
import {useTemplateLiteralResolver} from '@thunder/shared-hooks';
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

  const [showPasswordMap, setShowPasswordMap] = useState<Record<string, boolean>>({});
  const [formInputs, setFormInputs] = useState<Record<string, string>>({});
  const [formErrors, setFormErrors] = useState<Record<string, string>>({});

  const currentParams = searchParams.toString();
  const signInUrl = currentParams ? `${ROUTES.AUTH.SIGN_IN}?${currentParams}` : ROUTES.AUTH.SIGN_IN;

  const togglePasswordVisibility = (identifier: string): void => {
    setShowPasswordMap((prev: Record<string, boolean>) => ({
      ...prev,
      [identifier]: !prev[identifier],
    }));
  };

  const validateForm = (components: EmbeddedFlowComponent[]): boolean => {
    const errors: Record<string, string> = {};
    let isValid = true;

    components.forEach((component) => {
      if (
        ((component.type as EmbeddedFlowComponentType) === EmbeddedFlowComponentType.TextInput ||
          (component.type as EmbeddedFlowComponentType) === EmbeddedFlowComponentType.PasswordInput) &&
        component.required &&
        component.ref &&
        typeof component.ref === 'string' &&
        typeof component.label === 'string'
      ) {
        const value = formInputs[component.ref] || '';
        if (!value.trim()) {
          errors[component.ref] = `${t('validations:form.field.required', {field: t(resolve(component.label)!)})}`;
          isValid = false;
        }
      }
    });

    setFormErrors(errors);
    return isValid;
  };

  const updateInput = (ref: string, value: string): void => {
    setFormInputs((prev) => ({...prev, [ref]: value}));

    // Clear error when user starts typing
    if (formErrors[ref]) {
      setFormErrors((prev) => ({...prev, [ref]: ''}));
    }
  };

  return (
    <Stack gap={5}>
      <ColorSchemeImage
        src={{
          light: `${import.meta.env.BASE_URL}/assets/images/logo.svg`,
          dark: `${import.meta.env.BASE_URL}/assets/images/logo-inverted.svg`,
        }}
        alt={{light: 'Logo (Light)', dark: 'Logo (Dark)'}}
        height={30}
        width="auto"
      />
      <StyledPaper variant="outlined">
        <SignUp
          shouldRedirectAfterSignUp={false}
          onComplete={() => {
            // eslint-disable-next-line @typescript-eslint/no-floating-promises
            navigate(signInUrl);
          }}
        >
          {({errors, handleSubmit, isLoading, components}) => (
            <>
              {!components ? (
                <Box sx={{display: 'flex', justifyContent: 'center', p: 3}}>
                  <CircularProgress />
                </Box>
              ) : (
                <>
                  {errors && Object.keys(errors).length > 0 && (
                    <Alert severity="error" sx={{mb: 2}}>
                      <AlertTitle>{t('Error')}</AlertTitle>
                      {Object.entries(errors).map(([key, error]: [string, unknown]) => (
                        <Typography key={key} variant="body2">
                          {String(error)}
                        </Typography>
                      ))}
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
                                const variant = component.variant === EmbeddedFlowTextVariant.Heading1 ? 'h2' : 'body1';
                                return (
                                  <Typography key={component.id ?? index} variant={variant} sx={{mb: 1}}>
                                    {t(resolve(component.label)!)}
                                  </Typography>
                                );
                              }

                              // Handle BLOCK components (form blocks)
                              if ((component.type as EmbeddedFlowComponentType) === EmbeddedFlowComponentType.Block) {
                                return (
                                  <Box
                                    key={component.id ?? index}
                                    component="form"
                                    onSubmit={(event) => {
                                      event.preventDefault();
                                      if (validateForm(component.components ?? [])) {
                                        // Tracker: https://github.com/asgardeo/javascript/issues/222
                                        handleSubmit(component, formInputs).catch(() => {
                                          // Error handled by onError callback
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
                                          return (
                                            <FormControl key={subComponent.id ?? compIndex}>
                                              <FormLabel htmlFor={subComponent.ref}>
                                                {t(resolve(subComponent.label)!)}
                                              </FormLabel>
                                              <TextField
                                                error={!!formErrors[subComponent.ref]}
                                                helperText={formErrors[subComponent.ref]}
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
                                                color={formErrors[subComponent.ref] ? 'error' : 'primary'}
                                                disabled={isLoading}
                                                value={formInputs[subComponent.ref] || ''}
                                                onChange={(e) => updateInput(subComponent.ref, e.target.value)}
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
                                          const showPasswordForField = showPasswordMap[subComponent.ref] || false;

                                          return (
                                            <FormControl key={subComponent.id ?? compIndex}>
                                              <FormLabel htmlFor={subComponent.ref}>
                                                {t(resolve(subComponent.label)!)}
                                              </FormLabel>
                                              <TextField
                                                error={!!formErrors[subComponent.ref]}
                                                helperText={formErrors[subComponent.ref]}
                                                id={subComponent.ref}
                                                name={subComponent.ref}
                                                type={showPasswordForField ? 'text' : 'password'}
                                                placeholder={t(resolve(subComponent.placeholder)!)}
                                                autoComplete={subComponent.ref === 'password' ? 'new-password' : 'off'}
                                                required={subComponent.required}
                                                fullWidth
                                                variant="outlined"
                                                color={formErrors[subComponent.ref] ? 'error' : 'primary'}
                                                disabled={isLoading}
                                                value={formInputs[subComponent.ref] || ''}
                                                onChange={(e) => updateInput(subComponent.ref, e.target.value)}
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

                                        // Handle SELECT components
                                        if (
                                          subComponent.type === 'SELECT' &&
                                          subComponent.options &&
                                          subComponent.ref
                                        ) {
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
                                                error={!!formErrors[subComponent.ref]}
                                                value={formInputs[subComponent.ref] || ''}
                                                onChange={(e) => updateInput(subComponent.ref, e.target.value)}
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
                                              {formErrors[subComponent.ref] && (
                                                <Typography variant="caption" color="error.main" sx={{mt: 0.5}}>
                                                  {formErrors[subComponent.ref]}
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

                                        return null;
                                      },
                                    )}
                                  </Box>
                                );
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
