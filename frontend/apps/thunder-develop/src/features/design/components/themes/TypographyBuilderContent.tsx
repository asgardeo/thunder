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

import {Autocomplete, Box, TextField, Typography, type AutocompleteRenderInputParams} from '@wso2/oxygen-ui';
import {type JSX, type SyntheticEvent} from 'react';
import type {ThemeConfig} from '@thunder/shared-design';
import ConfigCard from '../common/ConfigCard';

/** Common browser-safe / system fonts surfaced as suggestions */
const BROWSER_SAFE_FONTS: string[] = [
  'Arial',
  'Arial Black',
  'Brush Script MT',
  'Comic Sans MS',
  'Courier New',
  'Georgia',
  'Helvetica',
  'Impact',
  'Inter',
  'Lucida Console',
  'Lucida Sans Unicode',
  'Palatino Linotype',
  'system-ui',
  'Tahoma',
  'Times New Roman',
  'Trebuchet MS',
  '-apple-system, BlinkMacSystemFont, sans-serif',
  'Verdana',
];

export interface TypographyBuilderContentProps {
  draft: ThemeConfig;
  onUpdate: (updater: (d: ThemeConfig) => void) => void;
}

/**
 * TypographyBuilderContent - Theme builder section for font family configuration.
 * Provides a freeSolo autocomplete with common browser-safe fonts; users can
 * also type any custom font stack.
 */
export default function TypographyBuilderContent({draft, onUpdate}: TypographyBuilderContentProps): JSX.Element {
  const fontFamily = draft.typography?.fontFamily ?? '';

  const handleChange = (_: SyntheticEvent, value: string | null): void => {
    onUpdate((d) => {
      if (d.typography) Object.assign(d.typography, {fontFamily: value ?? ''});
    });
  };

  const handleInputChange = (_: SyntheticEvent, value: string, reason: string): void => {
    if (reason === 'input')
      onUpdate((d) => {
        if (d.typography) Object.assign(d.typography, {fontFamily: value});
      });
  };

  return (
    <ConfigCard title="Font Family">
      <Autocomplete
        freeSolo
        disablePortal
        options={BROWSER_SAFE_FONTS}
        value={fontFamily || null}
        onChange={handleChange}
        onInputChange={handleInputChange}
        renderOption={(props, option: string) => (
          // eslint-disable-next-line react/jsx-props-no-spreading
          <Box component="li" {...props} key={option}>
            <Typography sx={{fontFamily: option, fontSize: '0.875rem'}}>{option}</Typography>
          </Box>
        )}
        renderInput={(params: AutocompleteRenderInputParams) => (
          <TextField
            {...params}
            size="small"
            placeholder="e.g. Inter, Arial, sans-serif"
            helperText="Choose a preset or type any CSS font stack"
          />
        )}
        sx={{mb: 1.5}}
      />

      {/* Live preview of the selected font */}
      {fontFamily && (
        <Box
          sx={{
            border: '1px solid',
            borderColor: 'divider',
            borderRadius: 1.5,
            p: 1.5,
            bgcolor: 'action.hover',
          }}
        >
          <Typography variant="caption" color="text.disabled" sx={{display: 'block', mb: 0.5, fontSize: '0.65rem'}}>
            Preview
          </Typography>
          <Typography sx={{fontFamily, fontSize: '1rem', lineHeight: 1.4}}>
            The quick brown fox jumps over the lazy dog.
          </Typography>
          <Typography sx={{fontFamily, fontSize: '0.75rem', color: 'text.secondary', mt: 0.5}}>
            ABCDEFGHIJKLMNOPQRSTUVWXYZ · 0123456789
          </Typography>
        </Box>
      )}
    </ConfigCard>
  );
}
