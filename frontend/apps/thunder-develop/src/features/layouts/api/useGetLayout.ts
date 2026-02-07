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
import {useAsgardeo} from '@asgardeo/react';
import {useConfig} from '@thunder/shared-contexts';
import type {Layout} from '../models/layout';
import LayoutQueryKeys from '../constants/layout-query-keys';

/**
 * React Query hook to fetch a single layout by ID
 *
 * @param layoutId - The ID of the layout to fetch
 * @returns Query result containing the layout
 */
export default function useGetLayout(layoutId: string): UseQueryResult<Layout> {
  const {http} = useAsgardeo();
  const {getServerUrl} = useConfig();

  return useQuery<Layout>({
    queryKey: [LayoutQueryKeys.LAYOUT, layoutId],
    queryFn: async (): Promise<Layout> => {
      const serverUrl: string = getServerUrl();

      const response: {data: Layout} = await http.request({
        url: `${serverUrl}/design/layouts/${layoutId}`,
        method: 'GET',
        headers: {'Content-Type': 'application/json'},
      } as unknown as Parameters<typeof http.request>[0]);

      return response.data;
    },
    enabled: Boolean(layoutId),
  });
}
