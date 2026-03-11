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

import {useState, useEffect, useCallback, useRef} from 'react';
import {useAsgardeo} from '@asgardeo/react';
import {useConfig} from '@thunder/shared-contexts';
import type {ApiUserSchema, ApiError} from '../types/user-types';

/**
 * Return type for the useGetUserType hook.
 */
export interface UseGetUserTypeReturn {
  data: ApiUserSchema | null;
  error: ApiError | null;
  loading: boolean;
  refetch: (newId?: string) => Promise<void>;
}

/**
 * Custom React hook to fetch a single user schema (user type) by ID from the Thunder server.
 *
 * @param id - The unique identifier of the user type to fetch
 * @returns Hook state and actions for fetching a user type
 */
export default function useGetUserType(id?: string): UseGetUserTypeReturn {
  const {http} = useAsgardeo();
  const {getServerUrl} = useConfig();

  const [data, setData] = useState<ApiUserSchema | null>(null);
  const [error, setError] = useState<ApiError | null>(null);
  const [loading, setLoading] = useState(false);

  // Track the id we last fetched to prevent double-fetching
  const lastFetchedIdRef = useRef<string | undefined>(undefined);

  // Keep stable refs to http and getServerUrl to avoid stale closure issues
  const httpRef = useRef(http);
  httpRef.current = http;
  const getServerUrlRef = useRef(getServerUrl);
  getServerUrlRef.current = getServerUrl;

  useEffect(() => {
    if (!id) {
      // Clear data when id becomes undefined
      setData(null);
      setError(null);
      setLoading(false);
      lastFetchedIdRef.current = undefined;
      return undefined;
    }

    // Prevent double-fetch for same id
    if (lastFetchedIdRef.current === id) {
      return undefined;
    }

    lastFetchedIdRef.current = id;

    const controller = new AbortController();
    const {signal} = controller;

    setLoading(true);
    setError(null);

    const currentHttp = httpRef.current;
    const currentGetServerUrl = getServerUrlRef.current;

    currentHttp
      .request({
        url: `${currentGetServerUrl() ?? ''}/user-schemas/${id}`,
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
        signal,
      } as unknown as Parameters<typeof currentHttp.request>[0])
      .then((response: unknown) => {
        if (signal.aborted) return;
        setData((response as {data: ApiUserSchema}).data);
        setLoading(false);
      })
      .catch((err: unknown) => {
        if (signal.aborted) return;
        setLoading(false);
        setError({
          code: 'FETCH_USER_TYPE_ERROR',
          message: err instanceof Error ? err.message : 'An unknown error occurred',
          description: 'Failed to fetch user type',
        });
      });

    return () => {
      controller.abort();
    };
  }, [id]);

  const refetch = useCallback(
    async (newId?: string): Promise<void> => {
      const targetId = newId ?? id;

      if (!targetId) {
        setError({
          code: 'INVALID_ID',
          message: 'Invalid schema ID',
          description: 'Schema ID is required',
        });
        return;
      }

      // Update the last fetched id to prevent re-fetch loop from useEffect
      lastFetchedIdRef.current = targetId;

      setLoading(true);
      setError(null);

      try {
        const serverUrl: string = getServerUrl() ?? '';
        const response: {data: ApiUserSchema} = await http.request({
          url: `${serverUrl}/user-schemas/${targetId}`,
          method: 'GET',
          headers: {
            'Content-Type': 'application/json',
          },
        } as unknown as Parameters<typeof http.request>[0]);

        setData(response.data);
        setLoading(false);
      } catch (err: unknown) {
        setLoading(false);
        setData(null);
        const apiError: ApiError = {
          code: 'FETCH_USER_TYPE_ERROR',
          message: err instanceof Error ? err.message : 'An unknown error occurred',
          description: 'Failed to fetch user type',
        };
        setError(apiError);
        throw err;
      }
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [id],
  );

  return {data, error, loading, refetch};
}
