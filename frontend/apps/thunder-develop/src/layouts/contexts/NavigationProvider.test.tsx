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
import {screen, renderHook, waitFor, render} from '@testing-library/react';
import {act} from 'react';
import {MemoryRouter} from 'react-router';
import NavigationProvider from './NavigationProvider';
import useNavigation from './useNavigation';

function TestComponent() {
  const {currentPage, sidebarOpen, toggleSidebar, setSidebarOpen, setCurrentPage} = useNavigation();

  return (
    <div>
      <div data-testid="current-page">{currentPage}</div>
      <div data-testid="sidebar-open">{String(sidebarOpen)}</div>
      <button type="button" onClick={toggleSidebar}>
        Toggle Sidebar
      </button>
      <button type="button" onClick={() => setSidebarOpen(false)}>
        Close Sidebar
      </button>
      <button type="button" onClick={() => setCurrentPage('custom')}>
        Set Custom Page
      </button>
    </div>
  );
}

describe('NavigationProvider', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('provides navigation context to children', () => {
    render(
      <MemoryRouter>
        <NavigationProvider>
          <TestComponent />
        </NavigationProvider>
      </MemoryRouter>,
    );

    expect(screen.getByTestId('current-page')).toBeInTheDocument();
    expect(screen.getByTestId('sidebar-open')).toBeInTheDocument();
  });

  it('sets current page to home by default', () => {
    render(
      <MemoryRouter initialEntries={['/']}>
        <NavigationProvider>
          <TestComponent />
        </NavigationProvider>
      </MemoryRouter>,
    );

    expect(screen.getByTestId('current-page')).toHaveTextContent('home');
  });

  it('sets sidebar open by default', () => {
    render(
      <MemoryRouter>
        <NavigationProvider>
          <TestComponent />
        </NavigationProvider>
      </MemoryRouter>,
    );

    expect(screen.getByTestId('sidebar-open')).toHaveTextContent('true');
  });

  it('updates current page based on route', () => {
    render(
      <MemoryRouter initialEntries={['/users']}>
        <NavigationProvider>
          <TestComponent />
        </NavigationProvider>
      </MemoryRouter>,
    );

    expect(screen.getByTestId('current-page')).toHaveTextContent('users');
  });

  it('extracts first path segment as current page', () => {
    render(
      <MemoryRouter initialEntries={['/users/123']}>
        <NavigationProvider>
          <TestComponent />
        </NavigationProvider>
      </MemoryRouter>,
    );

    expect(screen.getByTestId('current-page')).toHaveTextContent('users');
  });

  it('toggles sidebar when toggleSidebar is called', async () => {
    render(
      <MemoryRouter>
        <NavigationProvider>
          <TestComponent />
        </NavigationProvider>
      </MemoryRouter>,
    );

    expect(screen.getByTestId('sidebar-open')).toHaveTextContent('true');

    const toggleButton = screen.getByText('Toggle Sidebar');
    act(() => {
      toggleButton.click();
    });

    await waitFor(() => {
      expect(screen.getByTestId('sidebar-open')).toHaveTextContent('false');
    });

    act(() => {
      toggleButton.click();
    });

    await waitFor(() => {
      expect(screen.getByTestId('sidebar-open')).toHaveTextContent('true');
    });
  });

  it('closes sidebar when setSidebarOpen(false) is called', async () => {
    render(
      <MemoryRouter>
        <NavigationProvider>
          <TestComponent />
        </NavigationProvider>
      </MemoryRouter>,
    );

    const closeButton = screen.getByText('Close Sidebar');
    act(() => {
      closeButton.click();
    });

    await waitFor(() => {
      expect(screen.getByTestId('sidebar-open')).toHaveTextContent('false');
    });
  });

  it('updates current page when setCurrentPage is called', async () => {
    render(
      <MemoryRouter>
        <NavigationProvider>
          <TestComponent />
        </NavigationProvider>
      </MemoryRouter>,
    );

    const button = screen.getByText('Set Custom Page');
    act(() => {
      button.click();
    });

    await waitFor(() => {
      expect(screen.getByTestId('current-page')).toHaveTextContent('custom');
    });
  });

  it('memoizes context value', () => {
    const {rerender} = render(
      <MemoryRouter>
        <NavigationProvider>
          <TestComponent />
        </NavigationProvider>
      </MemoryRouter>,
    );

    const initialValue = screen.getByTestId('current-page').textContent;

    rerender(
      <MemoryRouter>
        <NavigationProvider>
          <TestComponent />
        </NavigationProvider>
      </MemoryRouter>,
    );

    expect(screen.getByTestId('current-page')).toHaveTextContent(initialValue);
  });
});

describe('useNavigation', () => {
  it('throws error when used outside NavigationProvider', () => {
    const TestComponentWithoutProvider = () => {
      try {
        useNavigation();
        return <div>Should not render</div>;
      } catch (error) {
        return <div data-testid="error">{(error as Error).message}</div>;
      }
    };

    render(
      <MemoryRouter>
        <TestComponentWithoutProvider />
      </MemoryRouter>,
    );

    expect(screen.getByTestId('error')).toHaveTextContent('useNavigation must be used within a NavigationProvider');
  });

  it('returns context when used inside NavigationProvider', () => {
    const {result} = renderHook(() => useNavigation(), {
      wrapper: ({children}) => (
        <MemoryRouter>
          <NavigationProvider>{children}</NavigationProvider>
        </MemoryRouter>
      ),
    });

    expect(result.current).toBeDefined();
    expect(result.current.currentPage).toBe('home');
    expect(result.current.sidebarOpen).toBe(true);
    expect(typeof result.current.toggleSidebar).toBe('function');
    expect(typeof result.current.setSidebarOpen).toBe('function');
    expect(typeof result.current.setCurrentPage).toBe('function');
  });
});
