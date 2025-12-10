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

import {type PropsWithChildren, type ReactElement, useCallback, useMemo, useRef, useState} from 'react';
import Notification, {NotificationType} from '../models/notification';
import {ValidationContext, type ValidationConfig} from './ValidationContext';

export interface ValidationProviderProps {
  /**
   * Validation configuration settings.
   */
  validationConfig?: ValidationConfig;
}

function ValidationProvider({
  children,
  validationConfig = {
    isOTPValidationEnabled: false,
    isRecoveryFactorValidationEnabled: false,
  },
}: PropsWithChildren<ValidationProviderProps>): ReactElement {
  const [notifications, setNotifications] = useState<Map<string, Notification>>(new Map());
  const [selectedNotification, setSelectedNotification] = useState<Notification | null>(null);
  const [openValidationPanel, setOpenValidationPanel] = useState<boolean>(false);
  const [currentActiveTab, setCurrentActiveTab] = useState<number>(0);

  // PERFORMANCE: Use ref to access current notifications without dependency
  // This prevents getNotification from changing on every notification update
  const notificationsRef = useRef(notifications);
  notificationsRef.current = notifications;

  // PERFORMANCE: Store setters in refs to create stable callbacks
  const setOpenValidationPanelRef = useRef(setOpenValidationPanel);
  const setSelectedNotificationRef = useRef(setSelectedNotification);
  setOpenValidationPanelRef.current = setOpenValidationPanel;
  setSelectedNotificationRef.current = setSelectedNotification;

  /**
   * Get the list of notifications.
   */
  const notificationList: Notification[] = useMemo(
    () => Array.from(notifications.values()),
    [notifications],
  );

  /**
   * Indicates whether the current state of the flow is valid.
   */
  const isValid: boolean = useMemo(
    () => notificationList.every((notification: Notification) => notification.getType() !== NotificationType.ERROR),
    [notificationList],
  );

  /**
   * Add a notification.
   * @param notification - The notification to add.
   */
  const addNotification: (notification: Notification) => void = useCallback((notification: Notification): void => {
    setNotifications((prev: Map<string, Notification>) => new Map(prev).set(notification.getId(), notification));
    setSelectedNotification(notification);
  }, []);

  /**
   * Remove a notification.
   * @param id - The ID of the notification to remove.
   */
  const removeNotification: (id: string) => void = useCallback((id: string): void => {
    setNotifications((prev: Map<string, Notification>) => {
      const updated = new Map<string, Notification>(prev);

      updated.delete(id);

      return updated;
    });
    setSelectedNotification((prev: Notification | null) => {
      if (prev?.getId() === id) {
        return null;
      }

      return prev;
    });
  }, []);

  /**
   * Gets a notification by its ID.
   * PERFORMANCE: Uses ref to avoid recreating this callback when notifications change.
   * This prevents cascading re-renders in useRequiredFields.
   * @param id - The ID of the notification to retrieve.
   * @returns The notification with the specified ID, or undefined if not found.
   */
  const getNotification: (id: string) => Notification | undefined = useCallback(
    (id: string): Notification | undefined => notificationsRef.current.get(id),
    [],
  );

  // PERFORMANCE: Stable callback for setOpenValidationPanel - never changes
  const setOpenValidationPanelStable = useCallback((isOpen: boolean): void => {
    setOpenValidationPanelRef.current(isOpen);
  }, []);

  // PERFORMANCE: Stable callback for setSelectedNotification - never changes
  const setSelectedNotificationStable = useCallback((notification: Notification | null): void => {
    setSelectedNotificationRef.current(notification);
  }, []);

  const contextValue = useMemo(
    () => ({
      addNotification,
      currentActiveTab,
      getNotification,
      isValid,
      notifications: notificationList,
      openValidationPanel,
      removeNotification,
      selectedNotification,
      setCurrentActiveTab,
      // PERFORMANCE: Use stable callbacks that never change
      setOpenValidationPanel: setOpenValidationPanelStable,
      setSelectedNotification: setSelectedNotificationStable,
      validationConfig,
    }),
    [
      addNotification,
      currentActiveTab,
      getNotification,
      isValid,
      notificationList,
      openValidationPanel,
      removeNotification,
      selectedNotification,
      setCurrentActiveTab,
      // PERFORMANCE: Stable callbacks - these never change so don't cause re-renders
      setOpenValidationPanelStable,
      setSelectedNotificationStable,
      validationConfig,
    ],
  );

  return <ValidationContext.Provider value={contextValue}>{children}</ValidationContext.Provider>;
}

export default ValidationProvider;
