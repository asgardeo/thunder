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
import {Stack, Typography, FormControl, FormLabel} from '@wso2/oxygen-ui';
import type {JSX} from 'react';
import {useTranslation} from 'react-i18next';
import OrganizationUnitTreePicker from '../../../organization-units/components/OrganizationUnitTreePicker';

export interface ConfigureOrganizationUnitProps {
  rootOuId: string;
  selectedOuId: string;
  onOuIdChange: (ouId: string) => void;
  onReadyChange?: (isReady: boolean) => void;
}

export default function ConfigureOrganizationUnit({
  rootOuId,
  selectedOuId,
  onOuIdChange,
  onReadyChange = undefined,
}: ConfigureOrganizationUnitProps): JSX.Element {
  const {t} = useTranslation();

  useEffect(() => {
    if (!selectedOuId) {
      onOuIdChange(rootOuId);
    }
  }, [selectedOuId, rootOuId, onOuIdChange]);

  useEffect((): void => {
    if (onReadyChange) {
      onReadyChange(selectedOuId.length > 0);
    }
  }, [selectedOuId, onReadyChange]);

  return (
    <Stack direction="column" spacing={4} data-testid="configure-organization-unit">
      <Typography variant="h1" gutterBottom>
        {t('users:createWizard.selectOrganizationUnit.title')}
      </Typography>
      <Typography variant="body1" color="text.secondary">
        {t('users:createWizard.selectOrganizationUnit.subtitle')}
      </Typography>

      <FormControl fullWidth required>
        <FormLabel>{t('users:createWizard.selectOrganizationUnit.fieldLabel')}</FormLabel>
        <OrganizationUnitTreePicker
          id="user-create-ou-picker"
          rootOuId={rootOuId}
          value={selectedOuId}
          onChange={onOuIdChange}
          maxHeight={500}
        />
      </FormControl>
    </Stack>
  );
}
