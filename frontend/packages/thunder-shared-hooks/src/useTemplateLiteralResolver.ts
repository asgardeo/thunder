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

import {
  TEMPLATE_LITERAL_REGEX,
  TemplateLiteralHandlers,
  TemplateLiteralResult,
  parseTemplateLiteral,
} from '@thunder/utils';
import {useMemo} from 'react';

/**
 * Return type for useTemplateLiteralResolver hook
 *
 * @interface TemplateLiteralResolverResult
 */
export interface TemplateLiteralResolverResult {
  /** Function to resolve template literals. If handlers are provided, calls the matching handler with the extracted key. */
  resolve: (value?: string, handlers?: TemplateLiteralHandlers) => string | undefined;
  /**
   * Like `resolve`, but performs a global substitution — replaces every `{{type(key)}}` occurrence
   * in the string with its handler result. Unresolved templates are left as-is.
   * Use this when template literals are embedded inside a larger string (e.g. HTML content).
   */
  resolveAll: (value?: string, handlers?: TemplateLiteralHandlers) => string | undefined;
}

/**
 * React hook to resolve template literals in strings
 *
 * This hook returns a resolve function that can parse strings containing template literals
 * wrapped in double braces and extract the keys for use with translation functions.
 *
 * Supported patterns:
 * - `{{ t(signin:heading) }}` -> extracts "signin:heading" for translation
 * - `{{ context(user:name) }}` -> extracts "user:name" for context resolution (future)
 *
 * @returns Object containing the resolve function
 *
 * @example
 * ```typescript
 * const { resolve } = useTemplateLiteralResolver();
 * const output = resolve('{{ t(signin:heading) }}'); // "signin:heading"
 *
 * const { t } = useTranslation();
 * const translatedText = t(output); // Use with your translation function
 * ```
 *
 * @example
 * ```typescript
 * // For non-template strings
 * const { resolve } = useTemplateLiteralResolver();
 * const output = resolve('plain text'); // 'plain text'
 * ```
 */
export default function useTemplateLiteralResolver(): TemplateLiteralResolverResult {
  const resolve = useMemo(
    () =>
      (value?: string, handlers?: TemplateLiteralHandlers): string | undefined => {
        if (!value || typeof value !== 'string') {
          return undefined;
        }

        const match: RegExpExecArray | null = TEMPLATE_LITERAL_REGEX.exec(value);

        if (!match) {
          return value;
        }

        const parsed: TemplateLiteralResult = parseTemplateLiteral(match[1].trim());

        if (parsed.key && handlers?.[parsed.type]) {
          return handlers[parsed.type]!(parsed.key);
        }

        return parsed.key ?? value;
      },
    [],
  );

  const resolveAll = useMemo(
    () =>
      (value?: string, handlers?: TemplateLiteralHandlers): string | undefined => {
        if (!value || typeof value !== 'string') {
          return undefined;
        }

        return value.replace(/\{\{\s*([^}]+)\s*\}\}/g, (fullMatch: string, inner: string) => {
          const parsed: TemplateLiteralResult = parseTemplateLiteral(inner.trim());

          if (parsed.key && handlers?.[parsed.type]) {
            return handlers[parsed.type]!(parsed.key);
          }

          return fullMatch;
        });
      },
    [],
  );

  return {resolve, resolveAll};
}
