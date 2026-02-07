/**
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
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

import type {JSX} from 'react';
import {useState, useRef, useCallback} from 'react';
import {Box, ClickAwayListener, FormControl, FormHelperText, FormLabel, InputAdornment, Popper, Stack, TextField, Typography} from '@mui/material';
import {HexColorPicker} from 'react-colorful';
import type {Control, FieldValues, Path} from 'react-hook-form';
import {Controller} from 'react-hook-form';

/**
 * Props for the ColorPickerField component.
 */
export interface ColorPickerFieldProps<TFieldValues extends FieldValues = FieldValues> {
  /** Field name for react-hook-form */
  name: Path<TFieldValues>;
  /** Form control from useForm */
  control: Control<TFieldValues>;
  /** Label for the field */
  label: string;
  /** Helper text to display below the field */
  helperText?: string;
  /** Whether the field is required */
  required?: boolean;
  /** Hex color to check contrast against (for WCAG validation) */
  contrastCheckAgainst?: string;
}

/**
 * Calculate relative luminance of a color (used for WCAG contrast calculation).
 * Formula from WCAG 2.0: https://www.w3.org/TR/WCAG20/#relativeluminancedef
 */
function getLuminance(hexColor: string): number {
  const rgb = hexColor
    .replace('#', '')
    .match(/.{2}/g)
    ?.map((hex) => parseInt(hex, 16) / 255) || [0, 0, 0];

  const [r, g, b] = rgb.map((val) => {
    return val <= 0.03928 ? val / 12.92 : Math.pow((val + 0.055) / 1.055, 2.4);
  });

  return 0.2126 * r + 0.7152 * g + 0.0722 * b;
}

/**
 * Calculate contrast ratio between two colors (WCAG 2.0 formula).
 * Returns a value between 1 and 21.
 */
function getContrastRatio(color1: string, color2: string): number {
  const lum1 = getLuminance(color1);
  const lum2 = getLuminance(color2);
  const lighter = Math.max(lum1, lum2);
  const darker = Math.min(lum1, lum2);
  return (lighter + 0.05) / (darker + 0.05);
}

/**
 * Color picker field component with visual picker, hex validation, and WCAG contrast checking.
 * Integrates with react-hook-form for form validation and state management.
 *
 * @param props - Component props
 * @returns ColorPickerField component
 *
 * @example
 * ```tsx
 * const { control, watch } = useForm<ThemeFormData>({
 *   resolver: zodResolver(themeSchema),
 *   defaultValues: initialTheme,
 * });
 *
 * const backgroundColor = watch('colorSchemes.dark.colors.background.default');
 *
 * <ColorPickerField
 *   name="colorSchemes.dark.colors.primary.main"
 *   control={control}
 *   label="Primary Color"
 *   required
 *   contrastCheckAgainst={backgroundColor}
 * />
 * ```
 */
export default function ColorPickerField<TFieldValues extends FieldValues = FieldValues>({
  name,
  control,
  label,
  helperText,
  required = false,
  contrastCheckAgainst,
}: ColorPickerFieldProps<TFieldValues>): JSX.Element {
  const [pickerOpen, setPickerOpen] = useState(false);
  const anchorRef = useRef<HTMLDivElement>(null);

  const handleClickAway = useCallback(() => {
    setPickerOpen(false);
  }, []);

  return (
    <Controller
      name={name}
      control={control}
      render={({field, fieldState: {error}}) => {
        const {value, onChange} = field;
        const isValidHex = /^#[0-9A-F]{6}$/i.test(value);

        // Calculate contrast ratio if both colors are valid hex
        let contrastRatio: number | null = null;
        let contrastWarning: string | null = null;

        if (isValidHex && contrastCheckAgainst && /^#[0-9A-F]{6}$/i.test(contrastCheckAgainst)) {
          contrastRatio = getContrastRatio(value, contrastCheckAgainst);
          // WCAG AA requires 4.5:1 for normal text, 3:1 for large text
          if (contrastRatio < 4.5) {
            contrastWarning = `Low contrast ratio: ${contrastRatio.toFixed(2)}:1 (WCAG AA requires 4.5:1)`;
          }
        }

        return (
          <FormControl error={!!error} fullWidth>
            <FormLabel required={required} sx={{mb: 1}}>
              {label}
            </FormLabel>
            <Stack direction="row" spacing={1}>
              {/* Color swatch button */}
              <Box
                ref={anchorRef}
                onClick={() => setPickerOpen(!pickerOpen)}
                sx={{
                  width: 40,
                  height: 40,
                  borderRadius: 1,
                  backgroundColor: isValidHex ? value : '#cccccc',
                  border: '1px solid',
                  borderColor: 'divider',
                  cursor: 'pointer',
                  flexShrink: 0,
                  transition: 'all 0.2s',
                  '&:hover': {
                    borderColor: 'primary.main',
                    boxShadow: 1,
                  },
                }}
                aria-label={`Pick color for ${label}`}
              />

              {/* Hex input */}
              <TextField
                {...field}
                size="small"
                fullWidth
                placeholder="#000000"
                error={!!error}
                InputProps={{
                  startAdornment: !value.startsWith('#') ? (
                    <InputAdornment position="start">
                      <Typography variant="body2" color="text.secondary">
                        #
                      </Typography>
                    </InputAdornment>
                  ) : null,
                }}
                inputProps={{
                  maxLength: 7,
                  style: {textTransform: 'uppercase'},
                }}
              />
            </Stack>

            {/* Helper text and validation messages */}
            {(helperText || error?.message || contrastWarning) && (
              <FormHelperText>
                {error?.message || contrastWarning || helperText}
              </FormHelperText>
            )}

            {/* Color picker popper */}
            {pickerOpen && (
              <ClickAwayListener onClickAway={handleClickAway}>
                <Popper
                  open={pickerOpen}
                  anchorEl={anchorRef.current}
                  placement="bottom-start"
                  sx={{zIndex: 1300}}
                >
                  <Box
                    sx={{
                      mt: 1,
                      p: 2,
                      bgcolor: 'background.paper',
                      borderRadius: 1,
                      boxShadow: 8,
                      border: '1px solid',
                      borderColor: 'divider',
                    }}
                  >
                    <HexColorPicker
                      color={isValidHex ? value : '#000000'}
                      onChange={onChange}
                      style={{width: 200, height: 150}}
                    />
                    {contrastRatio !== null && (
                      <Typography
                        variant="caption"
                        color={contrastRatio >= 4.5 ? 'success.main' : 'warning.main'}
                        sx={{display: 'block', mt: 1, textAlign: 'center'}}
                      >
                        Contrast: {contrastRatio.toFixed(2)}:1
                      </Typography>
                    )}
                  </Box>
                </Popper>
              </ClickAwayListener>
            )}
          </FormControl>
        );
      }}
    />
  );
}
