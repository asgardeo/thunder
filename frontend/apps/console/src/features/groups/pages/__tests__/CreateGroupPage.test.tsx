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
import {screen, waitFor} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {renderWithProviders} from '@thunder/test-utils';
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
    mutate: vi.fn(),
    isPending: false,
    error: null,
  }),
}));

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

const mockUseGetOrganizationUnits = vi.fn();
vi.mock('../../../organization-units/api/useGetOrganizationUnits', () => ({
  default: (...args: unknown[]): unknown => mockUseGetOrganizationUnits(...args),
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
    mockMutateAsync.mockResolvedValue({});
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe('with single OU', () => {
    beforeEach(() => {
      mockUseGetOrganizationUnits.mockReturnValue({
        data: {
          totalResults: 1,
          organizationUnits: [{id: 'ou-single', name: 'Default OU'}],
        },
        isLoading: false,
      });
    });

    it('should render name step with suggestions', () => {
      renderPage();

      expect(screen.getByTestId('configure-name')).toBeInTheDocument();
      expect(screen.getByText("Let's give a name to your group")).toBeInTheDocument();
    });

    it('should have disabled button initially', () => {
      renderPage();

      const button = screen.getByRole('button', {name: 'Continue'});
      expect(button).toBeDisabled();
    });

    it('should enable button when name is entered', async () => {
      const user = userEvent.setup();
      renderPage();

      const nameInput = screen.getByPlaceholderText('Enter group name');
      await user.type(nameInput, 'Test Group');

      await waitFor(() => {
        const button = screen.getByRole('button', {name: 'Continue'});
        expect(button).not.toBeDisabled();
      });
    });

    it('should submit directly without OU step when only one OU exists', async () => {
      const user = userEvent.setup();
      renderPage();

      const nameInput = screen.getByPlaceholderText('Enter group name');
      await user.type(nameInput, 'Test Group');

      await waitFor(() => {
        const button = screen.getByRole('button', {name: 'Continue'});
        expect(button).not.toBeDisabled();
      });

      await user.click(screen.getByRole('button', {name: 'Continue'}));

      await waitFor(() => {
        expect(mockMutateAsync).toHaveBeenCalledWith({
          name: 'Test Group',
          organizationUnitId: 'ou-single',
        });
      });
    });

    it('should navigate to groups list on successful creation', async () => {
      const user = userEvent.setup();
      renderPage();

      const nameInput = screen.getByPlaceholderText('Enter group name');
      await user.type(nameInput, 'Test Group');

      await waitFor(() => {
        const button = screen.getByRole('button', {name: 'Continue'});
        expect(button).not.toBeDisabled();
      });

      await user.click(screen.getByRole('button', {name: 'Continue'}));

      await waitFor(() => {
        expect(mockNavigate).toHaveBeenCalledWith('/groups');
      });
    });
  });

  describe('with multiple OUs', () => {
    beforeEach(() => {
      mockUseGetOrganizationUnits.mockReturnValue({
        data: {
          totalResults: 3,
          organizationUnits: [
            {id: 'ou-1', name: 'OU 1'},
            {id: 'ou-2', name: 'OU 2'},
          ],
        },
        isLoading: false,
      });
    });

    it('should show continue button on name step', async () => {
      const user = userEvent.setup();
      renderPage();

      const nameInput = screen.getByPlaceholderText('Enter group name');
      await user.type(nameInput, 'Test Group');

      await waitFor(() => {
        expect(screen.getByRole('button', {name: 'Continue'})).not.toBeDisabled();
      });
    });

    it('should navigate to OU step after name step', async () => {
      const user = userEvent.setup();
      renderPage();

      const nameInput = screen.getByPlaceholderText('Enter group name');
      await user.type(nameInput, 'Test Group');

      await waitFor(() => {
        expect(screen.getByRole('button', {name: 'Continue'})).not.toBeDisabled();
      });

      await user.click(screen.getByRole('button', {name: 'Continue'}));

      await waitFor(() => {
        expect(screen.getByTestId('configure-organization-unit')).toBeInTheDocument();
      });
    });

    it('should submit after selecting OU in step 2', async () => {
      const user = userEvent.setup();
      renderPage();

      // Step 1: Enter name
      const nameInput = screen.getByPlaceholderText('Enter group name');
      await user.type(nameInput, 'Test Group');

      await waitFor(() => {
        expect(screen.getByRole('button', {name: 'Continue'})).not.toBeDisabled();
      });

      await user.click(screen.getByRole('button', {name: 'Continue'}));

      // Step 2: Select OU
      await waitFor(() => {
        expect(screen.getByTestId('configure-organization-unit')).toBeInTheDocument();
      });

      await user.click(screen.getByTestId('select-ou'));

      await waitFor(() => {
        const button = screen.getByRole('button', {name: 'Continue'});
        expect(button).not.toBeDisabled();
      });

      await user.click(screen.getByRole('button', {name: 'Continue'}));

      await waitFor(() => {
        expect(mockMutateAsync).toHaveBeenCalledWith({
          name: 'Test Group',
          organizationUnitId: 'ou-123',
        });
      });
    });

    it('should navigate back to name step via breadcrumb click', async () => {
      const user = userEvent.setup();
      renderPage();

      const nameInput = screen.getByPlaceholderText('Enter group name');
      await user.type(nameInput, 'Test Group');

      await waitFor(() => {
        expect(screen.getByRole('button', {name: 'Continue'})).not.toBeDisabled();
      });

      await user.click(screen.getByRole('button', {name: 'Continue'}));

      await waitFor(() => {
        expect(screen.getByTestId('configure-organization-unit')).toBeInTheDocument();
      });

      // On OU step, "Create a Group" breadcrumb is a Typography with role="button"
      const nameStepBreadcrumb = screen.getByRole('button', {name: 'Create a Group'});
      await user.click(nameStepBreadcrumb);

      await waitFor(() => {
        expect(screen.getByTestId('configure-name')).toBeInTheDocument();
      });
    });

    it('should navigate back to name step via breadcrumb keyboard', async () => {
      const user = userEvent.setup();
      renderPage();

      const nameInput = screen.getByPlaceholderText('Enter group name');
      await user.type(nameInput, 'Test Group');

      await waitFor(() => {
        expect(screen.getByRole('button', {name: 'Continue'})).not.toBeDisabled();
      });

      await user.click(screen.getByRole('button', {name: 'Continue'}));

      await waitFor(() => {
        expect(screen.getByTestId('configure-organization-unit')).toBeInTheDocument();
      });

      // Navigate via keyboard Enter on the breadcrumb
      const nameStepBreadcrumb = screen.getByRole('button', {name: 'Create a Group'});
      nameStepBreadcrumb.focus();
      await user.keyboard('{Enter}');

      await waitFor(() => {
        expect(screen.getByTestId('configure-name')).toBeInTheDocument();
      });
    });

    it('should go back from OU step to name step', async () => {
      const user = userEvent.setup();
      renderPage();

      const nameInput = screen.getByPlaceholderText('Enter group name');
      await user.type(nameInput, 'Test Group');

      await waitFor(() => {
        expect(screen.getByRole('button', {name: 'Continue'})).not.toBeDisabled();
      });

      await user.click(screen.getByRole('button', {name: 'Continue'}));

      await waitFor(() => {
        expect(screen.getByTestId('configure-organization-unit')).toBeInTheDocument();
      });

      await user.click(screen.getByRole('button', {name: 'Back'}));

      await waitFor(() => {
        expect(screen.getByTestId('configure-name')).toBeInTheDocument();
      });
    });
  });

  it('should handle submission error gracefully', async () => {
    mockUseGetOrganizationUnits.mockReturnValue({
      data: {totalResults: 1, organizationUnits: [{id: 'ou-single', name: 'Default OU'}]},
      isLoading: false,
    });
    mockMutateAsync.mockRejectedValue(new Error('Create failed'));

    const user = userEvent.setup();
    renderPage();

    const nameInput = screen.getByPlaceholderText('Enter group name');
    await user.type(nameInput, 'Test Group');

    await waitFor(() => {
      expect(screen.getByRole('button', {name: 'Continue'})).not.toBeDisabled();
    });

    await user.click(screen.getByRole('button', {name: 'Continue'}));

    await waitFor(() => {
      expect(mockMutateAsync).toHaveBeenCalled();
    });

    // Should not navigate since submission failed
    expect(mockNavigate).not.toHaveBeenCalledWith('/groups');
  });

  it('should disable continue button while OUs are loading', () => {
    mockUseGetOrganizationUnits.mockReturnValue({
      data: null,
      isLoading: true,
    });

    renderPage();

    const button = screen.getByRole('button', {name: 'Continue'});
    expect(button).toBeDisabled();
  });

  it('should navigate back when close button is clicked', async () => {
    mockUseGetOrganizationUnits.mockReturnValue({
      data: {totalResults: 1, organizationUnits: [{id: 'ou-1', name: 'OU 1'}]},
      isLoading: false,
    });

    const user = userEvent.setup();
    renderPage();

    const closeButton = screen.getByRole('button', {name: 'Close'});
    await user.click(closeButton);

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/groups');
    });
  });

  it('should handle navigate rejection gracefully', async () => {
    mockUseGetOrganizationUnits.mockReturnValue({
      data: {totalResults: 1, organizationUnits: [{id: 'ou-1', name: 'OU 1'}]},
      isLoading: false,
    });
    mockNavigate.mockRejectedValue(new Error('Nav failed'));

    const user = userEvent.setup();
    renderPage();

    const closeButton = screen.getByRole('button', {name: 'Close'});
    await user.click(closeButton);

    // Should not throw - error is caught gracefully
    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/groups');
    });
  });
});
