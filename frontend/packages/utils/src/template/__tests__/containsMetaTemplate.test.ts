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
import containsMetaTemplate, {replaceMetaTemplate} from '../containsMetaTemplate';

describe('containsMetaTemplate', () => {
  describe('returns true when meta template is present', () => {
    it('should detect an exact meta template', () => {
      expect(containsMetaTemplate('{{meta(application.sign_up_url)}}', 'application.sign_up_url')).toBe(true);
    });

    it('should detect a meta template embedded in an HTML string', () => {
      expect(
        containsMetaTemplate('<a href="{{meta(application.sign_up_url)}}">Sign up</a>', 'application.sign_up_url'),
      ).toBe(true);
    });

    it('should allow whitespace around double braces', () => {
      expect(containsMetaTemplate('{{ meta(application.sign_up_url) }}', 'application.sign_up_url')).toBe(true);
    });

    it('should allow whitespace inside the braces only', () => {
      expect(containsMetaTemplate('{{  meta(application.sign_up_url)  }}', 'application.sign_up_url')).toBe(true);
    });

    it('should detect the template when it appears in the middle of text', () => {
      expect(containsMetaTemplate('Hello {{meta(ou.name)}} world', 'ou.name')).toBe(true);
    });

    it('should find the template among multiple tokens', () => {
      expect(
        containsMetaTemplate('{{meta(ou.name)}} and {{meta(application.sign_up_url)}}', 'application.sign_up_url'),
      ).toBe(true);
    });
  });

  describe('returns false when meta template is absent', () => {
    it('should return false for a plain URL string', () => {
      expect(containsMetaTemplate('https://example.com/signup', 'application.sign_up_url')).toBe(false);
    });

    it('should return false when a different key is present', () => {
      expect(containsMetaTemplate('{{meta(ou.name)}}', 'application.sign_up_url')).toBe(false);
    });

    it('should return false for an empty string', () => {
      expect(containsMetaTemplate('', 'application.sign_up_url')).toBe(false);
    });

    it('should return false for a partial key match', () => {
      // "sign_up_url" alone should not match "application.sign_up_url"
      expect(containsMetaTemplate('{{meta(sign_up_url)}}', 'application.sign_up_url')).toBe(false);
    });

    it('should return false when the meta() call is missing', () => {
      expect(containsMetaTemplate('{{application.sign_up_url}}', 'application.sign_up_url')).toBe(false);
    });
  });

  describe('special regex characters in key', () => {
    it('should escape dots in the key so they match literally', () => {
      // A dot in the key should not act as a regex wildcard
      expect(containsMetaTemplate('{{meta(applicationXsignUpUrl)}}', 'application.sign_up_url')).toBe(false);
    });
  });
});

describe('replaceMetaTemplate', () => {
  it('should replace a single meta template occurrence', () => {
    const result = replaceMetaTemplate(
      '{{meta(application.sign_up_url)}}',
      'application.sign_up_url',
      'https://example.com/signup',
    );

    expect(result).toBe('https://example.com/signup');
  });

  it('should replace a meta template embedded inside an HTML string', () => {
    const result = replaceMetaTemplate(
      '<a href="{{meta(application.sign_up_url)}}">Sign up</a>',
      'application.sign_up_url',
      'https://example.com/signup',
    );

    expect(result).toBe('<a href="https://example.com/signup">Sign up</a>');
  });

  it('should replace all occurrences globally', () => {
    const result = replaceMetaTemplate('{{meta(ou.name)}} — {{meta(ou.name)}}', 'ou.name', 'Acme');

    expect(result).toBe('Acme — Acme');
  });

  it('should replace when there is whitespace around braces', () => {
    const result = replaceMetaTemplate('{{ meta(application.sign_up_url) }}', 'application.sign_up_url', '/signup');

    expect(result).toBe('/signup');
  });

  it('should leave other tokens untouched', () => {
    const result = replaceMetaTemplate(
      '{{meta(ou.name)}} visits {{meta(application.sign_up_url)}}',
      'application.sign_up_url',
      '/signup',
    );

    expect(result).toBe('{{meta(ou.name)}} visits /signup');
  });

  it('should return the original string unchanged when the key is not found', () => {
    const original = 'No template here';

    expect(replaceMetaTemplate(original, 'application.sign_up_url', '/signup')).toBe(original);
  });

  it('should return an empty string unchanged', () => {
    expect(replaceMetaTemplate('', 'application.sign_up_url', '/signup')).toBe('');
  });
});
