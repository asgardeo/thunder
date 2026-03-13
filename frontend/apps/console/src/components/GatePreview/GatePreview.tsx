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

import {useState, type JSX} from 'react';
import {AsgardeoProvider, BaseSignIn, type EmbeddedFlowComponent} from '@asgardeo/react';
import {Box, CircularProgress, Typography, useColorScheme} from '@wso2/oxygen-ui';
import type {Theme, ColorSchemeOption} from '@thunder/shared-design';
import toAsgardeoTheme from '../../features/design/utils/asgardeoThemeTransformer';
import PreviewToolbar from '../../features/design/components/PreviewToolbar';
import {VIEWPORT_WIDTHS, VIEWPORT_HEIGHTS} from '../../features/design/components/viewportConstants';
import buildPreviewMock from './mocks/buildPreviewMock';

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
  colorScheme?: ColorSchemeOption;
  /** When true, the preview tracks the host app's color scheme instead of the toolbar toggle. */
  syncColorSchemeWithSystem?: boolean;
  mock?: EmbeddedFlowComponent[];
}

const ZOOM_STEPS = [25, 50, 75, 100, 125, 150];

export default function GatePreview({
  theme,
  displayName = '',
  showToolbar = true,
  viewport = undefined,
  mock = buildPreviewMock(),
  colorScheme = undefined,
  syncColorSchemeWithSystem = false,
}: GatePreviewProps): JSX.Element {
  const {mode, systemMode} = useColorScheme();
  const [previewColorScheme, setPreviewColorScheme] = useState<'light' | 'dark' | 'system'>('light');
  const [viewportState, setViewport] = useState<Viewport>('desktop');
  const [zoom, setZoom] = useState(100);

  const resolvedSystemMode: 'light' | 'dark' = (mode === 'system' ? systemMode : mode) === 'dark' ? 'dark' : 'light';
  const activeScheme = colorScheme !== 'system' ? colorScheme : undefined;
  let effectiveScheme: 'light' | 'dark';
  if (activeScheme) {
    effectiveScheme = activeScheme;
  } else if (syncColorSchemeWithSystem) {
    effectiveScheme = resolvedSystemMode;
  } else if (previewColorScheme !== 'system') {
    effectiveScheme = previewColorScheme;
  } else {
    effectiveScheme = resolvedSystemMode;
  }

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
            viewport={viewportState}
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
            width: viewport?.width ?? VIEWPORT_WIDTHS[viewportState],
            height: viewport?.height ?? VIEWPORT_HEIGHTS[viewportState],
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
                transform: `scale(${zoom / 140})`,
                transformOrigin: 'center',
                flexShrink: 0,
                transition: 'transform 0.15s ease',
              }}
            >
              <AsgardeoProvider
                platform="AsgardeoV2"
                instanceId={9999}
                baseUrl={undefined}
                preferences={{
                  theme: {
                    overrides: theme ? toAsgardeoTheme(theme, effectiveScheme) : {},
                  },
                }}
              >
                <BaseSignIn
                  components={mock}
                  isLoading={false}
                  onSubmit={async () => {}}
                  onError={() => {}}
                  error={null}
                  size="medium"
                  showTitle
                  showSubtitle
                  showLogo
                />
              </AsgardeoProvider>
            </Box>
          </Box>
        </Box>
      </Box>
    </Box>
  );
}
