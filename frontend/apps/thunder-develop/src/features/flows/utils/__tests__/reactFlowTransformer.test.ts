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
import type {Node, Edge} from '@xyflow/react';
import {
  transformReactFlow,
  validateFlowGraph,
  createFlowConfiguration,
  type ReactFlowCanvasData,
  type FlowGraph,
} from '../reactFlowTransformer';
import type {StepData} from '../../models/steps';
import {StepTypes, StaticStepTypes} from '../../models/steps';
import {ElementTypes, ElementCategories, ActionEventTypes, ButtonTypes} from '../../models/elements';
import type {Element} from '../../models/elements';

// Mock generateResourceId
vi.mock('../generateResourceId', () => ({
  default: vi.fn((prefix: string) => `${prefix}-generated-id`),
}));

describe('reactFlowTransformer', () => {
  const createNode = (
    id: string,
    type: string,
    position: {x: number; y: number} = {x: 0, y: 0},
    data: StepData = {},
  ): Node<StepData> => ({
    id,
    type,
    position,
    data,
  });

  const createEdge = (id: string, source: string, target: string, sourceHandle?: string): Edge => ({
    id,
    source,
    target,
    ...(sourceHandle && {sourceHandle}),
  });

  describe('transformReactFlow', () => {
    describe('Basic Node Transformation', () => {
      it('should transform START node correctly', () => {
        const canvasData: ReactFlowCanvasData = {
          nodes: [createNode('start-1', StaticStepTypes.Start, {x: 100, y: 100})],
          edges: [],
        };

        const result = transformReactFlow(canvasData);

        expect(result.nodes).toHaveLength(1);
        expect(result.nodes[0].id).toBe('start-1');
        expect(result.nodes[0].type).toBe('START');
        expect(result.nodes[0].layout.position).toEqual({x: 100, y: 100});
      });

      it('should transform END node correctly', () => {
        const canvasData: ReactFlowCanvasData = {
          nodes: [createNode('end-1', StepTypes.End, {x: 200, y: 200})],
          edges: [],
        };

        const result = transformReactFlow(canvasData);

        expect(result.nodes).toHaveLength(1);
        expect(result.nodes[0].type).toBe('END');
      });

      it('should transform VIEW node to PROMPT', () => {
        const canvasData: ReactFlowCanvasData = {
          nodes: [createNode('view-1', StepTypes.View, {x: 0, y: 0})],
          edges: [],
        };

        const result = transformReactFlow(canvasData);

        expect(result.nodes[0].type).toBe('PROMPT');
      });

      it('should transform EXECUTION node to TASK_EXECUTION', () => {
        const canvasData: ReactFlowCanvasData = {
          nodes: [createNode('exec-1', StepTypes.Execution, {x: 0, y: 0})],
          edges: [],
        };

        const result = transformReactFlow(canvasData);

        expect(result.nodes[0].type).toBe('TASK_EXECUTION');
      });

      it('should transform RULE node to DECISION', () => {
        const canvasData: ReactFlowCanvasData = {
          nodes: [createNode('rule-1', StepTypes.Rule, {x: 0, y: 0})],
          edges: [],
        };

        const result = transformReactFlow(canvasData);

        expect(result.nodes[0].type).toBe('DECISION');
      });

      it('should use measured dimensions when available', () => {
        const node: Node<StepData> = {
          id: 'node-1',
          type: StepTypes.View,
          position: {x: 0, y: 0},
          data: {},
          measured: {width: 300, height: 200},
        };

        const canvasData: ReactFlowCanvasData = {
          nodes: [node],
          edges: [],
        };

        const result = transformReactFlow(canvasData);

        expect(result.nodes[0].layout.size).toEqual({width: 300, height: 200});
      });

      it('should use default dimensions when measured is not available', () => {
        const canvasData: ReactFlowCanvasData = {
          nodes: [createNode('node-1', StepTypes.View)],
          edges: [],
        };

        const result = transformReactFlow(canvasData);

        expect(result.nodes[0].layout.size).toEqual({width: 200, height: 100});
      });
    });

    describe('Component Processing', () => {
      it('should clean and include components in meta for VIEW nodes', () => {
        const components: Element[] = [
          {
            id: 'text-input-1',
            type: ElementTypes.TextInput,
            category: ElementCategories.Field,
            resourceType: 'ELEMENT',
            version: '1.0.0',
            deprecated: false,
            deletable: true,
            display: {label: 'Username', image: '', showOnResourcePanel: true},
            config: {field: {name: 'username', type: {}}, styles: {}},
          } as Element,
        ];

        const canvasData: ReactFlowCanvasData = {
          nodes: [createNode('view-1', StepTypes.View, {x: 0, y: 0}, {components})],
          edges: [],
        };

        const result = transformReactFlow(canvasData);

        expect(result.nodes[0].meta).toBeDefined();
        expect(result.nodes[0].meta?.components).toHaveLength(1);
        // Verify internal properties are removed
        expect(result.nodes[0].meta?.components?.[0]).not.toHaveProperty('display');
        expect(result.nodes[0].meta?.components?.[0]).not.toHaveProperty('config');
      });

      it('should extract inputs from VIEW components', () => {
        const components: Element[] = [
          {
            id: 'text-input-1',
            type: ElementTypes.TextInput,
            category: ElementCategories.Field,
            name: 'username',
            required: true,
          } as unknown as Element,
          {
            id: 'password-input-1',
            type: ElementTypes.PasswordInput,
            category: ElementCategories.Field,
            name: 'password',
            required: true,
          } as unknown as Element,
        ];

        const canvasData: ReactFlowCanvasData = {
          nodes: [createNode('view-1', StepTypes.View, {x: 0, y: 0}, {components})],
          edges: [],
        };

        const result = transformReactFlow(canvasData);

        expect(result.nodes[0].inputs).toHaveLength(2);
        expect(result.nodes[0].inputs?.[0]).toEqual({
          ref: 'text-input-1',
          type: ElementTypes.TextInput,
          identifier: 'username',
          required: true,
        });
      });

      it('should extract actions from buttons', () => {
        const components: Element[] = [
          {
            id: 'button-1',
            type: ElementTypes.Action,
            category: ElementCategories.Action,
            action: {next: 'next-node'},
          } as Element,
        ];

        const canvasData: ReactFlowCanvasData = {
          nodes: [createNode('view-1', StepTypes.View, {x: 0, y: 0}, {components})],
          edges: [createEdge('edge-1', 'view-1', 'next-node', 'button-1_NEXT')],
        };

        const result = transformReactFlow(canvasData);

        expect(result.nodes[0].actions).toHaveLength(1);
        expect(result.nodes[0].actions?.[0]).toEqual({
          ref: 'button-1',
          nextNode: 'next-node',
        });
      });

      it('should handle nested components in forms', () => {
        const formComponent: Element = {
          id: 'form-1',
          type: 'BLOCK',
          category: ElementCategories.Block,
          components: [
            {
              id: 'input-1',
              type: ElementTypes.TextInput,
              category: ElementCategories.Field,
              name: 'email',
            } as unknown as Element,
          ],
        } as unknown as Element;

        const canvasData: ReactFlowCanvasData = {
          nodes: [createNode('view-1', StepTypes.View, {x: 0, y: 0}, {components: [formComponent]})],
          edges: [],
        };

        const result = transformReactFlow(canvasData);

        expect(result.nodes[0].inputs).toHaveLength(1);
        expect(result.nodes[0].inputs?.[0].identifier).toBe('email');
      });

      it('should handle components in END nodes', () => {
        const components: Element[] = [
          {
            id: 'text-1',
            type: ElementTypes.Text,
            category: ElementCategories.Display,
          } as Element,
        ];

        const canvasData: ReactFlowCanvasData = {
          nodes: [createNode('end-1', StepTypes.End, {x: 0, y: 0}, {components})],
          edges: [],
        };

        const result = transformReactFlow(canvasData);

        expect(result.nodes[0].meta?.components).toHaveLength(1);
      });
    });

    describe('Edge Connections', () => {
      it('should set onSuccess for START node from edges', () => {
        const canvasData: ReactFlowCanvasData = {
          nodes: [createNode('start-1', StaticStepTypes.Start), createNode('view-1', StepTypes.View)],
          edges: [createEdge('edge-1', 'start-1', 'view-1')],
        };

        const result = transformReactFlow(canvasData);

        const startNode = result.nodes.find((n) => n.type === 'START');
        expect(startNode?.onSuccess).toBe('view-1');
      });

      it('should set onSuccess for EXECUTION node from edges', () => {
        const canvasData: ReactFlowCanvasData = {
          nodes: [
            createNode(
              'exec-1',
              StepTypes.Execution,
              {x: 0, y: 0},
              {
                action: {executor: {name: 'TestExecutor'}},
              },
            ),
            createNode('end-1', StepTypes.End),
          ],
          edges: [createEdge('edge-1', 'exec-1', 'end-1')],
        };

        const result = transformReactFlow(canvasData);

        const execNode = result.nodes.find((n) => n.type === 'TASK_EXECUTION');
        expect(execNode?.onSuccess).toBe('end-1');
      });

      it('should set onFailure for EXECUTION node when failure handle exists', () => {
        const canvasData: ReactFlowCanvasData = {
          nodes: [
            createNode('exec-1', StepTypes.Execution),
            createNode('success-1', StepTypes.End),
            createNode('failure-1', StepTypes.End),
          ],
          edges: [createEdge('edge-1', 'exec-1', 'success-1'), createEdge('edge-2', 'exec-1', 'failure-1', 'failure')],
        };

        const result = transformReactFlow(canvasData);

        const execNode = result.nodes.find((n) => n.type === 'TASK_EXECUTION');
        expect(execNode?.onSuccess).toBe('success-1');
        expect(execNode?.onFailure).toBe('failure-1');
      });

      it('should set onSuccess for DECISION node from edges', () => {
        const canvasData: ReactFlowCanvasData = {
          nodes: [createNode('rule-1', StepTypes.Rule), createNode('view-1', StepTypes.View)],
          edges: [createEdge('edge-1', 'rule-1', 'view-1')],
        };

        const result = transformReactFlow(canvasData);

        const ruleNode = result.nodes.find((n) => n.type === 'DECISION');
        expect(ruleNode?.onSuccess).toBe('view-1');
      });

      it('should prefer edges over action.next for button connections', () => {
        const components: Element[] = [
          {
            id: 'button-1',
            type: ElementTypes.Action,
            category: ElementCategories.Action,
            action: {next: 'stale-node'}, // This is stale
          } as Element,
        ];

        const canvasData: ReactFlowCanvasData = {
          nodes: [createNode('view-1', StepTypes.View, {x: 0, y: 0}, {components})],
          edges: [createEdge('edge-1', 'view-1', 'current-node', 'button-1_NEXT')],
        };

        const result = transformReactFlow(canvasData);

        // Should use edge target, not action.next
        expect(result.nodes[0].actions?.[0].nextNode).toBe('current-node');
      });
    });

    describe('Execution Node Processing', () => {
      it('should include executor configuration', () => {
        const canvasData: ReactFlowCanvasData = {
          nodes: [
            createNode(
              'exec-1',
              StepTypes.Execution,
              {x: 0, y: 0},
              {
                action: {
                  executor: {
                    name: 'TestExecutor',
                    config: {key: 'value'},
                  },
                },
              },
            ),
          ],
          edges: [],
        };

        const result = transformReactFlow(canvasData);

        expect(result.nodes[0].executor).toEqual({
          name: 'TestExecutor',
          config: {key: 'value'},
        });
      });

      it('should include properties for EXECUTION nodes', () => {
        const canvasData: ReactFlowCanvasData = {
          nodes: [
            createNode(
              'exec-1',
              StepTypes.Execution,
              {x: 0, y: 0},
              {
                properties: {timeout: 5000, retries: 3},
              },
            ),
          ],
          edges: [],
        };

        const result = transformReactFlow(canvasData);

        expect(result.nodes[0].properties).toEqual({timeout: 5000, retries: 3});
      });

      it('should collect inputs from preceding PROMPT node', () => {
        const promptComponents: Element[] = [
          {
            id: 'input-1',
            type: ElementTypes.TextInput,
            category: ElementCategories.Field,
            name: 'username',
          } as unknown as Element,
        ];

        const canvasData: ReactFlowCanvasData = {
          nodes: [
            createNode('view-1', StepTypes.View, {x: 0, y: 0}, {components: promptComponents}),
            createNode(
              'exec-1',
              StepTypes.Execution,
              {x: 100, y: 0},
              {
                action: {executor: {name: 'PasswordValidator'}},
              },
            ),
          ],
          edges: [createEdge('edge-1', 'view-1', 'exec-1')],
        };

        const result = transformReactFlow(canvasData);

        const execNode = result.nodes.find((n) => n.type === 'TASK_EXECUTION');
        expect(execNode?.inputs).toHaveLength(1);
        expect(execNode?.inputs?.[0].identifier).toBe('username');
      });

      it('should use code input for OAuth executors', () => {
        const canvasData: ReactFlowCanvasData = {
          nodes: [
            createNode(
              'exec-1',
              StepTypes.Execution,
              {x: 0, y: 0},
              {
                action: {executor: {name: 'GoogleOIDCAuthExecutor'}},
              },
            ),
          ],
          edges: [],
        };

        const result = transformReactFlow(canvasData);

        const execNode = result.nodes.find((n) => n.type === 'TASK_EXECUTION');
        expect(execNode?.inputs).toHaveLength(1);
        expect(execNode?.inputs?.[0].identifier).toBe('code');
        expect(execNode?.inputs?.[0].type).toBe('TEXT_INPUT');
      });

      it('should use code input for GitHub OAuth executor', () => {
        const canvasData: ReactFlowCanvasData = {
          nodes: [
            createNode(
              'exec-1',
              StepTypes.Execution,
              {x: 0, y: 0},
              {
                action: {executor: {name: 'GithubOAuthExecutor'}},
              },
            ),
          ],
          edges: [],
        };

        const result = transformReactFlow(canvasData);

        const execNode = result.nodes.find((n) => n.type === 'TASK_EXECUTION');
        expect(execNode?.inputs?.[0].identifier).toBe('code');
      });
    });

    describe('Event Type Derivation', () => {
      it('should derive SUBMIT eventType for submit button', () => {
        const components: Element[] = [
          {
            id: 'button-1',
            type: ElementTypes.Action,
            category: ElementCategories.Action,
            buttonType: ButtonTypes.Submit,
          } as Element & {buttonType: string},
        ];

        const canvasData: ReactFlowCanvasData = {
          nodes: [createNode('view-1', StepTypes.View, {x: 0, y: 0}, {components})],
          edges: [],
        };

        const result = transformReactFlow(canvasData);

        expect(result.nodes[0].meta?.components?.[0].eventType).toBe(ActionEventTypes.Submit);
      });

      it('should derive TRIGGER eventType for regular button', () => {
        const components: Element[] = [
          {
            id: 'button-1',
            type: ElementTypes.Action,
            category: ElementCategories.Action,
            buttonType: ButtonTypes.Button,
          } as Element & {buttonType: string},
        ];

        const canvasData: ReactFlowCanvasData = {
          nodes: [createNode('view-1', StepTypes.View, {x: 0, y: 0}, {components})],
          edges: [],
        };

        const result = transformReactFlow(canvasData);

        expect(result.nodes[0].meta?.components?.[0].eventType).toBe(ActionEventTypes.Trigger);
      });

      it('should default to TRIGGER eventType when buttonType is missing', () => {
        const components: Element[] = [
          {
            id: 'button-1',
            type: ElementTypes.Action,
            category: ElementCategories.Action,
          } as Element,
        ];

        const canvasData: ReactFlowCanvasData = {
          nodes: [createNode('view-1', StepTypes.View, {x: 0, y: 0}, {components})],
          edges: [],
        };

        const result = transformReactFlow(canvasData);

        expect(result.nodes[0].meta?.components?.[0].eventType).toBe(ActionEventTypes.Trigger);
      });
    });

    describe('Input Field Processing', () => {
      it('should set ref for input fields', () => {
        const components: Element[] = [
          {
            id: 'input-1',
            type: ElementTypes.EmailInput,
            category: ElementCategories.Field,
            name: 'email',
          } as unknown as Element,
        ];

        const canvasData: ReactFlowCanvasData = {
          nodes: [createNode('view-1', StepTypes.View, {x: 0, y: 0}, {components})],
          edges: [],
        };

        const result = transformReactFlow(canvasData);

        expect(result.nodes[0].meta?.components?.[0].ref).toBe('email');
      });

      it('should use id as ref fallback when name is missing', () => {
        const components: Element[] = [
          {
            id: 'input-1',
            type: ElementTypes.TextInput,
            category: ElementCategories.Field,
          } as Element,
        ];

        const canvasData: ReactFlowCanvasData = {
          nodes: [createNode('view-1', StepTypes.View, {x: 0, y: 0}, {components})],
          edges: [],
        };

        const result = transformReactFlow(canvasData);

        expect(result.nodes[0].meta?.components?.[0].ref).toBe('input-1');
      });

      it('should handle all input element types', () => {
        const inputTypes = [
          ElementTypes.TextInput,
          ElementTypes.PasswordInput,
          ElementTypes.EmailInput,
          ElementTypes.PhoneInput,
          ElementTypes.NumberInput,
          ElementTypes.DateInput,
          ElementTypes.OtpInput,
          ElementTypes.Checkbox,
          ElementTypes.Dropdown,
        ];

        const components: Element[] = inputTypes.map((type, index) => ({
          id: `input-${index}`,
          type,
          category: ElementCategories.Field,
          name: `field-${index}`,
        })) as unknown as Element[];

        const canvasData: ReactFlowCanvasData = {
          nodes: [createNode('view-1', StepTypes.View, {x: 0, y: 0}, {components})],
          edges: [],
        };

        const result = transformReactFlow(canvasData);

        expect(result.nodes[0].inputs).toHaveLength(inputTypes.length);
      });
    });
  });

  describe('validateFlowGraph', () => {
    it('should return empty array for valid flow', () => {
      const flowGraph: FlowGraph = {
        nodes: [
          {
            id: 'start-1',
            type: 'START',
            layout: {size: {width: 100, height: 50}, position: {x: 0, y: 0}},
            onSuccess: 'end-1',
          },
          {id: 'end-1', type: 'END', layout: {size: {width: 100, height: 50}, position: {x: 100, y: 0}}},
        ],
      };

      const errors = validateFlowGraph(flowGraph);

      expect(errors).toHaveLength(0);
    });

    it('should detect duplicate node IDs', () => {
      const flowGraph: FlowGraph = {
        nodes: [
          {id: 'node-1', type: 'START', layout: {size: {width: 100, height: 50}, position: {x: 0, y: 0}}},
          {id: 'node-1', type: 'END', layout: {size: {width: 100, height: 50}, position: {x: 100, y: 0}}},
        ],
      };

      const errors = validateFlowGraph(flowGraph);

      expect(errors).toContain('Duplicate node IDs found: node-1');
    });

    it('should detect missing START node', () => {
      const flowGraph: FlowGraph = {
        nodes: [{id: 'end-1', type: 'END', layout: {size: {width: 100, height: 50}, position: {x: 0, y: 0}}}],
      };

      const errors = validateFlowGraph(flowGraph);

      expect(errors).toContain('Flow must have at least one START node');
    });

    it('should detect missing END node', () => {
      const flowGraph: FlowGraph = {
        nodes: [{id: 'start-1', type: 'START', layout: {size: {width: 100, height: 50}, position: {x: 0, y: 0}}}],
      };

      const errors = validateFlowGraph(flowGraph);

      expect(errors).toContain('Flow must have at least one END node');
    });

    it('should detect invalid onSuccess reference', () => {
      const flowGraph: FlowGraph = {
        nodes: [
          {
            id: 'start-1',
            type: 'START',
            layout: {size: {width: 100, height: 50}, position: {x: 0, y: 0}},
            onSuccess: 'non-existent',
          },
          {id: 'end-1', type: 'END', layout: {size: {width: 100, height: 50}, position: {x: 100, y: 0}}},
        ],
      };

      const errors = validateFlowGraph(flowGraph);

      expect(errors).toContain('Node start-1: onSuccess references non-existent node non-existent');
    });

    it('should detect invalid onFailure reference', () => {
      const flowGraph: FlowGraph = {
        nodes: [
          {id: 'start-1', type: 'START', layout: {size: {width: 100, height: 50}, position: {x: 0, y: 0}}},
          {
            id: 'exec-1',
            type: 'TASK_EXECUTION',
            layout: {size: {width: 100, height: 50}, position: {x: 50, y: 0}},
            onFailure: 'non-existent',
          },
          {id: 'end-1', type: 'END', layout: {size: {width: 100, height: 50}, position: {x: 100, y: 0}}},
        ],
      };

      const errors = validateFlowGraph(flowGraph);

      expect(errors).toContain('Node exec-1: onFailure references non-existent node non-existent');
    });

    it('should detect invalid action nextNode reference', () => {
      const flowGraph: FlowGraph = {
        nodes: [
          {
            id: 'start-1',
            type: 'START',
            layout: {size: {width: 100, height: 50}, position: {x: 0, y: 0}},
          },
          {
            id: 'prompt-1',
            type: 'PROMPT',
            layout: {size: {width: 100, height: 50}, position: {x: 50, y: 0}},
            actions: [{ref: 'button-1', nextNode: 'non-existent'}],
          },
          {id: 'end-1', type: 'END', layout: {size: {width: 100, height: 50}, position: {x: 100, y: 0}}},
        ],
      };

      const errors = validateFlowGraph(flowGraph);

      expect(errors).toContain('Node prompt-1, action button-1: nextNode references non-existent node non-existent');
    });
  });

  describe('createFlowConfiguration', () => {
    it('should create flow configuration with default values', () => {
      const canvasData: ReactFlowCanvasData = {
        nodes: [createNode('start-1', StaticStepTypes.Start), createNode('end-1', StepTypes.End)],
        edges: [],
      };

      const config = createFlowConfiguration(canvasData);

      expect(config.name).toBe('New Flow');
      expect(config.handle).toBe('new-flow');
      expect(config.flowType).toBe('AUTHENTICATION');
      expect(config.nodes).toHaveLength(2);
    });

    it('should create flow configuration with custom values', () => {
      const canvasData: ReactFlowCanvasData = {
        nodes: [createNode('start-1', StaticStepTypes.Start), createNode('end-1', StepTypes.End)],
        edges: [],
      };

      const config = createFlowConfiguration(canvasData, 'Custom Flow', 'custom-flow', 'REGISTRATION');

      expect(config.name).toBe('Custom Flow');
      expect(config.handle).toBe('custom-flow');
      expect(config.flowType).toBe('REGISTRATION');
    });
  });

  describe('Edge Cases', () => {
    it('should handle empty canvas data', () => {
      const canvasData: ReactFlowCanvasData = {
        nodes: [],
        edges: [],
      };

      const result = transformReactFlow(canvasData);

      expect(result.nodes).toHaveLength(0);
    });

    it('should handle node with unknown type', () => {
      const canvasData: ReactFlowCanvasData = {
        nodes: [createNode('unknown-1', 'UNKNOWN_TYPE')],
        edges: [],
      };

      const result = transformReactFlow(canvasData);

      expect(result.nodes[0].type).toBe('UNKNOWN_TYPE');
    });

    it('should handle node without type', () => {
      const node: Node<StepData> = {
        id: 'node-1',
        position: {x: 0, y: 0},
        data: {},
      };

      const canvasData: ReactFlowCanvasData = {
        nodes: [node],
        edges: [],
      };

      const result = transformReactFlow(canvasData);

      expect(result.nodes[0].type).toBe('UNKNOWN');
    });

    it('should handle VIEW node without components', () => {
      const canvasData: ReactFlowCanvasData = {
        nodes: [createNode('view-1', StepTypes.View)],
        edges: [],
      };

      const result = transformReactFlow(canvasData);

      expect(result.nodes[0].meta).toBeUndefined();
      expect(result.nodes[0].inputs).toBeUndefined();
      expect(result.nodes[0].actions).toBeUndefined();
    });

    it('should round position values', () => {
      const canvasData: ReactFlowCanvasData = {
        nodes: [createNode('node-1', StepTypes.View, {x: 100.7, y: 200.3})],
        edges: [],
      };

      const result = transformReactFlow(canvasData);

      expect(result.nodes[0].layout.position).toEqual({x: 101, y: 200});
    });

    it('should handle action without next node', () => {
      const components: Element[] = [
        {
          id: 'button-1',
          type: ElementTypes.Action,
          category: ElementCategories.Action,
          action: {}, // No next defined
        } as Element,
      ];

      const canvasData: ReactFlowCanvasData = {
        nodes: [createNode('view-1', StepTypes.View, {x: 0, y: 0}, {components})],
        edges: [], // No edges either
      };

      const result = transformReactFlow(canvasData);

      // Action without next should not be included
      expect(result.nodes[0].actions).toBeUndefined();
    });

    it('should include executor in action when present', () => {
      const components: Element[] = [
        {
          id: 'button-1',
          type: ElementTypes.Action,
          category: ElementCategories.Action,
          action: {
            next: 'exec-1',
            executor: {name: 'TestExecutor'},
          },
        } as unknown as Element,
      ];

      const canvasData: ReactFlowCanvasData = {
        nodes: [createNode('view-1', StepTypes.View, {x: 0, y: 0}, {components})],
        edges: [createEdge('edge-1', 'view-1', 'exec-1', 'button-1_NEXT')],
      };

      const result = transformReactFlow(canvasData);

      expect(result.nodes[0].actions?.[0].executor).toEqual({name: 'TestExecutor'});
    });
  });
});
