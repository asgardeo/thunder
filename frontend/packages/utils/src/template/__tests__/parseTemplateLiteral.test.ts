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

import {describe, expect, it} from 'vitest';
import parseTemplateLiteral, {
  TEMPLATE_LITERAL_REGEX,
  FUNCTION_CALL_REGEX,
  TemplateLiteralType,
} from '../parseTemplateLiteral';

describe('parseTemplateLiteral', () => {
  describe('translation (t) templates', () => {
    it('should parse a simple translation key', () => {
      const result = parseTemplateLiteral('t(signin:heading)');

      expect(result.type).toBe(TemplateLiteralType.TRANSLATION);
      expect(result.key).toBe('signin:heading');
      expect(result.originalValue).toBe('t(signin:heading)');
    });

    it('should strip surrounding single quotes from the key', () => {
      const result = parseTemplateLiteral("t('signin:heading')");

      expect(result.type).toBe(TemplateLiteralType.TRANSLATION);
      expect(result.key).toBe('signin:heading');
    });

    it('should strip surrounding double quotes from the key', () => {
      const result = parseTemplateLiteral('t("signin:heading")');

      expect(result.type).toBe(TemplateLiteralType.TRANSLATION);
      expect(result.key).toBe('signin:heading');
    });

    it('should trim whitespace from the key', () => {
      const result = parseTemplateLiteral('t( signin:heading )');

      expect(result.type).toBe(TemplateLiteralType.TRANSLATION);
      expect(result.key).toBe('signin:heading');
    });

    it('should handle a key without a namespace', () => {
      const result = parseTemplateLiteral('t(heading)');

      expect(result.type).toBe(TemplateLiteralType.TRANSLATION);
      expect(result.key).toBe('heading');
    });

    it('should handle a key with nested dots', () => {
      const result = parseTemplateLiteral('t(common:button.submit)');

      expect(result.type).toBe(TemplateLiteralType.TRANSLATION);
      expect(result.key).toBe('common:button.submit');
    });
  });

  describe('meta templates', () => {
    it('should parse a simple meta key', () => {
      const result = parseTemplateLiteral('meta(application.name)');

      expect(result.type).toBe(TemplateLiteralType.META);
      expect(result.key).toBe('application.name');
      expect(result.originalValue).toBe('meta(application.name)');
    });

    it('should parse a boolean meta key', () => {
      const result = parseTemplateLiteral('meta(is_registration_flow_enabled)');

      expect(result.type).toBe(TemplateLiteralType.META);
      expect(result.key).toBe('is_registration_flow_enabled');
    });

    it('should strip surrounding quotes from a meta key', () => {
      const result = parseTemplateLiteral("meta('application.sign_up_url')");

      expect(result.type).toBe(TemplateLiteralType.META);
      expect(result.key).toBe('application.sign_up_url');
    });
  });

  describe('unknown / unrecognised templates', () => {
    it('should return UNKNOWN type for a non-function-call format', () => {
      const result = parseTemplateLiteral('not-a-function-call');

      expect(result.type).toBe(TemplateLiteralType.UNKNOWN);
      expect(result.key).toBeUndefined();
      expect(result.originalValue).toBe('not-a-function-call');
    });

    it('should return UNKNOWN type for an unsupported function name', () => {
      const result = parseTemplateLiteral('custom(key)');

      expect(result.type).toBe(TemplateLiteralType.UNKNOWN);
      expect(result.key).toBeUndefined();
    });

    it('should return UNKNOWN type for an empty string', () => {
      const result = parseTemplateLiteral('');

      expect(result.type).toBe(TemplateLiteralType.UNKNOWN);
    });
  });
});

describe('TEMPLATE_LITERAL_REGEX', () => {
  it('should match a double-brace template expression', () => {
    expect(TEMPLATE_LITERAL_REGEX.test('{{t(signin:heading)}}')).toBe(true);
  });

  it('should match a template with whitespace inside braces', () => {
    expect(TEMPLATE_LITERAL_REGEX.test('{{ meta(application.name) }}')).toBe(true);
  });

  it('should not match a plain string without braces', () => {
    expect(TEMPLATE_LITERAL_REGEX.test('hello world')).toBe(false);
  });
});

describe('FUNCTION_CALL_REGEX', () => {
  it('should match and capture a function call', () => {
    const match = FUNCTION_CALL_REGEX.exec('t(signin:heading)');

    expect(match).not.toBeNull();
    expect(match?.[1]).toBe('t');
    expect(match?.[2]).toBe('signin:heading');
  });

  it('should match a meta function call', () => {
    const match = FUNCTION_CALL_REGEX.exec('meta(application.name)');

    expect(match).not.toBeNull();
    expect(match?.[1]).toBe('meta');
    expect(match?.[2]).toBe('application.name');
  });

  it('should return null for a non-function-call string', () => {
    expect(FUNCTION_CALL_REGEX.exec('not-a-function')).toBeNull();
  });
});
