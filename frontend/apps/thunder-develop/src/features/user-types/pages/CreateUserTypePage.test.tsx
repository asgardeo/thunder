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
import {screen, waitFor} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import render from '@/test/test-utils';
import CreateUserTypePage from './CreateUserTypePage';
import type {ApiError} from '../types/user-types';

const mockNavigate = vi.fn();
const mockCreateUserType = vi.fn();

// Mock react-router
vi.mock('react-router', async () => {
  const actual = await vi.importActual<typeof import('react-router')>('react-router');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

// Mock useCreateUserType hook
interface UseCreateUserTypeReturn {
  createUserType: (data: {name: string; schema: Record<string, unknown>}) => Promise<void>;
  loading: boolean;
  error: ApiError | null;
}

const mockUseCreateUserType = vi.fn<() => UseCreateUserTypeReturn>();

vi.mock('../api/useCreateUserType', () => ({
  default: () => mockUseCreateUserType(),
}));

describe('CreateUserTypePage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseCreateUserType.mockReturnValue({
      createUserType: mockCreateUserType,
      loading: false,
      error: null,
    });
  });

  it('renders the page with initial form', () => {
    render(<CreateUserTypePage />);

    expect(screen.getByRole('heading', {name: 'Create User Type'})).toBeInTheDocument();
    expect(screen.getByText('Define a new user type schema for your organization')).toBeInTheDocument();
    expect(screen.getByLabelText(/User Type Name/i)).toBeInTheDocument();
  });

  it('navigates back when Back button is clicked', async () => {
    const user = userEvent.setup();
    render(<CreateUserTypePage />);

    const backButton = screen.getByRole('button', {name: /go back/i});
    await user.click(backButton);

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/user-types');
    });
  });

  it('navigates back when Cancel button is clicked', async () => {
    const user = userEvent.setup();
    render(<CreateUserTypePage />);

    const cancelButton = screen.getByRole('button', {name: /cancel/i});
    await user.click(cancelButton);

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/user-types');
    });
  });

  it('allows user to enter user type name', async () => {
    const user = userEvent.setup();
    render(<CreateUserTypePage />);

    const nameInput = screen.getByLabelText(/User Type Name/i);
    await user.type(nameInput, 'Employee');

    expect(nameInput).toHaveValue('Employee');
  });

  it('allows adding a new property', async () => {
    const user = userEvent.setup();
    render(<CreateUserTypePage />);

    const addButton = screen.getByRole('button', {name: /add property/i});
    await user.click(addButton);
  });

  it('allows removing a property', async () => {
    const user = userEvent.setup();
    render(<CreateUserTypePage />);

    // Add a second property first
    const addButton = screen.getByRole('button', {name: /add property/i});
    await user.click(addButton);

    // Now remove the second property - find the X icon button
    const removeButtons = screen
      .getAllByRole('button')
      .filter((btn) => btn.classList.contains('MuiIconButton-colorError'));

    await user.click(removeButtons[removeButtons.length - 1]);
  });

  it('allows changing property name', async () => {
    const user = userEvent.setup();
    render(<CreateUserTypePage />);

    const propertyNameInput = screen.getByPlaceholderText(/e.g., email, age, address/i);
    await user.type(propertyNameInput, 'email');

    expect(propertyNameInput).toHaveValue('email');
  });

  it('allows changing property type', async () => {
    const user = userEvent.setup();
    render(<CreateUserTypePage />);

    const typeSelect = screen.getByRole('combobox');
    await user.click(typeSelect);

    const numberOption = await screen.findByText('Number');
    await user.click(numberOption);

    await waitFor(() => {
      expect(typeSelect).toHaveTextContent('Number');
    });
  });

  it('shows validation error when submitting without name', async () => {
    const user = userEvent.setup();
    render(<CreateUserTypePage />);

    const submitButton = screen.getByRole('button', {name: /create user type/i});
    await user.click(submitButton);

    await waitFor(() => {
      expect(screen.getByText('Please enter a user type name')).toBeInTheDocument();
    });

    expect(mockCreateUserType).not.toHaveBeenCalled();
  });

  it('shows validation error when submitting without property name', async () => {
    const user = userEvent.setup();
    render(<CreateUserTypePage />);

    const nameInput = screen.getByLabelText(/User Type Name/i);
    await user.type(nameInput, 'Employee');

    const submitButton = screen.getByRole('button', {name: /create user type/i});
    await user.click(submitButton);

    await waitFor(() => {
      expect(screen.getByText('Please add at least one property')).toBeInTheDocument();
    });

    expect(mockCreateUserType).not.toHaveBeenCalled();
  });

  it('shows validation error for duplicate property names', async () => {
    const user = userEvent.setup();
    render(<CreateUserTypePage />);

    const nameInput = screen.getByLabelText(/User Type Name/i);
    await user.type(nameInput, 'Employee');

    // Add first property
    const firstPropertyInput = screen.getByPlaceholderText(/e.g., email, age, address/i);
    await user.type(firstPropertyInput, 'email');

    // Add second property
    const addButton = screen.getByRole('button', {name: /add property/i});
    await user.click(addButton);

    // Set same name for second property
    const propertyInputs = screen.getAllByPlaceholderText(/e.g., email, age, address/i);
    await user.type(propertyInputs[1], 'email');

    const submitButton = screen.getByRole('button', {name: /create user type/i});
    await user.click(submitButton);

    await waitFor(() => {
      expect(screen.getByText(/Duplicate property names found: email/i)).toBeInTheDocument();
    });

    expect(mockCreateUserType).not.toHaveBeenCalled();
  });

  it('allows toggling required checkbox', async () => {
    const user = userEvent.setup();
    render(<CreateUserTypePage />);

    const requiredCheckbox = screen.getByRole('checkbox', {name: /required/i});
    expect(requiredCheckbox).not.toBeChecked();

    await user.click(requiredCheckbox);
    expect(requiredCheckbox).toBeChecked();

    await user.click(requiredCheckbox);
    expect(requiredCheckbox).not.toBeChecked();
  });

  it('allows toggling unique checkbox for string type', async () => {
    const user = userEvent.setup();
    render(<CreateUserTypePage />);

    const uniqueCheckbox = screen.getByRole('checkbox', {name: /unique/i});
    expect(uniqueCheckbox).not.toBeChecked();

    await user.click(uniqueCheckbox);
    expect(uniqueCheckbox).toBeChecked();
  });

  it('hides unique checkbox for boolean type', async () => {
    const user = userEvent.setup();
    render(<CreateUserTypePage />);

    // Initially unique checkbox should be visible for string type
    expect(screen.getByRole('checkbox', {name: /unique/i})).toBeInTheDocument();

    // Change to boolean type
    const typeSelect = screen.getByRole('combobox');
    await user.click(typeSelect);

    const booleanOption = await screen.findByText('Boolean');
    await user.click(booleanOption);

    await waitFor(() => {
      expect(screen.queryByRole('checkbox', {name: /unique/i})).not.toBeInTheDocument();
    });
  });

  it('allows adding regex pattern for string type', async () => {
    const user = userEvent.setup();
    render(<CreateUserTypePage />);

    const regexInput = screen.getByPlaceholderText(/e.g., \^/i);
    await user.click(regexInput);
    await user.paste('^[a-z]+$');

    expect(regexInput).toHaveValue('^[a-z]+$');
  });

  it('allows adding enum values for string type', async () => {
    const user = userEvent.setup();
    render(<CreateUserTypePage />);

    const enumInput = screen.getByPlaceholderText(/Add value and press Enter/i);
    await user.type(enumInput, 'admin');

    const addEnumButton = screen.getByRole('button', {name: /^add$/i});
    await user.click(addEnumButton);

    await waitFor(() => {
      expect(screen.getByText('admin')).toBeInTheDocument();
    });

    expect(enumInput).toHaveValue('');
  });

  it('allows adding enum value by pressing Enter', async () => {
    const user = userEvent.setup();
    render(<CreateUserTypePage />);

    const enumInput = screen.getByPlaceholderText(/Add value and press Enter/i);
    await user.type(enumInput, 'user{Enter}');

    await waitFor(() => {
      expect(screen.getByText('user')).toBeInTheDocument();
    });

    expect(enumInput).toHaveValue('');
  });

  it('allows removing enum values', async () => {
    const user = userEvent.setup();
    render(<CreateUserTypePage />);

    const enumInput = screen.getByPlaceholderText(/Add value and press Enter/i);
    await user.type(enumInput, 'admin{Enter}');

    await waitFor(() => {
      expect(screen.getByText('admin')).toBeInTheDocument();
    });

    // Find and click the remove button for the enum value
    const enumContainer = screen.getByText('admin').closest('div');
    const removeButton = enumContainer?.querySelector('button');
    if (removeButton) {
      await user.click(removeButton);
    }

    await waitFor(() => {
      expect(screen.queryByText('admin')).not.toBeInTheDocument();
    });
  });

  it('successfully creates user type with valid data', async () => {
    const user = userEvent.setup();
    mockCreateUserType.mockResolvedValue(undefined);

    render(<CreateUserTypePage />);

    // Fill in user type name
    const nameInput = screen.getByLabelText(/User Type Name/i);
    await user.type(nameInput, 'Employee');

    // Fill in property name
    const propertyNameInput = screen.getByPlaceholderText(/e.g., email, age, address/i);
    await user.type(propertyNameInput, 'email');

    // Mark as required
    const requiredCheckbox = screen.getByRole('checkbox', {name: /required/i});
    await user.click(requiredCheckbox);

    // Submit form
    const submitButton = screen.getByRole('button', {name: /create user type/i});
    await user.click(submitButton);

    await waitFor(() => {
      expect(mockCreateUserType).toHaveBeenCalledWith({
        name: 'Employee',
        schema: {
          email: {
            type: 'string',
            required: true,
          },
        },
      });
    });

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/user-types');
    });
  });

  it('displays error from API', async () => {
    const error: ApiError = {
      code: 'CREATE_ERROR',
      message: 'Failed to create user type',
      description: 'User type already exists',
    };

    mockUseCreateUserType.mockReturnValue({
      createUserType: mockCreateUserType,
      loading: false,
      error,
    });

    render(<CreateUserTypePage />);

    expect(screen.getByText('Failed to create user type')).toBeInTheDocument();
    expect(screen.getByText('User type already exists')).toBeInTheDocument();
  });

  it('shows loading state during submission', () => {
    mockUseCreateUserType.mockReturnValue({
      createUserType: mockCreateUserType,
      loading: true,
      error: null,
    });

    render(<CreateUserTypePage />);

    expect(screen.getByText('Creating...')).toBeInTheDocument();
    expect(screen.getByRole('button', {name: /creating.../i})).toBeDisabled();
    expect(screen.getByRole('button', {name: /cancel/i})).toBeDisabled();
  });

  it('closes snackbar when close button is clicked', async () => {
    const user = userEvent.setup();
    render(<CreateUserTypePage />);

    // Trigger validation error
    const submitButton = screen.getByRole('button', {name: /create user type/i});
    await user.click(submitButton);

    await waitFor(() => {
      expect(screen.getByText('Please enter a user type name')).toBeInTheDocument();
    });

    // Close snackbar
    const closeButton = screen.getByLabelText(/close/i);
    await user.click(closeButton);

    await waitFor(() => {
      expect(screen.queryByText('Please enter a user type name')).not.toBeInTheDocument();
    });
  });

  it('creates schema with all property types correctly', async () => {
    const user = userEvent.setup();
    mockCreateUserType.mockResolvedValue(undefined);

    render(<CreateUserTypePage />);

    // Set user type name
    const nameInput = screen.getByLabelText(/User Type Name/i);
    await user.type(nameInput, 'Complex Type');

    // Add string property with all features
    const firstPropertyInput = screen.getByPlaceholderText(/e.g., email, age, address/i);
    await user.type(firstPropertyInput, 'status');

    const requiredCheckbox = screen.getByRole('checkbox', {name: /required/i});
    await user.click(requiredCheckbox);

    const uniqueCheckbox = screen.getByRole('checkbox', {name: /unique/i});
    await user.click(uniqueCheckbox);

    const regexInput = screen.getByPlaceholderText(/e.g., \^/i);
    await user.click(regexInput);
    await user.paste('^[A-Z]+$');

    const enumInput = screen.getByPlaceholderText(/Add value and press Enter/i);
    await user.type(enumInput, 'ACTIVE{Enter}');
    await user.type(enumInput, 'INACTIVE{Enter}');

    // Submit
    const submitButton = screen.getByRole('button', {name: /create user type/i});
    await user.click(submitButton);

    await waitFor(() => {
      expect(mockCreateUserType).toHaveBeenCalledWith({
        name: 'Complex Type',
        schema: {
          status: {
            type: 'string',
            required: true,
            unique: true,
            regex: '^[A-Z]+$',
            enum: ['ACTIVE', 'INACTIVE'],
          },
        },
      });
    });
  });

  it('resets type-specific fields when type changes', async () => {
    const user = userEvent.setup();
    render(<CreateUserTypePage />);

    // Add enum value for string type
    const enumInput = screen.getByPlaceholderText(/Add value and press Enter/i);
    await user.type(enumInput, 'test{Enter}');

    await waitFor(() => {
      expect(screen.getByText('test')).toBeInTheDocument();
    });

    // Change type to number
    const typeSelect = screen.getByRole('combobox');
    await user.click(typeSelect);

    const numberOption = await screen.findByText('Number');
    await user.click(numberOption);

    // Enum values should be gone
    await waitFor(() => {
      expect(screen.queryByText('test')).not.toBeInTheDocument();
    });

    // Regex and enum inputs should not be visible for number type
    expect(screen.queryByPlaceholderText(/Add value and press Enter/i)).not.toBeInTheDocument();
    expect(screen.queryByPlaceholderText(/e.g., \^/i)).not.toBeInTheDocument();
  });
});
