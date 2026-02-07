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
import {useAsgardeo} from '@asgardeo/react';
import {useConfig} from '@thunder/shared-contexts';
import ThemeQueryKeys from '../constants/theme-query-keys';

/**
 * React Query hook to delete multiple themes
 *
 * @returns Mutation result for bulk deleting themes
 */
export default function useBulkDeleteThemes(): UseMutationResult<void, Error, string[]> {
  const {http} = useAsgardeo();
  const {getServerUrl} = useConfig();
  const queryClient = useQueryClient();

  return useMutation<void, Error, string[]>({
    mutationFn: async (themeIds: string[]): Promise<void> => {
      const serverUrl: string = getServerUrl();

      // Delete all themes in parallel
      await Promise.all(
        themeIds.map(async (themeId) =>
          http.request({
            url: `${serverUrl}/design/themes/${themeId}`,
            method: 'DELETE',
            headers: {'Content-Type': 'application/json'},
          } as unknown as Parameters<typeof http.request>[0])
        )
      );
    },
    onSuccess: () => {
      queryClient.invalidateQueries({queryKey: [ThemeQueryKeys.THEMES]}).catch(() => {});
    },
  });
}
