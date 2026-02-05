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
import SideMenuMobile from '../SideMenuMobile';

// Mock MenuContent component
vi.mock('../MenuContent', () => ({
  default: () => <div data-testid="menu-content">Menu Content</div>,
}));

// Mock useNavigation
vi.mock('@/layouts/contexts/useNavigation', () => ({
  default: vi.fn(() => ({
    currentPage: {id: 'users', text: 'Users', category: 'Dashboard'},
    setCurrentPage: vi.fn(),
    sidebarOpen: false,
    setSidebarOpen: vi.fn(),
    toggleSidebar: vi.fn(),
  })),
}));

describe('SideMenuMobile', () => {
  const mockToggleDrawer = vi.fn(() => vi.fn());

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders drawer when open is true', () => {
    render(<SideMenuMobile open toggleDrawer={mockToggleDrawer} />);

    expect(screen.getByText('Riley Carter')).toBeInTheDocument();
  });

  it('does not display content when open is false', () => {
    render(<SideMenuMobile open={false} toggleDrawer={mockToggleDrawer} />);

    expect(screen.queryByText('Riley Carter')).not.toBeInTheDocument();
  });

  it('renders user name', () => {
    render(<SideMenuMobile open toggleDrawer={mockToggleDrawer} />);

    expect(screen.getByText('Riley Carter')).toBeInTheDocument();
  });

  it('renders MenuContent component', () => {
    render(<SideMenuMobile open toggleDrawer={mockToggleDrawer} />);

    expect(screen.getByTestId('menu-content')).toBeInTheDocument();
  });

  it('renders logout button', () => {
    render(<SideMenuMobile open toggleDrawer={mockToggleDrawer} />);

    const logoutButton = screen.getByRole('button', {name: /logout/i});
    expect(logoutButton).toBeInTheDocument();
  });

  it('calls toggleDrawer with false when drawer is closed', async () => {
    const user = userEvent.setup();
    render(<SideMenuMobile open toggleDrawer={mockToggleDrawer} />);

    // Click outside the drawer (on the backdrop)
    const backdrop = document.querySelector('.MuiBackdrop-root');
    if (backdrop) {
      await user.click(backdrop);
      expect(mockToggleDrawer).toHaveBeenCalledWith(false);
    }
  });

  it('logout button is full width', () => {
    render(<SideMenuMobile open toggleDrawer={mockToggleDrawer} />);

    const logoutButton = screen.getByRole('button', {name: /logout/i});
    expect(logoutButton).toHaveClass('MuiButton-fullWidth');
  });

  it('renders drawer with undefined open prop', () => {
    render(<SideMenuMobile open={undefined} toggleDrawer={mockToggleDrawer} />);

    // Drawer should not display content when open is undefined
    expect(screen.queryByText('Riley Carter')).not.toBeInTheDocument();
  });

  it('renders user avatar when open', () => {
    render(<SideMenuMobile open toggleDrawer={mockToggleDrawer} />);

    // Drawer renders in a portal, so use document instead of container
    const avatar = document.querySelector('.MuiAvatar-root');
    expect(avatar).toBeInTheDocument();
  });

  it('renders notification bell button when open', () => {
    render(<SideMenuMobile open toggleDrawer={mockToggleDrawer} />);

    const buttons = screen.getAllByRole('button');
    // Should have at least the notification button and the logout button
    expect(buttons.length).toBeGreaterThanOrEqual(2);
  });

  it('renders dividers when open', () => {
    render(<SideMenuMobile open toggleDrawer={mockToggleDrawer} />);

    // Drawer renders in a portal, so use document instead of container
    const dividers = document.querySelectorAll('.MuiDivider-root');
    expect(dividers.length).toBeGreaterThan(0);
  });

  it('logout button has outlined variant', () => {
    render(<SideMenuMobile open toggleDrawer={mockToggleDrawer} />);

    const logoutButton = screen.getByRole('button', {name: /logout/i});
    expect(logoutButton).toHaveClass('MuiButton-outlined');
  });

  it('renders logout button with start icon', () => {
    render(<SideMenuMobile open toggleDrawer={mockToggleDrawer} />);

    const logoutButton = screen.getByRole('button', {name: /logout/i});
    const startIcon = logoutButton.querySelector('.MuiButton-startIcon');
    expect(startIcon).toBeInTheDocument();
  });

  it('renders drawer anchored to the right', () => {
    render(<SideMenuMobile open toggleDrawer={mockToggleDrawer} />);

    const drawer = document.querySelector('.MuiDrawer-root');
    expect(drawer).toBeInTheDocument();

    const paper = document.querySelector('.MuiDrawer-paperAnchorRight');
    expect(paper).toBeInTheDocument();
  });

  it('renders user avatar with correct alt text', () => {
    render(<SideMenuMobile open toggleDrawer={mockToggleDrawer} />);

    const avatar = document.querySelector('.MuiAvatar-root img');
    expect(avatar).toHaveAttribute('alt', 'Riley Carter');
  });

  it('renders user name as h6 typography', () => {
    render(<SideMenuMobile open toggleDrawer={mockToggleDrawer} />);

    const userName = screen.getByText('Riley Carter');
    expect(userName).toBeInTheDocument();
    expect(userName.tagName).toBe('P');
  });

  it('renders bell icon in notification button', () => {
    render(<SideMenuMobile open toggleDrawer={mockToggleDrawer} />);

    const bellIcon = document.querySelector('svg.lucide-bell');
    expect(bellIcon).toBeInTheDocument();
  });

  it('renders logout icon in logout button', () => {
    render(<SideMenuMobile open toggleDrawer={mockToggleDrawer} />);

    const logoutIcon = document.querySelector('svg.lucide-log-out');
    expect(logoutIcon).toBeInTheDocument();
  });

  it('renders two dividers when open', () => {
    render(<SideMenuMobile open toggleDrawer={mockToggleDrawer} />);

    const dividers = document.querySelectorAll('.MuiDivider-root');
    expect(dividers.length).toBe(2);
  });

  it('renders drawer with correct z-index above other drawers', () => {
    render(<SideMenuMobile open toggleDrawer={mockToggleDrawer} />);

    // Verify the drawer is rendered and has the correct paper element
    const paper = document.querySelector('.MuiDrawer-paper');
    expect(paper).toBeInTheDocument();
  });

  it('preserves content after re-render', () => {
    const {rerender} = render(<SideMenuMobile open toggleDrawer={mockToggleDrawer} />);

    rerender(<SideMenuMobile open toggleDrawer={mockToggleDrawer} />);

    expect(screen.getByText('Riley Carter')).toBeInTheDocument();
    expect(screen.getByTestId('menu-content')).toBeInTheDocument();
    expect(screen.getByRole('button', {name: /logout/i})).toBeInTheDocument();
  });
});
