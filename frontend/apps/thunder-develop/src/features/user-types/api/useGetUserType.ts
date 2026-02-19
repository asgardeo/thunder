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

import {useEffect, useRef, useState, useMemo} from 'react';
import {useAsgardeo} from '@asgardeo/react';
import {useConfig} from '@thunder/shared-contexts';
import {useLogger} from '@thunder/logger/react';
import type {ApiError, ApiUserType} from '../types/user-types';

/**
 * Custom hook to fetch a single user type by ID
 * Includes double-fetch prevention for React Strict Mode
 * @param id - The user type ID to fetch
 * @returns Object containing data, loading state, error, and refetch function
 */
export default function useGetUserType(id?: string) {
  const {http} = useAsgardeo();
  const {getServerUrl} = useConfig();
  const logger = useLogger();
  const [data, setData] = useState<ApiUserType | null>(null);
  const [error, setError] = useState<ApiError | null>(null);
  const [loading, setLoading] = useState(false);

  // Refs to prevent double-fetch in React Strict Mode
  const hasFetchedRef = useRef(false);
  const lastIdRef = useRef<string | undefined>(undefined);

  const API_BASE_URL: string = useMemo(
    () => getServerUrl() ?? (import.meta.env.VITE_ASGARDEO_BASE_URL as string),
    [getServerUrl],
  );

  useEffect(() => {
    if (!id) {
      setData(null);
      setError(null);
      return;
    }

    // Check if we've already fetched for this ID
    if (hasFetchedRef.current && lastIdRef.current === id) {
      return;
    }

    // Mark as fetched and store the ID
    hasFetchedRef.current = true;
    lastIdRef.current = id;

    const fetchUserType = async (): Promise<void> => {
      try {
        setLoading(true);
        setError(null);

        const response = await http.request({
          url: `${API_BASE_URL}/user-types/${id}`,
          method: 'GET',
        } as unknown as Parameters<typeof http.request>[0]);

        const jsonData = response.data as ApiUserType;
        setData(jsonData);
        setError(null);
      } catch (err) {
        const apiError: ApiError = {
          code: 'FETCH_USER_TYPE_ERROR',
          message: err instanceof Error ? err.message : 'An unknown error occurred',
          description: 'Failed to fetch user type',
        };
        setError(apiError);
        setData(null);
      } finally {
        setLoading(false);
      }
    };

    fetchUserType().catch((_error: unknown) => {
      logger.error('Failed to fetch user type', {error: _error, userTypeId: id});
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

    // Reset the hasFetched flag when explicitly refetching
    hasFetchedRef.current = false;
    lastIdRef.current = userTypeId;

    try {
      setLoading(true);
      setError(null);

      const response = await http.request({
        url: `${API_BASE_URL}/user-types/${userTypeId}`,
        method: 'GET',
      } as unknown as Parameters<typeof http.request>[0]);

      const jsonData = response.data as ApiUserType;
      setData(jsonData);
      setError(null);
    } catch (err) {
      const apiError: ApiError = {
        code: 'FETCH_USER_TYPE_ERROR',
        message: err instanceof Error ? err.message : 'An unknown error occurred',
        description: 'Failed to fetch user type',
      };
      setError(apiError);
      setData(null);
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
