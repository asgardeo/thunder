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

import {useState} from 'react';

import type {ApiError, ApiUserSchema, UpdateUserSchemaRequest} from '../types/user-types';

/**
 * Custom hook to update an existing user schema (user type)
 * @returns Object containing updateUserType function, data, loading state, error, and reset function
 */
export default function useUpdateUserType() {
  const [data, setData] = useState<ApiUserSchema | null>(null);
  const [error, setError] = useState<ApiError | null>(null);
  const [loading, setLoading] = useState(false);

  const updateUserType = async (userTypeId: string, requestData: UpdateUserSchemaRequest): Promise<void> => {
    try {
      setLoading(true);
      setError(null);
      setData(null);

      const response = await fetch(`https://localhost:8090/user-schemas/${userTypeId}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(requestData),
      });

      if (!response.ok) {
        let errorData: ApiError;
        try {
          errorData = (await response.json()) as ApiError;
        } catch {
          errorData = {
            code: 'UPDATE_USER_TYPE_ERROR',
            message: `HTTP error! status: ${response.status}`,
            description: await response.text(),
          };
        }
        setError(errorData);
        throw new Error(errorData.message);
      }

      const jsonData = (await response.json()) as ApiUserSchema;
      setData(jsonData);
      setError(null);
    } catch (err) {
      const apiError: ApiError = {
        code: 'UPDATE_USER_TYPE_ERROR',
        message: err instanceof Error ? err.message : 'An unknown error occurred',
        description: 'Failed to update user type',
      };
      setError(apiError);
      throw err;
    } finally {
      setLoading(false);
    }
  };

  const reset = () => {
    setData(null);
    setError(null);
  };

  return {
    updateUserType,
    data,
    loading,
    error,
    reset,
  };
}
