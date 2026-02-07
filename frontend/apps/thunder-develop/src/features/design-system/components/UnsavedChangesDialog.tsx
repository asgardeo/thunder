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
 * Props for the UnsavedChangesDialog component.
 */
export interface UnsavedChangesDialogProps {
  /** Whether the dialog is open */
  open: boolean;
  /** Callback to save changes and proceed with navigation */
  onSave: () => void | Promise<void>;
  /** Callback to discard changes and proceed with navigation */
  onDiscard: () => void;
  /** Callback to cancel navigation and stay on page */
  onCancel: () => void;
  /** Optional custom message */
  message?: string;
}

/**
 * Dialog that warns users about unsaved changes before navigation.
 * Provides options to save, discard, or cancel navigation.
 *
 * @param props - Component props
 * @returns UnsavedChangesDialog component
 *
 * @example
 * ```tsx
 * const { isDirty } = useForm();
 * const { showDialog, confirmNavigation, cancelNavigation } = useUnsavedChanges(isDirty);
 *
 * const handleSave = async () => {
 *   await saveTheme();
 *   confirmNavigation();
 * };
 *
 * return (
 *   <>
 *     <form>...</form>
 *     <UnsavedChangesDialog
 *       open={showDialog}
 *       onSave={handleSave}
 *       onDiscard={confirmNavigation}
 *       onCancel={cancelNavigation}
 *     />
 *   </>
 * );
 * ```
 */
export default function UnsavedChangesDialog({
  open,
  onSave,
  onDiscard,
  onCancel,
  message = 'You have unsaved changes. Do you want to save them?',
}: UnsavedChangesDialogProps): JSX.Element {
  const [isSaving, setIsSaving] = useState(false);

  const handleSave = async (): Promise<void> => {
    setIsSaving(true);
    try {
      await onSave();
    } finally {
      setIsSaving(false);
    }
  };

  return (
    <Dialog
      open={open}
      onClose={onCancel}
      aria-labelledby="unsaved-changes-dialog-title"
      aria-describedby="unsaved-changes-dialog-description"
      maxWidth="sm"
      fullWidth
    >
      <DialogTitle id="unsaved-changes-dialog-title">Unsaved Changes</DialogTitle>
      <DialogContent>
        <DialogContentText id="unsaved-changes-dialog-description">{message}</DialogContentText>
      </DialogContent>
      <DialogActions>
        <Button onClick={onCancel} disabled={isSaving}>
          Cancel
        </Button>
        <Button onClick={onDiscard} color="error" disabled={isSaving}>
          Discard Changes
        </Button>
        <Button onClick={handleSave} variant="contained" disabled={isSaving} autoFocus>
          Save Changes
        </Button>
      </DialogActions>
    </Dialog>
  );
}
