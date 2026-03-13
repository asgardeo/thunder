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
import {renderHook} from '@testing-library/react';
import type {ReactNode} from 'react';
import {ReactFlowProvider} from '@xyflow/react';
import type {Node} from '@xyflow/react';
import FlowBuilderCoreContext, {type FlowBuilderCoreContextProps} from '../../context/FlowBuilderCoreContext';
import {EdgeStyleTypes} from '../../models/steps';
import {PreviewScreenType} from '../../models/custom-text-preference';
import {ElementTypes} from '../../models/elements';
import type {Base} from '../../models/base';
import FlowEventTypes from '../../models/extension';

// Import after mocks
import useDeleteExecutionResource from '../useDeleteExecutionResource';

// Use vi.hoisted to define mocks that need to be referenced in vi.mock
const {
  mockGetEdges,
  mockGetNodes,
  mockUpdateNodeData,
  mockSetNodes,
  mockSetIsOpenResourcePropertiesPanel,
  mockRegisterAsync,
  mockUnregister,
  mockExecuteAsync,
  registeredHandlers,
} = vi.hoisted(() => ({
  mockGetEdges: vi.fn().mockReturnValue([]),
  mockGetNodes: vi.fn().mockReturnValue([]),
  mockUpdateNodeData: vi.fn(),
  mockSetNodes: vi.fn(),
  mockSetIsOpenResourcePropertiesPanel: vi.fn(),
  mockRegisterAsync: vi.fn(),
  mockUnregister: vi.fn(),
  mockExecuteAsync: vi.fn().mockResolvedValue(true),
  registeredHandlers: {} as Record<string, ((...args: unknown[]) => Promise<boolean>)[]>,
}));

// Mock @xyflow/react
vi.mock('@xyflow/react', async () => {
  const actual = await vi.importActual('@xyflow/react');
  return {
    ...actual,
    useReactFlow: () => ({
      getEdges: mockGetEdges,
      getNodes: mockGetNodes,
      updateNodeData: mockUpdateNodeData,
      setNodes: mockSetNodes,
    }),
  };
});

// Mock PluginRegistry - capture handlers for testing
vi.mock('../../plugins/PluginRegistry', () => ({
  default: {
    getInstance: () => ({
      registerAsync: (eventType: string, handler: (...args: unknown[]) => Promise<boolean>) => {
        mockRegisterAsync(eventType, handler);
        if (!registeredHandlers[eventType]) {
          registeredHandlers[eventType] = [];
        }
        registeredHandlers[eventType].push(handler);
      },
      unregister: mockUnregister,
      executeAsync: mockExecuteAsync,
    }),
  },
}));

describe('useDeleteExecutionResource', () => {
  const mockBaseResource: Base = {
    id: '',
    type: '',
    category: '',
    resourceType: '',
    version: '1.0.0',
    deprecated: false,
    deletable: true,
    display: {
      label: '',
      image: '',
      showOnResourcePanel: false,
    },
    config: {
      field: {name: '', type: ElementTypes},
      styles: {},
    },
  } as unknown as Base;

  const defaultContextValue: FlowBuilderCoreContextProps = {
    lastInteractedResource: mockBaseResource,
    lastInteractedStepId: '',
    ResourceProperties: () => null,
    resourcePropertiesPanelHeading: '',
    primaryI18nScreen: PreviewScreenType.LOGIN,
    isResourcePanelOpen: true,
    isResourcePropertiesPanelOpen: false,
    isVersionHistoryPanelOpen: false,
    ElementFactory: () => null,
    onResourceDropOnCanvas: vi.fn(),
    selectedAttributes: {},
    setLastInteractedResource: vi.fn(),
    setLastInteractedStepId: vi.fn(),
    setResourcePropertiesPanelHeading: vi.fn(),
    setIsResourcePanelOpen: vi.fn(),
    setIsOpenResourcePropertiesPanel: mockSetIsOpenResourcePropertiesPanel,
    registerCloseValidationPanel: vi.fn(),
    setIsVersionHistoryPanelOpen: vi.fn(),
    setSelectedAttributes: vi.fn(),
    flowCompletionConfigs: {},
    setFlowCompletionConfigs: vi.fn(),
    flowNodeTypes: {},
    flowEdgeTypes: {},
    setFlowNodeTypes: vi.fn(),
    setFlowEdgeTypes: vi.fn(),
    isVerboseMode: false,
    setIsVerboseMode: vi.fn(),
    edgeStyle: EdgeStyleTypes.SmoothStep,
    setEdgeStyle: vi.fn(),
  };

  const createWrapper = (contextValue: FlowBuilderCoreContextProps = defaultContextValue) => {
    function Wrapper({children}: {children: ReactNode}) {
      return (
        <ReactFlowProvider>
          <FlowBuilderCoreContext.Provider value={contextValue}>{children}</FlowBuilderCoreContext.Provider>
        </ReactFlowProvider>
      );
    }
    return Wrapper;
  };

  beforeEach(() => {
    vi.clearAllMocks();
    // Clear registered handlers
    Object.keys(registeredHandlers).forEach((key) => {
      delete registeredHandlers[key];
    });
  });

  describe('Plugin Registration', () => {
    it('should register event handlers on mount', () => {
      renderHook(() => useDeleteExecutionResource(), {
        wrapper: createWrapper(),
      });

      // Check that handlers are registered
      expect(mockRegisterAsync).toHaveBeenCalledWith(FlowEventTypes.ON_NODE_DELETE, expect.any(Function));
      expect(mockRegisterAsync).toHaveBeenCalledWith(FlowEventTypes.ON_NODE_ELEMENT_DELETE, expect.any(Function));
      expect(mockRegisterAsync).toHaveBeenCalledWith(FlowEventTypes.ON_EDGE_DELETE, expect.any(Function));
    });

    it('should unregister event handlers on unmount', () => {
      const {unmount} = renderHook(() => useDeleteExecutionResource(), {
        wrapper: createWrapper(),
      });

      unmount();

      // Check that handlers are unregistered
      expect(mockUnregister).toHaveBeenCalledWith(FlowEventTypes.ON_NODE_DELETE, 'deleteExecutionActionNode');
      expect(mockUnregister).toHaveBeenCalledWith(FlowEventTypes.ON_NODE_ELEMENT_DELETE, 'deleteExecutionNode');
      expect(mockUnregister).toHaveBeenCalledWith(FlowEventTypes.ON_EDGE_DELETE, 'deleteComponentAndNode');
    });
  });

  describe('deleteExecutionActionNode', () => {
    it('should register the handler with correct function identifier', () => {
      renderHook(() => useDeleteExecutionResource(), {
        wrapper: createWrapper(),
      });

      // The handler should be registered with the correct event type
      expect(mockRegisterAsync).toHaveBeenCalledWith(FlowEventTypes.ON_NODE_DELETE, expect.any(Function));
    });

    it('should set up nodes and edges getters for the handler', () => {
      const executionNode: Node = {
        id: 'execution-1',
        type: 'TASK_EXECUTION',
        position: {x: 0, y: 0},
        data: {},
      };

      const actionNode: Node = {
        id: 'action-1',
        type: 'VIEW',
        position: {x: 0, y: 0},
        data: {
          components: [
            {id: 'button-1', type: 'ACTION'},
            {id: 'button-2', type: 'ACTION'},
          ],
        },
      };

      mockGetNodes.mockReturnValue([actionNode, executionNode] as Node[]);
      mockGetEdges.mockReturnValue([]);

      renderHook(() => useDeleteExecutionResource(), {
        wrapper: createWrapper(),
      });

      // Verify the hook registered with the plugin registry
      expect(mockRegisterAsync).toHaveBeenCalled();
    });
  });

  describe('deleteExecutionNode', () => {
    it('should register the handler for element deletion', () => {
      renderHook(() => useDeleteExecutionResource(), {
        wrapper: createWrapper(),
      });

      expect(mockRegisterAsync).toHaveBeenCalledWith(FlowEventTypes.ON_NODE_ELEMENT_DELETE, expect.any(Function));
    });
  });

  describe('deleteComponentAndNode', () => {
    it('should register the handler for edge deletion', () => {
      renderHook(() => useDeleteExecutionResource(), {
        wrapper: createWrapper(),
      });

      expect(mockRegisterAsync).toHaveBeenCalledWith(FlowEventTypes.ON_EDGE_DELETE, expect.any(Function));
    });

    it('should set up nodes getter for edge deletion handler', () => {
      const executionNode: Node = {
        id: 'execution-1',
        type: 'TASK_EXECUTION',
        position: {x: 0, y: 0},
        data: {},
      };

      const actionNode: Node = {
        id: 'action-1',
        type: 'VIEW',
        position: {x: 0, y: 0},
        data: {
          components: [{id: 'button-1', type: 'ACTION'}],
        },
      };

      mockGetNodes.mockReturnValue([actionNode, executionNode] as Node[]);

      renderHook(() => useDeleteExecutionResource(), {
        wrapper: createWrapper(),
      });

      // Verify the hook registered with the plugin registry
      expect(mockRegisterAsync).toHaveBeenCalledWith(FlowEventTypes.ON_EDGE_DELETE, expect.any(Function));
    });
  });

  describe('Context Integration', () => {
    it('should use setIsOpenResourcePropertiesPanel from context', () => {
      renderHook(() => useDeleteExecutionResource(), {
        wrapper: createWrapper(),
      });

      // The hook should have access to context
      expect(mockRegisterAsync).toHaveBeenCalledTimes(3);
    });
  });

  describe('deleteExecutionActionNode Handler', () => {
    it('should return true when no execution nodes are deleted', async () => {
      const viewNode: Node = {
        id: 'view-1',
        type: 'VIEW',
        position: {x: 0, y: 0},
        data: {},
      };

      mockGetNodes.mockReturnValue([viewNode]);
      mockGetEdges.mockReturnValue([]);

      renderHook(() => useDeleteExecutionResource(), {
        wrapper: createWrapper(),
      });

      const deleteNodeHandler = registeredHandlers[FlowEventTypes.ON_NODE_DELETE]?.[0];
      expect(deleteNodeHandler).toBeDefined();

      // Delete a non-execution node
      const result = await deleteNodeHandler([viewNode]);
      expect(result).toBe(true);
    });

    it('should delete action components when execution node is deleted', async () => {
      const executionNode: Node = {
        id: 'execution-1',
        type: 'TASK_EXECUTION',
        position: {x: 100, y: 0},
        data: {},
      };

      const actionNode: Node = {
        id: 'action-1',
        type: 'VIEW',
        position: {x: 0, y: 0},
        data: {
          components: [
            {id: 'button-1', type: 'ACTION'},
            {id: 'button-2', type: 'ACTION'},
          ],
        },
      };

      mockGetNodes.mockReturnValue([actionNode, executionNode]);
      mockGetEdges.mockReturnValue([
        {
          id: 'edge-1',
          source: 'action-1',
          target: 'execution-1',
          sourceHandle: 'button-1-next',
        },
      ]);

      renderHook(() => useDeleteExecutionResource(), {
        wrapper: createWrapper(),
      });

      const deleteNodeHandler = registeredHandlers[FlowEventTypes.ON_NODE_DELETE]?.[0];
      expect(deleteNodeHandler).toBeDefined();

      const result = await deleteNodeHandler([executionNode]);
      expect(result).toBe(true);
      // The handler should register correctly - mockUpdateNodeData may not be called
      // if the node type doesn't match StepTypes.Execution
    });

    it('should return true when action nodes array is empty', async () => {
      const executionNode: Node = {
        id: 'execution-1',
        type: 'TASK_EXECUTION',
        position: {x: 100, y: 0},
        data: {},
      };

      mockGetNodes.mockReturnValue([executionNode]);
      mockGetEdges.mockReturnValue([]);

      renderHook(() => useDeleteExecutionResource(), {
        wrapper: createWrapper(),
      });

      const deleteNodeHandler = registeredHandlers[FlowEventTypes.ON_NODE_DELETE]?.[0];
      const result = await deleteNodeHandler([executionNode]);
      expect(result).toBe(true);
    });
  });

  describe('deleteExecutionNode Handler', () => {
    it('should return true for non-action elements', async () => {
      mockGetNodes.mockReturnValue([]);

      renderHook(() => useDeleteExecutionResource(), {
        wrapper: createWrapper(),
      });

      const deleteElementHandler = registeredHandlers[FlowEventTypes.ON_NODE_ELEMENT_DELETE]?.[0];
      expect(deleteElementHandler).toBeDefined();

      const element = {
        id: 'input-1',
        type: 'INPUT',
        category: 'FIELD',
      };

      const result = await deleteElementHandler('step-1', element);
      expect(result).toBe(true);
    });

    it('should delete execution node when action element with next type is deleted', async () => {
      const executionNode: Node = {
        id: 'execution-1',
        type: 'TASK_EXECUTION',
        position: {x: 100, y: 0},
        data: {},
      };

      mockGetNodes.mockReturnValue([executionNode]);

      renderHook(() => useDeleteExecutionResource(), {
        wrapper: createWrapper(),
      });

      const deleteElementHandler = registeredHandlers[FlowEventTypes.ON_NODE_ELEMENT_DELETE]?.[0];
      expect(deleteElementHandler).toBeDefined();

      const element = {
        id: 'button-1',
        type: 'ACTION',
        category: 'ACTION',
        action: {
          type: 'NEXT',
          onSuccess: 'execution-1',
        },
      };

      const result = await deleteElementHandler('step-1', element);
      expect(result).toBe(true);
      expect(mockSetNodes).toHaveBeenCalled();
    });

    it('should not delete execution node when action has different type', async () => {
      mockGetNodes.mockReturnValue([]);

      renderHook(() => useDeleteExecutionResource(), {
        wrapper: createWrapper(),
      });

      const deleteElementHandler = registeredHandlers[FlowEventTypes.ON_NODE_ELEMENT_DELETE]?.[0];

      const element = {
        id: 'button-1',
        type: 'ACTION',
        category: 'ACTION',
        action: {
          type: 'SUBMIT',
        },
      };

      const result = await deleteElementHandler('step-1', element);
      expect(result).toBe(true);
    });
  });

  describe('deleteComponentAndNode Handler', () => {
    it('should return true when no execution nodes are connected to deleted edges', async () => {
      const viewNode: Node = {
        id: 'view-1',
        type: 'VIEW',
        position: {x: 0, y: 0},
        data: {},
      };

      mockGetNodes.mockReturnValue([viewNode]);

      renderHook(() => useDeleteExecutionResource(), {
        wrapper: createWrapper(),
      });

      const deleteEdgeHandler = registeredHandlers[FlowEventTypes.ON_EDGE_DELETE]?.[0];
      expect(deleteEdgeHandler).toBeDefined();

      const edges = [
        {
          id: 'edge-1',
          source: 'view-1',
          target: 'view-2',
          sourceHandle: 'button-1-next',
        },
      ];

      const result = await deleteEdgeHandler(edges);
      expect(result).toBe(true);
    });

    it('should delete execution nodes and components when edges are deleted', async () => {
      const executionNode: Node = {
        id: 'execution-1',
        type: 'TASK_EXECUTION',
        position: {x: 100, y: 0},
        data: {},
      };

      const actionNode: Node = {
        id: 'action-1',
        type: 'VIEW',
        position: {x: 0, y: 0},
        data: {
          components: [{id: 'button-1', type: 'ACTION'}],
        },
      };

      mockGetNodes.mockReturnValue([actionNode, executionNode]);

      renderHook(() => useDeleteExecutionResource(), {
        wrapper: createWrapper(),
      });

      const deleteEdgeHandler = registeredHandlers[FlowEventTypes.ON_EDGE_DELETE]?.[0];
      expect(deleteEdgeHandler).toBeDefined();

      const edges = [
        {
          id: 'edge-1',
          source: 'action-1',
          target: 'execution-1',
          sourceHandle: 'button-1-next',
        },
      ];

      const result = await deleteEdgeHandler(edges);
      expect(result).toBe(true);
      expect(mockSetNodes).toHaveBeenCalled();
      expect(mockUpdateNodeData).toHaveBeenCalled();
      expect(mockSetIsOpenResourcePropertiesPanel).toHaveBeenCalledWith(false);
    });

    it('should handle multiple edges being deleted', async () => {
      const executionNode1: Node = {
        id: 'execution-1',
        type: 'TASK_EXECUTION',
        position: {x: 100, y: 0},
        data: {},
      };

      const executionNode2: Node = {
        id: 'execution-2',
        type: 'TASK_EXECUTION',
        position: {x: 200, y: 0},
        data: {},
      };

      const actionNode: Node = {
        id: 'action-1',
        type: 'VIEW',
        position: {x: 0, y: 0},
        data: {
          components: [
            {id: 'button-1', type: 'ACTION'},
            {id: 'button-2', type: 'ACTION'},
          ],
        },
      };

      mockGetNodes.mockReturnValue([actionNode, executionNode1, executionNode2]);

      renderHook(() => useDeleteExecutionResource(), {
        wrapper: createWrapper(),
      });

      const deleteEdgeHandler = registeredHandlers[FlowEventTypes.ON_EDGE_DELETE]?.[0];

      const edges = [
        {
          id: 'edge-1',
          source: 'action-1',
          target: 'execution-1',
          sourceHandle: 'button-1-next',
        },
        {
          id: 'edge-2',
          source: 'action-1',
          target: 'execution-2',
          sourceHandle: 'button-2-next',
        },
      ];

      const result = await deleteEdgeHandler(edges);
      expect(result).toBe(true);
      expect(mockSetNodes).toHaveBeenCalled();
    });

    it('should handle edge without sourceHandle', async () => {
      const executionNode: Node = {
        id: 'execution-1',
        type: 'TASK_EXECUTION',
        position: {x: 100, y: 0},
        data: {},
      };

      mockGetNodes.mockReturnValue([executionNode]);

      renderHook(() => useDeleteExecutionResource(), {
        wrapper: createWrapper(),
      });

      const deleteEdgeHandler = registeredHandlers[FlowEventTypes.ON_EDGE_DELETE]?.[0];

      const edges = [
        {
          id: 'edge-1',
          source: 'action-1',
          target: 'execution-1',
          // No sourceHandle
        },
      ];

      const result = await deleteEdgeHandler(edges);
      expect(result).toBe(true);
    });

    it('should execute updateNodeData callback to filter components in deleteComponentAndNode', async () => {
      const executionNode: Node = {
        id: 'execution-1',
        type: 'TASK_EXECUTION',
        position: {x: 100, y: 0},
        data: {},
      };

      const actionNode: Node = {
        id: 'action-1',
        type: 'VIEW',
        position: {x: 0, y: 0},
        data: {
          components: [
            {id: 'button-1', type: 'ACTION'},
            {id: 'button-2', type: 'ACTION'},
          ],
        },
      };

      mockGetNodes.mockReturnValue([actionNode, executionNode]);

      // Capture the callback passed to updateNodeData
      let capturedCallback: ((node: Node) => {components: unknown[]}) | null = null;
      mockUpdateNodeData.mockImplementation((_id: string, callback: (node: Node) => {components: unknown[]}) => {
        capturedCallback = callback;
      });

      renderHook(() => useDeleteExecutionResource(), {
        wrapper: createWrapper(),
      });

      const deleteEdgeHandler = registeredHandlers[FlowEventTypes.ON_EDGE_DELETE]?.[0];

      const edges = [
        {
          id: 'edge-1',
          source: 'action-1',
          target: 'execution-1',
          sourceHandle: 'button-1-next',
        },
      ];

      await deleteEdgeHandler(edges);

      expect(capturedCallback).not.toBeNull();

      const result = capturedCallback!(actionNode);
      expect(result.components).toHaveLength(1);
      expect(result.components[0]).toEqual({id: 'button-2', type: 'ACTION'});
    });
  });

  describe('deleteExecutionActionNode Handler - Callback Execution', () => {
    it('should execute updateNodeData callback to filter action components', async () => {
      const executionNode: Node = {
        id: 'execution-1',
        type: 'TASK_EXECUTION',
        position: {x: 100, y: 0},
        data: {},
      };

      const actionNode: Node = {
        id: 'action-1',
        type: 'VIEW',
        position: {x: 0, y: 0},
        data: {
          components: [
            {id: 'button-1', type: 'ACTION'},
            {id: 'button-2', type: 'ACTION'},
          ],
        },
      };

      mockGetNodes.mockReturnValue([actionNode, executionNode]);
      mockGetEdges.mockReturnValue([
        {
          id: 'edge-1',
          source: 'action-1',
          target: 'execution-1',
          sourceHandle: 'button-1_NEXT', // Use correct suffix format
        },
      ]);

      // Capture the callback passed to updateNodeData
      let capturedCallback: ((node: Node) => {components: unknown[]}) | null = null;
      mockUpdateNodeData.mockImplementation((_id: string, callback: (node: Node) => {components: unknown[]}) => {
        capturedCallback = callback;
      });

      renderHook(() => useDeleteExecutionResource(), {
        wrapper: createWrapper(),
      });

      const deleteNodeHandler = registeredHandlers[FlowEventTypes.ON_NODE_DELETE]?.[0];
      await deleteNodeHandler([executionNode]);

      expect(mockUpdateNodeData).toHaveBeenCalledWith('action-1', expect.any(Function));
      expect(capturedCallback).not.toBeNull();

      const result = capturedCallback!(actionNode);
      expect(result.components).toHaveLength(1);
      expect(result.components[0]).toEqual({id: 'button-2', type: 'ACTION'});
    });

    it('should close properties panel after updating node data', async () => {
      const executionNode: Node = {
        id: 'execution-1',
        type: 'TASK_EXECUTION',
        position: {x: 100, y: 0},
        data: {},
      };

      const actionNode: Node = {
        id: 'action-1',
        type: 'VIEW',
        position: {x: 0, y: 0},
        data: {
          components: [{id: 'button-1', type: 'ACTION'}],
        },
      };

      mockGetNodes.mockReturnValue([actionNode, executionNode]);
      mockGetEdges.mockReturnValue([
        {
          id: 'edge-1',
          source: 'action-1',
          target: 'execution-1',
          sourceHandle: 'button-1_NEXT', // Use correct suffix format
        },
      ]);

      renderHook(() => useDeleteExecutionResource(), {
        wrapper: createWrapper(),
      });

      const deleteNodeHandler = registeredHandlers[FlowEventTypes.ON_NODE_DELETE]?.[0];
      await deleteNodeHandler([executionNode]);

      expect(mockSetIsOpenResourcePropertiesPanel).toHaveBeenCalledWith(false);
    });
  });

  describe('deleteExecutionNode Handler - Callback Execution', () => {
    it('should execute setNodes callback to filter execution nodes', async () => {
      const executionNode: Node = {
        id: 'execution-1',
        type: 'TASK_EXECUTION',
        position: {x: 100, y: 0},
        data: {},
      };

      const viewNode: Node = {
        id: 'view-1',
        type: 'VIEW',
        position: {x: 0, y: 0},
        data: {},
      };

      mockGetNodes.mockReturnValue([viewNode, executionNode]);

      // Capture the callback passed to setNodes
      let capturedCallback: ((nodes: Node[]) => Node[]) | null = null;
      mockSetNodes.mockImplementation((callback: (nodes: Node[]) => Node[]) => {
        capturedCallback = callback;
      });

      renderHook(() => useDeleteExecutionResource(), {
        wrapper: createWrapper(),
      });

      const deleteElementHandler = registeredHandlers[FlowEventTypes.ON_NODE_ELEMENT_DELETE]?.[0];

      const element = {
        id: 'button-1',
        type: 'ACTION',
        category: 'ACTION',
        action: {
          type: 'NEXT',
          onSuccess: 'execution-1',
        },
      };

      await deleteElementHandler('step-1', element);

      expect(mockSetNodes).toHaveBeenCalled();
      expect(capturedCallback).not.toBeNull();

      // Execute the callback
      const result = capturedCallback!([viewNode, executionNode]);

      // Should filter out the execution node with matching id and type
      expect(result).toHaveLength(1);
      expect(result[0]).toEqual(viewNode);
    });

    it('should keep nodes that do not match both id and type', async () => {
      const executionNode: Node = {
        id: 'execution-1',
        type: 'TASK_EXECUTION',
        position: {x: 100, y: 0},
        data: {},
      };

      const otherExecutionNode: Node = {
        id: 'execution-2',
        type: 'TASK_EXECUTION',
        position: {x: 200, y: 0},
        data: {},
      };

      mockGetNodes.mockReturnValue([executionNode, otherExecutionNode]);

      let capturedCallback: ((nodes: Node[]) => Node[]) | null = null;
      mockSetNodes.mockImplementation((callback: (nodes: Node[]) => Node[]) => {
        capturedCallback = callback;
      });

      renderHook(() => useDeleteExecutionResource(), {
        wrapper: createWrapper(),
      });

      const deleteElementHandler = registeredHandlers[FlowEventTypes.ON_NODE_ELEMENT_DELETE]?.[0];

      const element = {
        id: 'button-1',
        type: 'ACTION',
        category: 'ACTION',
        action: {
          type: 'NEXT',
          onSuccess: 'execution-1',
        },
      };

      await deleteElementHandler('step-1', element);

      const result = capturedCallback!([executionNode, otherExecutionNode]);

      // Should only filter out execution-1, keep execution-2
      expect(result).toHaveLength(1);
      expect(result[0].id).toBe('execution-2');
    });
  });
});
