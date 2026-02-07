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

import type {JSX} from 'react';
import {Button, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle} from '@wso2/oxygen-ui';
import {useTranslation} from 'react-i18next';
import {useLogger} from '@thunder/logger';
import useDeleteTheme from '../api/useDeleteTheme';

/**
 * Props for ThemeDeleteDialog component
 */
interface ThemeDeleteDialogProps {
  open: boolean;
  themeId: string;
  onClose: () => void;
}

/**
 * Dialog component for confirming theme deletion
 */
export default function ThemeDeleteDialog({open, themeId, onClose}: ThemeDeleteDialogProps): JSX.Element {
  const {t} = useTranslation();
  const logger = useLogger('ThemeDeleteDialog');
  const {mutate: deleteTheme, isPending} = useDeleteTheme();

  const handleDelete = () => {
    deleteTheme(themeId, {
      onSuccess: () => {
        logger.info('Theme deleted successfully', {themeId});
        onClose();
      },
      onError: (error) => {
        logger.error('Failed to delete theme', {error, themeId});
      },
    });
  };

  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle>{t('themes:delete.title', {defaultValue: 'Delete Theme'})}</DialogTitle>
      <DialogContent>
        <DialogContentText>
          {t('themes:delete.message', {
            defaultValue: 'Are you sure you want to delete this theme? This action cannot be undone.',
          })}
        </DialogContentText>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose} disabled={isPending}>
          {t('common:actions.cancel', {defaultValue: 'Cancel'})}
        </Button>
        <Button onClick={handleDelete} color="error" variant="contained" disabled={isPending}>
          {t('common:actions.delete', {defaultValue: 'Delete'})}
        </Button>
      </DialogActions>
    </Dialog>
  );
}
