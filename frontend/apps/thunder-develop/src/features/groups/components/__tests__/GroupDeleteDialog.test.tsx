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

  it('should render dialog when open', async () => {
    await renderWithProviders(<GroupDeleteDialog {...defaultProps} />);

    await expect.element(page.getByText('Delete Group')).toBeInTheDocument();
    await expect.element(page.getByText('Are you sure you want to delete this group?')).toBeInTheDocument();
    await expect.element(page.getByText('This action cannot be undone. All group associations will be permanently removed.')).toBeInTheDocument();
  });

  it('should not render content when closed', async () => {
    await renderWithProviders(<GroupDeleteDialog {...defaultProps} open={false} />);

    await expect.element(page.getByText('Delete Group')).not.toBeInTheDocument();
  });

  it('should call onClose when cancel is clicked', async () => {
    await renderWithProviders(<GroupDeleteDialog {...defaultProps} />);

    await userEvent.click(page.getByText('Cancel'));

    expect(defaultProps.onClose).toHaveBeenCalled();
  });

  it('should call mutate when confirm is clicked', async () => {
    await renderWithProviders(<GroupDeleteDialog {...defaultProps} />);

    await userEvent.click(page.getByRole('button', {name: 'Delete'}));

    expect(mockMutate).toHaveBeenCalledWith('g1', expect.objectContaining({onSuccess: expect.any(Function) as unknown, onError: expect.any(Function) as unknown}));
  });

  it('should call onSuccess and onClose on successful delete', async () => {
    mockMutate.mockImplementation((_id: string, opts: {onSuccess: () => void}) => {
      opts.onSuccess();
    });

    await renderWithProviders(<GroupDeleteDialog {...defaultProps} />);

    await userEvent.click(page.getByRole('button', {name: 'Delete'}));

    expect(defaultProps.onClose).toHaveBeenCalled();
    expect(defaultProps.onSuccess).toHaveBeenCalled();
  });

  it('should display error on failed delete', async () => {
    mockMutate.mockImplementation((_id: string, opts: {onError: (err: Error) => void}) => {
      opts.onError(new Error('Delete failed'));
    });

    await renderWithProviders(<GroupDeleteDialog {...defaultProps} />);

    await userEvent.click(page.getByRole('button', {name: 'Delete'}));

    await expect.element(page.getByText('Delete failed')).toBeInTheDocument();
  });

  it('should not call mutate when groupId is null', async () => {
    await renderWithProviders(<GroupDeleteDialog {...defaultProps} groupId={null} />);

    // Button should be disabled when groupId is null, preventing mutation
    await expect.element(page.getByRole('button', {name: 'Delete'})).toBeDisabled();
    expect(mockMutate).not.toHaveBeenCalled();
  });

  it('should disable delete button when groupId is null', async () => {
    await renderWithProviders(<GroupDeleteDialog {...defaultProps} groupId={null} />);

    await expect.element(page.getByRole('button', {name: 'Delete'})).toBeDisabled();
  });
});
