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

// Strips CSS comments, control characters, and decodes unicode escapes to neutralize
// obfuscation techniques before applying sanitization rules.
function normalizeForSanitization(css: string): string {
  // Remove all CSS comments
  let normalized = css.replace(/\/\*[\s\S]*?\*\//g, '');

  // Remove null bytes and other control characters
  // eslint-disable-next-line no-control-regex
  normalized = normalized.replace(/[\x00-\x08\x0B\x0C\x0E-\x1F]/g, '');

  // Decode CSS unicode escapes (e.g., \65 -> e, \0065 -> e)
  normalized = normalized.replace(/\\([0-9a-fA-F]{1,6})\s?/g, (_match, hex: string) => {
    const codePoint = parseInt(hex, 16);

    if (codePoint > 0x10ffff) {
      return '';
    }

    return String.fromCodePoint(codePoint);
  });

  return normalized;
}

/**
 * Sanitizes inline CSS content by removing potentially dangerous constructs.
 * First normalizes the CSS to defeat obfuscation (comments, unicode escapes, null bytes),
 * then strips known dangerous patterns.
 *
 * @param css - The raw CSS string to sanitize
 * @returns The sanitized CSS string
 */
export function sanitizeCss(css: string): string {
  let sanitized = normalizeForSanitization(css);

  // Remove JavaScript expressions (IE legacy)
  sanitized = sanitized.replace(/expression\s*\([^)]*\)/gi, '');

  // Remove javascript: protocol references
  sanitized = sanitized.replace(/javascript\s*:/gi, '');

  // Remove url() with data: or javascript: protocols (handles whitespace/quote variations)
  sanitized = sanitized.replace(/url\s*\(\s*['"]?\s*(data|javascript)\s*:/gi, 'url(about:');

  // Remove @import rules (prevent loading external stylesheets)
  sanitized = sanitized.replace(/@import\s+[^;]+;/gi, '');

  // Remove @charset rules (prevent encoding-based attacks)
  sanitized = sanitized.replace(/@charset\s+[^;]+;/gi, '');

  // Remove -moz-binding (Firefox XBL injection) — anchored to property boundary
  sanitized = sanitized.replace(/(^|[{;]\s*)-moz-binding\s*:[^;]+;?/gi, '$1');

  // Remove behavior property (IE HTC injection) — anchored to avoid matching e.g. scroll-behavior
  sanitized = sanitized.replace(/(^|[{;]\s*)behavior\s*:[^;]+;?/gi, '$1');

  return sanitized;
}

/**
 * Validates that a stylesheet URL uses the https or http protocol.
 *
 * @param href - The URL to validate
 * @returns True if the URL is valid for stylesheet loading
 */
export function isValidStylesheetUrl(href: string): boolean {
  try {
    const url = new URL(href);
    return url.protocol === 'https:' || url.protocol === 'http:';
  } catch {
    return false;
  }
}

/**
 * Checks whether a stylesheet URL uses insecure http instead of https.
 *
 * @param href - The URL to check
 * @returns True if the URL uses http://
 */
export function isInsecureStylesheetUrl(href: string): boolean {
  try {
    const url = new URL(href);
    return url.protocol === 'http:';
  } catch {
    return false;
  }
}
