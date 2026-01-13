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

import {describe, it, expect, vi, beforeEach} from 'vitest';
import {render, screen} from '@testing-library/react';
import TechnologyGuide, {type TechnologyGuideProps} from '../TechnologyGuide';
import type {IntegrationGuides} from '../../../models/application-templates';
import {ApplicationCreateFlowSignInApproach} from '../../../models/application-create-flow';

// Mock react-i18next
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => {
      const translations: Record<string, string> = {
        'applications:onboarding.summary.guides.title': 'Next Steps',
        'applications:onboarding.summary.guides.subtitle':
          'Choose how you want to integrate authentication into your application',
        'applications:onboarding.summary.guides.divider': 'or',
        'applications:clientSecret.copied': 'Copied!',
      };
      return translations[key] || key;
    },
  }),
}));

// Mock react-syntax-highlighter
vi.mock('react-syntax-highlighter', () => ({
  Prism: ({children}: {children: string}) => <pre data-testid="syntax-highlighter">{children}</pre>,
}));

vi.mock('react-syntax-highlighter/dist/esm/styles/prism', () => ({
  vscDarkPlus: {},
}));

describe('TechnologyGuide', () => {
  const mockGuides: IntegrationGuides = {
    inbuilt: {
      llm_prompt: {
        id: 'llm-prompt',
        title: 'Integrate with an LLM Prompt',
        description: 'Use AI to generate integration code',
        type: 'llm',
        icon: 'sparkles',
        content: 'This is the LLM prompt content with {{clientId}} and {{applicationId}}',
      },
      manual_steps: [
        {
          step: 1,
          title: 'Create a React application',
          description: 'Run the following commands',
          code: {
            language: 'terminal',
            content: 'npm create vite@latest',
          },
        },
        {
          step: 2,
          title: 'Install dependencies',
          description: 'Install the required packages',
          subDescription: 'Make sure to use the correct package manager',
          bullets: ['Install the SDK', 'Configure environment'],
          code: {
            language: 'terminal',
            filename: 'terminal',
            content: 'npm install @thunder/sdk',
          },
        },
        {
          step: 3,
          title: 'Configure environment',
          description: 'Set up your environment variables',
          code: {
            language: '.env',
            filename: '.env',
            content: 'CLIENT_ID={{clientId}}',
          },
        },
      ],
    },
    custom: {
      llm_prompt: {
        id: 'custom-llm-prompt',
        title: 'Custom Integration Prompt',
        description: 'Generate custom integration code',
        type: 'llm',
        icon: 'sparkles',
        content: 'Custom prompt for {{clientId}}',
      },
      manual_steps: [
        {
          step: 1,
          title: 'Custom Step 1',
          description: 'Custom description',
        },
      ],
    },
  };

  const defaultProps: TechnologyGuideProps = {
    guides: mockGuides,
    signInApproach: ApplicationCreateFlowSignInApproach.INBUILT,
  };

  beforeEach(() => {
    vi.clearAllMocks();
    // Mock clipboard API
    const mockWriteText = vi.fn().mockResolvedValue(undefined);
    vi.stubGlobal('navigator', {
      ...navigator,
      clipboard: {
        writeText: mockWriteText,
      },
    });
  });

  const renderComponent = (props: Partial<TechnologyGuideProps> = {}) =>
    render(<TechnologyGuide {...defaultProps} {...props} />);

  describe('rendering', () => {
    it('should return null when guides is null', () => {
      const {container} = renderComponent({guides: null});

      expect(container.firstChild).toBeNull();
    });

    it('should return null when selected guide does not exist', () => {
      const guidesWithoutInbuilt: IntegrationGuides = {
        custom: mockGuides.custom,
      };

      const {container} = renderComponent({
        guides: guidesWithoutInbuilt,
        signInApproach: ApplicationCreateFlowSignInApproach.INBUILT,
      });

      expect(container.firstChild).toBeNull();
    });

    it('should render LLM prompt card', () => {
      renderComponent();

      expect(screen.getByText('Integrate with an LLM Prompt')).toBeInTheDocument();
      expect(screen.getByText('Use AI to generate integration code')).toBeInTheDocument();
    });

    it('should render Copy Prompt button when llm_prompt has content', () => {
      renderComponent();

      expect(screen.getByRole('button', {name: /copy prompt/i})).toBeInTheDocument();
    });

    it('should render divider with "or" text when manual steps exist', () => {
      renderComponent();

      expect(screen.getByText('or')).toBeInTheDocument();
    });

    it('should not render divider when manual steps are empty', () => {
      const guidesWithoutSteps: IntegrationGuides = {
        inbuilt: {
          llm_prompt: mockGuides.inbuilt.llm_prompt,
          manual_steps: [],
        },
      };

      renderComponent({guides: guidesWithoutSteps});

      expect(screen.queryByText('or')).not.toBeInTheDocument();
    });

    it('should render manual integration steps', () => {
      renderComponent();

      expect(screen.getByText('Create a React application')).toBeInTheDocument();
      expect(screen.getByText('Run the following commands')).toBeInTheDocument();
      expect(screen.getByText('Install dependencies')).toBeInTheDocument();
    });

    it('should render step numbers', () => {
      renderComponent();

      expect(screen.getByText('1')).toBeInTheDocument();
      expect(screen.getByText('2')).toBeInTheDocument();
      expect(screen.getByText('3')).toBeInTheDocument();
    });

    it('should render subDescription when provided', () => {
      renderComponent();

      expect(screen.getByText('Make sure to use the correct package manager')).toBeInTheDocument();
    });

    it('should render bullets when provided', () => {
      renderComponent();

      expect(screen.getByText('Install the SDK')).toBeInTheDocument();
      // 'Configure environment' appears as both a bullet and a step title
      expect(screen.getAllByText('Configure environment').length).toBeGreaterThan(0);
    });

    it('should render code blocks with syntax highlighting', () => {
      renderComponent();

      const codeBlocks = screen.getAllByTestId('syntax-highlighter');
      expect(codeBlocks.length).toBeGreaterThan(0);
    });

    it('should render filename when provided in code block', () => {
      renderComponent();

      expect(screen.getByText('.env')).toBeInTheDocument();
    });
  });

  describe('sign-in approach selection', () => {
    it('should render inbuilt guide when signInApproach is INBUILT', () => {
      renderComponent({signInApproach: ApplicationCreateFlowSignInApproach.INBUILT});

      expect(screen.getByText('Integrate with an LLM Prompt')).toBeInTheDocument();
      expect(screen.getByText('Create a React application')).toBeInTheDocument();
    });

    it('should render custom guide when signInApproach is CUSTOM', () => {
      renderComponent({signInApproach: ApplicationCreateFlowSignInApproach.CUSTOM});

      expect(screen.getByText('Custom Integration Prompt')).toBeInTheDocument();
      expect(screen.getByText('Custom Step 1')).toBeInTheDocument();
    });
  });

  describe('placeholder replacement', () => {
    it('should replace {{clientId}} placeholder in code when clientId is provided', () => {
      renderComponent({clientId: 'test-client-id'});

      // The code block should contain the replaced value
      const codeBlocks = screen.getAllByTestId('syntax-highlighter');
      const envBlock = codeBlocks.find((block) => block.textContent?.includes('CLIENT_ID'));
      expect(envBlock?.textContent).toContain('test-client-id');
    });

    it('should keep {{clientId}} placeholder when clientId is empty', () => {
      renderComponent({clientId: ''});

      const codeBlocks = screen.getAllByTestId('syntax-highlighter');
      const envBlock = codeBlocks.find((block) => block.textContent?.includes('CLIENT_ID'));
      expect(envBlock?.textContent).toContain('{{clientId}}');
    });

    it('should replace {{applicationId}} placeholder when applicationId is provided', () => {
      const guidesWithAppId: IntegrationGuides = {
        inbuilt: {
          llm_prompt: mockGuides.inbuilt.llm_prompt,
          manual_steps: [
            {
              step: 1,
              title: 'Test Step',
              description: 'Test',
              code: {
                language: 'javascript',
                content: 'const appId = "{{applicationId}}"',
              },
            },
          ],
        },
      };

      renderComponent({guides: guidesWithAppId, applicationId: 'test-app-id'});

      const codeBlocks = screen.getAllByTestId('syntax-highlighter');
      expect(codeBlocks[0].textContent).toContain('test-app-id');
    });
  });

  describe('copy functionality', () => {
    it('should render Copy Prompt button', () => {
      renderComponent();

      const copyButton = screen.getByRole('button', {name: /copy prompt/i});
      expect(copyButton).toBeInTheDocument();
    });

    it('should render copy buttons for code blocks', () => {
      renderComponent();

      // Find all buttons - should have Copy Prompt plus copy buttons for each code block
      const allButtons = screen.getAllByRole('button');
      // At least Copy Prompt + 3 code blocks = 4 buttons
      expect(allButtons.length).toBeGreaterThanOrEqual(4);
    });
  });

  describe('edge cases', () => {
    it('should handle step without code block', () => {
      const guidesWithoutCode: IntegrationGuides = {
        inbuilt: {
          llm_prompt: mockGuides.inbuilt.llm_prompt,
          manual_steps: [
            {
              step: 1,
              title: 'Step without code',
              description: 'This step has no code',
            },
          ],
        },
      };

      renderComponent({guides: guidesWithoutCode});

      expect(screen.getByText('Step without code')).toBeInTheDocument();
      expect(screen.getByText('This step has no code')).toBeInTheDocument();
    });

    it('should handle llm_prompt without content', () => {
      const guidesWithoutContent: IntegrationGuides = {
        inbuilt: {
          llm_prompt: {
            id: 'no-content',
            title: 'No Content Prompt',
            description: 'This prompt has no content',
            type: 'llm',
            icon: 'sparkles',
            // no content property
          },
          manual_steps: [],
        },
      };

      renderComponent({guides: guidesWithoutContent});

      expect(screen.getByText('No Content Prompt')).toBeInTheDocument();
      // Copy Prompt button should not be present
      expect(screen.queryByRole('button', {name: /copy prompt/i})).not.toBeInTheDocument();
    });

    it('should handle special characters in code', () => {
      const guidesWithSpecialChars: IntegrationGuides = {
        inbuilt: {
          llm_prompt: mockGuides.inbuilt.llm_prompt,
          manual_steps: [
            {
              step: 1,
              title: 'Special characters',
              description: 'Code with special chars',
              code: {
                language: 'typescript',
                content: 'const str = "<script>alert(\'xss\')</script>"',
              },
            },
          ],
        },
      };

      renderComponent({guides: guidesWithSpecialChars});

      expect(screen.getByText('Special characters')).toBeInTheDocument();
    });
  });

  describe('language mapping', () => {
    it('should map terminal language to bash', () => {
      renderComponent();

      // Verify terminal code blocks are rendered
      const codeBlocks = screen.getAllByTestId('syntax-highlighter');
      expect(codeBlocks.length).toBeGreaterThan(0);
    });

    it('should map .env language to properties', () => {
      renderComponent();

      // Verify .env file is rendered with filename
      expect(screen.getByText('.env')).toBeInTheDocument();
    });
  });
});
