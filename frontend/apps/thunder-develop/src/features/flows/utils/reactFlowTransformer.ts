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

import type {Edge, Node} from '@xyflow/react';
import type {Element} from '../models/elements';
import {ElementTypes, InputVariants, BlockTypes} from '../models/elements';
import type {StepData} from '../models/steps';
import {StepTypes, StaticStepTypes} from '../models/steps';

/**
 * Flow node definition structure
 */
interface FlowNode {
  id: string;
  type: string;
  meta?: {
    components?: FlowComponent[];
  };
  inputs?: FlowInput[];
  actions?: FlowAction[];
  properties?: Record<string, unknown>;
  executor?: {
    name: string;
    [key: string]: unknown;
  };
  onSuccess?: string;
  onFailure?: string;
}

/**
 * Flow UI component structure (for PROMPT nodes)
 */
interface FlowComponent {
  type: string;
  id: string;
  label?: string;
  variant?: string | Record<string, unknown>;
  length?: number;
  required?: boolean;
  placeholder?: string;
  components?: FlowComponent[];
  // Include all UI properties from the original element
  resourceType?: string;
  category?: string;
  properties?: unknown;
  action?: unknown;
}

/**
 * Flow input field definition
 */
interface FlowInput {
  ref: string;
  type: string;
  identifier: string;
  required: boolean;
}

/**
 * Flow action definition
 */
interface FlowAction {
  ref: string;
  nextNode: string;
}

/**
 * Flow graph structure
 */
interface FlowGraph {
  nodes: FlowNode[];
}

/**
 * Complete flow configuration with metadata
 */
interface FlowConfiguration {
  id: string;
  type: string;
  nodes: FlowNode[];
}

/**
 * React Flow canvas data structure
 */
interface ReactFlowCanvasData {
  nodes: Node<StepData>[];
  edges: Edge[];
  viewport?: {
    x: number;
    y: number;
    zoom: number;
  };
}

/**
 * Maps canvas step types to flow node types
 */
const STEP_TO_NODE_TYPE_MAP: Record<string, string> = {
  [StepTypes.View]: 'PROMPT',
  [StepTypes.Execution]: 'TASK_EXECUTION',
  [StepTypes.Rule]: 'DECISION',
  [StepTypes.End]: 'END',
  [StaticStepTypes.Start]: 'START',
  [StaticStepTypes.UserOnboard]: 'END',
};

/**
 * Maps canvas element types to flow component types
 */
const ELEMENT_TO_COMPONENT_TYPE_MAP: Record<string, string> = {
  [ElementTypes.Input]: 'TEXT_INPUT',
  [ElementTypes.Button]: 'ACTION',
  [ElementTypes.Divider]: 'DIVIDER',
  [ElementTypes.Typography]: 'TEXT',
  [ElementTypes.RichText]: 'TEXT',
  [ElementTypes.Image]: 'IMAGE',
  [ElementTypes.Captcha]: 'CAPTCHA',
  [ElementTypes.Choice]: 'CHOICE',
  [ElementTypes.Resend]: 'ACTION',
  [BlockTypes.Form]: 'BLOCK',
};

/**
 * Maps input variants to flow input types
 */
const INPUT_VARIANT_TO_TYPE_MAP: Record<string, string> = {
  [InputVariants.Text]: 'TEXT_INPUT',
  [InputVariants.Password]: 'PASSWORD_INPUT',
  [InputVariants.Email]: 'EMAIL_INPUT',
  [InputVariants.Telephone]: 'PHONE_INPUT',
  [InputVariants.Number]: 'NUMBER_INPUT',
  [InputVariants.Checkbox]: 'CHECKBOX',
  [InputVariants.OTP]: 'OTP_INPUT',
};

/**
 * Transforms a canvas UI element to a flow component
 */
function transformElement(element: Element): FlowComponent | null {
  if (!element || !element.type) {
    return null;
  }

  const component: FlowComponent = {
    type: ELEMENT_TO_COMPONENT_TYPE_MAP[element.type] || element.type,
    id: element.id,
  };

  // Include resourceType and category
  if (element.resourceType) {
    component.resourceType = element.resourceType;
  }

  if (element.category) {
    component.category = element.category;
  }

  // Extract text from config and set as label
  if (element.config && typeof element.config === 'object' && 'text' in element.config) {
    component.label = String(element.config.text);
  }

  // Include variant
  if (element.variant !== undefined) {
    component.variant = typeof element.variant === 'string' ? element.variant : (element.variant as Record<string, unknown>);
  }

  // Include all properties (contains UI-specific data from config)
  if (element.config && typeof element.config === 'object') {
    component.properties = element.config;
  }

  // Include action definition
  if (element.action) {
    component.action = element.action;
  }

  // Override label from config if present (for input fields, etc.)
  const elementConfig = element.config as unknown as Record<string, unknown> | undefined;
  if (elementConfig?.label) {
    component.label = String(elementConfig.label);
  }

  // Handle Input field variants
  if (element.type === ElementTypes.Input && element.variant) {
    const variant = String(element.variant);
    component.type = INPUT_VARIANT_TO_TYPE_MAP[variant] || 'TEXT_INPUT';
  }

  // Handle Button action variants
  if (element.type === ElementTypes.Button && element.variant) {
    component.variant = String(element.variant);
    component.type = 'ACTION';
  }

  // Handle Typography text variants
  if (element.type === ElementTypes.Typography && element.variant) {
    component.variant = String(element.variant);
  }

  // Add required flag for input fields (for backwards compatibility)
  if (elementConfig?.required !== undefined) {
    component.required = Boolean(elementConfig.required);
  }

  // Add placeholder for input fields (for backwards compatibility)
  if (elementConfig?.placeholder) {
    component.placeholder = String(elementConfig.placeholder);
  }

  // Add length for OTP input fields (for backwards compatibility)
  if (elementConfig?.length !== undefined) {
    component.length = Number(elementConfig.length);
  }

  // Handle nested components (for FORM blocks)
  if (element.components && element.components.length > 0) {
    component.components = element.components
      .map(transformElement)
      .filter((c): c is FlowComponent => c !== null);
  }

  return component;
}

/**
 * Extracts input field definitions from UI components
 */
function extractInputs(components: Element[]): FlowInput[] {
  const inputs: FlowInput[] = [];

  function processComponent(component: Element): void {
    // Check if this is an input field
    if (component.type === ElementTypes.Input) {
      const config = component.config as unknown as Record<string, unknown> | undefined;
      const variant = String(component.variant || InputVariants.Text);
      const inputType = INPUT_VARIANT_TO_TYPE_MAP[variant] || 'TEXT_INPUT';

      inputs.push({
        ref: component.id,
        type: inputType,
        identifier: String(config?.name || config?.identifier || component.id),
        required: Boolean(config?.required),
      });
    }

    // Recursively process nested components
    if (component.components && component.components.length > 0) {
      component.components.forEach(processComponent);
    }
  }

  components.forEach(processComponent);
  return inputs;
}

/**
 * Extracts action definitions from UI components and edges
 */
function extractActions(
  components: Element[],
  nodeId: string,
  edges: Edge[]
): FlowAction[] {
  const actions: FlowAction[] = [];

  function processComponent(component: Element): void {
    // Check if this is a button/action element
    if (component.type === ElementTypes.Button || component.type === ElementTypes.Resend) {
      // Check if the component has an action with a next reference
      if (component.action?.next) {
        actions.push({
          ref: component.id,
          nextNode: component.action.next,
        });
      } else {
        // Try to find the next node from edges connected to this button
        // Look for edges with sourceHandle matching the component ID
        const connectedEdge = edges.find(
          (edge) => edge.source === nodeId && edge.sourceHandle === component.id
        );

        if (connectedEdge) {
          actions.push({
            ref: component.id,
            nextNode: connectedEdge.target,
          });
        }
      }
    }

    // Recursively process nested components
    if (component.components && component.components.length > 0) {
      component.components.forEach(processComponent);
    }
  }

  components.forEach(processComponent);
  return actions;
}

/**
 * Finds the primary next node from edges or step action
 */
function findNextNode(canvasNode: Node<StepData>, edges: Edge[]): string | undefined {
  // First check if the step has an action.next defined
  if (canvasNode.data?.action?.next) {
    return canvasNode.data.action.next;
  }

  // Otherwise find from edges (use the default/first edge)
  const outgoingEdges = edges.filter((edge) => edge.source === canvasNode.id);

  if (outgoingEdges.length > 0) {
    // Prefer edges without sourceHandle (default connection)
    const defaultEdge = outgoingEdges.find((edge) => !edge.sourceHandle);
    if (defaultEdge) {
      return defaultEdge.target;
    }

    // Otherwise use the first edge
    return outgoingEdges[0].target;
  }

  return undefined;
}

/**
 * Transforms a React Flow canvas node to a flow node definition
 */
function transformNode(canvasNode: Node<StepData>, edges: Edge[]): FlowNode {
  const flowNode: FlowNode = {
    id: canvasNode.id,
    type: STEP_TO_NODE_TYPE_MAP[canvasNode.type || ''] || canvasNode.type || 'UNKNOWN',
  };

  const stepData = canvasNode.data;

  // Handle PROMPT nodes (VIEW steps with UI components)
  if (canvasNode.type === StepTypes.View && stepData?.components) {
    const components = stepData.components
      .map(transformElement)
      .filter((c): c is FlowComponent => c !== null);

    if (components.length > 0) {
      flowNode.meta = {
        components,
      };
    }

    // Extract input field definitions
    const inputs = extractInputs(stepData.components);
    if (inputs.length > 0) {
      flowNode.inputs = inputs;
    }

    // Extract action definitions from buttons
    const actions = extractActions(stepData.components, canvasNode.id, edges);
    if (actions.length > 0) {
      flowNode.actions = actions;
    }
  }

  // Handle TASK_EXECUTION nodes (EXECUTION steps)
  if (canvasNode.type === StepTypes.Execution) {
    // Add executor configuration
    if (stepData?.action?.executor && stepData.action.executor.name) {
      flowNode.executor = stepData.action.executor as {name: string; [key: string]: unknown};
    }

    // Add execution properties if present
    if (stepData?.properties && Object.keys(stepData.properties).length > 0) {
      flowNode.properties = stepData.properties as Record<string, unknown>;
    }

    // Add onSuccess connection
    const successNode = findNextNode(canvasNode, edges);
    if (successNode) {
      flowNode.onSuccess = successNode;
    }

    // Check for onFailure connection (if there's a sourceHandle named 'failure')
    const failureEdge = edges.find(
      (edge) => edge.source === canvasNode.id && edge.sourceHandle === 'failure'
    );
    if (failureEdge) {
      flowNode.onFailure = failureEdge.target;
    }
  }

  // Handle DECISION nodes (RULE steps)
  if (canvasNode.type === StepTypes.Rule) {
    // Find all outgoing edges
    const outgoingEdges = edges.filter((edge) => edge.source === canvasNode.id);
    const nextNodes = outgoingEdges.map((edge) => edge.target);

    if (nextNodes.length > 0) {
      // For DECISION nodes, we use onSuccess for the primary path
      flowNode.onSuccess = nextNodes[0];
    }

    // Add decision properties if present (for conditions)
    if (stepData?.properties && Object.keys(stepData.properties).length > 0) {
      flowNode.properties = stepData.properties as Record<string, unknown>;
    }
  }

  // Handle START nodes
  if (canvasNode.type === StaticStepTypes.Start) {
    const nextNode = findNextNode(canvasNode, edges);
    if (nextNode) {
      flowNode.onSuccess = nextNode;
    }
  }

  // Handle END nodes (no additional processing needed)
  if (canvasNode.type === StepTypes.End || canvasNode.type === StaticStepTypes.UserOnboard) {
    // END nodes don't need any connections
  }

  return flowNode;
}

/**
 * Main transformer function that converts React Flow canvas data to flow graph format
 *
 * @param canvasData - The output from React Flow's toObject() method
 * @returns The flow graph structure
 */
export function transformReactFlow(canvasData: ReactFlowCanvasData): FlowGraph {
  const flowNodes: FlowNode[] = [];

  // Transform each React Flow canvas node to a flow node
  for (const canvasNode of canvasData.nodes) {
    const flowNode = transformNode(canvasNode, canvasData.edges);
    flowNodes.push(flowNode);
  }

  return {
    nodes: flowNodes,
  };
}

/**
 * Validates the flow graph structure
 *
 * @param flowGraph - The flow graph to validate
 * @returns An array of validation errors (empty if valid)
 */
export function validateFlowGraph(flowGraph: FlowGraph): string[] {
  const errors: string[] = [];
  const nodeIds = new Set(flowGraph.nodes.map((node) => node.id));

  // Check for duplicate node IDs
  const duplicateIds = flowGraph.nodes
    .map((node) => node.id)
    .filter((id, index, arr) => arr.indexOf(id) !== index);

  if (duplicateIds.length > 0) {
    errors.push(`Duplicate node IDs found: ${duplicateIds.join(', ')}`);
  }

  // Validate node connections
  for (const node of flowGraph.nodes) {
    // Check onSuccess references
    if (node.onSuccess && !nodeIds.has(node.onSuccess)) {
      errors.push(`Node ${node.id}: onSuccess references non-existent node ${node.onSuccess}`);
    }

    // Check onFailure references
    if (node.onFailure && !nodeIds.has(node.onFailure)) {
      errors.push(`Node ${node.id}: onFailure references non-existent node ${node.onFailure}`);
    }

    // Check action nextNode references
    if (node.actions) {
      for (const action of node.actions) {
        if (!nodeIds.has(action.nextNode)) {
          errors.push(
            `Node ${node.id}, action ${action.ref}: nextNode references non-existent node ${action.nextNode}`
          );
        }
      }
    }
  }

  // Check for at least one START node
  const startNodes = flowGraph.nodes.filter((node) => node.type === 'START');
  if (startNodes.length === 0) {
    errors.push('Flow must have at least one START node');
  }

  // Check for at least one END node
  const endNodes = flowGraph.nodes.filter((node) => node.type === 'END');
  if (endNodes.length === 0) {
    errors.push('Flow must have at least one END node');
  }

  return errors;
}

/**
 * Creates a complete flow configuration with metadata
 *
 * @param canvasData - The output from React Flow's toObject() method
 * @param flowId - The unique identifier for this flow
 * @param flowType - The type of flow (e.g., 'AUTHENTICATION', 'LOGIN_FLOW')
 * @returns The complete flow configuration with metadata
 */
export function createFlowConfiguration(
  canvasData: ReactFlowCanvasData,
  flowId: string,
  flowType: string = 'AUTHENTICATION'
): FlowConfiguration {
  const flowGraph = transformReactFlow(canvasData);

  return {
    id: flowId,
    type: flowType,
    nodes: flowGraph.nodes,
  };
}

export type {
  FlowNode,
  FlowComponent,
  FlowInput,
  FlowAction,
  FlowGraph,
  FlowConfiguration,
  ReactFlowCanvasData,
};
