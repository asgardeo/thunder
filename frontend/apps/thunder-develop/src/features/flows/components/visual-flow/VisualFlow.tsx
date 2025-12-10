/**
 * Copyright (c) 2023-2025, WSO2 LLC. (https://www.wso2.com).
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
  Background,
  Controls,
  type Edge,
  type EdgeTypes,
  type Node,
  type NodeTypes,
  ReactFlow,
  type ReactFlowProps,
  SelectionMode,
} from '@xyflow/react';
import {CollisionPriority} from '@dnd-kit/abstract';
import {useColorScheme} from '@wso2/oxygen-ui';
import isEmpty from 'lodash-es/isEmpty';
import {memo, type ReactElement, useCallback, useEffect, useMemo, useRef} from 'react';
import '@xyflow/react/dist/style.css';
import Droppable from '../dnd/droppable';
import generateResourceId from '../../utils/generateResourceId';
import VisualFlowConstants from '../../constants/VisualFlowConstants';
import useFlowBuilderCore from '../../hooks/useFlowBuilderCore';
import getKnownEdgeTypes from '../../utils/getKnownEdgeTypes';
import BaseEdge from '../react-flow-overrides/BaseEdge';
import {StepTypes} from '../../models/steps';
import {
  computeObstacleBounds,
  createObstacleCacheKey,
  type CachedObstacle,
} from '../../utils/edgeCollisionUtils';
import './VisualFlow.scss';

/**
 * Props interface of {@link VisualFlow}
 */
export interface VisualFlowPropsInterface extends ReactFlowProps {
  /**
   * Custom edges to be rendered.
   */
  customEdgeTypes?: Record<string, Edge>;
  /**
   * Node types to be rendered.
   */
  nodeTypes?: NodeTypes;
}

/**
 * Wrapper component for React Flow used in the Visual Editor.
 *
 * @param props - Props injected to the component.
 * @returns Visual editor flow component.
 */
function VisualFlow({
  customEdgeTypes = {},
  nodeTypes = {},
  nodes,
  onNodesChange,
  edges,
  onEdgesChange,
  onConnect,
  onNodesDelete,
  onEdgesDelete,
  onNodeDragStop,
}: VisualFlowPropsInterface): ReactElement {
  const {setFlowNodeTypes, flowNodeTypes, setFlowEdgeTypes, flowEdgeTypes, isVerboseMode, edgeStyle, isCollisionAvoidanceEnabled} = useFlowBuilderCore();
  const {mode} = useColorScheme();

  // Use refs to store stable obstacle data that only updates after drag stops
  // This prevents edge re-renders during node dragging
  const obstaclesRef = useRef<CachedObstacle[]>([]);
  const edgesRef = useRef<Edge[]>([]);

  // Memoize execution node IDs separately - only changes when nodes change
  const executionNodeIds = useMemo(() => {
    if (!nodes) return new Set<string>();
    return new Set(
      nodes
        .filter((node) => node.type?.toUpperCase() === StepTypes.Execution)
        .map((node) => node.id)
    );
  }, [nodes]);

  // Filter nodes based on verbose mode
  // When not in verbose mode, hide execution nodes
  const filteredNodes: Node[] = useMemo(() => {
    if (isVerboseMode || !nodes) {
      return nodes ?? [];
    }
    // Use pre-computed executionNodeIds for faster filtering
    return nodes.filter((node) => !executionNodeIds.has(node.id));
  }, [nodes, isVerboseMode, executionNodeIds]);

  // Memoize the edge map separately - only changes when edges change
  const nodeEdgeMap = useMemo(() => {
    if (!edges) return new Map<string, { incoming: Edge[]; outgoing: Edge[] }>();

    const map = new Map<string, { incoming: Edge[]; outgoing: Edge[] }>();
    edges.forEach((edge) => {
      if (!map.has(edge.source)) {
        map.set(edge.source, { incoming: [], outgoing: [] });
      }
      if (!map.has(edge.target)) {
        map.set(edge.target, { incoming: [], outgoing: [] });
      }
      map.get(edge.source)!.outgoing.push(edge);
      map.get(edge.target)!.incoming.push(edge);
    });
    return map;
  }, [edges]);

  // Filter edges based on verbose mode - uses pre-computed maps for efficiency
  const filteredEdges = useMemo(() => {
    if (isVerboseMode || !edges || !nodes) {
      return edges ?? [];
    }

    // Memoized tracing function using closure over executionNodeIds and nodeEdgeMap
    // Cache traced results to avoid redundant traversals
    const traceCache = new Map<string, string[]>();

    const traceToNonExecTarget = (startExecId: string, visited: Set<string> = new Set()): string[] => {
      if (visited.has(startExecId)) return [];

      // Check cache first
      const cacheKey = `${startExecId}:${Array.from(visited).join(',')}`;
      if (traceCache.has(cacheKey)) {
        return traceCache.get(cacheKey)!;
      }

      visited.add(startExecId);
      const targets: string[] = [];
      const execEdges = nodeEdgeMap.get(startExecId);

      if (execEdges) {
        for (const outEdge of execEdges.outgoing) {
          if (executionNodeIds.has(outEdge.target)) {
            targets.push(...traceToNonExecTarget(outEdge.target, visited));
          } else {
            targets.push(outEdge.target);
          }
        }
      }

      traceCache.set(cacheKey, targets);
      return targets;
    };

    // Keep edges that don't involve execution nodes and aren't self-loops
    const directEdges = edges.filter(
      (edge) =>
        !executionNodeIds.has(edge.source) &&
        !executionNodeIds.has(edge.target) &&
        edge.source !== edge.target
    );

    // Create bypass edges more efficiently
    const bypassEdges: Edge[] = [];
    const addedBypassEdges = new Set<string>();

    for (const execNodeId of executionNodeIds) {
      const execNodeEdges = nodeEdgeMap.get(execNodeId);
      if (!execNodeEdges) continue;

      const { incoming, outgoing } = execNodeEdges;

      for (const inEdge of incoming) {
        // Skip if source is also an execution node
        if (executionNodeIds.has(inEdge.source)) continue;

        // Collect reachable non-execution targets
        const reachableTargets: string[] = [];
        for (const outEdge of outgoing) {
          if (executionNodeIds.has(outEdge.target)) {
            reachableTargets.push(...traceToNonExecTarget(outEdge.target));
          } else {
            reachableTargets.push(outEdge.target);
          }
        }

        // Create bypass edges
        for (const targetId of reachableTargets) {
          if (inEdge.source === targetId) continue; // Skip self-loops

          const bypassEdgeKey = `${inEdge.source}:${inEdge.sourceHandle ?? ''}->${targetId}`;
          if (addedBypassEdges.has(bypassEdgeKey)) continue;

          addedBypassEdges.add(bypassEdgeKey);
          bypassEdges.push({
            id: `bypass-${inEdge.source}-${targetId}-${inEdge.sourceHandle ?? 'default'}`,
            source: inEdge.source,
            sourceHandle: inEdge.sourceHandle,
            target: targetId,
            type: 'default',
          });
        }
      }
    }

    // Filter out duplicate bypass edges
    const existingDirectConnections = new Set(
      directEdges.map((e) => `${e.source}:${e.sourceHandle ?? ''}->${e.target}`)
    );

    const uniqueBypassEdges = bypassEdges.filter((edge) => {
      const connectionKey = `${edge.source}:${edge.sourceHandle ?? ''}->${edge.target}`;
      return !existingDirectConnections.has(connectionKey);
    });

    return [...directEdges, ...uniqueBypassEdges];
  }, [edges, nodes, isVerboseMode, executionNodeIds, nodeEdgeMap]);

  // Pre-compute obstacle bounds once for all edges (major performance optimization)
  // This avoids each edge component computing obstacles independently
  // We use a ref to store obstacles to prevent re-renders during drag
  const obstacleCacheKey = useMemo(
    () => (isCollisionAvoidanceEnabled && filteredNodes.length > 0 ? createObstacleCacheKey(filteredNodes) : ''),
    [isCollisionAvoidanceEnabled, filteredNodes],
  );

  // Compute obstacles but store in ref to avoid triggering edge re-renders
  // The ref is updated but doesn't cause re-renders - edges will use latest value on next render
  // Note: The return value is intentionally unused; the side effect of updating obstaclesRef is what matters
  useMemo(() => {
    if (!isCollisionAvoidanceEnabled || !obstacleCacheKey) {
      obstaclesRef.current = [];
      return;
    }
    obstaclesRef.current = computeObstacleBounds(filteredNodes);
  }, [isCollisionAvoidanceEnabled, obstacleCacheKey, filteredNodes]);

  // Update edges ref whenever filtered edges change
  // eslint-disable-next-line react-hooks/exhaustive-deps
  useEffect(() => {
    edgesRef.current = filteredEdges;
  }, [filteredEdges]);

  // Apply the user-selected edge style to all edges via data prop
  // We keep using 'base-edge' type to maintain collision avoidance,
  // but pass refs for obstacles (stable references) instead of the arrays directly
  // IMPORTANT: We only depend on filteredEdges, edgeStyle, and isCollisionAvoidanceEnabled
  // NOT on allObstacles - this prevents re-creating edge objects during node drag
  // BaseEdge will access obstaclesRef.current to get the latest obstacles lazily
  const styledEdges = useMemo(
    () =>
      filteredEdges.map((edge) => ({
        ...edge,
        type: 'base-edge', // Always use base-edge for collision avoidance
        data: {
          ...edge.data,
          edgeStyle, // Pass the style preference to BaseEdge
          isCollisionAvoidanceEnabled, // Pass collision avoidance toggle
          obstaclesRef, // Pass ref for lazy access (stable reference)
          edgesRef, // Pass ref for lazy access (stable reference)
        },
      })),
    // eslint-disable-next-line react-hooks/exhaustive-deps -- intentionally excluding allObstacles to prevent re-renders during drag
    [filteredEdges, edgeStyle, isCollisionAvoidanceEnabled],
  );

  const edgeTypes: EdgeTypes = useMemo(
    () => ({
      'base-edge': BaseEdge,
      ...getKnownEdgeTypes(),
      ...customEdgeTypes,
    }),
    [customEdgeTypes],
  );

  useEffect(() => {
    if (!isEmpty(flowNodeTypes)) {
      return;
    }

    setFlowNodeTypes(nodeTypes ?? {});
  }, [nodeTypes, flowNodeTypes, setFlowNodeTypes]);

  useEffect(() => {
    if (!isEmpty(flowEdgeTypes)) {
      return;
    }

    setFlowEdgeTypes(edgeTypes ?? {});
  }, [edgeTypes, flowEdgeTypes, setFlowEdgeTypes]);

  // Memoize fitViewOptions to prevent object recreation
  const fitViewOptions = useMemo(() => ({
    maxZoom: 0.8,
  }), []);

  // Memoize proOptions to prevent object recreation
  const proOptions = useMemo(() => ({
    hideAttribution: true,
  }), []);

  // Memoize default edge options for performance
  const defaultEdgeOptions = useMemo(() => ({
    // Disable edge animation during interactions for better performance
    animated: false,
  }), []);

  // Stable callback wrapper for onNodesChange - prevents re-renders from parent
  const handleNodesChange = useCallback(
    (changes: Parameters<NonNullable<typeof onNodesChange>>[0]) => {
      onNodesChange?.(changes);
    },
    [onNodesChange],
  );

  // Stable callback wrapper for onEdgesChange
  const handleEdgesChange = useCallback(
    (changes: Parameters<NonNullable<typeof onEdgesChange>>[0]) => {
      onEdgesChange?.(changes);
    },
    [onEdgesChange],
  );

  return (
    <Droppable
      id={generateResourceId(VisualFlowConstants.FLOW_BUILDER_CANVAS_ID)}
      type={VisualFlowConstants.FLOW_BUILDER_DROPPABLE_CANVAS_ID}
      accept={[...VisualFlowConstants.FLOW_BUILDER_CANVAS_ALLOWED_RESOURCE_TYPES]}
      collisionPriority={CollisionPriority.Low}
    >
      <ReactFlow
        fitView
        fitViewOptions={fitViewOptions}
        nodes={filteredNodes}
        edges={styledEdges}
        nodeTypes={nodeTypes}
        edgeTypes={edgeTypes}
        onConnect={onConnect}
        onNodesDelete={onNodesDelete}
        onEdgesDelete={onEdgesDelete}
        onNodeDragStop={onNodeDragStop}
        proOptions={proOptions}
        onNodesChange={handleNodesChange}
        onEdgesChange={handleEdgesChange}
        colorMode={mode}
        // PERFORMANCE OPTIMIZATIONS
        defaultEdgeOptions={defaultEdgeOptions}
        // Disable expensive features during drag
        nodesDraggable
        nodesConnectable
        elementsSelectable
        // Use partial selection mode for better performance with many nodes
        selectionMode={SelectionMode.Partial}
        // Disable zoom on scroll for smoother experience (can re-enable if needed)
        zoomOnScroll
        zoomOnPinch
        panOnScroll={false}
        // Disable auto-pan during drag for smoother dragging
        autoPanOnNodeDrag={false}
        // Optimize connection line rendering
        connectionLineStyle={connectionLineStyle}
        // Disable snap to grid for smoother dragging
        snapToGrid={false}
        // Minimum zoom level to prevent excessive node rendering
        minZoom={0.1}
        maxZoom={2}
        // Disable node extent to avoid expensive boundary calculations
        nodeExtent={undefined}
        // Optimize for large graphs
        elevateNodesOnSelect={false}
        elevateEdgesOnSelect={false}
      >
        <Controls position="top-center" orientation="horizontal" showInteractive={false} />
        <Background className="react-flow-background" gap={20} />
      </ReactFlow>
    </Droppable>
  );
}

// Stable connection line style - defined outside component to prevent recreation
const connectionLineStyle = { stroke: '#b1b1b7', strokeWidth: 2 };

// Memoize VisualFlow to prevent unnecessary re-renders from parent
export default memo(VisualFlow, (prevProps, nextProps) => {
  // Custom comparison - only re-render when necessary props change
  return (
    prevProps.nodes === nextProps.nodes &&
    prevProps.edges === nextProps.edges &&
    prevProps.nodeTypes === nextProps.nodeTypes &&
    prevProps.customEdgeTypes === nextProps.customEdgeTypes &&
    prevProps.onNodesChange === nextProps.onNodesChange &&
    prevProps.onEdgesChange === nextProps.onEdgesChange &&
    prevProps.onConnect === nextProps.onConnect &&
    prevProps.onNodesDelete === nextProps.onNodesDelete &&
    prevProps.onEdgesDelete === nextProps.onEdgesDelete &&
    prevProps.onNodeDragStop === nextProps.onNodeDragStop
  );
});
