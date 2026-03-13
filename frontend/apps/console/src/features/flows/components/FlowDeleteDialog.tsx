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

import {useState, type JSX} from 'react';
import {Dialog, DialogTitle, DialogContent, DialogContentText, DialogActions, Button, Alert} from '@wso2/oxygen-ui';
import {useTranslation} from 'react-i18next';
import useDeleteFlow from '../api/useDeleteFlow';

export interface FlowDeleteDialogProps {
  /**
   * Whether the dialog is open
   */
  open: boolean;
  /**
   * The ID of the flow to delete
   */
  flowId: string | null;
  /**
   * Callback when the dialog should be closed
   */
  onClose: () => void;
  /**
   * Callback when the flow is successfully deleted
   */
  onSuccess?: () => void;
}

/**
 * Dialog component for confirming flow deletion
 */
export default function FlowDeleteDialog({
  open,
  flowId,
  onClose,
  onSuccess = undefined,
}: FlowDeleteDialogProps): JSX.Element {
  const {t} = useTranslation();
  const deleteFlow = useDeleteFlow();
  const [error, setError] = useState<string | null>(null);

  const handleCancel = (): void => {
    setError(null);
    onClose();
  };

  const handleConfirm = (): void => {
    if (!flowId) return;

    deleteFlow.mutate(flowId, {
      onSuccess: (): void => {
        setError(null);
        onClose();
        onSuccess?.();
      },
      onError: (err: Error) => {
        setError(err.message ?? t('flows:delete.error'));
      },
    });
  };

  return (
    <Dialog open={open} onClose={handleCancel} maxWidth="sm" fullWidth>
      <DialogTitle>{t('flows:delete.title')}</DialogTitle>
      <DialogContent>
        <DialogContentText sx={{mb: 2}}>{t('flows:delete.message')}</DialogContentText>
        <Alert severity="warning" sx={{mb: 2}}>
          {t('flows:delete.disclaimer')}
        </Alert>
        {error && (
          <Alert severity="error" sx={{mt: 2}}>
            {error}
          </Alert>
        )}
      </DialogContent>
      <DialogActions>
        <Button onClick={handleCancel} disabled={deleteFlow.isPending}>
          {t('common:actions.cancel')}
        </Button>
        <Button onClick={handleConfirm} color="error" variant="contained" disabled={deleteFlow.isPending}>
          {deleteFlow.isPending ? t('common:status.deleting') : t('common:actions.delete')}
        </Button>
      </DialogActions>
    </Dialog>
  );
}
