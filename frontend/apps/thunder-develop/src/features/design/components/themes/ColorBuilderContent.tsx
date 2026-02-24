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

import {Stack, type CssVarsPalette} from '@wso2/oxygen-ui';
import type {JSX} from 'react';
import type {Theme} from '@thunder/shared-design';
import ColorEditRow from '../common/ColorEditRow';
import ConfigCard from '../common/ConfigCard';

export interface ColorBuilderContentProps {
  colors: CssVarsPalette;
  onUpdate: (updater: (c: Theme['colorSchemes']) => void) => void;
}

/**
 * ColorBuilderContent - Theme builder section for editing color scheme colors.
 * Organizes colors into ConfigCard sections matching the full palette structure.
 * Channel-only fields (e.g. mainChannel) are intentionally omitted.
 */
export default function ColorBuilderContent({colors, onUpdate}: ColorBuilderContentProps): JSX.Element {
  // Alias as any to work around the MUI Theme type not exposing colorSchemes directly.
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const c = colors as any;

  const field = (path: string[], label: string): JSX.Element => {
    const value: string = path.reduce((obj, key) => obj?.[key], c as any) ?? '';
    return (
      <ColorEditRow
        label={label}
        value={value}
        onChange={(v) =>
          onUpdate((palette: any) => {
            const target = path.slice(0, -1).reduce((obj, key) => (obj[key] ??= {}), palette);
            target[path[path.length - 1]] = v;
          })
        }
      />
    );
  };

  return (
    <Stack gap={1}>
      {/* Primary */}
      <ConfigCard title="Primary">
        {field(['primary', 'main'], 'Main')}
        {field(['primary', 'light'], 'Light')}
        {field(['primary', 'dark'], 'Dark')}
        {field(['primary', 'contrastText'], 'Contrast Text')}
      </ConfigCard>

      {/* Secondary */}
      <ConfigCard title="Secondary" defaultOpen={false}>
        {field(['secondary', 'main'], 'Main')}
        {field(['secondary', 'light'], 'Light')}
        {field(['secondary', 'dark'], 'Dark')}
        {field(['secondary', 'contrastText'], 'Contrast Text')}
      </ConfigCard>

      {/* Semantic: Error */}
      {c.error && (
        <ConfigCard title="Error" defaultOpen={false}>
          {field(['error', 'main'], 'Main')}
          {field(['error', 'light'], 'Light')}
          {field(['error', 'dark'], 'Dark')}
          {field(['error', 'contrastText'], 'Contrast Text')}
        </ConfigCard>
      )}

      {/* Semantic: Warning */}
      {c.warning && (
        <ConfigCard title="Warning" defaultOpen={false}>
          {field(['warning', 'main'], 'Main')}
          {field(['warning', 'light'], 'Light')}
          {field(['warning', 'dark'], 'Dark')}
          {field(['warning', 'contrastText'], 'Contrast Text')}
        </ConfigCard>
      )}

      {/* Semantic: Info */}
      {c.info && (
        <ConfigCard title="Info" defaultOpen={false}>
          {field(['info', 'main'], 'Main')}
          {field(['info', 'light'], 'Light')}
          {field(['info', 'dark'], 'Dark')}
          {field(['info', 'contrastText'], 'Contrast Text')}
        </ConfigCard>
      )}

      {/* Semantic: Success */}
      {c.success && (
        <ConfigCard title="Success" defaultOpen={false}>
          {field(['success', 'main'], 'Main')}
          {field(['success', 'light'], 'Light')}
          {field(['success', 'dark'], 'Dark')}
          {field(['success', 'contrastText'], 'Contrast Text')}
        </ConfigCard>
      )}

      {/* Backgrounds */}
      {c.background && (
        <ConfigCard title="Backgrounds" defaultOpen={false}>
          {field(['background', 'default'], 'Default')}
          {field(['background', 'paper'], 'Surface')}
          {c.background.acrylic !== undefined && field(['background', 'acrylic'], 'Acrylic')}
        </ConfigCard>
      )}

      {/* Text */}
      {c.text && (
        <ConfigCard title="Text" defaultOpen={false}>
          {field(['text', 'primary'], 'Primary')}
          {field(['text', 'secondary'], 'Secondary')}
          {c.text.disabled !== undefined && field(['text', 'disabled'], 'Disabled')}
        </ConfigCard>
      )}

      {/* Common */}
      {c.common && (
        <ConfigCard title="Common" defaultOpen={false}>
          {field(['common', 'black'], 'Black')}
          {field(['common', 'white'], 'White')}
          {c.common.background !== undefined && field(['common', 'background'], 'Background')}
          {c.common.onBackground !== undefined && field(['common', 'onBackground'], 'On Background')}
        </ConfigCard>
      )}

      {/* Borders & Dividers */}
      {c.divider !== undefined && (
        <ConfigCard title="Borders & Dividers" defaultOpen={false}>
          {field(['divider'], 'Divider')}
        </ConfigCard>
      )}
    </Stack>
  );
}
