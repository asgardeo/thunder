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
import type {ApplicationListResponse} from '../responses';
import type {BasicApplication} from '../application';

describe('Responses Models', () => {
  describe('ApplicationListResponse', () => {
    it('should have required properties', () => {
      const mockApplication: BasicApplication = {
        id: '550e8400-e29b-41d4-a716-446655440000',
        name: 'Test App',
        client_id: 'test_client_id',
      };

      const response: ApplicationListResponse = {
        totalResults: 1,
        count: 1,
        applications: [mockApplication],
      };

      expect(response).toHaveProperty('totalResults');
      expect(response).toHaveProperty('count');
      expect(response).toHaveProperty('applications');
    });

    it('should accept valid totalResults and count', () => {
      const response: ApplicationListResponse = {
        totalResults: 25,
        count: 10,
        applications: [],
      };

      expect(response.totalResults).toBe(25);
      expect(response.count).toBe(10);
      expect(Array.isArray(response.applications)).toBe(true);
    });

    it('should accept array of BasicApplication', () => {
      const mockApplications: BasicApplication[] = [
        {
          id: '1',
          name: 'App 1',
          client_id: 'client_1',
        },
        {
          id: '2',
          name: 'App 2',
          description: 'Test description',
          client_id: 'client_2',
          logo_url: 'https://example.com/logo.png',
        },
      ];

      const response: ApplicationListResponse = {
        totalResults: 2,
        count: 2,
        applications: mockApplications,
      };

      expect(response.applications).toHaveLength(2);
      expect(response.applications[0].name).toBe('App 1');
      expect(response.applications[1].description).toBe('Test description');
    });

    it('should handle empty applications array', () => {
      const response: ApplicationListResponse = {
        totalResults: 0,
        count: 0,
        applications: [],
      };

      expect(response.applications).toHaveLength(0);
      expect(response.totalResults).toBe(0);
      expect(response.count).toBe(0);
    });
  });
});
