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
   * Memoize filtered properties to avoid expensive operations on every render.
   * Only recomputes when lastInteractedResource or lastInteractedStepId changes.
   */
  const filteredProperties = useMemo((): Properties => {
    if (!lastInteractedResource) {
      return {} as Properties;
    }

    const props: Properties = {} as Properties;

    // Extract top-level editable properties (new format)
    // Note: startIcon and endIcon are handled by ButtonExtendedProperties, not displayed here
    const topLevelEditableProps = ['label', 'hint', 'placeholder', 'required', 'src', 'alt'];
    const resourceWithProps = lastInteractedResource as Resource & Record<string, unknown>;
    topLevelEditableProps.forEach((key) => {
      if (resourceWithProps[key] !== undefined && !ResourcePropertyPanelConstants.EXCLUDED_PROPERTIES.includes(key)) {
        (props as Record<string, unknown>)[key] = resourceWithProps[key];
      }
    });

    // Also extract from config for backwards compatibility
    if (lastInteractedResource.config) {
      Object.keys(lastInteractedResource.config).forEach((key: string) => {
        if (!ResourcePropertyPanelConstants.EXCLUDED_PROPERTIES.includes(key)) {
          (props as Record<string, unknown>)[key] = (
            lastInteractedResource.config as unknown as Record<string, unknown>
          )[key];
        }
      });
    }

    PluginRegistry.getInstance().executeSync(
      FlowEventTypes.ON_PROPERTY_PANEL_OPEN,
      lastInteractedResource,
      props,
      lastInteractedStepId,
    );

    return cloneDeep(props);
  }, [lastInteractedResource, lastInteractedStepId]);

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

    // Preserve the current label value when changing variants (for Typography elements)
    const currentLabel = (currentResource as Element & {label?: string}).label;
    if (currentLabel !== undefined) {
      (selectedVariant as Element & {label?: string}).label = currentLabel;
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
   * Create debounced handler using useRef to maintain stable reference.
   * Uses refs internally to access current values without stale closures.
   */
  const handlePropertyChangeRef = useRef(
    debounce(async (propertyKey: string, newValue: string | boolean | object, element: Element): Promise<void> => {
      const currentStepId = lastInteractedStepIdRef.current;
      const currentResource = lastInteractedResourceRef.current;

      // Execute plugins for ON_PROPERTY_CHANGE event.
      const pluginResult = await PluginRegistry.getInstance().executeAsync(
        FlowEventTypes.ON_PROPERTY_CHANGE,
        propertyKey,
        newValue,
        element,
        currentStepId,
      );

      // If plugin handled the change (returned false), still update the resource to trigger re-render
      // This ensures properties panel updates after plugin modifications (e.g., adding confirm password field)
      if (!pluginResult) {
        if (element.id === lastInteractedResourceIdRef.current && currentResource) {
          const updatedResource: Resource = cloneDeep(currentResource);
          set(updatedResource as unknown as Record<string, unknown>, propertyKey, newValue);
          setLastInteractedResourceRef.current(updatedResource);
        }
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
        } else if (propertyKey === 'data') {
          // When propertyKey is exactly 'data', replace the entire data object
          return {...(newValue as StepData)};
        } else {
          // Strip 'data.' prefix if present since we're already setting on the data object
          const actualKey = propertyKey.startsWith('data.') ? propertyKey.slice(5) : propertyKey;
          set(data as Record<string, unknown>, actualKey, newValue);
        }

        return {...data};
      });

      // Only update lastInteractedResource if the element being changed is still the currently selected one.
      // This prevents stale updates from overwriting the heading when user switches to a different element.
      // Use the ref to get the current resource ID at execution time (not from the stale closure).
      if (propertyKey !== 'action' && element.id === lastInteractedResourceIdRef.current && currentResource) {
        const updatedResource: Resource = cloneDeep(currentResource);

        // Top-level editable properties are set directly on the resource
        const topLevelEditableProps = ['label', 'hint', 'placeholder', 'required', 'src', 'alt', 'startIcon', 'endIcon'];
        if (propertyKey === 'data') {
          // When propertyKey is exactly 'data', replace the entire data object
          updatedResource.data = newValue as StepData;
        } else if (topLevelEditableProps.includes(propertyKey)) {
          set(updatedResource as unknown as Record<string, unknown>, propertyKey, newValue);
        } else if (propertyKey.startsWith('config.') || propertyKey.startsWith('data.')) {
          // Properties starting with 'config.' or 'data.' should be set on the resource directly
          set(updatedResource, propertyKey, newValue);
        } else {
          set(updatedResource.data as Record<string, unknown>, propertyKey, newValue);
        }
        setLastInteractedResourceRef.current(updatedResource);
      }
    }, 300),
  );

  const handlePropertyChange = useCallback(
    (propertyKey: string, newValue: string | boolean | object, element: Element) => {
      // eslint-disable-next-line @typescript-eslint/no-floating-promises
      handlePropertyChangeRef.current(propertyKey, newValue, element);
    },
    [],
  );

  if (!lastInteractedResource) {
    return (
      <Typography variant="body2" color="textSecondary" sx={{padding: 2}}>
        No properties available.
      </Typography>
    );
  }

  return (
    <Stack gap={2}>
      <ResourcePropertiesComponent
        resource={lastInteractedResource}
        properties={filteredProperties as Record<string, unknown>}
        onChange={handlePropertyChange}
        onVariantChange={changeSelectedVariant}
      />
    </Stack>
  );
}

export default memo(ResourceProperties);
