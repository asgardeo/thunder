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

import type {ReactNode} from 'react';
import {describe, it, expect, vi, beforeEach} from 'vitest';
import {render, screen, waitFor, userEvent, within} from '@thunder/test-utils';
import CreateUserTypePage from '../CreateUserTypePage';
import UserTypeCreateProvider from '../../contexts/UserTypeCreate/UserTypeCreateProvider';
import type {ApiError, CreateUserSchemaRequest} from '../../types/user-types';

const mockNavigate = vi.fn();
const mockCreateUserType = vi.fn();

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

// Mock useCreateUserType hook
interface UseCreateUserTypeReturn {
  createUserType: (data: CreateUserSchemaRequest) => Promise<void>;
  loading: boolean;
  error: ApiError | null;
}

const mockUseCreateUserType = vi.fn<() => UseCreateUserTypeReturn>();

vi.mock('../../api/useCreateUserType', () => ({
  default: () => mockUseCreateUserType(),
}));

// Mock useGetOrganizationUnits (used by ConfigureGeneral for auto-selecting the first OU)
vi.mock('../../../organization-units/api/useGetOrganizationUnits', () => ({
  default: () => ({
    data: {
      totalResults: 2,
      startIndex: 1,
      count: 2,
      organizationUnits: [
        {id: 'root-ou', name: 'Root Organization', handle: 'root', description: null, parent: null},
        {id: 'child-ou', name: 'Child Organization', handle: 'child', description: null, parent: 'root-ou'},
      ],
    },
    isLoading: false,
    error: null,
  }),
}));

// Mock OrganizationUnitTreePicker
vi.mock('../../../organization-units/components/OrganizationUnitTreePicker', () => ({
  default: ({value, onChange}: {value: string; onChange: (id: string) => void}) => (
    <div data-testid="ou-tree-picker">
      <span data-testid="ou-value">{value}</span>
      <button type="button" data-testid="select-ou" onClick={() => onChange('ou-123')}>
        Select OU
      </button>
    </div>
  ),
}));

vi.mock('../../../applications/utils/generateAppNameSuggestions', () => ({
  default: () => ['Alpha Users', 'Beta Users', 'Gamma Users'],
}));

/**
 * Helper to render the wizard page wrapped in provider.
 */
function renderPage() {
  return render(
    <UserTypeCreateProvider>
      <CreateUserTypePage />
    </UserTypeCreateProvider>,
  );
}

/**
 * Helper to navigate from step 1 (Name) to step 2 (General) by typing a name and clicking Continue.
 */
async function goToGeneralStep(user: ReturnType<typeof userEvent.setup>, name = 'Employee') {
  await user.type(screen.getByLabelText(/User Type Name/i), name);
  await user.click(screen.getByRole('button', {name: /Continue/i}));
  await waitFor(() => {
    expect(screen.getByTestId('configure-general')).toBeInTheDocument();
  });
}

/**
 * Helper to navigate from step 1 to step 3 (Properties) via step 2.
 * The first OU is auto-selected by ConfigureGeneral.
 */
async function goToPropertiesStep(user: ReturnType<typeof userEvent.setup>, name = 'Employee') {
  await goToGeneralStep(user, name);
  // OU is auto-selected from useGetOrganizationUnits data
  await waitFor(() => {
    const continueButton = screen.getByRole('button', {name: /Continue/i});
    expect(continueButton).not.toBeDisabled();
  });
  await user.click(screen.getByRole('button', {name: /Continue/i}));
  await waitFor(() => {
    expect(screen.getByTestId('configure-properties')).toBeInTheDocument();
  });
}

const getPropertyTypeSelect = (index = 0) => screen.getAllByRole('combobox')[index];

describe('CreateUserTypePage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseCreateUserType.mockReturnValue({
      createUserType: mockCreateUserType,
      loading: false,
      error: null,
    });
  });

  // ============================================================================
  // Step 1: Name
  // ============================================================================

  it('renders the wizard with Name step initially', () => {
    renderPage();

    expect(screen.getByTestId('configure-name')).toBeInTheDocument();
    expect(screen.getByText("Let's name your user type")).toBeInTheDocument();
    expect(screen.getByLabelText(/User Type Name/i)).toBeInTheDocument();
  });

  it('closes wizard when X button is clicked', async () => {
    const user = userEvent.setup();
    renderPage();

    // The X (close) button navigates back to /user-types
    const closeButtons = screen.getAllByRole('button');
    // The X close button is the first IconButton in the header
    const closeButton = closeButtons[0];
    await user.click(closeButton);

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/user-types');
    });
  });

  it('allows user to enter user type name', async () => {
    const user = userEvent.setup();
    renderPage();

    const nameInput = screen.getByLabelText(/User Type Name/i);
    await user.type(nameInput, 'Employee');

    expect(nameInput).toHaveValue('Employee');
  });

  it('disables Continue button when name is empty', () => {
    renderPage();

    const continueButton = screen.getByRole('button', {name: /Continue/i});
    expect(continueButton).toBeDisabled();
  });

  it('enables Continue button when name is entered', async () => {
    const user = userEvent.setup();
    renderPage();

    await user.type(screen.getByLabelText(/User Type Name/i), 'Employee');

    const continueButton = screen.getByRole('button', {name: /Continue/i});
    expect(continueButton).not.toBeDisabled();
  });

  it('navigates to General step when Continue is clicked', async () => {
    const user = userEvent.setup();
    renderPage();

    await goToGeneralStep(user);

    expect(screen.getByTestId('configure-general')).toBeInTheDocument();
  });

  // ============================================================================
  // Step 2: General
  // ============================================================================

  it('shows the organization unit tree picker on General step', async () => {
    const user = userEvent.setup();
    renderPage();

    await goToGeneralStep(user);

    expect(screen.getByTestId('ou-tree-picker')).toBeInTheDocument();
  });

  it('auto-selects the first organization unit', async () => {
    const user = userEvent.setup();
    renderPage();

    await goToGeneralStep(user);

    // First OU should be auto-selected from useGetOrganizationUnits data
    await waitFor(() => {
      expect(screen.getByTestId('ou-value')).toHaveTextContent('root-ou');
    });
  });

  it('allows selecting a different organization unit via tree picker', async () => {
    const user = userEvent.setup();
    renderPage();

    await goToGeneralStep(user);

    await user.click(screen.getByTestId('select-ou'));

    expect(screen.getByTestId('ou-value')).toHaveTextContent('ou-123');
  });

  it('enables Continue on General step when OU is auto-selected', async () => {
    const user = userEvent.setup();
    renderPage();

    await goToGeneralStep(user);

    await waitFor(() => {
      const continueButton = screen.getByRole('button', {name: /Continue/i});
      expect(continueButton).not.toBeDisabled();
    });
  });

  it('navigates back to Name step when Back is clicked on General step', async () => {
    const user = userEvent.setup();
    renderPage();

    await goToGeneralStep(user);

    await user.click(screen.getByRole('button', {name: /Back/i}));

    await waitFor(() => {
      expect(screen.getByTestId('configure-name')).toBeInTheDocument();
    });
  });

  // ============================================================================
  // Step 3: Properties
  // ============================================================================

  it('navigates to Properties step', async () => {
    const user = userEvent.setup();
    renderPage();

    await goToPropertiesStep(user);

    expect(screen.getByTestId('configure-properties')).toBeInTheDocument();
  });

  it('allows adding a new property', async () => {
    const user = userEvent.setup();
    renderPage();

    await goToPropertiesStep(user);

    const addButton = screen.getByRole('button', {name: /Add Property/i});
    await user.click(addButton);
  });

  it('allows removing a property', async () => {
    const user = userEvent.setup();
    renderPage();

    await goToPropertiesStep(user);

    // Add a second property first
    const addButton = screen.getByRole('button', {name: /Add Property/i});
    await user.click(addButton);

    // Now remove the second property
    const removeButtons = screen
      .getAllByRole('button')
      .filter((btn) => btn.classList.contains('MuiIconButton-colorError'));

    await user.click(removeButtons[removeButtons.length - 1]);
  });

  it('allows changing property name', async () => {
    const user = userEvent.setup();
    renderPage();

    await goToPropertiesStep(user);

    const propertyNameInput = screen.getByPlaceholderText(/e\.g\., email, age, address/i);
    await user.type(propertyNameInput, 'email');

    expect(propertyNameInput).toHaveValue('email');
  });

  it('allows changing property type', async () => {
    const user = userEvent.setup();
    renderPage();

    await goToPropertiesStep(user);

    const typeSelect = getPropertyTypeSelect();
    await user.click(typeSelect);

    const numberOption = await screen.findByText('Number');
    await user.click(numberOption);

    await waitFor(() => {
      expect(typeSelect).toHaveTextContent('Number');
    });
  });

  it('allows toggling required checkbox', async () => {
    const user = userEvent.setup();
    renderPage();

    await goToPropertiesStep(user);

    const requiredCheckbox = screen.getByRole('checkbox', {name: /Users must provide a value/i});
    expect(requiredCheckbox).not.toBeChecked();

    await user.click(requiredCheckbox);
    expect(requiredCheckbox).toBeChecked();

    await user.click(requiredCheckbox);
    expect(requiredCheckbox).not.toBeChecked();
  });

  it('allows toggling unique checkbox for string type', async () => {
    const user = userEvent.setup();
    renderPage();

    await goToPropertiesStep(user);

    const uniqueCheckbox = screen.getByRole('checkbox', {name: /Each user must have a distinct value/i});
    expect(uniqueCheckbox).not.toBeChecked();

    await user.click(uniqueCheckbox);
    expect(uniqueCheckbox).toBeChecked();
  });

  it('hides unique checkbox for boolean type', async () => {
    const user = userEvent.setup();
    renderPage();

    await goToPropertiesStep(user);

    expect(screen.getByRole('checkbox', {name: /Each user must have a distinct value/i})).toBeInTheDocument();

    const typeSelect = getPropertyTypeSelect();
    await user.click(typeSelect);

    const booleanOption = await screen.findByText('Boolean');
    await user.click(booleanOption);

    await waitFor(() => {
      expect(screen.queryByRole('checkbox', {name: /Each user must have a distinct value/i})).not.toBeInTheDocument();
    });
  });

  it('allows adding regex pattern for string type', async () => {
    const user = userEvent.setup();
    renderPage();

    await goToPropertiesStep(user);

    const regexInput = screen.getByPlaceholderText('e.g., ^[a-zA-Z0-9]+$');
    await user.click(regexInput);
    await user.paste('^[a-z]+$');

    expect(regexInput).toHaveValue('^[a-z]+$');
  });

  it(
    'allows adding enum values for enum type',
    async () => {
      const user = userEvent.setup();
      renderPage();

      await goToPropertiesStep(user);

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
    },
    15_000,
  );

  it('allows adding enum value by pressing Enter', async () => {
    const user = userEvent.setup();
    renderPage();

    await goToPropertiesStep(user);

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
    renderPage();

    await goToPropertiesStep(user);

    const typeSelect = getPropertyTypeSelect();
    await user.click(typeSelect);
    const enumOption = await screen.findByText('Enum');
    await user.click(enumOption);

    const enumInput = screen.getByPlaceholderText(/Add value and press Enter/i);

    const addEnumButton = screen.getByRole('button', {name: /^Add$/i});
    await user.click(addEnumButton);

    const enumContainer = enumInput.closest('div')?.querySelector('.MuiBox-root');
    expect(enumContainer).not.toBeInTheDocument();

    await user.type(enumInput, '   ');
    await user.click(addEnumButton);

    expect(enumContainer).not.toBeInTheDocument();
  });

  it('allows removing enum values', async () => {
    const user = userEvent.setup();
    renderPage();

    await goToPropertiesStep(user);

    const typeSelect = getPropertyTypeSelect();
    await user.click(typeSelect);
    const enumOption = await screen.findByText('Enum');
    await user.click(enumOption);

    const enumInput = screen.getByPlaceholderText(/Add value and press Enter/i);
    await user.type(enumInput, 'admin{Enter}');

    await waitFor(() => {
      expect(screen.getByText('admin')).toBeInTheDocument();
    });

    // The Chip component renders a delete icon as an SVG sibling to the label
    const chipElement = screen.getByText('admin').closest('.MuiChip-root');
    const deleteIcon = chipElement?.querySelector('.MuiChip-deleteIcon');
    if (deleteIcon) {
      await user.click(deleteIcon);
    }

    await waitFor(() => {
      expect(screen.queryByText('admin')).not.toBeInTheDocument();
    });
  });

  it('resets type-specific fields when type changes', async () => {
    const user = userEvent.setup();
    renderPage();

    await goToPropertiesStep(user);

    const typeSelect = getPropertyTypeSelect();
    await user.click(typeSelect);
    const enumTypeOption = await screen.findByText('Enum');
    await user.click(enumTypeOption);

    const enumInput = screen.getByPlaceholderText(/Add value and press Enter/i);
    await user.type(enumInput, 'test{Enter}');

    await waitFor(() => {
      expect(screen.getByText('test')).toBeInTheDocument();
    });

    await user.click(typeSelect);

    const numberOption = await screen.findByText('Number');
    await user.click(numberOption);

    await waitFor(() => {
      expect(screen.queryByText('test')).not.toBeInTheDocument();
    });

    expect(screen.queryByPlaceholderText(/Add value and press Enter/i)).not.toBeInTheDocument();
    expect(screen.queryByPlaceholderText(/\^\[a-zA-Z0-9\]\+\$/)).not.toBeInTheDocument();
  });

  it('navigates back to General step when Back is clicked on Properties step', async () => {
    const user = userEvent.setup();
    renderPage();

    await goToPropertiesStep(user);

    await user.click(screen.getByRole('button', {name: /Back/i}));

    await waitFor(() => {
      expect(screen.getByTestId('configure-general')).toBeInTheDocument();
    });
  });

  // ============================================================================
  // Full wizard flow: submission
  // ============================================================================

  it('successfully creates user type with valid data', async () => {
    const user = userEvent.setup();
    mockCreateUserType.mockResolvedValue(undefined);

    renderPage();

    // Step 1: Name
    await user.type(screen.getByLabelText(/User Type Name/i), 'Employee');
    await user.click(screen.getByRole('button', {name: /Continue/i}));

    // Step 2: General (OU auto-selected as root-ou)
    await waitFor(() => {
      expect(screen.getByTestId('configure-general')).toBeInTheDocument();
    });
    await waitFor(() => {
      expect(screen.getByRole('button', {name: /Continue/i})).not.toBeDisabled();
    });
    await user.click(screen.getByRole('button', {name: /Continue/i}));

    // Step 3: Properties
    await waitFor(() => {
      expect(screen.getByTestId('configure-properties')).toBeInTheDocument();
    });
    const propertyNameInput = screen.getByPlaceholderText(/e.g., email, age, address/i);
    await user.type(propertyNameInput, 'email');

    const requiredCheckbox = screen.getByRole('checkbox', {name: /Users must provide a value/i});
    await user.click(requiredCheckbox);

    // Submit
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

    renderPage();

    // Step 1: Name
    await user.type(screen.getByLabelText(/User Type Name/i), 'Employee');
    await user.click(screen.getByRole('button', {name: /Continue/i}));

    // Step 2: General - pick a different OU via tree picker and enable self-registration
    await waitFor(() => {
      expect(screen.getByTestId('configure-general')).toBeInTheDocument();
    });
    await user.click(screen.getByTestId('select-ou'));
    await user.click(screen.getByLabelText(/Allow Self Registration/i));
    await user.click(screen.getByRole('button', {name: /Continue/i}));

    // Step 3: Properties
    await waitFor(() => {
      expect(screen.getByTestId('configure-properties')).toBeInTheDocument();
    });
    const propertyNameInput = screen.getByPlaceholderText(/e\.g\., email, age, address/i);
    await user.type(propertyNameInput, 'email');

    await user.click(screen.getByRole('button', {name: /Create User Type/i}));

    await waitFor(() => {
      expect(mockCreateUserType).toHaveBeenCalledWith({
        name: 'Employee',
        ouId: 'ou-123',
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

  it('shows validation error when submitting without property name', async () => {
    const user = userEvent.setup();
    renderPage();

    await goToPropertiesStep(user);

    // The Create User Type button should be disabled when no property name is entered
    const submitButton = screen.getByRole('button', {name: /Create User Type/i});
    expect(submitButton).toBeDisabled();

    expect(mockCreateUserType).not.toHaveBeenCalled();
  });

  it('shows validation error for duplicate property names', async () => {
    const user = userEvent.setup();
    renderPage();

    await goToPropertiesStep(user);

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

  it('displays error from API', () => {
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

    renderPage();

    expect(screen.getByText('Failed to create user type')).toBeInTheDocument();
    expect(screen.getByText('User type already exists')).toBeInTheDocument();
  });

  it('shows loading state during submission on last step', async () => {
    const user = userEvent.setup();
    renderPage();

    // Navigate to properties step with loading=false (default from beforeEach)
    await goToPropertiesStep(user);

    // Type a property name so the step is "ready"
    await user.type(screen.getByPlaceholderText(/e\.g\., email, age, address/i), 'email');

    // Now switch to loading state
    mockUseCreateUserType.mockReturnValue({
      createUserType: mockCreateUserType,
      loading: true,
      error: null,
    });

    // Trigger a re-render by typing (the hook return value will update)
    await user.type(screen.getByPlaceholderText(/e\.g\., email, age, address/i), '2');

    await waitFor(() => {
      expect(screen.getByText('Saving...')).toBeInTheDocument();
    });
    expect(screen.getByRole('button', {name: /Saving/i})).toBeDisabled();
  });

  it('creates schema with enum property correctly', {timeout: 15_000}, async () => {
    const user = userEvent.setup();
    mockCreateUserType.mockResolvedValue(undefined);

    renderPage();

    await goToPropertiesStep(user, 'Complex Type');

    // Change type to enum
    const typeSelect = getPropertyTypeSelect();
    await user.click(typeSelect);
    const enumOption = await screen.findByText('Enum');
    await user.click(enumOption);

    // Add enum property with all features
    const firstPropertyInput = screen.getByPlaceholderText(/e.g., email, age, address/i);
    await user.type(firstPropertyInput, 'status');

    const requiredCheckbox = screen.getByRole('checkbox', {name: /Users must provide a value/i});
    await user.click(requiredCheckbox);

    const uniqueCheckbox = screen.getByRole('checkbox', {name: /Each user must have a distinct value/i});
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

  it('creates schema with string property containing regex pattern', async () => {
    const user = userEvent.setup();
    mockCreateUserType.mockResolvedValue(undefined);

    renderPage();

    await goToPropertiesStep(user, 'RegexTest');

    // Add property name
    const propertyNameInput = screen.getByPlaceholderText(/e.g., email, age, address/i);
    await user.type(propertyNameInput, 'code');

    // Add regex pattern
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

    renderPage();

    await goToPropertiesStep(user, 'NumberTest');

    // Add property name
    const propertyNameInput = screen.getByPlaceholderText(/e.g., email, age, address/i);
    await user.type(propertyNameInput, 'employeeId');

    // Change type to number
    const typeSelect = getPropertyTypeSelect();
    await user.click(typeSelect);
    const numberOption = await screen.findByText('Number');
    await user.click(numberOption);

    // Mark as unique
    const uniqueCheckbox = screen.getByRole('checkbox', {name: /Each user must have a distinct value/i});
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

    renderPage();

    await goToPropertiesStep(user);

    await user.type(screen.getByPlaceholderText(/e\.g\., email, age, address/i), 'email');

    const submitButton = screen.getByRole('button', {name: /Create User Type/i});
    await user.click(submitButton);

    await waitFor(() => {
      expect(mockCreateUserType).toHaveBeenCalled();
      expect(mockNavigate).not.toHaveBeenCalledWith('/user-types');
    });
  });

  // ============================================================================
  // Breadcrumb navigation
  // ============================================================================

  it('shows breadcrumbs reflecting current step', async () => {
    const user = userEvent.setup();
    renderPage();

    // Step 1: Only "Name" breadcrumb
    const breadcrumb1 = screen.getByRole('navigation', {name: /breadcrumb/i});
    expect(within(breadcrumb1).getByText('Name')).toBeInTheDocument();

    await goToGeneralStep(user);

    // Step 2: "Name" > "General" breadcrumbs
    const breadcrumb2 = screen.getByRole('navigation', {name: /breadcrumb/i});
    expect(within(breadcrumb2).getByText('Name')).toBeInTheDocument();
    expect(within(breadcrumb2).getByText('General')).toBeInTheDocument();
  });

  it('allows navigating back via breadcrumb click', async () => {
    const user = userEvent.setup();
    renderPage();

    await goToGeneralStep(user);

    // Click on "Name" breadcrumb to go back
    const breadcrumb = screen.getByRole('navigation', {name: /breadcrumb/i});
    await user.click(within(breadcrumb).getByText('Name'));

    await waitFor(() => {
      expect(screen.getByTestId('configure-name')).toBeInTheDocument();
    });
  });

  // ============================================================================
  // Progress bar
  // ============================================================================

  it('shows progress bar', () => {
    renderPage();

    expect(screen.getByRole('progressbar')).toBeInTheDocument();
  });
});
