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

import {Box, Stack, Typography} from '@wso2/oxygen-ui';
import type {JSX} from 'react';
import ConfigCard from '../common/ConfigCard';
import SelectRow from '../common/SelectRow';
import SliderRow from '../common/SliderRow';
import SlotEditor from './SlotEditor';

export interface ScreenEditorProps {
  screenDraft: Record<string, unknown>;
  onUpdate: (path: string[], value: unknown) => void;
}

/**
 * ScreenEditor - Edits background, spacing, and slot configurations for a layout screen.
 * Uses ConfigCard sections to organize different property groups.
 */
export default function ScreenEditor({screenDraft, onUpdate}: ScreenEditorProps): JSX.Element {
  const background = screenDraft.background as Record<string, unknown> | undefined;
  const spacing = screenDraft.spacing as Record<string, unknown> | undefined;
  const slots = screenDraft.slots as Record<string, unknown> | undefined;

  const num = (v: unknown): number => Number(v) || 0;
  const str = (v: unknown): string => String(v ?? '');

  return (
    <Box sx={{overflowY: 'auto', height: '100%', p: 1.5}}>
      {background && (
        <ConfigCard title="Background">
          {background.type !== undefined && (
            <SelectRow
              label="Type"
              value={str(background.type)}
              options={[
                {value: 'solid', label: 'Solid'},
                {value: 'gradient', label: 'Gradient'},
                {value: 'image', label: 'Image'},
                {value: 'none', label: 'None'},
              ]}
              onChange={(v) => onUpdate(['background', 'type'], v)}
            />
          )}
          {background.value !== undefined && (
            <Stack direction="row" alignItems="flex-start" justifyContent="space-between" spacing={1} sx={{py: 0.5}}>
              <Typography variant="caption" color="text.secondary" sx={{fontSize: '0.75rem', flexShrink: 0}}>
                Value
              </Typography>
              <Typography
                variant="caption"
                sx={{
                  fontFamily: 'monospace',
                  fontSize: '0.7rem',
                  textAlign: 'right',
                  wordBreak: 'break-all',
                  color: 'text.primary',
                }}
              >
                {str(background.value)}
              </Typography>
            </Stack>
          )}
        </ConfigCard>
      )}

      {spacing && (
        <ConfigCard title="Spacing">
          {spacing.componentGap !== undefined && (
            <SliderRow
              label="Component gap"
              value={num(spacing.componentGap)}
              min={0}
              max={64}
              onChange={(v) => onUpdate(['spacing', 'componentGap'], v)}
            />
          )}
          {spacing.sectionGap !== undefined && (
            <SliderRow
              label="Section gap"
              value={num(spacing.sectionGap)}
              min={0}
              max={64}
              onChange={(v) => onUpdate(['spacing', 'sectionGap'], v)}
            />
          )}
        </ConfigCard>
      )}

      {slots && Object.keys(slots).length > 0 && (
        <ConfigCard title="Slots">
          {Object.entries(slots).map(([slotName, slotDef]) => (
            <SlotEditor
              key={slotName}
              name={slotName}
              slot={slotDef as Record<string, unknown>}
              onUpdate={(path, value) => onUpdate(['slots', ...path], value)}
            />
          ))}
        </ConfigCard>
      )}

      {!background && !spacing && !slots && (
        <Box sx={{px: 2, py: 2}}>
          <Typography variant="caption" color="text.disabled" sx={{fontSize: '0.75rem'}}>
            No overrides — inherits from base screen
          </Typography>
        </Box>
      )}
    </Box>
  );
}
