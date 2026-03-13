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
import {render, screen} from '@testing-library/react';
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
    it('should provide null as default value when used outside provider', () => {
      render(<TestConsumer />);

      expect(screen.getByTestId('context-consumer')).toHaveAttribute('data-is-null', 'true');
    });

    it('should display null context value text', () => {
      render(<TestConsumer />);

      expect(screen.getByTestId('context-consumer')).toHaveTextContent('Context Value: null');
    });
  });

  describe('Provider Usage', () => {
    it('should allow providing custom value through Provider', () => {
      const customValue = {};

      render(
        <LoginFlowBuilderContext.Provider value={customValue}>
          <TestConsumer />
        </LoginFlowBuilderContext.Provider>,
      );

      expect(screen.getByTestId('context-consumer')).toHaveAttribute('data-is-null', 'false');
    });

    it('should allow providing null through Provider', () => {
      render(
        <LoginFlowBuilderContext.Provider value={null}>
          <TestConsumer />
        </LoginFlowBuilderContext.Provider>,
      );

      expect(screen.getByTestId('context-consumer')).toHaveAttribute('data-is-null', 'true');
    });
  });

  describe('Nested Providers', () => {
    it('should use closest provider value', () => {
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

      render(
        <LoginFlowBuilderContext.Provider value={outerValue}>
          <LoginFlowBuilderContext.Provider value={innerValue}>
            <InnerConsumer />
          </LoginFlowBuilderContext.Provider>
        </LoginFlowBuilderContext.Provider>,
      );

      expect(screen.getByTestId('inner-consumer')).toHaveAttribute('data-is-null', 'true');
    });
  });
});
