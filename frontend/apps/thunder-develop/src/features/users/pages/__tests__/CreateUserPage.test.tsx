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
import {render, screen, waitFor, userEvent} from '@thunder/test-utils';
import CreateUserPage from '../CreateUserPage';
import type {ApiError, UserSchemaListResponse, ApiUserSchema} from '../../types/users';
import type {CreateUserResponse} from '../../api/useCreateUser';

const mockNavigate = vi.fn();
const mockCreateUser = vi.fn();
const mockResetCreateUser = vi.fn();
const mockRefetchSchemas = vi.fn();
const mockRefetchSchema = vi.fn();

// Mock react-router
vi.mock('react-router', async () => {
  const actual = await vi.importActual<typeof import('react-router')>('react-router');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

// Mock hooks
interface UseCreateUserReturn {
  createUser: (data: {
    organizationUnit: string;
    type: string;
    attributes: Record<string, unknown>;
  }) => Promise<CreateUserResponse>;
  data: CreateUserResponse | null;
  loading: boolean;
  error: ApiError | null;
  reset: () => void;
}

interface UseGetUserSchemasReturn {
  data: UserSchemaListResponse | null;
  loading: boolean;
  error: ApiError | null;
  refetch: () => void;
}

interface UseGetUserSchemaReturn {
  data: ApiUserSchema | null;
  loading: boolean;
  error: ApiError | null;
  refetch: (id?: string) => void;
}

const mockUseCreateUser = vi.fn<() => UseCreateUserReturn>();
const mockUseGetUserSchemas = vi.fn<() => UseGetUserSchemasReturn>();
const mockUseGetUserSchema = vi.fn<() => UseGetUserSchemaReturn>();

vi.mock('../../api/useCreateUser', () => ({
  default: () => mockUseCreateUser(),
}));

vi.mock('../../api/useGetUserSchemas', () => ({
  default: () => mockUseGetUserSchemas(),
}));

vi.mock('../../api/useGetUserSchema', () => ({
  default: () => mockUseGetUserSchema(),
}));

describe('CreateUserPage', () => {
  const mockSchemasData: UserSchemaListResponse = {
    totalResults: 2,
    startIndex: 1,
    count: 2,
    schemas: [
      {id: 'schema1', name: 'Employee', ouId: 'root-ou'},
      {id: 'schema2', name: 'Contractor', ouId: 'child-ou'},
    ],
  };

  const mockSchemaData: ApiUserSchema = {
    id: 'schema1',
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
    },
  };

  beforeEach(() => {
    vi.clearAllMocks();
    mockUseCreateUser.mockReturnValue({
      createUser: mockCreateUser,
      data: null,
      loading: false,
      error: null,
      reset: mockResetCreateUser,
    });
    mockUseGetUserSchemas.mockReturnValue({
      data: mockSchemasData,
      loading: false,
      error: null,
      refetch: mockRefetchSchemas,
    });
    mockUseGetUserSchema.mockReturnValue({
      data: mockSchemaData,
      loading: false,
      error: null,
      refetch: mockRefetchSchema,
    });
  });

  it('renders the page with title and description', () => {
    render(<CreateUserPage />);

    expect(screen.getByRole('heading', {name: 'Create User'})).toBeInTheDocument();
    expect(screen.getByText('Add a new user to your organization')).toBeInTheDocument();
  });

  it('renders user type select with schemas', () => {
    render(<CreateUserPage />);

    expect(screen.getByText('User Type')).toBeInTheDocument();
    expect(screen.getByRole('combobox')).toBeInTheDocument();
    expect(screen.getByText('Employee')).toBeInTheDocument();
  });

  it('navigates back when Back button is clicked', async () => {
    const user = userEvent.setup();
    render(<CreateUserPage />);

    const backButton = screen.getByRole('button', {name: /go back/i});
    await user.click(backButton);

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/users');
    });
  });

  it('navigates back when Cancel button is clicked', async () => {
    const user = userEvent.setup();
    render(<CreateUserPage />);

    const cancelButton = screen.getByRole('button', {name: /cancel/i});
    await user.click(cancelButton);

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/users');
    });
  });

  it('renders Create button for user type', () => {
    render(<CreateUserPage />);

    const createButtons = screen.getAllByRole('button', {name: /create/i});
    const createTypeButton = createButtons.find(
      (button) =>
        button.textContent?.includes('Create') && button !== screen.getByRole('button', {name: /create user/i}),
    );

    expect(createTypeButton).toBeInTheDocument();
  });

  it('logs console when Create user type button is clicked', async () => {
    const consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => undefined);
    const user = userEvent.setup();
    render(<CreateUserPage />);

    const createButtons = screen.getAllByRole('button', {name: /create/i});
    const createTypeButton = createButtons.find((button) => button.textContent === 'Create');

    if (createTypeButton) {
      await user.click(createTypeButton);
      expect(consoleSpy).toHaveBeenCalledWith('Navigate to create user type page');
    }

    consoleSpy.mockRestore();
  });

  it('allows changing user type', async () => {
    const user = userEvent.setup();
    render(<CreateUserPage />);

    const select = screen.getByRole('combobox');
    await user.click(select);

    const contractorOption = await screen.findByText('Contractor');
    await user.click(contractorOption);

    await waitFor(() => {
      expect(select).toHaveTextContent('Contractor');
    });
  });

  it('displays loading state for schema fields', () => {
    mockUseGetUserSchema.mockReturnValue({
      data: null,
      loading: true,
      error: null,
      refetch: mockRefetchSchema,
    });

    render(<CreateUserPage />);

    expect(screen.getByText('Loading schema fields...')).toBeInTheDocument();
  });

  it('displays error when schema fails to load', () => {
    const error: ApiError = {
      code: 'SCHEMA_ERROR',
      message: 'Failed to load schema',
      description: 'Schema not found',
    };

    mockUseGetUserSchema.mockReturnValue({
      data: null,
      loading: false,
      error,
      refetch: mockRefetchSchema,
    });

    render(<CreateUserPage />);

    expect(screen.getByText(/Error loading schema: Failed to load schema/i)).toBeInTheDocument();
  });

  it('renders schema fields when loaded', () => {
    render(<CreateUserPage />);

    expect(screen.getByPlaceholderText(/Enter username/i)).toBeInTheDocument();
    expect(screen.getByPlaceholderText(/Enter age/i)).toBeInTheDocument();
  });

  it('allows entering values in schema fields', async () => {
    const user = userEvent.setup();
    render(<CreateUserPage />);

    const usernameInput = screen.getByPlaceholderText(/Enter username/i);
    await user.type(usernameInput, 'john_doe');

    expect(usernameInput).toHaveValue('john_doe');

    const ageInput = screen.getByPlaceholderText(/Enter age/i);
    await user.type(ageInput, '30');

    expect(ageInput).toHaveValue(30);
  });

  it('displays validation error when required fields are missing', async () => {
    const user = userEvent.setup();
    render(<CreateUserPage />);

    const submitButton = screen.getByRole('button', {name: /create user/i});
    await user.click(submitButton);

    await waitFor(() => {
      expect(screen.getByText('username is required')).toBeInTheDocument();
    });

    expect(mockCreateUser).not.toHaveBeenCalled();
  });

  it('successfully creates user with valid data', async () => {
    const user = userEvent.setup();
    const mockResponse: CreateUserResponse = {
      id: 'user123',
      organizationUnit: 'test-ou',
      type: 'Employee',
      attributes: {
        username: 'john_doe',
        age: 30,
      },
    };
    mockCreateUser.mockResolvedValue(mockResponse);

    render(<CreateUserPage />);

    const usernameInput = screen.getByPlaceholderText(/Enter username/i);
    await user.type(usernameInput, 'john_doe');

    const ageInput = screen.getByPlaceholderText(/Enter age/i);
    await user.type(ageInput, '30');

    const submitButton = screen.getByRole('button', {name: /create user/i});
    await user.click(submitButton);

    await waitFor(() => {
      expect(mockCreateUser).toHaveBeenCalledWith({
        organizationUnit: 'root-ou',
        type: 'Employee',
        attributes: {
          username: 'john_doe',
          age: 30,
        },
      });
    });

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/users');
    });
  });

  it('displays error from create user API', () => {
    const error: ApiError = {
      code: 'CREATE_ERROR',
      message: 'Failed to create user',
      description: 'User already exists',
    };

    mockUseCreateUser.mockReturnValue({
      createUser: mockCreateUser,
      data: null,
      loading: false,
      error,
      reset: mockResetCreateUser,
    });

    render(<CreateUserPage />);

    expect(screen.getByText('Failed to create user')).toBeInTheDocument();
    expect(screen.getByText('User already exists')).toBeInTheDocument();
  });

  it('does not submit when selected schema is missing organization unit', async () => {
    const user = userEvent.setup();
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => undefined);

    mockUseGetUserSchemas.mockReturnValue({
      data: {
        ...mockSchemasData,
        schemas: [{...mockSchemasData.schemas[0], ouId: ''}],
      },
      loading: false,
      error: null,
      refetch: mockRefetchSchemas,
    });

    render(<CreateUserPage />);

    const usernameInput = screen.getByPlaceholderText(/Enter username/i);
    await user.type(usernameInput, 'john_doe');

    const submitButton = screen.getByRole('button', {name: /create user/i});
    await user.click(submitButton);

    await waitFor(() => {
      expect(mockCreateUser).not.toHaveBeenCalled();
      expect(consoleSpy).toHaveBeenCalledWith('Failed to create user:', expect.any(Error));
    });

    consoleSpy.mockRestore();
  });

  it('shows loading state during submission', async () => {
    const user = userEvent.setup();
    let resolveCreateUser: ((value: CreateUserResponse) => void) | undefined;
    const createUserPromise = new Promise<CreateUserResponse>((resolve) => {
      resolveCreateUser = resolve;
    });
    mockCreateUser.mockReturnValue(createUserPromise);

    render(<CreateUserPage />);

    const usernameInput = screen.getByPlaceholderText(/Enter username/i);
    await user.type(usernameInput, 'john_doe');

    const submitButton = screen.getByRole('button', {name: /create user/i});
    await user.click(submitButton);

    // Wait for the loading state to appear
    await waitFor(() => {
      expect(screen.getByText('Creating...')).toBeInTheDocument();
      expect(screen.getByRole('button', {name: /creating.../i})).toBeDisabled();
      expect(screen.getByRole('button', {name: /cancel/i})).toBeDisabled();
    });

    // Resolve the promise to clean up and wait for state updates
    if (resolveCreateUser) {
      resolveCreateUser({
        id: 'user123',
        organizationUnit: 'test-ou',
        type: 'Employee',
        attributes: {username: 'john_doe'},
      });
    }

    // Wait for the promise to resolve and state to update
    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalled();
    });
  });

  it('disables submit and cancel buttons during submission', async () => {
    const user = userEvent.setup();
    mockCreateUser.mockImplementation(() => new Promise(() => {})); // Never resolves

    render(<CreateUserPage />);

    const usernameInput = screen.getByPlaceholderText(/Enter username/i);
    await user.type(usernameInput, 'john_doe');

    const submitButton = screen.getByRole('button', {name: /create user/i});
    await user.click(submitButton);

    await waitFor(() => {
      expect(screen.getByRole('button', {name: /creating.../i})).toBeDisabled();
      expect(screen.getByRole('button', {name: /cancel/i})).toBeDisabled();
    });
  });

  it('handles empty schemas list', () => {
    mockUseGetUserSchemas.mockReturnValue({
      data: {
        totalResults: 0,
        startIndex: 1,
        count: 0,
        schemas: [],
      },
      loading: false,
      error: null,
      refetch: mockRefetchSchemas,
    });

    render(<CreateUserPage />);

    expect(screen.getByText('Loading schemas...')).toBeInTheDocument();
  });

  it('handles null schemas data', () => {
    mockUseGetUserSchemas.mockReturnValue({
      data: null,
      loading: false,
      error: null,
      refetch: mockRefetchSchemas,
    });

    render(<CreateUserPage />);

    expect(screen.getByText('Loading schemas...')).toBeInTheDocument();
  });

  it('sets first schema as default when schemas load', () => {
    render(<CreateUserPage />);

    const select = screen.getByRole('combobox');
    expect(select).toHaveTextContent('Employee');
  });

  it('renders with different schema field types', () => {
    const complexSchema: ApiUserSchema = {
      id: 'schema1',
      name: 'Employee',
      schema: {
        email: {
          type: 'string',
          required: true,
        },
        salary: {
          type: 'number',
          required: true,
        },
        active: {
          type: 'boolean',
          required: false,
        },
        tags: {
          type: 'array',
          items: {
            type: 'string',
          },
          required: false,
        },
      },
    };

    mockUseGetUserSchema.mockReturnValue({
      data: complexSchema,
      loading: false,
      error: null,
      refetch: mockRefetchSchema,
    });

    render(<CreateUserPage />);

    expect(screen.getByPlaceholderText(/Enter email/i)).toBeInTheDocument();
    expect(screen.getByPlaceholderText(/Enter salary/i)).toBeInTheDocument();
    expect(screen.getByRole('checkbox')).toBeInTheDocument();
    expect(screen.getByPlaceholderText(/Add tags/i)).toBeInTheDocument();
  });

  it('handles form submission with all field types', async () => {
    const user = userEvent.setup();
    const complexSchema: ApiUserSchema = {
      id: 'schema1',
      name: 'Employee',
      schema: {
        email: {
          type: 'string',
          required: true,
        },
        salary: {
          type: 'number',
          required: true,
        },
        active: {
          type: 'boolean',
          required: false,
        },
      },
    };

    mockUseGetUserSchema.mockReturnValue({
      data: complexSchema,
      loading: false,
      error: null,
      refetch: mockRefetchSchema,
    });

    const mockResponse: CreateUserResponse = {
      id: 'user123',
      organizationUnit: 'test-ou',
      type: 'Employee',
      attributes: {
        email: 'john@example.com',
        salary: 50000,
        active: true,
      },
    };
    mockCreateUser.mockResolvedValue(mockResponse);

    render(<CreateUserPage />);

    const emailInput = screen.getByPlaceholderText(/Enter email/i);
    await user.type(emailInput, 'john@example.com');

    const salaryInput = screen.getByPlaceholderText(/Enter salary/i);
    await user.type(salaryInput, '50000');

    const activeCheckbox = screen.getByRole('checkbox');
    await user.click(activeCheckbox);

    const submitButton = screen.getByRole('button', {name: /create user/i});
    await user.click(submitButton);

    await waitFor(() => {
      expect(mockCreateUser).toHaveBeenCalledWith({
        organizationUnit: 'root-ou',
        type: 'Employee',
        attributes: {
          email: 'john@example.com',
          salary: 50000,
          active: true,
        },
      });
    });
  });

  it('handles exception during user creation and logs error', async () => {
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => undefined);
    const user = userEvent.setup();
    const error = new Error('Network error');
    mockCreateUser.mockRejectedValue(error);

    render(<CreateUserPage />);

    const usernameInput = screen.getByPlaceholderText(/Enter username/i);
    await user.type(usernameInput, 'john_doe');

    const submitButton = screen.getByRole('button', {name: /create user/i});
    await user.click(submitButton);

    await waitFor(() => {
      expect(mockCreateUser).toHaveBeenCalled();
    });

    await waitFor(() => {
      expect(consoleSpy).toHaveBeenCalledWith('Failed to create user:', error);
    });

    // Should reset submitting state
    expect(screen.getByRole('button', {name: /create user/i})).not.toBeDisabled();

    consoleSpy.mockRestore();
  });

  it('uses selectedSchema name when field value is undefined', async () => {
    const user = userEvent.setup();
    render(<CreateUserPage />);

    // The select should initially show the first schema name
    const select = screen.getByRole('combobox');
    expect(select).toHaveTextContent('Employee');

    // Change to another schema
    await user.click(select);
    const contractorOption = await screen.findByText('Contractor');
    await user.click(contractorOption);

    await waitFor(() => {
      expect(select).toHaveTextContent('Contractor');
    });
  });

  it('shows validation error when submitting without required fields', async () => {
    const user = userEvent.setup();

    mockUseGetUserSchemas.mockReturnValue({
      data: mockSchemasData,
      loading: false,
      error: null,
      refetch: mockRefetchSchemas,
    });

    render(<CreateUserPage />);

    // Submit without filling required fields
    const submitButton = screen.getByRole('button', {name: /create user/i});
    await user.click(submitButton);

    // Check that validation runs - username is required
    await waitFor(() => {
      expect(screen.getByText('username is required')).toBeInTheDocument();
    });
  });
});
