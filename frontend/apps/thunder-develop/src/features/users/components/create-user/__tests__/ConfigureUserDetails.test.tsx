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
import ConfigureUserDetails, {type ConfigureUserDetailsProps} from '../ConfigureUserDetails';
import type {ApiUserSchema} from '../../../types/users';

const mockSchema: ApiUserSchema = {
  id: 'schema-1',
  name: 'Employee',
  schema: {
    username: {
      type: 'string',
      required: true,
    },
    age: {
      type: 'number',
      required: false,
    },
    active: {
      type: 'boolean',
      required: false,
    },
  },
};

describe('ConfigureUserDetails', () => {
  const mockOnFormValuesChange = vi.fn();
  const mockOnReadyChange = vi.fn();

  const defaultProps: ConfigureUserDetailsProps = {
    schema: mockSchema,
    defaultValues: {},
    onFormValuesChange: mockOnFormValuesChange,
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  const renderComponent = (props: Partial<ConfigureUserDetailsProps> = {}) =>
    renderWithProviders(<ConfigureUserDetails {...defaultProps} {...props} />);

  it('renders the component with title and subtitle', async () => {
    await renderComponent();

    await expect.element(page.getByText('Enter user details')).toBeInTheDocument();
    await expect.element(page.getByText('Fill in the required information for the new user.')).toBeInTheDocument();
  });

  it('renders the data-testid attribute', async () => {
    await renderComponent();

    await expect.element(page.getByTestId('configure-user-details')).toBeInTheDocument();
  });

  it('renders string fields from the schema', async () => {
    await renderComponent();

    await expect.element(page.getByPlaceholder(/enter username/i)).toBeInTheDocument();
  });

  it('renders number fields from the schema', async () => {
    await renderComponent();

    await expect.element(page.getByPlaceholder(/enter age/i)).toBeInTheDocument();
  });

  it('renders boolean fields from the schema', async () => {
    await renderComponent();

    await expect.element(page.getByRole('checkbox')).toBeInTheDocument();
  });

  it('calls onFormValuesChange when form values change', async () => {
    await renderComponent();

    const usernameInput = page.getByPlaceholder(/enter username/i);
    await userEvent.fill(usernameInput, 'john');

    await vi.waitFor(() => {
      expect(mockOnFormValuesChange).toHaveBeenCalled();
      const lastCall = mockOnFormValuesChange.mock.calls[mockOnFormValuesChange.mock.calls.length - 1][0] as Record<string, unknown>;
      expect(lastCall).toHaveProperty('username', 'john');
    });
  });

  it('renders with default values pre-filled', async () => {
    await renderComponent({
      defaultValues: {username: 'existing_user', age: 25},
    });

    await expect.element(page.getByPlaceholder(/enter username/i)).toHaveValue('existing_user');
  });

  describe('onReadyChange callback', () => {
    it('calls onReadyChange with false when required fields are empty', async () => {
      await renderComponent({onReadyChange: mockOnReadyChange});

      // username is required and starts empty, so form is not valid
      await vi.waitFor(() => {
        expect(mockOnReadyChange).toHaveBeenCalledWith(false);
      });
    });

    it('calls onReadyChange with true when required fields are filled', async () => {
      await renderComponent({onReadyChange: mockOnReadyChange});

      const usernameInput = page.getByPlaceholder(/enter username/i);
      await userEvent.fill(usernameInput, 'john');

      await vi.waitFor(() => {
        expect(mockOnReadyChange).toHaveBeenCalledWith(true);
      });
    });

    it('does not crash when onReadyChange is undefined', async () => {
      expect(async () => {
        await renderComponent({onReadyChange: undefined});
      }).not.toThrow();
    });
  });

  it('renders credential fields as password inputs with toggle visibility', async () => {
    const schemaWithCredential: ApiUserSchema = {
      id: 'schema-cred',
      name: 'Employee',
      schema: {
        username: {
          type: 'string',
          required: true,
        },
        password: {
          type: 'string',
          required: true,
          credential: true,
        },
      },
    };

    await renderComponent({schema: schemaWithCredential});

    const passwordInput = page.getByPlaceholder(/enter password/i);
    await expect.element(passwordInput).toHaveAttribute('type', 'password');
    await expect.element(page.getByLabelText('show password')).toBeInTheDocument();

    const usernameInput = page.getByPlaceholder(/enter username/i);
    await expect.element(usernameInput).toHaveAttribute('type', 'text');
  });

  it('handles schema with no fields', async () => {
    const emptySchema: ApiUserSchema = {
      id: 'schema-empty',
      name: 'Empty',
      schema: {},
    };

    await renderComponent({schema: emptySchema});

    await expect.element(page.getByTestId('configure-user-details')).toBeInTheDocument();
  });
});
