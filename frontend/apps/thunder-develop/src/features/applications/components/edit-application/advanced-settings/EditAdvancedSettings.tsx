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
import type {OAuth2Config} from '../../../models/oauth';
import OAuth2ConfigSection from './OAuth2ConfigSection';
import CertificateSection from './CertificateSection';
import MetadataSection from './MetadataSection';

interface EditAdvancedSettingsProps {
  application: Application;
  editedApp: Partial<Application>;
  oauth2Config?: OAuth2Config;
  onFieldChange: (field: keyof Application, value: unknown) => void;
}

export default function EditAdvancedSettings({
  application,
  editedApp,
  oauth2Config,
  onFieldChange,
}: EditAdvancedSettingsProps) {
  return (
    <Stack spacing={3}>
      <OAuth2ConfigSection oauth2Config={oauth2Config} />
      <CertificateSection application={application} editedApp={editedApp} onFieldChange={onFieldChange} />
      <MetadataSection application={application} />
    </Stack>
  );
}
