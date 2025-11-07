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

import {useDragDropManager, useDroppable, type DragDropEventHandlers, type UseDroppableInput} from '@dnd-kit/react';
import {Box, type BoxProps} from '@wso2/oxygen-ui';
import {type PropsWithChildren, type ReactElement, useState, useEffect} from 'react';
import {pointerIntersection} from '@dnd-kit/collision';
import './droppable.scss';

/**
 * Resource type for drag and drop operations.
 */
interface Resource {
  type: string;
}

/**
 * A safe version of useDragDropMonitor that only subscribes when inside a DragDropProvider.
 * This prevents the "useDndMonitor hook was called outside of a DragDropProvider" warning.
 */
function useSafeDragDropMonitor(handlers: Partial<DragDropEventHandlers>): void {
  const manager = useDragDropManager();

  useEffect(() => {
    if (!manager) {
      return undefined;
    }

    const unsubscribers: (() => void)[] = [];

    if (handlers.onDragEnd) {
      unsubscribers.push(manager.monitor.addEventListener('dragend', handlers.onDragEnd));
    }

    if (handlers.onDragOver) {
      unsubscribers.push(manager.monitor.addEventListener('dragover', handlers.onDragOver));
    }

    if (handlers.onDragStart) {
      unsubscribers.push(manager.monitor.addEventListener('dragstart', handlers.onDragStart));
    }

    if (handlers.onDragMove) {
      unsubscribers.push(manager.monitor.addEventListener('dragmove', handlers.onDragMove));
    }

    return () => {
      unsubscribers.forEach((unsubscribe) => unsubscribe());
    };
  }, [manager, handlers]);
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

  const [draggedResource, setDraggedResource] = useState<Resource | null>(null);

  useSafeDragDropMonitor({
    onDragEnd() {
      setDraggedResource(null);
    },
    onDragOver(event) {
      const {source} = event.operation;
      const sourceData = source?.data as {dragged?: Resource} | undefined;

      if (sourceData?.dragged) {
        setDraggedResource(sourceData.dragged);
      }
    },
  });

  const droppableClassName = [
    'flow-builder-dnd-droppable',
    draggedResource && (accept as string[])?.includes(draggedResource.type) && 'allowed',
    draggedResource && !(accept as string[])?.includes(draggedResource.type) && 'disallowed',
    isDropTarget && typeof id === 'string' && id.includes((data as {stepId?: string})?.stepId ?? '') && 'is-dropping',
    className,
  ]
    .filter(Boolean)
    .join(' ');

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

export default Droppable;
