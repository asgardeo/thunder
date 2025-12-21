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
  ColorSchemeImage,
  Stack,
  IconButton,
  InputAdornment,
  type TypographyVariant,
  CircularProgress,
} from '@wso2/oxygen-ui';
import {useState} from 'react';
import {
  EmbeddedFlowComponentType,
  EmbeddedFlowEventType,
  EmbeddedFlowTextVariant,
  SignIn,
  SignUp,
  type EmbeddedFlowComponent,
} from '@asgardeo/react';
import {useTemplateLiteralResolver} from '@thunder/shared-hooks';
import {Eye, EyeClosed} from '@wso2/oxygen-ui-icons-react';
import {useNavigate, useSearchParams} from 'react-router';
import {Trans, useTranslation} from 'react-i18next';
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

export default function SignInBox(): JSX.Element {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const {resolve} = useTemplateLiteralResolver();
  const {t} = useTranslation();

  const [showPasswordMap, setShowPasswordMap] = useState<Record<string, boolean>>({});
  const [formInputs, setFormInputs] = useState<Record<string, string>>({});
  const [formErrors, setFormErrors] = useState<Record<string, string>>({});

  const togglePasswordVisibilityForField = (identifier: string): void => {
    setShowPasswordMap((prev: Record<string, boolean>) => ({
      ...prev,
      [identifier]: !prev[identifier],
    }));
  };

  const validateForm = (components: EmbeddedFlowComponent[]): boolean => {
    const errors: Record<string, string> = {};
    let isValid = true;

    components.forEach((component: EmbeddedFlowComponent) => {
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

  const updateInput = (ref: string | undefined, value: string): void => {
    if (!ref) return;

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
        sx={{
          display: {xs: 'flex', md: 'none'},
        }}
      />
      <StyledPaper variant="outlined">
        <SignIn
          onError={() => {
            // Handle authentication errors
            // Error will be displayed in the UI via the error prop
          }}
        >
          {({onSubmit, isLoading, components, error, isInitialized}) =>
            isLoading || !isInitialized ? (
              <Box sx={{display: 'flex', justifyContent: 'center', p: 3}}>
                <CircularProgress />
              </Box>
            ) : (
              <>
                {error && (
                  <Alert severity="error" sx={{mb: 2}}>
                    <AlertTitle>{t('signin:errors.signin.failed.message')}</AlertTitle>
                    {error.message || error.toString()}
                  </Alert>
                )}
                {(() => {
                  if (components && components.length > 0) {
                    return (
                      <Box sx={{display: 'flex', flexDirection: 'column', gap: 2}}>
                        {components.map((component: EmbeddedFlowComponent, index: number) => {
                          // Handle TEXT components (headings)
                          if ((component.type as EmbeddedFlowComponentType) === EmbeddedFlowComponentType.Text) {
                            const variant: TypographyVariant =
                              component.variant === EmbeddedFlowTextVariant.Heading1 ? 'h2' : 'body1';

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

                                  const blockComponents: EmbeddedFlowComponent[] = component.components ?? [];

                                  if (validateForm(blockComponents)) {
                                    // Find submit action
                                    const submitAction: EmbeddedFlowComponent = blockComponents.find(
                                      (blockComponent: EmbeddedFlowComponent) =>
                                        (blockComponent.type as EmbeddedFlowComponentType) ===
                                          EmbeddedFlowComponentType.Action &&
                                        blockComponent.eventType === EmbeddedFlowEventType.Submit,
                                    )!;

                                    // Tracker: https://github.com/asgardeo/javascript/issues/222
                                    onSubmit({inputs: formInputs, action: submitAction?.id}).catch(() => {
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

                                    if (
                                      (subComponent.type as EmbeddedFlowComponentType) ===
                                        EmbeddedFlowComponentType.TextInput &&
                                      subComponent.ref &&
                                      typeof subComponent.ref === 'string'
                                    ) {
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
                                            error={!!(subComponent.ref && formErrors[subComponent.ref])}
                                            helperText={subComponent.ref ? formErrors[subComponent.ref] : undefined}
                                            id={subComponent.ref}
                                            name={subComponent.ref}
                                            type="text"
                                            placeholder={t(resolve(subComponent.placeholder)!)}
                                            autoComplete={subComponent.ref === 'username' ? 'username' : 'off'}
                                            autoFocus={subComponent.ref === 'username'}
                                            required={subComponent.required}
                                            variant="outlined"
                                            color={
                                              subComponent.ref && formErrors[subComponent.ref] ? 'error' : 'primary'
                                            }
                                            disabled={isLoading}
                                            value={subComponent.ref ? (formInputs[subComponent.ref] ?? '') : ''}
                                            onChange={(e) => updateInput(subComponent.ref, e.target.value)}
                                          />
                                        </FormControl>
                                      );
                                    }

                                    // Handle PASSWORD_INPUT components
                                    if (
                                      (subComponent.type as EmbeddedFlowComponentType) ===
                                        EmbeddedFlowComponentType.PasswordInput &&
                                      subComponent.ref &&
                                      typeof subComponent.ref === 'string'
                                    ) {
                                      const showPasswordForField: boolean = showPasswordMap[subComponent.ref] ?? false;

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
                                            error={!!(subComponent.ref && formErrors[subComponent.ref])}
                                            helperText={subComponent.ref ? formErrors[subComponent.ref] : undefined}
                                            id={subComponent.ref}
                                            name={subComponent.ref}
                                            type={showPasswordForField ? 'text' : 'password'}
                                            placeholder={t(resolve(subComponent.placeholder)!)}
                                            autoComplete={subComponent.ref === 'password' ? 'current-password' : 'off'}
                                            required={subComponent.required}
                                            variant="outlined"
                                            color={
                                              subComponent.ref && formErrors[subComponent.ref] ? 'error' : 'primary'
                                            }
                                            disabled={isLoading}
                                            value={subComponent.ref ? (formInputs[subComponent.ref] ?? '') : ''}
                                            onChange={(e) => updateInput(subComponent.ref, e.target.value)}
                                            slotProps={{
                                              input: {
                                                endAdornment: (
                                                  <InputAdornment position="end">
                                                    <IconButton
                                                      aria-label="toggle password visibility"
                                                      onClick={() => togglePasswordVisibilityForField(subComponent.ref)}
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

                          return null;
                        })}
                      </Box>
                    );
                  }

                  // If no components, show loading
                  return (
                    <Box sx={{display: 'flex', justifyContent: 'center', p: 3}}>
                      <CircularProgress />
                    </Box>
                  );
                })()}
              </>
            )
          }
        </SignIn>

        <SignUp>
          {({components}) => {
            if (components && components.length > 0) {
              return (
                <Typography sx={{textAlign: 'center', mt: 2}}>
                  <Trans i18nKey="signin:redirect.to.signup">
                    Don&apos;t have an account?
                    <Button
                      variant="text"
                      onClick={() => {
                        const currentParams: string = searchParams.toString();
                        const createUrl: string = currentParams
                          ? `${ROUTES.AUTH.SIGN_UP}?${currentParams}`
                          : ROUTES.AUTH.SIGN_UP;

                        // eslint-disable-next-line @typescript-eslint/no-floating-promises
                        navigate(createUrl);
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
                      Sign up
                    </Button>
                  </Trans>
                </Typography>
              );
            }

            return null;
          }}
        </SignUp>
      </StyledPaper>
    </Stack>
  );
}
