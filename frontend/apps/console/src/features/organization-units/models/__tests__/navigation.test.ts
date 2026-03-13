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
import type {OUNavigationState} from '../navigation';

describe('Navigation Models', () => {
  describe('OUNavigationState', () => {
    it('should have required fromOU property with id and name', () => {
      const state: OUNavigationState = {
        fromOU: {
          id: '550e8400-e29b-41d4-a716-446655440000',
          name: 'Engineering Department',
        },
      };

      expect(state.fromOU.id).toBe('550e8400-e29b-41d4-a716-446655440000');
      expect(state.fromOU.name).toBe('Engineering Department');
    });

    it('should represent navigation from parent to child OU', () => {
      const parentId = 'parent-ou-id';
      const parentName = 'Parent OU';

      const state: OUNavigationState = {
        fromOU: {id: parentId, name: parentName},
      };

      expect(state.fromOU).toEqual({id: parentId, name: parentName});
    });
  });
});
