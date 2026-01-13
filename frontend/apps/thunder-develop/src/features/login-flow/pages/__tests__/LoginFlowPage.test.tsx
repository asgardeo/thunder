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
import {render, screen} from '@testing-library/react';
import LoginFlowBuilderPage from '../LoginFlowPage';

// Mock ReactFlowProvider
vi.mock('@xyflow/react', () => ({
  ReactFlowProvider: ({children}: {children: React.ReactNode}) => (
    <div data-testid="react-flow-provider">{children}</div>
  ),
}));

// Mock LoginFlowBuilder
vi.mock('../../components/LoginFlowBuilder', () => ({
  default: () => <div data-testid="login-flow-builder">Login Flow Builder</div>,
}));

// Mock LoginFlowBuilderProvider
vi.mock('../../context/LoginFlowBuilderProvider', () => ({
  default: ({children}: {children: React.ReactNode}) => (
    <div data-testid="login-flow-builder-provider">{children}</div>
  ),
}));

describe('LoginFlowBuilderPage', () => {
  describe('Rendering', () => {
    it('should render the LoginFlowBuilderProvider', () => {
      render(<LoginFlowBuilderPage />);

      expect(screen.getByTestId('login-flow-builder-provider')).toBeInTheDocument();
    });

    it('should render the ReactFlowProvider inside LoginFlowBuilderProvider', () => {
      render(<LoginFlowBuilderPage />);

      const provider = screen.getByTestId('login-flow-builder-provider');
      const reactFlowProvider = screen.getByTestId('react-flow-provider');

      expect(provider).toContainElement(reactFlowProvider);
    });

    it('should render the LoginFlowBuilder inside ReactFlowProvider', () => {
      render(<LoginFlowBuilderPage />);

      const reactFlowProvider = screen.getByTestId('react-flow-provider');
      const loginFlowBuilder = screen.getByTestId('login-flow-builder');

      expect(reactFlowProvider).toContainElement(loginFlowBuilder);
    });

    it('should render components in correct nesting order', () => {
      render(<LoginFlowBuilderPage />);

      const provider = screen.getByTestId('login-flow-builder-provider');
      const reactFlowProvider = screen.getByTestId('react-flow-provider');
      const loginFlowBuilder = screen.getByTestId('login-flow-builder');

      // Verify proper nesting: LoginFlowBuilderProvider > ReactFlowProvider > LoginFlowBuilder
      expect(provider).toContainElement(reactFlowProvider);
      expect(reactFlowProvider).toContainElement(loginFlowBuilder);
    });
  });

  describe('Component Integration', () => {
    it('should render LoginFlowBuilder content', () => {
      render(<LoginFlowBuilderPage />);

      expect(screen.getByText('Login Flow Builder')).toBeInTheDocument();
    });
  });

  describe('Page Structure', () => {
    it('should have all required provider wrappers', () => {
      render(<LoginFlowBuilderPage />);

      expect(screen.getByTestId('login-flow-builder-provider')).toBeInTheDocument();
      expect(screen.getByTestId('react-flow-provider')).toBeInTheDocument();
      expect(screen.getByTestId('login-flow-builder')).toBeInTheDocument();
    });

    it('should render without crashing', () => {
      expect(() => render(<LoginFlowBuilderPage />)).not.toThrow();
    });
  });
});
