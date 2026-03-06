/**
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
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

import {useState, useEffect, type JSX} from 'react';
import {Dialog, DialogTitle, DialogContent, DialogContentText, DialogActions, Button, Alert} from '@wso2/oxygen-ui';
import {useTranslation} from 'react-i18next';
import {useLogger} from '@thunder/logger';
import useRegenerateClientSecret from '../api/useRegenerateClientSecret';

/**
 * Props for the {@link RegenerateSecretDialog} component.
 */
export interface RegenerateSecretDialogProps {
  /**
   * Whether the dialog is open
   */
  open: boolean;
  /**
   * The ID of the application whose client secret will be regenerated
   */
  applicationId: string | null;
  /**
   * Callback when the dialog should be closed
   */
  onClose: () => void;
  /**
   * Callback when the client secret is successfully regenerated with the new client secret
   */
  onSuccess?: (newClientSecret: string) => void;
  /**
   * Callback when the regeneration fails
   */
  onError?: (message: string) => void;
}

/**
 * Dialog component for confirming client secret regeneration.
 *
 * This dialog warns users about the consequences of regenerating an application's
 * client secret before proceeding with the action.
 *
 * @param props - Component props
 * @returns The regenerate client secret confirmation dialog
 */
export default function RegenerateSecretDialog({
  open,
  applicationId,
  onClose,
  onSuccess = undefined,
  onError = undefined,
}: RegenerateSecretDialogProps): JSX.Element {
  const {t} = useTranslation();
  const logger = useLogger('RegenerateSecretDialog');
  const [error, setError] = useState<string | null>(null);
  const regenerateClientSecret = useRegenerateClientSecret();

  // Reset error state when dialog is opened
  useEffect(() => {
    if (open) {
      setError(null);
    }
  }, [open]);

  const handleCancel = (): void => {
    setError(null);
    onClose();
  };

  const handleConfirm = (): void => {
    if (!applicationId) {
      setError(t('applications:regenerateSecret.dialog.error'));
      return;
    }

    setError(null);
    logger.info('Regenerating application client secret', {applicationId});

    regenerateClientSecret.mutate(
      {applicationId},
      {
        onSuccess: ({clientSecret}) => {
          logger.info('Application client secret regenerated successfully. New client secret generated.', {
            applicationId,
          });
          onClose();
          onSuccess?.(clientSecret);
        },
        onError: (err) => {
          const errorMessage = err instanceof Error ? err.message : t('applications:regenerateSecret.dialog.error');
          logger.error('Failed to regenerate client secret', {
            applicationId,
            errorMessage,
            errorName: err instanceof Error ? err.name : 'UnknownError',
          });
          setError(errorMessage);
          onError?.(errorMessage);
        },
      },
    );
  };

  return (
    <Dialog open={open} onClose={handleCancel} maxWidth="sm" fullWidth>
      <DialogTitle>{t('applications:regenerateSecret.dialog.title')}</DialogTitle>
      <DialogContent>
        <DialogContentText sx={{mb: 2}}>{t('applications:regenerateSecret.dialog.message')}</DialogContentText>
        <Alert severity="warning" sx={{mb: 2}}>
          {t('applications:regenerateSecret.dialog.disclaimer')}
        </Alert>
        {error && (
          <Alert severity="error" sx={{mt: 2}}>
            {error}
          </Alert>
        )}
      </DialogContent>
      <DialogActions>
        <Button onClick={handleCancel} disabled={regenerateClientSecret.isPending}>
          {t('common:actions.cancel')}
        </Button>
        <Button
          onClick={handleConfirm}
          color="error"
          variant="contained"
          disabled={regenerateClientSecret.isPending || !applicationId}
        >
          {regenerateClientSecret.isPending
            ? t('applications:regenerateSecret.dialog.regenerating')
            : t('applications:regenerateSecret.dialog.confirmButton')}
        </Button>
      </DialogActions>
    </Dialog>
  );
}
