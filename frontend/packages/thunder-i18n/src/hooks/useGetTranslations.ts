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
 * Response from the translations API.
 */
export interface TranslationsResponse {
  language: string;
  totalResults?: number;
  translations: Record<string, Record<string, string>>;
}

/**
 * Options for the useGetTranslations hook.
 */
export interface UseGetTranslationsOptions {
  serverUrl: string;
  language: string;
  namespace?: string;
  enabled?: boolean;
}

/**
 * Custom hook to fetch translations for a language.
 *
 * @param options - Options for fetching translations
 * @returns TanStack Query object for fetching translations
 *
 * @example
 * ```tsx
 * function TranslationsDisplay() {
 *   const { data, isLoading, error } = useGetTranslations({
 *     serverUrl: 'https://api.example.com',
 *     language: 'en',
 *     namespace: 'flowCustomi18n',
 *   });
 *
 *   if (isLoading) return <Spinner />;
 *   if (error) return <Error message={error.message} />;
 *
 *   return (
 *     <ul>
 *       {Object.entries(data?.translations || {}).map(([ns, keys]) => (
 *         Object.entries(keys).map(([key, value]) => (
 *           <li key={`${ns}.${key}`}>{key}: {value}</li>
 *         ))
 *       ))}
 *     </ul>
 *   );
 * }
 * ```
 */
export default function useGetTranslations({
  serverUrl,
  language,
  namespace,
  enabled = true,
}: UseGetTranslationsOptions): UseQueryResult<TranslationsResponse, Error> {
  const {http} = useAsgardeo();

  return useQuery<TranslationsResponse, Error>({
    queryKey: namespace ? [I18nQueryKeys.TRANSLATIONS, language, namespace] : [I18nQueryKeys.TRANSLATIONS, language],
    queryFn: async (): Promise<TranslationsResponse> => {
      let url = `${serverUrl}/i18n/languages/${language}/translations/resolve`;

      if (namespace) {
        url += `?namespace=${encodeURIComponent(namespace)}`;
      }

      const response: {
        data: TranslationsResponse;
      } = await http.request({
        url,
        method: 'GET',
        attachToken: false,
        withCredentials: false,
      } as unknown as Parameters<typeof http.request>[0]);

      return response.data;
    },
    enabled: enabled && !!language && !!serverUrl,
  });
}
