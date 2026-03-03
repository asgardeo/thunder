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

import {Box, Chip, FormControl, FormLabel, Stack, TextField, Typography, useTheme} from '@wso2/oxygen-ui';
import {getDisplayNameForCode, toFlagEmoji} from '@thunder/i18n';
import type {LocaleOption} from '@thunder/i18n';
import {useEffect, useMemo, type JSX} from 'react';
import {useTranslation} from 'react-i18next';
import {Lightbulb} from '@wso2/oxygen-ui-icons-react';

/**
 * Props for the {@link ReviewLocaleCode} component.
 *
 * @public
 */
export interface ReviewLocaleCodeProps {
  /** The locale code derived from the previous steps — used as the default value. */
  derivedLocale: LocaleOption;
  /** The current override value entered by the user (controlled). */
  localeCode: string;
  /** Callback invoked when the user edits the locale code input. */
  onLocaleCodeChange: (code: string) => void;
  /** Callback invoked whenever the step readiness changes (e.g. input becomes non-empty). */
  onReadyChange?: (isReady: boolean) => void;
}

/**
 * Step component in the language creation wizard that allows the user to review
 * and optionally override the BCP 47 locale code derived from the country and
 * language selections.
 *
 * Shows a preview of the flag emoji and resolved display name for the effective
 * locale code, along with a helper tip about the BCP 47 format.
 *
 * @param props - The component props
 * @param props.derivedLocale - Locale derived from the previous wizard steps, used as the default
 * @param props.localeCode - Current user-entered override value (controlled)
 * @param props.onLocaleCodeChange - Callback invoked when the locale code input changes
 * @param props.onReadyChange - Callback invoked when step readiness changes
 *
 * @returns JSX element rendering the locale code review step
 *
 * @example
 * ```tsx
 * import ReviewLocaleCode from './ReviewLocaleCode';
 *
 * function Wizard() {
 *   const [code, setCode] = useState('');
 *   return (
 *     <ReviewLocaleCode
 *       derivedLocale={{code: 'fr-FR', displayName: 'French (France)', flag: '🇫🇷'}}
 *       localeCode={code}
 *       onLocaleCodeChange={setCode}
 *       onReadyChange={(ready) => setStepReady(ready)}
 *     />
 *   );
 * }
 * ```
 *
 * @public
 */
export default function ReviewLocaleCode({
  derivedLocale,
  localeCode,
  onLocaleCodeChange,
  onReadyChange = undefined,
}: ReviewLocaleCodeProps): JSX.Element {
  const theme = useTheme();
  const {t} = useTranslation('translations');

  const effectiveCode = localeCode.trim() || derivedLocale.code;
  const isLocaleCodeValid = /^[A-Za-z]{2,3}(?:-[A-Za-z]{4})?(?:-(?:[A-Za-z]{2}|\d{3}))?$/.test(effectiveCode);

  const resolvedName = useMemo(() => getDisplayNameForCode(effectiveCode), [effectiveCode]);

  const previewFlag = toFlagEmoji(effectiveCode.split('-')[1]?.toUpperCase() ?? '');

  useEffect(() => {
    onReadyChange?.(isLocaleCodeValid);
  }, [isLocaleCodeValid, onReadyChange]);

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

      <Stack direction="row" alignItems="center" spacing={1}>
        <Lightbulb size={20} color={theme.vars?.palette.warning.main} />
        <Typography variant="body2" color="text.secondary">
          {t('language.add.code.helperText')}
        </Typography>
      </Stack>
    </Stack>
  );
}
