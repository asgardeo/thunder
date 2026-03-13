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
import type {Group, GroupListResponse} from '../group';

describe('Group Models', () => {
  describe('Group', () => {
    it('should have required id, name, and organizationUnit properties', () => {
      const group: Group = {
        id: '7a1b2c3d-4e5f-6789-abcd-ef0123456789',
        name: 'Developers',
        organizationUnit: '550e8400-e29b-41d4-a716-446655440000',
      };

      expect(group.id).toBe('7a1b2c3d-4e5f-6789-abcd-ef0123456789');
      expect(group.name).toBe('Developers');
      expect(group.organizationUnit).toBe('550e8400-e29b-41d4-a716-446655440000');
    });
  });

  describe('GroupListResponse', () => {
    it('should have required properties', () => {
      const response: GroupListResponse = {
        totalResults: 1,
        startIndex: 0,
        count: 1,
        groups: [{id: 'group-1', name: 'Developers', organizationUnit: 'ou-id'}],
      };

      expect(response).toHaveProperty('totalResults');
      expect(response).toHaveProperty('startIndex');
      expect(response).toHaveProperty('count');
      expect(response).toHaveProperty('groups');
    });

    it('should accept valid pagination values', () => {
      const response: GroupListResponse = {
        totalResults: 15,
        startIndex: 0,
        count: 10,
        groups: [],
      };

      expect(response.totalResults).toBe(15);
      expect(response.startIndex).toBe(0);
      expect(response.count).toBe(10);
      expect(Array.isArray(response.groups)).toBe(true);
    });

    it('should accept array of Group', () => {
      const mockGroups: Group[] = [
        {id: '1', name: 'Developers', organizationUnit: 'ou-1'},
        {id: '2', name: 'Designers', organizationUnit: 'ou-1'},
        {id: '3', name: 'Managers', organizationUnit: 'ou-1'},
      ];

      const response: GroupListResponse = {
        totalResults: 3,
        startIndex: 0,
        count: 3,
        groups: mockGroups,
      };

      expect(response.groups).toHaveLength(3);
      expect(response.groups[0].name).toBe('Developers');
      expect(response.groups[2].name).toBe('Managers');
    });

    it('should handle empty groups array', () => {
      const response: GroupListResponse = {
        totalResults: 0,
        startIndex: 0,
        count: 0,
        groups: [],
      };

      expect(response.groups).toHaveLength(0);
      expect(response.totalResults).toBe(0);
    });

    it('should accept optional pagination links', () => {
      const response: GroupListResponse = {
        totalResults: 20,
        startIndex: 0,
        count: 10,
        groups: [],
        links: [{rel: 'next', href: '/groups?offset=10&limit=10'}],
      };

      expect(response.links).toHaveLength(1);
      expect(response.links![0].rel).toBe('next');
    });
  });
});
