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

import {useEffect} from 'react';
import {Stack, Typography, FormControl, FormLabel, Select, MenuItem} from '@wso2/oxygen-ui';
import type {JSX} from 'react';
import {useTranslation} from 'react-i18next';
import type {SchemaInterface} from '../../types/users';

/**
 * Props for the {@link ConfigureUserType} component.
 *
 * @public
 */
export interface ConfigureUserTypeProps {
  schemas: SchemaInterface[];
  selectedSchema: SchemaInterface | null;
  onSchemaChange: (schema: SchemaInterface | null) => void;
  onReadyChange?: (isReady: boolean) => void;
}

/**
 * Step 1 of the user creation wizard: select a user type (schema).
 *
 * @public
 */
export default function ConfigureUserType({
  schemas,
  selectedSchema,
  onSchemaChange,
  onReadyChange = undefined,
}: ConfigureUserTypeProps): JSX.Element {
  const {t} = useTranslation();

  useEffect((): void => {
    if (onReadyChange) {
      onReadyChange(selectedSchema !== null);
    }
  }, [selectedSchema, onReadyChange]);

  return (
    <Stack direction="column" spacing={4} data-testid="configure-user-type">
      <Typography variant="h1" gutterBottom>
        {t('users:createWizard.selectUserType.title')}
      </Typography>
      <Typography variant="body1" color="text.secondary">
        {t('users:createWizard.selectUserType.subtitle')}
      </Typography>

      <FormControl fullWidth required>
        <FormLabel htmlFor="user-type-select">{t('users:createWizard.selectUserType.fieldLabel')}</FormLabel>
        <Select
          id="user-type-select"
          value={selectedSchema?.id ?? ''}
          onChange={(e) => {
            const schema = schemas.find((s) => s.id === e.target.value);
            onSchemaChange(schema ?? null);
          }}
          displayEmpty
          data-testid="user-type-select"
        >
          <MenuItem value="" disabled>
            <em>{t('users:createWizard.selectUserType.placeholder')}</em>
          </MenuItem>
          {schemas.map((schema) => (
            <MenuItem key={schema.id} value={schema.id}>
              {schema.name}
            </MenuItem>
          ))}
        </Select>
      </FormControl>
    </Stack>
  );
}
