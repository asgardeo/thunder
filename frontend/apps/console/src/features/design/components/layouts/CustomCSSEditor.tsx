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

import type {Stylesheet} from '@thunderid/design';
import {Box, Button, Stack, Typography, useColorScheme} from '@wso2/oxygen-ui';
import {Plus} from '@wso2/oxygen-ui-icons-react';
import {forwardRef, useImperativeHandle, useRef, useState, type JSX} from 'react';
import {useTranslation} from 'react-i18next';
import StylesheetItem from './StylesheetItem';

export interface CustomCSSEditorHandle {
  /** Flush all pending debounced edits synchronously. */
  flush: () => void;
}

export interface CustomCSSEditorProps {
  stylesheets: Stylesheet[];
  onChange: (stylesheets: Stylesheet[]) => void;
}

/** Generates a short unique ID like "custom-1", "custom-2", etc. */
function nextId(stylesheets: Stylesheet[]): string {
  const existing = new Set(stylesheets.map((s) => s.id));
  let n = stylesheets.length + 1;
  while (existing.has(`custom-${n}`)) n += 1;
  return `custom-${n}`;
}

const CustomCSSEditor = forwardRef<CustomCSSEditorHandle, CustomCSSEditorProps>(
  ({stylesheets, onChange}, ref): JSX.Element => {
    const {t} = useTranslation('design');
    const {mode, systemMode} = useColorScheme();
    const colorMode: 'light' | 'dark' = (mode === 'system' ? systemMode : mode) === 'dark' ? 'dark' : 'light';

    const [expandedIdx, setExpandedIdx] = useState<number | null>(null);

    // Track flush callbacks from InlineCSSField instances
    const flushMapRef = useRef<Map<number, () => void>>(new Map());

    useImperativeHandle(ref, () => ({
      flush: () => {
        flushMapRef.current.forEach((fn) => fn());
      },
    }));

    // Stable React keys — not tied to the editable `id` field.
    const [keyCounter, setKeyCounter] = useState(stylesheets.length);
    const nextKeyRef = useRef(keyCounter);
    const nextKey = (): number => {
      nextKeyRef.current += 1;
      setKeyCounter(nextKeyRef.current);
      return nextKeyRef.current;
    };
    const [stableKeys, setStableKeys] = useState<number[]>(() =>
      Array.from({length: stylesheets.length}, (_, i) => i + 1),
    );

    // Sync stable keys when stylesheets are replaced externally (e.g. server load).
    const [prevLength, setPrevLength] = useState(stylesheets.length);
    if (stylesheets.length !== prevLength && stableKeys.length !== stylesheets.length) {
      setPrevLength(stylesheets.length);
      const newKeys = Array.from({length: stylesheets.length}, (_, i) => keyCounter + i + 1);
      setStableKeys(newKeys);
      setKeyCounter(keyCounter + stylesheets.length);
    }

    const handleAdd = (type: 'inline' | 'url'): void => {
      const id = nextId(stylesheets);
      const newSheet: Stylesheet = type === 'inline' ? {id, type: 'inline', content: ''} : {id, type: 'url', href: ''};
      const updated = [...stylesheets, newSheet];
      onChange(updated);
      setStableKeys((prev) => [...prev, nextKey()]);
      setExpandedIdx(updated.length - 1);
    };

    const handleRemove = (idx: number): void => {
      const updated = stylesheets.filter((_, i) => i !== idx);
      onChange(updated);
      setStableKeys((prev) => prev.filter((_, i) => i !== idx));
      if (expandedIdx === idx) setExpandedIdx(null);
      else if (expandedIdx !== null && expandedIdx > idx) setExpandedIdx(expandedIdx - 1);
    };

    const handleMove = (idx: number, direction: -1 | 1): void => {
      const target = idx + direction;
      if (target < 0 || target >= stylesheets.length) return;

      // Collapse so Monaco editors are not visible during reorder.
      setExpandedIdx(null);

      const updated = [...stylesheets];
      [updated[idx], updated[target]] = [updated[target], updated[idx]];
      onChange(updated);

      // Generate fresh keys so React fully unmounts/remounts items
      // instead of trying to move DOM nodes (which corrupts Monaco).
      setStableKeys(updated.map(() => nextKey()));
    };

    const handleUpdate = (idx: number, patch: Partial<Stylesheet>): void => {
      // Enforce unique IDs — reject rename if another stylesheet already uses the same id.
      if (patch.id !== undefined && stylesheets.some((s, i) => i !== idx && s.id === patch.id)) {
        return;
      }
      const updated = stylesheets.map((s, i) => (i === idx ? {...s, ...patch} : s));
      onChange(updated as Stylesheet[]);
    };

    return (
      <Stack gap={1}>
        {stylesheets.length === 0 && (
          <Box sx={{py: 3, display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 1}}>
            <Typography variant="body2" color="text.disabled" sx={{fontSize: '0.8rem', textAlign: 'center'}}>
              {t('layouts.config.custom_css.empty_state.message', 'No custom stylesheets yet.')}
            </Typography>
            <Typography variant="caption" color="text.disabled" sx={{fontSize: '0.7rem', textAlign: 'center'}}>
              {t(
                'layouts.config.custom_css.empty_state.description',
                'Add an inline stylesheet or link an external CSS file to customize the appearance.',
              )}
            </Typography>
          </Box>
        )}

        {stylesheets.map((sheet, idx) => (
          <StylesheetItem
            key={stableKeys[idx]}
            sheet={sheet}
            idx={idx}
            total={stylesheets.length}
            expanded={expandedIdx === idx}
            colorMode={colorMode}
            onToggle={() => setExpandedIdx(expandedIdx === idx ? null : idx)}
            onRemove={() => handleRemove(idx)}
            onMove={(dir) => handleMove(idx, dir)}
            onUpdate={(patch) => handleUpdate(idx, patch)}
            registerFlush={(flush) => {
              const key = stableKeys[idx];
              if (flush) flushMapRef.current.set(key, flush);
              else flushMapRef.current.delete(key);
            }}
          />
        ))}

        <Stack direction="row" spacing={0.75} sx={{mt: 0.25}}>
          <Button
            size="small"
            variant="text"
            startIcon={<Plus size={12} />}
            onClick={() => handleAdd('inline')}
            sx={{textTransform: 'none', fontSize: '0.7rem', color: 'text.secondary', px: 1, minWidth: 0}}
          >
            {t('layouts.config.custom_css.actions.add_inline.label', 'Inline')}
          </Button>
          <Button
            size="small"
            variant="text"
            startIcon={<Plus size={12} />}
            onClick={() => handleAdd('url')}
            sx={{textTransform: 'none', fontSize: '0.7rem', color: 'text.secondary', px: 1, minWidth: 0}}
          >
            {t('layouts.config.custom_css.actions.add_url.label', 'External URL')}
          </Button>
        </Stack>
      </Stack>
    );
  },
);

CustomCSSEditor.displayName = 'CustomCSSEditor';

export default CustomCSSEditor;
