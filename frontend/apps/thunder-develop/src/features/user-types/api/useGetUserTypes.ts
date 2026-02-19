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
import type {ApiError, UserTypeListParams, UserTypeListResponse} from '../types/user-types';

/**
 * Custom hook to fetch a paginated list of user types
 * @param params - Optional pagination parameters (limit, offset)
 * @returns Object containing data, loading state, error, and refetch function
 */
export default function useGetUserTypes(params?: UserTypeListParams) {
  const {http} = useAsgardeo();
  const {getServerUrl} = useConfig();
  const logger = useLogger();
  const [data, setData] = useState<UserTypeListResponse | null>(null);
  const [error, setError] = useState<ApiError | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(false);
  const abortControllerRef = useRef<AbortController | null>(null);

  const API_BASE_URL: string = useMemo(
    () => getServerUrl() ?? (import.meta.env.VITE_ASGARDEO_BASE_URL as string),
    [getServerUrl],
  );

  useEffect(() => {
    // Cancel previous request if it exists
    abortControllerRef.current?.abort();
    abortControllerRef.current = new AbortController();

    const fetchUserTypes = async (): Promise<void> => {
      setIsLoading(true);
      setError(null);

      try {
        // Build query parameters
        const queryParams = new URLSearchParams();
        if (params?.limit !== undefined) {
          queryParams.append('limit', params.limit.toString());
        }
        if (params?.offset !== undefined) {
          queryParams.append('offset', params.offset.toString());
        }

        const queryString = queryParams.toString();
        const url = `${API_BASE_URL}/user-types${queryString ? `?${queryString}` : ''}`;

        const response = await http.request({
          url,
          method: 'GET',
          signal: abortControllerRef.current?.signal,
        } as unknown as Parameters<typeof http.request>[0]);

        const jsonData = response.data as UserTypeListResponse;
        setData(jsonData);
        setError(null);
      } catch (err) {
        // Don't set error if request was aborted
        if (err instanceof Error && err.name === 'AbortError') {
          return;
        }

        const apiError: ApiError = {
          code: 'FETCH_USER_TYPES_ERROR',
          message: err instanceof Error ? err.message : 'An unknown error occurred',
          description: 'Failed to fetch user types',
        };
        setError(apiError);
        setData(null);
      } finally {
        setIsLoading(false);
      }
    };

    fetchUserTypes().catch((_error: unknown) => {
      logger.error('Failed to fetch user types', {error: _error});
    });

    // Cleanup: abort request on unmount
    return () => {
      abortControllerRef.current?.abort();
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [params]);

  const refetch = async (newParams?: UserTypeListParams): Promise<void> => {
    // Cancel previous request
    abortControllerRef.current?.abort();
    abortControllerRef.current = new AbortController();

    setIsLoading(true);
    setError(null);

    try {
      // Build query parameters
      const queryParams = new URLSearchParams();
      const finalParams = newParams ?? params;

      if (finalParams?.limit !== undefined) {
        queryParams.append('limit', finalParams.limit.toString());
      }
      if (finalParams?.offset !== undefined) {
        queryParams.append('offset', finalParams.offset.toString());
      }

      const queryString = queryParams.toString();
      const url = `${API_BASE_URL}/user-types${queryString ? `?${queryString}` : ''}`;

      const response = await http.request({
        url,
        method: 'GET',
        signal: abortControllerRef.current.signal,
      } as unknown as Parameters<typeof http.request>[0]);

      const jsonData = response.data as UserTypeListResponse;
      setData(jsonData);
      setError(null);
    } catch (err) {
      // Don't set error if request was aborted
      if (err instanceof Error && err.name === 'AbortError') {
        return;
      }

      const apiError: ApiError = {
        code: 'FETCH_USER_TYPES_ERROR',
        message: err instanceof Error ? err.message : 'An unknown error occurred',
        description: 'Failed to fetch user types',
      };
      setError(apiError);
      setData(null);
      throw err;
    } finally {
      setIsLoading(false);
    }
  };

  return {
    data,
    loading: isLoading,
    error,
    refetch,
  };
}
