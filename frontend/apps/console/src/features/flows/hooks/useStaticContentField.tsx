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
import {useCallback, useEffect} from 'react';
import useFlowPlugins from './useFlowPlugins';
import useGetFlowBuilderCoreResources from '../api/useGetFlowBuilderCoreResources';
import VisualFlowConstants from '../constants/VisualFlowConstants';
import type {Properties} from '../models/base';
import {type Element, ElementCategories, ElementTypes} from '../models/elements';
import {ExecutionTypes, StepTypes} from '../models/steps';
import generateResourceId from '../utils/generateResourceId';

const STATIC_CONTENT_ENABLED_PROPERTY = 'enableStaticContent';

/**
 * Custom hook to manage static content field in execution nodes.
 */
const useStaticContentField = (): void => {
  const {getNode, updateNodeData} = useReactFlow();
  const {data: resources} = useGetFlowBuilderCoreResources();
  const {onPropertyChange, onPropertyPanelOpen} = useFlowPlugins();

  /**
   * Adds static content to the execution node when staticContentEnabled is checked.
   */
  const addStaticContent = useCallback(
    (propertyKey: string, newValue: unknown, currentElement: Element, stepId: string): boolean => {
      if (currentElement?.type === StepTypes.Execution && propertyKey === STATIC_CONTENT_ENABLED_PROPERTY) {
        updateNodeData(stepId, (node: Node) => {
          const components: Element[] = cloneDeep(node?.data?.components ?? []) as Element[];

          if (!newValue) {
            return {
              ...node.data,
              components: [],
            };
          }

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
    },
    [resources, updateNodeData],
  );

  /**
   * Adds staticContentEnabled property to the execution step property panel.
   */
  const addStaticContentProperties = useCallback(
    (resource: Element, properties: Properties, stepId: string): boolean => {
      const node: Node | undefined = getNode(stepId);

      if (!node) {
        return true;
      }

      const resourceData = resource?.data as {action?: {executor?: {name?: ExecutionTypes}}} | undefined;
      const executorName = resourceData?.action?.executor?.name;

      if (
        resource?.type === StepTypes.Execution &&
        executorName &&
        VisualFlowConstants.FLOW_BUILDER_STATIC_CONTENT_ALLOWED_EXECUTION_TYPES.includes(executorName)
      ) {
        if (executorName === ExecutionTypes.MagicLinkExecutor) {
          return true;
        }
        const components: Element[] = (node?.data?.components as Element[]) || [];

        (properties as Record<string, unknown>)[STATIC_CONTENT_ENABLED_PROPERTY] = components.length > 0;
      }

      return true;
    },
    [getNode],
  );

  useEffect(() => onPropertyChange(addStaticContent), [onPropertyChange, addStaticContent]);
  useEffect(() => onPropertyPanelOpen(addStaticContentProperties), [onPropertyPanelOpen, addStaticContentProperties]);
};

export default useStaticContentField;
