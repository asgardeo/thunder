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

import {useEffect, useMemo, useRef} from 'react';
import type {EdgeTypes, NodeProps, NodeTypes} from '@xyflow/react';
import {StaticStepTypes, type Step} from '@/features/flows/models/steps';
import type {Element} from '@/features/flows/models/elements';
import type {Resources} from '@/features/flows/models/resources';
import BaseEdge from '@/features/flows/components/react-flow-overrides/BaseEdge';
import StepFactory from '../components/resources/steps/StepFactory';
import StaticStepFactory from '../components/resources/steps/StaticStepFactory';
import LoginFlowConstants from '../constants/LoginFlowConstants';

/**
 * Props for the useNodeTypes hook.
 */
export interface UseNodeTypesProps {
  /** Array of step definitions. */
  steps: Step[] | undefined;
  /** All flow builder resources. */
  resources: Resources;
  /** Callback to add an element to a view. */
  onAddElementToView: (element: Element, viewId: string) => void;
  /** Callback to add an element to a form. */
  onAddElementToForm: (element: Element, formId: string) => void;
}

/**
 * Return type for the useNodeTypes hook.
 */
export interface UseNodeTypesReturn {
  /** Node types for React Flow. */
  nodeTypes: NodeTypes;
  /** Edge types for React Flow. */
  edgeTypes: EdgeTypes;
}

/**
 * Hook to generate node and edge types for React Flow.
 * Uses refs to avoid unnecessary re-renders when callbacks change.
 *
 * @param props - Configuration options for the hook.
 * @returns Node and edge types for React Flow.
 */
const useNodeTypes = (props: UseNodeTypesProps): UseNodeTypesReturn => {
  const {steps, resources, onAddElementToView, onAddElementToForm} = props;

  // Use refs to store callbacks and resources to avoid nodeTypes recreation
  const onAddElementToViewRef = useRef(onAddElementToView);
  const onAddElementToFormRef = useRef(onAddElementToForm);
  const resourcesRef = useRef(resources);
  const stepsByTypeRef = useRef<Record<string, Step[]>>({});

  // Update refs when data changes (doesn't trigger re-render)
  useEffect(() => {
    onAddElementToViewRef.current = onAddElementToView;
    onAddElementToFormRef.current = onAddElementToForm;
    resourcesRef.current = resources;
  }, [onAddElementToView, onAddElementToForm, resources]);

  // Organize steps by type
  useEffect(() => {
    if (steps) {
      stepsByTypeRef.current = steps.reduce((acc: Record<string, Step[]>, step: Step) => {
        if (!acc[step.type]) {
          acc[step.type] = [];
        }
        acc[step.type].push(step);
        return acc;
      }, {});
    }
  }, [steps]);

  /**
   * Node types for React Flow.
   * Uses refs to read data at render time, avoiding recreation when callbacks change.
   */
  const nodeTypes = useMemo((): NodeTypes => {
    // Get unique step types from steps (only to determine which types we need)
    const stepTypeSet = new Set(steps?.map((s) => s.type) ?? []);

    const stepNodes: NodeTypes = Array.from(stepTypeSet).reduce((acc: NodeTypes, stepType: string) => {
      // Create a stable component that reads from refs at render time
      acc[stepType] = (nodeProps: NodeProps) => (
        // @ts-expect-error NodeProps doesn't include all required properties but they're provided at runtime
        <StepFactory
          resourceId={nodeProps.id}
          resources={stepsByTypeRef.current[stepType] ?? []}
          allResources={resourcesRef.current}
          onAddElement={(element: Element) => onAddElementToViewRef.current?.(element, nodeProps.id)}
          onAddElementToForm={(element: Element, formId: string) =>
            onAddElementToFormRef.current?.(element, formId)
          }
          {...nodeProps}
        />
      );
      return acc;
    }, {});

    const staticStepNodes: NodeTypes = Object.values(StaticStepTypes).reduce(
      (acc: NodeTypes, type: StaticStepTypes) => {
        acc[type] = (nodeProps: NodeProps) => (
          // @ts-expect-error NodeProps doesn't include all required properties but they're provided at runtime
          <StaticStepFactory {...nodeProps} type={type} />
        );

        return acc;
      },
      {},
    );

    return {
      ...staticStepNodes,
      ...stepNodes,
    };
    // IMPORTANT: Only depend on the step types array, not resources
    // The actual data is accessed via refs at render time
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [steps?.map((s) => s.type).join(',')]);

  /**
   * Edge types for React Flow.
   * Register BaseEdge which uses SmartStepEdge for intelligent routing around nodes.
   */
  const edgeTypes: EdgeTypes = useMemo(
    () => ({
      default: BaseEdge,
      smoothstep: BaseEdge,
      step: BaseEdge,
      [LoginFlowConstants.DEFAULT_EDGE_TYPE]: BaseEdge,
    }),
    [],
  );

  return {
    nodeTypes,
    edgeTypes,
  };
};

export default useNodeTypes;
