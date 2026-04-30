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

import {useLogger} from '@thunder/logger/react';
import {useCallback, useState, type JSX} from 'react';
import UserCreatePage from './UserCreatePage';
import UserInvitePage from './UserInvitePage';
import UserCreateProvider from '../contexts/UserCreate/UserCreateProvider';

export default function UserAddPage(): JSX.Element {
  const logger = useLogger('UserAddPage');
  const [useManualCreateFallback, setUseManualCreateFallback] = useState(false);

  const handleFlowUnavailable = useCallback(
    (error: unknown) => {
      logger.warn('User onboarding flow unavailable. Falling back to manual create user flow.', {error});
      setUseManualCreateFallback(true);
    },
    [logger],
  );

  if (useManualCreateFallback) {
    return (
      <UserCreateProvider>
        <UserCreatePage />
      </UserCreateProvider>
    );
  }

  return <UserInvitePage onFlowUnavailable={handleFlowUnavailable} />;
}
