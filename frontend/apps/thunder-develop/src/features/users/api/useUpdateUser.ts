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
import type {ApiUser, UpdateUserRequest} from '../types/users';
import UserQueryKeys from '../constants/user-query-keys';

/**
 * Variables for the update user mutation.
 */
export interface UpdateUserVariables {
  userId: string;
  data: UpdateUserRequest;
}

/**
 * Custom hook to update an existing user.
 *
 * @returns TanStack Query mutation object for updating users
 */
export default function useUpdateUser(): UseMutationResult<ApiUser, Error, UpdateUserVariables> {
  const {http} = useAsgardeo();
  const {getServerUrl} = useConfig();
  const queryClient: ReturnType<typeof useQueryClient> = useQueryClient();

  return useMutation<ApiUser, Error, UpdateUserVariables>({
    mutationFn: async ({userId, data}: UpdateUserVariables): Promise<ApiUser> => {
      const serverUrl: string = getServerUrl();

      const response: {
        data: ApiUser;
      } = await http.request({
        url: `${serverUrl}/users/${userId}`,
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        data: JSON.stringify(data),
      } as unknown as Parameters<typeof http.request>[0]);

      return response.data;
    },
    onSuccess: (_data, variables) => {
      queryClient
        .invalidateQueries({queryKey: [UserQueryKeys.USER, variables.userId]})
        .catch(() => {
          // Ignore invalidation errors
        });
      queryClient.invalidateQueries({queryKey: [UserQueryKeys.USERS]}).catch(() => {
        // Ignore invalidation errors
      });
    },
  });
}
