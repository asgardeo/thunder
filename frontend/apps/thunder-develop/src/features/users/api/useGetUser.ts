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

import {useState, useCallback, useEffect} from 'react';
import type {ApiUser, ApiError} from '../types/users';

const API_BASE_URL = 'https://localhost:8090';

/**
 * Hook to fetch a single user by ID
 * GET https://localhost:8090/users/{userId}
 */
export default function useGetUser(userId: string | undefined) {
  const [data, setData] = useState<ApiUser | null>(null);
  const [error, setError] = useState<ApiError | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  const fetchUser = useCallback(async () => {
    if (!userId) {
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      const url = `${API_BASE_URL}/users/${userId}`;

      const response = await fetch(url, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        const contentType = response.headers.get('content-type');
        if (contentType?.includes('application/json')) {
          const errorData = (await response.json()) as ApiError;
          setError(errorData);
          throw new Error(errorData.message ?? 'Failed to fetch user');
        } else {
          const errorText = await response.text();
          const apiError: ApiError = {
            code: `HTTP_${response.status}`,
            message: response.statusText,
            description: errorText ?? 'Failed to fetch user',
          };
          setError(apiError);
          throw new Error(apiError.message);
        }
      }

      const result = (await response.json()) as ApiUser;
      setData(result);
    } catch (err) {
      if (err instanceof Error) {
        setError({
          code: 'FETCH_ERROR',
          message: err.message,
          description: 'An error occurred while fetching user',
        });
      }
      throw err;
    } finally {
      setIsLoading(false);
    }
  }, [userId]);

  useEffect(() => {
    fetchUser().catch(() => {
      // Error already handled
    });
  }, [fetchUser]);

  const refetch = useCallback(() => fetchUser(), [fetchUser]);

  return {
    data,
    loading: isLoading,
    error,
    refetch,
  };
}
