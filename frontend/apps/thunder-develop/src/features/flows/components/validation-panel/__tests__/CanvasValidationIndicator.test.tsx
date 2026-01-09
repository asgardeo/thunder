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

import {describe, it, expect, vi, beforeEach} from 'vitest';
import {render, screen, fireEvent} from '@testing-library/react';
import CanvasValidationIndicator from '../CanvasValidationIndicator';
import Notification, {NotificationType} from '../../../models/notification';

// Mock react-i18next
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => {
      const translations: Record<string, string> = {
        'flows:core.notificationPanel.tabs.errors': 'Errors',
        'flows:core.notificationPanel.tabs.warnings': 'Warnings',
        'flows:core.notificationPanel.tabs.info': 'Info',
        'flows:core.notificationPanel.trigger.label': 'Validation Status',
      };
      return translations[key] || key;
    },
  }),
}));

// Mock @xyflow/react
vi.mock('@xyflow/react', () => ({
  Panel: ({children, className}: {children: React.ReactNode; className: string}) => (
    <div data-testid="panel" className={className}>
      {children}
    </div>
  ),
}));

// Mock hooks
const mockSetCurrentActiveTab = vi.fn();
const mockSetOpenValidationPanel = vi.fn();
const mockSetIsOpenResourcePropertiesPanel = vi.fn();

let mockNotifications: Notification[] = [];
let mockOpenValidationPanel = false;
let mockIsResourcePropertiesPanelOpen = false;

vi.mock('../../../hooks/useValidationStatus', () => ({
  default: () => ({
    notifications: mockNotifications,
    setCurrentActiveTab: mockSetCurrentActiveTab,
    openValidationPanel: mockOpenValidationPanel,
    setOpenValidationPanel: mockSetOpenValidationPanel,
  }),
}));

vi.mock('../../../hooks/useFlowBuilderCore', () => ({
  default: () => ({
    setIsOpenResourcePropertiesPanel: mockSetIsOpenResourcePropertiesPanel,
    isResourcePropertiesPanelOpen: mockIsResourcePropertiesPanelOpen,
  }),
}));

describe('CanvasValidationIndicator', () => {
  const createNotification = (id: string, message: string, type: NotificationType): Notification => new Notification(id, message, type);

  beforeEach(() => {
    vi.clearAllMocks();
    mockNotifications = [];
    mockOpenValidationPanel = false;
    mockIsResourcePropertiesPanelOpen = false;
  });

  describe('Visibility', () => {
    it('should return null when there are no notifications', () => {
      mockNotifications = [];

      const {container} = render(<CanvasValidationIndicator />);

      expect(container.firstChild).toBeNull();
    });

    it('should render when there are notifications', () => {
      mockNotifications = [createNotification('1', 'Error', NotificationType.ERROR)];

      render(<CanvasValidationIndicator />);

      expect(screen.getByTestId('panel')).toBeInTheDocument();
    });
  });

  describe('Error Notifications', () => {
    it('should display error count', () => {
      mockNotifications = [
        createNotification('1', 'Error 1', NotificationType.ERROR),
        createNotification('2', 'Error 2', NotificationType.ERROR),
      ];

      render(<CanvasValidationIndicator />);

      expect(screen.getByText('2')).toBeInTheDocument();
    });

    it('should have error button with correct aria-label', () => {
      mockNotifications = [createNotification('1', 'Error', NotificationType.ERROR)];

      render(<CanvasValidationIndicator />);

      expect(screen.getByRole('button', {name: '1 Errors'})).toBeInTheDocument();
    });

    it('should open validation panel on error tab when error button clicked', () => {
      mockNotifications = [createNotification('1', 'Error', NotificationType.ERROR)];

      render(<CanvasValidationIndicator />);

      const errorButton = screen.getByRole('button', {name: '1 Errors'});
      fireEvent.click(errorButton);

      expect(mockSetCurrentActiveTab).toHaveBeenCalledWith(0);
      expect(mockSetIsOpenResourcePropertiesPanel).toHaveBeenCalledWith(false);
      expect(mockSetOpenValidationPanel).toHaveBeenCalledWith(true);
    });
  });

  describe('Warning Notifications', () => {
    it('should display warning count', () => {
      mockNotifications = [
        createNotification('1', 'Warning 1', NotificationType.WARNING),
        createNotification('2', 'Warning 2', NotificationType.WARNING),
        createNotification('3', 'Warning 3', NotificationType.WARNING),
      ];

      render(<CanvasValidationIndicator />);

      expect(screen.getByText('3')).toBeInTheDocument();
    });

    it('should have warning button with correct aria-label', () => {
      mockNotifications = [createNotification('1', 'Warning', NotificationType.WARNING)];

      render(<CanvasValidationIndicator />);

      expect(screen.getByRole('button', {name: '1 Warnings'})).toBeInTheDocument();
    });

    it('should open validation panel on warnings tab when warning button clicked', () => {
      mockNotifications = [createNotification('1', 'Warning', NotificationType.WARNING)];

      render(<CanvasValidationIndicator />);

      const warningButton = screen.getByRole('button', {name: '1 Warnings'});
      fireEvent.click(warningButton);

      expect(mockSetCurrentActiveTab).toHaveBeenCalledWith(1);
      expect(mockSetOpenValidationPanel).toHaveBeenCalledWith(true);
    });
  });

  describe('Info Notifications', () => {
    it('should display info count', () => {
      mockNotifications = [createNotification('1', 'Info', NotificationType.INFO)];

      render(<CanvasValidationIndicator />);

      expect(screen.getByText('1')).toBeInTheDocument();
    });

    it('should have info button with correct aria-label', () => {
      mockNotifications = [createNotification('1', 'Info', NotificationType.INFO)];

      render(<CanvasValidationIndicator />);

      expect(screen.getByRole('button', {name: '1 Info'})).toBeInTheDocument();
    });

    it('should open validation panel on info tab when info button clicked', () => {
      mockNotifications = [createNotification('1', 'Info', NotificationType.INFO)];

      render(<CanvasValidationIndicator />);

      const infoButton = screen.getByRole('button', {name: '1 Info'});
      fireEvent.click(infoButton);

      expect(mockSetCurrentActiveTab).toHaveBeenCalledWith(2);
      expect(mockSetOpenValidationPanel).toHaveBeenCalledWith(true);
    });
  });

  describe('Mixed Notifications', () => {
    it('should display counts for all notification types', () => {
      mockNotifications = [
        createNotification('1', 'Error', NotificationType.ERROR),
        createNotification('2', 'Warning 1', NotificationType.WARNING),
        createNotification('3', 'Warning 2', NotificationType.WARNING),
        createNotification('4', 'Info', NotificationType.INFO),
      ];

      render(<CanvasValidationIndicator />);

      expect(screen.getByRole('button', {name: '1 Errors'})).toBeInTheDocument();
      expect(screen.getByRole('button', {name: '2 Warnings'})).toBeInTheDocument();
      expect(screen.getByRole('button', {name: '1 Info'})).toBeInTheDocument();
    });

    it('should only show buttons for notification types that exist', () => {
      mockNotifications = [
        createNotification('1', 'Error', NotificationType.ERROR),
        createNotification('2', 'Info', NotificationType.INFO),
      ];

      render(<CanvasValidationIndicator />);

      expect(screen.getByRole('button', {name: '1 Errors'})).toBeInTheDocument();
      expect(screen.getByRole('button', {name: '1 Info'})).toBeInTheDocument();
      expect(screen.queryByRole('button', {name: /Warnings/})).not.toBeInTheDocument();
    });
  });

  describe('Panel Toggle', () => {
    it('should close validation panel if already open when button clicked', () => {
      mockNotifications = [createNotification('1', 'Error', NotificationType.ERROR)];
      mockOpenValidationPanel = true;

      render(<CanvasValidationIndicator />);

      const errorButton = screen.getByRole('button', {name: '1 Errors'});
      fireEvent.click(errorButton);

      expect(mockSetOpenValidationPanel).toHaveBeenCalledWith(false);
      expect(mockSetCurrentActiveTab).not.toHaveBeenCalled();
    });
  });

  describe('Panel Open CSS Class', () => {
    it('should have panel-open class when validation panel is open', () => {
      mockNotifications = [createNotification('1', 'Error', NotificationType.ERROR)];
      mockOpenValidationPanel = true;

      render(<CanvasValidationIndicator />);

      const panel = screen.getByTestId('panel');
      expect(panel.className).toContain('panel-open');
    });

    it('should have panel-open class when resource properties panel is open', () => {
      mockNotifications = [createNotification('1', 'Error', NotificationType.ERROR)];
      mockIsResourcePropertiesPanelOpen = true;

      render(<CanvasValidationIndicator />);

      const panel = screen.getByTestId('panel');
      expect(panel.className).toContain('panel-open');
    });

    it('should not have panel-open class when no panels are open', () => {
      mockNotifications = [createNotification('1', 'Error', NotificationType.ERROR)];
      mockOpenValidationPanel = false;
      mockIsResourcePropertiesPanelOpen = false;

      render(<CanvasValidationIndicator />);

      const panel = screen.getByTestId('panel');
      expect(panel.className).not.toContain('panel-open');
    });
  });
});
