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

import type {CustomTextPreferenceResult, PreviewScreenType} from '../models/custom-text-preference';

/**
 * Hook to resolve custom text preferences for multiple screen types using useSWR with caching.
 * This is optimized for handling multiple screen types efficiently.
 *
 * @param name - Resource Name.
 * @param screenTypes - Array of Resource Screen types.
 * @param locale - Resource Locale.
 * @param type - Resource Type.
 * @param subOrg - Whether it's for a sub-organization.
 * @param shouldFetch - Should fetch the data.
 * @returns SWR response object containing the data, error, isLoading, mutate.
 */
const useResolveCustomTextPreferences = (
  name: string,
  screenTypes: PreviewScreenType[],
  locale: string,
  type: string,
  shouldFetch = true,
): CustomTextPreferenceResult => {
  // generate dummy data
  const data: Partial<Record<PreviewScreenType, Record<string, string>>> = {};

  screenTypes.forEach((screen) => {
    data[screen] = {
      resolved_key_1: `Resolved value 1 for ${name} on ${screen} in ${locale} of type ${type}`,
      resolved_key_2: `Resolved value 2 for ${name} on ${screen} in ${locale} of type ${type}`,
      resolved_key_3: `Resolved value 3 for ${name} on ${screen} in ${locale} of type ${type}`,
    };
  });

  return {
    data: shouldFetch ? data : undefined,
    error: undefined,
    isLoading: false,
    mutate(): void {
      throw new Error('Function not implemented.');
    },
  };
};

export default useResolveCustomTextPreferences;
