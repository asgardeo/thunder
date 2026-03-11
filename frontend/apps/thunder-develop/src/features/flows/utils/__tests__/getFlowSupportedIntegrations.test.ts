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
import getFlowSupportedIntegrations from '../getFlowSupportedIntegrations';
import {AuthenticatorTypes} from '../../../integrations/models/authenticators';

describe('getFlowSupportedIntegrations', () => {
  describe('Single Integration Detection', () => {
    it('should detect basic auth from handle containing "basic"', async () => {
      const result = getFlowSupportedIntegrations('basic-flow');
      expect(result).toContain(AuthenticatorTypes.BASIC_AUTH);
    });

    it('should detect google from handle containing "google"', async () => {
      const result = getFlowSupportedIntegrations('google-flow');
      expect(result).toContain('google');
    });

    it('should detect github from handle containing "github"', async () => {
      const result = getFlowSupportedIntegrations('github-flow');
      expect(result).toContain('github');
    });

    it('should detect sms from handle containing "sms"', async () => {
      const result = getFlowSupportedIntegrations('sms-flow');
      expect(result).toContain('sms-otp');
    });

    it('should detect passkey from handle containing "passkey"', async () => {
      const result = getFlowSupportedIntegrations('passkey-flow');
      expect(result).toContain(AuthenticatorTypes.PASSKEY);
    });
  });

  describe('Multiple Integration Detection', () => {
    it('should detect basic and google from combined handle', async () => {
      const result = getFlowSupportedIntegrations('basic-google-flow');
      expect(result).toContain(AuthenticatorTypes.BASIC_AUTH);
      expect(result).toContain('google');
      expect(result).toHaveLength(2);
    });

    it('should detect basic and github from combined handle', async () => {
      const result = getFlowSupportedIntegrations('basic-github-flow');
      expect(result).toContain(AuthenticatorTypes.BASIC_AUTH);
      expect(result).toContain('github');
      expect(result).toHaveLength(2);
    });

    it('should detect all three: basic, google, and github', async () => {
      const result = getFlowSupportedIntegrations('basic-google-github-flow');
      expect(result).toContain(AuthenticatorTypes.BASIC_AUTH);
      expect(result).toContain('google');
      expect(result).toContain('github');
      expect(result).toHaveLength(3);
    });

    it('should detect basic, passkey, and google', async () => {
      const result = getFlowSupportedIntegrations('basic-passkey-google-flow');
      expect(result).toContain(AuthenticatorTypes.BASIC_AUTH);
      expect(result).toContain(AuthenticatorTypes.PASSKEY);
      expect(result).toContain('google');
      expect(result).toHaveLength(3);
    });

    it('should detect basic and sms from combined handle', async () => {
      const result = getFlowSupportedIntegrations('basic-sms-otp-flow');
      expect(result).toContain(AuthenticatorTypes.BASIC_AUTH);
      expect(result).toContain('sms-otp');
      expect(result).toHaveLength(2);
    });
  });

  describe('No Integration Detection', () => {
    it('should return empty array for handle with no recognized integrations', async () => {
      const result = getFlowSupportedIntegrations('custom-flow');
      expect(result).toEqual([]);
    });

    it('should return empty array for empty handle', async () => {
      const result = getFlowSupportedIntegrations('');
      expect(result).toEqual([]);
    });
  });

  describe('Case Sensitivity', () => {
    it('should detect basic in lowercase handle', async () => {
      const result = getFlowSupportedIntegrations('basic-auth-flow');
      expect(result).toContain(AuthenticatorTypes.BASIC_AUTH);
    });

    it('should handle mixed case handles', async () => {
      // The function uses includes which is case-sensitive
      // Testing actual behavior
      const result = getFlowSupportedIntegrations('Basic-flow');
      // Should not contain basic_auth since 'Basic' !== 'basic'
      expect(result).not.toContain(AuthenticatorTypes.BASIC_AUTH);
    });
  });

  describe('Edge Cases', () => {
    it('should handle handle with integration name at start', async () => {
      const result = getFlowSupportedIntegrations('googleauth');
      expect(result).toContain('google');
    });

    it('should handle handle with integration name at end', async () => {
      const result = getFlowSupportedIntegrations('authgoogle');
      expect(result).toContain('google');
    });

    it('should handle handle with integration name in middle', async () => {
      const result = getFlowSupportedIntegrations('my-google-auth-flow');
      expect(result).toContain('google');
    });

    it('should handle partial matches', async () => {
      // 'basics' should match 'basic'
      const result = getFlowSupportedIntegrations('basics-flow');
      expect(result).toContain(AuthenticatorTypes.BASIC_AUTH);
    });
  });

  describe('Return Value Structure', () => {
    it('should return an array', async () => {
      const result = getFlowSupportedIntegrations('any-flow');
      expect(Array.isArray(result)).toBe(true);
    });

    it('should not return duplicates', async () => {
      // Even with multiple occurrences, should only add once
      const result = getFlowSupportedIntegrations('basic-basic-flow');
      const basicCount = result.filter((i) => i === AuthenticatorTypes.BASIC_AUTH).length;
      expect(basicCount).toBe(1);
    });
  });
});
