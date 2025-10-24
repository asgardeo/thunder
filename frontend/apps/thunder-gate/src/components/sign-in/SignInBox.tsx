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

import type {JSX} from 'react';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import MuiCard from '@mui/material/Card';
import Checkbox from '@mui/material/Checkbox';
import Divider from '@mui/material/Divider';
import FormLabel from '@mui/material/FormLabel';
import FormControl from '@mui/material/FormControl';
import Alert from '@mui/material/Alert';
import FormControlLabel from '@mui/material/FormControlLabel';
import Link from '@mui/material/Link';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import {styled} from '@mui/material/styles';
import {useState} from 'react';
import {SignIn} from '@asgardeo/react';
import {Smartphone} from 'lucide-react';
import {Google, Facebook, GitHub} from '@thunder/ui';
import AlertTitle from '@mui/material/AlertTitle';

const Card = styled(MuiCard)(({theme}) => ({
  display: 'flex',
  flexDirection: 'column',
  alignSelf: 'center',
  width: '100%',
  padding: theme.spacing(4),
  gap: theme.spacing(2),
  boxShadow: 'hsla(220, 30%, 5%, 0.05) 0px 5px 15px 0px, hsla(220, 25%, 10%, 0.05) 0px 15px 35px -5px',
  [theme.breakpoints.up('sm')]: {
    width: '450px',
  },
  ...theme.applyStyles('dark', {
    boxShadow: 'hsla(220, 30%, 5%, 0.5) 0px 5px 15px 0px, hsla(220, 25%, 10%, 0.08) 0px 15px 35px -5px',
  }),
}));

export default function SignInBox(): JSX.Element {
  const [emailError, setEmailError] = useState(false);
  const [emailErrorMessage, setEmailErrorMessage] = useState('');
  const [passwordError, setPasswordError] = useState(false);
  const [passwordErrorMessage, setPasswordErrorMessage] = useState('');

  const validateInputs = () => {
    const email = document.getElementById('email') as HTMLInputElement;
    const password = document.getElementById('password') as HTMLInputElement;

    let isValid = true;

    if (!email.value || !/\S+@\S+\.\S+/.test(email.value)) {
      setEmailError(true);
      setEmailErrorMessage('Please enter a valid email address.');
      isValid = false;
    } else {
      setEmailError(false);
      setEmailErrorMessage('');
    }

    if (!password.value || password.value.length < 6) {
      setPasswordError(true);
      setPasswordErrorMessage('Password must be at least 6 characters long.');
      isValid = false;
    } else {
      setPasswordError(false);
      setPasswordErrorMessage('');
    }

    return isValid;
  };

  return (
    <Card variant="outlined">
      <SignIn
        onError={(error) => {
          setEmailError(true);
          setEmailErrorMessage(error?.message || 'Authentication failed');
        }}
      >
        {({onSubmit, isLoading, components, error, isInitialized}) => (
          <>
            <Typography component="h1" variant="h4" sx={{width: '100%', fontSize: 'clamp(2rem, 10vw, 2.15rem)'}}>
              Sign in
            </Typography>

            {!isInitialized ? (
              <Box sx={{display: 'flex', justifyContent: 'center', p: 3}}>
                <Typography>Loading authentication...</Typography>
              </Box>
            ) : (
              <>
                {error && (
                  <Alert severity="error" sx={{mb: 2}}>
                    <AlertTitle>Error</AlertTitle>
                    {error.message || error.toString()}
                  </Alert>
                )}

                {/* Handle different flow types */}
                {(() => {
                  // Check if we have button components (multi-option flow)
                  if (components && components.length > 0 && components.some((c) => c.type === 'BUTTON')) {
                    return (
                      <>
                        {/* Always render BasicAuth form first if basic_auth is available */}
                        {components.some((c) => c.config?.actionId === 'basic_auth') && (
                          <Box
                            component="form"
                            onSubmit={(event) => {
                              event.preventDefault();
                              if (validateInputs()) {
                                const data = new FormData(event.currentTarget);
                                // eslint-disable-next-line @typescript-eslint/no-unsafe-argument, @typescript-eslint/no-explicit-any
                                onSubmit({
                                  actionId: 'basic_auth',
                                  inputs: {
                                    username: data.get('email') as string,
                                    password: data.get('password') as string,
                                  },
                                  // eslint-disable-next-line @typescript-eslint/no-explicit-any
                                } as any).catch(() => {
                                  // Error handled by onError callback
                                });
                              }
                            }}
                            noValidate
                            sx={{display: 'flex', flexDirection: 'column', width: '100%', gap: 2, mb: 2}}
                          >
                            <FormControl>
                              <FormLabel htmlFor="email">Email</FormLabel>
                              <TextField
                                error={emailError}
                                helperText={emailErrorMessage}
                                id="email"
                                type="email"
                                name="email"
                                placeholder="your@email.com"
                                autoComplete="email"
                                autoFocus
                                required
                                fullWidth
                                variant="outlined"
                                color={emailError ? 'error' : 'primary'}
                                disabled={isLoading}
                              />
                            </FormControl>
                            <FormControl>
                              <Box sx={{display: 'flex', justifyContent: 'space-between'}}>
                                <FormLabel htmlFor="password">Password</FormLabel>
                                {/* eslint-disable jsx-a11y/anchor-is-valid */}
                                <Link
                                  component="button"
                                  onClick={() => null}
                                  sx={{p: 0, minWidth: 'auto', fontSize: '0.875rem', alignSelf: 'baseline'}}
                                >
                                  Forgot your password?
                                </Link>
                                {/* eslint-enable jsx-a11y/anchor-is-valid */}
                              </Box>
                              <TextField
                                error={passwordError}
                                helperText={passwordErrorMessage}
                                name="password"
                                placeholder="••••••"
                                type="password"
                                id="password"
                                autoComplete="current-password"
                                required
                                fullWidth
                                variant="outlined"
                                color={passwordError ? 'error' : 'primary'}
                                disabled={isLoading}
                              />
                            </FormControl>
                            <FormControlLabel
                              control={<Checkbox value="remember" color="primary" disabled={isLoading} />}
                              label="Remember me"
                            />
                            <Button type="submit" fullWidth variant="contained" disabled={isLoading}>
                              {isLoading ? 'Signing in...' : 'Sign in'}
                            </Button>
                          </Box>
                        )}

                        {/* Show divider if there are other auth options besides basic_auth */}
                        {components.some((c) => c.type === 'BUTTON' && c.config?.actionId !== 'basic_auth') && (
                          <Divider>or</Divider>
                        )}

                        {/* Show other authentication options as buttons */}
                        <Box sx={{display: 'flex', flexDirection: 'column', gap: 2}}>
                          {components
                            .filter((c) => c.type === 'BUTTON')
                            .filter((c) => c.config?.actionId !== 'basic_auth')
                            .map((button) => {
                              const actionId = String(button.config?.actionId ?? '');

                              const getIcon = () => {
                                if (actionId.includes('google')) return <Google />;
                                if (actionId.includes('facebook')) return <Facebook />;
                                if (actionId.includes('github')) return <GitHub />;
                                if (actionId.includes('mobile')) return <Smartphone />;
                                return null;
                              };

                              const getLabel = () => {
                                if (button.config?.text && typeof button.config.text === 'string') {
                                  return button.config.text;
                                }
                                if (actionId.includes('google')) return 'Sign in with Google';
                                if (actionId.includes('github')) return 'Sign in with GitHub';
                                if (actionId.includes('facebook')) return 'Sign in with Facebook';
                                if (actionId.includes('mobile')) return 'Sign in with SMS OTP';
                                return `Sign in with ${actionId.replace('_auth', '').replace('_', ' ')}`;
                              };

                              return (
                                <Button
                                  key={button.id}
                                  fullWidth
                                  variant="outlined"
                                  onClick={() => {
                                    // eslint-disable-next-line @typescript-eslint/no-unsafe-argument, @typescript-eslint/no-explicit-any
                                    onSubmit({actionId} as any).catch(() => {
                                      // Error handled by onError callback
                                    });
                                  }}
                                  disabled={isLoading}
                                  startIcon={getIcon()}
                                >
                                  {getLabel()}
                                </Button>
                              );
                            })}
                        </Box>
                      </>
                    );
                  }

                  // Handle FORM components
                  if (components && components.length > 0 && components.some((c) => c.type === 'FORM')) {
                    return (
                      <Box sx={{display: 'flex', flexDirection: 'column', gap: 2}}>
                        {components
                          .filter((c) => c.type === 'FORM')
                          .map((form) => (
                            <Box
                              key={form.id}
                              component="form"
                              onSubmit={(event) => {
                                event.preventDefault();
                                const data = new FormData(event.currentTarget);
                                const inputs: Record<string, string> = {};

                                // Extract inputs from form components
                                form.components
                                  ?.filter((c) => c.type === 'INPUT')
                                  .forEach((input) => {
                                    if (input.config?.identifier) {
                                      inputs[input.config.identifier] = data.get(input.config.identifier) as string;
                                    }
                                  });

                                // eslint-disable-next-line @typescript-eslint/no-unsafe-argument, @typescript-eslint/no-explicit-any
                                onSubmit({inputs} as any).catch(() => {
                                  // Error handled by onError callback
                                });
                              }}
                              noValidate
                              sx={{display: 'flex', flexDirection: 'column', width: '100%', gap: 2}}
                            >
                              {form.components?.map((component) => {
                                if (component.type === 'INPUT') {
                                  return (
                                    <FormControl key={component.id}>
                                      <FormLabel htmlFor={component.config?.identifier}>
                                        {component.config?.label}
                                      </FormLabel>
                                      <TextField
                                        id={component.config?.identifier}
                                        name={component.config?.identifier}
                                        type={component.config?.type === 'text' ? 'text' : component.config?.type}
                                        placeholder={component.config?.placeholder}
                                        required={component.config?.required}
                                        fullWidth
                                        variant="outlined"
                                        disabled={isLoading}
                                      />
                                      {component.config?.hint && (
                                        <Typography variant="caption" color="text.secondary">
                                          {component.config.hint}
                                        </Typography>
                                      )}
                                    </FormControl>
                                  );
                                }

                                if (component.type === 'BUTTON' && component.config?.type === 'submit') {
                                  return (
                                    <Button
                                      key={component.id}
                                      type="submit"
                                      fullWidth
                                      variant={component.variant === 'PRIMARY' ? 'contained' : 'outlined'}
                                      disabled={isLoading}
                                    >
                                      {isLoading ? 'Submitting...' : component.config?.text}
                                    </Button>
                                  );
                                }
                                return null;
                              })}
                            </Box>
                          ))}
                      </Box>
                    );
                  }

                  return null;
                })()}

                <Box sx={{textAlign: 'center'}}>
                  <Typography sx={{textAlign: 'center'}}>
                    Don&apos;t have an account?{' '}
                    <span>
                      <Link
                        href="/material-ui/getting-started/templates/sign-in/"
                        variant="body2"
                        sx={{alignSelf: 'center'}}
                      >
                        Sign up
                      </Link>
                    </span>
                  </Typography>
                </Box>
              </>
            )}
          </>
        )}
      </SignIn>
    </Card>
  );
}
