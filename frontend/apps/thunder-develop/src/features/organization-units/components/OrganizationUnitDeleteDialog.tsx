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
import useDeleteOrganizationUnit from '../api/useDeleteOrganizationUnit';

export interface OrganizationUnitDeleteDialogProps {
  /**
   * Whether the dialog is open
   */
  open: boolean;
  /**
   * The ID of the organization unit to delete
   */
  organizationUnitId: string | null;
  /**
   * Callback when the dialog should be closed
   */
  onClose: () => void;
  /**
   * Callback when the organization unit is successfully deleted
   */
  onSuccess?: () => void;
}

/**
 * Dialog component for confirming organization unit deletion
 */
export default function OrganizationUnitDeleteDialog({
  open,
  organizationUnitId,
  onClose,
  onSuccess = undefined,
}: OrganizationUnitDeleteDialogProps): JSX.Element {
  const {t} = useTranslation();
  const deleteOrganizationUnit = useDeleteOrganizationUnit();
  const [error, setError] = useState<string | null>(null);

  const handleCancel = (): void => {
    setError(null);
    onClose();
  };

  const handleConfirm = (): void => {
    if (!organizationUnitId) return;

    deleteOrganizationUnit.mutate(organizationUnitId, {
      onSuccess: (): void => {
        setError(null);
        onClose();
        onSuccess?.();
      },
      onError: (err: Error) => {
        setError(err.message ?? t('organizationUnits:delete.error'));
      },
    });
  };

  return (
    <Dialog open={open} onClose={handleCancel} maxWidth="sm" fullWidth>
      <DialogTitle>{t('organizationUnits:delete.title')}</DialogTitle>
      <DialogContent>
        <DialogContentText sx={{mb: 2}}>{t('organizationUnits:delete.message')}</DialogContentText>
        <Alert severity="warning" sx={{mb: 2}}>
          {t('organizationUnits:delete.disclaimer')}
        </Alert>
        {error && (
          <Alert severity="error" sx={{mt: 2}}>
            {error}
          </Alert>
        )}
      </DialogContent>
      <DialogActions>
        <Button onClick={handleCancel} disabled={deleteOrganizationUnit.isPending}>
          {t('common:actions.cancel')}
        </Button>
        <Button onClick={handleConfirm} color="error" variant="contained" disabled={deleteOrganizationUnit.isPending}>
          {deleteOrganizationUnit.isPending ? t('common:status.deleting') : t('common:actions.delete')}
        </Button>
      </DialogActions>
    </Dialog>
  );
}
