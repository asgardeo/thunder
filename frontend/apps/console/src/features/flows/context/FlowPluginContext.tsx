/**
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
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

import type {Edge, Node} from '@xyflow/react';
import type {Context} from 'react';
import {createContext} from 'react';
import type {Properties} from '../models/base';
import type {Element} from '../models/elements';
import type {Template} from '../models/templates';

/**
 * Typed handler signatures for each plugin event.
 */
export type PropertyChangeHandler = (
  propertyKey: string,
  newValue: unknown,
  element: Element,
  stepId: string,
) => boolean;
export type PropertyPanelOpenHandler = (resource: Element, properties: Properties, stepId: string) => boolean;
export type ElementFilterHandler = (element: Element) => boolean;
export type EdgeDeleteHandler = (deletedEdges: Edge[]) => boolean;
export type NodeDeleteHandler = (deleted: Node[]) => boolean;
export type NodeElementDeleteHandler = (stepId: string, element: Element) => boolean;
export type TemplateLoadHandler = (template: Template) => boolean;

/**
 * Props interface of {@link FlowPluginContext}
 *
 * Provides a typed, React-scoped plugin registry for interceptor-style event handling
 * within the flow builder. Unlike FlowEventsContext (fire-and-forget), plugin handlers
 * return boolean values that are aggregated with `.every()` to support interception.
 */
export interface FlowPluginContextProps {
  /** Subscribe to property change events. Returns an unsubscribe function. */
  onPropertyChange: (handler: PropertyChangeHandler) => () => void;
  /** Emit property change event. Returns true if all handlers returned true. */
  emitPropertyChange: (propertyKey: string, newValue: unknown, element: Element, stepId: string) => boolean;

  /** Subscribe to property panel open events. Returns an unsubscribe function. */
  onPropertyPanelOpen: (handler: PropertyPanelOpenHandler) => () => void;
  /** Emit property panel open event. Returns true if all handlers returned true. */
  emitPropertyPanelOpen: (resource: Element, properties: Properties, stepId: string) => boolean;

  /** Subscribe to element filter events. Returns an unsubscribe function. */
  onElementFilter: (handler: ElementFilterHandler) => () => void;
  /** Emit element filter event. Returns true if all handlers returned true (element should be shown). */
  emitElementFilter: (element: Element) => boolean;

  /** Subscribe to edge delete events. Returns an unsubscribe function. */
  onEdgeDelete: (handler: EdgeDeleteHandler) => () => void;
  /** Emit edge delete event. Returns true if all handlers returned true. */
  emitEdgeDelete: (deletedEdges: Edge[]) => boolean;

  /** Subscribe to node delete events. Returns an unsubscribe function. */
  onNodeDelete: (handler: NodeDeleteHandler) => () => void;
  /** Emit node delete event. Returns true if all handlers returned true. */
  emitNodeDelete: (deleted: Node[]) => boolean;

  /** Subscribe to node element delete events. Returns an unsubscribe function. */
  onNodeElementDelete: (handler: NodeElementDeleteHandler) => () => void;
  /** Emit node element delete event. Returns true if all handlers returned true. */
  emitNodeElementDelete: (stepId: string, element: Element) => boolean;

  /** Subscribe to template load events. Returns an unsubscribe function. */
  onTemplateLoad: (handler: TemplateLoadHandler) => () => void;
  /** Emit template load event. Returns true if all handlers returned true. */
  emitTemplateLoad: (template: Template) => boolean;
}

const FlowPluginContext: Context<FlowPluginContextProps | undefined> = createContext<
  FlowPluginContextProps | undefined
>(undefined);

FlowPluginContext.displayName = 'FlowPluginContext';

export default FlowPluginContext;
