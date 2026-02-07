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

import type {JSX, ReactNode} from 'react';
import {createContext, useCallback, useState} from 'react';
import {Alert, Snackbar} from '@mui/material';
import type {UseNotificationReturn} from '../hooks/useNotification';

/**
 * Notification severity types.
 */
type NotificationSeverity = 'success' | 'error' | 'info' | 'warning';

/**
 * Internal notification state.
 */
interface NotificationState {
  open: boolean;
  message: string;
  severity: NotificationSeverity;
}

/**
 * Props for the NotificationProvider component.
 */
interface NotificationProviderProps {
  /** Child components */
  children: ReactNode;
  /** Auto-dismiss duration in milliseconds (default: 6000) */
  autoHideDuration?: number;
}

/**
 * Context for notification system.
 */
export const NotificationContext = createContext<UseNotificationReturn | undefined>(undefined);

/**
 * Provider component for centralized notification management.
 * Wraps the application to provide toast notifications via Snackbar.
 *
 * @param props - Component props
 * @returns NotificationProvider component
 *
 * @example
 * ```tsx
 * // In AppWithConfig.tsx
 * <QueryClientProvider client={queryClient}>
 *   <NotificationProvider>
 *     <I18nProvider>
 *       <App />
 *     </I18nProvider>
 *   </NotificationProvider>
 * </QueryClientProvider>
 * ```
 */
export default function NotificationProvider({
  children,
  autoHideDuration = 6000,
}: NotificationProviderProps): JSX.Element {
  const [notification, setNotification] = useState<NotificationState>({
    open: false,
    message: '',
    severity: 'info',
  });

  const showNotification = useCallback(
    (message: string, severity: NotificationSeverity) => {
      setNotification({
        open: true,
        message,
        severity,
      });
    },
    []
  );

  const showSuccess = useCallback(
    (message: string) => {
      showNotification(message, 'success');
    },
    [showNotification]
  );

  const showError = useCallback(
    (message: string) => {
      showNotification(message, 'error');
    },
    [showNotification]
  );

  const showInfo = useCallback(
    (message: string) => {
      showNotification(message, 'info');
    },
    [showNotification]
  );

  const showWarning = useCallback(
    (message: string) => {
      showNotification(message, 'warning');
    },
    [showNotification]
  );

  const handleClose = useCallback((_event?: React.SyntheticEvent | Event, reason?: string) => {
    // Don't close on clickaway
    if (reason === 'clickaway') {
      return;
    }

    setNotification((prev) => ({
      ...prev,
      open: false,
    }));
  }, []);

  const contextValue: UseNotificationReturn = {
    showSuccess,
    showError,
    showInfo,
    showWarning,
  };

  return (
    <NotificationContext.Provider value={contextValue}>
      {children}
      <Snackbar
        open={notification.open}
        autoHideDuration={autoHideDuration}
        onClose={handleClose}
        anchorOrigin={{vertical: 'bottom', horizontal: 'right'}}
      >
        <Alert onClose={handleClose} severity={notification.severity} variant="filled" sx={{width: '100%'}}>
          {notification.message}
        </Alert>
      </Snackbar>
    </NotificationContext.Provider>
  );
}
