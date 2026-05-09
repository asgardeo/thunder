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
import {InboundAuthTypes, type InboundAuthType, type InboundAuthConfig} from '../inbound-auth';
import type {OAuth2Config} from '../oauth';

describe('inbound-auth', () => {
  describe('InboundAuthTypes', () => {
    it('should have OAUTH2 type defined', () => {
      expect(InboundAuthTypes.OAUTH2).toBe('oauth2');
    });

    it('should be a constant object', () => {
      expect(typeof InboundAuthTypes).toBe('object');
      expect(InboundAuthTypes).toHaveProperty('OAUTH2');
    });

    it('should have correct value for OAUTH2', () => {
      const authType: InboundAuthType = 'oauth2';
      expect(authType).toBe(InboundAuthTypes.OAUTH2);
    });
  });

  describe('InboundAuthConfig interface', () => {
    it('should allow creating a valid OAuth2 inbound auth config', () => {
      const mockOAuth2Config: OAuth2Config = {
        redirectUris: ['https://example.com/callback'],
        grantTypes: ['authorization_code'],
        responseTypes: ['code'],
        scopes: ['openid', 'profile'],
      };

      const inboundAuthConfig: InboundAuthConfig = {
        type: InboundAuthTypes.OAUTH2,
        config: mockOAuth2Config,
      };

      expect(inboundAuthConfig.type).toBe('oauth2');
      expect(inboundAuthConfig.config).toEqual(mockOAuth2Config);
    });

    it('should work with string type for type field', () => {
      const mockConfig: OAuth2Config = {
        redirectUris: ['https://app.com/auth'],
        grantTypes: ['authorization_code', 'refresh_token'],
        responseTypes: ['code'],
        scopes: ['openid'],
      };

      const config: InboundAuthConfig = {
        type: 'oauth2',
        config: mockConfig,
      };

      expect(config.type).toBe(InboundAuthTypes.OAUTH2);
    });

    it('should allow PKCE configuration for SPAs', () => {
      const spaConfig: OAuth2Config = {
        redirectUris: ['http://localhost:3000/callback'],
        grantTypes: ['authorization_code', 'refresh_token'],
        responseTypes: ['code'],
        pkceRequired: true,
        publicClient: true,
        scopes: ['openid', 'profile', 'email'],
      };

      const inboundAuthConfig: InboundAuthConfig = {
        type: InboundAuthTypes.OAUTH2,
        config: spaConfig,
      };

      expect(inboundAuthConfig.config.pkceRequired).toBe(true);
      expect(inboundAuthConfig.config.publicClient).toBe(true);
    });
  });
});
