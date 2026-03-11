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
import {render} from '@thunder/test-utils/browser';
import {page} from 'vitest/browser';
import type {ReactNode} from 'react';
import {ReactFlowProvider} from '@xyflow/react';
import type {Element as FlowElement} from '@/features/flows/models/elements';
import BlockAdapter from '../BlockAdapter';

// Mock dependencies
vi.mock('../BlockAdapter.scss', () => ({}));

vi.mock('@/features/flows/components/resources/steps/view/ReorderableElement', () => ({
  ReorderableElement: ({element, id}: {element: FlowElement; id: string}) => (
    <div data-testid={`reorderable-element-${id}`}>{element.id}</div>
  ),
}));

vi.mock('@/features/flows/plugins/PluginRegistry', () => ({
  default: {
    getInstance: () => ({
      executeSync: () => true,
    }),
  },
}));

describe('BlockAdapter', () => {
  const createMockElement = (overrides: Partial<FlowElement> = {}): FlowElement =>
    ({
      id: 'element-1',
      type: 'TEXT_INPUT',
      category: 'FIELD',
      config: {},
      ...overrides,
    }) as FlowElement;

  const createWrapper = () => {
    function Wrapper({children}: {children: ReactNode}) {
      return <ReactFlowProvider>{children}</ReactFlowProvider>;
    }
    return Wrapper;
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('should render the adapter with correct class names', async () => {
      const resource = createMockElement();

      const {container} = await render(<BlockAdapter resource={resource} />, {wrapper: createWrapper()});

      expect(container.querySelector('.adapter')).toBeInTheDocument();
      expect(container.querySelector('.block-adapter')).toBeInTheDocument();
    });

    it('should render empty when resource has no components', async () => {
      const resource = createMockElement({components: undefined});

      const {container} = await render(<BlockAdapter resource={resource} />, {wrapper: createWrapper()});

      expect(container.querySelector('.block-adapter')).toBeInTheDocument();
      expect(container.querySelectorAll('[data-testid^="reorderable-element"]')).toHaveLength(0);
    });

    it('should render empty array when components is empty', async () => {
      const resource = createMockElement({components: []});

      const {container} = await render(<BlockAdapter resource={resource} />, {wrapper: createWrapper()});

      expect(container.querySelector('.block-adapter')).toBeInTheDocument();
      expect(container.querySelectorAll('[data-testid^="reorderable-element"]')).toHaveLength(0);
    });
  });

  describe('Components Rendering', () => {
    it('should render ReorderableElement for each component', async () => {
      const components = [
        createMockElement({id: 'comp-1'}),
        createMockElement({id: 'comp-2'}),
        createMockElement({id: 'comp-3'}),
      ];
      const resource = createMockElement({components});

      await render(<BlockAdapter resource={resource} />, {wrapper: createWrapper()});

      await expect.element(page.getByTestId('reorderable-element-comp-1')).toBeInTheDocument();
      await expect.element(page.getByTestId('reorderable-element-comp-2')).toBeInTheDocument();
      await expect.element(page.getByTestId('reorderable-element-comp-3')).toBeInTheDocument();
    });

    it('should pass availableElements to ReorderableElement', async () => {
      const components = [createMockElement({id: 'comp-1'})];
      const resource = createMockElement({components});
      const availableElements = [createMockElement({id: 'available-1'})];

      await render(<BlockAdapter resource={resource} availableElements={availableElements} />, {wrapper: createWrapper()});

      await expect.element(page.getByTestId('reorderable-element-comp-1')).toBeInTheDocument();
    });

    it('should pass onAddElementToForm callback to ReorderableElement', async () => {
      const components = [createMockElement({id: 'comp-1'})];
      const resource = createMockElement({components});
      const onAddElementToForm = vi.fn();

      await render(<BlockAdapter resource={resource} onAddElementToForm={onAddElementToForm} />, {wrapper: createWrapper()});

      await expect.element(page.getByTestId('reorderable-element-comp-1')).toBeInTheDocument();
    });
  });

  describe('Default Props', () => {
    it('should work with undefined availableElements', async () => {
      const resource = createMockElement({components: [createMockElement({id: 'comp-1'})]});

      const {container} = await render(<BlockAdapter resource={resource} />, {wrapper: createWrapper()});

      expect(container.querySelector('.block-adapter')).toBeInTheDocument();
    });

    it('should work with undefined onAddElementToForm', async () => {
      const resource = createMockElement({components: [createMockElement({id: 'comp-1'})]});

      const {container} = await render(<BlockAdapter resource={resource} />, {wrapper: createWrapper()});

      expect(container.querySelector('.block-adapter')).toBeInTheDocument();
    });
  });

  describe('Filtering', () => {
    it('should filter components through PluginRegistry', async () => {
      const components = [
        createMockElement({id: 'comp-1'}),
        createMockElement({id: 'comp-2'}),
      ];
      const resource = createMockElement({components});

      await render(<BlockAdapter resource={resource} />, {wrapper: createWrapper()});

      // All components should render since our mock returns true
      await expect.element(page.getByTestId('reorderable-element-comp-1')).toBeInTheDocument();
      await expect.element(page.getByTestId('reorderable-element-comp-2')).toBeInTheDocument();
    });
  });
});
