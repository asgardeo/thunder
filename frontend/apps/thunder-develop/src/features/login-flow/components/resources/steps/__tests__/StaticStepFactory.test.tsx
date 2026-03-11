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
    it('should render CommonStaticStepFactory for Start type', async () => {
      const props = createNodeProps({type: StaticStepTypes.Start});

      await render(<StaticStepFactory {...props} />);

      await expect.element(page.getByTestId('common-static-step-factory')).toBeInTheDocument();
    });

    it('should pass type prop to CommonStaticStepFactory', async () => {
      const props = createNodeProps({type: StaticStepTypes.Start});

      await render(<StaticStepFactory {...props} />);

      await expect.element(page.getByTestId('common-static-step-factory')).toHaveAttribute('data-type', StaticStepTypes.Start);
    });

    it('should pass UserOnboard type to CommonStaticStepFactory', async () => {
      const props = createNodeProps({type: StaticStepTypes.UserOnboard});

      await render(<StaticStepFactory {...props} />);

      await expect.element(page.getByTestId('common-static-step-factory')).toHaveAttribute(
        'data-type',
        StaticStepTypes.UserOnboard,
      );
    });
  });

  describe('Props Forwarding', () => {
    it('should forward additional node props', async () => {
      const props = createNodeProps({
        id: 'custom-node-id',
        position: {x: 100, y: 200},
        data: {customData: 'test'},
      });

      await render(<StaticStepFactory {...props} />);

      await expect.element(page.getByTestId('common-static-step-factory')).toBeInTheDocument();
    });

    it('should handle different node IDs', async () => {
      const props1 = createNodeProps({id: 'node-1', type: StaticStepTypes.Start});
      const props2 = createNodeProps({id: 'node-2', type: StaticStepTypes.Start});

      const {rerender} = await render(<StaticStepFactory {...props1} />);

      await expect.element(page.getByTestId('common-static-step-factory')).toBeInTheDocument();

      await rerender(<StaticStepFactory {...props2} />);

      await expect.element(page.getByTestId('common-static-step-factory')).toBeInTheDocument();
    });
  });

  describe('Type Handling', () => {
    it('should handle undefined type gracefully', async () => {
      const props = createNodeProps({type: undefined});

      await render(<StaticStepFactory {...props} />);

      await expect.element(page.getByTestId('common-static-step-factory')).toBeInTheDocument();
    });

    it('should handle empty string type', async () => {
      const props = createNodeProps({type: ''});

      await render(<StaticStepFactory {...props} />);

      await expect.element(page.getByTestId('common-static-step-factory')).toHaveAttribute('data-type', '');
    });

    it('should handle custom type string', async () => {
      const props = createNodeProps({type: 'CUSTOM_TYPE'});

      await render(<StaticStepFactory {...props} />);

      await expect.element(page.getByTestId('common-static-step-factory')).toHaveAttribute('data-type', 'CUSTOM_TYPE');
    });
  });
});
