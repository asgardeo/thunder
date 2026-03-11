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

/* eslint-disable react/require-default-props */

import {describe, it, expect, vi, beforeEach} from 'vitest';
import {renderWithProviders} from '@thunder/test-utils/browser';
import {page, userEvent} from 'vitest/browser';
import DashboardLayout from '../DashboardLayout';

const mockSignIn = vi.fn();
const mockSignOut = vi.fn();
const mockLoggerError = vi.fn();
const mockUserData = vi.fn();

// Mock Asgardeo
vi.mock('@asgardeo/react', () => ({
  useAsgardeo: () => ({
    signIn: mockSignIn,
  }),
  User: ({children}: {children: (user: unknown) => React.ReactNode}) => children(mockUserData()),
  SignOutButton: ({children}: {children: (props: {signOut: () => void}) => React.ReactNode}) =>
    children({signOut: mockSignOut}),
}));

// Mock @thunder/logger/react
vi.mock('@thunder/logger/react', () => ({
  useLogger: () => ({
    error: mockLoggerError,
    info: vi.fn(),
    warn: vi.fn(),
    debug: vi.fn(),
  }),
}));

// Mock Outlet
vi.mock('react-router', async () => {
  const actual = await vi.importActual<typeof import('react-router')>('react-router');
  return {
    ...actual,
    Outlet: () => <div data-testid="outlet">Outlet Content</div>,
    Link: ({children, to}: {children: React.ReactNode; to: string}) => (
      <a href={to} data-testid="router-link">
        {children}
      </a>
    ),
  };
});

// Mock @wso2/oxygen-ui to avoid infinite re-render loops in browser mode.
// AppShell, Sidebar, Header, etc. use ResizeObserver/media queries that cause
// "Maximum update depth exceeded" in real browser rendering.
vi.mock('@wso2/oxygen-ui', () => {
  function MockChildren({children}: {children?: React.ReactNode}) {
    return <div>{children}</div>;
  }

  function Shell({children}: {children?: React.ReactNode}) {
    return <div data-testid="app-shell">{children}</div>;
  }
  Shell.Navbar = MockChildren;
  Shell.Sidebar = MockChildren;
  Shell.Main = function MockMain({children}: {children?: React.ReactNode}) {
    return <main data-testid="main">{children}</main>;
  };
  Shell.Footer = MockChildren;

  function HeaderComp({children}: {children?: React.ReactNode}) {
    return <header data-testid="header">{children}</header>;
  }
  HeaderComp.Toggle = function MockToggle() {
    return <button data-testid="header-toggle" type="button">Toggle</button>;
  };
  HeaderComp.Brand = MockChildren;
  HeaderComp.BrandLogo = MockChildren;
  HeaderComp.BrandTitle = MockChildren;
  HeaderComp.Spacer = function MockSpacer() {
    return <div data-testid="header-spacer" />;
  };
  HeaderComp.Actions = MockChildren;

  function SidebarComp({children}: {children?: React.ReactNode}) {
    return <div data-testid="sidebar">{children}</div>;
  }
  SidebarComp.Nav = MockChildren;
  SidebarComp.Category = MockChildren;
  SidebarComp.CategoryLabel = MockChildren;
  SidebarComp.Item = MockChildren;
  SidebarComp.ItemIcon = MockChildren;
  SidebarComp.ItemLabel = MockChildren;

  function UMenu({children}: {children?: React.ReactNode}) {
    return <div data-testid="user-menu">{children}</div>;
  }
  UMenu.Trigger = function MockTrigger({name}: {name?: string}) {
    return <button type="button" aria-label={name}>{name}</button>;
  };
  UMenu.Header = function MockHeader({name, email}: {name?: string; email?: string}) {
    return <div data-testid="user-menu-header">{name} {email}</div>;
  };
  UMenu.Divider = function MockDivider() {
    return <hr />;
  };
  UMenu.Logout = function MockLogout({label, onClick}: {label?: string; onClick?: () => void}) {
    return <button type="button" onClick={onClick}>{label}</button>;
  };

  function FooterComp({children}: {children?: React.ReactNode}) {
    return <div data-testid="footer">{children}</div>;
  }
  FooterComp.Copyright = MockChildren;
  FooterComp.Divider = function MockFooterDivider() {
    return <hr />;
  };
  FooterComp.Version = MockChildren;

  function MockColorSchemeImage() {
    return <img data-testid="color-scheme-image" alt="logo" />;
  }

  function MockColorSchemeToggle() {
    return <button data-testid="color-scheme-toggle" type="button">Theme</button>;
  }

  function MockDividerComp({orientation}: {orientation?: string}) {
    return <hr data-orientation={orientation} />;
  }

  function MockThemeProvider({children}: {children?: React.ReactNode}) {
    return children;
  }

  return {
    AppShell: Shell,
    Header: HeaderComp,
    Sidebar: SidebarComp,
    UserMenu: UMenu,
    Footer: FooterComp,
    ColorSchemeImage: MockColorSchemeImage,
    ColorSchemeToggle: MockColorSchemeToggle,
    Divider: MockDividerComp,
    OxygenUIThemeProvider: MockThemeProvider,
  };
});

describe('DashboardLayout', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUserData.mockReturnValue({name: 'Test User', email: 'test@example.com'});
  });

  it('renders AppShell layout', async () => {
    await renderWithProviders(<DashboardLayout />);

    // Check that the outlet is rendered
    await expect.element(page.getByTestId('outlet')).toBeInTheDocument();
  });

  it('renders Outlet for nested routes', async () => {
    await renderWithProviders(<DashboardLayout />);

    await expect.element(page.getByTestId('outlet')).toBeInTheDocument();
    await expect.element(page.getByTestId('outlet')).toHaveTextContent('Outlet Content');
  });

  it('renders navigation categories', async () => {
    await renderWithProviders(<DashboardLayout />);

    // Check for category labels
    await expect.element(page.getByText('Identities')).toBeInTheDocument();
    await expect.element(page.getByText('Resources')).toBeInTheDocument();
  });

  it('renders navigation items', async () => {
    await renderWithProviders(<DashboardLayout />);

    // Check for navigation items using translation keys
    await expect.element(page.getByText('Users')).toBeInTheDocument();
    await expect.element(page.getByText('User Types')).toBeInTheDocument();
    await expect.element(page.getByText('Applications')).toBeInTheDocument();
    await expect.element(page.getByText('Integrations')).toBeInTheDocument();
    await expect.element(page.getByText('Flows')).toBeInTheDocument();
  });

  it('renders footer', async () => {
    await renderWithProviders(<DashboardLayout />);

    const currentYear = new Date().getFullYear();
    await expect.element(page.getByText(new RegExp(currentYear.toString()))).toBeInTheDocument();
  });

  it('calls signIn after successful signOut when sign out is clicked', async () => {
    mockSignOut.mockResolvedValue(undefined);
    mockSignIn.mockResolvedValue(undefined);

    await renderWithProviders(<DashboardLayout />);

    // Open the user menu first
    await userEvent.click(page.getByLabelText('Test User'));

    // Click sign out menu item
    await userEvent.click(page.getByText('Sign Out'));

    await vi.waitFor(() => {
      expect(mockSignOut).toHaveBeenCalled();
      expect(mockSignIn).toHaveBeenCalled();
    });
  });

  it('logs error when signOut fails', async () => {
    const signOutError = new Error('Sign out failed');
    mockSignOut.mockRejectedValue(signOutError);

    await renderWithProviders(<DashboardLayout />);

    // Open the user menu first
    await userEvent.click(page.getByLabelText('Test User'));

    // Click sign out menu item
    await userEvent.click(page.getByText('Sign Out'));

    await vi.waitFor(() => {
      expect(mockSignOut).toHaveBeenCalled();
      expect(mockLoggerError).toHaveBeenCalledWith('Sign out/in failed', {error: signOutError});
    });
  });

  it('renders with fallback values when user data is missing', async () => {
    mockUserData.mockReturnValue(null);

    await renderWithProviders(<DashboardLayout />);

    await expect.element(page.getByTestId('outlet')).toBeInTheDocument();
  });

  it('renders with undefined user name and email', async () => {
    mockUserData.mockReturnValue({name: undefined, email: undefined});

    await renderWithProviders(<DashboardLayout />);

    await expect.element(page.getByTestId('outlet')).toBeInTheDocument();
  });
});
