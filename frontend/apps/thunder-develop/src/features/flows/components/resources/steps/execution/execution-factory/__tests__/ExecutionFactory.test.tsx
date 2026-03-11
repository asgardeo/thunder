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
import {render} from '@thunder/test-utils/browser';
import {page} from 'vitest/browser';
import type {Step} from '@/features/flows/models/steps';
import {ExecutionTypes} from '@/features/flows/models/steps';
import ExecutionFactory from '../ExecutionFactory';

// Use vi.hoisted to define mock function before vi.mock hoisting
const mockUseColorScheme = vi.hoisted(() =>
  vi.fn(() => ({
    mode: 'light',
    systemMode: 'light',
  })),
);

// Mock useColorScheme
vi.mock('@wso2/oxygen-ui', async () => {
  const actual = await vi.importActual('@wso2/oxygen-ui');
  return {
    ...actual,
    useColorScheme: () => mockUseColorScheme(),
  };
});

// Mock resolveStaticResourcePath
vi.mock('@/features/flows/utils/resolveStaticResourcePath', () => ({
  default: (path: string) => `/static/${path}`,
}));

// Mock GoogleExecution
vi.mock('../GoogleExecution', () => ({
  default: ({resource}: {resource: Step}) => (
    <div data-testid="google-execution" data-resource-id={resource?.id}>
      GoogleExecution
    </div>
  ),
}));

// Mock GithubExecution
vi.mock('../GithubExecution', () => ({
  default: ({resource}: {resource: Step}) => (
    <div data-testid="github-execution" data-resource-id={resource?.id}>
      GithubExecution
    </div>
  ),
}));

// Mock SmsOtpExecution
vi.mock('../SmsOtpExecution', () => ({
  default: ({resource}: {resource: Step}) => (
    <div data-testid="sms-otp-execution" data-resource-id={resource?.id}>
      SmsOtpExecution
    </div>
  ),
}));

// Create mock resource
const createMockResource = (overrides: Partial<Step> = {}): Step =>
  ({
    id: 'execution-1',
    type: 'TASK_EXECUTION',
    position: {x: 0, y: 0},
    size: {width: 200, height: 100},
    display: {
      label: 'Test Executor',
      image: 'assets/images/icons/test.svg',
      showOnResourcePanel: true,
    },
    data: {
      action: {
        executor: {
          name: 'TestExecutor',
        },
      },
    },
    config: {},
    ...overrides,
  }) as Step;

describe('ExecutionFactory', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseColorScheme.mockReturnValue({
      mode: 'light',
      systemMode: 'light',
    });
  });

  describe('Google Federation', () => {
    it('should render GoogleExecution for GoogleOIDCAuthExecutor', async () => {
      const resource = createMockResource({
        data: {
          action: {
            executor: {
              name: ExecutionTypes.GoogleFederation,
            },
          },
        },
      });
      await render(<ExecutionFactory resource={resource} />);

      await expect.element(page.getByTestId('google-execution')).toBeInTheDocument();
      await expect.element(page.getByTestId('github-execution')).not.toBeInTheDocument();
      await expect.element(page.getByTestId('sms-otp-execution')).not.toBeInTheDocument();
    });

    it('should pass resource to GoogleExecution', async () => {
      const resource = createMockResource({
        id: 'google-resource-1',
        data: {
          action: {
            executor: {
              name: ExecutionTypes.GoogleFederation,
            },
          },
        },
      });
      await render(<ExecutionFactory resource={resource} />);

      const googleExecution = page.getByTestId('google-execution');
      expect(googleExecution).toHaveAttribute('data-resource-id', 'google-resource-1');
    });
  });

  describe('GitHub Federation', () => {
    it('should render GithubExecution for GithubOAuthExecutor', async () => {
      const resource = createMockResource({
        data: {
          action: {
            executor: {
              name: ExecutionTypes.GithubFederation,
            },
          },
        },
      });
      await render(<ExecutionFactory resource={resource} />);

      await expect.element(page.getByTestId('github-execution')).toBeInTheDocument();
      await expect.element(page.getByTestId('google-execution')).not.toBeInTheDocument();
      await expect.element(page.getByTestId('sms-otp-execution')).not.toBeInTheDocument();
    });

    it('should pass resource to GithubExecution', async () => {
      const resource = createMockResource({
        id: 'github-resource-1',
        data: {
          action: {
            executor: {
              name: ExecutionTypes.GithubFederation,
            },
          },
        },
      });
      await render(<ExecutionFactory resource={resource} />);

      const githubExecution = page.getByTestId('github-execution');
      expect(githubExecution).toHaveAttribute('data-resource-id', 'github-resource-1');
    });
  });

  describe('SMS OTP Auth', () => {
    it('should render SmsOtpExecution for SMSOTPAuthExecutor', async () => {
      const resource = createMockResource({
        data: {
          action: {
            executor: {
              name: ExecutionTypes.SMSOTPAuth,
            },
          },
        },
      });
      await render(<ExecutionFactory resource={resource} />);

      await expect.element(page.getByTestId('sms-otp-execution')).toBeInTheDocument();
      await expect.element(page.getByTestId('google-execution')).not.toBeInTheDocument();
      await expect.element(page.getByTestId('github-execution')).not.toBeInTheDocument();
    });

    it('should pass resource to SmsOtpExecution', async () => {
      const resource = createMockResource({
        id: 'sms-otp-resource-1',
        data: {
          action: {
            executor: {
              name: ExecutionTypes.SMSOTPAuth,
            },
          },
        },
      });
      await render(<ExecutionFactory resource={resource} />);

      const smsOtpExecution = page.getByTestId('sms-otp-execution');
      expect(smsOtpExecution).toHaveAttribute('data-resource-id', 'sms-otp-resource-1');
    });
  });

  describe('Generic Executor with Display Image', () => {
    it('should render image and label for executors with display.image', async () => {
      const resource = createMockResource({
        display: {
          label: 'Custom Executor',
          image: 'assets/images/icons/custom.svg',
          showOnResourcePanel: true,
        },
        data: {
          action: {
            executor: {
              name: 'CustomExecutor',
            },
          },
        },
      });
      await render(<ExecutionFactory resource={resource} />);

      const img = page.getByRole('img');
      expect(img).toHaveAttribute('src', '/static/assets/images/icons/custom.svg');
      expect(img).toHaveAttribute('alt', 'Custom Executor-icon');
      await expect.element(page.getByText('Custom Executor')).toBeInTheDocument();
    });

    it('should use default alt text when displayLabel is undefined', async () => {
      const resource = createMockResource({
        display: {
          label: undefined as unknown as string,
          image: 'assets/images/icons/custom.svg',
          showOnResourcePanel: true,
        },
        data: {
          action: {
            executor: {
              name: 'CustomExecutor',
            },
          },
        },
      });
      await render(<ExecutionFactory resource={resource} />);

      const img = page.getByRole('img');
      expect(img).toHaveAttribute('alt', 'executor-icon');
    });

    it('should use translation key for default label when displayLabel is undefined', async () => {
      const resource = createMockResource({
        display: {
          label: undefined as unknown as string,
          image: 'assets/images/icons/custom.svg',
          showOnResourcePanel: true,
        },
        data: {
          action: {
            executor: {
              name: 'CustomExecutor',
            },
          },
        },
      });
      await render(<ExecutionFactory resource={resource} />);

      await expect.element(page.getByText('Execution')).toBeInTheDocument();
    });

    it('should apply dark mode filter when in dark mode', async () => {
      mockUseColorScheme.mockReturnValue({
        mode: 'dark',
        systemMode: 'dark',
      });

      const resource = createMockResource({
        display: {
          label: 'Dark Mode Executor',
          image: 'assets/images/icons/custom.svg',
          showOnResourcePanel: true,
        },
        data: {
          action: {
            executor: {
              name: 'DarkModeExecutor',
            },
          },
        },
      });
      await render(<ExecutionFactory resource={resource} />);

      const img = page.getByRole('img');
      expect(img).toHaveStyle({filter: 'brightness(0.9) invert(1)'});
    });

    it('should not apply filter in light mode', async () => {
      mockUseColorScheme.mockReturnValue({
        mode: 'light',
        systemMode: 'light',
      });

      const resource = createMockResource({
        display: {
          label: 'Light Mode Executor',
          image: 'assets/images/icons/custom.svg',
          showOnResourcePanel: true,
        },
        data: {
          action: {
            executor: {
              name: 'LightModeExecutor',
            },
          },
        },
      });
      await render(<ExecutionFactory resource={resource} />);

      const img = page.getByRole('img');
      expect(img).toHaveStyle({filter: 'none'});
    });

    it('should use systemMode when mode is set to system', async () => {
      mockUseColorScheme.mockReturnValue({
        mode: 'system',
        systemMode: 'dark',
      });

      const resource = createMockResource({
        display: {
          label: 'System Mode Executor',
          image: 'assets/images/icons/custom.svg',
          showOnResourcePanel: true,
        },
        data: {
          action: {
            executor: {
              name: 'SystemModeExecutor',
            },
          },
        },
      });
      await render(<ExecutionFactory resource={resource} />);

      const img = page.getByRole('img');
      expect(img).toHaveStyle({filter: 'brightness(0.9) invert(1)'});
    });
  });

  describe('Fallback Executor without Display Image', () => {
    it('should render only label when display.image is not provided', async () => {
      const resource = createMockResource({
        display: {
          label: 'No Image Executor',
          image: '',
          showOnResourcePanel: true,
        },
        data: {
          action: {
            executor: {
              name: 'NoImageExecutor',
            },
          },
        },
      });
      await render(<ExecutionFactory resource={resource} />);

      await expect.element(page.getByText('No Image Executor')).toBeInTheDocument();
      await expect.element(page.getByRole('img')).not.toBeInTheDocument();
    });

    it('should use translation key for default label when displayLabel is not provided and no image', async () => {
      const resource = createMockResource({
        display: undefined,
        data: {
          action: {
            executor: {
              name: 'FallbackExecutor',
            },
          },
        },
      });
      await render(<ExecutionFactory resource={resource} />);

      await expect.element(page.getByText('Execution')).toBeInTheDocument();
    });

    it('should render fallback when display is completely undefined', async () => {
      const resource = createMockResource({
        display: undefined,
        data: {
          action: {
            executor: {
              name: 'UndefinedDisplayExecutor',
            },
          },
        },
      });
      await render(<ExecutionFactory resource={resource} />);

      await expect.element(page.getByText('Execution')).toBeInTheDocument();
    });
  });

  describe('Edge Cases', () => {
    it('should handle undefined data', async () => {
      const resource = createMockResource({
        display: undefined,
        data: undefined,
      });
      await render(<ExecutionFactory resource={resource} />);

      // Should render fallback
      await expect.element(page.getByText('Execution')).toBeInTheDocument();
    });

    it('should handle undefined action', async () => {
      const resource = createMockResource({
        display: undefined,
        data: {},
      });
      await render(<ExecutionFactory resource={resource} />);

      // Should render fallback
      await expect.element(page.getByText('Execution')).toBeInTheDocument();
    });

    it('should handle undefined executor', async () => {
      const resource = createMockResource({
        display: undefined,
        data: {
          action: {},
        },
      });
      await render(<ExecutionFactory resource={resource} />);

      // Should render fallback
      await expect.element(page.getByText('Execution')).toBeInTheDocument();
    });

    it('should handle undefined executor name', async () => {
      const resource = createMockResource({
        display: undefined,
        data: {
          action: {
            executor: {},
          },
        },
      });
      await render(<ExecutionFactory resource={resource} />);

      // Should render fallback
      await expect.element(page.getByText('Execution')).toBeInTheDocument();
    });
  });
});
