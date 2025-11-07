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

import {type EdgeTypes, type NodeTypes, ReactFlowProvider, useReactFlow} from '@xyflow/react';
import merge from 'lodash-es/merge';
import startCase from 'lodash-es/startCase';
import {
  type FunctionComponent,
  type MutableRefObject,
  type PropsWithChildren,
  type ReactElement,
  type ReactNode,
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
} from 'react';
import {Stack, Typography} from '@wso2/oxygen-ui';
import useUserPreferences from '@/features/common/hooks/useUserPreferences';
import {Settings} from '@wso2/oxygen-ui-icons-react';
import ValidationProvider from './ValidationProvider';
import FlowConstants from '../constants/FlowConstants';
import FlowBuilderCoreContext from './FlowBuilderCoreContext';
import type {FlowCompletionConfigsInterface, FlowsHistoryInterface} from '../models/flows';
import {type Resource, ResourceTypes} from '../models/resources';
import {StepTypes, EdgeStyleTypes, type EdgeStyleTypes as EdgeStyleTypesType} from '../models/steps';
import type {Claim, FlowTypes} from '../models/metadata';
import {PreviewScreenType} from '../models/custom-text-preference';
import type {ValidationConfig} from './ValidationContext';

/**
 * Props interface for ElementFactory component
 */
export interface ElementFactoryProps {
  resource?: Resource;
  stepId: string;
  [key: string]: unknown;
}

/**
 * Props interface for ResourceProperties component
 */
export interface ResourcePropertiesProps {
  properties?: Record<string, unknown>;
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
  onChange: (propertyKey: string, newValue: string | boolean | object, resource: Resource) => void;
  /**
   * The event handler for the variant change.
   * @param variant - The variant of the element.
   * @param resource - Partial resource properties to override.
   */
  onVariantChange?: (variant: string, resource?: Partial<Resource>) => void;
}

/**
 * Props interface of {@link FlowBuilderCoreProvider}
 */
export interface FlowBuilderCoreProviderProps {
  /**
   * The factory for creating nodes.
   */
  ElementFactory: FunctionComponent<ElementFactoryProps>;
  /**
   * The factory for creating element properties.
   */
  ResourceProperties: FunctionComponent<ResourcePropertiesProps>;
  /**
   * The type of the flow.
   */
  flowType: FlowTypes;
  /**
   * Screen types for the i18n text.
   * First provided screen type will be used as the primary screen type.
   */
  screenTypes: PreviewScreenType[];
  /**
   * Validation configuration settings.
   */
  validationConfig?: ValidationConfig;
}

/**
 * Inner component that uses useReactFlow hook and provides the core context.
 */
function FlowContextWrapper({
  children = null,
  ElementFactory,
  ResourceProperties,
  flowType,
  screenTypes,
  validationConfig = {},
}: PropsWithChildren<FlowBuilderCoreProviderProps>): ReactElement {
  const {toObject} = useReactFlow();
  const {flows, setPreferences} = useUserPreferences();

  const [isResourcePanelOpen, setIsResourcePanelOpen] = useState<boolean>(true);
  const [isResourcePropertiesPanelOpen, setIsOpenResourcePropertiesPanel] = useState<boolean>(false);
  const [isVersionHistoryPanelOpen, setIsVersionHistoryPanelOpen] = useState<boolean>(false);
  const [resourcePropertiesPanelHeading, setResourcePropertiesPanelHeading] = useState<ReactNode>(null);
  const [lastInteractedElementInternal, setLastInteractedElementInternal] = useState<Resource>();
  const [lastInteractedStepId, setLastInteractedStepId] = useState<string>('');
  const [selectedAttributes, setSelectedAttributes] = useState<Record<string, Claim[]>>({});
  const [language, setLanguage] = useState<string>('en-US');
  const [flowCompletionConfigs, setFlowCompletionConfigs] = useState<FlowCompletionConfigsInterface>({});
  const [isAutoSaveLocalHistoryEnabled, setIsAutoSaveLocalHistoryEnabled] = useState<boolean>(true);
  const [isAutoSavingLocalHistory, setIsAutoSavingLocalHistory] = useState<boolean>(false);
  const [lastLocalHistoryAutoSaveTimestamp, setLastLocalHistoryAutoSaveTimestamp] = useState<number | null>(null);
  const [hasLocalHistory, setHasLocalHistory] = useState<boolean>(false);
  const [flowNodeTypes, setFlowNodeTypes] = useState<NodeTypes>({});
  const [flowEdgeTypes, setFlowEdgeTypes] = useState<EdgeTypes>({});
  const [isVerboseMode, setIsVerboseMode] = useState<boolean>(true);
  const [edgeStyle, setEdgeStyle] = useState<EdgeStyleTypesType>(EdgeStyleTypes.Bezier);
  const [isCollisionAvoidanceEnabled, setIsCollisionAvoidanceEnabled] = useState<boolean>(true);

  const intervalRef: MutableRefObject<NodeJS.Timeout | null> = useRef<NodeJS.Timeout | null>(null);

  // Temp variables for data fetching and error handling.
  const flowMetadata = undefined;
  const textPreference = null;
  const fallbackTextPreference = null;
  const brandingPreference = null;
  const supportedLocales = undefined;
  const textPreferenceLoading = false;
  const fallbackTextPreferenceLoading = false;
  const customTextPreferenceMetaLoading = false;
  const isFlowMetadataLoading = false;
  const isI18nSubmitting = false;
  const userName = 'unknown';

  // TODO: Implement i18n key update logic
  const updateI18nKey = useCallback(
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    async (_screenType: string, _language: string, _i18nText: Record<string, string>): Promise<boolean> =>
      Promise.resolve(false),
    [],
  );

  /**
   * Memoized drafts for the current flow type.
   */
  const localHistory: FlowsHistoryInterface[] = useMemo(() => {
    const flowsData = flows as Record<string, {history?: FlowsHistoryInterface[]}> | undefined;
    return flowsData?.[flowType]?.history ?? [];
  }, [flows, flowType]);

  /**
   * Memoized i18n text combining both text preference and fallback.
   */
  const i18nText: Partial<Record<PreviewScreenType, Record<string, string>>> = useMemo(() => {
    if (!textPreference || !fallbackTextPreference) {
      return {};
    }

    return merge({}, fallbackTextPreference, textPreference);
  }, [textPreference, fallbackTextPreference]);

  /**
   * Memoized branding enabled status based on the branding preference.
   */
  const isBrandingEnabled: boolean = useMemo(() => {
    const preference = brandingPreference as {preference?: {configs?: {isBrandingEnabled?: boolean}}} | null;
    return preference?.preference?.configs?.isBrandingEnabled ?? false;
  }, [brandingPreference]);

  /**
   * Memoized primary i18n screen based on the screen types.
   */
  const primaryI18nScreen: PreviewScreenType = useMemo(
    () => screenTypes?.[0] || PreviewScreenType.COMMON,
    [screenTypes],
  );

  /**
   * Check for existing drafts on component mount.
   */
  useEffect(() => {
    setHasLocalHistory(localHistory.length > 0);
  }, [localHistory]);

  /**
   * Manually trigger an auto-save operation.
   */
  const triggerLocalHistoryAutoSave: () => Promise<boolean> = useCallback(async (): Promise<boolean> => {
    if (!toObject || isAutoSavingLocalHistory) {
      return false;
    }

    setIsAutoSavingLocalHistory(true);

    try {
      const flowData: Record<string, unknown> = toObject();

      if (!flowData || Object.keys(flowData).length === 0) {
        return false;
      }

      // Check if the new flow data is different from the most recent history item
      const mostRecentHistoryItem: FlowsHistoryInterface = localHistory[localHistory.length - 1];

      if (mostRecentHistoryItem && JSON.stringify(mostRecentHistoryItem.flowData) === JSON.stringify(flowData)) {
        // Data is the same as the most recent item, no need to save
        return false;
      }

      const timestamp: number = Date.now();

      const history: FlowsHistoryInterface[] = [
        ...localHistory,
        {
          author: {
            userName,
          },
          flowData,
          timestamp,
        },
      ].slice(-FlowConstants.MAX_HISTORY_ITEMS);

      // Save a draft using setPreferences in localstorage.
      setPreferences({
        flows: {
          [flowType]: {
            history,
          },
        },
      });

      setLastLocalHistoryAutoSaveTimestamp(timestamp);
      setHasLocalHistory(true);

      return true;
    } catch {
      return false;
    } finally {
      setIsAutoSavingLocalHistory(false);
    }
  }, [toObject, isAutoSavingLocalHistory, flowType, setPreferences, localHistory]);

  /**
   * Clear all saved drafts for this flow type.
   */
  const clearLocalHistory: () => Promise<boolean> = useCallback(async (): Promise<boolean> => {
    try {
      // Clear drafts from preferences
      setPreferences({
        flows: {
          [flowType]: {
            history: [],
          },
        },
      });

      setHasLocalHistory(false);
      setLastLocalHistoryAutoSaveTimestamp(null);

      return true;
    } catch {
      return false;
    }
  }, [flowType, setPreferences]);

  /**
   * Restore flow from a specific history item.
   */
  const restoreFromHistory: (historyItem: FlowsHistoryInterface) => Promise<boolean> = useCallback(
    async (historyItem: FlowsHistoryInterface): Promise<boolean> => {
      try {
        const {flowData} = historyItem;

        if (!flowData?.nodes || !flowData.edges) {
          // eslint-disable-next-line no-console
          console.error('flows:core.notifications.restoreFromHistory.invalidData');

          return false;
        }

        // Extract nodes and edges from the flow data
        const {nodes, edges} = flowData as {nodes: unknown[]; edges: unknown[]};

        // Apply the restored flow data to the current flow using React Flow's methods
        // Note: This assumes the provider has access to setNodes and setEdges methods
        // which will be provided by the specific flow builder implementation
        if (typeof window !== 'undefined') {
          // Dispatch a custom event that the flow builder cores can listen to
          const restoreEvent: CustomEvent = new CustomEvent('restoreFromHistory', {
            detail: {edges, historyItem, nodes},
          });

          window.dispatchEvent(restoreEvent);

          // TODO: Show success message
        }

        return true;
      } catch (error) {
        // eslint-disable-next-line no-console
        console.error('flows:core.notifications.restoreFromHistory.error', error);

        return false;
      }
    },
    [],
  );

  /**
   * Set up auto-save interval.
   */
  useEffect(() => {
    if (isAutoSaveLocalHistoryEnabled && FlowConstants.AUTO_SAVE_INTERVAL > 0) {
      intervalRef.current = setInterval(() => {
        triggerLocalHistoryAutoSave().catch(() => {
          // Ignore errors
        });
      }, FlowConstants.AUTO_SAVE_INTERVAL);

      return () => {
        if (intervalRef.current) {
          clearInterval(intervalRef.current);
          intervalRef.current = null;
        }
      };
    }
    if (intervalRef.current) {
      clearInterval(intervalRef.current);
      intervalRef.current = null;
    }

    return undefined;
  }, [isAutoSaveLocalHistoryEnabled, triggerLocalHistoryAutoSave]);

  /**
   * Cleanup interval on unmount.
   */
  useEffect(
    () => () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
      }
    },
    [],
  );

  const setLastInteractedResource = useCallback((resource: Resource): void => {
    // TODO: Internationalize this string and get from a mapping.
    setResourcePropertiesPanelHeading(
      <Stack direction="row" className="sub-title" gap={1} alignItems="center">
        <Settings />
        <Typography variant="h5">{startCase(resource?.type?.toLowerCase())} Properties</Typography>
      </Stack>,
    );
    setLastInteractedElementInternal(resource);
    // If the element is a step node, do not open the properties panel for now.
    // TODO: Figure out if there are properties for a step.
    if (
      (resource.category === ResourceTypes.Step && resource.type === StepTypes.View) ||
      resource.resourceType === ResourceTypes.Template ||
      resource.resourceType === ResourceTypes.Widget
    ) {
      setIsOpenResourcePropertiesPanel(false);

      return;
    }

    setIsOpenResourcePropertiesPanel(true);
  }, []);

  const onResourceDropOnCanvas = useCallback(
    (resource: Resource, stepId: string): void => {
      setLastInteractedResource(resource);
      setLastInteractedStepId(stepId);
    },
    [setLastInteractedResource],
  );

  /**
   * Function to check if a given i18n key is custom.
   */
  const isCustomI18nKey: (key: string, excludePrimaryScreen?: boolean) => boolean = useCallback(
    (key: string, excludePrimaryScreen = true): boolean =>
      fallbackTextPreference
        ? Object.keys(fallbackTextPreference).every(
            (screenType: string) =>
              (screenType === primaryI18nScreen && excludePrimaryScreen) ||
              !fallbackTextPreference[screenType as PreviewScreenType][key],
          )
        : false,
    [fallbackTextPreference, primaryI18nScreen],
  );

  const contextValue = useMemo(
    () => ({
      ElementFactory,
      ResourceProperties,
      clearLocalHistory,
      flowCompletionConfigs,
      flowEdgeTypes,
      flowNodeTypes,
      hasLocalHistory,
      i18nText,
      i18nTextLoading: textPreferenceLoading || fallbackTextPreferenceLoading || customTextPreferenceMetaLoading,
      isAutoSaveLocalHistoryEnabled,
      isAutoSavingLocalHistory,
      isBrandingEnabled,
      isCustomI18nKey,
      isFlowMetadataLoading,
      isI18nSubmitting,
      isResourcePanelOpen,
      isResourcePropertiesPanelOpen,
      isVersionHistoryPanelOpen,
      language,
      lastInteractedResource: lastInteractedElementInternal!,
      lastInteractedStepId,
      lastLocalHistoryAutoSaveTimestamp,
      localHistory,
      metadata: flowMetadata,
      onResourceDropOnCanvas,
      primaryI18nScreen,
      resourcePropertiesPanelHeading,
      restoreFromHistory,
      selectedAttributes,
      setFlowCompletionConfigs,
      setFlowEdgeTypes,
      setFlowNodeTypes,
      setIsOpenResourcePropertiesPanel,
      setIsResourcePanelOpen,
      setIsVersionHistoryPanelOpen,
      setLanguage,
      setLastInteractedResource,
      setLastInteractedStepId,
      setLocalHistoryAutoSaveEnabled: setIsAutoSaveLocalHistoryEnabled,
      setResourcePropertiesPanelHeading,
      setSelectedAttributes,
      supportedLocales,
      triggerLocalHistoryAutoSave,
      updateI18nKey,
      isVerboseMode,
      setIsVerboseMode,
      edgeStyle,
      setEdgeStyle,
      isCollisionAvoidanceEnabled,
      setIsCollisionAvoidanceEnabled,
    }),
    [
      ElementFactory,
      ResourceProperties,
      clearLocalHistory,
      customTextPreferenceMetaLoading,
      fallbackTextPreferenceLoading,
      flowCompletionConfigs,
      flowEdgeTypes,
      flowMetadata,
      flowNodeTypes,
      hasLocalHistory,
      i18nText,
      isAutoSaveLocalHistoryEnabled,
      isAutoSavingLocalHistory,
      isBrandingEnabled,
      isCustomI18nKey,
      isFlowMetadataLoading,
      isI18nSubmitting,
      isResourcePanelOpen,
      isResourcePropertiesPanelOpen,
      isVersionHistoryPanelOpen,
      language,
      lastInteractedElementInternal,
      lastInteractedStepId,
      lastLocalHistoryAutoSaveTimestamp,
      localHistory,
      onResourceDropOnCanvas,
      primaryI18nScreen,
      resourcePropertiesPanelHeading,
      restoreFromHistory,
      selectedAttributes,
      setLastInteractedResource,
      supportedLocales,
      textPreferenceLoading,
      triggerLocalHistoryAutoSave,
      updateI18nKey,
      isVerboseMode,
      edgeStyle,
      isCollisionAvoidanceEnabled,
    ],
  );

  return (
    <FlowBuilderCoreContext.Provider value={contextValue}>
      <ValidationProvider validationConfig={validationConfig}>{children}</ValidationProvider>
    </FlowBuilderCoreContext.Provider>
  );
}

/**
 * FlowBuilderCoreProvider component.
 * This component provides flow builder core related context to its children.
 * It wraps the internal component with ReactFlowProvider to enable useReactFlow hook usage.
 *
 * @param props - Props injected to the component.
 * @returns The FlowBuilderCoreProvider component.
 */
function FlowBuilderCoreProvider({
  children = null,
  ...props
}: PropsWithChildren<FlowBuilderCoreProviderProps>): ReactElement {
  return (
    <ReactFlowProvider>
      <FlowContextWrapper {...props}>{children}</FlowContextWrapper>
    </ReactFlowProvider>
  );
}

export default FlowBuilderCoreProvider;
