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

import {type JSX} from 'react';
import {useTranslation} from 'react-i18next';
import {
  Box,
  Divider,
  FormHelperText,
  ListItemIcon,
  ListItemText,
  MenuItem,
  Select,
  Stack,
  Typography,
} from '@wso2/oxygen-ui';
import {Palette, Sliders, Type} from '@wso2/oxygen-ui-icons-react';
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

const SECTION_IDS: ThemeSection[] = ['colors', 'shape', 'typography'];
const SECTION_ICONS: Record<ThemeSection, JSX.Element> = {
  colors: <Palette size={18} />,
  shape: <Sliders size={18} />,
  typography: <Type size={18} />,
  // eslint-disable-next-line react/jsx-no-useless-fragment
  general: <></>,
};

// Exported for external use (e.g. ThemeBuilderPage) — use with t() for labels
const SECTIONS: {id: ThemeSection; icon: JSX.Element}[] = SECTION_IDS.map((id) => ({id, icon: SECTION_ICONS[id]}));

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

const SECTION_LABEL_KEYS: Record<ThemeSection, [string, string]> = {
  colors: ['themes.builder.sections.colors.label', 'Colors'],
  shape: ['themes.builder.sections.shape.label', 'Shape'],
  typography: ['themes.builder.sections.typography.label', 'Typography'],
  general: ['themes.builder.sections.general.label', 'General'],
};

const SECTION_DESCRIPTION_KEYS: Record<ThemeSection, [string, string]> = {
  colors: ['themes.builder.sections.colors.description', 'Light & dark color schemes'],
  shape: ['themes.builder.sections.shape.description', 'Border radius & corner styles'],
  typography: ['themes.builder.sections.typography.description', 'Font family & type scale'],
  general: ['themes.builder.sections.general.description', 'General settings'],
};

export default function ThemeBuilderLeftPanel({
  onBack,
  onPanelToggle,
  draftTheme,
  setDraftTheme,
  setIsDirty,
  activeSection,
  setActiveSection,
}: ThemeBuilderLeftPanelProps): JSX.Element {
  const {t} = useTranslation('design');

  return (
    <>
      <BuilderPanelHeader
        onBack={onBack}
        backLabel={t('themes.builder.actions.back_to_design.label', 'Back to Design')}
        onPanelToggle={onPanelToggle}
        hidePanelTooltip={t('themes.builder.tooltips.hide_sections', 'Hide sections')}
      />
      <Stack gap={2}>
        {/* Top-level global settings */}
        {draftTheme && (
          <>
            <Box>
              <Typography variant="h6" gutterBottom>
                {t('themes.forms.settings.fields.default_color_scheme.label', 'Default Color Scheme')}
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
                      {option && t(`common.color_scheme.options.${option.id}.label`, option.label)}
                    </Box>
                  );
                }}
              >
                {ColorSchemeOptions.map((o) => (
                  <MenuItem key={o.id} value={o.id}>
                    <ListItemIcon>{o.icon}</ListItemIcon>
                    <ListItemText>{t(`common.color_scheme.options.${o.id}.label`, o.label)}</ListItemText>
                  </MenuItem>
                ))}
              </Select>
              <FormHelperText>
                {t(
                  'themes.forms.settings.fields.default_color_scheme.helper_text',
                  'Select whether you want a light, dark or system color scheme as the default.',
                )}
              </FormHelperText>
            </Box>
            <Box>
              <Typography variant="h6" gutterBottom>
                {t('themes.forms.settings.fields.default_text_direction.label', 'Default Text Direction')}
              </Typography>
              <Select
                value={draftTheme.direction ?? 'ltr'}
                onChange={(e) => {
                  const next = JSON.parse(JSON.stringify(draftTheme)) as Theme;
                  next.direction = String(e.target.value) as Theme['direction'];
                  setDraftTheme(next);
                  setIsDirty(true);
                }}
                size="small"
                fullWidth
              >
                <MenuItem value="ltr">
                  {t('themes.forms.settings.fields.default_text_direction.options.ltr.label', 'Left-to-Right (LTR)')}
                </MenuItem>
                <MenuItem value="rtl">
                  {t('themes.forms.settings.fields.default_text_direction.options.rtl.label', 'Right-to-Left (RTL)')}
                </MenuItem>
              </Select>
              <FormHelperText>
                {t(
                  'themes.forms.settings.fields.default_text_direction.helper_text',
                  'Select the default text direction for your theme. This will affect the layout and alignment of components.',
                )}
              </FormHelperText>
            </Box>
          </>
        )}
        <Divider />
        <Box>
          <Typography variant="h6" gutterBottom>
            {t('themes.forms.settings.heading', 'Settings')}
          </Typography>
          <Stack gap={1}>
            {SECTION_IDS.map((id) => (
              <SectionCard
                key={id}
                label={t(...SECTION_LABEL_KEYS[id])}
                description={t(...SECTION_DESCRIPTION_KEYS[id])}
                icon={SECTION_ICONS[id]}
                isSelected={activeSection === id}
                onClick={() => setActiveSection(id)}
              />
            ))}
          </Stack>
        </Box>
      </Stack>
    </>
  );
}
