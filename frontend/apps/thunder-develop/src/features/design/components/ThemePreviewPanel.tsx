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

import {Box, CircularProgress, IconButton, Tooltip, Typography, useColorScheme} from '@wso2/oxygen-ui';
import {useState, type JSX} from 'react';
import {BaseSignIn, ThemeProvider} from '@asgardeo/react';
import type {EmbeddedFlowComponent} from '@asgardeo/react';
import {Minus, Monitor, Moon, Plus, Smartphone, Sun, SunMoon, Tablet} from '@wso2/oxygen-ui-icons-react';
import useThemeBuilder from '../contexts/ThemeBuilder/useThemeBuilder';
import toAsgardeoTheme from '../utils/asgardeoThemeTransformer';

export type Viewport = 'desktop' | 'tablet' | 'mobile';

type ColorSchemeOption = 'light' | 'dark' | 'system';

interface ThemePreviewPanelProps {
  themeId: string | null;
}

const VIEWPORT_WIDTHS: Record<Viewport, number> = {
  desktop: 1440,
  tablet: 768,
  mobile: 390,
};

const ZOOM_STEPS = [25, 50, 75, 100, 125, 150];

const VIEWPORT_OPTIONS: {id: Viewport; label: string; icon: JSX.Element}[] = [
  {id: 'mobile', label: 'Mobile (390px)', icon: <Smartphone size={14} />},
  {id: 'tablet', label: 'Tablet (768px)', icon: <Tablet size={14} />},
  {id: 'desktop', label: 'Desktop (1440px)', icon: <Monitor size={14} />},
];

const COLOR_SCHEME_OPTIONS: {id: ColorSchemeOption; label: string; icon: JSX.Element}[] = [
  {id: 'light', label: 'Light', icon: <Sun size={14} />},
  {id: 'dark', label: 'Dark', icon: <Moon size={14} />},
  {id: 'system', label: 'System', icon: <SunMoon size={14} />},
];

// Thin vertical divider for toolbar groups
function ToolbarDivider(): JSX.Element {
  return <Box sx={{width: '1px', height: 16, bgcolor: 'divider', mx: 0.5, flexShrink: 0}} />;
}

// ── Mock flow components ─────────────────────────────────────────────────────

const mockComponents: EmbeddedFlowComponent[] = [
  {
    id: 'text_001',
    label: 'Sign In',
    type: 'TEXT',
    variant: 'HEADING_1',
  },
  {
    id: 'block_001',
    type: 'BLOCK',
    components: [
      {
        id: 'input_001',
        label: 'Username',
        placeholder: 'Enter your username',
        type: 'TEXT_INPUT',
        required: true,
      },
      {
        id: 'input_002',
        label: 'Password',
        placeholder: 'Enter your password',
        type: 'PASSWORD_INPUT',
        required: true,
      },
      {
        id: 'action_001',
        label: 'Sign In',
        type: 'ACTION',
        eventType: 'SUBMIT',
        actionRef: 'basic_auth',
      },
    ],
  },
] as EmbeddedFlowComponent[];

export default function ThemePreviewPanel({themeId}: ThemePreviewPanelProps): JSX.Element {
  const {draftTheme, displayName, previewColorScheme, setPreviewColorScheme, viewport, setViewport} = useThemeBuilder();
  const {mode, systemMode} = useColorScheme();
  const [zoom, setZoom] = useState(100);

  // Resolve the effective scheme — 'system' defers to the OS/app color mode
  const resolvedSystemMode = mode === 'system' ? systemMode : mode;
  const resolvedIfSystem: 'light' | 'dark' = resolvedSystemMode === 'dark' ? 'dark' : 'light';
  const effectiveScheme: 'light' | 'dark' = previewColorScheme !== 'system' ? previewColorScheme : resolvedIfSystem;

  const viewportWidth = VIEWPORT_WIDTHS[viewport];
  const zoomIdx = ZOOM_STEPS.indexOf(zoom);

  if (!themeId && !draftTheme) {
    return (
      <Box
        sx={{height: '100%', display: 'flex', alignItems: 'center', justifyContent: 'center', color: 'text.secondary'}}
      >
        <Typography variant="body2">Select a theme to preview</Typography>
      </Box>
    );
  }

  if (!draftTheme) {
    return (
      <Box sx={{p: 4, height: '100%', display: 'flex', alignItems: 'center', justifyContent: 'center'}}>
        <CircularProgress size={32} />
      </Box>
    );
  }

  // No-op handlers for preview
  const handleSubmit = async (): Promise<void> => {
    // Preview mode - no actual submission
  };

  const handleError = (): void => {
    // Preview mode - no error handling needed
  };

  return (
    <Box sx={{height: '100%', display: 'flex', flexDirection: 'column'}}>
      {/* Browser chrome label bar */}
      <Box
        sx={{
          px: 3,
          py: 1.5,
          borderBottom: '1px solid',
          borderColor: 'divider',
          display: 'flex',
          alignItems: 'center',
          gap: 1,
          flexShrink: 0,
        }}
      >
        <Box sx={{width: 8, height: 8, borderRadius: '50%', bgcolor: '#fc5c57'}} />
        <Box sx={{width: 8, height: 8, borderRadius: '50%', bgcolor: '#febc2e'}} />
        <Box sx={{width: 8, height: 8, borderRadius: '50%', bgcolor: '#29c840'}} />
        <Box
          sx={{
            flex: 1,
            mx: 2,
            height: 22,
            bgcolor: 'action.hover',
            borderRadius: 1,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
          }}
        >
          <Typography variant="caption" color="text.disabled" sx={{fontSize: 10}}>
            {displayName} — Preview
          </Typography>
        </Box>
      </Box>

      {/* Preview canvas */}
      <Box
        sx={{
          flex: 1,
          overflow: 'auto',
          position: 'relative',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          py: 6,
          px: 2,
        }}
      >
        {/* Zoomed sign-in preview */}
        <Box
          sx={{
            transform: `scale(${zoom / 100})`,
            transformOrigin: 'center',
            flexShrink: 0,
            transition: 'transform 0.15s ease',
          }}
        >
          <ThemeProvider mode={effectiveScheme} theme={toAsgardeoTheme(draftTheme, effectiveScheme)}>
            <BaseSignIn
              components={mockComponents}
              isLoading={false}
              onSubmit={handleSubmit}
              onError={handleError}
              error={null}
              size="medium"
              showTitle
              showSubtitle
              showLogo
            />
          </ThemeProvider>
        </Box>

        {/* ── Floating toolbar ────────────────────────────────────────── */}
        <Box
          sx={{
            position: 'absolute',
            bottom: 20,
            left: '50%',
            transform: 'translateX(-50%)',
            zIndex: 10,
            display: 'flex',
            alignItems: 'center',
            gap: 0.25,
            px: 1.5,
            py: 0.75,
            bgcolor: 'background.paper',
            borderRadius: 3,
            boxShadow: '0 4px 20px rgba(0,0,0,0.12), 0 1px 4px rgba(0,0,0,0.06)',
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

          {/* Resolution display */}
          <Typography
            variant="caption"
            sx={{
              px: 1,
              fontSize: '0.7rem',
              color: 'text.secondary',
              minWidth: 54,
              textAlign: 'center',
              fontVariantNumeric: 'tabular-nums',
            }}
          >
            {viewportWidth}px
          </Typography>

          <ToolbarDivider />

          {/* Color scheme */}
          {COLOR_SCHEME_OPTIONS.map((cs) => (
            <Tooltip key={cs.id} title={`${cs.label} mode`}>
              <IconButton
                size="small"
                onClick={() => setPreviewColorScheme(cs.id)}
                sx={{
                  borderRadius: 1,
                  color: previewColorScheme === cs.id ? 'primary.main' : 'text.secondary',
                  bgcolor: previewColorScheme === cs.id ? 'primary.50' : 'transparent',
                  '&:hover': {bgcolor: previewColorScheme === cs.id ? 'primary.100' : 'action.hover'},
                }}
              >
                {cs.icon}
              </IconButton>
            </Tooltip>
          ))}

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
      </Box>
    </Box>
  );
}
