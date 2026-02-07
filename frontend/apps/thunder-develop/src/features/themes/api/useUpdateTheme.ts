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
import type {UpdateThemeRequest} from '../models/requests';
import ThemeQueryKeys from '../constants/theme-query-keys';

/**
 * Parameters for updating a theme
 */
export interface UpdateThemeParams {
  themeId: string;
  themeData: UpdateThemeRequest;
}

/**
 * React Query hook to update an existing theme
 *
 * @returns Mutation result for updating a theme
 */
export default function useUpdateTheme(): UseMutationResult<Theme, Error, UpdateThemeParams> {
  const {http} = useAsgardeo();
  const {getServerUrl} = useConfig();
  const queryClient = useQueryClient();

  return useMutation<Theme, Error, UpdateThemeParams>({
    mutationFn: async ({themeId, themeData}: UpdateThemeParams): Promise<Theme> => {
      const serverUrl: string = getServerUrl();

      const response: {data: Theme} = await http.request({
        url: `${serverUrl}/design/themes/${themeId}`,
        method: 'PUT',
        headers: {'Content-Type': 'application/json'},
        data: JSON.stringify(themeData),
      } as unknown as Parameters<typeof http.request>[0]);

      return response.data;
    },
    onSuccess: (_, {themeId}) => {
      queryClient.invalidateQueries({queryKey: [ThemeQueryKeys.THEMES]}).catch(() => {});
      queryClient.invalidateQueries({queryKey: [ThemeQueryKeys.THEME, themeId]}).catch(() => {});
    },
  });
}
