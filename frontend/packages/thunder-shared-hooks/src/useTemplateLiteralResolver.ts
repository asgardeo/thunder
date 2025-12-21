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

import {useMemo} from 'react';

/**
 * Template literal types supported by the resolver
 *
 * @enum {string}
 */
export enum TemplateLiteralType {
  /** Translation template literal using t() function */
  TRANSLATION = 't',
  /** Unknown or unsupported template literal format */
  UNKNOWN = 'unknown',
}

/**
 * Result of parsing a template literal
 *
 * @interface TemplateLiteralResult
 */
export interface TemplateLiteralResult {
  /** The type of template literal that was detected */
  type: TemplateLiteralType;
  /** The extracted key from the template literal (e.g., "signin:heading" from "{{ t(signin:heading) }}") */
  key?: string;
  /** Reserved for future use - the resolved value after processing */
  resolvedValue?: string;
  /** The original template literal content before parsing */
  originalValue: string;
}

/**
 * Return type for useTemplateLiteralResolver hook
 *
 * @interface TemplateLiteralResolverResult
 */
export interface TemplateLiteralResolverResult {
  /** Function to resolve template literals and extract keys */
  resolve: (value?: string) => string | undefined;
}

/**
 * Parse a template literal content and extract type and key
 *
 * Supports parsing function calls like:
 * - `t(signin:heading)` -> extracts "signin:heading" as a translation key
 * - Future: `context(user:name)` -> extracts "user:name" as a context key
 *
 * @param content - The content inside template literal braces (without {{ }})
 * @returns Parsed template literal information including type, key, and original value
 *
 * @example
 * ```typescript
 * parseTemplateLiteral('t(signin:heading)')
 * // Returns: { type: 'TRANSLATION', key: 'signin:heading', originalValue: 't(signin:heading)' }
 * ```
 */
function parseTemplateLiteral(content: string): TemplateLiteralResult {
  const originalValue: string = content;
  const functionCallRegex = /^(\w+)\(([^)]+)\)$/;
  const match: RegExpExecArray | null = functionCallRegex.exec(content);

  if (!match) {
    return {
      type: TemplateLiteralType.UNKNOWN,
      originalValue,
    };
  }

  const [, functionName, key] = match;

  switch (functionName) {
    case 't':
      return {
        type: TemplateLiteralType.TRANSLATION,
        key: key.trim(),
        originalValue,
      };
    default:
      return {
        type: TemplateLiteralType.UNKNOWN,
        originalValue,
      };
  }
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
 * const output = resolve('plain text'); // undefined
 * ```
 */
export default function useTemplateLiteralResolver(): TemplateLiteralResolverResult {
  const resolve = useMemo(
    () => (value?: string): string | undefined => {
      if (!value || typeof value !== 'string') {
        return undefined;
      }

      // Check if the string contains template literals
      const templateLiteralRegex = /\{\{\s*([^}]+)\s*\}\}/;
      const match: RegExpExecArray | null = templateLiteralRegex.exec(value);

      if (!match) {
        return undefined;
      }

      const parsed: TemplateLiteralResult = parseTemplateLiteral(match[1].trim());

      return parsed.key;
    },
    []
  );

  return { resolve };
}
