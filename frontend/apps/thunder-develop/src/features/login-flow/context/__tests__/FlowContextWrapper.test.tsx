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
import FlowContextWrapper from '../FlowContextWrapper';
import LoginFlowBuilderContext from '../LoginFlowBuilderContext';

// Test consumer component
function TestConsumer() {
  const context = useContext(LoginFlowBuilderContext);
  return (
    <div data-testid="context-consumer" data-context-value={context === null ? 'null' : 'object'}>
      Context Consumer
    </div>
  );
}

describe('FlowContextWrapper', () => {
  describe('Provider Setup', () => {
    it('should render children', () => {
      render(
        <FlowContextWrapper>
          <div data-testid="child">Child Content</div>
        </FlowContextWrapper>,
      );

      expect(screen.getByTestId('child')).toHaveTextContent('Child Content');
    });

    it('should provide LoginFlowBuilderContext', () => {
      render(
        <FlowContextWrapper>
          <TestConsumer />
        </FlowContextWrapper>,
      );

      expect(screen.getByTestId('context-consumer')).toBeInTheDocument();
    });

    it('should provide null context value', () => {
      render(
        <FlowContextWrapper>
          <TestConsumer />
        </FlowContextWrapper>,
      );

      expect(screen.getByTestId('context-consumer')).toHaveAttribute('data-context-value', 'null');
    });
  });

  describe('Children Rendering', () => {
    it('should render single child', () => {
      render(
        <FlowContextWrapper>
          <span data-testid="single-child">Single Child</span>
        </FlowContextWrapper>,
      );

      expect(screen.getByTestId('single-child')).toBeInTheDocument();
    });

    it('should render multiple children', () => {
      render(
        <FlowContextWrapper>
          <div data-testid="child-1">First</div>
          <div data-testid="child-2">Second</div>
          <div data-testid="child-3">Third</div>
        </FlowContextWrapper>,
      );

      expect(screen.getByTestId('child-1')).toBeInTheDocument();
      expect(screen.getByTestId('child-2')).toBeInTheDocument();
      expect(screen.getByTestId('child-3')).toBeInTheDocument();
    });

    it('should render nested components', () => {
      render(
        <FlowContextWrapper>
          <div data-testid="parent">
            <div data-testid="nested-child">Nested Content</div>
          </div>
        </FlowContextWrapper>,
      );

      const parent = screen.getByTestId('parent');
      const nestedChild = screen.getByTestId('nested-child');

      expect(parent).toContainElement(nestedChild);
    });
  });
});
