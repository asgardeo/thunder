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
import type {UpdateLayoutRequest} from '../models/requests';
import LayoutQueryKeys from '../constants/layout-query-keys';

/**
 * Parameters for updating a layout
 */
export interface UpdateLayoutParams {
  layoutId: string;
  layoutData: UpdateLayoutRequest;
}

/**
 * React Query hook to update an existing layout
 *
 * @returns Mutation result for updating a layout
 */
export default function useUpdateLayout(): UseMutationResult<Layout, Error, UpdateLayoutParams> {
  const {http} = useAsgardeo();
  const {getServerUrl} = useConfig();
  const queryClient = useQueryClient();

  return useMutation<Layout, Error, UpdateLayoutParams>({
    mutationFn: async ({layoutId, layoutData}: UpdateLayoutParams): Promise<Layout> => {
      const serverUrl: string = getServerUrl();

      const response: {data: Layout} = await http.request({
        url: `${serverUrl}/design/layouts/${layoutId}`,
        method: 'PUT',
        headers: {'Content-Type': 'application/json'},
        data: JSON.stringify(layoutData),
      } as unknown as Parameters<typeof http.request>[0]);

      return response.data;
    },
    onSuccess: (_, {layoutId}) => {
      queryClient.invalidateQueries({queryKey: [LayoutQueryKeys.LAYOUTS]}).catch(() => {});
      queryClient.invalidateQueries({queryKey: [LayoutQueryKeys.LAYOUT, layoutId]}).catch(() => {});
    },
  });
}
