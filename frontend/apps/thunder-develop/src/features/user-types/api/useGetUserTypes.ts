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
import type {UserSchemaListParams, UserSchemaListResponse, ApiError} from '../types/user-types';

/**
 * Return type for the useGetUserTypes hook.
 */
export interface UseGetUserTypesReturn {
  data: UserSchemaListResponse | null;
  error: ApiError | null;
  loading: boolean;
  refetch: (newParams?: UserSchemaListParams) => Promise<void>;
}

function buildUrl(serverUrl: string, params?: UserSchemaListParams): string {
  const queryParams = new URLSearchParams();
  if (params?.limit !== undefined) {
    queryParams.append('limit', params.limit.toString());
  }
  if (params?.offset !== undefined) {
    queryParams.append('offset', params.offset.toString());
  }
  const queryString = queryParams.toString();
  return `${serverUrl}/user-schemas${queryString ? `?${queryString}` : ''}`;
}

function isAbortError(err: unknown): boolean {
  return (
    (err instanceof Error && err.name === 'AbortError') || (err instanceof Error && err.message === 'Aborted')
  );
}

/**
 * Custom React hook to fetch a paginated list of user schemas (user types) from the Thunder server.
 *
 * @param params - Optional pagination parameters
 * @returns Hook state and actions for fetching user types
 */
export default function useGetUserTypes(params?: UserSchemaListParams): UseGetUserTypesReturn {
  const {http} = useAsgardeo();
  const {getServerUrl} = useConfig();

  const [data, setData] = useState<UserSchemaListResponse | null>(null);
  const [error, setError] = useState<ApiError | null>(null);
  const [loading, setLoading] = useState(false);

  // Keep stable refs to http and getServerUrl to avoid stale closure issues
  const httpRef = useRef(http);
  httpRef.current = http;
  const getServerUrlRef = useRef(getServerUrl);
  getServerUrlRef.current = getServerUrl;

  // Ref to track current params for use inside refetch
  const paramsRef = useRef(params);
  paramsRef.current = params;

  const limit = params?.limit;
  const offset = params?.offset;

  useEffect(() => {
    const controller = new AbortController();
    const {signal} = controller;

    setLoading(true);

    const currentHttp = httpRef.current;
    const currentGetServerUrl = getServerUrlRef.current;
    const serverUrl = currentGetServerUrl() ?? '';
    const url = buildUrl(serverUrl, paramsRef.current);

    currentHttp
      .request({
        url,
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
        signal,
      } as unknown as Parameters<typeof currentHttp.request>[0])
      .then((response: unknown) => {
        if (signal.aborted) return;
        setData((response as {data: UserSchemaListResponse}).data);
        setError(null);
        setLoading(false);
      })
      .catch((err: unknown) => {
        if (signal.aborted) return;
        setLoading(false);
        if (!isAbortError(err)) {
          setError({
            code: 'FETCH_USER_TYPES_ERROR',
            message: err instanceof Error ? err.message : 'An unknown error occurred',
            description: 'Failed to fetch user types',
          });
        }
      });

    return () => {
      controller.abort();
    };
    // Depend on limit and offset individually to trigger re-fetch when params change
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [limit, offset]);

  const refetch = useCallback(
    async (newParams?: UserSchemaListParams): Promise<void> => {
      const fetchParams = newParams ?? paramsRef.current;

      setLoading(true);
      setError(null);

      try {
        const serverUrl: string = getServerUrl() ?? '';
        const url = buildUrl(serverUrl, fetchParams);
        const response: {data: UserSchemaListResponse} = await http.request({
          url,
          method: 'GET',
          headers: {
            'Content-Type': 'application/json',
          },
        } as unknown as Parameters<typeof http.request>[0]);

        setData(response.data);
        setLoading(false);
      } catch (err: unknown) {
        setLoading(false);

        if (!isAbortError(err)) {
          const apiError: ApiError = {
            code: 'FETCH_USER_TYPES_ERROR',
            message: err instanceof Error ? err.message : 'An unknown error occurred',
            description: 'Failed to fetch user types',
          };
          setError(apiError);
        }
        throw err;
      }
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [],
  );

  return {data, error, loading, refetch};
}
