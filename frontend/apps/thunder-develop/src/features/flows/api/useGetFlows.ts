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
import {useConfig} from '@thunder/commons-contexts';
import {useAsgardeo} from '@asgardeo/react';
import type {FlowListResponse, FlowType} from '../models/responses';
import FlowQueryKeys from '../constants/flow-query-keys';

/**
 * Parameters for the {@link useGetFlows} hook.
 */
export interface UseGetFlowsParams {
  /**
   * Filter by flow type (AUTHENTICATION or REGISTRATION)
   */
  flowType?: FlowType;
  /**
   * Maximum number of records to return (default: 30, max: 100)
   */
  limit?: number;
  /**
   * Number of records to skip for pagination (default: 0)
   */
  offset?: number;
}

/**
 * Custom React hook to fetch a paginated list of flows from the Thunder server.
 *
 * This hook uses TanStack Query to manage the server state and provides automatic
 * caching, refetching, and background updates. The query is keyed by the pagination
 * and filter parameters to ensure proper cache management.
 *
 * @param params - Optional pagination and filter parameters
 * @param params.flowType - Filter by flow type (AUTHENTICATION or REGISTRATION)
 * @param params.limit - Maximum number of records to return (default: 30)
 * @param params.offset - Number of records to skip for pagination (default: 0)
 * @returns TanStack Query result object containing flows list data, loading state, and error information
 *
 * @example
 * ```tsx
 * function FlowsList() {
 *   const { data, isLoading, error } = useGetFlows({ limit: 10, offset: 0 });
 *
 *   if (isLoading) return <div>Loading...</div>;
 *   if (error) return <div>Error: {error.message}</div>;
 *
 *   return (
 *     <ul>
 *       {data?.flows.map((flow) => (
 *         <li key={flow.id}>{flow.name}</li>
 *       ))}
 *     </ul>
 *   );
 * }
 * ```
 */
export default function useGetFlows(params?: UseGetFlowsParams): UseQueryResult<FlowListResponse> {
  const {http} = useAsgardeo();
  const {getServerUrl} = useConfig();
  const {flowType, limit = 30, offset = 0} = params ?? {};

  return useQuery<FlowListResponse>({
    queryKey: [FlowQueryKeys.FLOWS, {flowType, limit, offset}],
    queryFn: async (): Promise<FlowListResponse> => {
      const serverUrl: string = getServerUrl();
      const queryParams: URLSearchParams = new URLSearchParams({
        limit: limit.toString(),
        offset: offset.toString(),
      });

      if (flowType) {
        queryParams.append('flowType', flowType);
      }

      const response: {
        data: FlowListResponse;
      } = await http.request({
        url: `${serverUrl}/flows?${queryParams.toString()}`,
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
      } as unknown as Parameters<typeof http.request>[0]);

      return response.data;
    },
  });
}
