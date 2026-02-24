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

import {Autocomplete, Box, CircularProgress, Divider, TextField, Typography} from '@wso2/oxygen-ui';
import {Plus} from '@wso2/oxygen-ui-icons-react';
import {type JSX, type SyntheticEvent} from 'react';
import {useTranslation} from 'react-i18next';

const ADD_NEW_SENTINEL = '__add_new__';

export interface LanguageAutocompleteProps {
  options: string[];
  value: string | null;
  onChange: (value: string | null) => void;
  onAddNew: () => void;
  loading?: boolean;
}

export default function LanguageAutocomplete({
  options,
  value,
  onChange,
  onAddNew,
  loading = false,
}: LanguageAutocompleteProps): JSX.Element {
  const {t} = useTranslation('translations');

  const allOptions = [...options, ADD_NEW_SENTINEL];

  const handleChange = (_event: SyntheticEvent, newValue: string | null) => {
    if (newValue === ADD_NEW_SENTINEL) {
      onAddNew();
      return;
    }
    onChange(newValue);
  };

  return (
    <Autocomplete
      options={allOptions}
      value={value}
      onChange={handleChange}
      loading={loading}
      size="small"
      disableClearable={false}
      getOptionLabel={(opt) => (opt === ADD_NEW_SENTINEL ? '' : opt)}
      isOptionEqualToValue={(opt, val) => opt === val}
      renderInput={(params) => (
        <TextField
          {...params}
          label={t('language.selectPlaceholder')}
          InputProps={{
            ...params.InputProps,
            endAdornment: (
              <>
                {loading && <CircularProgress color="inherit" size={16} />}
                {params.InputProps.endAdornment}
              </>
            ),
          }}
        />
      )}
      renderOption={(props, option) => {
        if (option === ADD_NEW_SENTINEL) {
          const {key, ...rest} = props as {key: unknown} & React.HTMLAttributes<HTMLLIElement>;
          return (
            <Box key={String(key)} component="li" {...rest} sx={{py: 0, px: 0}}>
              <Divider sx={{width: '100%', mx: 0}} />
              <Box
                sx={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 1,
                  px: 2,
                  py: 1,
                  width: '100%',
                  color: 'primary.main',
                }}
              >
                <Plus size={14} />
                <Typography variant="body2" color="primary">
                  {t('language.addOption')}
                </Typography>
              </Box>
            </Box>
          );
        }
        const {key, ...rest} = props as {key: unknown} & React.HTMLAttributes<HTMLLIElement>;
        return (
          <Box key={String(key)} component="li" {...rest}>
            <Typography variant="body2">{option}</Typography>
          </Box>
        );
      }}
      filterOptions={(opts, state) => {
        const filtered = opts.filter(
          (opt) => opt === ADD_NEW_SENTINEL || opt.toLowerCase().includes(state.inputValue.toLowerCase()),
        );
        // Always include the "add new" sentinel
        if (!filtered.includes(ADD_NEW_SENTINEL)) filtered.push(ADD_NEW_SENTINEL);
        return filtered;
      }}
    />
  );
}
