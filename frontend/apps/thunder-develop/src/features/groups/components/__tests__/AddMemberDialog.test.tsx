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
import {screen} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {renderWithProviders} from '@thunder/test-utils';
import type * as OxygenUI from '@wso2/oxygen-ui';
import AddMemberDialog from '../edit-group/members-settings/AddMemberDialog';

interface MockDataGridProps {
  rows?: {id: string; [key: string]: unknown}[];
  loading?: boolean;
  checkboxSelection?: boolean;
  onRowSelectionModelChange?: (model: {type: string; ids: Set<string>}) => void;
}

vi.mock('@wso2/oxygen-ui', async () => {
  const actual = await vi.importActual<typeof OxygenUI>('@wso2/oxygen-ui');
  return {
    ...actual,
    DataGrid: {
      ...(actual.DataGrid ?? {}),
      DataGrid: ({rows = [], loading = false, checkboxSelection = false, onRowSelectionModelChange = undefined}: MockDataGridProps) => (
        <div data-testid="users-grid" data-loading={loading}>
          {rows.map((row) => (
            <div key={row.id} data-testid={`user-${row.id}`}>
              {checkboxSelection && (
                <input
                  type="checkbox"
                  data-testid={`checkbox-${row.id}`}
                  onChange={() => {
                    onRowSelectionModelChange?.({type: 'include', ids: new Set([row.id])});
                  }}
                />
              )}
              {row.id}
            </div>
          ))}
        </div>
      ),
    },
  };
});

const mockUseGetUsers = vi.fn();
vi.mock('../../../users/api/useGetUsers', () => ({
  default: (...args: unknown[]): unknown => mockUseGetUsers(...args),
}));

describe('AddMemberDialog', () => {
  const defaultProps = {
    open: true,
    onClose: vi.fn(),
    onAdd: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
    mockUseGetUsers.mockReturnValue({
      data: {
        totalResults: 2,
        startIndex: 0,
        count: 2,
        users: [
          {id: 'u1', organizationUnit: 'ou1', type: 'Person'},
          {id: 'u2', organizationUnit: 'ou2', type: 'Person'},
        ],
      },
      loading: false,
    });
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('should render dialog when open', () => {
    renderWithProviders(<AddMemberDialog {...defaultProps} />);

    expect(screen.getByText('addMember.title')).toBeInTheDocument();
  });

  it('should not render when closed', () => {
    renderWithProviders(<AddMemberDialog {...defaultProps} open={false} />);

    expect(screen.queryByText('addMember.title')).not.toBeInTheDocument();
  });

  it('should render users in the grid', () => {
    renderWithProviders(<AddMemberDialog {...defaultProps} />);

    expect(screen.getByTestId('user-u1')).toBeInTheDocument();
    expect(screen.getByTestId('user-u2')).toBeInTheDocument();
  });

  it('should show loading state', () => {
    mockUseGetUsers.mockReturnValue({
      data: null,
      loading: true,
    });
    renderWithProviders(<AddMemberDialog {...defaultProps} />);

    expect(screen.getByTestId('users-grid')).toHaveAttribute('data-loading', 'true');
  });

  it('should show no results alert when no users', () => {
    mockUseGetUsers.mockReturnValue({
      data: {totalResults: 0, startIndex: 0, count: 0, users: []},
      loading: false,
    });
    renderWithProviders(<AddMemberDialog {...defaultProps} />);

    expect(screen.getByText('addMember.noResults')).toBeInTheDocument();
  });

  it('should disable add button when no selection', () => {
    renderWithProviders(<AddMemberDialog {...defaultProps} />);

    const addButton = screen.getByText('addMember.add').closest('button');
    expect(addButton).toBeDisabled();
  });

  it('should enable add button after selecting a user', async () => {
    const user = userEvent.setup();
    renderWithProviders(<AddMemberDialog {...defaultProps} />);

    await user.click(screen.getByTestId('checkbox-u1'));

    const addButton = screen.getByText('addMember.add').closest('button');
    expect(addButton).not.toBeDisabled();
  });

  it('should call onAdd with selected members', async () => {
    const user = userEvent.setup();
    renderWithProviders(<AddMemberDialog {...defaultProps} />);

    await user.click(screen.getByTestId('checkbox-u1'));
    await user.click(screen.getByText('addMember.add'));

    expect(defaultProps.onAdd).toHaveBeenCalledWith([{id: 'u1', type: 'user'}]);
  });

  it('should call onClose when cancel is clicked', async () => {
    const user = userEvent.setup();
    renderWithProviders(<AddMemberDialog {...defaultProps} />);

    await user.click(screen.getByText('Cancel'));

    expect(defaultProps.onClose).toHaveBeenCalled();
  });

  it('should call useGetUsers with pagination params', () => {
    renderWithProviders(<AddMemberDialog {...defaultProps} />);

    expect(mockUseGetUsers).toHaveBeenCalledWith({limit: 10, offset: 0});
  });

  it('should show error alert when users fetch fails', () => {
    mockUseGetUsers.mockReturnValue({
      data: null,
      loading: false,
      error: {code: 'FETCH_ERROR', message: 'Network error', description: 'Failed to fetch users'},
    });
    renderWithProviders(<AddMemberDialog {...defaultProps} />);

    expect(screen.getByText('Network error')).toBeInTheDocument();
    expect(screen.queryByText('addMember.noResults')).not.toBeInTheDocument();
  });
});
