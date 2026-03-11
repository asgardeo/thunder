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
import {useAsgardeo} from '@asgardeo/react';
import {useConfig} from '@thunder/shared-contexts';
import type {ApiUserSchema, CreateUserSchemaRequest, ApiError} from '../types/user-types';

/**
 * Return type for the useCreateUserType hook.
 */
export interface UseCreateUserTypeReturn {
  data: ApiUserSchema | null;
  error: ApiError | null;
  loading: boolean;
  createUserType: (requestData: CreateUserSchemaRequest) => Promise<ApiUserSchema>;
  reset: () => void;
}

/**
 * Custom React hook to create a new user schema (user type) in the Thunder server.
 *
 * @returns Hook state and actions for creating user types
 */
export default function useCreateUserType(): UseCreateUserTypeReturn {
  const {http} = useAsgardeo();
  const {getServerUrl} = useConfig();

  const [data, setData] = useState<ApiUserSchema | null>(null);
  const [error, setError] = useState<ApiError | null>(null);
  const [loading, setLoading] = useState(false);

  const createUserType = useCallback(
    async (requestData: CreateUserSchemaRequest): Promise<ApiUserSchema> => {
      setLoading(true);
      setError(null);

      try {
        const serverUrl: string = getServerUrl() ?? '';
        const response: {data: ApiUserSchema} = await http.request({
          url: `${serverUrl}/user-schemas`,
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          data: requestData,
        } as unknown as Parameters<typeof http.request>[0]);

        setData(response.data);
        setLoading(false);
        return response.data;
      } catch (err: unknown) {
        setLoading(false);
        const apiError: ApiError = {
          code: 'CREATE_USER_TYPE_ERROR',
          message: err instanceof Error ? err.message : 'An unknown error occurred',
          description: 'Failed to create user type',
        };
        setError(apiError);
        throw err;
      }
    },
    [getServerUrl, http],
  );

  const reset = useCallback(() => {
    setData(null);
    setError(null);
    setLoading(false);
  }, []);

  return {data, error, loading, createUserType, reset};
}
