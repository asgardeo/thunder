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
import type {Layout} from '../models/layout';

interface LayoutPreviewPlaceholderProps {
  layout: Layout;
  screenType?: string;
}

/**
 * Visual preview for layout cards showing actual layout structure
 */
export default function LayoutPreviewPlaceholder({
  layout,
  screenType = 'signin',
}: LayoutPreviewPlaceholderProps): JSX.Element {
  // Provide fallback if layout data is not available
  if (!layout?.layout?.screens) {
    return (
      <Box
        sx={{
          width: '100%',
          height: 140,
          borderRadius: 1,
          overflow: 'hidden',
          bgcolor: 'background.default',
          border: '1px solid',
          borderColor: 'divider',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
        }}
      >
        <Box sx={{color: 'text.secondary', fontSize: '0.875rem'}}>Loading preview...</Box>
      </Box>
    );
  }

  const screen = layout.layout.screens[screenType] || layout.layout.screens.signin;

  // Provide default values for optional properties
  const container = screen?.container || {
    maxWidth: 400,
    padding: 24,
    borderRadius: 8,
    elevation: 2,
    background: '#ffffff',
  };
  const header = screen?.header || {enabled: false, height: 60, padding: 16};
  const footer = screen?.footer || {enabled: false, height: 40, padding: 16};
  const borderRadius = parseInt(container.borderRadius?.toString() || '8') || 8;

  // Parse background
  let backgroundFill = '#667eea';
  if (screen?.background?.type === 'solid') {
    backgroundFill = screen.background.value;
  } else if (screen?.background?.type === 'gradient') {
    // Use first color from gradient if possible
    backgroundFill = screen.background.value.split(',')[0]?.replace('linear-gradient(', '').trim() || '#667eea';
  }

  return (
    <Box
      sx={{
        width: '100%',
        height: 140,
        borderRadius: 1,
        overflow: 'hidden',
        bgcolor: 'background.default',
        border: '1px solid',
        borderColor: 'divider',
      }}
    >
      <svg width="100%" height="100%" viewBox="0 0 300 140" xmlns="http://www.w3.org/2000/svg">
        <defs>
          <linearGradient id={`layoutGradient-${layout.id}`} x1="0%" y1="0%" x2="100%" y2="100%">
            <stop offset="0%" style={{stopColor: backgroundFill, stopOpacity: 1}} />
            <stop offset="100%" style={{stopColor: backgroundFill, stopOpacity: 0.7}} />
          </linearGradient>
          {/* Shadow filter */}
          <filter id={`shadow-${layout.id}`} x="-50%" y="-50%" width="200%" height="200%">
            <feGaussianBlur in="SourceAlpha" stdDeviation="2" />
            <feOffset dx="0" dy="1" result="offsetblur" />
            <feComponentTransfer>
              <feFuncA type="linear" slope="0.3" />
            </feComponentTransfer>
            <feMerge>
              <feMergeNode />
              <feMergeNode in="SourceGraphic" />
            </feMerge>
          </filter>
        </defs>

        {/* Background */}
        <rect
          width="300"
          height="140"
          fill={screen?.background?.type === 'gradient' ? `url(#layoutGradient-${layout.id})` : backgroundFill}
        />

        {/* Decorative overlay */}
        <rect width="300" height="140" fill="white" opacity="0.02" />

        {/* Browser chrome bar */}
        <rect x="0" y="0" width="300" height="24" fill="black" opacity="0.12" />
        <circle cx="12" cy="12" r="3" fill="white" opacity="0.4" />
        <circle cx="24" cy="12" r="3" fill="white" opacity="0.4" />
        <circle cx="36" cy="12" r="3" fill="white" opacity="0.4" />

        {/* Container with shadow */}
        <rect
          x="50"
          y="35"
          width="200"
          height="90"
          rx={borderRadius}
          fill={container.background}
          opacity="0.98"
          filter={`url(#shadow-${layout.id})`}
        />

        {/* Header if enabled */}
        {header.enabled && (
          <>
            <rect x="60" y="45" width="80" height="8" rx="4" fill={backgroundFill} opacity="0.65" />
            {/* Header icons */}
            <circle cx="235" cy="49" r="3" fill={backgroundFill} opacity="0.4" />
            <circle cx="225" cy="49" r="3" fill={backgroundFill} opacity="0.4" />
          </>
        )}

        {/* Content area - form-like structure */}
        <rect x="60" y="65" width="180" height="10" rx={borderRadius / 2} fill={backgroundFill} opacity="0.15" />
        <rect x="60" y="80" width="180" height="10" rx={borderRadius / 2} fill={backgroundFill} opacity="0.15" />

        {/* Primary button */}
        <rect x="60" y="95" width="120" height="12" rx={borderRadius / 2} fill={backgroundFill} opacity="0.85" />
        <rect x="185" y="95" width="55" height="12" rx={borderRadius / 2} fill={backgroundFill} opacity="0.3" />

        {/* Footer if enabled */}
        {footer.enabled && (
          <>
            <rect x="60" y="112" width="40" height="4" rx="2" fill="#A0AEC0" opacity="0.5" />
            <rect x="105" y="112" width="35" height="4" rx="2" fill="#A0AEC0" opacity="0.5" />
          </>
        )}
      </svg>
    </Box>
  );
}
