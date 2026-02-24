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

import {Box, Button, Drawer, IconButton, Tooltip, Typography, useColorScheme} from '@wso2/oxygen-ui';
import {ArrowLeft, Layers, Save} from '@wso2/oxygen-ui-icons-react';
import {useRef, type JSX} from 'react';
import {useNavigate} from 'react-router';
import LayoutPreviewPanel from '../components/LayoutPreviewPanel';
import LayoutConfigPanel from '../components/LayoutConfigPanel';
import ScreenListItem from '../components/layouts/ScreenListItem';
import AddScreenRow from '../components/layouts/AddScreenRow';
import DesignUIConstants from '../constants/design-ui-constants';
import useLayoutBuilder from '../contexts/LayoutBuilder/useLayoutBuilder';

export default function LayoutBuilderPage(): JSX.Element {
  const {mode, systemMode} = useColorScheme();
  const navigate = useNavigate();

  const {
    layoutId,
    displayName,
    selectedScreen,
    setSelectedScreen,
    screenDraft,
    isDirty,
    addScreen,
    getAllScreens,
    getBaseScreenNames,
    setScreenDraft,
    setIsDirty,
  } = useLayoutBuilder();

  const saveHandlerRef = useRef<() => void>(() => {});

  const allScreens = getAllScreens();
  const screenNames = Object.keys(allScreens);
  const baseScreenNames = getBaseScreenNames();

  const handleAddScreen = (name: string, extendsBase: string): void => {
    addScreen(name, extendsBase);
  };

  return (
    <Box sx={{width: '100%', height: '100vh', display: 'flex', flexDirection: 'column'}}>
      {/* ── Top bar ───────────────────────────────────────────────────────── */}
      <Box
        sx={{
          height: 48,
          flexShrink: 0,
          display: 'flex',
          alignItems: 'center',
          px: 1.5,
          borderBottom: '1px solid',
          borderColor: 'divider',
          bgcolor: 'background.paper',
          gap: 1,
        }}
      >
        <Tooltip title="Back to Design">
          <IconButton size="small" onClick={() => navigate('/design')} sx={{mr: 0.5}}>
            <ArrowLeft size={16} />
          </IconButton>
        </Tooltip>

        <Box sx={{flex: 1, display: 'flex', justifyContent: 'center', pointerEvents: 'none'}}>
          <Typography variant="body2" sx={{fontWeight: 600, fontSize: '0.875rem', color: 'text.primary'}}>
            {displayName ?? '—'}
          </Typography>
        </Box>

        <Button
          size="small"
          variant="contained"
          disableElevation
          disabled={!isDirty}
          onClick={() => saveHandlerRef.current()}
          startIcon={<Save size={14} />}
          sx={{textTransform: 'none', fontSize: '0.8125rem', borderRadius: 1.5}}
        >
          Save
        </Button>
      </Box>

      {/* ── Main area ─────────────────────────────────────────────────────── */}
      <Box
        sx={{
          flex: 1,
          display: 'flex',
          overflow: 'hidden',
          p: 1,
          bgcolor: (systemMode ?? mode) === 'dark' ? '#141414' : '#f6f7f9',
        }}
      >
        {/* ── Left panel: screen list ───────────────────────────────────── */}
        <Drawer
          variant="persistent"
          anchor="left"
          open
          sx={{
            width: DesignUIConstants.LEFT_PANEL_WIDTH,
            flexShrink: 0,
            '& .MuiDrawer-paper': {
              width: DesignUIConstants.LEFT_PANEL_WIDTH,
              position: 'relative',
              border: 'none',
              borderRight: '1px solid',
              borderColor: 'divider',
              overflow: 'hidden',
              display: 'flex',
              flexDirection: 'column',
            },
          }}
        >
          <Box
            sx={{
              px: 1.25,
              pt: 1.5,
              pb: 0.75,
              display: 'flex',
              alignItems: 'center',
              gap: 0.75,
            }}
          >
            <Layers size={14} style={{opacity: 0.5}} />
            <Typography
              variant="caption"
              sx={{
                fontWeight: 600,
                fontSize: '0.68rem',
                textTransform: 'uppercase',
                letterSpacing: '0.06em',
                color: 'text.secondary',
              }}
            >
              Screens
            </Typography>
            <Box sx={{flex: 1}} />
            <Typography variant="caption" sx={{fontSize: '0.65rem', color: 'text.disabled'}}>
              {screenNames.length}
            </Typography>
          </Box>

          <Box sx={{flex: 1, overflowY: 'auto', px: 1.25, pb: 1, display: 'flex', flexDirection: 'column', gap: 0.5}}>
            {screenNames.map((name) => (
              <ScreenListItem
                key={name}
                name={name}
                extendsBase={allScreens[name]?.extends as string | undefined}
                isSelected={selectedScreen === name}
                onClick={() => setSelectedScreen(name)}
              />
            ))}
          </Box>

          {/* Add screen */}
          <Box sx={{px: 1.25, pb: 1.25, pt: 0.5, borderTop: '1px solid', borderColor: 'divider'}}>
            <AddScreenRow baseScreens={baseScreenNames} onAdd={handleAddScreen} />
          </Box>
        </Drawer>

        {/* ── Center: canvas with rulers ────────────────────────────────── */}
        <Box
          component="main"
          sx={{
            flexGrow: 1,
            height: '100%',
            overflow: 'hidden',
            display: 'flex',
            flexDirection: 'column',
          }}
        >
          <LayoutPreviewPanel
            layoutId={layoutId ?? null}
            selectedScreen={selectedScreen}
            screenDraft={screenDraft}
            showRulers
          />
        </Box>

        {/* ── Right panel: screen config ────────────────────────────────── */}
        <Drawer
          variant="persistent"
          anchor="right"
          open
          sx={{
            width: DesignUIConstants.RIGHT_PANEL_WIDTH,
            flexShrink: 0,
            '& .MuiDrawer-paper': {
              width: DesignUIConstants.RIGHT_PANEL_WIDTH,
              position: 'relative',
              border: 'none',
              borderLeft: '1px solid',
              borderColor: 'divider',
              overflow: 'hidden',
              display: 'flex',
              flexDirection: 'column',
            },
          }}
        >
          <Box
            sx={{
              height: 40,
              flexShrink: 0,
              px: 2,
              display: 'flex',
              alignItems: 'center',
              borderBottom: '1px solid',
              borderColor: 'divider',
            }}
          >
            <Typography
              variant="caption"
              sx={{
                fontWeight: 600,
                fontSize: '0.7rem',
                textTransform: 'uppercase',
                letterSpacing: '0.06em',
                color: 'text.secondary',
              }}
            >
              {selectedScreen ? `Screen — ${selectedScreen}` : 'Constraints'}
            </Typography>
          </Box>
          <Box sx={{flex: 1, overflow: 'hidden'}}>
            <LayoutConfigPanel
              layoutId={layoutId ?? null}
              selectedScreen={selectedScreen}
              onScreenChange={setSelectedScreen}
              screenDraft={screenDraft}
              onScreenDraftChange={setScreenDraft}
              onDirtyChange={setIsDirty}
              saveHandlerRef={saveHandlerRef}
            />
          </Box>
        </Drawer>
      </Box>
    </Box>
  );
}
