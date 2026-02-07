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
import type {Layout} from '../models/layout';
import type {CreateLayoutRequest} from '../models/requests';
import LayoutQueryKeys from '../constants/layout-query-keys';

/**
 * React Query hook to create a new layout
 *
 * @returns Mutation result for creating a layout
 */
export default function useCreateLayout(): UseMutationResult<Layout, Error, CreateLayoutRequest> {
  const {http} = useAsgardeo();
  const {getServerUrl} = useConfig();
  const queryClient = useQueryClient();

  return useMutation<Layout, Error, CreateLayoutRequest>({
    mutationFn: async (layoutData: CreateLayoutRequest): Promise<Layout> => {
      const serverUrl: string = getServerUrl();

      const response: {data: Layout} = await http.request({
        url: `${serverUrl}/design/layouts`,
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        data: JSON.stringify(layoutData),
      } as unknown as Parameters<typeof http.request>[0]);

      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({queryKey: [LayoutQueryKeys.LAYOUTS]}).catch(() => {});
    },
  });
}
