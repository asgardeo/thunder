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

import type {ReactElement} from 'react';
import {ExecutionTypes} from '@/features/flows/models/steps';
import {Box, Typography} from '@wso2/oxygen-ui';
import {useTranslation} from 'react-i18next';
import type {ExecutionMinimalPropsInterface} from '../ExecutionMinimal';
import GoogleExecution from './GoogleExecution';
import GithubExecution from './GithubExecution';

/**
 * Props interface of {@link CommonStepFactory}
 */
export type ExecutionFactoryPropsInterface = ExecutionMinimalPropsInterface;

/**
 * Factory for creating execution types.
 *
 * @param props - Props injected to the component.
 * @returns The ExecutionFactory component.
 */
function ExecutionFactory({resource}: ExecutionFactoryPropsInterface): ReactElement {
  const {t} = useTranslation();

  const action = resource.data?.action;
  const executorName = action?.executor?.name;

  if (executorName === ExecutionTypes.GoogleFederation) {
    return <GoogleExecution resource={resource} />;
  }

  if (executorName === ExecutionTypes.GithubFederation) {
    return <GithubExecution resource={resource} />;
  }

  if (executorName === ExecutionTypes.PasskeyEnrollment) {
    return (
      <Box display="flex" gap={1}>
        <img src="https://www.svgrepo.com/show/246819/fingerprint.svg" alt="fingerprint-icon" height="20" />
        <Typography variant="body1">{t('flows:core.executions.names.passkeyEnrollment')}</Typography>
      </Box>
    );
  }

  if (executorName === ExecutionTypes.MagicLinkExecutor) {
    return (
      <Box display="flex" gap={1} className="flow-builder-execution magic-link">
        <img src="https://www.svgrepo.com/show/524687/link.svg" alt="link-icon" height="20" />
        <Typography variant="body1">{t('flows:core.executions.names.magicLink')}</Typography>
      </Box>
    );
  }

  if (executorName === ExecutionTypes.ConfirmationCode) {
    return (
      <Box display="flex" gap={1} className="flow-builder-execution confirmation-code">
        <img src="https://www.svgrepo.com/show/468264/check-mark-square-2.svg" alt="check-mark-icon" height="20" />
        <Typography variant="body1">{t('flows:core.executions.names.confirmationCode')}</Typography>
      </Box>
    );
  }

  if (executorName === ExecutionTypes.SendEmailOTP) {
    return (
      <Box display="flex" gap={1} className="flow-builder-execution send-email-otp">
        <Typography variant="body1">{t('flows:core.executions.names.sendEmailOTP')}</Typography>
      </Box>
    );
  }

  if (executorName === ExecutionTypes.VerifyEmailOTP) {
    return (
      <Box display="flex" gap={1} className="flow-builder-execution verify-email-otp">
        <Typography variant="body1">{t('flows:core.executions.names.verifyEmailOTP')}</Typography>
      </Box>
    );
  }

  if (executorName === ExecutionTypes.SendSMS) {
    return (
      <Box display="flex" gap={1} className="flow-builder-execution send-sms">
        <Typography variant="body1">{t('flows:core.executions.names.sendSMS')}</Typography>
      </Box>
    );
  }

  if (executorName === ExecutionTypes.VerifySMSOTP) {
    return (
      <Box display="flex" gap={1} className="flow-builder-execution verify-sms-otp">
        <Typography variant="body1">{t('flows:core.executions.names.verifySMSOTP')}</Typography>
      </Box>
    );
  }

  return (
    <Box display="flex" gap={1}>
      <Typography variant="body1">{t('flows:core.executions.names.default')}</Typography>
    </Box>
  );
}

export default ExecutionFactory;
