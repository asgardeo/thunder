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

import {describe, it, expect, beforeEach, afterEach, vi} from 'vitest';
import {page, userEvent} from 'vitest/browser';
import {renderWithProviders} from '@thunder/test-utils/browser';
import GroupCreateProvider from '../../contexts/GroupCreate/GroupCreateProvider';
import CreateGroupPage from '../CreateGroupPage';

const mockNavigate = vi.fn();
vi.mock('react-router', async () => {
  const actual = await vi.importActual<typeof import('react-router')>('react-router');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

const mockMutateAsync = vi.fn();
vi.mock('../../api/useCreateGroup', () => ({
  default: () => ({
    mutateAsync: mockMutateAsync,
    isPending: false,
    error: null,
  }),
}));

const mockUseGetOrganizationUnits = vi.fn();
vi.mock('../../../organization-units/api/useGetOrganizationUnits', () => ({
  default: (...args: unknown[]): unknown => mockUseGetOrganizationUnits(...args),
}));

const mockGenerateRandomHumanReadableIdentifiers = vi.hoisted(() => vi.fn());
vi.mock('@thunder/utils', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@thunder/utils')>();
  return {
    ...actual,
    generateRandomHumanReadableIdentifiers: mockGenerateRandomHumanReadableIdentifiers,
  };
});

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

function renderPage() {
  return renderWithProviders(
    <GroupCreateProvider>
      <CreateGroupPage />
    </GroupCreateProvider>,
  );
}

describe('CreateGroupPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockNavigate.mockResolvedValue(undefined);
    mockGenerateRandomHumanReadableIdentifiers.mockReturnValue([]);
    // Single OU — no OU step needed
    mockUseGetOrganizationUnits.mockReturnValue({
      data: {
        totalResults: 1,
        startIndex: 0,
        count: 1,
        organizationUnits: [{id: 'ou-1', name: 'Root OU'}],
      },
      isLoading: false,
    });
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('should render the wizard step breadcrumb', async () => {
    await renderPage();

    // The NAME step label is "Create a Group"
    await expect.element(page.getByText('Create a Group')).toBeInTheDocument();
  });

  it('should render ConfigureName on the first step', async () => {
    await renderPage();

    await expect.element(page.getByTestId('configure-name')).toBeInTheDocument();
  });

  it('should render Continue button', async () => {
    await renderPage();

    await expect.element(page.getByRole('button', {name: 'Continue'})).toBeInTheDocument();
  });

  it('should render Close button', async () => {
    await renderPage();

    await expect.element(page.getByRole('button', {name: 'Close'})).toBeInTheDocument();
  });

  it('should have Continue button disabled when name is empty', async () => {
    await renderPage();

    await expect.element(page.getByRole('button', {name: 'Continue'})).toBeDisabled();
  });

  it('should enable Continue button when name is entered', async () => {
    await renderPage();

    await userEvent.type(page.getByRole('textbox'), 'My Group');

    await expect.element(page.getByRole('button', {name: 'Continue'})).not.toBeDisabled();
  });

  it('should call mutateAsync and navigate on submit with single OU', async () => {
    mockMutateAsync.mockResolvedValue(undefined);
    await renderPage();

    await userEvent.type(page.getByRole('textbox'), 'My Group');
    await userEvent.click(page.getByRole('button', {name: 'Continue'}));

    await vi.waitFor(() => {
      expect(mockMutateAsync).toHaveBeenCalledWith({
        name: 'My Group',
        organizationUnitId: 'ou-1',
      });
    });

    await vi.waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/groups');
    });
  });

  it('should advance to OU step when multiple OUs exist', async () => {
    mockUseGetOrganizationUnits.mockReturnValue({
      data: {
        totalResults: 2,
        startIndex: 0,
        count: 2,
        organizationUnits: [
          {id: 'ou-1', name: 'Root OU'},
          {id: 'ou-2', name: 'Child OU'},
        ],
      },
      isLoading: false,
    });

    await renderPage();

    await userEvent.type(page.getByRole('textbox'), 'My Group');
    await userEvent.click(page.getByRole('button', {name: 'Continue'}));

    // Should now be on the OU step
    await expect.element(page.getByTestId('configure-organization-unit')).toBeInTheDocument();
  });

  it('should navigate to /groups when Close button is clicked', async () => {
    await renderPage();

    await userEvent.click(page.getByRole('button', {name: 'Close'}));

    await vi.waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/groups');
    });
  });

  it('should handle navigate rejection in close handler gracefully', async () => {
    mockNavigate.mockRejectedValue(new Error('Nav failed'));
    await renderPage();

    await userEvent.click(page.getByRole('button', {name: 'Close'}));

    await vi.waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/groups');
    });
  });

  it('should handle navigate rejection on successful creation gracefully', async () => {
    mockNavigate.mockRejectedValue(new Error('Nav failed'));
    mockMutateAsync.mockResolvedValue(undefined);

    await renderPage();

    await userEvent.type(page.getByRole('textbox'), 'My Group');
    await userEvent.click(page.getByRole('button', {name: 'Continue'}));

    await vi.waitFor(() => {
      expect(mockMutateAsync).toHaveBeenCalled();
    });
  });

  it('should show Back button on OU step', async () => {
    mockUseGetOrganizationUnits.mockReturnValue({
      data: {
        totalResults: 2,
        startIndex: 0,
        count: 2,
        organizationUnits: [
          {id: 'ou-1', name: 'Root OU'},
          {id: 'ou-2', name: 'Child OU'},
        ],
      },
      isLoading: false,
    });

    await renderPage();

    await userEvent.type(page.getByRole('textbox'), 'My Group');
    await userEvent.click(page.getByRole('button', {name: 'Continue'}));

    await expect.element(page.getByRole('button', {name: 'Back'})).toBeInTheDocument();
  });

  it('should go back to name step when Back is clicked on OU step', async () => {
    mockUseGetOrganizationUnits.mockReturnValue({
      data: {
        totalResults: 2,
        startIndex: 0,
        count: 2,
        organizationUnits: [
          {id: 'ou-1', name: 'Root OU'},
          {id: 'ou-2', name: 'Child OU'},
        ],
      },
      isLoading: false,
    });

    await renderPage();

    await userEvent.type(page.getByRole('textbox'), 'My Group');
    await userEvent.click(page.getByRole('button', {name: 'Continue'}));

    await expect.element(page.getByTestId('configure-organization-unit')).toBeInTheDocument();

    await userEvent.click(page.getByRole('button', {name: 'Back'}));

    await expect.element(page.getByTestId('configure-name')).toBeInTheDocument();
  });

  it('should submit with selected OU when multiple OUs exist', async () => {
    mockMutateAsync.mockResolvedValue(undefined);
    mockUseGetOrganizationUnits.mockReturnValue({
      data: {
        totalResults: 2,
        startIndex: 0,
        count: 2,
        organizationUnits: [
          {id: 'ou-1', name: 'Root OU'},
          {id: 'ou-2', name: 'Child OU'},
        ],
      },
      isLoading: false,
    });

    await renderPage();

    await userEvent.type(page.getByRole('textbox'), 'My Group');
    await userEvent.click(page.getByRole('button', {name: 'Continue'}));

    // Select OU
    await userEvent.click(page.getByTestId('select-ou'));
    await userEvent.click(page.getByRole('button', {name: 'Continue'}));

    await vi.waitFor(() => {
      expect(mockMutateAsync).toHaveBeenCalledWith({
        name: 'My Group',
        organizationUnitId: 'ou-123',
      });
    });
  });
});
