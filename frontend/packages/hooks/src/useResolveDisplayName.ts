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

import {isI18nTemplatePattern, type TemplateLiteralHandlers} from '@thunderid/utils';
import {useCallback} from 'react';
import useTemplateLiteralResolver from './useTemplateLiteralResolver';

/**
 * Options for the useResolveDisplayName hook.
 */
export interface UseResolveDisplayNameOptions {
  /** Template literal handlers (typically `{ t }` from `useTranslation()`). */
  handlers: TemplateLiteralHandlers;
}

/**
 * Return type for useResolveDisplayName hook.
 */
export interface ResolveDisplayNameResult {
  /**
   * Resolves a display name string. If the value is an i18n template pattern,
   * it resolves the translation. Returns the resolved string, or an empty string
   * if resolution fails or the value is empty.
   */
  resolveDisplayName: (displayName: string) => string;
}

/**
 * React hook that provides a memoized function for resolving display names,
 * handling both plain text and i18n template patterns (e.g. `{{t(namespace:key)}}`).
 *
 * For i18n patterns, it checks whether the translation actually resolved to a
 * meaningful value (not just the raw key) before returning it.
 *
 * @param options - Options containing template literal handlers.
 * @returns Object containing the resolveDisplayName function.
 *
 * @example
 * ```tsx
 * const { t } = useTranslation();
 * const { resolveDisplayName } = useResolveDisplayName({ handlers: { t } });
 * resolveDisplayName('First Name');          // "First Name"
 * resolveDisplayName('{{t(custom:fname)}}'); // "First Name" (if translation exists)
 * resolveDisplayName('{{t(custom:fname)}}'); // "" (if translation key is missing)
 * ```
 */
export default function useResolveDisplayName({handlers}: UseResolveDisplayNameOptions): ResolveDisplayNameResult {
  const {resolve} = useTemplateLiteralResolver();

  const resolveDisplayName = useCallback(
    (displayName: string): string => {
      if (!displayName.trim()) return '';
      if (isI18nTemplatePattern(displayName)) {
        const resolved = resolve(displayName, handlers);
        const rawKey = resolve(displayName);
        // t() strips namespace prefix on fallback (e.g. "custom:key" -> "key"), check both forms
        const keyWithoutNs = rawKey?.includes(':') ? rawKey.split(':').pop() : rawKey;
        if (!resolved || resolved === rawKey || resolved === keyWithoutNs) return '';
        return resolved;
      }
      return displayName;
    },
    [handlers, resolve],
  );

  return {resolveDisplayName};
}
