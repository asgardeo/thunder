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
import type {UserSchemaListResponse, SchemaListParams, ApiError} from '../types/users';

/**
 * Custom hook to fetch a list of user schemas
 * @param params - Optional query parameters for pagination
 * @returns Object containing data, loading state, error, and refetch function
 */
export default function useGetUserSchemas(params?: SchemaListParams) {
  const {http} = useAsgardeo();
  const {getServerUrl} = useConfig();
  const [data, setData] = useState<UserSchemaListResponse | null>(null);
  const [error, setError] = useState<ApiError | null>(null);
  const [loading, setLoading] = useState(false);
  const abortControllerRef = useRef<AbortController | null>(null);

  const API_BASE_URL: string = useMemo(
    () => getServerUrl() ?? (import.meta.env.VITE_ASGARDEO_BASE_URL as string),
    [getServerUrl],
  );

  useEffect(() => {
    // Cancel previous request
    abortControllerRef.current?.abort();
    abortControllerRef.current = new AbortController();

    const fetchUserSchemas = async (): Promise<void> => {
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

        const queryString = searchParams.toString();

        const response = await http.request({
          url: `${API_BASE_URL}/user-schemas${queryString ? `?${queryString}` : ''}`,
          method: 'GET',
          headers: {
            'Content-Type': 'application/json',
          },
          signal: abortControllerRef.current?.signal,
        } as unknown as Parameters<typeof http.request>[0]);

        const jsonData = response.data as UserSchemaListResponse;
        setData(jsonData);
        setError(null);
      } catch (err) {
        // Don't set error if request was aborted
        if (err instanceof Error && err.name === 'AbortError') {
          return;
        }

        const apiError: ApiError = {
          code: 'FETCH_ERROR',
          message: err instanceof Error ? err.message : 'An unknown error occurred',
          description: 'Failed to fetch user schemas',
        };
        setError(apiError);
      } finally {
        setLoading(false);
      }
    };

    fetchUserSchemas().catch(() => {
      // Error already handled
    });

    // Cleanup on unmount
    return () => {
      abortControllerRef.current?.abort();
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [params]);

  const refetch = async (newParams?: SchemaListParams): Promise<void> => {
    // Cancel previous request
    abortControllerRef.current?.abort();
    abortControllerRef.current = new AbortController();

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

      const queryString = searchParams.toString();

      const response = await http.request({
        url: `${API_BASE_URL}/user-schemas${queryString ? `?${queryString}` : ''}`,
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
        signal: abortControllerRef.current?.signal,
      } as unknown as Parameters<typeof http.request>[0]);

      const jsonData = response.data as UserSchemaListResponse;
      setData(jsonData);
      setError(null);
    } catch (err) {
      // Don't set error if request was aborted
      if (err instanceof Error && err.name === 'AbortError') {
        return;
      }

      const apiError: ApiError = {
        code: 'FETCH_ERROR',
        message: err instanceof Error ? err.message : 'An unknown error occurred',
        description: 'Failed to fetch user schemas',
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
