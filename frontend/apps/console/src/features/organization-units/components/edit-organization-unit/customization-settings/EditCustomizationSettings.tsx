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

import type {JSX} from 'react';
import {Stack} from '@wso2/oxygen-ui';
import type {OrganizationUnit} from '../../../models/organization-unit';
import AppearanceSection from './AppearanceSection';

/**
 * Props for the {@link EditCustomizationSettings} component.
 */
interface EditCustomizationSettingsProps {
  /**
   * The organization unit being edited
   */
  organizationUnit: OrganizationUnit;
  /**
   * Partial organization unit object containing edited fields
   */
  editedOU: Partial<OrganizationUnit>;
  /**
   * Callback function to handle field value changes
   * @param field - The organization unit field being updated
   * @param value - The new value for the field
   */
  onFieldChange: (field: keyof OrganizationUnit, value: unknown) => void;
}

/**
 * Customization tab content for the Organization Unit edit page.
 *
 * Displays sections for:
 * - Appearance (theme selection)
 *
 * @param props - Component props
 * @returns Customization settings sections wrapped in a Stack
 */
export default function EditCustomizationSettings({
  organizationUnit,
  editedOU,
  onFieldChange,
}: EditCustomizationSettingsProps): JSX.Element {
  return (
    <Stack spacing={3}>
      <AppearanceSection organizationUnit={organizationUnit} editedOU={editedOU} onFieldChange={onFieldChange} />
    </Stack>
  );
}
