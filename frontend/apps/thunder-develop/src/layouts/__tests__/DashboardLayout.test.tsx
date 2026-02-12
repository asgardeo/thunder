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

import {describe, it, expect, vi} from 'vitest';
import {render, screen} from '@thunder/test-utils';
import DashboardLayout from '../DashboardLayout';

// Mock components
vi.mock('../../components/Sidebar/SideMenu', () => ({
  default: ({defaultExpanded}: {defaultExpanded?: boolean}) => (
    <div data-testid="side-menu" data-default-expanded={defaultExpanded}>
      SideMenu
    </div>
  ),
}));

vi.mock('../../components/Header/Header', () => ({
  default: () => <div data-testid="header">Header</div>,
}));

vi.mock('../contexts/NavigationProvider', () => ({
  default: ({children}: {children: React.ReactNode}) => <div data-testid="navigation-provider">{children}</div>,
}));

// Mock Outlet
vi.mock('react-router', async () => {
  const actual = await vi.importActual<typeof import('react-router')>('react-router');
  return {
    ...actual,
    Outlet: () => <div data-testid="outlet">Outlet Content</div>,
  };
});

// Mock @wso2/oxygen-ui Layout
vi.mock('@wso2/oxygen-ui', async () => {
  const actual = await vi.importActual<typeof import('@wso2/oxygen-ui')>('@wso2/oxygen-ui');
  return {
    ...actual,
    Layout: Object.assign(
      ({children, ...props}: {children: React.ReactNode}) => <div data-testid="layout-root" {...props}>{children}</div>,
      {
        Sidebar: ({children}: {children: React.ReactNode}) => <div data-testid="layout-sidebar">{children}</div>,
        Content: ({children}: {children: React.ReactNode}) => <div data-testid="layout-content">{children}</div>,
        Header: ({children}: {children: React.ReactNode}) => <div data-testid="layout-header">{children}</div>,
      }
    ),
  };
});

describe('DashboardLayout', () => {
  it('renders NavigationProvider', () => {
    render(<DashboardLayout />);

    expect(screen.getByTestId('navigation-provider')).toBeInTheDocument();
  });

  it('renders Layout component', () => {
    render(<DashboardLayout />);

    expect(screen.getByTestId('layout-root')).toBeInTheDocument();
  });

  it('renders Layout.Sidebar with SideMenu', () => {
    render(<DashboardLayout />);

    expect(screen.getByTestId('layout-sidebar')).toBeInTheDocument();
    expect(screen.getByTestId('side-menu')).toBeInTheDocument();
  });

  it('renders Layout.Content', () => {
    render(<DashboardLayout />);

    expect(screen.getByTestId('layout-content')).toBeInTheDocument();
  });

  it('renders Layout.Header with Header component', () => {
    render(<DashboardLayout />);

    expect(screen.getByTestId('layout-header')).toBeInTheDocument();
    expect(screen.getByTestId('header')).toBeInTheDocument();
  });

  it('renders Outlet for nested routes', () => {
    render(<DashboardLayout />);

    expect(screen.getByTestId('outlet')).toBeInTheDocument();
  });

  it('renders complete layout structure', () => {
    render(<DashboardLayout />);

    expect(screen.getByTestId('navigation-provider')).toBeInTheDocument();
    expect(screen.getByTestId('layout-root')).toBeInTheDocument();
    expect(screen.getByTestId('layout-sidebar')).toBeInTheDocument();
    expect(screen.getByTestId('layout-content')).toBeInTheDocument();
    expect(screen.getByTestId('side-menu')).toBeInTheDocument();
    expect(screen.getByTestId('header')).toBeInTheDocument();
    expect(screen.getByTestId('outlet')).toBeInTheDocument();
  });

  it('passes defaultExpanded=true to SideMenu when dense=false by default', () => {
    render(<DashboardLayout />);

    const sideMenu = screen.getByTestId('side-menu');
    expect(sideMenu).toHaveAttribute('data-default-expanded', 'true');
  });

  it('passes defaultExpanded=false to SideMenu when dense=true', () => {
    render(<DashboardLayout dense />);

    const sideMenu = screen.getByTestId('side-menu');
    expect(sideMenu).toHaveAttribute('data-default-expanded', 'false');
  });
});
