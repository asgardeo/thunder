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
import isEmpty from '../isEmpty';

describe('isEmpty', () => {
  describe('null and undefined', () => {
    it('should return true for null', () => {
      expect(isEmpty(null)).toBe(true);
    });

    it('should return true for undefined', () => {
      expect(isEmpty(undefined)).toBe(true);
    });
  });

  describe('primitives', () => {
    it('should return true for boolean true', () => {
      expect(isEmpty(true)).toBe(true);
    });

    it('should return true for boolean false', () => {
      expect(isEmpty(false)).toBe(true);
    });

    it('should return true for zero', () => {
      expect(isEmpty(0)).toBe(true);
    });

    it('should return true for a non-zero number', () => {
      expect(isEmpty(42)).toBe(true);
    });

    it('should return true for a Symbol', () => {
      expect(isEmpty(Symbol('x'))).toBe(true);
    });
  });

  describe('strings', () => {
    it('should return true for an empty string', () => {
      expect(isEmpty('')).toBe(true);
    });

    it('should return false for a non-empty string', () => {
      expect(isEmpty('hello')).toBe(false);
    });
  });

  describe('arrays', () => {
    it('should return true for an empty array', () => {
      expect(isEmpty([])).toBe(true);
    });

    it('should return false for a non-empty array', () => {
      expect(isEmpty([1, 2, 3])).toBe(false);
    });
  });

  describe('objects', () => {
    it('should return true for an empty plain object', () => {
      expect(isEmpty({})).toBe(true);
    });

    it('should return false for a non-empty plain object', () => {
      expect(isEmpty({a: 1})).toBe(false);
    });

    it('should only consider own enumerable keys', () => {
      const obj: object = Object.create({inherited: true}) as object;
      expect(isEmpty(obj)).toBe(true);
    });
  });

  describe('Map', () => {
    it('should return true for an empty Map', () => {
      expect(isEmpty(new Map())).toBe(true);
    });

    it('should return false for a non-empty Map', () => {
      expect(isEmpty(new Map([['key', 'value']]))).toBe(false);
    });
  });

  describe('Set', () => {
    it('should return true for an empty Set', () => {
      expect(isEmpty(new Set())).toBe(true);
    });

    it('should return false for a non-empty Set', () => {
      expect(isEmpty(new Set([1, 2]))).toBe(false);
    });
  });

  describe('array-like objects', () => {
    it('should return true for an object with length 0', () => {
      expect(isEmpty({length: 0})).toBe(true);
    });

    it('should return false for an object with length > 0', () => {
      expect(isEmpty({length: 3})).toBe(false);
    });
  });
});
