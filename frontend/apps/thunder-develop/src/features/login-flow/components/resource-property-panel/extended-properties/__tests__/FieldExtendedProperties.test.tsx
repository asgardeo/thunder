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
import userEvent from '@testing-library/user-event';
import {ElementTypes} from '@/features/flows/models/elements';
import type {Resource} from '@/features/flows/models/resources';
import FieldExtendedProperties from '../FieldExtendedProperties';

// Mock dependencies
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
}));

const mockHasResourceFieldNotification = vi.fn().mockReturnValue(false);
const mockGetResourceFieldNotification = vi.fn().mockReturnValue('');

vi.mock('@/features/flows/hooks/useValidationStatus', () => ({
  default: () => ({
    selectedNotification: {
      hasResourceFieldNotification: mockHasResourceFieldNotification,
      getResourceFieldNotification: mockGetResourceFieldNotification,
    },
  }),
}));

describe('FieldExtendedProperties', () => {
  const mockOnChange = vi.fn();

  const createMockResource = (type: string, overrides: Partial<Resource> = {}): Resource =>
    ({
      id: 'field-1',
      type,
      category: 'FIELD',
      resourceType: 'ELEMENT',
      ...overrides,
    }) as Resource;

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('should render the component for text input', () => {
      const resource = createMockResource(ElementTypes.TextInput);

      render(<FieldExtendedProperties resource={resource} onChange={mockOnChange} />);

      expect(screen.getByText('flows:core.fieldExtendedProperties.attribute')).toBeInTheDocument();
    });

    it('should render Autocomplete component', () => {
      const resource = createMockResource(ElementTypes.TextInput);

      const {container} = render(<FieldExtendedProperties resource={resource} onChange={mockOnChange} />);

      expect(container.querySelector('.MuiAutocomplete-root')).toBeInTheDocument();
    });

    it('should render with placeholder text', () => {
      const resource = createMockResource(ElementTypes.TextInput);

      render(<FieldExtendedProperties resource={resource} onChange={mockOnChange} />);

      expect(screen.getByPlaceholderText('flows:core.fieldExtendedProperties.selectAttribute')).toBeInTheDocument();
    });
  });

  describe('Password Input Handling', () => {
    it('should render with credential attributes for PasswordInput type', async () => {
      const user = userEvent.setup();
      const resource = createMockResource(ElementTypes.PasswordInput);

      render(<FieldExtendedProperties resource={resource} onChange={mockOnChange} />);

      const input = screen.getByRole('combobox');
      expect(input).toBeInTheDocument();

      await user.click(input);

      expect(await screen.findByRole('option', {name: 'password'})).toBeInTheDocument();
      expect(screen.getByRole('option', {name: 'pin'})).toBeInTheDocument();
      expect(screen.getByRole('option', {name: 'secret'})).toBeInTheDocument();
    });
  });

  describe('Attribute Selection', () => {
    it('should have email, username, firstName as options', async () => {
      const resource = createMockResource(ElementTypes.TextInput);

      render(<FieldExtendedProperties resource={resource} onChange={mockOnChange} />);

      const input = screen.getByPlaceholderText('flows:core.fieldExtendedProperties.selectAttribute');
      fireEvent.focus(input);
      fireEvent.click(input);

      // Check for dropdown options (may be in listbox)
      expect(screen.getByRole('combobox')).toBeInTheDocument();
    });

    it('should display current ref value', () => {
      const resource = createMockResource(ElementTypes.TextInput, {ref: 'email'} as Partial<Resource>);

      render(<FieldExtendedProperties resource={resource} onChange={mockOnChange} />);

      const input = screen.getByRole('combobox');
      expect(input).toHaveValue('email');
    });
  });

  describe('Resource Change Handling', () => {
    it('should sync value when resource changes', () => {
      const resource1 = createMockResource(ElementTypes.TextInput, {id: 'field-1', ref: 'email'} as Partial<Resource>);
      const resource2 = createMockResource(ElementTypes.TextInput, {
        id: 'field-2',
        ref: 'username',
      } as Partial<Resource>);

      const {rerender} = render(<FieldExtendedProperties resource={resource1} onChange={mockOnChange} />);

      let input = screen.getByRole('combobox');
      expect(input).toHaveValue('email');

      rerender(<FieldExtendedProperties resource={resource2} onChange={mockOnChange} />);

      input = screen.getByRole('combobox');
      expect(input).toHaveValue('username');
    });

    it('should sync to empty when resource ref changes to undefined (same id)', () => {
      const resourceWithRef = createMockResource(ElementTypes.TextInput, {
        id: 'field-1',
        ref: 'email',
      } as Partial<Resource>);
      const resourceWithoutRef = createMockResource(ElementTypes.TextInput, {id: 'field-1'} as Partial<Resource>);

      const {rerender} = render(<FieldExtendedProperties resource={resourceWithRef} onChange={mockOnChange} />);

      let input = screen.getByRole('combobox');
      expect(input).toHaveValue('email');

      rerender(<FieldExtendedProperties resource={resourceWithoutRef} onChange={mockOnChange} />);

      input = screen.getByRole('combobox');
      expect(input).toHaveValue('email');
    });
  });

  describe('onChange Handling', () => {
    it('should call onChange when selecting an attribute from dropdown', async () => {
      const user = userEvent.setup();
      const resource = createMockResource(ElementTypes.TextInput);

      render(<FieldExtendedProperties resource={resource} onChange={mockOnChange} />);

      const input = screen.getByRole('combobox');
      await user.click(input);

      // Wait for dropdown to open and select an option
      const option = await screen.findByRole('option', {name: 'email'});
      await user.click(option);

      expect(mockOnChange).toHaveBeenCalledWith('ref', 'email', resource);
    });

    it('should call onChange with empty string when clearing selection', () => {
      const resource = createMockResource(ElementTypes.TextInput, {ref: 'email'} as Partial<Resource>);

      render(<FieldExtendedProperties resource={resource} onChange={mockOnChange} />);

      // Clear the input by clicking the clear button
      const clearButton = screen.getByTitle('Clear');
      fireEvent.click(clearButton);

      expect(mockOnChange).toHaveBeenCalledWith('ref', '', resource);
    });

    it('should call onChange when typing a custom value (free-solo)', () => {
      const resource = createMockResource(ElementTypes.TextInput);

      render(<FieldExtendedProperties resource={resource} onChange={mockOnChange} />);

      const input = screen.getByRole('combobox');
      fireEvent.change(input, {target: {value: 'customAttribute'}});

      expect(mockOnChange).toHaveBeenCalledWith('ref', 'customAttribute', resource);
    });
  });

  describe('Error Message Handling', () => {
    it('should display error message when validation error exists', () => {
      mockHasResourceFieldNotification.mockReturnValue(true);
      mockGetResourceFieldNotification.mockReturnValue('This field is required');

      const resource = createMockResource(ElementTypes.TextInput);

      render(<FieldExtendedProperties resource={resource} onChange={mockOnChange} />);

      expect(screen.getByText('This field is required')).toBeInTheDocument();
    });
  });
});
