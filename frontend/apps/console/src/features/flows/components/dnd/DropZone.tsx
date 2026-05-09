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

import {useDragOperation, type UseDroppableInput} from '@dnd-kit/react';
import {useSortable} from '@dnd-kit/react/sortable';
import {Box} from '@wso2/oxygen-ui';
import {memo, useMemo, type ReactElement} from 'react';

/**
 * Props interface of {@link DropZone}
 */
export interface DropZoneProps {
  id: string;
  index: number;
  position: 'start' | 'end';
  accept?: UseDroppableInput['accept'];
  droppableData?: Record<string, unknown>;
}

/**
 * Sortable drop zone rendered at the edges of a droppable container.
 * Used at the top (position="start") to allow dropping before the first element
 * and at the bottom (position="end") to allow dropping after the last element.
 */
function DropZone({id, index, position, accept = undefined, droppableData = undefined}: DropZoneProps): ReactElement {
  const {ref, sortable, isDropTarget} = useSortable({
    id: `${id}-${position}`,
    index,
    accept,
    data: {...droppableData, [`is${position === 'start' ? 'Start' : 'End'}Zone`]: true, isReordering: true},
  });

  const {source} = useDragOperation();
  const showIndicator = useMemo(
    () => Boolean(source && isDropTarget && sortable.accepts(source)),
    [source, isDropTarget, sortable],
  );

  const indicatorEdge = position === 'start' ? 'bottom' : 'top';

  return (
    <Box
      ref={ref}
      sx={{
        minHeight: position === 'start' ? '20px' : '40px',
        width: '100%',
        flexShrink: 0,
        position: 'relative',
        backgroundColor: showIndicator ? 'rgba(var(--oxygen-palette-success-mainChannel) / 0.1)' : 'transparent',
        transition: 'background-color 0.2s ease',
        ...(showIndicator && {
          [`&::${position === 'start' ? 'after' : 'before'}`]: {
            content: '""',
            position: 'absolute',
            left: 0,
            right: 0,
            [indicatorEdge]: 0,
            height: '3px',
            backgroundColor: 'primary.main',
            borderRadius: '2px',
            zIndex: 100,
            pointerEvents: 'none',
          },
        }),
      }}
    />
  );
}

export default memo(DropZone);
