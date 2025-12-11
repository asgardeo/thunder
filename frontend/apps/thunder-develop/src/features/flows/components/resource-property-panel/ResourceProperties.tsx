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
import {useRef, useMemo, useCallback, memo, type ReactElement} from 'react';
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
 * PERFORMANCE: This component has been optimized to:
 * 1. Memoize filteredProperties to avoid expensive cloneDeep + PluginRegistry calls on every render
 * 2. Use useCallback for changeSelectedVariant to prevent recreation
 * 3. Use useRef for debounced handlePropertyChange to maintain stable reference
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

  // PERFORMANCE: Store refs for values used in debounced/callback functions
  const lastInteractedResourceRef = useRef(lastInteractedResource);
  const lastInteractedStepIdRef = useRef(lastInteractedStepId);
  const setLastInteractedResourceRef = useRef(setLastInteractedResource);
  const updateNodeDataRef = useRef(updateNodeData);

  // Keep refs in sync
  lastInteractedResourceRef.current = lastInteractedResource;
  lastInteractedStepIdRef.current = lastInteractedStepId;
  setLastInteractedResourceRef.current = setLastInteractedResource;
  updateNodeDataRef.current = updateNodeData;

  /**
   * PERFORMANCE: Memoize filtered properties to avoid expensive operations on every render.
   * Only recomputes when lastInteractedResource or lastInteractedStepId changes.
   */
  const filteredProperties = useMemo((): Properties => {
    if (!lastInteractedResource?.config) {
      return {} as Properties;
    }

    const props: Properties = Object.keys(lastInteractedResource.config).reduce(
      (acc: Record<string, unknown>, key: string) => {
        if (!ResourcePropertyPanelConstants.EXCLUDED_PROPERTIES.includes(key)) {
          acc[key] = (lastInteractedResource.config as unknown as Record<string, unknown>)[key];
        }
        return acc;
      },
      {} as Record<string, unknown>,
    ) as Properties;

    PluginRegistry.getInstance().executeSync(
      FlowEventTypes.ON_PROPERTY_PANEL_OPEN,
      lastInteractedResource,
      props,
      lastInteractedStepId,
    );

    return cloneDeep(props);
  }, [lastInteractedResource, lastInteractedStepId]);

  // PERFORMANCE: Memoize changeSelectedVariant to prevent recreation on every render
  const changeSelectedVariant = useCallback((selected: string, element?: Partial<Element>) => {
    const currentResource = lastInteractedResourceRef.current;
    const currentStepId = lastInteractedStepIdRef.current;

    if (!currentResource) return;

    let selectedVariant: Element | undefined = cloneDeep(
      currentResource.variants?.find((resource: Element) => resource.variant === selected),
    );

    if (!selectedVariant) {
      return;
    }

    if (element) {
      selectedVariant = merge(selectedVariant, element);
    }

    // Preserve the current text value when changing variants
    const currentText = (currentResource.config as {text?: string})?.text;
    if (currentText && selectedVariant.config) {
      (selectedVariant.config as {text?: string}).text = currentText;
    }

    const updateComponent = (components: Element[]): Element[] =>
      components.map((component: Element) => {
        if (component.id === currentResource.id) {
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

    updateNodeDataRef.current(currentStepId, (node: FlowNode<StepData>) => {
      const components: Element[] = updateComponent(cloneDeep(node?.data?.components) ?? []);

      setLastInteractedResourceRef.current(merge(cloneDeep(currentResource), selectedVariant));

      return {
        components,
      };
    });
  }, []);

  /**
   * PERFORMANCE: Create debounced handler using useRef to maintain stable reference.
   * Uses refs internally to access current values without stale closures.
   */
  const handlePropertyChangeRef = useRef(
    debounce(
      async (propertyKey: string, newValue: string | boolean | object, element: Element): Promise<void> => {
        const currentStepId = lastInteractedStepIdRef.current;
        const currentResource = lastInteractedResourceRef.current;

        // Execute plugins for ON_PROPERTY_CHANGE event.
        if (
          !(await PluginRegistry.getInstance().executeAsync(
            FlowEventTypes.ON_PROPERTY_CHANGE,
            propertyKey,
            newValue,
            element,
            currentStepId,
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

        updateNodeDataRef.current(currentStepId, (node: FlowNode<StepData>) => {
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
        if (propertyKey !== 'action' && element.id === lastInteractedResourceIdRef.current && currentResource) {
          const updatedResource: Resource = cloneDeep(currentResource);

          if (propertyKey.startsWith('config.')) {
            set(updatedResource, propertyKey, newValue);
          } else {
            set(updatedResource.data as Record<string, unknown>, propertyKey, newValue);
          }
          setLastInteractedResourceRef.current(updatedResource);
        }
      },
      300,
    )
  );

  // PERFORMANCE: Memoize onChange handler to prevent recreation
  const handlePropertyChange = useCallback(
    (propertyKey: string, newValue: string | boolean | object, element: Element) => {
      // eslint-disable-next-line @typescript-eslint/no-floating-promises
      handlePropertyChangeRef.current(propertyKey, newValue, element);
    },
    [],
  );

  return (
    <div className="flow-builder-element-properties">
      {lastInteractedResource ? (
        <Stack gap={2}>
          {lastInteractedResource && (
            <ResourcePropertiesComponent
              resource={lastInteractedResource}
              properties={filteredProperties as Record<string, unknown>}
              onChange={handlePropertyChange}
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

// PERFORMANCE: Memoize component to prevent re-renders from parent
export default memo(ResourceProperties);
