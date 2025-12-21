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

import {type UseDraggableInput, useDraggable} from '@dnd-kit/react';
import {Box} from '@wso2/oxygen-ui';
import type {BoxProps} from '@wso2/oxygen-ui';
import {type CSSProperties, type PropsWithChildren, type ReactElement} from 'react';

/**
 * Props interface of {@link Draggable}
 */
export interface DraggableProps extends UseDraggableInput {
  /**
   * Set of types that the draggable can be dropped into.
   */
  accept: string[];
}

/**
 * Draggable component.
 *
 * @param props - Props injected to the component.
 * @returns Draggable component.
 */
function Draggable({id, children = null, ...rest}: PropsWithChildren<DraggableProps>): ReactElement {
  const {ref, isDragging} = useDraggable({
    id,
    ...rest,
  });

  const style: CSSProperties = {
    opacity: isDragging ? 0.4 : 1,
    transform: isDragging ? 'scale(1.01)' : 'none',
    transition: isDragging ? 'none' : 'transform 0.2s ease, opacity 0.2s ease',
  };

  const className = ['dnd-draggable', isDragging && 'is-dragging'].filter(Boolean).join(' ');

  return (
    <Box ref={ref as BoxProps['ref']} sx={{height: '100%', width: '100%', ...style}} className={className}>
      {children}
    </Box>
  );
}

export default Draggable;
