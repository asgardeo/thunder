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
import GroupDeleteDialog from '../GroupDeleteDialog';

const mockMutate = vi.fn();
vi.mock('../../api/useDeleteGroup', () => ({
  default: () => ({
    mutate: mockMutate,
    isPending: false,
  }),
}));

describe('GroupDeleteDialog', () => {
  const defaultProps = {
    open: true,
    groupId: 'g1',
    onClose: vi.fn(),
    onSuccess: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('should render dialog when open', () => {
    renderWithProviders(<GroupDeleteDialog {...defaultProps} />);

    expect(screen.getByText('delete.title')).toBeInTheDocument();
    expect(screen.getByText('delete.message')).toBeInTheDocument();
    expect(screen.getByText('delete.disclaimer')).toBeInTheDocument();
  });

  it('should not render content when closed', () => {
    renderWithProviders(<GroupDeleteDialog {...defaultProps} open={false} />);

    expect(screen.queryByText('delete.title')).not.toBeInTheDocument();
  });

  it('should call onClose when cancel is clicked', async () => {
    const user = userEvent.setup();
    renderWithProviders(<GroupDeleteDialog {...defaultProps} />);

    await user.click(screen.getByText('Cancel'));

    expect(defaultProps.onClose).toHaveBeenCalled();
  });

  it('should call mutate when confirm is clicked', async () => {
    const user = userEvent.setup();
    renderWithProviders(<GroupDeleteDialog {...defaultProps} />);

    await user.click(screen.getByText('Delete'));

    expect(mockMutate).toHaveBeenCalledWith('g1', expect.objectContaining({onSuccess: expect.any(Function) as unknown, onError: expect.any(Function) as unknown}));
  });

  it('should call onSuccess and onClose on successful delete', async () => {
    mockMutate.mockImplementation((_id: string, opts: {onSuccess: () => void}) => {
      opts.onSuccess();
    });

    const user = userEvent.setup();
    renderWithProviders(<GroupDeleteDialog {...defaultProps} />);

    await user.click(screen.getByText('Delete'));

    expect(defaultProps.onClose).toHaveBeenCalled();
    expect(defaultProps.onSuccess).toHaveBeenCalled();
  });

  it('should display error on failed delete', async () => {
    mockMutate.mockImplementation((_id: string, opts: {onError: (err: Error) => void}) => {
      opts.onError(new Error('Delete failed'));
    });

    const user = userEvent.setup();
    renderWithProviders(<GroupDeleteDialog {...defaultProps} />);

    await user.click(screen.getByText('Delete'));

    await waitFor(() => {
      expect(screen.getByText('Delete failed')).toBeInTheDocument();
    });
  });

  it('should not call mutate when groupId is null', () => {
    renderWithProviders(<GroupDeleteDialog {...defaultProps} groupId={null} />);

    // Button should be disabled when groupId is null, preventing mutation
    const deleteButton = screen.getByText('Delete').closest('button');
    expect(deleteButton).toBeDisabled();
    expect(mockMutate).not.toHaveBeenCalled();
  });

  it('should disable delete button when groupId is null', () => {
    renderWithProviders(<GroupDeleteDialog {...defaultProps} groupId={null} />);

    const deleteButton = screen.getByText('Delete').closest('button');
    expect(deleteButton).toBeDisabled();
  });
});
