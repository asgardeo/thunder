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
import {render, screen, userEvent} from '@thunder/test-utils';
import MenuContent from '../MenuContent';

const mockSetCurrentPage = vi.fn();

// Mock useNavigation hook
vi.mock('@/layouts/contexts/useNavigation', () => ({
  default: vi.fn(() => ({
    currentPage: 'users',
    setCurrentPage: mockSetCurrentPage,
    sidebarOpen: false,
    setSidebarOpen: vi.fn(),
    toggleSidebar: vi.fn(),
  })),
}));

describe('MenuContent', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders main menu items', () => {
    render(<MenuContent />);

    expect(screen.getByText('Users')).toBeInTheDocument();
    expect(screen.getByText('User Types')).toBeInTheDocument();
    expect(screen.getByText('Integrations')).toBeInTheDocument();
    expect(screen.getByText('Applications')).toBeInTheDocument();
  });

  it('renders icons for main menu items', () => {
    const {container} = render(<MenuContent />);

    expect(container.querySelector('svg.lucide-users-round')).toBeInTheDocument();
    expect(container.querySelector('svg.lucide-user')).toBeInTheDocument();
    expect(container.querySelector('svg.lucide-blocks')).toBeInTheDocument();
    expect(container.querySelector('svg.lucide-layout-grid')).toBeInTheDocument();
  });

  it('calls setCurrentPage when a menu item is clicked', async () => {
    const user = userEvent.setup();
    render(<MenuContent />);

    const usersLink = screen.getByText('Users');
    await user.click(usersLink);

    expect(mockSetCurrentPage).toHaveBeenCalledWith('users');
  });

  it('renders NavLink components with correct paths', () => {
    const {container} = render(<MenuContent />);

    const usersLink = container.querySelector('a[href="/users"]');
    expect(usersLink).toBeInTheDocument();

    const userTypesLink = container.querySelector('a[href="/user-types"]');
    expect(userTypesLink).toBeInTheDocument();
  });

  it('marks current page as selected', () => {
    render(<MenuContent />);

    // Users should be selected based on mock
    const usersButton = screen.getByText('Users').closest('a');
    expect(usersButton).toHaveClass('Mui-selected');
  });

  it('all menu items are clickable links', () => {
    render(<MenuContent />);

    const usersLink = screen.getByText('Users').closest('a');
    expect(usersLink).toBeInTheDocument();

    const userTypesLink = screen.getByText('User Types').closest('a');
    expect(userTypesLink).toBeInTheDocument();

    const integrationsLink = screen.getByText('Integrations').closest('a');
    expect(integrationsLink).toBeInTheDocument();

    const applicationsLink = screen.getByText('Applications').closest('a');
    expect(applicationsLink).toBeInTheDocument();
  });

  it('hides text labels in mini mode', async () => {
    const mockContext = {
      mini: true,
      fullyExpanded: false,
      fullyCollapsed: true,
      hasDrawerTransitions: true,
    };

    const {default: SidebarContext} = await import('../context/SidebarContext');

    const {container} = render(
      <SidebarContext.Provider value={mockContext}>
        <MenuContent />
      </SidebarContext.Provider>
    );

    // Text elements should have display: none in mini mode
    const listItemText = container.querySelector('.MuiListItemText-root');
    expect(listItemText).toHaveStyle({display: 'none'});
  });

  it('shows text labels in expanded mode', async () => {
    const mockContext = {
      mini: false,
      fullyExpanded: true,
      fullyCollapsed: false,
      hasDrawerTransitions: true,
    };

    const {default: SidebarContext} = await import('../context/SidebarContext');

    const {container} = render(
      <SidebarContext.Provider value={mockContext}>
        <MenuContent />
      </SidebarContext.Provider>
    );

    // Text elements should have display: block in expanded mode
    const listItemText = container.querySelector('.MuiListItemText-root');
    expect(listItemText).toHaveStyle({display: 'block'});
  });

  it('calls setCurrentPage with correct id for each menu item', async () => {
    const user = userEvent.setup();
    render(<MenuContent />);

    const userTypesLink = screen.getByText('User Types');
    await user.click(userTypesLink);
    expect(mockSetCurrentPage).toHaveBeenCalledWith('user-types');

    const integrationsLink = screen.getByText('Integrations');
    await user.click(integrationsLink);
    expect(mockSetCurrentPage).toHaveBeenCalledWith('integrations');

    const applicationsLink = screen.getByText('Applications');
    await user.click(applicationsLink);
    expect(mockSetCurrentPage).toHaveBeenCalledWith('applications');
  });

  it('marks different page as selected', async () => {
    const mockUseNavigation = await import('@/layouts/contexts/useNavigation');
    vi.mocked(mockUseNavigation.default).mockReturnValue({
      currentPage: 'integrations',
      setCurrentPage: mockSetCurrentPage,
      sidebarOpen: false,
      setSidebarOpen: vi.fn(),
      toggleSidebar: vi.fn(),
    });

    render(<MenuContent />);

    const integrationsButton = screen.getByText('Integrations').closest('a');
    expect(integrationsButton).toHaveClass('Mui-selected');
  });
});
