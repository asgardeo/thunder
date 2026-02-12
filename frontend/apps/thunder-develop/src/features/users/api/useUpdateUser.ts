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
import type {ApiError, ApiUser} from '../types/users';

/**
 * Request body for updating a user
 * All fields are required as per PUT semantics
 */
export interface UpdateUserRequest {
  organizationUnit: string;
  type: string;
  groups?: string[];
  attributes: Record<string, unknown>;
}

/**
 * Custom hook to update an existing user
 * @returns Object containing updateUser function, data, loading state, error, and reset function
 */
export default function useUpdateUser() {
  const {http} = useAsgardeo();
  const {getServerUrl} = useConfig();
  const [data, setData] = useState<ApiUser | null>(null);
  const [error, setError] = useState<ApiError | null>(null);
  const [loading, setLoading] = useState(false);

  const API_BASE_URL: string = useMemo(
    () => getServerUrl() ?? (import.meta.env.VITE_ASGARDEO_BASE_URL as string),
    [getServerUrl],
  );

  const updateUser = async (userId: string, userData: UpdateUserRequest): Promise<void> => {
    try {
      setLoading(true);
      setError(null);
      setData(null);

      const response = await http.request({
        url: `${API_BASE_URL}/users/${userId}`,
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        data: userData,
      } as unknown as Parameters<typeof http.request>[0]);

      const jsonData = response.data as ApiUser;
      setData(jsonData);
      setError(null);
    } catch (err) {
      const apiError: ApiError = {
        code: 'UPDATE_USER_ERROR',
        message: err instanceof Error ? err.message : 'An unknown error occurred',
        description: 'Failed to update user',
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
    updateUser,
    data,
    loading,
    error,
    reset,
  };
}
