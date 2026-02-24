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

import {Box, FormLabel, Stack, Tooltip, Typography} from '@wso2/oxygen-ui';
import {type ChangeEvent, type JSX} from 'react';

const PRESET_COLORS = [
  {label: 'Indigo', value: '#4f46e5'},
  {label: 'Blue', value: '#2563eb'},
  {label: 'Cyan', value: '#0891b2'},
  {label: 'Teal', value: '#0d9488'},
  {label: 'Green', value: '#16a34a'},
  {label: 'Orange', value: '#ea580c'},
  {label: 'Red', value: '#dc2626'},
  {label: 'Pink', value: '#db2777'},
  {label: 'Purple', value: '#9333ea'},
  {label: 'Slate', value: '#475569'},
];

export interface ConfigureThemeColorProps {
  themeName: string;
  primaryColor: string;
  onPrimaryColorChange: (color: string) => void;
}

export default function ConfigureThemeColor({
  themeName,
  primaryColor,
  onPrimaryColorChange,
}: ConfigureThemeColorProps): JSX.Element {
  const isCustomColor = !PRESET_COLORS.some((c) => c.value === primaryColor);

  return (
    <Stack direction="column" spacing={4}>
      <Stack direction="column" spacing={1}>
        <Typography variant="h1">Pick a primary color</Typography>
        <Typography variant="body1" color="text.secondary">
          This sets the primary accent color for <strong>{themeName}</strong>. You can fine-tune everything in the theme builder
          after creating.
        </Typography>
      </Stack>

      <Stack direction="column" spacing={2}>
        <FormLabel>Primary color</FormLabel>
        <Box sx={{display: 'flex', flexWrap: 'wrap', gap: 1.5}}>
          {PRESET_COLORS.map((c) => (
            <Tooltip key={c.value} title={c.label}>
              <Box
                onClick={() => onPrimaryColorChange(c.value)}
                sx={{
                  width: 80,
                  height: 80,
                  borderRadius: 1,
                  bgcolor: c.value,
                  cursor: 'pointer',
                  border: '3px solid',
                  borderColor: primaryColor === c.value ? 'text.primary' : 'transparent',
                  transition: 'border-color 0.15s, transform 0.1s',
                  '&:hover': {transform: 'scale(1.12)'},
                }}
              />
            </Tooltip>
          ))}

          {/* Custom color picker */}
          <Tooltip title="Custom color">
            <Box sx={{position: 'relative', width: 40, height: 40}}>
              <Box
                sx={{
                  width: 80,
                  height: 80,
                  borderRadius: 1,
                  bgcolor: isCustomColor ? primaryColor : 'action.hover',
                  border: '3px solid',
                  borderColor: isCustomColor ? 'text.primary' : 'divider',
                  cursor: 'pointer',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  transition: 'border-color 0.15s',
                }}
              >
                {!isCustomColor && (
                  <Typography variant="caption" sx={{fontSize: '1rem', lineHeight: 1, pointerEvents: 'none'}}>
                    ＋
                  </Typography>
                )}
              </Box>
              <Box
                component="input"
                type="color"
                value={primaryColor}
                onChange={(e: ChangeEvent<HTMLInputElement>) => onPrimaryColorChange(e.target.value)}
                sx={{position: 'absolute', inset: 0, opacity: 0, cursor: 'pointer', width: '100%', height: '100%'}}
              />
            </Box>
          </Tooltip>
        </Box>
      </Stack>
    </Stack>
  );
}
