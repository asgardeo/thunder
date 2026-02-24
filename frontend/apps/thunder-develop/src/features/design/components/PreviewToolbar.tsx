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

import {Box, IconButton, Tooltip, Typography} from '@wso2/oxygen-ui';
import {type JSX} from 'react';
import {Minus, Monitor, Plus, Smartphone, Tablet} from '@wso2/oxygen-ui-icons-react';
import type {ColorSchemeOption} from '@thunder/shared-design';
import type {Viewport} from './ThemePreviewPanel';
import {VIEWPORT_WIDTHS, VIEWPORT_HEIGHTS} from './viewportConstants';
import ColorSchemeOptions from '../constants/ColorSchemeOptions';

export interface PreviewToolbarProps {
  viewport: Viewport;
  setViewport: (v: Viewport) => void;
  previewColorScheme: ColorSchemeOption;
  setPreviewColorScheme: (cs: ColorSchemeOption) => void;
  zoom: number;
  setZoom: (z: number) => void;
  zoomIdx: number;
}

const ZOOM_STEPS = [25, 50, 75, 100, 125, 150];

const VIEWPORT_OPTIONS: {id: Viewport; label: string; icon: JSX.Element}[] = [
  {id: 'mobile', label: 'Mobile (390px)', icon: <Smartphone size={14} />},
  {id: 'tablet', label: 'Tablet (768px)', icon: <Tablet size={14} />},
  {id: 'desktop', label: 'Desktop (1440px)', icon: <Monitor size={14} />},
];


function ToolbarDivider(): JSX.Element {
  return <Box sx={{width: '1px', height: 16, bgcolor: 'divider', mx: 0.5, flexShrink: 0}} />;
}

const ASPECT_RATIO_LABELS: Record<Viewport, string> = {
  desktop: '16:9',
  tablet: '3:4',
  mobile: '9:19',
};

export default function PreviewToolbar({
  viewport,
  setViewport,
  previewColorScheme,
  setPreviewColorScheme,
  showDimensions = false,
  zoom,
  setZoom,
  zoomIdx,
}: PreviewToolbarProps): JSX.Element {
  return (
    <Box
      sx={{
        zIndex: 10,
        display: 'flex',
        alignItems: 'center',
        gap: 0.5,
        px: 2,
        py: 1,
        bgcolor: 'background.paper',
        borderRadius: 1,
        boxShadow: '0 8px 32px rgba(0,0,0,0.18), 0 2px 8px rgba(0,0,0,0.08)',
        border: '1px solid',
        borderColor: 'divider',
        whiteSpace: 'nowrap',
      }}
    >
      {/* Viewport */}
      {VIEWPORT_OPTIONS.map((vp) => (
        <Tooltip key={vp.id} title={vp.label}>
          <IconButton
            size="small"
            onClick={() => setViewport(vp.id)}
            sx={{
              borderRadius: 1,
              color: viewport === vp.id ? 'primary.main' : 'text.secondary',
              bgcolor: viewport === vp.id ? 'primary.50' : 'transparent',
              '&:hover': {bgcolor: viewport === vp.id ? 'primary.100' : 'action.hover'},
            }}
          >
            {vp.icon}
          </IconButton>
        </Tooltip>
      ))}

      <ToolbarDivider />

      {/* Dimensions */}
      {showDimensions && (
        <>
          <Box sx={{display: 'flex', alignItems: 'center', gap: 0.5, px: 0.5}}>
            <Typography
              variant="caption"
              sx={{color: 'text.secondary', fontSize: '0.7rem', fontVariantNumeric: 'tabular-nums'}}
            >
              {VIEWPORT_WIDTHS[viewport]}
            </Typography>
            <Typography variant="caption" sx={{color: 'text.disabled', fontSize: '0.7rem', lineHeight: 1}}>
              ×
            </Typography>
            <Typography
              variant="caption"
              sx={{color: 'text.secondary', fontSize: '0.7rem', fontVariantNumeric: 'tabular-nums'}}
            >
              {VIEWPORT_HEIGHTS[viewport]}
            </Typography>
            <Typography variant="caption" sx={{color: 'text.disabled', fontSize: '0.65rem', lineHeight: 1, pl: 0.25}}>
              {ASPECT_RATIO_LABELS[viewport]}
            </Typography>
          </Box>
          <ToolbarDivider />
        </>
      )}

      {/* Color scheme 3-way switch */}
      <Box sx={{display: 'flex', alignItems: 'center', gap: 0.75}}>
        <Typography variant="caption" sx={{fontSize: '0.7rem', color: 'text.secondary', userSelect: 'none', lineHeight: 1}}>
          Color Scheme
        </Typography>
        <Box
          sx={{
            display: 'flex',
            alignItems: 'center',
            border: '1px solid',
            borderColor: 'divider',
            borderRadius: 1,
            overflow: 'hidden',
          }}
        >
          {ColorSchemeOptions.map((cs) => (
            <Tooltip key={cs.id} title={`${cs.label} mode`}>
              <Box
                onClick={() => setPreviewColorScheme(cs.id)}
                sx={{
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  px: 0.75,
                  py: 0.5,
                  cursor: 'pointer',
                  color: previewColorScheme === cs.id ? 'primary.main' : 'text.secondary',
                  bgcolor: previewColorScheme === cs.id ? 'primary.50' : 'transparent',
                  '&:hover': {bgcolor: previewColorScheme === cs.id ? 'primary.100' : 'action.hover'},
                }}
              >
                {cs.icon}
              </Box>
            </Tooltip>
          ))}
        </Box>
      </Box>

      <ToolbarDivider />

      {/* Zoom out */}
      <Tooltip title="Zoom out">
        <span>
          <IconButton
            size="small"
            disabled={zoomIdx <= 0}
            onClick={() => setZoom(ZOOM_STEPS[zoomIdx - 1] ?? zoom)}
            sx={{borderRadius: 1, color: 'text.secondary'}}
          >
            <Minus size={12} />
          </IconButton>
        </span>
      </Tooltip>

      <Typography
        variant="caption"
        sx={{
          fontSize: '0.7rem',
          minWidth: 36,
          textAlign: 'center',
          color: 'text.secondary',
          fontVariantNumeric: 'tabular-nums',
        }}
      >
        {zoom}%
      </Typography>

      {/* Zoom in */}
      <Tooltip title="Zoom in">
        <span>
          <IconButton
            size="small"
            disabled={zoomIdx >= ZOOM_STEPS.length - 1}
            onClick={() => setZoom(ZOOM_STEPS[zoomIdx + 1] ?? zoom)}
            sx={{borderRadius: 1, color: 'text.secondary'}}
          >
            <Plus size={12} />
          </IconButton>
        </span>
      </Tooltip>
    </Box>
  );
}
