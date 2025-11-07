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
} from '@xyflow/react';
import {CollisionPriority} from '@dnd-kit/abstract';
import {useColorScheme} from '@wso2/oxygen-ui';
import isEmpty from 'lodash-es/isEmpty';
import {type ReactElement, useEffect, useMemo} from 'react';
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

  // Filter nodes and edges based on verbose mode
  // When not in verbose mode, hide execution nodes and their connected edges
  const filteredNodes: Node[] = useMemo(() => {
    if (isVerboseMode || !nodes) {
      return nodes ?? [];
    }
    // Filter out execution nodes
    return nodes.filter((node) => node.type?.toUpperCase() !== StepTypes.Execution);
  }, [nodes, isVerboseMode]);

  const filteredEdges = useMemo(() => {
    if (isVerboseMode || !edges || !nodes) {
      return edges ?? [];
    }

    // Get IDs of execution nodes
    const executionNodeIds = new Set(
      nodes
        .filter((node) => node.type?.toUpperCase() === StepTypes.Execution)
        .map((node) => node.id)
    );

    // Build a map of edges for quick lookup
    // Map: nodeId -> { incoming: Edge[], outgoing: Edge[] }
    const nodeEdgeMap = new Map<string, { incoming: Edge[]; outgoing: Edge[] }>();

    edges.forEach((edge) => {
      if (!nodeEdgeMap.has(edge.source)) {
        nodeEdgeMap.set(edge.source, { incoming: [], outgoing: [] });
      }
      if (!nodeEdgeMap.has(edge.target)) {
        nodeEdgeMap.set(edge.target, { incoming: [], outgoing: [] });
      }
      nodeEdgeMap.get(edge.source)!.outgoing.push(edge);
      nodeEdgeMap.get(edge.target)!.incoming.push(edge);
    });

    // Handle chains of execution nodes (A -> Exec1 -> Exec2 -> B)
    // We need to trace through connected execution nodes to find the ultimate non-exec targets
    const traceToNonExecTarget = (startExecId: string, visited: Set<string> = new Set()): string[] => {
      if (visited.has(startExecId)) return [];
      visited.add(startExecId);

      const targets: string[] = [];
      const execEdges = nodeEdgeMap.get(startExecId);
      if (!execEdges) return targets;

      execEdges.outgoing.forEach((outEdge) => {
        if (executionNodeIds.has(outEdge.target)) {
          // Target is another execution node, trace further
          targets.push(...traceToNonExecTarget(outEdge.target, visited));
        } else {
          targets.push(outEdge.target);
        }
      });

      return targets;
    };

    // Keep edges that don't involve execution nodes
    // Also filter out self-loop edges (source === target) as these typically represent
    // actions that trigger executors without navigation (like "Resend OTP")
    const directEdges = edges.filter(
      (edge) =>
        !executionNodeIds.has(edge.source) &&
        !executionNodeIds.has(edge.target) &&
        edge.source !== edge.target // Filter out self-loops
    );

    // Create bypass edges: for each execution node, connect its sources to its targets
    // Only create bypass if there's a valid non-execution target reachable
    const bypassEdges: Edge[] = [];
    const addedBypassEdges = new Set<string>();

    executionNodeIds.forEach((execNodeId) => {
      const execNodeEdges = nodeEdgeMap.get(execNodeId);
      if (!execNodeEdges) return;

      const { incoming, outgoing } = execNodeEdges;

      // For each incoming edge to the execution node
      incoming.forEach((inEdge) => {
        // Skip if source is also an execution node (will be handled by that node's bypass)
        if (executionNodeIds.has(inEdge.source)) return;

        // Collect all reachable non-execution targets from this execution node
        const reachableTargets: string[] = [];

        outgoing.forEach((outEdge) => {
          if (executionNodeIds.has(outEdge.target)) {
            // Target is another execution node, trace through it
            reachableTargets.push(...traceToNonExecTarget(outEdge.target));
          } else {
            // Direct non-execution target
            reachableTargets.push(outEdge.target);
          }
        });

        // Only create bypass edges if there are reachable non-execution targets
        // Also skip creating bypass edges that would be self-loops (source === target)
        reachableTargets.forEach((targetId) => {
          // Skip self-loops - these are actions like "Resend OTP" that trigger an executor
          // and return to the same view without navigating elsewhere
          if (inEdge.source === targetId) return;

          const bypassEdgeKey = `${inEdge.source}:${inEdge.sourceHandle ?? ''}->${targetId}`;

          // Avoid duplicates
          if (addedBypassEdges.has(bypassEdgeKey)) return;
          addedBypassEdges.add(bypassEdgeKey);

          bypassEdges.push({
            id: `bypass-${inEdge.source}-${targetId}-${inEdge.sourceHandle ?? 'default'}`,
            source: inEdge.source,
            sourceHandle: inEdge.sourceHandle,
            target: targetId,
            type: 'default', // Use Bézier curve for bypass edges
          });
        });
      });
    });

    // Filter directEdges - these are edges that don't involve execution nodes at all
    // They should all be kept as is
    const filteredDirectEdges = directEdges;

    // Also filter out bypass edges that would create duplicate connections
    // (same source handle to same target)
    const existingDirectConnections = new Set(
      filteredDirectEdges.map((e) => `${e.source}:${e.sourceHandle ?? ''}->${e.target}`)
    );

    const uniqueBypassEdges = bypassEdges.filter((edge) => {
      const connectionKey = `${edge.source}:${edge.sourceHandle ?? ''}->${edge.target}`;
      return !existingDirectConnections.has(connectionKey);
    });

    return [...filteredDirectEdges, ...uniqueBypassEdges];
  }, [edges, nodes, isVerboseMode]);

  // Pre-compute obstacle bounds once for all edges (major performance optimization)
  // This avoids each edge component computing obstacles independently
  const obstacleCacheKey = useMemo(
    () => (isCollisionAvoidanceEnabled && filteredNodes.length > 0 ? createObstacleCacheKey(filteredNodes) : ''),
    [isCollisionAvoidanceEnabled, filteredNodes],
  );

  const allObstacles: CachedObstacle[] = useMemo(() => {
    if (!isCollisionAvoidanceEnabled || !obstacleCacheKey) {
      return [];
    }
    return computeObstacleBounds(filteredNodes);
  }, [isCollisionAvoidanceEnabled, obstacleCacheKey, filteredNodes]);

  // Apply the user-selected edge style to all edges via data prop
  // We keep using 'base-edge' type to maintain collision avoidance,
  // but pass the style preference and pre-computed obstacles through edge data
  const styledEdges = useMemo(
    () =>
      filteredEdges.map((edge) => ({
        ...edge,
        type: 'base-edge', // Always use base-edge for collision avoidance
        data: {
          ...edge.data,
          edgeStyle, // Pass the style preference to BaseEdge
          isCollisionAvoidanceEnabled, // Pass collision avoidance toggle
          allObstacles, // Pass pre-computed obstacles (avoids per-edge computation)
          allEdges: filteredEdges, // Pass edges array for staggering calculation
        },
      })),
    [filteredEdges, edgeStyle, isCollisionAvoidanceEnabled, allObstacles],
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

  return (
    <Droppable
      id={generateResourceId(VisualFlowConstants.FLOW_BUILDER_CANVAS_ID)}
      type={VisualFlowConstants.FLOW_BUILDER_DROPPABLE_CANVAS_ID}
      accept={[...VisualFlowConstants.FLOW_BUILDER_CANVAS_ALLOWED_RESOURCE_TYPES]}
      collisionPriority={CollisionPriority.Low}
    >
      <ReactFlow
        fitView
        fitViewOptions={{
          maxZoom: 0.8,
        }}
        nodes={filteredNodes}
        edges={styledEdges}
        nodeTypes={useMemo(() => nodeTypes, [nodeTypes])}
        edgeTypes={edgeTypes}
        onConnect={onConnect}
        onNodesDelete={onNodesDelete}
        onEdgesDelete={onEdgesDelete}
        onNodeDragStop={onNodeDragStop}
        proOptions={{hideAttribution: true}}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        colorMode={mode}
      >
        <Controls position="top-center" orientation="horizontal" />
        <Background className="react-flow-background" />
      </ReactFlow>
    </Droppable>
  );
}

export default VisualFlow;
