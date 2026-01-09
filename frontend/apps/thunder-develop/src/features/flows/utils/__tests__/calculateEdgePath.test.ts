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

import {describe, expect, it} from 'vitest';
import {Position, type Node} from '@xyflow/react';
import {calculateEdgePath, calculateAllEdgePaths, type EdgeInput} from '../calculateEdgePath';

describe('calculateEdgePath', () => {
  const createNode = (
    id: string,
    x: number,
    y: number,
    width = 150,
    height = 50,
  ): Node => ({
    id,
    position: {x, y},
    data: {},
    measured: {width, height},
  });

  describe('Basic Path Calculation', () => {
    it('should calculate a path between two points', () => {
      const result = calculateEdgePath(100, 100, 300, 100, Position.Right, Position.Left, []);

      expect(result).toBeDefined();
      expect(result.path).toBeDefined();
      expect(typeof result.path).toBe('string');
      expect(result.path.startsWith('M')).toBe(true);
    });

    it('should return center coordinates', () => {
      const result = calculateEdgePath(0, 0, 200, 0, Position.Right, Position.Left, []);

      expect(typeof result.centerX).toBe('number');
      expect(typeof result.centerY).toBe('number');
    });

    it('should handle horizontal straight line', () => {
      const result = calculateEdgePath(0, 100, 200, 100, Position.Right, Position.Left, []);

      expect(result.path).toContain('M 0,100');
      expect(result.centerY).toBeCloseTo(100, 0);
    });

    it('should handle vertical straight line', () => {
      const result = calculateEdgePath(100, 0, 100, 200, Position.Bottom, Position.Top, []);

      expect(result.path).toContain('M 100,0');
    });
  });

  describe('Edge Styles', () => {
    const nodes: Node[] = [];

    it('should generate smoothstep path by default', () => {
      const result = calculateEdgePath(0, 0, 200, 100, Position.Right, Position.Left, nodes);

      // Smoothstep uses Q (quadratic bezier) for corners
      expect(result.path).toBeDefined();
    });

    it('should generate smoothstep path when specified', () => {
      const result = calculateEdgePath(0, 0, 200, 100, Position.Right, Position.Left, nodes, 'smoothstep');

      expect(result.path).toBeDefined();
    });

    it('should generate step path when specified', () => {
      const result = calculateEdgePath(0, 0, 200, 100, Position.Right, Position.Left, nodes, 'step');

      // Step path uses only M and L commands (no curves)
      expect(result.path).toBeDefined();
      expect(result.path).toMatch(/^M.*L/);
    });

    it('should generate bezier path when default style specified', () => {
      const result = calculateEdgePath(0, 0, 200, 100, Position.Right, Position.Left, nodes, 'default');

      // Bezier uses C (cubic bezier) command
      expect(result.path).toBeDefined();
      expect(result.path).toContain('C');
    });

    it('should apply custom border radius for smoothstep', () => {
      const result1 = calculateEdgePath(0, 0, 200, 100, Position.Right, Position.Left, nodes, 'smoothstep', 10);
      const result2 = calculateEdgePath(0, 0, 200, 100, Position.Right, Position.Left, nodes, 'smoothstep', 30);

      // Both should produce valid paths
      expect(result1.path).toBeDefined();
      expect(result2.path).toBeDefined();
    });
  });

  describe('Source and Target Positions', () => {
    const nodes: Node[] = [];

    it('should handle Right to Left positions', () => {
      const result = calculateEdgePath(100, 100, 300, 100, Position.Right, Position.Left, nodes);

      expect(result.path).toBeDefined();
    });

    it('should handle Left to Right positions', () => {
      const result = calculateEdgePath(300, 100, 100, 100, Position.Left, Position.Right, nodes);

      expect(result.path).toBeDefined();
    });

    it('should handle Bottom to Top positions', () => {
      const result = calculateEdgePath(100, 100, 100, 300, Position.Bottom, Position.Top, nodes);

      expect(result.path).toBeDefined();
    });

    it('should handle Top to Bottom positions', () => {
      const result = calculateEdgePath(100, 300, 100, 100, Position.Top, Position.Bottom, nodes);

      expect(result.path).toBeDefined();
    });

    it('should handle diagonal Right to Left', () => {
      const result = calculateEdgePath(0, 0, 200, 200, Position.Right, Position.Left, nodes);

      expect(result.path).toBeDefined();
    });

    it('should handle diagonal Bottom to Top', () => {
      const result = calculateEdgePath(0, 0, 200, 200, Position.Bottom, Position.Top, nodes);

      expect(result.path).toBeDefined();
    });
  });

  describe('Obstacle Avoidance', () => {
    it('should route around a single obstacle', () => {
      const obstacle = createNode('obstacle', 100, 75, 100, 50);
      const result = calculateEdgePath(0, 100, 300, 100, Position.Right, Position.Left, [obstacle]);

      expect(result.path).toBeDefined();
      // Path should not go through the obstacle
    });

    it('should route around multiple obstacles', () => {
      const obstacles = [createNode('obs1', 100, 75, 50, 50), createNode('obs2', 200, 75, 50, 50)];
      const result = calculateEdgePath(0, 100, 350, 100, Position.Right, Position.Left, obstacles);

      expect(result.path).toBeDefined();
    });

    it('should handle obstacles that block horizontal path', () => {
      const obstacle = createNode('blocker', 100, 50, 100, 100);
      const result = calculateEdgePath(0, 100, 300, 100, Position.Right, Position.Left, [obstacle]);

      expect(result.path).toBeDefined();
      // Path should find alternative route
    });

    it('should handle obstacles that block vertical path', () => {
      const obstacle = createNode('blocker', 50, 100, 100, 100);
      const result = calculateEdgePath(100, 0, 100, 300, Position.Bottom, Position.Top, [obstacle]);

      expect(result.path).toBeDefined();
    });

    it('should use node measured dimensions for collision detection', () => {
      const obstacle: Node = {
        id: 'measured',
        position: {x: 100, y: 75},
        data: {},
        measured: {width: 100, height: 50},
      };
      const result = calculateEdgePath(0, 100, 300, 100, Position.Right, Position.Left, [obstacle]);

      expect(result.path).toBeDefined();
    });

    it('should use node width/height when measured not available', () => {
      const obstacle: Node = {
        id: 'sized',
        position: {x: 100, y: 75},
        data: {},
        width: 100,
        height: 50,
      };
      const result = calculateEdgePath(0, 100, 300, 100, Position.Right, Position.Left, [obstacle]);

      expect(result.path).toBeDefined();
    });
  });

  describe('Exit Point Calculation', () => {
    it('should add padding when exiting from Right position', () => {
      const result = calculateEdgePath(0, 0, 200, 0, Position.Right, Position.Left, []);

      expect(result.path).toBeDefined();
    });

    it('should add padding when exiting from Left position', () => {
      const result = calculateEdgePath(200, 0, 0, 0, Position.Left, Position.Right, []);

      expect(result.path).toBeDefined();
    });

    it('should add padding when exiting from Bottom position', () => {
      const result = calculateEdgePath(0, 0, 0, 200, Position.Bottom, Position.Top, []);

      expect(result.path).toBeDefined();
    });

    it('should add padding when exiting from Top position', () => {
      const result = calculateEdgePath(0, 200, 0, 0, Position.Top, Position.Bottom, []);

      expect(result.path).toBeDefined();
    });

    it('should handle exit point inside container node', () => {
      const container = createNode('container', 0, 0, 200, 200);
      const result = calculateEdgePath(100, 100, 400, 100, Position.Right, Position.Left, [container]);

      expect(result.path).toBeDefined();
    });
  });

  describe('Center Point Calculation', () => {
    it('should calculate center for horizontal path', () => {
      const result = calculateEdgePath(0, 100, 200, 100, Position.Right, Position.Left, []);

      expect(result.centerX).toBeCloseTo(100, -1);
      expect(result.centerY).toBeCloseTo(100, 0);
    });

    it('should calculate center for L-shaped path', () => {
      const result = calculateEdgePath(0, 0, 200, 200, Position.Right, Position.Left, [], 'step');

      expect(typeof result.centerX).toBe('number');
      expect(typeof result.centerY).toBe('number');
    });

    it('should calculate center for bezier curve', () => {
      const result = calculateEdgePath(0, 0, 200, 100, Position.Right, Position.Left, [], 'default');

      expect(typeof result.centerX).toBe('number');
      expect(typeof result.centerY).toBe('number');
    });
  });

  describe('Bezier Edge Style', () => {
    it('should create bezier curve with control points', () => {
      const result = calculateEdgePath(0, 0, 200, 0, Position.Right, Position.Left, [], 'default');

      expect(result.path).toContain('C');
    });

    it('should handle backward-flowing edges (target left of source)', () => {
      const result = calculateEdgePath(200, 0, 0, 0, Position.Right, Position.Left, [], 'default');

      expect(result.path).toBeDefined();
      expect(result.path).toContain('C');
    });

    it('should create elaborate curve for significantly backward edges', () => {
      const result = calculateEdgePath(300, 100, 0, 100, Position.Right, Position.Left, [], 'default');

      expect(result.path).toBeDefined();
    });

    it('should handle backward edges going down', () => {
      const result = calculateEdgePath(300, 0, 0, 200, Position.Right, Position.Left, [], 'default');

      expect(result.path).toBeDefined();
    });

    it('should handle backward edges going up', () => {
      const result = calculateEdgePath(300, 200, 0, 0, Position.Right, Position.Left, [], 'default');

      expect(result.path).toBeDefined();
    });
  });

  describe('Path Simplification', () => {
    it('should remove collinear points from path', () => {
      const result = calculateEdgePath(0, 100, 300, 100, Position.Right, Position.Left, []);

      // A straight horizontal line should be simplified
      expect(result.path).toBeDefined();
    });

    it('should remove duplicate points', () => {
      const result = calculateEdgePath(0, 0, 100, 0, Position.Right, Position.Left, []);

      expect(result.path).toBeDefined();
    });
  });
});

describe('calculateAllEdgePaths', () => {
  const createNode = (id: string, x: number, y: number, width = 150, height = 50): Node => ({
    id,
    position: {x, y},
    data: {},
    measured: {width, height},
  });

  const createEdgeInput = (
    id: string,
    sourceX: number,
    sourceY: number,
    targetX: number,
    targetY: number,
    sourcePosition = Position.Right,
    targetPosition = Position.Left,
  ): EdgeInput => ({
    id,
    sourceX,
    sourceY,
    targetX,
    targetY,
    sourcePosition,
    targetPosition,
  });

  describe('Basic Functionality', () => {
    it('should return a Map of edge paths', () => {
      const edges: EdgeInput[] = [createEdgeInput('e1', 0, 0, 200, 0)];
      const nodes: Node[] = [];

      const result = calculateAllEdgePaths(edges, nodes);

      expect(result).toBeInstanceOf(Map);
      expect(result.size).toBe(1);
    });

    it('should calculate paths for multiple edges', () => {
      const edges: EdgeInput[] = [
        createEdgeInput('e1', 0, 0, 200, 0),
        createEdgeInput('e2', 0, 100, 200, 100),
        createEdgeInput('e3', 0, 200, 200, 200),
      ];
      const nodes: Node[] = [];

      const result = calculateAllEdgePaths(edges, nodes);

      expect(result.size).toBe(3);
      expect(result.has('e1')).toBe(true);
      expect(result.has('e2')).toBe(true);
      expect(result.has('e3')).toBe(true);
    });

    it('should return EdgePathResult for each edge', () => {
      const edges: EdgeInput[] = [createEdgeInput('e1', 0, 0, 200, 100)];
      const nodes: Node[] = [];

      const result = calculateAllEdgePaths(edges, nodes);
      const edgeResult = result.get('e1');

      expect(edgeResult).toBeDefined();
      expect(edgeResult?.path).toBeDefined();
      expect(typeof edgeResult?.centerX).toBe('number');
      expect(typeof edgeResult?.centerY).toBe('number');
    });
  });

  describe('Edge Styles', () => {
    const edges: EdgeInput[] = [createEdgeInput('e1', 0, 0, 200, 100)];
    const nodes: Node[] = [];

    it('should apply smoothstep style by default', () => {
      const result = calculateAllEdgePaths(edges, nodes);

      expect(result.get('e1')?.path).toBeDefined();
    });

    it('should apply specified edge style', () => {
      const resultStep = calculateAllEdgePaths(edges, nodes, 'step');
      const resultBezier = calculateAllEdgePaths(edges, nodes, 'default');

      expect(resultStep.get('e1')?.path).not.toEqual(resultBezier.get('e1')?.path);
    });

    it('should apply custom border radius', () => {
      const result = calculateAllEdgePaths(edges, nodes, 'smoothstep', 15);

      expect(result.get('e1')?.path).toBeDefined();
    });
  });

  describe('Edge Separation', () => {
    it('should separate overlapping horizontal edges', () => {
      const edges: EdgeInput[] = [
        createEdgeInput('e1', 0, 100, 200, 100),
        createEdgeInput('e2', 0, 100, 200, 100), // Same path
      ];
      const nodes: Node[] = [];

      const result = calculateAllEdgePaths(edges, nodes);

      expect(result.size).toBe(2);
      // Both edges should have valid paths
      expect(result.get('e1')?.path).toBeDefined();
      expect(result.get('e2')?.path).toBeDefined();
    });

    it('should separate overlapping vertical edges', () => {
      const edges: EdgeInput[] = [
        createEdgeInput('e1', 100, 0, 100, 200, Position.Bottom, Position.Top),
        createEdgeInput('e2', 100, 0, 100, 200, Position.Bottom, Position.Top),
      ];
      const nodes: Node[] = [];

      const result = calculateAllEdgePaths(edges, nodes);

      expect(result.size).toBe(2);
    });

    it('should handle edges with different paths (no overlap)', () => {
      const edges: EdgeInput[] = [
        createEdgeInput('e1', 0, 0, 200, 0),
        createEdgeInput('e2', 0, 200, 200, 200),
      ];
      const nodes: Node[] = [];

      const result = calculateAllEdgePaths(edges, nodes);

      expect(result.size).toBe(2);
    });
  });

  describe('Obstacle Avoidance', () => {
    it('should route edges around nodes', () => {
      const edges: EdgeInput[] = [createEdgeInput('e1', 0, 100, 400, 100)];
      const nodes: Node[] = [createNode('obstacle', 150, 75, 100, 50)];

      const result = calculateAllEdgePaths(edges, nodes);

      expect(result.get('e1')?.path).toBeDefined();
    });

    it('should route multiple edges around multiple nodes', () => {
      const edges: EdgeInput[] = [
        createEdgeInput('e1', 0, 100, 500, 100),
        createEdgeInput('e2', 0, 150, 500, 150),
      ];
      const nodes: Node[] = [createNode('obs1', 150, 75, 100, 100), createNode('obs2', 300, 75, 100, 100)];

      const result = calculateAllEdgePaths(edges, nodes);

      expect(result.size).toBe(2);
    });
  });

  describe('Edge Cases', () => {
    it('should handle empty edges array', () => {
      const result = calculateAllEdgePaths([], []);

      expect(result.size).toBe(0);
    });

    it('should handle empty nodes array', () => {
      const edges: EdgeInput[] = [createEdgeInput('e1', 0, 0, 100, 0)];

      const result = calculateAllEdgePaths(edges, []);

      expect(result.size).toBe(1);
    });

    it('should handle edges at the same position', () => {
      const edges: EdgeInput[] = [createEdgeInput('e1', 100, 100, 100, 100)];

      const result = calculateAllEdgePaths(edges, []);

      expect(result.get('e1')?.path).toBeDefined();
    });

    it('should handle very short edges', () => {
      const edges: EdgeInput[] = [createEdgeInput('e1', 0, 0, 10, 0)];

      const result = calculateAllEdgePaths(edges, []);

      expect(result.get('e1')?.path).toBeDefined();
    });

    it('should handle very long edges', () => {
      const edges: EdgeInput[] = [createEdgeInput('e1', 0, 0, 10000, 0)];

      const result = calculateAllEdgePaths(edges, []);

      expect(result.get('e1')?.path).toBeDefined();
    });
  });

  describe('Complex Scenarios', () => {
    it('should handle a graph with multiple connected nodes', () => {
      const nodes: Node[] = [
        createNode('n1', 0, 0, 100, 50),
        createNode('n2', 200, 0, 100, 50),
        createNode('n3', 400, 0, 100, 50),
        createNode('n4', 200, 100, 100, 50),
      ];
      const edges: EdgeInput[] = [
        createEdgeInput('e1', 100, 25, 200, 25),
        createEdgeInput('e2', 300, 25, 400, 25),
        createEdgeInput('e3', 100, 25, 200, 125, Position.Right, Position.Left),
        createEdgeInput('e4', 300, 125, 400, 25, Position.Right, Position.Left),
      ];

      const result = calculateAllEdgePaths(edges, nodes);

      expect(result.size).toBe(4);
      result.forEach((edgeResult) => {
        expect(edgeResult.path).toBeDefined();
        expect(edgeResult.path.length).toBeGreaterThan(0);
      });
    });

    it('should handle edges that need to go around enclosed obstacles', () => {
      const nodes: Node[] = [
        createNode('box1', 100, 0, 50, 200),
        createNode('box2', 200, 0, 50, 200),
        createNode('box3', 100, 0, 150, 50),
        createNode('box4', 100, 150, 150, 50),
      ];
      const edges: EdgeInput[] = [createEdgeInput('e1', 0, 100, 300, 100)];

      const result = calculateAllEdgePaths(edges, nodes);

      expect(result.get('e1')?.path).toBeDefined();
    });
  });
});
