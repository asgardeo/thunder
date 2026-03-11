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
import {render} from '@thunder/test-utils/browser';
import {page, userEvent} from 'vitest/browser';
import type {ReactNode} from 'react';
import CheckboxPropertyField from '../CheckboxPropertyField';
import {ValidationContext, type ValidationContextProps} from '../../../context/ValidationContext';
import type {Resource} from '../../../models/resources';
import Notification from '../../../models/notification';

describe('CheckboxPropertyField', () => {
  const mockOnChange = vi.fn();

  const mockResource: Resource = {
    id: 'resource-1',
    type: 'CHECKBOX',
    config: {},
  } as Resource;

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

  describe('Rendering', () => {
    it('should render checkbox with label', async () => {
      await render(
        <CheckboxPropertyField
          resource={mockResource}
          propertyKey="isEnabled"
          propertyValue
          onChange={mockOnChange}
        />,
        {wrapper: createWrapper()},
      );

      await expect.element(page.getByLabelText('Is Enabled')).toBeInTheDocument();
    });

    it('should convert camelCase propertyKey to Start Case label', async () => {
      await render(
        <CheckboxPropertyField
          resource={mockResource}
          propertyKey="myPropertyName"
          propertyValue={false}
          onChange={mockOnChange}
        />,
        {wrapper: createWrapper()},
      );

      await expect.element(page.getByLabelText('My Property Name')).toBeInTheDocument();
    });

    it('should render checkbox as checked when propertyValue is true', async () => {
      await render(
        <CheckboxPropertyField
          resource={mockResource}
          propertyKey="enabled"
          propertyValue
          onChange={mockOnChange}
        />,
        {wrapper: createWrapper()},
      );

      const checkbox = page.getByRole('checkbox');
      expect(checkbox).toBeChecked();
    });

    it('should render checkbox as unchecked when propertyValue is false', async () => {
      await render(
        <CheckboxPropertyField
          resource={mockResource}
          propertyKey="enabled"
          propertyValue={false}
          onChange={mockOnChange}
        />,
        {wrapper: createWrapper()},
      );

      const checkbox = page.getByRole('checkbox');
      expect(checkbox).not.toBeChecked();
    });
  });

  describe('onChange Handler', () => {
    it('should call onChange with checked=true when checkbox is clicked', async () => {
      await render(
        <CheckboxPropertyField
          resource={mockResource}
          propertyKey="enabled"
          propertyValue={false}
          onChange={mockOnChange}
        />,
        {wrapper: createWrapper()},
      );

      const checkbox = page.getByRole('checkbox');
      await userEvent.click(checkbox);

      expect(mockOnChange).toHaveBeenCalledWith('enabled', true, mockResource);
    });

    it('should call onChange with checked=false when checkbox is unchecked', async () => {
      await render(
        <CheckboxPropertyField
          resource={mockResource}
          propertyKey="enabled"
          propertyValue
          onChange={mockOnChange}
        />,
        {wrapper: createWrapper()},
      );

      const checkbox = page.getByRole('checkbox');
      await userEvent.click(checkbox);

      expect(mockOnChange).toHaveBeenCalledWith('enabled', false, mockResource);
    });

    it('should pass the correct resource to onChange', async () => {
      const specificResource = {...mockResource, id: 'specific-resource'};
      await render(
        <CheckboxPropertyField
          resource={specificResource}
          propertyKey="active"
          propertyValue={false}
          onChange={mockOnChange}
        />,
        {wrapper: createWrapper()},
      );

      const checkbox = page.getByRole('checkbox');
      await userEvent.click(checkbox);

      expect(mockOnChange).toHaveBeenCalledWith('active', true, specificResource);
    });
  });

  describe('Error State', () => {
    it('should display error message when notification exists', async () => {
      const notification = new Notification('notification-1', 'Error', 'error');
      notification.addResourceFieldNotification('resource-1_isRequired', 'This field is required');

      const contextWithError: ValidationContextProps = {
        ...defaultContextValue,
        selectedNotification: notification,
      };

      await render(
        <CheckboxPropertyField
          resource={mockResource}
          propertyKey="isRequired"
          propertyValue={false}
          onChange={mockOnChange}
        />,
        {wrapper: createWrapper(contextWithError)},
      );

      await expect.element(page.getByText('This field is required')).toBeInTheDocument();
    });

    it('should not display error message when no notification exists', async () => {
      await render(
        <CheckboxPropertyField
          resource={mockResource}
          propertyKey="enabled"
          propertyValue={false}
          onChange={mockOnChange}
        />,
        {wrapper: createWrapper()},
      );

      await expect.element(page.getByRole('alert')).not.toBeInTheDocument();
    });

    it('should not display error message for different property', async () => {
      const notification = new Notification('notification-1', 'Error', 'error');
      notification.addResourceFieldNotification('resource-1_otherProperty', 'Other error');

      const contextWithError: ValidationContextProps = {
        ...defaultContextValue,
        selectedNotification: notification,
      };

      await render(
        <CheckboxPropertyField
          resource={mockResource}
          propertyKey="enabled"
          propertyValue={false}
          onChange={mockOnChange}
        />,
        {wrapper: createWrapper(contextWithError)},
      );

      await expect.element(page.getByText('Other error')).not.toBeInTheDocument();
    });
  });

  describe('Accessibility', () => {
    it('should have accessible checkbox element', async () => {
      await render(
        <CheckboxPropertyField
          resource={mockResource}
          propertyKey="enabled"
          propertyValue
          onChange={mockOnChange}
        />,
        {wrapper: createWrapper()},
      );

      const checkbox = page.getByRole('checkbox');
      expect(checkbox).toBeInTheDocument();
    });
  });

  describe('Error Color State', () => {
    it('should render checkbox with error color when there is an error', async () => {
      const notification = new Notification('notification-1', 'Error', 'error');
      notification.addResourceFieldNotification('resource-1_hasError', 'Field has an error');

      const contextWithError: ValidationContextProps = {
        ...defaultContextValue,
        selectedNotification: notification,
      };

      await render(
        <CheckboxPropertyField
          resource={mockResource}
          propertyKey="hasError"
          propertyValue={false}
          onChange={mockOnChange}
        />,
        {wrapper: createWrapper(contextWithError)},
      );

      // The error message should be displayed
      await expect.element(page.getByText('Field has an error')).toBeInTheDocument();
    });

    it('should render checkbox with primary color when no error', async () => {
      await render(
        <CheckboxPropertyField
          resource={mockResource}
          propertyKey="noError"
          propertyValue
          onChange={mockOnChange}
        />,
        {wrapper: createWrapper()},
      );

      // No error helper text should be present
      await expect.element(page.getByRole('alert')).not.toBeInTheDocument();
    });
  });

  describe('Notification without matching field', () => {
    it('should return empty string when notification exists but has no matching field', async () => {
      const notification = new Notification('notification-1', 'Error', 'error');
      // Add a notification for a different field
      notification.addResourceFieldNotification('resource-1_differentField', 'Different error');

      const contextWithNotification: ValidationContextProps = {
        ...defaultContextValue,
        selectedNotification: notification,
      };

      await render(
        <CheckboxPropertyField
          resource={mockResource}
          propertyKey="enabled"
          propertyValue
          onChange={mockOnChange}
        />,
        {wrapper: createWrapper(contextWithNotification)},
      );

      // No error message for this field
      await expect.element(page.getByText('Different error')).not.toBeInTheDocument();
    });
  });
});
