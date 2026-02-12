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
import {render, screen} from '@thunder/test-utils';
import NavbarBreadcrumbs from '../NavbarBreadcrumbs';

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
      currentPage: 'dashboard',
      setCurrentPage: vi.fn(),
      sidebarOpen: false,
      setSidebarOpen: vi.fn(),
      toggleSidebar: vi.fn(),
    });

    render(<NavbarBreadcrumbs />);

    expect(screen.getByText('Develop')).toBeInTheDocument();
    expect(screen.getByText('Dashboard')).toBeInTheDocument();
  });

  it('renders breadcrumbs with different category and page', async () => {
    const mockUseNavigation = await import('@/layouts/contexts/useNavigation');
    vi.mocked(mockUseNavigation.default).mockReturnValue({
      currentPage: 'users',
      setCurrentPage: vi.fn(),
      sidebarOpen: false,
      setSidebarOpen: vi.fn(),
      toggleSidebar: vi.fn(),
    });

    render(<NavbarBreadcrumbs />);

    expect(screen.getByText('Develop')).toBeInTheDocument();
    expect(screen.getByText('Users')).toBeInTheDocument();
  });

  it('has correct aria-label', async () => {
    const mockUseNavigation = await import('@/layouts/contexts/useNavigation');
    vi.mocked(mockUseNavigation.default).mockReturnValue({
      currentPage: 'users',
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
      currentPage: 'dashboard',
      setCurrentPage: vi.fn(),
      sidebarOpen: false,
      setSidebarOpen: vi.fn(),
      toggleSidebar: vi.fn(),
    });

    const {container} = render(<NavbarBreadcrumbs />);

    // Check for lucide-react ChevronRight icon
    const separator = container.querySelector('svg');
    expect(separator).toBeInTheDocument();
  });

  it('renders breadcrumbs for user-types page', async () => {
    const mockUseNavigation = await import('@/layouts/contexts/useNavigation');
    vi.mocked(mockUseNavigation.default).mockReturnValue({
      currentPage: 'user-types',
      setCurrentPage: vi.fn(),
      sidebarOpen: false,
      setSidebarOpen: vi.fn(),
      toggleSidebar: vi.fn(),
    });

    render(<NavbarBreadcrumbs />);

    expect(screen.getByText('Develop')).toBeInTheDocument();
    expect(screen.getByText('User Types')).toBeInTheDocument();
  });

  it('renders breadcrumbs for integrations page', async () => {
    const mockUseNavigation = await import('@/layouts/contexts/useNavigation');
    vi.mocked(mockUseNavigation.default).mockReturnValue({
      currentPage: 'integrations',
      setCurrentPage: vi.fn(),
      sidebarOpen: false,
      setSidebarOpen: vi.fn(),
      toggleSidebar: vi.fn(),
    });

    render(<NavbarBreadcrumbs />);

    expect(screen.getByText('Develop')).toBeInTheDocument();
    expect(screen.getByText('Integrations')).toBeInTheDocument();
  });

  it('renders breadcrumbs for applications page', async () => {
    const mockUseNavigation = await import('@/layouts/contexts/useNavigation');
    vi.mocked(mockUseNavigation.default).mockReturnValue({
      currentPage: 'applications',
      setCurrentPage: vi.fn(),
      sidebarOpen: false,
      setSidebarOpen: vi.fn(),
      toggleSidebar: vi.fn(),
    });

    render(<NavbarBreadcrumbs />);

    expect(screen.getByText('Develop')).toBeInTheDocument();
    expect(screen.getByText('Applications')).toBeInTheDocument();
  });

  it('falls back to page ID when translation key not found', async () => {
    const mockUseNavigation = await import('@/layouts/contexts/useNavigation');
    vi.mocked(mockUseNavigation.default).mockReturnValue({
      currentPage: 'unknown-page',
      setCurrentPage: vi.fn(),
      sidebarOpen: false,
      setSidebarOpen: vi.fn(),
      toggleSidebar: vi.fn(),
    });

    render(<NavbarBreadcrumbs />);

    expect(screen.getByText('Develop')).toBeInTheDocument();
    // Should display the page ID when translation not found
    expect(screen.getByText('unknown-page')).toBeInTheDocument();
  });

  it('applies correct styling for styled breadcrumbs', async () => {
    const mockUseNavigation = await import('@/layouts/contexts/useNavigation');
    vi.mocked(mockUseNavigation.default).mockReturnValue({
      currentPage: 'users',
      setCurrentPage: vi.fn(),
      sidebarOpen: false,
      setSidebarOpen: vi.fn(),
      toggleSidebar: vi.fn(),
    });

    const {container} = render(<NavbarBreadcrumbs />);

    const breadcrumbs = container.querySelector('.MuiBreadcrumbs-root');
    expect(breadcrumbs).toBeInTheDocument();
  });
});
