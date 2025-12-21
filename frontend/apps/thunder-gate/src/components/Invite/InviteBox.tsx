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
  CircularProgress,
} from '@wso2/oxygen-ui';
import {Eye, EyeClosed} from '@wso2/oxygen-ui-icons-react';
import {useNavigate, useSearchParams} from 'react-router';
import {useState, useEffect, useCallback} from 'react';
import {useTranslation} from 'react-i18next';
import ROUTES from '../../constants/routes';

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

interface InvitationValidationResponse {
  valid: boolean;
  userID?: string;
  email?: string;
  message?: string;
}

interface InviteBoxState {
  isLoading: boolean;
  isValidating: boolean;
  isSubmitting: boolean;
  error: string | null;
  invitationValid: boolean;
  userEmail: string;
  userID: string;
  password: string;
  confirmPassword: string;
  showPassword: boolean;
  showConfirmPassword: boolean;
  success: boolean;
}

export default function InviteBox(): JSX.Element {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const {t} = useTranslation();

  const token = searchParams.get('token');
  const appId = searchParams.get('app_id') || searchParams.get('appId');

  const [state, setState] = useState<InviteBoxState>({
    isLoading: true,
    isValidating: true,
    isSubmitting: false,
    error: null,
    invitationValid: false,
    userEmail: '',
    userID: '',
    password: '',
    confirmPassword: '',
    showPassword: false,
    showConfirmPassword: false,
    success: false,
  });

  const validateToken = useCallback(async () => {
    if (!token) {
      setState((prev) => ({
        ...prev,
        isLoading: false,
        isValidating: false,
        error: t('invite:error.no_token', 'Invalid invitation link. No token provided.'),
      }));
      return;
    }

    try {
      const response = await fetch(`/api/invitations/validate?token=${encodeURIComponent(token)}`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        throw new Error('Failed to validate invitation');
      }

      const data: InvitationValidationResponse = await response.json() as InvitationValidationResponse;

      if (data.valid) {
        setState((prev) => ({
          ...prev,
          isLoading: false,
          isValidating: false,
          invitationValid: true,
          userEmail: data.email || '',
          userID: data.userID || '',
        }));
      } else {
        setState((prev) => ({
          ...prev,
          isLoading: false,
          isValidating: false,
          error: data.message || t('invite:error.invalid', 'This invitation is invalid or has expired.'),
        }));
      }
    } catch {
      setState((prev) => ({
        ...prev,
        isLoading: false,
        isValidating: false,
        error: t('invite:error.validation_failed', 'Failed to validate invitation. Please try again later.'),
      }));
    }
  }, [token, t]);

  useEffect(() => {
    void validateToken();
  }, [validateToken]);

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();

    // Validate passwords
    if (state.password !== state.confirmPassword) {
      setState((prev) => ({
        ...prev,
        error: t('invite:error.password_mismatch', 'Passwords do not match.'),
      }));
      return;
    }

    if (state.password.length < 8) {
      setState((prev) => ({
        ...prev,
        error: t('invite:error.password_too_short', 'Password must be at least 8 characters long.'),
      }));
      return;
    }

    setState((prev) => ({...prev, isSubmitting: true, error: null}));

    try {
      const response = await fetch('/api/invitations/redeem', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          token,
          password: state.password,
          applicationId: appId,
        }),
      });

      if (!response.ok) {
        const errorData = await response.json() as {error?: string; message?: string};
        throw new Error(errorData.message || errorData.error || 'Failed to complete registration');
      }

      setState((prev) => ({
        ...prev,
        isSubmitting: false,
        success: true,
      }));

      // Redirect to sign in after a short delay
      setTimeout(() => {
        void navigate(ROUTES.AUTH.SIGN_IN);
      }, 2000);
    } catch (err) {
      setState((prev) => ({
        ...prev,
        isSubmitting: false,
        error: err instanceof Error ? err.message : t('invite:error.submission_failed', 'Failed to complete registration. Please try again.'),
      }));
    }
  };

  const togglePasswordVisibility = (field: 'showPassword' | 'showConfirmPassword') => {
    setState((prev) => ({
      ...prev,
      [field]: !prev[field],
    }));
  };

  // Loading state
  if (state.isLoading) {
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
          <Box sx={{display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: 200}}>
            <CircularProgress />
          </Box>
        </StyledPaper>
      </Stack>
    );
  }

  // Error state (invalid invitation)
  if (state.error && !state.invitationValid) {
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
          <Alert severity="error">
            <AlertTitle>{t('invite:error.title', 'Invalid Invitation')}</AlertTitle>
            {state.error}
          </Alert>
          <Button
            variant="contained"
            onClick={() => navigate(ROUTES.AUTH.SIGN_IN)}
            sx={{mt: 2}}
          >
            {t('invite:action.go_to_signin', 'Go to Sign In')}
          </Button>
        </StyledPaper>
      </Stack>
    );
  }

  // Success state
  if (state.success) {
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
          <Alert severity="success">
            <AlertTitle>{t('invite:success.title', 'Registration Complete!')}</AlertTitle>
            {t('invite:success.message', 'Your account has been activated. Redirecting to sign in...')}
          </Alert>
        </StyledPaper>
      </Stack>
    );
  }

  // Normal form state
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
        <Typography variant="h2" sx={{mb: 1}}>
          {t('invite:title', 'Complete Your Registration')}
        </Typography>
        <Typography variant="body1" sx={{mb: 2, color: 'text.secondary'}}>
          {t('invite:subtitle', 'You have been invited to create an account. Please set your password to continue.')}
        </Typography>

        {state.error && (
          <Alert severity="error" sx={{mb: 2}}>
            <AlertTitle>{t('Error')}</AlertTitle>
            {state.error}
          </Alert>
        )}

        {state.userEmail && (
          <Alert severity="info" sx={{mb: 2}}>
            {t('invite:info.email', 'You are registering with: {{email}}', {email: state.userEmail})}
          </Alert>
        )}

        <Box
          component="form"
          onSubmit={(e) => {
            void handleSubmit(e);
          }}
          noValidate
          sx={{display: 'flex', flexDirection: 'column', width: '100%', gap: 2}}
        >
          <FormControl>
            <FormLabel htmlFor="password">{t('invite:field.password', 'Password')}</FormLabel>
            <TextField
              id="password"
              name="password"
              type={state.showPassword ? 'text' : 'password'}
              placeholder={t('invite:field.password_placeholder', 'Enter your password')}
              autoComplete="new-password"
              autoFocus
              required
              fullWidth
              variant="outlined"
              disabled={state.isSubmitting}
              value={state.password}
              onChange={(e) => setState((prev) => ({...prev, password: e.target.value, error: null}))}
              slotProps={{
                input: {
                  endAdornment: (
                    <InputAdornment position="end">
                      <IconButton
                        aria-label="toggle password visibility"
                        onClick={() => togglePasswordVisibility('showPassword')}
                        edge="end"
                        disabled={state.isSubmitting}
                      >
                        {state.showPassword ? <EyeClosed /> : <Eye />}
                      </IconButton>
                    </InputAdornment>
                  ),
                },
              }}
            />
          </FormControl>

          <FormControl>
            <FormLabel htmlFor="confirmPassword">{t('invite:field.confirm_password', 'Confirm Password')}</FormLabel>
            <TextField
              id="confirmPassword"
              name="confirmPassword"
              type={state.showConfirmPassword ? 'text' : 'password'}
              placeholder={t('invite:field.confirm_password_placeholder', 'Confirm your password')}
              autoComplete="new-password"
              required
              fullWidth
              variant="outlined"
              disabled={state.isSubmitting}
              value={state.confirmPassword}
              onChange={(e) => setState((prev) => ({...prev, confirmPassword: e.target.value, error: null}))}
              slotProps={{
                input: {
                  endAdornment: (
                    <InputAdornment position="end">
                      <IconButton
                        aria-label="toggle confirm password visibility"
                        onClick={() => togglePasswordVisibility('showConfirmPassword')}
                        edge="end"
                        disabled={state.isSubmitting}
                      >
                        {state.showConfirmPassword ? <EyeClosed /> : <Eye />}
                      </IconButton>
                    </InputAdornment>
                  ),
                },
              }}
            />
          </FormControl>

          <Button
            type="submit"
            fullWidth
            variant="contained"
            disabled={state.isSubmitting || !state.password || !state.confirmPassword}
            sx={{mt: 2}}
          >
            {state.isSubmitting ? t('invite:action.completing', 'Completing Registration...') : t('invite:action.submit', 'Complete Registration')}
          </Button>
        </Box>
      </StyledPaper>
    </Stack>
  );
}
