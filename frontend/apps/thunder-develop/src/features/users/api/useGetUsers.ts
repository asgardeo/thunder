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
import type {ApiError, UserListParams, UserListResponse} from '../types/users';

/**
 * Custom hook to fetch a list of users
 * @param params - Optional query parameters for filtering and pagination
 * @returns Object containing data, loading state, error, and refetch function
 */
export default function useGetUsers(params?: UserListParams) {
  const {http} = useAsgardeo();
  const {getServerUrl} = useConfig();
  const [data, setData] = useState<UserListResponse | null>(null);
  const [error, setError] = useState<ApiError | null>(null);
  const [loading, setLoading] = useState(false);

  const API_BASE_URL: string = useMemo(
    () => getServerUrl() ?? (import.meta.env.VITE_ASGARDEO_BASE_URL as string),
    [getServerUrl],
  );

  useEffect(() => {
    const fetchUsers = async (): Promise<void> => {
      try {
        setLoading(true);
        setError(null);

        const searchParams = new URLSearchParams();

        if (params?.limit !== undefined) {
          searchParams.append('limit', String(params.limit));
        }
        if (params?.offset !== undefined) {
          searchParams.append('offset', String(params.offset));
        }
        if (params?.filter) {
          searchParams.append('filter', params.filter);
        }

        const queryString = searchParams.toString();

        const response = await http.request({
          url: `${API_BASE_URL}/users${queryString ? `?${queryString}` : ''}`,
          method: 'GET',
          headers: {
            Accept: 'application/json',
            'Content-Type': 'application/json',
          },
        } as unknown as Parameters<typeof http.request>[0]);

        const jsonData = response.data as UserListResponse;
        setData(jsonData);
        setError(null);
      } catch (err) {
        const apiError: ApiError = {
          code: 'FETCH_ERROR',
          message: err instanceof Error ? err.message : 'An unknown error occurred',
          description: 'Failed to fetch users',
        };
        setError(apiError);
      } finally {
        setLoading(false);
      }
    };

    fetchUsers().catch(() => {
      // Error already handled
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [params]);

  const refetch = async (newParams?: UserListParams): Promise<void> => {
    try {
      setLoading(true);
      setError(null);

      const searchParams = new URLSearchParams();
      const finalParams = newParams ?? params;

      if (finalParams?.limit !== undefined) {
        searchParams.append('limit', String(finalParams.limit));
      }
      if (finalParams?.offset !== undefined) {
        searchParams.append('offset', String(finalParams.offset));
      }
      if (finalParams?.filter) {
        searchParams.append('filter', finalParams.filter);
      }

      const queryString = searchParams.toString();

      const response = await http.request({
        url: `${API_BASE_URL}/users${queryString ? `?${queryString}` : ''}`,
        method: 'GET',
        headers: {
          Accept: 'application/json',
          'Content-Type': 'application/json',
        },
      } as unknown as Parameters<typeof http.request>[0]);

      const jsonData = response.data as UserListResponse;
      setData(jsonData);
      setError(null);
    } catch (err) {
      const apiError: ApiError = {
        code: 'FETCH_ERROR',
        message: err instanceof Error ? err.message : 'An unknown error occurred',
        description: 'Failed to fetch users',
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
