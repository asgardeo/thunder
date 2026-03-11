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
import {render} from '@thunder/test-utils/browser';
import {page} from 'vitest/browser';
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
    it('should render the LoginFlowBuilderProvider', async () => {
      await render(<LoginFlowBuilderPage />);

      await expect.element(page.getByTestId('login-flow-builder-provider')).toBeInTheDocument();
    });

    it('should render the ReactFlowProvider inside LoginFlowBuilderProvider', async () => {
      await render(<LoginFlowBuilderPage />);

      const provider = page.getByTestId('login-flow-builder-provider');
      const reactFlowProvider = page.getByTestId('react-flow-provider');

      expect(provider.element()).toContainElement(reactFlowProvider.element());
    });

    it('should render the LoginFlowBuilder inside ReactFlowProvider', async () => {
      await render(<LoginFlowBuilderPage />);

      const reactFlowProvider = page.getByTestId('react-flow-provider');
      const loginFlowBuilder = page.getByTestId('login-flow-builder');

      expect(reactFlowProvider.element()).toContainElement(loginFlowBuilder.element());
    });

    it('should render components in correct nesting order', async () => {
      await render(<LoginFlowBuilderPage />);

      const provider = page.getByTestId('login-flow-builder-provider');
      const reactFlowProvider = page.getByTestId('react-flow-provider');
      const loginFlowBuilder = page.getByTestId('login-flow-builder');

      // Verify proper nesting: LoginFlowBuilderProvider > ReactFlowProvider > LoginFlowBuilder
      expect(provider.element()).toContainElement(reactFlowProvider.element());
      expect(reactFlowProvider.element()).toContainElement(loginFlowBuilder.element());
    });
  });

  describe('Component Integration', () => {
    it('should render LoginFlowBuilder content', async () => {
      await render(<LoginFlowBuilderPage />);

      await expect.element(page.getByText('Login Flow Builder')).toBeInTheDocument();
    });
  });

  describe('Page Structure', () => {
    it('should have all required provider wrappers', async () => {
      await render(<LoginFlowBuilderPage />);

      await expect.element(page.getByTestId('login-flow-builder-provider')).toBeInTheDocument();
      await expect.element(page.getByTestId('react-flow-provider')).toBeInTheDocument();
      await expect.element(page.getByTestId('login-flow-builder')).toBeInTheDocument();
    });

    it('should render without crashing', () => {
      expect(() => render(<LoginFlowBuilderPage />)).not.toThrow();
    });
  });
});
