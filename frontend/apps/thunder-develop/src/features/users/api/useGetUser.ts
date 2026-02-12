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

import {useState, useEffect, useMemo} from 'react';
import {useAsgardeo} from '@asgardeo/react';
import {useConfig} from '@thunder/shared-contexts';
import type {ApiUser, ApiError} from '../types/users';

/**
 * Custom hook to fetch a single user by ID
 * @param userId - The ID of the user to fetch
 * @returns Object containing data, loading state, error, and refetch function
 */
export default function useGetUser(userId: string | undefined) {
  const {http} = useAsgardeo();
  const {getServerUrl} = useConfig();
  const [data, setData] = useState<ApiUser | null>(null);
  const [error, setError] = useState<ApiError | null>(null);
  const [loading, setLoading] = useState(false);

  const API_BASE_URL: string = useMemo(
    () => getServerUrl() ?? (import.meta.env.VITE_ASGARDEO_BASE_URL as string),
    [getServerUrl],
  );

  useEffect(() => {
    if (!userId) {
      return;
    }

    const fetchUser = async (): Promise<void> => {
      try {
        setLoading(true);
        setError(null);

        const response = await http.request({
          url: `${API_BASE_URL}/users/${userId}`,
          method: 'GET',
          headers: {
            'Content-Type': 'application/json',
          },
        } as unknown as Parameters<typeof http.request>[0]);

        const jsonData = response.data as ApiUser;
        setData(jsonData);
        setError(null);
      } catch (err) {
        const apiError: ApiError = {
          code: 'FETCH_ERROR',
          message: err instanceof Error ? err.message : 'An unknown error occurred',
          description: 'Failed to fetch user',
        };
        setError(apiError);
      } finally {
        setLoading(false);
      }
    };

    fetchUser().catch(() => {
      // Error already handled
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [userId]);

  const refetch = async (): Promise<void> => {
    if (!userId) {
      return;
    }

    try {
      setLoading(true);
      setError(null);

      const response = await http.request({
        url: `${API_BASE_URL}/users/${userId}`,
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
      } as unknown as Parameters<typeof http.request>[0]);

      const jsonData = response.data as ApiUser;
      setData(jsonData);
      setError(null);
    } catch (err) {
      const apiError: ApiError = {
        code: 'FETCH_ERROR',
        message: err instanceof Error ? err.message : 'An unknown error occurred',
        description: 'Failed to fetch user',
      };
      setError(apiError);
      throw err;
    } finally {
      setLoading(false);
    }
  };

  return {
    data,
    loading,
    error,
    refetch,
  };
}
