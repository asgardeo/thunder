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

import {BROWSER_SAFE_FONTS, type Theme} from '@thunderid/design';
import {
  Autocomplete,
  Box,
  Link,
  Stack,
  TextField,
  Typography,
  type AutocompleteRenderInputParams,
} from '@wso2/oxygen-ui';
import {type JSX, type SyntheticEvent} from 'react';
import {useTranslation} from 'react-i18next';
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

/** MUI's mapping: which fontWeight reference each typography variant uses by default. */
const VARIANT_WEIGHT_REF: Record<
  string,
  'fontWeightLight' | 'fontWeightRegular' | 'fontWeightMedium' | 'fontWeightBold'
> = {
  h1: 'fontWeightLight',
  h2: 'fontWeightLight',
  h3: 'fontWeightRegular',
  h4: 'fontWeightRegular',
  h5: 'fontWeightRegular',
  h6: 'fontWeightMedium',
  subtitle1: 'fontWeightRegular',
  subtitle2: 'fontWeightMedium',
  body1: 'fontWeightRegular',
  body2: 'fontWeightRegular',
  button: 'fontWeightMedium',
  caption: 'fontWeightRegular',
  overline: 'fontWeightRegular',
};

type TypographyRecord = Record<string, unknown>;

/** Propagate a fontWeight reference change to all variants that use it. */
function propagateWeight(typography: TypographyRecord, weightKey: string, value: number): void {
  for (const [variant, ref] of Object.entries(VARIANT_WEIGHT_REF)) {
    if (ref === weightKey && typography[variant]) {
      (typography[variant] as Record<string, unknown>).fontWeight = value;
    }
  }
}

/** Default font weight values. */
const DEFAULT_WEIGHTS = {fontWeightLight: 300, fontWeightRegular: 400, fontWeightMedium: 500, fontWeightBold: 700};

/** Default base size values. */
const DEFAULT_SIZES = {fontSize: 14, htmlFontSize: 16};

/** All typography variant keys. */
const VARIANT_KEYS = [
  'h1',
  'h2',
  'h3',
  'h4',
  'h5',
  'h6',
  'subtitle1',
  'subtitle2',
  'body1',
  'body2',
  'button',
  'caption',
  'overline',
];

/** Clear computed variant properties so extendTheme recomputes them from the base values.
 *  Removes fontSize, lineHeight, and letterSpacing from all variants. */
function clearVariantSizes(typography: TypographyRecord): void {
  for (const key of VARIANT_KEYS) {
    const variantObj = typography[key] as Record<string, unknown> | undefined;
    if (!variantObj) continue;
    delete variantObj.fontSize;
    delete variantObj.lineHeight;
    delete variantObj.letterSpacing;
  }
}

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
  const {t} = useTranslation('design');
  const fontFamily = (draft.typography?.fontFamily as string) ?? '';
  const typo = draft.typography;

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
      <ConfigCard title={t('themes.forms.typography_builder.font_family.title', 'Font Family')}>
        <Autocomplete
          freeSolo
          options={BROWSER_SAFE_FONTS}
          value={fontFamily || null}
          onChange={handleChange}
          onInputChange={handleInputChange}
          renderOption={(props, option: string) => (
            <Box component="li" {...props} key={option}>
              <Typography sx={{fontFamily: option, fontSize: '0.875rem'}}>{option}</Typography>
            </Box>
          )}
          renderInput={(params: AutocompleteRenderInputParams) => (
            <TextField
              {...params}
              size="small"
              placeholder={t(
                'themes.forms.typography_builder.fields.font_family.placeholder',
                'e.g. Inter, Arial, sans-serif',
              )}
              helperText={t(
                'themes.forms.typography_builder.fields.font_family.helper_text',
                'Choose a preset or type any CSS font stack',
              )}
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
              {t('themes.forms.typography_builder.fields.preview.label', 'Preview')}
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

      <ConfigCard
        title={t('themes.forms.typography_builder.font_weights.title', 'Font Weights')}
        defaultOpen={false}
        action={
          <Link
            component="button"
            variant="caption"
            onClick={() => {
              onUpdate((d) => {
                if (!d.typography) return;
                Object.assign(d.typography, DEFAULT_WEIGHTS);
                for (const [wKey, wVal] of Object.entries(DEFAULT_WEIGHTS)) {
                  propagateWeight(d.typography as unknown as TypographyRecord, wKey, wVal);
                }
              });
            }}
            sx={{fontSize: '0.7rem', mr: 1, cursor: 'pointer', background: 'none', border: 'none', p: 0}}
          >
            {t('themes.forms.typography_builder.actions.reset.label', 'Reset')}
          </Link>
        }
      >
        <SelectRow
          label={t('themes.forms.typography_builder.fields.light.label', 'Light')}
          value={String((typo?.fontWeightLight as number | undefined) ?? 300)}
          options={FONT_WEIGHT_OPTIONS}
          onChange={(v) =>
            onUpdate((d) => {
              if (!d.typography) return;
              const num = Number(v);
              Object.assign(d.typography, {fontWeightLight: num});
              propagateWeight(d.typography as unknown as TypographyRecord, 'fontWeightLight', num);
            })
          }
        />
        <SelectRow
          label={t('themes.forms.typography_builder.fields.regular.label', 'Regular')}
          value={String((typo?.fontWeightRegular as number | undefined) ?? 400)}
          options={FONT_WEIGHT_OPTIONS}
          onChange={(v) =>
            onUpdate((d) => {
              if (!d.typography) return;
              const num = Number(v);
              Object.assign(d.typography, {fontWeightRegular: num});
              propagateWeight(d.typography as unknown as TypographyRecord, 'fontWeightRegular', num);
            })
          }
        />
        <SelectRow
          label={t('themes.forms.typography_builder.fields.medium.label', 'Medium')}
          value={String((typo?.fontWeightMedium as number | undefined) ?? 500)}
          options={FONT_WEIGHT_OPTIONS}
          onChange={(v) =>
            onUpdate((d) => {
              if (!d.typography) return;
              const num = Number(v);
              Object.assign(d.typography, {fontWeightMedium: num});
              propagateWeight(d.typography as unknown as TypographyRecord, 'fontWeightMedium', num);
            })
          }
        />
        <SelectRow
          label={t('themes.forms.typography_builder.fields.bold.label', 'Bold')}
          value={String((typo?.fontWeightBold as number | undefined) ?? 700)}
          options={FONT_WEIGHT_OPTIONS}
          onChange={(v) =>
            onUpdate((d) => {
              if (!d.typography) return;
              const num = Number(v);
              Object.assign(d.typography, {fontWeightBold: num});
              propagateWeight(d.typography as unknown as TypographyRecord, 'fontWeightBold', num);
            })
          }
        />
      </ConfigCard>

      <ConfigCard
        title={t('themes.forms.typography_builder.base_sizes.title', 'Base Sizes')}
        defaultOpen={false}
        action={
          <Link
            component="button"
            variant="caption"
            onClick={() => {
              onUpdate((d) => {
                if (!d.typography) return;
                Object.assign(d.typography, DEFAULT_SIZES);
                clearVariantSizes(d.typography as unknown as TypographyRecord);
              });
            }}
            sx={{fontSize: '0.7rem', mr: 1, cursor: 'pointer', background: 'none', border: 'none', p: 0}}
          >
            {t('themes.forms.typography_builder.actions.reset.label', 'Reset')}
          </Link>
        }
      >
        <SliderRow
          label={t('themes.forms.typography_builder.fields.base_font_size.label', 'Base Font Size')}
          value={(typo?.fontSize as number | undefined) ?? 14}
          min={10}
          max={24}
          unit="px"
          onChange={(v) =>
            onUpdate((d) => {
              if (!d.typography) return;
              Object.assign(d.typography, {fontSize: v});
              clearVariantSizes(d.typography as unknown as TypographyRecord);
            })
          }
        />
        <SliderRow
          label={t('themes.forms.typography_builder.fields.html_font_size.label', 'HTML Font Size')}
          value={(typo?.htmlFontSize as number | undefined) ?? 16}
          min={10}
          max={24}
          unit="px"
          onChange={(v) =>
            onUpdate((d) => {
              if (!d.typography) return;
              Object.assign(d.typography, {htmlFontSize: v});
              clearVariantSizes(d.typography as unknown as TypographyRecord);
            })
          }
        />
      </ConfigCard>

      <ConfigCard
        title={t('themes.forms.typography_builder.type_scale.title', 'Type Scale')}
        defaultOpen={false}
        action={
          <Link
            component="button"
            variant="caption"
            onClick={() => {
              onUpdate((d) => {
                if (!d.typography) return;
                clearVariantSizes(d.typography as unknown as TypographyRecord);
              });
            }}
            sx={{fontSize: '0.7rem', mr: 1, cursor: 'pointer', background: 'none', border: 'none', p: 0}}
          >
            {t('themes.forms.typography_builder.actions.reset.label', 'Reset')}
          </Link>
        }
      >
        {TYPE_SCALE_VARIANTS.map(({key, label}) => {
          const typoRecord = typo as unknown as Record<string, {fontSize?: string} | undefined>;
          return (
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
                value={typoRecord?.[key]?.fontSize ?? ''}
                onChange={(e) =>
                  onUpdate((d) => {
                    const typoMap = d.typography as unknown as Record<string, {fontSize?: string} | undefined>;
                    if (typoMap?.[key]) Object.assign(typoMap[key], {fontSize: e.target.value});
                  })
                }
                placeholder={t('themes.forms.typography_builder.fields.type_scale.placeholder', 'e.g. 1.5rem')}
                sx={{width: 110, '& .MuiInputBase-input': {fontSize: '0.75rem', py: 0.5}}}
              />
            </Stack>
          );
        })}
      </ConfigCard>
    </Stack>
  );
}
