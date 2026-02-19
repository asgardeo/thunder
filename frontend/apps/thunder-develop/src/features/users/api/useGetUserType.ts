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

import {useState, useEffect, useRef, useMemo} from 'react';
import {useAsgardeo} from '@asgardeo/react';
import {useConfig} from '@thunder/shared-contexts';
import type {ApiUserType, ApiError} from '../types/users';

/**
 * Custom hook to fetch a single user type by ID
 * @param id - The ID of the user type to fetch
 * @returns Object containing data, loading state, error, and refetch function
 */
export default function useGetUserType(id?: string) {
  const {http} = useAsgardeo();
  const {getServerUrl} = useConfig();
  const [data, setData] = useState<ApiUserType | null>(null);
  const [error, setError] = useState<ApiError | null>(null);
  const [loading, setLoading] = useState(false);
  const hasFetchedRef = useRef(false);
  const lastIdRef = useRef<string | undefined>(undefined);

  const API_BASE_URL: string = useMemo(
    () => getServerUrl() ?? (import.meta.env.VITE_ASGARDEO_BASE_URL as string),
    [getServerUrl],
  );

  useEffect(() => {
    if (!id) {
      return;
    }

    // Prevent double fetch in React Strict Mode and check if ID changed
    if (hasFetchedRef.current && lastIdRef.current === id) {
      return;
    }
    hasFetchedRef.current = true;
    lastIdRef.current = id;

    const fetchUserType = async (): Promise<void> => {
      try {
        setLoading(true);
        setError(null);

        const response = await http.request({
          url: `${API_BASE_URL}/user-types/${id}`,
          method: 'GET',
          headers: {
            'Content-Type': 'application/json',
          },
        } as unknown as Parameters<typeof http.request>[0]);

        const jsonData = response.data as ApiUserType;
        setData(jsonData);
        setError(null);
      } catch (err) {
        const apiError: ApiError = {
          code: 'FETCH_ERROR',
          message: err instanceof Error ? err.message : 'An unknown error occurred',
          description: 'Failed to fetch user type',
        };
        setError(apiError);
      } finally {
        setLoading(false);
      }
    };

    fetchUserType().catch(() => {
      // Error already handled
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id]);

  const refetch = async (newId?: string): Promise<void> => {
    const userTypeId = newId ?? id;
    if (!userTypeId) {
      setError({
        code: 'INVALID_ID',
        message: 'Invalid user type ID',
        description: 'User type ID is required',
      });
      return;
    }

    try {
      setLoading(true);
      setError(null);

      const response = await http.request({
        url: `${API_BASE_URL}/user-types/${userTypeId}`,
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
      } as unknown as Parameters<typeof http.request>[0]);

      const jsonData = response.data as ApiUserType;
      setData(jsonData);
      setError(null);
    } catch (err) {
      const apiError: ApiError = {
        code: 'FETCH_ERROR',
        message: err instanceof Error ? err.message : 'An unknown error occurred',
        description: 'Failed to fetch user type',
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
