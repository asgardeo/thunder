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

export interface SliderRowProps {
  label: string;
  value: number;
  min: number;
  max: number;
  unit?: string;
  onChange: (v: number) => void;
}

/**
 * SliderRow - A labeled slider control with value display.
 * Used for numeric configuration options like padding, spacing, etc.
 */
export default function SliderRow({label, value, min, max, unit = 'px', onChange}: SliderRowProps): JSX.Element {
  return (
    <Box sx={{mb: 0.5}}>
      <Stack direction="row" alignItems="center" justifyContent="space-between">
        <Typography variant="caption" color="text.secondary" sx={{fontSize: '0.75rem'}}>
          {label}
        </Typography>
        <Typography variant="caption" sx={{fontFamily: 'monospace', fontSize: '0.7rem', color: 'text.primary'}}>
          {value}
          {unit}
        </Typography>
      </Stack>
      <Slider size="small" min={min} max={max} step={1} value={value} onChange={(_, v) => onChange(v)} sx={{py: 0.5}} />
    </Box>
  );
}
