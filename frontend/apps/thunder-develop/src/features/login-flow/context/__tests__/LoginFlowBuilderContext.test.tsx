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

import {describe, it, expect} from 'vitest';
import {render} from '@thunder/test-utils/browser';
import {page} from 'vitest/browser';
import {useContext} from 'react';
import LoginFlowBuilderContext from '../LoginFlowBuilderContext';

// Test consumer component
function TestConsumer() {
  const context = useContext(LoginFlowBuilderContext);
  return (
    <div
      data-testid="context-consumer"
      data-is-null={context === null ? 'true' : 'false'}
    >
      Context Value: {context === null ? 'null' : 'object'}
    </div>
  );
}

describe('LoginFlowBuilderContext', () => {
  describe('Context Creation', () => {
    it('should have displayName set', () => {
      expect(LoginFlowBuilderContext.displayName).toBe('LoginFlowBuilderContext');
    });
  });

  describe('Default Value', () => {
    it('should provide null as default value when used outside provider', async () => {
      await render(<TestConsumer />);

      await expect.element(page.getByTestId('context-consumer')).toHaveAttribute('data-is-null', 'true');
    });

    it('should display null context value text', async () => {
      await render(<TestConsumer />);

      await expect.element(page.getByTestId('context-consumer')).toHaveTextContent('Context Value: null');
    });
  });

  describe('Provider Usage', () => {
    it('should allow providing custom value through Provider', async () => {
      const customValue = {};

      await render(
        <LoginFlowBuilderContext.Provider value={customValue}>
          <TestConsumer />
        </LoginFlowBuilderContext.Provider>,
      );

      await expect.element(page.getByTestId('context-consumer')).toHaveAttribute('data-is-null', 'false');
    });

    it('should allow providing null through Provider', async () => {
      await render(
        <LoginFlowBuilderContext.Provider value={null}>
          <TestConsumer />
        </LoginFlowBuilderContext.Provider>,
      );

      await expect.element(page.getByTestId('context-consumer')).toHaveAttribute('data-is-null', 'true');
    });
  });

  describe('Nested Providers', () => {
    it('should use closest provider value', async () => {
      const outerValue = {};
      const innerValue = null;

      function InnerConsumer() {
        const context = useContext(LoginFlowBuilderContext);
        return (
          <div data-testid="inner-consumer" data-is-null={context === null ? 'true' : 'false'}>
            Inner
          </div>
        );
      }

      await render(
        <LoginFlowBuilderContext.Provider value={outerValue}>
          <LoginFlowBuilderContext.Provider value={innerValue}>
            <InnerConsumer />
          </LoginFlowBuilderContext.Provider>
        </LoginFlowBuilderContext.Provider>,
      );

      await expect.element(page.getByTestId('inner-consumer')).toHaveAttribute('data-is-null', 'true');
    });
  });
});
