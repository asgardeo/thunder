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
import CreateGroupPage from '../CreateGroupPage';

const mockNavigate = vi.fn();
vi.mock('react-router', async () => {
  const actual = await vi.importActual<typeof import('react-router')>('react-router');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

const mockMutate = vi.fn();
vi.mock('../../api/useCreateGroup', () => ({
  default: () => ({
    mutate: mockMutate,
    isPending: false,
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

describe('CreateGroupPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockNavigate.mockResolvedValue(undefined);
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('should render page title and heading', () => {
    renderWithProviders(<CreateGroupPage />);

    expect(screen.getByText('create.title')).toBeInTheDocument();
    expect(screen.getByText('create.heading')).toBeInTheDocument();
  });

  it('should render form fields', () => {
    renderWithProviders(<CreateGroupPage />);

    expect(screen.getByText('create.form.name.label')).toBeInTheDocument();
    expect(screen.getByText('create.form.description.label')).toBeInTheDocument();
    expect(screen.getByText('create.form.organizationUnit.label')).toBeInTheDocument();
  });

  it('should render OU tree picker', () => {
    renderWithProviders(<CreateGroupPage />);

    expect(screen.getByTestId('ou-tree-picker')).toBeInTheDocument();
  });

  it('should have disabled submit button initially', () => {
    renderWithProviders(<CreateGroupPage />);

    const submitButton = screen.getByText('Create').closest('button');
    expect(submitButton).toBeDisabled();
  });

  it('should enable submit button when form is valid', async () => {
    const user = userEvent.setup();
    renderWithProviders(<CreateGroupPage />);

    const nameInput = screen.getByPlaceholderText('create.form.name.placeholder');
    await user.type(nameInput, 'Test Group');

    await user.click(screen.getByTestId('select-ou'));

    await waitFor(() => {
      const submitButton = screen.getByText('Create').closest('button');
      expect(submitButton).not.toBeDisabled();
    });
  });

  it('should call mutate on form submit', async () => {
    const user = userEvent.setup();
    renderWithProviders(<CreateGroupPage />);

    const nameInput = screen.getByPlaceholderText('create.form.name.placeholder');
    await user.type(nameInput, 'Test Group');

    await user.click(screen.getByTestId('select-ou'));

    await waitFor(() => {
      const submitButton = screen.getByText('Create').closest('button');
      expect(submitButton).not.toBeDisabled();
    });

    await user.click(screen.getByText('Create'));

    await waitFor(() => {
      expect(mockMutate).toHaveBeenCalledWith(
        {
          name: 'Test Group',
          description: undefined,
          organizationUnitId: 'ou-123',
        },
        expect.objectContaining({
          onSuccess: expect.any(Function) as unknown,
          onError: expect.any(Function) as unknown,
        }),
      );
    });
  });

  it('should navigate to groups list on successful creation', async () => {
    mockMutate.mockImplementation((_data: unknown, opts: {onSuccess: () => void}) => {
      opts.onSuccess();
    });

    const user = userEvent.setup();
    renderWithProviders(<CreateGroupPage />);

    const nameInput = screen.getByPlaceholderText('create.form.name.placeholder');
    await user.type(nameInput, 'Test Group');

    await user.click(screen.getByTestId('select-ou'));

    await waitFor(() => {
      const submitButton = screen.getByText('Create').closest('button');
      expect(submitButton).not.toBeDisabled();
    });

    await user.click(screen.getByText('Create'));

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/groups');
    });
  });

  it('should display error on failed creation', async () => {
    mockMutate.mockImplementation((_data: unknown, opts: {onError: (err: Error) => void}) => {
      opts.onError(new Error('Creation failed'));
    });

    const user = userEvent.setup();
    renderWithProviders(<CreateGroupPage />);

    const nameInput = screen.getByPlaceholderText('create.form.name.placeholder');
    await user.type(nameInput, 'Test Group');

    await user.click(screen.getByTestId('select-ou'));

    await waitFor(() => {
      const submitButton = screen.getByText('Create').closest('button');
      expect(submitButton).not.toBeDisabled();
    });

    await user.click(screen.getByText('Create'));

    await waitFor(() => {
      expect(screen.getByText('Creation failed')).toBeInTheDocument();
    });
  });

  it('should submit with description when provided', async () => {
    const user = userEvent.setup();
    renderWithProviders(<CreateGroupPage />);

    const nameInput = screen.getByPlaceholderText('create.form.name.placeholder');
    await user.type(nameInput, 'Test Group');

    const descInput = screen.getByPlaceholderText('create.form.description.placeholder');
    await user.type(descInput, 'A test description');

    await user.click(screen.getByTestId('select-ou'));

    await waitFor(() => {
      const submitButton = screen.getByText('Create').closest('button');
      expect(submitButton).not.toBeDisabled();
    });

    await user.click(screen.getByText('Create'));

    await waitFor(() => {
      expect(mockMutate).toHaveBeenCalledWith(
        {
          name: 'Test Group',
          description: 'A test description',
          organizationUnitId: 'ou-123',
        },
        expect.objectContaining({
          onSuccess: expect.any(Function) as unknown,
          onError: expect.any(Function) as unknown,
        }),
      );
    });
  });

  it('should navigate back when close button is clicked', async () => {
    const user = userEvent.setup();
    renderWithProviders(<CreateGroupPage />);

    const closeButton = screen.getByRole('button', {name: 'Close'});
    await user.click(closeButton);
    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/groups');
    });
  });

  it('should handle navigate rejection in close handler gracefully', async () => {
    mockNavigate.mockRejectedValue(new Error('Nav failed'));
    const user = userEvent.setup();
    renderWithProviders(<CreateGroupPage />);

    const closeButton = screen.getByRole('button', {name: 'Close'});
    await user.click(closeButton);

    // Should not throw - error is caught gracefully
    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/groups');
    });
  });

  it('should handle navigate rejection in onSuccess gracefully', async () => {
    mockNavigate.mockRejectedValue(new Error('Nav failed'));
    mockMutate.mockImplementation((_data: unknown, opts: {onSuccess: () => void}) => {
      opts.onSuccess();
    });

    const user = userEvent.setup();
    renderWithProviders(<CreateGroupPage />);

    const nameInput = screen.getByPlaceholderText('create.form.name.placeholder');
    await user.type(nameInput, 'Test Group');
    await user.click(screen.getByTestId('select-ou'));

    await waitFor(() => {
      const submitButton = screen.getByText('Create').closest('button');
      expect(submitButton).not.toBeDisabled();
    });

    await user.click(screen.getByText('Create'));

    // Should not throw - error is caught gracefully
    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/groups');
    });
  });

  it('should close error alert when dismiss is clicked', async () => {
    mockMutate.mockImplementation((_data: unknown, opts: {onError: (err: Error) => void}) => {
      opts.onError(new Error('Creation failed'));
    });

    const user = userEvent.setup();
    renderWithProviders(<CreateGroupPage />);

    const nameInput = screen.getByPlaceholderText('create.form.name.placeholder');
    await user.type(nameInput, 'Test Group');
    await user.click(screen.getByTestId('select-ou'));

    await waitFor(() => {
      const submitButton = screen.getByText('Create').closest('button');
      expect(submitButton).not.toBeDisabled();
    });

    await user.click(screen.getByText('Create'));

    await waitFor(() => {
      expect(screen.getByText('Creation failed')).toBeInTheDocument();
    });

    // Close the error alert (find the close button inside the alert)
    const alert = screen.getByRole('alert');
    const closeAlertButton = alert.querySelector('button')!;
    await user.click(closeAlertButton);

    await waitFor(() => {
      expect(screen.queryByText('Creation failed')).not.toBeInTheDocument();
    });
  });
});
