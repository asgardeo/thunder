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

import {render, screen} from '@testing-library/react';
import {describe, it, expect, vi, beforeEach} from 'vitest';
import IntegrationGuides from '../IntegrationGuides';
import type {Application} from '../../../../models/application';
import type {OAuth2Config} from '../../../../models/oauth';
import getIntegrationGuidesForTemplate from '../../../../utils/getIntegrationGuidesForTemplate';
import IntegrationGuide from '../IntegrationGuide';

// Mock the integration guide utility
vi.mock('../../../../utils/getIntegrationGuidesForTemplate', () => ({
  default: vi.fn(),
}));

// Mock the child component
vi.mock('../IntegrationGuide', () => ({
  default: vi.fn(() => <div data-testid="integration-guide">Integration Guide</div>),
}));

const mockApplication: Application = {
  id: 'app-123',
  name: 'Test Application',
  template: 'react',
  description: 'Test description',
  allowed_user_types: ['admin', 'user'],
};

const mockOAuth2Config: OAuth2Config = {
  client_id: 'client-123',
  client_secret: 'secret-456',
  grant_types: ['authorization_code'],
  response_types: ['code'],
  pkce_required: true,
  public_client: false,
  redirect_uris: ['https://example.com/callback'],
};

const mockIntegrationGuides = {
  INBUILT: {
    llm_prompt: {
      id: 'llm-1',
      title: 'Use AI to integrate',
      description: 'Copy prompt for AI',
      type: 'llm' as const,
      icon: 'sparkles',
      content: 'LLM prompt content',
    },
    manual_steps: [
      {
        step: 1,
        title: 'Install dependencies',
        description: 'Install required packages',
        code: {
          language: 'bash',
          content: 'npm install',
        },
      },
    ],
  },
};

describe('IntegrationGuides', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('should render integration guide when guides are available', () => {
      vi.mocked(getIntegrationGuidesForTemplate).mockReturnValue(mockIntegrationGuides);

      render(<IntegrationGuides application={mockApplication} oauth2Config={mockOAuth2Config} />);

      expect(screen.getByTestId('integration-guide')).toBeInTheDocument();
    });

    it('should render fallback message when no guides are available', () => {
      vi.mocked(getIntegrationGuidesForTemplate).mockReturnValue(null);

      render(<IntegrationGuides application={mockApplication} />);

      expect(screen.getByText('No integration guides available for this application type.')).toBeInTheDocument();
    });
  });

  describe('Props Propagation', () => {
    it('should pass clientId from oauth2Config to IntegrationGuide', () => {
      vi.mocked(getIntegrationGuidesForTemplate).mockReturnValue(mockIntegrationGuides);

      render(<IntegrationGuides application={mockApplication} oauth2Config={mockOAuth2Config} />);

      expect(IntegrationGuide).toHaveBeenCalledWith(
        {
          clientId: 'client-123',
          applicationId: 'app-123',
          integrationGuides: mockIntegrationGuides,
          templateId: 'react',
        },
        undefined,
      );
    });

    it('should pass empty clientId when oauth2Config is not provided', () => {
      vi.mocked(getIntegrationGuidesForTemplate).mockReturnValue(mockIntegrationGuides);

      render(<IntegrationGuides application={mockApplication} />);

      expect(IntegrationGuide).toHaveBeenCalledWith(
        {
          clientId: '',
          applicationId: 'app-123',
          integrationGuides: mockIntegrationGuides,
          templateId: 'react',
        },
        undefined,
      );
    });

    it('should pass applicationId to IntegrationGuide', () => {
      vi.mocked(getIntegrationGuidesForTemplate).mockReturnValue(mockIntegrationGuides);

      render(<IntegrationGuides application={mockApplication} />);

      expect(IntegrationGuide).toHaveBeenCalledWith(
        {
          clientId: '',
          applicationId: 'app-123',
          integrationGuides: mockIntegrationGuides,
          templateId: 'react',
        },
        undefined,
      );
    });

    it('should pass integrationGuides to IntegrationGuide', () => {
      vi.mocked(getIntegrationGuidesForTemplate).mockReturnValue(mockIntegrationGuides);

      render(<IntegrationGuides application={mockApplication} />);

      expect(IntegrationGuide).toHaveBeenCalledWith(
        {
          clientId: '',
          applicationId: 'app-123',
          integrationGuides: mockIntegrationGuides,
          templateId: 'react',
        },
        undefined,
      );
    });

    it('should pass templateId from application to IntegrationGuide', () => {
      vi.mocked(getIntegrationGuidesForTemplate).mockReturnValue(mockIntegrationGuides);

      render(<IntegrationGuides application={mockApplication} />);

      expect(IntegrationGuide).toHaveBeenCalledWith(
        {
          clientId: '',
          applicationId: 'app-123',
          integrationGuides: mockIntegrationGuides,
          templateId: 'react',
        },
        undefined,
      );
    });
  });

  describe('Template Utility Integration', () => {
    it('should call getIntegrationGuidesForTemplate with template from application', () => {
      vi.mocked(getIntegrationGuidesForTemplate).mockReturnValue(mockIntegrationGuides);

      render(<IntegrationGuides application={mockApplication} />);

      expect(getIntegrationGuidesForTemplate).toHaveBeenCalledWith('react');
    });

    it('should call getIntegrationGuidesForTemplate with empty string when template is not defined', () => {
      vi.mocked(getIntegrationGuidesForTemplate).mockReturnValue(null);
      const appWithoutTemplate = {...mockApplication, template: undefined};

      render(<IntegrationGuides application={appWithoutTemplate} />);

      expect(getIntegrationGuidesForTemplate).toHaveBeenCalledWith('');
    });
  });
});
