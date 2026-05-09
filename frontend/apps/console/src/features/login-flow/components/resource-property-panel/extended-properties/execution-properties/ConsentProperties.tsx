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

import {FormHelperText, FormLabel, Stack, TextField, Typography} from '@wso2/oxygen-ui';
import {useMemo, type ReactNode} from 'react';
import {useTranslation} from 'react-i18next';
import type {CommonResourcePropertiesPropsInterface} from './types';
import type {StepData} from '@/features/flows/models/steps';

function ConsentProperties({resource, onChange}: CommonResourcePropertiesPropsInterface): ReactNode {
  const {t} = useTranslation();

  const currentTimeout = useMemo(() => {
    const stepData = resource?.data as StepData | undefined;
    return (stepData?.properties as {timeout?: string})?.timeout ?? '0';
  }, [resource]);

  const handleTimeoutChange = (value: string): void => {
    if (value === '') {
      onChange('data.properties.timeout', '0', resource, true);
      return;
    }

    const num = Number(value);
    if (Number.isNaN(num)) {
      return;
    }

    const parsed = Math.max(0, Math.floor(num));
    onChange('data.properties.timeout', String(parsed), resource, true);
  };

  return (
    <Stack gap={2}>
      <Typography variant="body2" color="text.secondary">
        {t('flows:core.executions.consent.description')}
      </Typography>

      <div>
        <FormLabel htmlFor="consent-timeout">{t('flows:core.executions.consent.timeout.label')}</FormLabel>
        <TextField
          id="consent-timeout"
          value={currentTimeout}
          onChange={(e) => handleTimeoutChange(e.target.value)}
          placeholder={t('flows:core.executions.consent.timeout.placeholder')}
          fullWidth
          size="small"
          type="number"
          inputProps={{min: 0}}
        />
        <FormHelperText>{t('flows:core.executions.consent.timeout.hint')}</FormHelperText>
      </div>
    </Stack>
  );
}

export default ConsentProperties;
