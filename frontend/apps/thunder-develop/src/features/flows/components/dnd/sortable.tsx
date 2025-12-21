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
import {
  memo,
  type PropsWithChildren,
  type ReactElement,
  type ReactNode,
  type RefObject,
  useMemo,
  useSyncExternalStore,
} from 'react';
import {type UseSortableInput, useSortable} from '@dnd-kit/react/sortable';
import {RestrictToVerticalAxis} from '@dnd-kit/abstract/modifiers';
import {useDragDropManager} from '@dnd-kit/react';
import classNames from 'classnames';
import './sortable.scss';

interface DragOperationState {
  isDragging: boolean;
  sourceIndex: number | undefined;
  isReordering: boolean;
}

let globalDragOperationState: DragOperationState = {
  isDragging: false,
  sourceIndex: undefined,
  isReordering: false,
};
const dragOperationListeners = new Set<() => void>();

function setGlobalDragOperationState(state: DragOperationState): void {
  // Only notify if state actually changed
  if (
    globalDragOperationState.isDragging !== state.isDragging ||
    globalDragOperationState.sourceIndex !== state.sourceIndex ||
    globalDragOperationState.isReordering !== state.isReordering
  ) {
    globalDragOperationState = state;
    dragOperationListeners.forEach((listener) => listener());
  }
}

function subscribeToDragOperation(callback: () => void): () => void {
  dragOperationListeners.add(callback);
  return () => dragOperationListeners.delete(callback);
}

function getDragOperationState(): DragOperationState {
  return globalDragOperationState;
}

/**
 * Hook to subscribe to global drag operation state with minimal re-renders.
 */
function useGlobalDragOperationState(): DragOperationState {
  return useSyncExternalStore(subscribeToDragOperation, getDragOperationState, getDragOperationState);
}

// Use WeakMap keyed by manager instance to track setup
// This ensures listeners are added only once per manager, and handles HMR correctly
const sortableSetupManagersMap = new WeakMap<object, boolean>();

/**
 * Hook to set up drag operation monitoring once per manager instance.
 */
function useDragOperationMonitorSetup(): void {
  const manager = useDragDropManager();

  // Skip if no manager or already set up for this manager instance
  if (!manager || sortableSetupManagersMap.has(manager)) return;

  // Mark this manager as set up
  sortableSetupManagersMap.set(manager, true);

  manager.monitor.addEventListener('dragstart', (event) => {
    const {source} = event.operation;
    const sourceIndex = (source as {index?: number} | undefined)?.index;
    const isReordering = (source?.data as {isReordering?: boolean} | undefined)?.isReordering === true;

    setGlobalDragOperationState({
      isDragging: true,
      sourceIndex,
      isReordering,
    });
  });

  manager.monitor.addEventListener('dragend', () => {
    setGlobalDragOperationState({
      isDragging: false,
      sourceIndex: undefined,
      isReordering: false,
    });
  });
}

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
 * Props interface for SortablePresentation
 */
interface SortablePresentationProps {
  children: ReactNode;
  isDragging: boolean;
  showIndicatorBefore: boolean;
  showIndicatorAfter: boolean;
  elementStyle: CSSProperties;
}

/**
 * Memoized presentation component for Sortable content.
 * PERFORMANCE FIX: Based on dnd-kit issue #389 - separate presentation from hook
 * This prevents children from re-rendering when useSortable causes parent re-renders.
 * @see https://github.com/clauderic/dnd-kit/issues/389
 *
 * @param props - Props injected to the component.
 * @returns SortablePresentation component.
 */
function SortablePresentation({
  children,
  isDragging,
  showIndicatorBefore,
  showIndicatorAfter,
  elementStyle,
}: SortablePresentationProps): ReactElement {
  return (
    <Box
      sx={{height: '100%', width: '100%', ...elementStyle}}
      className={classNames('dnd-sortable-content', {
        'is-dragging': isDragging,
        'show-drop-indicator-before': showIndicatorBefore,
        'show-drop-indicator-after': showIndicatorAfter,
      })}
    >
      {children}
    </Box>
  );
}

const MemoizedSortablePresentation = memo(SortablePresentation);

/**
 * Sortable component.
 * PERFORMANCE FIX: Uses memoized presentation pattern from dnd-kit issue #389
 * The useSortable hook causes re-renders during drag operations, but by memoizing
 * the children separately, those re-renders become cheap (only the wrapper re-renders).
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

  useDragOperationMonitorSetup();

  // This only re-renders when drag state actually changes, not on every mouse move
  const {
    isDragging: isDragActive,
    sourceIndex: dragSourceIndex,
    isReordering: isReorderingOperation,
  } = useGlobalDragOperationState();

  const {showIndicatorBefore, showIndicatorAfter} = useMemo(() => {
    // Determine if the drop indicator should be shown above this element
    // Show indicator when: dragging is active, this element is the drop target,
    // and we're not the element being dragged
    const showDropIndicator = isDragActive && isDropTarget && !isDragging;

    // Determine indicator position (before or after this element)
    // For reordering: If dragging from below (higher index) to above (lower index), show indicator at top
    // For new items from resource panel: Always show indicator at top (insert before)
    const indicatorBefore =
      showDropIndicator &&
      (isReorderingOperation
        ? typeof dragSourceIndex === 'number' && typeof index === 'number' && dragSourceIndex > index
        : true); // For new items, always show before
    const indicatorAfter =
      showDropIndicator &&
      isReorderingOperation &&
      typeof dragSourceIndex === 'number' &&
      typeof index === 'number' &&
      dragSourceIndex < index;

    return {showIndicatorBefore: indicatorBefore, showIndicatorAfter: indicatorAfter};
  }, [isDragActive, isDropTarget, isDragging, isReorderingOperation, dragSourceIndex, index]);

  const elementStyle: CSSProperties = useMemo(
    () => ({
      opacity: isDragging ? 0.4 : 1,
      transform: isDragging ? 'scale(1.01)' : 'none',
      transition: isDragging ? 'none' : 'all 0.2s ease',
    }),
    [isDragging],
  );

  return (
    <Box
      ref={ref}
      className={classNames('dnd-sortable', {
        'is-dragging': isDragging,
        'show-drop-indicator-before': showIndicatorBefore,
        'show-drop-indicator-after': showIndicatorAfter,
      })}
    >
      <MemoizedSortablePresentation
        isDragging={isDragging}
        showIndicatorBefore={showIndicatorBefore}
        showIndicatorAfter={showIndicatorAfter}
        elementStyle={elementStyle}
      >
        {children}
      </MemoizedSortablePresentation>
    </Box>
  );
}

export default memo(Sortable);
