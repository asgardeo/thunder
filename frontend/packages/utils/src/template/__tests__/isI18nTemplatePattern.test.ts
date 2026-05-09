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
import isI18nTemplatePattern, {I18N_PATTERN, I18N_KEY_PATTERN} from '../isI18nTemplatePattern';

describe('isI18nTemplatePattern', () => {
  describe('valid i18n patterns', () => {
    it('should return true for a simple i18n key', () => {
      expect(isI18nTemplatePattern('{{t(signin:heading)}}')).toBe(true);
    });

    it('should return true for a key without a namespace', () => {
      expect(isI18nTemplatePattern('{{t(heading)}}')).toBe(true);
    });

    it('should return true for a key with nested dots', () => {
      expect(isI18nTemplatePattern('{{t(common:button.submit)}}')).toBe(true);
    });

    it('should return true when the value has leading and trailing whitespace', () => {
      expect(isI18nTemplatePattern('  {{t(signin:heading)}}  ')).toBe(true);
    });
  });

  describe('invalid i18n patterns', () => {
    it('should return false for a plain string', () => {
      expect(isI18nTemplatePattern('hello world')).toBe(false);
    });

    it('should return false for a meta template', () => {
      expect(isI18nTemplatePattern('{{meta(application.name)}}')).toBe(false);
    });

    it('should return false when the template is embedded in other text', () => {
      expect(isI18nTemplatePattern('Click {{t(signin:heading)}} here')).toBe(false);
    });

    it('should return false for missing closing braces', () => {
      expect(isI18nTemplatePattern('{{t(signin:heading)')).toBe(false);
    });

    it('should return false for an empty string', () => {
      expect(isI18nTemplatePattern('')).toBe(false);
    });

    it('should return false for empty parentheses', () => {
      expect(isI18nTemplatePattern('{{t()}}')).toBe(false);
    });
  });
});

describe('I18N_PATTERN', () => {
  it('should match a full i18n template string', () => {
    expect(I18N_PATTERN.test('{{t(signin:heading)}}')).toBe(true);
  });

  it('should not match a partial string', () => {
    expect(I18N_PATTERN.test('prefix{{t(signin:heading)}}')).toBe(false);
  });
});

describe('I18N_KEY_PATTERN', () => {
  it('should capture the key from an i18n template', () => {
    const match = I18N_KEY_PATTERN.exec('{{t(signin:heading)}}');

    expect(match).not.toBeNull();
    expect(match?.[1]).toBe('signin:heading');
  });

  it('should return null for a non-matching string', () => {
    expect(I18N_KEY_PATTERN.exec('hello')).toBeNull();
  });
});
