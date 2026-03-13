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
import useConfirmPasswordField from '../useConfirmPasswordField';
import FlowEventTypes from '../../models/extension';
import {ElementTypes, BlockTypes} from '../../models/elements';
import FlowBuilderElementConstants from '../../constants/FlowBuilderElementConstants';
import type {Element} from '../../models/elements';

// Use vi.hoisted to define mocks
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
  default: vi.fn().mockReturnValue('generated-field-id'),
}));

describe('useConfirmPasswordField', () => {
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
      renderHook(() => useConfirmPasswordField(), {
        wrapper: createWrapper(),
      });

      // Should register two async handlers for ON_PROPERTY_CHANGE
      expect(mockRegisterAsync).toHaveBeenCalledWith(FlowEventTypes.ON_PROPERTY_CHANGE, expect.any(Function));
      expect(mockRegisterSync).toHaveBeenCalledWith(FlowEventTypes.ON_PROPERTY_PANEL_OPEN, expect.any(Function));
      expect(mockRegisterAsync).toHaveBeenCalledWith(FlowEventTypes.ON_NODE_ELEMENT_DELETE, expect.any(Function));
    });

    it('should unregister event handlers on unmount', () => {
      const {unmount} = renderHook(() => useConfirmPasswordField(), {
        wrapper: createWrapper(),
      });

      unmount();

      expect(mockUnregister).toHaveBeenCalledWith(FlowEventTypes.ON_PROPERTY_CHANGE, 'addConfirmPasswordField');
      expect(mockUnregister).toHaveBeenCalledWith(
        FlowEventTypes.ON_PROPERTY_CHANGE,
        'updateConfirmPasswordFieldProperties',
      );
      expect(mockUnregister).toHaveBeenCalledWith(
        FlowEventTypes.ON_PROPERTY_PANEL_OPEN,
        'addConfirmPasswordFieldProperties',
      );
      expect(mockUnregister).toHaveBeenCalledWith(FlowEventTypes.ON_NODE_ELEMENT_DELETE, 'deleteConfirmPasswordField');
    });
  });

  describe('addConfirmPasswordField Handler', () => {
    it('should return true for non-password input types', async () => {
      renderHook(() => useConfirmPasswordField(), {
        wrapper: createWrapper(),
      });

      const handlers = registeredHandlers.async[FlowEventTypes.ON_PROPERTY_CHANGE];
      const addConfirmPasswordFieldHandler = handlers?.find(
        // eslint-disable-next-line no-underscore-dangle
        (h) => (h as unknown as Record<string, string>).__FLOW_BUILDER_PLUGIN_FUNCTION_IDENTIFIER === 'addConfirmPasswordField' ||
               handlers.indexOf(h) === 0,
      );

      const element = {
        id: 'text-input-1',
        type: ElementTypes.TextInput,
      } as Element;

      const result = await addConfirmPasswordFieldHandler!('requireConfirmation', true, element, 'step-1');
      expect(result).toBe(true);
    });

    it('should return true for properties other than requireConfirmation', async () => {
      renderHook(() => useConfirmPasswordField(), {
        wrapper: createWrapper(),
      });

      const handlers = registeredHandlers.async[FlowEventTypes.ON_PROPERTY_CHANGE];
      const addConfirmPasswordFieldHandler = handlers?.[0];

      const element = {
        id: 'password-1',
        type: ElementTypes.PasswordInput,
      } as Element;

      const result = await addConfirmPasswordFieldHandler('someOtherProperty', true, element, 'step-1');
      expect(result).toBe(true);
    });

    it('should add confirm password field when requireConfirmation is true', async () => {
      const stepNode: Node = {
        id: 'step-1',
        type: 'VIEW',
        position: {x: 0, y: 0},
        data: {
          components: [
            {
              id: 'form-1',
              type: BlockTypes.Form,
              components: [
                {id: 'password-1', type: ElementTypes.PasswordInput},
              ],
            },
          ],
        },
      };

      mockGetNode.mockReturnValue(stepNode);

      renderHook(() => useConfirmPasswordField(), {
        wrapper: createWrapper(),
      });

      const handlers = registeredHandlers.async[FlowEventTypes.ON_PROPERTY_CHANGE];
      const addConfirmPasswordFieldHandler = handlers?.[0];

      const element = {
        id: 'password-1',
        type: ElementTypes.PasswordInput,
      } as Element;

      const result = await addConfirmPasswordFieldHandler('requireConfirmation', true, element, 'step-1');
      expect(result).toBe(false);
      expect(mockUpdateNodeData).toHaveBeenCalled();
    });

    it('should remove confirm password field when requireConfirmation is false', async () => {
      const stepNode: Node = {
        id: 'step-1',
        type: 'VIEW',
        position: {x: 0, y: 0},
        data: {
          components: [
            {
              id: 'form-1',
              type: BlockTypes.Form,
              components: [
                {id: 'password-1', type: ElementTypes.PasswordInput},
                {
                  id: 'confirm-password-1',
                  type: ElementTypes.PasswordInput,
                  identifier: FlowBuilderElementConstants.CONFIRM_PASSWORD_IDENTIFIER,
                },
              ],
            },
          ],
        },
      };

      mockGetNode.mockReturnValue(stepNode);

      renderHook(() => useConfirmPasswordField(), {
        wrapper: createWrapper(),
      });

      const handlers = registeredHandlers.async[FlowEventTypes.ON_PROPERTY_CHANGE];
      const addConfirmPasswordFieldHandler = handlers?.[0];

      const element = {
        id: 'password-1',
        type: ElementTypes.PasswordInput,
      } as Element;

      const result = await addConfirmPasswordFieldHandler('requireConfirmation', false, element, 'step-1');
      expect(result).toBe(false);
      expect(mockUpdateNodeData).toHaveBeenCalled();
    });

    it('should execute updateNodeData callback correctly when adding confirm field', async () => {
      let capturedCallback: ((node: Node) => Record<string, unknown>) | null = null;
      mockUpdateNodeData.mockImplementation(
        (_stepId: string, callback: (node: Node) => Record<string, unknown>) => {
          capturedCallback = callback;
        },
      );

      renderHook(() => useConfirmPasswordField(), {
        wrapper: createWrapper(),
      });

      const handlers = registeredHandlers.async[FlowEventTypes.ON_PROPERTY_CHANGE];
      const addConfirmPasswordFieldHandler = handlers?.[0];

      const element = {
        id: 'password-1',
        type: ElementTypes.PasswordInput,
      } as Element;

      await addConfirmPasswordFieldHandler('requireConfirmation', true, element, 'step-1');

      expect(capturedCallback).not.toBeNull();

      const mockNode: Node = {
        id: 'step-1',
        type: 'VIEW',
        position: {x: 0, y: 0},
        data: {
          components: [
            {
              id: 'form-1',
              type: BlockTypes.Form,
              components: [{id: 'password-1', type: ElementTypes.PasswordInput}],
            },
          ],
        },
      };

      const result = capturedCallback!(mockNode);
      expect(result.components).toBeDefined();
    });

    it('should return empty object when node has no components', async () => {
      let capturedCallback: ((node: Node) => Record<string, unknown>) | null = null;
      mockUpdateNodeData.mockImplementation(
        (_stepId: string, callback: (node: Node) => Record<string, unknown>) => {
          capturedCallback = callback;
        },
      );

      renderHook(() => useConfirmPasswordField(), {
        wrapper: createWrapper(),
      });

      const handlers = registeredHandlers.async[FlowEventTypes.ON_PROPERTY_CHANGE];
      const addConfirmPasswordFieldHandler = handlers?.[0];

      const element = {
        id: 'password-1',
        type: ElementTypes.PasswordInput,
      } as Element;

      await addConfirmPasswordFieldHandler('requireConfirmation', true, element, 'step-1');

      const mockNode: Node = {
        id: 'step-1',
        type: 'VIEW',
        position: {x: 0, y: 0},
        data: {},
      };

      const result = capturedCallback!(mockNode);
      expect(result).toEqual({});
    });

    it('should execute callback to add confirm password field after password field', async () => {
      let capturedCallback: ((node: Node) => Record<string, unknown>) | null = null;
      mockUpdateNodeData.mockImplementation(
        (_stepId: string, callback: (node: Node) => Record<string, unknown>) => {
          capturedCallback = callback;
        },
      );

      renderHook(() => useConfirmPasswordField(), {
        wrapper: createWrapper(),
      });

      const handlers = registeredHandlers.async[FlowEventTypes.ON_PROPERTY_CHANGE];
      const addConfirmPasswordFieldHandler = handlers?.[0];

      const element = {
        id: 'password-1',
        type: ElementTypes.PasswordInput,
      } as Element;

      await addConfirmPasswordFieldHandler('requireConfirmation', true, element, 'step-1');

      const mockNode: Node = {
        id: 'step-1',
        type: 'VIEW',
        position: {x: 0, y: 0},
        data: {
          components: [
            {
              id: 'form-1',
              type: BlockTypes.Form,
              components: [
                {id: 'input-1', type: 'TEXT_INPUT'},
                {id: 'password-1', type: ElementTypes.PasswordInput},
                {id: 'button-1', type: 'BUTTON'},
              ],
            },
          ],
        },
      };

      const result = capturedCallback!(mockNode);
      expect(result.components).toBeDefined();
      const components = result.components as Element[];
      const form = components[0] as Element & {components?: Element[]};
      // Confirm password should be inserted after password-1 (at index 2)
      expect(form.components?.length).toBe(4);
      expect((form.components?.[2] as Element & {identifier?: string})?.identifier).toBe(
        FlowBuilderElementConstants.CONFIRM_PASSWORD_IDENTIFIER,
      );
    });

    it('should not add confirm field when password is not in a form', async () => {
      let capturedCallback: ((node: Node) => Record<string, unknown>) | null = null;
      mockUpdateNodeData.mockImplementation(
        (_stepId: string, callback: (node: Node) => Record<string, unknown>) => {
          capturedCallback = callback;
        },
      );

      renderHook(() => useConfirmPasswordField(), {
        wrapper: createWrapper(),
      });

      const handlers = registeredHandlers.async[FlowEventTypes.ON_PROPERTY_CHANGE];
      const addConfirmPasswordFieldHandler = handlers?.[0];

      const element = {
        id: 'password-1',
        type: ElementTypes.PasswordInput,
      } as Element;

      await addConfirmPasswordFieldHandler('requireConfirmation', true, element, 'step-1');

      // Node with components that are not forms
      const mockNode: Node = {
        id: 'step-1',
        type: 'VIEW',
        position: {x: 0, y: 0},
        data: {
          components: [
            {id: 'button-1', type: 'BUTTON'},
            {id: 'password-1', type: ElementTypes.PasswordInput},
          ],
        },
      };

      const result = capturedCallback!(mockNode);
      expect(result.components).toBeDefined();
      // No confirm field should be added since password is not in a form
      expect((result.components as Element[]).length).toBe(2);
    });

    it('should execute callback to remove confirm password field when unchecking', async () => {
      let capturedCallback: ((node: Node) => Record<string, unknown>) | null = null;
      mockUpdateNodeData.mockImplementation(
        (_stepId: string, callback: (node: Node) => Record<string, unknown>) => {
          capturedCallback = callback;
        },
      );

      renderHook(() => useConfirmPasswordField(), {
        wrapper: createWrapper(),
      });

      const handlers = registeredHandlers.async[FlowEventTypes.ON_PROPERTY_CHANGE];
      const addConfirmPasswordFieldHandler = handlers?.[0];

      const element = {
        id: 'password-1',
        type: ElementTypes.PasswordInput,
      } as Element;

      await addConfirmPasswordFieldHandler('requireConfirmation', false, element, 'step-1');

      const mockNode: Node = {
        id: 'step-1',
        type: 'VIEW',
        position: {x: 0, y: 0},
        data: {
          components: [
            {
              id: 'form-1',
              type: BlockTypes.Form,
              components: [
                {id: 'password-1', type: ElementTypes.PasswordInput},
                {
                  id: 'confirm-password-1',
                  type: ElementTypes.PasswordInput,
                  identifier: FlowBuilderElementConstants.CONFIRM_PASSWORD_IDENTIFIER,
                },
              ],
            },
          ],
        },
      };

      const result = capturedCallback!(mockNode);
      expect(result.components).toBeDefined();
      const components = result.components as Element[];
      const form = components[0] as Element & {components?: Element[]};
      // Confirm password should be removed
      expect(form.components?.length).toBe(1);
      expect(form.components?.[0].id).toBe('password-1');
    });

    it('should not remove confirm field if not found', async () => {
      let capturedCallback: ((node: Node) => Record<string, unknown>) | null = null;
      mockUpdateNodeData.mockImplementation(
        (_stepId: string, callback: (node: Node) => Record<string, unknown>) => {
          capturedCallback = callback;
        },
      );

      renderHook(() => useConfirmPasswordField(), {
        wrapper: createWrapper(),
      });

      const handlers = registeredHandlers.async[FlowEventTypes.ON_PROPERTY_CHANGE];
      const addConfirmPasswordFieldHandler = handlers?.[0];

      const element = {
        id: 'password-1',
        type: ElementTypes.PasswordInput,
      } as Element;

      await addConfirmPasswordFieldHandler('requireConfirmation', false, element, 'step-1');

      const mockNode: Node = {
        id: 'step-1',
        type: 'VIEW',
        position: {x: 0, y: 0},
        data: {
          components: [
            {
              id: 'form-1',
              type: BlockTypes.Form,
              components: [
                {id: 'password-1', type: ElementTypes.PasswordInput},
                // No confirm password field
              ],
            },
          ],
        },
      };

      const result = capturedCallback!(mockNode);
      expect(result.components).toBeDefined();
      const components = result.components as Element[];
      const form = components[0] as Element & {components?: Element[]};
      // Components should remain unchanged
      expect(form.components?.length).toBe(1);
    });
  });

  describe('addConfirmPasswordFieldProperties Handler', () => {
    it('should return true for non-password input types', () => {
      renderHook(() => useConfirmPasswordField(), {
        wrapper: createWrapper(),
      });

      const addPropertiesHandler = registeredHandlers.sync[FlowEventTypes.ON_PROPERTY_PANEL_OPEN]?.[0];

      const resource = {
        id: 'text-input-1',
        type: ElementTypes.TextInput,
      } as Element;

      const properties: Record<string, unknown> = {};
      const result = addPropertiesHandler(resource, properties, 'step-1');
      expect(result).toBe(true);
      expect(properties.requireConfirmation).toBeUndefined();
    });

    it('should return true for password without PASSWORD_IDENTIFIER', () => {
      renderHook(() => useConfirmPasswordField(), {
        wrapper: createWrapper(),
      });

      const addPropertiesHandler = registeredHandlers.sync[FlowEventTypes.ON_PROPERTY_PANEL_OPEN]?.[0];

      const resource = {
        id: 'password-1',
        type: ElementTypes.PasswordInput,
        identifier: 'OTHER_IDENTIFIER',
      } as Element & {identifier?: string};

      const properties: Record<string, unknown> = {};
      const result = addPropertiesHandler(resource, properties, 'step-1');
      expect(result).toBe(true);
    });

    it('should add requireConfirmation property for password with PASSWORD_IDENTIFIER', () => {
      const stepNode: Node = {
        id: 'step-1',
        type: 'VIEW',
        position: {x: 0, y: 0},
        data: {
          components: [
            {
              id: 'form-1',
              type: BlockTypes.Form,
              components: [
                {
                  id: 'password-1',
                  type: ElementTypes.PasswordInput,
                  identifier: FlowBuilderElementConstants.PASSWORD_IDENTIFIER,
                },
                {
                  id: 'confirm-password-1',
                  type: ElementTypes.PasswordInput,
                  identifier: FlowBuilderElementConstants.CONFIRM_PASSWORD_IDENTIFIER,
                  hint: 'Confirm hint',
                  label: 'Confirm Label',
                  placeholder: 'Confirm placeholder',
                },
              ],
            },
          ],
        },
      };

      mockGetNode.mockReturnValue(stepNode);

      renderHook(() => useConfirmPasswordField(), {
        wrapper: createWrapper(),
      });

      const addPropertiesHandler = registeredHandlers.sync[FlowEventTypes.ON_PROPERTY_PANEL_OPEN]?.[0];

      const resource = {
        id: 'password-1',
        type: ElementTypes.PasswordInput,
        identifier: FlowBuilderElementConstants.PASSWORD_IDENTIFIER,
      } as Element & {identifier?: string};

      const properties: Record<string, unknown> = {};
      const result = addPropertiesHandler(resource, properties, 'step-1');
      expect(result).toBe(true);
      expect(properties.requireConfirmation).toBe(true);
      expect(properties.confirmHint).toBe('Confirm hint');
      expect(properties.confirmLabel).toBe('Confirm Label');
      expect(properties.confirmPlaceholder).toBe('Confirm placeholder');
    });

    it('should set requireConfirmation to false when no confirm field exists', () => {
      const stepNode: Node = {
        id: 'step-1',
        type: 'VIEW',
        position: {x: 0, y: 0},
        data: {
          components: [
            {
              id: 'form-1',
              type: BlockTypes.Form,
              components: [
                {
                  id: 'password-1',
                  type: ElementTypes.PasswordInput,
                  identifier: FlowBuilderElementConstants.PASSWORD_IDENTIFIER,
                },
              ],
            },
          ],
        },
      };

      mockGetNode.mockReturnValue(stepNode);

      renderHook(() => useConfirmPasswordField(), {
        wrapper: createWrapper(),
      });

      const addPropertiesHandler = registeredHandlers.sync[FlowEventTypes.ON_PROPERTY_PANEL_OPEN]?.[0];

      const resource = {
        id: 'password-1',
        type: ElementTypes.PasswordInput,
        identifier: FlowBuilderElementConstants.PASSWORD_IDENTIFIER,
      } as Element & {identifier?: string};

      const properties: Record<string, unknown> = {};
      const result = addPropertiesHandler(resource, properties, 'step-1');
      expect(result).toBe(true);
      expect(properties.requireConfirmation).toBe(false);
    });

    it('should use resource requireConfirmation value when explicitly set', () => {
      const stepNode: Node = {
        id: 'step-1',
        type: 'VIEW',
        position: {x: 0, y: 0},
        data: {
          components: [
            {
              id: 'form-1',
              type: BlockTypes.Form,
              components: [
                {
                  id: 'password-1',
                  type: ElementTypes.PasswordInput,
                  identifier: FlowBuilderElementConstants.PASSWORD_IDENTIFIER,
                },
              ],
            },
          ],
        },
      };

      mockGetNode.mockReturnValue(stepNode);

      renderHook(() => useConfirmPasswordField(), {
        wrapper: createWrapper(),
      });

      const addPropertiesHandler = registeredHandlers.sync[FlowEventTypes.ON_PROPERTY_PANEL_OPEN]?.[0];

      const resource = {
        id: 'password-1',
        type: ElementTypes.PasswordInput,
        identifier: FlowBuilderElementConstants.PASSWORD_IDENTIFIER,
        requireConfirmation: true,
      } as Element & {identifier?: string; requireConfirmation?: boolean};

      const properties: Record<string, unknown> = {};
      const result = addPropertiesHandler(resource, properties, 'step-1');
      expect(result).toBe(true);
      expect(properties.requireConfirmation).toBe(true);
    });
  });

  describe('updateConfirmPasswordFieldProperties Handler', () => {
    it('should return true for non-password input types', async () => {
      renderHook(() => useConfirmPasswordField(), {
        wrapper: createWrapper(),
      });

      const handlers = registeredHandlers.async[FlowEventTypes.ON_PROPERTY_CHANGE];
      const updatePropertiesHandler = handlers?.[1];

      const element = {
        id: 'text-input-1',
        type: ElementTypes.TextInput,
      } as Element;

      const result = await updatePropertiesHandler('confirmHint', 'New hint', element, 'step-1');
      expect(result).toBe(true);
    });

    it('should update confirm password field properties', async () => {
      renderHook(() => useConfirmPasswordField(), {
        wrapper: createWrapper(),
      });

      const handlers = registeredHandlers.async[FlowEventTypes.ON_PROPERTY_CHANGE];
      const updatePropertiesHandler = handlers?.[1];

      const element = {
        id: 'password-1',
        type: ElementTypes.PasswordInput,
      } as Element;

      const result = await updatePropertiesHandler('confirmHint', 'New hint', element, 'step-1');
      expect(result).toBe(false);
      expect(mockUpdateNodeData).toHaveBeenCalled();
    });

    it('should update confirmLabel property', async () => {
      renderHook(() => useConfirmPasswordField(), {
        wrapper: createWrapper(),
      });

      const handlers = registeredHandlers.async[FlowEventTypes.ON_PROPERTY_CHANGE];
      const updatePropertiesHandler = handlers?.[1];

      const element = {
        id: 'password-1',
        type: ElementTypes.PasswordInput,
      } as Element;

      const result = await updatePropertiesHandler('confirmLabel', 'New label', element, 'step-1');
      expect(result).toBe(false);
    });

    it('should update confirmPlaceholder property', async () => {
      renderHook(() => useConfirmPasswordField(), {
        wrapper: createWrapper(),
      });

      const handlers = registeredHandlers.async[FlowEventTypes.ON_PROPERTY_CHANGE];
      const updatePropertiesHandler = handlers?.[1];

      const element = {
        id: 'password-1',
        type: ElementTypes.PasswordInput,
      } as Element;

      const result = await updatePropertiesHandler('confirmPlaceholder', 'New placeholder', element, 'step-1');
      expect(result).toBe(false);
    });

    it('should return true for required property (not return false)', async () => {
      renderHook(() => useConfirmPasswordField(), {
        wrapper: createWrapper(),
      });

      const handlers = registeredHandlers.async[FlowEventTypes.ON_PROPERTY_CHANGE];
      const updatePropertiesHandler = handlers?.[1];

      const element = {
        id: 'password-1',
        type: ElementTypes.PasswordInput,
      } as Element;

      const result = await updatePropertiesHandler('required', true, element, 'step-1');
      expect(result).toBe(true);
      expect(mockUpdateNodeData).toHaveBeenCalled();
    });

    it('should return true for other properties', async () => {
      renderHook(() => useConfirmPasswordField(), {
        wrapper: createWrapper(),
      });

      const handlers = registeredHandlers.async[FlowEventTypes.ON_PROPERTY_CHANGE];
      const updatePropertiesHandler = handlers?.[1];

      const element = {
        id: 'password-1',
        type: ElementTypes.PasswordInput,
      } as Element;

      const result = await updatePropertiesHandler('someOtherProperty', 'value', element, 'step-1');
      expect(result).toBe(true);
    });

    it('should execute updateNodeData callback to update confirmHint on confirm field', async () => {
      let capturedCallback: ((node: Node) => Record<string, unknown>) | null = null;
      mockUpdateNodeData.mockImplementation(
        (_stepId: string, callback: (node: Node) => Record<string, unknown>) => {
          capturedCallback = callback;
        },
      );

      renderHook(() => useConfirmPasswordField(), {
        wrapper: createWrapper(),
      });

      const handlers = registeredHandlers.async[FlowEventTypes.ON_PROPERTY_CHANGE];
      const updatePropertiesHandler = handlers?.[1];

      const element = {
        id: 'password-1',
        type: ElementTypes.PasswordInput,
      } as Element;

      await updatePropertiesHandler('confirmHint', 'New hint value', element, 'step-1');

      expect(capturedCallback).not.toBeNull();

      const mockNode: Node = {
        id: 'step-1',
        type: 'VIEW',
        position: {x: 0, y: 0},
        data: {
          components: [
            {
              id: 'form-1',
              type: BlockTypes.Form,
              components: [
                {id: 'password-1', type: ElementTypes.PasswordInput},
                {
                  id: 'confirm-password-1',
                  type: ElementTypes.PasswordInput,
                  identifier: FlowBuilderElementConstants.CONFIRM_PASSWORD_IDENTIFIER,
                  hint: 'Old hint',
                },
              ],
            },
          ],
        },
      };

      const result = capturedCallback!(mockNode);
      expect(result.components).toBeDefined();
      const components = result.components as Element[];
      const form = components[0] as Element & {components?: Element[]};
      const confirmField = form.components?.find(
        (c: Element) => (c as Element & {identifier?: string}).identifier === FlowBuilderElementConstants.CONFIRM_PASSWORD_IDENTIFIER,
      ) as Element & {hint?: string};
      expect(confirmField?.hint).toBe('New hint value');
    });

    it('should execute updateNodeData callback to update confirmLabel on confirm field', async () => {
      let capturedCallback: ((node: Node) => Record<string, unknown>) | null = null;
      mockUpdateNodeData.mockImplementation(
        (_stepId: string, callback: (node: Node) => Record<string, unknown>) => {
          capturedCallback = callback;
        },
      );

      renderHook(() => useConfirmPasswordField(), {
        wrapper: createWrapper(),
      });

      const handlers = registeredHandlers.async[FlowEventTypes.ON_PROPERTY_CHANGE];
      const updatePropertiesHandler = handlers?.[1];

      const element = {
        id: 'password-1',
        type: ElementTypes.PasswordInput,
      } as Element;

      await updatePropertiesHandler('confirmLabel', 'New Label', element, 'step-1');

      const mockNode: Node = {
        id: 'step-1',
        type: 'VIEW',
        position: {x: 0, y: 0},
        data: {
          components: [
            {
              id: 'form-1',
              type: BlockTypes.Form,
              components: [
                {id: 'password-1', type: ElementTypes.PasswordInput},
                {
                  id: 'confirm-password-1',
                  type: ElementTypes.PasswordInput,
                  identifier: FlowBuilderElementConstants.CONFIRM_PASSWORD_IDENTIFIER,
                  label: 'Old Label',
                },
              ],
            },
          ],
        },
      };

      const result = capturedCallback!(mockNode);
      const components = result.components as Element[];
      const form = components[0] as Element & {components?: Element[]};
      const confirmField = form.components?.find(
        (c: Element) => (c as Element & {identifier?: string}).identifier === FlowBuilderElementConstants.CONFIRM_PASSWORD_IDENTIFIER,
      ) as Element & {label?: string};
      expect(confirmField?.label).toBe('New Label');
    });

    it('should execute updateNodeData callback to update confirmPlaceholder on confirm field', async () => {
      let capturedCallback: ((node: Node) => Record<string, unknown>) | null = null;
      mockUpdateNodeData.mockImplementation(
        (_stepId: string, callback: (node: Node) => Record<string, unknown>) => {
          capturedCallback = callback;
        },
      );

      renderHook(() => useConfirmPasswordField(), {
        wrapper: createWrapper(),
      });

      const handlers = registeredHandlers.async[FlowEventTypes.ON_PROPERTY_CHANGE];
      const updatePropertiesHandler = handlers?.[1];

      const element = {
        id: 'password-1',
        type: ElementTypes.PasswordInput,
      } as Element;

      await updatePropertiesHandler('confirmPlaceholder', 'New Placeholder', element, 'step-1');

      const mockNode: Node = {
        id: 'step-1',
        type: 'VIEW',
        position: {x: 0, y: 0},
        data: {
          components: [
            {
              id: 'form-1',
              type: BlockTypes.Form,
              components: [
                {id: 'password-1', type: ElementTypes.PasswordInput},
                {
                  id: 'confirm-password-1',
                  type: ElementTypes.PasswordInput,
                  identifier: FlowBuilderElementConstants.CONFIRM_PASSWORD_IDENTIFIER,
                  placeholder: 'Old Placeholder',
                },
              ],
            },
          ],
        },
      };

      const result = capturedCallback!(mockNode);
      const components = result.components as Element[];
      const form = components[0] as Element & {components?: Element[]};
      const confirmField = form.components?.find(
        (c: Element) => (c as Element & {identifier?: string}).identifier === FlowBuilderElementConstants.CONFIRM_PASSWORD_IDENTIFIER,
      ) as Element & {placeholder?: string};
      expect(confirmField?.placeholder).toBe('New Placeholder');
    });

    it('should execute updateNodeData callback to update required on confirm field', async () => {
      let capturedCallback: ((node: Node) => Record<string, unknown>) | null = null;
      mockUpdateNodeData.mockImplementation(
        (_stepId: string, callback: (node: Node) => Record<string, unknown>) => {
          capturedCallback = callback;
        },
      );

      renderHook(() => useConfirmPasswordField(), {
        wrapper: createWrapper(),
      });

      const handlers = registeredHandlers.async[FlowEventTypes.ON_PROPERTY_CHANGE];
      const updatePropertiesHandler = handlers?.[1];

      const element = {
        id: 'password-1',
        type: ElementTypes.PasswordInput,
      } as Element;

      await updatePropertiesHandler('required', true, element, 'step-1');

      const mockNode: Node = {
        id: 'step-1',
        type: 'VIEW',
        position: {x: 0, y: 0},
        data: {
          components: [
            {
              id: 'form-1',
              type: BlockTypes.Form,
              components: [
                {id: 'password-1', type: ElementTypes.PasswordInput},
                {
                  id: 'confirm-password-1',
                  type: ElementTypes.PasswordInput,
                  identifier: FlowBuilderElementConstants.CONFIRM_PASSWORD_IDENTIFIER,
                  required: false,
                },
              ],
            },
          ],
        },
      };

      const result = capturedCallback!(mockNode);
      const components = result.components as Element[];
      const form = components[0] as Element & {components?: Element[]};
      const confirmField = form.components?.find(
        (c: Element) => (c as Element & {identifier?: string}).identifier === FlowBuilderElementConstants.CONFIRM_PASSWORD_IDENTIFIER,
      ) as Element & {required?: boolean};
      expect(confirmField?.required).toBe(true);
    });

    it('should return empty object when node has no components during update', async () => {
      let capturedCallback: ((node: Node) => Record<string, unknown>) | null = null;
      mockUpdateNodeData.mockImplementation(
        (_stepId: string, callback: (node: Node) => Record<string, unknown>) => {
          capturedCallback = callback;
        },
      );

      renderHook(() => useConfirmPasswordField(), {
        wrapper: createWrapper(),
      });

      const handlers = registeredHandlers.async[FlowEventTypes.ON_PROPERTY_CHANGE];
      const updatePropertiesHandler = handlers?.[1];

      const element = {
        id: 'password-1',
        type: ElementTypes.PasswordInput,
      } as Element;

      await updatePropertiesHandler('confirmHint', 'New hint', element, 'step-1');

      const mockNode: Node = {
        id: 'step-1',
        type: 'VIEW',
        position: {x: 0, y: 0},
        data: {},
      };

      const result = capturedCallback!(mockNode);
      expect(result).toEqual({});
    });

    it('should not update non-form components during property update', async () => {
      let capturedCallback: ((node: Node) => Record<string, unknown>) | null = null;
      mockUpdateNodeData.mockImplementation(
        (_stepId: string, callback: (node: Node) => Record<string, unknown>) => {
          capturedCallback = callback;
        },
      );

      renderHook(() => useConfirmPasswordField(), {
        wrapper: createWrapper(),
      });

      const handlers = registeredHandlers.async[FlowEventTypes.ON_PROPERTY_CHANGE];
      const updatePropertiesHandler = handlers?.[1];

      const element = {
        id: 'password-1',
        type: ElementTypes.PasswordInput,
      } as Element;

      await updatePropertiesHandler('confirmHint', 'New hint', element, 'step-1');

      const mockNode: Node = {
        id: 'step-1',
        type: 'VIEW',
        position: {x: 0, y: 0},
        data: {
          components: [
            {id: 'button-1', type: 'BUTTON'},
            {id: 'password-1', type: ElementTypes.PasswordInput},
          ],
        },
      };

      const result = capturedCallback!(mockNode);
      // Components should remain unchanged since there's no form
      expect((result.components as Element[]).length).toBe(2);
    });
  });

  describe('deleteConfirmPasswordField Handler', () => {
    it('should return true for non-password input types', async () => {
      renderHook(() => useConfirmPasswordField(), {
        wrapper: createWrapper(),
      });

      const deleteHandler = registeredHandlers.async[FlowEventTypes.ON_NODE_ELEMENT_DELETE]?.[0];

      const element = {
        id: 'text-input-1',
        type: ElementTypes.TextInput,
      } as Element;

      const result = await deleteHandler('step-1', element);
      expect(result).toBe(true);
    });

    it('should return true for password without PASSWORD_IDENTIFIER', async () => {
      renderHook(() => useConfirmPasswordField(), {
        wrapper: createWrapper(),
      });

      const deleteHandler = registeredHandlers.async[FlowEventTypes.ON_NODE_ELEMENT_DELETE]?.[0];

      const element = {
        id: 'password-1',
        type: ElementTypes.PasswordInput,
        identifier: 'OTHER_IDENTIFIER',
      } as Element & {identifier?: string};

      const result = await deleteHandler('step-1', element);
      expect(result).toBe(true);
    });

    it('should delete confirm password field when password field is deleted', async () => {
      renderHook(() => useConfirmPasswordField(), {
        wrapper: createWrapper(),
      });

      const deleteHandler = registeredHandlers.async[FlowEventTypes.ON_NODE_ELEMENT_DELETE]?.[0];

      const element = {
        id: 'password-1',
        type: ElementTypes.PasswordInput,
        identifier: FlowBuilderElementConstants.PASSWORD_IDENTIFIER,
      } as Element & {identifier?: string};

      const result = await deleteHandler('step-1', element);
      expect(result).toBe(true);
      expect(mockUpdateNodeData).toHaveBeenCalled();
    });

    it('should execute updateNodeData callback correctly when deleting', async () => {
      let capturedCallback: ((node: Node) => Record<string, unknown>) | null = null;
      mockUpdateNodeData.mockImplementation(
        (_stepId: string, callback: (node: Node) => Record<string, unknown>) => {
          capturedCallback = callback;
        },
      );

      renderHook(() => useConfirmPasswordField(), {
        wrapper: createWrapper(),
      });

      const deleteHandler = registeredHandlers.async[FlowEventTypes.ON_NODE_ELEMENT_DELETE]?.[0];

      const element = {
        id: 'password-1',
        type: ElementTypes.PasswordInput,
        identifier: FlowBuilderElementConstants.PASSWORD_IDENTIFIER,
      } as Element & {identifier?: string};

      await deleteHandler('step-1', element);

      expect(capturedCallback).not.toBeNull();

      const mockNode: Node = {
        id: 'step-1',
        type: 'VIEW',
        position: {x: 0, y: 0},
        data: {
          components: [
            {
              id: 'form-1',
              type: BlockTypes.Form,
              components: [
                {id: 'password-1', type: ElementTypes.PasswordInput},
                {
                  id: 'confirm-password-1',
                  type: ElementTypes.PasswordInput,
                  identifier: FlowBuilderElementConstants.CONFIRM_PASSWORD_IDENTIFIER,
                },
              ],
            },
          ],
        },
      };

      const result = capturedCallback!(mockNode);
      expect(result.components).toBeDefined();
    });

    it('should return empty object when node has no components', async () => {
      let capturedCallback: ((node: Node) => Record<string, unknown>) | null = null;
      mockUpdateNodeData.mockImplementation(
        (_stepId: string, callback: (node: Node) => Record<string, unknown>) => {
          capturedCallback = callback;
        },
      );

      renderHook(() => useConfirmPasswordField(), {
        wrapper: createWrapper(),
      });

      const deleteHandler = registeredHandlers.async[FlowEventTypes.ON_NODE_ELEMENT_DELETE]?.[0];

      const element = {
        id: 'password-1',
        type: ElementTypes.PasswordInput,
        identifier: FlowBuilderElementConstants.PASSWORD_IDENTIFIER,
      } as Element & {identifier?: string};

      await deleteHandler('step-1', element);

      const mockNode: Node = {
        id: 'step-1',
        type: 'VIEW',
        position: {x: 0, y: 0},
        data: {},
      };

      const result = capturedCallback!(mockNode);
      expect(result).toEqual({});
    });

    it('should execute callback to properly remove confirm password field from form', async () => {
      let capturedCallback: ((node: Node) => Record<string, unknown>) | null = null;
      mockUpdateNodeData.mockImplementation(
        (_stepId: string, callback: (node: Node) => Record<string, unknown>) => {
          capturedCallback = callback;
        },
      );

      renderHook(() => useConfirmPasswordField(), {
        wrapper: createWrapper(),
      });

      const deleteHandler = registeredHandlers.async[FlowEventTypes.ON_NODE_ELEMENT_DELETE]?.[0];

      const element = {
        id: 'password-1',
        type: ElementTypes.PasswordInput,
        identifier: FlowBuilderElementConstants.PASSWORD_IDENTIFIER,
      } as Element & {identifier?: string};

      await deleteHandler('step-1', element);

      const mockNode: Node = {
        id: 'step-1',
        type: 'VIEW',
        position: {x: 0, y: 0},
        data: {
          components: [
            {
              id: 'form-1',
              type: BlockTypes.Form,
              components: [
                {id: 'input-1', type: 'TEXT_INPUT'},
                {id: 'password-1', type: ElementTypes.PasswordInput},
                {
                  id: 'confirm-password-1',
                  type: ElementTypes.PasswordInput,
                  identifier: FlowBuilderElementConstants.CONFIRM_PASSWORD_IDENTIFIER,
                },
                {id: 'button-1', type: 'BUTTON'},
              ],
            },
          ],
        },
      };

      const result = capturedCallback!(mockNode);
      expect(result.components).toBeDefined();
      const components = result.components as Element[];
      const form = components[0] as Element & {components?: Element[]};
      // Confirm password should be removed
      expect(form.components?.length).toBe(3);
      expect(form.components?.find((c: Element) =>
        (c as Element & {identifier?: string}).identifier === FlowBuilderElementConstants.CONFIRM_PASSWORD_IDENTIFIER,
      )).toBeUndefined();
    });

    it('should not modify form if no confirm password field exists during delete', async () => {
      let capturedCallback: ((node: Node) => Record<string, unknown>) | null = null;
      mockUpdateNodeData.mockImplementation(
        (_stepId: string, callback: (node: Node) => Record<string, unknown>) => {
          capturedCallback = callback;
        },
      );

      renderHook(() => useConfirmPasswordField(), {
        wrapper: createWrapper(),
      });

      const deleteHandler = registeredHandlers.async[FlowEventTypes.ON_NODE_ELEMENT_DELETE]?.[0];

      const element = {
        id: 'password-1',
        type: ElementTypes.PasswordInput,
        identifier: FlowBuilderElementConstants.PASSWORD_IDENTIFIER,
      } as Element & {identifier?: string};

      await deleteHandler('step-1', element);

      const mockNode: Node = {
        id: 'step-1',
        type: 'VIEW',
        position: {x: 0, y: 0},
        data: {
          components: [
            {
              id: 'form-1',
              type: BlockTypes.Form,
              components: [
                {id: 'password-1', type: ElementTypes.PasswordInput},
                {id: 'button-1', type: 'BUTTON'},
              ],
            },
          ],
        },
      };

      const result = capturedCallback!(mockNode);
      const components = result.components as Element[];
      const form = components[0] as Element & {components?: Element[]};
      // Form components should remain unchanged
      expect(form.components?.length).toBe(2);
    });

    it('should not modify non-form components during delete', async () => {
      let capturedCallback: ((node: Node) => Record<string, unknown>) | null = null;
      mockUpdateNodeData.mockImplementation(
        (_stepId: string, callback: (node: Node) => Record<string, unknown>) => {
          capturedCallback = callback;
        },
      );

      renderHook(() => useConfirmPasswordField(), {
        wrapper: createWrapper(),
      });

      const deleteHandler = registeredHandlers.async[FlowEventTypes.ON_NODE_ELEMENT_DELETE]?.[0];

      const element = {
        id: 'password-1',
        type: ElementTypes.PasswordInput,
        identifier: FlowBuilderElementConstants.PASSWORD_IDENTIFIER,
      } as Element & {identifier?: string};

      await deleteHandler('step-1', element);

      const mockNode: Node = {
        id: 'step-1',
        type: 'VIEW',
        position: {x: 0, y: 0},
        data: {
          components: [
            {id: 'button-1', type: 'BUTTON'},
            {id: 'password-1', type: ElementTypes.PasswordInput},
          ],
        },
      };

      const result = capturedCallback!(mockNode);
      // Components should remain unchanged since there's no form
      expect((result.components as Element[]).length).toBe(2);
    });
  });
});
