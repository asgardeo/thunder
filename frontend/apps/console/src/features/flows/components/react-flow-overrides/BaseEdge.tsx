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

import {BaseEdge as XYFlowBaseEdge, EdgeLabelRenderer, useReactFlow, useNodes, type EdgeProps} from '@xyflow/react';
import {useState, useContext, type ReactElement, type SyntheticEvent} from 'react';
import {XIcon} from '@wso2/oxygen-ui-icons-react';
import {Box} from '@wso2/oxygen-ui';
import {calculateEdgePath, type EdgeStyle} from '../../utils/calculateEdgePath';
import FlowBuilderCoreContext from '../../context/FlowBuilderCoreContext';

/**
 * Props interface of {@link BaseEdge}
 */
export type BaseEdgePropsInterface = EdgeProps;

/**
 * Border radius for smooth step edges in pixels.
 */
const SMOOTH_STEP_BORDER_RADIUS = 20;

/**
 * Enhanced edge component with custom routing algorithm to avoid nodes.
 * Includes custom delete button and label functionality with hover effects.
 * Supports multiple edge styles: Bezier, Smooth Step (with rounded corners), and Step.
 */
function BaseEdge({
  id,
  source,
  target,
  sourceX,
  sourceY,
  targetX,
  targetY,
  sourcePosition,
  targetPosition,
  label,
  style,
  deletable,
  ...rest
}: BaseEdgePropsInterface): ReactElement {
  const {deleteElements} = useReactFlow();
  const [isHovered, setIsHovered] = useState<boolean>(false);
  const nodes = useNodes();
  const {edgeStyle} = useContext(FlowBuilderCoreContext);

  // Calculate smart path that routes around nodes with the selected edge style
  const {
    path: edgePath,
    centerX: labelX,
    centerY: labelY,
  } = calculateEdgePath(
    sourceX,
    sourceY,
    targetX,
    targetY,
    sourcePosition,
    targetPosition,
    nodes,
    edgeStyle as EdgeStyle,
    SMOOTH_STEP_BORDER_RADIUS,
  );

  const handleDelete = (event: SyntheticEvent) => {
    event.stopPropagation();
    deleteElements({edges: [{id}]}).catch(() => {});
  };

  const handleDeleteKeyDown = (event: React.KeyboardEvent) => {
    if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault();
      event.stopPropagation();
      deleteElements({edges: [{id}]}).catch(() => {});
    }
  };

  return (
    <g onMouseEnter={() => setIsHovered(true)} onMouseLeave={() => setIsHovered(false)}>
      {/* Invisible wider path for hover detection */}
      <path d={edgePath} fill="none" stroke="transparent" strokeWidth={20} style={{cursor: 'pointer'}} />
      <XYFlowBaseEdge
        id={id}
        path={edgePath}
        style={{
          ...style,
          strokeWidth: isHovered ? 3 : 2,
          transition: 'stroke-width 0.2s ease',
        }}
        interactionWidth={20}
        markerEnd={rest.markerEnd}
        markerStart={rest.markerStart}
      />
      <EdgeLabelRenderer>
        {label && (
          <Box
            className="nodrag nopan"
            onMouseEnter={() => setIsHovered(true)}
            onMouseLeave={() => setIsHovered(false)}
            sx={{
              pointerEvents: 'auto',
              position: 'absolute',
              transform: `translate(-50%, -50%) translate(${labelX}px, ${labelY}px)`,
              zIndex: 1000,
            }}
          >
            {label}
          </Box>
        )}
        {isHovered && deletable !== false && (
          <Box
            className="nodrag nopan"
            onClick={handleDelete}
            onKeyDown={handleDeleteKeyDown}
            role="button"
            tabIndex={0}
            aria-label="Delete edge"
            onMouseEnter={() => setIsHovered(true)}
            onMouseLeave={() => setIsHovered(false)}
            sx={{
              position: 'absolute',
              transform: `translate(-50%, -50%) translate(${labelX}px, ${labelY}px)`,
              pointerEvents: 'auto',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              width: '24px',
              height: '24px',
              backgroundColor: 'error.main',
              borderRadius: '50%',
              cursor: 'pointer',
              boxShadow: 2,
              transition: 'all 0.2s ease',
              zIndex: 10000,
              '&:hover, &:focus': {
                backgroundColor: 'error.dark',
                transform: `translate(-50%, -50%) translate(${labelX}px, ${labelY}px) scale(1.1)`,
              },
            }}
          >
            <XIcon size={16} style={{color: 'white'}} />
          </Box>
        )}
      </EdgeLabelRenderer>
    </g>
  );
}

export default BaseEdge;
