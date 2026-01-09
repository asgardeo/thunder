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
import ValidationErrorBoundary from '../ValidationErrorBoundary';
import {ValidationContext, type ValidationContextProps} from '../../../context/ValidationContext';
import type {Resource} from '../../../models/resources';
import Notification, {NotificationType} from '../../../models/notification';

// Mock the SCSS file
vi.mock('../ValidationErrorBoundary.scss', () => ({}));

describe('ValidationErrorBoundary', () => {
  const mockResource: Resource = {
    id: 'resource-1',
    type: 'TEXT_INPUT',
    config: {},
  } as Resource;

  // Helper to create a notification with a resource
  const createNotificationWithResource = (
    id: string,
    message: string,
    type: NotificationType,
    resourceId: string,
  ): Notification => {
    const notification = new Notification(id, message, type);
    // addResource expects a Resource object, so we create one
    notification.addResource({id: resourceId, type: 'TEXT_INPUT', config: {}} as Resource);
    return notification;
  };

  const defaultContextValue: ValidationContextProps = {
    isValid: true,
    notifications: [],
    getNotification: vi.fn(),
    validationConfig: {
      isOTPValidationEnabled: false,
      isRecoveryFactorValidationEnabled: false,
      isPasswordExecutorValidationEnabled: false,
    },
  };

  const createWrapper = (contextValue: ValidationContextProps = defaultContextValue) => {
    function Wrapper({children}: {children: ReactNode}) {
      return <ValidationContext.Provider value={contextValue}>{children}</ValidationContext.Provider>;
    }
    return Wrapper;
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering Without Notifications', () => {
    it('should render children without error boundary styling when no notifications exist', () => {
      render(
        <ValidationErrorBoundary resource={mockResource}>
          <div data-testid="child-content">Child Content</div>
        </ValidationErrorBoundary>,
        {wrapper: createWrapper()},
      );

      expect(screen.getByTestId('child-content')).toBeInTheDocument();
      expect(screen.getByTestId('child-content').textContent).toBe('Child Content');
    });

    it('should not show alert icon when no notifications exist', () => {
      render(
        <ValidationErrorBoundary resource={mockResource}>
          <div>Content</div>
        </ValidationErrorBoundary>,
        {wrapper: createWrapper()},
      );

      // CircleAlertIcon should not be rendered
      expect(document.querySelector('.circle-alert-icon')).not.toBeInTheDocument();
    });
  });

  describe('Error Notifications', () => {
    it('should show error styling when error notification exists for resource', () => {
      const errorNotification = createNotificationWithResource(
        'notification-1',
        'Error Message',
        NotificationType.ERROR,
        'resource-1',
      );

      const contextWithError: ValidationContextProps = {
        ...defaultContextValue,
        notifications: [errorNotification],
      };

      const {container} = render(
        <ValidationErrorBoundary resource={mockResource}>
          <div>Content</div>
        </ValidationErrorBoundary>,
        {wrapper: createWrapper(contextWithError)},
      );

      expect(container.querySelector('.validation-error-boundary')).toBeInTheDocument();
      expect(container.querySelector('.error')).toBeInTheDocument();
    });

    it('should show alert icon for error notification', () => {
      const errorNotification = createNotificationWithResource(
        'notification-1',
        'Error Message',
        NotificationType.ERROR,
        'resource-1',
      );

      const contextWithError: ValidationContextProps = {
        ...defaultContextValue,
        notifications: [errorNotification],
      };

      render(
        <ValidationErrorBoundary resource={mockResource}>
          <div>Content</div>
        </ValidationErrorBoundary>,
        {wrapper: createWrapper(contextWithError)},
      );

      expect(document.querySelector('.circle-alert-icon')).toBeInTheDocument();
    });
  });

  describe('Warning Notifications', () => {
    it('should show warning styling when warning notification exists for resource', () => {
      const warningNotification = createNotificationWithResource(
        'notification-1',
        'Warning Message',
        NotificationType.WARNING,
        'resource-1',
      );

      const contextWithWarning: ValidationContextProps = {
        ...defaultContextValue,
        notifications: [warningNotification],
      };

      const {container} = render(
        <ValidationErrorBoundary resource={mockResource}>
          <div>Content</div>
        </ValidationErrorBoundary>,
        {wrapper: createWrapper(contextWithWarning)},
      );

      expect(container.querySelector('.validation-error-boundary')).toBeInTheDocument();
      expect(container.querySelector('.warning')).toBeInTheDocument();
    });
  });

  describe('Info Notifications', () => {
    it('should show info styling when info notification exists for resource', () => {
      const infoNotification = createNotificationWithResource(
        'notification-1',
        'Info Message',
        NotificationType.INFO,
        'resource-1',
      );

      const contextWithInfo: ValidationContextProps = {
        ...defaultContextValue,
        notifications: [infoNotification],
      };

      const {container} = render(
        <ValidationErrorBoundary resource={mockResource}>
          <div>Content</div>
        </ValidationErrorBoundary>,
        {wrapper: createWrapper(contextWithInfo)},
      );

      expect(container.querySelector('.validation-error-boundary')).toBeInTheDocument();
      expect(container.querySelector('.info')).toBeInTheDocument();
    });
  });

  describe('Notification Priority', () => {
    it('should prioritize error over warning notification', () => {
      const errorNotification = createNotificationWithResource(
        'error-1',
        'Error Message',
        NotificationType.ERROR,
        'resource-1',
      );

      const warningNotification = createNotificationWithResource(
        'warning-1',
        'Warning Message',
        NotificationType.WARNING,
        'resource-1',
      );

      const contextWithBoth: ValidationContextProps = {
        ...defaultContextValue,
        notifications: [warningNotification, errorNotification],
      };

      const {container} = render(
        <ValidationErrorBoundary resource={mockResource}>
          <div>Content</div>
        </ValidationErrorBoundary>,
        {wrapper: createWrapper(contextWithBoth)},
      );

      expect(container.querySelector('.error')).toBeInTheDocument();
      expect(container.querySelector('.warning')).not.toBeInTheDocument();
    });

    it('should prioritize warning over info notification', () => {
      const warningNotification = createNotificationWithResource(
        'warning-1',
        'Warning Message',
        NotificationType.WARNING,
        'resource-1',
      );

      const infoNotification = createNotificationWithResource(
        'info-1',
        'Info Message',
        NotificationType.INFO,
        'resource-1',
      );

      const contextWithBoth: ValidationContextProps = {
        ...defaultContextValue,
        notifications: [infoNotification, warningNotification],
      };

      const {container} = render(
        <ValidationErrorBoundary resource={mockResource}>
          <div>Content</div>
        </ValidationErrorBoundary>,
        {wrapper: createWrapper(contextWithBoth)},
      );

      expect(container.querySelector('.warning')).toBeInTheDocument();
      expect(container.querySelector('.info')).not.toBeInTheDocument();
    });
  });

  describe('Hover Behavior', () => {
    it('should hide error boundary on hover when disableErrorBoundaryOnHover is true', () => {
      const errorNotification = createNotificationWithResource(
        'notification-1',
        'Error Message',
        NotificationType.ERROR,
        'resource-1',
      );

      const contextWithError: ValidationContextProps = {
        ...defaultContextValue,
        notifications: [errorNotification],
      };

      const {container} = render(
        <ValidationErrorBoundary resource={mockResource} disableErrorBoundaryOnHover>
          <div>Content</div>
        </ValidationErrorBoundary>,
        {wrapper: createWrapper(contextWithError)},
      );

      const boundaryDiv = container.firstChild as HTMLElement;

      // Initially should not have 'active' class
      expect(boundaryDiv).not.toHaveClass('active');

      // Simulate mouse over
      fireEvent.mouseOver(boundaryDiv);

      // Should now have 'active' class
      expect(boundaryDiv).toHaveClass('active');
    });

    it('should restore error boundary when mouse leaves after hover', () => {
      const errorNotification = createNotificationWithResource(
        'notification-1',
        'Error Message',
        NotificationType.ERROR,
        'resource-1',
      );

      const contextWithError: ValidationContextProps = {
        ...defaultContextValue,
        notifications: [errorNotification],
      };

      const {container} = render(
        <ValidationErrorBoundary resource={mockResource} disableErrorBoundaryOnHover>
          <div>Content</div>
        </ValidationErrorBoundary>,
        {wrapper: createWrapper(contextWithError)},
      );

      const boundaryDiv = container.firstChild as HTMLElement;

      // Mouse over
      fireEvent.mouseOver(boundaryDiv);
      expect(boundaryDiv).toHaveClass('active');

      // Mouse out
      fireEvent.mouseOut(boundaryDiv);
      expect(boundaryDiv).not.toHaveClass('active');
    });

    it('should handle focus/blur events for accessibility', () => {
      const errorNotification = createNotificationWithResource(
        'notification-1',
        'Error Message',
        NotificationType.ERROR,
        'resource-1',
      );

      const contextWithError: ValidationContextProps = {
        ...defaultContextValue,
        notifications: [errorNotification],
      };

      const {container} = render(
        <ValidationErrorBoundary resource={mockResource} disableErrorBoundaryOnHover>
          <div>Content</div>
        </ValidationErrorBoundary>,
        {wrapper: createWrapper(contextWithError)},
      );

      const boundaryDiv = container.firstChild as HTMLElement;

      // Focus
      fireEvent.focus(boundaryDiv);
      expect(boundaryDiv).toHaveClass('active');

      // Blur
      fireEvent.blur(boundaryDiv);
      expect(boundaryDiv).not.toHaveClass('active');
    });

    it('should not activate on hover when disableErrorBoundaryOnHover is false', () => {
      const errorNotification = createNotificationWithResource(
        'notification-1',
        'Error Message',
        NotificationType.ERROR,
        'resource-1',
      );

      const contextWithError: ValidationContextProps = {
        ...defaultContextValue,
        notifications: [errorNotification],
      };

      const {container} = render(
        <ValidationErrorBoundary resource={mockResource} disableErrorBoundaryOnHover={false}>
          <div>Content</div>
        </ValidationErrorBoundary>,
        {wrapper: createWrapper(contextWithError)},
      );

      const boundaryDiv = container.firstChild as HTMLElement;

      fireEvent.mouseOver(boundaryDiv);
      expect(boundaryDiv).not.toHaveClass('active');
    });
  });

  describe('Alert Icon Visibility', () => {
    it('should hide alert icon when active and disableErrorBoundaryOnHover is true', () => {
      const errorNotification = createNotificationWithResource(
        'notification-1',
        'Error Message',
        NotificationType.ERROR,
        'resource-1',
      );

      const contextWithError: ValidationContextProps = {
        ...defaultContextValue,
        notifications: [errorNotification],
      };

      const {container} = render(
        <ValidationErrorBoundary resource={mockResource} disableErrorBoundaryOnHover>
          <div>Content</div>
        </ValidationErrorBoundary>,
        {wrapper: createWrapper(contextWithError)},
      );

      // Alert icon should be visible initially
      expect(document.querySelector('.circle-alert-icon')).toBeInTheDocument();

      // Mouse over
      const boundaryDiv = container.firstChild as HTMLElement;
      fireEvent.mouseOver(boundaryDiv);

      // Alert icon should be hidden
      expect(document.querySelector('.circle-alert-icon')).not.toBeInTheDocument();
    });
  });

  describe('Padded Class', () => {
    it('should add padded class when notification exists and disableErrorBoundaryOnHover is false', () => {
      const errorNotification = createNotificationWithResource(
        'notification-1',
        'Error Message',
        NotificationType.ERROR,
        'resource-1',
      );

      const contextWithError: ValidationContextProps = {
        ...defaultContextValue,
        notifications: [errorNotification],
      };

      const {container} = render(
        <ValidationErrorBoundary resource={mockResource} disableErrorBoundaryOnHover={false}>
          <div>Content</div>
        </ValidationErrorBoundary>,
        {wrapper: createWrapper(contextWithError)},
      );

      expect(container.querySelector('.padded')).toBeInTheDocument();
    });

    it('should not add padded class when disableErrorBoundaryOnHover is true', () => {
      const errorNotification = createNotificationWithResource(
        'notification-1',
        'Error Message',
        NotificationType.ERROR,
        'resource-1',
      );

      const contextWithError: ValidationContextProps = {
        ...defaultContextValue,
        notifications: [errorNotification],
      };

      const {container} = render(
        <ValidationErrorBoundary resource={mockResource} disableErrorBoundaryOnHover>
          <div>Content</div>
        </ValidationErrorBoundary>,
        {wrapper: createWrapper(contextWithError)},
      );

      expect(container.querySelector('.padded')).not.toBeInTheDocument();
    });
  });

  describe('Different Resources', () => {
    it('should not show error boundary for notifications targeting different resource', () => {
      const errorNotification = createNotificationWithResource(
        'notification-1',
        'Error Message',
        NotificationType.ERROR,
        'different-resource-id',
      );

      const contextWithError: ValidationContextProps = {
        ...defaultContextValue,
        notifications: [errorNotification],
      };

      const {container} = render(
        <ValidationErrorBoundary resource={mockResource}>
          <div>Content</div>
        </ValidationErrorBoundary>,
        {wrapper: createWrapper(contextWithError)},
      );

      expect(container.querySelector('.validation-error-boundary')).not.toBeInTheDocument();
    });
  });

  describe('Default Props', () => {
    it('should default disableErrorBoundaryOnHover to false', () => {
      const errorNotification = createNotificationWithResource(
        'notification-1',
        'Error Message',
        NotificationType.ERROR,
        'resource-1',
      );

      const contextWithError: ValidationContextProps = {
        ...defaultContextValue,
        notifications: [errorNotification],
      };

      const {container} = render(
        <ValidationErrorBoundary resource={mockResource}>
          <div>Content</div>
        </ValidationErrorBoundary>,
        {wrapper: createWrapper(contextWithError)},
      );

      // Should have padded class (which means disableErrorBoundaryOnHover is false)
      expect(container.querySelector('.padded')).toBeInTheDocument();
    });

    it('should render with null children by default', () => {
      const {container} = render(<ValidationErrorBoundary resource={mockResource} />, {wrapper: createWrapper()});

      expect(container.firstChild).toBeInTheDocument();
    });
  });
});
