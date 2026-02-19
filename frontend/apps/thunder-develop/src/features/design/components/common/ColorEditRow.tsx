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

import {Box, Stack, TextField, Typography} from '@wso2/oxygen-ui';
import {useEffect, useRef, useState, type JSX} from 'react';

export interface ColorEditRowProps {
  label: string;
  value: string;
  onChange: (v: string) => void;
  compact?: boolean;
}

/**
 * ColorEditRow - A color picker and text input for editing color values.
 * Supports two modes:
 * - compact: Inline row with small swatch (for default accordion mode)
 * - full: Large swatch and text input (for builder mode)
 */
export default function ColorEditRow({label, value, onChange, compact = false}: ColorEditRowProps): JSX.Element {
  const pickerRef = useRef<HTMLInputElement>(null);
  const isFocused = useRef(false);
  const [editValue, setEditValue] = useState(value);
  const isHex = /^#[0-9a-fA-F]{6}$/i.test(value);

  // Sync when external value changes, but not while the text input is focused
  useEffect(() => {
    if (!isFocused.current) setEditValue(value);
  }, [value]);

  const handleText = (e: React.ChangeEvent<HTMLInputElement>): void => {
    const v = e.target.value;
    setEditValue(v);
    if (/^#[0-9a-fA-F]{6}$/i.test(v)) onChange(v);
  };

  if (compact) {
    // Legacy compact row for default accordion mode
    return (
      <Stack direction="row" alignItems="center" justifyContent="space-between" sx={{py: 0.5}}>
        <Typography variant="caption" color="text.secondary" sx={{fontSize: '0.75rem'}}>
          {label}
        </Typography>
        <Stack direction="row" alignItems="center" spacing={0.75}>
          <Box
            onClick={() => isHex && pickerRef.current?.click()}
            sx={{
              width: 16,
              height: 16,
              borderRadius: 0.5,
              bgcolor: value,
              border: '1px solid rgba(0,0,0,0.2)',
              cursor: isHex ? 'pointer' : 'default',
              flexShrink: 0,
              '&:hover': isHex ? {outline: '2px solid', outlineColor: 'primary.main', outlineOffset: '1px'} : {},
            }}
          />
          {isHex && (
            <input
              ref={pickerRef}
              type="color"
              value={value}
              onChange={(e) => onChange(e.target.value)}
              style={{position: 'absolute', opacity: 0, width: 0, height: 0, pointerEvents: 'none'}}
            />
          )}
          <Typography variant="caption" sx={{fontFamily: 'monospace', fontSize: '0.7rem', color: 'text.primary'}}>
            {value}
          </Typography>
        </Stack>
      </Stack>
    );
  }

  // Full builder-mode row
  return (
    <Box sx={{mb: 2}}>
      {/* Label + hex hint */}
      <Stack direction="row" alignItems="baseline" justifyContent="space-between" mb={0.875}>
        <Typography variant="body2" sx={{fontWeight: 600, fontSize: '0.875rem'}}>
          {label}
        </Typography>
        <Typography variant="caption" sx={{fontFamily: 'monospace', fontSize: '0.68rem', color: 'text.disabled'}}>
          {isHex ? value.toLowerCase() : value}
        </Typography>
      </Stack>

      {/* Swatch + text input */}
      <Stack direction="row" spacing={1} alignItems="stretch">
        {/* Swatch — opens native color picker */}
        <Box
          onClick={() => pickerRef.current?.click()}
          sx={{
            width: 44,
            height: 44,
            flexShrink: 0,
            bgcolor: value,
            borderRadius: 1,
            border: '1.5px solid',
            borderColor: 'divider',
            cursor: 'pointer',
            transition: 'outline 0.12s',
            '&:hover': {outline: '2.5px solid', outlineColor: 'primary.main', outlineOffset: '2px'},
          }}
        />
        <input
          ref={pickerRef}
          type="color"
          value={isHex ? value : '#000000'}
          onChange={(e) => {
            const v = e.target.value;
            setEditValue(v);
            onChange(v);
          }}
          style={{position: 'absolute', opacity: 0, width: 0, height: 0, pointerEvents: 'none'}}
        />

        <TextField
          value={editValue}
          onChange={handleText}
          onFocus={() => {
            isFocused.current = true;
          }}
          onBlur={() => {
            isFocused.current = false;
            setEditValue(value);
          }}
          spellCheck={false}
          size="small"
          fullWidth
          sx={{
            flex: 1,
            '& .MuiInputBase-root': {
              height: 44,
              fontFamily: 'monospace',
              fontSize: '0.875rem',
            },
          }}
        />
      </Stack>
    </Box>
  );
}
