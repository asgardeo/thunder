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

import {describe, expect, it} from 'vitest';
import {IdentityProviderTypes, type IdentityProvider} from '@/features/integrations/models/identity-provider';
import {AuthenticatorTypes} from '../../../integrations/models/authenticators';
import resolveAuthFlowHandle from '../resolveAuthFlowHandle';
import {AUTH_FLOW_HANDLES} from '../../models/auth-flow-graphs';

describe('resolveAuthFlowHandle', () => {
  describe('Constants', () => {
    it('should export AuthenticatorTypes.BASIC_AUTH', () => {
      expect(AuthenticatorTypes.BASIC_AUTH).toBe('basic_auth');
    });
  });

  describe('Single Authentication Method', () => {
    it('should return BASIC flow when only username/password is enabled', () => {
      const result = resolveAuthFlowHandle({
        hasUsernamePassword: true,
        identityProviders: [],
      });

      expect(result).toBe(AUTH_FLOW_HANDLES.BASIC);
      expect(result).toBe('default-basic-flow');
    });

    it('should return GOOGLE flow when only Google is selected', () => {
      const googleProvider: IdentityProvider = {
        id: 'google-123',
        name: 'Google',
        type: IdentityProviderTypes.GOOGLE,
      };

      const result = resolveAuthFlowHandle({
        hasUsernamePassword: false,
        identityProviders: [googleProvider],
      });

      expect(result).toBe(AUTH_FLOW_HANDLES.GOOGLE);
      expect(result).toBe('default-google-flow');
    });

    it('should return GITHUB flow when only GitHub is selected', () => {
      const githubProvider: IdentityProvider = {
        id: 'github-456',
        name: 'GitHub',
        type: IdentityProviderTypes.GITHUB,
      };

      const result = resolveAuthFlowHandle({
        hasUsernamePassword: false,
        identityProviders: [githubProvider],
      });

      expect(result).toBe(AUTH_FLOW_HANDLES.GITHUB);
      expect(result).toBe('default-github-flow');
    });
  });

  describe('Combined Authentication Methods', () => {
    it('should return BASIC_GOOGLE flow when username/password + Google', () => {
      const googleProvider: IdentityProvider = {
        id: 'google-123',
        name: 'Google',
        type: IdentityProviderTypes.GOOGLE,
      };

      const result = resolveAuthFlowHandle({
        hasUsernamePassword: true,
        identityProviders: [googleProvider],
      });

      expect(result).toBe(AUTH_FLOW_HANDLES.BASIC_GOOGLE);
      expect(result).toBe('basic-google-flow');
    });

    it('should return BASIC_GOOGLE_GITHUB flow when username/password + Google + GitHub', () => {
      const googleProvider: IdentityProvider = {
        id: 'google-123',
        name: 'Google',
        type: IdentityProviderTypes.GOOGLE,
      };

      const githubProvider: IdentityProvider = {
        id: 'github-456',
        name: 'GitHub',
        type: IdentityProviderTypes.GITHUB,
      };

      const result = resolveAuthFlowHandle({
        hasUsernamePassword: true,
        identityProviders: [googleProvider, githubProvider],
      });

      expect(result).toBe(AUTH_FLOW_HANDLES.BASIC_GOOGLE_GITHUB);
      expect(result).toBe('basic-google-github-flow');
    });

    it('should return BASIC_GITHUB flow when username/password + GitHub', () => {
      const githubProvider: IdentityProvider = {
        id: 'github-456',
        name: 'GitHub',
        type: IdentityProviderTypes.GITHUB,
      };

      const result = resolveAuthFlowHandle({
        hasUsernamePassword: true,
        identityProviders: [githubProvider],
      });

      expect(result).toBe(AUTH_FLOW_HANDLES.BASIC_GITHUB);
      expect(result).toBe('basic-github-flow');
    });

    it('should return GOOGLE_GITHUB flow when Google + GitHub without username/password', () => {
      const googleProvider: IdentityProvider = {
        id: 'google-123',
        name: 'Google',
        type: IdentityProviderTypes.GOOGLE,
      };

      const githubProvider: IdentityProvider = {
        id: 'github-456',
        name: 'GitHub',
        type: IdentityProviderTypes.GITHUB,
      };

      const result = resolveAuthFlowHandle({
        hasUsernamePassword: false,
        identityProviders: [googleProvider, githubProvider],
      });

      expect(result).toBe(AUTH_FLOW_HANDLES.GOOGLE_GITHUB);
      expect(result).toBe('google-github-flow');
    });
  });

  describe('Multiple Identity Providers', () => {
    it('should handle multiple Google providers (duplicate types)', () => {
      const googleProvider1: IdentityProvider = {
        id: 'google-123',
        name: 'Google 1',
        type: IdentityProviderTypes.GOOGLE,
      };

      const googleProvider2: IdentityProvider = {
        id: 'google-456',
        name: 'Google 2',
        type: IdentityProviderTypes.GOOGLE,
      };

      const result = resolveAuthFlowHandle({
        hasUsernamePassword: false,
        identityProviders: [googleProvider1, googleProvider2],
      });

      // Should still treat as single Google provider
      expect(result).toBe(AUTH_FLOW_HANDLES.GOOGLE);
    });

    it('should handle providers in different order', () => {
      const googleProvider: IdentityProvider = {
        id: 'google-123',
        name: 'Google',
        type: IdentityProviderTypes.GOOGLE,
      };

      const githubProvider: IdentityProvider = {
        id: 'github-456',
        name: 'GitHub',
        type: IdentityProviderTypes.GITHUB,
      };

      const result1 = resolveAuthFlowHandle({
        hasUsernamePassword: true,
        identityProviders: [googleProvider, githubProvider],
      });

      const result2 = resolveAuthFlowHandle({
        hasUsernamePassword: true,
        identityProviders: [githubProvider, googleProvider],
      });

      // Order shouldn't matter
      expect(result1).toBe(result2);
      expect(result1).toBe(AUTH_FLOW_HANDLES.BASIC_GOOGLE_GITHUB);
    });
  });

  describe('Edge Cases', () => {
    it('should return BASIC flow when no authentication method is selected (fallback)', () => {
      const result = resolveAuthFlowHandle({
        hasUsernamePassword: false,
        identityProviders: [],
      });

      // Default fallback
      expect(result).toBe(AUTH_FLOW_HANDLES.BASIC);
      expect(result).toBe('default-basic-flow');
    });

    it('should handle empty identity providers array', () => {
      const result = resolveAuthFlowHandle({
        hasUsernamePassword: true,
        identityProviders: [],
      });

      expect(result).toBe(AUTH_FLOW_HANDLES.BASIC);
    });

    it('should handle identity provider with different properties', () => {
      const customProvider: IdentityProvider = {
        id: 'google-999',
        name: 'My Custom Google',
        type: IdentityProviderTypes.GOOGLE,
      };

      const result = resolveAuthFlowHandle({
        hasUsernamePassword: false,
        identityProviders: [customProvider],
      });

      // Should only care about the type
      expect(result).toBe(AUTH_FLOW_HANDLES.GOOGLE);
    });
  });

  describe('Type Extraction Logic', () => {
    it('should correctly extract provider types from identity providers', () => {
      const googleProvider: IdentityProvider = {
        id: 'google-123',
        name: 'Google',
        type: IdentityProviderTypes.GOOGLE,
      };

      const githubProvider: IdentityProvider = {
        id: 'github-456',
        name: 'GitHub',
        type: IdentityProviderTypes.GITHUB,
      };

      const result = resolveAuthFlowHandle({
        hasUsernamePassword: false,
        identityProviders: [googleProvider, githubProvider],
      });

      // Should detect both types and return appropriate flow (social-only)
      expect(result).toBe(AUTH_FLOW_HANDLES.GOOGLE_GITHUB);
    });

    it('should use type property from IdentityProvider', () => {
      const provider: IdentityProvider = {
        id: 'test-123',
        name: 'Test Provider',
        type: IdentityProviderTypes.GOOGLE,
      };

      const result = resolveAuthFlowHandle({
        hasUsernamePassword: false,
        identityProviders: [provider],
      });

      expect(result).toBe(AUTH_FLOW_HANDLES.GOOGLE);
    });
  });

  describe('Comprehensive Scenario Coverage', () => {
    it('should handle all valid two-method combinations', () => {
      const scenarios = [
        {
          config: {hasUsernamePassword: true, identityProviders: []},
          expected: AUTH_FLOW_HANDLES.BASIC,
        },
        {
          config: {
            hasUsernamePassword: true,
            identityProviders: [{id: '1', name: 'Google', type: IdentityProviderTypes.GOOGLE} as IdentityProvider],
          },
          expected: AUTH_FLOW_HANDLES.BASIC_GOOGLE,
        },
        {
          config: {
            hasUsernamePassword: true,
            identityProviders: [{id: '2', name: 'GitHub', type: IdentityProviderTypes.GITHUB} as IdentityProvider],
          },
          expected: AUTH_FLOW_HANDLES.BASIC_GITHUB,
        },
        {
          config: {
            hasUsernamePassword: false,
            identityProviders: [{id: '1', name: 'Google', type: IdentityProviderTypes.GOOGLE} as IdentityProvider],
          },
          expected: AUTH_FLOW_HANDLES.GOOGLE,
        },
        {
          config: {
            hasUsernamePassword: false,
            identityProviders: [{id: '2', name: 'GitHub', type: IdentityProviderTypes.GITHUB} as IdentityProvider],
          },
          expected: AUTH_FLOW_HANDLES.GITHUB,
        },
      ];

      scenarios.forEach(({config, expected}) => {
        const result = resolveAuthFlowHandle(config);
        expect(result).toBe(expected);
      });
    });

    it('should handle all valid three-method combinations', () => {
      const googleProvider: IdentityProvider = {
        id: 'google-123',
        name: 'Google',
        type: IdentityProviderTypes.GOOGLE,
      };

      const githubProvider: IdentityProvider = {
        id: 'github-456',
        name: 'GitHub',
        type: IdentityProviderTypes.GITHUB,
      };

      const scenarios = [
        {
          config: {hasUsernamePassword: true, identityProviders: [googleProvider, githubProvider]},
          expected: AUTH_FLOW_HANDLES.BASIC_GOOGLE_GITHUB,
        },
        {
          config: {hasUsernamePassword: false, identityProviders: [googleProvider, githubProvider]},
          expected: AUTH_FLOW_HANDLES.GOOGLE_GITHUB,
        },
      ];

      scenarios.forEach(({config, expected}) => {
        const result = resolveAuthFlowHandle(config);
        expect(result).toBe(expected);
      });
    });
  });

  describe('Return Value Verification', () => {
    it('should always return a string', () => {
      const configs = [
        {hasUsernamePassword: true, identityProviders: []},
        {hasUsernamePassword: false, identityProviders: []},
        {
          hasUsernamePassword: true,
          identityProviders: [{id: '1', name: 'Google', type: IdentityProviderTypes.GOOGLE} as IdentityProvider],
        },
      ];

      configs.forEach((config) => {
        const result = resolveAuthFlowHandle(config);
        expect(typeof result).toBe('string');
        expect(result.length).toBeGreaterThan(0);
      });
    });

    it('should always return a value from AUTH_FLOW_HANDLES', () => {
      const configs = [
        {hasUsernamePassword: true, identityProviders: []},
        {hasUsernamePassword: false, identityProviders: []},
        {
          hasUsernamePassword: true,
          identityProviders: [{id: '1', name: 'Google', type: IdentityProviderTypes.GOOGLE} as IdentityProvider],
        },
        {
          hasUsernamePassword: false,
          identityProviders: [{id: '1', name: 'Google', type: IdentityProviderTypes.GOOGLE} as IdentityProvider],
        },
        {
          hasUsernamePassword: false,
          identityProviders: [{id: '2', name: 'GitHub', type: IdentityProviderTypes.GITHUB} as IdentityProvider],
        },
      ];

      const validFlowIds = Object.values(AUTH_FLOW_HANDLES);

      configs.forEach((config) => {
        const result = resolveAuthFlowHandle(config);
        expect(validFlowIds).toContain(result);
      });
    });

    it('should never return undefined or null', () => {
      const configs = [
        {hasUsernamePassword: true, identityProviders: []},
        {hasUsernamePassword: false, identityProviders: []},
      ];

      configs.forEach((config) => {
        const result = resolveAuthFlowHandle(config);
        expect(result).not.toBeUndefined();
        expect(result).not.toBeNull();
      });
    });
  });

  describe('Identity Provider Type Constants', () => {
    it('should use IdentityProviderTypes.GOOGLE constant', () => {
      expect(IdentityProviderTypes.GOOGLE).toBe('GOOGLE');
    });

    it('should use IdentityProviderTypes.GITHUB constant', () => {
      expect(IdentityProviderTypes.GITHUB).toBe('GITHUB');
    });
  });
});
