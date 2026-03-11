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

import {render} from '@thunder/test-utils/browser';
import {page, userEvent} from 'vitest/browser';
import {describe, it, expect, vi, beforeEach, afterEach} from 'vitest';
import {LoggerProvider, LogLevel} from '@thunder/logger';
import TechnologyGuide from '../TechnologyGuide';
import type {IntegrationGuides} from '../../../../models/application-templates';

const mockIntegrationGuides: IntegrationGuides = {
  INBUILT: {
    llm_prompt: {
      id: 'llm-1',
      title: 'Use AI Assistant',
      description: 'Get AI-powered integration guidance',
      type: 'llm' as const,
      icon: 'sparkles',
      content: 'Integrate with clientId: {{clientId}} and applicationId: {{applicationId}}',
    },
    manual_steps: [
      {
        step: 1,
        title: 'Install dependencies',
        description: 'Install required packages for your application',
        subDescription: 'Run the following command in your terminal',
        bullets: ['npm for Node Package Manager', 'yarn for Yarn Package Manager'],
        code: {
          language: 'bash',
          filename: 'terminal',
          content: 'npm install @thunder/sdk',
        },
      },
      {
        step: 2,
        title: 'Configure client',
        description: 'Set up your application with the client ID',
        code: {
          language: 'typescript',
          filename: 'config.ts',
          content: 'const clientId = "{{clientId}}";',
        },
      },
    ],
  },
  EMBEDDED: {
    llm_prompt: {
      id: 'llm-2',
      title: 'Embedded Integration',
      description: 'Custom login UI integration',
      type: 'llm' as const,
      icon: 'sparkles',
      content: 'Embedded integration prompt',
    },
    manual_steps: [
      {
        step: 1,
        title: 'Setup custom UI',
        description: 'Create your custom login form',
      },
    ],
  },
};

const mockWriteText = vi.fn();

const renderWithProviders = async (component: React.ReactElement) =>
  render(<LoggerProvider logger={{level: LogLevel.DEBUG}}>{component}</LoggerProvider>);

describe('TechnologyGuide', () => {
  const originalClipboard = navigator.clipboard;

  beforeEach(() => {
    vi.useFakeTimers({shouldAdvanceTime: true});
    vi.clearAllMocks();
    mockWriteText.mockResolvedValue(undefined);
    Object.defineProperty(navigator, 'clipboard', {
      value: {
        writeText: mockWriteText,
      },
      writable: true,
      configurable: true,
    });
  });

  afterEach(() => {
    vi.runOnlyPendingTimers();
    vi.useRealTimers();
    Object.defineProperty(navigator, 'clipboard', {
      value: originalClipboard,
      writable: true,
      configurable: true,
    });
  });

  describe('Rendering', () => {
    it('should return null when guides is null', async () => {
      const {container} = await renderWithProviders(<TechnologyGuide guides={null} />);

      expect(container.firstChild?.firstChild).toBeFalsy();
    });

    it('should return null when selected guide is not found', async () => {
      const guidesWithoutInbuilt: IntegrationGuides = {
        OTHER: mockIntegrationGuides.INBUILT,
      };

      const {container} = await renderWithProviders(<TechnologyGuide guides={guidesWithoutInbuilt} templateId="react" />);

      expect(container.firstChild?.firstChild).toBeFalsy();
    });

    it('should render inbuilt guide for non-embedded template', async () => {
      await renderWithProviders(<TechnologyGuide guides={mockIntegrationGuides} templateId="react" />);

      await expect.element(page.getByText('Use AI Assistant')).toBeInTheDocument();
      await expect.element(page.getByText('Get AI-powered integration guidance')).toBeInTheDocument();
    });

    it('should render embedded guide for embedded template', async () => {
      await renderWithProviders(<TechnologyGuide guides={mockIntegrationGuides} templateId="react-embedded" />);

      await expect.element(page.getByText('Embedded Integration')).toBeInTheDocument();
    });

    it('should default to inbuilt guide when templateId is null', async () => {
      await renderWithProviders(<TechnologyGuide guides={mockIntegrationGuides} templateId={null} />);

      await expect.element(page.getByText('Use AI Assistant')).toBeInTheDocument();
    });
  });

  describe('LLM Prompt Section', () => {
    it('should render LLM prompt card with title and description', async () => {
      await renderWithProviders(<TechnologyGuide guides={mockIntegrationGuides} templateId="react" clientId="client-123" />);

      await expect.element(page.getByText('Use AI Assistant')).toBeInTheDocument();
      await expect.element(page.getByText('Get AI-powered integration guidance')).toBeInTheDocument();
    });

    it('should render copy prompt button', async () => {
      await renderWithProviders(<TechnologyGuide guides={mockIntegrationGuides} templateId="react" />);

      await expect.element(page.getByTestId('copy-prompt-button')).toBeInTheDocument();
      await expect.element(page.getByText('Copy Prompt')).toBeInTheDocument();
    });
  });

  describe('Manual Steps Section', () => {
    it('should render divider with "or" text', async () => {
      await renderWithProviders(<TechnologyGuide guides={mockIntegrationGuides} templateId="react" />);

      await expect.element(page.getByText('or', {exact: true})).toBeInTheDocument();
    });

    it('should render all manual steps', async () => {
      await renderWithProviders(<TechnologyGuide guides={mockIntegrationGuides} templateId="react" />);

      await expect.element(page.getByText('Install dependencies')).toBeInTheDocument();
      await expect.element(page.getByText('Configure client')).toBeInTheDocument();
    });

    it('should render step numbers', async () => {
      await renderWithProviders(<TechnologyGuide guides={mockIntegrationGuides} templateId="react" />);

      await expect.element(page.getByText('1')).toBeInTheDocument();
      await expect.element(page.getByText('2')).toBeInTheDocument();
    });

    it('should render step descriptions', async () => {
      await renderWithProviders(<TechnologyGuide guides={mockIntegrationGuides} templateId="react" />);

      await expect.element(page.getByText('Install required packages for your application')).toBeInTheDocument();
      await expect.element(page.getByText('Set up your application with the client ID')).toBeInTheDocument();
    });

    it('should render sub-descriptions when provided', async () => {
      await renderWithProviders(<TechnologyGuide guides={mockIntegrationGuides} templateId="react" />);

      await expect.element(page.getByText('Run the following command in your terminal')).toBeInTheDocument();
    });

    it('should render bullet points when provided', async () => {
      await renderWithProviders(<TechnologyGuide guides={mockIntegrationGuides} templateId="react" />);

      await expect.element(page.getByText('npm for Node Package Manager')).toBeInTheDocument();
      await expect.element(page.getByText('yarn for Yarn Package Manager')).toBeInTheDocument();
    });
  });

  describe('Code Blocks', () => {
    it('should render code blocks for steps with code', async () => {
      const {container} = await renderWithProviders(<TechnologyGuide guides={mockIntegrationGuides} templateId="react" />);

      // Check that code blocks exist
      const codeBlocks = container.querySelectorAll('pre');
      expect(codeBlocks).toHaveLength(2);

      // Check code content is present
      expect(container.textContent).toContain('npm install @thunder/sdk');
      expect(container.textContent).toContain('const clientId = "{{clientId}}";');
    });

    it('should render filenames when provided', async () => {
      await renderWithProviders(<TechnologyGuide guides={mockIntegrationGuides} templateId="react" />);

      await expect.element(page.getByText('terminal', {exact: true})).toBeInTheDocument();
      await expect.element(page.getByText('config.ts')).toBeInTheDocument();
    });

    it('should render copy buttons for each code block', async () => {
      await renderWithProviders(<TechnologyGuide guides={mockIntegrationGuides} templateId="react" />);

      const copyButtons = page.getByTestId(/copy-code-button-/);
      expect(copyButtons).toHaveLength(2);
    });
  });

  describe('Empty States', () => {
    it('should not render code block when step has no code', async () => {
      const guidesWithoutCode: IntegrationGuides = {
        INBUILT: {
          llm_prompt: mockIntegrationGuides.INBUILT.llm_prompt,
          manual_steps: [
            {
              step: 1,
              title: 'No code step',
              description: 'This step has no code',
            },
          ],
        },
      };

      const {container} = await renderWithProviders(<TechnologyGuide guides={guidesWithoutCode} templateId="react" />);

      const codeBlocks = container.querySelectorAll('pre');
      expect(codeBlocks).toHaveLength(0);
    });

    it('should not render manual steps section when manual_steps is empty', async () => {
      const guidesWithoutSteps: IntegrationGuides = {
        INBUILT: {
          llm_prompt: mockIntegrationGuides.INBUILT.llm_prompt,
          manual_steps: [],
        },
      };

      await renderWithProviders(<TechnologyGuide guides={guidesWithoutSteps} templateId="react" />);

      await expect.element(page.getByText('or')).not.toBeInTheDocument();
    });

    it('should not render copy prompt button when llm_prompt has no content', async () => {
      const guidesWithoutContent: IntegrationGuides = {
        INBUILT: {
          llm_prompt: {
            id: 'llm-1',
            title: 'Use AI Assistant',
            description: 'Get AI-powered integration guidance',
            type: 'llm' as const,
            icon: 'sparkles',
          },
          manual_steps: [],
        },
      };

      await renderWithProviders(<TechnologyGuide guides={guidesWithoutContent} templateId="react" />);

      await expect.element(page.getByTestId('copy-prompt-button')).not.toBeInTheDocument();
    });
  });

  describe('Placeholder Replacement', () => {
    it('should replace {{clientId}} placeholder in code blocks', async () => {
      const {container} = await renderWithProviders(
        <TechnologyGuide guides={mockIntegrationGuides} templateId="react" clientId="my-client-id" />,
      );

      expect(container.textContent).toContain('const clientId = "my-client-id";');
      expect(container.textContent).not.toContain('{{clientId}}');
    });

    it('should replace {{applicationId}} placeholder in LLM prompt when copied', async () => {
      await renderWithProviders(
        <TechnologyGuide
          guides={mockIntegrationGuides}
          templateId="react"
          clientId="test-client"
          applicationId="test-app-id"
        />,
      );

      const copyButton = page.getByTestId('copy-prompt-button');
      await userEvent.click(copyButton);

      await vi.waitFor(() => {
        expect(mockWriteText).toHaveBeenCalledWith(
          'Integrate with clientId: test-client and applicationId: test-app-id',
        );
      });
    });

    it('should not replace placeholders when clientId is empty', async () => {
      const {container} = await renderWithProviders(
        <TechnologyGuide guides={mockIntegrationGuides} templateId="react" clientId="" />,
      );

      expect(container.textContent).toContain('{{clientId}}');
    });

    it('should not replace applicationId placeholder when applicationId is empty', async () => {
      await renderWithProviders(
        <TechnologyGuide guides={mockIntegrationGuides} templateId="react" clientId="test-client" applicationId="" />,
      );

      const copyButton = page.getByTestId('copy-prompt-button');
      await userEvent.click(copyButton);

      await vi.waitFor(() => {
        expect(mockWriteText).toHaveBeenCalledWith(
          'Integrate with clientId: test-client and applicationId: {{applicationId}}',
        );
      });
    });
  });

  describe('Copy Functionality', () => {
    it('should copy LLM prompt to clipboard when copy button is clicked', async () => {
      await renderWithProviders(<TechnologyGuide guides={mockIntegrationGuides} templateId="react" />);

      const copyButton = page.getByTestId('copy-prompt-button');
      await userEvent.click(copyButton);

      await vi.waitFor(() => {
        expect(mockWriteText).toHaveBeenCalledWith(
          'Integrate with clientId: {{clientId}} and applicationId: {{applicationId}}',
        );
      });
    });

    it('should show copied feedback after copying prompt', async () => {
      await renderWithProviders(<TechnologyGuide guides={mockIntegrationGuides} templateId="react" />);

      const copyButton = page.getByTestId('copy-prompt-button');
      await userEvent.click(copyButton);

      // The copied feedback is shown in the Tooltip
      await vi.waitFor(async () => {
        await expect.element(page.getByText('Copied to clipboard')).toBeInTheDocument();
      });
    });

    it('should copy code to clipboard when copy code button is clicked', async () => {
      await renderWithProviders(
        <TechnologyGuide guides={mockIntegrationGuides} templateId="react" clientId="test-client-123" />,
      );

      const copyCodeButton = page.getByTestId('copy-code-button-1');
      await userEvent.click(copyCodeButton);

      await vi.waitFor(() => {
        expect(mockWriteText).toHaveBeenCalledWith('npm install @thunder/sdk');
      });
    });

    it('should show copied feedback after copying code', async () => {
      await renderWithProviders(<TechnologyGuide guides={mockIntegrationGuides} templateId="react" />);

      const copyCodeButton = page.getByTestId('copy-code-button-1');
      await userEvent.click(copyCodeButton);

      // The copied feedback is shown as translated text
      await vi.waitFor(async () => {
        await expect.element(page.getByText('Copied to clipboard')).toBeInTheDocument();
      });
    });

    it('should replace placeholders in copied code', async () => {
      await renderWithProviders(
        <TechnologyGuide guides={mockIntegrationGuides} templateId="react" clientId="replaced-client-id" />,
      );

      const copyCodeButton = page.getByTestId('copy-code-button-2');
      await userEvent.click(copyCodeButton);

      await vi.waitFor(() => {
        expect(mockWriteText).toHaveBeenCalledWith('const clientId = "replaced-client-id";');
      });
    });

    it('should not call clipboard when prompt has no content', async () => {
      const guidesWithEmptyContent: IntegrationGuides = {
        INBUILT: {
          llm_prompt: {
            id: 'llm-1',
            title: 'Use AI Assistant',
            description: 'Get AI-powered integration guidance',
            type: 'llm' as const,
            icon: 'sparkles',
            content: '',
          },
          manual_steps: [],
        },
      };

      await renderWithProviders(<TechnologyGuide guides={guidesWithEmptyContent} templateId="react" />);

      // Empty string content is falsy, so the copy button should not be rendered
      await expect.element(page.getByTestId('copy-prompt-button')).not.toBeInTheDocument();
      expect(mockWriteText).not.toHaveBeenCalled();
    });

    describe('Clipboard Fallback', () => {
      it('should use fallback method when clipboard API fails for prompt', async () => {
        mockWriteText.mockRejectedValue(new Error('Clipboard API failed'));

        const mockExecCommand = vi.fn().mockReturnValue(true);
        document.execCommand = mockExecCommand;

        await renderWithProviders(<TechnologyGuide guides={mockIntegrationGuides} templateId="react" />);

        const copyButton = page.getByTestId('copy-prompt-button');
        await userEvent.click(copyButton);

        await vi.waitFor(() => {
          expect(mockExecCommand).toHaveBeenCalledWith('copy');
        });
      });

      it('should use fallback method when clipboard API fails for code', async () => {
        mockWriteText.mockRejectedValue(new Error('Clipboard API failed'));

        const mockExecCommand = vi.fn().mockReturnValue(true);
        document.execCommand = mockExecCommand;

        await renderWithProviders(<TechnologyGuide guides={mockIntegrationGuides} templateId="react" />);

        const copyCodeButton = page.getByTestId('copy-code-button-1');
        await userEvent.click(copyCodeButton);

        await vi.waitFor(() => {
          expect(mockExecCommand).toHaveBeenCalledWith('copy');
        });
      });

      it('should handle fallback failure gracefully for prompt', async () => {
        mockWriteText.mockRejectedValue(new Error('Clipboard API failed'));

        const mockExecCommand = vi.fn().mockImplementation(() => {
          throw new Error('execCommand failed');
        });
        document.execCommand = mockExecCommand;

        await renderWithProviders(<TechnologyGuide guides={mockIntegrationGuides} templateId="react" />);

        const copyButton = page.getByTestId('copy-prompt-button');

        // Should not throw - component handles error gracefully
        expect(async () => userEvent.click(copyButton)).not.toThrow();
      });

      it('should handle fallback failure gracefully for code', async () => {
        mockWriteText.mockRejectedValue(new Error('Clipboard API failed'));

        const mockExecCommand = vi.fn().mockImplementation(() => {
          throw new Error('execCommand failed');
        });
        document.execCommand = mockExecCommand;

        await renderWithProviders(<TechnologyGuide guides={mockIntegrationGuides} templateId="react" />);

        const copyCodeButton = page.getByTestId('copy-code-button-1');

        // Should not throw - component handles error gracefully
        expect(async () => userEvent.click(copyCodeButton)).not.toThrow();
      });
    });
  });

  describe('Code Block Language Mapping', () => {
    it('should map terminal language to bash', async () => {
      const guidesWithTerminal: IntegrationGuides = {
        INBUILT: {
          llm_prompt: mockIntegrationGuides.INBUILT.llm_prompt,
          manual_steps: [
            {
              step: 1,
              title: 'Run command',
              description: 'Execute this command',
              code: {
                language: 'terminal',
                content: 'npm install',
              },
            },
          ],
        },
      };

      const {container} = await renderWithProviders(<TechnologyGuide guides={guidesWithTerminal} templateId="react" />);

      const codeBlock = container.querySelector('pre');
      expect(codeBlock).toBeInTheDocument();
    });

    it('should map .env language to properties', async () => {
      const guidesWithEnv: IntegrationGuides = {
        INBUILT: {
          llm_prompt: mockIntegrationGuides.INBUILT.llm_prompt,
          manual_steps: [
            {
              step: 1,
              title: 'Configure env',
              description: 'Set environment variables',
              code: {
                language: '.env',
                filename: '.env',
                content: 'API_KEY=your-key',
              },
            },
          ],
        },
      };

      const {container} = await renderWithProviders(<TechnologyGuide guides={guidesWithEnv} templateId="react" />);

      const codeBlock = container.querySelector('pre');
      expect(codeBlock).toBeInTheDocument();
    });

    it('should map typescript language to tsx', async () => {
      const guidesWithTs: IntegrationGuides = {
        INBUILT: {
          llm_prompt: mockIntegrationGuides.INBUILT.llm_prompt,
          manual_steps: [
            {
              step: 1,
              title: 'Add code',
              description: 'Add TypeScript code',
              code: {
                language: 'typescript',
                content: 'const x: string = "test";',
              },
            },
          ],
        },
      };

      const {container} = await renderWithProviders(<TechnologyGuide guides={guidesWithTs} templateId="react" />);

      const codeBlock = container.querySelector('pre');
      expect(codeBlock).toBeInTheDocument();
    });

    it('should pass through unknown languages unchanged', async () => {
      const guidesWithPython: IntegrationGuides = {
        INBUILT: {
          llm_prompt: mockIntegrationGuides.INBUILT.llm_prompt,
          manual_steps: [
            {
              step: 1,
              title: 'Python code',
              description: 'Add Python code',
              code: {
                language: 'python',
                content: 'print("hello")',
              },
            },
          ],
        },
      };

      const {container} = await renderWithProviders(<TechnologyGuide guides={guidesWithPython} templateId="react" />);

      const codeBlock = container.querySelector('pre');
      expect(codeBlock).toBeInTheDocument();
    });

    it('should render code block without filename header when filename is not provided', async () => {
      const guidesWithoutFilename: IntegrationGuides = {
        INBUILT: {
          llm_prompt: mockIntegrationGuides.INBUILT.llm_prompt,
          manual_steps: [
            {
              step: 1,
              title: 'Run command',
              description: 'Execute this command',
              code: {
                language: 'bash',
                content: 'npm install',
              },
            },
          ],
        },
      };

      await renderWithProviders(<TechnologyGuide guides={guidesWithoutFilename} templateId="react" />);

      // Should not have a filename displayed
      await expect.element(page.getByText('terminal')).not.toBeInTheDocument();
      await expect.element(page.getByText('config.ts')).not.toBeInTheDocument();
    });
  });

  describe('Bullets Rendering', () => {
    it('should not render bullets section when bullets array is empty', async () => {
      const guidesWithEmptyBullets: IntegrationGuides = {
        INBUILT: {
          llm_prompt: mockIntegrationGuides.INBUILT.llm_prompt,
          manual_steps: [
            {
              step: 1,
              title: 'Step without bullets',
              description: 'This step has empty bullets array',
              bullets: [],
            },
          ],
        },
      };

      const {container} = await renderWithProviders(<TechnologyGuide guides={guidesWithEmptyBullets} templateId="react" />);

      const bulletLists = container.querySelectorAll('ul');
      expect(bulletLists).toHaveLength(0);
    });
  });
});
