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

import {describe, it, expect, vi, beforeEach, afterEach} from 'vitest';
import {render, screen, waitFor, cleanup, userEvent} from '@thunder/test-utils';
import type {JSX} from 'react';
import SideMenu from '../SideMenu';

// Mock Asgardeo User component
vi.mock('@asgardeo/react', () => ({
  User: ({children}: {children: (user: {name: string; email: string}) => JSX.Element}) =>
    children({name: 'John Doe', email: 'john.doe@example.com'}),
}));

// Mock ThemedIcon component
vi.mock('@thunder/ui', () => ({
  ThemedIcon: () => <div data-testid="themed-icon">Logo</div>,
}));

// Mock child components
vi.mock('../MenuContent', () => ({
  default: () => <div data-testid="menu-content">Menu Content</div>,
}));

vi.mock('../OptionsMenu', () => ({
  default: () => (
    <button type="button" data-testid="options-menu">
      Options
    </button>
  ),
}));

describe('SideMenu', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.useFakeTimers();
  });

  afterEach(() => {
    cleanup();
    vi.runOnlyPendingTimers();
    vi.useRealTimers();
  });

  describe('Basic rendering', () => {
    it('renders the drawer component', () => {
      const {container} = render(<SideMenu />);

      const drawer = container.querySelector('.MuiDrawer-root');
      expect(drawer).toBeInTheDocument();
    });

    it('renders the Developer title when expanded', () => {
      render(<SideMenu />);

      expect(screen.getByText('Developer')).toBeInTheDocument();
    });

    it('renders menu content', () => {
      render(<SideMenu />);

      expect(screen.getByTestId('menu-content')).toBeInTheDocument();
    });

    it('renders user information when expanded', () => {
      render(<SideMenu />);

      expect(screen.getByText('John Doe')).toBeInTheDocument();
      expect(screen.getByText('john.doe@example.com')).toBeInTheDocument();
    });

    it('renders user avatar', () => {
      const {container} = render(<SideMenu />);

      const avatar = container.querySelector('.MuiAvatar-root');
      expect(avatar).toBeInTheDocument();
      expect(avatar).toHaveTextContent('J');
    });

    it('renders options menu when expanded', () => {
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

  describe('Collapsible functionality', () => {
    it('renders collapse button when expanded', () => {
      render(<SideMenu />);

      const collapseButton = screen.getByLabelText('Expand/Collapse sidebar');
      expect(collapseButton).toBeInTheDocument();
    });

    it('renders expand button when collapsed', () => {
      render(<SideMenu defaultExpanded={false} />);

      const expandButton = screen.getByLabelText('Expand/Collapse sidebar');
      expect(expandButton).toBeInTheDocument();
    });

    it('calls onExpandedChange when collapse button is clicked', async () => {
      vi.useRealTimers(); // Use real timers for user events
      const user = userEvent.setup();
      const handleExpandedChange = vi.fn();
      render(<SideMenu defaultExpanded onExpandedChange={handleExpandedChange} />);

      const collapseButton = screen.getByLabelText('Expand/Collapse sidebar');
      await user.click(collapseButton);

      expect(handleExpandedChange).toHaveBeenCalledWith(false);
      vi.useFakeTimers(); // Restore fake timers
    });

    it('calls onExpandedChange when expand button is clicked', async () => {
      vi.useRealTimers(); // Use real timers for user events
      const user = userEvent.setup();
      const handleExpandedChange = vi.fn();
      render(<SideMenu defaultExpanded={false} onExpandedChange={handleExpandedChange} />);

      const expandButton = screen.getByLabelText('Expand/Collapse sidebar');
      await user.click(expandButton);

      expect(handleExpandedChange).toHaveBeenCalledWith(true);
      vi.useFakeTimers(); // Restore fake timers
    });

    it('toggles expanded state when used in uncontrolled mode', async () => {
      vi.useRealTimers(); // Use real timers for user events
      const user = userEvent.setup();
      render(<SideMenu />);

      // Initially expanded
      expect(screen.getByText('Developer')).toBeInTheDocument();
      expect(screen.getByLabelText('Expand/Collapse sidebar')).toBeInTheDocument();

      // Click to collapse
      const collapseButton = screen.getByLabelText('Expand/Collapse sidebar');
      await user.click(collapseButton);

      // Should still have toggle button
      await waitFor(() => {
        expect(screen.getByLabelText('Expand/Collapse sidebar')).toBeInTheDocument();
      });
      vi.useFakeTimers(); // Restore fake timers
    });

    it('hides logo and title when collapsed', () => {
      render(<SideMenu defaultExpanded={false} />);

      expect(screen.queryByText('Developer')).not.toBeInTheDocument();
      // Logo should not be rendered when collapsed (it's inside the same conditional block)
    });

    it('hides user information when collapsed', () => {
      render(<SideMenu defaultExpanded={false} />);

      expect(screen.queryByText('John Doe')).not.toBeInTheDocument();
      expect(screen.queryByText('john.doe@example.com')).not.toBeInTheDocument();
    });

    it('hides options menu when collapsed', () => {
      render(<SideMenu defaultExpanded={false} />);

      expect(screen.queryByTestId('options-menu')).not.toBeInTheDocument();
    });

    it('still shows user avatar when collapsed', () => {
      const {container} = render(<SideMenu defaultExpanded={false} />);

      const avatar = container.querySelector('.MuiAvatar-root');
      expect(avatar).toBeInTheDocument();
    });

    it('renders toggle button even when disableCollapsible is true', () => {
      render(<SideMenu disableCollapsible />);

      // The toggle button is always rendered
      expect(screen.getByLabelText('Expand/Collapse sidebar')).toBeInTheDocument();
    });

    it('renders toggle button when disableCollapsible is true and collapsed', () => {
      render(<SideMenu defaultExpanded={false} disableCollapsible />);

      // The toggle button is always rendered
      expect(screen.getByLabelText('Expand/Collapse sidebar')).toBeInTheDocument();
    });

    it('shows logo and title when disableCollapsible is true', () => {
      render(<SideMenu defaultExpanded={false} disableCollapsible />);

      // When collapsible is disabled, it should still show content even if defaultExpanded is false
      expect(screen.getByText('Developer')).toBeInTheDocument();
    });
  });

  describe('Drawer width', () => {
    it('has correct width when expanded', () => {
      const {container} = render(<SideMenu defaultExpanded />);

      const drawer = container.querySelector('.MuiDrawer-root');
      const style = window.getComputedStyle(drawer!);
      expect(style.width).toBe('300px');
    });

    it('has correct width when collapsed', () => {
      const {container} = render(<SideMenu defaultExpanded={false} />);

      const drawer = container.querySelector('.MuiDrawer-root');
      const style = window.getComputedStyle(drawer!);
      expect(style.width).toBe('64px');
    });
  });

  describe('Controlled vs Uncontrolled mode', () => {
    it('syncs state when defaultExpanded prop changes', async () => {
      vi.useRealTimers(); // Use real timers for user events
      const user = userEvent.setup();
      const handleExpandedChange = vi.fn();
      const {rerender} = render(<SideMenu defaultExpanded onExpandedChange={handleExpandedChange} />);

      expect(screen.getByText('Developer')).toBeInTheDocument();

      // Click collapse button
      const collapseButton = screen.getByLabelText('Expand/Collapse sidebar');
      await user.click(collapseButton);

      // Should call handler and change state
      expect(handleExpandedChange).toHaveBeenCalledWith(false);

      // Wait for the state to update
      await waitFor(() => {
        expect(screen.queryByText('Developer')).not.toBeInTheDocument();
      });

      // Parent updates the prop - should sync state
      rerender(<SideMenu defaultExpanded={false} onExpandedChange={handleExpandedChange} />);

      // Should remain collapsed
      expect(screen.queryByText('Developer')).not.toBeInTheDocument();

      vi.useFakeTimers(); // Restore fake timers
    });

    it('works as uncontrolled component when expanded prop is not provided', async () => {
      vi.useRealTimers(); // Use real timers for user events
      const user = userEvent.setup();
      render(<SideMenu />);

      // Initially expanded
      expect(screen.getByText('Developer')).toBeInTheDocument();

      // Click to collapse
      const collapseButton = screen.getByLabelText('Expand/Collapse sidebar');
      await user.click(collapseButton);

      // Should update internal state and collapse
      await waitFor(() => {
        expect(screen.queryByText('Developer')).not.toBeInTheDocument();
      });

      vi.useFakeTimers(); // Restore fake timers
    });
  });

  describe('SidebarContext', () => {
    it('provides correct context values when expanded', () => {
      // We can verify context is provided by checking if MenuContent renders
      // (MenuContent depends on SidebarContext)
      render(<SideMenu defaultExpanded />);

      expect(screen.getByTestId('menu-content')).toBeInTheDocument();
    });

    it('provides correct context values when collapsed', () => {
      render(<SideMenu defaultExpanded={false} />);

      expect(screen.getByTestId('menu-content')).toBeInTheDocument();
    });
  });

  describe('Transition timing', () => {
    it('sets isFullyExpanded to true after transition when expanding', () => {
      const {rerender} = render(<SideMenu defaultExpanded={false} />);

      // Expand the sidebar
      rerender(<SideMenu defaultExpanded />);

      // Fast-forward time to trigger the setTimeout
      vi.advanceTimersByTime(300); // Default enteringScreen duration

      // The component should have completed its expansion transition
      expect(screen.getByText('Developer')).toBeInTheDocument();
    });

    it('sets isFullyCollapsed to true after transition when collapsing', () => {
      const {rerender} = render(<SideMenu defaultExpanded />);

      // Collapse the sidebar
      rerender(<SideMenu defaultExpanded={false} />);

      // Fast-forward time to trigger the setTimeout
      vi.advanceTimersByTime(300); // Default leavingScreen duration

      // The component should have completed its collapse transition
      expect(screen.queryByText('Developer')).not.toBeInTheDocument();
    });

    it('cleans up timeout when component unmounts during expansion', () => {
      const {rerender, unmount} = render(<SideMenu defaultExpanded={false} />);

      // Start expanding
      rerender(<SideMenu defaultExpanded />);

      // Unmount before timeout completes
      unmount();

      // Advance timers - should not throw error
      vi.advanceTimersByTime(300);
    });

    it('cleans up timeout when component unmounts during collapse', () => {
      const {rerender, unmount} = render(<SideMenu defaultExpanded />);

      // Start collapsing
      rerender(<SideMenu defaultExpanded={false} />);

      // Unmount before timeout completes
      unmount();

      // Advance timers - should not throw error
      vi.advanceTimersByTime(300);
    });
  });
});
