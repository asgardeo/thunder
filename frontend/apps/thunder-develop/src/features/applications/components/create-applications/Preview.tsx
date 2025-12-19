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

import {BaseSignIn, ThemeProvider} from '@asgardeo/react';
import {
  Box,
  Typography,
  Divider,
  useColorScheme,
  useTheme,
  Button,
  TextField,
  FormControl,
  FormLabel,
  Paper,
  Avatar,
  Stack,
} from '@wso2/oxygen-ui';
import {AppWindowMac} from '@wso2/oxygen-ui-icons-react';
import type {JSX} from 'react';
import {useTranslation} from 'react-i18next';
import {type IdentityProvider} from '@/features/integrations/models/identity-provider';
import getIntegrationIcon from '@/features/integrations/utils/getIntegrationIcon';
import {AuthenticatorTypes} from '@/features/integrations/models/authenticators';
import useIdentityProviders from '../../../integrations/api/useIdentityProviders';

/**
 * Props for the {@link Preview} component that displays a live preview of the application sign-in page.
 *
 * @public
 */
export interface PreviewProps {
  /**
   * The name of the application to display in the preview
   */
  appName: string | null;

  /**
   * URL of the application logo to display in the preview
   */
  appLogo: string | null;

  /**
   * The primary color to use for branding elements (e.g., sign-in button)
   */
  selectedColor: string;

  /**
   * Record of enabled authentication integrations
   * Keys are integration IDs, values indicate whether they are enabled
   */
  integrations: Record<string, boolean>;

  /**
   * Whether a custom flow is selected (preview will be blurred)
   */
  isCustomFlowSelected?: boolean;

  /**
   * Name of the custom flow (for display in the overlay)
   */
  customFlowName?: string;
}

/**
 * React component that renders a live preview of the application's sign-in page
 * based on the user's design and authentication configuration choices.
 *
 * This component displays a simulated browser window containing a login interface that
 * reflects the user's selections including:
 * - Application name and logo
 * - Primary brand color for buttons and interactive elements
 * - Enabled authentication methods (username/password, social logins)
 * - Identity provider buttons with appropriate branding
 *
 * The preview updates in real-time as users make changes in the onboarding flow,
 * providing immediate visual feedback of their customization choices. The component
 * fetches identity provider data and displays enabled providers with their respective
 * icons and labels.
 *
 * @param props - The component props
 * @param props.appName - The application name to display in the preview
 * @param props.appLogo - URL of the logo to display in the preview
 * @param props.selectedColor - Hex color code for branding elements
 * @param props.integrations - Record of enabled authentication integrations
 *
 * @returns JSX element displaying the sign-in page preview in a browser mockup
 *
 * @example
 * ```tsx
 * import Preview from './Preview';
 *
 * function OnboardingFlow() {
 *   return (
 *     <Preview
 *       appName="My Application"
 *       appLogo="https://example.com/logo.png"
 *       selectedColor="#FF5733"
 *       integrations={{
 *         'username-password': true,
 *         'google-idp': true
 *       }}
 *     />
 *   );
 * }
 * ```
 *
 * @public
 */
export default function Preview({appName, appLogo, selectedColor, integrations, 
  isCustomFlowSelected = false, customFlowName = undefined}: PreviewProps): JSX.Element {
  const {t} = useTranslation();
  const {mode} = useColorScheme();
  const theme = useTheme();
  const {data: identityProviders} = useIdentityProviders();

  const hasUsernamePassword: boolean = integrations[AuthenticatorTypes.BASIC_AUTH] ?? false;
  const selectedProviders: IdentityProvider[] =
    identityProviders?.filter((idp: IdentityProvider): boolean => integrations[idp.id]) ?? [];
  const hasSocialLogins: boolean = selectedProviders.length > 0;

  return (
    <Box
      sx={{
        backgroundColor: theme.vars?.palette.background.default,
        backgroundAttachment: 'fixed',
        backgroundImage: `
          radial-gradient(circle at 25% 15%, rgba(255, 117, 2, 0.2) 0%, rgba(255,255,255,0) 60%),
          radial-gradient(circle at 50% 40%, rgba(69, 30, 175, 0.1) 0%, rgba(255,255,255,0) 20%),
          radial-gradient(circle at center, rgba(255,255,255,0.9) 0%, ${theme.vars?.palette.background.default} 100%)
        `,
        backgroundBlendMode: mode === 'dark' ? 'screen' : 'normal',
        height: '100%',
        borderRadius: theme.vars?.shape.borderRadius,
        boxShadow: `0 0 0 1px rgba(199, 211, 234, 0.08) inset, 0 24px 48px 0 rgba(168, 216, 245, 0.06) inset, 0 1px 1px 0 rgba(216, 236, 248, 0.20) inset;`,
        position: 'relative',
        ...theme.applyStyles('dark', {
          backgroundImage: `radial-gradient(circle at 30% 10%, rgba(255, 117, 2, 0.3) 0%, rgba(0,0,0,0) 40%),
          radial-gradient(circle at 60% 40%, rgba(69, 30, 175, 0.14) 0%, rgba(0,0,0,0) 70%),
          radial-gradient(circle at center, rgba(0,0,0,0.6) 0%, ${theme.vars?.palette.background.default} 100%)
        `,
        }),
      }}
    >
      <Box
        sx={{
          px: 2,
          py: 1,
          background: `radial-gradient(70.71% 70.71% at 50% 50%, rgba(255, 230, 200, 0.00) 50%, rgba(0, 0, 0, 0.00) 51%), radial-gradient(70.71% 70.71% at 50% 50%, rgba(255, 230, 200, 0.00) 50%, rgba(0, 0, 0, 0.00) 51%), radial-gradient(70.71% 70.71% at 50% 50%, rgba(255, 230, 200, 0.00) 50%, rgba(0, 0, 0, 0.00) 51%), rgba(255, 223, 198, 0.04)`,
          borderBottom: `1px solid ${theme.vars?.palette.divider}`,
          borderTopLeftRadius: theme.vars?.shape.borderRadius,
          borderTopRightRadius: theme.vars?.shape.borderRadius,
        }}
      >
        <Typography
          variant="subtitle2"
          sx={{
            display: 'flex',
            gap: 1,
            alignItems: 'center',
          }}
        >
          <AppWindowMac />
          {t('applications:onboarding.preview.title')}
        </Typography>
      </Box>
      <Box
        sx={{
          display: 'flex',
          justifyContent: 'center',
          flexDirection: 'column',
          alignItems: 'center',
          height: '100%',
          position: 'relative',
        }}
      >
        {/* Blur overlay when custom flow selected - covers only inner content */}
        {isCustomFlowSelected && (
          <Box
            sx={{
              position: 'absolute',
              top: 0,
              left: 0,
              right: 0,
              bottom: 0,
              zIndex: 10,
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              justifyContent: 'center',
              backdropFilter: 'blur(6px)',
              backgroundColor: 'rgba(255, 255, 255, 0.25)',
              ...theme.applyStyles('dark', {
                backgroundColor: 'rgba(0, 0, 0, 0.2)',
              }),
            }}
          >
            <Paper
              elevation={6}
              sx={{
                p: 3,
                textAlign: 'center',
                maxWidth: 280,
                border: `1px solid ${theme.vars?.palette.divider}`,
              }}
            >
              <Typography variant="subtitle1" gutterBottom>
                {t('applications:onboarding.preview.customFlowSelected')}
              </Typography>
              {customFlowName && (
                <Typography variant="body2" color="text.secondary" sx={{mb: 2}}>
                  {customFlowName}
                </Typography>
              )}
              <Typography variant="caption" color="text.secondary">
                {t('applications:onboarding.preview.customFlowHint')}
              </Typography>
            </Paper>
          </Box>
        )}
        {appLogo && (
          <Box
            sx={{
              display: 'flex',
              justifyContent: 'center',
              mb: 2,
            }}
          >
            <Avatar
              src={appLogo}
              sx={{
                width: 64,
                height: 64,
                p: 1,
                ...theme.applyStyles('light', {
                  backgroundColor: selectedColor,
                }),
                ...theme.applyStyles('dark', {
                  backgroundColor: selectedColor,
                }),
              }}
            />
          </Box>
        )}
        <Paper sx={{pointerEvents: 'none', width: 400, position: 'relative'}}>
          <ThemeProvider mode={mode}>
            <Box>
              <BaseSignIn onError={() => {}} onSuccess={() => {}}>
                {() => (
                  <Box sx={{display: 'flex', flexDirection: 'column', gap: 2, p: 4}}>
                    <Stack alignItems="center" spacing={1} sx={{mb: 2}}>
                      <Typography component="h1" variant="h5" sx={{width: '100%', mb: 2, textAlign: 'center'}}>
                        {t('applications:onboarding.preview.signInTo', {appName})}
                      </Typography>
                      <Typography variant="caption">{t('applications:onboarding.preview.welcomeMessage')}</Typography>
                    </Stack>

                    {/* Username/Password form - Conditionally rendered */}
                    {hasUsernamePassword && (
                      <Box
                        component="form"
                        onSubmit={(e) => e.preventDefault()}
                        sx={{display: 'flex', flexDirection: 'column', gap: 2, mb: hasSocialLogins ? 2 : 0}}
                      >
                        <FormControl required>
                          <FormLabel htmlFor="preview-username">
                            {t('applications:onboarding.preview.username')}
                          </FormLabel>
                          <TextField
                            id="preview-username"
                            type="text"
                            placeholder={t('applications:onboarding.preview.usernamePlaceholder')}
                            fullWidth
                            variant="outlined"
                            disabled
                          />
                        </FormControl>
                        <FormControl required>
                          <FormLabel htmlFor="preview-password">
                            {t('applications:onboarding.preview.password')}
                          </FormLabel>
                          <TextField
                            id="preview-password"
                            type="password"
                            placeholder={t('applications:onboarding.preview.passwordPlaceholder')}
                            fullWidth
                            variant="outlined"
                            disabled
                          />
                        </FormControl>
                        <Button
                          type="submit"
                          fullWidth
                          variant="contained"
                          color="secondary"
                          sx={{
                            color: '#fff',
                            backgroundColor: selectedColor,
                            '&:hover': {
                              backgroundColor: selectedColor,
                            },
                          }}
                        >
                          {t('applications:onboarding.preview.signInButton')}
                        </Button>
                      </Box>
                    )}

                    {/* Divider - Show only when both username/password and social logins exist */}
                    {hasUsernamePassword && hasSocialLogins && (
                      <Divider>{t('applications:onboarding.preview.dividerText')}</Divider>
                    )}

                    {/* Social login buttons with actual provider names */}
                    {hasSocialLogins && (
                      <Box sx={{display: 'flex', flexDirection: 'column', gap: 2}}>
                        {selectedProviders.map(
                          (provider: IdentityProvider): JSX.Element => (
                            <Button
                              key={provider.id}
                              fullWidth
                              variant="outlined"
                              disabled
                              startIcon={getIntegrationIcon(provider.type)}
                            >
                              {t('applications:onboarding.preview.continueWith', {providerName: provider.name})}
                            </Button>
                          ),
                        )}
                      </Box>
                    )}
                  </Box>
                )}
              </BaseSignIn>
            </Box>
          </ThemeProvider>
        </Paper>
      </Box>
    </Box>
  );
}
