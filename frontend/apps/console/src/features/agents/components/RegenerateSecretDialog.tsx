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

import {useLogger} from '@thunderid/logger';
import {Dialog, DialogTitle, DialogContent, DialogContentText, DialogActions, Button, Alert} from '@wso2/oxygen-ui';
import {useState, type JSX} from 'react';
import {useTranslation} from 'react-i18next';
import useRegenerateAgentSecret from '../api/useRegenerateAgentSecret';

export interface RegenerateSecretDialogProps {
  open: boolean;
  agentId: string | null;
  onClose: () => void;
  onSuccess?: (newClientSecret: string) => void;
  onError?: (message: string) => void;
}

export default function RegenerateSecretDialog({
  open,
  agentId,
  onClose,
  onSuccess = undefined,
  onError = undefined,
}: RegenerateSecretDialogProps): JSX.Element {
  const {t} = useTranslation();
  const logger = useLogger('RegenerateSecretDialog');
  const [error, setError] = useState<string | null>(null);
  const regenerateClientSecret = useRegenerateAgentSecret();

  const handleCancel = (): void => {
    setError(null);
    onClose();
  };

  const handleConfirm = (): void => {
    if (!agentId) {
      setError(t('agents:regenerateSecret.dialog.error', 'Failed to regenerate client secret'));
      return;
    }

    setError(null);
    logger.info('Regenerating agent client secret', {agentId});

    regenerateClientSecret.mutate(
      {agentId},
      {
        onSuccess: ({clientSecret}) => {
          logger.info('Agent client secret regenerated successfully.', {agentId});
          onClose();
          onSuccess?.(clientSecret);
        },
        onError: (err) => {
          const errorMessage =
            err instanceof Error
              ? err.message
              : t('agents:regenerateSecret.dialog.error', 'Failed to regenerate client secret');
          logger.error('Failed to regenerate agent client secret', {
            agentId,
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
      <DialogTitle>{t('agents:regenerateSecret.dialog.title', 'Regenerate client secret?')}</DialogTitle>
      <DialogContent>
        <DialogContentText sx={{mb: 2}}>
          {t(
            'agents:regenerateSecret.dialog.message',
            'A new client secret will be generated for this agent. Any service using the current client secret will stop working immediately.',
          )}
        </DialogContentText>
        <Alert severity="warning" sx={{mb: 2}}>
          {t(
            'agents:regenerateSecret.dialog.disclaimer',
            'This action cannot be undone. The current client secret will be invalidated as soon as you confirm.',
          )}
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
          disabled={regenerateClientSecret.isPending || !agentId}
        >
          {regenerateClientSecret.isPending
            ? t('agents:regenerateSecret.dialog.regenerating', 'Regenerating…')
            : t('agents:regenerateSecret.dialog.confirmButton', 'Regenerate')}
        </Button>
      </DialogActions>
    </Dialog>
  );
}
