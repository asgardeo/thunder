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

import {describe, it, expect, vi} from 'vitest';
import {render} from '@thunder/test-utils/browser';
import {page} from 'vitest/browser';
import OAuth2ConfigSection from '../OAuth2ConfigSection';
import type {OAuth2Config} from '../../../../models/oauth';

describe('OAuth2ConfigSection', () => {
  describe('Rendering', () => {
    it('should return null when oauth2Config is not provided', async () => {
      const {container} = await render(<OAuth2ConfigSection />);

      expect(container.firstChild).toBeNull();
    });

    it('should return null when oauth2Config is undefined', async () => {
      const {container} = await render(<OAuth2ConfigSection oauth2Config={undefined} />);

      expect(container.firstChild).toBeNull();
    });

    it('should render OAuth2 config section with all elements', async () => {
      const oauth2Config: OAuth2Config = {
        grant_types: ['authorization_code', 'refresh_token'],
        response_types: ['code'],
        pkce_required: true,
        public_client: false,
      };

      await render(<OAuth2ConfigSection oauth2Config={oauth2Config} />);

      await expect.element(page.getByText('OAuth2 Configuration', {exact: true})).toBeInTheDocument();
      await expect.element(page.getByText('These OAuth 2.0 settings are read-only and were configured during application creation.')).toBeInTheDocument();
    });
  });

  describe('Grant Types Display', () => {
    it('should display all grant types as chips', async () => {
      const oauth2Config: OAuth2Config = {
        grant_types: ['authorization_code', 'refresh_token', 'client_credentials'],
        response_types: ['code'],
        pkce_required: false,
        public_client: false,
      };

      await render(<OAuth2ConfigSection oauth2Config={oauth2Config} />);

      await expect.element(page.getByText('Grant Types', {exact: true})).toBeInTheDocument();
      await expect.element(page.getByText('authorization_code', {exact: true})).toBeInTheDocument();
      await expect.element(page.getByText('refresh_token', {exact: true})).toBeInTheDocument();
      await expect.element(page.getByText('client_credentials', {exact: true})).toBeInTheDocument();
      await expect.element(page.getByText('OAuth 2.0 flows this application can use (e.g., authorization_code, client_credentials, refresh_token).')).toBeInTheDocument();
    });

    it('should handle single grant type', async () => {
      const oauth2Config: OAuth2Config = {
        grant_types: ['authorization_code'],
        response_types: ['code'],
        pkce_required: false,
        public_client: false,
      };

      await render(<OAuth2ConfigSection oauth2Config={oauth2Config} />);

      await expect.element(page.getByText('authorization_code', {exact: true})).toBeInTheDocument();
      await expect.element(page.getByText('refresh_token', {exact: true})).not.toBeInTheDocument();
    });

    it('should handle empty grant types array', async () => {
      const oauth2Config: OAuth2Config = {
        grant_types: [],
        response_types: ['code'],
        pkce_required: false,
        public_client: false,
      };

      await render(<OAuth2ConfigSection oauth2Config={oauth2Config} />);

      await expect.element(page.getByText('Grant Types', {exact: true})).toBeInTheDocument();
      await expect.element(page.getByRole('button')).not.toBeInTheDocument();
    });
  });

  describe('Response Types Display', () => {
    it('should display all response types as chips', async () => {
      const oauth2Config: OAuth2Config = {
        grant_types: ['authorization_code'],
        response_types: ['code', 'token'],
        pkce_required: false,
        public_client: false,
      };

      await render(<OAuth2ConfigSection oauth2Config={oauth2Config} />);

      await expect.element(page.getByText('Response Types', {exact: true})).toBeInTheDocument();
      await expect.element(page.getByText('code', {exact: true})).toBeInTheDocument();
      await expect.element(page.getByText('token', {exact: true})).toBeInTheDocument();
    });

    it('should handle single response type', async () => {
      const oauth2Config: OAuth2Config = {
        grant_types: ['authorization_code'],
        response_types: ['code'],
        pkce_required: false,
        public_client: false,
      };

      await render(<OAuth2ConfigSection oauth2Config={oauth2Config} />);

      await expect.element(page.getByText('code', {exact: true})).toBeInTheDocument();
    });

    it('should handle empty response types array', async () => {
      const oauth2Config: OAuth2Config = {
        grant_types: ['authorization_code'],
        response_types: [],
        pkce_required: false,
        public_client: false,
      };

      await render(<OAuth2ConfigSection oauth2Config={oauth2Config} />);

      await expect.element(page.getByText('Response Types', {exact: true})).toBeInTheDocument();
    });
  });

  describe('Public Client Status', () => {
    it('should display public client as yes when true', async () => {
      const oauth2Config: OAuth2Config = {
        grant_types: ['authorization_code'],
        response_types: ['code'],
        pkce_required: false,
        public_client: true,
      };

      await render(<OAuth2ConfigSection oauth2Config={oauth2Config} />);

      await expect.element(page.getByText('Public Client', {exact: true})).toBeInTheDocument();
      await expect.element(page.getByText('Yes', {exact: true})).toBeInTheDocument();
      await expect.element(page.getByText('This is a public client (SPA, mobile app) that cannot securely store credentials.')).toBeInTheDocument();
    });

    it('should display public client as no when false', async () => {
      const oauth2Config: OAuth2Config = {
        grant_types: ['authorization_code'],
        response_types: ['code'],
        pkce_required: false,
        public_client: false,
      };

      await render(<OAuth2ConfigSection oauth2Config={oauth2Config} />);

      await expect.element(page.getByText('Public Client', {exact: true})).toBeInTheDocument();
      // Both public_client and pkce_required are false, so "No" appears twice - check description instead
      await expect.element(page.getByText('This is a confidential client (server-side app) that can securely store credentials.')).toBeInTheDocument();
    });

    it('should handle undefined public_client as false', async () => {
      const oauth2Config: OAuth2Config = {
        grant_types: ['authorization_code'],
        response_types: ['code'],
        pkce_required: false,
      };

      await render(<OAuth2ConfigSection oauth2Config={oauth2Config} />);

      // public_client defaults to false - verify via description text
      await expect.element(page.getByText('This is a confidential client (server-side app) that can securely store credentials.')).toBeInTheDocument();
    });
  });

  describe('PKCE Requirement Status', () => {
    it('should display PKCE as required when true', async () => {
      const oauth2Config: OAuth2Config = {
        grant_types: ['authorization_code'],
        response_types: ['code'],
        pkce_required: true,
        public_client: false,
      };

      await render(<OAuth2ConfigSection oauth2Config={oauth2Config} />);

      await expect.element(page.getByText('PKCE Required', {exact: true})).toBeInTheDocument();
      await expect.element(page.getByText('Yes', {exact: true})).toBeInTheDocument();
      await expect.element(page.getByText('PKCE is required for authorization code flow, providing additional security.')).toBeInTheDocument();
    });

    it('should display PKCE as not required when false', async () => {
      const oauth2Config: OAuth2Config = {
        grant_types: ['authorization_code'],
        response_types: ['code'],
        pkce_required: false,
        public_client: false,
      };

      await render(<OAuth2ConfigSection oauth2Config={oauth2Config} />);

      await expect.element(page.getByText('PKCE Required', {exact: true})).toBeInTheDocument();
      // Both public_client and pkce_required are false, so "No" appears twice - check description instead
      await expect.element(page.getByText('PKCE is not required. Consider enabling for public clients (SPAs, mobile apps).')).toBeInTheDocument();
    });

    it('should handle undefined pkce_required as false', async () => {
      const oauth2Config: OAuth2Config = {
        grant_types: ['authorization_code'],
        response_types: ['code'],
        public_client: false,
      };

      await render(<OAuth2ConfigSection oauth2Config={oauth2Config} />);

      // pkce_required defaults to false - verify via description text
      await expect.element(page.getByText('PKCE is not required. Consider enabling for public clients (SPAs, mobile apps).')).toBeInTheDocument();
    });
  });

  describe('Layout and Styling', () => {
    it('should render grant type chips with correct styling', async () => {
      const oauth2Config: OAuth2Config = {
        grant_types: ['authorization_code'],
        response_types: ['code'],
        pkce_required: false,
        public_client: false,
      };

      await render(<OAuth2ConfigSection oauth2Config={oauth2Config} />);

      const chip = page.getByText('authorization_code', {exact: true}).element().closest('.MuiChip-root');
      expect(chip).toHaveClass('MuiChip-outlined');
      expect(chip).toHaveClass('MuiChip-sizeSmall');
    });

    it('should render in a Stack with proper spacing', async () => {
      const oauth2Config: OAuth2Config = {
        grant_types: ['authorization_code'],
        response_types: ['code'],
        pkce_required: false,
        public_client: false,
      };

      const {container} = await render(<OAuth2ConfigSection oauth2Config={oauth2Config} />);

      const stack = container.querySelector('.MuiStack-root');
      expect(stack).toBeInTheDocument();
    });
  });

  describe('Edge Cases', () => {
    it('should handle minimal OAuth2 config', async () => {
      const oauth2Config: OAuth2Config = {
        grant_types: ['authorization_code'],
        response_types: ['code'],
      };

      await render(<OAuth2ConfigSection oauth2Config={oauth2Config} />);

      await expect.element(page.getByText('OAuth2 Configuration', {exact: true})).toBeInTheDocument();
      await expect.element(page.getByText('authorization_code', {exact: true})).toBeInTheDocument();
      await expect.element(page.getByText('code', {exact: true})).toBeInTheDocument();
    });

    it('should handle multiple grant types correctly', async () => {
      const oauth2Config: OAuth2Config = {
        grant_types: [
          'authorization_code',
          'refresh_token',
          'client_credentials',
          'urn:ietf:params:oauth:grant-type:token-exchange',
        ],
        response_types: ['code'],
        pkce_required: true,
        public_client: false,
      };

      await render(<OAuth2ConfigSection oauth2Config={oauth2Config} />);

      await expect.element(page.getByText('authorization_code', {exact: true})).toBeInTheDocument();
      await expect.element(page.getByText('refresh_token', {exact: true})).toBeInTheDocument();
      await expect.element(page.getByText('client_credentials', {exact: true})).toBeInTheDocument();
      await expect.element(page.getByText('urn:ietf:params:oauth:grant-type:token-exchange', {exact: true})).toBeInTheDocument();
    });
  });
});
