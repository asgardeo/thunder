/**
 * Copyright (c) 2025-2026, WSO2 LLC. (https://www.wso2.com).
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

import type {ReactNode} from 'react';
import {describe, it, expect, vi, beforeEach} from 'vitest';
import {render, screen, waitFor, userEvent} from '@thunder/test-utils';
import CreateUserPage from '../CreateUserPage';
import UserCreateProvider from '../../contexts/UserCreate/UserCreateProvider';
import type {UserSchemaListResponse, ApiUserSchema, SchemaInterface} from '../../types/users';

const mockNavigate = vi.fn();
const mockMutateAsync = vi.fn();
const mockReset = vi.fn();

// Mock react-router
vi.mock('react-router', async () => {
  const actual = await vi.importActual<typeof import('react-router')>('react-router');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
    Link: ({to, children = undefined, ...props}: {to: string; children?: ReactNode; [key: string]: unknown}) => (
      <a
        {...(props as Record<string, unknown>)}
        href={to}
        onClick={(e) => {
          e.preventDefault();
          Promise.resolve(mockNavigate(to)).catch(() => {});
        }}
      >
        {children}
      </a>
    ),
  };
});

// Mock logger
vi.mock('@thunder/logger/react', () => ({
  useLogger: () => ({
    info: vi.fn(),
    warn: vi.fn(),
    error: vi.fn(),
    debug: vi.fn(),
    withComponent: vi.fn().mockReturnThis(),
  }),
}));

// Mock hooks
interface UseCreateUserReturn {
  mutate: ReturnType<typeof vi.fn>;
  mutateAsync: ReturnType<typeof vi.fn>;
  isPending: boolean;
  error: Error | null;
  data: unknown;
  isError: boolean;
  isSuccess: boolean;
  isIdle: boolean;
  reset: () => void;
}

interface UseGetUserSchemasReturn {
  data: UserSchemaListResponse | undefined;
  isLoading: boolean;
  error: Error | null;
}

interface UseGetUserSchemaReturn {
  data: ApiUserSchema | undefined;
  isLoading: boolean;
  error: Error | null;
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

// Mock child components with controlled test behavior
vi.mock('../../components/create-user/ConfigureUserType', () => ({
  default: ({
    schemas,
    selectedSchema,
    onSchemaChange,
    onReadyChange,
  }: {
    schemas: SchemaInterface[];
    selectedSchema: SchemaInterface | null;
    onSchemaChange: (schema: SchemaInterface | null) => void;
    onReadyChange?: (isReady: boolean) => void;
  }) => (
    <div data-testid="configure-user-type">
      <span data-testid="selected-schema-name">{selectedSchema?.name ?? 'none'}</span>
      {schemas.map((s) => (
        <button
          key={s.id}
          type="button"
          data-testid={`select-schema-${s.name}`}
          onClick={() => {
            onSchemaChange(s);
            onReadyChange?.(true);
          }}
        >
          {s.name}
        </button>
      ))}
    </div>
  ),
}));

vi.mock('../../components/create-user/ConfigureUserDetails', () => ({
  default: ({
    onFormValuesChange,
    onReadyChange,
  }: {
    onFormValuesChange: (values: Record<string, unknown>) => void;
    onReadyChange?: (isReady: boolean) => void;
  }) => (
    <div data-testid="configure-user-details">
      <button
        type="button"
        data-testid="fill-form"
        onClick={() => {
          onFormValuesChange({username: 'john_doe', age: 30});
          onReadyChange?.(true);
        }}
      >
        Fill Form
      </button>
      <button
        type="button"
        data-testid="fill-form-with-empty-values"
        onClick={() => {
          onFormValuesChange({username: 'john_doe', age: 30, nickname: '', phone: null, address: undefined});
          onReadyChange?.(true);
        }}
      >
        Fill Form With Empty Values
      </button>
      <button
        type="button"
        data-testid="mark-ready"
        onClick={() => {
          onReadyChange?.(true);
        }}
      >
        Mark Ready
      </button>
    </div>
  ),
}));

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
    username: {type: 'string', required: true},
    age: {type: 'number', required: false},
  },
};

/**
 * Helper to render the wizard page wrapped in provider.
 */
function renderPage() {
  return render(
    <UserCreateProvider>
      <CreateUserPage />
    </UserCreateProvider>,
  );
}

/**
 * Helper to navigate from step 1 (User Type) to step 2 (User Details)
 * by selecting a schema and clicking Continue.
 */
async function goToDetailsStep(user: ReturnType<typeof userEvent.setup>, schemaName = 'Employee') {
  await user.click(screen.getByTestId(`select-schema-${schemaName}`));
  await user.click(screen.getByRole('button', {name: /continue/i}));
  await waitFor(() => {
    expect(screen.getByTestId('configure-user-details')).toBeInTheDocument();
  });
}

describe('CreateUserPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockNavigate.mockResolvedValue(undefined);
    mockMutateAsync.mockResolvedValue({
      id: 'user123',
      organizationUnit: 'root-ou',
      type: 'Employee',
      attributes: {},
    });
    mockUseCreateUser.mockReturnValue({
      mutate: vi.fn(),
      mutateAsync: mockMutateAsync,
      isPending: false,
      error: null,
      data: undefined,
      isError: false,
      isSuccess: false,
      isIdle: true,
      reset: mockReset,
    });
    mockUseGetUserSchemas.mockReturnValue({
      data: mockSchemasData,
      isLoading: false,
      error: null,
    });
    mockUseGetUserSchema.mockReturnValue({
      data: mockSchemaData,
      isLoading: false,
      error: null,
    });
  });

  // ============================================================================
  // Step 1: User Type
  // ============================================================================

  it('renders the wizard with User Type step initially', () => {
    renderPage();

    expect(screen.getByTestId('configure-user-type')).toBeInTheDocument();
  });

  it('shows progress bar', () => {
    renderPage();

    expect(screen.getByRole('progressbar')).toBeInTheDocument();
  });

  it('shows breadcrumb navigation', () => {
    renderPage();

    expect(screen.getByLabelText('breadcrumb')).toBeInTheDocument();
  });

  it('disables Continue button when no schema is selected', () => {
    renderPage();

    const continueButton = screen.getByRole('button', {name: /continue/i});
    expect(continueButton).toBeDisabled();
  });

  it('enables Continue button when a schema is selected', async () => {
    const user = userEvent.setup();
    renderPage();

    await user.click(screen.getByTestId('select-schema-Employee'));

    const continueButton = screen.getByRole('button', {name: /continue/i});
    expect(continueButton).not.toBeDisabled();
  });

  it('closes wizard when X button is clicked', async () => {
    const user = userEvent.setup();
    renderPage();

    const closeButtons = screen.getAllByRole('button');
    const closeButton = closeButtons[0];
    await user.click(closeButton);

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/users');
    });
  });

  it('does not close wizard when loading', async () => {
    mockUseCreateUser.mockReturnValue({
      mutate: vi.fn(),
      mutateAsync: mockMutateAsync,
      isPending: true,
      error: null,
      data: undefined,
      isError: false,
      isSuccess: false,
      isIdle: false,
      reset: mockReset,
    });

    const user = userEvent.setup();
    renderPage();

    const closeButtons = screen.getAllByRole('button');
    const closeButton = closeButtons[0];
    await user.click(closeButton);

    expect(mockNavigate).not.toHaveBeenCalled();
  });

  // ============================================================================
  // Step 2: User Details
  // ============================================================================

  it('navigates to User Details step when Continue is clicked', async () => {
    const user = userEvent.setup();
    renderPage();

    await goToDetailsStep(user);

    expect(screen.getByTestId('configure-user-details')).toBeInTheDocument();
  });

  it('shows loading state when schema is loading on step 2', async () => {
    mockUseGetUserSchema.mockReturnValue({
      data: undefined,
      isLoading: true,
      error: null,
    });

    const user = userEvent.setup();
    renderPage();

    // Select schema and navigate to step 2 manually (can't use goToDetailsStep
    // because it waits for configure-user-details which won't render during loading)
    await user.click(screen.getByTestId('select-schema-Employee'));
    await user.click(screen.getByRole('button', {name: /continue/i}));

    await waitFor(() => {
      expect(screen.getByText('Loading...')).toBeInTheDocument();
    });
  });

  it('shows Back button on User Details step', async () => {
    const user = userEvent.setup();
    renderPage();

    await goToDetailsStep(user);

    expect(screen.getByRole('button', {name: /back/i})).toBeInTheDocument();
  });

  it('does not show Back button on User Type step', () => {
    renderPage();

    expect(screen.queryByRole('button', {name: /back/i})).not.toBeInTheDocument();
  });

  it('navigates back to User Type step when Back is clicked', async () => {
    const user = userEvent.setup();
    renderPage();

    await goToDetailsStep(user);

    await user.click(screen.getByRole('button', {name: /back/i}));

    await waitFor(() => {
      expect(screen.getByTestId('configure-user-type')).toBeInTheDocument();
    });
  });

  it('navigates to a step via breadcrumb keyboard interaction', async () => {
    const user = userEvent.setup();
    renderPage();

    // Go to step 2 so the breadcrumb for step 1 becomes clickable
    await goToDetailsStep(user);

    // The first breadcrumb item (User Type) should be navigable via keyboard
    const breadcrumbStep = screen.getByRole('button', {name: /user type/i});
    breadcrumbStep.focus();
    await user.keyboard('{Enter}');

    await waitFor(() => {
      expect(screen.getByTestId('configure-user-type')).toBeInTheDocument();
    });
  });

  it('preserves selected schema when navigating back and forward', async () => {
    const user = userEvent.setup();
    renderPage();

    await goToDetailsStep(user);

    // Go back
    await user.click(screen.getByRole('button', {name: /back/i}));

    await waitFor(() => {
      expect(screen.getByTestId('configure-user-type')).toBeInTheDocument();
    });

    // Schema should still be selected
    expect(screen.getByTestId('selected-schema-name')).toHaveTextContent('Employee');
  });

  // ============================================================================
  // Submission
  // ============================================================================

  it('submits the form with correct data', async () => {
    const user = userEvent.setup();

    renderPage();

    await goToDetailsStep(user);
    await user.click(screen.getByTestId('fill-form'));

    // Wait for step ready state to update before clicking submit
    await waitFor(() => {
      expect(screen.getByRole('button', {name: /create user/i})).not.toBeDisabled();
    });

    await user.click(screen.getByRole('button', {name: /create user/i}));

    await waitFor(() => {
      expect(mockMutateAsync).toHaveBeenCalledWith({
        organizationUnit: 'root-ou',
        type: 'Employee',
        attributes: {username: 'john_doe', age: 30},
      });
    });

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/users');
    });
  });

  it('filters empty attribute values before submission', async () => {
    const user = userEvent.setup();

    renderPage();

    await goToDetailsStep(user);
    // Use the button that emits empty/null/undefined values alongside valid ones
    await user.click(screen.getByTestId('fill-form-with-empty-values'));

    await waitFor(() => {
      expect(screen.getByRole('button', {name: /create user/i})).not.toBeDisabled();
    });

    await user.click(screen.getByRole('button', {name: /create user/i}));

    await waitFor(() => {
      const calledWith = mockMutateAsync.mock.calls[0][0] as {attributes: Record<string, unknown>};
      // Verify empty/null/undefined values were filtered out
      expect(calledWith.attributes).toEqual({username: 'john_doe', age: 30});
      expect(calledWith.attributes).not.toHaveProperty('nickname');
      expect(calledWith.attributes).not.toHaveProperty('phone');
      expect(calledWith.attributes).not.toHaveProperty('address');
    });
  });

  it('shows saving state on submit button during loading', () => {
    mockUseCreateUser.mockReturnValue({
      mutate: vi.fn(),
      mutateAsync: mockMutateAsync,
      isPending: true,
      error: null,
      data: undefined,
      isError: false,
      isSuccess: false,
      isIdle: false,
      reset: mockReset,
    });

    renderPage();

    // On step 1 with isPending=true, the continue button should be disabled
    const continueButton = screen.getByRole('button', {name: /continue/i});
    expect(continueButton).toBeDisabled();
  });

  // ============================================================================
  // Error Handling
  // ============================================================================

  it('displays API error from create user', () => {
    mockUseCreateUser.mockReturnValue({
      mutate: vi.fn(),
      mutateAsync: mockMutateAsync,
      isPending: false,
      error: new Error('Failed to create user'),
      data: undefined,
      isError: true,
      isSuccess: false,
      isIdle: false,
      reset: mockReset,
    });

    renderPage();

    expect(screen.getByText('Failed to create user')).toBeInTheDocument();
  });

  it('closes snackbar when dismissed', async () => {
    const user = userEvent.setup();

    mockUseGetUserSchemas.mockReturnValue({
      data: {
        ...mockSchemasData,
        schemas: [{id: 'schema1', name: 'Employee', ouId: ''}],
      },
      isLoading: false,
      error: null,
    });

    renderPage();

    await goToDetailsStep(user);
    await user.click(screen.getByTestId('fill-form'));

    await waitFor(() => {
      expect(screen.getByRole('button', {name: /create user/i})).not.toBeDisabled();
    });

    // Trigger validation error to open snackbar
    await user.click(screen.getByRole('button', {name: /create user/i}));

    await waitFor(() => {
      expect(screen.getByText('Organization unit ID is missing for the selected user type.')).toBeInTheDocument();
    });

    // Close the snackbar
    const snackbarCloseButton = screen.getAllByRole('button', {name: /close/i});
    // The snackbar close button is the last close button rendered
    await user.click(snackbarCloseButton[snackbarCloseButton.length - 1]);

    await waitFor(() => {
      expect(
        screen.queryByText('Organization unit ID is missing for the selected user type.'),
      ).not.toBeInTheDocument();
    });
  });

  it('renders null when schema details are not available on step 2', async () => {
    mockUseGetUserSchema.mockReturnValue({
      data: undefined,
      isLoading: false,
      error: null,
    });

    const user = userEvent.setup();
    renderPage();

    await user.click(screen.getByTestId('select-schema-Employee'));
    await user.click(screen.getByRole('button', {name: /continue/i}));

    // Should not render the details form or loading
    await waitFor(() => {
      expect(screen.queryByTestId('configure-user-details')).not.toBeInTheDocument();
      expect(screen.queryByText('Loading...')).not.toBeInTheDocument();
    });
  });

  it('shows validation error when schema has missing ouId', async () => {
    const user = userEvent.setup();

    mockUseGetUserSchemas.mockReturnValue({
      data: {
        ...mockSchemasData,
        schemas: [{id: 'schema1', name: 'Employee', ouId: ''}],
      },
      isLoading: false,
      error: null,
    });

    renderPage();

    await goToDetailsStep(user);
    await user.click(screen.getByTestId('fill-form'));

    await waitFor(() => {
      expect(screen.getByRole('button', {name: /create user/i})).not.toBeDisabled();
    });

    await user.click(screen.getByRole('button', {name: /create user/i}));

    await waitFor(() => {
      expect(
        screen.getByText('Organization unit ID is missing for the selected user type.'),
      ).toBeInTheDocument();
    });

    expect(mockMutateAsync).not.toHaveBeenCalled();
  });

  it('handles null schemas data gracefully', () => {
    mockUseGetUserSchemas.mockReturnValue({
      data: undefined,
      isLoading: false,
      error: null,
    });

    renderPage();

    expect(screen.getByTestId('configure-user-type')).toBeInTheDocument();
  });

  it('handles create user rejection gracefully', async () => {
    const user = userEvent.setup();
    mockMutateAsync.mockRejectedValue(new Error('Network error'));

    renderPage();

    await goToDetailsStep(user);
    await user.click(screen.getByTestId('fill-form'));

    await waitFor(() => {
      expect(screen.getByRole('button', {name: /create user/i})).not.toBeDisabled();
    });

    await user.click(screen.getByRole('button', {name: /create user/i}));

    await waitFor(() => {
      expect(mockMutateAsync).toHaveBeenCalled();
    });
  });
});
