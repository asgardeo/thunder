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

import {getSmoothStepPath, getBezierPath, type Node, type Edge, type Position} from '@xyflow/react';
import {EdgeStyleTypes, type EdgeStyleTypes as EdgeStyleType} from '../models/steps';

// Interface for cached obstacle data
export interface CachedObstacle {
  id: string;
  x: number;
  y: number;
  width: number;
  height: number;
  right: number;
  bottom: number;
  type: string;
}

// Interface for pre-computed edge path
export interface ComputedEdgePath {
  edgePath: string;
  labelX: number;
  labelY: number;
}

// Round position to grid for stable caching (reduces recalculations during drag)
const POSITION_GRID = 10;

// Create a stable cache key from node positions (only VIEW and EXECUTION nodes)
export const createObstacleCacheKey = (nodes: Node[]): string =>
  nodes
    .filter((n) => {
      const type = n.type?.toUpperCase();
      return type === 'VIEW' || type === 'EXECUTION';
    })
    .map((n) => {
      const roundedX = Math.round(n.position.x / POSITION_GRID) * POSITION_GRID;
      const roundedY = Math.round(n.position.y / POSITION_GRID) * POSITION_GRID;
      return `${n.id}:${roundedX},${roundedY}`;
    })
    .join('|');

// Pre-compute obstacle bounds for faster collision checks
export const computeObstacleBounds = (nodes: Node[]): CachedObstacle[] =>
  nodes
    .filter((n) => {
      const type = n.type?.toUpperCase();
      return type === 'VIEW' || type === 'EXECUTION';
    })
    .map((n) => {
      const {position} = n;
      const width = n.measured?.width ?? n.width ?? 350;
      const height = n.measured?.height ?? n.height ?? 200;
      return {
        id: n.id,
        x: position.x,
        y: position.y,
        width,
        height,
        right: position.x + width,
        bottom: position.y + height,
        type: n.type?.toUpperCase() ?? '',
      };
    });

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
): ComputedEdgePath => {
  const r = getCornerRadius(pathStyle);
  const minVerticalSpace = Math.max(2 * r, 40);

  const x1 = sx + Math.max(2 * r, 40);
  const x2 = tx - Math.max(2 * r, 40);

  if (x1 >= x2) {
    if (pathStyle === EdgeStyleTypes.Bezier) {
      const midX = (sx + tx) / 2;
      return {
        edgePath: `M ${sx} ${sy} C ${midX} ${sy} ${midX} ${ty} ${tx} ${ty}`,
        labelX: midX,
        labelY: (sy + ty) / 2,
      };
    }
    return {
      edgePath: `M ${sx} ${sy} L ${tx} ${ty}`,
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
    return {edgePath: path, labelX: midX, labelY: adjustedRouteY};
  }

  let path = `M ${sx} ${sy}`;

  if (r === 0) {
    path += ` L ${x1} ${sy}`;
    path += ` L ${x1} ${adjustedRouteY}`;
    path += ` L ${x2} ${adjustedRouteY}`;
    path += ` L ${x2} ${ty}`;
    path += ` L ${tx} ${ty}`;
    return {edgePath: path, labelX: (x1 + x2) / 2, labelY: adjustedRouteY};
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
  return {edgePath: path, labelX: (x1 + x2) / 2, labelY: adjustedRouteY};
};

// Helper function to create a backward path (right to left)
const createBackwardPathAroundObstacles = (
  sx: number,
  sy: number,
  tx: number,
  ty: number,
  routeY: number,
  pathStyle: EdgeStyleType = EdgeStyleTypes.SmoothStep,
): ComputedEdgePath => {
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
    return {edgePath: path, labelX: midX, labelY: adjustedRouteY};
  }

  let path = `M ${sx} ${sy}`;

  if (r === 0) {
    path += ` L ${x1} ${sy}`;
    path += ` L ${x1} ${adjustedRouteY}`;
    path += ` L ${x2} ${adjustedRouteY}`;
    path += ` L ${x2} ${ty}`;
    path += ` L ${tx} ${ty}`;
    return {edgePath: path, labelX: (x1 + x2) / 2, labelY: adjustedRouteY};
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

  return {edgePath: path, labelX: (x1 + x2) / 2, labelY: adjustedRouteY};
};

export interface ComputeEdgePathParams {
  sourceX: number;
  sourceY: number;
  targetX: number;
  targetY: number;
  sourcePosition: Position;
  targetPosition: Position;
  source: string;
  target: string;
  edgeId: string;
  edgeStyle: EdgeStyleType;
  isCollisionAvoidanceEnabled: boolean;
  allObstacles: CachedObstacle[];
  edges: Edge[];
}

// Compute edge path with obstacle avoidance - centralized function
export const computeEdgePath = ({
  sourceX,
  sourceY,
  targetX,
  targetY,
  sourcePosition,
  targetPosition,
  source,
  target,
  edgeId,
  edgeStyle,
  isCollisionAvoidanceEnabled,
  allObstacles,
  edges,
}: ComputeEdgePathParams): ComputedEdgePath => {
  // If collision avoidance is disabled, use simple paths
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

  const isBackwardEdge = targetX < sourceX;

  let parentViewOfSource: CachedObstacle | null = null;
  if (isBackwardEdge) {
    parentViewOfSource = allObstacles.find((obs) => {
      if (obs.type !== 'VIEW') return false;
      return sourceX >= obs.x && sourceX <= obs.right && sourceY >= obs.y && sourceY <= obs.bottom;
    }) ?? null;
  }

  const obstacleNodes = allObstacles.filter((obs) => obs.id !== source && obs.id !== target);
  const effectiveObstacles = parentViewOfSource
    ? [...obstacleNodes, parentViewOfSource]
    : obstacleNodes;

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

  const edgeSpacing = 35;
  const currentEdgeIndex = edges.findIndex((e) => e.id === edgeId);

  const edgesWithOverlappingPaths =
    currentEdgeIndex > 0 && (isBackwardEdge || obstaclesInPath.length > 0)
      ? edges.slice(0, currentEdgeIndex).filter((e) => {
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

  if (isBackwardEdge) {
    if (obstaclesInPath.length > 0) {
      const lowestObstacleBottom = Math.max(...obstaclesInPath.map((obs) => obs.bottom + 40));
      const staggeredRouteY = lowestObstacleBottom + edgeOffset;
      return createBackwardPathAroundObstacles(sourceX, sourceY, targetX, targetY, staggeredRouteY, edgeStyle);
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
    return createPathAroundObstacles(sourceX, sourceY, targetX, targetY, staggeredRouteY, edgeStyle);
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
    offset: 0,
  });
  return {edgePath: path, labelX: lx, labelY: ly};
};
