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
import {screen} from '@testing-library/react';
import render from '@/test/test-utils';
import NavbarBreadcrumbs from './NavbarBreadcrumbs';

// Mock the useNavigation hook
vi.mock('@/layouts/contexts/useNavigation', () => ({
  default: vi.fn(),
}));

describe('NavbarBreadcrumbs', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders breadcrumbs with category and page text', async () => {
    const mockUseNavigation = await import('@/layouts/contexts/useNavigation');
    vi.mocked(mockUseNavigation.default).mockReturnValue({
      currentPage: {
        id: 'home',
        category: 'Dashboard',
        text: 'Home',
      },
      setCurrentPage: vi.fn(),
      sidebarOpen: false,
      setSidebarOpen: vi.fn(),
      toggleSidebar: vi.fn(),
    });

    render(<NavbarBreadcrumbs />);

    expect(screen.getByText('Dashboard')).toBeInTheDocument();
    expect(screen.getByText('Home')).toBeInTheDocument();
  });

  it('renders breadcrumbs with different category and page', async () => {
    const mockUseNavigation = await import('@/layouts/contexts/useNavigation');
    vi.mocked(mockUseNavigation.default).mockReturnValue({
      currentPage: {
        id: 'users-list',
        category: 'Users',
        text: 'User List',
      },
      setCurrentPage: vi.fn(),
      sidebarOpen: false,
      setSidebarOpen: vi.fn(),
      toggleSidebar: vi.fn(),
    });

    render(<NavbarBreadcrumbs />);

    expect(screen.getByText('Users')).toBeInTheDocument();
    expect(screen.getByText('User List')).toBeInTheDocument();
  });

  it('has correct aria-label', async () => {
    const mockUseNavigation = await import('@/layouts/contexts/useNavigation');
    vi.mocked(mockUseNavigation.default).mockReturnValue({
      currentPage: {
        id: 'profile',
        category: 'Settings',
        text: 'Profile',
      },
      setCurrentPage: vi.fn(),
      sidebarOpen: false,
      setSidebarOpen: vi.fn(),
      toggleSidebar: vi.fn(),
    });

    render(<NavbarBreadcrumbs />);

    const breadcrumbs = screen.getByLabelText('breadcrumb');
    expect(breadcrumbs).toBeInTheDocument();
  });

  it('renders separator icon between breadcrumbs', async () => {
    const mockUseNavigation = await import('@/layouts/contexts/useNavigation');
    vi.mocked(mockUseNavigation.default).mockReturnValue({
      currentPage: {
        id: 'home',
        category: 'Dashboard',
        text: 'Home',
      },
      setCurrentPage: vi.fn(),
      sidebarOpen: false,
      setSidebarOpen: vi.fn(),
      toggleSidebar: vi.fn(),
    });

    const {container} = render(<NavbarBreadcrumbs />);

    const separator = container.querySelector('svg[data-testid="NavigateNextRoundedIcon"]');
    expect(separator).toBeInTheDocument();
  });
});
