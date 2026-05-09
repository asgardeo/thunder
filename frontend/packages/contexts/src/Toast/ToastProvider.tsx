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

import {Alert, Snackbar} from '@wso2/oxygen-ui';
import {useState, useCallback, useMemo} from 'react';
import type {SyntheticEvent, PropsWithChildren, JSX} from 'react';
import ToastContext, {type ToastSeverity} from './ToastContext';

/**
 * Internal state shape for the active toast notification.
 */
interface ToastState {
  open: boolean;
  message: string;
  severity: ToastSeverity;
}

/**
 * Props for the ToastProvider component.
 *
 * @public
 */
export type ToastProviderProps = PropsWithChildren;

/**
 * React context provider component that enables toast notifications throughout the application.
 *
 * This component manages the lifecycle of a single snackbar notification rendered at the
 * bottom-right of the viewport. It exposes a `showToast` function via `ToastContext` so
 * that any descendant component or hook can trigger a notification without needing to manage
 * local state.
 *
 * Wrap your application (or a subtree) with this provider and consume notifications using
 * the `useToast` hook.
 *
 * @example
 * Basic setup in the application root:
 * ```tsx
 * import ToastProvider from './ToastProvider';
 *
 * function App() {
 *   return (
 *     <ToastProvider>
 *       <Routes />
 *     </ToastProvider>
 *   );
 * }
 * ```
 *
 * @example
 * Triggering a toast from a mutation hook:
 * ```ts
 * import useToast from './useToast';
 *
 * function useCreateItem() {
 *   const { showToast } = useToast();
 *
 *   return useMutation({
 *     mutationFn: createItem,
 *     onSuccess: () => showToast('Item created successfully.', 'success'),
 *     onError: () => showToast('Failed to create item.', 'error'),
 *   });
 * }
 * ```
 *
 * @public
 */
export default function ToastProvider({children}: ToastProviderProps): JSX.Element {
  const [toast, setToast] = useState<ToastState>({
    open: false,
    message: '',
    severity: 'success',
  });

  const showToast = useCallback((message: string, severity: ToastSeverity = 'success'): void => {
    setToast({open: true, message, severity});
  }, []);

  const handleClose = useCallback((_event?: SyntheticEvent | Event, reason?: string): void => {
    if (reason === 'clickaway') return;
    setToast((prev) => ({...prev, open: false}));
  }, []);

  const contextValue = useMemo(() => ({showToast}), [showToast]);

  return (
    <ToastContext.Provider value={contextValue}>
      {children}
      <Snackbar
        open={toast.open}
        autoHideDuration={6000}
        onClose={handleClose}
        anchorOrigin={{vertical: 'bottom', horizontal: 'right'}}
      >
        <Alert onClose={handleClose} severity={toast.severity} sx={{width: '100%'}}>
          {toast.message}
        </Alert>
      </Snackbar>
    </ToastContext.Provider>
  );
}
