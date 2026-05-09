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

/**
 * Props interface of {@link FlowEventsContext}
 *
 * Provides a typed, React-scoped event bus for cross-component communication
 * within the flow builder, replacing window-level CustomEvent dispatch/listen patterns.
 */
export interface FlowEventsContextProps {
  /**
   * Notify listeners that a flow element (step, template, widget) was added to the canvas.
   * @param type - The type of element added ('step' | 'template' | 'widget').
   */
  notifyElementAdded: (type: string) => void;
  /**
   * Register a handler for element-added events. Returns an unsubscribe function.
   */
  onElementAdded: (handler: (type: string) => void) => () => void;
  /**
   * Trigger auto-layout of the flow canvas.
   */
  triggerAutoLayout: () => void;
  /**
   * Register a handler for auto-layout events. Returns an unsubscribe function.
   */
  onAutoLayout: (handler: () => void) => () => void;
  /**
   * Restore nodes and edges from version history.
   */
  restoreFromHistory: (nodes: Node[], edges: Edge[]) => void;
  /**
   * Register a handler for restore-from-history events. Returns an unsubscribe function.
   */
  onRestoreFromHistory: (handler: (nodes: Node[], edges: Edge[]) => void) => () => void;
}

const FlowEventsContext: Context<FlowEventsContextProps | undefined> = createContext<
  FlowEventsContextProps | undefined
>(undefined);

FlowEventsContext.displayName = 'FlowEventsContext';

export default FlowEventsContext;
