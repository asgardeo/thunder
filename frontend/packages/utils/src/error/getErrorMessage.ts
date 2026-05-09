/**
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
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

import type {ApiError} from '@thunderid/types';

/**
 * A minimal subset of the i18next TFunction interface used by {@link getErrorMessage}.
 */
type TranslateFn = (key: string, options?: {defaultValue: string}) => string;

/**
 * Extracts a localized error message from an API error response.
 *
 * Attempts to resolve a specific i18n message for the error code returned
 * by the API (e.g. `errors.APP-1020`). If no specific translation exists,
 * falls back to the provided generic key.
 *
 * @param error - The error thrown by the mutation
 * @param t - The i18next translation function scoped to the relevant namespace
 * @param fallbackKey - i18n key to use when no specific message is found (e.g. `'create.error'`)
 * @returns Localized error message string
 *
 * @example
 * ```typescript
 * onError: (error) => {
 *   showToast(getErrorMessage(error, t, 'create.error'), 'error');
 * }
 * ```
 *
 * @public
 */
export default function getErrorMessage(error: Error, t: TranslateFn, fallbackKey: string): string {
  const apiError = (error as {response?: {data?: ApiError}}).response?.data;

  if (apiError?.code) {
    const specific = t(`errors.${apiError.code}`, {defaultValue: ''});

    if (specific) {
      return specific;
    }
  }

  return t(fallbackKey);
}
