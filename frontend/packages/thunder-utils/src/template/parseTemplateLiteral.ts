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

/**
 * Regular expression to detect a template literal wrapped in double braces.
 * Matches patterns like `{{ t(key) }}`, `{{meta(key)}}`, etc.
 */
export const TEMPLATE_LITERAL_REGEX = /\{\{\s*([^}]+)\s*\}\}/;

/**
 * Regular expression to parse a function-call expression inside template braces.
 * Matches `funcName(arg)` and captures the function name and argument.
 */
export const FUNCTION_CALL_REGEX = /^(\w+)\(([^)]+)\)$/;

/**
 * Template literal types supported by the resolver.
 */
export enum TemplateLiteralType {
  /** Translation template literal using t() function */
  TRANSLATION = 't',
  /** Meta template literal using meta() function — resolves against flow/page meta data */
  META = 'meta',
  /** Unknown or unsupported template literal format */
  UNKNOWN = 'unknown',
}

/**
 * Result of parsing a template literal.
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
 * Map of handler functions keyed by TemplateLiteralType.
 * When provided to resolve(), the matching handler is called with the extracted key.
 *
 * Since TemplateLiteralType.TRANSLATION = 't', you can pass `{ t }` directly from useTranslation().
 *
 * @example
 * ```typescript
 * const { t } = useTranslation();
 * resolve('{{ t(signin:heading) }}', { t }); // calls t('signin:heading')
 * ```
 */
export type TemplateLiteralHandlers = Partial<Record<TemplateLiteralType, (key: string) => string>>;

/**
 * Parse a template literal content string and extract its type and key.
 *
 * Supports function-call expressions like:
 * - `t(signin:heading)` -> type TRANSLATION, key "signin:heading"
 *
 * @param content - The content inside the template literal braces (without `{{ }}`).
 * @returns Parsed template literal information including type, key, and original value.
 *
 * @example
 * ```typescript
 * parseTemplateLiteral('t(signin:heading)')
 * // Returns: { type: TemplateLiteralType.TRANSLATION, key: 'signin:heading', originalValue: 't(signin:heading)' }
 * ```
 */
export default function parseTemplateLiteral(content: string): TemplateLiteralResult {
  const originalValue: string = content;
  const match: RegExpExecArray | null = FUNCTION_CALL_REGEX.exec(content);

  if (!match) {
    return {type: TemplateLiteralType.UNKNOWN, originalValue};
  }

  const [, functionName, key] = match;

  const cleanKey = key.trim().replace(/^['"]|['"]$/g, '');

  switch (functionName as TemplateLiteralType) {
    case TemplateLiteralType.TRANSLATION:
      return {type: TemplateLiteralType.TRANSLATION, key: cleanKey, originalValue};
    case TemplateLiteralType.META:
      return {type: TemplateLiteralType.META, key: cleanKey, originalValue};
    default:
      return {type: TemplateLiteralType.UNKNOWN, originalValue};
  }
}
