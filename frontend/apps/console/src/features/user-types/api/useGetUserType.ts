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

import {useQuery, type UseQueryResult} from '@tanstack/react-query';
import {useAsgardeo} from '@asgardeo/react';
import {useConfig} from '@thunder/shared-contexts';
import type {ApiUserSchema} from '../types/user-types';
import UserTypeQueryKeys from '../constants/userTypeQueryKeys';

/**
 * Custom React hook to fetch a single user schema (user type) by ID from the Thunder server.
 *
 * @param id - The unique identifier of the user type to fetch
 * @returns TanStack Query result object containing user type data, loading state, and error information
 */
export default function useGetUserType(id?: string): UseQueryResult<ApiUserSchema> {
  const {http} = useAsgardeo();
  const {getServerUrl} = useConfig();

  return useQuery<ApiUserSchema>({
    queryKey: [UserTypeQueryKeys.USER_TYPE, id],
    queryFn: async (): Promise<ApiUserSchema> => {
      const serverUrl: string = getServerUrl();

      const response: {
        data: ApiUserSchema;
      } = await http.request({
        url: `${serverUrl}/user-schemas/${id}`,
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
      } as unknown as Parameters<typeof http.request>[0]);

      return response.data;
    },
    enabled: Boolean(id),
  });
}
