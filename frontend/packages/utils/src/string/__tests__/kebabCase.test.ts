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
import kebabCase from '../kebabCase';

describe('kebabCase', () => {
  it('converts space-separated words', () => {
    expect(kebabCase('Acrylic Orange')).toBe('acrylic-orange');
  });

  it('converts camelCase', () => {
    expect(kebabCase('fooBar')).toBe('foo-bar');
  });

  it('converts PascalCase', () => {
    expect(kebabCase('FooBar')).toBe('foo-bar');
  });

  it('converts ALL_CAPS_SNAKE_CASE', () => {
    expect(kebabCase('FOO_BAR')).toBe('foo-bar');
  });

  it('converts strings with existing hyphens', () => {
    expect(kebabCase('high-contrast')).toBe('high-contrast');
  });

  it('strips special characters', () => {
    expect(kebabCase('hello! world?')).toBe('hello-world');
  });

  it('handles leading and trailing whitespace', () => {
    expect(kebabCase('  hello world ')).toBe('hello-world');
  });

  it('collapses multiple separators', () => {
    expect(kebabCase('pale  indigo')).toBe('pale-indigo');
  });

  it('handles an empty string', () => {
    expect(kebabCase('')).toBe('');
  });

  it('handles a single word', () => {
    expect(kebabCase('Classic')).toBe('classic');
  });

  it('handles XMLParser-style acronym transitions', () => {
    expect(kebabCase('XMLParser')).toBe('xml-parser');
  });
});
