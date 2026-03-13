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
import type {CreateOrganizationUnitRequest, UpdateOrganizationUnitRequest, OrganizationUnitListParams} from '../requests';

describe('Request Models', () => {
  describe('CreateOrganizationUnitRequest', () => {
    it('should accept required handle and name properties', () => {
      const request: CreateOrganizationUnitRequest = {
        handle: 'engineering',
        name: 'Engineering Department',
      };

      expect(request.handle).toBe('engineering');
      expect(request.name).toBe('Engineering Department');
    });

    it('should accept optional description and parent', () => {
      const request: CreateOrganizationUnitRequest = {
        handle: 'engineering',
        name: 'Engineering Department',
        description: 'Software engineering team',
        parent: 'root-ou-id',
      };

      expect(request.description).toBe('Software engineering team');
      expect(request.parent).toBe('root-ou-id');
    });

    it('should accept null for nullable optional fields', () => {
      const request: CreateOrganizationUnitRequest = {
        handle: 'engineering',
        name: 'Engineering',
        description: null,
        parent: null,
      };

      expect(request.description).toBeNull();
      expect(request.parent).toBeNull();
    });

    it('should accept undefined for optional fields', () => {
      const request: CreateOrganizationUnitRequest = {
        handle: 'root',
        name: 'Root',
        description: undefined,
        parent: undefined,
      };

      expect(request.description).toBeUndefined();
      expect(request.parent).toBeUndefined();
    });
  });

  describe('UpdateOrganizationUnitRequest', () => {
    it('should accept required handle and name properties', () => {
      const request: UpdateOrganizationUnitRequest = {
        handle: 'engineering',
        name: 'Engineering Department',
      };

      expect(request.handle).toBe('engineering');
      expect(request.name).toBe('Engineering Department');
    });

    it('should accept all updatable fields', () => {
      const request: UpdateOrganizationUnitRequest = {
        handle: 'engineering',
        name: 'Engineering Department (Updated)',
        description: 'Updated description',
        parent: 'root-ou-id',
        theme_id: '96c62e6d-9297-4295-8195-d28dfe0c9ff7',
        logo_url: 'https://example.com/new-logo.png',
      };

      expect(request.description).toBe('Updated description');
      expect(request.parent).toBe('root-ou-id');
      expect(request.theme_id).toBe('96c62e6d-9297-4295-8195-d28dfe0c9ff7');
      expect(request.logo_url).toBe('https://example.com/new-logo.png');
    });

    it('should accept null for nullable fields', () => {
      const request: UpdateOrganizationUnitRequest = {
        handle: 'engineering',
        name: 'Engineering',
        description: null,
        parent: null,
        theme_id: null,
      };

      expect(request.description).toBeNull();
      expect(request.parent).toBeNull();
      expect(request.theme_id).toBeNull();
    });
  });

  describe('OrganizationUnitListParams', () => {
    it('should accept empty params', () => {
      const params: OrganizationUnitListParams = {};

      expect(params.limit).toBeUndefined();
      expect(params.offset).toBeUndefined();
    });

    it('should accept limit and offset', () => {
      const params: OrganizationUnitListParams = {
        limit: 10,
        offset: 20,
      };

      expect(params.limit).toBe(10);
      expect(params.offset).toBe(20);
    });

    it('should accept only limit', () => {
      const params: OrganizationUnitListParams = {
        limit: 50,
      };

      expect(params.limit).toBe(50);
      expect(params.offset).toBeUndefined();
    });
  });
});
