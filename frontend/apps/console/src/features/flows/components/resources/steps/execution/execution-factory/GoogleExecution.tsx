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

import {Box, Typography} from '@wso2/oxygen-ui';
import type {ReactElement} from 'react';
import {useTranslation} from 'react-i18next';
import type {ExecutionMinimalPropsInterface} from '../ExecutionMinimal';
import resolveStaticResourcePath from '@/features/flows/utils/resolveStaticResourcePath';

/**
 * Props interface of {@link GoogleExecution}.
 */
export type GoogleExecutionPropsInterface = ExecutionMinimalPropsInterface;

function GoogleExecution({resource}: GoogleExecutionPropsInterface): ReactElement {
  const {t} = useTranslation();

  // display.label contains the action/mode (e.g., "Google")
  const displayLabel = resource.display?.label;

  return (
    <Box display="flex" gap={1}>
      <img src={resolveStaticResourcePath('assets/images/icons/google.svg')} alt="google-icon" height="20" />
      <Typography variant="body1">{displayLabel ?? t('flows:core.executions.names.google')}</Typography>
    </Box>
  );
}

export default GoogleExecution;
