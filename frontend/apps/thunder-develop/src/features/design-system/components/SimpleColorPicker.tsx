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
import {
  Box,
  ClickAwayListener,
  FormControl,
  FormLabel,
  InputAdornment,
  Popper,
  Stack,
  TextField,
  Typography,
} from '@mui/material';
import {HexColorPicker} from 'react-colorful';

/**
 * Props for the SimpleColorPicker component.
 */
export interface SimpleColorPickerProps {
  /** Label for the field */
  label: string;
  /** Current color value */
  value: string;
  /** Callback when color changes */
  onChange: (color: string) => void;
  /** Whether the field is required */
  required?: boolean;
}

/**
 * Simple color picker component without react-hook-form dependency.
 * Used for standalone color picking in theme controls.
 */
export default function SimpleColorPicker({
  label,
  value,
  onChange,
  required = false,
}: SimpleColorPickerProps): JSX.Element {
  const [pickerOpen, setPickerOpen] = useState(false);
  const anchorRef = useRef<HTMLDivElement>(null);

  const handleClickAway = useCallback(() => {
    setPickerOpen(false);
  }, []);

  const isValidHex = /^#[0-9A-F]{6}$/i.test(value);

  return (
    <FormControl fullWidth>
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
          value={value}
          onChange={(e) => onChange(e.target.value)}
          size="small"
          fullWidth
          placeholder="#000000"
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

      {/* Color picker popper */}
      {pickerOpen && (
        <ClickAwayListener onClickAway={handleClickAway}>
          <Popper open={pickerOpen} anchorEl={anchorRef.current} placement="bottom-start" sx={{zIndex: 1300}}>
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
            </Box>
          </Popper>
        </ClickAwayListener>
      )}
    </FormControl>
  );
}
