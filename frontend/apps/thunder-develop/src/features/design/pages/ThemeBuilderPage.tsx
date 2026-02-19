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

import {Box, Button, Drawer, Stack, Typography, useColorScheme} from '@wso2/oxygen-ui';
import {Palette, Save, Settings, Sliders, Type} from '@wso2/oxygen-ui-icons-react';
import {useCallback, useRef, useState, type JSX} from 'react';
import {useNavigate} from 'react-router';
import BuilderLayout from '../../../components/BuilderLayout/BuilderLayout';
import BuilderPanelHeader from '../../../components/BuilderLayout/BuilderPanelHeader';
import ThemePreviewPanel from '../components/ThemePreviewPanel';
import ThemeConfigPanel from '../components/ThemeConfigPanel';
import SectionCard from '../components/themes/SectionCard';
import {type ThemeSection} from '../models/theme-builder';
import DesignUIConstants from '../constants/design-ui-constants';
import useThemeBuilder from '../contexts/ThemeBuilder/useThemeBuilder';

interface SectionDef {
  id: ThemeSection;
  label: string;
  description: string;
  icon: JSX.Element;
}

const SECTIONS: SectionDef[] = [
  {id: 'colors', label: 'Colors', description: 'Light & dark color schemes', icon: <Palette size={18} />},
  {id: 'shape', label: 'Shape', description: 'Border radius & corner styles', icon: <Sliders size={18} />},
  {id: 'typography', label: 'Typography', description: 'Font family & type scale', icon: <Type size={18} />},
  {id: 'general', label: 'General', description: 'Direction & default scheme', icon: <Settings size={18} />},
];

export default function ThemeBuilderPage(): JSX.Element {
  const {mode, systemMode} = useColorScheme();
  const navigate = useNavigate();

  const {themeId, displayName, activeSection, setActiveSection, isDirty} = useThemeBuilder();

  const saveHandlerRef = useRef<() => void>(() => {});
  const [isPanelOpen, setIsPanelOpen] = useState(true);

  const handleTogglePanel = useCallback(() => {
    setIsPanelOpen((prev) => !prev);
  }, []);

  const handleBack = useCallback(() => {
    // eslint-disable-next-line @typescript-eslint/no-floating-promises
    navigate('/design');
  }, [navigate]);

  const bgColor = (systemMode ?? mode) === 'dark' ? '#141414' : '#f6f7f9';

  const leftPanelContent = (
    <>
      <BuilderPanelHeader
        onBack={handleBack}
        backLabel="Back to Design"
        onPanelToggle={handleTogglePanel}
        hidePanelTooltip="Hide sections"
      />
      <Box sx={{px: 0.25, pb: 0.75}}>
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
          Theme sections
        </Typography>
      </Box>
      <Stack>
        {SECTIONS.map((s) => (
          <SectionCard
            key={s.id}
            label={s.label}
            description={s.description}
            icon={s.icon}
            isSelected={activeSection === s.id}
            onClick={() => setActiveSection(s.id)}
          />
        ))}
      </Stack>
    </>
  );

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
        }}
      >
        {/* Theme name — centred */}
        <Box sx={{flex: 1, display: 'flex', justifyContent: 'center', pointerEvents: 'none'}}>
          <Typography variant="body2" sx={{fontWeight: 600, fontSize: '0.875rem', color: 'text.primary'}}>
            {displayName ?? '—'}
          </Typography>
        </Box>

        {/* Save */}
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

      {/* ── Three-column builder area ──────────────────────────────────────── */}
      <Box sx={{flex: 1, overflow: 'hidden'}}>
        <BuilderLayout
          open={isPanelOpen}
          onPanelToggle={handleTogglePanel}
          panelWidth={DesignUIConstants.LEFT_PANEL_WIDTH}
          panelContent={leftPanelContent}
          expandTooltip="Show sections"
          panelPaperSx={{
            overflow: 'hidden',
            display: 'flex',
            flexDirection: 'column',
            borderRight: '1px solid',
            borderColor: 'divider',
          }}
        >
          <Box sx={{display: 'flex', height: '100%', overflow: 'hidden', p: 1, bgcolor: bgColor}}>
            {/* ── Center: canvas preview ─────────────────────────────────── */}
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
              <ThemePreviewPanel themeId={themeId ?? null} />
            </Box>

            {/* ── Right panel: section config ────────────────────────────── */}
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
              {/* Right panel header */}
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
                  {SECTIONS.find((s) => s.id === activeSection)?.label ?? 'Config'}
                </Typography>
              </Box>
              <Box sx={{flex: 1, overflow: 'hidden'}}>
                <ThemeConfigPanel
                  themeId={themeId ?? null}
                  activeSection={activeSection}
                  saveHandlerRef={saveHandlerRef}
                />
              </Box>
            </Drawer>
          </Box>
        </BuilderLayout>
      </Box>
    </Box>
  );
}
