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

import {Stack, Typography} from '@wso2/oxygen-ui';
import {useReactFlow, type Node as FlowNode} from '@xyflow/react';
import {useRef, type ReactElement} from 'react';
import cloneDeep from 'lodash-es/cloneDeep';
import merge from 'lodash-es/merge';
import debounce from 'lodash-es/debounce';
import set from 'lodash-es/set';
import isEmpty from 'lodash-es/isEmpty';
import type {Properties} from '../../models/base';
import type {Resource} from '../../models/resources';
import './ResourceProperties.scss';
import useFlowBuilderCore from '../../hooks/useFlowBuilderCore';
import ResourcePropertyPanelConstants from '../../constants/ResourcePropertyPanelConstants';
import PluginRegistry from '../../plugins/PluginRegistry';
import FlowEventTypes from '../../models/extension';
import type {StepData} from '../../models/steps';
import type {Element} from '../../models/elements';

/**
 * Props interface of {@link ResourceProperties}
 */
export interface CommonResourcePropertiesPropsInterface {
  properties?: Properties;
  /**
   * The resource associated with the property.
   */
  resource: Resource;
  /**
   * The event handler for the property change.
   * @param propertyKey - The key of the property.
   * @param newValue - The new value of the property.
   * @param resource - The element associated with the property.
   */
  onChange: (propertyKey: string, newValue: unknown, resource: Resource) => void;
  /**
   * The event handler for the variant change.
   * @param variant - The variant of the element.
   * @param resource - Partial resource properties to override.
   */
  onVariantChange?: (variant: string, resource?: Partial<Resource>) => void;
}

/**
 * Component to generate the properties panel for the selected resource.
 *
 * @param props - Props injected to the component.
 * @returns The ResourceProperties component.
 */
function ResourceProperties(): ReactElement {
  const {updateNodeData} = useReactFlow();
  const {
    lastInteractedResource,
    setLastInteractedResource,
    ResourceProperties: ResourcePropertiesComponent,
    lastInteractedStepId,
  } = useFlowBuilderCore();

  // Use a ref to track the current resource ID for debounced functions
  const lastInteractedResourceIdRef = useRef<string>(lastInteractedResource?.id);
  lastInteractedResourceIdRef.current = lastInteractedResource?.id;

  /**
   * Get the filtered properties of the last interacted resource.
   *
   * @returns Filtered properties of the last interacted resource.
   */
  const getFilteredProperties = (): Properties => {
    const filteredProperties: Properties = Object.keys(lastInteractedResource?.config || {}).reduce(
      (acc: Record<string, unknown>, key: string) => {
        if (!ResourcePropertyPanelConstants.EXCLUDED_PROPERTIES.includes(key)) {
          acc[key] = (lastInteractedResource?.config as unknown as Record<string, unknown>)[key];
        }

        return acc;
      },
      {} as Record<string, unknown>,
    ) as Properties;

    PluginRegistry.getInstance().executeSync(
      FlowEventTypes.ON_PROPERTY_PANEL_OPEN,
      lastInteractedResource,
      filteredProperties,
      lastInteractedStepId,
    );

    return cloneDeep(filteredProperties);
  };

  const changeSelectedVariant = (selected: string, element?: Partial<Element>) => {
    let selectedVariant: Element | undefined = cloneDeep(
      lastInteractedResource?.variants?.find((resource: Element) => resource.variant === selected),
    );

    if (!selectedVariant) {
      return;
    }

    if (element) {
      selectedVariant = merge(selectedVariant, element);
    }

    // Preserve the current text value when changing variants
    const currentText = (lastInteractedResource?.config as {text?: string})?.text;
    if (currentText && selectedVariant.config) {
      (selectedVariant.config as {text?: string}).text = currentText;
    }

    const updateComponent = (components: Element[]): Element[] =>
      components.map((component: Element) => {
        if (component.id === lastInteractedResource.id) {
          return merge(cloneDeep(component), selectedVariant);
        }

        if (component.components) {
          return {
            ...component,
            components: updateComponent(component.components),
          };
        }

        return component;
      });

    updateNodeData(lastInteractedStepId, (node: FlowNode<StepData>) => {
      const components: Element[] = updateComponent(cloneDeep(node?.data?.components) ?? []);

      setLastInteractedResource(merge(cloneDeep(lastInteractedResource), selectedVariant));

      return {
        components,
      };
    });
  };

  /**
   * Handles the property change event.
   */
  const handlePropertyChange = debounce(
    async (propertyKey: string, newValue: string | boolean | object, element: Element): Promise<void> => {
      // Execute plugins for ON_PROPERTY_CHANGE event.
      if (
        !(await PluginRegistry.getInstance().executeAsync(
          FlowEventTypes.ON_PROPERTY_CHANGE,
          propertyKey,
          newValue,
          element,
          lastInteractedStepId,
        ))
      ) {
        return;
      }

      const updateComponent = (components: Element[]): Element[] =>
        components.map((component: Element) => {
          if (component.id === element.id) {
            const updated = {...component};

            set(updated, propertyKey, newValue);

            return updated;
          }

          if (component.components) {
            return {
              ...component,
              components: updateComponent(component.components),
            };
          }

          return component;
        });

      updateNodeData(lastInteractedStepId, (node: FlowNode<StepData>) => {
        const data: StepData = node?.data ?? {};

        if (!isEmpty(node?.data?.components)) {
          data.components = updateComponent(cloneDeep(node?.data?.components) ?? []);
        } else {
          set(data as Record<string, unknown>, propertyKey, newValue);
        }

        return {...data};
      });

      // Only update lastInteractedResource if the element being changed is still the currently selected one.
      // This prevents stale updates from overwriting the heading when user switches to a different element.
      // Use the ref to get the current resource ID at execution time (not from the stale closure).
      if (propertyKey !== 'action' && element.id === lastInteractedResourceIdRef.current) {
        const updatedResource: Resource = cloneDeep(lastInteractedResource);

        if (propertyKey.startsWith('config.')) {
          set(updatedResource, propertyKey, newValue);
        } else {
          set(updatedResource.data as Record<string, unknown>, propertyKey, newValue);
        }
        setLastInteractedResource(updatedResource);
      }
    },
    300,
  );

  return (
    <div className="flow-builder-element-properties">
      {lastInteractedResource ? (
        <Stack gap={2}>
          {lastInteractedResource && (
            <ResourcePropertiesComponent
              resource={lastInteractedResource}
              properties={getFilteredProperties() as Record<string, unknown>}
              onChange={(propertyKey: string, newValue: string | boolean | object, element: Element) => {
                // eslint-disable-next-line @typescript-eslint/no-floating-promises
                handlePropertyChange(propertyKey, newValue, element);
              }}
              onVariantChange={changeSelectedVariant}
            />
          )}
        </Stack>
      ) : (
        <Typography variant="body2" color="textSecondary" sx={{padding: 2}}>
          No properties available.
        </Typography>
      )}
    </div>
  );
}

export default ResourceProperties;
