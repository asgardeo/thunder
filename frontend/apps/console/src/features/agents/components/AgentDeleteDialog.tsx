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

import {Dialog, DialogTitle, DialogContent, DialogContentText, DialogActions, Button, Alert} from '@wso2/oxygen-ui';
import {useState, type JSX} from 'react';
import {useTranslation} from 'react-i18next';
import useDeleteAgent from '../api/useDeleteAgent';

export interface AgentDeleteDialogProps {
  open: boolean;
  agentId: string | null;
  onClose: () => void;
  onSuccess?: () => void;
}

export default function AgentDeleteDialog({
  open,
  agentId,
  onClose,
  onSuccess = undefined,
}: AgentDeleteDialogProps): JSX.Element {
  const {t} = useTranslation();
  const deleteAgent = useDeleteAgent();
  const [error, setError] = useState<string | null>(null);

  const handleCancel = (): void => {
    if (deleteAgent.isPending) return;
    setError(null);
    onClose();
  };

  const handleConfirm = (): void => {
    if (!agentId) return;

    deleteAgent.mutate(agentId, {
      onSuccess: (): void => {
        setError(null);
        onClose();
        onSuccess?.();
      },
      onError: (err: Error) => {
        setError(err.message ?? t('agents:delete.error', 'Failed to delete agent. Please try again.'));
      },
    });
  };

  return (
    <Dialog open={open} onClose={handleCancel} maxWidth="sm" fullWidth>
      <DialogTitle>{t('agents:delete.title', 'Delete agent')}</DialogTitle>
      <DialogContent>
        <DialogContentText sx={{mb: 2}}>
          {t('agents:delete.message', 'Are you sure you want to delete this agent? This action cannot be undone.')}
        </DialogContentText>
        <Alert severity="warning" sx={{mb: 2}}>
          {t('agents:delete.disclaimer', 'Deleting this agent will revoke all its credentials and access tokens.')}
        </Alert>
        {error && (
          <Alert severity="error" sx={{mt: 2}}>
            {error}
          </Alert>
        )}
      </DialogContent>
      <DialogActions>
        <Button onClick={handleCancel} disabled={deleteAgent.isPending}>
          {t('common:actions.cancel')}
        </Button>
        <Button onClick={handleConfirm} color="error" variant="contained" disabled={deleteAgent.isPending}>
          {deleteAgent.isPending ? t('common:status.deleting') : t('common:actions.delete')}
        </Button>
      </DialogActions>
    </Dialog>
  );
}
