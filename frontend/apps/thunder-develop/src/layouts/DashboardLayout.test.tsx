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
import {screen} from '@testing-library/react';
import render from '@/test/test-utils';
import DashboardLayout from './DashboardLayout';

// Mock components
vi.mock('../components/Sidebar/SideMenu', () => ({
  default: () => <div data-testid="side-menu">SideMenu</div>,
}));

vi.mock('../components/Header/Header', () => ({
  default: () => <div data-testid="header">Header</div>,
}));

vi.mock('./contexts/NavigationProvider', () => ({
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

// Mock @thunder/ui Layout
vi.mock('@thunder/ui', () => ({
  Layout: {
    Provider: ({children}: {children: React.ReactNode}) => <div data-testid="layout-provider">{children}</div>,
    Root: ({children}: {children: React.ReactNode}) => <div data-testid="layout-root">{children}</div>,
    Sidebar: ({children}: {children: React.ReactNode}) => <div data-testid="layout-sidebar">{children}</div>,
    Content: ({children}: {children: React.ReactNode}) => <div data-testid="layout-content">{children}</div>,
    Header: ({children}: {children: React.ReactNode}) => <div data-testid="layout-header">{children}</div>,
  },
}));

describe('DashboardLayout', () => {
  it('renders NavigationProvider', () => {
    render(<DashboardLayout />);

    expect(screen.getByTestId('navigation-provider')).toBeInTheDocument();
  });

  it('renders Layout.Provider', () => {
    render(<DashboardLayout />);

    expect(screen.getByTestId('layout-provider')).toBeInTheDocument();
  });

  it('renders Layout.Root', () => {
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
    expect(screen.getByTestId('layout-provider')).toBeInTheDocument();
    expect(screen.getByTestId('layout-root')).toBeInTheDocument();
    expect(screen.getByTestId('layout-sidebar')).toBeInTheDocument();
    expect(screen.getByTestId('layout-content')).toBeInTheDocument();
    expect(screen.getByTestId('side-menu')).toBeInTheDocument();
    expect(screen.getByTestId('header')).toBeInTheDocument();
    expect(screen.getByTestId('outlet')).toBeInTheDocument();
  });
});
