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

import userEvent from '@testing-library/user-event';
import {render, screen, waitFor} from '@thunderid/test-utils';
import {describe, it, expect, vi, beforeEach, afterEach} from 'vitest';
import AddAssignmentDialog from '../AddAssignmentDialog';

vi.mock('@thunderid/configure-users');
vi.mock('../../../../../groups/api/useGetGroups');
vi.mock('../../../../../applications/api/useGetApplications');
vi.mock('../../../../api/useGetRoleAssignments');
vi.mock('@thunderid/hooks', () => ({
  useDataGridLocaleText: vi.fn(),
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => {
      const translations: Record<string, string> = {
        'roles:assignments.dialog.title': 'Add Assignments',
        'roles:assignments.dialog.tabs.users': 'Users',
        'roles:assignments.dialog.tabs.groups': 'Groups',
        'roles:assignments.dialog.tabs.apps': 'Apps',
        'roles:assignments.dialog.columns.displayName': 'Display Name',
        'roles:assignments.dialog.columns.userType': 'User Type',
        'roles:assignments.dialog.columns.name': 'Name',
        'roles:assignments.dialog.columns.description': 'Description',
        'roles:assignments.dialog.add': 'Add',
        'roles:assignments.dialog.fetchError': 'Failed to fetch data',
        'common:actions.cancel': 'Cancel',
      };
      return translations[key] || key;
    },
  }),
}));

vi.mock('@wso2/oxygen-ui', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@wso2/oxygen-ui')>();
  return {
    ...actual,
    DataGrid: {
      DataGrid: ({
        rows,
        columns,
        checkboxSelection = false,
        onRowSelectionModelChange = undefined,
        getRowId = undefined,
      }: {
        rows: Record<string, unknown>[];
        columns: {
          field: string;
          renderCell?: (params: {row: Record<string, unknown>}) => React.ReactElement;
          valueGetter?: (value: unknown, row: Record<string, unknown>) => string;
        }[];
        checkboxSelection?: boolean;
        onRowSelectionModelChange?: (model: {type: string; ids: Set<string>}) => void;
        getRowId?: (row: Record<string, unknown>) => string;
      }) => (
        <div role="grid" data-testid="data-grid">
          {rows.map((row) => {
            const rowId = getRowId ? getRowId(row) : (row.id as string);
            return (
              <div key={rowId} role="row" data-row-id={rowId}>
                {checkboxSelection && (
                  <input
                    type="checkbox"
                    data-testid={`checkbox-${String(row.id)}`}
                    onChange={(e) => {
                      if (onRowSelectionModelChange) {
                        const ids = new Set<string>();
                        if (e.target.checked) ids.add(row.id as string);
                        onRowSelectionModelChange({type: 'include', ids});
                      }
                    }}
                  />
                )}
                {columns.map((col) => {
                  if (col.renderCell) return <div key={col.field}>{col.renderCell({row})}</div>;
                  if (col.valueGetter) return <div key={col.field}>{col.valueGetter(null, row)}</div>;
                  return <div key={col.field}>{row[col.field] as string}</div>;
                })}
              </div>
            );
          })}
        </div>
      ),
    },
  };
});

const {useGetUsers} = await import('@thunderid/configure-users');
const {default: useGetGroups} = await import('../../../../../groups/api/useGetGroups');
const {default: useGetApplications} = await import('../../../../../applications/api/useGetApplications');
const {default: useGetRoleAssignments} = await import('../../../../api/useGetRoleAssignments');
const {useDataGridLocaleText} = await import('@thunderid/hooks');

describe('AddAssignmentDialog', () => {
  const mockOnClose = vi.fn();
  const mockOnAdd = vi.fn();

  const defaultProps = {
    open: true,
    roleId: 'role-1',
    onClose: mockOnClose,
    onAdd: mockOnAdd,
  };

  const mockUsersData = {
    totalResults: 2,
    startIndex: 0,
    count: 2,
    users: [
      {id: 'user-1', ouId: 'ou-1', type: 'local', display: 'John Doe'},
      {id: 'user-2', ouId: 'ou-1', type: 'federated', display: 'Jane Smith'},
    ],
  };

  const mockGroupsData = {
    totalResults: 2,
    startIndex: 0,
    count: 2,
    groups: [
      {id: 'group-1', name: 'Admins', description: 'Admin group', ouId: 'ou-1'},
      {id: 'group-2', name: 'Viewers', description: undefined, ouId: 'ou-1'},
    ],
  };

  const mockExistingUserAssignments = {
    totalResults: 1,
    startIndex: 0,
    count: 1,
    assignments: [{id: 'user-2', type: 'user', display: 'Jane Smith'}],
  };

  const mockExistingGroupAssignments = {
    totalResults: 0,
    startIndex: 0,
    count: 0,
    assignments: [],
  };
  const mockApplicationsData = {
    totalResults: 2,
    count: 2,
    applications: [
      {id: 'app-1', name: 'Orders API', description: 'Orders backend service'},
      {id: 'app-2', name: 'Billing API', description: 'Billing backend service'},
    ],
  };
  const mockExistingAppAssignments = {
    totalResults: 1,
    startIndex: 0,
    count: 1,
    assignments: [{id: 'app-2', type: 'app', display: 'Billing API'}],
  };

  const renderComponent = (props = {}) => render(<AddAssignmentDialog {...defaultProps} {...props} />);

  beforeEach(() => {
    vi.mocked(useDataGridLocaleText).mockReturnValue({});

    vi.mocked(useGetUsers).mockReturnValue({
      data: mockUsersData,
      isLoading: false,
      error: null,
    } as unknown as ReturnType<typeof useGetUsers>);

    vi.mocked(useGetGroups).mockReturnValue({
      data: mockGroupsData,
      isLoading: false,
      error: null,
    } as unknown as ReturnType<typeof useGetGroups>);
    vi.mocked(useGetApplications).mockReturnValue({
      data: mockApplicationsData,
      isLoading: false,
      error: null,
    } as unknown as ReturnType<typeof useGetApplications>);

    vi.mocked(useGetRoleAssignments).mockImplementation(
      (params: {type?: string}) =>
        ({
          data:
            params.type === 'user'
              ? mockExistingUserAssignments
              : params.type === 'group'
                ? mockExistingGroupAssignments
                : mockExistingAppAssignments,
          isLoading: false,
          error: null,
        }) as unknown as ReturnType<typeof useGetRoleAssignments>,
    );
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('should render dialog when open is true', () => {
      renderComponent();

      expect(screen.getByRole('dialog')).toBeInTheDocument();
      expect(screen.getByText('Add Assignments')).toBeInTheDocument();
    });

    it('should not render dialog when open is false', () => {
      renderComponent({open: false});

      expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    });

    it('should render Users tab as active by default', () => {
      renderComponent();

      expect(screen.getByText('Users')).toBeInTheDocument();
      expect(screen.getByText('Groups')).toBeInTheDocument();
      expect(screen.getByText('Apps')).toBeInTheDocument();
    });

    it('should filter out already-assigned users', () => {
      renderComponent();

      expect(screen.getByText('John Doe')).toBeInTheDocument();
      expect(screen.queryByText('Jane Smith')).not.toBeInTheDocument();
    });

    it('should render Cancel and Add buttons', () => {
      renderComponent();

      expect(screen.getByRole('button', {name: /Cancel/})).toBeInTheDocument();
      expect(screen.getByRole('button', {name: /Add/})).toBeInTheDocument();
    });

    it('should show Add button as disabled when no selections', () => {
      renderComponent();

      expect(screen.getByRole('button', {name: /Add/})).toBeDisabled();
    });

    it('should render with Groups tab active when initialTab is 1', () => {
      renderComponent({initialTab: 1});

      expect(screen.getByText('Admins')).toBeInTheDocument();
    });
  });

  describe('User Interactions', () => {
    it('should switch to Groups tab when clicked', async () => {
      const user = userEvent.setup();
      renderComponent();

      await user.click(screen.getByText('Groups'));

      expect(screen.getByText('Admins')).toBeInTheDocument();
    });

    it('should switch to Apps tab when clicked', async () => {
      const user = userEvent.setup();
      renderComponent();

      await user.click(screen.getByText('Apps'));

      expect(screen.getByText('Orders API')).toBeInTheDocument();
      expect(screen.queryByText('Billing API')).not.toBeInTheDocument();
    });

    it('should call onClose when Cancel is clicked', async () => {
      const user = userEvent.setup();
      renderComponent();

      await user.click(screen.getByRole('button', {name: /Cancel/}));

      expect(mockOnClose).toHaveBeenCalledTimes(1);
    });

    it('should call onAdd with user assignments when Add is clicked after selecting', async () => {
      const user = userEvent.setup();
      renderComponent();

      const checkbox = screen.getByTestId('checkbox-user-1');
      await user.click(checkbox);

      const addButton = screen.getByRole('button', {name: /Add/});
      await user.click(addButton);

      expect(mockOnAdd).toHaveBeenCalledTimes(1);
      expect(mockOnAdd).toHaveBeenCalledWith([{id: 'user-1', type: 'user'}]);
    });

    it('should show selection count on Add button', async () => {
      const user = userEvent.setup();
      renderComponent();

      const checkbox = screen.getByTestId('checkbox-user-1');
      await user.click(checkbox);

      await waitFor(() => {
        expect(screen.getByRole('button', {name: /Add.*\(1\)/})).toBeInTheDocument();
      });
    });
  });

  describe('Error States', () => {
    it('should show error alert when users fetch fails', () => {
      vi.mocked(useGetUsers).mockReturnValue({
        data: undefined,
        isLoading: false,
        error: new Error('User fetch failed'),
      } as unknown as ReturnType<typeof useGetUsers>);

      renderComponent();

      expect(screen.getByText('User fetch failed')).toBeInTheDocument();
    });

    it('should show error alert when groups fetch fails', async () => {
      const user = userEvent.setup();
      vi.mocked(useGetGroups).mockReturnValue({
        data: undefined,
        isLoading: false,
        error: new Error('Group fetch failed'),
      } as unknown as ReturnType<typeof useGetGroups>);

      renderComponent();
      await user.click(screen.getByText('Groups'));

      expect(screen.getByText('Group fetch failed')).toBeInTheDocument();
    });

    it('should show error alert when apps fetch fails', async () => {
      const user = userEvent.setup();
      vi.mocked(useGetApplications).mockReturnValue({
        data: undefined,
        isLoading: false,
        error: new Error('App fetch failed'),
      } as unknown as ReturnType<typeof useGetApplications>);

      renderComponent();
      await user.click(screen.getByText('Apps'));

      expect(screen.getByText('App fetch failed')).toBeInTheDocument();
    });

    it('should not show error alert while loading', () => {
      vi.mocked(useGetUsers).mockReturnValue({
        data: undefined,
        isLoading: true,
        error: new Error('User fetch failed'),
      } as unknown as ReturnType<typeof useGetUsers>);

      renderComponent();

      expect(screen.queryByText('User fetch failed')).not.toBeInTheDocument();
    });
  });

  describe('Edge Cases', () => {
    it('should handle empty user list', () => {
      vi.mocked(useGetUsers).mockReturnValue({
        data: {totalResults: 0, startIndex: 0, count: 0, users: []},
        isLoading: false,
        error: null,
      } as unknown as ReturnType<typeof useGetUsers>);

      renderComponent();

      expect(screen.getByRole('grid')).toBeInTheDocument();
      expect(screen.queryByRole('row')).not.toBeInTheDocument();
    });

    it('should display user ID when display name is missing', () => {
      vi.mocked(useGetUsers).mockReturnValue({
        data: {
          totalResults: 1,
          startIndex: 0,
          count: 1,
          users: [{id: 'user-no-display', ouId: 'ou-1', type: 'local'}],
        },
        isLoading: false,
        error: null,
      } as unknown as ReturnType<typeof useGetUsers>);

      vi.mocked(useGetRoleAssignments).mockImplementation(
        () =>
          ({
            data: {totalResults: 0, startIndex: 0, count: 0, assignments: []},
            isLoading: false,
            error: null,
          }) as unknown as ReturnType<typeof useGetRoleAssignments>,
      );

      renderComponent();

      const matches = screen.getAllByText('user-no-display');
      expect(matches.length).toBeGreaterThanOrEqual(1);
    });

    it('should show "-" for group with no description', async () => {
      const user = userEvent.setup();
      renderComponent();

      await user.click(screen.getByText('Groups'));

      expect(screen.getByText('-')).toBeInTheDocument();
    });
  });
});
