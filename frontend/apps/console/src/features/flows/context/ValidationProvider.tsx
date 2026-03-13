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

import {type PropsWithChildren, type ReactElement, useCallback, useEffect, useMemo, useState} from 'react';
import Notification, {NotificationType} from '../models/notification';
import {ValidationContext, type ValidationConfig} from './ValidationContext';
import useFlowBuilderCore from '../hooks/useFlowBuilderCore';

export interface ValidationProviderProps {
  /**
   * Validation configuration settings.
   * @defaultValue { isOTPValidationEnabled: false, isRecoveryFactorValidationEnabled: false }
   */
  validationConfig?: ValidationConfig;
  /**
   * Provider children.
   * @defaultValue undefined
   */
  children?: React.ReactNode;
}

function ValidationProvider({
  children = undefined,
  validationConfig = {
    isOTPValidationEnabled: false,
    isRecoveryFactorValidationEnabled: false,
  },
}: PropsWithChildren<ValidationProviderProps>): ReactElement {
  const {setIsOpenResourcePropertiesPanel, registerCloseValidationPanel} = useFlowBuilderCore();

  const [notifications, setNotifications] = useState<Map<string, Notification>>(new Map());
  const [selectedNotification, setSelectedNotification] = useState<Notification | null>(null);
  const [openValidationPanel, setOpenValidationPanelInternal] = useState<boolean>(false);
  const [currentActiveTab, setCurrentActiveTab] = useState<number>(0);

  /**
   * Wrapper for setOpenValidationPanel that closes the resource properties panel
   * when opening the validation panel (mutual exclusion).
   */
  const setOpenValidationPanel = useCallback(
    (open: boolean): void => {
      if (open) {
        // Close resource properties panel when opening validation panel
        setIsOpenResourcePropertiesPanel(false);
      }
      setOpenValidationPanelInternal(open);
    },
    [setIsOpenResourcePropertiesPanel],
  );

  // Register the close callback with FlowBuilderCoreContext for mutual exclusion
  useEffect(() => {
    registerCloseValidationPanel(() => {
      setOpenValidationPanelInternal(false);
    });

    // Cleanup: clear the registration when unmounting
    return () => {
      registerCloseValidationPanel(() => {});
    };
  }, [registerCloseValidationPanel]);

  /**
   * Get the list of notifications.
   */
  const notificationList: Notification[] = useMemo(() => Array.from(notifications.values()), [notifications]);

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
   * @param id - The ID of the notification to retrieve.
   * @returns The notification with the specified ID, or undefined if not found.
   */
  const getNotification: (id: string) => Notification | undefined = useCallback(
    (id: string): Notification | undefined => notifications.get(id),
    [notifications],
  );

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
      setOpenValidationPanel,
      setSelectedNotification,
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
      setOpenValidationPanel,
      setSelectedNotification,
      validationConfig,
    ],
  );

  return <ValidationContext.Provider value={contextValue}>{children}</ValidationContext.Provider>;
}

export default ValidationProvider;
