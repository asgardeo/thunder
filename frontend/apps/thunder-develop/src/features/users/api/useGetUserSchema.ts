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
import type {ApiUserSchema, ApiError} from '../types/users';

/**
 * Custom hook to fetch a single user schema by ID
 * @param id - The ID of the user schema to fetch
 * @returns Object containing data, loading state, error, and refetch function
 */
export default function useGetUserSchema(id?: string) {
  const {http} = useAsgardeo();
  const {getServerUrl} = useConfig();
  const [data, setData] = useState<ApiUserSchema | null>(null);
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

    const fetchUserSchema = async (): Promise<void> => {
      try {
        setLoading(true);
        setError(null);

        const response = await http.request({
          url: `${API_BASE_URL}/user-schemas/${id}`,
          method: 'GET',
          headers: {
            'Content-Type': 'application/json',
          },
        } as unknown as Parameters<typeof http.request>[0]);

        const jsonData = response.data as ApiUserSchema;
        setData(jsonData);
        setError(null);
      } catch (err) {
        const apiError: ApiError = {
          code: 'FETCH_ERROR',
          message: err instanceof Error ? err.message : 'An unknown error occurred',
          description: 'Failed to fetch user schema',
        };
        setError(apiError);
      } finally {
        setLoading(false);
      }
    };

    fetchUserSchema().catch(() => {
      // Error already handled
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id]);

  const refetch = async (newId?: string): Promise<void> => {
    const schemaId = newId ?? id;
    if (!schemaId) {
      setError({
        code: 'INVALID_ID',
        message: 'Invalid schema ID',
        description: 'Schema ID is required',
      });
      return;
    }

    try {
      setLoading(true);
      setError(null);

      const response = await http.request({
        url: `${API_BASE_URL}/user-schemas/${schemaId}`,
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
      } as unknown as Parameters<typeof http.request>[0]);

      const jsonData = response.data as ApiUserSchema;
      setData(jsonData);
      setError(null);
    } catch (err) {
      const apiError: ApiError = {
        code: 'FETCH_ERROR',
        message: err instanceof Error ? err.message : 'An unknown error occurred',
        description: 'Failed to fetch user schema',
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
