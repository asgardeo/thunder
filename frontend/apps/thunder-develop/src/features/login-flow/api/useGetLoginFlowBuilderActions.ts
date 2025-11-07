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

import useGetFlowBuilderCoreActions from '@/features/flows/api/useGetFlowBuilderCoreActions';
import type {Actions} from '@/features/flows/models/actions';
import actions from '../data/actions.json';

/**
 * Hook to get the actions supported by the flow builder.
 * This hook will aggregate the core actions and the password recovery specific actions.
 *
 * This function calls the GET method of the following endpoint to get the actions.
 * - TODO: Fill this
 * For more details, refer to the documentation:
 * {@link https://TODO:<fillthis>)}
 *
 * @returns SWR response object containing the data, error, isLoading, isValidating, mutate.
 */
const useGetLoginFlowBuilderActions = <Data = Actions>() => {
  const {data: coreActions} = useGetFlowBuilderCoreActions();

  return {
    data: [...coreActions, ...actions] as unknown as Data,
    error: null,
    isLoading: false,
    isValidating: false,
    mutate: () => null,
  };
};

export default useGetLoginFlowBuilderActions;
