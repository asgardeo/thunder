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

import {useMemo, type JSX} from 'react';
import {createTheme, ThemeProvider} from '@mui/material/styles';
import {
  Avatar,
  Box,
  Button,
  CircularProgress,
  Divider,
  FormControl,
  FormLabel,
  Paper,
  Stack,
  TextField,
  Typography,
} from '@wso2/oxygen-ui';
import {AppWindowMac} from '@wso2/oxygen-ui-icons-react';
import {useTranslation} from 'react-i18next';
import type {ThemeConfig} from '../models/theme';
import type {LayoutConfig} from '@/features/layouts/models/layout';

/**
 * Props for ThemePreview component
 */
interface ThemePreviewProps {
  /** Theme configuration to preview */
  theme: ThemeConfig;
  /** Optional layout configuration to apply */
  layout?: LayoutConfig | null;
  /** Which color scheme to preview (defaults to defaultColorScheme) */
  activeColorScheme?: 'light' | 'dark';
  /** Preview width */
  width?: number;
  /** Preview height */
  height?: number;
}

/**
 * Enhanced preview component that shows an actual login page with the theme applied.
 */
export default function ThemePreview({
  theme: themeConfig,
  layout,
  activeColorScheme,
  width,
  height,
}: ThemePreviewProps): JSX.Element {
  const {t} = useTranslation();

  // Handle null themeConfig
  if (!themeConfig) {
    return (
      <Box
        sx={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          height: '100%',
          bgcolor: 'background.default',
        }}
      >
        <CircularProgress />
      </Box>
    );
  }

  // Determine which color scheme to show
  const displayScheme = activeColorScheme || themeConfig.defaultColorScheme;
  const colors = themeConfig.colorSchemes[displayScheme].colors;

  // Create a live MUI theme from the configuration
  const previewTheme = useMemo(() => {
    return createTheme({
      palette: {
        mode: displayScheme,
        primary: {
          main: colors.primary.main,
          dark: colors.primary.dark,
          contrastText: colors.primary.contrastText,
        },
        secondary: {
          main: colors.secondary.main,
          dark: colors.secondary.dark,
          contrastText: colors.secondary.contrastText,
        },
        background: {
          default: colors.background.default,
          paper: colors.background.paper,
        },
        text: {
          primary: colors.text.primary,
          secondary: colors.text.secondary,
        },
      },
      shape: {
        borderRadius: parseInt(themeConfig.shape.borderRadius) || 8,
      },
      typography: {
        fontFamily: themeConfig.typography.fontFamily,
      },
    });
  }, [themeConfig, displayScheme, colors]);

  return (
    <ThemeProvider theme={previewTheme}>
      <Box
        sx={{
          bgcolor: 'background.default',
          backgroundAttachment: 'fixed',
          backgroundImage: `
            radial-gradient(circle at 25% 15%, ${colors.primary.main}33 0%, rgba(255,255,255,0) 60%),
            radial-gradient(circle at 50% 40%, ${colors.secondary.main}22 0%, rgba(255,255,255,0) 20%),
            radial-gradient(circle at center, ${colors.background.default}ee 0%, ${colors.background.default} 100%)
          `,
          backgroundBlendMode: displayScheme === 'dark' ? 'screen' : 'normal',
          height: '100%',
          minHeight: 500,
          borderRadius: 2,
          border: '1px solid',
          borderColor: 'divider',
          overflow: 'hidden',
        }}
      >
        {/* Browser mockup header */}
        <Box
          sx={{
            px: 2,
            py: 1,
            bgcolor: 'background.paper',
            borderBottom: '1px solid',
            borderColor: 'divider',
            opacity: 0.95,
          }}
        >
          <Typography
            variant="subtitle2"
            sx={{
              display: 'flex',
              gap: 1,
              alignItems: 'center',
              color: 'text.secondary',
            }}
          >
            <AppWindowMac />
            {t('themes:builder.preview', {defaultValue: 'Preview'})} - Sign In
          </Typography>
        </Box>

        {/* Login page content */}
        <Box
          sx={{
            display: 'flex',
            justifyContent: 'center',
            flexDirection: 'column',
            alignItems: 'center',
            height: 'calc(100% - 48px)',
            p: 3,
          }}
        >
          {/* App logo placeholder */}
          <Box
            sx={{
              display: 'flex',
              justifyContent: 'center',
              mb: 3,
            }}
          >
            <Avatar
              sx={{
                width: 56,
                height: 56,
                bgcolor: 'primary.main',
                color: 'primary.contrastText',
                fontSize: '1.5rem',
                fontWeight: 600,
              }}
            >
              A
            </Avatar>
          </Box>

          {/* Login form */}
          <Paper
            elevation={layout?.screens?.signin?.container?.elevation ?? 2}
            sx={{
              width: '100%',
              maxWidth: layout?.screens?.signin?.container?.maxWidth ?? 400,
              bgcolor: layout?.screens?.signin?.container?.background ?? 'background.paper',
              borderRadius: layout?.screens?.signin?.container?.borderRadius ?? undefined,
            }}
          >
            <Box sx={{p: layout?.screens?.signin?.container?.padding ? `${layout.screens.signin.container.padding}px` : 4}}>
              <Stack alignItems="center" spacing={1} sx={{mb: 3}}>
                <Typography variant="h5" sx={{width: '100%', textAlign: 'center', color: 'text.primary'}}>
                  {t('themes:preview.signIn', {defaultValue: 'Sign In'})}
                </Typography>
                <Typography variant="body2" sx={{color: 'text.secondary', textAlign: 'center'}}>
                  {t('themes:preview.welcomeBack', {defaultValue: 'Welcome back! Please sign in to continue.'})}
                </Typography>
              </Stack>

              {/* Username/Password form */}
              <Box
                component="form"
                onSubmit={(e) => e.preventDefault()}
                sx={{display: 'flex', flexDirection: 'column', gap: 2}}
              >
                <FormControl required>
                  <FormLabel htmlFor="preview-username" sx={{color: 'text.primary', mb: 0.5}}>
                    {t('themes:preview.username', {defaultValue: 'Username'})}
                  </FormLabel>
                  <TextField
                    id="preview-username"
                    type="text"
                    placeholder={t('themes:preview.usernamePlaceholder', {defaultValue: 'Enter your username'})}
                    fullWidth
                    size="small"
                    disabled
                    sx={{
                      '& .MuiInputBase-root': {
                        bgcolor: 'background.default',
                      },
                    }}
                  />
                </FormControl>

                <FormControl required>
                  <FormLabel htmlFor="preview-password" sx={{color: 'text.primary', mb: 0.5}}>
                    {t('themes:preview.password', {defaultValue: 'Password'})}
                  </FormLabel>
                  <TextField
                    id="preview-password"
                    type="password"
                    placeholder={t('themes:preview.passwordPlaceholder', {defaultValue: 'Enter your password'})}
                    fullWidth
                    size="small"
                    disabled
                    sx={{
                      '& .MuiInputBase-root': {
                        bgcolor: 'background.default',
                      },
                    }}
                  />
                </FormControl>

                <Button
                  type="submit"
                  fullWidth
                  variant="contained"
                  color="primary"
                  sx={{
                    mt: 1,
                    py: 1,
                    textTransform: 'none',
                    fontSize: '0.9375rem',
                  }}
                >
                  {t('themes:preview.signInButton', {defaultValue: 'Sign In'})}
                </Button>
              </Box>

              {/* Divider */}
              <Divider sx={{my: 3, color: 'text.secondary'}}>
                {t('themes:preview.dividerText', {defaultValue: 'or'})}
              </Divider>

              {/* Social login buttons */}
              <Stack spacing={1.5}>
                <Button
                  fullWidth
                  variant="outlined"
                  disabled
                  sx={{
                    textTransform: 'none',
                    color: 'text.primary',
                    borderColor: 'divider',
                  }}
                >
                  {t('themes:preview.continueWithGoogle', {defaultValue: 'Continue with Google'})}
                </Button>
                <Button
                  fullWidth
                  variant="outlined"
                  disabled
                  sx={{
                    textTransform: 'none',
                    color: 'text.primary',
                    borderColor: 'divider',
                  }}
                >
                  {t('themes:preview.continueWithGithub', {defaultValue: 'Continue with GitHub'})}
                </Button>
              </Stack>

              {/* Footer links */}
              <Stack direction="row" justifyContent="center" spacing={1} sx={{mt: 3}}>
                <Typography variant="body2" sx={{color: 'text.secondary'}}>
                  {t('themes:preview.noAccount', {defaultValue: "Don't have an account?"})}
                </Typography>
                <Typography
                  variant="body2"
                  sx={{
                    color: 'primary.main',
                    fontWeight: 500,
                    cursor: 'pointer',
                    '&:hover': {
                      textDecoration: 'underline',
                    },
                  }}
                >
                  {t('themes:preview.signUp', {defaultValue: 'Sign Up'})}
                </Typography>
              </Stack>
            </Box>
          </Paper>

          {/* Privacy notice */}
          <Typography
            variant="caption"
            sx={{
              mt: 3,
              color: 'text.secondary',
              textAlign: 'center',
              maxWidth: 400,
            }}
          >
            {t('themes:preview.privacyNotice', {
              defaultValue: 'By signing in, you agree to our Terms of Service and Privacy Policy.',
            })}
          </Typography>
        </Box>
      </Box>
    </ThemeProvider>
  );
}
