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

import {useState, useCallback} from 'react';
import {useAsgardeo} from '@asgardeo/react';
import {useConfig} from '@thunder/shared-contexts';
import type {ApiError} from '../types/user-types';

/**
 * Return type for the useDeleteUserType hook.
 */
export interface UseDeleteUserTypeReturn {
  error: ApiError | null;
  loading: boolean;
  deleteUserType: (userTypeId: string) => Promise<boolean>;
  reset: () => void;
}

/**
 * Custom React hook to delete a user schema (user type) from the Thunder server.
 *
 * @returns Hook state and actions for deleting user types
 */
export default function useDeleteUserType(): UseDeleteUserTypeReturn {
  const {http} = useAsgardeo();
  const {getServerUrl} = useConfig();

  const [error, setError] = useState<ApiError | null>(null);
  const [loading, setLoading] = useState(false);

  const deleteUserType = useCallback(
    async (userTypeId: string): Promise<boolean> => {
      setLoading(true);
      setError(null);

      try {
        const serverUrl: string = getServerUrl() ?? '';
        await http.request({
          url: `${serverUrl}/user-schemas/${userTypeId}`,
          method: 'DELETE',
          headers: {
            'Content-Type': 'application/json',
          },
        } as unknown as Parameters<typeof http.request>[0]);

        setLoading(false);
        return true;
      } catch (err: unknown) {
        setLoading(false);
        const apiError: ApiError = {
          code: 'DELETE_USER_TYPE_ERROR',
          message: err instanceof Error ? err.message : 'An unknown error occurred',
          description: 'Failed to delete user type',
        };
        setError(apiError);
        throw err;
      }
    },
    [getServerUrl, http],
  );

  const reset = useCallback(() => {
    setError(null);
    setLoading(false);
  }, []);

  return {error, loading, deleteUserType, reset};
}
