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

import {buildCountryOptions} from '@thunderid/i18n';
import type {CountryOption} from '@thunderid/i18n';
import {Autocomplete, Box, Chip, FormControl, FormLabel, Stack, TextField, Typography, useTheme} from '@wso2/oxygen-ui';
import {Lightbulb} from '@wso2/oxygen-ui-icons-react';
import {useEffect, useMemo, type JSX} from 'react';
import {useTranslation} from 'react-i18next';

/**
 * Props for the {@link SelectCountry} component.
 *
 * @public
 */
export interface SelectCountryProps {
  /** Currently selected country, or null if none has been chosen. */
  selectedCountry: CountryOption | null;
  /** Callback invoked when the user selects or clears a country. */
  onCountryChange: (country: CountryOption | null) => void;
  /** Callback invoked whenever step readiness changes (i.e. a country becomes selected). */
  onReadyChange?: (isReady: boolean) => void;
}

/**
 * First step in the language creation wizard where the user selects the country
 * associated with the new language.
 *
 * Renders a searchable autocomplete populated with all available country options.
 * Each option shows the country flag, name, and ISO region code. A helper tip
 * below explains how the country selection influences the generated locale code.
 *
 * @param props - The component props
 * @param props.selectedCountry - Currently selected country option
 * @param props.onCountryChange - Callback invoked when the country selection changes
 * @param props.onReadyChange - Callback invoked when step readiness changes
 *
 * @returns JSX element rendering the country selection step
 *
 * @example
 * ```tsx
 * import SelectCountry from './SelectCountry';
 *
 * function Wizard() {
 *   const [country, setCountry] = useState<CountryOption | null>(null);
 *   return (
 *     <SelectCountry
 *       selectedCountry={country}
 *       onCountryChange={setCountry}
 *       onReadyChange={(ready) => setStepReady(ready)}
 *     />
 *   );
 * }
 * ```
 *
 * @public
 */
export default function SelectCountry({
  selectedCountry,
  onCountryChange,
  onReadyChange = undefined,
}: SelectCountryProps): JSX.Element {
  const theme = useTheme();
  const {t} = useTranslation('translations');

  const countryOptions = useMemo(() => buildCountryOptions(), []);

  useEffect(() => {
    onReadyChange?.(!!selectedCountry);
  }, [selectedCountry, onReadyChange]);

  return (
    <Stack spacing={4}>
      <Box>
        <Typography variant="h1" gutterBottom>
          {t('language.create.country.title')}
        </Typography>
        <Typography variant="body1" color="text.secondary">
          {t('language.create.country.subtitle')}
        </Typography>
      </Box>

      <FormControl required fullWidth>
        <FormLabel htmlFor="country-select">{t('language.create.countryLabel')}</FormLabel>
        <Autocomplete
          id="country-select"
          options={countryOptions}
          value={selectedCountry}
          onChange={(_, v) => onCountryChange(v)}
          getOptionLabel={(opt) => opt.name}
          filterOptions={(opts, state) => {
            const input = state.inputValue.toLowerCase();
            return opts.filter(
              (opt) => opt.name.toLowerCase().includes(input) || opt.regionCode.toLowerCase().includes(input),
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
                  {opt.name}
                </Typography>
                <Chip
                  label={opt.regionCode}
                  size="small"
                  variant="outlined"
                  sx={{fontFamily: 'monospace', fontSize: '0.7rem'}}
                />
              </Box>
            );
          }}
          renderInput={(params) => <TextField placeholder={t('language.create.country.placeholder')} {...params} />}
        />
      </FormControl>

      <Stack direction="row" alignItems="center" spacing={1}>
        <Lightbulb size={20} color={theme.vars?.palette.warning.main} />
        <Typography variant="body2" color="text.secondary">
          {t('language.create.country.helperText')}
        </Typography>
      </Stack>
    </Stack>
  );
}
