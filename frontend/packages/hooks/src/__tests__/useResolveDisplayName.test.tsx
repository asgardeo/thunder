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

import {renderHook} from '@thunderid/test-utils';
import {describe, it, expect, vi} from 'vitest';
import useResolveDisplayName from '../useResolveDisplayName';

describe('useResolveDisplayName', () => {
  describe('plain text display names', () => {
    it('should return plain text as-is', () => {
      const t = vi.fn();
      const {result} = renderHook(() => useResolveDisplayName({handlers: {t}}));

      expect(result.current.resolveDisplayName('First Name')).toBe('First Name');
      expect(t).not.toHaveBeenCalled();
    });

    it('should return empty string for empty input', () => {
      const t = vi.fn();
      const {result} = renderHook(() => useResolveDisplayName({handlers: {t}}));

      expect(result.current.resolveDisplayName('')).toBe('');
    });

    it('should return empty string for whitespace-only input', () => {
      const t = vi.fn();
      const {result} = renderHook(() => useResolveDisplayName({handlers: {t}}));

      expect(result.current.resolveDisplayName('   ')).toBe('');
    });
  });

  describe('i18n template patterns', () => {
    it('should resolve a translated i18n pattern', () => {
      const t = vi.fn((key: string) => {
        if (key === 'custom:firstName') return 'First Name';
        return key;
      });
      const {result} = renderHook(() => useResolveDisplayName({handlers: {t}}));

      expect(result.current.resolveDisplayName('{{t(custom:firstName)}}')).toBe('First Name');
    });

    it('should return empty string when translation key is missing (t returns raw key)', () => {
      const t = vi.fn((key: string) => key);
      const {result} = renderHook(() => useResolveDisplayName({handlers: {t}}));

      expect(result.current.resolveDisplayName('{{t(custom:missingKey)}}')).toBe('');
    });

    it('should return empty string when t returns namespace-stripped key (fallback behavior)', () => {
      // i18next strips namespace on fallback: "custom:myKey" -> "myKey"
      const t = vi.fn(() => 'myKey');
      const {result} = renderHook(() => useResolveDisplayName({handlers: {t}}));

      expect(result.current.resolveDisplayName('{{t(custom:myKey)}}')).toBe('');
    });

    it('should return empty string when t returns undefined', () => {
      const t = vi.fn(() => undefined as unknown as string);
      const {result} = renderHook(() => useResolveDisplayName({handlers: {t}}));

      expect(result.current.resolveDisplayName('{{t(custom:someKey)}}')).toBe('');
    });

    it('should handle keys without namespace', () => {
      const t = vi.fn((key: string) => {
        if (key === 'greeting') return 'Hello';
        return key;
      });
      const {result} = renderHook(() => useResolveDisplayName({handlers: {t}}));

      expect(result.current.resolveDisplayName('{{t(greeting)}}')).toBe('Hello');
    });

    it('should return empty string for non-namespaced key when t returns the key unchanged', () => {
      const t = vi.fn((key: string) => key);
      const {result} = renderHook(() => useResolveDisplayName({handlers: {t}}));

      expect(result.current.resolveDisplayName('{{t(missingKey)}}')).toBe('');
    });
  });
});
