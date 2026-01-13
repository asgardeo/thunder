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
import type {ReactNode} from 'react';
import ValidationStatusLabels from '../ValidationStatusLabels';
import {ValidationContext, type ValidationContextProps} from '../../../context/ValidationContext';
import FlowBuilderCoreContext, {type FlowBuilderCoreContextProps} from '../../../context/FlowBuilderCoreContext';
import Notification, {NotificationType} from '../../../models/notification';
import {EdgeStyleTypes} from '../../../models/steps';
import {PreviewScreenType} from '../../../models/custom-text-preference';
import {ElementTypes} from '../../../models/elements';
import type {Base} from '../../../models/base';

// Mock react-i18next
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
}));

// Mock the SCSS file
vi.mock('../ValidationStatusLabels.scss', () => ({}));

describe('ValidationStatusLabels', () => {
  const mockSetCurrentActiveTab = vi.fn();
  const mockSetOpenValidationPanel = vi.fn();
  const mockSetIsOpenResourcePropertiesPanel = vi.fn();

  const mockBaseResource: Base = {
    id: 'resource-1',
    type: 'TEXT_INPUT',
    category: 'FIELD',
    resourceType: 'ELEMENT',
    version: '1.0.0',
    deprecated: false,
    deletable: true,
    display: {
      label: 'Test Resource',
      image: '',
      showOnResourcePanel: false,
    },
    config: {
      field: {name: '', type: ElementTypes},
      styles: {},
    },
  };

  const defaultValidationContext: ValidationContextProps = {
    isValid: true,
    notifications: [],
    getNotification: vi.fn(),
    validationConfig: {
      isOTPValidationEnabled: false,
      isRecoveryFactorValidationEnabled: false,
      isPasswordExecutorValidationEnabled: false,
    },
    setCurrentActiveTab: mockSetCurrentActiveTab,
    openValidationPanel: false,
    setOpenValidationPanel: mockSetOpenValidationPanel,
  };

  const defaultFlowBuilderContext: FlowBuilderCoreContextProps = {
    lastInteractedResource: mockBaseResource,
    lastInteractedStepId: 'step-1',
    ResourceProperties: () => null,
    resourcePropertiesPanelHeading: 'Test Panel Heading',
    primaryI18nScreen: PreviewScreenType.LOGIN,
    isResourcePanelOpen: true,
    isResourcePropertiesPanelOpen: false,
    isVersionHistoryPanelOpen: false,
    ElementFactory: () => null,
    onResourceDropOnCanvas: vi.fn(),
    selectedAttributes: {},
    setLastInteractedResource: vi.fn(),
    setLastInteractedStepId: vi.fn(),
    setResourcePropertiesPanelHeading: vi.fn(),
    setIsResourcePanelOpen: vi.fn(),
    setIsOpenResourcePropertiesPanel: mockSetIsOpenResourcePropertiesPanel,
    registerCloseValidationPanel: vi.fn(),
    setIsVersionHistoryPanelOpen: vi.fn(),
    setSelectedAttributes: vi.fn(),
    flowCompletionConfigs: {},
    setFlowCompletionConfigs: vi.fn(),
    flowNodeTypes: {},
    flowEdgeTypes: {},
    setFlowNodeTypes: vi.fn(),
    setFlowEdgeTypes: vi.fn(),
    isVerboseMode: false,
    setIsVerboseMode: vi.fn(),
    edgeStyle: EdgeStyleTypes.SmoothStep,
    setEdgeStyle: vi.fn(),
  };

  const createWrapper = (
    validationContext: ValidationContextProps = defaultValidationContext,
    flowBuilderContext: FlowBuilderCoreContextProps = defaultFlowBuilderContext,
  ) => {
    function Wrapper({children}: {children: ReactNode}) {
      return (
        <FlowBuilderCoreContext.Provider value={flowBuilderContext}>
          <ValidationContext.Provider value={validationContext}>{children}</ValidationContext.Provider>
        </FlowBuilderCoreContext.Provider>
      );
    }
    return Wrapper;
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('should render the notification bell button', () => {
      render(<ValidationStatusLabels />, {wrapper: createWrapper()});

      expect(screen.getByRole('button')).toBeInTheDocument();
    });

    it('should render with tooltip', () => {
      render(<ValidationStatusLabels />, {wrapper: createWrapper()});

      // The button should have the tooltip (translation key)
      expect(screen.getByRole('button')).toBeInTheDocument();
    });
  });

  describe('Badge Display - No Notifications', () => {
    it('should display badge with 0 when no notifications exist', () => {
      render(<ValidationStatusLabels />, {wrapper: createWrapper()});

      // Badge with 0 content
      const badge = document.querySelector('.MuiBadge-badge');
      expect(badge).toBeInTheDocument();
    });
  });

  describe('Badge Display - Error Notifications', () => {
    it('should display error count in badge when errors exist', () => {
      const errorNotification1 = new Notification('error-1', 'Error 1', NotificationType.ERROR);
      const errorNotification2 = new Notification('error-2', 'Error 2', NotificationType.ERROR);

      const contextWithErrors: ValidationContextProps = {
        ...defaultValidationContext,
        notifications: [errorNotification1, errorNotification2],
      };

      render(<ValidationStatusLabels />, {wrapper: createWrapper(contextWithErrors)});

      // Check that badge exists
      const badge = document.querySelector('.MuiBadge-badge');
      expect(badge).toBeInTheDocument();
      expect(badge?.textContent).toBe('2');
    });

    it('should show error badge color when errors exist', () => {
      const errorNotification = new Notification('error-1', 'Error 1', NotificationType.ERROR);

      const contextWithError: ValidationContextProps = {
        ...defaultValidationContext,
        notifications: [errorNotification],
      };

      render(<ValidationStatusLabels />, {wrapper: createWrapper(contextWithError)});

      const badge = document.querySelector('.MuiBadge-colorError');
      expect(badge).toBeInTheDocument();
    });
  });

  describe('Badge Display - Warning Notifications', () => {
    it('should display warning count in badge when only warnings exist', () => {
      const warningNotification1 = new Notification('warning-1', 'Warning 1', NotificationType.WARNING);
      const warningNotification2 = new Notification('warning-2', 'Warning 2', NotificationType.WARNING);
      const warningNotification3 = new Notification('warning-3', 'Warning 3', NotificationType.WARNING);

      const contextWithWarnings: ValidationContextProps = {
        ...defaultValidationContext,
        notifications: [warningNotification1, warningNotification2, warningNotification3],
      };

      render(<ValidationStatusLabels />, {wrapper: createWrapper(contextWithWarnings)});

      const badge = document.querySelector('.MuiBadge-badge');
      expect(badge?.textContent).toBe('3');
    });

    it('should show warning badge color when only warnings exist', () => {
      const warningNotification = new Notification('warning-1', 'Warning 1', NotificationType.WARNING);

      const contextWithWarning: ValidationContextProps = {
        ...defaultValidationContext,
        notifications: [warningNotification],
      };

      render(<ValidationStatusLabels />, {wrapper: createWrapper(contextWithWarning)});

      const badge = document.querySelector('.MuiBadge-colorWarning');
      expect(badge).toBeInTheDocument();
    });
  });

  describe('Badge Display - Info Notifications', () => {
    it('should display info count in badge when only info notifications exist', () => {
      const infoNotification = new Notification('info-1', 'Info 1', NotificationType.INFO);

      const contextWithInfo: ValidationContextProps = {
        ...defaultValidationContext,
        notifications: [infoNotification],
      };

      render(<ValidationStatusLabels />, {wrapper: createWrapper(contextWithInfo)});

      const badge = document.querySelector('.MuiBadge-badge');
      expect(badge?.textContent).toBe('1');
    });

    it('should show info badge color when only info notifications exist', () => {
      const infoNotification = new Notification('info-1', 'Info 1', NotificationType.INFO);

      const contextWithInfo: ValidationContextProps = {
        ...defaultValidationContext,
        notifications: [infoNotification],
      };

      render(<ValidationStatusLabels />, {wrapper: createWrapper(contextWithInfo)});

      const badge = document.querySelector('.MuiBadge-colorInfo');
      expect(badge).toBeInTheDocument();
    });
  });

  describe('Badge Priority', () => {
    it('should prioritize error count over warning count', () => {
      const errorNotification = new Notification('error-1', 'Error 1', NotificationType.ERROR);
      const warningNotification1 = new Notification('warning-1', 'Warning 1', NotificationType.WARNING);
      const warningNotification2 = new Notification('warning-2', 'Warning 2', NotificationType.WARNING);

      const contextWithMixed: ValidationContextProps = {
        ...defaultValidationContext,
        notifications: [errorNotification, warningNotification1, warningNotification2],
      };

      render(<ValidationStatusLabels />, {wrapper: createWrapper(contextWithMixed)});

      const badge = document.querySelector('.MuiBadge-badge');
      expect(badge?.textContent).toBe('1'); // Error count
      expect(document.querySelector('.MuiBadge-colorError')).toBeInTheDocument();
    });

    it('should prioritize warning count over info count', () => {
      const warningNotification = new Notification('warning-1', 'Warning 1', NotificationType.WARNING);
      const infoNotification1 = new Notification('info-1', 'Info 1', NotificationType.INFO);
      const infoNotification2 = new Notification('info-2', 'Info 2', NotificationType.INFO);

      const contextWithMixed: ValidationContextProps = {
        ...defaultValidationContext,
        notifications: [warningNotification, infoNotification1, infoNotification2],
      };

      render(<ValidationStatusLabels />, {wrapper: createWrapper(contextWithMixed)});

      const badge = document.querySelector('.MuiBadge-badge');
      expect(badge?.textContent).toBe('1'); // Warning count
      expect(document.querySelector('.MuiBadge-colorWarning')).toBeInTheDocument();
    });
  });

  describe('Click Behavior - Opening Panel', () => {
    it('should open validation panel on click when panel is closed', () => {
      render(<ValidationStatusLabels />, {wrapper: createWrapper()});

      const button = screen.getByRole('button');
      fireEvent.click(button);

      expect(mockSetOpenValidationPanel).toHaveBeenCalledWith(true);
    });

    it('should close resource properties panel when opening validation panel', () => {
      render(<ValidationStatusLabels />, {wrapper: createWrapper()});

      const button = screen.getByRole('button');
      fireEvent.click(button);

      expect(mockSetIsOpenResourcePropertiesPanel).toHaveBeenCalledWith(false);
    });

    it('should set active tab to 0 (errors) when errors exist', () => {
      const errorNotification = new Notification('error-1', 'Error 1', NotificationType.ERROR);

      const contextWithError: ValidationContextProps = {
        ...defaultValidationContext,
        notifications: [errorNotification],
      };

      render(<ValidationStatusLabels />, {wrapper: createWrapper(contextWithError)});

      const button = screen.getByRole('button');
      fireEvent.click(button);

      expect(mockSetCurrentActiveTab).toHaveBeenCalledWith(0);
    });

    it('should set active tab to 1 (warnings) when only warnings exist', () => {
      const warningNotification = new Notification('warning-1', 'Warning 1', NotificationType.WARNING);

      const contextWithWarning: ValidationContextProps = {
        ...defaultValidationContext,
        notifications: [warningNotification],
      };

      render(<ValidationStatusLabels />, {wrapper: createWrapper(contextWithWarning)});

      const button = screen.getByRole('button');
      fireEvent.click(button);

      expect(mockSetCurrentActiveTab).toHaveBeenCalledWith(1);
    });

    it('should set active tab to 0 when no notifications exist', () => {
      render(<ValidationStatusLabels />, {wrapper: createWrapper()});

      const button = screen.getByRole('button');
      fireEvent.click(button);

      expect(mockSetCurrentActiveTab).toHaveBeenCalledWith(0);
    });
  });

  describe('Click Behavior - Closing Panel', () => {
    it('should close validation panel on click when panel is already open', () => {
      const contextWithOpenPanel: ValidationContextProps = {
        ...defaultValidationContext,
        openValidationPanel: true,
      };

      render(<ValidationStatusLabels />, {wrapper: createWrapper(contextWithOpenPanel)});

      const button = screen.getByRole('button');
      fireEvent.click(button);

      expect(mockSetOpenValidationPanel).toHaveBeenCalledWith(false);
    });

    it('should not set active tab when closing panel', () => {
      const contextWithOpenPanel: ValidationContextProps = {
        ...defaultValidationContext,
        openValidationPanel: true,
      };

      render(<ValidationStatusLabels />, {wrapper: createWrapper(contextWithOpenPanel)});

      const button = screen.getByRole('button');
      fireEvent.click(button);

      expect(mockSetCurrentActiveTab).not.toHaveBeenCalled();
    });

    it('should not close resource properties panel when closing validation panel', () => {
      const contextWithOpenPanel: ValidationContextProps = {
        ...defaultValidationContext,
        openValidationPanel: true,
      };

      render(<ValidationStatusLabels />, {wrapper: createWrapper(contextWithOpenPanel)});

      const button = screen.getByRole('button');
      fireEvent.click(button);

      expect(mockSetIsOpenResourcePropertiesPanel).not.toHaveBeenCalled();
    });
  });

  describe('Empty Notifications Array', () => {
    it('should handle empty notifications array gracefully', () => {
      const contextWithEmptyNotifications: ValidationContextProps = {
        ...defaultValidationContext,
        notifications: [],
      };

      render(<ValidationStatusLabels />, {wrapper: createWrapper(contextWithEmptyNotifications)});

      const badge = document.querySelector('.MuiBadge-badge');
      expect(badge).toBeInTheDocument();
    });
  });

  describe('Null/Undefined Notifications', () => {
    it('should handle undefined notifications gracefully', () => {
      const contextWithUndefinedNotifications: ValidationContextProps = {
        ...defaultValidationContext,
        notifications: undefined as unknown as Notification[],
      };

      render(<ValidationStatusLabels />, {wrapper: createWrapper(contextWithUndefinedNotifications)});

      // Should not throw and should render
      expect(screen.getByRole('button')).toBeInTheDocument();
    });
  });
});
