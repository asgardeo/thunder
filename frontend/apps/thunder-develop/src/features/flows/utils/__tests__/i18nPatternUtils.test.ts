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

import {describe, expect, it, vi, beforeEach} from 'vitest';
import {isI18nPattern, extractI18nKey, resolveI18nValue} from '../i18nPatternUtils';

describe('i18nPatternUtils', () => {
  describe('isI18nPattern', () => {
    describe('Valid Patterns', () => {
      it('should return true for valid i18n pattern', () => {
        expect(isI18nPattern('{{t(login.title)}}')).toBe(true);
      });

      it('should return true for pattern with dots in key', () => {
        expect(isI18nPattern('{{t(common.buttons.submit)}}')).toBe(true);
      });

      it('should return true for pattern with underscores', () => {
        expect(isI18nPattern('{{t(login_page_title)}}')).toBe(true);
      });

      it('should return true for pattern with dashes', () => {
        expect(isI18nPattern('{{t(login-page-title)}}')).toBe(true);
      });

      it('should return true for single word key', () => {
        expect(isI18nPattern('{{t(title)}}')).toBe(true);
      });

      it('should return true for pattern with whitespace and stripHtml option', () => {
        expect(isI18nPattern('  {{t(key)}}  ')).toBe(true);
      });
    });

    describe('Invalid Patterns', () => {
      it('should return false for undefined value', () => {
        expect(isI18nPattern(undefined)).toBe(false);
      });

      it('should return false for empty string', () => {
        expect(isI18nPattern('')).toBe(false);
      });

      it('should return false for plain text', () => {
        expect(isI18nPattern('Hello World')).toBe(false);
      });

      it('should return false for incomplete pattern - missing closing braces', () => {
        expect(isI18nPattern('{{t(key)')).toBe(false);
      });

      it('should return false for incomplete pattern - missing t function', () => {
        expect(isI18nPattern('{{key}}')).toBe(false);
      });

      it('should return false for incomplete pattern - missing parentheses', () => {
        expect(isI18nPattern('{{t key}}')).toBe(false);
      });

      it('should return false for pattern with text before', () => {
        expect(isI18nPattern('prefix {{t(key)}}')).toBe(false);
      });

      it('should return false for pattern with text after', () => {
        expect(isI18nPattern('{{t(key)}} suffix')).toBe(false);
      });

      it('should return false for empty key', () => {
        expect(isI18nPattern('{{t()}}')).toBe(false);
      });
    });

    describe('HTML Stripping', () => {
      it('should handle HTML wrapped pattern with stripHtml=true', () => {
        expect(isI18nPattern('<p>{{t(key)}}</p>', true)).toBe(true);
      });

      it('should handle nested HTML with stripHtml=true', () => {
        expect(isI18nPattern('<div><span>{{t(key)}}</span></div>', true)).toBe(true);
      });

      it('should fail HTML wrapped pattern without stripHtml', () => {
        expect(isI18nPattern('<p>{{t(key)}}</p>')).toBe(false);
      });

      it('should handle empty HTML tags', () => {
        expect(isI18nPattern('<p></p>{{t(key)}}<br/>', true)).toBe(true);
      });
    });
  });

  describe('extractI18nKey', () => {
    describe('Successful Extraction', () => {
      it('should extract key from valid pattern', () => {
        expect(extractI18nKey('{{t(login.title)}}')).toBe('login.title');
      });

      it('should extract key with dots', () => {
        expect(extractI18nKey('{{t(common.buttons.submit)}}')).toBe('common.buttons.submit');
      });

      it('should extract key with underscores', () => {
        expect(extractI18nKey('{{t(login_page_title)}}')).toBe('login_page_title');
      });

      it('should extract key with dashes', () => {
        expect(extractI18nKey('{{t(login-page-title)}}')).toBe('login-page-title');
      });

      it('should extract single word key', () => {
        expect(extractI18nKey('{{t(title)}}')).toBe('title');
      });

      it('should handle whitespace around pattern', () => {
        expect(extractI18nKey('  {{t(key)}}  ')).toBe('key');
      });
    });

    describe('Failed Extraction', () => {
      it('should return null for undefined value', () => {
        expect(extractI18nKey(undefined)).toBeNull();
      });

      it('should return null for empty string', () => {
        expect(extractI18nKey('')).toBeNull();
      });

      it('should return null for plain text', () => {
        expect(extractI18nKey('Hello World')).toBeNull();
      });

      it('should return null for invalid pattern', () => {
        expect(extractI18nKey('{{key}}')).toBeNull();
      });

      it('should return null for partial pattern', () => {
        expect(extractI18nKey('{{t(key)')).toBeNull();
      });
    });

    describe('HTML Stripping', () => {
      it('should extract key from HTML wrapped pattern with stripHtml=true', () => {
        expect(extractI18nKey('<p>{{t(key)}}</p>', true)).toBe('key');
      });

      it('should extract key from nested HTML with stripHtml=true', () => {
        expect(extractI18nKey('<div><span>{{t(nested.key)}}</span></div>', true)).toBe('nested.key');
      });

      it('should return null for HTML wrapped pattern without stripHtml', () => {
        expect(extractI18nKey('<p>{{t(key)}}</p>')).toBeNull();
      });
    });
  });

  describe('resolveI18nValue', () => {
    const mockTranslate = vi.fn((key: string) => `Translated: ${key}`);

    beforeEach(() => {
      mockTranslate.mockClear();
    });

    describe('Successful Resolution', () => {
      it('should resolve i18n pattern using translate function', () => {
        const result = resolveI18nValue('{{t(login.title)}}', mockTranslate);

        expect(mockTranslate).toHaveBeenCalledWith('login.title');
        expect(result).toBe('Translated: login.title');
      });

      it('should pass the correct key to translate function', () => {
        resolveI18nValue('{{t(common.buttons.submit)}}', mockTranslate);

        expect(mockTranslate).toHaveBeenCalledWith('common.buttons.submit');
      });

      it('should handle complex keys', () => {
        const result = resolveI18nValue('{{t(auth.login_form.title-text)}}', mockTranslate);

        expect(result).toBe('Translated: auth.login_form.title-text');
      });
    });

    describe('Failed Resolution', () => {
      it('should return empty string for undefined value', () => {
        const result = resolveI18nValue(undefined, mockTranslate);

        expect(result).toBe('');
        expect(mockTranslate).not.toHaveBeenCalled();
      });

      it('should return empty string for empty string value', () => {
        const result = resolveI18nValue('', mockTranslate);

        expect(result).toBe('');
        expect(mockTranslate).not.toHaveBeenCalled();
      });

      it('should return empty string for plain text', () => {
        const result = resolveI18nValue('Hello World', mockTranslate);

        expect(result).toBe('');
        expect(mockTranslate).not.toHaveBeenCalled();
      });

      it('should return empty string for invalid pattern', () => {
        const result = resolveI18nValue('{{key}}', mockTranslate);

        expect(result).toBe('');
        expect(mockTranslate).not.toHaveBeenCalled();
      });
    });

    describe('HTML Stripping', () => {
      it('should resolve HTML wrapped pattern with stripHtml=true', () => {
        const result = resolveI18nValue('<p>{{t(key)}}</p>', mockTranslate, true);

        expect(mockTranslate).toHaveBeenCalledWith('key');
        expect(result).toBe('Translated: key');
      });

      it('should return empty string for HTML without stripHtml', () => {
        const result = resolveI18nValue('<p>{{t(key)}}</p>', mockTranslate, false);

        expect(result).toBe('');
        expect(mockTranslate).not.toHaveBeenCalled();
      });
    });

    describe('Translation Function Behavior', () => {
      it('should call translate function exactly once', () => {
        resolveI18nValue('{{t(key)}}', mockTranslate);

        expect(mockTranslate).toHaveBeenCalledTimes(1);
      });

      it('should return the exact value from translate function', () => {
        const customTranslate = vi.fn(() => 'Custom Translation');
        const result = resolveI18nValue('{{t(key)}}', customTranslate);

        expect(result).toBe('Custom Translation');
      });

      it('should handle translate function returning empty string', () => {
        const emptyTranslate = vi.fn(() => '');
        const result = resolveI18nValue('{{t(key)}}', emptyTranslate);

        expect(result).toBe('');
      });
    });
  });
});
