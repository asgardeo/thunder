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

import type {CommonResourcePropertiesPropsInterface} from '@/features/flows/components/resource-property-panel/ResourceProperties';
import {useMemo, type ReactNode} from 'react';
import {useTranslation} from 'react-i18next';
import {Alert, FormHelperText, FormLabel, MenuItem, Select, Stack, TextField, Typography} from '@wso2/oxygen-ui';
import type {StepData} from '@/features/flows/models/steps';
import {ExecutionTypes} from '@/features/flows/models/steps';
import useValidationStatus from '@/features/flows/hooks/useValidationStatus';
import useIdentityProviders from '@/features/integrations/api/useIdentityProviders';
import {IdentityProviderTypes, type IdentityProviderType} from '@/features/integrations/models/identity-provider';
import useNotificationSenders from '@/features/notification-senders/api/useNotificationSenders';

/**
 * Maps executor names to their corresponding identity provider types.
 */
const EXECUTOR_TO_IDP_TYPE_MAP: Record<string, IdentityProviderType> = {
  [ExecutionTypes.GoogleFederation]: IdentityProviderTypes.GOOGLE,
  [ExecutionTypes.GithubFederation]: IdentityProviderTypes.GITHUB,
};

/**
 * Available modes for SMS OTP executor.
 */
const SMS_OTP_MODES = [
  {value: 'send', translationKey: 'flows:core.executions.smsOtp.mode.send', displayLabel: 'Send SMS OTP'},
  {value: 'verify', translationKey: 'flows:core.executions.smsOtp.mode.verify', displayLabel: 'Verify SMS OTP'},
] as const;

/**
 * Available modes for Passkey executor.
 */
const PASSKEY_MODES = [
  {
    value: 'challenge',
    translationKey: 'flows:core.executions.passkey.mode.challenge',
    displayLabel: 'Request Passkey',
  },
  {value: 'verify', translationKey: 'flows:core.executions.passkey.mode.verify', displayLabel: 'Verify Passkey'},
  {
    value: 'register_start',
    translationKey: 'flows:core.executions.passkey.mode.registerStart',
    displayLabel: 'Start Passkey Registration',
  },
  {
    value: 'register_finish',
    translationKey: 'flows:core.executions.passkey.mode.registerFinish',
    displayLabel: 'Finish Passkey Registration',
  },
] as const;

/**
 * Passkey modes that require relying party configuration.
 */
const PASSKEY_MODES_WITH_RELYING_PARTY = ['challenge', 'register_start'] as const;

/**
 * Props interface of {@link ExecutionExtendedProperties}
 */
export type ExecutionExtendedPropertiesPropsInterface = CommonResourcePropertiesPropsInterface;

/**
 * Extended properties for execution step elements (Google, GitHub, etc.).
 * Provides a dropdown to select the IDP connection for the executor.
 *
 * @param props - Props injected to the component.
 * @returns The ExecutionExtendedProperties component.
 */
function ExecutionExtendedProperties({resource, onChange}: ExecutionExtendedPropertiesPropsInterface): ReactNode {
  const {t} = useTranslation();
  const {selectedNotification} = useValidationStatus();
  const {data: identityProviders, isLoading: isLoadingIdps} = useIdentityProviders();
  const {data: notificationSenders, isLoading: isLoadingSenders} = useNotificationSenders();

  // Get the executor name from the resource
  const executorName = useMemo(() => {
    const stepData = resource?.data as StepData | undefined;
    return stepData?.action?.executor?.name;
  }, [resource]);

  // Get the current IDP ID from the resource properties
  const currentIdpId = useMemo(() => {
    const stepData = resource?.data as StepData | undefined;
    return (stepData?.properties as {idpId?: string})?.idpId ?? '';
  }, [resource]);

  // Get the current mode for SMS OTP executor
  const currentMode = useMemo(() => {
    const stepData = resource?.data as StepData | undefined;
    return (stepData?.action?.executor as {mode?: string})?.mode ?? '';
  }, [resource]);

  // Get the current sender ID for SMS OTP executor
  const currentSenderId = useMemo(() => {
    const stepData = resource?.data as StepData | undefined;
    return (stepData?.properties as {senderId?: string})?.senderId ?? '';
  }, [resource]);

  // Get the current relying party ID for Passkey executor
  const currentRelyingPartyId = useMemo(() => {
    const stepData = resource?.data as StepData | undefined;
    return (stepData?.properties as {relyingPartyId?: string})?.relyingPartyId ?? '';
  }, [resource]);

  // Get the current relying party name for Passkey executor
  const currentRelyingPartyName = useMemo(() => {
    const stepData = resource?.data as StepData | undefined;
    return (stepData?.properties as {relyingPartyName?: string})?.relyingPartyName ?? '';
  }, [resource]);

  // Check if this is an SMS OTP executor
  const isSmsOtpExecutor = executorName === ExecutionTypes.SMSOTPAuth;

  // Check if this is a Passkey executor
  const isPasskeyExecutor = executorName === ExecutionTypes.PasskeyAuth;

  // Check if current Passkey mode requires relying party configuration
  const showRelyingPartyConfig =
    isPasskeyExecutor &&
    PASSKEY_MODES_WITH_RELYING_PARTY.includes(currentMode as (typeof PASSKEY_MODES_WITH_RELYING_PARTY)[number]);

  // Get the IDP type for the current executor
  const idpType = useMemo(() => {
    if (!executorName) {
      return null;
    }
    return EXECUTOR_TO_IDP_TYPE_MAP[executorName] ?? null;
  }, [executorName]);

  // Get available connections for this executor by filtering IDPs by type
  const availableConnections = useMemo(() => {
    if (!idpType || !identityProviders) {
      return [];
    }

    return identityProviders.filter((idp) => idp.type === idpType);
  }, [idpType, identityProviders]);

  // Check if current value is a placeholder or empty
  const isPlaceholder = currentIdpId === '{{IDP_ID}}' || currentIdpId === '';

  /**
   * Get the error message for the connection field.
   */
  const errorMessage: string = useMemo(() => {
    const key = `${resource?.id}_data.properties.idpId`;

    if (selectedNotification?.hasResourceFieldNotification(key)) {
      return selectedNotification?.getResourceFieldNotification(key);
    }

    return '';
  }, [resource, selectedNotification]);

  /**
   * Get the error message for the sender field.
   */
  const senderErrorMessage: string = useMemo(() => {
    const key = `${resource?.id}_data.properties.senderId`;

    if (selectedNotification?.hasResourceFieldNotification(key)) {
      return selectedNotification?.getResourceFieldNotification(key);
    }

    return '';
  }, [resource, selectedNotification]);

  // Handle connection selection - store the IDP ID in properties.idpId
  const handleConnectionChange = (selectedIdpId: string): void => {
    onChange('data.properties.idpId', selectedIdpId, resource);
  };

  // Handle mode selection for SMS OTP executor
  const handleModeChange = (selectedMode: string): void => {
    // Update the display label based on the selected mode
    const modeConfig = SMS_OTP_MODES.find((mode) => mode.value === selectedMode);

    // Build the updated data object with both mode and display label
    // This avoids the debounce issue where multiple rapid onChange calls would drop intermediate values
    const updatedData = {
      ...((resource?.data as StepData) ?? {}),
      action: {
        ...((resource?.data as StepData)?.action ?? {}),
        executor: {
          ...((resource?.data as StepData)?.action?.executor ?? {}),
          mode: selectedMode,
        },
      },
      display: {
        ...((resource?.data as StepData)?.display ?? {}),
        label: modeConfig?.displayLabel ?? 'SMS OTP',
      },
    };

    // Single update call with the complete data object
    onChange('data', updatedData, resource);
  };

  // Handle sender selection for SMS OTP executor
  const handleSenderChange = (selectedSenderId: string): void => {
    onChange('data.properties.senderId', selectedSenderId, resource);
  };

  // Handle mode selection for Passkey executor
  const handlePasskeyModeChange = (selectedMode: string): void => {
    // Update the display label based on the selected mode
    const modeConfig = PASSKEY_MODES.find((mode) => mode.value === selectedMode);

    // Build the updated data object with both mode and display label
    const updatedData = {
      ...((resource?.data as StepData) ?? {}),
      action: {
        ...((resource?.data as StepData)?.action ?? {}),
        executor: {
          ...((resource?.data as StepData)?.action?.executor ?? {}),
          mode: selectedMode,
        },
      },
      display: {
        ...((resource?.data as StepData)?.display ?? {}),
        label: modeConfig?.displayLabel ?? 'Passkey',
      },
    };

    // Single update call with the complete data object
    onChange('data', updatedData, resource);
  };

  // Handle relying party ID change for Passkey executor
  const handleRelyingPartyIdChange = (value: string): void => {
    onChange('data.properties.relyingPartyId', value, resource);
  };

  // Handle relying party name change for Passkey executor
  const handleRelyingPartyNameChange = (value: string): void => {
    onChange('data.properties.relyingPartyName', value, resource);
  };

  // Render SMS OTP mode and sender selector
  if (isSmsOtpExecutor) {
    const hasSenders = (notificationSenders?.length ?? 0) > 0;
    const isSenderPlaceholder = currentSenderId === '{{SENDER_ID}}' || currentSenderId === '';
    const showSenderError = isSenderPlaceholder || !!senderErrorMessage;

    return (
      <Stack gap={2}>
        <Typography variant="body2" color="text.secondary">
          {t('flows:core.executions.smsOtp.description')}
        </Typography>

        <div>
          <FormLabel htmlFor="mode-select">{t('flows:core.executions.smsOtp.mode.label')}</FormLabel>
          <Select
            id="mode-select"
            value={currentMode}
            onChange={(e) => handleModeChange(e.target.value)}
            displayEmpty
            fullWidth
          >
            <MenuItem value="" disabled>
              {t('flows:core.executions.smsOtp.mode.placeholder')}
            </MenuItem>
            {SMS_OTP_MODES.map((mode) => (
              <MenuItem key={mode.value} value={mode.value}>
                {t(mode.translationKey)}
              </MenuItem>
            ))}
          </Select>
        </div>

        <div>
          <FormLabel htmlFor="sender-select">{t('flows:core.executions.smsOtp.sender.label')}</FormLabel>
          <Select
            id="sender-select"
            value={isSenderPlaceholder ? '' : currentSenderId}
            onChange={(e) => handleSenderChange(e.target.value)}
            displayEmpty
            fullWidth
            error={showSenderError}
            disabled={isLoadingSenders || !hasSenders}
          >
            <MenuItem value="" disabled>
              {isLoadingSenders ? t('common:loading') : t('flows:core.executions.smsOtp.sender.placeholder')}
            </MenuItem>
            {notificationSenders?.map((sender) => (
              <MenuItem key={sender.id} value={sender.id}>
                {sender.name}
              </MenuItem>
            ))}
          </Select>
          {showSenderError && (
            <FormHelperText error>
              {senderErrorMessage || t('flows:core.executions.smsOtp.sender.required')}
            </FormHelperText>
          )}
        </div>

        {!isLoadingSenders && !hasSenders && (
          <Alert severity="warning">{t('flows:core.executions.smsOtp.sender.noSenders')}</Alert>
        )}
      </Stack>
    );
  }

  // Render Passkey mode selector and relying party configuration
  if (isPasskeyExecutor) {
    return (
      <Stack gap={2}>
        <Typography variant="body2" color="text.secondary">
          {t('flows:core.executions.passkey.description')}
        </Typography>

        <div>
          <FormLabel htmlFor="passkey-mode-select">{t('flows:core.executions.passkey.mode.label')}</FormLabel>
          <Select
            id="passkey-mode-select"
            value={currentMode}
            onChange={(e) => handlePasskeyModeChange(e.target.value)}
            displayEmpty
            fullWidth
          >
            <MenuItem value="" disabled>
              {t('flows:core.executions.passkey.mode.placeholder')}
            </MenuItem>
            {PASSKEY_MODES.map((mode) => (
              <MenuItem key={mode.value} value={mode.value}>
                {t(mode.translationKey)}
              </MenuItem>
            ))}
          </Select>
        </div>

        {showRelyingPartyConfig && (
          <>
            <div>
              <FormLabel htmlFor="relying-party-id">
                {t('flows:core.executions.passkey.relyingPartyId.label')}
              </FormLabel>
              <TextField
                id="relying-party-id"
                value={currentRelyingPartyId}
                onChange={(e) => handleRelyingPartyIdChange(e.target.value)}
                placeholder={t('flows:core.executions.passkey.relyingPartyId.placeholder')}
                fullWidth
                size="small"
              />
              <FormHelperText>{t('flows:core.executions.passkey.relyingPartyId.hint')}</FormHelperText>
            </div>

            <div>
              <FormLabel htmlFor="relying-party-name">
                {t('flows:core.executions.passkey.relyingPartyName.label')}
              </FormLabel>
              <TextField
                id="relying-party-name"
                value={currentRelyingPartyName}
                onChange={(e) => handleRelyingPartyNameChange(e.target.value)}
                placeholder={t('flows:core.executions.passkey.relyingPartyName.placeholder')}
                fullWidth
                size="small"
              />
              <FormHelperText>{t('flows:core.executions.passkey.relyingPartyName.hint')}</FormHelperText>
            </div>
          </>
        )}
      </Stack>
    );
  }

  // If no executor name or no matching IDP type, don't render
  if (!executorName || !idpType) {
    return null;
  }

  const hasConnections = availableConnections.length > 0;
  const showError = isPlaceholder || !!errorMessage;

  return (
    <Stack gap={2}>
      <Typography variant="body2" color="text.secondary">
        Select a connection from the following list to link it with the login flow.
      </Typography>

      <div>
        <FormLabel htmlFor="connection-select">Connection</FormLabel>
        <Select
          id="connection-select"
          value={isPlaceholder ? '' : currentIdpId}
          onChange={(e) => handleConnectionChange(e.target.value)}
          displayEmpty
          fullWidth
          error={showError}
          disabled={isLoadingIdps || !hasConnections}
        >
          <MenuItem value="" disabled>
            {isLoadingIdps ? t('common:loading') : 'Select a connection'}
          </MenuItem>
          {availableConnections.map((idp) => (
            <MenuItem key={idp.id} value={idp.id}>
              {idp.name}
            </MenuItem>
          ))}
        </Select>
        {showError && (
          <FormHelperText error>{errorMessage || 'Connection is required and must be selected.'}</FormHelperText>
        )}
      </div>

      {!isLoadingIdps && !hasConnections && (
        <Alert severity="warning">
          No connections available. Please create a connection to link with the login flow.
        </Alert>
      )}
    </Stack>
  );
}

export default ExecutionExtendedProperties;
