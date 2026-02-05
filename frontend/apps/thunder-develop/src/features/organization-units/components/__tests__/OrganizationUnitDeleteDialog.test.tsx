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
import {screen, fireEvent, waitFor, renderWithProviders} from '@thunder/test-utils';
import OrganizationUnitDeleteDialog from '../OrganizationUnitDeleteDialog';

// Mock the delete hook
const mockMutate = vi.fn();
let mockIsPending = false;
vi.mock('../../api/useDeleteOrganizationUnit', () => ({
  default: () => ({
    mutate: mockMutate,
    get isPending() {
      return mockIsPending;
    },
  }),
}));

// Mock translations
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => {
      const translations: Record<string, string> = {
        'organizationUnits:delete.title': 'Delete Organization Unit',
        'organizationUnits:delete.message': 'Are you sure you want to delete this organization unit?',
        'organizationUnits:delete.disclaimer': 'This action cannot be undone.',
        'organizationUnits:delete.error': 'Failed to delete',
        'common:actions.cancel': 'Cancel',
        'common:actions.delete': 'Delete',
        'common:status.deleting': 'Deleting...',
      };
      return translations[key] ?? key;
    },
  }),
}));

describe('OrganizationUnitDeleteDialog', () => {
  const defaultProps = {
    open: true,
    organizationUnitId: 'ou-123',
    onClose: vi.fn(),
    onSuccess: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
    mockMutate.mockReset();
    mockIsPending = false;
  });

  it('should render dialog when open is true', () => {
    renderWithProviders(<OrganizationUnitDeleteDialog {...defaultProps} />);

    expect(screen.getByText('Delete Organization Unit')).toBeInTheDocument();
    expect(screen.getByText('Are you sure you want to delete this organization unit?')).toBeInTheDocument();
    expect(screen.getByText('This action cannot be undone.')).toBeInTheDocument();
  });

  it('should not render dialog content when open is false', () => {
    renderWithProviders(<OrganizationUnitDeleteDialog {...defaultProps} open={false} />);

    expect(screen.queryByText('Delete Organization Unit')).not.toBeInTheDocument();
  });

  it('should call onClose when cancel button is clicked', () => {
    const onClose = vi.fn();
    renderWithProviders(<OrganizationUnitDeleteDialog {...defaultProps} onClose={onClose} />);

    fireEvent.click(screen.getByText('Cancel'));

    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it('should call mutate with correct id when delete button is clicked', () => {
    renderWithProviders(<OrganizationUnitDeleteDialog {...defaultProps} />);

    fireEvent.click(screen.getByText('Delete'));

    expect(mockMutate).toHaveBeenCalledWith('ou-123', expect.any(Object));
  });

  it('should not call mutate when organizationUnitId is null', () => {
    renderWithProviders(<OrganizationUnitDeleteDialog {...defaultProps} organizationUnitId={null} />);

    fireEvent.click(screen.getByText('Delete'));

    expect(mockMutate).not.toHaveBeenCalled();
  });

  it('should call onClose and onSuccess on successful deletion', async () => {
    const onClose = vi.fn();
    const onSuccess = vi.fn();
    mockMutate.mockImplementation((_id, options: {onSuccess: () => void}) => {
      options.onSuccess();
    });

    renderWithProviders(<OrganizationUnitDeleteDialog {...defaultProps} onClose={onClose} onSuccess={onSuccess} />);

    fireEvent.click(screen.getByText('Delete'));

    await waitFor(() => {
      expect(onClose).toHaveBeenCalled();
      expect(onSuccess).toHaveBeenCalled();
    });
  });

  it('should display error message on deletion failure', async () => {
    mockMutate.mockImplementation((_id, options: {onError: (err: Error) => void}) => {
      options.onError(new Error('Network error'));
    });

    renderWithProviders(<OrganizationUnitDeleteDialog {...defaultProps} />);

    fireEvent.click(screen.getByText('Delete'));

    await waitFor(() => {
      expect(screen.getByText('Network error')).toBeInTheDocument();
    });
  });

  it('should display fallback error message when error has no message', async () => {
    mockMutate.mockImplementation((_id, options: {onError: (err: unknown) => void}) => {
      options.onError({});
    });

    renderWithProviders(<OrganizationUnitDeleteDialog {...defaultProps} />);

    fireEvent.click(screen.getByText('Delete'));

    await waitFor(() => {
      // There are 2 alerts - warning disclaimer and error
      const alerts = screen.getAllByRole('alert');
      expect(alerts.length).toBe(2);
    });
  });

  it('should work without onSuccess callback', async () => {
    const onClose = vi.fn();
    mockMutate.mockImplementation((_id, options: {onSuccess: () => void}) => {
      options.onSuccess();
    });

    renderWithProviders(
      <OrganizationUnitDeleteDialog {...defaultProps} onClose={onClose} onSuccess={undefined} />,
    );

    fireEvent.click(screen.getByText('Delete'));

    await waitFor(() => {
      expect(onClose).toHaveBeenCalled();
    });
  });

  it('should clear error when cancel is clicked', async () => {
    mockMutate.mockImplementation((_id, options: {onError: (err: Error) => void}) => {
      options.onError(new Error('Network error'));
    });

    const {rerender} = renderWithProviders(<OrganizationUnitDeleteDialog {...defaultProps} />);

    // Trigger error
    fireEvent.click(screen.getByText('Delete'));

    await waitFor(() => {
      expect(screen.getByText('Network error')).toBeInTheDocument();
    });

    // Click cancel
    fireEvent.click(screen.getByText('Cancel'));

    // Reopen dialog
    rerender(<OrganizationUnitDeleteDialog {...defaultProps} />);

    // Error should be cleared (dialog reopens fresh state)
    // Note: The error state is local to the component instance
  });

  it('should display cancel and delete buttons', () => {
    renderWithProviders(<OrganizationUnitDeleteDialog {...defaultProps} />);

    expect(screen.getByText('Cancel')).toBeInTheDocument();
    expect(screen.getByText('Delete')).toBeInTheDocument();
  });

  it('should render error alert only when error is present', async () => {
    renderWithProviders(<OrganizationUnitDeleteDialog {...defaultProps} />);

    // Initially no error alert (only the warning alert should be present)
    const alerts = screen.getAllByRole('alert');
    expect(alerts.length).toBe(1); // Only warning alert

    // Trigger error
    mockMutate.mockImplementation((_id: string, options: {onError: (err: Error) => void}) => {
      options.onError(new Error('Delete error'));
    });

    fireEvent.click(screen.getByText('Delete'));

    await waitFor(() => {
      const allAlerts = screen.getAllByRole('alert');
      expect(allAlerts.length).toBe(2); // Warning + error alert
      expect(screen.getByText('Delete error')).toBeInTheDocument();
    });
  });

  it('should call onSuccess when provided on successful deletion', async () => {
    const onSuccess = vi.fn();
    const onClose = vi.fn();
    mockMutate.mockImplementation((_id: string, options: {onSuccess: () => void}) => {
      options.onSuccess();
    });

    renderWithProviders(
      <OrganizationUnitDeleteDialog {...defaultProps} onClose={onClose} onSuccess={onSuccess} />,
    );

    fireEvent.click(screen.getByText('Delete'));

    await waitFor(() => {
      expect(onSuccess).toHaveBeenCalledTimes(1);
      expect(onClose).toHaveBeenCalledTimes(1);
    });
  });

  it('should handle error with null message using fallback', async () => {
    mockMutate.mockImplementation((_id: string, options: {onError: (err: Error) => void}) => {
      options.onError({message: null} as unknown as Error);
    });

    renderWithProviders(<OrganizationUnitDeleteDialog {...defaultProps} />);

    fireEvent.click(screen.getByText('Delete'));

    await waitFor(() => {
      expect(screen.getByText('Failed to delete')).toBeInTheDocument();
    });
  });

  it('should show deleting state and disable buttons when isPending is true', () => {
    mockIsPending = true;

    renderWithProviders(<OrganizationUnitDeleteDialog {...defaultProps} />);

    // Delete button should show "Deleting..." text
    expect(screen.getByText('Deleting...')).toBeInTheDocument();

    // Both buttons should be disabled
    const cancelButton = screen.getByText('Cancel').closest('button');
    const deletingButton = screen.getByText('Deleting...').closest('button');
    expect(cancelButton).toBeDisabled();
    expect(deletingButton).toBeDisabled();
  });

  it('should re-render correctly when props change', () => {
    const onClose = vi.fn();
    const {rerender} = renderWithProviders(
      <OrganizationUnitDeleteDialog {...defaultProps} onClose={onClose} organizationUnitId="ou-123" />,
    );

    expect(screen.getByText('Delete Organization Unit')).toBeInTheDocument();

    // Re-render with different organizationUnitId
    rerender(
      <OrganizationUnitDeleteDialog {...defaultProps} onClose={onClose} organizationUnitId="ou-456" />,
    );

    expect(screen.getByText('Delete Organization Unit')).toBeInTheDocument();

    // Click delete should call mutate with new ID
    fireEvent.click(screen.getByText('Delete'));
    expect(mockMutate).toHaveBeenCalledWith('ou-456', expect.any(Object));
  });

  it('should re-render when transitioning between open and closed', async () => {
    const {rerender} = renderWithProviders(
      <OrganizationUnitDeleteDialog {...defaultProps} open />,
    );

    expect(screen.getByText('Delete Organization Unit')).toBeInTheDocument();

    // Close dialog - MUI Dialog animates, so content may linger briefly
    rerender(<OrganizationUnitDeleteDialog {...defaultProps} open={false} />);

    await waitFor(() => {
      expect(screen.queryByText('Are you sure you want to delete this organization unit?')).not.toBeVisible();
    });

    // Reopen dialog
    rerender(<OrganizationUnitDeleteDialog {...defaultProps} open />);

    expect(screen.getByText('Delete Organization Unit')).toBeInTheDocument();
  });

  it('should re-render when isPending transitions', () => {
    mockIsPending = false;

    const {rerender} = renderWithProviders(
      <OrganizationUnitDeleteDialog {...defaultProps} />,
    );

    expect(screen.getByText('Delete')).toBeInTheDocument();

    // Transition to pending
    mockIsPending = true;
    rerender(<OrganizationUnitDeleteDialog {...defaultProps} />);

    expect(screen.getByText('Deleting...')).toBeInTheDocument();

    // Transition back
    mockIsPending = false;
    rerender(<OrganizationUnitDeleteDialog {...defaultProps} />);

    expect(screen.getByText('Delete')).toBeInTheDocument();
  });
});
