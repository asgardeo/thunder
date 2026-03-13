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

import type {JSX} from 'react';
import {Typography, Button} from '@wso2/oxygen-ui';
import {useTranslation} from 'react-i18next';
import SettingsCard from '@/components/SettingsCard';

/**
 * Props for the {@link DangerZoneSection} component.
 */
interface DangerZoneSectionProps {
  /**
   * Callback function to open the delete confirmation dialog
   */
  onDeleteClick: () => void;
}

/**
 * Section component displaying the danger zone with destructive actions.
 *
 * Displays a delete button for permanently removing the organization unit.
 *
 * @param props - Component props
 * @returns Danger zone UI within a Paper
 */
export default function DangerZoneSection({onDeleteClick}: DangerZoneSectionProps): JSX.Element {
  const {t} = useTranslation();

  return (
    <SettingsCard
      title={t('organizationUnits:edit.general.sections.dangerZone.title')}
      description={t('organizationUnits:edit.general.sections.dangerZone.description')}
    >
      <Typography variant="h6" gutterBottom color="error">
        {t('organizationUnits:edit.general.sections.dangerZone.deleteOU.title')}
      </Typography>
      <Typography variant="body2" color="text.secondary" sx={{mb: 3}}>
        {t('organizationUnits:edit.general.sections.dangerZone.deleteOU.description')}
      </Typography>
      <Button variant="contained" color="error" onClick={onDeleteClick}>
        {t('organizationUnits:edit.general.dangerZone.delete.button.label')}
      </Button>
    </SettingsCard>
  );
}
