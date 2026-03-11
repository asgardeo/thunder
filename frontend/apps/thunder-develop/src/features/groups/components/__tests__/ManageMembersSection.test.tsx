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
import {page} from 'vitest/browser';
import {renderWithProviders} from '@thunder/test-utils/browser';
import type * as OxygenUI from '@wso2/oxygen-ui';
import ManageMembersSection from '../edit-group/members-settings/ManageMembersSection';

interface MockDataGridProps {
  rows?: {id: string; [key: string]: unknown}[];
  columns?: {
    field?: string;
    renderCell?: (params: {row: Record<string, unknown>; field: string; value: unknown; id: string}) => React.ReactNode;
  }[];
  loading?: boolean;
}

vi.mock('@wso2/oxygen-ui', async () => {
  const actual = await vi.importActual<typeof OxygenUI>('@wso2/oxygen-ui');
  return {
    ...actual,
    DataGrid: {
      ...(actual.DataGrid ?? {}),
      DataGrid: ({rows = [], columns = [], loading = false}: MockDataGridProps) => (
        <div data-testid="members-grid" data-loading={loading}>
          {rows.map((row) => (
            <div key={row.id} data-testid={`member-${row.id}`}>
              {columns.map((column) => {
                if (!column?.field || !column.renderCell) return null;
                return (
                  <span key={`${row.id}-${column.field}`}>
                    {column.renderCell({row, field: column.field, value: row[column.field], id: String(row.id)})}
                  </span>
                );
              })}
            </div>
          ))}
        </div>
      ),
    },
  };
});

const mockUseGetGroupMembers = vi.fn();
vi.mock('../../api/useGetGroupMembers', () => ({
  default: (...args: unknown[]): unknown => mockUseGetGroupMembers(...args),
}));

describe('ManageMembersSection', () => {
  const defaultProps = {
    groupId: 'g1',
    onRemoveMember: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
    mockUseGetGroupMembers.mockReturnValue({
      data: {
        totalResults: 2,
        startIndex: 0,
        count: 2,
        members: [
          {id: 'u1', type: 'user'},
          {id: 'g2', type: 'group'},
        ],
      },
      isLoading: false,
    });
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('should render the section title', async () => {
    await renderWithProviders(<ManageMembersSection {...defaultProps} />);

    // Use getByRole to find the heading with accessible name "Members"
    // SettingsCard renders the title as Typography h5 which has role "heading"
    await expect.element(page.getByRole('heading', {name: 'Members', level: 5})).toBeInTheDocument();
  });

  it('should render members in the data grid', async () => {
    await renderWithProviders(<ManageMembersSection {...defaultProps} />);

    await expect.element(page.getByTestId('member-u1')).toBeInTheDocument();
    await expect.element(page.getByTestId('member-g2')).toBeInTheDocument();
  });

  it('should show loading state', async () => {
    mockUseGetGroupMembers.mockReturnValue({
      data: null,
      isLoading: true,
    });
    await renderWithProviders(<ManageMembersSection {...defaultProps} />);

    await expect.element(page.getByTestId('members-grid')).toHaveAttribute('data-loading', 'true');
  });

  it('should render header action when provided', async () => {
    await renderWithProviders(
      <ManageMembersSection {...defaultProps} headerAction={<button type="button">Add Member Action</button>} />,
    );

    await expect.element(page.getByRole('button', {name: 'Add Member Action'})).toBeInTheDocument();
  });

  it('should call useGetGroupMembers with groupId and pagination params', async () => {
    await renderWithProviders(<ManageMembersSection {...defaultProps} />);

    expect(mockUseGetGroupMembers).toHaveBeenCalledWith('g1', {limit: 10, offset: 0});
  });

  it('should call onRemoveMember when remove button is clicked', async () => {
    await renderWithProviders(<ManageMembersSection {...defaultProps} />);

    // The actions column renderCell creates an IconButton with aria-label "Remove"
    const removeButtons = page.getByRole('button', {name: /remove/i}).all();
    expect(removeButtons.length).toBeGreaterThan(0);
    await removeButtons[0].click();

    expect(defaultProps.onRemoveMember).toHaveBeenCalledWith({id: 'u1', type: 'user'});
  });
});
