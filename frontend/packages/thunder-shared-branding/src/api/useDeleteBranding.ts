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
import BrandingQueryKeys from '../constants/branding-query-keys';

/**
 * Custom hook to delete a branding configuration from the Thunder server.
 *
 * @returns TanStack Query mutation object for deleting branding configurations
 *
 * @example
 * ```tsx
 * function DeleteBrandingButton({ brandingId }: { brandingId: string }) {
 *   const deleteBranding = useDeleteBranding();
 *
 *   const handleDelete = () => {
 *     if (confirm('Are you sure you want to delete this branding configuration?')) {
 *       deleteBranding.mutate(brandingId, {
 *         onSuccess: () => {
 *           console.log('Branding deleted successfully');
 *         },
 *         onError: (error) => {
 *           console.error('Failed to delete branding:', error);
 *         }
 *       });
 *     }
 *   };
 *
 *   return (
 *     <button onClick={handleDelete} disabled={deleteBranding.isPending}>
 *       {deleteBranding.isPending ? 'Deleting...' : 'Delete Branding'}
 *     </button>
 *   );
 * }
 * ```
 */
export default function useDeleteBranding(): UseMutationResult<void, Error, string> {
  const {http} = useAsgardeo();
  const {getServerUrl} = useConfig();
  const queryClient: ReturnType<typeof useQueryClient> = useQueryClient();

  return useMutation<void, Error, string>({
    mutationFn: async (brandingId: string): Promise<void> => {
      const serverUrl: string = getServerUrl();
      await http.request({
        url: `${serverUrl}/branding/${brandingId}`,
        method: 'DELETE',
        headers: {
          'Content-Type': 'application/json',
        },
      } as unknown as Parameters<typeof http.request>[0]);
    },
    onSuccess: (_, brandingId) => {
      // Invalidate the specific branding and the list after successful deletion
      queryClient.invalidateQueries({queryKey: [BrandingQueryKeys.BRANDING, brandingId]}).catch(() => {
        // Ignore invalidation errors
      });
      queryClient.invalidateQueries({queryKey: [BrandingQueryKeys.BRANDINGS]}).catch(() => {
        // Ignore invalidation errors
      });
    },
  });
}
