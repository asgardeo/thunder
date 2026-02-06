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
import Header from '../Header';

// Mock the child components and dependencies
vi.mock('../../Navbar/NavbarBreadcrumbs', () => ({
  default: () => <div data-testid="navbar-breadcrumbs">Breadcrumbs</div>,
}));

vi.mock('../Search', () => ({
  default: () => <input data-testid="search" placeholder="Search" />,
}));

vi.mock('@thunder/ui', () => ({
  ColorModeIconDropdown: () => (
    <button type="button" data-testid="theme-toggle">
      Theme Toggle
    </button>
  ),
}));

vi.mock('@/layouts/contexts/useNavigation', () => ({
  default: vi.fn(() => ({
    currentPage: {category: 'Dashboard', text: 'Users'},
    navigate: vi.fn(),
  })),
}));

describe('Header', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders all header components', () => {
    render(<Header />);

    expect(screen.getByTestId('navbar-breadcrumbs')).toBeInTheDocument();
    expect(screen.getByTestId('search')).toBeInTheDocument();
    expect(screen.getByTestId('theme-toggle')).toBeInTheDocument();
    expect(screen.getByRole('button', {name: /open notifications/i})).toBeInTheDocument();
    expect(screen.getByRole('button', {name: /change language/i})).toBeInTheDocument();
  });

  it('renders notification button with bell icon', () => {
    const {container} = render(<Header />);

    const bellIcon = container.querySelector('svg.lucide-bell');
    expect(bellIcon).toBeInTheDocument();
  });

  it('renders menu button for mobile navigation', () => {
    render(<Header />);

    const menuButton = screen.getByRole('button', {name: /menu/i});
    expect(menuButton).toBeInTheDocument();
  });

  it('renders with correct layout structure', () => {
    const {container} = render(<Header />);

    const stacks = container.querySelectorAll('.MuiStack-root');
    expect(stacks.length).toBeGreaterThanOrEqual(1);
  });

  it('renders notifications tooltip', () => {
    render(<Header />);

    const notificationsButton = screen.getByRole('button', {name: /open notifications/i});
    expect(notificationsButton).toBeInTheDocument();
  });

  it('renders header with correct spacing and alignment', () => {
    const {container} = render(<Header />);

    // Verify the outermost Stack renders as a row
    const outerStack = container.querySelector('.MuiStack-root');
    expect(outerStack).toBeInTheDocument();
    expect(outerStack).toHaveStyle({flexDirection: 'row'});
  });

  it('renders LanguageSwitcher component', () => {
    render(<Header />);

    const languageButton = screen.getByRole('button', {name: /change language/i});
    expect(languageButton).toBeInTheDocument();
  });

  it('renders ColorSchemeToggle component', () => {
    render(<Header />);

    expect(screen.getByTestId('theme-toggle')).toBeInTheDocument();
  });

  it('renders both left and right sections', () => {
    const {container} = render(<Header />);

    // Left section: menu button + breadcrumbs
    expect(screen.getByRole('button', {name: /menu/i})).toBeInTheDocument();
    expect(screen.getByTestId('navbar-breadcrumbs')).toBeInTheDocument();

    // Right section: search, notifications, language, theme
    expect(screen.getByTestId('search')).toBeInTheDocument();
    expect(screen.getByRole('button', {name: /open notifications/i})).toBeInTheDocument();

    // Verify multiple stacks are rendered for the layout
    const stacks = container.querySelectorAll('.MuiStack-root');
    expect(stacks.length).toBeGreaterThanOrEqual(3);
  });

  it('preserves content after re-render', () => {
    const {rerender} = render(<Header />);

    // Trigger a re-render to exercise memoization branches
    rerender(<Header />);

    expect(screen.getByTestId('navbar-breadcrumbs')).toBeInTheDocument();
    expect(screen.getByTestId('search')).toBeInTheDocument();
    expect(screen.getByRole('button', {name: /open notifications/i})).toBeInTheDocument();
    expect(screen.getByRole('button', {name: /change language/i})).toBeInTheDocument();
  });
});
