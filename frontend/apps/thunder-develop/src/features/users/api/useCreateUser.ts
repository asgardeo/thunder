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
import type {ApiError} from '../types/users';

/**
 * Request body for creating a new user
 */
export interface CreateUserRequest {
  organizationUnit: string;
  type: string;
  groups?: string[];
  attributes: Record<string, unknown>;
}

/**
 * Response after creating a user
 */
export interface CreateUserResponse {
  id: string;
  organizationUnit: string;
  type: string;
  attributes: Record<string, unknown>;
}

/**
 * Custom hook to create a new user
 * @returns Object containing createUser function, data, loading state, error, and reset function
 */
export default function useCreateUser() {
  const {http} = useAsgardeo();
  const {getServerUrl} = useConfig();
  const [data, setData] = useState<CreateUserResponse | null>(null);
  const [error, setError] = useState<ApiError | null>(null);
  const [loading, setLoading] = useState(false);

  const API_BASE_URL: string = useMemo(
    () => getServerUrl() ?? (import.meta.env.VITE_ASGARDEO_BASE_URL as string),
    [getServerUrl],
  );

  const createUser = async (userData: CreateUserRequest): Promise<void> => {
    try {
      setLoading(true);
      setError(null);
      setData(null);

      const response = await http.request({
        url: `${API_BASE_URL}/users`,
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        data: userData,
      } as unknown as Parameters<typeof http.request>[0]);

      const jsonData = response.data as CreateUserResponse;
      setData(jsonData);
      setError(null);
    } catch (err) {
      const apiError: ApiError = {
        code: 'CREATE_USER_ERROR',
        message: err instanceof Error ? err.message : 'An unknown error occurred',
        description: 'Failed to create user',
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
    createUser,
    data,
    loading,
    error,
    reset,
  };
}
