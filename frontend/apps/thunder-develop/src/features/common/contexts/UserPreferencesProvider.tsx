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

import get from 'lodash-es/get';
import merge from 'lodash-es/merge';
import {type ReactElement, useCallback, useEffect, useMemo, useState} from 'react';
import UserPreferencesContext from './UserPreferencesContext';

/**
 * Props interface of {@link UserPreferencesProvider}
 */
export interface UserPreferencesProviderProps {
  /**
   * Storage strategy. Default is "localstorage".
   */
  storageStrategy?: 'localstorage' | 'sessionstorage';
  /**
   * UUI of the user.
   */
  userId: string;
}

const USER_PREFERENCES_STORAGE_KEY = 'user-preferences';

/**
 * Props interface for UserPreferencesProvider with children
 */
interface UserPreferencesProviderPropsWithChildren extends UserPreferencesProviderProps {
  children: ReactElement;
}

/**
 * Provider for the Application local settings.
 * This is a generic provider which can be used to store any type of settings.
 * The type of the preferences should be passed as a generic type.
 *
 * @example
 * `<UserPreferencesProvider<Preference> initialPreferences={ { "orgId": { "key": "value" } } }>`
 *
 * @param props - Props for the client.
 * @returns App settings provider.
 */
function UserPreferencesProvider({
  children,
  storageStrategy = 'localstorage',
  userId: _userId,
}: UserPreferencesProviderPropsWithChildren): ReactElement {
  const [preferencesInContext, setPreferencesInContext] = useState<Record<string, unknown> | null>(null);

  /**
   * Set the initial preferences.
   * If the preferences are already set in storage, they will be overridden.
   */
  useEffect(() => {
    const storedPreferences: string | null = (() => {
      switch (storageStrategy) {
        case 'localstorage':
          return localStorage.getItem(USER_PREFERENCES_STORAGE_KEY);
        case 'sessionstorage':
          return sessionStorage.getItem(USER_PREFERENCES_STORAGE_KEY);
        default:
          return null;
      }
    })();

    if (storedPreferences) {
      const preferences = JSON.parse(storedPreferences) as Record<string, unknown>;

      setPreferencesInContext(preferences);
    }
  }, [storageStrategy]);

  /**
   * Set the preferences in storage.
   *
   * @example
   * `setPreferences({ "key": "value" }, "orgId")`
   *
   * @param preferencesToUpdate - The new preferences to update.
   * @param userId - Optional user Id. If provided, the preferences for the passed in user-id will be updated.
   */
  const setPreferences = useCallback(
    (preferencesToUpdate: Record<string, unknown>, userId?: string): void => {
      const updatedPreferences = (merge as (
        ...sources: (Record<string, unknown> | null)[]
      ) => Record<string, unknown>)({}, preferencesInContext, {
        [userId ?? _userId]: {
          ...preferencesToUpdate,
        },
      });

      setPreferencesInContext(updatedPreferences);

      switch (storageStrategy) {
        case 'localstorage':
          localStorage.setItem(USER_PREFERENCES_STORAGE_KEY, JSON.stringify(updatedPreferences));

          break;
        case 'sessionstorage':
          sessionStorage.setItem(USER_PREFERENCES_STORAGE_KEY, JSON.stringify(updatedPreferences));

          break;
        default:
          break;
      }
    },
    [preferencesInContext, _userId, storageStrategy],
  );

  /**
   * Get the preferences from storage.
   *
   * @example
   * `getPreferences("key.nested", "orgId")`
   *
   * @param key - The key of the preference to retrieve.
   * @param userId - Optional user Id. If provided, the preferences for the passed in user-id will be updated.
   */
  const getPreferences = useCallback(
    (key: string, userId?: string): unknown => {
      const userKey: string = userId ?? _userId;
      const userPreferences = get(preferencesInContext, userKey, {}) as Record<string, unknown>;

      return userPreferences[key] ?? null;
    },
    [preferencesInContext, _userId],
  );

  /**
   * Get all flat-level preferences for the specified organization.
   *
   * @example
   * `getFlatPreferences("orgId")`
   *
   * @param userId - Optional user Id. If provided, the preferences for the passed in user-id will be updated.
   */
  const getFlatPreferences = useCallback(
    (userId?: string): Record<string, unknown> => get(preferencesInContext, userId ?? _userId, {}) as Record<string, unknown>,
    [preferencesInContext, _userId],
  );

  const contextValue = useMemo(
    () => ({
      getPreferences,
      setPreferences,
      ...getFlatPreferences(),
    }),
    [getPreferences, setPreferences, getFlatPreferences],
  );

  return <UserPreferencesContext.Provider value={contextValue}>{children}</UserPreferencesContext.Provider>;
}

export default UserPreferencesProvider;
