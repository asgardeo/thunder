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
import type {Application, BasicApplication} from '../application';

describe('Application Models', () => {
  describe('BasicApplication', () => {
    it('should have required id and name properties', () => {
      const app: BasicApplication = {
        id: '550e8400-e29b-41d4-a716-446655440000',
        name: 'Test Application',
      };

      expect(app.id).toBe('550e8400-e29b-41d4-a716-446655440000');
      expect(app.name).toBe('Test Application');
    });

    it('should accept optional properties', () => {
      const app: BasicApplication = {
        id: '1',
        name: 'Test App',
        description: 'Test description',
        logo_url: 'https://example.com/logo.png',
        client_id: 'test_client_id',
        auth_flow_id: 'auth_flow_123',
        registration_flow_id: 'reg_flow_456',
        is_registration_flow_enabled: true,
        template: 'react',
      };

      expect(app.description).toBe('Test description');
      expect(app.logo_url).toBe('https://example.com/logo.png');
      expect(app.client_id).toBe('test_client_id');
      expect(app.auth_flow_id).toBe('auth_flow_123');
      expect(app.registration_flow_id).toBe('reg_flow_456');
      expect(app.is_registration_flow_enabled).toBe(true);
      expect(app.template).toBe('react');
    });
  });

  describe('Application', () => {
    it('should accept minimal application object', () => {
      const app: Application = {
        id: '1',
        name: 'Test App',
        inbound_auth_config: [],
      };

      expect(app.id).toBe('1');
      expect(app.name).toBe('Test App');
      expect(app.inbound_auth_config).toEqual([]);
    });

    it('should accept full application object with all properties', () => {
      const app: Application = {
        id: '550e8400-e29b-41d4-a716-446655440000',
        name: 'My Web Application',
        description: 'Customer portal application',
        url: 'https://myapp.com',
        logo_url: 'https://myapp.com/logo.png',
        tos_uri: 'https://myapp.com/terms',
        policy_uri: 'https://myapp.com/privacy',
        contacts: ['admin@myapp.com'],
        auth_flow_id: 'flow_123',
        registration_flow_id: 'reg_123',
        is_registration_flow_enabled: true,
        template: 'nextjs',
        inbound_auth_config: [],
        theme_id: 'theme_123',
        certificate: {
          type: 'PEM',
          value: 'cert_value',
        },
        allowed_user_types: ['INTERNAL'],
      };

      expect(app).toHaveProperty('id');
      expect(app).toHaveProperty('name');
      expect(app).toHaveProperty('description');
      expect(app).toHaveProperty('url');
      expect(app).toHaveProperty('logo_url');
      expect(app).toHaveProperty('tos_uri');
      expect(app).toHaveProperty('policy_uri');
      expect(app).toHaveProperty('contacts');
      expect(app).toHaveProperty('auth_flow_id');
      expect(app).toHaveProperty('registration_flow_id');
      expect(app).toHaveProperty('is_registration_flow_enabled');
      expect(app).toHaveProperty('template');
      expect(app).toHaveProperty('inbound_auth_config');
      expect(app).toHaveProperty('theme_id');
      expect(app).toHaveProperty('certificate');
      expect(app).toHaveProperty('allowed_user_types');
    });
  });
});
