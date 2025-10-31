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

import {useState, useCallback} from 'react';
import type {ApiError, ApiUser} from '../types/users';

const API_BASE_URL = 'https://localhost:8090';

/**
 * Request body for updating a user
 * All fields are required as per PUT semantics
 */
export interface UpdateUserRequest {
  organizationUnit: string;
  type: string;
  groups?: string[];
  attributes: Record<string, unknown>;
}

/**
 * Hook to update an existing user
 * PUT https://localhost:8090/users/{userId}
 */
export default function useUpdateUser() {
  const [data, setData] = useState<ApiUser | null>(null);
  const [error, setError] = useState<ApiError | null>(null);
  const [loading, setLoading] = useState(false);

  const updateUser = useCallback(async (userId: string, userData: UpdateUserRequest) => {
    try {
      setLoading(true);
      setError(null);
      setData(null);

      const url = `${API_BASE_URL}/users/${userId}`;

      const response = await fetch(url, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(userData),
      });

      if (!response.ok) {
        // Handle error response
        const contentType = response.headers.get('content-type');
        if (contentType?.includes('application/json')) {
          const errorData = (await response.json()) as ApiError;
          setError(errorData);
          throw new Error(errorData.message ?? 'Failed to update user');
        } else {
          const errorText = await response.text();
          const apiError: ApiError = {
            code: `HTTP_${response.status}`,
            message: response.statusText,
            description: errorText ?? 'Failed to update user',
          };
          setError(apiError);
          throw new Error(apiError.message);
        }
      }

      const result = (await response.json()) as ApiUser;
      setData(result);

      return result;
    } catch (err) {
      if (err instanceof Error) {
        const apiError: ApiError = {
          code: 'UPDATE_USER_ERROR',
          message: err.message,
          description: 'An error occurred while updating the user',
        };
        setError(apiError);
        throw err;
      }
      throw err;
    } finally {
      setLoading(false);
    }
  }, []);

  const reset = useCallback(() => {
    setData(null);
    setError(null);
  }, []);

  return {
    updateUser,
    data,
    loading,
    error,
    reset,
  };
}
