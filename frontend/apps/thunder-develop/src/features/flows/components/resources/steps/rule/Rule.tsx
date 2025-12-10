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

import {memo, useCallback, useMemo, useRef, type DragEvent, type ReactElement} from 'react';
import './Rule.scss';
import {Handle, Position, useNodeId, useReactFlow} from '@xyflow/react';
import useFlowBuilderCore from '@/features/flows/hooks/useFlowBuilderCore';
import {Box, IconButton, Tooltip, Typography} from '@wso2/oxygen-ui';
import {CrossIcon} from '@wso2/oxygen-ui-icons-react';
import type {Resource} from '@/features/flows/models/resources';
import type {CommonStepFactoryPropsInterface} from '../CommonStepFactory';

/**
 * Props interface of {@link Rule}
 */
export type RulePropsInterface = CommonStepFactoryPropsInterface;

/**
 * Representation of an empty step in the flow builder.
 *
 * @param props - Props injected to the component.
 * @returns Rule component.
 */
// PERFORMANCE: Use data prop instead of useNodesData to prevent re-renders on ANY node change
function Rule({data, id}: RulePropsInterface): ReactElement {
  const nodeId: string | null = useNodeId();
  // PERFORMANCE: Removed useNodesData hook - it caused re-renders on ANY node change
  // The `data` and `id` props already contain the node's data, passed down from React Flow
  const {deleteElements} = useReactFlow();
  const {setLastInteractedResource} = useFlowBuilderCore();

  const ref = useRef<HTMLDivElement>(null);

  const handleDragOver: (event: DragEvent<HTMLDivElement>) => void = useCallback(
    (event: DragEvent<HTMLDivElement>) => {
      event.preventDefault();
      const {dataTransfer} = event;
      if (dataTransfer) {
        dataTransfer.dropEffect = 'move';
      }
    },
    [],
  );

  const handleDrop: (e: DragEvent<HTMLDivElement>) => void = useCallback((event: DragEvent<HTMLDivElement>) => {
    event.preventDefault();
  }, []);

  // Memoize ruleStep to prevent recreation on each render
  const ruleStep: Resource = useMemo(() => ({
    ...(typeof data === 'object' && data !== null ? data : {}),
    id: id ?? nodeId ?? '',
  } as Resource), [data, id, nodeId]);

  return (
    <div ref={ref} className="flow-builder-rule" onDrop={handleDrop} onDrag={handleDragOver}>
      <Handle type="target" position={Position.Left} />
      <Box
        display="flex"
        justifyContent="space-between"
        className="flow-builder-rule-action-panel"
        onClick={() => setLastInteractedResource(ruleStep)}
      >
        <Typography variant="body2" className="flow-builder-rule-id">
          Conditional Rule
        </Typography>
        <Tooltip title="Remove">
          <IconButton
            size="small"
            onClick={() => {
              if (nodeId) {
                // eslint-disable-next-line @typescript-eslint/no-floating-promises
                deleteElements({nodes: [{id: nodeId}]});
              }
            }}
            className="flow-builder-rule-remove-button"
          >
            <CrossIcon />
          </IconButton>
        </Tooltip>
      </Box>
      <Handle type="source" position={Position.Right} id="a" />
    </div>
  );
}

// Memoize to prevent re-renders during drag operations
export default memo(Rule, (prevProps, nextProps) =>
  prevProps.id === nextProps.id &&
  prevProps.data === nextProps.data
);
