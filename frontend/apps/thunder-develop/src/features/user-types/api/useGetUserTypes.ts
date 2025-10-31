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

import {useCallback, useEffect, useRef, useState} from 'react';

import type {ApiError, UserSchemaListParams, UserSchemaListResponse} from '../types/user-types';

/**
 * Custom hook to fetch a paginated list of user schemas (user types)
 * @param params - Optional pagination parameters (limit, offset)
 * @returns Object containing data, loading state, error, and refetch function
 */
export default function useGetUserTypes(params?: UserSchemaListParams) {
  const [data, setData] = useState<UserSchemaListResponse | null>(null);
  const [error, setError] = useState<ApiError | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(false);
  const abortControllerRef = useRef<AbortController | null>(null);

  const fetchUserTypes = useCallback(async () => {
    // Cancel previous request if it exists
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
    }

    // Create new abort controller for this request
    abortControllerRef.current = new AbortController();

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
      const url = `https://localhost:8090/user-schemas${queryString ? `?${queryString}` : ''}`;

      const response = await fetch(url, {
        signal: abortControllerRef.current.signal,
      });

      if (!response.ok) {
        let errorData: ApiError;
        try {
          errorData = (await response.json()) as ApiError;
        } catch {
          errorData = {
            code: 'FETCH_USER_TYPES_ERROR',
            message: `HTTP error! status: ${response.status}`,
            description: await response.text(),
          };
        }
        setError(errorData);
        setData(null);
        return;
      }

      const jsonData = (await response.json()) as UserSchemaListResponse;
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
  }, [params]);

  useEffect(() => {
    (async () => {
      await fetchUserTypes();
    })().catch(() => {
      // TODO: Log the errors
      // Tracker: https://github.com/asgardeo/thunder/issues/618
    });

    // Cleanup: abort request on unmount
    return () => {
      if (abortControllerRef.current) {
        abortControllerRef.current.abort();
      }
    };
  }, [fetchUserTypes]);

  return {
    data,
    loading: isLoading,
    error,
    refetch: fetchUserTypes,
  };
}
