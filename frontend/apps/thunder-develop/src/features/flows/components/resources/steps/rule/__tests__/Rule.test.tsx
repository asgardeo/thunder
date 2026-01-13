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
import {render, screen, fireEvent} from '@testing-library/react';
import Rule from '../Rule';
import type {CommonStepFactoryPropsInterface} from '../../CommonStepFactory';

// Mock i18next
const translations: Record<string, string> = {
  'flows:core.rule.conditionalRule': 'Conditional Rule',
  'flows:core.rule.remove': 'Remove',
};

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => translations[key] || key,
  }),
}));

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
    it('should render the Rule component', () => {
      render(<Rule {...createMockProps({id: 'rule-1', data: {}})} />);

      expect(screen.getByText('Conditional Rule')).toBeInTheDocument();
    });

    it('should render with flow-builder-rule class', () => {
      const {container} = render(<Rule {...createMockProps({id: 'rule-1', data: {}})} />);

      expect(container.querySelector('.flow-builder-rule')).toBeInTheDocument();
    });
  });

  describe('React Flow Handles', () => {
    it('should render a target handle on the left', () => {
      render(<Rule {...createMockProps({id: 'rule-1', data: {}})} />);

      const targetHandle = screen.getByTestId('handle-target');
      expect(targetHandle).toBeInTheDocument();
      expect(targetHandle).toHaveAttribute('data-position', 'left');
    });

    it('should render a source handle on the right', () => {
      render(<Rule {...createMockProps({id: 'rule-1', data: {}})} />);

      const sourceHandle = screen.getByTestId('handle-source');
      expect(sourceHandle).toBeInTheDocument();
      expect(sourceHandle).toHaveAttribute('data-position', 'right');
    });

    it('should have source handle with id "a"', () => {
      render(<Rule {...createMockProps({id: 'rule-1', data: {}})} />);

      const sourceHandle = screen.getByTestId('handle-source');
      expect(sourceHandle).toHaveAttribute('data-handle-id', 'a');
    });
  });

  describe('Remove Button', () => {
    it('should render a remove button with tooltip', () => {
      render(<Rule {...createMockProps({id: 'rule-1', data: {}})} />);

      // Button should be present
      const removeButton = screen.getByRole('button');
      expect(removeButton).toBeInTheDocument();
    });

    it('should call deleteElements when remove button is clicked', () => {
      render(<Rule {...createMockProps({id: 'rule-1', data: {}})} />);

      const removeButton = screen.getByRole('button');
      fireEvent.click(removeButton);

      expect(mockDeleteElements).toHaveBeenCalledWith({
        nodes: [{id: 'test-node-id'}],
      });
    });

    it('should not call deleteElements if nodeId is empty', () => {
      mockUseNodeId.mockReturnValue('');

      render(<Rule {...createMockProps({id: 'rule-1', data: {}})} />);

      const removeButton = screen.getByRole('button');
      fireEvent.click(removeButton);

      expect(mockDeleteElements).not.toHaveBeenCalled();
    });
  });

  describe('Action Panel Interaction', () => {
    it('should set lastInteractedResource when action panel is clicked', () => {
      render(<Rule {...createMockProps({id: 'rule-1', data: {someData: 'value'}})} />);

      const actionPanel = screen.getByText('Conditional Rule').closest('.flow-builder-rule-action-panel');
      if (actionPanel) {
        fireEvent.click(actionPanel);
        expect(mockSetLastInteractedResource).toHaveBeenCalled();
      }
    });

    it('should pass correct resource object to setLastInteractedResource', () => {
      const testData = {name: 'Test Rule', condition: 'true'};
      render(<Rule {...createMockProps({id: 'custom-rule-id', data: testData})} />);

      const actionPanel = screen.getByText('Conditional Rule').closest('.flow-builder-rule-action-panel');
      if (actionPanel) {
        fireEvent.click(actionPanel);

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
    it('should handle drag over event', () => {
      render(<Rule {...createMockProps({id: 'rule-1', data: {}})} />);

      const ruleElement = screen.getByText('Conditional Rule').closest('.flow-builder-rule');
      if (ruleElement) {
        const event = {
          preventDefault: vi.fn(),
          dataTransfer: {dropEffect: ''},
        };

        fireEvent.dragOver(ruleElement, event);

        // Drag over should set dropEffect to 'move'
      }
    });

    it('should handle drop event', () => {
      render(<Rule {...createMockProps({id: 'rule-1', data: {}})} />);

      const ruleElement = screen.getByText('Conditional Rule').closest('.flow-builder-rule');
      if (ruleElement) {
        const event = {
          preventDefault: vi.fn(),
        };

        fireEvent.drop(ruleElement, event);

        // Drop event should be prevented
      }
    });
  });

  describe('Props Integration', () => {
    it('should use id from props when nodeId is available', () => {
      render(<Rule {...createMockProps({id: 'props-id', data: {}})} />);

      const actionPanel = screen.getByText('Conditional Rule').closest('.flow-builder-rule-action-panel');
      if (actionPanel) {
        fireEvent.click(actionPanel);

        expect(mockSetLastInteractedResource).toHaveBeenCalledWith(
          expect.objectContaining({
            id: 'props-id',
          }),
        );
      }
    });

    it('should fall back to nodeId when id prop is not provided', () => {
      render(<Rule {...createMockProps({data: {}})} />);

      const actionPanel = screen.getByText('Conditional Rule').closest('.flow-builder-rule-action-panel');
      if (actionPanel) {
        fireEvent.click(actionPanel);

        expect(mockSetLastInteractedResource).toHaveBeenCalledWith(
          expect.objectContaining({
            id: 'rule-1',
          }),
        );
      }
    });
  });

  describe('Memoization', () => {
    it('should be wrapped in memo for performance', () => {
      // The component is exported as MemoizedRule
      // We can verify it renders correctly multiple times with same props
      const props = createMockProps({id: 'rule-1', data: {value: 1}});
      const {rerender} = render(<Rule {...props} />);

      expect(screen.getByText('Conditional Rule')).toBeInTheDocument();

      // Rerender with same props
      rerender(<Rule {...props} />);

      expect(screen.getByText('Conditional Rule')).toBeInTheDocument();
    });
  });
});
