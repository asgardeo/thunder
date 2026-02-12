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
import type {UpdateBrandingRequest} from '../models/requests';
import BrandingQueryKeys from '../constants/branding-query-keys';

interface UpdateBrandingParams {
  brandingId: string;
  data: UpdateBrandingRequest;
}

/**
 * Custom hook to update an existing branding configuration in the Thunder server.
 *
 * @returns TanStack Query mutation object for updating branding configurations
 *
 * @example
 * ```tsx
 * function EditBrandingForm({ brandingId }: { brandingId: string }) {
 *   const updateBranding = useUpdateBranding();
 *
 *   const handleSubmit = (data: UpdateBrandingRequest) => {
 *     updateBranding.mutate(
 *       { brandingId, data },
 *       {
 *         onSuccess: (branding) => {
 *           console.log('Branding updated:', branding);
 *         },
 *         onError: (error) => {
 *           console.error('Failed to update branding:', error);
 *         }
 *       }
 *     );
 *   };
 *
 *   return (
 *     <button onClick={() => handleSubmit(data)} disabled={updateBranding.isPending}>
 *       {updateBranding.isPending ? 'Updating...' : 'Update Branding'}
 *     </button>
 *   );
 * }
 * ```
 */
export default function useUpdateBranding(): UseMutationResult<Branding, Error, UpdateBrandingParams> {
  const {http} = useAsgardeo();
  const {getServerUrl} = useConfig();
  const queryClient: ReturnType<typeof useQueryClient> = useQueryClient();

  return useMutation<Branding, Error, UpdateBrandingParams>({
    mutationFn: async ({brandingId, data}: UpdateBrandingParams): Promise<Branding> => {
      const serverUrl: string = getServerUrl();
      const response: {
        data: Branding;
      } = await http.request({
        url: `${serverUrl}/branding/${brandingId}`,
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        data: JSON.stringify(data),
      } as unknown as Parameters<typeof http.request>[0]);

      return response.data;
    },
    onSuccess: (_, {brandingId}) => {
      // Invalidate the specific branding and the list after successful update
      queryClient.invalidateQueries({queryKey: [BrandingQueryKeys.BRANDING, brandingId]}).catch(() => {
        // Ignore invalidation errors
      });
      queryClient.invalidateQueries({queryKey: [BrandingQueryKeys.BRANDINGS]}).catch(() => {
        // Ignore invalidation errors
      });
    },
  });
}
