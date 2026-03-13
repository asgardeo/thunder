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
import type {TokenConfig} from '../token';

describe('Token Models', () => {
  describe('TokenConfig', () => {
    it('should have required validity_period and user_attributes', () => {
      const tokenConfig: TokenConfig = {
        validity_period: 3600,
        user_attributes: ['email', 'username'],
      };

      expect(tokenConfig).toHaveProperty('validity_period');
      expect(tokenConfig).toHaveProperty('user_attributes');
      expect(tokenConfig.validity_period).toBe(3600);
      expect(tokenConfig.user_attributes).toEqual(['email', 'username']);
    });


    it('should accept validity period and user attributes', () => {
      const tokenConfig: TokenConfig = {
        validity_period: 7200,
        user_attributes: ['sub', 'email', 'name'],
      };

      expect(tokenConfig.validity_period).toBe(7200);
      expect(tokenConfig.user_attributes).toHaveLength(3);
    });

    it('should accept empty user_attributes array', () => {
      const tokenConfig: TokenConfig = {
        validity_period: 1800,
        user_attributes: [],
      };

      expect(tokenConfig.user_attributes).toEqual([]);
      expect(tokenConfig.user_attributes).toHaveLength(0);
    });

    it('should accept various validity periods', () => {
      const shortLived: TokenConfig = {
        validity_period: 300, // 5 minutes
        user_attributes: [],
      };

      const longLived: TokenConfig = {
        validity_period: 86400, // 24 hours
        user_attributes: [],
      };

      expect(shortLived.validity_period).toBe(300);
      expect(longLived.validity_period).toBe(86400);
    });

    it('should accept multiple user attributes', () => {
      const tokenConfig: TokenConfig = {
        validity_period: 3600,
        user_attributes: ['sub', 'email', 'email_verified', 'name', 'given_name', 'family_name', 'picture', 'roles'],
      };

      expect(tokenConfig.user_attributes).toHaveLength(8);
      expect(tokenConfig.user_attributes).toContain('email');
      expect(tokenConfig.user_attributes).toContain('roles');
    });
  });
});
