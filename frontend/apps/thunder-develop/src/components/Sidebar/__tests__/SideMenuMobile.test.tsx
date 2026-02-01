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
});
