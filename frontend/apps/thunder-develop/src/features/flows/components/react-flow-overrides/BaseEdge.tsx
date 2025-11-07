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

import {
  BaseEdge as XYFlowBaseEdge,
  EdgeLabelRenderer,
  useReactFlow,
  type EdgeProps,
  type Edge,
  getSmoothStepPath,
  getBezierPath,
} from '@xyflow/react';
import {useState, useMemo, memo, type ReactElement, type SyntheticEvent} from 'react';
import {XIcon} from '@wso2/oxygen-ui-icons-react';
import {EdgeStyleTypes, type EdgeStyleTypes as EdgeStyleType} from '../../models/steps';
import type {CachedObstacle} from '../../utils/edgeCollisionUtils';

/**
 * Props interface of {@link BaseEdge}
 */
export type BaseEdgePropsInterface = EdgeProps;

// Helper function to get corner radius based on edge style
const getCornerRadius = (pathStyle: EdgeStyleType): number => {
  if (pathStyle === EdgeStyleTypes.Step) return 0;
  if (pathStyle === EdgeStyleTypes.Bezier) return 30;
  return 20; // SmoothStep default
};

// Helper function to create a custom path that routes around obstacles
const createPathAroundObstacles = (
  sx: number,
  sy: number,
  tx: number,
  ty: number,
  routeY: number,
  pathStyle: EdgeStyleType = EdgeStyleTypes.SmoothStep,
): {path: string; labelX: number; labelY: number} => {
  const r = getCornerRadius(pathStyle);
  const minVerticalSpace = Math.max(2 * r, 40);
  const x1 = sx + Math.max(2 * r, 40);
  const x2 = tx - Math.max(2 * r, 40);

  if (x1 >= x2) {
    if (pathStyle === EdgeStyleTypes.Bezier) {
      const midX = (sx + tx) / 2;
      return {
        path: `M ${sx} ${sy} C ${midX} ${sy} ${midX} ${ty} ${tx} ${ty}`,
        labelX: midX,
        labelY: (sy + ty) / 2,
      };
    }
    return {
      path: `M ${sx} ${sy} L ${tx} ${ty}`,
      labelX: (sx + tx) / 2,
      labelY: (sy + ty) / 2,
    };
  }

  let adjustedRouteY = routeY;
  const dist1 = Math.abs(routeY - sy);
  const dist2 = Math.abs(routeY - ty);

  if (routeY >= sy && routeY >= ty) {
    const minRouteY = Math.max(sy, ty) + minVerticalSpace;
    adjustedRouteY = Math.max(routeY, minRouteY);
  } else if (routeY <= sy && routeY <= ty) {
    const maxRouteY = Math.min(sy, ty) - minVerticalSpace;
    adjustedRouteY = Math.min(routeY, maxRouteY);
  } else {
    if (dist1 < minVerticalSpace) {
      adjustedRouteY = routeY > sy ? sy + minVerticalSpace : sy - minVerticalSpace;
    }
    if (dist2 < minVerticalSpace) {
      adjustedRouteY = routeY > ty ? ty + minVerticalSpace : ty - minVerticalSpace;
    }
  }

  if (pathStyle === EdgeStyleTypes.Bezier) {
    const controlOffset = Math.min(Math.abs(adjustedRouteY - sy), Math.abs(adjustedRouteY - ty), 80);
    const midX = (sx + tx) / 2;
    const path = `M ${sx} ${sy} C ${sx + controlOffset} ${sy} ${midX} ${adjustedRouteY} ${midX} ${adjustedRouteY} C ${midX} ${adjustedRouteY} ${tx - controlOffset} ${ty} ${tx} ${ty}`;
    return {path, labelX: midX, labelY: adjustedRouteY};
  }

  let path = `M ${sx} ${sy}`;

  if (r === 0) {
    path += ` L ${x1} ${sy}`;
    path += ` L ${x1} ${adjustedRouteY}`;
    path += ` L ${x2} ${adjustedRouteY}`;
    path += ` L ${x2} ${ty}`;
    path += ` L ${tx} ${ty}`;
    return {path, labelX: (x1 + x2) / 2, labelY: adjustedRouteY};
  }

  path += ` L ${x1 - r} ${sy}`;
  if (adjustedRouteY > sy) {
    path += ` Q ${x1} ${sy} ${x1} ${sy + r}`;
    path += ` L ${x1} ${adjustedRouteY - r}`;
    path += ` Q ${x1} ${adjustedRouteY} ${x1 + r} ${adjustedRouteY}`;
  } else if (adjustedRouteY < sy) {
    path += ` Q ${x1} ${sy} ${x1} ${sy - r}`;
    path += ` L ${x1} ${adjustedRouteY + r}`;
    path += ` Q ${x1} ${adjustedRouteY} ${x1 + r} ${adjustedRouteY}`;
  } else {
    path += ` L ${x1 + r} ${sy}`;
  }

  path += ` L ${x2 - r} ${adjustedRouteY}`;

  if (ty < adjustedRouteY) {
    path += ` Q ${x2} ${adjustedRouteY} ${x2} ${adjustedRouteY - r}`;
    path += ` L ${x2} ${ty + r}`;
    path += ` Q ${x2} ${ty} ${x2 + r} ${ty}`;
  } else if (ty > adjustedRouteY) {
    path += ` Q ${x2} ${adjustedRouteY} ${x2} ${adjustedRouteY + r}`;
    path += ` L ${x2} ${ty - r}`;
    path += ` Q ${x2} ${ty} ${x2 + r} ${ty}`;
  } else {
    path += ` L ${x2 + r} ${adjustedRouteY}`;
  }

  path += ` L ${tx} ${ty}`;
  return {path, labelX: (x1 + x2) / 2, labelY: adjustedRouteY};
};

// Helper function to create a backward path (right to left)
const createBackwardPathAroundObstacles = (
  sx: number,
  sy: number,
  tx: number,
  ty: number,
  routeY: number,
  pathStyle: EdgeStyleType = EdgeStyleTypes.SmoothStep,
): {path: string; labelX: number; labelY: number} => {
  const r = getCornerRadius(pathStyle);
  const minVerticalSpace = Math.max(2 * r, 40);
  const x1 = sx + Math.max(2 * r, 40);
  const x2 = tx - Math.max(2 * r, 40);

  let adjustedRouteY = routeY;
  if (routeY >= sy && routeY >= ty) {
    const minRouteY = Math.max(sy, ty) + minVerticalSpace;
    adjustedRouteY = Math.max(routeY, minRouteY);
  } else if (routeY <= sy && routeY <= ty) {
    const maxRouteY = Math.min(sy, ty) - minVerticalSpace;
    adjustedRouteY = Math.min(routeY, maxRouteY);
  }

  if (pathStyle === EdgeStyleTypes.Bezier) {
    const controlOffset = Math.min(Math.abs(adjustedRouteY - sy), Math.abs(adjustedRouteY - ty), 80);
    const midX = (x1 + x2) / 2;
    const path = `M ${sx} ${sy} C ${sx + controlOffset} ${sy} ${x1} ${adjustedRouteY} ${midX} ${adjustedRouteY} C ${x2} ${adjustedRouteY} ${tx - controlOffset} ${ty} ${tx} ${ty}`;
    return {path, labelX: midX, labelY: adjustedRouteY};
  }

  let path = `M ${sx} ${sy}`;

  if (r === 0) {
    path += ` L ${x1} ${sy}`;
    path += ` L ${x1} ${adjustedRouteY}`;
    path += ` L ${x2} ${adjustedRouteY}`;
    path += ` L ${x2} ${ty}`;
    path += ` L ${tx} ${ty}`;
    return {path, labelX: (x1 + x2) / 2, labelY: adjustedRouteY};
  }

  path += ` L ${x1 - r} ${sy}`;
  path += ` Q ${x1} ${sy} ${x1} ${sy + r}`;
  path += ` L ${x1} ${adjustedRouteY - r}`;
  path += ` Q ${x1} ${adjustedRouteY} ${x1 - r} ${adjustedRouteY}`;
  path += ` L ${x2 + r} ${adjustedRouteY}`;
  path += ` Q ${x2} ${adjustedRouteY} ${x2} ${adjustedRouteY - r}`;
  path += ` L ${x2} ${ty + r}`;
  path += ` Q ${x2} ${ty} ${x2 + r} ${ty}`;
  path += ` L ${tx} ${ty}`;

  return {path, labelX: (x1 + x2) / 2, labelY: adjustedRouteY};
};

/**
 * Optimized edge component that receives pre-computed obstacle data via props.
 * This component is memoized to prevent unnecessary re-renders during node drag.
 *
 * IMPORTANT: This component does NOT use useNodes() or useEdges() hooks to avoid
 * triggering re-renders of ALL edges when ANY node position changes.
 */
function BaseEdgeComponent({
  id,
  sourceX,
  sourceY,
  targetX,
  targetY,
  sourcePosition,
  targetPosition,
  label,
  style,
  deletable,
  source,
  target,
  data,
  ...rest
}: BaseEdgePropsInterface): ReactElement {
  const {deleteElements} = useReactFlow();
  const [isHovered, setIsHovered] = useState<boolean>(false);

  // Get pre-computed data from parent (VisualFlow) via edge data
  // This avoids using useNodes()/useEdges() which cause all edges to re-render
  const edgeStyle: EdgeStyleType = (data?.edgeStyle as EdgeStyleType) ?? EdgeStyleTypes.SmoothStep;
  const isCollisionAvoidanceEnabled = (data?.isCollisionAvoidanceEnabled as boolean) ?? true;

  // Use stable references from data to avoid creating new arrays on every render
  const allObstaclesFromData = data?.allObstacles as CachedObstacle[] | undefined;
  const allEdgesFromData = data?.allEdges as Edge[] | undefined;

  // Memoize the fallback empty arrays to maintain stable references
  const emptyObstacles = useMemo<CachedObstacle[]>(() => [], []);
  const emptyEdges = useMemo<Edge[]>(() => [], []);

  const allObstacles = allObstaclesFromData ?? emptyObstacles;
  const allEdges = allEdgesFromData ?? emptyEdges;

  // Calculate edge path with pre-computed obstacles
  const {edgePath, labelX, labelY} = useMemo(() => {
    // If collision avoidance is disabled, use simple paths for better performance
    if (!isCollisionAvoidanceEnabled) {
      if (edgeStyle === EdgeStyleTypes.Bezier) {
        const [path, lx, ly] = getBezierPath({
          sourcePosition,
          sourceX,
          sourceY,
          targetPosition,
          targetX,
          targetY,
        });
        return {edgePath: path, labelX: lx, labelY: ly};
      }

      const [path, lx, ly] = getSmoothStepPath({
        sourcePosition,
        sourceX,
        sourceY,
        targetPosition,
        targetX,
        targetY,
        borderRadius: edgeStyle === EdgeStyleTypes.Step ? 0 : 20,
        offset: 0,
      });
      return {edgePath: path, labelX: lx, labelY: ly};
    }

    // Detect if this is a backward/cycle edge
    const isBackwardEdge = targetX < sourceX;

    // Find parent view for backward edges
    let parentViewOfSource: CachedObstacle | null = null;
    if (isBackwardEdge) {
      parentViewOfSource = allObstacles.find((obs) => {
        if (obs.type !== 'VIEW') return false;
        return sourceX >= obs.x && sourceX <= obs.right && sourceY >= obs.y && sourceY <= obs.bottom;
      }) ?? null;
    }

    // Get obstacle nodes (excluding source and target)
    const obstacleNodes = allObstacles.filter((obs) => obs.id !== source && obs.id !== target);
    const effectiveObstacles = parentViewOfSource
      ? [...obstacleNodes, parentViewOfSource]
      : obstacleNodes;

    // Check for obstacles in the edge path
    const minX = Math.min(sourceX, targetX);
    const maxX = Math.max(sourceX, targetX);
    const minY = Math.min(sourceY, targetY);
    const maxY = Math.max(sourceY, targetY);

    const obstaclesInPath = effectiveObstacles.filter((obs) => {
      const horizontalBuffer = 50;
      const horizontalOverlap = obs.right > (minX - horizontalBuffer) && obs.x < (maxX + horizontalBuffer);
      if (!horizontalOverlap) return false;

      if (isBackwardEdge) {
        const nodeInVerticalRange = obs.bottom >= minY && obs.y <= maxY;
        const nodeBlocksPath = nodeInVerticalRange && obs.right > minX && obs.x < maxX;
        return nodeBlocksPath;
      }

      const sourceLinePassesThroughNode = sourceY >= obs.y && sourceY <= obs.bottom && obs.right > sourceX && obs.x < targetX;
      const targetLinePassesThroughNode = targetY >= obs.y && targetY <= obs.bottom && obs.right > sourceX && obs.x < targetX;
      const nodeInVerticalRange = obs.bottom >= minY && obs.y <= maxY;
      const nodeBlocksHorizontalPath = obs.x < targetX && obs.right > sourceX;
      const nodeBlocksDefaultPath = nodeInVerticalRange && nodeBlocksHorizontalPath;

      return sourceLinePassesThroughNode || targetLinePassesThroughNode || nodeBlocksDefaultPath;
    });

    // Calculate edge offset for staggering
    const edgeSpacing = 35;
    const currentEdgeIndex = allEdges.findIndex((e) => e.id === id);

    const edgesWithOverlappingPaths =
      currentEdgeIndex > 0 && (isBackwardEdge || obstaclesInPath.length > 0)
        ? allEdges.slice(0, currentEdgeIndex).filter((e) => {
            const eSourceObs = allObstacles.find((o) => o.id === e.source);
            const eTargetObs = allObstacles.find((o) => o.id === e.target);
            if (!eSourceObs || !eTargetObs) return false;

            const eSourceRight = eSourceObs.right;
            const eTargetX = eTargetObs.x;
            const eIsBackward = eTargetX < eSourceRight;

            const eMinX = Math.min(eSourceRight, eTargetX);
            const eMaxX = Math.max(eSourceRight, eTargetX);
            if (eMaxX < minX || eMinX > maxX) return false;

            if (isBackwardEdge && eIsBackward) return true;
            if (!isBackwardEdge && !eIsBackward && obstaclesInPath.length > 0) return true;
            return false;
          }).length
        : 0;

    const edgeOffset = edgesWithOverlappingPaths * edgeSpacing;

    // For backward edges with obstacles
    if (isBackwardEdge) {
      if (obstaclesInPath.length > 0) {
        const lowestObstacleBottom = Math.max(...obstaclesInPath.map((obs) => obs.bottom + 40));
        const staggeredRouteY = lowestObstacleBottom + edgeOffset;
        const result = createBackwardPathAroundObstacles(sourceX, sourceY, targetX, targetY, staggeredRouteY, edgeStyle);
        return {edgePath: result.path, labelX: result.labelX, labelY: result.labelY};
      }

      if (edgeStyle === EdgeStyleTypes.Bezier) {
        const [path, lx, ly] = getBezierPath({
          sourcePosition,
          sourceX,
          sourceY,
          targetPosition,
          targetX,
          targetY,
        });
        return {edgePath: path, labelX: lx, labelY: ly};
      }

      const [path, lx, ly] = getSmoothStepPath({
        sourcePosition,
        sourceX,
        sourceY,
        targetPosition,
        targetX,
        targetY,
        borderRadius: edgeStyle === EdgeStyleTypes.Step ? 0 : 20,
        offset: 80,
      });
      return {edgePath: path, labelX: lx, labelY: ly};
    }

    // For forward edges with obstacles
    if (obstaclesInPath.length > 0) {
      let lowestObstacleBottom = Math.max(...obstaclesInPath.map((obs) => obs.bottom + 40));

      const midX1 = sourceX + 40;
      const midX2 = targetX - 40;

      const verticalObstacles = effectiveObstacles.filter((obs) => {
        const midX1PassesThrough =
          midX1 >= obs.x &&
          midX1 <= obs.right &&
          ((sourceY <= obs.bottom && lowestObstacleBottom >= obs.y) || (obs.y <= sourceY && obs.bottom >= sourceY));

        const midX2PassesThrough =
          midX2 >= obs.x &&
          midX2 <= obs.right &&
          ((targetY <= obs.bottom && lowestObstacleBottom >= obs.y) || (obs.y <= targetY && obs.bottom >= targetY));

        return midX1PassesThrough || midX2PassesThrough;
      });

      if (verticalObstacles.length > 0) {
        const additionalBottom = Math.max(...verticalObstacles.map((obs) => obs.bottom + 40));
        lowestObstacleBottom = Math.max(lowestObstacleBottom, additionalBottom);
      }

      const staggeredRouteY = lowestObstacleBottom + edgeOffset;
      const result = createPathAroundObstacles(sourceX, sourceY, targetX, targetY, staggeredRouteY, edgeStyle);
      return {edgePath: result.path, labelX: result.labelX, labelY: result.labelY};
    }

    // No obstacles - use simple path
    if (edgeStyle === EdgeStyleTypes.Bezier) {
      const [path, lx, ly] = getBezierPath({
        sourcePosition,
        sourceX,
        sourceY,
        targetPosition,
        targetX,
        targetY,
      });
      return {edgePath: path, labelX: lx, labelY: ly};
    }

    const [path, lx, ly] = getSmoothStepPath({
      sourcePosition,
      sourceX,
      sourceY,
      targetPosition,
      targetX,
      targetY,
      borderRadius: edgeStyle === EdgeStyleTypes.Step ? 0 : 20,
      offset: 0,
    });
    return {edgePath: path, labelX: lx, labelY: ly};
  }, [sourceX, sourceY, targetX, targetY, sourcePosition, targetPosition, allObstacles, source, target, allEdges, id, edgeStyle, isCollisionAvoidanceEnabled]);

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
    <>
      <g onMouseEnter={() => setIsHovered(true)} onMouseLeave={() => setIsHovered(false)}>
        <XYFlowBaseEdge
          id={id}
          path={edgePath}
          style={{
            ...style,
            pointerEvents: 'visibleStroke',
            strokeWidth: isHovered ? 3 : 2,
            transition: 'stroke-width 0.2s ease',
          }}
          {...rest}
        />
      </g>
      <EdgeLabelRenderer>
        {label && (
          <div
            style={{
              pointerEvents: 'auto',
              position: 'absolute',
              transform: `translate(-50%, -50%) translate(${labelX}px,${labelY}px)`,
              zIndex: 1000,
            }}
            className="edge-label-renderer__deletable-edge nodrag nopan"
            onMouseEnter={() => setIsHovered(true)}
            onMouseLeave={() => setIsHovered(false)}
          >
            {label}
          </div>
        )}
        {isHovered && deletable && (
          <div
            className="edge-delete-button nodrag nopan"
            onClick={handleDelete}
            onKeyDown={handleDeleteKeyDown}
            role="button"
            tabIndex={0}
            onMouseEnter={() => setIsHovered(true)}
            onMouseLeave={() => setIsHovered(false)}
            style={{
              position: 'absolute',
              transform: `translate(-50%, -50%) translate(${labelX}px,${labelY}px)`,
              pointerEvents: 'auto',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              width: '24px',
              height: '24px',
              backgroundColor: '#ef4444',
              borderRadius: '50%',
              cursor: 'pointer',
              boxShadow: '0 2px 8px rgba(0,0,0,0.3)',
              transition: 'all 0.2s ease',
              zIndex: 10000,
            }}
            onMouseOver={(e) => {
              (e.currentTarget as HTMLElement).style.backgroundColor = '#dc2626';
              (e.currentTarget as HTMLElement).style.transform =
                `translate(-50%, -50%) translate(${labelX}px,${labelY}px) scale(1.1)`;
            }}
            onMouseOut={(e) => {
              (e.currentTarget as HTMLElement).style.backgroundColor = '#ef4444';
              (e.currentTarget as HTMLElement).style.transform =
                `translate(-50%, -50%) translate(${labelX}px,${labelY}px) scale(1)`;
            }}
            onFocus={(e) => {
              (e.currentTarget as HTMLElement).style.backgroundColor = '#dc2626';
              (e.currentTarget as HTMLElement).style.transform =
                `translate(-50%, -50%) translate(${labelX}px,${labelY}px) scale(1.1)`;
            }}
            onBlur={(e) => {
              (e.currentTarget as HTMLElement).style.backgroundColor = '#ef4444';
              (e.currentTarget as HTMLElement).style.transform =
                `translate(-50%, -50%) translate(${labelX}px,${labelY}px) scale(1)`;
            }}
          >
            <XIcon size={16} style={{color: 'white'}} />
          </div>
        )}
      </EdgeLabelRenderer>
    </>
  );
}

// Memoize the component to prevent unnecessary re-renders
// Only re-render when the edge's own props change
const BaseEdge = memo(BaseEdgeComponent, (prevProps, nextProps) =>
  // Custom comparison function for better performance
  // Only re-render if relevant props have changed
  prevProps.id === nextProps.id &&
  prevProps.sourceX === nextProps.sourceX &&
  prevProps.sourceY === nextProps.sourceY &&
  prevProps.targetX === nextProps.targetX &&
  prevProps.targetY === nextProps.targetY &&
  prevProps.sourcePosition === nextProps.sourcePosition &&
  prevProps.targetPosition === nextProps.targetPosition &&
  prevProps.label === nextProps.label &&
  prevProps.deletable === nextProps.deletable &&
  prevProps.source === nextProps.source &&
  prevProps.target === nextProps.target &&
  prevProps.data?.edgeStyle === nextProps.data?.edgeStyle &&
  prevProps.data?.isCollisionAvoidanceEnabled === nextProps.data?.isCollisionAvoidanceEnabled &&
  prevProps.data?.allObstacles === nextProps.data?.allObstacles &&
  prevProps.data?.allEdges === nextProps.data?.allEdges
);

export default BaseEdge;
