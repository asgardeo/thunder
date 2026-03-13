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
import useStaticContentField from '../useStaticContentField';
import FlowEventTypes from '../../models/extension';
import {StepTypes, ExecutionTypes} from '../../models/steps';
import {ElementTypes} from '../../models/elements';

// Use vi.hoisted to define mocks that need to be referenced in vi.mock
const {mockGetNode, mockUpdateNodeData, mockRegisterAsync, mockRegisterSync, mockUnregister} = vi.hoisted(() => ({
  mockGetNode: vi.fn(),
  mockUpdateNodeData: vi.fn(),
  mockRegisterAsync: vi.fn(),
  mockRegisterSync: vi.fn(),
  mockUnregister: vi.fn(),
}));

// Store registered handlers for testing
const registeredHandlers: {
  async: Record<string, ((...args: unknown[]) => Promise<boolean>)[]>;
  sync: Record<string, ((...args: unknown[]) => boolean)[]>;
} = {
  async: {},
  sync: {},
};

// Mock @xyflow/react
vi.mock('@xyflow/react', async () => {
  const actual = await vi.importActual('@xyflow/react');
  return {
    ...actual,
    useReactFlow: () => ({
      getNode: mockGetNode,
      updateNodeData: mockUpdateNodeData,
    }),
  };
});

// Mock useGetFlowBuilderCoreResources
vi.mock('../../api/useGetFlowBuilderCoreResources', () => ({
  default: () => ({
    data: {
      elements: [
        {
          id: 'rich-text-element',
          type: ElementTypes.RichText,
          config: {text: ''},
        },
      ],
    },
    isLoading: false,
    error: null,
  }),
}));

// Mock PluginRegistry - capture handlers for testing
vi.mock('../../plugins/PluginRegistry', () => ({
  default: {
    getInstance: () => ({
      registerAsync: (eventType: string, handler: (...args: unknown[]) => Promise<boolean>) => {
        mockRegisterAsync(eventType, handler);
        if (!registeredHandlers.async[eventType]) {
          registeredHandlers.async[eventType] = [];
        }
        registeredHandlers.async[eventType].push(handler);
      },
      registerSync: (eventType: string, handler: (...args: unknown[]) => boolean) => {
        mockRegisterSync(eventType, handler);
        if (!registeredHandlers.sync[eventType]) {
          registeredHandlers.sync[eventType] = [];
        }
        registeredHandlers.sync[eventType].push(handler);
      },
      unregister: mockUnregister,
    }),
  },
}));

// Mock generateResourceId
vi.mock('../../utils/generateResourceId', () => ({
  default: vi.fn().mockReturnValue('generated-id'),
}));

describe('useStaticContentField', () => {
  const createWrapper = () => {
    function Wrapper({children}: {children: ReactNode}) {
      return <ReactFlowProvider>{children}</ReactFlowProvider>;
    }
    return Wrapper;
  };

  beforeEach(() => {
    vi.clearAllMocks();
    // Clear registered handlers
    Object.keys(registeredHandlers.async).forEach((key) => {
      delete registeredHandlers.async[key];
    });
    Object.keys(registeredHandlers.sync).forEach((key) => {
      delete registeredHandlers.sync[key];
    });
  });

  describe('Plugin Registration', () => {
    it('should register event handlers on mount', () => {
      renderHook(() => useStaticContentField(), {
        wrapper: createWrapper(),
      });

      expect(mockRegisterAsync).toHaveBeenCalledWith(FlowEventTypes.ON_PROPERTY_CHANGE, expect.any(Function));
      expect(mockRegisterSync).toHaveBeenCalledWith(FlowEventTypes.ON_PROPERTY_PANEL_OPEN, expect.any(Function));
    });

    it('should unregister event handlers on unmount', () => {
      const {unmount} = renderHook(() => useStaticContentField(), {
        wrapper: createWrapper(),
      });

      unmount();

      expect(mockUnregister).toHaveBeenCalledWith(FlowEventTypes.ON_PROPERTY_CHANGE, 'addStaticContent');
      expect(mockUnregister).toHaveBeenCalledWith(FlowEventTypes.ON_PROPERTY_PANEL_OPEN, 'addStaticContentProperties');
    });
  });

  describe('addStaticContent Handler', () => {
    it('should return true for non-execution step types', async () => {
      renderHook(() => useStaticContentField(), {
        wrapper: createWrapper(),
      });

      const addStaticContentHandler = registeredHandlers.async[FlowEventTypes.ON_PROPERTY_CHANGE]?.[0];
      expect(addStaticContentHandler).toBeDefined();

      const element = {
        id: 'view-1',
        type: StepTypes.View,
      };

      const result = await addStaticContentHandler('enableStaticContent', true, element, 'step-1');
      expect(result).toBe(true);
    });

    it('should return true for properties other than enableStaticContent', async () => {
      renderHook(() => useStaticContentField(), {
        wrapper: createWrapper(),
      });

      const addStaticContentHandler = registeredHandlers.async[FlowEventTypes.ON_PROPERTY_CHANGE]?.[0];

      const element = {
        id: 'execution-1',
        type: StepTypes.Execution,
      };

      const result = await addStaticContentHandler('someOtherProperty', true, element, 'step-1');
      expect(result).toBe(true);
    });

    it('should add static content when enableStaticContent is true', async () => {
      const executionNode: Node = {
        id: 'execution-1',
        type: StepTypes.Execution,
        position: {x: 0, y: 0},
        data: {
          components: [],
        },
      };

      mockGetNode.mockReturnValue(executionNode);

      renderHook(() => useStaticContentField(), {
        wrapper: createWrapper(),
      });

      const addStaticContentHandler = registeredHandlers.async[FlowEventTypes.ON_PROPERTY_CHANGE]?.[0];

      const element = {
        id: 'execution-1',
        type: StepTypes.Execution,
      };

      const result = await addStaticContentHandler('enableStaticContent', true, element, 'execution-1');
      expect(result).toBe(false);
      expect(mockUpdateNodeData).toHaveBeenCalled();
    });

    it('should remove static content when enableStaticContent is false', async () => {
      const executionNode: Node = {
        id: 'execution-1',
        type: StepTypes.Execution,
        position: {x: 0, y: 0},
        data: {
          components: [{id: 'rich-text-1', type: ElementTypes.RichText}],
        },
      };

      mockGetNode.mockReturnValue(executionNode);

      renderHook(() => useStaticContentField(), {
        wrapper: createWrapper(),
      });

      const addStaticContentHandler = registeredHandlers.async[FlowEventTypes.ON_PROPERTY_CHANGE]?.[0];

      const element = {
        id: 'execution-1',
        type: StepTypes.Execution,
      };

      const result = await addStaticContentHandler('enableStaticContent', false, element, 'execution-1');
      expect(result).toBe(false);
      expect(mockUpdateNodeData).toHaveBeenCalled();
    });

    it('should call updateNodeData callback correctly when adding content', async () => {
      let capturedCallback: ((node: Node) => Record<string, unknown>) | null = null;
      mockUpdateNodeData.mockImplementation(
        (_stepId: string, callback: (node: Node) => Record<string, unknown>) => {
          capturedCallback = callback;
        },
      );

      renderHook(() => useStaticContentField(), {
        wrapper: createWrapper(),
      });

      const addStaticContentHandler = registeredHandlers.async[FlowEventTypes.ON_PROPERTY_CHANGE]?.[0];

      const element = {
        id: 'execution-1',
        type: StepTypes.Execution,
      };

      await addStaticContentHandler('enableStaticContent', true, element, 'execution-1');

      // Execute the captured callback
      const mockNode: Node = {
        id: 'execution-1',
        type: StepTypes.Execution,
        position: {x: 0, y: 0},
        data: {components: []},
      };

      expect(capturedCallback).not.toBeNull();
      const result = capturedCallback!(mockNode);
      expect(result.components).toBeDefined();
    });

    it('should call updateNodeData callback correctly when removing content', async () => {
      let capturedCallback: ((node: Node) => Record<string, unknown>) | null = null;
      mockUpdateNodeData.mockImplementation(
        (_stepId: string, callback: (node: Node) => Record<string, unknown>) => {
          capturedCallback = callback;
        },
      );

      renderHook(() => useStaticContentField(), {
        wrapper: createWrapper(),
      });

      const addStaticContentHandler = registeredHandlers.async[FlowEventTypes.ON_PROPERTY_CHANGE]?.[0];

      const element = {
        id: 'execution-1',
        type: StepTypes.Execution,
      };

      await addStaticContentHandler('enableStaticContent', false, element, 'execution-1');

      // Execute the captured callback
      const mockNode: Node = {
        id: 'execution-1',
        type: StepTypes.Execution,
        position: {x: 0, y: 0},
        data: {
          components: [{id: 'rich-text-1', type: ElementTypes.RichText}],
        },
      };

      expect(capturedCallback).not.toBeNull();
      const result = capturedCallback!(mockNode);
      expect(result.components).toEqual([]);
    });
  });

  describe('addStaticContentProperties Handler', () => {
    it('should return true when node is not found', () => {
      mockGetNode.mockReturnValue(undefined);

      renderHook(() => useStaticContentField(), {
        wrapper: createWrapper(),
      });

      const addPropertiesHandler = registeredHandlers.sync[FlowEventTypes.ON_PROPERTY_PANEL_OPEN]?.[0];
      expect(addPropertiesHandler).toBeDefined();

      const resource = {
        id: 'execution-1',
        type: StepTypes.Execution,
        data: {
          action: {
            executor: {name: ExecutionTypes.SendEmailOTP},
          },
        },
      };

      const properties: Record<string, unknown> = {};
      const result = addPropertiesHandler(resource, properties, 'step-1');
      expect(result).toBe(true);
      expect(properties.enableStaticContent).toBeUndefined();
    });

    it('should return true for non-execution step types', () => {
      const viewNode: Node = {
        id: 'view-1',
        type: StepTypes.View,
        position: {x: 0, y: 0},
        data: {},
      };

      mockGetNode.mockReturnValue(viewNode);

      renderHook(() => useStaticContentField(), {
        wrapper: createWrapper(),
      });

      const addPropertiesHandler = registeredHandlers.sync[FlowEventTypes.ON_PROPERTY_PANEL_OPEN]?.[0];

      const resource = {
        id: 'view-1',
        type: StepTypes.View,
      };

      const properties: Record<string, unknown> = {};
      const result = addPropertiesHandler(resource, properties, 'view-1');
      expect(result).toBe(true);
      expect(properties.enableStaticContent).toBeUndefined();
    });

    it('should not add enableStaticContent property for non-allowed execution types', () => {
      const executionNode: Node = {
        id: 'execution-1',
        type: StepTypes.Execution,
        position: {x: 0, y: 0},
        data: {
          components: [{id: 'rich-text-1', type: ElementTypes.RichText}],
        },
      };

      mockGetNode.mockReturnValue(executionNode);

      renderHook(() => useStaticContentField(), {
        wrapper: createWrapper(),
      });

      const addPropertiesHandler = registeredHandlers.sync[FlowEventTypes.ON_PROPERTY_PANEL_OPEN]?.[0];

      // SendEmailOTP is not in the allowed types list
      const resource = {
        id: 'execution-1',
        type: StepTypes.Execution,
        data: {
          action: {
            executor: {name: ExecutionTypes.SendEmailOTP},
          },
        },
      };

      const properties: Record<string, unknown> = {};
      const result = addPropertiesHandler(resource, properties, 'execution-1');
      expect(result).toBe(true);
      // Property not set because SendEmailOTP is not in allowed types
      expect(properties.enableStaticContent).toBeUndefined();
    });

    it('should not add property for non-allowed execution types even without components', () => {
      const executionNode: Node = {
        id: 'execution-1',
        type: StepTypes.Execution,
        position: {x: 0, y: 0},
        data: {
          components: [],
        },
      };

      mockGetNode.mockReturnValue(executionNode);

      renderHook(() => useStaticContentField(), {
        wrapper: createWrapper(),
      });

      const addPropertiesHandler = registeredHandlers.sync[FlowEventTypes.ON_PROPERTY_PANEL_OPEN]?.[0];

      // SendEmailOTP is not in the allowed types list
      const resource = {
        id: 'execution-1',
        type: StepTypes.Execution,
        data: {
          action: {
            executor: {name: ExecutionTypes.SendEmailOTP},
          },
        },
      };

      const properties: Record<string, unknown> = {};
      const result = addPropertiesHandler(resource, properties, 'execution-1');
      expect(result).toBe(true);
      // Property not set because SendEmailOTP is not in allowed types
      expect(properties.enableStaticContent).toBeUndefined();
    });

    it('should return true for MagicLinkExecutor without adding property', () => {
      const executionNode: Node = {
        id: 'execution-1',
        type: StepTypes.Execution,
        position: {x: 0, y: 0},
        data: {
          components: [],
        },
      };

      mockGetNode.mockReturnValue(executionNode);

      renderHook(() => useStaticContentField(), {
        wrapper: createWrapper(),
      });

      const addPropertiesHandler = registeredHandlers.sync[FlowEventTypes.ON_PROPERTY_PANEL_OPEN]?.[0];

      const resource = {
        id: 'execution-1',
        type: StepTypes.Execution,
        data: {
          action: {
            executor: {name: ExecutionTypes.MagicLinkExecutor},
          },
        },
      };

      const properties: Record<string, unknown> = {};
      const result = addPropertiesHandler(resource, properties, 'execution-1');
      expect(result).toBe(true);
      expect(properties.enableStaticContent).toBeUndefined();
    });

    it('should return true when executor name is not in allowed types', () => {
      const executionNode: Node = {
        id: 'execution-1',
        type: StepTypes.Execution,
        position: {x: 0, y: 0},
        data: {
          components: [],
        },
      };

      mockGetNode.mockReturnValue(executionNode);

      renderHook(() => useStaticContentField(), {
        wrapper: createWrapper(),
      });

      const addPropertiesHandler = registeredHandlers.sync[FlowEventTypes.ON_PROPERTY_PANEL_OPEN]?.[0];

      const resource = {
        id: 'execution-1',
        type: StepTypes.Execution,
        data: {
          action: {
            executor: {name: 'SomeOtherExecutor'},
          },
        },
      };

      const properties: Record<string, unknown> = {};
      const result = addPropertiesHandler(resource, properties, 'execution-1');
      expect(result).toBe(true);
    });

    it('should return true when executor is undefined', () => {
      const executionNode: Node = {
        id: 'execution-1',
        type: StepTypes.Execution,
        position: {x: 0, y: 0},
        data: {
          components: [],
        },
      };

      mockGetNode.mockReturnValue(executionNode);

      renderHook(() => useStaticContentField(), {
        wrapper: createWrapper(),
      });

      const addPropertiesHandler = registeredHandlers.sync[FlowEventTypes.ON_PROPERTY_PANEL_OPEN]?.[0];

      const resource = {
        id: 'execution-1',
        type: StepTypes.Execution,
        data: {},
      };

      const properties: Record<string, unknown> = {};
      const result = addPropertiesHandler(resource, properties, 'execution-1');
      expect(result).toBe(true);
    });
  });
});
