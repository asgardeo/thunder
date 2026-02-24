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

import {Box, CircularProgress, Typography, useColorScheme, type BoxProps} from '@wso2/oxygen-ui';
import {useState, type JSX} from 'react';
import {BaseSignIn, ThemeProvider} from '@asgardeo/react';
import type {EmbeddedFlowComponent} from '@asgardeo/react';
import type {Theme, ColorSchemeOption} from '@thunder/shared-design';
import toAsgardeoTheme from '../features/design/utils/asgardeoThemeTransformer';
import PreviewToolbar from '../features/design/components/PreviewToolbar';
import {VIEWPORT_WIDTHS, VIEWPORT_HEIGHTS} from '../features/design/components/viewportConstants';

export type Viewport = 'desktop' | 'tablet' | 'mobile';

export interface GatePreviewProps {
  /** The theme to render. Null shows a loading spinner; undefined shows an empty prompt. */
  theme: Theme | null | undefined;
  displayName?: string;
  showToolbar?: boolean;
  viewport?: {
    width: string | number;
    height: string | number;
  };
  colorSchemeOverride?: ColorSchemeOption;
}

const ZOOM_STEPS = [25, 50, 75, 100, 125, 150];

const mockComponents: EmbeddedFlowComponent[] = [
  {id: 'text_001', label: 'Sign In', type: 'TEXT', variant: 'HEADING_1'},
  {
    id: 'block_001',
    type: 'BLOCK',
    components: [
      {id: 'input_001', label: 'Username', placeholder: 'Enter your username', type: 'TEXT_INPUT', required: true},
      {id: 'input_002', label: 'Password', placeholder: 'Enter your password', type: 'PASSWORD_INPUT', required: true},
      {
        id: 'action_001',
        label: 'Sign In',
        type: 'ACTION',
        eventType: 'SUBMIT',
        actionRef: 'basic_auth',
        variant: 'PRIMARY',
      },
    ],
  },
] as EmbeddedFlowComponent[];

export default function GatePreview({
  theme,
  displayName,
  showToolbar = true,
  viewport,
  colorSchemeOverride,
}: GatePreviewProps): JSX.Element {
  const {mode, systemMode} = useColorScheme();
  const [previewColorScheme, setPreviewColorScheme] = useState<'light' | 'dark' | 'system'>('light');
  const [_viewport, setViewport] = useState<Viewport>('desktop');
  const [zoom, setZoom] = useState(75);

  const resolvedSystemMode = mode === 'system' ? systemMode : mode;
  const effectiveScheme: 'light' | 'dark' =
    previewColorScheme !== 'system' ? previewColorScheme : resolvedSystemMode === 'dark' ? 'dark' : 'light';

  const zoomIdx = ZOOM_STEPS.indexOf(zoom);

  if (theme === null) {
    return (
      <Box sx={{height: '100%', display: 'flex', alignItems: 'center', justifyContent: 'center'}}>
        <CircularProgress size={32} />
      </Box>
    );
  }

  return (
    <Box sx={{height: '100%', display: 'flex', flexDirection: 'column'}}>
      {/* Toolbar */}
      {showToolbar && (
        <Box sx={{display: 'flex', justifyContent: 'center', py: 1.5, flexShrink: 0}}>
          <PreviewToolbar
            viewport={_viewport}
            setViewport={setViewport}
            previewColorScheme={previewColorScheme}
            setPreviewColorScheme={setPreviewColorScheme}
            zoom={zoom}
            setZoom={setZoom}
            zoomIdx={zoomIdx}
          />
        </Box>
      )}

      {/* Viewport container */}
      <Box sx={{flex: 1, overflow: 'hidden', display: 'flex', justifyContent: 'center', alignItems: 'center', p: 2}}>
        <Box
          sx={{
            backgroundColor: 'background.paper',
            borderRadius: 1,
            width: viewport?.width ?? VIEWPORT_WIDTHS[_viewport],
            height: viewport?.height ?? VIEWPORT_HEIGHTS[_viewport],
            transition: 'width 0.2s ease, height 0.2s ease',
            display: 'flex',
            flexDirection: 'column',
          }}
        >
          {/* Browser chrome */}
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
                {displayName ? `${displayName} — Preview` : 'Preview'}
              </Typography>
            </Box>
          </Box>

          {/* Canvas */}
          <Box
            sx={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              py: 4,
              px: 2,
              overflow: 'hidden',
              height: '100%',
              backgroundColor: theme?.colorSchemes?.[effectiveScheme]?.palette?.background?.default,
            }}
          >
            <Box
              sx={{
                transform: `scale(${zoom / 100})`,
                transformOrigin: 'center',
                flexShrink: 0,
                transition: 'transform 0.15s ease',
              }}
            >
              <ThemeProvider
                mode={colorSchemeOverride ?? effectiveScheme}
                theme={
                  theme
                    ? toAsgardeoTheme(theme, (colorSchemeOverride as 'light' | 'dark' | undefined) ?? effectiveScheme)
                    : undefined
                }
              >
                <BaseSignIn
                  components={mockComponents}
                  isLoading={false}
                  onSubmit={async () => {}}
                  onError={() => {}}
                  error={null}
                  size="medium"
                  showTitle
                  showSubtitle
                  showLogo
                />
              </ThemeProvider>
            </Box>
          </Box>
        </Box>
      </Box>
    </Box>
  );
}
