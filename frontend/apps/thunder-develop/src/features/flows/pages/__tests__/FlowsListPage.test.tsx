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
import {render} from '@thunder/test-utils/browser';
import {page, userEvent} from 'vitest/browser';
import FlowsListPage from '../FlowsListPage';

// Mock @thunder/logger/react
const mockLoggerError = vi.fn();

vi.mock('@thunder/logger/react', () => ({
  useLogger: () => ({
    debug: vi.fn(),
    info: vi.fn(),
    warn: vi.fn(),
    error: mockLoggerError,
  }),
}));

// Mock useNavigate
const mockNavigate = vi.fn();
vi.mock('react-router', async () => {
  const actual = await vi.importActual('react-router');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

// Mock FlowsList component
vi.mock('../../components/FlowsList', () => ({
  default: () => <div data-testid="flows-list">FlowsList Component</div>,
}));

describe('FlowsListPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('should render the page title', async () => {
      await render(<FlowsListPage />);

      // Use heading role to uniquely find the title
      await expect.element(page.getByRole('heading', {name: 'Flows'})).toBeInTheDocument();
    });

    it('should render the page subtitle', async () => {
      await render(<FlowsListPage />);

      // Actual translation: 'Create and manage authentication and registration flows for your applications'
      await expect.element(
        page.getByText('Create and manage authentication and registration flows for your applications'),
      ).toBeInTheDocument();
    });

    it('should render the Create New Flow button', async () => {
      await render(<FlowsListPage />);

      // Actual translation: 'Create New Flow'
      await expect.element(page.getByRole('button', {name: /create new flow/i})).toBeInTheDocument();
    });

    it('should render FlowsList component', async () => {
      await render(<FlowsListPage />);

      await expect.element(page.getByTestId('flows-list')).toBeInTheDocument();
    });
  });

  describe('Create New Flow Button', () => {
    it('should navigate to login-builder when Create New Flow is clicked', async () => {
      await render(<FlowsListPage />);

      const addButton = page.getByRole('button', {name: /create new flow/i});
      await userEvent.click(addButton);

      await vi.waitFor(() => {
        expect(mockNavigate).toHaveBeenCalledWith('/flows/signin');
      });
    });

    it('should render button with contained variant', async () => {
      await render(<FlowsListPage />);

      const addButton = page.getByRole('button', {name: /create new flow/i});
      expect(addButton).toHaveClass('MuiButton-contained');
    });
  });

  describe('Layout', () => {
    it('should render title as h1', async () => {
      await render(<FlowsListPage />);

      const title = page.getByRole('heading', {level: 1});
      expect(title).toHaveTextContent('Flows');
    });

    it('should have proper structure with header and list', async () => {
      const {container} = await render(<FlowsListPage />);

      // Check that the page has a box container
      expect(container.querySelector('.MuiBox-root')).toBeInTheDocument();
    });
  });

  describe('Navigation Error Handling', () => {
    it('should handle navigation errors when Create New Flow button is clicked', async () => {
      const navigationError = new Error('Navigation failed');
      mockNavigate.mockRejectedValueOnce(navigationError);

      await render(<FlowsListPage />);

      const addButton = page.getByRole('button', {name: /create new flow/i});
      await userEvent.click(addButton);

      await vi.waitFor(() => {
        expect(mockNavigate).toHaveBeenCalledWith('/flows/signin');
      });

      // Verify that the error was caught and logged
      await vi.waitFor(() => {
        expect(mockLoggerError).toHaveBeenCalledWith('Failed to navigate to flow builder page', {
          error: navigationError,
        });
      });

      // Component should still be rendered (no crash)
      expect(addButton).toBeInTheDocument();
    });
  });
});
