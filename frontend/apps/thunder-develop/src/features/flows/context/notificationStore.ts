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

import {useSyncExternalStore} from 'react';
import type Notification from '../models/notification';

// PERFORMANCE: External store for notifications to allow fine-grained subscriptions
// This prevents components from re-rendering when unrelated context values change
// (e.g., when selectedNotification changes, ValidationErrorBoundary won't re-render)

const notificationStoreListeners = new Set<() => void>();
let notificationStoreSnapshot: Notification[] = [];

/**
 * Updates the notification store and notifies all subscribers.
 * Called by ValidationProvider when notifications change.
 */
export function updateNotificationStore(notifications: Notification[]): void {
  // Always update the snapshot - the array reference from useMemo is new each time
  // notifications state changes, so we just need to update and notify
  notificationStoreSnapshot = notifications;
  notificationStoreListeners.forEach((listener) => listener());
}

/**
 * Hook to subscribe ONLY to notifications, not the entire ValidationContext.
 * PERFORMANCE: This prevents re-renders when selectedNotification or openValidationPanel change.
 *
 * Use this hook in components that only need to read notifications (like ValidationErrorBoundary)
 * instead of useValidationStatus() which subscribes to the entire context.
 */
export function useNotificationsOnly(): Notification[] {
  return useSyncExternalStore(
    (callback) => {
      notificationStoreListeners.add(callback);
      return () => notificationStoreListeners.delete(callback);
    },
    () => notificationStoreSnapshot,
    () => notificationStoreSnapshot,
  );
}
