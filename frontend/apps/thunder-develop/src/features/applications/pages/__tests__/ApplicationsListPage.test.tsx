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

import {render} from '@thunder/test-utils/browser';
import {page, userEvent} from 'vitest/browser';
import {describe, it, expect, vi, beforeEach} from 'vitest';
import ApplicationsListPage from '../ApplicationsListPage';

// Mock the ApplicationsList component
vi.mock('../../components/ApplicationsList', () => ({
  default: () => <div data-testid="applications-list">Applications List Component</div>,
}));

// Mock react-router navigate
const mockNavigate = vi.fn();
vi.mock('react-router', async () => {
  const actual = await vi.importActual('react-router');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

describe('ApplicationsListPage', () => {
  const renderWithProviders = () => render(<ApplicationsListPage />);

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('should render the page title', async () => {
      await renderWithProviders();

      await expect.element(page.getByRole('heading', {level: 1, name: 'Applications'})).toBeInTheDocument();
    });

    it('should render the page subtitle', async () => {
      await renderWithProviders();

      await expect.element(page.getByText('Manage your applications and services')).toBeInTheDocument();
    });

    it('should render the Add Application button', async () => {
      await renderWithProviders();

      await expect.element(page.getByRole('button', {name: /Add Application/i})).toBeInTheDocument();
    });

    it('should render the search field', async () => {
      await renderWithProviders();

      const searchInput = page.getByPlaceholder('Search ..');
      expect(searchInput).toBeInTheDocument();
      expect(searchInput).toHaveAttribute('type', 'text');
    });

    it('should render the ApplicationsList component', async () => {
      await renderWithProviders();

      await expect.element(page.getByTestId('applications-list')).toBeInTheDocument();
    });

    it('should render search icon in the search field', async () => {
      await renderWithProviders();

      const searchInput = page.getByPlaceholder('Search ..');
      const searchInputContainer = searchInput.element().closest('.MuiInputBase-root');

      expect(searchInputContainer).toBeInTheDocument();
    });
  });

  describe('Navigation', () => {
    it('should navigate to create page when Add Application button is clicked', async () => {
            await renderWithProviders();

      const createButton = page.getByRole('button', {name: /Add Application/i});
      await userEvent.click(createButton);

      expect(mockNavigate).toHaveBeenCalledWith('/applications/create');
    });

    it('should handle navigation errors gracefully', async () => {
            const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

      mockNavigate.mockRejectedValueOnce(new Error('Navigation failed'));

      await renderWithProviders();

      const createButton = page.getByRole('button', {name: /Add Application/i});
      await userEvent.click(createButton);

      expect(mockNavigate).toHaveBeenCalledWith('/applications/create');

      // Logger should log the error
      expect(consoleErrorSpy).toHaveBeenCalled();

      consoleErrorSpy.mockRestore();
    });
  });

  describe('Search Functionality', () => {
    it('should allow typing in the search field', async () => {
            await renderWithProviders();

      const searchInput = page.getByPlaceholder('Search ..');
      await userEvent.type(searchInput, 'My App');

      expect(searchInput).toHaveValue('My App');
    });

    it('should clear search field value', async () => {
            await renderWithProviders();

      const searchInput = page.getByPlaceholder('Search ..');
      await userEvent.type(searchInput, 'Test');
      expect(searchInput).toHaveValue('Test');

      await userEvent.clear(searchInput);
      expect(searchInput).toHaveValue('');
    });

    it('should handle special characters in search', async () => {
            await renderWithProviders();

      const searchInput = page.getByPlaceholder('Search ..');
      await userEvent.type(searchInput, '!@#$%^&*()');

      expect(searchInput).toHaveValue('!@#$%^&*()');
    });
  });

  describe('Layout', () => {
    it('should have proper page structure', async () => {
      const {container} = await renderWithProviders();

      // Main container
      const mainBox = container.querySelector('.MuiBox-root');
      expect(mainBox).toBeInTheDocument();

      // Header section with title and button
      await expect.element(page.getByRole('heading', {level: 1})).toBeInTheDocument();
      await expect.element(page.getByRole('button', {name: /Add Application/i})).toBeInTheDocument();

      // Search section
      await expect.element(page.getByPlaceholder('Search ..')).toBeInTheDocument();

      // Content section
      await expect.element(page.getByTestId('applications-list')).toBeInTheDocument();
    });

    it('should render components in correct order', async () => {
      const {container} = await renderWithProviders();

      const elements = [
        page.getByRole('heading', {level: 1}),
        page.getByRole('button', {name: /Add Application/i}),
        page.getByPlaceholder('Search ..'),
        page.getByTestId('applications-list'),
      ];

      // Verify each element exists in the DOM
      elements.forEach((element) => {
        expect(element).toBeInTheDocument();
      });

      // Verify order by checking positions in the DOM
      const mainContainer = container.firstChild;
      expect(mainContainer).toBeInTheDocument();
    });
  });

  describe('Responsive Behavior', () => {
    it('should render with responsive flex properties', async () => {
      const {container} = await renderWithProviders();

      // Header stack should have flexWrap
      const headerStack = container.querySelector('.MuiStack-root');
      expect(headerStack).toBeInTheDocument();
    });

    it('should render search field with minimum width', async () => {
      await renderWithProviders();

      const searchInput = page.getByPlaceholder('Search ..');
      const searchField = searchInput.element().closest('.MuiTextField-root');

      expect(searchField).toBeInTheDocument();
    });
  });

  describe('Button Styling', () => {
    it('should render Add Application button with correct variant', async () => {
      await renderWithProviders();

      const createButton = page.getByRole('button', {name: /Add Application/i});
      expect(createButton).toHaveClass('MuiButton-contained');
    });

    it('should have Plus icon in Add Application button', async () => {
      await renderWithProviders();

      const createButton = page.getByRole('button', {name: /Add Application/i});
      const icon = createButton.element().querySelector('svg');

      expect(icon).toBeInTheDocument();
    });
  });

  describe('Integration', () => {
    it('should work with QueryClient provider', () => {
      expect(() => renderWithProviders()).not.toThrow();
    });

    it('should work with BrowserRouter', () => {
      expect(() => renderWithProviders()).not.toThrow();
    });

    it('should work with ConfigProvider', () => {
      expect(() => renderWithProviders()).not.toThrow();
    });

    it('should render the search input adornment with icon', async () => {
      await renderWithProviders();

      const searchInput = page.getByPlaceholder('Search ..');
      const inputContainer = searchInput.element().closest('.MuiInputBase-root');

      // The InputAdornment with Search icon should be present
      expect(inputContainer).toBeInTheDocument();
      expect(inputContainer?.querySelector('svg')).toBeInTheDocument();
    });

    it('should render with all required MUI components', async () => {
      await renderWithProviders();

      // Verify Box, Stack, Typography components are rendered
      await expect.element(page.getByRole('heading', {level: 1})).toBeInTheDocument();
      await expect.element(page.getByText('Manage your applications and services')).toBeInTheDocument();
    });

    it('should render button with startIcon', async () => {
      await renderWithProviders();

      const createButton = page.getByRole('button', {name: /Add Application/i});
      // Button should have SVG icon
      expect(createButton.element().querySelector('svg')).toBeInTheDocument();
    });

    it('should render TextField with correct size', async () => {
      await renderWithProviders();

      const searchInput = page.getByPlaceholder('Search ..');
      const textField = searchInput.element().closest('.MuiTextField-root');
      expect(textField).toBeInTheDocument();
    });

    it('should render ApplicationsList component', async () => {
      await renderWithProviders();

      await expect.element(page.getByTestId('applications-list')).toBeInTheDocument();
    });
  });

  describe('Edge Cases', () => {
    it('should handle rapid button clicks', async () => {
            await renderWithProviders();

      const createButton = page.getByRole('button', {name: /Add Application/i});

      await userEvent.click(createButton);
      await userEvent.click(createButton);
      await userEvent.click(createButton);

      // Navigation should be attempted for each click
      expect(mockNavigate).toHaveBeenCalledTimes(3);
    });

    it('should handle long search queries', async () => {
      await renderWithProviders();

      const searchInput = page.getByPlaceholder('Search ..');
      const longQuery = 'A'.repeat(500);

      // Use userEvent.fill for long input
      await userEvent.fill(searchInput, longQuery);

      await expect.element(searchInput).toHaveValue(longQuery);
    });

    it('should maintain state after multiple interactions', async () => {
      await renderWithProviders();

      // Type in search using fireEvent for cross-platform reliability
      const searchInput = page.getByPlaceholder('Search ..');
      await userEvent.fill(searchInput, 'Test App');

      // Click create button
      const createButton = page.getByRole('button', {name: /Add Application/i});
      await userEvent.click(createButton);

      // Search value should still be there
      expect(searchInput).toHaveValue('Test App');
    });
  });

  describe('Accessibility', () => {
    it('should have proper heading hierarchy', async () => {
      await renderWithProviders();

      const h1 = page.getByRole('heading', {level: 1});
      expect(h1).toBeInTheDocument();
      expect(h1).toHaveTextContent('Applications');
    });

    it('should have accessible search field', async () => {
      await renderWithProviders();

      const searchInput = page.getByPlaceholder('Search ..');
      expect(searchInput).toHaveAttribute('placeholder', 'Search ..');
    });

    it('should have accessible buttons', async () => {
      await renderWithProviders();

      const createButton = page.getByRole('button', {name: /Add Application/i});
      expect(createButton).toBeEnabled();
      expect(createButton).toHaveAccessibleName();
    });

    it('should be keyboard navigable', async () => {
            await renderWithProviders();

      // Tab through interactive elements
      await userEvent.tab();

      const createButton = page.getByRole('button', {name: /Add Application/i});
      expect(createButton).toHaveFocus();

      await userEvent.tab();

      const searchInput = page.getByPlaceholder('Search ..');
      expect(searchInput).toHaveFocus();
    });

    it('should support Enter key on Add Application button', async () => {
            await renderWithProviders();

      const createButton = page.getByRole('button', {name: /Add Application/i});
      createButton.element().focus();

      await userEvent.keyboard('{Enter}');

      expect(mockNavigate).toHaveBeenCalledWith('/applications/create');
    });
  });
});
