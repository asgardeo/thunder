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
import {StaticStepTypes} from '@/features/flows/models/steps';
import StaticStepFactory from '../StaticStepFactory';

// Mock CommonStaticStepFactory
vi.mock('@/features/flows/components/resources/steps/CommonStaticStepFactory', () => ({
  CommonStaticStepFactory: ({type}: {type: string}) => (
    <div data-testid="common-static-step-factory" data-type={type}>
      Common Static Step Factory
    </div>
  ),
}));

describe('StaticStepFactory', () => {
  const createNodeProps = (overrides: Record<string, unknown> = {}) => ({
    id: 'node-1',
    type: StaticStepTypes.Start,
    position: {x: 0, y: 0},
    data: {},
    ...overrides,
  });

  describe('Rendering', () => {
    it('should render CommonStaticStepFactory for Start type', () => {
      const props = createNodeProps({type: StaticStepTypes.Start});

      render(<StaticStepFactory {...props} />);

      expect(screen.getByTestId('common-static-step-factory')).toBeInTheDocument();
    });

    it('should pass type prop to CommonStaticStepFactory', () => {
      const props = createNodeProps({type: StaticStepTypes.Start});

      render(<StaticStepFactory {...props} />);

      expect(screen.getByTestId('common-static-step-factory')).toHaveAttribute('data-type', StaticStepTypes.Start);
    });

    it('should pass UserOnboard type to CommonStaticStepFactory', () => {
      const props = createNodeProps({type: StaticStepTypes.UserOnboard});

      render(<StaticStepFactory {...props} />);

      expect(screen.getByTestId('common-static-step-factory')).toHaveAttribute(
        'data-type',
        StaticStepTypes.UserOnboard,
      );
    });
  });

  describe('Props Forwarding', () => {
    it('should forward additional node props', () => {
      const props = createNodeProps({
        id: 'custom-node-id',
        position: {x: 100, y: 200},
        data: {customData: 'test'},
      });

      render(<StaticStepFactory {...props} />);

      expect(screen.getByTestId('common-static-step-factory')).toBeInTheDocument();
    });

    it('should handle different node IDs', () => {
      const props1 = createNodeProps({id: 'node-1', type: StaticStepTypes.Start});
      const props2 = createNodeProps({id: 'node-2', type: StaticStepTypes.Start});

      const {rerender} = render(<StaticStepFactory {...props1} />);

      expect(screen.getByTestId('common-static-step-factory')).toBeInTheDocument();

      rerender(<StaticStepFactory {...props2} />);

      expect(screen.getByTestId('common-static-step-factory')).toBeInTheDocument();
    });
  });
});
