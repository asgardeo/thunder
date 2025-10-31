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
import Header from './Header';

// Mock the child components and dependencies
vi.mock('../Navbar/NavbarBreadcrumbs', () => ({
  default: () => <div data-testid="navbar-breadcrumbs">Breadcrumbs</div>,
}));

vi.mock('./Search', () => ({
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
  });

  it('renders notifications button with badge', () => {
    render(<Header />);

    const notificationsButton = screen.getByRole('button', {name: /open notifications/i});
    expect(notificationsButton).toBeInTheDocument();
  });

  it('renders notifications icon', () => {
    const {container} = render(<Header />);

    const notificationIcon = container.querySelector('svg.lucide-bell');
    expect(notificationIcon).toBeInTheDocument();
  });

  it('has correct layout structure', () => {
    const {container} = render(<Header />);

    const stack = container.querySelector('.MuiStack-root');
    expect(stack).toBeInTheDocument();
  });

  it('contains search functionality', () => {
    render(<Header />);

    const searchInput = screen.getByTestId('search');
    expect(searchInput).toBeInTheDocument();
  });

  it('contains theme toggle', () => {
    render(<Header />);

    const themeToggle = screen.getByTestId('theme-toggle');
    expect(themeToggle).toBeInTheDocument();
  });
});
