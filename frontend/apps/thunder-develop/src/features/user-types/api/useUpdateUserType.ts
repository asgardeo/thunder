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
import type {ApiUserSchema, UpdateUserSchemaRequest} from '../types/user-types';
import UserTypeQueryKeys from '../constants/userTypeQueryKeys';

/**
 * Variables for the {@link useUpdateUserType} mutation.
 */
export interface UpdateUserTypeVariables {
  /**
   * The unique identifier of the user type to update
   */
  userTypeId: string;
  /**
   * The updated user type data
   */
  data: UpdateUserSchemaRequest;
}

/**
 * Custom React hook to update an existing user schema (user type) in the Thunder server.
 *
 * @returns TanStack Query mutation object for updating user types
 */
export default function useUpdateUserType(): UseMutationResult<ApiUserSchema, Error, UpdateUserTypeVariables> {
  const {http} = useAsgardeo();
  const {getServerUrl} = useConfig();
  const queryClient: ReturnType<typeof useQueryClient> = useQueryClient();

  return useMutation<ApiUserSchema, Error, UpdateUserTypeVariables>({
    mutationFn: async ({userTypeId, data}: UpdateUserTypeVariables): Promise<ApiUserSchema> => {
      const serverUrl: string = getServerUrl();
      const response: {
        data: ApiUserSchema;
      } = await http.request({
        url: `${serverUrl}/user-schemas/${userTypeId}`,
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
        .invalidateQueries({queryKey: [UserTypeQueryKeys.USER_TYPE, variables.userTypeId]})
        .catch(() => {
          // Ignore invalidation errors
        });
      queryClient.invalidateQueries({queryKey: [UserTypeQueryKeys.USER_TYPES]}).catch(() => {
        // Ignore invalidation errors
      });
    },
  });
}
