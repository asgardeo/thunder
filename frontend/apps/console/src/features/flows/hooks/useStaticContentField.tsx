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

import {type Node, useReactFlow} from '@xyflow/react';
import cloneDeep from 'lodash-es/cloneDeep';
import {useEffect} from 'react';
import useGetFlowBuilderCoreResources from '../api/useGetFlowBuilderCoreResources';
import VisualFlowConstants from '../constants/VisualFlowConstants';
import {type Element, ElementCategories, ElementTypes} from '../models/elements';
import FlowEventTypes from '../models/extension';
import {ExecutionTypes, StepTypes} from '../models/steps';
import PluginRegistry from '../plugins/PluginRegistry';
import generateResourceId from '../utils/generateResourceId';

const STATIC_CONTENT_ENABLED_PROPERTY = 'enableStaticContent';

/**
 * Custom hook to manage static content field in execution nodes.
 */
const useStaticContentField = (): void => {
  const {getNode, updateNodeData} = useReactFlow();
  const {data: resources} = useGetFlowBuilderCoreResources();

  useEffect(() => {
    /**
     * Adds static content to the execution node when staticContentEnabled is checked.
     *
     * @param args - The arguments passed to the handler:
     *   - args[0]: propertyKey - The key of the property being changed.
     *   - args[1]: newValue - The new value of the property.
     *   - args[2]: element - The element being modified.
     *   - args[3]: stepId - The ID of the step where the element is located.
     * @returns Returns false if the static content is added/removed, true otherwise.
     */
    const addStaticContent = async (...args: unknown[]): Promise<boolean> => {
      const [propertyKey, newValue, currentElement, stepId] = args as [string, unknown, Element, string];
      // Check if this is a execution step and the property is staticContentEnabled.
      if (currentElement?.type === StepTypes.Execution && propertyKey === STATIC_CONTENT_ENABLED_PROPERTY) {
        updateNodeData(stepId, (node: Node) => {
          const components: Element[] = cloneDeep(node?.data?.components ?? []) as Element[];

          if (!newValue) {
            // Remove static content if it exists.
            return {
              ...node.data,
              components: [],
            };
          }
          // Add static content if it doesn't exist.
          if (components.length === 0) {
            const richTextElement: Element | undefined = resources?.elements?.find(
              (elem: Element) => elem.type === ElementTypes.RichText,
            );

            if (richTextElement) {
              const staticContentElement: Element = cloneDeep(richTextElement);
              (staticContentElement.config as unknown as Record<string, unknown>).text = '<h3>Static Content</h3>';
              staticContentElement.id = generateResourceId(ElementCategories.Display);
              components.push(staticContentElement);
            }
          }

          return {
            ...node.data,
            components,
          };
        });

        return false;
      }

      return true;
    };

    /**
     * Adds staticContentEnabled property to the execution step property panel.
     *
     * @param args - The arguments passed to the handler:
     *   - args[0]: resource - The resource element to which properties are being added.
     *   - args[1]: properties - The properties object to which static content property will be added.
     *   - args[2]: stepId - The ID of the step where the resource is located.
     * @returns return true.
     */
    const addStaticContentProperties = (...args: unknown[]): boolean => {
      const [resource, properties, stepId] = args as [Element, Record<string, unknown>, string];
      const node: Node | undefined = getNode(stepId);

      if (!node) {
        return true;
      }

      const resourceData = resource?.data as {action?: {executor?: {name?: ExecutionTypes}}} | undefined;
      const executorName = resourceData?.action?.executor?.name;

      // Check if this is a execution step.
      if (
        resource?.type === StepTypes.Execution &&
        executorName &&
        VisualFlowConstants.FLOW_BUILDER_STATIC_CONTENT_ALLOWED_EXECUTION_TYPES.includes(executorName)
      ) {
        if (executorName === ExecutionTypes.MagicLinkExecutor) {
          return true;
        }
        const components: Element[] = (node?.data?.components as Element[]) || [];

        properties[STATIC_CONTENT_ENABLED_PROPERTY] = components.length > 0;
      }

      return true;
    };

    (addStaticContent as unknown as Record<string, unknown>)[
      VisualFlowConstants.FLOW_BUILDER_PLUGIN_FUNCTION_IDENTIFIER
    ] = 'addStaticContent';
    (addStaticContentProperties as unknown as Record<string, unknown>)[
      VisualFlowConstants.FLOW_BUILDER_PLUGIN_FUNCTION_IDENTIFIER
    ] = 'addStaticContentProperties';

    PluginRegistry.getInstance().registerAsync(FlowEventTypes.ON_PROPERTY_CHANGE, addStaticContent);
    PluginRegistry.getInstance().registerSync(FlowEventTypes.ON_PROPERTY_PANEL_OPEN, addStaticContentProperties);

    return () => {
      PluginRegistry.getInstance().unregister(
        FlowEventTypes.ON_PROPERTY_CHANGE,
        (addStaticContent as unknown as Record<string, unknown>)[
          VisualFlowConstants.FLOW_BUILDER_PLUGIN_FUNCTION_IDENTIFIER
        ] as string,
      );
      PluginRegistry.getInstance().unregister(
        FlowEventTypes.ON_PROPERTY_PANEL_OPEN,
        (addStaticContentProperties as unknown as Record<string, unknown>)[
          VisualFlowConstants.FLOW_BUILDER_PLUGIN_FUNCTION_IDENTIFIER
        ] as string,
      );
    };
  }, [getNode, resources, updateNodeData]);
};

export default useStaticContentField;
