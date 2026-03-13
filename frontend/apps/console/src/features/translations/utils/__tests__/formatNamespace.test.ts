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
import formatNamespace from '../formatNamespace';

describe('formatNamespace', () => {
  describe('camelCase inputs', () => {
    it('formats a camelCase string by inserting spaces before capitals', () => {
      expect(formatNamespace('userProfileSettings')).toBe('User Profile Settings');
    });

    it('formats a single-word camelCase string by capitalizing the first letter', () => {
      expect(formatNamespace('users')).toBe('Users');
    });

    it('formats a two-word camelCase string correctly', () => {
      expect(formatNamespace('signIn')).toBe('Sign In');
    });

    it('formats a multi-word camelCase string correctly', () => {
      expect(formatNamespace('myApplicationSettings')).toBe('My Application Settings');
    });
  });

  describe('PascalCase inputs', () => {
    it('formats a PascalCase string correctly', () => {
      expect(formatNamespace('AdminPanel')).toBe('Admin Panel');
    });

    it('formats a single PascalCase word correctly', () => {
      expect(formatNamespace('Users')).toBe('Users');
    });

    it('formats multi-word PascalCase correctly', () => {
      expect(formatNamespace('UserProfileSettings')).toBe('User Profile Settings');
    });
  });

  describe('edge cases', () => {
    it('returns empty string for empty input', () => {
      expect(formatNamespace('')).toBe('');
    });

    it('formats a lowercase single word by capitalizing it', () => {
      expect(formatNamespace('common')).toBe('Common');
    });

    it('handles consecutive capital letters correctly', () => {
      expect(formatNamespace('myAPIConfig')).toBe('My A P I Config');
    });

    it('returns a trimmed result when input has leading/trailing spaces', () => {
      const result = formatNamespace('  myNamespace  ');
      expect(result).not.toMatch(/^\s|\s$/);
    });
  });

  describe('real namespace examples', () => {
    it('formats "applications" correctly', () => {
      expect(formatNamespace('applications')).toBe('Applications');
    });

    it('formats "translations" correctly', () => {
      expect(formatNamespace('translations')).toBe('Translations');
    });

    it('formats "loginFlow" correctly', () => {
      expect(formatNamespace('loginFlow')).toBe('Login Flow');
    });

    it('formats "identityProviders" correctly', () => {
      expect(formatNamespace('identityProviders')).toBe('Identity Providers');
    });
  });
});
