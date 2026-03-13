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
import type {Node} from '@xyflow/react';
import type {UpdateNodeInternals} from '@xyflow/system';
import cloneDeep from 'lodash-es/cloneDeep';
import {BlockTypes, ElementCategories, type Element} from '@/features/flows/models/elements';
import {ResourceTypes} from '@/features/flows/models/resources';
import {StepTypes} from '@/features/flows/models/steps';
import generateIdsForResources from '@/features/flows/utils/generateIdsForResources';
import useGenerateStepElement from '@/features/flows/hooks/useGenerateStepElement';
import {INPUT_ELEMENT_TYPES, mutateComponents} from '../utils/componentMutations';

/**
 * Props for the useElementAddition hook.
 */
export interface UseElementAdditionProps {
  /** Function to set nodes in the flow. */
  setNodes: React.Dispatch<React.SetStateAction<Node[]>>;
  /** Function to update node internals for rendering. */
  updateNodeInternals: UpdateNodeInternals;
}

/**
 * Return type for the useElementAddition hook.
 */
export interface UseElementAdditionReturn {
  /** Add an element to a view from the context menu. */
  handleAddElementToView: (element: Element, viewId: string) => void;
  /** Add an element to a form from the context menu. */
  handleAddElementToForm: (element: Element, formId: string) => void;
}

/**
 * Hook to handle adding elements to views and forms.
 * Manages the logic for inserting elements into the correct containers
 * with proper form creation and mutation.
 *
 * @param props - Configuration options for the hook.
 * @returns Element addition handlers.
 */
const useElementAddition = (props: UseElementAdditionProps): UseElementAdditionReturn => {
  const {setNodes, updateNodeInternals} = props;
  const {generateStepElement} = useGenerateStepElement();

  /**
   * Callback for adding an element to a view from the context menu.
   * Adds the element to the specified View.
   */
  const handleAddElementToView = useCallback(
    (element: Element, viewId: string): void => {
      // Use generateStepElement to properly apply variants and generate unique IDs
      const generatedElement: Element = generateStepElement(element);
      let viewStepId: string | null = null;

      setNodes((prevNodes: Node[]) => {
        // Find the View node with the given viewId
        const existingViewStep = prevNodes.find((node) => node.id === viewId && node.type === StepTypes.View);

        if (!existingViewStep) {
          return prevNodes; // No View exists, do nothing
        }

        // Store view ID for later use
        viewStepId = existingViewStep.id;

        // For INPUT elements, add them to Form (create Form if needed)
        if (INPUT_ELEMENT_TYPES.has(element.type)) {
          return prevNodes.map((node) => {
            if (node.id === existingViewStep.id) {
              const nodeData = node.data as {components?: Element[]} | undefined;
              const existingComponents: Element[] = cloneDeep(nodeData?.components ?? []);

              // Find existing Form in the View
              const existingForm = existingComponents.find((comp: Element) => comp.type === BlockTypes.Form);

              if (existingForm) {
                // Add input to existing Form
                const updatedForm: Element = {
                  ...existingForm,
                  components: [...(existingForm.components ?? []), generatedElement],
                };

                const componentsWithoutForm = existingComponents.filter(
                  (comp: Element) => comp.type !== BlockTypes.Form,
                );

                return {
                  ...node,
                  data: {
                    ...nodeData,
                    components: mutateComponents([...componentsWithoutForm, updatedForm]),
                  },
                };
              }

              // No Form exists - create a new Form with the input
              const newForm: Element = generateIdsForResources<Element>({
                id: '{{ID}}',
                resourceType: ResourceTypes.Element,
                category: ElementCategories.Block,
                type: BlockTypes.Form,
                version: '0.1.0',
                deprecated: false,
                display: {
                  label: 'Form',
                  image: 'assets/images/icons/form.svg',
                },
                config: {},
                components: [generatedElement],
              } as Element);

              return {
                ...node,
                data: {
                  ...nodeData,
                  components: mutateComponents([...existingComponents, newForm]),
                },
              };
            }
            return node;
          });
        }

        // For other elements (Buttons, etc.), add to existing View
        return prevNodes.map((node) => {
          if (node.id === existingViewStep.id) {
            const nodeData = node.data as {components?: Element[]} | undefined;
            const existingComponents: Element[] = cloneDeep(nodeData?.components ?? []);

            return {
              ...node,
              data: {
                ...nodeData,
                components: mutateComponents([...existingComponents, generatedElement]),
              },
            };
          }
          return node;
        });
      });

      // Schedule node internals update after state has been updated
      if (viewStepId) {
        queueMicrotask(() => {
          updateNodeInternals(viewStepId!);
        });
      }
    },
    [setNodes, updateNodeInternals, generateStepElement],
  );

  /**
   * Callback for adding an element to a form from the context menu.
   * Adds the element to the specified form.
   */
  const handleAddElementToForm = useCallback(
    (element: Element, formId: string): void => {
      // Use generateStepElement to properly apply variants and generate unique IDs
      const generatedElement: Element = generateStepElement(element);
      let viewStepId: string | null = null;

      setNodes((prevNodes: Node[]) => {
        // Find the View node that contains the form with the given formId
        const existingViewStep = prevNodes.find((node) => {
          if (node.type !== StepTypes.View) return false;
          const nodeData = node.data as {components?: Element[]} | undefined;
          const components = nodeData?.components ?? [];
          // Check if this view contains the form with the given formId
          return components.some((component: Element) => component.id === formId && component.type === BlockTypes.Form);
        });

        if (!existingViewStep) {
          return prevNodes;
        }

        viewStepId = existingViewStep.id;

        return prevNodes.map((node) => {
          if (node.id === existingViewStep.id) {
            const nodeData = node.data as {components?: Element[]} | undefined;
            const existingComponents: Element[] = cloneDeep(nodeData?.components ?? []);

            // Find the form and add the element to it
            const updatedComponents = existingComponents.map((component: Element) => {
              if (component.id === formId && component.type === BlockTypes.Form) {
                return {
                  ...component,
                  components: [...(component.components ?? []), generatedElement],
                };
              }
              return component;
            });

            return {
              ...node,
              data: {
                ...nodeData,
                components: mutateComponents(updatedComponents),
              },
            };
          }
          return node;
        });
      });

      // Schedule node internals update after state has been updated
      if (viewStepId) {
        queueMicrotask(() => {
          updateNodeInternals(viewStepId!);
        });
      }
    },
    [setNodes, updateNodeInternals, generateStepElement],
  );

  return {
    handleAddElementToView,
    handleAddElementToForm,
  };
};

export default useElementAddition;
