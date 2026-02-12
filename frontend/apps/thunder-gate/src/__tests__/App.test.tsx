/**
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
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
import {render} from '@testing-library/react';
import App from '../App';

// Create a variable to hold the mock routes that we can change per test
let mockAppRoutes: {
  path: string;
  element: React.ReactNode;
  children?: {
    path?: string;
    index?: boolean;
    element: React.ReactNode;
  }[];
}[] = [];

// Mock the app routes module
vi.mock('../config/appRoutes', () => ({
  get default() {
    return mockAppRoutes;
  },
}));

describe('App', () => {
  beforeEach(() => {
    // Reset routes before each test
    mockAppRoutes = [];
  });

  it('renders without crashing with empty routes', () => {
    mockAppRoutes = [];
    const {container} = render(<App />);
    expect(container).toBeInTheDocument();
  });

  it('renders routes with children that have paths', () => {
    mockAppRoutes = [
      {
        path: '/parent',
        element: <div data-testid="parent">Parent</div>,
        children: [
          {
            path: 'child1',
            element: <div data-testid="child1">Child 1</div>,
          },
          {
            path: 'child2',
            element: <div data-testid="child2">Child 2</div>,
          },
        ],
      },
    ];
    const {container} = render(<App />);
    expect(container).toBeInTheDocument();
  });

  it('renders routes with index child (no path) without crashing', () => {
    // This test ensures that routes with index children (no path property)
    // render without runtime errors
    mockAppRoutes = [
      {
        path: '/dashboard',
        element: <div data-testid="dashboard">Dashboard</div>,
        children: [
          {
            index: true, // No path, uses index
            element: <div data-testid="index-child">Index Child</div>,
          },
          {
            path: 'settings',
            element: <div data-testid="settings">Settings</div>,
          },
        ],
      },
    ];
    const {container} = render(<App />);
    expect(container).toBeInTheDocument();
  });

  it('renders routes without children', () => {
    mockAppRoutes = [
      {
        path: '/standalone',
        element: <div data-testid="standalone">Standalone Route</div>,
      },
    ];
    const {container} = render(<App />);
    expect(container).toBeInTheDocument();
  });

  it('renders multiple parent routes with mixed children', () => {
    mockAppRoutes = [
      {
        path: '/auth',
        element: <div data-testid="auth">Auth</div>,
        children: [
          {
            index: true, // Index route without path
            element: <div>Auth Index</div>,
          },
        ],
      },
      {
        path: '/app',
        element: <div data-testid="app">App</div>,
        children: [
          {
            path: 'home',
            element: <div>Home</div>,
          },
          {
            index: true, // Another index route without path
            element: <div>App Index</div>,
          },
        ],
      },
    ];
    const {container} = render(<App />);
    expect(container).toBeInTheDocument();
  });

  it('renders route with empty children array', () => {
    mockAppRoutes = [
      {
        path: '/empty-children',
        element: <div data-testid="empty-children">Empty Children</div>,
        children: [],
      },
    ];
    const {container} = render(<App />);
    expect(container).toBeInTheDocument();
  });
});
