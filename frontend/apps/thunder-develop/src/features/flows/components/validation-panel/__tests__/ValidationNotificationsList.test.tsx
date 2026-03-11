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

/* eslint-disable @typescript-eslint/no-explicit-any, @typescript-eslint/no-unsafe-argument, @typescript-eslint/no-unsafe-assignment */

import {describe, it, expect, vi, beforeEach} from 'vitest';
import {render} from '@thunder/test-utils/browser';
import {page, userEvent} from 'vitest/browser';
import ValidationNotificationsList from '../ValidationNotificationsList';
import Notification, {NotificationType} from '../../../models/notification';

describe('ValidationNotificationsList', () => {
  const mockOnNotificationClick = vi.fn();

  const createNotification = (
    id: string,
    message: string,
    type: NotificationType,
    hasResources = false,
    hasPanelNotification = false,
  ): Notification => {
    const notification = new Notification(id, message, type);
    if (hasResources) {
      notification.addResource({id: 'resource-1', type: 'TEST', category: 'TEST'} as any);
    }
    if (hasPanelNotification) {
      notification.setPanelNotification(<div>Panel Notification</div>);
    }
    return notification;
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Empty State', () => {
    it('should display empty message when no notifications', async () => {
      await render(
        <ValidationNotificationsList
          notifications={[]}
          emptyMessage="No errors found"
          onNotificationClick={mockOnNotificationClick}
        />,
      );

      await expect.element(page.getByText('No errors found')).toBeInTheDocument();
    });

    it('should display empty message for undefined notifications', async () => {
      await render(
        <ValidationNotificationsList
          notifications={undefined as any}
          emptyMessage="No notifications"
          onNotificationClick={mockOnNotificationClick}
        />,
      );

      await expect.element(page.getByText('No notifications')).toBeInTheDocument();
    });
  });

  describe('Notification Rendering', () => {
    it('should render notification messages', async () => {
      const notifications = [
        createNotification('1', 'Error message 1', NotificationType.ERROR),
        createNotification('2', 'Error message 2', NotificationType.ERROR),
      ];

      await render(
        <ValidationNotificationsList
          notifications={notifications}
          emptyMessage="No errors"
          onNotificationClick={mockOnNotificationClick}
        />,
      );

      await expect.element(page.getByText('Error message 1')).toBeInTheDocument();
      await expect.element(page.getByText('Error message 2')).toBeInTheDocument();
    });

    it('should render different notification types', async () => {
      const notifications = [
        createNotification('1', 'Error notification', NotificationType.ERROR),
        createNotification('2', 'Warning notification', NotificationType.WARNING),
        createNotification('3', 'Info notification', NotificationType.INFO),
      ];

      await render(
        <ValidationNotificationsList
          notifications={notifications}
          emptyMessage="No notifications"
          onNotificationClick={mockOnNotificationClick}
        />,
      );

      await expect.element(page.getByText('Error notification')).toBeInTheDocument();
      await expect.element(page.getByText('Warning notification')).toBeInTheDocument();
      await expect.element(page.getByText('Info notification')).toBeInTheDocument();
    });
  });

  describe('Show Button', () => {
    it('should show "Show" button when notification has resources', async () => {
      const notifications = [createNotification('1', 'Error with resources', NotificationType.ERROR, true)];

      await render(
        <ValidationNotificationsList
          notifications={notifications}
          emptyMessage="No errors"
          onNotificationClick={mockOnNotificationClick}
        />,
      );

      await expect.element(page.getByRole('button', {name: 'Show'})).toBeInTheDocument();
    });

    it('should show "Show" button when notification has panel notification', async () => {
      const notifications = [createNotification('1', 'Error with panel', NotificationType.ERROR, false, true)];

      await render(
        <ValidationNotificationsList
          notifications={notifications}
          emptyMessage="No errors"
          onNotificationClick={mockOnNotificationClick}
        />,
      );

      await expect.element(page.getByRole('button', {name: 'Show'})).toBeInTheDocument();
    });

    it('should not show "Show" button when notification has no resources or panel notification', async () => {
      const notifications = [createNotification('1', 'Simple error', NotificationType.ERROR)];

      await render(
        <ValidationNotificationsList
          notifications={notifications}
          emptyMessage="No errors"
          onNotificationClick={mockOnNotificationClick}
        />,
      );

      await expect.element(page.getByRole('button', {name: 'Show'})).not.toBeInTheDocument();
    });

    it('should call onNotificationClick when Show button is clicked', async () => {
      const notifications = [createNotification('1', 'Error with resources', NotificationType.ERROR, true)];

      await render(
        <ValidationNotificationsList
          notifications={notifications}
          emptyMessage="No errors"
          onNotificationClick={mockOnNotificationClick}
        />,
      );

      const showButton = page.getByRole('button', {name: 'Show'});
      await userEvent.click(showButton);

      expect(mockOnNotificationClick).toHaveBeenCalledTimes(1);
      expect(mockOnNotificationClick).toHaveBeenCalledWith(notifications[0]);
    });
  });

  describe('Multiple Notifications', () => {
    it('should render all notifications in a list', async () => {
      const notifications = [
        createNotification('1', 'First error', NotificationType.ERROR, true),
        createNotification('2', 'Second error', NotificationType.ERROR, true),
        createNotification('3', 'Third error', NotificationType.ERROR),
      ];

      await render(
        <ValidationNotificationsList
          notifications={notifications}
          emptyMessage="No errors"
          onNotificationClick={mockOnNotificationClick}
        />,
      );

      await expect.element(page.getByText('First error')).toBeInTheDocument();
      await expect.element(page.getByText('Second error')).toBeInTheDocument();
      await expect.element(page.getByText('Third error')).toBeInTheDocument();

      // Only first two should have Show buttons
      const showButtons = page.getByRole('button', {name: 'Show'}).all();
      expect(showButtons).toHaveLength(2);
    });
  });
});
