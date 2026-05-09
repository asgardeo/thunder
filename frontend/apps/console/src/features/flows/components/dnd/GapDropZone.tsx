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

import {useDroppable, useDragOperation, type UseDroppableInput} from '@dnd-kit/react';
import {Box} from '@wso2/oxygen-ui';
import {memo, useMemo, type ReactElement} from 'react';

/**
 * Props interface of {@link GapDropZone}
 */
export interface GapDropZoneProps {
  id: string;
  accept?: UseDroppableInput['accept'];
  data?: Record<string, unknown>;
}

/**
 * A thin drop zone placed between sortable elements to provide
 * reliable drop targets in gaps where sortable collision detection
 * may not reach (e.g. inside React Flow transformed nodes).
 */
function GapDropZone({id, accept = undefined, data = undefined}: GapDropZoneProps): ReactElement {
  const {ref, droppable, isDropTarget} = useDroppable({
    id,
    accept,
    data,
  });

  const {source} = useDragOperation();

  const canAccept = useMemo(() => {
    if (!source) return false;
    return droppable.accepts(source);
  }, [source, droppable]);

  const showIndicator = Boolean(source && isDropTarget && canAccept);

  return (
    <Box
      ref={ref}
      sx={{
        minHeight: showIndicator ? '12px' : '8px',
        width: '100%',
        position: 'relative',
        transition: 'min-height 0.15s ease',
        ...(showIndicator && {
          '&::before': {
            content: '""',
            position: 'absolute',
            left: 0,
            right: 0,
            top: '50%',
            transform: 'translateY(-50%)',
            height: '2px',
            backgroundColor: 'primary.main',
            borderRadius: '1px',
            zIndex: 100,
            pointerEvents: 'none',
          },
        }),
      }}
    />
  );
}

export default memo(GapDropZone);
