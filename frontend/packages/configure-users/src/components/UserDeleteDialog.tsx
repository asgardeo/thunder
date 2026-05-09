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
import useDeleteUser from '../api/useDeleteUser';

export interface UserDeleteDialogProps {
  open: boolean;
  userId: string | null;
  onClose: () => void;
  onSuccess?: () => void;
}

/**
 * Dialog component for confirming user deletion.
 */
export default function UserDeleteDialog({
  open,
  userId,
  onClose,
  onSuccess = undefined,
}: UserDeleteDialogProps): JSX.Element {
  const {t} = useTranslation();
  const deleteUser = useDeleteUser();
  const [error, setError] = useState<string | null>(null);

  const handleCancel = (): void => {
    if (deleteUser.isPending) return;
    setError(null);
    onClose();
  };

  const handleConfirm = (): void => {
    if (!userId) return;

    setError(null);
    deleteUser.mutate(userId, {
      onSuccess: (): void => {
        setError(null);
        onClose();
        onSuccess?.();
      },
      onError: (err: Error) => {
        setError(err.message ?? t('users:delete.error', 'Failed to delete user'));
      },
    });
  };

  return (
    <Dialog open={open} onClose={handleCancel} maxWidth="sm" fullWidth>
      <DialogTitle>{t('users:delete.title', 'Delete User')}</DialogTitle>
      <DialogContent>
        <DialogContentText sx={{mb: 2}}>
          {t('users:delete.message', 'Are you sure you want to delete this user? This action cannot be undone.')}
        </DialogContentText>
        <Alert severity="warning" sx={{mb: 2}}>
          {t('users:delete.disclaimer', 'All associated data will be permanently removed.')}
        </Alert>
        {error && (
          <Alert severity="error" sx={{mt: 2}}>
            {error}
          </Alert>
        )}
      </DialogContent>
      <DialogActions>
        <Button onClick={handleCancel} disabled={deleteUser.isPending}>
          {t('common:actions.cancel')}
        </Button>
        <Button onClick={handleConfirm} color="error" variant="contained" disabled={deleteUser.isPending || !userId}>
          {deleteUser.isPending ? t('common:status.deleting', 'Deleting...') : t('common:actions.delete', 'Delete')}
        </Button>
      </DialogActions>
    </Dialog>
  );
}
