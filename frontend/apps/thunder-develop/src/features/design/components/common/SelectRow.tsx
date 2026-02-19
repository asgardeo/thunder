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

import {MenuItem, Select, Stack, Typography} from '@wso2/oxygen-ui';
import type {JSX} from 'react';

export interface SelectRowProps {
  label: string;
  value: string;
  options: {value: string; label: string}[];
  onChange: (v: string) => void;
}

/**
 * SelectRow - A labeled dropdown select control.
 * Used for enumerated configuration options.
 */
export default function SelectRow({label, value, options, onChange}: SelectRowProps): JSX.Element {
  return (
    <Stack direction="row" alignItems="center" justifyContent="space-between" sx={{py: 0.5}}>
      <Typography variant="caption" color="text.secondary" sx={{fontSize: '0.75rem'}}>
        {label}
      </Typography>
      <Select
        value={value}
        onChange={(e) => onChange(String(e.target.value))}
        size="small"
        sx={{fontSize: '0.75rem', height: 28, minWidth: 90}}
      >
        {options.map((opt) => (
          <MenuItem key={opt.value} value={opt.value} sx={{fontSize: '0.75rem'}}>
            {opt.label}
          </MenuItem>
        ))}
      </Select>
    </Stack>
  );
}
