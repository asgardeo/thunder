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

/**
 * Props for the {@link ConfigureOrganizationUnit} component.
 *
 * @public
 */
export interface ConfigureOrganizationUnitProps {
  selectedOuId: string;
  onOuIdChange: (ouId: string) => void;
  onReadyChange?: (isReady: boolean) => void;
}

/**
 * Step 2 of the group creation wizard: select the organization unit.
 * This step is only shown when there is more than one OU available.
 *
 * @public
 */
export default function ConfigureOrganizationUnit({
  selectedOuId,
  onOuIdChange,
  onReadyChange = undefined,
}: ConfigureOrganizationUnitProps): JSX.Element {
  const {t} = useTranslation();

  useEffect((): void => {
    if (onReadyChange) {
      onReadyChange(selectedOuId.length > 0);
    }
  }, [selectedOuId, onReadyChange]);

  return (
    <Stack direction="column" spacing={4} data-testid="configure-organization-unit">
      <Typography variant="h1" gutterBottom>
        {t('groups:createWizard.organizationUnit.title')}
      </Typography>
      <Typography variant="body1" color="text.secondary">
        {t('groups:createWizard.organizationUnit.subtitle')}
      </Typography>

      <FormControl fullWidth required>
        <FormLabel>{t('groups:create.form.organizationUnit.label')}</FormLabel>
        <OrganizationUnitTreePicker
          id="group-create-ou-picker"
          value={selectedOuId}
          onChange={onOuIdChange}
          maxHeight={500}
        />
      </FormControl>
    </Stack>
  );
}
