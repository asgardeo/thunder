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

import {useDroppable, useDragOperation, type UseDroppableInput} from '@dnd-kit/react';
import {Box, type BoxProps} from '@wso2/oxygen-ui';
import {memo, type PropsWithChildren, type ReactElement, type ReactNode, useMemo, Children} from 'react';
import {pointerIntersection} from '@dnd-kit/collision';
import {useSortable} from '@dnd-kit/react/sortable';

/**
 * Keyframe animation for drop indicator pulse effect.
 * Defined once as a constant to avoid recreation on every render.
 */
const DROP_INDICATOR_KEYFRAMES = {
  '@keyframes dropIndicatorPulse': {
    '0%, 100%': {
      opacity: 1,
    },
    '50%': {
      opacity: 0.6,
    },
  },
};

/**
 * Props interface of {@link Droppable}
 */
export type DroppableProps = UseDroppableInput<Record<string, unknown>> &
  BoxProps & {
    /** Minimum height of the invisible end-zone drop target. Defaults to 40. */
    bottomZoneMinHeight?: number;
  };

/**
 * Props interface for DroppablePresentation
 */
interface DroppablePresentationProps {
  children: ReactNode;
  className?: string;
  sx?: BoxProps['sx'];
}

/**
 * Memoized presentation component for Droppable content.
 * PERFORMANCE FIX: Based on dnd-kit issue #389 - separate presentation from hook
 * This prevents children from re-rendering when useDroppable causes parent re-renders.
 * @see https://github.com/clauderic/dnd-kit/issues/389
 *
 * @param props - Props injected to the component.
 * @returns DroppablePresentation component.
 */
function DroppablePresentation({children, className = undefined, sx = {}}: DroppablePresentationProps): ReactElement {
  return (
    <Box
      className={className}
      sx={{
        display: 'inline-flex',
        flexDirection: 'column',
        gap: '10px',
        height: '100%',
        width: '100%',
        ...sx,
      }}
    >
      {children}
    </Box>
  );
}

const MemoizedDroppablePresentation = memo(DroppablePresentation);

/**
 * Invisible sortable at bottom to show drop indicator for inserting at end.
 */
function BottomZone({
  id,
  count,
  accept = undefined,
  droppableData = undefined,
  minHeight = 40,
}: {
  id: string;
  count: number;
  accept?: UseDroppableInput['accept'];
  droppableData?: Record<string, unknown>;
  minHeight?: number;
}) {
  const {ref, sortable, isDropTarget} = useSortable({
    id: `${id}-end`,
    index: count,
    accept,
    data: {...droppableData, isEndZone: true, isReordering: true},
  });

  const {source} = useDragOperation();
  const show = useMemo(
    () => source && isDropTarget && sortable.accepts(source),
    [source, isDropTarget, sortable],
  );

  const dropIndicatorStyles = useMemo(() => ({
    minHeight: `${minHeight}px`,
    width: '100%',
    position: 'relative' as const,
    marginTop: '4px',
    marginBottom: '4px',
    ...(show && {
      '&::before': {
        content: '""',
        position: 'absolute' as const,
        left: 0,
        right: 0,
        top: '-8px',
        height: '3px',
        backgroundColor: 'primary.main',
        borderRadius: '2px',
        zIndex: 100,
        pointerEvents: 'none' as const,
        animation: 'dropIndicatorPulse 1s ease-in-out infinite',
      },
      '&::after': {
        content: '""',
        position: 'absolute' as const,
        left: '-4px',
        right: '-4px',
        top: '-16px',
        height: 'calc(8px * 2)',
        backgroundColor: 'rgba(var(--mui-palette-primary-mainChannel) / 0.1)',
        borderRadius: '4px',
        zIndex: 99,
        pointerEvents: 'none' as const,
      },
    }),
    ...DROP_INDICATOR_KEYFRAMES,
  }), [show]);

  return <Box ref={ref} sx={dropIndicatorStyles} />;
}

const MemoizedBottomZone = memo(BottomZone);

/**
 * Droppable component.
 * PERFORMANCE FIX: Uses memoized presentation pattern from dnd-kit issue #389
 * The useDroppable hook causes re-renders during drag operations, but by memoizing
 * the children separately, those re-renders become cheap (only the wrapper re-renders).
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
  bottomZoneMinHeight = 40,
  ...rest
}: PropsWithChildren<DroppableProps>): ReactElement {
  const {ref, droppable, isDropTarget} = useDroppable<Record<string, unknown>>({
    accept,
    collisionDetector,
    data,
    id,
    ...rest,
  });

  const {source} = useDragOperation();
  const count = useMemo(() => Children.count(children), [children]);

  const canAcceptDrop = useMemo(() => {
    if (!source) return true;
    return droppable.accepts(source);
  }, [source, droppable]);

  const isDraggingOver = useMemo(() => Boolean(source && isDropTarget), [source, isDropTarget]);

  const dropStyles = useMemo(() => {
    if (!isDraggingOver) return {};

    if (canAcceptDrop) {
      return {
        backgroundColor: 'rgba(var(--mui-palette-success-mainChannel) / 0.1)',
        border: '2px dashed',
        borderColor: 'success.light',
      };
    }

    return {
      backgroundColor: 'rgba(var(--mui-palette-error-mainChannel) / 0.1)',
      border: '2px dashed',
      borderColor: 'error.light',
    };
  }, [isDraggingOver, canAcceptDrop]);

  return (
    <Box
      ref={ref as BoxProps['ref']}
      className={className}
      sx={{
        display: 'inline-flex',
        flexDirection: 'column',
        height: '100%',
        width: '100%',
        borderRadius: '4px',
        ...dropStyles,
      }}
    >
      <MemoizedDroppablePresentation sx={sx}>{children}</MemoizedDroppablePresentation>
      <MemoizedBottomZone id={id} count={count} accept={accept} droppableData={data} minHeight={bottomZoneMinHeight} />
    </Box>
  );
}

export default memo(Droppable);
