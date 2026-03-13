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

import {useTranslation} from 'react-i18next';
import type {ReactElement} from 'react';
import {Alert, Box, Button, List, ListItem, Typography} from '@wso2/oxygen-ui';
import Notification from '../../models/notification';

/**
 * Props interface of {@link ValidationNotificationsList}
 */
export interface ValidationNotificationsListPropsInterface {
  /**
   * Array of notifications to display.
   */
  notifications: Notification[];
  /**
   * Message to display when no notifications are available.
   */
  emptyMessage: string;
  /**
   * Callback fired when a notification is clicked.
   */
  onNotificationClick: (notification: Notification) => void;
}

/**
 * Component to render a list of validation notifications.
 *
 * @param props - Props injected to the component.
 * @returns The ValidationNotificationsList component.
 */
function ValidationNotificationsList({
  notifications,
  emptyMessage,
  onNotificationClick,
}: ValidationNotificationsListPropsInterface): ReactElement {
  const {t} = useTranslation();

  if (!notifications || notifications.length === 0) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" minHeight="200px">
        <Typography variant="body2" color="textSecondary" fontStyle="italic">
          {emptyMessage}
        </Typography>
      </Box>
    );
  }

  return (
    <Box>
      <List>
        {notifications.map((notification: Notification) => (
          <ListItem key={notification.getId()}>
            <Alert icon={false} className="notification-item" severity={notification.getType()}>
              {notification.getMessage()}
              {(notification.hasResources() || notification.hasPanelNotification()) && (
                <Box textAlign="right">
                  <Button
                    variant="text"
                    size="small"
                    color={notification.getType()}
                    onClick={() => onNotificationClick(notification)}
                    disableRipple
                    className="notification-action-button"
                  >
                    {t('common:show')}
                  </Button>
                </Box>
              )}
            </Alert>
          </ListItem>
        ))}
      </List>
    </Box>
  );
}

export default ValidationNotificationsList;
