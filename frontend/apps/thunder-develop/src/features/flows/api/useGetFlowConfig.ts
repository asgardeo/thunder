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

import type {FlowConfigInterface} from '../models/flows';
import type {FlowTypes} from '../models/metadata';

/**
 * Hook to fetch the flow configuration.
 *
 * @param shouldFetch - Should fetch data from the API.
 * @returns Flow configuration response.
 */
const useGetFlowConfig = <Data = FlowConfigInterface>(flowType: FlowTypes, shouldFetch = true) => {
  // return mock data for now
  const data: FlowConfigInterface = {
    flowCompletionConfigs: {},
    flowType,
    isEnabled: true,
  };

  return {
    data: shouldFetch ? (data as unknown as Data) : null,
    error: null,
    isLoading: false,
    isValidating: false,
    mutate: () => null,
  };
};

export default useGetFlowConfig;
