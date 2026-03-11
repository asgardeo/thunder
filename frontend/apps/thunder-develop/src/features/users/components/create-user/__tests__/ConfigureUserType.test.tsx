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

import {describe, it, expect, vi, beforeEach} from 'vitest';
import {page, userEvent} from 'vitest/browser';
import {renderWithProviders} from '@thunder/test-utils/browser';
import ConfigureUserType, {type ConfigureUserTypeProps} from '../ConfigureUserType';
import type {SchemaInterface} from '../../../types/users';

const mockSchemas: SchemaInterface[] = [
  {id: 'schema-1', name: 'Employee', ouId: 'ou-1'},
  {id: 'schema-2', name: 'Contractor', ouId: 'ou-2'},
  {id: 'schema-3', name: 'Vendor', ouId: 'ou-3'},
];

describe('ConfigureUserType', () => {
  const mockOnSchemaChange = vi.fn();
  const mockOnReadyChange = vi.fn();

  const defaultProps: ConfigureUserTypeProps = {
    schemas: mockSchemas,
    selectedSchema: null,
    onSchemaChange: mockOnSchemaChange,
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  const renderComponent = (props: Partial<ConfigureUserTypeProps> = {}) =>
    renderWithProviders(<ConfigureUserType {...defaultProps} {...props} />);

  it('renders the component with title and subtitle', async () => {
    await renderComponent();

    await expect.element(page.getByText('Select a user type')).toBeInTheDocument();
    await expect.element(page.getByText('Choose a user type (schema) for the new user.')).toBeInTheDocument();
  });

  it('renders the user type select field', async () => {
    await renderComponent();

    await expect.element(page.getByText('User Type')).toBeInTheDocument();
    await expect.element(page.getByTestId('configure-user-type')).toBeInTheDocument();
  });

  it('renders placeholder when no schema is selected', async () => {
    await renderComponent();

    await expect.element(page.getByText('Select a user type')).toBeInTheDocument();
  });

  it('renders all schema options in the select', async () => {
    await renderComponent();

    const select = page.getByRole('combobox');
    await userEvent.click(select);

    const listbox = page.getByRole('listbox');
    await expect.element(listbox.getByText('Employee')).toBeInTheDocument();
    await expect.element(listbox.getByText('Contractor')).toBeInTheDocument();
    await expect.element(listbox.getByText('Vendor')).toBeInTheDocument();
  });

  it('calls onSchemaChange when a schema is selected', async () => {
    await renderComponent();

    const select = page.getByRole('combobox');
    await userEvent.click(select);

    const listbox = page.getByRole('listbox');
    await userEvent.click(listbox.getByText('Employee'));

    await vi.waitFor(() => {
      expect(mockOnSchemaChange).toHaveBeenCalledWith(mockSchemas[0]);
    });
  });

  it('calls onSchemaChange when selecting a different schema', async () => {
    await renderComponent({selectedSchema: mockSchemas[0]});

    const select = page.getByRole('combobox');
    await userEvent.click(select);

    const listbox = page.getByRole('listbox');
    await userEvent.click(listbox.getByText('Contractor'));

    await vi.waitFor(() => {
      expect(mockOnSchemaChange).toHaveBeenCalledWith(mockSchemas[1]);
    });
  });

  it('displays the selected schema name', async () => {
    await renderComponent({selectedSchema: mockSchemas[0]});

    await expect.element(page.getByText('Employee')).toBeInTheDocument();
  });

  describe('onReadyChange callback', () => {
    it('calls onReadyChange with true when a schema is selected', async () => {
      await renderComponent({
        selectedSchema: mockSchemas[0],
        onReadyChange: mockOnReadyChange,
      });

      await vi.waitFor(() => {
        expect(mockOnReadyChange).toHaveBeenCalledWith(true);
      });
    });

    it('calls onReadyChange with false when no schema is selected', async () => {
      await renderComponent({
        selectedSchema: null,
        onReadyChange: mockOnReadyChange,
      });

      await vi.waitFor(() => {
        expect(mockOnReadyChange).toHaveBeenCalledWith(false);
      });
    });

    it('does not crash when onReadyChange is undefined', async () => {
      expect(async () => {
        await renderComponent({selectedSchema: mockSchemas[0], onReadyChange: undefined});
      }).not.toThrow();
    });

    it('calls onReadyChange when selectedSchema transitions from null to non-null', async () => {
      const {rerender} = await renderWithProviders(
        <ConfigureUserType {...defaultProps} selectedSchema={null} onReadyChange={mockOnReadyChange} />,
      );

      await vi.waitFor(() => {
        expect(mockOnReadyChange).toHaveBeenCalledWith(false);
      });
      mockOnReadyChange.mockClear();

      await rerender(
        <ConfigureUserType {...defaultProps} selectedSchema={mockSchemas[0]} onReadyChange={mockOnReadyChange} />,
      );

      await vi.waitFor(() => {
        expect(mockOnReadyChange).toHaveBeenCalledWith(true);
      });
    });
  });

  it('handles empty schemas list', async () => {
    await renderComponent({schemas: []});

    await expect.element(page.getByTestId('configure-user-type')).toBeInTheDocument();
  });
});
