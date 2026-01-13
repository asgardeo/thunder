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
import TextPropertyField from '../TextPropertyField';
import {ValidationContext, type ValidationContextProps} from '../../../context/ValidationContext';
import type {Resource} from '../../../models/resources';
import Notification from '../../../models/notification';

// Mock react-i18next
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, options?: {propertyName?: string}) =>
      options?.propertyName ? `Enter ${options.propertyName}` : key,
  }),
}));

describe('TextPropertyField', () => {
  const mockOnChange = vi.fn();

  const mockResource: Resource = {
    id: 'resource-1',
    type: 'TEXT',
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
    it('should render text field with label', () => {
      render(
        <TextPropertyField
          resource={mockResource}
          propertyKey="userName"
          propertyValue="test"
          onChange={mockOnChange}
        />,
        {wrapper: createWrapper()},
      );

      // Check the label text is rendered
      expect(screen.getByText('User Name')).toBeInTheDocument();
      expect(screen.getByRole('textbox')).toBeInTheDocument();
    });

    it('should convert camelCase propertyKey to Start Case label', () => {
      render(
        <TextPropertyField
          resource={mockResource}
          propertyKey="myPropertyName"
          propertyValue=""
          onChange={mockOnChange}
        />,
        {wrapper: createWrapper()},
      );

      // Check the label text is rendered in Start Case
      expect(screen.getByText('My Property Name')).toBeInTheDocument();
    });

    it('should render text field with default value', () => {
      render(
        <TextPropertyField
          resource={mockResource}
          propertyKey="title"
          propertyValue="Hello World"
          onChange={mockOnChange}
        />,
        {wrapper: createWrapper()},
      );

      const textField = screen.getByRole('textbox');
      expect(textField).toHaveValue('Hello World');
    });

    it('should render empty text field when value is empty', () => {
      render(
        <TextPropertyField resource={mockResource} propertyKey="title" propertyValue="" onChange={mockOnChange} />,
        {wrapper: createWrapper()},
      );

      const textField = screen.getByRole('textbox');
      expect(textField).toHaveValue('');
    });
  });

  describe('onChange Handler', () => {
    it('should call onChange when text is entered', () => {
      render(
        <TextPropertyField resource={mockResource} propertyKey="label" propertyValue="" onChange={mockOnChange} />,
        {wrapper: createWrapper()},
      );

      const textField = screen.getByRole('textbox');
      fireEvent.change(textField, {target: {value: 'New Value'}});

      expect(mockOnChange).toHaveBeenCalledWith('label', 'New Value', mockResource);
    });

    it('should pass the correct resource to onChange', () => {
      const specificResource = {...mockResource, id: 'specific-resource'};
      render(
        <TextPropertyField
          resource={specificResource}
          propertyKey="description"
          propertyValue=""
          onChange={mockOnChange}
        />,
        {wrapper: createWrapper()},
      );

      const textField = screen.getByRole('textbox');
      fireEvent.change(textField, {target: {value: 'Test'}});

      expect(mockOnChange).toHaveBeenCalledWith('description', 'Test', specificResource);
    });
  });

  describe('Error State', () => {
    it('should display error message when notification exists', () => {
      const notification = new Notification('notification-1', 'Error', 'error');
      notification.addResourceFieldNotification('resource-1_title', 'Title is required');

      const contextWithError: ValidationContextProps = {
        ...defaultContextValue,
        selectedNotification: notification,
      };

      render(
        <TextPropertyField resource={mockResource} propertyKey="title" propertyValue="" onChange={mockOnChange} />,
        {wrapper: createWrapper(contextWithError)},
      );

      expect(screen.getByText('Title is required')).toBeInTheDocument();
    });

    it('should not display error message when no notification exists', () => {
      render(
        <TextPropertyField resource={mockResource} propertyKey="title" propertyValue="" onChange={mockOnChange} />,
        {wrapper: createWrapper()},
      );

      expect(screen.queryByRole('alert')).not.toBeInTheDocument();
    });

    it('should not display error message for different property', () => {
      const notification = new Notification('notification-1', 'Error', 'error');
      notification.addResourceFieldNotification('resource-1_otherProperty', 'Other error');

      const contextWithError: ValidationContextProps = {
        ...defaultContextValue,
        selectedNotification: notification,
      };

      render(
        <TextPropertyField resource={mockResource} propertyKey="title" propertyValue="" onChange={mockOnChange} />,
        {wrapper: createWrapper(contextWithError)},
      );

      expect(screen.queryByText('Other error')).not.toBeInTheDocument();
    });
  });

  describe('I18n Pattern', () => {
    it('should display i18n placeholder when value matches i18n pattern', () => {
      render(
        <TextPropertyField
          resource={mockResource}
          propertyKey="label"
          propertyValue="{{t(common.submit)}}"
          onChange={mockOnChange}
        />,
        {wrapper: createWrapper()},
      );

      expect(screen.getByText('{{t(common.submit)}}')).toBeInTheDocument();
    });

    it('should not display i18n placeholder for regular text', () => {
      render(
        <TextPropertyField
          resource={mockResource}
          propertyKey="label"
          propertyValue="Regular Text"
          onChange={mockOnChange}
        />,
        {wrapper: createWrapper()},
      );

      // Should not have the i18n placeholder container
      expect(screen.queryByText('{{t(common.submit)}}')).not.toBeInTheDocument();
    });
  });

  describe('Accessibility', () => {
    it('should have accessible text input', () => {
      render(
        <TextPropertyField
          resource={mockResource}
          propertyKey="username"
          propertyValue=""
          onChange={mockOnChange}
        />,
        {wrapper: createWrapper()},
      );

      const textField = screen.getByRole('textbox');
      expect(textField).toBeInTheDocument();
    });

    it('should render label element with htmlFor attribute', () => {
      render(
        <TextPropertyField
          resource={mockResource}
          propertyKey="email"
          propertyValue=""
          onChange={mockOnChange}
        />,
        {wrapper: createWrapper()},
      );

      // Check that the label is rendered with the correct text
      const label = screen.getByText('Email');
      expect(label).toBeInTheDocument();
      // Check that the label has a for attribute (htmlFor in React)
      expect(label).toHaveAttribute('for');
    });
  });

  describe('TextField Props', () => {
    it('should clear text field value when i18n pattern is detected', () => {
      render(
        <TextPropertyField
          resource={mockResource}
          propertyKey="label"
          propertyValue="{{t(common.button)}}"
          onChange={mockOnChange}
        />,
        {wrapper: createWrapper()},
      );

      const textField = screen.getByRole('textbox');
      // When i18n pattern is detected, value is set to empty string
      expect(textField).toHaveValue('');
    });

    it('should render placeholder when not i18n pattern', () => {
      render(
        <TextPropertyField
          resource={mockResource}
          propertyKey="title"
          propertyValue=""
          onChange={mockOnChange}
        />,
        {wrapper: createWrapper()},
      );

      const textField = screen.getByRole('textbox');
      // The placeholder should be set based on the translation
      expect(textField).toHaveAttribute('placeholder', 'Enter Title');
    });

    it('should not render placeholder when i18n pattern is detected', () => {
      render(
        <TextPropertyField
          resource={mockResource}
          propertyKey="label"
          propertyValue="{{t(common.label)}}"
          onChange={mockOnChange}
        />,
        {wrapper: createWrapper()},
      );

      const textField = screen.getByRole('textbox');
      // When i18n pattern is detected, placeholder is empty
      expect(textField).toHaveAttribute('placeholder', '');
    });
  });

  describe('Additional Props', () => {
    it('should pass additional props to TextField', () => {
      render(
        <TextPropertyField
          resource={mockResource}
          propertyKey="username"
          propertyValue=""
          onChange={mockOnChange}
          disabled
        />,
        {wrapper: createWrapper()},
      );

      const textField = screen.getByRole('textbox');
      expect(textField).toBeDisabled();
    });

    it('should handle multiline prop', () => {
      render(
        <TextPropertyField
          resource={mockResource}
          propertyKey="description"
          propertyValue=""
          onChange={mockOnChange}
          multiline
          rows={4}
        />,
        {wrapper: createWrapper()},
      );

      // Check if textarea is rendered (multiline makes TextField render as textarea)
      const textArea = screen.getByRole('textbox');
      expect(textArea).toBeInTheDocument();
    });
  });

  describe('I18n Key Extraction', () => {
    it('should extract i18n key from pattern with simple key', () => {
      render(
        <TextPropertyField
          resource={mockResource}
          propertyKey="buttonLabel"
          propertyValue="{{t(login.submit)}}"
          onChange={mockOnChange}
        />,
        {wrapper: createWrapper()},
      );

      // The i18n pattern should be displayed
      expect(screen.getByText('{{t(login.submit)}}')).toBeInTheDocument();
    });

    it('should extract i18n key from pattern with nested key', () => {
      render(
        <TextPropertyField
          resource={mockResource}
          propertyKey="message"
          propertyValue="{{t(flows.login.welcome.message)}}"
          onChange={mockOnChange}
        />,
        {wrapper: createWrapper()},
      );

      expect(screen.getByText('{{t(flows.login.welcome.message)}}')).toBeInTheDocument();
    });
  });

  describe('Error State with TextField', () => {
    it('should set error prop on TextField when error message exists', () => {
      const notification = new Notification('notification-1', 'Error', 'error');
      notification.addResourceFieldNotification('resource-1_name', 'Name is required');

      const contextWithError: ValidationContextProps = {
        ...defaultContextValue,
        selectedNotification: notification,
      };

      render(
        <TextPropertyField resource={mockResource} propertyKey="name" propertyValue="" onChange={mockOnChange} />,
        {wrapper: createWrapper(contextWithError)},
      );

      // Check that error styling is applied
      const textField = screen.getByRole('textbox');
      expect(textField).toBeInTheDocument();
      expect(screen.getByText('Name is required')).toBeInTheDocument();
    });
  });

  describe('Resource ID Edge Cases', () => {
    it('should handle resource with undefined id', () => {
      const resourceWithUndefinedId = {...mockResource, id: undefined} as unknown as Resource;

      render(
        <TextPropertyField
          resource={resourceWithUndefinedId}
          propertyKey="label"
          propertyValue="test"
          onChange={mockOnChange}
        />,
        {wrapper: createWrapper()},
      );

      expect(screen.getByRole('textbox')).toBeInTheDocument();
    });

    it('should handle resource with empty id', () => {
      const resourceWithEmptyId = {...mockResource, id: ''};

      render(
        <TextPropertyField
          resource={resourceWithEmptyId}
          propertyKey="label"
          propertyValue="test"
          onChange={mockOnChange}
        />,
        {wrapper: createWrapper()},
      );

      expect(screen.getByRole('textbox')).toBeInTheDocument();
    });
  });

  describe('Notification Edge Cases', () => {
    it('should return empty string when selectedNotification is undefined', () => {
      const contextWithNoNotification: ValidationContextProps = {
        ...defaultContextValue,
        selectedNotification: undefined,
      };

      render(
        <TextPropertyField resource={mockResource} propertyKey="title" propertyValue="" onChange={mockOnChange} />,
        {wrapper: createWrapper(contextWithNoNotification)},
      );

      // Should not display any error message
      const formHelperTexts = document.querySelectorAll('.MuiFormHelperText-root');
      expect(formHelperTexts.length).toBe(0);
    });

    it('should handle notification without matching field notification', () => {
      const notification = new Notification('notification-1', 'Error', 'error');
      // No resource field notification added

      const contextWithNotification: ValidationContextProps = {
        ...defaultContextValue,
        selectedNotification: notification,
      };

      render(
        <TextPropertyField resource={mockResource} propertyKey="title" propertyValue="" onChange={mockOnChange} />,
        {wrapper: createWrapper(contextWithNotification)},
      );

      // Should not display error message
      expect(screen.queryByText('Title is required')).not.toBeInTheDocument();
    });
  });
});
