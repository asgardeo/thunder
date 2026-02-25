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

import {Typography, Stack, FormControl, FormLabel, Checkbox, FormControlLabel} from '@wso2/oxygen-ui';
import type {JSX} from 'react';
import {useEffect} from 'react';
import {useTranslation} from 'react-i18next';
import useGetOrganizationUnits from '../../../organization-units/api/useGetOrganizationUnits';
import OrganizationUnitTreePicker from '../../../organization-units/components/OrganizationUnitTreePicker';

/**
 * Props for the {@link ConfigureGeneral} component.
 *
 * @public
 */
export interface ConfigureGeneralProps {
  ouId: string;
  onOuIdChange: (ouId: string) => void;
  allowSelfRegistration: boolean;
  onAllowSelfRegistrationChange: (allow: boolean) => void;
  onReadyChange?: (isReady: boolean) => void;
}

/**
 * Step 2 of the user type creation wizard: configure organization unit and self-registration.
 *
 * @public
 */
export default function ConfigureGeneral({
  ouId,
  onOuIdChange,
  allowSelfRegistration,
  onAllowSelfRegistrationChange,
  onReadyChange = undefined,
}: ConfigureGeneralProps): JSX.Element {
  const {t} = useTranslation();
  const {data} = useGetOrganizationUnits();

  // Auto-select first organization unit
  useEffect(() => {
    if (!ouId && data?.organizationUnits && data.organizationUnits.length > 0) {
      onOuIdChange(data.organizationUnits[0].id);
    }
  }, [data, ouId, onOuIdChange]);

  // Broadcast readiness
  useEffect((): void => {
    if (onReadyChange) {
      onReadyChange(ouId.trim().length > 0);
    }
  }, [ouId, onReadyChange]);

  return (
    <Stack direction="column" spacing={4} data-testid="configure-general">
      <Stack direction="column" spacing={1}>
        <Typography variant="h1" gutterBottom>
          {t('userTypes:createWizard.general.title')}
        </Typography>
        <Typography variant="subtitle1" gutterBottom>
          {t('userTypes:createWizard.general.subtitle')}
        </Typography>
      </Stack>

      <FormControl fullWidth required>
        <FormLabel>{t('userTypes:organizationUnit')}</FormLabel>
        <OrganizationUnitTreePicker
          id="user-type-ou-picker"
          value={ouId}
          onChange={onOuIdChange}
        />
      </FormControl>

      <FormControlLabel
        control={
          <Checkbox
            checked={allowSelfRegistration}
            onChange={(e) => onAllowSelfRegistrationChange(e.target.checked)}
          />
        }
        label={t('userTypes:allowSelfRegistration')}
      />
    </Stack>
  );
}
