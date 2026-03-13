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

import {describe, it, expect} from 'vitest';
import TokenConstants from '../token-constants';

describe('TokenConstants', () => {
  describe('DEFAULT_TOKEN_ATTRIBUTES', () => {
    it('should be defined', () => {
      expect(TokenConstants.DEFAULT_TOKEN_ATTRIBUTES).toBeDefined();
    });

    it('should be an array', () => {
      expect(Array.isArray(TokenConstants.DEFAULT_TOKEN_ATTRIBUTES)).toBe(true);
    });

    it('should contain standard JWT claims', () => {
      const standardClaims = ['aud', 'exp', 'iat', 'iss', 'sub', 'nbf', 'jti'];

      standardClaims.forEach((claim) => {
        expect(TokenConstants.DEFAULT_TOKEN_ATTRIBUTES).toContain(claim);
      });
    });

    it('should contain OAuth2 specific claims', () => {
      const oauth2Claims = ['client_id', 'grant_type', 'scope'];

      oauth2Claims.forEach((claim) => {
        expect(TokenConstants.DEFAULT_TOKEN_ATTRIBUTES).toContain(claim);
      });
    });

    it('should have the expected number of attributes', () => {
      expect(TokenConstants.DEFAULT_TOKEN_ATTRIBUTES).toHaveLength(10);
    });

    it('should not contain duplicate values', () => {
      const unique = new Set(TokenConstants.DEFAULT_TOKEN_ATTRIBUTES);
      expect(unique.size).toBe(TokenConstants.DEFAULT_TOKEN_ATTRIBUTES.length);
    });

    it('should contain all expected attributes in correct order', () => {
      const expectedAttributes = [
        'aud',
        'client_id',
        'exp',
        'grant_type',
        'iat',
        'iss',
        'jti',
        'nbf',
        'scope',
        'sub',
      ];

      expect(TokenConstants.DEFAULT_TOKEN_ATTRIBUTES).toEqual(expectedAttributes);
    });
  });

  describe('USER_INFO_DEFAULT_ATTRIBUTES', () => {
    it('should be defined', () => {
      expect(TokenConstants.USER_INFO_DEFAULT_ATTRIBUTES).toBeDefined();
    });

    it('should be an array', () => {
      expect(Array.isArray(TokenConstants.USER_INFO_DEFAULT_ATTRIBUTES)).toBe(true);
    });

    it('should contain sub attribute', () => {
      expect(TokenConstants.USER_INFO_DEFAULT_ATTRIBUTES).toContain('sub');
    });

    it('should match expected defaults', () => {
      expect(TokenConstants.USER_INFO_DEFAULT_ATTRIBUTES).toEqual(['sub']);
    });
  });

  describe('ADDITIONAL_USER_ATTRIBUTES', () => {
    it('should be defined', () => {
      expect(TokenConstants.ADDITIONAL_USER_ATTRIBUTES).toBeDefined();
    });

    it('should be an array', () => {
      expect(Array.isArray(TokenConstants.ADDITIONAL_USER_ATTRIBUTES)).toBe(true);
    });

    it('should contain expected attributes', () => {
      expect(TokenConstants.ADDITIONAL_USER_ATTRIBUTES).toContain('ouHandle');
      expect(TokenConstants.ADDITIONAL_USER_ATTRIBUTES).toContain('ouId');
      expect(TokenConstants.ADDITIONAL_USER_ATTRIBUTES).toContain('ouName');
      expect(TokenConstants.ADDITIONAL_USER_ATTRIBUTES).toContain('userType');
    });
  });
});
