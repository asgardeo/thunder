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

import {useMemo, type ReactElement} from 'react';
import {Trans, useTranslation} from 'react-i18next';
import type {RequiredFieldInterface} from '@/features/flows/hooks/useRequiredFields';
import useRequiredFields from '@/features/flows/hooks/useRequiredFields';
import {Box, Typography} from '@wso2/oxygen-ui';
import resolveStaticResourcePath from '@/features/flows/utils/resolveStaticResourcePath';
import type {ExecutionMinimalPropsInterface} from '../ExecutionMinimal';

/**
 * Props interface of {@link GoogleExecution}.
 */
export type GoogleExecutionPropsInterface = ExecutionMinimalPropsInterface;

function GoogleExecution({resource}: GoogleExecutionPropsInterface): ReactElement {
  const {t} = useTranslation();

  const generalMessage: ReactElement = useMemo(
    () => (
      <Trans i18nKey="flows:core.validation.fields.executor.general" values={{id: resource?.id}} components={[<code />]}>
        {'The executor <0>{{id}}</0> is not properly configured.'}
      </Trans>
    ),
    [resource?.id],
  );

  const fields: RequiredFieldInterface[] = useMemo(
    () => [
      {
        errorMessage: t('flows:core.validation.fields.input.idpId'),
        name: 'data.properties.idpId',
      },
    ],
    [t],
  );

  useRequiredFields(resource, generalMessage, fields);

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
