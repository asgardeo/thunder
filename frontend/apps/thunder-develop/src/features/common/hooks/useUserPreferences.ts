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

import {useContext} from 'react';
import UserPreferenceContext, {type UserPreferencesContextProps} from '../contexts/UserPreferencesContext';

/**
 * Hook that provides access to the user's preference context.
 * @returns An object containing the context values of {@link UserPreferenceContext}.
 */
const useUserPreferences = <T = Record<string, unknown>>(): UserPreferencesContextProps<T> => {
  const context = useContext(UserPreferenceContext);

  if (context === null || context === undefined) {
    throw new Error('useUserPreferences must be used within a UserPreferencesProvider');
  }

  return context as UserPreferencesContextProps<T>;
};

export default useUserPreferences;
