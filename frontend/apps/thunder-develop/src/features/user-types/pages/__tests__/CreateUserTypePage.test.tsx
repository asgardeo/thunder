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
import type {OrganizationUnitListParams} from '@/features/organization-units/types/organization-units';
import CreateUserTypePage from '../CreateUserTypePage';
import type {ApiError, CreateUserSchemaRequest} from '../../types/user-types';

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
  createUserType: (data: CreateUserSchemaRequest) => Promise<void>;
  loading: boolean;
  error: ApiError | null;
}

const mockUseCreateUserType = vi.fn<() => UseCreateUserTypeReturn>();

interface UseGetOrganizationUnitsReturn {
  data: {
    totalResults: number;
    startIndex: number;
    count: number;
    organizationUnits: {
      id: string;
      name: string;
      handle: string;
      description?: string | null;
      parent?: string | null;
    }[];
  } | null;
  isLoading: boolean;
  error: ApiError | null;
  refetch: (newParams?: OrganizationUnitListParams) => Promise<void>;
}

const mockUseGetOrganizationUnits = vi.fn<() => UseGetOrganizationUnitsReturn>();
const mockRefetchOrganizationUnits = vi.fn();

vi.mock('../../api/useCreateUserType', () => ({
  default: () => mockUseCreateUserType(),
}));

vi.mock('../../../organization-units/api/useGetOrganizationUnits', () => ({
  default: () => mockUseGetOrganizationUnits(),
}));

const getOrganizationUnitSelect = () => screen.getAllByRole('combobox')[0];
const getPropertyTypeSelect = (index = 0) => screen.getAllByRole('combobox')[index + 1];

const mockOrganizationUnitsResponse = {
  totalResults: 2,
  startIndex: 1,
  count: 2,
  organizationUnits: [
    {id: 'root-ou', name: 'Root Organization', handle: 'root', description: null, parent: null},
    {id: 'child-ou', name: 'Child Organization', handle: 'child', description: null, parent: 'root-ou'},
  ],
};

describe('CreateUserTypePage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseCreateUserType.mockReturnValue({
      createUserType: mockCreateUserType,
      loading: false,
      error: null,
    });
    mockUseGetOrganizationUnits.mockReturnValue({
      data: mockOrganizationUnitsResponse,
      isLoading: false,
      error: null,
      refetch: mockRefetchOrganizationUnits,
    });
  });

  it('renders the page with initial form', () => {
    render(<CreateUserTypePage />);

    expect(screen.getByRole('heading', {name: 'Add User Type'})).toBeInTheDocument();
    expect(screen.getByText('Define a new user type schema for your organization')).toBeInTheDocument();
    expect(screen.getByLabelText(/Type Name/i)).toBeInTheDocument();
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

    const cancelButton = screen.getByRole('button', {name: /Cancel/i});
    await user.click(cancelButton);

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/user-types');
    });
  });

  it('allows user to enter user type name', async () => {
    const user = userEvent.setup();
    render(<CreateUserTypePage />);

    const nameInput = screen.getByLabelText(/Type Name/i);
    await user.type(nameInput, 'Employee');

    expect(nameInput).toHaveValue('Employee');
  });

  it('allows adding a new property', async () => {
    const user = userEvent.setup();
    render(<CreateUserTypePage />);

    const addButton = screen.getByRole('button', {name: /Add Property/i});
    await user.click(addButton);
  });

  it('allows removing a property', async () => {
    const user = userEvent.setup();
    render(<CreateUserTypePage />);

    // Add a second property first
    const addButton = screen.getByRole('button', {name: /Add Property/i});
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

    const propertyNameInput = screen.getByPlaceholderText(/e\.g\., email, age, address/i);
    await user.type(propertyNameInput, 'email');

    expect(propertyNameInput).toHaveValue('email');
  });

  it('allows changing property type', async () => {
    const user = userEvent.setup();
    render(<CreateUserTypePage />);

    const typeSelect = getPropertyTypeSelect();
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

    const submitButton = screen.getByRole('button', {name: /Create User Type/i});
    await user.click(submitButton);

    await waitFor(() => {
      expect(screen.getByText('Please enter a user type name')).toBeInTheDocument();
    });

    expect(mockCreateUserType).not.toHaveBeenCalled();
  });

  it('renders organization units and allows selecting by name', async () => {
    const user = userEvent.setup();
    mockCreateUserType.mockResolvedValue(undefined);

    render(<CreateUserTypePage />);

    const ouSelect = getOrganizationUnitSelect();

    await waitFor(() => {
      expect(ouSelect).toHaveTextContent('Root Organization');
    });

    await user.click(ouSelect);
    const childOption = await screen.findByText('Child Organization');
    await user.click(childOption);

    await waitFor(() => {
      expect(ouSelect).toHaveTextContent('Child Organization');
    });

    await user.type(screen.getByLabelText(/Type Name/i), 'Employee');
    await user.type(screen.getByPlaceholderText(/e\.g\., email, age, address/i), 'email');

    const submitButton = screen.getByRole('button', {name: /Create User Type/i});
    await user.click(submitButton);

    await waitFor(() => {
      expect(mockCreateUserType).toHaveBeenCalledWith(
        expect.objectContaining({
          ouId: 'child-ou',
        }),
      );
    });
  });

  it('shows validation error when submitting without property name', async () => {
    const user = userEvent.setup();
    render(<CreateUserTypePage />);

    const nameInput = screen.getByLabelText(/Type Name/i);
    await user.type(nameInput, 'Employee');

    const submitButton = screen.getByRole('button', {name: /Create User Type/i});
    await user.click(submitButton);

    await waitFor(() => {
      expect(screen.getByText('Please add at least one property')).toBeInTheDocument();
    });

    expect(mockCreateUserType).not.toHaveBeenCalled();
  });

  it('shows validation error when submitting without organization unit id', async () => {
    const user = userEvent.setup();
    mockUseGetOrganizationUnits.mockReturnValue({
      data: {...mockOrganizationUnitsResponse, organizationUnits: []},
      isLoading: false,
      error: null,
      refetch: mockRefetchOrganizationUnits,
    });
    render(<CreateUserTypePage />);

    await user.type(screen.getByLabelText(/Type Name/i), 'Employee');
    await user.type(screen.getByPlaceholderText(/e\.g\., email, age, address/i), 'email');

    const submitButton = screen.getByRole('button', {name: /Create User Type/i});
    await user.click(submitButton);

    await waitFor(() => {
      expect(screen.getByText('Please provide an organization unit ID')).toBeInTheDocument();
    });

    expect(mockCreateUserType).not.toHaveBeenCalled();
  });

  it('shows validation error for duplicate property names', async () => {
    const user = userEvent.setup();
    render(<CreateUserTypePage />);

    const nameInput = screen.getByLabelText(/Type Name/i);
    await user.type(nameInput, 'Employee');

    // Add first property
    const firstPropertyInput = screen.getByPlaceholderText(/e\.g\., email, age, address/i);
    await user.type(firstPropertyInput, 'email');

    // Add second property
    const addButton = screen.getByRole('button', {name: /Add Property/i});
    await user.click(addButton);

    // Set same name for second property
    const propertyInputs = screen.getAllByPlaceholderText(/e\.g\., email, age, address/i);
    await user.type(propertyInputs[1], 'email');

    const submitButton = screen.getByRole('button', {name: /Create User Type/i});
    await user.click(submitButton);

    await waitFor(() => {
      expect(screen.getByText(/Duplicate property names found/i)).toBeInTheDocument();
    });

    expect(mockCreateUserType).not.toHaveBeenCalled();
  });

  it('allows toggling required checkbox', async () => {
    const user = userEvent.setup();
    render(<CreateUserTypePage />);

    const requiredCheckbox = screen.getByRole('checkbox', {name: /Required/i});
    expect(requiredCheckbox).not.toBeChecked();

    await user.click(requiredCheckbox);
    expect(requiredCheckbox).toBeChecked();

    await user.click(requiredCheckbox);
    expect(requiredCheckbox).not.toBeChecked();
  });

  it('allows toggling unique checkbox for string type', async () => {
    const user = userEvent.setup();
    render(<CreateUserTypePage />);

    const uniqueCheckbox = screen.getByRole('checkbox', {name: /Unique/i});
    expect(uniqueCheckbox).not.toBeChecked();

    await user.click(uniqueCheckbox);
    expect(uniqueCheckbox).toBeChecked();
  });

  it('hides unique checkbox for boolean type', async () => {
    const user = userEvent.setup();
    render(<CreateUserTypePage />);

    // Initially unique checkbox should be visible for string type
    expect(screen.getByRole('checkbox', {name: /Unique/i})).toBeInTheDocument();

    // Change to boolean type
    const typeSelect = getPropertyTypeSelect();
    await user.click(typeSelect);

    const booleanOption = await screen.findByText('Boolean');
    await user.click(booleanOption);

    await waitFor(() => {
      expect(screen.queryByRole('checkbox', {name: /Unique/i})).not.toBeInTheDocument();
    });
  });

  it('allows adding regex pattern for string type', async () => {
    const user = userEvent.setup();
    render(<CreateUserTypePage />);

    const regexInput = screen.getByPlaceholderText('e.g., ^[a-zA-Z0-9]+$');
    await user.click(regexInput);
    await user.paste('^[a-z]+$');

    expect(regexInput).toHaveValue('^[a-z]+$');
  });

  it('allows adding enum values for enum type', async () => {
    const user = userEvent.setup();
    render(<CreateUserTypePage />);

    // Change type to enum
    const typeSelect = getPropertyTypeSelect();
    await user.click(typeSelect);
    const enumOption = await screen.findByText('Enum');
    await user.click(enumOption);

    const enumInput = screen.getByPlaceholderText(/Add value and press Enter/i);
    await user.type(enumInput, 'admin');

    const addEnumButton = screen.getByRole('button', {name: /^Add$/i});
    await user.click(addEnumButton);

    await waitFor(() => {
      expect(screen.getByText('admin')).toBeInTheDocument();
    });

    expect(enumInput).toHaveValue('');
  });

  it('allows adding enum value by pressing Enter', async () => {
    const user = userEvent.setup();
    render(<CreateUserTypePage />);

    // Change type to enum
    const typeSelect = getPropertyTypeSelect();
    await user.click(typeSelect);
    const enumOption = await screen.findByText('Enum');
    await user.click(enumOption);

    const enumInput = screen.getByPlaceholderText(/Add value and press Enter/i);
    await user.type(enumInput, 'user{Enter}');

    await waitFor(() => {
      expect(screen.getByText('user')).toBeInTheDocument();
    });

    expect(enumInput).toHaveValue('');
  });

  it('does not add enum value when input is empty or whitespace', async () => {
    const user = userEvent.setup();
    render(<CreateUserTypePage />);

    // Change type to enum
    const typeSelect = getPropertyTypeSelect();
    await user.click(typeSelect);
    const enumOption = await screen.findByText('Enum');
    await user.click(enumOption);

    const enumInput = screen.getByPlaceholderText(/Add value and press Enter/i);

    // Try to add empty value by clicking button
    const addEnumButton = screen.getByRole('button', {name: /^Add$/i});
    await user.click(addEnumButton);

    // Should not add any enum values
    const enumContainer = enumInput.closest('div')?.querySelector('.MuiBox-root');
    expect(enumContainer).not.toBeInTheDocument();

    // Try to add whitespace value
    await user.type(enumInput, '   ');
    await user.click(addEnumButton);

    // Should still not add any enum values
    expect(enumContainer).not.toBeInTheDocument();
  });

  it('allows removing enum values', async () => {
    const user = userEvent.setup();
    render(<CreateUserTypePage />);

    // Change type to enum
    const typeSelect = getPropertyTypeSelect();
    await user.click(typeSelect);
    const enumOption = await screen.findByText('Enum');
    await user.click(enumOption);

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
    const nameInput = screen.getByLabelText(/Type Name/i);
    await user.type(nameInput, 'Employee');

    // Fill in property name
    const propertyNameInput = screen.getByPlaceholderText(/e.g., email, age, address/i);
    await user.type(propertyNameInput, 'email');

    // Mark as required
    const requiredCheckbox = screen.getByRole('checkbox', {name: /Required/i});
    await user.click(requiredCheckbox);

    // Submit form
    const submitButton = screen.getByRole('button', {name: /Create User Type/i});
    await user.click(submitButton);

    await waitFor(() => {
      expect(mockCreateUserType).toHaveBeenCalledWith({
        name: 'Employee',
        ouId: 'root-ou',
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

  it('submits organization unit and registration flag when provided', async () => {
    const user = userEvent.setup();
    mockCreateUserType.mockResolvedValue(undefined);

    render(<CreateUserTypePage />);

    await user.type(screen.getByLabelText(/Type Name/i), 'Employee');
    const ouSelect = getOrganizationUnitSelect();
    await user.click(ouSelect);
    const secondOuOption = await screen.findByText('Child Organization');
    await user.click(secondOuOption);

    const propertyNameInput = screen.getByPlaceholderText(/e\.g\., email, age, address/i);
    await user.type(propertyNameInput, 'email');

    await user.click(screen.getByLabelText(/Allow Self Registration/i));

    await user.click(screen.getByRole('button', {name: /Create User Type/i}));

    await waitFor(() => {
      expect(mockCreateUserType).toHaveBeenCalledWith({
        name: 'Employee',
        ouId: 'child-ou',
        allowSelfRegistration: true,
        schema: {
          email: {
            type: 'string',
            required: false,
          },
        },
      });
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

    expect(screen.getByText('Saving...')).toBeInTheDocument();
    expect(screen.getByRole('button', {name: /Saving/i})).toBeDisabled();
    expect(screen.getByRole('button', {name: /Cancel/i})).toBeDisabled();
  });

  it('closes snackbar when close button is clicked', async () => {
    const user = userEvent.setup();
    render(<CreateUserTypePage />);

    // Trigger validation error
    const submitButton = screen.getByRole('button', {name: /Create User Type/i});
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

  it('creates schema with enum property correctly', async () => {
    const user = userEvent.setup();
    mockCreateUserType.mockResolvedValue(undefined);

    render(<CreateUserTypePage />);

    // Set user type name
    const nameInput = screen.getByLabelText(/Type Name/i);
    await user.type(nameInput, 'Complex Type');

    // Change type to enum
    const typeSelect = getPropertyTypeSelect();
    await user.click(typeSelect);
    const enumOption = await screen.findByText('Enum');
    await user.click(enumOption);

    // Add enum property with all features
    const firstPropertyInput = screen.getByPlaceholderText(/e.g., email, age, address/i);
    await user.type(firstPropertyInput, 'status');

    const requiredCheckbox = screen.getByRole('checkbox', {name: /Required/i});
    await user.click(requiredCheckbox);

    const uniqueCheckbox = screen.getByRole('checkbox', {name: /Unique/i});
    await user.click(uniqueCheckbox);

    const enumInput = screen.getByPlaceholderText(/Add value and press Enter/i);
    await user.type(enumInput, 'ACTIVE{Enter}');
    await user.type(enumInput, 'INACTIVE{Enter}');

    // Submit
    const submitButton = screen.getByRole('button', {name: /Create User Type/i});
    await user.click(submitButton);

    await waitFor(() => {
      expect(mockCreateUserType).toHaveBeenCalledWith({
        name: 'Complex Type',
        ouId: 'root-ou',
        schema: {
          status: {
            type: 'string',
            required: true,
            unique: true,
            enum: ['ACTIVE', 'INACTIVE'],
          },
        },
      });
    });
  });

  it('resets type-specific fields when type changes', async () => {
    const user = userEvent.setup();
    render(<CreateUserTypePage />);

    // Change type to enum first
    const typeSelect = getPropertyTypeSelect();
    await user.click(typeSelect);
    const enumTypeOption = await screen.findByText('Enum');
    await user.click(enumTypeOption);

    // Add enum value for enum type
    const enumInput = screen.getByPlaceholderText(/Add value and press Enter/i);
    await user.type(enumInput, 'test{Enter}');

    await waitFor(() => {
      expect(screen.getByText('test')).toBeInTheDocument();
    });

    // Change type to number
    await user.click(typeSelect);

    const numberOption = await screen.findByText('Number');
    await user.click(numberOption);

    // Enum values should be gone
    await waitFor(() => {
      expect(screen.queryByText('test')).not.toBeInTheDocument();
    });

    // Regex and enum inputs should not be visible for number type
    expect(screen.queryByPlaceholderText(/userTypes:enumPlaceholder/i)).not.toBeInTheDocument();
    expect(screen.queryByPlaceholderText(/userTypes:regexPlaceholder/i)).not.toBeInTheDocument();
  });

  it('creates schema with string property containing regex pattern', async () => {
    const user = userEvent.setup();
    mockCreateUserType.mockResolvedValue(undefined);

    render(<CreateUserTypePage />);

    // Set user type name
    const nameInput = screen.getByLabelText(/Type Name/i);
    await user.type(nameInput, 'RegexTest');

    // Add property name
    const propertyNameInput = screen.getByPlaceholderText(/e.g., email, age, address/i);
    await user.type(propertyNameInput, 'code');

    // Add regex pattern using paste to avoid special character parsing issues
    const regexInput = screen.getByPlaceholderText('e.g., ^[a-zA-Z0-9]+$');
    await user.click(regexInput);
    await user.paste('^[A-Z]{3}$');

    // Submit
    const submitButton = screen.getByRole('button', {name: /Create User Type/i});
    await user.click(submitButton);

    await waitFor(() => {
      expect(mockCreateUserType).toHaveBeenCalledWith({
        name: 'RegexTest',
        ouId: 'root-ou',
        schema: {
          code: {
            type: 'string',
            required: false,
            regex: '^[A-Z]{3}$',
          },
        },
      });
    });
  });

  it('creates schema with number property that is unique', async () => {
    const user = userEvent.setup();
    mockCreateUserType.mockResolvedValue(undefined);

    render(<CreateUserTypePage />);

    // Set user type name
    const nameInput = screen.getByLabelText(/Type Name/i);
    await user.type(nameInput, 'NumberTest');

    // Add property name
    const propertyNameInput = screen.getByPlaceholderText(/e.g., email, age, address/i);
    await user.type(propertyNameInput, 'employeeId');

    // Change type to number
    const typeSelect = getPropertyTypeSelect();
    await user.click(typeSelect);
    const numberOption = await screen.findByText('Number');
    await user.click(numberOption);

    // Mark as unique
    const uniqueCheckbox = screen.getByRole('checkbox', {name: /Unique/i});
    await user.click(uniqueCheckbox);

    // Submit
    const submitButton = screen.getByRole('button', {name: /Create User Type/i});
    await user.click(submitButton);

    await waitFor(() => {
      expect(mockCreateUserType).toHaveBeenCalledWith({
        name: 'NumberTest',
        ouId: 'root-ou',
        schema: {
          employeeId: {
            type: 'number',
            required: false,
            unique: true,
          },
        },
      });
    });
  });

  it('handles create error gracefully', async () => {
    const user = userEvent.setup();
    mockCreateUserType.mockRejectedValue(new Error('Create failed'));

    render(<CreateUserTypePage />);

    await user.type(screen.getByLabelText(/Type Name/i), 'Employee');
    await user.type(screen.getByPlaceholderText(/e\.g\., email, age, address/i), 'email');

    const submitButton = screen.getByRole('button', {name: /Create User Type/i});
    await user.click(submitButton);

    await waitFor(() => {
      expect(mockCreateUserType).toHaveBeenCalled();
      // Should not navigate on error
      expect(mockNavigate).not.toHaveBeenCalledWith('/user-types');
    });
  });

  it('shows organization units error alert when fetch fails', () => {
    const orgError: ApiError = {
      code: 'ORG_ERROR',
      message: 'API error message',
      description: 'Error description',
    };

    mockUseGetOrganizationUnits.mockReturnValue({
      data: null,
      isLoading: false,
      error: orgError,
      refetch: mockRefetchOrganizationUnits,
    });

    render(<CreateUserTypePage />);

    // The title uses the translation key, and the message is the API error message
    expect(screen.getByText('Failed to load organization units')).toBeInTheDocument();
    expect(screen.getByText('API error message')).toBeInTheDocument();
  });

  it('shows loading state in organization unit dropdown', async () => {
    const user = userEvent.setup();

    mockUseGetOrganizationUnits.mockReturnValue({
      data: null,
      isLoading: true,
      error: null,
      refetch: mockRefetchOrganizationUnits,
    });

    render(<CreateUserTypePage />);

    const ouSelect = getOrganizationUnitSelect();
    await user.click(ouSelect);

    await waitFor(() => {
      expect(screen.getByText('Loading...')).toBeInTheDocument();
    });
  });

  it('shows no organization units message when list is empty', async () => {
    const user = userEvent.setup();

    mockUseGetOrganizationUnits.mockReturnValue({
      data: {...mockOrganizationUnitsResponse, organizationUnits: []},
      isLoading: false,
      error: null,
      refetch: mockRefetchOrganizationUnits,
    });

    render(<CreateUserTypePage />);

    const ouSelect = getOrganizationUnitSelect();
    await user.click(ouSelect);

    await waitFor(() => {
      expect(screen.getByText('No organization units available')).toBeInTheDocument();
    });
  });

  it('shows error in organization unit dropdown when fetch fails', async () => {
    const user = userEvent.setup();
    const orgError: ApiError = {
      code: 'ORG_ERROR',
      message: 'Org fetch failed',
      description: 'Error',
    };

    mockUseGetOrganizationUnits.mockReturnValue({
      data: null,
      isLoading: false,
      error: orgError,
      refetch: mockRefetchOrganizationUnits,
    });

    render(<CreateUserTypePage />);

    const ouSelect = getOrganizationUnitSelect();
    await user.click(ouSelect);

    // The error message appears in the dropdown menu item
    await waitFor(() => {
      const menuItems = screen.getAllByRole('option');
      expect(menuItems.some((item) => item.textContent?.includes('Org fetch failed'))).toBe(true);
    });
  });

  it('displays organization unit placeholder when no value is selected', async () => {
    mockUseGetOrganizationUnits.mockReturnValue({
      data: {...mockOrganizationUnitsResponse, organizationUnits: []},
      isLoading: false,
      error: null,
      refetch: mockRefetchOrganizationUnits,
    });

    render(<CreateUserTypePage />);

    const ouSelect = getOrganizationUnitSelect();
    expect(ouSelect).toHaveTextContent('Select an organization unit');
  });

  it('displays selected organization unit name in renderValue', async () => {
    const user = userEvent.setup();

    render(<CreateUserTypePage />);

    // The component should auto-select the first organization unit
    const ouSelect = getOrganizationUnitSelect();

    await waitFor(() => {
      expect(ouSelect).toHaveTextContent('Root Organization');
    });

    // Click to open dropdown and select a different unit
    await user.click(ouSelect);
    const childOption = await screen.findByText('Child Organization');
    await user.click(childOption);

    await waitFor(() => {
      expect(ouSelect).toHaveTextContent('Child Organization');
    });
  });
});
