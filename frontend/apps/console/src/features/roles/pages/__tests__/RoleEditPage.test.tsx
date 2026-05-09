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
import {render, screen} from '@thunderid/test-utils';
import type {NavigateFunction} from 'react-router';
import {describe, it, expect, vi, beforeEach, afterEach} from 'vitest';
import type {Role} from '../../models/role';
import RoleEditPage from '../RoleEditPage';

// Mock dependencies
vi.mock('../../api/useGetRole');
vi.mock('../../api/useUpdateRole');

vi.mock('../../components/RoleDeleteDialog', () => ({
  default: ({open}: {open: boolean}) => (open ? <div data-testid="delete-dialog">Delete Dialog</div> : null),
}));

vi.mock('../../components/edit-role/general-settings/EditGeneralSettings', () => ({
  default: ({onDeleteClick}: {onDeleteClick: () => void}) => (
    <div data-testid="edit-general-settings">
      <button type="button" onClick={onDeleteClick}>
        Delete
      </button>
    </div>
  ),
}));

vi.mock('../../components/edit-role/assignments-settings/EditAssignmentsSettings', () => ({
  default: () => <div data-testid="edit-assignments-settings">Assignments Settings</div>,
}));

vi.mock('@thunderid/components', () => ({
  CopyableId: vi.fn(() => null),
}));

vi.mock('react-router', async () => {
  const actual = await vi.importActual('react-router');
  return {
    ...actual,
    useParams: vi.fn(),
    useNavigate: vi.fn(),
    Link: ({children, to}: {children: React.ReactNode; to: string}) => <a href={to}>{children}</a>,
  };
});

vi.mock('@thunderid/logger/react', () => ({
  useLogger: () => ({
    error: vi.fn(),
    info: vi.fn(),
    warn: vi.fn(),
    debug: vi.fn(),
  }),
}));

vi.mock('@thunderid/contexts', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@thunderid/contexts')>();
  return {
    ...actual,
    useToast: () => ({showToast: vi.fn()}),
  };
});

const {default: useGetRole} = await import('../../api/useGetRole');
const {default: useUpdateRole} = await import('../../api/useUpdateRole');
const {useParams, useNavigate} = await import('react-router');

describe('RoleEditPage', () => {
  let mockNavigate: ReturnType<typeof vi.fn>;

  const mockRole: Role = {
    id: 'role-1',
    name: 'Admin Role',
    description: 'Administrator role',
    ouId: 'ou-1',
    permissions: [],
  };

  beforeEach(() => {
    mockNavigate = vi.fn();

    vi.mocked(useParams).mockReturnValue({roleId: 'role-1'});
    vi.mocked(useNavigate).mockReturnValue(mockNavigate as unknown as NavigateFunction);

    vi.mocked(useGetRole).mockReturnValue({
      data: mockRole,
      isLoading: false,
      error: null,
      refetch: vi.fn(),
    } as unknown as ReturnType<typeof useGetRole>);

    vi.mocked(useUpdateRole).mockReturnValue({
      mutate: vi.fn(),
      mutateAsync: vi.fn(),
      isPending: false,
      isError: false,
      isSuccess: false,
      error: null,
      data: undefined,
      reset: vi.fn(),
      context: undefined,
      failureCount: 0,
      failureReason: null,
      isIdle: true,
      isPaused: false,
      status: 'idle',
      submittedAt: 0,
      variables: undefined,
    });
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe('Loading State', () => {
    it('should show CircularProgress while loading', () => {
      vi.mocked(useGetRole).mockReturnValue({
        data: undefined,
        isLoading: true,
        error: null,
        refetch: vi.fn(),
      } as unknown as ReturnType<typeof useGetRole>);

      render(<RoleEditPage />);

      expect(screen.getByRole('progressbar')).toBeInTheDocument();
    });
  });

  describe('Error State', () => {
    it('should show error alert when fetch fails', () => {
      vi.mocked(useGetRole).mockReturnValue({
        data: undefined,
        isLoading: false,
        error: new Error('Failed to load'),
        refetch: vi.fn(),
      } as unknown as ReturnType<typeof useGetRole>);

      render(<RoleEditPage />);

      expect(screen.getByRole('alert')).toBeInTheDocument();
    });
  });

  describe('Rendering (with role data)', () => {
    it('should display role name in header', () => {
      render(<RoleEditPage />);

      expect(screen.getByText('Admin Role')).toBeInTheDocument();
    });

    it('should render role description', () => {
      render(<RoleEditPage />);

      expect(screen.getByText('Administrator role')).toBeInTheDocument();
    });

    it('should render two tabs', () => {
      render(<RoleEditPage />);

      const tabs = screen.getAllByRole('tab');
      expect(tabs).toHaveLength(2);
    });

    it('should show General tab panel by default', () => {
      render(<RoleEditPage />);

      expect(screen.getByTestId('edit-general-settings')).toBeInTheDocument();
    });
  });

  describe('Tab Navigation', () => {
    it('should switch to Assignments tab panel when Assignments tab clicked', async () => {
      const user = userEvent.setup();
      render(<RoleEditPage />);

      const tabs = screen.getAllByRole('tab');
      await user.click(tabs[1]);

      expect(screen.getByTestId('edit-assignments-settings')).toBeInTheDocument();
    });
  });

  describe('Delete Flow', () => {
    it('should open delete dialog when delete is triggered', async () => {
      const user = userEvent.setup();
      render(<RoleEditPage />);

      const deleteButton = screen.getByRole('button', {name: 'Delete'});
      await user.click(deleteButton);

      expect(screen.getByTestId('delete-dialog')).toBeInTheDocument();
    });
  });
});
