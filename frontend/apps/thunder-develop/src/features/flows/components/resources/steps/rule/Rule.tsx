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

import {useCallback, useRef, type DragEvent, type ReactElement} from 'react';
import './Rule.scss';
import {Handle, Position, useNodeId, useNodesData, useReactFlow, type Node} from '@xyflow/react';
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
// eslint-disable-next-line @typescript-eslint/no-unused-vars
function Rule(_props: RulePropsInterface): ReactElement {
  const nodeId: string | null = useNodeId();
  const node: Pick<Node, 'data' | 'type' | 'id'> | null = useNodesData(nodeId ?? '');
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

  const ruleStep: Resource = {
    ...(typeof node?.data === 'object' && node.data !== null ? node.data : {}),
    id: node?.id ?? '',
  } as Resource;

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

export default Rule;
