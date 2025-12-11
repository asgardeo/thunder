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

import {useReactFlow} from '@xyflow/react';
import {useCallback} from 'react';
import {type Element} from '../models/elements';

/**
 * Represents the contract returned by the `useComponentDelete` hook.
 */
export interface UseComponentDelete {
  /**
   * Deletes a given component from a specified step.
   *
   * @param stepId - ID of the step (node) from which the component should be deleted.
   * @param component - The component to be removed.
   */
  deleteComponent: (stepId: string, component: Element) => void;
}

/**
 * Custom React hook to provide functionality for deleting a component from a step node.
 *
 * It leverages the `useReactFlow` hook from `@xyflow/react` to update node data and
 * performs a deep recursive deletion by matching component IDs.
 *
 * @returns An object with a `deleteComponent` function to remove a component from a step.
 */
const useComponentDelete = (): UseComponentDelete => {
  const {updateNodeData} = useReactFlow();

  /**
   * Deletes a component from the list of components in a step node by ID.
   *
   * @param stepId - The unique identifier of the step node.
   * @param component - The component object to be deleted.
   */
  const deleteComponent = useCallback(
    (stepId: string, component: Element): void => {
      /**
       * Recursively filters out the specified component and its children.
       * Creates new arrays/objects immutably without needing cloneDeep.
       *
       * @param components - Array of components to update.
       * @returns A new array of components with the specified one removed.
       */
      const updateComponent = (components: Element[]): Element[] =>
        components?.reduce((acc: Element[], _component: Element) => {
          if (_component.id === component.id) {
            return acc;
          }

          const updatedComponent = _component.components
            ? {..._component, components: updateComponent(_component.components)}
            : _component;

          acc.push(updatedComponent);

          return acc;
        }, []);

      updateNodeData(stepId, (node: {data?: {components?: Element[]}}) => {
        const components: Element[] = updateComponent(node?.data?.components ?? []);

        return {
          components,
        };
      });
    },
    [updateNodeData],
  );

  return {deleteComponent};
};

export default useComponentDelete;
