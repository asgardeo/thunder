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

import {describe, it, expect, vi, beforeEach} from 'vitest';
import {
  registerI18nCacheInvalidator,
  unregisterI18nCacheInvalidator,
  invalidateI18nCache,
} from '../invalidate-i18n-cache';

describe('invalidate-i18n-cache', () => {
  const I18N_CACHE_INVALIDATE_KEY = 'thunderI18nCacheInvalidate';

  beforeEach(() => {
    // Clean up the global key before each test
    delete (window as unknown as Record<string, unknown>)[I18N_CACHE_INVALIDATE_KEY];
  });

  describe('registerI18nCacheInvalidator', () => {
    it('should register the invalidation function on window', () => {
      const mockInvalidateFn = vi.fn();

      registerI18nCacheInvalidator(mockInvalidateFn);

      expect((window as unknown as Record<string, unknown>)[I18N_CACHE_INVALIDATE_KEY]).toBe(mockInvalidateFn);
    });

    it('should overwrite existing invalidation function', () => {
      const firstFn = vi.fn();
      const secondFn = vi.fn();

      registerI18nCacheInvalidator(firstFn);
      registerI18nCacheInvalidator(secondFn);

      expect((window as unknown as Record<string, unknown>)[I18N_CACHE_INVALIDATE_KEY]).toBe(secondFn);
    });
  });

  describe('unregisterI18nCacheInvalidator', () => {
    it('should remove the invalidation function from window', () => {
      const mockInvalidateFn = vi.fn();
      registerI18nCacheInvalidator(mockInvalidateFn);

      unregisterI18nCacheInvalidator();

      expect((window as unknown as Record<string, unknown>)[I18N_CACHE_INVALIDATE_KEY]).toBeUndefined();
    });

    it('should not throw when no function is registered', () => {
      expect(() => unregisterI18nCacheInvalidator()).not.toThrow();
    });
  });

  describe('invalidateI18nCache', () => {
    it('should call the registered invalidation function', () => {
      const mockInvalidateFn = vi.fn();
      registerI18nCacheInvalidator(mockInvalidateFn);

      invalidateI18nCache();

      expect(mockInvalidateFn).toHaveBeenCalledTimes(1);
    });

    it('should not throw when no function is registered', () => {
      expect(() => invalidateI18nCache()).not.toThrow();
    });

    it('should not call function after it has been unregistered', () => {
      const mockInvalidateFn = vi.fn();
      registerI18nCacheInvalidator(mockInvalidateFn);
      unregisterI18nCacheInvalidator();

      invalidateI18nCache();

      expect(mockInvalidateFn).not.toHaveBeenCalled();
    });
  });
});
