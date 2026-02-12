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
import {Typography, Paper, Button} from '@wso2/oxygen-ui';
import {useTranslation} from 'react-i18next';

/**
 * Props for the {@link EditAdvancedSettings} component.
 */
interface EditAdvancedSettingsProps {
  /**
   * Callback function to open the delete confirmation dialog
   */
  onDeleteClick: () => void;
}

/**
 * Advanced Settings tab content for the Organization Unit edit page.
 *
 * Displays the danger zone section with a delete button for
 * permanently removing the organization unit.
 *
 * @param props - Component props
 * @returns Advanced settings tab content
 */
export default function EditAdvancedSettings({onDeleteClick}: EditAdvancedSettingsProps): JSX.Element {
  const {t} = useTranslation();

  return (
    <Paper sx={{p: 3, mb: 3}}>
      <Typography variant="h6" gutterBottom color="error">
        {t('organizationUnits:view.advanced.dangerZone')}
      </Typography>
      <Typography variant="body2" color="text.secondary" sx={{mb: 3}}>
        {t('organizationUnits:view.advanced.dangerZoneDescription')}
      </Typography>

      <Button variant="outlined" color="error" onClick={onDeleteClick}>
        {t('organizationUnits:view.advanced.deleteButton')}
      </Button>
    </Paper>
  );
}
