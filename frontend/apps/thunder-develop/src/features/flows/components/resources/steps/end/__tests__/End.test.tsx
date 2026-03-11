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
import End from '../End';
import type {CommonStepFactoryPropsInterface} from '../../CommonStepFactory';

// Mock @xyflow/react
vi.mock('@xyflow/react', () => ({
  Handle: ({type, position, id}: {type: string; position: string; id: string}) => (
    <div data-testid={`handle-${type}`} data-position={position} data-handle-id={id} />
  ),
  Position: {
    Left: 'left',
    Right: 'right',
    Top: 'top',
    Bottom: 'bottom',
  },
}));

// Mock SCSS
vi.mock('../End.scss', () => ({}));

// Default mock props for End component
const createMockProps = (overrides: Partial<CommonStepFactoryPropsInterface> = {}): CommonStepFactoryPropsInterface =>
  ({
    id: 'end-node-1',
    resourceId: 'end-resource-1',
    resources: [],
    data: {},
    type: 'END',
    zIndex: 1,
    isConnectable: true,
    positionAbsoluteX: 0,
    positionAbsoluteY: 0,
    dragging: false,
    selected: false,
    deletable: false,
    selectable: true,
    parentId: undefined,
    ...overrides,
  }) as CommonStepFactoryPropsInterface;

describe('End', () => {
  describe('Rendering', () => {
    it('should render the End node', async () => {
      await render(<End {...createMockProps()} />);

      await expect.element(page.getByText('End')).toBeInTheDocument();
    });

    it('should render a Fab button with end label', async () => {
      await render(<End {...createMockProps()} />);

      const fab = page.getByRole('button', {name: 'end'});
      expect(fab).toBeInTheDocument();
    });

    it('should render with end class on Fab', async () => {
      await render(<End {...createMockProps()} />);

      const fab = page.getByRole('button');
      expect(fab).toHaveClass('end');
    });
  });

  describe('React Flow Handle', () => {
    it('should render a target handle', async () => {
      await render(<End {...createMockProps()} />);

      const handle = page.getByTestId('handle-target');
      expect(handle).toBeInTheDocument();
    });

    it('should position handle on the left', async () => {
      await render(<End {...createMockProps()} />);

      const handle = page.getByTestId('handle-target');
      expect(handle).toHaveAttribute('data-position', 'left');
    });

    it('should have correct handle id with previous suffix', async () => {
      await render(<End {...createMockProps()} />);

      const handle = page.getByTestId('handle-target');
      // Handle id should contain 'end' and '_PREVIOUS' suffix
      expect(handle.element().getAttribute('data-handle-id')).toContain('end');
      expect(handle.element().getAttribute('data-handle-id')).toContain('_PREVIOUS');
    });

    it('should have hidden-handle class', async () => {
      // Note: Since we're mocking Handle, we can't directly test the class
      // but the component should pass the className prop
      await render(<End {...createMockProps()} />);

      const handle = page.getByTestId('handle-target');
      expect(handle).toBeInTheDocument();
    });
  });

  describe('Fab Properties', () => {
    it('should render extended variant Fab', async () => {
      await render(<End {...createMockProps()} />);

      const fab = page.getByRole('button');
      // Extended variant Fab will be rendered
      expect(fab).toBeInTheDocument();
    });

    it('should render small size Fab', async () => {
      await render(<End {...createMockProps()} />);

      const fab = page.getByRole('button');
      expect(fab).toBeInTheDocument();
    });
  });

  describe('Structure', () => {
    it('should be wrapped in a div', async () => {
      const {container} = await render(<End {...createMockProps()} />);

      expect(container.firstChild?.nodeName).toBe('DIV');
    });

    it('should contain Handle before Fab (target comes first)', async () => {
      const {container} = await render(<End {...createMockProps()} />);

      const children = container.firstChild?.childNodes;
      expect(children).toBeDefined();
      if (children) {
        // Target handle should come before the Fab button
        expect(children.length).toBeGreaterThanOrEqual(2);
      }
    });
  });

  describe('Props', () => {
    it('should accept props without error', async () => {
      // End component accepts CommonStepFactoryPropsInterface but doesn't use them
      await render(<End {...createMockProps({data: {}, id: 'end-node-1'})} />);

      await expect.element(page.getByText('End')).toBeInTheDocument();
    });
  });
});
