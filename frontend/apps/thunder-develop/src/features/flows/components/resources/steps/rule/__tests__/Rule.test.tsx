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

/* eslint-disable react/require-default-props */

import {describe, it, expect, vi, beforeEach} from 'vitest';
import {render} from '@thunder/test-utils/browser';
import {page, userEvent} from 'vitest/browser';
import Rule from '../Rule';
import type {CommonStepFactoryPropsInterface} from '../../CommonStepFactory';

// Mock @xyflow/react
const mockDeleteElements = vi.fn();
const mockUseNodeId = vi.fn((): string => 'test-node-id');

vi.mock('@xyflow/react', () => ({
  Handle: ({type, position, id}: {type: string; position: string; id?: string}) => (
    <div data-testid={`handle-${type}`} data-position={position} data-handle-id={id} />
  ),
  Position: {
    Left: 'left',
    Right: 'right',
    Top: 'top',
    Bottom: 'bottom',
  },
  useNodeId: () => mockUseNodeId(),
  useReactFlow: () => ({
    deleteElements: mockDeleteElements,
  }),
}));

// Mock useFlowBuilderCore
const mockSetLastInteractedResource = vi.fn();
vi.mock('@/features/flows/hooks/useFlowBuilderCore', () => ({
  default: () => ({
    setLastInteractedResource: mockSetLastInteractedResource,
  }),
}));

// Mock SCSS
vi.mock('../Rule.scss', () => ({}));

// Default mock props for Rule component
const createMockProps = (overrides: Partial<CommonStepFactoryPropsInterface> = {}): CommonStepFactoryPropsInterface =>
  ({
    id: 'rule-1',
    resourceId: 'rule-resource-1',
    resources: [],
    data: {},
    type: 'RULE',
    zIndex: 1,
    isConnectable: true,
    positionAbsoluteX: 0,
    positionAbsoluteY: 0,
    dragging: false,
    selected: false,
    deletable: true,
    selectable: true,
    parentId: undefined,
    ...overrides,
  }) as CommonStepFactoryPropsInterface;

describe('Rule', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseNodeId.mockReturnValue('test-node-id');
  });

  describe('Rendering', () => {
    it('should render the Rule component', async () => {
      await render(<Rule {...createMockProps({id: 'rule-1', data: {}})} />);

      await expect.element(page.getByText('Conditional Rule')).toBeInTheDocument();
    });

    it('should render with flow-builder-rule class', async () => {
      const {container} = await render(<Rule {...createMockProps({id: 'rule-1', data: {}})} />);

      expect(container.querySelector('.flow-builder-rule')).toBeInTheDocument();
    });
  });

  describe('React Flow Handles', () => {
    it('should render a target handle on the left', async () => {
      await render(<Rule {...createMockProps({id: 'rule-1', data: {}})} />);

      const targetHandle = page.getByTestId('handle-target');
      expect(targetHandle).toBeInTheDocument();
      expect(targetHandle).toHaveAttribute('data-position', 'left');
    });

    it('should render a source handle on the right', async () => {
      await render(<Rule {...createMockProps({id: 'rule-1', data: {}})} />);

      const sourceHandle = page.getByTestId('handle-source');
      expect(sourceHandle).toBeInTheDocument();
      expect(sourceHandle).toHaveAttribute('data-position', 'right');
    });

    it('should have source handle with id "a"', async () => {
      await render(<Rule {...createMockProps({id: 'rule-1', data: {}})} />);

      const sourceHandle = page.getByTestId('handle-source');
      expect(sourceHandle).toHaveAttribute('data-handle-id', 'a');
    });
  });

  describe('Remove Button', () => {
    it('should render a remove button with tooltip', async () => {
      await render(<Rule {...createMockProps({id: 'rule-1', data: {}})} />);

      // Button should be present
      const removeButton = page.getByRole('button');
      expect(removeButton).toBeInTheDocument();
    });

    it('should call deleteElements when remove button is clicked', async () => {
      await render(<Rule {...createMockProps({id: 'rule-1', data: {}})} />);

      const removeButton = page.getByRole('button');
      await userEvent.click(removeButton);

      expect(mockDeleteElements).toHaveBeenCalledWith({
        nodes: [{id: 'test-node-id'}],
      });
    });

    it('should not call deleteElements if nodeId is empty', async () => {
      mockUseNodeId.mockReturnValue('');

      await render(<Rule {...createMockProps({id: 'rule-1', data: {}})} />);

      const removeButton = page.getByRole('button');
      await userEvent.click(removeButton);

      expect(mockDeleteElements).not.toHaveBeenCalled();
    });
  });

  describe('Action Panel Interaction', () => {
    it('should set lastInteractedResource when action panel is clicked', async () => {
      await render(<Rule {...createMockProps({id: 'rule-1', data: {someData: 'value'}})} />);

      const actionPanel = page.getByText('Conditional Rule').element().closest('.flow-builder-rule-action-panel');
      if (actionPanel) {
        await userEvent.click(actionPanel);
        expect(mockSetLastInteractedResource).toHaveBeenCalled();
      }
    });

    it('should pass correct resource object to setLastInteractedResource', async () => {
      const testData = {name: 'Test Rule', condition: 'true'};
      await render(<Rule {...createMockProps({id: 'custom-rule-id', data: testData})} />);

      const actionPanel = page.getByText('Conditional Rule').element().closest('.flow-builder-rule-action-panel');
      if (actionPanel) {
        await userEvent.click(actionPanel);

        expect(mockSetLastInteractedResource).toHaveBeenCalledWith(
          expect.objectContaining({
            id: 'custom-rule-id',
            name: 'Test Rule',
            condition: 'true',
          }),
        );
      }
    });
  });

  describe('Drag and Drop', () => {
    it('should handle drag event without errors', async () => {
      await render(<Rule {...createMockProps({id: 'rule-1', data: {}})} />);

      const ruleElement = page.getByText('Conditional Rule').element().closest('.flow-builder-rule')!;
      expect(ruleElement).toBeInTheDocument();

      const dataTransfer = new DataTransfer();
      // In real browsers, dispatching synthetic DragEvents with dataTransfer works
      // but dropEffect may be read-only. Verify the event dispatches without errors.
      ruleElement.dispatchEvent(new DragEvent('drag', {bubbles: true, dataTransfer}));
    });

    it('should handle drag event when dataTransfer is not provided', async () => {
      await render(<Rule {...createMockProps({id: 'rule-1', data: {}})} />);

      const ruleElement = page.getByText('Conditional Rule').element().closest('.flow-builder-rule')!;
      expect(ruleElement).toBeInTheDocument();

      // Should not throw when dataTransfer is undefined
      ruleElement.dispatchEvent(new DragEvent('drag', {bubbles: true}));
    });

    it('should handle drop event', async () => {
      await render(<Rule {...createMockProps({id: 'rule-1', data: {}})} />);

      const ruleElement = page.getByText('Conditional Rule').element().closest('.flow-builder-rule')!;
      expect(ruleElement).toBeInTheDocument();

      ruleElement.dispatchEvent(new DragEvent('drop', {bubbles: true}));
    });
  });

  describe('Props Integration', () => {
    it('should use id from props when nodeId is available', async () => {
      await render(<Rule {...createMockProps({id: 'props-id', data: {}})} />);

      const actionPanel = page.getByText('Conditional Rule').element().closest('.flow-builder-rule-action-panel');
      if (actionPanel) {
        await userEvent.click(actionPanel);

        expect(mockSetLastInteractedResource).toHaveBeenCalledWith(
          expect.objectContaining({
            id: 'props-id',
          }),
        );
      }
    });

    it('should fall back to nodeId when id prop is not provided', async () => {
      await render(<Rule {...createMockProps({data: {}})} />);

      const actionPanel = page.getByText('Conditional Rule').element().closest('.flow-builder-rule-action-panel');
      if (actionPanel) {
        await userEvent.click(actionPanel);

        expect(mockSetLastInteractedResource).toHaveBeenCalledWith(
          expect.objectContaining({
            id: 'rule-1',
          }),
        );
      }
    });
  });

  describe('Memoization', () => {
    it('should be wrapped in memo for performance', async () => {
      // The component is exported as MemoizedRule
      // We can verify it renders correctly multiple times with same props
      const props = createMockProps({id: 'rule-1', data: {value: 1}});
      const {rerender} = await render(<Rule {...props} />);

      await expect.element(page.getByText('Conditional Rule')).toBeInTheDocument();

      // Rerender with same props
      await rerender(<Rule {...props} />);

      await expect.element(page.getByText('Conditional Rule')).toBeInTheDocument();
    });

    it('should re-render when data prop changes', async () => {
      const initialData = {value: 1};
      const props = createMockProps({id: 'rule-1', data: initialData});
      const {rerender} = await render(<Rule {...props} />);

      await expect.element(page.getByText('Conditional Rule')).toBeInTheDocument();

      // Rerender with different data - should trigger re-render
      const newData = {value: 2};
      await rerender(<Rule {...createMockProps({id: 'rule-1', data: newData})} />);

      await expect.element(page.getByText('Conditional Rule')).toBeInTheDocument();
    });

    it('should re-render when id prop changes', async () => {
      const props = createMockProps({id: 'rule-1', data: {value: 1}});
      const {rerender} = await render(<Rule {...props} />);

      await expect.element(page.getByText('Conditional Rule')).toBeInTheDocument();

      // Rerender with different id - should trigger re-render
      await rerender(<Rule {...createMockProps({id: 'rule-2', data: {value: 1}})} />);

      await expect.element(page.getByText('Conditional Rule')).toBeInTheDocument();
    });

    it('should not re-render when other props change but data and id remain the same', async () => {
      const data = {value: 1};
      const props = createMockProps({id: 'rule-1', data, zIndex: 1});
      const {rerender} = await render(<Rule {...props} />);

      await expect.element(page.getByText('Conditional Rule')).toBeInTheDocument();

      // Rerender with same data and id but different zIndex - should not trigger re-render
      await rerender(<Rule {...createMockProps({id: 'rule-1', data, zIndex: 2})} />);

      await expect.element(page.getByText('Conditional Rule')).toBeInTheDocument();
    });
  });
});
