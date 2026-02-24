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

import {Box, Chip, FormControl, FormHelperText, FormLabel, Stack, TextField, Typography} from '@wso2/oxygen-ui';
import {getDisplayNameForCode, toFlagEmoji} from '@thunder/i18n';
import type {LocaleOption} from '@thunder/i18n';
import {useEffect, useMemo, type JSX} from 'react';
import {useTranslation} from 'react-i18next';

export interface ReviewLocaleCodeProps {
  /** The locale code derived from the previous steps — used as the default value. */
  derivedLocale: LocaleOption;
  /** The current override value (controlled). */
  localeCode: string;
  onLocaleCodeChange: (code: string) => void;
  onReadyChange?: (isReady: boolean) => void;
}

export default function ReviewLocaleCode({
  derivedLocale,
  localeCode,
  onLocaleCodeChange,
  onReadyChange,
}: ReviewLocaleCodeProps): JSX.Element {
  const {t} = useTranslation('translations');

  const effectiveCode = localeCode.trim() || derivedLocale.code;

  const resolvedName = useMemo(() => getDisplayNameForCode(effectiveCode), [effectiveCode]);

  const previewFlag = toFlagEmoji(effectiveCode.split('-')[1]?.toUpperCase() ?? '');

  useEffect(() => {
    onReadyChange?.(effectiveCode.length > 0);
  }, [effectiveCode, onReadyChange]);

  return (
    <Stack spacing={4}>
      <Box>
        <Typography variant="h1" gutterBottom>
          {t('language.create.localeCode.title')}
        </Typography>
        <Typography variant="body1" color="text.secondary">
          {t('language.create.localeCode.subtitle')}
        </Typography>
      </Box>

      <Box>
        <FormControl required fullWidth>
          <FormLabel htmlFor="locale-code-input">{t('language.add.code.label')}</FormLabel>
          <TextField
            id="locale-code-input"
            placeholder={derivedLocale.code}
            value={localeCode}
            onChange={(e) => onLocaleCodeChange(e.target.value)}
            fullWidth
            autoFocus
          />
          <FormHelperText>{t('language.add.codeHelper')}</FormHelperText>
        </FormControl>

        {effectiveCode && (
          <Stack direction="row" alignItems="center" spacing={1} sx={{mt: 1.5}}>
            <Typography sx={{fontSize: '1.1rem', lineHeight: 1}}>{previewFlag}</Typography>
            {resolvedName && (
              <Typography variant="body2" color="text.secondary">
                {resolvedName}
              </Typography>
            )}
            <Chip
              label={effectiveCode}
              size="small"
              variant="outlined"
              sx={{fontFamily: 'monospace', fontSize: '0.7rem'}}
            />
          </Stack>
        )}
      </Box>
    </Stack>
  );
}
