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

import {useCallback} from 'react';
import {MarkerType, type Edge, type Node} from '@xyflow/react';
import VisualFlowConstants from '@/features/flows/constants/VisualFlowConstants';
import {BlockTypes, ElementTypes, type Element} from '@/features/flows/models/elements';
import {StepTypes, type Step} from '@/features/flows/models/steps';
import LoginFlowConstants from '../constants/LoginFlowConstants';

// Use centralized constants
const {START_STEP_ID: INITIAL_FLOW_START_STEP_ID, END_STEP_ID: INITIAL_FLOW_USER_ONBOARD_STEP_ID, DEFAULT_EDGE_TYPE} =
  LoginFlowConstants;

/**
 * Helper to create an edge with standard configuration.
 */
const createEdge = (
  id: string,
  source: string,
  sourceHandle: string,
  target: string,
): Edge => ({
  animated: false,
  id,
  markerEnd: {
    type: MarkerType.Arrow,
  },
  source,
  sourceHandle,
  target,
  type: DEFAULT_EDGE_TYPE,
});

/**
 * Props for useEdgeGeneration hook.
 */
export interface UseEdgeGenerationProps {
  /**
   * The ID of the start step.
   * @default 'START'
   */
  startStepId?: string;
  /**
   * The ID of the end/user onboard step.
   * @default 'END'
   */
  endStepId?: string;
}

/**
 * Return type for useEdgeGeneration hook.
 */
export interface UseEdgeGenerationReturn {
  /**
   * Generates edges for the flow based on step configuration.
   */
  generateEdges: (flowSteps: Step[]) => Edge[];
  /**
   * Validates edges to ensure all targets exist.
   * Accepts Node[] or Step[] since it only needs the id property.
   */
  validateEdges: (edges: Edge[], nodes: Node[]) => Edge[];
}

/**
 * Hook for generating and validating edges in the login flow builder.
 *
 * This hook encapsulates the edge generation logic which:
 * 1. Creates edge from START to first step
 * 2. Creates edges based on button actions within steps
 * 3. Creates edges based on step-level actions
 * 4. Ensures connection to END step if not explicitly configured
 *
 * @param props - Configuration for edge generation.
 * @returns Functions for generating and validating edges.
 */
const useEdgeGeneration = (props?: UseEdgeGenerationProps): UseEdgeGenerationReturn => {
  const startStepId = props?.startStepId ?? INITIAL_FLOW_START_STEP_ID;
  const defaultEndStepId = props?.endStepId ?? INITIAL_FLOW_USER_ONBOARD_STEP_ID;

  /**
   * Generate edges for the flow based on step configurations.
   */
  const generateEdges = useCallback(
    (flowSteps: Step[]): Edge[] => {
      const generatedEdges: Edge[] = [];

      // Get all step IDs for validation
      const stepIds = new Set(flowSteps.map((step: Step) => step.id));

      // Find the user onboard step
      const userOnboardStep = flowSteps.find((step: Step) => step.type === StepTypes.End);
      const userOnboardStepId = userOnboardStep?.id ?? defaultEndStepId;

      // Track if we've created an edge to the user onboard step
      let userOnboardEdgeCreated = false;

      /**
       * Create edges for a button based on its action configuration.
       */
      const createEdgesForButton = (step: Step, button: Element): void => {
        const sourceHandle = `${button.id}${VisualFlowConstants.FLOW_BUILDER_NEXT_HANDLE_SUFFIX}`;

        if (button.action?.next) {
          if (stepIds.has(button.action.next)) {
            // Valid step reference
            generatedEdges.push(createEdge(button.id, step.id, sourceHandle, button.action.next));
            if (button.action.next === userOnboardStepId) {
              userOnboardEdgeCreated = true;
            }
          } else if (button.action.next === StepTypes.End) {
            // Reference to END type - connect to actual end step
            generatedEdges.push(createEdge(button.id, step.id, sourceHandle, userOnboardStepId));
            userOnboardEdgeCreated = true;
          }
        } else {
          // No explicit next - connect to user onboard step
          generatedEdges.push(createEdge(button.id, step.id, sourceHandle, userOnboardStepId));
          userOnboardEdgeCreated = true;
        }
      };

      // Create edge from START to first step
      if (flowSteps.length > 0) {
        let [firstStep] = flowSteps;

        // Skip START node if it's the first step
        if (firstStep.id === startStepId && flowSteps.length > 1) {
          [, firstStep] = flowSteps;
        }

        if (firstStep && firstStep.id !== startStepId) {
          generatedEdges.push(
            createEdge(
              `${startStepId}-${firstStep.id}`,
              startStepId,
              `${startStepId}${VisualFlowConstants.FLOW_BUILDER_NEXT_HANDLE_SUFFIX}`,
              firstStep.id,
            ),
          );
        }
      }

      // Process each step for action-based edges
      flowSteps
        .filter((step) => step.type !== StepTypes.End)
        .forEach((step) => {
          // Process components with actions
          if (step.data?.components) {
            step.data.components.forEach((component) => {
              // Process buttons inside forms
              if (component.type === BlockTypes.Form && component.components) {
                component.components
                  .filter((formComponent) => formComponent.type === ElementTypes.Button)
                  .forEach((formComponent) => createEdgesForButton(step, formComponent));
              }

              // Process direct button components
              if (component.type === ElementTypes.Button) {
                createEdgesForButton(step, component);
              }
            });
          }

          // Process step-level actions
          if (step.data?.action?.next) {
            const sourceHandle = `${step.id}${VisualFlowConstants.FLOW_BUILDER_NEXT_HANDLE_SUFFIX}`;

            if (stepIds.has(step.data.action.next)) {
              generatedEdges.push(
                createEdge(`${step.id}-to-${step.data.action.next}`, step.id, sourceHandle, step.data.action.next),
              );
              if (step.data.action.next === userOnboardStepId) {
                userOnboardEdgeCreated = true;
              }
            } else if (step.data.action.next === StepTypes.End) {
              generatedEdges.push(
                createEdge(`${step.id}-to-${userOnboardStepId}`, step.id, sourceHandle, userOnboardStepId),
              );
              userOnboardEdgeCreated = true;
            }
          }
        });

      // If no edge to user onboard was created, connect the last view step
      if (!userOnboardEdgeCreated && flowSteps.length > 0) {
        const viewSteps = flowSteps.filter((step: Step) => step.type === StepTypes.View);

        if (viewSteps.length > 0) {
          const lastViewStep = viewSteps[viewSteps.length - 1];
          let buttonId: string | null = null;

          // Try to find a button to use for the connection
          if (lastViewStep.data?.components) {
            const formComponent = lastViewStep.data.components.find(
              (component: Element) => component.type === BlockTypes.Form,
            );

            if (formComponent?.components) {
              const button = formComponent.components.find((elem: Element) => elem.type === ElementTypes.Button);
              if (button) {
                buttonId = button.id;
              }
            }
          }

          const edgeId = buttonId ?? `${lastViewStep.id}-to-${userOnboardStepId}`;
          const sourceHandle = buttonId
            ? `${buttonId}${VisualFlowConstants.FLOW_BUILDER_NEXT_HANDLE_SUFFIX}`
            : `${lastViewStep.id}${VisualFlowConstants.FLOW_BUILDER_NEXT_HANDLE_SUFFIX}`;

          generatedEdges.push(createEdge(edgeId, lastViewStep.id, sourceHandle, userOnboardStepId));
        }
      }

      return generatedEdges;
    },
    [startStepId, defaultEndStepId],
  );

  /**
   * Validate edges to ensure all targets exist in the flow.
   * Accepts Node[] or Step[] since it only needs the id property.
   */
  const validateEdges = useCallback((edges: Edge[], nodes: Node[]): Edge[] => {
    const stepIds = new Set(nodes.map((node) => node.id));

    // Always include START and END as valid targets
    stepIds.add(INITIAL_FLOW_START_STEP_ID);
    stepIds.add(INITIAL_FLOW_USER_ONBOARD_STEP_ID);

    return edges.filter((edge) => {
      const targetExists = stepIds.has(edge.target);
      const sourceExists = stepIds.has(edge.source);

      return targetExists && sourceExists;
    });
  }, []);

  return {
    generateEdges,
    validateEdges,
  };
};

export default useEdgeGeneration;
