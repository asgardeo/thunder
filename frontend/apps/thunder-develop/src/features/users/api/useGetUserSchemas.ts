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

import {useState, useEffect, useCallback, useTransition} from 'react';
import type {UserSchemaListResponse, SchemaListParams, ApiError} from '../types/users';

const API_BASE_URL = 'https://localhost:8090';

/**
 * Hook to fetch user schemas from the API
 * GET https://localhost:8090/user-schemas
 *
 * Uses React's useTransition for non-blocking loading states
 */
export default function useGetUserSchemas(params?: SchemaListParams) {
  const [data, setData] = useState<UserSchemaListResponse | null>(null);
  const [error, setError] = useState<ApiError | null>(null);
  const [isPending, startFetchTransition] = useTransition();

  const fetchUserSchemas = useCallback(
    async (queryParams?: SchemaListParams) => {
      try {
        setError(null);

        // Build query string
        const searchParams = new URLSearchParams();
        const finalParams = queryParams || params;

        if (finalParams?.limit !== undefined) {
          searchParams.append('limit', String(finalParams.limit));
        }
        if (finalParams?.offset !== undefined) {
          searchParams.append('offset', String(finalParams.offset));
        }

        const queryString = searchParams.toString();
        const url = `${API_BASE_URL}/user-schemas${queryString ? `?${queryString}` : ''}`;

        const response = await fetch(url, {
          method: 'GET',
          headers: {
            'Content-Type': 'application/json',
          },
        });

        if (!response.ok) {
          // Handle error response
          const contentType = response.headers.get('content-type');
          if (contentType?.includes('application/json')) {
            const errorData = (await response.json()) as ApiError;
            setError(errorData);
            throw new Error(errorData.message || 'Failed to fetch user schemas');
          } else {
            const errorText = await response.text();
            const apiError: ApiError = {
              code: `HTTP_${response.status}`,
              message: response.statusText,
              description: errorText || 'Failed to fetch user schemas',
            };
            setError(apiError);
            throw new Error(apiError.message);
          }
        }

        const result = (await response.json()) as UserSchemaListResponse;

        // Use startFetchTransition to update state without blocking
        startFetchTransition(() => {
          setData(result);
        });

        return result;
      } catch (err) {
        if (err instanceof Error) {
          setError({
            code: 'FETCH_ERROR',
            message: err.message,
            description: 'An error occurred while fetching user schemas',
          });
        }
        throw err;
      }
    },
    [params],
  );

  useEffect(() => {
    startFetchTransition(() => {
      fetchUserSchemas().catch(() => {
        // Error is already handled in fetchUserSchemas
      });
    });
  }, [fetchUserSchemas]);

  const refetch = useCallback(
    (newParams?: SchemaListParams) => {
      startFetchTransition(() => {
        fetchUserSchemas(newParams).catch(() => {
          // Error is already handled in fetchUserSchemas
        });
      });
    },
    [fetchUserSchemas, startFetchTransition],
  );

  return {
    data,
    loading: isPending,
    error,
    refetch,
  };
}
