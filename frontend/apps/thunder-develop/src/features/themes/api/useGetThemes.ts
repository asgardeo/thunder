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
import type {ThemeListResponse} from '../models/responses';
import ThemeQueryKeys from '../constants/theme-query-keys';

/**
 * Parameters for useGetThemes hook
 */
export interface UseGetThemesParams {
  limit?: number;
  offset?: number;
}

/**
 * React Query hook to fetch themes list
 *
 * @param params - Optional pagination parameters
 * @returns Query result containing theme list
 */
export default function useGetThemes(params?: UseGetThemesParams): UseQueryResult<ThemeListResponse> {
  const {http} = useAsgardeo();
  const {getServerUrl} = useConfig();
  const {limit = 30, offset = 0} = params ?? {};

  return useQuery<ThemeListResponse>({
    queryKey: [ThemeQueryKeys.THEMES, {limit, offset}],
    queryFn: async (): Promise<ThemeListResponse> => {
      const serverUrl: string = getServerUrl();
      const queryParams: URLSearchParams = new URLSearchParams({
        limit: limit.toString(),
        offset: offset.toString(),
      });

      const response: {data: ThemeListResponse} = await http.request({
        url: `${serverUrl}/design/themes?${queryParams.toString()}`,
        method: 'GET',
        headers: {'Content-Type': 'application/json'},
      } as unknown as Parameters<typeof http.request>[0]);

      return response.data;
    },
  });
}
