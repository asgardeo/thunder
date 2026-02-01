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
 * Variables for the {@link useUpdateApplication} mutation.
 *
 * @public
 */
export interface UpdateApplicationVariables {
  /**
   * The unique identifier of the application to update
   */
  applicationId: string;
  /**
   * The updated application data
   */
  data: CreateApplicationRequest;
}

/**
 * Custom React hook to update an existing application in the Thunder server.
 *
 * This hook uses TanStack Query mutations to handle the application update process,
 * providing loading states and error handling. Upon successful update, it automatically
 * updates the cached application data and invalidates related queries to ensure
 * the UI reflects the latest changes.
 *
 * @returns TanStack Query mutation object for updating applications with mutate function, loading state, and error information
 *
 * @example
 * ```tsx
 * function UpdateApplicationForm({ applicationId }: { applicationId: string }) {
 *   const updateApp = useUpdateApplication();
 *
 *   const handleSubmit = (data: CreateApplicationRequest) => {
 *     updateApp.mutate(
 *       { applicationId, data },
 *       {
 *         onSuccess: (application) => {
 *           console.log('Application updated:', application);
 *         },
 *         onError: (error) => {
 *           console.error('Failed to update application:', error);
 *         }
 *       }
 *     );
 *   };
 *
 *   return (
 *     <button onClick={() => handleSubmit(data)} disabled={updateApp.isPending}>
 *       {updateApp.isPending ? 'Updating...' : 'Update Application'}
 *     </button>
 *   );
 * }
 * ```
 *
 * @public
 */
export default function useUpdateApplication(): UseMutationResult<Application, Error, UpdateApplicationVariables> {
  const {http} = useAsgardeo();
  const {getServerUrl} = useConfig();
  const queryClient: ReturnType<typeof useQueryClient> = useQueryClient();

  return useMutation<Application, Error, UpdateApplicationVariables>({
    mutationFn: async ({applicationId, data}: UpdateApplicationVariables): Promise<Application> => {
      const serverUrl: string = getServerUrl();
      const response: {
        data: Application;
      } = await http.request({
        url: `${serverUrl}/applications/${applicationId}`,
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        data: JSON.stringify(data),
      } as unknown as Parameters<typeof http.request>[0]);

      return response.data;
    },
    onSuccess: (_data, variables) => {
      // Invalidate and refetch the specific application
      queryClient
        .invalidateQueries({queryKey: [ApplicationQueryKeys.APPLICATION, variables.applicationId]})
        .catch(() => {
          // Ignore invalidation errors
        });
      // Invalidate and refetch applications list
      queryClient.invalidateQueries({queryKey: [ApplicationQueryKeys.APPLICATIONS]}).catch(() => {
        // Ignore invalidation errors
      });
    },
  });
}
