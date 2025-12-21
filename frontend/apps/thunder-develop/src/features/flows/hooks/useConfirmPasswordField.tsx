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
import set from 'lodash-es/set';
import {useCallback, useEffect} from 'react';
import PluginRegistry from '../plugins/PluginRegistry';
import FlowEventTypes from '../models/extension';
import VisualFlowConstants from '../constants/VisualFlowConstants';
import {BlockTypes, ElementTypes, InputVariants} from '../models/elements';
import FlowBuilderElementConstants from '../constants/FlowBuilderElementConstants';
import type {Element} from '../models/elements';
import type {Properties} from '../models/base';
import generateResourceId from '../utils/generateResourceId';

/**
 * Type for async property change handler with identifier property.
 * Extends the generic plugin handler signature for compatibility with PluginRegistry.
 */
type AsyncPropertyChangeHandler = ((...args: unknown[]) => Promise<boolean>) & {
  [VisualFlowConstants.FLOW_BUILDER_PLUGIN_FUNCTION_IDENTIFIER]: string;
};

/**
 * Type for async node delete handler with identifier property.
 * Extends the generic plugin handler signature for compatibility with PluginRegistry.
 */
type AsyncNodeDeleteHandler = ((...args: unknown[]) => Promise<boolean>) & {
  [VisualFlowConstants.FLOW_BUILDER_PLUGIN_FUNCTION_IDENTIFIER]: string;
};

/**
 * Type for sync property panel handler with identifier property.
 * Extends the generic plugin handler signature for compatibility with PluginRegistry.
 */
type SyncPropertyPanelHandler = ((...args: unknown[]) => boolean) & {
  [VisualFlowConstants.FLOW_BUILDER_PLUGIN_FUNCTION_IDENTIFIER]: string;
};


const useConfirmPasswordField = (): void => {
  const {getNode, updateNodeData} = useReactFlow();

  /**
   * Adds a confirm password field to the form when the password field's confirm checkbox is checked.
   *
   * @param propertyKey - The key of the property being changed.
   * @param newValue - The new value of the property.
   * @param element - The element being modified.
   * @param stepId - The ID of the step where the element is located.
   * @param renderProperties - Function to re-render properties after modification.
   * @returns Returns false if the confirm password field is added, true otherwise.
   */
  const addConfirmPasswordField = useCallback(async (
    propertyKey: string,
    newValue: unknown,
    element: Element,
    stepId: string,
  ): Promise<boolean> => {
    if (element.type === ElementTypes.Input && element.variant === InputVariants.Password) {
      if (propertyKey === 'requireConfirmation' && !newValue) {
        updateNodeData(stepId, (node: Node) => {
          if (!node.data.components) {
            return {};
          }

          const components: Element[] = cloneDeep(node.data.components) as Element[];
          let formFound = false;
          let passwordConfirmFieldIndex: number | undefined;

          components.every((component: Element) => {
            if (component.type === BlockTypes.Form) {
              component?.components?.every((c: Element) => {
                if (c.id === element.id) {
                  formFound = true;

                  return false;
                }

                return true;
              });

              if (formFound) {
                component?.components?.every((c: Element, index: number) => {
                  const cWithIdentifier = c as Element & {identifier?: string};
                  if (cWithIdentifier?.identifier === FlowBuilderElementConstants.CONFIRM_PASSWORD_IDENTIFIER && formFound) {
                    passwordConfirmFieldIndex = index;

                    return false;
                  }

                  return true;
                });
              }

              if (passwordConfirmFieldIndex === undefined) {
                return true;
              }

              component?.components?.splice(passwordConfirmFieldIndex, 1);

              return false;
            }

            return true;
          });

          return {
            components,
          };
        });

        return false;
      }

      if (propertyKey === 'requireConfirmation' && newValue) {
        updateNodeData(stepId, (node: Node) => {
          if (!node.data.components) {
            return {};
          }

          const components: Element[] = cloneDeep(node.data.components) as Element[];
          let passwordFieldIndex: number;

          components.every((component: Element) => {
            if (component.type === BlockTypes.Form) {
              component?.components?.every((c: Element, index: number) => {
                if (c.id === element.id) {
                  passwordFieldIndex = index;

                  return false;
                }

                return true;
              });

              if (passwordFieldIndex === undefined) {
                return true;
              }

              const confirmPasswordField: Element = cloneDeep(element);

              // Set properties at top level (new format)
              (confirmPasswordField as Element & {identifier?: string}).identifier = FlowBuilderElementConstants.CONFIRM_PASSWORD_IDENTIFIER;
              (confirmPasswordField as Element & {label?: string}).label = 'Confirm Password';
              (confirmPasswordField as Element & {placeholder?: string}).placeholder = 'Enter your password confirmation';
              confirmPasswordField.id = generateResourceId('field');

              component?.components?.splice(passwordFieldIndex + 1, 0, confirmPasswordField);

              return false;
            }

            return true;
          });

          return {
            components,
          };
        });

        return false;
      }
    }

    return true;
  }, [updateNodeData]);

  /**
   * Adds properties to the confirm password field when confirmation is enabled for a password field.
   *
   * @param resource - The resource element to which properties are being added.
   * @param properties - The properties object to which confirm password field properties will be added.
   * @param stepId - The ID of the step where the resource is located.
   * @returns return true.
   */
  const addConfirmPasswordFieldProperties = useCallback((resource: Element, properties: Properties, stepId: string): boolean => {
    if (resource.type === ElementTypes.Input && resource.variant === InputVariants.Password) {
      const resourceIdentifier = (resource as Element & {identifier?: string})?.identifier;
      if (resourceIdentifier === FlowBuilderElementConstants.PASSWORD_IDENTIFIER) {
        let passwordConfirmationField: Element | undefined;
        let formFound = false;
        const node = getNode(stepId);

        if (node?.data?.components) {
          const components: Element[] = cloneDeep(node.data.components) as Element[];

          components.every((component: Element) => {
            if (component.type === BlockTypes.Form) {
              component?.components?.every((c: Element) => {
                if (c.id === resource.id) {
                  formFound = true;

                  return false;
                }

                return true;
              });

              if (formFound) {
                component?.components?.every((c: Element) => {
                  const cIdentifier = (c as Element & {identifier?: string})?.identifier;
                  if (cIdentifier === FlowBuilderElementConstants.CONFIRM_PASSWORD_IDENTIFIER && formFound) {
                    passwordConfirmationField = c;

                    return false;
                  }

                  return true;
                });
              }

              if (formFound) {
                return false;
              }
            }

            return true;
          });
        }

        // Check both the node data (for existing confirm fields) and the resource property
        // The resource property is updated immediately when the checkbox is clicked,
        // while the node data update may be asynchronous
        const resourceWithConfirm = resource as Element & {requireConfirmation?: boolean};

        // If the resource has an explicit requireConfirmation value, use it (handles immediate UI updates)
        // Otherwise, fall back to checking if the confirm field exists in the node data
        let hasConfirmation: boolean;
        if (resourceWithConfirm.requireConfirmation !== undefined) {
          hasConfirmation = resourceWithConfirm.requireConfirmation;
        } else {
          hasConfirmation = !!passwordConfirmationField;
        }

        type FieldWithProps = Element & {hint?: string; label?: string; placeholder?: string};
        const fieldWithProps = passwordConfirmationField as FieldWithProps | undefined;

        Object.assign(properties, {
          requireConfirmation: hasConfirmation,
          // Show confirm field properties if confirmation is enabled
          ...(hasConfirmation && {
            confirmHint: fieldWithProps?.hint ?? '',
            confirmLabel: fieldWithProps?.label ?? 'Confirm Password',
            confirmPlaceholder: fieldWithProps?.placeholder ?? 'Enter your password confirmation',
          }),
        });
      }
    }

    return true;
  }, [getNode]);

  /**
   * Updates the properties of the confirm password field when the password field's properties are changed.
   *
   * @param propertyKey - The key of the property being changed.
   * @param newValue - The new value of the property.
   * @param element - The element being modified.
   * @param stepId - The ID of the step where the element is located.
   * @returns Returns false if the confirm password field properties are updated, true otherwise.
   */
  const updateConfirmPasswordFieldProperties = useCallback(async (
    propertyKey: string,
    newValue: unknown,
    element: Element,
    stepId: string,
  ): Promise<boolean> => {
    if (element.type === ElementTypes.Input && element.variant === InputVariants.Password) {
      if (
        propertyKey === 'confirmHint' ||
        propertyKey === 'confirmLabel' ||
        propertyKey === 'confirmPlaceholder' ||
        propertyKey === 'required'
      ) {
        updateNodeData(stepId, (node: Node) => {
          if (!node.data.components) {
            return {};
          }

          const components: Element[] = cloneDeep(node.data.components) as Element[];
          let passwordFieldIndex: number;

          components.every((component: Element) => {
            if (component.type === BlockTypes.Form) {
              component?.components?.every((c: Element, index: number) => {
                if (c.id === element.id) {
                  passwordFieldIndex = index;

                  return false;
                }

                return true;
              });

              if (passwordFieldIndex === undefined) {
                return true;
              }

              // Extract the property name (e.g., 'confirmHint' -> 'hint', 'confirmLabel' -> 'label')
              let propertyName: string = propertyKey;

              if (propertyName.startsWith('confirm')) {
                // Remove 'confirm' prefix and lowercase the first letter (e.g., 'confirmHint' -> 'hint')
                propertyName = propertyName.substring(7);
                propertyName = propertyName.charAt(0).toLowerCase() + propertyName.slice(1);
              }

              if (component.components) {
                set(component.components[passwordFieldIndex + 1], propertyName, newValue);
              }

              return false;
            }

            return true;
          });

          return {
            components,
          };
        });

        if (propertyKey !== 'required') {
          return false;
        }
      }
    }

    return true;
  }, [updateNodeData]);

  /**
   * Deletes the confirm password field from the form when the password field is removed.
   *
   * @param stepId - The ID of the step where the resource is located.
   * @param resource - The resource element to be checked.
   * @returns Returns true.
   */
  const deleteConfirmPasswordField = useCallback(async (stepId: string, resource: Element): Promise<boolean> => {
    const resourceIdentifier = (resource as Element & {identifier?: string})?.identifier;
    if (
      resource.type === ElementTypes.Input &&
      resource.variant === InputVariants.Password &&
      resourceIdentifier === FlowBuilderElementConstants.PASSWORD_IDENTIFIER
    ) {
      updateNodeData(stepId, (node: Node) => {
        if (!node.data.components) {
          return {};
        }

        const components: Element[] = cloneDeep(node.data.components) as Element[];
        let formFound = false;

        components.every((component: Element, componentIndex: number) => {
          if (component.type === BlockTypes.Form) {
            component?.components?.every((c: Element) => {
              if (c.id === resource.id) {
                formFound = true;

                return false;
              }

              return true;
            });

            if (formFound && component.components) {
              components[componentIndex] = {
                ...component,
                components: component.components.filter(
                  (c: Element) => {
                    const cIdentifier = (c as Element & {identifier?: string})?.identifier;
                    return cIdentifier !== FlowBuilderElementConstants.CONFIRM_PASSWORD_IDENTIFIER;
                  },
                ),
              };

              return false;
            }
          }

          return true;
        });

        return {
          components,
        };
      });
    }

    return true;
  }, [updateNodeData]);

  useEffect(() => {
    (addConfirmPasswordField as AsyncPropertyChangeHandler)[
      VisualFlowConstants.FLOW_BUILDER_PLUGIN_FUNCTION_IDENTIFIER
    ] = 'addConfirmPasswordField';
    (addConfirmPasswordFieldProperties as SyncPropertyPanelHandler)[
      VisualFlowConstants.FLOW_BUILDER_PLUGIN_FUNCTION_IDENTIFIER
    ] = 'addConfirmPasswordFieldProperties';
    (updateConfirmPasswordFieldProperties as AsyncPropertyChangeHandler)[
      VisualFlowConstants.FLOW_BUILDER_PLUGIN_FUNCTION_IDENTIFIER
    ] = 'updateConfirmPasswordFieldProperties';
    (deleteConfirmPasswordField as AsyncNodeDeleteHandler)[
      VisualFlowConstants.FLOW_BUILDER_PLUGIN_FUNCTION_IDENTIFIER
    ] = 'deleteConfirmPasswordField';

    PluginRegistry.getInstance().registerAsync(
      FlowEventTypes.ON_PROPERTY_CHANGE,
      addConfirmPasswordField as AsyncPropertyChangeHandler,
    );
    PluginRegistry.getInstance().registerAsync(
      FlowEventTypes.ON_PROPERTY_CHANGE,
      updateConfirmPasswordFieldProperties as AsyncPropertyChangeHandler,
    );
    PluginRegistry.getInstance().registerSync(
      FlowEventTypes.ON_PROPERTY_PANEL_OPEN,
      addConfirmPasswordFieldProperties as SyncPropertyPanelHandler,
    );
    // NOTE: renderConfirmPasswordField and skipConfirmPasswordField are not registered
    // because the custom rendering is not fully implemented (TODO in renderConfirmPasswordField).
    // The confirm password field will render as a normal INPUT element.
    PluginRegistry.getInstance().registerAsync(
      FlowEventTypes.ON_NODE_ELEMENT_DELETE,
      deleteConfirmPasswordField as AsyncNodeDeleteHandler,
    );

    return () => {
      PluginRegistry.getInstance().unregister(
        FlowEventTypes.ON_PROPERTY_CHANGE,
        (addConfirmPasswordField as AsyncPropertyChangeHandler)[
          VisualFlowConstants.FLOW_BUILDER_PLUGIN_FUNCTION_IDENTIFIER
        ],
      );
      PluginRegistry.getInstance().unregister(
        FlowEventTypes.ON_PROPERTY_CHANGE,
        (updateConfirmPasswordFieldProperties as AsyncPropertyChangeHandler)[
          VisualFlowConstants.FLOW_BUILDER_PLUGIN_FUNCTION_IDENTIFIER
        ],
      );
      PluginRegistry.getInstance().unregister(
        FlowEventTypes.ON_PROPERTY_PANEL_OPEN,
        (addConfirmPasswordFieldProperties as SyncPropertyPanelHandler)[
          VisualFlowConstants.FLOW_BUILDER_PLUGIN_FUNCTION_IDENTIFIER
        ],
      );
      // NOTE: renderConfirmPasswordField and skipConfirmPasswordField are not unregistered
      // because they were not registered (see comment above).
      PluginRegistry.getInstance().unregister(
        FlowEventTypes.ON_NODE_ELEMENT_DELETE,
        (deleteConfirmPasswordField as AsyncNodeDeleteHandler)[
          VisualFlowConstants.FLOW_BUILDER_PLUGIN_FUNCTION_IDENTIFIER
        ],
      );
    };
  }, [
    addConfirmPasswordField,
    addConfirmPasswordFieldProperties,
    deleteConfirmPasswordField,
    updateConfirmPasswordFieldProperties,
  ]);
};

export default useConfirmPasswordField;
