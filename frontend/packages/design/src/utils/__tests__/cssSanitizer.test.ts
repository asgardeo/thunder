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
import {sanitizeCss, isValidStylesheetUrl, isInsecureStylesheetUrl} from '../cssSanitizer';

describe('sanitizeCss', () => {
  it('should pass through safe CSS unchanged', () => {
    const css = '.my-class { color: red; font-size: 14px; }';
    expect(sanitizeCss(css)).toBe(css);
  });

  it('should remove expression() calls', () => {
    expect(sanitizeCss('width: expression(document.body.clientWidth)')).toBe('width: ');
  });

  it('should remove javascript: protocol references', () => {
    expect(sanitizeCss('color: javascript:alert(1)')).toBe('color: alert(1)');
  });

  it('should neutralize javascript: inside url()', () => {
    const result = sanitizeCss('background: url(javascript:alert(1))');
    expect(result).not.toContain('javascript');
  });

  it('should remove data: protocol in url()', () => {
    expect(sanitizeCss('background: url(data:text/html,<script>alert(1)</script>)')).toBe(
      'background: url(about:text/html,<script>alert(1)</script>)',
    );
  });

  it('should remove @import rules', () => {
    expect(sanitizeCss('@import url("https://evil.com/steal.css"); .foo { color: red; }')).toBe(
      ' .foo { color: red; }',
    );
  });

  it('should remove @charset rules', () => {
    expect(sanitizeCss('@charset "UTF-7"; .foo { color: red; }')).toBe(' .foo { color: red; }');
  });

  it('should remove -moz-binding property', () => {
    expect(sanitizeCss('.foo { -moz-binding: url("http://evil.com/xbl"); }')).toBe('.foo {  }');
  });

  it('should remove behavior property', () => {
    expect(sanitizeCss('.foo { behavior: url(script.htc); }')).toBe('.foo {  }');
  });

  // Obfuscation bypass tests
  it('should handle expression() obfuscated with CSS comments', () => {
    const obfuscated = 'width: exp/*comment*/ression(document.body.clientWidth)';
    const result = sanitizeCss(obfuscated);
    expect(result).not.toContain('expression');
  });

  it('should handle unicode-escaped expression()', () => {
    // \\65 is unicode escape for 'e', so \\65xpression -> expression
    const obfuscated = 'width: \\65 xpression(document.body.clientWidth)';
    const result = sanitizeCss(obfuscated);
    expect(result).not.toContain('expression');
  });

  it('should strip null bytes used for obfuscation', () => {
    const withNullBytes = 'expres\x00sion(alert(1))';
    const result = sanitizeCss(withNullBytes);
    expect(result).not.toContain('expression');
  });

  it('should handle multiple dangerous patterns in one input', () => {
    const css = '@import url("evil.css"); .foo { behavior: url(x.htc); -moz-binding: url(y); }';
    const result = sanitizeCss(css);
    expect(result).not.toContain('@import');
    expect(result).not.toContain('behavior');
    expect(result).not.toContain('-moz-binding');
  });
});

describe('isValidStylesheetUrl', () => {
  it('should accept https URLs', () => {
    expect(isValidStylesheetUrl('https://cdn.example.com/styles.css')).toBe(true);
  });

  it('should accept http URLs', () => {
    expect(isValidStylesheetUrl('http://cdn.example.com/styles.css')).toBe(true);
  });

  it('should reject javascript: URLs', () => {
    expect(isValidStylesheetUrl('javascript:alert(1)')).toBe(false);
  });

  it('should reject data: URLs', () => {
    expect(isValidStylesheetUrl('data:text/css,body{color:red}')).toBe(false);
  });

  it('should reject invalid URLs', () => {
    expect(isValidStylesheetUrl('not a url')).toBe(false);
  });

  it('should reject empty strings', () => {
    expect(isValidStylesheetUrl('')).toBe(false);
  });
});

describe('isInsecureStylesheetUrl', () => {
  it('should return true for http URLs', () => {
    expect(isInsecureStylesheetUrl('http://cdn.example.com/styles.css')).toBe(true);
  });

  it('should return false for https URLs', () => {
    expect(isInsecureStylesheetUrl('https://cdn.example.com/styles.css')).toBe(false);
  });

  it('should return false for invalid URLs', () => {
    expect(isInsecureStylesheetUrl('not a url')).toBe(false);
  });
});
