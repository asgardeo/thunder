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

/**
 * Global key for the i18n cache invalidation function.
 */
const I18N_CACHE_INVALIDATE_KEY = 'thunderI18nCacheInvalidate';

/**
 * Type for the window with i18n cache invalidation function.
 */
interface WindowWithI18nCache extends Window {
  [I18N_CACHE_INVALIDATE_KEY]?: () => void;
}

/**
 * Registers the i18n cache invalidation function.
 * Called by I18nProvider to expose the invalidation function.
 *
 * @param invalidateFn - Function to invalidate the i18n cache
 */
export function registerI18nCacheInvalidator(invalidateFn: () => void): void {
  (window as WindowWithI18nCache)[I18N_CACHE_INVALIDATE_KEY] = invalidateFn;
}

/**
 * Unregisters the i18n cache invalidation function.
 * Called by I18nProvider on cleanup.
 */
export function unregisterI18nCacheInvalidator(): void {
  delete (window as WindowWithI18nCache)[I18N_CACHE_INVALIDATE_KEY];
}

/**
 * Invalidates the i18n translations cache, triggering a refetch from the API.
 * Call this after creating or updating translations.
 */
export function invalidateI18nCache(): void {
  (window as WindowWithI18nCache)[I18N_CACHE_INVALIDATE_KEY]?.();
}
