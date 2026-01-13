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
import {render, screen, fireEvent} from '@testing-library/react';
import ValidationNotificationsList from '../ValidationNotificationsList';
import Notification, {NotificationType} from '../../../models/notification';

// Mock react-i18next
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => {
      const translations: Record<string, string> = {
        'common:show': 'Show',
      };
      return translations[key] || key;
    },
  }),
}));

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
    it('should display empty message when no notifications', () => {
      render(
        <ValidationNotificationsList
          notifications={[]}
          emptyMessage="No errors found"
          onNotificationClick={mockOnNotificationClick}
        />,
      );

      expect(screen.getByText('No errors found')).toBeInTheDocument();
    });

    it('should display empty message for undefined notifications', () => {
      render(
        <ValidationNotificationsList
          notifications={undefined as any}
          emptyMessage="No notifications"
          onNotificationClick={mockOnNotificationClick}
        />,
      );

      expect(screen.getByText('No notifications')).toBeInTheDocument();
    });
  });

  describe('Notification Rendering', () => {
    it('should render notification messages', () => {
      const notifications = [
        createNotification('1', 'Error message 1', NotificationType.ERROR),
        createNotification('2', 'Error message 2', NotificationType.ERROR),
      ];

      render(
        <ValidationNotificationsList
          notifications={notifications}
          emptyMessage="No errors"
          onNotificationClick={mockOnNotificationClick}
        />,
      );

      expect(screen.getByText('Error message 1')).toBeInTheDocument();
      expect(screen.getByText('Error message 2')).toBeInTheDocument();
    });

    it('should render different notification types', () => {
      const notifications = [
        createNotification('1', 'Error notification', NotificationType.ERROR),
        createNotification('2', 'Warning notification', NotificationType.WARNING),
        createNotification('3', 'Info notification', NotificationType.INFO),
      ];

      render(
        <ValidationNotificationsList
          notifications={notifications}
          emptyMessage="No notifications"
          onNotificationClick={mockOnNotificationClick}
        />,
      );

      expect(screen.getByText('Error notification')).toBeInTheDocument();
      expect(screen.getByText('Warning notification')).toBeInTheDocument();
      expect(screen.getByText('Info notification')).toBeInTheDocument();
    });
  });

  describe('Show Button', () => {
    it('should show "Show" button when notification has resources', () => {
      const notifications = [createNotification('1', 'Error with resources', NotificationType.ERROR, true)];

      render(
        <ValidationNotificationsList
          notifications={notifications}
          emptyMessage="No errors"
          onNotificationClick={mockOnNotificationClick}
        />,
      );

      expect(screen.getByRole('button', {name: 'Show'})).toBeInTheDocument();
    });

    it('should show "Show" button when notification has panel notification', () => {
      const notifications = [createNotification('1', 'Error with panel', NotificationType.ERROR, false, true)];

      render(
        <ValidationNotificationsList
          notifications={notifications}
          emptyMessage="No errors"
          onNotificationClick={mockOnNotificationClick}
        />,
      );

      expect(screen.getByRole('button', {name: 'Show'})).toBeInTheDocument();
    });

    it('should not show "Show" button when notification has no resources or panel notification', () => {
      const notifications = [createNotification('1', 'Simple error', NotificationType.ERROR)];

      render(
        <ValidationNotificationsList
          notifications={notifications}
          emptyMessage="No errors"
          onNotificationClick={mockOnNotificationClick}
        />,
      );

      expect(screen.queryByRole('button', {name: 'Show'})).not.toBeInTheDocument();
    });

    it('should call onNotificationClick when Show button is clicked', () => {
      const notifications = [createNotification('1', 'Error with resources', NotificationType.ERROR, true)];

      render(
        <ValidationNotificationsList
          notifications={notifications}
          emptyMessage="No errors"
          onNotificationClick={mockOnNotificationClick}
        />,
      );

      const showButton = screen.getByRole('button', {name: 'Show'});
      fireEvent.click(showButton);

      expect(mockOnNotificationClick).toHaveBeenCalledTimes(1);
      expect(mockOnNotificationClick).toHaveBeenCalledWith(notifications[0]);
    });
  });

  describe('Multiple Notifications', () => {
    it('should render all notifications in a list', () => {
      const notifications = [
        createNotification('1', 'First error', NotificationType.ERROR, true),
        createNotification('2', 'Second error', NotificationType.ERROR, true),
        createNotification('3', 'Third error', NotificationType.ERROR),
      ];

      render(
        <ValidationNotificationsList
          notifications={notifications}
          emptyMessage="No errors"
          onNotificationClick={mockOnNotificationClick}
        />,
      );

      expect(screen.getByText('First error')).toBeInTheDocument();
      expect(screen.getByText('Second error')).toBeInTheDocument();
      expect(screen.getByText('Third error')).toBeInTheDocument();

      // Only first two should have Show buttons
      const showButtons = screen.getAllByRole('button', {name: 'Show'});
      expect(showButtons).toHaveLength(2);
    });
  });
});
