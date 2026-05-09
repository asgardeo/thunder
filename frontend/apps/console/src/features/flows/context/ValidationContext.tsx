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

import {type Context, createContext} from 'react';
import Notification from '../models/notification';

export interface ValidationConfig {
  /**
   * Whether OTP validation is enabled.
   */
  isOTPValidationEnabled?: boolean;
  /**
   * Whether recovery factor validation is enabled.
   */
  isRecoveryFactorValidationEnabled?: boolean;
  /**
   * Whether password executor validation is enabled.
   */
  isPasswordExecutorValidationEnabled?: boolean;
}

/**
 * Context for managing validation state and notifications.
 */
export interface ValidationContextProps {
  /**
   * Validation configuration settings.
   */
  validationConfig?: ValidationConfig;
  /**
   * Indicates whether the current state of the flow is valid.
   */
  isValid: boolean;
  /**
   * List of notifications related to the validation state.
   */
  notifications: Notification[];
  /**
   * Adds a notification to the list of notifications.
   *
   * @param notification - The notification to add.
   * @returns The updated list of notifications.
   */
  addNotification?: (notification: Notification) => void;
  /**
   * Removes a notification from the list of notifications.
   *
   * @param notificationId - The ID of the notification to remove.
   * @returns The updated list of notifications.
   */
  removeNotification?: (notificationId: string) => void;
  /**
   * The currently selected notification.
   */
  selectedNotification?: Notification | null;
  /**
   * Sets the currently selected notification.
   *
   * @param notification - The notification to select.
   */
  setSelectedNotification?: (notification: Notification | null) => void;
  /**
   * Indicates whether the validation panel is open.
   */
  openValidationPanel?: boolean;
  /**
   * Sets the visibility of the validation panel.
   *
   * @param open - Whether the validation panel should be open.
   */
  setOpenValidationPanel?: (open: boolean) => void;
  /**
   * The currently active tab index.
   */
  currentActiveTab?: number;
  /**
   * Sets the currently active tab index.
   *
   * @param tab - The index of the tab to set as active.
   */
  setCurrentActiveTab?: (tab: number) => void;
  /**
   * Gets a notification by its ID.
   * @param id - The ID of the notification to retrieve.
   * @returns The notification with the specified ID, or undefined if not found.
   */
  getNotification: (id: string) => Notification | undefined;
}

/**
 * Validation context for managing flow validation state and notifications.
 */
export const ValidationContext: Context<ValidationContextProps> = createContext<ValidationContextProps>({
  addNotification: undefined,
  currentActiveTab: 0,
  getNotification: () => undefined,
  isValid: true,
  notifications: [],
  openValidationPanel: false,
  removeNotification: undefined,
  selectedNotification: undefined,
  setCurrentActiveTab: undefined,
  setOpenValidationPanel: undefined,
  validationConfig: {
    isOTPValidationEnabled: false,
    isPasswordExecutorValidationEnabled: false,
    isRecoveryFactorValidationEnabled: false,
  },
});
