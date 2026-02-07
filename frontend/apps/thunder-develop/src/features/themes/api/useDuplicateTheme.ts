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
import type {Theme} from '../models/theme';
import type {CreateThemeRequest} from '../models/requests';
import ThemeQueryKeys from '../constants/theme-query-keys';

/**
 * React Query hook to duplicate a theme
 *
 * @returns Mutation result for duplicating a theme
 */
export default function useDuplicateTheme(): UseMutationResult<Theme, Error, string> {
  const {http} = useAsgardeo();
  const {getServerUrl} = useConfig();
  const queryClient = useQueryClient();

  return useMutation<Theme, Error, string>({
    mutationFn: async (themeId: string): Promise<Theme> => {
      const serverUrl: string = getServerUrl();

      // Fetch the original theme
      const getResponse: {data: Theme} = await http.request({
        url: `${serverUrl}/design/themes/${themeId}`,
        method: 'GET',
        headers: {'Content-Type': 'application/json'},
      } as unknown as Parameters<typeof http.request>[0]);

      const originalTheme = getResponse.data;

      // Create a copy with "(Copy)" suffix
      const duplicateRequest: CreateThemeRequest = {
        displayName: `${originalTheme.displayName} (Copy)`,
        theme: originalTheme.theme,
      };

      // Create the duplicate theme
      const createResponse: {data: Theme} = await http.request({
        url: `${serverUrl}/design/themes`,
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        data: JSON.stringify(duplicateRequest),
      } as unknown as Parameters<typeof http.request>[0]);

      return createResponse.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({queryKey: [ThemeQueryKeys.THEMES]}).catch(() => {});
    },
  });
}
