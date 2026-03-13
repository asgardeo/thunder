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
import ManageGroupsSection from './ManageGroupsSection';

/**
 * Props for the {@link EditGroupSettings} component.
 */
interface EditGroupSettingsProps {
  /**
   * The ID of the organization unit
   */
  organizationUnitId: string;
}

/**
 * Groups tab content for the Organization Unit edit page.
 *
 * Displays sections for:
 * - Managing groups belonging to the organization unit (DataGrid)
 *
 * @param props - Component props
 * @returns Groups tab content
 */
export default function EditGroupSettings({organizationUnitId}: EditGroupSettingsProps): JSX.Element {
  return (
    <Stack spacing={3}>
      <ManageGroupsSection organizationUnitId={organizationUnitId} />
    </Stack>
  );
}
