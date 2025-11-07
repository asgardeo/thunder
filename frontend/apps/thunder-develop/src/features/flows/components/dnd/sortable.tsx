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

import {Box, type CSSProperties} from '@wso2/oxygen-ui';
import type {PropsWithChildren, RefObject} from 'react';
import {type UseSortableInput, useSortable} from '@dnd-kit/react/sortable';
import {RestrictToVerticalAxis} from '@dnd-kit/abstract/modifiers';
import {useDragDropManager} from '@dnd-kit/react';
import classNames from 'classnames';
import './sortable.scss';

/**
 * Props interface of {@link Sortable}
 */
export interface SortableProps extends UseSortableInput {
  /**
   * Handle reference.
   */
  handleRef?: RefObject<HTMLElement | null>;
}

/**
 * Sortable component.
 *
 * @param props - Props injected to the component.
 * @returns Sortable component.
 */
function Sortable({
  id,
  index,
  children = null,
  handleRef = undefined,
  collisionDetector,
  ...rest
}: PropsWithChildren<SortableProps>) {
  const {ref, isDragging, isDropTarget} = useSortable({
    collisionDetector,
    handle: handleRef,
    id,
    index,
    modifiers: [RestrictToVerticalAxis],
    ...rest,
  });

  const manager = useDragDropManager();
  const isDragActive = manager?.dragOperation?.status?.dragging ?? false;
  const dragSource = manager?.dragOperation?.source;
  // Get the source index from the data property where it's stored
  const dragSourceIndex = (dragSource as {index?: number} | undefined)?.index;
  // Check if this is a reordering operation (dragging within the list) vs new item from resource panel
  const isReorderingOperation = (dragSource?.data as {isReordering?: boolean} | undefined)?.isReordering === true;

  // Determine if the drop indicator should be shown above this element
  // Show indicator when: dragging is active, this element is the drop target,
  // and we're not the element being dragged
  const showDropIndicator = isDragActive && isDropTarget && !isDragging;

  // Determine indicator position (before or after this element)
  // For reordering: If dragging from below (higher index) to above (lower index), show indicator at top
  // For new items from resource panel: Always show indicator at top (insert before)
  const showIndicatorBefore =
    showDropIndicator &&
    (isReorderingOperation
      ? typeof dragSourceIndex === 'number' && typeof index === 'number' && dragSourceIndex > index
      : true); // For new items, always show before
  const showIndicatorAfter =
    showDropIndicator &&
    isReorderingOperation &&
    typeof dragSourceIndex === 'number' &&
    typeof index === 'number' &&
    dragSourceIndex < index;

  const elementStyle: CSSProperties = {
    opacity: isDragging ? 0.4 : 1,
    transform: isDragging ? 'scale(1.01)' : 'none',
    transition: isDragging ? 'none' : 'all 0.2s ease',
  };

  return (
    <Box
      ref={ref}
      sx={{height: '100%', width: '100%', ...elementStyle}}
      className={classNames('dnd-sortable', {
        'is-dragging': isDragging,
        'show-drop-indicator-before': showIndicatorBefore,
        'show-drop-indicator-after': showIndicatorAfter,
      })}
    >
      {children}
    </Box>
  );
}

export default Sortable;
