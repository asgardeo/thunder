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
import {render, screen} from '@testing-library/react';
import type {ReactNode} from 'react';
import CommonElementPropertyFactory from '../CommonElementPropertyFactory';
import {ValidationContext, type ValidationContextProps} from '../../../context/ValidationContext';
import type {Resource} from '../../../models/resources';
import {ElementTypes} from '../../../models/elements';
import FlowBuilderElementConstants from '../../../constants/FlowBuilderElementConstants';

// Mock child components
vi.mock('../rich-text/RichTextWithTranslation', () => ({
  default: ({onChange}: {onChange: (html: string) => void}) => (
    <div data-testid="rich-text-with-translation">
      <button type="button" onClick={() => onChange('<p>test</p>')}>
        Rich Text Editor
      </button>
    </div>
  ),
}));

vi.mock('../CheckboxPropertyField', () => ({
  default: ({
    resource,
    propertyKey,
    propertyValue,
    onChange,
  }: {
    resource: Resource;
    propertyKey: string;
    propertyValue: boolean;
    onChange: (key: string, value: boolean, resource: Resource) => void;
  }) => (
    <div data-testid="checkbox-property-field">
      <input
        type="checkbox"
        checked={propertyValue}
        onChange={(e) => onChange(propertyKey, e.target.checked, resource)}
        data-property-key={propertyKey}
      />
    </div>
  ),
}));

vi.mock('../TextPropertyField', () => ({
  default: ({
    resource,
    propertyKey,
    propertyValue,
    onChange,
  }: {
    resource: Resource;
    propertyKey: string;
    propertyValue: string;
    onChange: (key: string, value: string, resource: Resource) => void;
  }) => (
    <div data-testid="text-property-field">
      <input
        type="text"
        value={propertyValue}
        onChange={(e) => onChange(propertyKey, e.target.value, resource)}
        data-property-key={propertyKey}
      />
    </div>
  ),
}));

describe('CommonElementPropertyFactory', () => {
  const mockOnChange = vi.fn();

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

  describe('RichText Element', () => {
    it('should render RichTextWithTranslation for label property when resource is RichText', () => {
      const richTextResource: Resource = {
        id: 'resource-1',
        type: ElementTypes.RichText,
        config: {},
      } as Resource;

      render(
        <CommonElementPropertyFactory
          resource={richTextResource}
          propertyKey="label"
          propertyValue="<p>Test content</p>"
          onChange={mockOnChange}
        />,
        {wrapper: createWrapper()},
      );

      expect(screen.getByTestId('rich-text-with-translation')).toBeInTheDocument();
    });

    it('should not render RichTextWithTranslation for non-label properties on RichText', () => {
      const richTextResource: Resource = {
        id: 'resource-1',
        type: ElementTypes.RichText,
        config: {},
      } as Resource;

      render(
        <CommonElementPropertyFactory
          resource={richTextResource}
          propertyKey="other"
          propertyValue="test"
          onChange={mockOnChange}
        />,
        {wrapper: createWrapper()},
      );

      expect(screen.queryByTestId('rich-text-with-translation')).not.toBeInTheDocument();
      expect(screen.getByTestId('text-property-field')).toBeInTheDocument();
    });
  });

  describe('Boolean Properties', () => {
    it('should render CheckboxPropertyField for boolean property values', () => {
      const resource: Resource = {
        id: 'resource-1',
        type: ElementTypes.TextInput,
        config: {},
      } as Resource;

      render(
        <CommonElementPropertyFactory
          resource={resource}
          propertyKey="required"
          propertyValue
          onChange={mockOnChange}
        />,
        {wrapper: createWrapper()},
      );

      expect(screen.getByTestId('checkbox-property-field')).toBeInTheDocument();
    });

    it('should render CheckboxPropertyField for false boolean values', () => {
      const resource: Resource = {
        id: 'resource-1',
        type: ElementTypes.TextInput,
        config: {},
      } as Resource;

      render(
        <CommonElementPropertyFactory
          resource={resource}
          propertyKey="disabled"
          propertyValue={false}
          onChange={mockOnChange}
        />,
        {wrapper: createWrapper()},
      );

      expect(screen.getByTestId('checkbox-property-field')).toBeInTheDocument();
    });
  });

  describe('String Properties', () => {
    it('should render TextPropertyField for string property values', () => {
      const resource: Resource = {
        id: 'resource-1',
        type: ElementTypes.TextInput,
        config: {},
      } as Resource;

      render(
        <CommonElementPropertyFactory
          resource={resource}
          propertyKey="placeholder"
          propertyValue="Enter text"
          onChange={mockOnChange}
        />,
        {wrapper: createWrapper()},
      );

      expect(screen.getByTestId('text-property-field')).toBeInTheDocument();
    });

    it('should render TextPropertyField for empty string values', () => {
      const resource: Resource = {
        id: 'resource-1',
        type: ElementTypes.TextInput,
        config: {},
      } as Resource;

      render(
        <CommonElementPropertyFactory
          resource={resource}
          propertyKey="hint"
          propertyValue=""
          onChange={mockOnChange}
        />,
        {wrapper: createWrapper()},
      );

      expect(screen.getByTestId('text-property-field')).toBeInTheDocument();
    });
  });

  describe('Captcha Element', () => {
    it('should render TextField with default provider for Captcha resource', () => {
      const captchaResource: Resource = {
        id: 'resource-1',
        type: ElementTypes.Captcha,
        config: {},
      } as Resource;

      render(
        <CommonElementPropertyFactory
          resource={captchaResource}
          propertyKey="provider"
          propertyValue={undefined}
          onChange={mockOnChange}
        />,
        {wrapper: createWrapper()},
      );

      const textField = screen.getByRole('textbox');
      expect(textField).toBeInTheDocument();
      expect(textField).toHaveValue(FlowBuilderElementConstants.DEFAULT_CAPTCHA_PROVIDER);
    });

    it('should render disabled TextField for Captcha provider', () => {
      const captchaResource: Resource = {
        id: 'resource-1',
        type: ElementTypes.Captcha,
        config: {},
      } as Resource;

      render(
        <CommonElementPropertyFactory
          resource={captchaResource}
          propertyKey="provider"
          propertyValue={null}
          onChange={mockOnChange}
        />,
        {wrapper: createWrapper()},
      );

      const textField = screen.getByRole('textbox');
      expect(textField).toBeDisabled();
    });
  });

  describe('Null Cases', () => {
    it('should return null for unsupported property types', () => {
      const resource: Resource = {
        id: 'resource-1',
        type: ElementTypes.TextInput,
        config: {},
      } as Resource;

      const {container} = render(
        <CommonElementPropertyFactory
          resource={resource}
          propertyKey="complexProp"
          propertyValue={{nested: 'object'}}
          onChange={mockOnChange}
        />,
        {wrapper: createWrapper()},
      );

      expect(container.firstChild).toBeNull();
    });

    it('should return null for number property values', () => {
      const resource: Resource = {
        id: 'resource-1',
        type: ElementTypes.TextInput,
        config: {},
      } as Resource;

      const {container} = render(
        <CommonElementPropertyFactory
          resource={resource}
          propertyKey="count"
          propertyValue={42}
          onChange={mockOnChange}
        />,
        {wrapper: createWrapper()},
      );

      expect(container.firstChild).toBeNull();
    });

    it('should return null for array property values', () => {
      const resource: Resource = {
        id: 'resource-1',
        type: ElementTypes.TextInput,
        config: {},
      } as Resource;

      const {container} = render(
        <CommonElementPropertyFactory
          resource={resource}
          propertyKey="items"
          propertyValue={['item1', 'item2']}
          onChange={mockOnChange}
        />,
        {wrapper: createWrapper()},
      );

      expect(container.firstChild).toBeNull();
    });
  });

  describe('Additional Props', () => {
    it('should pass additional props to child components', () => {
      const resource: Resource = {
        id: 'resource-1',
        type: ElementTypes.TextInput,
        config: {},
      } as Resource;

      render(
        <CommonElementPropertyFactory
          resource={resource}
          propertyKey="label"
          propertyValue="Test Label"
          onChange={mockOnChange}
          data-custom-prop="custom-value"
        />,
        {wrapper: createWrapper()},
      );

      expect(screen.getByTestId('text-property-field')).toBeInTheDocument();
    });
  });

  describe('Label Property for Non-RichText', () => {
    it('should render TextPropertyField for label property on non-RichText elements', () => {
      const resource: Resource = {
        id: 'resource-1',
        type: ElementTypes.TextInput,
        config: {},
      } as Resource;

      render(
        <CommonElementPropertyFactory
          resource={resource}
          propertyKey="label"
          propertyValue="My Label"
          onChange={mockOnChange}
        />,
        {wrapper: createWrapper()},
      );

      expect(screen.getByTestId('text-property-field')).toBeInTheDocument();
    });
  });
});
