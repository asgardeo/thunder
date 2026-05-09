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

import {buildLocaleOptions} from '@thunderid/i18n';
import type {CountryOption, LocaleOption} from '@thunderid/i18n';
import {Autocomplete, Box, Chip, FormControl, FormLabel, Stack, TextField, Typography, useTheme} from '@wso2/oxygen-ui';
import {Lightbulb} from '@wso2/oxygen-ui-icons-react';
import {useEffect, useMemo, type JSX} from 'react';
import {useTranslation} from 'react-i18next';

/**
 * Props for the {@link SelectLanguage} component.
 *
 * @public
 */
export interface SelectLanguageProps {
  /** The country selected in the previous wizard step, used to derive available locale options. */
  selectedCountry: CountryOption;
  /** Currently selected locale option, or null if none has been chosen. */
  selectedLocale: LocaleOption | null;
  /** Callback invoked when the user selects or clears a locale. */
  onLocaleChange: (locale: LocaleOption | null) => void;
  /** Callback invoked whenever step readiness changes (i.e. a locale becomes selected). */
  onReadyChange?: (isReady: boolean) => void;
}

/**
 * Second step in the language creation wizard where the user selects the specific
 * language variant spoken in the previously chosen country.
 *
 * Locale options are derived from the selected country's region code. Each option
 * displays the language flag, display name, and BCP 47 code. A helper tip explains
 * how the language selection contributes to the final locale code.
 *
 * @param props - The component props
 * @param props.selectedCountry - Country chosen in the preceding wizard step
 * @param props.selectedLocale - Currently selected locale option
 * @param props.onLocaleChange - Callback invoked when the locale selection changes
 * @param props.onReadyChange - Callback invoked when step readiness changes
 *
 * @returns JSX element rendering the language selection step
 *
 * @example
 * ```tsx
 * import SelectLanguage from './SelectLanguage';
 *
 * function Wizard() {
 *   const [locale, setLocale] = useState<LocaleOption | null>(null);
 *   return (
 *     <SelectLanguage
 *       selectedCountry={{name: 'France', regionCode: 'FR', flag: '🇫🇷'}}
 *       selectedLocale={locale}
 *       onLocaleChange={setLocale}
 *       onReadyChange={(ready) => setStepReady(ready)}
 *     />
 *   );
 * }
 * ```
 *
 * @public
 */
export default function SelectLanguage({
  selectedCountry,
  selectedLocale,
  onLocaleChange,
  onReadyChange = undefined,
}: SelectLanguageProps): JSX.Element {
  const theme = useTheme();
  const {t} = useTranslation('translations');

  const languageOptions = useMemo(() => buildLocaleOptions(selectedCountry.regionCode), [selectedCountry.regionCode]);

  useEffect(() => {
    onReadyChange?.(!!selectedLocale);
  }, [selectedLocale, onReadyChange]);

  return (
    <Stack spacing={4}>
      <Box>
        <Typography variant="h1" gutterBottom>
          {t('language.create.language.title')}
        </Typography>
        <Typography variant="body1" color="text.secondary">
          {t('language.create.language.subtitle', {country: selectedCountry.name})}
        </Typography>
      </Box>

      <FormControl required fullWidth>
        <FormLabel htmlFor="language-select">{t('language.create.language.label')}</FormLabel>
        <Autocomplete
          id="language-select"
          options={languageOptions}
          value={selectedLocale}
          onChange={(_, v) => onLocaleChange(v)}
          getOptionLabel={(opt) => opt.displayName}
          filterOptions={(opts, state) => {
            const input = state.inputValue.toLowerCase();
            return opts.filter(
              (opt) => opt.code.toLowerCase().includes(input) || opt.displayName.toLowerCase().includes(input),
            );
          }}
          renderOption={(props, opt) => {
            // eslint-disable-next-line react/prop-types
            const {key, ...rest} = props as {key: unknown} & React.HTMLAttributes<HTMLLIElement>;
            return (
              <Box key={String(key)} component="li" {...rest} sx={{display: 'flex', alignItems: 'center', gap: 1.5}}>
                <Typography
                  sx={{fontSize: '1.2rem', lineHeight: 1, userSelect: 'none', width: 28, textAlign: 'center'}}
                >
                  {opt.flag}
                </Typography>
                <Typography variant="body2" sx={{flex: 1}}>
                  {opt.displayName}
                </Typography>
                <Chip
                  label={opt.code}
                  size="small"
                  variant="outlined"
                  sx={{fontFamily: 'monospace', fontSize: '0.7rem'}}
                />
              </Box>
            );
          }}
          renderInput={(params) => <TextField {...params} placeholder={t('language.create.language.placeholder')} />}
        />
      </FormControl>

      <Stack direction="row" alignItems="center" spacing={1}>
        <Lightbulb size={20} color={theme.vars?.palette.warning.main} />
        <Typography variant="body2" color="text.secondary">
          {t('language.create.language.helperText')}
        </Typography>
      </Stack>
    </Stack>
  );
}
