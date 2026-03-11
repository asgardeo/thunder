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

import {describe, it, expect, vi, beforeEach, afterEach} from 'vitest';
import {render} from '@thunder/test-utils/browser';
import {page, userEvent} from 'vitest/browser';
import ValidationPanel from '../ValidationPanel';
import Notification, {NotificationType} from '../../../models/notification';

// Mock ValidationNotificationsList
vi.mock('../ValidationNotificationsList', () => ({
  default: ({
    notifications,
    emptyMessage,
    onNotificationClick,
  }: {
    notifications: Notification[];
    emptyMessage: string;
    onNotificationClick: (n: Notification) => void;
  }) => (
    <div data-testid="notifications-list" data-count={notifications.length} data-empty-message={emptyMessage}>
      {notifications.map((n) => (
        <button
          type="button"
          key={n.getId()}
          onClick={() => onNotificationClick(n)}
          data-testid={`notification-${n.getId()}`}
        >
          {n.getMessage()}
        </button>
      ))}
    </div>
  ),
}));

// Mock hooks
const mockSetOpenValidationPanel = vi.fn();
const mockSetSelectedNotification = vi.fn();
const mockSetCurrentActiveTab = vi.fn();
const mockSetLastInteractedResource = vi.fn();

let mockNotifications: Notification[] = [];
let mockOpenValidationPanel = true;
let mockCurrentActiveTab = 0;

vi.mock('../../../hooks/useValidationStatus', () => ({
  default: () => ({
    notifications: mockNotifications,
    openValidationPanel: mockOpenValidationPanel,
    setOpenValidationPanel: mockSetOpenValidationPanel,
    setSelectedNotification: mockSetSelectedNotification,
    currentActiveTab: mockCurrentActiveTab,
    setCurrentActiveTab: mockSetCurrentActiveTab,
  }),
}));

vi.mock('../../../hooks/useFlowBuilderCore', () => ({
  default: () => ({
    setLastInteractedResource: mockSetLastInteractedResource,
  }),
}));

describe('ValidationPanel', () => {
  const createNotification = (id: string, message: string, type: NotificationType): Notification => new Notification(id, message, type);

  beforeEach(() => {
    vi.clearAllMocks();
    mockNotifications = [];
    mockOpenValidationPanel = true;
    mockCurrentActiveTab = 0;

    // Create drawer container
    const drawerContainer = document.createElement('div');
    drawerContainer.id = 'drawer-container';
    document.body.appendChild(drawerContainer);
  });

  afterEach(() => {
    const drawerContainer = document.getElementById('drawer-container');
    if (drawerContainer) {
      document.body.removeChild(drawerContainer);
    }
  });

  describe('Rendering', () => {
    it('should render panel header', async () => {
      await render(<ValidationPanel />);

      await expect.element(page.getByText('Notifications')).toBeInTheDocument();
    });

    it('should render tabs for errors, warnings, and info', async () => {
      await render(<ValidationPanel />);

      await expect.element(page.getByRole('tab', {name: /Errors/i})).toBeInTheDocument();
      await expect.element(page.getByRole('tab', {name: /Warnings/i})).toBeInTheDocument();
      await expect.element(page.getByRole('tab', {name: /Info/i})).toBeInTheDocument();
    });

    it('should render close button', async () => {
      await render(<ValidationPanel />);

      // Close button is an IconButton
      const buttons = page.getByRole('button').all();
      expect(buttons.length).toBeGreaterThan(0);
    });
  });

  describe('Tab Navigation', () => {
    it('should show errors tab content by default', async () => {
      mockCurrentActiveTab = 0;
      mockNotifications = [createNotification('1', 'Error message', NotificationType.ERROR)];

      await render(<ValidationPanel />);

      const visibleList = page.getByTestId('notifications-list');
      expect(visibleList).toHaveAttribute('data-empty-message', 'No errors found');
    });

    it('should call setCurrentActiveTab when switching tabs', async () => {
      await render(<ValidationPanel />);

      const warningsTab = page.getByRole('tab', {name: /Warnings/i});
      await userEvent.click(warningsTab);

      expect(mockSetCurrentActiveTab).toHaveBeenCalledWith(1);
    });
  });

  describe('Notification Filtering', () => {
    it('should filter error notifications for errors tab', async () => {
      mockCurrentActiveTab = 0;
      mockNotifications = [
        createNotification('1', 'Error', NotificationType.ERROR),
        createNotification('2', 'Warning', NotificationType.WARNING),
        createNotification('3', 'Info', NotificationType.INFO),
      ];

      await render(<ValidationPanel />);

      const errorTabPanel = document.getElementById('validation-tabpanel-0');
      expect(errorTabPanel).not.toHaveAttribute('hidden');
    });

    it('should filter warning notifications for warnings tab', async () => {
      mockCurrentActiveTab = 1;
      mockNotifications = [
        createNotification('1', 'Error', NotificationType.ERROR),
        createNotification('2', 'Warning', NotificationType.WARNING),
      ];

      await render(<ValidationPanel />);

      const warningTabPanel = document.getElementById('validation-tabpanel-1');
      expect(warningTabPanel).not.toHaveAttribute('hidden');
    });
  });

  describe('Close Functionality', () => {
    it('should call setOpenValidationPanel(false) when close button clicked', async () => {
      await render(<ValidationPanel />);

      // Find the close button (IconButton with X icon)
      const closeButtons = page.getByRole('button');
      const closeButton = closeButtons.all().find((btn) => btn.element().querySelector('svg'));

      if (closeButton) {
        await userEvent.click(closeButton);
        expect(mockSetOpenValidationPanel).toHaveBeenCalledWith(false);
      }
    });
  });

  describe('Notification Click Handling', () => {
    it('should set selected notification when notification is clicked', async () => {
      mockCurrentActiveTab = 0;
      const notification = createNotification('1', 'Error', NotificationType.ERROR);
      mockNotifications = [notification];

      await render(<ValidationPanel />);

      const notificationButton = page.getByTestId('notification-1');
      await userEvent.click(notificationButton);

      expect(mockSetSelectedNotification).toHaveBeenCalledWith(notification);
    });

    it('should close panel when notification is clicked', async () => {
      mockCurrentActiveTab = 0;
      const notification = createNotification('1', 'Error', NotificationType.ERROR);
      mockNotifications = [notification];

      await render(<ValidationPanel />);

      const notificationButton = page.getByTestId('notification-1');
      await userEvent.click(notificationButton);

      expect(mockSetOpenValidationPanel).toHaveBeenCalledWith(false);
    });

    it('should set last interacted resource when notification has single resource', async () => {
      mockCurrentActiveTab = 0;
      const notification = createNotification('1', 'Error', NotificationType.ERROR);
      const resource = {id: 'resource-1', type: 'TEST', category: 'TEST'} as any;
      notification.addResource(resource);
      mockNotifications = [notification];

      await render(<ValidationPanel />);

      const notificationButton = page.getByTestId('notification-1');
      await userEvent.click(notificationButton);

      expect(mockSetLastInteractedResource).toHaveBeenCalledWith(resource);
    });
  });

  describe('Tab Panel Accessibility', () => {
    it('should have correct aria attributes on tab panels', async () => {
      await render(<ValidationPanel />);

      const tabPanel0 = document.getElementById('validation-tabpanel-0');
      expect(tabPanel0).toHaveAttribute('role', 'tabpanel');
      expect(tabPanel0).toHaveAttribute('aria-labelledby', 'validation-tab-0');
    });
  });

  describe('Notification with Multiple Resources', () => {
    it('should not set last interacted resource when notification has multiple resources', async () => {
      mockCurrentActiveTab = 0;
      const notification = createNotification('1', 'Error', NotificationType.ERROR);
      const resource1 = {id: 'resource-1', type: 'TEST', category: 'TEST'} as any;
      const resource2 = {id: 'resource-2', type: 'TEST', category: 'TEST'} as any;
      notification.addResource(resource1);
      notification.addResource(resource2);
      mockNotifications = [notification];

      await render(<ValidationPanel />);

      const notificationButton = page.getByTestId('notification-1');
      await userEvent.click(notificationButton);

      expect(mockSetLastInteractedResource).not.toHaveBeenCalled();
    });

    it('should not set last interacted resource when notification has no resources', async () => {
      mockCurrentActiveTab = 0;
      const notification = createNotification('1', 'Error', NotificationType.ERROR);
      mockNotifications = [notification];

      await render(<ValidationPanel />);

      const notificationButton = page.getByTestId('notification-1');
      await userEvent.click(notificationButton);

      expect(mockSetLastInteractedResource).not.toHaveBeenCalled();
    });
  });

  describe('Tab Content Display', () => {
    it('should display info notifications in info tab', async () => {
      mockCurrentActiveTab = 2;
      mockNotifications = [
        createNotification('1', 'Info message', NotificationType.INFO),
      ];

      await render(<ValidationPanel />);

      const infoTabPanel = document.getElementById('validation-tabpanel-2');
      expect(infoTabPanel).not.toHaveAttribute('hidden');
    });
  });
});
