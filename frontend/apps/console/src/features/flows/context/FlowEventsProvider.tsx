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
import {type PropsWithChildren, type ReactElement, useCallback, useMemo, useRef} from 'react';
import FlowEventsContext from './FlowEventsContext';

/**
 * Provider for the flow events context.
 * Uses ref-based handler sets so callbacks are stable and never cause re-renders.
 */
function FlowEventsProvider({children}: PropsWithChildren): ReactElement {
  const elementAddedHandlers = useRef(new Set<(type: string) => void>());
  const autoLayoutHandlers = useRef(new Set<() => void>());
  const restoreHandlers = useRef(new Set<(nodes: Node[], edges: Edge[]) => void>());

  const notifyElementAdded = useCallback((type: string): void => {
    elementAddedHandlers.current.forEach((handler) => handler(type));
  }, []);

  const onElementAdded = useCallback((handler: (type: string) => void): (() => void) => {
    elementAddedHandlers.current.add(handler);
    return () => {
      elementAddedHandlers.current.delete(handler);
    };
  }, []);

  const triggerAutoLayout = useCallback((): void => {
    autoLayoutHandlers.current.forEach((handler) => handler());
  }, []);

  const onAutoLayout = useCallback((handler: () => void): (() => void) => {
    autoLayoutHandlers.current.add(handler);
    return () => {
      autoLayoutHandlers.current.delete(handler);
    };
  }, []);

  const restoreFromHistory = useCallback((nodes: Node[], edges: Edge[]): void => {
    restoreHandlers.current.forEach((handler) => handler(nodes, edges));
  }, []);

  const onRestoreFromHistory = useCallback((handler: (nodes: Node[], edges: Edge[]) => void): (() => void) => {
    restoreHandlers.current.add(handler);
    return () => {
      restoreHandlers.current.delete(handler);
    };
  }, []);

  const value = useMemo(
    () => ({
      notifyElementAdded,
      onElementAdded,
      triggerAutoLayout,
      onAutoLayout,
      restoreFromHistory,
      onRestoreFromHistory,
    }),
    [notifyElementAdded, onElementAdded, triggerAutoLayout, onAutoLayout, restoreFromHistory, onRestoreFromHistory],
  );

  return <FlowEventsContext.Provider value={value}>{children}</FlowEventsContext.Provider>;
}

export default FlowEventsProvider;
