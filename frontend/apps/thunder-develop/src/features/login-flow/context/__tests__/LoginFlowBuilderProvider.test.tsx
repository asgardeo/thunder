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

/* eslint-disable @typescript-eslint/no-unsafe-assignment, @typescript-eslint/prefer-nullish-coalescing, @typescript-eslint/no-unsafe-member-access */

import {describe, it, expect, vi} from 'vitest';
import {render} from '@thunder/test-utils/browser';
import {page} from 'vitest/browser';
import LoginFlowBuilderProvider from '../LoginFlowBuilderProvider';
import {PreviewScreenType} from '../../../flows/models/custom-text-preference';

// Mock FlowBuilderCoreProvider
vi.mock('@/features/flows/context/FlowBuilderCoreProvider', () => ({
  default: ({children, screenTypes}: {children: React.ReactNode; screenTypes: PreviewScreenType[]}) => (
    <div data-testid="flow-builder-core-provider" data-screen-types={JSON.stringify(screenTypes)}>
      {children}
    </div>
  ),
}));

// Mock FlowContextWrapper
vi.mock('../FlowContextWrapper', () => ({
  default: ({children}: {children: React.ReactNode}) => (
    <div data-testid="flow-context-wrapper">{children}</div>
  ),
}));

// Mock ResourceProperties
vi.mock('../../components/resource-property-panel/ResourceProperties', () => ({
  default: () => <div>Resource Properties</div>,
}));

// Mock ElementFactory
vi.mock('../../components/resources/elements/ElementFactory', () => ({
  default: () => <div>Element Factory</div>,
}));

describe('LoginFlowBuilderProvider', () => {
  describe('Component Structure', () => {
    it('should render FlowBuilderCoreProvider', async () => {
      await render(
        <LoginFlowBuilderProvider>
          <div data-testid="child">Child Content</div>
        </LoginFlowBuilderProvider>,
      );

      await expect.element(page.getByTestId('flow-builder-core-provider')).toBeInTheDocument();
    });

    it('should render FlowContextWrapper inside FlowBuilderCoreProvider', async () => {
      await render(
        <LoginFlowBuilderProvider>
          <div data-testid="child">Child Content</div>
        </LoginFlowBuilderProvider>,
      );

      const coreProvider = page.getByTestId('flow-builder-core-provider');
      const contextWrapper = page.getByTestId('flow-context-wrapper');

      expect(coreProvider.element()).toContainElement(contextWrapper.element());
    });

    it('should render children inside FlowContextWrapper', async () => {
      await render(
        <LoginFlowBuilderProvider>
          <div data-testid="child">Child Content</div>
        </LoginFlowBuilderProvider>,
      );

      const contextWrapper = page.getByTestId('flow-context-wrapper');
      const child = page.getByTestId('child');

      expect(contextWrapper.element()).toContainElement(child.element());
    });
  });

  describe('Screen Types Configuration', () => {
    it('should pass correct screen types to FlowBuilderCoreProvider', async () => {
      await render(
        <LoginFlowBuilderProvider>
          <div>Content</div>
        </LoginFlowBuilderProvider>,
      );

      const coreProvider = page.getByTestId('flow-builder-core-provider');
      const screenTypes = JSON.parse(coreProvider.element().getAttribute('data-screen-types') || '[]');

      expect(screenTypes).toContain(PreviewScreenType.SIGN_UP);
      expect(screenTypes).toContain(PreviewScreenType.COMMON);
      expect(screenTypes).toContain(PreviewScreenType.EMAIL_LINK_EXPIRY);
      expect(screenTypes).toContain(PreviewScreenType.SMS_OTP);
      expect(screenTypes).toContain(PreviewScreenType.EMAIL_OTP);
    });

    it('should have 5 screen types configured', async () => {
      await render(
        <LoginFlowBuilderProvider>
          <div>Content</div>
        </LoginFlowBuilderProvider>,
      );

      const coreProvider = page.getByTestId('flow-builder-core-provider');
      const screenTypes = JSON.parse(coreProvider.element().getAttribute('data-screen-types') || '[]');

      expect(screenTypes).toHaveLength(5);
    });

    it('should have SIGN_UP as the first screen type (primary)', async () => {
      await render(
        <LoginFlowBuilderProvider>
          <div>Content</div>
        </LoginFlowBuilderProvider>,
      );

      const coreProvider = page.getByTestId('flow-builder-core-provider');
      const screenTypes = JSON.parse(coreProvider.element().getAttribute('data-screen-types') || '[]');

      expect(screenTypes[0]).toBe(PreviewScreenType.SIGN_UP);
    });
  });

  describe('Children Rendering', () => {
    it('should render children content', async () => {
      await render(
        <LoginFlowBuilderProvider>
          <div data-testid="child">Child Content</div>
        </LoginFlowBuilderProvider>,
      );

      await expect.element(page.getByTestId('child')).toHaveTextContent('Child Content');
    });

    it('should render multiple children', async () => {
      await render(
        <LoginFlowBuilderProvider>
          <div data-testid="child-1">First Child</div>
          <div data-testid="child-2">Second Child</div>
        </LoginFlowBuilderProvider>,
      );

      await expect.element(page.getByTestId('child-1')).toBeInTheDocument();
      await expect.element(page.getByTestId('child-2')).toBeInTheDocument();
    });
  });
});
