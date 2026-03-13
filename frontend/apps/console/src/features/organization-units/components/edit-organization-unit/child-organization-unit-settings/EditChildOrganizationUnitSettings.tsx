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
import {Stack} from '@wso2/oxygen-ui';
import ManageChildOUsSection from './ManageChildOrganizationUnitSection';

/**
 * Props for the {@link EditChildOrganizationUnitSettings} component.
 */
interface EditChildOrganizationUnitSettingsProps {
  /**
   * The ID of the parent organization unit
   */
  organizationUnitId: string;
  /**
   * The name of the parent organization unit (for back navigation)
   */
  organizationUnitName: string;
}

/**
 * Child Organization Units tab content for the Organization Unit edit page.
 *
 * Displays sections for:
 * - Managing child organization units (DataGrid with navigation)
 *
 * @param props - Component props
 * @returns Child OUs tab content
 */
export default function EditChildOrganizationUnitSettings({
  organizationUnitId,
  organizationUnitName,
}: EditChildOrganizationUnitSettingsProps): JSX.Element {
  return (
    <Stack spacing={3}>
      <ManageChildOUsSection organizationUnitId={organizationUnitId} organizationUnitName={organizationUnitName} />
    </Stack>
  );
}
