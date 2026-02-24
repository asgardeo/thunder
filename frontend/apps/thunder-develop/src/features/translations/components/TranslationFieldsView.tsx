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

import {Box, IconButton, TextField, Tooltip, Typography} from '@wso2/oxygen-ui';
import {RotateCcw} from '@wso2/oxygen-ui-icons-react';
import {useMemo, type JSX} from 'react';
import {useTranslation} from 'react-i18next';

export interface TranslationFieldsViewProps {
  /** Current display values (server values merged with local edits). */
  localValues: Record<string, string>;
  /** Original server values — used to detect dirtiness and to reset. */
  serverValues: Record<string, string>;
  search: string;
  onChange: (key: string, value: string) => void;
  onResetField: (key: string) => void;
}

export default function TranslationFieldsView({
  localValues,
  serverValues,
  search,
  onChange,
  onResetField,
}: TranslationFieldsViewProps): JSX.Element {
  const {t} = useTranslation('translations');

  const allKeys = Object.keys(localValues);

  const filteredKeys = useMemo(() => {
    const q = search.toLowerCase();
    if (!q) return allKeys;
    return allKeys.filter(
      (k) => k.toLowerCase().includes(q) || (localValues[k] ?? '').toLowerCase().includes(q),
    );
  }, [allKeys, localValues, search]);

  if (filteredKeys.length === 0) {
    return (
      <Box sx={{py: 4, textAlign: 'center', color: 'text.secondary'}}>
        <Typography variant="body2">{t(search ? 'editor.noResults' : 'editor.noKeys')}</Typography>
      </Box>
    );
  }

  return (
    <Box sx={{display: 'flex', flexDirection: 'column', gap: 2}}>
      {filteredKeys.map((key) => {
        const value = localValues[key] ?? '';
        const serverValue = serverValues[key] ?? '';
        const isDirty = value !== serverValue;

        return (
          <Box key={key}>
            <Typography
              variant="caption"
              color="text.secondary"
              sx={{
                display: 'block',
                mb: 0.5,
                fontFamily: 'monospace',
                fontSize: '0.7rem',
                fontWeight: isDirty ? 600 : 400,
                color: isDirty ? 'warning.main' : 'text.secondary',
              }}
            >
              {key}
            </Typography>
            <Box sx={{display: 'flex', gap: 0.5, alignItems: 'flex-start'}}>
              <TextField
                size="small"
                fullWidth
                multiline
                minRows={1}
                maxRows={5}
                value={value}
                onChange={(e) => onChange(key, e.target.value)}
                sx={{
                  '& .MuiOutlinedInput-root': isDirty
                    ? {
                        '& fieldset': {borderColor: 'warning.main'},
                        '&:hover fieldset': {borderColor: 'warning.dark'},
                        '&.Mui-focused fieldset': {borderColor: 'warning.main'},
                      }
                    : {},
                }}
              />
              {isDirty && (
                <Tooltip title={t('editor.resetField')}>
                  <IconButton size="small" onClick={() => onResetField(key)} sx={{mt: 0.25, flexShrink: 0}}>
                    <RotateCcw size={14} />
                  </IconButton>
                </Tooltip>
              )}
            </Box>
          </Box>
        );
      })}
    </Box>
  );
}
