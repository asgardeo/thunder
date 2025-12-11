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
 * Hook to get the platform default branding preference text customizations from the distribution for multiple screens.
 *
 * @param screens - Array of Resource Screen names.
 * @param locale - Resource Locale.
 * @param shouldFetch - Should fetch the data.
 * @returns SWR response object containing the data, error, isValidating, mutate.
 */
const useGetCustomTextPreferenceFallbacks = (
  screens: PreviewScreenType[],
  locale: string,
  shouldFetch = true,
): CustomTextPreferenceResult => {
  // generate dummy data
  const data: Partial<Record<PreviewScreenType, Record<string, string>>> = {};

  screens.forEach((screen) => {
    data[screen] = {
      sample_key_1: `Sample value 1 for ${screen} in ${locale}`,
      sample_key_2: `Sample value 2 for ${screen} in ${locale}`,
      sample_key_3: `Sample value 3 for ${screen} in ${locale}`,
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

export default useGetCustomTextPreferenceFallbacks;
