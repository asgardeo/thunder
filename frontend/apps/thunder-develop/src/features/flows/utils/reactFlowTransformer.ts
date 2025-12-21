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
import {ElementCategories, ElementTypes, InputVariants, ActionEventTypes, ButtonTypes} from '../models/elements';
import type {StepData} from '../models/steps';
import {StepTypes, StaticStepTypes} from '../models/steps';
import {ActionTypes} from '../models/actions';
import generateResourceId from './generateResourceId';

/**
 * Suffix used in edge sourceHandle to identify the connection point
 */
const NEXT_HANDLE_SUFFIX = `_${ActionTypes.Next}`;

/**
 * Layout information for a node
 */
interface NodeLayout {
  size: {
    width: number;
    height: number;
  };
  position: {
    x: number;
    y: number;
  };
}

/**
 * Default layout dimensions
 */
const DEFAULT_LAYOUT = {
  width: 200,
  height: 100,
};

/**
 * Flow node definition structure
 */
interface FlowNode {
  id: string;
  type: string;
  layout: NodeLayout;
  meta?: {
    components?: Record<string, unknown>[];
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
  executor?: {
    name: string;
    [key: string]: unknown;
  };
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
  name: string;
  handle: string;
  flowType: string;
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
 * Derives the eventType for ACTION category components based on buttonType
 */
function deriveEventType(component?: Element & {buttonType?: string}): string {
  const buttonType = component?.buttonType;

  if (!buttonType) {
    return ActionEventTypes.Trigger;
  }

  switch (buttonType) {
    case ButtonTypes.Submit:
      return ActionEventTypes.Submit;
    case ButtonTypes.Button:
    default:
      return ActionEventTypes.Trigger;
  }
}

/**
 * Removes internal properties (variants, display, config, action) from components recursively.
 * These transformations prepare the component for the API payload.
 * Note: action is removed because actions are defined separately in the node's actions array.
 */
function cleanComponents(components: Element[]): Record<string, unknown>[] {
  return components.map((component) => {
    // Extract and remove internal properties (including action which is defined in node.actions)
    // eslint-disable-next-line @typescript-eslint/no-unused-vars -- config is excluded from output
    const {variants, display, config, action, ...rest} = component as Element & {
      variants?: unknown;
      display?: unknown;
      config?: unknown;
      action?: unknown;
    };

    // Build the cleaned component
    const cleanedComponent: Record<string, unknown> = {
      ...rest,
    };

    // For ACTION category components, ensure eventType is set
    if (component.category === ElementCategories.Action && !cleanedComponent.eventType) {
      cleanedComponent.eventType = deriveEventType(component as Element & {buttonType?: string});
    }

    // Recursively clean nested components if present
    const nestedComponents = cleanedComponent.components as Element[] | undefined;
    if (nestedComponents && nestedComponents.length > 0) {
      cleanedComponent.components = cleanComponents(nestedComponents);
    }

    return cleanedComponent;
  });
}

/**
 * Extracts input field definitions from UI components
 */
function extractInputs(components: Element[]): FlowInput[] {
  const inputs: FlowInput[] = [];

  function processComponent(component: Element): void {
    // Check if this is an input field
    if (component.type === ElementTypes.Input) {
      const variantValue = component.variant;
      const variant = typeof variantValue === 'string' ? variantValue : InputVariants.Text;
      const inputType = INPUT_VARIANT_TO_TYPE_MAP[variant] ?? 'TEXT_INPUT';

      // Extract identifier from top-level properties
      const componentWithProps = component as Element & {name?: string; identifier?: string; required?: boolean};
      let identifier: string;
      if (typeof componentWithProps.name === 'string') {
        identifier = componentWithProps.name;
      } else if (typeof componentWithProps.identifier === 'string') {
        identifier = componentWithProps.identifier;
      } else {
        identifier = component.id;
      }

      const isRequired = componentWithProps.required ?? false;

      inputs.push({
        ref: component.id,
        type: inputType,
        identifier,
        required: isRequired,
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
 * Extracts action definitions from UI components and edges.
 * Edges are the source of truth for connections - they represent the current
 * state of the canvas. The action.next property may be stale from when the
 * flow was loaded.
 */
function extractActions(components: Element[], nodeId: string, edges: Edge[]): FlowAction[] {
  const actions: FlowAction[] = [];

  function processComponent(component: Element): void {
    // Check if this is a button/action element
    if (component.type === ElementTypes.Button || component.type === ElementTypes.Resend) {
      // Build the action object
      const action: FlowAction = {
        ref: component.id,
        nextNode: '',
      };

      // First try to find the next node from edges (source of truth for connections)
      // The sourceHandle includes a suffix (e.g., "button_id_NEXT")
      const expectedHandle = `${component.id}${NEXT_HANDLE_SUFFIX}`;
      const connectedEdge = edges.find((edge) => edge.source === nodeId && edge.sourceHandle === expectedHandle);

      if (connectedEdge) {
        action.nextNode = connectedEdge.target;
      } else if (component.action?.next) {
        // Fall back to action.next only if no edge exists
        action.nextNode = component.action.next;
      }

      // Include executor information if present (for EXECUTOR action type)
      if (component.action?.executor) {
        action.executor = component.action.executor as {name: string; [key: string]: unknown};
      }

      // Only add the action if we have a valid nextNode
      if (action.nextNode) {
        actions.push(action);
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
 * Finds the primary next node from edges or step action.
 * Edges are the source of truth for connections - they represent the current
 * state of the canvas. The action.next property may be stale from when the
 * flow was loaded.
 */
function findNextNode(canvasNode: Node<StepData>, edges: Edge[]): string | undefined {
  // First try to find from edges (these are the source of truth for connections)
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

  // Fall back to action.next only if no edges exist (should be rare)
  if (canvasNode.data?.action?.next) {
    return canvasNode.data.action.next;
  }

  return undefined;
}

/**
 * Transforms a React Flow canvas node to a flow node definition
 */
function transformNode(canvasNode: Node<StepData>, edges: Edge[]): FlowNode {
  const stepData = canvasNode.data;

  // Build the layout from canvas node position and measured dimensions
  const layout: NodeLayout = {
    size: {
      width: canvasNode.measured?.width ?? canvasNode.width ?? DEFAULT_LAYOUT.width,
      height: canvasNode.measured?.height ?? canvasNode.height ?? DEFAULT_LAYOUT.height,
    },
    position: {
      x: Math.round(canvasNode.position.x),
      y: Math.round(canvasNode.position.y),
    },
  };

  const flowNode: FlowNode = {
    id: canvasNode.id,
    type: STEP_TO_NODE_TYPE_MAP[canvasNode.type ?? ''] ?? canvasNode.type ?? 'UNKNOWN',
    layout,
  };

  // Handle PROMPT nodes (VIEW steps with UI components)
  // Clean components to remove internal properties like variants
  if (canvasNode.type === StepTypes.View && stepData?.components) {
    flowNode.meta = {
      components: cleanComponents(stepData.components),
    };

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

  // Handle END nodes with components
  if (canvasNode.type === StepTypes.End && stepData?.components) {
    flowNode.meta = {
      components: cleanComponents(stepData.components),
    };
  }

  // Handle TASK_EXECUTION nodes (EXECUTION steps)
  if (canvasNode.type === StepTypes.Execution) {
    // Add executor configuration
    if (stepData?.action?.executor?.name) {
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
    const failureEdge = edges.find((edge) => edge.source === canvasNode.id && edge.sourceHandle === 'failure');
    if (failureEdge) {
      flowNode.onFailure = failureEdge.target;
    }

    // Note: inputs for TASK_EXECUTION nodes are collected in a second pass
    // after all nodes are transformed, since we need to look at preceding PROMPT nodes
  }

  // Handle DECISION nodes (RULE steps)
  if (canvasNode.type === StepTypes.Rule) {
    // Find all outgoing edges
    const outgoingEdges = edges.filter((edge) => edge.source === canvasNode.id);
    const nextNodes = outgoingEdges.map((edge) => edge.target);

    if (nextNodes.length > 0) {
      // For DECISION nodes, we use onSuccess for the primary path
      [flowNode.onSuccess] = nextNodes;
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

  // Handle END nodes (no additional processing needed for connections)
  // Components are already handled above

  return flowNode;
}

/**
 * Finds the PROMPT node that connects to the given TASK_EXECUTION node
 * by tracing back through the edges and actions.
 */
function findPrecedingPromptNode(
  targetNodeId: string,
  canvasNodes: Node<StepData>[],
  edges: Edge[],
): Node<StepData> | undefined {
  // Find edges that point to this node
  const incomingEdges = edges.filter((edge) => edge.target === targetNodeId);

  // Find the source nodes for incoming edges
  const sourceNodes = incomingEdges
    .map((edge) => canvasNodes.find((node) => node.id === edge.source))
    .filter((node): node is Node<StepData> => node !== undefined);

  // First, check if any source is directly a PROMPT (VIEW) node
  const directPromptNode = sourceNodes.find((node) => node.type === StepTypes.View);
  if (directPromptNode) {
    return directPromptNode;
  }

  // Check if any source is a START node and follow to find PROMPT
  const startNode = sourceNodes.find((node) => node.type === StaticStepTypes.Start);
  if (startNode) {
    const nextFromStart = findNextNode(startNode, edges);
    if (nextFromStart) {
      const nextNode = canvasNodes.find((node) => node.id === nextFromStart);
      if (nextNode?.type === StepTypes.View) {
        return nextNode;
      }
    }
  }

  return undefined;
}

/**
 * List of OAuth/OIDC executor names that require a 'code' input for OAuth callback.
 * These executors handle external authentication and don't inherit form inputs.
 */
const OAUTH_EXECUTOR_NAMES = new Set(['GoogleOIDCAuthExecutor', 'GithubOAuthExecutor']);

/**
 * Creates the standard OAuth code input for OAuth/OIDC executors.
 */
function createOAuthCodeInput(): FlowInput {
  return {
    ref: generateResourceId('input'),
    type: 'TEXT_INPUT',
    identifier: 'code',
    required: true,
  };
}

/**
 * Collects inputs for TASK_EXECUTION nodes from their preceding PROMPT nodes.
 * This is done in a second pass after all nodes are transformed.
 * Returns a new array of flow nodes with inputs added where applicable.
 *
 * Note: OAuth/OIDC executors (Google, GitHub, etc.) receive a standard 'code' input
 * for handling the OAuth callback, rather than inheriting form inputs.
 */
function collectInputsForExecutionNodes(
  flowNodes: FlowNode[],
  canvasNodes: Node<StepData>[],
  edges: Edge[],
): FlowNode[] {
  return flowNodes.map((flowNode) => {
    if (flowNode.type !== 'TASK_EXECUTION') {
      return flowNode;
    }

    const executorName = flowNode.executor?.name;

    // OAuth/OIDC executors get a standard 'code' input for OAuth callback
    if (executorName && OAUTH_EXECUTOR_NAMES.has(executorName)) {
      return {
        ...flowNode,
        inputs: [createOAuthCodeInput()],
      };
    }

    // Find the preceding PROMPT node
    const precedingPromptNode = findPrecedingPromptNode(flowNode.id, canvasNodes, edges);

    if (!precedingPromptNode?.data?.components) {
      return flowNode;
    }

    // Extract inputs from the PROMPT node's components
    const inputs = extractInputs(precedingPromptNode.data.components);

    if (inputs.length > 0) {
      return {
        ...flowNode,
        inputs,
      };
    }

    return flowNode;
  });
}

/**
 * Main transformer function that converts React Flow canvas data to flow graph format
 *
 * @param canvasData - The output from React Flow's toObject() method
 * @returns The flow graph structure
 */
export function transformReactFlow(canvasData: ReactFlowCanvasData): FlowGraph {
  // Transform each React Flow canvas node to a flow node
  const flowNodes: FlowNode[] = canvasData.nodes.map((canvasNode) => transformNode(canvasNode, canvasData.edges));

  // Second pass: collect inputs for TASK_EXECUTION nodes from preceding PROMPT nodes
  const nodesWithInputs = collectInputsForExecutionNodes(flowNodes, canvasData.nodes, canvasData.edges);

  return {
    nodes: nodesWithInputs,
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
  const duplicateIds = flowGraph.nodes.map((node) => node.id).filter((id, index, arr) => arr.indexOf(id) !== index);

  if (duplicateIds.length > 0) {
    errors.push(`Duplicate node IDs found: ${duplicateIds.join(', ')}`);
  }

  // Validate node connections
  flowGraph.nodes.forEach((node) => {
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
      node.actions.forEach((action) => {
        if (!nodeIds.has(action.nextNode)) {
          errors.push(
            `Node ${node.id}, action ${action.ref}: nextNode references non-existent node ${action.nextNode}`,
          );
        }
      });
    }
  });

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
 * @param flowName - The name of the flow
 * @param flowType - The type of flow (e.g., 'AUTHENTICATION', 'LOGIN_FLOW')
 * @returns The complete flow configuration with metadata
 */
export function createFlowConfiguration(
  canvasData: ReactFlowCanvasData,
  flowName = 'New Flow',
  flowHandle = 'new-flow',
  flowType = 'AUTHENTICATION',
): FlowConfiguration {
  const flowGraph = transformReactFlow(canvasData);

  return {
    name: flowName,
    handle: flowHandle,
    flowType,
    nodes: flowGraph.nodes,
  };
}

export type {FlowNode, FlowInput, FlowAction, FlowGraph, FlowConfiguration, ReactFlowCanvasData};
