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

/**
 * Event types in the flow builder extension.
 */
const EventTypes = {
  /**
   * Event triggered when a node is removed.
   */
  ON_NODE_DELETE: 'onNodeDelete',
  /**
   * Event triggered when an edge is removed.
   */
  ON_EDGE_DELETE: 'onEdgeDelete',
  /**
   * Event triggered when an element of a node is deleted.
   */
  ON_NODE_ELEMENT_DELETE: 'onNodeElementDelete',
  /**
   * Event triggered when a flow is updated.
   */
  ON_FLOW_UPDATE: 'onFlowUpdate',
  /**
   * Event triggered when the property panel is opened.
   */
  ON_PROPERTY_PANEL_OPEN: 'onPropertyPanelOpen',
  /**
   * Event triggered when a property is changed.
   */
  ON_PROPERTY_CHANGE: 'onPropertyChange',
  /**
   * Event triggered before a node element is rendered.
   */
  ON_NODE_ELEMENT_RENDER: 'onNodeElementRender',
  /**
   * Event triggered to filter node elements.
   */
  ON_NODE_ELEMENT_FILTER: 'onNodeElementFilter',
  /**
   * Event triggered when a template is loaded.
   */
  ON_TEMPLATE_LOAD: 'onTemplateLoad',
  /**
   * Event triggered when a custom nodes are registered to the flow.
   */
  ON_CUSTOM_NODE_REGISTER: 'onCustomNodeRegister',
} as const;

export default EventTypes;
export type EventTypes = (typeof EventTypes)[keyof typeof EventTypes];
