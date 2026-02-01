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
import type {BrandingResponse} from '../models/responses';
import BrandingQueryKeys from '../constants/branding-query-keys';

/**
 * Types for branding resolve functionality
 */
type BrandingResolveType = 'APP' | 'OU';

interface BrandingResolveParams {
  type: BrandingResolveType;
  id: string;
}

/**
 * Custom hook to resolve branding configuration by type and ID from the Thunder server.
 * Uses the /branding/resolve endpoint to fetch branding based on application or organizational unit.
 *
 * @param params - Object containing type ('APP' or 'OU') and id of the entity
 * @param options - Optional React Query configuration options
 * @returns TanStack Query result object with resolved branding data
 *
 * @example
 * ```tsx
 * function ApplicationBranding({ applicationId }: { applicationId: string }) {
 *   const { data, isLoading, error } = useGetBrandingResolve({
 *     type: 'APP',
 *     id: applicationId
 *   });
 *
 *   if (isLoading) return <div>Loading...</div>;
 *   if (error) return <div>Error: {error.message}</div>;
 *   if (!data) return <div>No branding found</div>;
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
export default function useGetBrandingResolve(
  params: BrandingResolveParams,
  options?: {enabled?: boolean},
): UseQueryResult<BrandingResponse> {
  const {getServerUrl} = useConfig();

  const isEnabled = options?.enabled ?? Boolean(params?.type && params?.id && params.id.trim().length > 0);

  return useQuery<BrandingResponse>({
    queryKey: [BrandingQueryKeys.BRANDING_RESOLVE, params.type, params.id],
    queryFn: async (): Promise<BrandingResponse> => {
      const serverUrl: string = getServerUrl();
      const queryParams = new URLSearchParams({
        type: params.type,
        id: params.id,
      });

      const requestUrl = `${serverUrl}/branding/resolve?${queryParams.toString()}`;

      const response = await fetch(requestUrl, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      return response.json() as Promise<BrandingResponse>;
    },
    enabled: isEnabled,
    retry: false,
  });
}
