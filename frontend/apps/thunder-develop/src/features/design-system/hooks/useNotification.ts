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

import {useContext} from 'react';
import {NotificationContext} from '../providers/NotificationProvider';

/**
 * Return type for the useNotification hook.
 */
export interface UseNotificationReturn {
  /** Show a success notification */
  showSuccess: (message: string) => void;
  /** Show an error notification */
  showError: (message: string) => void;
  /** Show an info notification */
  showInfo: (message: string) => void;
  /** Show a warning notification */
  showWarning: (message: string) => void;
}

/**
 * Hook to access the notification system.
 * Must be used within a NotificationProvider.
 *
 * @returns Notification functions
 *
 * @example
 * ```tsx
 * const { showSuccess, showError } = useNotification();
 *
 * const handleSave = async () => {
 *   try {
 *     await saveTheme();
 *     showSuccess('Theme saved successfully');
 *   } catch (error) {
 *     showError('Failed to save theme');
 *   }
 * };
 * ```
 */
export default function useNotification(): UseNotificationReturn {
  const context = useContext(NotificationContext);

  if (context === undefined) {
    throw new Error('useNotification must be used within a NotificationProvider');
  }

  return context;
}
