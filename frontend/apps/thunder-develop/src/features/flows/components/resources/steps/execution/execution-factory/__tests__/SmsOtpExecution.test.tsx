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
import SmsOtpExecution from '../SmsOtpExecution';

// Use vi.hoisted to define mock functions before vi.mock hoisting
const {mockUseColorScheme, mockUseRequiredFields} = vi.hoisted(() => ({
  mockUseColorScheme: vi.fn(() => ({
    mode: 'light',
    systemMode: 'light',
  })),
  mockUseRequiredFields: vi.fn(),
}));

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

// Mock useRequiredFields
vi.mock('@/features/flows/hooks/useRequiredFields', () => ({
  default: mockUseRequiredFields,
}));

// Create mock resource
const createMockResource = (overrides: Partial<Step> = {}): Step =>
  ({
    id: 'sms-otp-execution-1',
    type: 'TASK_EXECUTION',
    position: {x: 0, y: 0},
    size: {width: 200, height: 100},
    display: {
      label: 'SMS OTP',
      image: 'assets/images/icons/sms-otp.svg',
      showOnResourcePanel: true,
    },
    data: {
      action: {
        executor: {
          name: 'SMSOTPAuthExecutor',
        },
      },
      properties: {
        senderId: 'sms-sender-123',
      },
    },
    config: {},
    ...overrides,
  }) as Step;

describe('SmsOtpExecution', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseColorScheme.mockReturnValue({
      mode: 'light',
      systemMode: 'light',
    });
  });

  describe('Rendering', () => {
    it('should render the SMS OTP execution component with label', async () => {
      const resource = createMockResource();
      await render(<SmsOtpExecution resource={resource} />);

      await expect.element(page.getByText('SMS OTP')).toBeInTheDocument();
    });

    it('should render image when display.image is provided', async () => {
      const resource = createMockResource({
        display: {
          label: 'SMS OTP',
          image: 'assets/images/icons/sms-otp.svg',
          showOnResourcePanel: true,
        },
      });
      await render(<SmsOtpExecution resource={resource} />);

      const img = page.getByRole('img');
      expect(img).toHaveAttribute('src', '/static/assets/images/icons/sms-otp.svg');
      expect(img).toHaveAttribute('alt', 'SMS OTP-icon');
      expect(img).toHaveAttribute('height', '20');
    });

    it('should not render image when display.image is not provided', async () => {
      const resource = createMockResource({
        display: {
          label: 'SMS OTP',
          image: '',
          showOnResourcePanel: true,
        },
      });
      await render(<SmsOtpExecution resource={resource} />);

      await expect.element(page.getByRole('img')).not.toBeInTheDocument();
    });

    it('should have the correct CSS class', async () => {
      const resource = createMockResource();
      await render(<SmsOtpExecution resource={resource} />);

      const container = page.getByText('SMS OTP').element().parentElement;
      expect(container).toHaveClass('flow-builder-execution');
    });

    it('should use default alt text when displayLabel is undefined', async () => {
      const resource = createMockResource({
        display: {
          label: undefined as unknown as string,
          image: 'assets/images/icons/sms-otp.svg',
          showOnResourcePanel: true,
        },
      });
      await render(<SmsOtpExecution resource={resource} />);

      const img = page.getByRole('img');
      expect(img).toHaveAttribute('alt', 'sms-otp-icon');
    });
  });

  describe('Display Label', () => {
    it('should use display.label when provided', async () => {
      const resource = createMockResource({
        display: {
          label: 'Custom SMS Label',
          image: 'assets/images/icons/sms-otp.svg',
          showOnResourcePanel: true,
        },
      });
      await render(<SmsOtpExecution resource={resource} />);

      await expect.element(page.getByText('Custom SMS Label')).toBeInTheDocument();
    });

    it('should use translation key for default label when displayLabel is undefined', async () => {
      const resource = createMockResource({
        display: {
          label: undefined as unknown as string,
          image: 'assets/images/icons/sms-otp.svg',
          showOnResourcePanel: true,
        },
      });
      await render(<SmsOtpExecution resource={resource} />);

      await expect.element(page.getByText('Execution')).toBeInTheDocument();
    });

    it('should use translation key when display is undefined', async () => {
      const resource = createMockResource({
        display: undefined,
      });
      await render(<SmsOtpExecution resource={resource} />);

      await expect.element(page.getByText('Execution')).toBeInTheDocument();
    });
  });

  describe('Dark Mode', () => {
    it('should apply dark mode filter when in dark mode', async () => {
      mockUseColorScheme.mockReturnValue({
        mode: 'dark',
        systemMode: 'dark',
      });

      const resource = createMockResource({
        display: {
          label: 'SMS OTP',
          image: 'assets/images/icons/sms-otp.svg',
          showOnResourcePanel: true,
        },
      });
      await render(<SmsOtpExecution resource={resource} />);

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
          label: 'SMS OTP',
          image: 'assets/images/icons/sms-otp.svg',
          showOnResourcePanel: true,
        },
      });
      await render(<SmsOtpExecution resource={resource} />);

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
          label: 'SMS OTP',
          image: 'assets/images/icons/sms-otp.svg',
          showOnResourcePanel: true,
        },
      });
      await render(<SmsOtpExecution resource={resource} />);

      const img = page.getByRole('img');
      expect(img).toHaveStyle({filter: 'brightness(0.9) invert(1)'});
    });

    it('should use systemMode light when mode is system and systemMode is light', async () => {
      mockUseColorScheme.mockReturnValue({
        mode: 'system',
        systemMode: 'light',
      });

      const resource = createMockResource({
        display: {
          label: 'SMS OTP',
          image: 'assets/images/icons/sms-otp.svg',
          showOnResourcePanel: true,
        },
      });
      await render(<SmsOtpExecution resource={resource} />);

      const img = page.getByRole('img');
      expect(img).toHaveStyle({filter: 'none'});
    });
  });

  describe('Required Fields Validation', () => {
    it('should call useRequiredFields with resource and senderId field', async () => {
      const resource = createMockResource();
      await render(<SmsOtpExecution resource={resource} />);

      expect(mockUseRequiredFields).toHaveBeenCalledWith(
        resource,
        expect.anything(),
        expect.arrayContaining([
          expect.objectContaining({
            name: 'data.properties.senderId',
            errorMessage: 'Notification sender is required',
          }),
        ]),
      );
    });

    it('should pass generalMessage as ReactElement to useRequiredFields', async () => {
      const resource = createMockResource({id: 'test-sms-id'});
      await render(<SmsOtpExecution resource={resource} />);

      expect(mockUseRequiredFields).toHaveBeenCalledWith(
        resource,
        expect.objectContaining({
          props: expect.objectContaining({
            i18nKey: 'flows:core.validation.fields.executor.general',
          }) as Record<string, unknown>,
        }),
        expect.any(Array),
      );
    });

    it('should memoize fields array', async () => {
      const resource = createMockResource();
      const {rerender} = await render(<SmsOtpExecution resource={resource} />);

      const firstCallFields = mockUseRequiredFields.mock.calls[0][2] as unknown[];

      await rerender(<SmsOtpExecution resource={resource} />);

      const secondCallFields = mockUseRequiredFields.mock.calls[1][2] as unknown[];

      // Fields should be the same reference due to memoization
      expect(firstCallFields).toBe(secondCallFields);
    });
  });

  describe('Resource Handling', () => {
    it('should handle resource with senderId configured', async () => {
      const resource = createMockResource({
        data: {
          action: {
            executor: {
              name: 'SMSOTPAuthExecutor',
            },
          },
          properties: {
            senderId: 'configured-sender-id',
          },
        },
      });
      await render(<SmsOtpExecution resource={resource} />);

      await expect.element(page.getByText('SMS OTP')).toBeInTheDocument();
    });

    it('should handle resource without senderId', async () => {
      const resource = createMockResource({
        data: {
          action: {
            executor: {
              name: 'SMSOTPAuthExecutor',
            },
          },
          properties: {},
        },
      });
      await render(<SmsOtpExecution resource={resource} />);

      await expect.element(page.getByText('SMS OTP')).toBeInTheDocument();
    });

    it('should handle resource with undefined properties', async () => {
      const resource = createMockResource({
        data: {
          action: {
            executor: {
              name: 'SMSOTPAuthExecutor',
            },
          },
        },
      });
      await render(<SmsOtpExecution resource={resource} />);

      await expect.element(page.getByText('SMS OTP')).toBeInTheDocument();
    });
  });

  describe('Memoization', () => {
    it('should memoize generalMessage based on resource.id', async () => {
      const resource = createMockResource({id: 'sms-1'});
      const {rerender} = await render(<SmsOtpExecution resource={resource} />);

      const firstCallMessage = mockUseRequiredFields.mock.calls[0][1] as unknown;

      await rerender(<SmsOtpExecution resource={resource} />);

      const secondCallMessage = mockUseRequiredFields.mock.calls[1][1] as unknown;

      // Message should be the same reference due to memoization
      expect(firstCallMessage).toBe(secondCallMessage);
    });

    it('should update generalMessage when resource.id changes', async () => {
      const resource1 = createMockResource({id: 'sms-1'});
      const resource2 = createMockResource({id: 'sms-2'});

      const {rerender} = await render(<SmsOtpExecution resource={resource1} />);

      const firstCallMessage = mockUseRequiredFields.mock.calls[0][1] as unknown;

      await rerender(<SmsOtpExecution resource={resource2} />);

      const secondCallMessage = mockUseRequiredFields.mock.calls[1][1] as unknown;

      // Message should be different due to different resource.id
      expect(firstCallMessage).not.toBe(secondCallMessage);
    });
  });
});
