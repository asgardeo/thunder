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

import {Stack} from '@wso2/oxygen-ui';
import type {Application} from '../../../models/application';
import AuthenticationFlowSection from './AuthenticationFlowSection';
import RegistrationFlowSection from './RegistrationFlowSection';

/**
 * Props for the {@link EditFlowsSettings} component.
 */
interface EditFlowsSettingsProps {
  /**
   * The application being edited
   */
  application: Application;
  /**
   * Partial application object containing edited fields
   */
  editedApp: Partial<Application>;
  /**
   * Callback function to handle field value changes
   * @param field - The application field being updated
   * @param value - The new value for the field
   */
  onFieldChange: (field: keyof Application, value: unknown) => void;
}

/**
 * Container component for authentication and registration flow settings.
 *
 * Displays sections for:
 * - Authentication flow selection
 * - Registration flow selection (with enable/disable toggle)
 *
 * @param props - Component props
 * @returns Flow settings sections wrapped in a Stack
 */
export default function EditFlowsSettings({application, editedApp, onFieldChange}: EditFlowsSettingsProps) {
  return (
    <Stack spacing={3}>
      <AuthenticationFlowSection application={application} editedApp={editedApp} onFieldChange={onFieldChange} />
      <RegistrationFlowSection application={application} editedApp={editedApp} onFieldChange={onFieldChange} />
    </Stack>
  );
}
