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
import type {SchemaListParams, UserSchemaListResponse} from '../types/users';
import UserQueryKeys from '../constants/user-query-keys';

/**
 * Custom hook to fetch a list of user schemas.
 *
 * @param params - Optional query parameters for pagination
 * @returns TanStack Query result object containing schema list data, loading state, and error information
 */
export default function useGetUserSchemas(params?: SchemaListParams): UseQueryResult<UserSchemaListResponse> {
  const {http} = useAsgardeo();
  const {getServerUrl} = useConfig();
  const {limit, offset} = params ?? {};

  return useQuery<UserSchemaListResponse>({
    queryKey: [UserQueryKeys.USER_SCHEMAS, {limit, offset}],
    queryFn: async (): Promise<UserSchemaListResponse> => {
      const serverUrl: string = getServerUrl();
      const searchParams: URLSearchParams = new URLSearchParams();

      if (limit !== undefined) {
        searchParams.append('limit', String(limit));
      }
      if (offset !== undefined) {
        searchParams.append('offset', String(offset));
      }

      const queryString: string = searchParams.toString();

      const response: {
        data: UserSchemaListResponse;
      } = await http.request({
        url: `${serverUrl}/user-schemas${queryString ? `?${queryString}` : ''}`,
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
      } as unknown as Parameters<typeof http.request>[0]);

      return response.data;
    },
  });
}
