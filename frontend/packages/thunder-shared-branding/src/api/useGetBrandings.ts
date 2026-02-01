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
import {useConfig} from '@thunder/shared-contexts';
import {useAsgardeo} from '@asgardeo/react';
import type {BrandingListResponse} from '../models/responses';
import BrandingQueryKeys from '../constants/branding-query-keys';

interface UseGetBrandingsParams {
  limit?: number;
  offset?: number;
}

/**
 * Custom hook to fetch the list of branding configurations from the Thunder server.
 *
 * @param params - Optional query parameters
 * @param params.limit - Maximum number of records to return (default: 30)
 * @param params.offset - Number of records to skip for pagination (default: 0)
 * @returns TanStack Query result object with branding list data
 *
 * @example
 * ```tsx
 * function BrandingList() {
 *   const { data, isLoading, error } = useGetBrandings({ limit: 10, offset: 0 });
 *
 *   if (isLoading) return <div>Loading...</div>;
 *   if (error) return <div>Error: {error.message}</div>;
 *
 *   return (
 *     <ul>
 *       {data?.brandings.map((branding) => (
 *         <li key={branding.id}>{branding.displayName}</li>
 *       ))}
 *     </ul>
 *   );
 * }
 * ```
 */
export default function useGetBrandings(params?: UseGetBrandingsParams): UseQueryResult<BrandingListResponse> {
  const {http} = useAsgardeo();
  const {getServerUrl} = useConfig();
  const {limit = 30, offset = 0} = params ?? {};

  return useQuery<BrandingListResponse>({
    queryKey: [BrandingQueryKeys.BRANDINGS, {limit, offset}],
    queryFn: async (): Promise<BrandingListResponse> => {
      const serverUrl: string = getServerUrl();
      const queryParams: URLSearchParams = new URLSearchParams({
        limit: limit.toString(),
        offset: offset.toString(),
      });

      const response: {
        data: BrandingListResponse;
      } = await http.request({
        url: `${serverUrl}/branding?${queryParams.toString()}`,
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
      } as unknown as Parameters<typeof http.request>[0]);

      return response.data;
    },
  });
}
