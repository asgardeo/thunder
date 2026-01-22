/**
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
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

import {TextField} from '@wso2/oxygen-ui';
import {useTranslation} from 'react-i18next';
import {Controller} from 'react-hook-form';
import type {Control, FieldErrors} from 'react-hook-form';
import SettingsCard from '../SettingsCard';

interface TokenValidationSectionProps {
  control: Control<{validityPeriod?: number; accessTokenValidity?: number; idTokenValidity?: number}>;
  errors: FieldErrors<{validityPeriod?: number; accessTokenValidity?: number; idTokenValidity?: number}>;
  tokenType: 'shared' | 'access' | 'id';
}

export default function TokenValidationSection({control, errors, tokenType}: TokenValidationSectionProps) {
  const {t} = useTranslation();

  const getTitle = () => {
    if (tokenType === 'access') return t('applications:edit.token.accessTokenValidation.title');
    if (tokenType === 'id') return t('applications:edit.token.idTokenValidation.title');
    return t('applications:edit.token.tokenValidation.title');
  };

  const getDescription = () => {
    if (tokenType === 'access') return t('applications:edit.token.accessTokenValidation.description');
    if (tokenType === 'id') return t('applications:edit.token.idTokenValidation.description');
    return t('applications:edit.token.tokenValidation.description');
  };

  const fieldName =
    tokenType === 'shared' ? 'validityPeriod' : tokenType === 'access' ? 'accessTokenValidity' : 'idTokenValidity';

  return (
    <SettingsCard title={getTitle()} description={getDescription()}>
      <Controller
        name={fieldName}
        control={control}
        render={({field}) => (
          <TextField
            {...field}
            fullWidth
            label={t('applications:edit.token.labels.tokenValidity')}
            type="number"
            onChange={(e) => field.onChange(parseInt(e.target.value, 10))}
            error={!!errors[fieldName]}
            helperText={errors[fieldName]?.message ?? t('applications:edit.token.validity.hint')}
            inputProps={{min: 1}}
          />
        )}
      />
    </SettingsCard>
  );
}
