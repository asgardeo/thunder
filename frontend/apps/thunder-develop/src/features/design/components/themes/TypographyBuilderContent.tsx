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

import {Autocomplete, Box, Stack, TextField, Typography, type AutocompleteRenderInputParams} from '@wso2/oxygen-ui';
import {type JSX, type SyntheticEvent} from 'react';
import type {Theme} from '@thunder/shared-design';
import ConfigCard from '../common/ConfigCard';
import SelectRow from '../common/SelectRow';
import SliderRow from '../common/SliderRow';

const FONT_WEIGHT_OPTIONS = [
  {value: '100', label: '100 — Thin'},
  {value: '200', label: '200 — Extra Light'},
  {value: '300', label: '300 — Light'},
  {value: '400', label: '400 — Regular'},
  {value: '500', label: '500 — Medium'},
  {value: '600', label: '600 — Semi Bold'},
  {value: '700', label: '700 — Bold'},
  {value: '800', label: '800 — Extra Bold'},
  {value: '900', label: '900 — Black'},
];

const TYPE_SCALE_VARIANTS: {key: string; label: string}[] = [
  {key: 'h1', label: 'h1'},
  {key: 'h2', label: 'h2'},
  {key: 'h3', label: 'h3'},
  {key: 'h4', label: 'h4'},
  {key: 'h5', label: 'h5'},
  {key: 'h6', label: 'h6'},
  {key: 'subtitle1', label: 'subtitle1'},
  {key: 'subtitle2', label: 'subtitle2'},
  {key: 'body1', label: 'body1'},
  {key: 'body2', label: 'body2'},
  {key: 'button', label: 'button'},
  {key: 'caption', label: 'caption'},
  {key: 'overline', label: 'overline'},
];

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
  draft: Theme;
  onUpdate: (updater: (d: Theme) => void) => void;
}

/**
 * TypographyBuilderContent - Theme builder section for font family configuration.
 * Provides a freeSolo autocomplete with common browser-safe fonts; users can
 * also type any custom font stack.
 */
export default function TypographyBuilderContent({draft, onUpdate}: TypographyBuilderContentProps): JSX.Element {
  const fontFamily = draft.typography?.fontFamily ?? '';
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const typo = draft.typography as any;

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
    <Stack gap={1}>
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

    <ConfigCard title="Font Weights" defaultOpen={false}>
      <SelectRow
        label="Light"
        value={String(typo?.fontWeightLight ?? 300)}
        options={FONT_WEIGHT_OPTIONS}
        onChange={(v) =>
          onUpdate((d) => {
            // eslint-disable-next-line @typescript-eslint/no-explicit-any
            (d.typography as any).fontWeightLight = Number(v);
          })
        }
      />
      <SelectRow
        label="Regular"
        value={String(typo?.fontWeightRegular ?? 400)}
        options={FONT_WEIGHT_OPTIONS}
        onChange={(v) =>
          onUpdate((d) => {
            // eslint-disable-next-line @typescript-eslint/no-explicit-any
            (d.typography as any).fontWeightRegular = Number(v);
          })
        }
      />
      <SelectRow
        label="Medium"
        value={String(typo?.fontWeightMedium ?? 500)}
        options={FONT_WEIGHT_OPTIONS}
        onChange={(v) =>
          onUpdate((d) => {
            // eslint-disable-next-line @typescript-eslint/no-explicit-any
            (d.typography as any).fontWeightMedium = Number(v);
          })
        }
      />
      <SelectRow
        label="Bold"
        value={String(typo?.fontWeightBold ?? 700)}
        options={FONT_WEIGHT_OPTIONS}
        onChange={(v) =>
          onUpdate((d) => {
            // eslint-disable-next-line @typescript-eslint/no-explicit-any
            (d.typography as any).fontWeightBold = Number(v);
          })
        }
      />
    </ConfigCard>

    <ConfigCard title="Base Sizes" defaultOpen={false}>
      <SliderRow
        label="Base Font Size"
        value={typo?.fontSize ?? 14}
        min={10}
        max={24}
        unit="px"
        onChange={(v) =>
          onUpdate((d) => {
            // eslint-disable-next-line @typescript-eslint/no-explicit-any
            (d.typography as any).fontSize = v;
          })
        }
      />
      <SliderRow
        label="HTML Font Size"
        value={typo?.htmlFontSize ?? 16}
        min={10}
        max={24}
        unit="px"
        onChange={(v) =>
          onUpdate((d) => {
            // eslint-disable-next-line @typescript-eslint/no-explicit-any
            (d.typography as any).htmlFontSize = v;
          })
        }
      />
    </ConfigCard>

    <ConfigCard title="Type Scale" defaultOpen={false}>
      {TYPE_SCALE_VARIANTS.map(({key, label}) => (
        <Stack key={key} direction="row" alignItems="center" justifyContent="space-between" sx={{py: 0.4}}>
          <Typography
            variant="caption"
            color="text.secondary"
            sx={{fontSize: '0.75rem', fontFamily: 'monospace', minWidth: 72, flexShrink: 0}}
          >
            {label}
          </Typography>
          <TextField
            size="small"
            value={typo?.[key]?.fontSize ?? ''}
            onChange={(e) =>
              onUpdate((d) => {
                // eslint-disable-next-line @typescript-eslint/no-explicit-any
                const t = d.typography as any;
                if (t?.[key]) t[key].fontSize = e.target.value;
              })
            }
            placeholder="e.g. 1.5rem"
            sx={{width: 110, '& .MuiInputBase-input': {fontSize: '0.75rem', py: 0.5}}}
          />
        </Stack>
      ))}
    </ConfigCard>
    </Stack>
  );
}
