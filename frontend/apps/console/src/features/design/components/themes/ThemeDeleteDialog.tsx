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

import {useDeleteTheme} from '@thunderid/design';
import {Alert, Button, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle} from '@wso2/oxygen-ui';
import {useState, type JSX} from 'react';
import {useTranslation} from 'react-i18next';

export interface ThemeDeleteDialogProps {
  open: boolean;
  themeId: string | null;
  themeName: string | null;
  onClose: () => void;
  onSuccess?: () => void;
}

export default function ThemeDeleteDialog({
  open,
  themeId,
  themeName,
  onClose,
  onSuccess = undefined,
}: ThemeDeleteDialogProps): JSX.Element {
  const {t} = useTranslation('design');
  const deleteTheme = useDeleteTheme();
  const [error, setError] = useState<string | null>(null);

  const handleCancel = (): void => {
    if (deleteTheme.isPending) return;
    setError(null);
    onClose();
  };

  const handleConfirm = (): void => {
    if (!themeId) return;

    setError(null);
    deleteTheme.mutate(themeId, {
      onSuccess: (): void => {
        setError(null);
        onClose();
        onSuccess?.();
      },
      onError: (err: Error) => {
        setError(err.message ?? t('themes.delete.error'));
      },
    });
  };

  const message = themeName ? t('themes.delete.message', {name: themeName}) : t('themes.delete.messageUnnamed');

  return (
    <Dialog open={open} onClose={handleCancel} maxWidth="sm" fullWidth>
      <DialogTitle>{t('themes.delete.title')}</DialogTitle>
      <DialogContent>
        <DialogContentText sx={{mb: 2}}>{message}</DialogContentText>
        <Alert severity="warning" sx={{mb: 2}}>
          {t('themes.delete.disclaimer')}
        </Alert>
        {error && (
          <Alert severity="error" sx={{mt: 2}}>
            {error}
          </Alert>
        )}
      </DialogContent>
      <DialogActions>
        <Button onClick={handleCancel} disabled={deleteTheme.isPending}>
          {t('common:actions.cancel')}
        </Button>
        <Button onClick={handleConfirm} color="error" variant="contained" disabled={deleteTheme.isPending || !themeId}>
          {deleteTheme.isPending ? t('common:status.deleting') : t('common:actions.delete')}
        </Button>
      </DialogActions>
    </Dialog>
  );
}
