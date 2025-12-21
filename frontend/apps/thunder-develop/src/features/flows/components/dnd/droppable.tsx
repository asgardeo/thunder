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

import {useDroppable, type UseDroppableInput} from '@dnd-kit/react';
import {Box, type BoxProps} from '@wso2/oxygen-ui';
import {memo, type PropsWithChildren, type ReactElement, type ReactNode} from 'react';
import {pointerIntersection} from '@dnd-kit/collision';
import './droppable.scss';

/**
 * Props interface of {@link Droppable}
 */
export type DroppableProps = UseDroppableInput<Record<string, unknown>> & BoxProps;

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
      className={className ?? 'flow-builder-dnd-droppable-content'}
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
  ...rest
}: PropsWithChildren<DroppableProps>): ReactElement {
  const {ref} = useDroppable<Record<string, unknown>>({
    accept,
    collisionDetector,
    data,
    id,
    ...rest,
  });

  return (
    <Box
      ref={ref as BoxProps['ref']}
      className={className ?? 'flow-builder-dnd-droppable'}
      sx={{
        display: 'inline-flex',
        flexDirection: 'column',
        height: '100%',
        width: '100%',
      }}
    >
      <MemoizedDroppablePresentation sx={sx}>{children}</MemoizedDroppablePresentation>
    </Box>
  );
}

export default memo(Droppable);
