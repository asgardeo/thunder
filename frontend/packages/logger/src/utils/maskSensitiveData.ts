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
 * Mask sensitive data in a string.
 * Replaces characters with asterisks, leaving first and last characters visible.
 * @param value - The value to mask
 * @param visibleChars - Number of characters to leave visible at start and end
 * @returns Masked string
 */
export function maskString(value: string, visibleChars = 2): string {
  if (value.length <= visibleChars * 2) {
    return '*'.repeat(value.length);
  }

  const start = value.substring(0, visibleChars);
  const end = value.substring(value.length - visibleChars);
  const maskLength = value.length - visibleChars * 2;

  return `${start}${'*'.repeat(maskLength)}${end}`;
}

/**
 * Common sensitive field names to mask in logs.
 */
const SENSITIVE_FIELD_NAMES = new Set([
  'password',
  'passwd',
  'pwd',
  'secret',
  'token',
  'accesstoken',
  'access_token',
  'refreshtoken',
  'refresh_token',
  'apikey',
  'api_key',
  'auth',
  'authorization',
  'cookie',
  'session',
  'sessionid',
  'session_id',
  'ssn',
  'credit_card',
  'creditcard',
  'cvv',
  'pin',
]);

/**
 * Recursively mask sensitive data in an object.
 * @param obj - The object to process
 * @returns New object with sensitive fields masked
 */
export default function maskSensitiveData(obj: unknown): unknown {
  if (obj === null || obj === undefined) {
    return obj;
  }

  if (typeof obj !== 'object') {
    return obj;
  }

  if (obj instanceof Error) {
    return {
      name: obj.name,
      message: obj.message,
      stack: obj.stack,
    };
  }

  if (Array.isArray(obj)) {
    return obj.map((item) => maskSensitiveData(item));
  }

  const masked: Record<string, unknown> = {};

  Object.entries(obj).forEach(([key, value]) => {
    const lowerKey = key.toLowerCase();

    if (SENSITIVE_FIELD_NAMES.has(lowerKey)) {
      masked[key] = typeof value === 'string' ? maskString(value) : '[REDACTED]';
    } else if (typeof value === 'object' && value !== null) {
      masked[key] = maskSensitiveData(value);
    } else {
      masked[key] = value;
    }
  });

  return masked;
}
