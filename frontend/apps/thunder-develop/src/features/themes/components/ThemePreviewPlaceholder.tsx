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
import type {Theme} from '../models/theme';

interface ThemePreviewPlaceholderProps {
  theme: Theme;
  colorScheme?: 'light' | 'dark';
}

/**
 * Visual preview for theme cards showing actual theme colors
 */
export default function ThemePreviewPlaceholder({
  theme,
  colorScheme = 'dark',
}: ThemePreviewPlaceholderProps): JSX.Element {
  // Provide defaults if theme data is not available
  const defaultColors = {
    primary: {main: '#6366f1', dark: '#4f46e5', contrastText: '#ffffff'},
    secondary: {main: '#8b5cf6', dark: '#7c3aed', contrastText: '#ffffff'},
    background: {
      default: colorScheme === 'dark' ? '#1e293b' : '#ffffff',
      paper: colorScheme === 'dark' ? '#334155' : '#f8fafc',
    },
    text: {
      primary: colorScheme === 'dark' ? '#f1f5f9' : '#0f172a',
      secondary: colorScheme === 'dark' ? '#94a3b8' : '#64748b',
    },
  };

  const colors = theme?.theme?.colorSchemes?.[colorScheme]?.colors || defaultColors;
  const borderRadius = parseInt(theme?.theme?.shape?.borderRadius || '8') || 8;

  return (
    <Box
      sx={{
        width: '100%',
        height: 140,
        borderRadius: 1,
        overflow: 'hidden',
        bgcolor: colors.background.default,
        border: '1px solid',
        borderColor: 'divider',
        position: 'relative',
      }}
    >
      <svg width="100%" height="100%" viewBox="0 0 300 140" xmlns="http://www.w3.org/2000/svg">
        <defs>
          {/* Gradient for primary color */}
          <linearGradient id={`primaryGrad-${colorScheme}`} x1="0%" y1="0%" x2="100%" y2="0%">
            <stop offset="0%" style={{stopColor: colors.primary.main, stopOpacity: 1}} />
            <stop offset="100%" style={{stopColor: colors.primary.dark, stopOpacity: 1}} />
          </linearGradient>
          {/* Gradient for secondary color */}
          <linearGradient id={`secondaryGrad-${colorScheme}`} x1="0%" y1="0%" x2="100%" y2="0%">
            <stop offset="0%" style={{stopColor: colors.secondary.main, stopOpacity: 1}} />
            <stop offset="100%" style={{stopColor: colors.secondary.dark, stopOpacity: 1}} />
          </linearGradient>
        </defs>
        {/* Background */}
        <rect width="300" height="140" fill={colors.background.default} />
        {/* Top bar with color indicators */}{' '}
        <rect x="0" y="0" width="300" height="24" fill={colors.background.paper} opacity="0.6" />
        {/* Browser chrome dots */}
        <circle cx="12" cy="12" r="3" fill={colors.text.secondary} opacity="0.4" />
        <circle cx="24" cy="12" r="3" fill={colors.text.secondary} opacity="0.4" />
        <circle cx="36" cy="12" r="3" fill={colors.text.secondary} opacity="0.4" />
        {/* Color swatches (top right) */}
        <circle cx="272" cy="12" r="5" fill={`url(#primaryGrad-${colorScheme})`} opacity="0.95" />
        <circle cx="286" cy="12" r="5" fill={`url(#secondaryGrad-${colorScheme})`} opacity="0.95" />
        {/* Title text */}
        <rect x="32" y="40" width="80" height="8" rx="4" fill={colors.text.primary} opacity="0.6" />
        {/* Content card */}
        <rect x="32" y="60" width="236" height="60" rx={borderRadius} fill={colors.background.paper} />
        {/* Progress bar / Button with primary gradient */}
        <rect
          x="40"
          y="72"
          width="120"
          height="12"
          rx={borderRadius / 2}
          fill={`url(#primaryGrad-${colorScheme})`}
          opacity="0.9"
        />
        <rect x="165" y="72" width="90" height="12" rx={borderRadius / 2} fill={colors.primary.main} opacity="0.25" />
        {/* Secondary element with gradient */}
        <rect
          x="40"
          y="92"
          width="60"
          height="20"
          rx={borderRadius}
          fill={`url(#secondaryGrad-${colorScheme})`}
          opacity="0.85"
        />
        {/* Sidebar element */}
        <rect x="270" y="60" width="28" height="60" rx={borderRadius} fill={colors.text.secondary} opacity="0.1" />
      </svg>
    </Box>
  );
}
