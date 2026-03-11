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

import {describe, it, expect, vi, beforeEach} from 'vitest';
import {renderWithProviders} from '@thunder/test-utils/browser';
import {page, userEvent} from 'vitest/browser';
import CreateOrganizationUnitPage from '../CreateOrganizationUnitPage';

// Mock navigate and location
const mockNavigate = vi.fn();
let mockLocationState: Record<string, unknown> | null = null;
vi.mock('react-router', async () => {
  const actual = await vi.importActual('react-router');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
    useLocation: () => ({
      pathname: '/organization-units/create',
      search: '',
      hash: '',
      state: mockLocationState,
      key: 'default',
    }),
  };
});

// Mock logger
vi.mock('@thunder/logger/react', () => ({
  useLogger: () => ({
    error: vi.fn(),
    info: vi.fn(),
    debug: vi.fn(),
  }),
}));

// Mock create hook
const mockMutate = vi.fn();
vi.mock('../../api/useCreateOrganizationUnit', () => ({
  default: () => ({
    mutate: mockMutate,
    isPending: false,
  }),
}));

// Mock useOrganizationUnit hook
vi.mock('../../contexts/useOrganizationUnit', () => ({
  default: () => ({
    resetTreeState: vi.fn(),
  }),
}));

// Mock name suggestions utility
vi.mock('@thunder/utils', () => ({
  generateRandomHumanReadableIdentifiers: () => ['Suggested Name One', 'Suggested Name Two', 'Suggested Name Three'],
}));

describe('CreateOrganizationUnitPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockNavigate.mockReset();
    mockMutate.mockReset();
    mockLocationState = null;
  });

  it('should render page title and heading', async () => {
    await renderWithProviders(<CreateOrganizationUnitPage />);

    await expect.element(page.getByText('Create Organization Unit')).toBeInTheDocument();
    await expect.element(page.getByText("Let's set up your organization unit")).toBeInTheDocument();
  });

  it('should render name input field', async () => {
    await renderWithProviders(<CreateOrganizationUnitPage />);

    await expect.element(page.getByLabelText(/Name/i)).toBeInTheDocument();
  });

  it('should render handle input field', async () => {
    await renderWithProviders(<CreateOrganizationUnitPage />);

    await expect.element(page.getByLabelText(/Handle/i)).toBeInTheDocument();
  });

  it('should render description input field', async () => {
    await renderWithProviders(<CreateOrganizationUnitPage />);

    await expect.element(page.getByLabelText(/Description/i)).toBeInTheDocument();
  });

  it('should render name suggestions', async () => {
    await renderWithProviders(<CreateOrganizationUnitPage />);

    await expect.element(page.getByRole('button', {name: 'Suggested Name One'})).toBeInTheDocument();
    await expect.element(page.getByRole('button', {name: 'Suggested Name Two'})).toBeInTheDocument();
    await expect.element(page.getByRole('button', {name: 'Suggested Name Three'})).toBeInTheDocument();
  });

  it('should auto-generate handle from name', async () => {
    await renderWithProviders(<CreateOrganizationUnitPage />);

    const nameInput = page.getByLabelText(/Name/i);
    await userEvent.fill(nameInput, 'Test Organization');

    const handleInput = page.getByLabelText(/Handle/i);
    expect(handleInput).toHaveValue('test-organization');
  });

  it('should fill name when suggestion is clicked', async () => {
    await renderWithProviders(<CreateOrganizationUnitPage />);

    await userEvent.click(page.getByRole('button', {name: 'Suggested Name One'}));

    const nameInput = page.getByLabelText(/Name/i);
    expect(nameInput).toHaveValue('Suggested Name One');
  });

  it('should auto-generate handle when suggestion is clicked', async () => {
    await renderWithProviders(<CreateOrganizationUnitPage />);

    await userEvent.click(page.getByRole('button', {name: 'Suggested Name One'}));

    const handleInput = page.getByLabelText(/Handle/i);
    expect(handleInput).toHaveValue('suggested-name-one');
  });

  it('should not auto-generate handle after manual edit', async () => {
    await renderWithProviders(<CreateOrganizationUnitPage />);

    const handleInput = page.getByLabelText(/Handle/i);
    await userEvent.fill(handleInput, 'my-custom-handle');

    const nameInput = page.getByLabelText(/Name/i);
    await userEvent.fill(nameInput, 'Test Organization');

    expect(handleInput).toHaveValue('my-custom-handle');
  });

  it('should disable create button when form is invalid', async () => {
    await renderWithProviders(<CreateOrganizationUnitPage />);

    const createButton = page.getByRole('button', {name: 'Create'});
    await expect.element(createButton).toBeDisabled();
  });

  it('should enable create button when form is valid', async () => {
    await renderWithProviders(<CreateOrganizationUnitPage />);

    const nameInput = page.getByLabelText(/Name/i);
    const handleInput = page.getByLabelText(/Handle/i);

    await userEvent.fill(nameInput, 'Test Organization');
    await userEvent.fill(handleInput, 'test-org');

    // Wait for form validation to complete
    await vi.waitFor(() => {
      const createButton = page.getByRole('button', {name: 'Create'});
      expect(createButton).not.toBeDisabled();
    });
  });

  it('should call mutate on form submit', async () => {
    await renderWithProviders(<CreateOrganizationUnitPage />);

    const nameInput = page.getByLabelText(/Name/i);
    await userEvent.fill(nameInput, 'Test Organization');

    // Wait for form validation to complete
    await vi.waitFor(() => {
      const createButton = page.getByRole('button', {name: 'Create'});
      expect(createButton).not.toBeDisabled();
    });

    const createButton = page.getByRole('button', {name: 'Create'});
    await userEvent.click(createButton);

    await vi.waitFor(() => {
      expect(mockMutate).toHaveBeenCalledWith(
        expect.objectContaining({
          name: 'Test Organization',
          handle: 'test-organization',
        }),
        expect.any(Object),
      );
    });
  });

  it('should navigate back when close button is clicked', async () => {
    await renderWithProviders(<CreateOrganizationUnitPage />);

    // Find the close button (X icon button)
    const closeButton = page.locator('button').first();
    await userEvent.click(closeButton);

    await vi.waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/organization-units');
    });
  });

  it('should navigate on successful creation', async () => {
    mockMutate.mockImplementation((_data, options: {onSuccess: () => void}) => {
      options.onSuccess();
    });

    await renderWithProviders(<CreateOrganizationUnitPage />);

    const nameInput = page.getByLabelText(/Name/i);
    await userEvent.fill(nameInput, 'Test Organization');

    // Wait for form validation to complete
    await vi.waitFor(() => {
      const createButton = page.getByRole('button', {name: 'Create'});
      expect(createButton).not.toBeDisabled();
    });

    const createButton = page.getByRole('button', {name: 'Create'});
    await userEvent.click(createButton);

    await vi.waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/organization-units');
    });
  });

  it('should display error on creation failure', async () => {
    mockMutate.mockImplementation((_data, options: {onError: (err: Error) => void}) => {
      options.onError(new Error('Network error'));
    });

    await renderWithProviders(<CreateOrganizationUnitPage />);

    const nameInput = page.getByLabelText(/Name/i);
    await userEvent.fill(nameInput, 'Test Organization');

    // Wait for form validation to complete
    await vi.waitFor(() => {
      const createButton = page.getByRole('button', {name: 'Create'});
      expect(createButton).not.toBeDisabled();
    });

    const createButton = page.getByRole('button', {name: 'Create'});
    await userEvent.click(createButton);

    await vi.waitFor(async () => {
      await expect.element(page.getByText('Network error')).toBeInTheDocument();
    });
  });

  it('should close error alert when close button is clicked', async () => {
    mockMutate.mockImplementation((_data, options: {onError: (err: Error) => void}) => {
      options.onError(new Error('Network error'));
    });

    await renderWithProviders(<CreateOrganizationUnitPage />);

    const nameInput = page.getByLabelText(/Name/i);
    await userEvent.fill(nameInput, 'Test Organization');

    // Wait for form validation to complete
    await vi.waitFor(() => {
      const createButton = page.getByRole('button', {name: 'Create'});
      expect(createButton).not.toBeDisabled();
    });

    const createButton = page.getByRole('button', {name: 'Create'});
    await userEvent.click(createButton);

    await vi.waitFor(async () => {
      await expect.element(page.getByText('Network error')).toBeInTheDocument();
    });

    // Close the alert
    const alertCloseButton = page.getByRole('button', {name: /close/i});
    await userEvent.click(alertCloseButton);

    await vi.waitFor(async () => {
      await expect.element(page.getByText('Network error')).not.toBeInTheDocument();
    });
  });

  it('should include description in request when provided', async () => {
    await renderWithProviders(<CreateOrganizationUnitPage />);

    const nameInput = page.getByLabelText(/Name/i);
    const descriptionInput = page.getByLabelText(/Description/i);

    await userEvent.fill(nameInput, 'Test Organization');
    await userEvent.fill(descriptionInput, 'A test description');

    // Wait for form validation to complete
    await vi.waitFor(() => {
      const createButton = page.getByRole('button', {name: 'Create'});
      expect(createButton).not.toBeDisabled();
    });

    const createButton = page.getByRole('button', {name: 'Create'});
    await userEvent.click(createButton);

    await vi.waitFor(() => {
      expect(mockMutate).toHaveBeenCalledWith(
        expect.objectContaining({
          description: 'A test description',
        }),
        expect.any(Object),
      );
    });
  });

  it('should set description to null when empty', async () => {
    await renderWithProviders(<CreateOrganizationUnitPage />);

    const nameInput = page.getByLabelText(/Name/i);
    await userEvent.fill(nameInput, 'Test Organization');

    // Wait for form validation to complete
    await vi.waitFor(() => {
      const createButton = page.getByRole('button', {name: 'Create'});
      expect(createButton).not.toBeDisabled();
    });

    const createButton = page.getByRole('button', {name: 'Create'});
    await userEvent.click(createButton);

    await vi.waitFor(() => {
      expect(mockMutate).toHaveBeenCalledWith(
        expect.objectContaining({
          description: null,
        }),
        expect.any(Object),
      );
    });
  });

  it('should show "Root Organization Unit" in parent field when no parent is provided', async () => {
    await renderWithProviders(<CreateOrganizationUnitPage />);

    const parentInput = page.getByLabelText(/Parent Organization Unit/i);
    expect(parentInput).toHaveValue('Root Organization Unit');
    expect(parentInput).toHaveAttribute('readOnly');
  });

  it('should set parent to null when no parent is in navigation state', async () => {
    await renderWithProviders(<CreateOrganizationUnitPage />);

    const nameInput = page.getByLabelText(/Name/i);
    await userEvent.fill(nameInput, 'Test Organization');

    // Wait for form validation to complete
    await vi.waitFor(() => {
      const createButton = page.getByRole('button', {name: 'Create'});
      expect(createButton).not.toBeDisabled();
    });

    const createButton = page.getByRole('button', {name: 'Create'});
    await userEvent.click(createButton);

    await vi.waitFor(() => {
      expect(mockMutate).toHaveBeenCalledWith(
        expect.objectContaining({
          parent: null,
        }),
        expect.any(Object),
      );
    });
  });

  it('should display parent name and handle when navigated with parent state', async () => {
    mockLocationState = {parentId: 'ou-1', parentName: 'Engineering', parentHandle: 'engineering'};

    await renderWithProviders(<CreateOrganizationUnitPage />);

    const parentInput = page.getByLabelText(/Parent Organization Unit/i);
    expect(parentInput).toHaveValue('Engineering (engineering)');
    expect(parentInput).toHaveAttribute('readOnly');
  });

  it('should display parent name without handle when handle is not provided', async () => {
    mockLocationState = {parentId: 'ou-1', parentName: 'Engineering'};

    await renderWithProviders(<CreateOrganizationUnitPage />);

    const parentInput = page.getByLabelText(/Parent Organization Unit/i);
    expect(parentInput).toHaveValue('Engineering');
  });

  it('should submit with parent ID from navigation state', async () => {
    mockLocationState = {parentId: 'ou-1', parentName: 'Engineering', parentHandle: 'engineering'};

    await renderWithProviders(<CreateOrganizationUnitPage />);

    const nameInput = page.getByLabelText(/Name/i);
    await userEvent.fill(nameInput, 'Child Organization');

    await vi.waitFor(() => {
      const createButton = page.getByRole('button', {name: 'Create'});
      expect(createButton).not.toBeDisabled();
    });

    const createButton = page.getByRole('button', {name: 'Create'});
    await userEvent.click(createButton);

    await vi.waitFor(() => {
      expect(mockMutate).toHaveBeenCalledWith(
        expect.objectContaining({
          parent: 'ou-1',
        }),
        expect.any(Object),
      );
    });
  });

  it('should keep handle unchanged after manual edit when suggestion is clicked', async () => {
    await renderWithProviders(<CreateOrganizationUnitPage />);

    const handleInput = page.getByLabelText(/Handle/i);
    await userEvent.fill(handleInput, 'my-custom-handle');

    await userEvent.click(page.getByRole('button', {name: 'Suggested Name Two'}));

    // Handle should not change after suggestion click since it was manually edited
    expect(handleInput).toHaveValue('my-custom-handle');
  });

  it('should handle error without message', async () => {
    mockMutate.mockImplementation((_data, options: {onError: (err: unknown) => void}) => {
      options.onError({});
    });

    await renderWithProviders(<CreateOrganizationUnitPage />);

    const nameInput = page.getByLabelText(/Name/i);
    await userEvent.fill(nameInput, 'Test Organization');

    // Wait for form validation to complete
    await vi.waitFor(() => {
      const createButton = page.getByRole('button', {name: 'Create'});
      expect(createButton).not.toBeDisabled();
    });

    const createButton = page.getByRole('button', {name: 'Create'});
    await userEvent.click(createButton);

    await vi.waitFor(async () => {
      await expect.element(page.getByRole('alert')).toBeInTheDocument();
    });
  });

  it('should handle close navigation error gracefully', async () => {
    mockNavigate.mockRejectedValue(new Error('Navigation failed'));

    await renderWithProviders(<CreateOrganizationUnitPage />);

    const closeButton = page.locator('button').first();
    await userEvent.click(closeButton);

    // Should not throw - error is logged
    await vi.waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/organization-units');
    });
  });

  it('should handle success navigation error gracefully', async () => {
    mockNavigate.mockRejectedValue(new Error('Navigation failed'));
    mockMutate.mockImplementation((_data, options: {onSuccess: () => void}) => {
      options.onSuccess();
    });

    await renderWithProviders(<CreateOrganizationUnitPage />);

    const nameInput = page.getByLabelText(/Name/i);
    await userEvent.fill(nameInput, 'Test Organization');

    // Wait for form validation to complete
    await vi.waitFor(() => {
      const createButton = page.getByRole('button', {name: 'Create'});
      expect(createButton).not.toBeDisabled();
    });

    const createButton = page.getByRole('button', {name: 'Create'});
    await userEvent.click(createButton);

    // Should not throw - error is logged
    await vi.waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/organization-units');
    });
  });

  it('should trim whitespace from inputs on submit', async () => {
    await renderWithProviders(<CreateOrganizationUnitPage />);

    const nameInput = page.getByLabelText(/Name/i);
    const handleInput = page.getByLabelText(/Handle/i);
    const descriptionInput = page.getByLabelText(/Description/i);

    await userEvent.fill(nameInput, '  Test Organization  ');
    await userEvent.fill(handleInput, '  test-org  ');
    await userEvent.fill(descriptionInput, '  A description  ');

    // Wait for form validation to complete
    await vi.waitFor(() => {
      const createButton = page.getByRole('button', {name: 'Create'});
      expect(createButton).not.toBeDisabled();
    });

    const createButton = page.getByRole('button', {name: 'Create'});
    await userEvent.click(createButton);

    await vi.waitFor(() => {
      expect(mockMutate).toHaveBeenCalledWith(
        expect.objectContaining({
          name: 'Test Organization',
          handle: 'test-org',
          description: 'A description',
        }),
        expect.any(Object),
      );
    });
  });

  it('should render progress bar', async () => {
    await renderWithProviders(<CreateOrganizationUnitPage />);

    await expect.element(page.getByRole('progressbar')).toBeInTheDocument();
  });

  it('should render suggestions label', async () => {
    await renderWithProviders(<CreateOrganizationUnitPage />);

    await expect.element(page.getByText('In a hurry? Pick a random name:')).toBeInTheDocument();
  });
});
