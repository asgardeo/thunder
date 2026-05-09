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
import type {ApiError} from '../api-error';

describe('API Error Models', () => {
  describe('ApiError', () => {
    it('should have required code, message, and description properties', () => {
      const error: ApiError = {
        code: 'OU-60001',
        message: 'Organization unit not found',
        description: 'No organization unit exists with the given ID',
      };

      expect(error.code).toBe('OU-60001');
      expect(error.message).toBe('Organization unit not found');
      expect(error.description).toBe('No organization unit exists with the given ID');
    });

    it('should represent a validation error', () => {
      const error: ApiError = {
        code: 'OU-60002',
        message: 'Invalid request',
        description: 'The handle field is required and must be unique',
      };

      expect(error).toHaveProperty('code');
      expect(error).toHaveProperty('message');
      expect(error).toHaveProperty('description');
    });
  });
});
