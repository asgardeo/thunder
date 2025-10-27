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
import type {JSX} from 'react';
import render from '@/test/test-utils';
import SideMenu from './SideMenu';

interface MockUser {
  name: string;
  email: string;
}

// Mock the Asgardeo User component
vi.mock('@asgardeo/react', () => ({
  User: ({children}: {children: (user: MockUser) => JSX.Element}) => {
    const mockUser: MockUser = {
      name: 'John Doe',
      email: 'john.doe@example.com',
    };
    return <div data-testid="user-component">{children(mockUser)}</div>;
  },
}));

// Mock child components
vi.mock('./MenuContent', () => ({
  default: () => <div data-testid="menu-content">Menu Content</div>,
}));

vi.mock('./OptionsMenu', () => ({
  default: () => (
    <button type="button" data-testid="options-menu">
      Options
    </button>
  ),
}));

vi.mock('@thunder/ui', () => ({
  ThemedIcon: ({alt}: {alt: {light: string; dark: string}}) => <img data-testid="themed-icon" alt={alt.light} />,
}));

describe('SideMenu', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders the drawer component', () => {
    const {container} = render(<SideMenu />);

    const drawer = container.querySelector('.MuiDrawer-root');
    expect(drawer).toBeInTheDocument();
  });

  it('renders the logo', () => {
    render(<SideMenu />);

    const logo = screen.getByTestId('themed-icon');
    expect(logo).toBeInTheDocument();
  });

  it('renders the "Develop" title', () => {
    render(<SideMenu />);

    expect(screen.getByText('Develop')).toBeInTheDocument();
  });

  it('renders menu content', () => {
    render(<SideMenu />);

    expect(screen.getByTestId('menu-content')).toBeInTheDocument();
  });

  it('renders user information', () => {
    render(<SideMenu />);

    expect(screen.getByText('John Doe')).toBeInTheDocument();
    expect(screen.getByText('john.doe@example.com')).toBeInTheDocument();
  });

  it('renders user avatar with first letter', () => {
    render(<SideMenu />);

    const avatar = screen.getByText('J');
    expect(avatar).toBeInTheDocument();
  });

  it('renders options menu', () => {
    render(<SideMenu />);

    expect(screen.getByTestId('options-menu')).toBeInTheDocument();
  });

  it('has permanent variant', () => {
    const {container} = render(<SideMenu />);

    const drawer = container.querySelector('.MuiDrawer-root');
    expect(drawer).toHaveClass('MuiDrawer-docked');
  });

  it('renders divider after logo section', () => {
    const {container} = render(<SideMenu />);

    const dividers = container.querySelectorAll('.MuiDivider-root');
    expect(dividers.length).toBeGreaterThan(0);
  });
});
