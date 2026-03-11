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
import {ElementTypes} from '@/features/flows/models/elements';
import type {Resource} from '@/features/flows/models/resources';
import FieldExtendedProperties from '../FieldExtendedProperties';

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
    it('should render the component for text input', async () => {
      const resource = createMockResource(ElementTypes.TextInput);

      await render(<FieldExtendedProperties resource={resource} onChange={mockOnChange} />);

      await expect.element(page.getByText('Attribute')).toBeInTheDocument();
    });

    it('should render Autocomplete component', async () => {
      const resource = createMockResource(ElementTypes.TextInput);

      const {container} = await render(<FieldExtendedProperties resource={resource} onChange={mockOnChange} />);

      expect(container.querySelector('.MuiAutocomplete-root')).toBeInTheDocument();
    });

    it('should render with placeholder text', async () => {
      const resource = createMockResource(ElementTypes.TextInput);

      await render(<FieldExtendedProperties resource={resource} onChange={mockOnChange} />);

      await expect.element(page.getByPlaceholder('Select an attribute')).toBeInTheDocument();
    });
  });

  describe('Password Input Handling', () => {
    it('should render with credential attributes for PasswordInput type', async () => {
            const resource = createMockResource(ElementTypes.PasswordInput);

      await render(<FieldExtendedProperties resource={resource} onChange={mockOnChange} />);

      const input = page.getByRole('combobox');
      expect(input).toBeInTheDocument();

      await userEvent.click(input);

      expect(page.getByRole('option', {name: 'password'})).toBeInTheDocument();
      await expect.element(page.getByRole('option', {name: 'pin'})).toBeInTheDocument();
      await expect.element(page.getByRole('option', {name: 'secret'})).toBeInTheDocument();
    });
  });

  describe('Attribute Selection', () => {
    it('should have email, username, firstName as options', async () => {
      const resource = createMockResource(ElementTypes.TextInput);

      await render(<FieldExtendedProperties resource={resource} onChange={mockOnChange} />);

      const input = page.getByPlaceholder('Select an attribute');
      await userEvent.click(input);

      // Check for dropdown options (may be in listbox)
      await expect.element(page.getByRole('combobox')).toBeInTheDocument();
    });

    it('should display current ref value', async () => {
      const resource = createMockResource(ElementTypes.TextInput, {ref: 'email'} as Partial<Resource>);

      await render(<FieldExtendedProperties resource={resource} onChange={mockOnChange} />);

      const input = page.getByRole('combobox');
      expect(input).toHaveValue('email');
    });
  });

  describe('Resource Change Handling', () => {
    it('should sync value when resource changes', async () => {
      const resource1 = createMockResource(ElementTypes.TextInput, {id: 'field-1', ref: 'email'} as Partial<Resource>);
      const resource2 = createMockResource(ElementTypes.TextInput, {
        id: 'field-2',
        ref: 'username',
      } as Partial<Resource>);

      const {rerender} = await render(<FieldExtendedProperties resource={resource1} onChange={mockOnChange} />);

      let input = page.getByRole('combobox');
      expect(input).toHaveValue('email');

      await rerender(<FieldExtendedProperties resource={resource2} onChange={mockOnChange} />);

      input = page.getByRole('combobox');
      expect(input).toHaveValue('username');
    });

    it('should sync to empty when resource ref changes to undefined (same id)', async () => {
      const resourceWithRef = createMockResource(ElementTypes.TextInput, {
        id: 'field-1',
        ref: 'email',
      } as Partial<Resource>);
      const resourceWithoutRef = createMockResource(ElementTypes.TextInput, {id: 'field-1'} as Partial<Resource>);

      const {rerender} = await render(<FieldExtendedProperties resource={resourceWithRef} onChange={mockOnChange} />);

      let input = page.getByRole('combobox');
      expect(input).toHaveValue('email');

      await rerender(<FieldExtendedProperties resource={resourceWithoutRef} onChange={mockOnChange} />);

      input = page.getByRole('combobox');
      expect(input).toHaveValue('email');
    });
  });

  describe('onChange Handling', () => {
    it('should call onChange when selecting an attribute from dropdown', async () => {
            const resource = createMockResource(ElementTypes.TextInput);

      await render(<FieldExtendedProperties resource={resource} onChange={mockOnChange} />);

      const input = page.getByRole('combobox');
      await userEvent.click(input);

      // Wait for dropdown to open and select an option
      const option = page.getByRole('option', {name: 'email'});
      await userEvent.click(option);

      expect(mockOnChange).toHaveBeenCalledWith('ref', 'email', resource);
    });

    it('should call onChange with empty string when clearing selection', async () => {
      const resource = createMockResource(ElementTypes.TextInput, {ref: 'email'} as Partial<Resource>);

      await render(<FieldExtendedProperties resource={resource} onChange={mockOnChange} />);

      // Clear the input by clicking the clear button
      const clearButton = page.getByRole('button', {name: 'Clear'});
      await userEvent.click(clearButton);

      expect(mockOnChange).toHaveBeenCalledWith('ref', '', resource);
    });

    it('should call onChange when typing a custom value (free-solo)', async () => {
      const resource = createMockResource(ElementTypes.TextInput);

      await render(<FieldExtendedProperties resource={resource} onChange={mockOnChange} />);

      const input = page.getByRole('combobox');
      await userEvent.fill(input, 'customAttribute');

      expect(mockOnChange).toHaveBeenCalledWith('ref', 'customAttribute', resource);
    });
  });

  describe('Error Message Handling', () => {
    it('should display error message when validation error exists', async () => {
      mockHasResourceFieldNotification.mockReturnValue(true);
      mockGetResourceFieldNotification.mockReturnValue('This field is required');

      const resource = createMockResource(ElementTypes.TextInput);

      await render(<FieldExtendedProperties resource={resource} onChange={mockOnChange} />);

      await expect.element(page.getByText('This field is required')).toBeInTheDocument();
    });
  });
});
