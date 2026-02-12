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
import type {BrandingResponse} from '../models/responses';
import BrandingQueryKeys from '../constants/branding-query-keys';

/**
 * Custom hook to fetch a single branding configuration by ID from the Thunder server.
 *
 * @param brandingId - The unique identifier of the branding configuration
 * @returns TanStack Query result object with branding data
 *
 * @example
 * ```tsx
 * function BrandingDetails({ id }: { id: string }) {
 *   const { data, isLoading, error } = useGetBranding(id);
 *
 *   if (isLoading) return <div>Loading...</div>;
 *   if (error) return <div>Error: {error.message}</div>;
 *   if (!data) return <div>Not found</div>;
 *
 *   return (
 *     <div>
 *       <h1>{data.displayName}</h1>
 *       <pre>{JSON.stringify(data.preferences, null, 2)}</pre>
 *     </div>
 *   );
 * }
 * ```
 */
export default function useGetBranding(brandingId: string): UseQueryResult<BrandingResponse> {
  const {http} = useAsgardeo();
  const {getServerUrl} = useConfig();

  return useQuery<BrandingResponse>({
    queryKey: [BrandingQueryKeys.BRANDING, brandingId],
    queryFn: async (): Promise<BrandingResponse> => {
      const serverUrl: string = getServerUrl();

      const response: {
        data: BrandingResponse;
      } = await http.request({
        url: `${serverUrl}/branding/${brandingId}`,
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
      } as unknown as Parameters<typeof http.request>[0]);

      return response.data;
    },
    enabled: Boolean(brandingId),
  });
}
