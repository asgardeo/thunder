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
import type {Application} from '../models/application';
import type {CreateApplicationRequest} from '../models/requests';
import ApplicationQueryKeys from '../constants/application-query-keys';

/**
 * Custom React hook to create a new application in the Thunder server.
 *
 * This hook uses TanStack Query mutations to handle the application creation process,
 * providing loading states, error handling, and automatic cache invalidation. Upon
 * successful creation, it automatically invalidates the applications list query to
 * trigger a refetch.
 *
 * @returns TanStack Query mutation object for creating applications with mutate function, loading state, and error information
 *
 * @example
 * ```tsx
 * function CreateApplicationForm() {
 *   const createApp = useCreateApplication();
 *
 *   const handleSubmit = (data: CreateApplicationRequest) => {
 *     createApp.mutate(data, {
 *       onSuccess: (application) => {
 *         console.log('Application created:', application);
 *       },
 *       onError: (error) => {
 *         console.error('Failed to create application:', error);
 *       }
 *     });
 *   };
 *
 *   return (
 *     <button onClick={() => handleSubmit(data)} disabled={createApp.isPending}>
 *       {createApp.isPending ? 'Creating...' : 'Create Application'}
 *     </button>
 *   );
 * }
 * ```
 *
 * @public
 */
export default function useCreateApplication(): UseMutationResult<Application, Error, CreateApplicationRequest> {
  const {http} = useAsgardeo();
  const {getServerUrl} = useConfig();
  const queryClient: ReturnType<typeof useQueryClient> = useQueryClient();

  return useMutation<Application, Error, CreateApplicationRequest>({
    mutationFn: async (applicationData: CreateApplicationRequest): Promise<Application> => {
      const serverUrl: string = getServerUrl();
      const response: {
        data: Application;
      } = await http.request({
        url: `${serverUrl}/applications`,
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        data: JSON.stringify(applicationData),
      } as unknown as Parameters<typeof http.request>[0]);

      return response.data;
    },
    onSuccess: () => {
      // Invalidate and refetch applications list after successful creation
      queryClient.invalidateQueries({queryKey: [ApplicationQueryKeys.APPLICATIONS]}).catch(() => {
        // Ignore invalidation errors
      });
    },
  });
}
