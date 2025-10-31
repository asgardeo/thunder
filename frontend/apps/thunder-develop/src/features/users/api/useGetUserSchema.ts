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
import type {ApiUserSchema, ApiError} from '../types/users';

const API_BASE_URL = 'https://localhost:8090';

/**
 * Hook to fetch a single user schema by ID from the API
 * GET https://localhost:8090/user-schemas/{id}
 */
export default function useGetUserSchema(id?: string) {
  const [data, setData] = useState<ApiUserSchema | null>(null);
  const [error, setError] = useState<ApiError | null>(null);
  const [loading, setLoading] = useState(false);
  const hasFetchedRef = useRef(false);
  const lastIdRef = useRef<string | undefined>(undefined);

  const fetchUserSchema = useCallback(async (schemaId: string) => {
    try {
      setLoading(true);
      setError(null);

      const url = `${API_BASE_URL}/user-schemas/${schemaId}`;

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
          throw new Error(errorData.message ?? 'Failed to fetch user schema');
        } else {
          const errorText = await response.text();
          const apiError: ApiError = {
            code: `HTTP_${response.status}`,
            message: response.statusText,
            description: errorText ?? 'Failed to fetch user schema',
          };
          setError(apiError);
          throw new Error(apiError.message);
        }
      }

      const result = (await response.json()) as ApiUserSchema;
      setData(result);

      return result;
    } catch (err) {
      if (err instanceof Error) {
        setError({
          code: 'FETCH_ERROR',
          message: err.message,
          description: 'An error occurred while fetching user schema',
        });
      }
      throw err;
    } finally {
      setLoading(false);
    }
  }, []);

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

    fetchUserSchema(id).catch(() => {
      // Error is already handled in fetchUserSchema
    });
  }, [id, fetchUserSchema]);

  const refetch = useCallback(
    (newId?: string) => {
      const schemaId = newId ?? id;
      if (!schemaId) {
        setError({
          code: 'INVALID_ID',
          message: 'Invalid schema ID',
          description: 'Schema ID is required',
        });
        return;
      }

      fetchUserSchema(schemaId).catch(() => {
        // Error is already handled in fetchUserSchema
      });
    },
    [id, fetchUserSchema],
  );

  return {
    data,
    loading,
    error,
    refetch,
  };
}
