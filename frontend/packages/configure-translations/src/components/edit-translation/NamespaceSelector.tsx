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

import {Autocomplete, Box, FormControl, FormHelperText, FormLabel, TextField} from '@wso2/oxygen-ui';
import {type JSX} from 'react';
import {useTranslation} from 'react-i18next';

/**
 * Props for the {@link NamespaceSelector} component.
 *
 * @public
 */
export interface NamespaceSelectorProps {
  /** Available namespace options. */
  namespaces: string[];
  /** Currently selected namespace, or null if none. */
  value: string | null;
  /** Whether the namespace list is still loading. */
  loading: boolean;
  /** Called when the user selects a different namespace. */
  onChange: (namespace: string) => void;
}

/**
 * Autocomplete control for selecting a translation namespace.
 *
 * Formats camelCase namespace keys into human-readable labels and shows a
 * helper text below the input.
 *
 * @param props - The component props
 *
 * @returns JSX element rendering the namespace selector
 *
 * @public
 */
export default function NamespaceSelector({namespaces, value, loading, onChange}: NamespaceSelectorProps): JSX.Element {
  const {t} = useTranslation('translations');

  return (
    <Box sx={{display: 'flex', gap: 2, alignItems: 'center', mb: 2}}>
      <FormControl sx={{maxWidth: 600}}>
        <FormLabel htmlFor="namespace-selector">{t('editor.namespace')}</FormLabel>
        <Autocomplete
          id="namespace-selector"
          options={namespaces}
          value={value ?? ''}
          onChange={(_, v) => v && onChange(v)}
          disableClearable
          size="small"
          loading={loading}
          renderInput={(params) => <TextField {...params} />}
          getOptionLabel={(opt) =>
            opt
              .replace(/([A-Z])/g, ' $1')
              .replace(/^./, (c) => c.toUpperCase())
              .trim()
          }
        />
        <FormHelperText>{t('editor.namespace.helperText')}</FormHelperText>
      </FormControl>
    </Box>
  );
}
