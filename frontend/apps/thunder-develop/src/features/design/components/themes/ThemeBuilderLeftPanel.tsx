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

import {
  Box,
  Divider,
  FormHelperText,
  FormLabel,
  ListItemIcon,
  ListItemText,
  MenuItem,
  Select,
  Stack,
  Typography,
} from '@wso2/oxygen-ui';
import {Palette, Sliders, Type} from '@wso2/oxygen-ui-icons-react';
import {type JSX} from 'react';
import type {Theme} from '@thunder/shared-design';
import ColorSchemeOptions from '../../constants/ColorSchemeOptions';
import BuilderPanelHeader from '../../../../components/BuilderLayout/BuilderPanelHeader';
import SectionCard from './SectionCard';
import {type ThemeSection} from '../../models/theme-builder';

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
];

export {SECTIONS};
export type {SectionDef};

interface ThemeBuilderLeftPanelProps {
  onBack: () => void;
  onPanelToggle: () => void;
  draftTheme: Theme | null | undefined;
  setDraftTheme: (theme: Theme) => void;
  setIsDirty: (dirty: boolean) => void;
  activeSection: ThemeSection;
  setActiveSection: (section: ThemeSection) => void;
}

export default function ThemeBuilderLeftPanel({
  onBack,
  onPanelToggle,
  draftTheme,
  setDraftTheme,
  setIsDirty,
  activeSection,
  setActiveSection,
}: ThemeBuilderLeftPanelProps): JSX.Element {
  return (
    <>
      <BuilderPanelHeader
        onBack={onBack}
        backLabel="Back to Design"
        onPanelToggle={onPanelToggle}
        hidePanelTooltip="Hide sections"
      />
      <Stack gap={2}>
        {/* Top-level global settings */}
        {draftTheme && (
          <>
            <Box>
              <Typography variant="h6" gutterBottom>
                Default Color Scheme
              </Typography>
              <Select
                value={draftTheme.defaultColorScheme ?? 'light'}
                onChange={(e) => {
                  const next = JSON.parse(JSON.stringify(draftTheme)) as Theme;
                  next.defaultColorScheme = String(e.target.value) as Theme['defaultColorScheme'];
                  setDraftTheme(next);
                  setIsDirty(true);
                }}
                fullWidth
                renderValue={(value) => {
                  const option = ColorSchemeOptions.find((o) => o.id === value);
                  return (
                    <Box sx={{display: 'flex', alignItems: 'center', gap: 1}}>
                      {option?.icon}
                      {option?.label}
                    </Box>
                  );
                }}
              >
                {ColorSchemeOptions.map((o) => (
                  <MenuItem key={o.id} value={o.id}>
                    <ListItemIcon>{o.icon}</ListItemIcon>
                    <ListItemText>{o.label}</ListItemText>
                  </MenuItem>
                ))}
              </Select>
              <FormHelperText>
                Select whether you want a light, dark or system color scheme as the default.
              </FormHelperText>
            </Box>
            <Box>
              <Typography variant="h6" gutterBottom>
                Default Text Direction
              </Typography>
              <Select
                value={draftTheme.direction ?? 'ltr'}
                onChange={(e) => {
                  const next = JSON.parse(JSON.stringify(draftTheme)) as Theme;
                  next.direction = String(e.target.value);
                  setDraftTheme(next);
                  setIsDirty(true);
                }}
                size="small"
                fullWidth
              >
                <MenuItem value="ltr">Left-to-Right (LTR)</MenuItem>
                <MenuItem value="rtl">Right-to-Left (RTL)</MenuItem>
              </Select>
              <FormHelperText>
                Select the default text direction for your theme. This will affect the layout and alignment of
                components.
              </FormHelperText>
            </Box>
          </>
        )}
        <Divider />
        <Box>
          <Typography variant="h6" gutterBottom>
            Settings
          </Typography>
          <Stack gap={1}>
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
        </Box>
      </Stack>
    </>
  );
}
