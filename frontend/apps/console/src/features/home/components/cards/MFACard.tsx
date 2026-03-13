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

import {KeyRound} from '@wso2/oxygen-ui-icons-react';
import {useTranslation} from 'react-i18next';
import type {JSX} from 'react';
import HomeNextStepCard from './HomeNextStepCard';

export default function MFACard(): JSX.Element {
  const {t} = useTranslation('home');

  return (
    <HomeNextStepCard
      icon={<KeyRound size={24} />}
      title={t('next_steps.mfa.title', 'Multi-factor Authentication')}
      description={t(
        'next_steps.mfa.description',
        'Protect users by enabling an additional verification factor to the sign-in process.',
      )}
      primaryLabel={t('next_steps.mfa.actions.primary.label', 'Configure Flows')}
      primaryRoute="/flows"
    />
  );
}
