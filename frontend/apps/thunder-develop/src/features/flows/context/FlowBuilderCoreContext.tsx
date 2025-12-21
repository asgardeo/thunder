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

import type {EdgeTypes, NodeTypes} from '@xyflow/react';
import type {Context, Dispatch, FunctionComponent, ReactNode, SetStateAction} from 'react';
import {createContext} from 'react';
import {type Base} from '../models/base';
import type {FlowCompletionConfigsInterface} from '../models/flows';
import type {MetadataInterface, Claim} from '../models/metadata';
import type {Resource} from '../models/resources';
import {PreviewScreenType} from '../models/custom-text-preference';
import {type EdgeStyleTypes, EdgeStyleTypes as EdgeStyleTypesEnum} from '../models/steps';

/**
 * Props interface of {@link FlowBuilderCoreContext}
 */
export interface FlowBuilderCoreContextProps {
  /**
   * The properties of the last interacted resource.
   */
  lastInteractedResource: Base;
  /**
   * The ID of the last user interacted resource node.
   */
  lastInteractedStepId: string;
  /**
   * The wrapper for the resource properties factory.
   */
  ResourceProperties: FunctionComponent<{
    properties?: Record<string, unknown>;
    resource: Resource;
    onChange: (propertyKey: string, newValue: string | boolean | object, resource: Resource) => void;
    onVariantChange?: (variant: string, resource?: Partial<Resource>) => void;
  }>;
  /**
   * The heading for the element properties panel.
   */
  resourcePropertiesPanelHeading: ReactNode;
  /**
   * The primary i18n screen for the flow.
   */
  primaryI18nScreen: PreviewScreenType;
  /**
   * Indicates whether the element panel is open.
   */
  isResourcePanelOpen: boolean;
  /**
   * Indicates whether the element properties panel is open.
   */
  isResourcePropertiesPanelOpen: boolean;
  /**
   * Indicates whether the version history panel is open.
   */
  isVersionHistoryPanelOpen: boolean;
  /**
   * The factory for creating components.
   */
  ElementFactory: FunctionComponent<{resource?: Resource; stepId: string; [key: string]: unknown}>;
  /**
   * Function to be called when an element is dropped on the canvas.
   * @param element - The element that was dropped on the canvas.
   * @param nodeId - The ID of the node on which the element was dropped.
   */
  onResourceDropOnCanvas: (element: Base, nodeId: string) => void;
  /**
   * The set of attributes that are selected for the flow that are maintained per node.
   */
  selectedAttributes: Record<string, Claim[]>;
  /**
   * Sets the latest interacted resource inside the canvas.
   * @param resource - The resource that was interacted with.
   * @param openPanel - Whether to open the properties panel. Defaults to true.
   */
  setLastInteractedResource: (resource: Resource, openPanel?: boolean) => void;
  /**
   * Sets the active element node ID.
   */
  setLastInteractedStepId: (stepId: string) => void;
  /**
   * Sets the heading for the element properties panel.
   *
   * @param heading - The heading to set for the element properties panel.
   */
  setResourcePropertiesPanelHeading: Dispatch<SetStateAction<ReactNode>>;
  /**
   * Function to set the state of the element panel.
   *
   * @param isOpen - Boolean indicating whether the element panel should be open.
   */
  setIsResourcePanelOpen: Dispatch<SetStateAction<boolean>>;
  /**
   * Function to set the state of the element properties panel.
   *
   * @param isOpen - Boolean indicating whether the element properties panel should be open.
   */
  setIsOpenResourcePropertiesPanel: (isOpen: boolean) => void;
  /**
   * Function to set the state of the version history panel.
   */
  setIsVersionHistoryPanelOpen: Dispatch<SetStateAction<boolean>>;
  /**
   * Sets the selected attributes for the flow.
   */
  setSelectedAttributes: Dispatch<SetStateAction<Record<string, Claim[]>>>;
  /**
   * Metadata for the current flow builder.
   */
  metadata?: MetadataInterface;
  /**
   * Configured i18n text from the branding or default fallback.
   */
  i18nText?: Partial<Record<PreviewScreenType, Record<string, string>>>;
  /**
   * Indicates whether the i18n text is still loading.
   */
  i18nTextLoading?: boolean;
  /**
   * The language of the i18n text.
   */
  language?: string;
  /**
   * Sets the language for the i18n text.
   *
   * @param language - The language to set.
   */
  setLanguage?: (language: string) => void;
  /**
   * Updates an existing i18n key for the specified screen.
   *
   * @param screenType - The screen type for the i18n key.
   * @param language - The language for the i18n key.
   * @param i18nText - The i18n text to update.
   * @returns Promise indicating the success or failure of the update operation.
   */
  updateI18nKey?: (screenType: string, language: string, i18nText: Record<string, string>) => Promise<boolean>;
  /**
   * Indicates whether the i18n key related operations are in progress.
   */
  isI18nSubmitting?: boolean;
  /**
   * Function to check if a given i18n key is custom.
   *
   * @param key - The i18n key to check.
   * @param excludePrimaryScreen - Whether to exclude the primary screen from the check.
   * @returns True if the i18n key is custom, false otherwise.
   */
  isCustomI18nKey?: (key: string, excludePrimaryScreen?: boolean) => boolean;
  /**
   * Supported locales for the custom text preferences.
   */
  supportedLocales?: Record<string, {code: string; name: string; flag: string}>;
  /**
   * Indicates whether branding is enabled for the organization.
   */
  isBrandingEnabled?: boolean;
  /**
   * Flow completion configurations.
   */
  flowCompletionConfigs: FlowCompletionConfigsInterface;
  /**
   * Set the flow completion configurations.
   */
  setFlowCompletionConfigs: Dispatch<SetStateAction<FlowCompletionConfigsInterface>>;
  /**
   * Indicates whether the flow metadata is still loading.
   */
  isFlowMetadataLoading?: boolean;
  /**
   * Node types active in the flow.
   */
  flowNodeTypes: NodeTypes;
  /**
   * Edge types active in the flow.
   */
  flowEdgeTypes: EdgeTypes;
  /**
   * Function to set the node types active in the flow.
   */
  setFlowNodeTypes: Dispatch<SetStateAction<NodeTypes>>;
  /**
   * Function to set the edge types active in the flow.
   */
  setFlowEdgeTypes: Dispatch<SetStateAction<EdgeTypes>>;
  /**
   * Function to add a resource (element, widget, template, step) to the flow.
   *
   * @param resource - The resource to add.
   */
  addResourceToFlow?: (resource: Resource) => void;
  /**
   * Publishes the current flow configuration.
   * Transforms React Flow canvas data to the format expected by the backend.
   *
   * @returns Promise indicating the success or failure of the publish operation.
   */
  publishFlow?: () => Promise<boolean>;
  /**
   * Indicates whether the flow is in verbose mode (showing executors and their edges).
   * When false, execution nodes and their connected edges are hidden from the UI
   * but the data is preserved for backend operations.
   */
  isVerboseMode: boolean;
  /**
   * Function to toggle or set the verbose mode state.
   *
   * @param isVerbose - Boolean indicating whether verbose mode should be enabled.
   */
  setIsVerboseMode: Dispatch<SetStateAction<boolean>>;
  /**
   * The current edge style type for the flow canvas.
   * Controls how edges are rendered (Bézier, Smooth Step, or Step).
   */
  edgeStyle: EdgeStyleTypes;
  /**
   * Function to set the edge style type for the flow canvas.
   *
   * @param style - The edge style type to apply.
   */
  setEdgeStyle: Dispatch<SetStateAction<EdgeStyleTypes>>;
}

/**
 * Context object for managing the flow builder core context.
 */
const FlowBuilderCoreContext: Context<FlowBuilderCoreContextProps> = createContext<FlowBuilderCoreContextProps>({
  ElementFactory: () => null,
  ResourceProperties: () => null,
  flowCompletionConfigs: {},
  flowEdgeTypes: {},
  flowNodeTypes: {},
  i18nText: undefined,
  i18nTextLoading: false,
  isBrandingEnabled: false,
  isCustomI18nKey: () => false,
  isFlowMetadataLoading: false,
  isResourcePanelOpen: true,
  isResourcePropertiesPanelOpen: false,
  isVersionHistoryPanelOpen: false,
  language: '',
  lastInteractedResource: {} as Base,
  lastInteractedStepId: '',
  metadata: undefined,
  onResourceDropOnCanvas: () => {},
  primaryI18nScreen: PreviewScreenType.LOGIN,
  resourcePropertiesPanelHeading: null,
  selectedAttributes: {},
  setFlowCompletionConfigs: () => {},
  setFlowEdgeTypes: () => {},
  setFlowNodeTypes: () => {},
  setIsOpenResourcePropertiesPanel: () => {},
  setIsResourcePanelOpen: () => {},
  setIsVersionHistoryPanelOpen: () => {},
  setLanguage: () => {},
  setLastInteractedResource: () => {},
  setLastInteractedStepId: () => {},
  setResourcePropertiesPanelHeading: () => {},
  setSelectedAttributes: () => {},
  supportedLocales: {},
  addResourceToFlow: undefined,
  publishFlow: () => Promise.resolve(false),
  isVerboseMode: false,
  setIsVerboseMode: () => {},
  edgeStyle: EdgeStyleTypesEnum.SmoothStep,
  setEdgeStyle: () => {},
});

FlowBuilderCoreContext.displayName = 'FlowBuilderCoreContext';

export default FlowBuilderCoreContext;
