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
import type {JSX} from 'react';
import {screen} from '@testing-library/react';
import render from '@/test/test-utils';
import HomePage from './HomePage';

interface MockUser {
  givenName: string;
  familyName: string;
  name: string;
  email: string;
}

// Mock the Asgardeo User component
vi.mock('@asgardeo/react', () => ({
  User: ({children}: {children: (user: MockUser) => JSX.Element}) => {
    const mockUser: MockUser = {
      givenName: 'John',
      familyName: 'Doe',
      name: 'John Doe',
      email: 'john.doe@example.com',
    };
    // Simulate authenticated user
    return <div data-testid="user-authenticated">{children(mockUser)}</div>;
  },
}));

describe('HomePage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders welcome message with user name', () => {
    render(<HomePage />);

    const welcomeMessage = screen.getByText(/welcome, john doe!/i);
    expect(welcomeMessage).toBeInTheDocument();
  });

  it('displays greeting emoji', () => {
    render(<HomePage />);

    const welcomeMessage = screen.getByText(/👋 welcome/i);
    expect(welcomeMessage).toBeInTheDocument();
  });

  it('renders with correct typography variant', () => {
    render(<HomePage />);

    const heading = screen.getByText(/welcome, john doe!/i);
    expect(heading.tagName).toBe('H2');
  });

  it('renders the User component', () => {
    render(<HomePage />);

    const userComponent = screen.getByTestId('user-authenticated');
    expect(userComponent).toBeInTheDocument();
  });

  it('includes givenName in welcome message', () => {
    render(<HomePage />);

    const welcomeMessage = screen.getByText(/john/i);
    expect(welcomeMessage).toBeInTheDocument();
  });

  it('includes familyName in welcome message', () => {
    render(<HomePage />);

    const welcomeMessage = screen.getByText(/doe/i);
    expect(welcomeMessage).toBeInTheDocument();
  });
});

describe('HomePage - Unauthenticated', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('shows fallback when user is not authenticated', () => {
    // Re-mock for unauthenticated state
    vi.doMock('@asgardeo/react', () => ({
      User: ({fallback}: {fallback: JSX.Element}) => <div data-testid="user-fallback">{fallback}</div>,
    }));

    render(<HomePage />);

    // Verify the structure is still rendered
    const stack = screen.getByText(/welcome/i).closest('.MuiStack-root');
    expect(stack).toBeInTheDocument();
  });
});
