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
import isMetaTemplatePattern, {META_PATTERN, META_KEY_PATTERN} from '../isMetaTemplatePattern';

describe('isMetaTemplatePattern', () => {
  describe('valid meta patterns', () => {
    it('should return true for a simple meta key', () => {
      expect(isMetaTemplatePattern('{{meta(application.name)}}')).toBe(true);
    });

    it('should return true for a key with nested dots', () => {
      expect(isMetaTemplatePattern('{{meta(ou.description)}}')).toBe(true);
    });

    it('should return true for a boolean meta key', () => {
      expect(isMetaTemplatePattern('{{meta(is_registration_flow_enabled)}}')).toBe(true);
    });

    it('should return true when the value has leading and trailing whitespace', () => {
      expect(isMetaTemplatePattern('  {{meta(application.name)}}  ')).toBe(true);
    });
  });

  describe('invalid meta patterns', () => {
    it('should return false for a plain string', () => {
      expect(isMetaTemplatePattern('hello world')).toBe(false);
    });

    it('should return false for an i18n template', () => {
      expect(isMetaTemplatePattern('{{t(signin:heading)}}')).toBe(false);
    });

    it('should return false when the template is embedded in other text', () => {
      expect(isMetaTemplatePattern('Visit {{meta(application.sign_up_url)}} today')).toBe(false);
    });

    it('should return false for missing closing braces', () => {
      expect(isMetaTemplatePattern('{{meta(application.name)')).toBe(false);
    });

    it('should return false for an empty string', () => {
      expect(isMetaTemplatePattern('')).toBe(false);
    });

    it('should return false for empty parentheses', () => {
      expect(isMetaTemplatePattern('{{meta()}}')).toBe(false);
    });
  });
});

describe('META_PATTERN', () => {
  it('should match a full meta template string', () => {
    expect(META_PATTERN.test('{{meta(application.name)}}')).toBe(true);
  });

  it('should not match a partial string', () => {
    expect(META_PATTERN.test('prefix{{meta(application.name)}}')).toBe(false);
  });
});

describe('META_KEY_PATTERN', () => {
  it('should capture the key from a meta template', () => {
    const match = META_KEY_PATTERN.exec('{{meta(application.sign_up_url)}}');

    expect(match).not.toBeNull();
    expect(match?.[1]).toBe('application.sign_up_url');
  });

  it('should return null for a non-matching string', () => {
    expect(META_KEY_PATTERN.exec('hello')).toBeNull();
  });
});
