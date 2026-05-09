/**
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
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

import {render, screen, waitFor, within, userEvent} from '@thunderid/test-utils';
import {describe, it, expect, vi, beforeEach} from 'vitest';
import type {SchemaPropertyInput} from '../../../../types/user-types';
import EditSchemaSettings from '../EditSchemaSettings';

describe('EditSchemaSettings', () => {
  const mockOnPropertiesChange = vi.fn();

  const baseProperties: SchemaPropertyInput[] = [
    {
      id: '0',
      name: 'email',
      displayName: '',
      type: 'string',
      required: true,
      unique: true,
      credential: false,
      enum: [],
      regex: '',
    },
    {
      id: '1',
      name: 'age',
      displayName: '',
      type: 'number',
      required: false,
      unique: false,
      credential: false,
      enum: [],
      regex: '',
    },
  ];

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('resets unique and credential when changing type to boolean', async () => {
    const user = userEvent.setup();
    const props = {
      properties: [{...baseProperties[0], unique: true, credential: true}],
      onPropertiesChange: mockOnPropertiesChange,
      userTypeName: 'Test',
    };

    render(<EditSchemaSettings {...props} />);

    // Change type from string to boolean
    const typeSelect = screen.getByRole('combobox');
    await user.click(typeSelect);
    const booleanOption = await screen.findByRole('option', {name: 'Boolean'});
    await user.click(booleanOption);

    expect(mockOnPropertiesChange).toHaveBeenCalledWith(
      expect.arrayContaining([
        expect.objectContaining({
          type: 'boolean',
          unique: false,
          credential: false,
        }),
      ]),
    );
  });

  it('preserves enum values when changing type to enum', async () => {
    const user = userEvent.setup();
    const props = {
      properties: [{...baseProperties[0], enum: ['ACTIVE', 'INACTIVE']}],
      onPropertiesChange: mockOnPropertiesChange,
      userTypeName: 'Test',
    };

    render(<EditSchemaSettings {...props} />);

    const typeSelect = screen.getByRole('combobox');
    await user.click(typeSelect);
    const enumOption = await screen.findByRole('option', {name: 'Enum'});
    await user.click(enumOption);

    expect(mockOnPropertiesChange).toHaveBeenCalledWith(
      expect.arrayContaining([
        expect.objectContaining({
          type: 'enum',
          enum: ['ACTIVE', 'INACTIVE'],
        }),
      ]),
    );
  });

  it('clears enum values when changing from enum to number', async () => {
    const user = userEvent.setup();
    const props = {
      properties: [{...baseProperties[0], type: 'enum' as const, enum: ['A', 'B']}],
      onPropertiesChange: mockOnPropertiesChange,
      userTypeName: 'Test',
    };

    render(<EditSchemaSettings {...props} />);

    const typeSelect = screen.getByRole('combobox');
    await user.click(typeSelect);
    const numberOption = await screen.findByRole('option', {name: 'Number'});
    await user.click(numberOption);

    expect(mockOnPropertiesChange).toHaveBeenCalledWith(
      expect.arrayContaining([
        expect.objectContaining({
          type: 'number',
          enum: [],
          regex: '',
        }),
      ]),
    );
  });

  it('does not add duplicate enum value', async () => {
    const user = userEvent.setup();
    const props = {
      properties: [{...baseProperties[0], type: 'enum' as const, enum: ['ACTIVE']}],
      onPropertiesChange: mockOnPropertiesChange,
      userTypeName: 'Test',
    };

    render(<EditSchemaSettings {...props} />);

    const enumInput = screen.getByPlaceholderText(/add value and press enter/i);
    await user.type(enumInput, 'ACTIVE');
    const addButton = screen.getByRole('button', {name: /^add$/i});
    await user.click(addButton);

    // onPropertiesChange should NOT have been called for a duplicate
    expect(mockOnPropertiesChange).not.toHaveBeenCalled();
  });

  it('shows credential removal confirmation dialog when unchecking credential', async () => {
    const user = userEvent.setup();
    const props = {
      properties: [{...baseProperties[0], credential: true}],
      onPropertiesChange: mockOnPropertiesChange,
      userTypeName: 'Test',
    };

    render(<EditSchemaSettings {...props} />);

    const credentialCheckbox = screen.getByRole('checkbox', {name: /values will be hashed/i});
    await user.click(credentialCheckbox);

    await waitFor(() => {
      expect(screen.getByText(/removing the credential flag/i)).toBeInTheDocument();
    });
  });

  it('confirms credential removal via dialog', async () => {
    const user = userEvent.setup();
    const props = {
      properties: [{...baseProperties[0], credential: true}],
      onPropertiesChange: mockOnPropertiesChange,
      userTypeName: 'Test',
    };

    render(<EditSchemaSettings {...props} />);

    // Uncheck credential
    const credentialCheckbox = screen.getByRole('checkbox', {name: /values will be hashed/i});
    await user.click(credentialCheckbox);

    // Confirm removal
    const dialog = screen.getByRole('dialog');
    const confirmButton = within(dialog).getByRole('button', {name: /remove credential/i});
    await user.click(confirmButton);

    await waitFor(() => {
      expect(mockOnPropertiesChange).toHaveBeenCalledWith(
        expect.arrayContaining([
          expect.objectContaining({
            credential: false,
          }),
        ]),
      );
    });
  });

  it('cancels credential removal via dialog', async () => {
    const user = userEvent.setup();
    const props = {
      properties: [{...baseProperties[0], credential: true}],
      onPropertiesChange: mockOnPropertiesChange,
      userTypeName: 'Test',
    };

    render(<EditSchemaSettings {...props} />);

    const credentialCheckbox = screen.getByRole('checkbox', {name: /values will be hashed/i});
    await user.click(credentialCheckbox);

    const dialog = screen.getByRole('dialog');
    const cancelButton = within(dialog).getByRole('button', {name: /cancel/i});
    await user.click(cancelButton);

    await waitFor(() => {
      expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    });

    // Should not have changed properties
    expect(mockOnPropertiesChange).not.toHaveBeenCalled();
  });

  it('adds new property with incremented id', async () => {
    const user = userEvent.setup();
    const props = {
      properties: baseProperties,
      onPropertiesChange: mockOnPropertiesChange,
      userTypeName: 'Test',
    };

    render(<EditSchemaSettings {...props} />);

    const addButton = screen.getByRole('button', {name: /add property/i});
    await user.click(addButton);

    expect(mockOnPropertiesChange).toHaveBeenCalledWith(
      expect.arrayContaining([
        ...baseProperties,
        expect.objectContaining({
          id: '2',
          name: '',
          type: 'string',
        }),
      ]),
    );
  });

  it('does not show remove button when there is only one property', () => {
    const props = {
      properties: [baseProperties[0]],
      onPropertiesChange: mockOnPropertiesChange,
      userTypeName: 'Test',
    };

    render(<EditSchemaSettings {...props} />);

    expect(screen.queryByRole('button', {name: /remove property/i})).not.toBeInTheDocument();
  });
});
