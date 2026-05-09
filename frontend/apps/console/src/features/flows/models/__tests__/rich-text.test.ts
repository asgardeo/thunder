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

import {describe, it, expect} from 'vitest';
import {UPDATE_TYPES} from '../rich-text';
import type {UpdateType} from '../rich-text';

describe('rich-text', () => {
  describe('UPDATE_TYPES', () => {
    it('should have INTERNAL type with value "internal"', () => {
      expect(UPDATE_TYPES.INTERNAL).toBe('internal');
    });

    it('should have EXTERNAL type with value "external"', () => {
      expect(UPDATE_TYPES.EXTERNAL).toBe('external');
    });

    it('should have NONE type with value "none"', () => {
      expect(UPDATE_TYPES.NONE).toBe('none');
    });

    it('should have exactly three update types', () => {
      const keys = Object.keys(UPDATE_TYPES);
      expect(keys).toHaveLength(3);
      expect(keys).toContain('INTERNAL');
      expect(keys).toContain('EXTERNAL');
      expect(keys).toContain('NONE');
    });

    it('should be immutable (const assertion)', () => {
      const updateType: UpdateType = UPDATE_TYPES.INTERNAL;
      expect(updateType).toBe('internal');

      const externalType: UpdateType = UPDATE_TYPES.EXTERNAL;
      expect(externalType).toBe('external');

      const noneType: UpdateType = UPDATE_TYPES.NONE;
      expect(noneType).toBe('none');
    });
  });
});
