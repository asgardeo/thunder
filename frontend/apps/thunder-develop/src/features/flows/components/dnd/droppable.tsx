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

import {useDragDropManager, useDroppable, type UseDroppableInput} from '@dnd-kit/react';
import {Box, type BoxProps} from '@wso2/oxygen-ui';
import {memo, type PropsWithChildren, type ReactElement, useRef, useCallback, useSyncExternalStore} from 'react';
import {pointerIntersection} from '@dnd-kit/collision';
import './droppable.scss';

/**
 * Resource type for drag and drop operations.
 */
interface Resource {
  type: string;
}

// PERFORMANCE: Global drag state managed outside React to avoid per-component state updates
// This prevents ALL Droppable components from re-rendering on every drag move
let globalDraggedResource: Resource | null = null;
const dragStateListeners = new Set<() => void>();

function setGlobalDraggedResource(resource: Resource | null): void {
  if (globalDraggedResource?.type !== resource?.type) {
    globalDraggedResource = resource;
    // Notify all subscribers only when the type actually changes
    dragStateListeners.forEach((listener) => listener());
  }
}

function subscribeToGlobalDragState(callback: () => void): () => void {
  dragStateListeners.add(callback);
  return () => dragStateListeners.delete(callback);
}

function getGlobalDraggedResource(): Resource | null {
  return globalDraggedResource;
}

/**
 * Hook to subscribe to global drag state with minimal re-renders.
 * Only triggers re-render when the dragged resource type changes.
 */
function useGlobalDragState(): Resource | null {
  return useSyncExternalStore(subscribeToGlobalDragState, getGlobalDraggedResource, getGlobalDraggedResource);
}

/**
 * Hook to set up drag monitoring once per provider, not per Droppable.
 * Uses refs to avoid re-subscriptions.
 */
function useDragMonitorSetup(): void {
  const manager = useDragDropManager();
  const isSetupRef = useRef(false);

  // Set up listeners only once when manager is available
  if (manager && !isSetupRef.current) {
    isSetupRef.current = true;

    manager.monitor.addEventListener('dragend', () => {
      setGlobalDraggedResource(null);
    });

    manager.monitor.addEventListener('dragstart', (event) => {
      const {source} = event.operation;
      const sourceData = source?.data as {dragged?: Resource} | undefined;
      if (sourceData?.dragged) {
        setGlobalDraggedResource(sourceData.dragged);
      }
    });
  }
}

/**
 * Props interface of {@link Droppable}
 */
export type DroppableProps = UseDroppableInput<Record<string, unknown>> & BoxProps;

/**
 * Droppable component.
 *
 * @param props - Props injected to the component.
 * @returns Droppable component.
 */
function Droppable({
  id,
  children = null,
  sx = {},
  className,
  collisionDetector = pointerIntersection,
  data,
  accept,
  ...rest
}: PropsWithChildren<DroppableProps>): ReactElement {
  const {ref, isDropTarget} = useDroppable<Record<string, unknown>>({
    accept,
    collisionDetector,
    data,
    id,
    ...rest,
  });

  // PERFORMANCE: Set up drag monitoring (only runs setup once due to ref guard)
  useDragMonitorSetup();

  // PERFORMANCE: Subscribe to global drag state instead of per-component state
  // This only triggers re-render when the dragged resource TYPE changes, not on every drag move
  const draggedResource = useGlobalDragState();

  // PERFORMANCE: Memoize className computation
  const droppableClassName = useCallback(() => {
    return [
      'flow-builder-dnd-droppable',
      draggedResource && (accept as string[])?.includes(draggedResource.type) && 'allowed',
      draggedResource && !(accept as string[])?.includes(draggedResource.type) && 'disallowed',
      isDropTarget && typeof id === 'string' && id.includes((data as {stepId?: string})?.stepId ?? '') && 'is-dropping',
      className,
    ]
      .filter(Boolean)
      .join(' ');
  }, [draggedResource, accept, isDropTarget, id, data, className])();

  return (
    <Box
      ref={ref as BoxProps['ref']}
      className={droppableClassName}
      sx={{
        display: 'inline-flex',
        flexDirection: 'column',
        gap: '10px',
        height: '100%',
        transition: 'background-color 0.2s ease',
        width: '100%',
        ...sx,
      }}
    >
      {children}
    </Box>
  );
}

// PERFORMANCE: Memoize to prevent re-renders from parent components
export default memo(Droppable);
