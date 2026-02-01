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
import I18nQueryKeys from '../constants/I18nQueryKeys';

/**
 * Response from the languages API.
 */
export interface LanguagesResponse {
  languages: string[];
}

/**
 * Options for the useGetLanguages hook.
 */
export interface UseGetLanguagesOptions {
  serverUrl: string;
  enabled?: boolean;
}

/**
 * Custom hook to fetch available languages.
 *
 * @param options - Options for fetching languages
 * @returns TanStack Query object for fetching languages
 *
 * @example
 * ```tsx
 * function LanguageSelector() {
 *   const { data, isLoading, error } = useGetLanguages({
 *     serverUrl: 'https://api.example.com',
 *   });
 *
 *   if (isLoading) return <Spinner />;
 *   if (error) return <Error message={error.message} />;
 *
 *   return (
 *     <Select>
 *       {data?.languages.map(lang => (
 *         <Option key={lang} value={lang}>{lang}</Option>
 *       ))}
 *     </Select>
 *   );
 * }
 * ```
 */
export default function useGetLanguages({
  serverUrl,
  enabled = true,
}: UseGetLanguagesOptions): UseQueryResult<LanguagesResponse, Error> {
  const {http} = useAsgardeo();

  return useQuery<LanguagesResponse, Error>({
    queryKey: [I18nQueryKeys.LANGUAGES],
    queryFn: async (): Promise<LanguagesResponse> => {
      const response: {
        data: LanguagesResponse;
      } = await http.request({
        url: `${serverUrl}/i18n/languages`,
        method: 'GET',
        attachToken: false,
        withCredentials: false,
      } as unknown as Parameters<typeof http.request>[0]);

      return response.data;
    },
    enabled: enabled && !!serverUrl,
  });
}
