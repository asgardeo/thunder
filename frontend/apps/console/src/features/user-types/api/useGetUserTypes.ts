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
import type {UserSchemaListParams, UserSchemaListResponse} from '../types/user-types';
import UserTypeQueryKeys from '../constants/userTypeQueryKeys';

/**
 * Custom React hook to fetch a paginated list of user schemas (user types) from the Thunder server.
 *
 * @param params - Optional pagination parameters
 * @param params.limit - Maximum number of records to return
 * @param params.offset - Number of records to skip for pagination
 * @returns TanStack Query result object containing user types list data, loading state, and error information
 */
export default function useGetUserTypes(params?: UserSchemaListParams): UseQueryResult<UserSchemaListResponse> {
  const {http} = useAsgardeo();
  const {getServerUrl} = useConfig();
  const {limit, offset} = params ?? {};

  return useQuery<UserSchemaListResponse>({
    queryKey: [UserTypeQueryKeys.USER_TYPES, {limit, offset}],
    queryFn: async (): Promise<UserSchemaListResponse> => {
      const serverUrl: string = getServerUrl();
      const queryParams: URLSearchParams = new URLSearchParams();

      if (limit !== undefined) {
        queryParams.append('limit', limit.toString());
      }
      if (offset !== undefined) {
        queryParams.append('offset', offset.toString());
      }

      const queryString: string = queryParams.toString();
      const url = `${serverUrl}/user-schemas${queryString ? `?${queryString}` : ''}`;

      const response: {
        data: UserSchemaListResponse;
      } = await http.request({
        url,
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
      } as unknown as Parameters<typeof http.request>[0]);

      return response.data;
    },
  });
}
