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

import {Box, Slider, Stack, Typography} from '@wso2/oxygen-ui';
import type {JSX} from 'react';
import type {ThemeConfig} from '@thunder/shared-design';
import ConfigCard from '../common/ConfigCard';

export interface ShapeBuilderContentProps {
  draft: ThemeConfig;
  onUpdate: (updater: (d: ThemeConfig) => void) => void;
}

/**
 * ShapeBuilderContent - Theme builder section for editing border radius.
 * Provides slider control and preset swatches for quick selection.
 */
export default function ShapeBuilderContent({draft, onUpdate}: ShapeBuilderContentProps): JSX.Element {
  const borderRadiusNum =
    typeof draft.shape?.borderRadius === 'number'
      ? draft.shape.borderRadius
      : parseInt(String(draft.shape?.borderRadius ?? '8'), 10) || 8;

  const PRESETS = [0, 4, 8, 12, 16, 24];

  return (
    <ConfigCard title="Border Radius">
      <Stack direction="row" alignItems="center" justifyContent="space-between" mb={0.5}>
        <Typography variant="body2" sx={{fontWeight: 600, fontSize: '0.875rem', color: 'text.secondary'}}>
          Radius
        </Typography>
        <Box
          sx={{
            px: 1.25,
            py: 0.25,
            border: '1.5px solid',
            borderColor: 'divider',
            borderRadius: 1.5,
            minWidth: 52,
            textAlign: 'center',
          }}
        >
          <Typography sx={{fontFamily: 'monospace', fontSize: '0.8125rem', fontWeight: 600}}>
            {borderRadiusNum}px
          </Typography>
        </Box>
      </Stack>

      <Slider
        size="small"
        min={0}
        max={24}
        step={1}
        value={borderRadiusNum}
        onChange={(_, v) =>
          onUpdate((d) => {
            if (d.shape) d.shape.borderRadius = v;
          })
        }
        sx={{py: 1}}
      />

      {/* Preset swatches */}
      <Stack direction="row" spacing={1} sx={{mt: 0.5, flexWrap: 'wrap', gap: 0.75}}>
        {PRESETS.map((r) => (
          <Box
            key={r}
            onClick={() =>
              onUpdate((d) => {
                if (d.shape) d.shape.borderRadius = r;
              })
            }
            sx={{
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              gap: 0.5,
              cursor: 'pointer',
            }}
          >
            <Box
              sx={{
                width: 32,
                height: 32,
                bgcolor: borderRadiusNum === r ? 'primary.main' : 'action.selected',
                borderRadius: `${r}px`,
                border: '1.5px solid',
                borderColor: borderRadiusNum === r ? 'primary.main' : 'transparent',
                transition: 'all 0.15s',
                '&:hover': {bgcolor: borderRadiusNum === r ? 'primary.dark' : 'action.hover'},
              }}
            />
            <Typography
              variant="caption"
              sx={{
                fontSize: '0.6rem',
                color: borderRadiusNum === r ? 'primary.main' : 'text.disabled',
                fontWeight: borderRadiusNum === r ? 700 : 400,
              }}
            >
              {r}
            </Typography>
          </Box>
        ))}
      </Stack>
    </ConfigCard>
  );
}
