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

import {OrganizationUnitTreePicker, useHasMultipleOUs} from '@thunderid/configure-organization-units';
import {Typography, Stack, FormControl, FormLabel, Checkbox, FormControlLabel} from '@wso2/oxygen-ui';
import type {JSX} from 'react';
import {useEffect} from 'react';
import {useTranslation} from 'react-i18next';

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
  const {hasMultipleOUs, ouList} = useHasMultipleOUs();

  // Auto-select first organization unit when none is selected
  useEffect(() => {
    if (!ouId && ouList.length > 0) {
      onOuIdChange(ouList[0].id);
    }
  }, [ouList, ouId, onOuIdChange]);

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

      {hasMultipleOUs && (
        <FormControl fullWidth required>
          <FormLabel>{t('userTypes:organizationUnit')}</FormLabel>
          <OrganizationUnitTreePicker id="user-type-ou-picker" value={ouId} onChange={onOuIdChange} />
        </FormControl>
      )}

      <FormControlLabel
        control={
          <Checkbox checked={allowSelfRegistration} onChange={(e) => onAllowSelfRegistrationChange(e.target.checked)} />
        }
        label={t('userTypes:allowSelfRegistration')}
      />
    </Stack>
  );
}
