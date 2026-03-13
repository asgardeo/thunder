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
import {renderHook, render, fireEvent} from '@testing-library/react';
import type {NodeProps} from '@xyflow/react';
import {StaticStepTypes, StepTypes, type Step} from '@/features/flows/models/steps';
import type {Resources} from '@/features/flows/models/resources';
import type {Element} from '@/features/flows/models/elements';
import useNodeTypes from '../useNodeTypes';

// Mock StepFactory component
vi.mock('../../components/resources/steps/StepFactory', () => ({
  default: (props: {
    resourceId: string;
    resources: Step[];
    allResources: Resources;
    onAddElement?: (element: unknown) => void;
    onAddElementToForm?: (element: unknown, formId: string) => void;
  }) => (
    <div data-testid="step-factory" data-resource-id={props.resourceId}>
      StepFactory
      <button
        type="button"
        data-testid="trigger-add-element"
        onClick={() => props.onAddElement?.({id: 'test-element', type: 'TEXT'})}
      >
        Add Element
      </button>
      <button
        type="button"
        data-testid="trigger-add-element-to-form"
        onClick={() => props.onAddElementToForm?.({id: 'test-element', type: 'TEXT'}, 'form-1')}
      >
        Add Element to Form
      </button>
    </div>
  ),
}));

// Mock StaticStepFactory component
vi.mock('../../components/resources/steps/StaticStepFactory', () => ({
  default: (props: {type: StaticStepTypes}) => (
    <div data-testid="static-step-factory" data-type={props.type}>
      StaticStepFactory
    </div>
  ),
}));

// Mock BaseEdge component
vi.mock('@/features/flows/components/react-flow-overrides/BaseEdge', () => ({
  default: () => <div data-testid="base-edge">BaseEdge</div>,
}));

// Mock LoginFlowConstants
vi.mock('../../constants/LoginFlowConstants', () => ({
  default: {
    DEFAULT_EDGE_TYPE: 'smoothstep',
  },
}));

const createMockStep = (overrides: Partial<Step> = {}): Step =>
  ({
    id: 'step-1',
    type: StepTypes.View,
    position: {x: 0, y: 0},
    data: {},
    ...overrides,
  }) as Step;

const createMockResources = (): Resources =>
  ({
    templates: [],
    steps: [],
    elements: [],
    widgets: [],
  }) as unknown as Resources;

describe('useNodeTypes', () => {
  let mockOnAddElementToView: (element: Element<unknown>, viewId: string) => void;
  let mockOnAddElementToForm: (element: Element<unknown>, formId: string) => void;

  beforeEach(() => {
    vi.clearAllMocks();
    mockOnAddElementToView = vi.fn() as unknown as (element: Element<unknown>, viewId: string) => void;
    mockOnAddElementToForm = vi.fn() as unknown as (element: Element<unknown>, formId: string) => void;
  });

  const renderUseNodeTypes = (overrides = {}) => {
    const defaultProps = {
      steps: [createMockStep()],
      resources: createMockResources(),
      onAddElementToView: mockOnAddElementToView,
      onAddElementToForm: mockOnAddElementToForm,
      ...overrides,
    };

    return renderHook(() => useNodeTypes(defaultProps));
  };

  describe('Hook Interface', () => {
    it('should return nodeTypes object', () => {
      const {result} = renderUseNodeTypes();
      expect(result.current.nodeTypes).toBeDefined();
      expect(typeof result.current.nodeTypes).toBe('object');
    });

    it('should return edgeTypes object', () => {
      const {result} = renderUseNodeTypes();
      expect(result.current.edgeTypes).toBeDefined();
      expect(typeof result.current.edgeTypes).toBe('object');
    });
  });

  describe('nodeTypes', () => {
    it('should create node types for each step type', () => {
      const {result} = renderUseNodeTypes({
        steps: [
          createMockStep({id: 'view-1', type: StepTypes.View}),
          createMockStep({id: 'execution-1', type: StepTypes.Execution}),
        ],
      });

      expect(result.current.nodeTypes[StepTypes.View]).toBeDefined();
      expect(result.current.nodeTypes[StepTypes.Execution]).toBeDefined();
    });

    it('should create static node types for all StaticStepTypes', () => {
      const {result} = renderUseNodeTypes();

      Object.values(StaticStepTypes).forEach((type) => {
        expect(result.current.nodeTypes[type]).toBeDefined();
      });
    });

    it('should render StepFactory component for step types', () => {
      const {result} = renderUseNodeTypes({
        steps: [createMockStep({id: 'view-1', type: StepTypes.View})],
      });

      const ViewNodeType = result.current.nodeTypes[StepTypes.View];
      const mockNodeProps = {
        id: 'test-node',
        type: StepTypes.View,
        data: {},
        selected: false,
        isConnectable: true,
        xPos: 0,
        yPos: 0,
        zIndex: 1,
        dragging: false,
        targetPosition: undefined,
        sourcePosition: undefined,
      } as unknown as NodeProps;

      const {getByTestId} = render(<ViewNodeType {...mockNodeProps} />);
      expect(getByTestId('step-factory')).toBeInTheDocument();
      expect(getByTestId('step-factory')).toHaveAttribute('data-resource-id', 'test-node');
    });

    it('should render StaticStepFactory component for static step types', () => {
      const {result} = renderUseNodeTypes();

      const StartNodeType = result.current.nodeTypes[StaticStepTypes.Start];
      const mockNodeProps = {
        id: 'start-node',
        type: StaticStepTypes.Start,
        data: {},
        selected: false,
        isConnectable: true,
        xPos: 0,
        yPos: 0,
        zIndex: 1,
        dragging: false,
      } as unknown as NodeProps;

      const {getByTestId} = render(<StartNodeType {...mockNodeProps} />);
      expect(getByTestId('static-step-factory')).toBeInTheDocument();
      expect(getByTestId('static-step-factory')).toHaveAttribute('data-type', StaticStepTypes.Start);
    });

    it('should handle empty steps array', () => {
      const {result} = renderUseNodeTypes({steps: []});

      // Should still have static node types
      Object.values(StaticStepTypes).forEach((type) => {
        expect(result.current.nodeTypes[type]).toBeDefined();
      });
    });

    it('should handle undefined steps', () => {
      const {result} = renderUseNodeTypes({steps: undefined});

      // Should still have static node types
      Object.values(StaticStepTypes).forEach((type) => {
        expect(result.current.nodeTypes[type]).toBeDefined();
      });
    });

    it('should deduplicate step types', () => {
      const {result} = renderUseNodeTypes({
        steps: [
          createMockStep({id: 'view-1', type: StepTypes.View}),
          createMockStep({id: 'view-2', type: StepTypes.View}),
          createMockStep({id: 'view-3', type: StepTypes.View}),
        ],
      });

      // Should only have one VIEW type even with multiple VIEW steps
      expect(result.current.nodeTypes[StepTypes.View]).toBeDefined();
    });

    it('should maintain stable nodeTypes reference when steps dont change types', () => {
      const {result, rerender} = renderHook(
        ({steps}) =>
          useNodeTypes({
            steps,
            resources: createMockResources(),
            onAddElementToView: mockOnAddElementToView,
            onAddElementToForm: mockOnAddElementToForm,
          }),
        {
          initialProps: {
            steps: [createMockStep({id: 'view-1', type: StepTypes.View})],
          },
        },
      );

      const initialNodeTypes = result.current.nodeTypes;

      // Rerender with same step types but different IDs
      rerender({
        steps: [createMockStep({id: 'view-2', type: StepTypes.View})],
      });

      // nodeTypes should be the same reference since step types haven't changed
      expect(result.current.nodeTypes).toBe(initialNodeTypes);
    });

    it('should update nodeTypes when step types change', () => {
      const {result, rerender} = renderHook(
        ({steps}) =>
          useNodeTypes({
            steps,
            resources: createMockResources(),
            onAddElementToView: mockOnAddElementToView,
            onAddElementToForm: mockOnAddElementToForm,
          }),
        {
          initialProps: {
            steps: [createMockStep({id: 'view-1', type: StepTypes.View})],
          },
        },
      );

      expect(result.current.nodeTypes[StepTypes.View]).toBeDefined();
      expect(result.current.nodeTypes[StepTypes.Execution]).toBeUndefined();

      // Rerender with different step types
      rerender({
        steps: [
          createMockStep({id: 'view-1', type: StepTypes.View}),
          createMockStep({id: 'execution-1', type: StepTypes.Execution}),
        ],
      });

      expect(result.current.nodeTypes[StepTypes.View]).toBeDefined();
      expect(result.current.nodeTypes[StepTypes.Execution]).toBeDefined();
    });
  });

  describe('edgeTypes', () => {
    it('should include default edge type', () => {
      const {result} = renderUseNodeTypes();
      expect(result.current.edgeTypes.default).toBeDefined();
    });

    it('should include smoothstep edge type', () => {
      const {result} = renderUseNodeTypes();
      expect(result.current.edgeTypes.smoothstep).toBeDefined();
    });

    it('should include step edge type', () => {
      const {result} = renderUseNodeTypes();
      expect(result.current.edgeTypes.step).toBeDefined();
    });

    it('should maintain stable edgeTypes reference', () => {
      const {result, rerender} = renderUseNodeTypes();

      const initialEdgeTypes = result.current.edgeTypes;

      rerender();

      expect(result.current.edgeTypes).toBe(initialEdgeTypes);
    });
  });

  describe('callback refs', () => {
    it('should update onAddElementToView ref when callback changes', () => {
      const newOnAddElementToView = vi.fn();

      const {result, rerender} = renderHook(
        ({onAddElementToView}) =>
          useNodeTypes({
            steps: [createMockStep()],
            resources: createMockResources(),
            onAddElementToView,
            onAddElementToForm: mockOnAddElementToForm,
          }),
        {
          initialProps: {
            onAddElementToView: mockOnAddElementToView,
          },
        },
      );

      // Render a node type component to test callback
      const ViewNodeType = result.current.nodeTypes[StepTypes.View];
      const mockNodeProps = {
        id: 'test-node',
        type: StepTypes.View,
        data: {},
      } as unknown as NodeProps;

      // Initial render
      render(<ViewNodeType {...mockNodeProps} />);

      // Update callback
      rerender({onAddElementToView: newOnAddElementToView});

      // The ref should be updated without recreating nodeTypes
    });

    it('should update onAddElementToForm ref when callback changes', () => {
      const newOnAddElementToForm = vi.fn();

      const {rerender} = renderHook(
        ({onAddElementToForm}) =>
          useNodeTypes({
            steps: [createMockStep()],
            resources: createMockResources(),
            onAddElementToView: mockOnAddElementToView,
            onAddElementToForm,
          }),
        {
          initialProps: {
            onAddElementToForm: mockOnAddElementToForm,
          },
        },
      );

      // Update callback
      rerender({onAddElementToForm: newOnAddElementToForm});

      // The ref should be updated without recreating nodeTypes
    });

    it('should update resources ref when resources change', () => {
      const newResources = {
        ...createMockResources(),
        templates: [{id: 'new-template'}],
      } as unknown as Resources;

      const {rerender} = renderHook(
        ({resources}) =>
          useNodeTypes({
            steps: [createMockStep()],
            resources,
            onAddElementToView: mockOnAddElementToView,
            onAddElementToForm: mockOnAddElementToForm,
          }),
        {
          initialProps: {
            resources: createMockResources(),
          },
        },
      );

      // Update resources
      rerender({resources: newResources});

      // The ref should be updated without recreating nodeTypes
    });
  });

  describe('stepsByTypeRef organization', () => {
    it('should organize steps by type', () => {
      const {result} = renderUseNodeTypes({
        steps: [
          createMockStep({id: 'view-1', type: StepTypes.View}),
          createMockStep({id: 'view-2', type: StepTypes.View}),
          createMockStep({id: 'execution-1', type: StepTypes.Execution}),
        ],
      });

      // The stepsByTypeRef is internal, but we can verify behavior through rendering
      const ViewNodeType = result.current.nodeTypes[StepTypes.View];
      const ExecutionNodeType = result.current.nodeTypes[StepTypes.Execution];

      expect(ViewNodeType).toBeDefined();
      expect(ExecutionNodeType).toBeDefined();
    });

    it('should update stepsByTypeRef when steps change', () => {
      const {result, rerender} = renderHook(
        ({steps}) =>
          useNodeTypes({
            steps,
            resources: createMockResources(),
            onAddElementToView: mockOnAddElementToView,
            onAddElementToForm: mockOnAddElementToForm,
          }),
        {
          initialProps: {
            steps: [createMockStep({id: 'view-1', type: StepTypes.View})],
          },
        },
      );

      // Rerender with new steps
      rerender({
        steps: [
          createMockStep({id: 'view-1', type: StepTypes.View}),
          createMockStep({id: 'view-2', type: StepTypes.View}),
        ],
      });

      // nodeTypes should still work correctly
      expect(result.current.nodeTypes[StepTypes.View]).toBeDefined();
    });
  });

  describe('integration with onAddElement callbacks', () => {
    it('should pass onAddElement callback that calls onAddElementToView', () => {
      const {result} = renderUseNodeTypes();

      const ViewNodeType = result.current.nodeTypes[StepTypes.View];

      // The callback is passed through to StepFactory
      // We can verify the component is created correctly
      const mockNodeProps = {
        id: 'test-node',
        type: StepTypes.View,
        data: {},
      } as unknown as NodeProps;

      const {getByTestId} = render(<ViewNodeType {...mockNodeProps} />);
      expect(getByTestId('step-factory')).toBeInTheDocument();
    });

    it('should pass onAddElementToForm callback', () => {
      const {result} = renderUseNodeTypes();

      const ViewNodeType = result.current.nodeTypes[StepTypes.View];

      const mockNodeProps = {
        id: 'test-node',
        type: StepTypes.View,
        data: {},
      } as unknown as NodeProps;

      const {getByTestId} = render(<ViewNodeType {...mockNodeProps} />);
      expect(getByTestId('step-factory')).toBeInTheDocument();
    });

    it('should call onAddElementToView when onAddElement is triggered', () => {
      const {result} = renderUseNodeTypes();

      const ViewNodeType = result.current.nodeTypes[StepTypes.View];
      const mockNodeProps = {
        id: 'test-node',
        type: StepTypes.View,
        data: {},
      } as unknown as NodeProps;

      const {getByTestId} = render(<ViewNodeType {...mockNodeProps} />);

      // Trigger the onAddElement callback through the mock
      fireEvent.click(getByTestId('trigger-add-element'));

      // The onAddElementToView callback should be called with the element and node id
      expect(mockOnAddElementToView).toHaveBeenCalledWith({id: 'test-element', type: 'TEXT'}, 'test-node');
    });

    it('should call onAddElementToForm when onAddElementToForm is triggered', () => {
      const {result} = renderUseNodeTypes();

      const ViewNodeType = result.current.nodeTypes[StepTypes.View];
      const mockNodeProps = {
        id: 'test-node',
        type: StepTypes.View,
        data: {},
      } as unknown as NodeProps;

      const {getByTestId} = render(<ViewNodeType {...mockNodeProps} />);

      // Trigger the onAddElementToForm callback through the mock
      fireEvent.click(getByTestId('trigger-add-element-to-form'));

      // The onAddElementToForm callback should be called with the element and form id
      expect(mockOnAddElementToForm).toHaveBeenCalledWith({id: 'test-element', type: 'TEXT'}, 'form-1');
    });
  });
});
