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
import isRelativeUrl from '../isRelativeUrl';

describe('isRelativeUrl', () => {
  describe('relative URLs', () => {
    it('should return true for a current-directory relative path', () => {
      expect(isRelativeUrl('./foo')).toBe(true);
    });

    it('should return true for a parent-directory relative path', () => {
      expect(isRelativeUrl('../foo')).toBe(true);
    });

    it('should return true for a bare segment', () => {
      expect(isRelativeUrl('foo/bar')).toBe(true);
    });

    it('should return true for a root-relative path', () => {
      expect(isRelativeUrl('/foo/bar')).toBe(true);
    });

    it('should return true for an empty string', () => {
      expect(isRelativeUrl('')).toBe(true);
    });
  });

  describe('absolute URLs', () => {
    it('should return false for an https URL', () => {
      expect(isRelativeUrl('https://example.com/foo')).toBe(false);
    });

    it('should return false for an http URL', () => {
      expect(isRelativeUrl('http://example.com/foo')).toBe(false);
    });

    it('should return false for a protocol-relative URL', () => {
      expect(isRelativeUrl('//example.com/foo')).toBe(false);
    });

    it('should return false for a custom scheme URL', () => {
      expect(isRelativeUrl('ftp://files.example.com/resource')).toBe(false);
    });
  });
});
