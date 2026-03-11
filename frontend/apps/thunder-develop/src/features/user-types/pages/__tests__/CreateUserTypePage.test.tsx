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
import {render} from '@thunder/test-utils/browser';
import {page, userEvent} from 'vitest/browser';
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
async function renderPage() {
  return render(
    <UserTypeCreateProvider>
      <CreateUserTypePage />
    </UserTypeCreateProvider>,
  );
}

/**
 * Helper to navigate from step 1 (Name) to step 2 (General) by typing a name and clicking Continue.
 */
async function goToGeneralStep(name = 'Employee') {
  await userEvent.fill(page.getByLabelText(/User Type Name/i), name);
  await userEvent.click(page.getByRole('button', {name: /Continue/i}));
  await expect.element(page.getByTestId('configure-general')).toBeInTheDocument();
}

/**
 * Helper to navigate from step 1 to step 3 (Properties) via step 2.
 * The first OU is auto-selected by ConfigureGeneral.
 */
async function goToPropertiesStep(name = 'Employee') {
  await goToGeneralStep(name);
  // OU is auto-selected from useGetOrganizationUnits data
  await expect.element(page.getByRole('button', {name: /Continue/i})).not.toBeDisabled();
  await userEvent.click(page.getByRole('button', {name: /Continue/i}));
  await expect.element(page.getByTestId('configure-properties')).toBeInTheDocument();
}

const getPropertyTypeSelect = async (index = 0) => (page.getByRole('combobox').all())[index];

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

  it('renders the wizard with Name step initially', async () => {
    await renderPage();

    await expect.element(page.getByTestId('configure-name')).toBeInTheDocument();
    await expect.element(page.getByText("Let's name your user type")).toBeInTheDocument();
    await expect.element(page.getByLabelText(/User Type Name/i)).toBeInTheDocument();
  });

  it('closes wizard when X button is clicked', async () => {
    await renderPage();

    // The X (close) button navigates back to /user-types
    const closeButtons = page.getByRole('button').all();
    // The X close button is the first IconButton in the header
    const closeButton = closeButtons[0];
    await userEvent.click(closeButton);

    await vi.waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/user-types');
    });
  });

  it('allows user to enter user type name', async () => {
    await renderPage();

    const nameInput = page.getByLabelText(/User Type Name/i);
    await userEvent.fill(nameInput, 'Employee');

    await expect.element(nameInput).toHaveValue('Employee');
  });

  it('disables Continue button when name is empty', async () => {
    await renderPage();

    const continueButton = page.getByRole('button', {name: /Continue/i});
    await expect.element(continueButton).toBeDisabled();
  });

  it('enables Continue button when name is entered', async () => {
    await renderPage();

    await userEvent.fill(page.getByLabelText(/User Type Name/i), 'Employee');

    const continueButton = page.getByRole('button', {name: /Continue/i});
    await expect.element(continueButton).not.toBeDisabled();
  });

  it('navigates to General step when Continue is clicked', async () => {
    await renderPage();

    await goToGeneralStep();

    await expect.element(page.getByTestId('configure-general')).toBeInTheDocument();
  });

  // ============================================================================
  // Step 2: General
  // ============================================================================

  it('shows the organization unit tree picker on General step', async () => {
    await renderPage();

    await goToGeneralStep();

    await expect.element(page.getByTestId('ou-tree-picker')).toBeInTheDocument();
  });

  it('auto-selects the first organization unit', async () => {
    await renderPage();

    await goToGeneralStep();

    // First OU should be auto-selected from useGetOrganizationUnits data
    await expect.element(page.getByTestId('ou-value')).toHaveTextContent('root-ou');
  });

  it('allows selecting a different organization unit via tree picker', async () => {
    await renderPage();

    await goToGeneralStep();

    await userEvent.click(page.getByTestId('select-ou'));

    await expect.element(page.getByTestId('ou-value')).toHaveTextContent('ou-123');
  });

  it('enables Continue on General step when OU is auto-selected', async () => {
    await renderPage();

    await goToGeneralStep();

    await expect.element(page.getByRole('button', {name: /Continue/i})).not.toBeDisabled();
  });

  it('navigates back to Name step when Back is clicked on General step', async () => {
    await renderPage();

    await goToGeneralStep();

    await userEvent.click(page.getByRole('button', {name: /Back/i}));

    await expect.element(page.getByTestId('configure-name')).toBeInTheDocument();
  });

  // ============================================================================
  // Step 3: Properties
  // ============================================================================

  it('navigates to Properties step', async () => {
    await renderPage();

    await goToPropertiesStep();

    await expect.element(page.getByTestId('configure-properties')).toBeInTheDocument();
  });

  it('allows adding a new property', async () => {
    await renderPage();

    await goToPropertiesStep();

    const addButton = page.getByRole('button', {name: /Add Property/i});
    await userEvent.click(addButton);
  });

  it('allows removing a property', async () => {
    await renderPage();

    await goToPropertiesStep();

    // Add a second property first
    const addButton = page.getByRole('button', {name: /Add Property/i});
    await userEvent.click(addButton);

    // Now remove the second property
    const removeButtons = page.getByRole('button').all();
    const errorButtons = removeButtons.filter(
      (btn) => btn.element().classList.contains('MuiIconButton-colorError'),
    );

    await userEvent.click(errorButtons[errorButtons.length - 1]);
  });

  it('allows changing property name', async () => {
    await renderPage();

    await goToPropertiesStep();

    const propertyNameInput = page.getByPlaceholder(/e\.g\., email, age, address/i);
    await userEvent.fill(propertyNameInput, 'email');

    await expect.element(propertyNameInput).toHaveValue('email');
  });

  it('allows changing property type', async () => {
    await renderPage();

    await goToPropertiesStep();

    const typeSelect = await getPropertyTypeSelect();
    await userEvent.click(typeSelect);

    await vi.waitFor(async () => {
      await expect.element(page.getByRole('listbox')).toBeInTheDocument();
    });

    const numberOption = page.getByRole('option', {name: 'Number'});
    await userEvent.click(numberOption);

    await vi.waitFor(async () => {
      await expect.element(page.getByRole('listbox')).not.toBeInTheDocument();
    });

    // Re-acquire the combobox locator after selection to avoid stale locator issues
    const updatedTypeSelect = await getPropertyTypeSelect();
    await expect.element(updatedTypeSelect).toHaveTextContent('Number');
  });

  it('allows toggling required checkbox', async () => {
    await renderPage();

    await goToPropertiesStep();

    const requiredCheckbox = page.getByRole('checkbox', {name: /Users must provide a value/i});
    await expect.element(requiredCheckbox).not.toBeChecked();

    await userEvent.click(requiredCheckbox);
    await expect.element(requiredCheckbox).toBeChecked();

    await userEvent.click(requiredCheckbox);
    await expect.element(requiredCheckbox).not.toBeChecked();
  });

  it('allows toggling unique checkbox for string type', async () => {
    await renderPage();

    await goToPropertiesStep();

    const uniqueCheckbox = page.getByRole('checkbox', {name: /Each user must have a distinct value/i});
    await expect.element(uniqueCheckbox).not.toBeChecked();

    await userEvent.click(uniqueCheckbox);
    await expect.element(uniqueCheckbox).toBeChecked();
  });

  it('hides unique checkbox for boolean type', async () => {
    await renderPage();

    await goToPropertiesStep();

    await expect.element(page.getByRole('checkbox', {name: /Each user must have a distinct value/i})).toBeInTheDocument();

    const typeSelect = await getPropertyTypeSelect();
    await userEvent.click(typeSelect);

    const booleanOption = page.getByRole('option', {name: 'Boolean'});
    await userEvent.click(booleanOption);

    await expect.element(page.getByRole('checkbox', {name: /Each user must have a distinct value/i})).not.toBeInTheDocument();
  });

  it('allows adding regex pattern for string type', async () => {
    await renderPage();

    await goToPropertiesStep();

    const regexInput = page.getByPlaceholder('e.g., ^[a-zA-Z0-9]+$');
    await userEvent.click(regexInput);
    await userEvent.fill(regexInput, '^[a-z]+$');

    await expect.element(regexInput).toHaveValue('^[a-z]+$');
  });

  it(
    'allows adding enum values for enum type',
    async () => {
      await renderPage();

      await goToPropertiesStep();

      const typeSelect = await getPropertyTypeSelect();
      await userEvent.click(typeSelect);
      const enumOption = page.getByRole('option', {name: 'Enum'});
      await userEvent.click(enumOption);

      const enumInput = page.getByPlaceholder(/Add value and press Enter/i);
      await userEvent.fill(enumInput, 'admin');

      const addEnumButton = page.getByRole('button', {name: /^Add$/i});
      await userEvent.click(addEnumButton);

      await expect.element(page.getByText('admin')).toBeInTheDocument();

      await expect.element(enumInput).toHaveValue('');
    },
    15_000,
  );

  it('allows adding enum value by pressing Enter', async () => {
    await renderPage();

    await goToPropertiesStep();

    const typeSelect = await getPropertyTypeSelect();
    await userEvent.click(typeSelect);
    const enumOption = page.getByRole('option', {name: 'Enum'});
    await userEvent.click(enumOption);

    const enumInput = page.getByPlaceholder(/Add value and press Enter/i);
    await userEvent.fill(enumInput, 'user');
    await userEvent.keyboard('{Enter}');

    await expect.element(page.getByText('user', {exact: true})).toBeInTheDocument();

    await expect.element(enumInput).toHaveValue('');
  });

  it('does not add enum value when input is empty or whitespace', async () => {
    await renderPage();

    await goToPropertiesStep();

    const typeSelect = await getPropertyTypeSelect();
    await userEvent.click(typeSelect);
    const enumOption = page.getByRole('option', {name: 'Enum'});
    await userEvent.click(enumOption);

    const enumInput = page.getByPlaceholder(/Add value and press Enter/i);

    const addEnumButton = page.getByRole('button', {name: /^Add$/i});
    await userEvent.click(addEnumButton);

    const enumContainer = enumInput.element().closest('div')?.querySelector('.MuiBox-root');
    expect(enumContainer).not.toBeInTheDocument();

    await userEvent.fill(enumInput, '   ');
    await userEvent.click(addEnumButton);

    expect(enumContainer).not.toBeInTheDocument();
  });

  it('allows removing enum values', async () => {
    await renderPage();

    await goToPropertiesStep();

    const typeSelect = await getPropertyTypeSelect();
    await userEvent.click(typeSelect);
    const enumOption = page.getByRole('option', {name: 'Enum'});
    await userEvent.click(enumOption);

    const enumInput = page.getByPlaceholder(/Add value and press Enter/i);
    await userEvent.fill(enumInput, 'admin');
    await userEvent.keyboard('{Enter}');

    await expect.element(page.getByText('admin')).toBeInTheDocument();

    // The Chip component renders a delete icon as an SVG sibling to the label
    const chipElement = page.getByText('admin').element().closest('.MuiChip-root');
    const deleteIcon = chipElement?.querySelector('.MuiChip-deleteIcon');
    if (deleteIcon) {
      await userEvent.click(deleteIcon as HTMLElement);
    }

    await expect.element(page.getByText('admin')).not.toBeInTheDocument();
  });

  it('resets type-specific fields when type changes', async () => {
    await renderPage();

    await goToPropertiesStep();

    const typeSelect = await getPropertyTypeSelect();
    await userEvent.click(typeSelect);
    await vi.waitFor(async () => {
      await expect.element(page.getByRole('listbox')).toBeInTheDocument();
    });
    const enumTypeOption = page.getByRole('option', {name: 'Enum'});
    await userEvent.click(enumTypeOption);
    await vi.waitFor(async () => {
      await expect.element(page.getByRole('listbox')).not.toBeInTheDocument();
    });

    const enumInput = page.getByPlaceholder(/Add value and press Enter/i);
    await userEvent.fill(enumInput, 'test');
    await userEvent.keyboard('{Enter}');

    await expect.element(page.getByText('test')).toBeInTheDocument();

    // Re-acquire the typeSelect locator since the DOM may have re-rendered after enum value addition
    const typeSelectAfterEnum = await getPropertyTypeSelect();
    await userEvent.click(typeSelectAfterEnum);
    await vi.waitFor(async () => {
      await expect.element(page.getByRole('listbox')).toBeInTheDocument();
    });

    const numberOption = page.getByRole('option', {name: 'Number'});
    await userEvent.click(numberOption);
    await vi.waitFor(async () => {
      await expect.element(page.getByRole('listbox')).not.toBeInTheDocument();
    });

    await expect.element(page.getByText('test')).not.toBeInTheDocument();

    await expect.element(page.getByPlaceholder(/Add value and press Enter/i)).not.toBeInTheDocument();
    await expect.element(page.getByPlaceholder(/\^\[a-zA-Z0-9\]\+\$/)).not.toBeInTheDocument();
  });

  it('navigates back to General step when Back is clicked on Properties step', async () => {
    await renderPage();

    await goToPropertiesStep();

    await userEvent.click(page.getByRole('button', {name: /Back/i}));

    await expect.element(page.getByTestId('configure-general')).toBeInTheDocument();
  });

  // ============================================================================
  // Full wizard flow: submission
  // ============================================================================

  it('successfully creates user type with valid data', async () => {
    mockCreateUserType.mockResolvedValue(undefined);

    await renderPage();

    // Step 1: Name
    await userEvent.fill(page.getByLabelText(/User Type Name/i), 'Employee');
    await userEvent.click(page.getByRole('button', {name: /Continue/i}));

    // Step 2: General (OU auto-selected as root-ou)
    await expect.element(page.getByTestId('configure-general')).toBeInTheDocument();
    await expect.element(page.getByRole('button', {name: /Continue/i})).not.toBeDisabled();
    await userEvent.click(page.getByRole('button', {name: /Continue/i}));

    // Step 3: Properties
    await expect.element(page.getByTestId('configure-properties')).toBeInTheDocument();
    const propertyNameInput = page.getByPlaceholder(/e.g., email, age, address/i);
    await userEvent.fill(propertyNameInput, 'email');

    const requiredCheckbox = page.getByRole('checkbox', {name: /Users must provide a value/i});
    await userEvent.click(requiredCheckbox);

    // Submit
    const submitButton = page.getByRole('button', {name: /Create User Type/i});
    await userEvent.click(submitButton);

    await vi.waitFor(() => {
      expect(mockCreateUserType).toHaveBeenCalledWith({
        name: 'Employee',
        ouId: 'root-ou',
        schema: {
          email: {
            type: 'string',
            required: true,
          },
        },
        systemAttributes: {display: 'email'},
      });
    });

    await vi.waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/user-types');
    });
  });

  it('submits organization unit and registration flag when provided', async () => {
    mockCreateUserType.mockResolvedValue(undefined);

    await renderPage();

    // Step 1: Name
    await userEvent.fill(page.getByLabelText(/User Type Name/i), 'Employee');
    await userEvent.click(page.getByRole('button', {name: /Continue/i}));

    // Step 2: General - pick a different OU via tree picker and enable self-registration
    await expect.element(page.getByTestId('configure-general')).toBeInTheDocument();
    await userEvent.click(page.getByTestId('select-ou'));
    await userEvent.click(page.getByLabelText(/Allow Self Registration/i));
    await userEvent.click(page.getByRole('button', {name: /Continue/i}));

    // Step 3: Properties
    await expect.element(page.getByTestId('configure-properties')).toBeInTheDocument();
    const propertyNameInput = page.getByPlaceholder(/e\.g\., email, age, address/i);
    await userEvent.fill(propertyNameInput, 'email');

    await userEvent.click(page.getByRole('button', {name: /Create User Type/i}));

    await vi.waitFor(() => {
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
        systemAttributes: {display: 'email'},
      });
    });
  });

  it('shows validation error when submitting without property name', async () => {
    await renderPage();

    await goToPropertiesStep();

    // The Create User Type button should be disabled when no property name is entered
    const submitButton = page.getByRole('button', {name: /Create User Type/i});
    await expect.element(submitButton).toBeDisabled();

    expect(mockCreateUserType).not.toHaveBeenCalled();
  });

  it('shows validation error for duplicate property names', async () => {
    await renderPage();

    await goToPropertiesStep();

    // Add first property
    const firstPropertyInput = page.getByPlaceholder(/e\.g\., email, age, address/i);
    await userEvent.fill(firstPropertyInput, 'email');

    // Add second property
    const addButton = page.getByRole('button', {name: /Add Property/i});
    await userEvent.click(addButton);

    // Set same name for second property
    const propertyInputs = page.getByPlaceholder(/e\.g\., email, age, address/i).all();
    await userEvent.fill(propertyInputs[1], 'email');

    const submitButton = page.getByRole('button', {name: /Create User Type/i});
    await userEvent.click(submitButton);

    await expect.element(page.getByText(/Duplicate property names found/i)).toBeInTheDocument();

    expect(mockCreateUserType).not.toHaveBeenCalled();
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

    await renderPage();

    await expect.element(page.getByText('Failed to create user type')).toBeInTheDocument();
    await expect.element(page.getByText('User type already exists')).toBeInTheDocument();
  });

  it('shows loading state during submission on last step', async () => {
    await renderPage();

    // Navigate to properties step with loading=false (default from beforeEach)
    await goToPropertiesStep();

    // Type a property name so the step is "ready"
    await userEvent.fill(page.getByPlaceholder(/e\.g\., email, age, address/i), 'email');

    // Now switch to loading state
    mockUseCreateUserType.mockReturnValue({
      createUserType: mockCreateUserType,
      loading: true,
      error: null,
    });

    // Trigger a re-render by typing (the hook return value will update)
    await userEvent.fill(page.getByPlaceholder(/e\.g\., email, age, address/i), 'email2');

    await expect.element(page.getByText('Saving...')).toBeInTheDocument();
    await expect.element(page.getByRole('button', {name: /Saving/i})).toBeDisabled();
  });

  it('creates schema with enum property correctly', {timeout: 15_000}, async () => {
    mockCreateUserType.mockResolvedValue(undefined);

    await renderPage();

    await goToPropertiesStep('Complex Type');

    // Change type to enum
    const typeSelect = await getPropertyTypeSelect();
    await userEvent.click(typeSelect);
    const enumOption = page.getByRole('option', {name: 'Enum'});
    await userEvent.click(enumOption);

    // Add enum property with all features
    const firstPropertyInput = page.getByPlaceholder(/e.g., email, age, address/i);
    await userEvent.fill(firstPropertyInput, 'status');

    const requiredCheckbox = page.getByRole('checkbox', {name: /Users must provide a value/i});
    await userEvent.click(requiredCheckbox);

    const uniqueCheckbox = page.getByRole('checkbox', {name: /Each user must have a distinct value/i});
    await userEvent.click(uniqueCheckbox);

    const enumInput = page.getByPlaceholder(/Add value and press Enter/i);
    await userEvent.fill(enumInput, 'ACTIVE');
    await userEvent.keyboard('{Enter}');
    await userEvent.fill(enumInput, 'INACTIVE');
    await userEvent.keyboard('{Enter}');

    // Submit
    const submitButton = page.getByRole('button', {name: /Create User Type/i});
    await userEvent.click(submitButton);

    await vi.waitFor(() => {
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
        systemAttributes: {display: 'status'},
      });
    });
  });

  it('creates schema with string property containing regex pattern', async () => {
    mockCreateUserType.mockResolvedValue(undefined);

    await renderPage();

    await goToPropertiesStep('RegexTest');

    // Add property name
    const propertyNameInput = page.getByPlaceholder(/e.g., email, age, address/i);
    await userEvent.fill(propertyNameInput, 'code');

    // Add regex pattern
    const regexInput = page.getByPlaceholder('e.g., ^[a-zA-Z0-9]+$');
    await userEvent.click(regexInput);
    await userEvent.fill(regexInput, '^[A-Z]{3}$');

    // Submit
    const submitButton = page.getByRole('button', {name: /Create User Type/i});
    await userEvent.click(submitButton);

    await vi.waitFor(() => {
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
        systemAttributes: {display: 'code'},
      });
    });
  });

  it('creates schema with number property that is unique', async () => {
    mockCreateUserType.mockResolvedValue(undefined);

    await renderPage();

    await goToPropertiesStep('NumberTest');

    // Add property name
    const propertyNameInput = page.getByPlaceholder(/e.g., email, age, address/i);
    await userEvent.fill(propertyNameInput, 'employeeId');

    // Change type to number
    const typeSelect = await getPropertyTypeSelect();
    await userEvent.click(typeSelect);
    const numberOption = page.getByRole('option', {name: 'Number'});
    await userEvent.click(numberOption);

    // Mark as unique
    const uniqueCheckbox = page.getByRole('checkbox', {name: /Each user must have a distinct value/i});
    await userEvent.click(uniqueCheckbox);

    // Submit
    const submitButton = page.getByRole('button', {name: /Create User Type/i});
    await userEvent.click(submitButton);

    await vi.waitFor(() => {
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
        systemAttributes: {display: 'employeeId'},
      });
    });
  });

  it('handles create error gracefully', async () => {
    mockCreateUserType.mockRejectedValue(new Error('Create failed'));

    await renderPage();

    await goToPropertiesStep();

    await userEvent.fill(page.getByPlaceholder(/e\.g\., email, age, address/i), 'email');

    const submitButton = page.getByRole('button', {name: /Create User Type/i});
    await userEvent.click(submitButton);

    await vi.waitFor(() => {
      expect(mockCreateUserType).toHaveBeenCalled();
      expect(mockNavigate).not.toHaveBeenCalledWith('/user-types');
    });
  });

  // ============================================================================
  // Breadcrumb navigation
  // ============================================================================

  it('shows breadcrumbs reflecting current step', async () => {
    await renderPage();

    // Step 1: Only "Name" breadcrumb
    const breadcrumb1 = page.getByRole('navigation', {name: /breadcrumb/i});
    await expect.element(breadcrumb1.getByText('Name')).toBeInTheDocument();

    await goToGeneralStep();

    // Step 2: "Name" > "General" breadcrumbs
    const breadcrumb2 = page.getByRole('navigation', {name: /breadcrumb/i});
    await expect.element(breadcrumb2.getByText('Name')).toBeInTheDocument();
    await expect.element(breadcrumb2.getByText('General')).toBeInTheDocument();
  });

  it('allows navigating back via breadcrumb click', async () => {
    await renderPage();

    await goToGeneralStep();

    // Click on "Name" breadcrumb to go back
    const breadcrumb = page.getByRole('navigation', {name: /breadcrumb/i});
    await userEvent.click(breadcrumb.getByText('Name'));

    await expect.element(page.getByTestId('configure-name')).toBeInTheDocument();
  });

  // ============================================================================
  // Progress bar
  // ============================================================================

  it('shows progress bar', async () => {
    await renderPage();

    await expect.element(page.getByRole('progressbar')).toBeInTheDocument();
  });
});
