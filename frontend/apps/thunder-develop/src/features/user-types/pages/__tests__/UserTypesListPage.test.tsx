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
import {render, screen, waitFor, userEvent} from '@thunder/test-utils';
import UserTypesListPage from '../UserTypesListPage';

const mockNavigate = vi.fn();
const mockReload = vi.fn();
const mockLoggerError = vi.fn();

// Mock react-router
vi.mock('react-router', async () => {
  const actual = await vi.importActual<typeof import('react-router')>('react-router');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

// Mock logger
vi.mock('@thunder/logger/react', () => ({
  useLogger: () => ({
    error: mockLoggerError,
    info: vi.fn(),
    warn: vi.fn(),
    debug: vi.fn(),
  }),
}));

// Mock the UserTypesList component
vi.mock('../../components/UserTypesList', () => ({
  default: () => (
    <div data-testid="user-types-list">
      <span>Organization Unit: Root Organization</span>
    </div>
  ),
}));

describe('UserTypesListPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockLoggerError.mockReset();
    Object.defineProperty(window, 'location', {
      value: {reload: mockReload},
      writable: true,
    });
  });

  it('renders page title', () => {
    render(<UserTypesListPage />);

    expect(screen.getByText('User Types')).toBeInTheDocument();
  });

  it('renders page description', () => {
    render(<UserTypesListPage />);

    expect(screen.getByText('Define a new user type schema for your organization')).toBeInTheDocument();
  });

  it('renders create user type button', () => {
    render(<UserTypesListPage />);

    const createButton = screen.getByRole('button', {name: /create user type/i});
    expect(createButton).toBeInTheDocument();
  });

  it('renders UserTypesList component', () => {
    render(<UserTypesListPage />);

    expect(screen.getByTestId('user-types-list')).toBeInTheDocument();
  });

  it('shows organization unit name within the list component', () => {
    render(<UserTypesListPage />);

    expect(screen.getByText('Organization Unit: Root Organization')).toBeInTheDocument();
  });

  it('navigates to create page when create button is clicked', async () => {
    const user = userEvent.setup();
    render(<UserTypesListPage />);

    const createButton = screen.getByRole('button', {name: /create user type/i});
    await user.click(createButton);

    expect(mockNavigate).toHaveBeenCalledWith('/user-types/create');
  });

  it('handles navigation error when create button is clicked and logs error', async () => {
    const user = userEvent.setup();
    const testError = new Error('Navigation failed');
    mockNavigate.mockRejectedValue(testError);

    render(<UserTypesListPage />);

    const createButton = screen.getByRole('button', {name: /create user type/i});
    await user.click(createButton);

    expect(mockNavigate).toHaveBeenCalledWith('/user-types/create');

    // Wait for the error to be logged
    await waitFor(() => {
      expect(mockLoggerError).toHaveBeenCalledWith(
        'Failed to navigate to create user type page',
        expect.objectContaining({error: testError}),
      );
    });
  });
});
