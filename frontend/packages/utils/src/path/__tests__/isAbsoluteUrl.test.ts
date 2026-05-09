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
import isAbsoluteUrl from '../isAbsoluteUrl';

describe('isAbsoluteUrl', () => {
  describe('absolute URLs', () => {
    it('should return true for an https URL', () => {
      expect(isAbsoluteUrl('https://example.com/foo')).toBe(true);
    });

    it('should return true for an http URL', () => {
      expect(isAbsoluteUrl('http://example.com/foo')).toBe(true);
    });

    it('should return true for a protocol-relative URL', () => {
      expect(isAbsoluteUrl('//example.com/foo')).toBe(true);
    });

    it('should return true for a custom scheme URL', () => {
      expect(isAbsoluteUrl('ftp://files.example.com/resource')).toBe(true);
    });
  });

  describe('relative URLs', () => {
    it('should return false for a current-directory relative path', () => {
      expect(isAbsoluteUrl('./foo')).toBe(false);
    });

    it('should return false for a parent-directory relative path', () => {
      expect(isAbsoluteUrl('../foo')).toBe(false);
    });

    it('should return false for a bare segment', () => {
      expect(isAbsoluteUrl('foo/bar')).toBe(false);
    });

    it('should return false for a root-relative path', () => {
      expect(isAbsoluteUrl('/foo/bar')).toBe(false);
    });

    it('should return false for an empty string', () => {
      expect(isAbsoluteUrl('')).toBe(false);
    });
  });
});
