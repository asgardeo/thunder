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
import type {OrganizationUnit} from '../organization-unit';

describe('Organization Unit Models', () => {
  describe('OrganizationUnit', () => {
    it('should have required id, handle, and name properties', () => {
      const ou: OrganizationUnit = {
        id: '550e8400-e29b-41d4-a716-446655440000',
        handle: 'engineering',
        name: 'Engineering Department',
      };

      expect(ou.id).toBe('550e8400-e29b-41d4-a716-446655440000');
      expect(ou.handle).toBe('engineering');
      expect(ou.name).toBe('Engineering Department');
    });

    it('should accept optional properties', () => {
      const ou: OrganizationUnit = {
        id: '1',
        handle: 'engineering',
        name: 'Engineering',
        description: 'Software engineering team',
        parent: 'root-ou-id',
        theme_id: '96c62e6d-9297-4295-8195-d28dfe0c9ff7',
        layout_id: 'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
        logo_url: 'https://example.com/logo.png',
      };

      expect(ou.description).toBe('Software engineering team');
      expect(ou.parent).toBe('root-ou-id');
      expect(ou.theme_id).toBe('96c62e6d-9297-4295-8195-d28dfe0c9ff7');
      expect(ou.layout_id).toBe('a1b2c3d4-e5f6-7890-abcd-ef1234567890');
      expect(ou.logo_url).toBe('https://example.com/logo.png');
    });

    it('should accept null for nullable optional properties', () => {
      const ou: OrganizationUnit = {
        id: '1',
        handle: 'root',
        name: 'Root',
        description: null,
        parent: null,
        theme_id: null,
        layout_id: null,
      };

      expect(ou.description).toBeNull();
      expect(ou.parent).toBeNull();
      expect(ou.theme_id).toBeNull();
      expect(ou.layout_id).toBeNull();
    });

    it('should accept undefined for optional properties', () => {
      const ou: OrganizationUnit = {
        id: '1',
        handle: 'root',
        name: 'Root',
        description: undefined,
        parent: undefined,
      };

      expect(ou.description).toBeUndefined();
      expect(ou.parent).toBeUndefined();
    });
  });
});
