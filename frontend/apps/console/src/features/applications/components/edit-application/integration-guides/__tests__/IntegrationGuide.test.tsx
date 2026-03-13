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
import IntegrationGuide from '../IntegrationGuide';
import type {IntegrationGuides} from '../../../../models/application-templates';
import TechnologyGuide from '../TechnologyGuide';

// Mock the TechnologyGuide component
vi.mock('../TechnologyGuide', () => ({
  default: vi.fn(() => <div data-testid="technology-guide">Technology Guide</div>),
}));

const mockIntegrationGuides: IntegrationGuides = {
  inbuilt: {
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

describe('IntegrationGuide', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('should render TechnologyGuide when integrationGuides are provided', () => {
      render(
        <IntegrationGuide
          clientId="client-123"
          applicationId="app-123"
          integrationGuides={mockIntegrationGuides}
          templateId="react"
        />,
      );

      expect(screen.getByTestId('technology-guide')).toBeInTheDocument();
    });

    it('should not render TechnologyGuide when integrationGuides is null', () => {
      render(
        <IntegrationGuide clientId="client-123" applicationId="app-123" integrationGuides={null} templateId="react" />,
      );

      expect(screen.queryByTestId('technology-guide')).not.toBeInTheDocument();
    });

    it('should not render TechnologyGuide when integrationGuides is undefined', () => {
      render(<IntegrationGuide clientId="client-123" applicationId="app-123" templateId="react" />);

      expect(screen.queryByTestId('technology-guide')).not.toBeInTheDocument();
    });
  });

  describe('Props Propagation', () => {
    it('should pass all props to TechnologyGuide', () => {
      render(
        <IntegrationGuide
          clientId="client-123"
          applicationId="app-123"
          integrationGuides={mockIntegrationGuides}
          templateId="react"
        />,
      );

      expect(TechnologyGuide).toHaveBeenCalledWith(
        {
          guides: mockIntegrationGuides,
          templateId: 'react',
          clientId: 'client-123',
          applicationId: 'app-123',
        },
        undefined,
      );
    });

    it('should pass empty string for clientId when not provided', () => {
      render(<IntegrationGuide applicationId="app-123" integrationGuides={mockIntegrationGuides} templateId="react" />);

      expect(TechnologyGuide).toHaveBeenCalledWith(
        expect.objectContaining({
          clientId: '',
        }),
        undefined,
      );
    });

    it('should pass null for applicationId when not provided', () => {
      render(<IntegrationGuide clientId="client-123" integrationGuides={mockIntegrationGuides} templateId="react" />);

      expect(TechnologyGuide).toHaveBeenCalledWith(
        expect.objectContaining({
          applicationId: undefined,
        }),
        undefined,
      );
    });

    it('should handle undefined templateId', () => {
      render(
        <IntegrationGuide clientId="client-123" applicationId="app-123" integrationGuides={mockIntegrationGuides} />,
      );

      expect(TechnologyGuide).toHaveBeenCalledWith(
        expect.objectContaining({
          templateId: null,
        }),
        undefined,
      );
    });
  });

  describe('Layout', () => {
    it('should render content in a Stack with proper styling', () => {
      const {container} = render(
        <IntegrationGuide
          clientId="client-123"
          applicationId="app-123"
          integrationGuides={mockIntegrationGuides}
          templateId="react"
        />,
      );

      const stack = container.querySelector('.MuiStack-root');
      expect(stack).toBeInTheDocument();
    });
  });
});
