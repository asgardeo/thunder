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

import {
  Box,
  Typography,
  Stack,
  Switch,
  CircularProgress,
  Alert,
  useTheme,
  List,
  ListItem,
  ListItemButton,
  ListItemIcon,
  ListItemText,
  Divider,
  Autocomplete,
  TextField,
} from '@wso2/oxygen-ui';
import type {JSX} from 'react';
import {useEffect} from 'react';
import {Lightbulb, UserRound, Google, GitHub} from '@wso2/oxygen-ui-icons-react';
import {useTranslation} from 'react-i18next';
import type {BasicFlowDefinition} from '@/features/flows/models/responses';
import {type IdentityProvider, IdentityProviderTypes} from '@/features/integrations/models/identity-provider';
import getIntegrationIcon from '@/features/integrations/utils/getIntegrationIcon';
import {AuthenticatorTypes} from '@/features/integrations/models/authenticators';
import {AUTH_FLOW_HANDLES, SYSTEM_FLOW_HANDLES} from '../../models/auth-flow-graphs';
import useIdentityProviders from '../../../integrations/api/useIdentityProviders';

/**
 * Props for the {@link ConfigureSignInOptions} component.
 *
 * @public
 */
export interface ConfigureSignInOptionsProps {
  /**
   * Record of enabled authentication integrations
   * Keys are integration IDs, values indicate whether they are enabled
   */
  integrations: Record<string, boolean>;

  /**
   * Callback function when an integration toggle state changes
   */
  onIntegrationToggle: (connectionId: string) => void;

  /**
   * Callback function to broadcast whether this step is ready to proceed
   */
  onReadyChange?: (isReady: boolean) => void;

  /**
   * Function to check if a flow is available for a given handle
   * Used to disable sign-in options when their required flows are not configured
   */
  isFlowAvailable?: (handle: string) => boolean;

  /**
   * Whether flow data is currently being loaded
   */
  isLoadingFlows?: boolean;

  /**
   * Currently selected custom flow ID, or null if using toggle-based selection
   */
  customFlowId?: string | null;

  /**
   * Callback when custom flow selection changes
   */
  onCustomFlowChange?: (flowId: string | null) => void;

  /**
   * List of available authentication flows for the dropdown
   */
  authFlows?: BasicFlowDefinition[];
}

/**
 * Check if at least one authentication option is selected
 *
 * @param integrations - Record of integration states
 * @returns True if at least one integration is enabled
 *
 * @internal
 */
const hasAtLeastOneSelected = (integrations: Record<string, boolean>): boolean =>
  Object.values(integrations).some((isEnabled) => isEnabled);

/**
 * React component that renders the sign-in options configuration step in the
 * application creation onboarding flow.
 *
 * This component allows users to configure authentication methods for their application
 * by toggling between:
 * 1. Username & Password authentication (default enabled)
 * 2. Social/Enterprise identity provider integrations (Google, GitHub, etc.)
 *
 * The component fetches available identity providers and displays them as toggleable
 * list items with appropriate icons. Users can enable/disable multiple authentication
 * methods. The step is marked as ready only when at least one authentication option
 * is selected, ensuring applications have a valid sign-in mechanism.
 *
 * @param props - The component props
 * @param props.integrations - Record of enabled integrations (key: integration ID, value: enabled state)
 * @param props.onIntegrationToggle - Callback invoked when an integration is toggled
 * @param props.onReadyChange - Optional callback to notify parent of step readiness
 *
 * @returns JSX element displaying the sign-in options configuration interface
 *
 * @example
 * ```tsx
 * import ConfigureSignInOptions from './ConfigureSignInOptions';
 *
 * function OnboardingFlow() {
 *   const [integrations, setIntegrations] = useState({
 *     'username-password': true,
 *     'google-idp-id': false,
 *   });
 *
 *   const handleToggle = (id: string) => {
 *     setIntegrations(prev => ({
 *       ...prev,
 *       [id]: !prev[id]
 *     }));
 *   };
 *
 *   return (
 *     <ConfigureSignInOptions
 *       integrations={integrations}
 *       onIntegrationToggle={handleToggle}
 *       onReadyChange={(isReady) => console.log('Ready:', isReady)}
 *     />
 *   );
 * }
 * ```
 *
 * @public
 */
export default function ConfigureSignInOptions({
  integrations,
  onIntegrationToggle,
  onReadyChange = undefined,
  isFlowAvailable = undefined,
  isLoadingFlows = false,
  customFlowId = null,
  onCustomFlowChange = undefined,
  authFlows = [],
}: ConfigureSignInOptionsProps): JSX.Element {
  const {t} = useTranslation();
  const theme = useTheme();
  const {data, isLoading, error} = useIdentityProviders();

  /**
   * Broadcast readiness whenever integrations change.
   */
  useEffect((): void => {
    // Ready if using custom flow OR at least one toggle is selected
    const isReady: boolean = customFlowId !== null || hasAtLeastOneSelected(integrations);
    if (onReadyChange) {
      onReadyChange(isReady);
    }
  }, [integrations, customFlowId, onReadyChange]);

  if (isLoading || isLoadingFlows) {
    return (
      <Box sx={{display: 'flex', justifyContent: 'center', alignItems: 'center', py: 8}}>
        <CircularProgress />
      </Box>
    );
  }

  if (error) {
    return (
      <Alert severity="error" sx={{mb: 4}}>
        {t('applications:onboarding.configure.SignInOptions.error', {error: error.message ?? 'Unknown error'})}
      </Alert>
    );
  }

  const availableIntegrations: IdentityProvider[] = data ?? [];
  const googleProvider: IdentityProvider | undefined = availableIntegrations.find(
    (idp: IdentityProvider): boolean => idp.type === IdentityProviderTypes.GOOGLE,
  );
  const githubProvider: IdentityProvider | undefined = availableIntegrations.find(
    (idp: IdentityProvider): boolean => idp.type === IdentityProviderTypes.GITHUB,
  );
  const hasAtLeastOneSelectedOption: boolean = hasAtLeastOneSelected(integrations);
  const hasUsernamePassword: boolean = integrations[AuthenticatorTypes.BASIC_AUTH] ?? false;
  const isUsingCustomFlow: boolean = customFlowId !== null;

  // Check flow availability for validation
  const isBasicFlowAvailable = isFlowAvailable?.(AUTH_FLOW_HANDLES.BASIC) ?? true;
  const isGoogleFlowAvailable = isFlowAvailable?.(AUTH_FLOW_HANDLES.GOOGLE) ?? true;
  const isGitHubFlowAvailable = isFlowAvailable?.(AUTH_FLOW_HANDLES.GITHUB) ?? true;
  const isBasicGoogleFlowAvailable = isFlowAvailable?.(AUTH_FLOW_HANDLES.BASIC_GOOGLE) ?? true;
  const isBasicGitHubFlowAvailable = isFlowAvailable?.(AUTH_FLOW_HANDLES.BASIC_GITHUB) ?? true;
  const isGoogleGitHubFlowAvailable = isFlowAvailable?.(AUTH_FLOW_HANDLES.GOOGLE_GITHUB) ?? true;
  const isBasicGoogleGitHubFlowAvailable = isFlowAvailable?.(AUTH_FLOW_HANDLES.BASIC_GOOGLE_GITHUB) ?? true;

  // Determine if each option should be disabled based on required flows
  // Username/Password requires at least basic flow or any flow that includes basic
  const isBasicAuthDisabled = !isBasicFlowAvailable && !isBasicGoogleFlowAvailable && 
    !isBasicGitHubFlowAvailable && !isBasicGoogleGitHubFlowAvailable;
  // Google requires google flow or any flow that includes google
  const isGoogleDisabled = !isGoogleFlowAvailable && !isBasicGoogleFlowAvailable && 
    !isGoogleGitHubFlowAvailable && !isBasicGoogleGitHubFlowAvailable;
  // GitHub requires github flow or any flow that includes github
  const isGitHubDisabled = !isGitHubFlowAvailable && !isBasicGitHubFlowAvailable && 
    !isGoogleGitHubFlowAvailable && !isBasicGoogleGitHubFlowAvailable;

  return (
    <Stack direction="column" spacing={4}>
      <Stack direction="column" spacing={1}>
        <Typography variant="h1" gutterBottom>
          {t('applications:onboarding.configure.SignInOptions.title')}
        </Typography>
        <Typography variant="subtitle1" gutterBottom>
          {t('applications:onboarding.configure.SignInOptions.subtitle')}
        </Typography>
      </Stack>

      {/* Validation warning if no options selected */}
      {!hasAtLeastOneSelectedOption && !isUsingCustomFlow && (
        <Alert severity="warning" sx={{mb: 2}}>
          {t('applications:onboarding.configure.SignInOptions.noSelectionWarning')}
        </Alert>
      )}

      {/* Toggle options - disabled when using custom flow */}
      <Box sx={{opacity: isUsingCustomFlow ? 0.5 : 1, pointerEvents: isUsingCustomFlow ? 'none' : 'auto'}}>
      <List sx={{bgcolor: 'background.paper', borderRadius: 1, border: 1, borderColor: 'divider'}}>
        {/* Username & Password Option - Disabled if basic flow unavailable */}
        <ListItem
          disablePadding
          secondaryAction={
            <Switch
              edge="end"
              checked={hasUsernamePassword}
              onChange={(): void => onIntegrationToggle(AuthenticatorTypes.BASIC_AUTH)}
              color="primary"
              disabled={isBasicAuthDisabled}
            />
          }
        >
          <ListItemButton
            onClick={(): void => onIntegrationToggle(AuthenticatorTypes.BASIC_AUTH)}
            disabled={isBasicAuthDisabled}
          >
            <ListItemIcon>
              <UserRound size={24} />
            </ListItemIcon>
            <ListItemText
              primary={t('applications:onboarding.configure.SignInOptions.usernamePassword')}
              secondary={isBasicAuthDisabled 
                ? t('applications:onboarding.configure.SignInOptions.flowNotAvailable') 
                : undefined}
            />
          </ListItemButton>
        </ListItem>

        <Divider component="li" />

        {/* Google Option - Always shown if provider exists, disabled if flow unavailable */}
        {googleProvider ? (
          <ListItem
            disablePadding
            secondaryAction={
              !isGoogleDisabled ? (
                <Switch
                  edge="end"
                  checked={integrations[googleProvider.id] ?? false}
                  onChange={(): void => onIntegrationToggle(googleProvider.id)}
                  color="primary"
                />
              ) : null
            }
          >
            <ListItemButton
              onClick={!isGoogleDisabled ? (): void => onIntegrationToggle(googleProvider.id) : undefined}
              disabled={isGoogleDisabled}
            >
              <ListItemIcon>
                <Google size={24} />
              </ListItemIcon>
              <ListItemText
                primary={t('applications:onboarding.configure.SignInOptions.google')}
                secondary={
                  isGoogleDisabled
                    ? t('applications:onboarding.configure.SignInOptions.flowNotAvailable')
                    : null
                }
              />
            </ListItemButton>
          </ListItem>
        ) : (
          <ListItem disablePadding>
            <ListItemButton disabled>
              <ListItemIcon>
                <Google size={24} />
              </ListItemIcon>
              <ListItemText
                primary={t('applications:onboarding.configure.SignInOptions.google')}
                secondary={t('applications:onboarding.configure.SignInOptions.notConfigured')}
              />
            </ListItemButton>
          </ListItem>
        )}
        <Divider component="li" />

        {/* GitHub Option - Always shown if provider exists, disabled if flow unavailable */}
        {githubProvider ? (
          <ListItem
            disablePadding
            secondaryAction={
              !isGitHubDisabled ? (
                <Switch
                  edge="end"
                  checked={integrations[githubProvider.id] ?? false}
                  onChange={(): void => onIntegrationToggle(githubProvider.id)}
                  color="primary"
                />
              ) : null
            }
          >
            <ListItemButton
              onClick={!isGitHubDisabled ? (): void => onIntegrationToggle(githubProvider.id) : undefined}
              disabled={isGitHubDisabled}
            >
              <ListItemIcon>
                <GitHub size={24} />
              </ListItemIcon>
              <ListItemText
                primary={t('applications:onboarding.configure.SignInOptions.github')}
                secondary={
                  isGitHubDisabled
                    ? t('applications:onboarding.configure.SignInOptions.flowNotAvailable')
                    : null
                }
              />
            </ListItemButton>
          </ListItem>
        ) : (
          <ListItem disablePadding>
            <ListItemButton disabled>
              <ListItemIcon>
                <GitHub size={24} />
              </ListItemIcon>
              <ListItemText
                primary={t('applications:onboarding.configure.SignInOptions.github')}
                secondary={t('applications:onboarding.configure.SignInOptions.notConfigured')}
              />
            </ListItemButton>
          </ListItem>
        )}

        {/* Other Social Login Providers (if any) */}
        {availableIntegrations
          .filter(
            (provider: IdentityProvider): boolean =>
              provider.type !== IdentityProviderTypes.GOOGLE && provider.type !== IdentityProviderTypes.GITHUB,
          )
          .map(
            (provider: IdentityProvider, index: number, filteredProviders: IdentityProvider[]): JSX.Element => (
              <>
                <ListItem
                  key={provider.id}
                  disablePadding
                  secondaryAction={
                    <Switch
                      edge="end"
                      checked={integrations[provider.id] ?? false}
                      onChange={(): void => onIntegrationToggle(provider.id)}
                      color="primary"
                    />
                  }
                >
                  <ListItemButton onClick={(): void => onIntegrationToggle(provider.id)}>
                    <ListItemIcon>{getIntegrationIcon(provider.type)}</ListItemIcon>
                    <ListItemText primary={provider.name} />
                  </ListItemButton>
                </ListItem>
                {index < filteredProviders.length - 1 && <Divider component="li" />}
              </>
            ),
          )}
      </List>
      </Box>

      {/* Custom Flow Selection Section - Only show when flows are available */}
      {authFlows.filter((flow: BasicFlowDefinition) => flow.handle !== SYSTEM_FLOW_HANDLES.DEVELOP_APP).length > 0 && (
        <>
          <Divider sx={{my: 2}}>
            <Typography variant="caption" color="text.secondary">
              {t('applications:onboarding.configure.SignInOptions.orDivider')}
            </Typography>
          </Divider>

          <Autocomplete
            id="custom-flow-select"
            size="small"
            options={authFlows.filter((flow: BasicFlowDefinition) => flow.handle !== SYSTEM_FLOW_HANDLES.DEVELOP_APP)}
            getOptionLabel={(option): string => {
              if (typeof option === 'string') {
                const flow = authFlows.find((f) => f.id === option);
                return flow?.name ?? flow?.handle ?? '';
              }
              return option.name ?? option.handle;
            }}
            value={authFlows.find((f) => f.id === customFlowId) ?? null}
            onChange={(_event, newValue): void => {
              if (onCustomFlowChange) {
                onCustomFlowChange(newValue ? newValue.id : null);
              }
            }}
            isOptionEqualToValue={(option, value): boolean => option.id === value.id}
            renderInput={(params): JSX.Element => (
              <TextField
                {...params}
                label={t('applications:onboarding.configure.SignInOptions.selectExistingFlow')}
              />
            )}
            renderOption={(props, option): JSX.Element => (
              <li {...props} key={option.id}>
                <Box>
                  <Typography variant="body1">{option.name ?? option.handle}</Typography>
                  <Typography variant="caption" color="text.secondary">
                    {option.handle}
                  </Typography>
                </Box>
              </li>
            )}
            filterOptions={(options, state): BasicFlowDefinition[] => {
              const inputValue = state.inputValue.toLowerCase();
              if (!inputValue) {
                return options;
              }
              return options.filter(
                (option) =>
                  option.name?.toLowerCase().includes(inputValue) ||
                  option.handle.toLowerCase().includes(inputValue),
              );
            }}
            ListboxProps={{
              sx: {
                maxHeight: 300,
              },
            }}
            noOptionsText={t('applications:onboarding.configure.SignInOptions.noFlowsFound')}
            clearText={t('applications:onboarding.configure.SignInOptions.clear')}
          />
        </>
      )}

      <Stack direction="row" alignItems="center" spacing={1}>
        <Lightbulb size={20} color={theme?.vars?.palette.warning.main} />
        <Typography variant="body2" color="text.secondary">
          {t('applications:onboarding.configure.SignInOptions.hint')}
        </Typography>
      </Stack>
    </Stack>
  );
}
