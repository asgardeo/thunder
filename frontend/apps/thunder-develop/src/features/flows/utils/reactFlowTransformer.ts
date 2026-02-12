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
import {ElementCategories, ElementTypes, ActionEventTypes, ButtonTypes} from '../models/elements';
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
  prompts?: FlowPrompt[];
  properties?: Record<string, unknown>;
  executor?: {
    name: string;
    inputs?: FlowInput[];
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
 * Flow prompt definition - groups inputs with an action
 */
interface FlowPrompt {
  inputs?: FlowInput[];
  action?: FlowAction;
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
 * Set of input element types for quick lookup
 */
const INPUT_ELEMENT_TYPES = new Set<string>([
  ElementTypes.TextInput,
  ElementTypes.PasswordInput,
  ElementTypes.EmailInput,
  ElementTypes.PhoneInput,
  ElementTypes.NumberInput,
  ElementTypes.DateInput,
  ElementTypes.OtpInput,
  ElementTypes.Checkbox,
  ElementTypes.Dropdown,
]);

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

    // For input field components, ensure ref property is set
    // ref is the attribute selected from the dropdown (e.g., 'username', 'email')
    if (INPUT_ELEMENT_TYPES.has(component.type)) {
      const componentWithProps = component as Element & {name?: string; ref?: string};
      const ref = componentWithProps.name ?? componentWithProps.ref ?? component.id;
      cleanedComponent.ref = ref;
    }

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
    // Check if this is an input field (type is now directly the input type like TEXT_INPUT, PASSWORD_INPUT, etc.)
    if (INPUT_ELEMENT_TYPES.has(component.type)) {
      // Extract ref (attribute) from top-level properties
      // ref is the attribute selected from the dropdown (e.g., 'username', 'email')
      const componentWithProps = component as Element & {name?: string; ref?: string; required?: boolean};
      let identifier: string;
      if (typeof componentWithProps.name === 'string') {
        identifier = componentWithProps.name;
      } else if (typeof componentWithProps.ref === 'string') {
        identifier = componentWithProps.ref;
      } else {
        identifier = component.id;
      }

      const isRequired = componentWithProps.required ?? false;

      inputs.push({
        ref: component.id,
        type: component.type, // The type is already the API type (TEXT_INPUT, PASSWORD_INPUT, etc.)
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
 * Extracts prompts from UI components.
 * Each prompt groups an action with its associated inputs based on container structure.
 *
 * Scoping Rules:
 * 1. Inputs are scoped to their container block hierarchy.
 * 2. An action receives all inputs defined in its immediate container block AND any ancestor blocks.
 * 3. Actions without inputs (at top level or in a block with no inputs) get no inputs.
 *
 * Example Structure:
 * BLOCK A (Input 1)
 *   -> BLOCK B (Input 2, Action Submit)
 * Result: Prompt { action: Submit, inputs: [Input 1, Input 2] }
 */
function extractPrompts(components: Element[], nodeId: string, edges: Edge[]): FlowPrompt[] {
  const prompts: FlowPrompt[] = [];

  /**
   * Extracts inputs from a component tree (for a specific container)
   */
  function extractInputsFromContainer(containerComponents: Element[]): FlowInput[] {
    const inputs: FlowInput[] = [];

    function processForInputs(component: Element): void {
      if (INPUT_ELEMENT_TYPES.has(component.type)) {
        const componentWithProps = component as Element & {name?: string; ref?: string; required?: boolean};
        let identifier: string;
        if (typeof componentWithProps.name === 'string') {
          identifier = componentWithProps.name;
        } else if (typeof componentWithProps.ref === 'string') {
          identifier = componentWithProps.ref;
        } else {
          identifier = component.id;
        }

        inputs.push({
          ref: component.id,
          type: component.type,
          identifier,
          required: componentWithProps.required ?? false,
        });
      }

      if (component.components && component.components.length > 0) {
        component.components.forEach(processForInputs);
      }
    }

    containerComponents.forEach(processForInputs);
    return inputs;
  }

  /**
   * Extracts action from a component (single action only)
   */
  function extractActionFromComponent(component: Element): FlowAction | undefined {
    if (component.type === ElementTypes.Action || component.type === ElementTypes.Resend) {
      const action: FlowAction = {
        ref: component.id,
        nextNode: '',
      };

      const expectedHandle = `${component.id}${NEXT_HANDLE_SUFFIX}`;
      const connectedEdge = edges.find((edge) => edge.source === nodeId && edge.sourceHandle === expectedHandle);

      if (connectedEdge) {
        action.nextNode = connectedEdge.target;
      } else if (component.action?.onSuccess) {
        action.nextNode = component.action.onSuccess;
      }

      if (component.action?.executor) {
        action.executor = component.action.executor as {name: string; [key: string]: unknown};
      }

      return action.nextNode ? action : undefined;
    }
    return undefined;
  }

  /**
   * Processes a component and its children to extract prompts
   */
  function processComponent(component: Element, parentInputs: FlowInput[] = []): void {
    // Check if this component is an action
    const action = extractActionFromComponent(component);
    if (action) {
      // This action gets the parent's inputs (all accumulated up to this point)
      const prompt: FlowPrompt = { action };
      if (parentInputs.length > 0) {
        prompt.inputs = parentInputs;
      }
      prompts.push(prompt);
      return;
    }

    // If this is a BLOCK or container, extract its inputs and process children
    if (component.components && component.components.length > 0) {
      // Extract inputs strictly defined at this level (not deep recursive yet, as recursion happens via processComponent)
      // Note: extractInputsFromContainer is deep recursive, effectively grabbing all inputs in the subtree.
      // But for scoping, we want to pass these down.
      // Logic:
      // 1. Get inputs from this block (and sub-blocks if any, effectively "this container's inputs")
      // 2. Combine with parent inputs
      // 3. Pass to children
      const currentLevelInputs = extractInputsFromContainer([component]);
      
      // Combine parent inputs with current level inputs to support deep nesting (Block A -> Block B)
      const combinedInputs = [...parentInputs, ...currentLevelInputs];

      // Remove duplicates if any (though unlikely given unique IDs, strict accumulation is safer)
      const uniqueInputs = Array.from(new Map(combinedInputs.map(item => [item.ref, item])).values());

      // Process each child component, passing the combined inputs
      component.components.forEach((child) => {
        processComponent(child, uniqueInputs);
      });
    }
  }

  // Process top-level components
  components.forEach((component) => {
    processComponent(component, []);
  });

  return prompts;
}

/**
 * Finds the primary next node from edges or step action.
 * Edges are the source of truth for connections - they represent the current
 * state of the canvas. The action.onSuccess property may be stale from when the
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

  // Fall back to action.onSuccess only if no edges exist (should be rare)
  if (canvasNode.data?.action?.onSuccess) {
    return canvasNode.data.action.onSuccess;
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

    // Extract prompts with proper input-action association
    // Each action inside a BLOCK gets the inputs from that BLOCK
    // Actions without associated inputs (e.g. OAuth buttons) get no inputs
    const prompts = extractPrompts(stepData.components, canvasNode.id, edges);
    if (prompts.length > 0) {
      flowNode.prompts = prompts;
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
        executor: {
          ...flowNode.executor,
          name: executorName,
          inputs: [createOAuthCodeInput()],
        },
      };
    }

    // Find the preceding PROMPT node
    const precedingPromptNode = findPrecedingPromptNode(flowNode.id, canvasNodes, edges);

    if (!precedingPromptNode?.data?.components) {
      return flowNode;
    }

    // Extract inputs from the PROMPT node's components
    const inputs = extractInputs(precedingPromptNode.data.components);

    if (inputs.length > 0 && flowNode.executor?.name) {
      return {
        ...flowNode,
        executor: {
          ...flowNode.executor,
          name: flowNode.executor.name,
          inputs,
        },
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

    // Check action nextNode references (via prompts)
    if (node.prompts) {
      node.prompts.forEach((prompt: FlowPrompt) => {
        if (prompt.action && !nodeIds.has(prompt.action.nextNode)) {
          errors.push(
            `Node ${node.id}, action ${prompt.action.ref}: nextNode references non-existent node ${prompt.action.nextNode}`,
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
