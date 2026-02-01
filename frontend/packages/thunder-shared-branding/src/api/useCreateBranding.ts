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

import {useMutation, useQueryClient, type UseMutationResult} from '@tanstack/react-query';
import {useConfig} from '@thunder/shared-contexts';
import {useAsgardeo} from '@asgardeo/react';
import type {Branding} from '../models/branding';
import type {CreateBrandingRequest} from '../models/requests';
import BrandingQueryKeys from '../constants/branding-query-keys';

/**
 * Custom hook to create a new branding configuration in the Thunder server.
 *
 * @returns TanStack Query mutation object for creating branding configurations
 *
 * @example
 * ```tsx
 * function CreateBrandingForm() {
 *   const createBranding = useCreateBranding();
 *
 *   const handleSubmit = (data: CreateBrandingRequest) => {
 *     createBranding.mutate(data, {
 *       onSuccess: (branding) => {
 *         console.log('Branding created:', branding);
 *       },
 *       onError: (error) => {
 *         console.error('Failed to create branding:', error);
 *       }
 *     });
 *   };
 *
 *   return (
 *     <button onClick={() => handleSubmit(data)} disabled={createBranding.isPending}>
 *       {createBranding.isPending ? 'Creating...' : 'Create Branding'}
 *     </button>
 *   );
 * }
 * ```
 */
export default function useCreateBranding(): UseMutationResult<Branding, Error, CreateBrandingRequest> {
  const {http} = useAsgardeo();
  const {getServerUrl} = useConfig();
  const queryClient: ReturnType<typeof useQueryClient> = useQueryClient();

  return useMutation<Branding, Error, CreateBrandingRequest>({
    mutationFn: async (brandingData: CreateBrandingRequest): Promise<Branding> => {
      const serverUrl: string = getServerUrl();
      const response: {
        data: Branding;
      } = await http.request({
        url: `${serverUrl}/branding`,
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        data: JSON.stringify(brandingData),
      } as unknown as Parameters<typeof http.request>[0]);

      return response.data;
    },
    onSuccess: () => {
      // Invalidate and refetch branding list after successful creation
      queryClient.invalidateQueries({queryKey: [BrandingQueryKeys.BRANDINGS]}).catch(() => {
        // Ignore invalidation errors
      });
    },
  });
}
