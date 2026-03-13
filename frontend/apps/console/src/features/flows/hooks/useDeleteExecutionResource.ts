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

import {type Edge, type Node, useReactFlow} from '@xyflow/react';
import {useEffect} from 'react';
import {ActionTypes} from '../models/actions';
import {type Element, ElementCategories} from '../models/elements';
import FlowEventTypes from '../models/extension';
import {StepTypes} from '../models/steps';
import PluginRegistry from '../plugins/PluginRegistry';
import VisualFlowConstants from '../constants/VisualFlowConstants';
import useFlowBuilderCore from './useFlowBuilderCore';

/**
 * Custom hook to handle the deletion of execution resources in the flow builder.
 *
 * This hook registers an event listener for node deletion events and ensures that
 * any associated execution action nodes are also deleted when an execution node is removed.
 */
const useDeleteExecutionResource = (): void => {
  const {setIsOpenResourcePropertiesPanel} = useFlowBuilderCore();
  const {getEdges, getNodes, updateNodeData, setNodes} = useReactFlow();

  /**
   * Deletes associated execution components when execution nodes are removed.
   *
   * This utility function ensures that when an execution node is deleted from the flow,
   * any related execution initiation action components are also removed to maintain consistency.
   *
   * @param deleted - An array of nodes that have been deleted from the flow.
   */
  async function deleteExecutionActionNode(deleted: Node[]): Promise<boolean> {
    const nodes: Node[] = getNodes();
    const edges: Edge[] = getEdges();
    const actionNodes: Node[] = [];
    const actionComponentIds: string[] = [];

    deleted.forEach((node: Node) => {
      if (node?.type === StepTypes.Execution) {
        const actionNode: Node[] = nodes?.filter((n: Node) =>
          edges?.some((edge: Edge) => {
            if (
              edge.target === node.id &&
              edge.source === n.id &&
              edge?.sourceHandle?.includes(VisualFlowConstants.FLOW_BUILDER_NEXT_HANDLE_SUFFIX)
            ) {
              actionComponentIds.push(
                edge.sourceHandle.slice(0, -VisualFlowConstants.FLOW_BUILDER_NEXT_HANDLE_SUFFIX.length),
              );

              return true;
            }

            return false;
          }),
        );

        if (actionNode?.length > 0) {
          actionNodes.push(...actionNode);
        }
      }
    });

    // If no action nodes are found, return true to indicate no further action is needed.
    if (actionNodes.length === 0) {
      return true;
    }

    actionNodes.forEach((actionNode: Node) => {
      updateNodeData(actionNode.id, (node: Node) => {
        const components: Element[] = (node.data.components as Element[])?.filter(
          (component: Element) => !actionComponentIds.includes(component.id),
        );

        return {
          components,
        };
      });
    });
    setIsOpenResourcePropertiesPanel(false);

    return true;
  }

  /**
   * Deletes the execution node when a execution action is removed.
   *
   * @param _stepId - The ID of the step from which the element is being deleted.
   * @param element - The element being deleted, which is expected to be a execution action.
   * @returns Returns true if the deletion was successful.
   */
  async function deleteExecutionNode(_stepId: string, element: Element): Promise<boolean> {
    const action = element.action as {type?: string; onSuccess?: string} | undefined;

    if (element.category === ElementCategories.Action && action?.type === ActionTypes.Next) {
      setNodes((nodes: Node[]) =>
        nodes?.filter((node: Node) => node.id !== action?.onSuccess || node.type !== StepTypes.Execution),
      );
    }

    return true;
  }

  /**
   * Deletes the component and node associated with the deleted edges.
   *
   * @param deleted - The deleted edges from the flow.
   * @returns Returns true if the deletion was successful.
   */
  async function deleteComponentAndNode(deleted: Edge[]): Promise<boolean> {
    const nodes: Node[] = getNodes();
    const executionNodeIds: string[] = [];
    const actionNodeIds: string[] = [];
    const actionComponentIds: string[] = [];

    deleted.forEach((edge: Edge) => {
      nodes.forEach((node: Node) => {
        if (node.type === StepTypes.Execution && edge.target === node.id && edge.sourceHandle) {
          actionComponentIds.push(
            edge.sourceHandle.slice(0, -VisualFlowConstants.FLOW_BUILDER_NEXT_HANDLE_SUFFIX.length),
          );
          actionNodeIds.push(edge.source);
          executionNodeIds.push(edge.target);
        }
      });
    });

    if (executionNodeIds.length === 0) {
      return true;
    }

    setNodes((nds: Node[]) => nds?.filter((node: Node) => !executionNodeIds.includes(node.id)));

    actionNodeIds.forEach((actionNodeId: string) => {
      updateNodeData(actionNodeId, (node: Node) => {
        const components: Element[] = (node.data.components as Element[])?.filter(
          (component: Element) => !actionComponentIds.includes(component.id),
        );

        return {
          components,
        };
      });
    });
    setIsOpenResourcePropertiesPanel(false);

    return true;
  }

  useEffect(() => {
    // Attach unique identifiers to the functions for plugin registration
    (
      deleteComponentAndNode as typeof deleteComponentAndNode & Record<string, string>
    )[VisualFlowConstants.FLOW_BUILDER_PLUGIN_FUNCTION_IDENTIFIER] = 'deleteComponentAndNode';
    (
      deleteExecutionNode as typeof deleteExecutionNode & Record<string, string>
    )[VisualFlowConstants.FLOW_BUILDER_PLUGIN_FUNCTION_IDENTIFIER] = 'deleteExecutionNode';
    (
      deleteExecutionActionNode as typeof deleteExecutionActionNode & Record<string, string>
    )[VisualFlowConstants.FLOW_BUILDER_PLUGIN_FUNCTION_IDENTIFIER] = 'deleteExecutionActionNode';

    PluginRegistry.getInstance().registerAsync(
      FlowEventTypes.ON_NODE_DELETE,
      deleteExecutionActionNode as (...args: unknown[]) => Promise<boolean>,
    );
    PluginRegistry.getInstance().registerAsync(
      FlowEventTypes.ON_NODE_ELEMENT_DELETE,
      deleteExecutionNode as (...args: unknown[]) => Promise<boolean>,
    );
    PluginRegistry.getInstance().registerAsync(
      FlowEventTypes.ON_EDGE_DELETE,
      deleteComponentAndNode as (...args: unknown[]) => Promise<boolean>,
    );

    return () => {
      PluginRegistry.getInstance().unregister(
        FlowEventTypes.ON_NODE_DELETE,
        (deleteExecutionActionNode as typeof deleteExecutionActionNode & Record<string, string>)[
          VisualFlowConstants.FLOW_BUILDER_PLUGIN_FUNCTION_IDENTIFIER
        ],
      );
      PluginRegistry.getInstance().unregister(
        FlowEventTypes.ON_NODE_ELEMENT_DELETE,
        (deleteExecutionNode as typeof deleteExecutionNode & Record<string, string>)[
          VisualFlowConstants.FLOW_BUILDER_PLUGIN_FUNCTION_IDENTIFIER
        ],
      );
      PluginRegistry.getInstance().unregister(
        FlowEventTypes.ON_EDGE_DELETE,
        (deleteComponentAndNode as typeof deleteComponentAndNode & Record<string, string>)[
          VisualFlowConstants.FLOW_BUILDER_PLUGIN_FUNCTION_IDENTIFIER
        ],
      );
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);
};

export default useDeleteExecutionResource;
