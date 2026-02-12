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

import {useState, useMemo} from 'react';
import {useAsgardeo} from '@asgardeo/react';
import {useConfig} from '@thunder/shared-contexts';

import type {ApiError, ApiUserSchema, UpdateUserSchemaRequest} from '../types/user-types';

/**
 * Custom hook to update an existing user schema (user type)
 * @returns Object containing updateUserType function, data, loading state, error, and reset function
 */
export default function useUpdateUserType() {
  const {http} = useAsgardeo();
  const {getServerUrl} = useConfig();
  const [data, setData] = useState<ApiUserSchema | null>(null);
  const [error, setError] = useState<ApiError | null>(null);
  const [loading, setLoading] = useState(false);

  const API_BASE_URL: string = useMemo(
    () => getServerUrl() ?? (import.meta.env.VITE_ASGARDEO_BASE_URL as string),
    [getServerUrl],
  );

  const updateUserType = async (userTypeId: string, requestData: UpdateUserSchemaRequest): Promise<void> => {
    try {
      setLoading(true);
      setError(null);
      setData(null);

      const response = await http.request({
        url: `${API_BASE_URL}/user-schemas/${userTypeId}`,
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        data: requestData,
      } as unknown as Parameters<typeof http.request>[0]);

      const jsonData = response.data as ApiUserSchema;
      setData(jsonData);
      setError(null);
    } catch (err) {
      const apiError: ApiError = {
        code: 'UPDATE_USER_TYPE_ERROR',
        message: err instanceof Error ? err.message : 'An unknown error occurred',
        description: 'Failed to update user type',
      };
      setError(apiError);
      throw err;
    } finally {
      setLoading(false);
    }
  };

  const reset = () => {
    setData(null);
    setError(null);
  };

  return {
    updateUserType,
    data,
    loading,
    error,
    reset,
  };
}
