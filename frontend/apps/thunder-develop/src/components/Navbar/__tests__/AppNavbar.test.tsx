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
import AppNavbar, {CustomIcon} from '../AppNavbar';

// Mock child components
vi.mock('../../Sidebar/SideMenuMobile', () => ({
  default: ({open, toggleDrawer}: {open: boolean; toggleDrawer: (open: boolean) => () => void}) => (
    <div data-testid="side-menu-mobile" data-open={open}>
      <button type="button" onClick={toggleDrawer(false)}>
        Close
      </button>
    </div>
  ),
}));

vi.mock('@wso2/oxygen-ui', async () => {
  const actual = await vi.importActual('@wso2/oxygen-ui');
  return {
    ...actual,
    ColorSchemeToggle: () => (
      <button type="button" data-testid="theme-toggle">
        Theme
      </button>
    ),
  };
});

describe('AppNavbar', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders the app bar', () => {
    const {container} = render(<AppNavbar />);

    const appBar = container.querySelector('.MuiAppBar-root');
    expect(appBar).toBeInTheDocument();
  });

  it('renders Dashboard title', () => {
    render(<AppNavbar />);

    expect(screen.getByText('Dashboard')).toBeInTheDocument();
  });

  it('renders menu button', () => {
    render(<AppNavbar />);

    const menuButton = screen.getByRole('button', {name: /menu/i});
    expect(menuButton).toBeInTheDocument();
  });

  it('renders menu icon', () => {
    const {container} = render(<AppNavbar />);

    const menuIcon = container.querySelector('svg.lucide-menu');
    expect(menuIcon).toBeInTheDocument();
  });

  it('renders theme toggle button', () => {
    render(<AppNavbar />);

    expect(screen.getByTestId('theme-toggle')).toBeInTheDocument();
  });

  it('opens mobile menu when menu button is clicked', async () => {
    const user = userEvent.setup();
    render(<AppNavbar />);

    const menuButton = screen.getByRole('button', {name: /menu/i});
    await user.click(menuButton);

    const sideMenu = screen.getByTestId('side-menu-mobile');
    expect(sideMenu).toHaveAttribute('data-open', 'true');
  });

  it('closes mobile menu when toggleDrawer is called with false', async () => {
    const user = userEvent.setup();
    render(<AppNavbar />);

    // Open the menu first
    const menuButton = screen.getByRole('button', {name: /menu/i});
    await user.click(menuButton);

    // Close the menu
    const closeButton = screen.getByText('Close');
    await user.click(closeButton);

    const sideMenu = screen.getByTestId('side-menu-mobile');
    expect(sideMenu).toHaveAttribute('data-open', 'false');
  });

  it('has fixed position', () => {
    const {container} = render(<AppNavbar />);

    const appBar = container.querySelector('.MuiAppBar-root');
    expect(appBar).toHaveClass('MuiAppBar-positionFixed');
  });

  it('renders toolbar', () => {
    const {container} = render(<AppNavbar />);

    const toolbar = container.querySelector('.MuiToolbar-root');
    expect(toolbar).toBeInTheDocument();
  });

  it('renders Dashboard as h1 heading', () => {
    render(<AppNavbar />);

    const heading = screen.getByRole('heading', {level: 1, name: /dashboard/i});
    expect(heading).toBeInTheDocument();
  });
});

describe('CustomIcon', () => {
  it('renders the custom icon', () => {
    const {container} = render(<CustomIcon />);

    const box = container.querySelector('.MuiBox-root');
    expect(box).toBeInTheDocument();
  });

  it('renders dashboard icon inside custom icon', () => {
    const {container} = render(<CustomIcon />);

    const dashboardIcon = container.querySelector('svg.lucide-layout-dashboard');
    expect(dashboardIcon).toBeInTheDocument();
  });
});
