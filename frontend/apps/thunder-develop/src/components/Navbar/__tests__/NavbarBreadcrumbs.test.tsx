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

  it('renders breadcrumbs for organization-units page', async () => {
    const mockUseNavigation = await import('@/layouts/contexts/useNavigation');
    vi.mocked(mockUseNavigation.default).mockReturnValue({
      currentPage: 'organization-units',
      setCurrentPage: vi.fn(),
      sidebarOpen: false,
      setSidebarOpen: vi.fn(),
      toggleSidebar: vi.fn(),
    });

    render(<NavbarBreadcrumbs />);

    expect(screen.getByText('Develop')).toBeInTheDocument();
    expect(screen.getByText('Organization Units')).toBeInTheDocument();
  });

  it('renders breadcrumbs for flows page', async () => {
    const mockUseNavigation = await import('@/layouts/contexts/useNavigation');
    vi.mocked(mockUseNavigation.default).mockReturnValue({
      currentPage: 'flows',
      setCurrentPage: vi.fn(),
      sidebarOpen: false,
      setSidebarOpen: vi.fn(),
      toggleSidebar: vi.fn(),
    });

    render(<NavbarBreadcrumbs />);

    expect(screen.getByText('Develop')).toBeInTheDocument();
    expect(screen.getByText('Flows')).toBeInTheDocument();
  });

  it('renders both breadcrumb segments', async () => {
    const mockUseNavigation = await import('@/layouts/contexts/useNavigation');
    vi.mocked(mockUseNavigation.default).mockReturnValue({
      currentPage: 'users',
      setCurrentPage: vi.fn(),
      sidebarOpen: false,
      setSidebarOpen: vi.fn(),
      toggleSidebar: vi.fn(),
    });

    const {container} = render(<NavbarBreadcrumbs />);

    const typographyElements = container.querySelectorAll('.MuiTypography-root');
    expect(typographyElements.length).toBeGreaterThanOrEqual(2);
  });

  it('renders current page text with body1 Typography variant', async () => {
    const mockUseNavigation = await import('@/layouts/contexts/useNavigation');
    vi.mocked(mockUseNavigation.default).mockReturnValue({
      currentPage: 'users',
      setCurrentPage: vi.fn(),
      sidebarOpen: false,
      setSidebarOpen: vi.fn(),
      toggleSidebar: vi.fn(),
    });

    render(<NavbarBreadcrumbs />);

    const usersElement = screen.getByText('Users');
    expect(usersElement).toBeInTheDocument();
    expect(usersElement).toHaveClass('MuiTypography-body1');
  });

  it('renders separator with ChevronRight icon wrapped in Box', async () => {
    const mockUseNavigation = await import('@/layouts/contexts/useNavigation');
    vi.mocked(mockUseNavigation.default).mockReturnValue({
      currentPage: 'dashboard',
      setCurrentPage: vi.fn(),
      sidebarOpen: false,
      setSidebarOpen: vi.fn(),
      toggleSidebar: vi.fn(),
    });

    const {container} = render(<NavbarBreadcrumbs />);

    const separators = container.querySelectorAll('.MuiBreadcrumbs-separator');
    expect(separators.length).toBeGreaterThan(0);

    // Separator should contain a Box with an SVG icon
    const separatorBox = separators[0].querySelector('.MuiBox-root');
    expect(separatorBox).toBeInTheDocument();
    const svgIcon = separatorBox?.querySelector('svg');
    expect(svgIcon).toBeInTheDocument();
  });

  it('renders styled breadcrumbs with separator alignment', async () => {
    const mockUseNavigation = await import('@/layouts/contexts/useNavigation');
    vi.mocked(mockUseNavigation.default).mockReturnValue({
      currentPage: 'users',
      setCurrentPage: vi.fn(),
      sidebarOpen: false,
      setSidebarOpen: vi.fn(),
      toggleSidebar: vi.fn(),
    });

    const {container} = render(<NavbarBreadcrumbs />);

    // Verify the breadcrumbs ol has aligned items
    const breadcrumbOl = container.querySelector('.MuiBreadcrumbs-ol');
    expect(breadcrumbOl).toBeInTheDocument();
    expect(breadcrumbOl).toHaveStyle({alignItems: 'center'});
  });

  it('preserves content after re-render', async () => {
    const mockUseNavigation = await import('@/layouts/contexts/useNavigation');
    vi.mocked(mockUseNavigation.default).mockReturnValue({
      currentPage: 'users',
      setCurrentPage: vi.fn(),
      sidebarOpen: false,
      setSidebarOpen: vi.fn(),
      toggleSidebar: vi.fn(),
    });

    const {rerender} = render(<NavbarBreadcrumbs />);

    rerender(<NavbarBreadcrumbs />);

    expect(screen.getByText('Develop')).toBeInTheDocument();
    expect(screen.getByText('Users')).toBeInTheDocument();
  });
});
