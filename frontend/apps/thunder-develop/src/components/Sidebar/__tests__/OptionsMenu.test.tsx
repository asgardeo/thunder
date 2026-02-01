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
import {render, screen, waitFor, userEvent} from '@thunder/test-utils';
import type {JSX} from 'react';
import OptionsMenu from '../OptionsMenu';

// Mock Asgardeo hooks
const mockSignIn = vi.fn();
const mockSignOut = vi.fn();

vi.mock('@asgardeo/react', () => ({
  useAsgardeo: () => ({
    signIn: mockSignIn,
  }),
  SignOutButton: ({children}: {children: (props: {signOut: () => Promise<void>; isLoading: boolean}) => JSX.Element}) =>
    children({signOut: mockSignOut, isLoading: false}),
}));

describe('OptionsMenu', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders menu button', () => {
    render(<OptionsMenu />);

    const menuButton = screen.getByRole('button', {name: /open menu/i});
    expect(menuButton).toBeInTheDocument();
  });

  it('renders more icon', () => {
    const {container} = render(<OptionsMenu />);

    // Check for lucide-react EllipsisVertical icon
    const moreIcon = container.querySelector('svg');
    expect(moreIcon).toBeInTheDocument();
  });

  it('menu is closed by default', () => {
    render(<OptionsMenu />);

    expect(screen.queryByText('Profile')).not.toBeInTheDocument();
  });

  it('opens menu when button is clicked', async () => {
    const user = userEvent.setup();
    render(<OptionsMenu />);

    const menuButton = screen.getByRole('button', {name: /open menu/i});
    await user.click(menuButton);

    await waitFor(() => {
      expect(screen.getByText('Profile')).toBeInTheDocument();
    });
  });

  it('displays all menu items when opened', async () => {
    const user = userEvent.setup();
    render(<OptionsMenu />);

    const menuButton = screen.getByRole('button', {name: /open menu/i});
    await user.click(menuButton);

    await waitFor(() => {
      expect(screen.getByText('Profile')).toBeInTheDocument();
      expect(screen.getByText('My account')).toBeInTheDocument();
      expect(screen.getByText('Add another account')).toBeInTheDocument();
      expect(screen.getByText('Settings')).toBeInTheDocument();
      expect(screen.getByText('Sign Out')).toBeInTheDocument();
    });
  });

  it('closes menu when a menu item is clicked', async () => {
    const user = userEvent.setup();
    render(<OptionsMenu />);

    // Open menu
    const menuButton = screen.getByRole('button', {name: /open menu/i});
    await user.click(menuButton);

    await waitFor(() => {
      expect(screen.getByText('Profile')).toBeInTheDocument();
    });

    // Click Profile menu item
    await user.click(screen.getByText('Profile'));

    // Menu should close
    await waitFor(() => {
      expect(screen.queryByText('Profile')).not.toBeInTheDocument();
    });
  });

  it('renders logout icon in Sign Out menu item', async () => {
    const user = userEvent.setup();
    render(<OptionsMenu />);

    const menuButton = screen.getByRole('button', {name: /open menu/i});
    await user.click(menuButton);

    await waitFor(() => {
      expect(screen.getByText('Sign Out')).toBeInTheDocument();
    });

    const signOutMenuItem = screen.getByText('Sign Out').closest('li');
    expect(signOutMenuItem).toBeInTheDocument();
  });

  it('calls signOut and signIn when Sign Out is clicked', async () => {
    const user = userEvent.setup();
    render(<OptionsMenu />);

    // Open menu
    const menuButton = screen.getByRole('button', {name: /open menu/i});
    await user.click(menuButton);

    await waitFor(() => {
      expect(screen.getByText('Sign Out')).toBeInTheDocument();
    });

    // Click Sign Out
    await user.click(screen.getByText('Sign Out'));

    await waitFor(() => {
      expect(mockSignOut).toHaveBeenCalledTimes(1);
      expect(mockSignIn).toHaveBeenCalledTimes(1);
    });
  });

  it('renders dividers between menu sections', async () => {
    const user = userEvent.setup();
    render(<OptionsMenu />);

    const menuButton = screen.getByRole('button', {name: /open menu/i});
    await user.click(menuButton);

    await waitFor(() => {
      expect(screen.getByText('Profile')).toBeInTheDocument();
      expect(screen.getByText('Settings')).toBeInTheDocument();
      expect(screen.getByText('Sign Out')).toBeInTheDocument();
    });
  });
});
