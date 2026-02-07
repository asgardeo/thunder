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
import {useState} from 'react';
import {Button, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle} from '@mui/material';

/**
 * Severity level for the confirmation dialog.
 */
export type ConfirmDialogSeverity = 'warning' | 'error' | 'info';

/**
 * Props for the ConfirmDialog component.
 */
export interface ConfirmDialogProps {
  /** Whether the dialog is open */
  open: boolean;
  /** Callback when dialog is closed (via cancel or backdrop) */
  onClose: () => void;
  /** Callback when user confirms the action */
  onConfirm: () => void | Promise<void>;
  /** Dialog title */
  title: string;
  /** Dialog message/description */
  message: string;
  /** Confirm button label */
  confirmLabel?: string;
  /** Cancel button label */
  cancelLabel?: string;
  /** Severity level affects button color */
  severity?: ConfirmDialogSeverity;
}

/**
 * Reusable confirmation dialog with focus trap and keyboard support.
 * Automatically handles loading state during async confirm operations.
 *
 * @param props - Component props
 * @returns ConfirmDialog component
 *
 * @example
 * ```tsx
 * const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
 *
 * const handleDelete = async () => {
 *   await deleteTheme(themeId);
 *   setDeleteDialogOpen(false);
 * };
 *
 * return (
 *   <>
 *     <Button onClick={() => setDeleteDialogOpen(true)}>Delete</Button>
 *     <ConfirmDialog
 *       open={deleteDialogOpen}
 *       onClose={() => setDeleteDialogOpen(false)}
 *       onConfirm={handleDelete}
 *       title="Delete Theme"
 *       message="Are you sure you want to delete this theme? This action cannot be undone."
 *       confirmLabel="Delete"
 *       severity="error"
 *     />
 *   </>
 * );
 * ```
 */
export default function ConfirmDialog({
  open,
  onClose,
  onConfirm,
  title,
  message,
  confirmLabel = 'Confirm',
  cancelLabel = 'Cancel',
  severity = 'warning',
}: ConfirmDialogProps): JSX.Element {
  const [isLoading, setIsLoading] = useState(false);

  const handleConfirm = async (): Promise<void> => {
    setIsLoading(true);
    try {
      await onConfirm();
    } finally {
      setIsLoading(false);
    }
  };

  const confirmButtonColor = severity === 'error' ? 'error' : severity === 'warning' ? 'warning' : 'primary';

  return (
    <Dialog
      open={open}
      onClose={onClose}
      aria-labelledby="confirm-dialog-title"
      aria-describedby="confirm-dialog-description"
      maxWidth="sm"
      fullWidth
    >
      <DialogTitle id="confirm-dialog-title">{title}</DialogTitle>
      <DialogContent>
        <DialogContentText id="confirm-dialog-description">{message}</DialogContentText>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose} disabled={isLoading}>
          {cancelLabel}
        </Button>
        <Button onClick={handleConfirm} color={confirmButtonColor} variant="contained" disabled={isLoading} autoFocus>
          {confirmLabel}
        </Button>
      </DialogActions>
    </Dialog>
  );
}
