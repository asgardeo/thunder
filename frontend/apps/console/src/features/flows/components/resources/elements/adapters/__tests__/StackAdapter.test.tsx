/**
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
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
import {render, screen, fireEvent} from '@testing-library/react';
import type {ReactNode} from 'react';
import type {Element as FlowElement} from '@/features/flows/models/elements';
import StackAdapter, {type StackElement} from '../StackAdapter';

const mockUpdateNodeData = vi.fn();

vi.mock('@xyflow/react', () => ({
  useReactFlow: () => ({
    updateNodeData: mockUpdateNodeData,
  }),
}));

vi.mock('@/features/flows/plugins/PluginRegistry', () => ({
  default: {
    getInstance: () => ({
      executeSync: () => true,
    }),
  },
}));

vi.mock('@/features/flows/utils/generateResourceId', () => ({
  default: (prefix: string) => `${prefix}-generated`,
}));

vi.mock('@/features/flows/components/resources/steps/view/ReorderableElement', () => ({
  default: ({
    element,
    id,
    extraActions,
  }: {
    element: FlowElement;
    id: string;
    index: number;
    extraActions?: ReactNode;
  }) => (
    <div data-testid={`reorderable-element-${id}`}>
      <span data-testid={`element-label-${id}`}>{element.id}</span>
      {extraActions && <div data-testid={`extra-actions-${id}`}>{extraActions}</div>}
    </div>
  ),
}));

vi.mock('@/features/flows/components/dnd/Droppable', () => ({
  default: ({children, id}: {children: ReactNode; id: string}) => (
    <div data-testid="droppable" data-droppable-id={id}>
      {children}
    </div>
  ),
}));

vi.mock('@/features/flows/components/dnd/Handle', () => ({
  default: ({
    children,
    label,
    onClick,
  }: {
    children: ReactNode;
    label: string;
    onClick: () => void;
  }) => (
    <button type="button" data-testid={`handle-${label}`} aria-label={label} onClick={onClick}>
      {children}
    </button>
  ),
}));

describe('StackAdapter', () => {
  const createMockElement = (overrides: Partial<StackElement> = {}): StackElement =>
    ({
      id: 'stack-1',
      type: 'BLOCK',
      category: 'BLOCK',
      config: {},
      ...overrides,
    }) as StackElement;

  const createChildElement = (id: string): FlowElement =>
    ({
      id,
      type: 'ELEMENT',
      category: 'FIELD',
      config: {},
    }) as FlowElement;

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('should render droppable container', () => {
      const resource = createMockElement();

      render(<StackAdapter resource={resource} stepId="step-1" />);

      expect(screen.getByTestId('droppable')).toBeInTheDocument();
    });

    it('should render child elements via ReorderableFlowElement', () => {
      const resource = createMockElement({
        components: [createChildElement('child-1'), createChildElement('child-2')],
      });

      render(<StackAdapter resource={resource} stepId="step-1" />);

      expect(screen.getByTestId('reorderable-element-child-1')).toBeInTheDocument();
      expect(screen.getByTestId('reorderable-element-child-2')).toBeInTheDocument();
    });

    it('should show placeholder when no children exist in flex mode', () => {
      const resource = createMockElement({components: []});

      render(<StackAdapter resource={resource} stepId="step-1" />);

      expect(screen.getByText('Drop here')).toBeInTheDocument();
    });

    it('should show placeholder when components is undefined', () => {
      const resource = createMockElement({components: undefined});

      render(<StackAdapter resource={resource} stepId="step-1" />);

      expect(screen.getByText('Drop here')).toBeInTheDocument();
    });

    it('should not show placeholder when children exist in flex mode', () => {
      const resource = createMockElement({
        components: [createChildElement('child-1')],
      });

      render(<StackAdapter resource={resource} stepId="step-1" />);

      expect(screen.queryByText('Drop here')).not.toBeInTheDocument();
    });
  });

  describe('Grid mode (items >= 2)', () => {
    it('should show empty placeholder slots for unoccupied grid positions', () => {
      const resource = createMockElement({
        items: 3,
        components: [createChildElement('child-1')],
      });

      render(<StackAdapter resource={resource} stepId="step-1" />);

      // 3 items - 1 child = 2 empty slots
      const dropHereTexts = screen.getAllByText('Drop here');
      expect(dropHereTexts).toHaveLength(2);
    });

    it('should show no empty slots when all grid positions are filled', () => {
      const resource = createMockElement({
        items: 2,
        components: [createChildElement('child-1'), createChildElement('child-2')],
      });

      render(<StackAdapter resource={resource} stepId="step-1" />);

      expect(screen.queryByText('Drop here')).not.toBeInTheDocument();
    });

    it('should show no empty slots when more children than grid items', () => {
      const resource = createMockElement({
        items: 2,
        components: [
          createChildElement('child-1'),
          createChildElement('child-2'),
          createChildElement('child-3'),
        ],
      });

      render(<StackAdapter resource={resource} stepId="step-1" />);

      expect(screen.queryByText('Drop here')).not.toBeInTheDocument();
    });
  });

  describe('Move actions for row direction', () => {
    it('should show Move Right but not Move Left for first element', () => {
      const resource = createMockElement({
        direction: 'row',
        components: [createChildElement('child-1'), createChildElement('child-2')],
      });

      render(<StackAdapter resource={resource} stepId="step-1" />);

      const firstActions = screen.getByTestId('extra-actions-child-1');
      expect(firstActions.querySelector('[data-testid="handle-Move Right"]')).toBeInTheDocument();
      expect(
        firstActions.querySelector('[data-testid="handle-Move Left"]'),
      ).not.toBeInTheDocument();
    });

    it('should show Move Left but not Move Right for last element', () => {
      const resource = createMockElement({
        direction: 'row',
        components: [createChildElement('child-1'), createChildElement('child-2')],
      });

      render(<StackAdapter resource={resource} stepId="step-1" />);

      const lastActions = screen.getByTestId('extra-actions-child-2');
      expect(lastActions.querySelector('[data-testid="handle-Move Left"]')).toBeInTheDocument();
      expect(
        lastActions.querySelector('[data-testid="handle-Move Right"]'),
      ).not.toBeInTheDocument();
    });

    it('should show both Move Left and Move Right for middle element', () => {
      const resource = createMockElement({
        direction: 'row',
        components: [
          createChildElement('child-1'),
          createChildElement('child-2'),
          createChildElement('child-3'),
        ],
      });

      render(<StackAdapter resource={resource} stepId="step-1" />);

      const middleActions = screen.getByTestId('extra-actions-child-2');
      expect(middleActions.querySelector('[data-testid="handle-Move Left"]')).toBeInTheDocument();
      expect(middleActions.querySelector('[data-testid="handle-Move Right"]')).toBeInTheDocument();
    });
  });

  describe('Move actions for column direction', () => {
    it('should show Move Down but not Move Up for first element', () => {
      const resource = createMockElement({
        direction: 'column',
        components: [createChildElement('child-1'), createChildElement('child-2')],
      });

      render(<StackAdapter resource={resource} stepId="step-1" />);

      const firstActions = screen.getByTestId('extra-actions-child-1');
      expect(firstActions.querySelector('[data-testid="handle-Move Down"]')).toBeInTheDocument();
      expect(firstActions.querySelector('[data-testid="handle-Move Up"]')).not.toBeInTheDocument();
    });

    it('should show Move Up but not Move Down for last element', () => {
      const resource = createMockElement({
        direction: 'column',
        components: [createChildElement('child-1'), createChildElement('child-2')],
      });

      render(<StackAdapter resource={resource} stepId="step-1" />);

      const lastActions = screen.getByTestId('extra-actions-child-2');
      expect(lastActions.querySelector('[data-testid="handle-Move Up"]')).toBeInTheDocument();
      expect(
        lastActions.querySelector('[data-testid="handle-Move Down"]'),
      ).not.toBeInTheDocument();
    });
  });

  describe('Move functionality', () => {
    it('should call updateNodeData when Move Right is clicked', () => {
      const resource = createMockElement({
        direction: 'row',
        components: [createChildElement('child-1'), createChildElement('child-2')],
      });

      render(<StackAdapter resource={resource} stepId="step-1" />);

      const moveRightButton = screen.getByTestId('handle-Move Right');
      fireEvent.click(moveRightButton);

      expect(mockUpdateNodeData).toHaveBeenCalledWith('step-1', expect.any(Function));
    });

    it('should call updateNodeData when Move Left is clicked', () => {
      const resource = createMockElement({
        direction: 'row',
        components: [createChildElement('child-1'), createChildElement('child-2')],
      });

      render(<StackAdapter resource={resource} stepId="step-1" />);

      const moveLeftButton = screen.getByTestId('handle-Move Left');
      fireEvent.click(moveLeftButton);

      expect(mockUpdateNodeData).toHaveBeenCalledWith('step-1', expect.any(Function));
    });

    it('should swap elements when move callback is executed', () => {
      const child1 = createChildElement('child-1');
      const child2 = createChildElement('child-2');
      const resource = createMockElement({
        id: 'stack-1',
        direction: 'row',
        components: [child1, child2],
      });

      render(<StackAdapter resource={resource} stepId="step-1" />);

      // Click Move Right on first child
      fireEvent.click(screen.getByTestId('handle-Move Right'));

      expect(mockUpdateNodeData).toHaveBeenCalledWith('step-1', expect.any(Function));

      // Execute the callback to verify the swap logic
      const updateFn = mockUpdateNodeData.mock.calls[0][1] as (node: {data: unknown}) => unknown;
      const result = updateFn({
        data: {
          components: [
            {
              id: 'stack-1',
              components: [child1, child2],
            },
          ],
        },
      });

      const updatedStack = (result as {components: FlowElement[]}).components[0];
      expect(updatedStack.components![0].id).toBe('child-2');
      expect(updatedStack.components![1].id).toBe('child-1');
    });

    it('should not swap when element is not found in stack', () => {
      const child1 = createChildElement('child-1');
      const child2 = createChildElement('child-2');
      const resource = createMockElement({
        id: 'stack-1',
        direction: 'row',
        components: [child1, child2],
      });

      render(<StackAdapter resource={resource} stepId="step-1" />);

      fireEvent.click(screen.getByTestId('handle-Move Right'));

      const updateFn = mockUpdateNodeData.mock.calls[0][1] as (node: {data: unknown}) => unknown;
      // Pass node data where the stack has different children
      const result = updateFn({
        data: {
          components: [
            {
              id: 'stack-1',
              components: [createChildElement('other-1'), createChildElement('other-2')],
            },
          ],
        },
      });

      // Should remain unchanged since child-1 is not in this stack
      const updatedStack = (result as {components: FlowElement[]}).components[0];
      expect(updatedStack.components![0].id).toBe('other-1');
      expect(updatedStack.components![1].id).toBe('other-2');
    });

    it('should recursively search nested elements for the stack', () => {
      const child1 = createChildElement('child-1');
      const child2 = createChildElement('child-2');
      const resource = createMockElement({
        id: 'stack-1',
        direction: 'row',
        components: [child1, child2],
      });

      render(<StackAdapter resource={resource} stepId="step-1" />);

      fireEvent.click(screen.getByTestId('handle-Move Right'));

      const updateFn = mockUpdateNodeData.mock.calls[0][1] as (node: {data: unknown}) => unknown;
      // Stack is nested inside a parent element
      const result = updateFn({
        data: {
          components: [
            {
              id: 'parent-block',
              components: [
                {
                  id: 'stack-1',
                  components: [child1, child2],
                },
              ],
            },
          ],
        },
      });

      const parentBlock = (result as {components: FlowElement[]}).components[0];
      const updatedStack = parentBlock.components![0];
      expect(updatedStack.components![0].id).toBe('child-2');
      expect(updatedStack.components![1].id).toBe('child-1');
    });
  });

  describe('Single element', () => {
    it('should not show any move actions for a single element', () => {
      const resource = createMockElement({
        direction: 'row',
        components: [createChildElement('child-1')],
      });

      render(<StackAdapter resource={resource} stepId="step-1" />);

      const actions = screen.getByTestId('extra-actions-child-1');
      expect(actions.querySelector('[data-testid^="handle-"]')).not.toBeInTheDocument();
    });
  });

  describe('Default direction', () => {
    it('should default to row direction and show left/right actions', () => {
      const resource = createMockElement({
        // No direction specified
        components: [createChildElement('child-1'), createChildElement('child-2')],
      });

      render(<StackAdapter resource={resource} stepId="step-1" />);

      // Should use row direction by default — Move Right for first, Move Left for last
      expect(screen.getByTestId('handle-Move Right')).toBeInTheDocument();
      expect(screen.getByTestId('handle-Move Left')).toBeInTheDocument();
    });
  });

  describe('Plugin filtering', () => {
    it('should filter components through PluginRegistry', () => {
      // PluginRegistry mock returns true for all, so all children render
      const resource = createMockElement({
        components: [createChildElement('child-1'), createChildElement('child-2')],
      });

      render(<StackAdapter resource={resource} stepId="step-1" />);

      expect(screen.getByTestId('reorderable-element-child-1')).toBeInTheDocument();
      expect(screen.getByTestId('reorderable-element-child-2')).toBeInTheDocument();
    });
  });
});
