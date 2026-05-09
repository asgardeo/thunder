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
import type {CreateApplicationRequest} from '../requests';

describe('Requests Models', () => {
  describe('CreateApplicationRequest', () => {
    it('should accept minimal application request', () => {
      const request: CreateApplicationRequest = {
        name: 'Test App',
        inboundAuthConfig: [],
      };

      expect(request.name).toBe('Test App');
      expect(request.inboundAuthConfig).toEqual([]);
    });

    it('should accept full application request', () => {
      const request: CreateApplicationRequest = {
        name: 'My Web Application',
        description: 'Customer portal application',
        url: 'https://myapp.com',
        logoUrl: 'https://myapp.com/logo.png',
        tosUri: 'https://myapp.com/terms',
        policyUri: 'https://myapp.com/privacy',
        contacts: ['admin@myapp.com', 'support@myapp.com'],
        authFlowId: 'edc013d0-e893-4dc0-990c-3e1d203e005b',
        registrationFlowId: '80024fb3-29ed-4c33-aa48-8aee5e96d522',
        isRegistrationFlowEnabled: true,
        inboundAuthConfig: [],
      };

      expect(request).toHaveProperty('name');
      expect(request).toHaveProperty('description');
      expect(request).toHaveProperty('url');
      expect(request).toHaveProperty('logoUrl');
      expect(request).toHaveProperty('tosUri');
      expect(request).toHaveProperty('policyUri');
      expect(request).toHaveProperty('contacts');
      expect(request).toHaveProperty('authFlowId');
      expect(request).toHaveProperty('registrationFlowId');
      expect(request).toHaveProperty('isRegistrationFlowEnabled');
      expect(request).toHaveProperty('inboundAuthConfig');
    });

    it('should accept optional fields as undefined', () => {
      const request: CreateApplicationRequest = {
        name: 'Test App',
        description: undefined,
        url: undefined,
        inboundAuthConfig: [],
      };

      expect(request.description).toBeUndefined();
      expect(request.url).toBeUndefined();
    });

    it('should accept array of contacts', () => {
      const request: CreateApplicationRequest = {
        name: 'Test App',
        contacts: ['user1@example.com', 'user2@example.com', 'admin@example.com'],
        inboundAuthConfig: [],
      };

      expect(request.contacts).toHaveLength(3);
      expect(request.contacts).toContain('admin@example.com');
    });

    it('should accept template field', () => {
      const request: CreateApplicationRequest = {
        name: 'Test App',
        template: 'react',
        inboundAuthConfig: [],
      };

      expect(request.template).toBe('react');
    });
  });
});
