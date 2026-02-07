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

import {useCallback, useEffect, useState} from 'react';

/**
 * Return type for the useUnsavedChanges hook.
 */
export interface UseUnsavedChangesReturn {
  /** Whether unsaved changes dialog should be shown */
  showDialog: boolean;
  /** Show the unsaved changes dialog */
  showUnsavedDialog: () => void;
  /** Hide the unsaved changes dialog */
  hideUnsavedDialog: () => void;
  /** Confirm navigation (proceed with leaving the page) */
  confirmNavigation: () => void;
  /** Cancel navigation (stay on the page) */
  cancelNavigation: () => void;
  /** Callback to check before navigation */
  checkUnsavedChanges: () => boolean;
}

/**
 * Hook to track unsaved changes and warn before navigation.
 * Shows a browser confirmation dialog before leaving the page and
 * provides manual control for in-app navigation.
 *
 * Note: This hook uses beforeunload for browser navigation (refresh, close tab).
 * For in-app navigation, you need to manually check isDirty and call showUnsavedDialog.
 *
 * @param isDirty - Whether there are unsaved changes
 * @returns Dialog state and navigation handlers
 *
 * @example
 * ```tsx
 * const { isDirty } = useForm();
 * const { showDialog, showUnsavedDialog, confirmNavigation, cancelNavigation } = useUnsavedChanges(isDirty);
 *
 * const handleBack = () => {
 *   if (isDirty) {
 *     showUnsavedDialog();
 *   } else {
 *     navigate('/themes');
 *   }
 * };
 *
 * return (
 *   <>
 *     <Button onClick={handleBack}>Back</Button>
 *     <UnsavedChangesDialog
 *       open={showDialog}
 *       onSave={() => {
 *         handleSave();
 *         confirmNavigation();
 *       }}
 *       onDiscard={confirmNavigation}
 *       onCancel={cancelNavigation}
 *     />
 *   </>
 * );
 * ```
 */
export default function useUnsavedChanges(isDirty: boolean): UseUnsavedChangesReturn {
  const [showDialog, setShowDialog] = useState(false);
  const [pendingNavigation, setPendingNavigation] = useState<(() => void) | null>(null);

  // Show unsaved changes dialog
  const showUnsavedDialog = useCallback(() => {
    setShowDialog(true);
  }, []);

  // Hide unsaved changes dialog
  const hideUnsavedDialog = useCallback(() => {
    setShowDialog(false);
  }, []);

  // Confirm navigation - proceed and execute pending navigation if any
  const confirmNavigation = useCallback(() => {
    setShowDialog(false);
    if (pendingNavigation) {
      pendingNavigation();
      setPendingNavigation(null);
    }
  }, [pendingNavigation]);

  // Cancel navigation - stay on current page
  const cancelNavigation = useCallback(() => {
    setShowDialog(false);
    setPendingNavigation(null);
  }, []);

  // Check if there are unsaved changes (for manual navigation checks)
  const checkUnsavedChanges = useCallback(() => {
    return isDirty;
  }, [isDirty]);

  // Warn before page reload/close
  useEffect(() => {
    const handleBeforeUnload = (event: BeforeUnloadEvent): string | undefined => {
      if (isDirty) {
        event.preventDefault();
        return (event.returnValue = '');
      }
    };

    window.addEventListener('beforeunload', handleBeforeUnload);

    return () => {
      window.removeEventListener('beforeunload', handleBeforeUnload);
    };
  }, [isDirty]);

  return {
    showDialog,
    showUnsavedDialog,
    hideUnsavedDialog,
    confirmNavigation,
    cancelNavigation,
    checkUnsavedChanges,
  };
}
