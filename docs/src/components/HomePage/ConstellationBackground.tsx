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

import React, {JSX} from 'react';
import useIsDarkMode from '../../hooks/useIsDarkMode';

/**
 * Thunder (lightning bolt) outline vertices, based on the path:
 *   M13.5,1 L4,18 L11,18 L9.5,31 L20,14 L13,14 L13.5,1 Z
 * Normalized to a 24x32 unit coordinate space.
 */
const BOLT_VERTICES: [number, number][] = [
  [13.5, 1],   // top point
  [4, 18],     // lower-left bend
  [11, 18],    // inner-left step
  [9.5, 31],   // bottom point
  [20, 14],    // upper-right bend
  [13, 14],    // inner-right step
];

const BOLT_CX = 12;
const BOLT_CY = 16;

/** Generate SVG path `d` for a thunder bolt at given position and scale. */
function boltPath(cx: number, cy: number, scale: number): string {
  return BOLT_VERTICES
    .map(([px, py], i) => {
      const x = (px - BOLT_CX) * scale + cx;
      const y = (py - BOLT_CY) * scale + cy;

      return `${i === 0 ? 'M' : 'L'}${x.toFixed(1)},${y.toFixed(1)}`;
    })
    .join(' ') + ' Z';
}

/** Get all vertices of the thunder bolt for drawing nodes. */
function boltVertices(cx: number, cy: number, scale: number): {x: number; y: number}[] {
  return BOLT_VERTICES.map(([px, py]) => ({
    x: (px - BOLT_CX) * scale + cx,
    y: (py - BOLT_CY) * scale + cy,
  }));
}

/**
 * Single large thunder bolt outline as the hero background,
 * rendered with dashed lines and nodes at vertices.
 */
const ConstellationBackground = React.memo(function ConstellationBackground(): JSX.Element {
  const isDark = useIsDarkMode();

  // One large thunder bolt positioned to the right side
  const shapeCx = 1050;
  const shapeCy = 400;
  const shapeScale = 22;

  const vertices = boltVertices(shapeCx, shapeCy, shapeScale);

  return (
    <svg
      style={{
        position: 'absolute',
        top: 0,
        left: 0,
        width: '100%',
        height: '100%',
        pointerEvents: 'none',
      }}
      viewBox="0 0 1440 900"
      preserveAspectRatio="xMidYMid slice"
      xmlns="http://www.w3.org/2000/svg"
    >
      {/* Ambient orange glow */}
      <defs>
        <radialGradient id="hero-orange-glow" cx="60%" cy="35%" r="45%">
          <stop offset="0%" stopColor={isDark ? 'rgba(255, 107, 0, 0.18)' : 'rgba(255, 107, 0, 0.08)'} />
          <stop offset="50%" stopColor={isDark ? 'rgba(255, 107, 0, 0.06)' : 'rgba(255, 107, 0, 0.03)'} />
          <stop offset="100%" stopColor="transparent" />
        </radialGradient>
        <radialGradient id="hero-orange-glow-2" cx="85%" cy="25%" r="30%">
          <stop offset="0%" stopColor={isDark ? 'rgba(255, 107, 0, 0.12)' : 'rgba(255, 107, 0, 0.06)'} />
          <stop offset="100%" stopColor="transparent" />
        </radialGradient>
      </defs>
      <rect width="1440" height="900" fill="url(#hero-orange-glow)" />
      <rect width="1440" height="900" fill="url(#hero-orange-glow-2)" />

      {/* Single large thunder bolt outline — dashed stroke with slow float */}
      <g style={{transformOrigin: `${shapeCx}px ${shapeCy}px`, animation: 'subtleFloat 20s ease-in-out infinite'}}>
        <path
          d={boltPath(shapeCx, shapeCy, shapeScale)}
          fill="none"
          stroke={isDark ? 'rgba(255, 255, 255, 0.12)' : 'rgba(0, 0, 0, 0.08)'}
          strokeWidth="1"
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeDasharray="6 8"
          style={{animation: 'dashFlow 8s linear infinite'}}
        />

        {/* Nodes at all vertices */}
        <g fill={isDark ? 'white' : '#1a1a2e'}>
          {vertices.map((v, i) => (
            <circle key={i} cx={v.x} cy={v.y} r="3" opacity={isDark ? 0.35 : 0.2}>
              <animate attributeName="opacity" values={`${isDark ? 0.2 : 0.1};${isDark ? 0.5 : 0.3};${isDark ? 0.2 : 0.1}`} dur={`${3 + (i % 3)}s`} begin={`${i * 0.3}s`} repeatCount="indefinite" />
            </circle>
          ))}
        </g>
      </g>
    </svg>
  );
});

export default ConstellationBackground;
