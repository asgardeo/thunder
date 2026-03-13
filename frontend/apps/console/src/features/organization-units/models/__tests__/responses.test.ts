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
import type {OrganizationUnitListResponse} from '../responses';
import type {OrganizationUnit} from '../organization-unit';

describe('Response Models', () => {
  describe('OrganizationUnitListResponse', () => {
    it('should have required properties', () => {
      const mockOU: OrganizationUnit = {
        id: '550e8400-e29b-41d4-a716-446655440000',
        handle: 'engineering',
        name: 'Engineering Department',
      };

      const response: OrganizationUnitListResponse = {
        totalResults: 1,
        startIndex: 0,
        count: 1,
        organizationUnits: [mockOU],
      };

      expect(response).toHaveProperty('totalResults');
      expect(response).toHaveProperty('startIndex');
      expect(response).toHaveProperty('count');
      expect(response).toHaveProperty('organizationUnits');
    });

    it('should accept valid pagination values', () => {
      const response: OrganizationUnitListResponse = {
        totalResults: 50,
        startIndex: 10,
        count: 10,
        organizationUnits: [],
      };

      expect(response.totalResults).toBe(50);
      expect(response.startIndex).toBe(10);
      expect(response.count).toBe(10);
      expect(Array.isArray(response.organizationUnits)).toBe(true);
    });

    it('should accept array of OrganizationUnit', () => {
      const mockOUs: OrganizationUnit[] = [
        {id: '1', handle: 'engineering', name: 'Engineering'},
        {id: '2', handle: 'marketing', name: 'Marketing', description: 'Marketing team'},
      ];

      const response: OrganizationUnitListResponse = {
        totalResults: 2,
        startIndex: 0,
        count: 2,
        organizationUnits: mockOUs,
      };

      expect(response.organizationUnits).toHaveLength(2);
      expect(response.organizationUnits[0].handle).toBe('engineering');
      expect(response.organizationUnits[1].description).toBe('Marketing team');
    });

    it('should handle empty organization units array', () => {
      const response: OrganizationUnitListResponse = {
        totalResults: 0,
        startIndex: 0,
        count: 0,
        organizationUnits: [],
      };

      expect(response.organizationUnits).toHaveLength(0);
      expect(response.totalResults).toBe(0);
      expect(response.count).toBe(0);
    });

    it('should accept optional pagination links', () => {
      const response: OrganizationUnitListResponse = {
        totalResults: 50,
        startIndex: 0,
        count: 10,
        organizationUnits: [],
        links: [
          {rel: 'next', href: '/organization-units?offset=10&limit=10'},
          {rel: 'prev', href: '/organization-units?offset=0&limit=10'},
        ],
      };

      expect(response.links).toHaveLength(2);
      expect(response.links![0].rel).toBe('next');
      expect(response.links![1].rel).toBe('prev');
    });
  });
});
