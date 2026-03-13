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
import {render, screen} from '@testing-library/react';
import OAuth2ConfigSection from '../OAuth2ConfigSection';
import type {OAuth2Config} from '../../../../models/oauth';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
}));

describe('OAuth2ConfigSection', () => {
  describe('Rendering', () => {
    it('should return null when oauth2Config is not provided', () => {
      const {container} = render(<OAuth2ConfigSection />);

      expect(container.firstChild).toBeNull();
    });

    it('should return null when oauth2Config is undefined', () => {
      const {container} = render(<OAuth2ConfigSection oauth2Config={undefined} />);

      expect(container.firstChild).toBeNull();
    });

    it('should render OAuth2 config section with all elements', () => {
      const oauth2Config: OAuth2Config = {
        grant_types: ['authorization_code', 'refresh_token'],
        response_types: ['code'],
        pkce_required: true,
        public_client: false,
      };

      render(<OAuth2ConfigSection oauth2Config={oauth2Config} />);

      expect(screen.getByText('applications:edit.advanced.labels.oauth2Config')).toBeInTheDocument();
      expect(screen.getByText('applications:edit.advanced.oauth2Config.intro')).toBeInTheDocument();
    });
  });

  describe('Grant Types Display', () => {
    it('should display all grant types as chips', () => {
      const oauth2Config: OAuth2Config = {
        grant_types: ['authorization_code', 'refresh_token', 'client_credentials'],
        response_types: ['code'],
        pkce_required: false,
        public_client: false,
      };

      render(<OAuth2ConfigSection oauth2Config={oauth2Config} />);

      expect(screen.getByText('applications:edit.advanced.labels.grantTypes')).toBeInTheDocument();
      expect(screen.getByText('authorization_code')).toBeInTheDocument();
      expect(screen.getByText('refresh_token')).toBeInTheDocument();
      expect(screen.getByText('client_credentials')).toBeInTheDocument();
      expect(screen.getByText('applications:edit.advanced.grantTypes.hint')).toBeInTheDocument();
    });

    it('should handle single grant type', () => {
      const oauth2Config: OAuth2Config = {
        grant_types: ['authorization_code'],
        response_types: ['code'],
        pkce_required: false,
        public_client: false,
      };

      render(<OAuth2ConfigSection oauth2Config={oauth2Config} />);

      expect(screen.getByText('authorization_code')).toBeInTheDocument();
      expect(screen.queryByText('refresh_token')).not.toBeInTheDocument();
    });

    it('should handle empty grant types array', () => {
      const oauth2Config: OAuth2Config = {
        grant_types: [],
        response_types: ['code'],
        pkce_required: false,
        public_client: false,
      };

      render(<OAuth2ConfigSection oauth2Config={oauth2Config} />);

      expect(screen.getByText('applications:edit.advanced.labels.grantTypes')).toBeInTheDocument();
      expect(screen.queryByRole('button')).not.toBeInTheDocument();
    });
  });

  describe('Response Types Display', () => {
    it('should display all response types as chips', () => {
      const oauth2Config: OAuth2Config = {
        grant_types: ['authorization_code'],
        response_types: ['code', 'token'],
        pkce_required: false,
        public_client: false,
      };

      render(<OAuth2ConfigSection oauth2Config={oauth2Config} />);

      expect(screen.getByText('applications:edit.advanced.labels.responseTypes')).toBeInTheDocument();
      expect(screen.getByText('code')).toBeInTheDocument();
      expect(screen.getByText('token')).toBeInTheDocument();
    });

    it('should handle single response type', () => {
      const oauth2Config: OAuth2Config = {
        grant_types: ['authorization_code'],
        response_types: ['code'],
        pkce_required: false,
        public_client: false,
      };

      render(<OAuth2ConfigSection oauth2Config={oauth2Config} />);

      expect(screen.getByText('code')).toBeInTheDocument();
    });

    it('should handle empty response types array', () => {
      const oauth2Config: OAuth2Config = {
        grant_types: ['authorization_code'],
        response_types: [],
        pkce_required: false,
        public_client: false,
      };

      render(<OAuth2ConfigSection oauth2Config={oauth2Config} />);

      expect(screen.getByText('applications:edit.advanced.labels.responseTypes')).toBeInTheDocument();
    });
  });

  describe('Public Client Status', () => {
    it('should display public client as yes when true', () => {
      const oauth2Config: OAuth2Config = {
        grant_types: ['authorization_code'],
        response_types: ['code'],
        pkce_required: false,
        public_client: true,
      };

      render(<OAuth2ConfigSection oauth2Config={oauth2Config} />);

      expect(screen.getByText('applications:edit.advanced.labels.publicClient')).toBeInTheDocument();
      expect(screen.getByText('applications:edit.advanced.publicClient.yes')).toBeInTheDocument();
      expect(screen.getByText('applications:edit.advanced.publicClient.public')).toBeInTheDocument();
    });

    it('should display public client as no when false', () => {
      const oauth2Config: OAuth2Config = {
        grant_types: ['authorization_code'],
        response_types: ['code'],
        pkce_required: false,
        public_client: false,
      };

      render(<OAuth2ConfigSection oauth2Config={oauth2Config} />);

      expect(screen.getByText('applications:edit.advanced.labels.publicClient')).toBeInTheDocument();
      expect(screen.getByText('applications:edit.advanced.publicClient.no')).toBeInTheDocument();
      expect(screen.getByText('applications:edit.advanced.publicClient.confidential')).toBeInTheDocument();
    });

    it('should handle undefined public_client as false', () => {
      const oauth2Config: OAuth2Config = {
        grant_types: ['authorization_code'],
        response_types: ['code'],
        pkce_required: false,
      };

      render(<OAuth2ConfigSection oauth2Config={oauth2Config} />);

      expect(screen.getByText('applications:edit.advanced.publicClient.no')).toBeInTheDocument();
    });
  });

  describe('PKCE Requirement Status', () => {
    it('should display PKCE as required when true', () => {
      const oauth2Config: OAuth2Config = {
        grant_types: ['authorization_code'],
        response_types: ['code'],
        pkce_required: true,
        public_client: false,
      };

      render(<OAuth2ConfigSection oauth2Config={oauth2Config} />);

      expect(screen.getByText('applications:edit.advanced.labels.pkceRequired')).toBeInTheDocument();
      expect(screen.getByText('applications:edit.advanced.pkce.yes')).toBeInTheDocument();
      expect(screen.getByText('applications:edit.advanced.pkce.enabled')).toBeInTheDocument();
    });

    it('should display PKCE as not required when false', () => {
      const oauth2Config: OAuth2Config = {
        grant_types: ['authorization_code'],
        response_types: ['code'],
        pkce_required: false,
        public_client: false,
      };

      render(<OAuth2ConfigSection oauth2Config={oauth2Config} />);

      expect(screen.getByText('applications:edit.advanced.labels.pkceRequired')).toBeInTheDocument();
      expect(screen.getByText('applications:edit.advanced.pkce.no')).toBeInTheDocument();
      expect(screen.getByText('applications:edit.advanced.pkce.disabled')).toBeInTheDocument();
    });

    it('should handle undefined pkce_required as false', () => {
      const oauth2Config: OAuth2Config = {
        grant_types: ['authorization_code'],
        response_types: ['code'],
        public_client: false,
      };

      render(<OAuth2ConfigSection oauth2Config={oauth2Config} />);

      expect(screen.getByText('applications:edit.advanced.pkce.no')).toBeInTheDocument();
    });
  });

  describe('Layout and Styling', () => {
    it('should render grant type chips with correct styling', () => {
      const oauth2Config: OAuth2Config = {
        grant_types: ['authorization_code'],
        response_types: ['code'],
        pkce_required: false,
        public_client: false,
      };

      render(<OAuth2ConfigSection oauth2Config={oauth2Config} />);

      const chip = screen.getByText('authorization_code').closest('.MuiChip-root');
      expect(chip).toHaveClass('MuiChip-outlined');
      expect(chip).toHaveClass('MuiChip-sizeSmall');
    });

    it('should render in a Stack with proper spacing', () => {
      const oauth2Config: OAuth2Config = {
        grant_types: ['authorization_code'],
        response_types: ['code'],
        pkce_required: false,
        public_client: false,
      };

      const {container} = render(<OAuth2ConfigSection oauth2Config={oauth2Config} />);

      const stack = container.querySelector('.MuiStack-root');
      expect(stack).toBeInTheDocument();
    });
  });

  describe('Edge Cases', () => {
    it('should handle minimal OAuth2 config', () => {
      const oauth2Config: OAuth2Config = {
        grant_types: ['authorization_code'],
        response_types: ['code'],
      };

      render(<OAuth2ConfigSection oauth2Config={oauth2Config} />);

      expect(screen.getByText('applications:edit.advanced.labels.oauth2Config')).toBeInTheDocument();
      expect(screen.getByText('authorization_code')).toBeInTheDocument();
      expect(screen.getByText('code')).toBeInTheDocument();
    });

    it('should handle multiple grant types correctly', () => {
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

      render(<OAuth2ConfigSection oauth2Config={oauth2Config} />);

      expect(screen.getByText('authorization_code')).toBeInTheDocument();
      expect(screen.getByText('refresh_token')).toBeInTheDocument();
      expect(screen.getByText('client_credentials')).toBeInTheDocument();
      expect(screen.getByText('urn:ietf:params:oauth:grant-type:token-exchange')).toBeInTheDocument();
    });
  });
});
