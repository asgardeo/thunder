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

import type {JSX} from 'react';
import {Box} from '@wso2/oxygen-ui';

/**
 * Props for Ruler component
 */
interface RulerProps {
  /** Ruler direction */
  direction: 'horizontal' | 'vertical';
  /** Size of the ruler in pixels */
  size: number;
  /** Zoom level (percentage) */
  zoom?: number;
}

/**
 * Ruler component for design canvas
 */
export default function Ruler({direction, size, zoom = 100}: RulerProps): JSX.Element {
  const isHorizontal = direction === 'horizontal';
  const rulerSize = 24; // Height/width of the ruler bar
  const step = 50; // Pixels between markers
  const adjustedStep = (step * zoom) / 100;

  // Calculate number of markers
  const count = Math.ceil(size / adjustedStep);
  const markers = Array.from({length: count}, (_, i) => i * adjustedStep);

  return (
    <Box
      sx={{
        position: 'relative',
        width: isHorizontal ? '100%' : rulerSize,
        height: isHorizontal ? rulerSize : '100%',
        bgcolor: 'background.paper',
        borderRight: isHorizontal ? 'none' : '1px solid',
        borderBottom: isHorizontal ? '1px solid' : 'none',
        borderColor: 'divider',
        overflow: 'hidden',
      }}
    >
      {/* Ruler markers */}
      {markers.map((position, index) => (
        <Box
          key={index}
          sx={{
            position: 'absolute',
            ...(isHorizontal
              ? {
                  left: position,
                  top: 0,
                  width: '1px',
                  height: index % 5 === 0 ? '12px' : '6px',
                }
              : {
                  top: position,
                  left: 0,
                  height: '1px',
                  width: index % 5 === 0 ? '12px' : '6px',
                }),
            bgcolor: 'text.secondary',
            opacity: 0.3,
          }}
        />
      ))}

      {/* Ruler labels (every 200px) */}
      {markers
        .filter((_, index) => index % 4 === 0)
        .map((position, index) => (
          <Box
            key={`label-${index}`}
            sx={{
              position: 'absolute',
              ...(isHorizontal
                ? {
                    left: position + 4,
                    top: 4,
                  }
                : {
                    top: position + 4,
                    left: 4,
                    transform: 'rotate(-90deg)',
                    transformOrigin: 'top left',
                  }),
              fontSize: '9px',
              color: 'text.secondary',
              userSelect: 'none',
            }}
          >
            {Math.round((index * 4 * step * 100) / zoom)}
          </Box>
        ))}
    </Box>
  );
}
