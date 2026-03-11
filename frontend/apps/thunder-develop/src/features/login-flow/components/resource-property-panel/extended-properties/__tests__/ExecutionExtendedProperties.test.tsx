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
import {page, userEvent} from 'vitest/browser';
import {ExecutionTypes} from '@/features/flows/models/steps';
import type {Resource} from '@/features/flows/models/resources';
import {IdentityProviderTypes} from '@/features/integrations/models/identity-provider';
import ExecutionExtendedProperties from '../ExecutionExtendedProperties';

// Mock useValidationStatus
const mockSelectedNotification = {
  hasResourceFieldNotification: vi.fn(() => false),
  getResourceFieldNotification: vi.fn(() => ''),
};

vi.mock('@/features/flows/hooks/useValidationStatus', () => ({
  default: () => ({
    selectedNotification: mockSelectedNotification,
  }),
}));

// Mock useIdentityProviders
const mockIdentityProviders = vi.fn<() => {data: unknown[]; isLoading: boolean}>();
vi.mock('@/features/integrations/api/useIdentityProviders', () => ({
  // eslint-disable-next-line @typescript-eslint/no-unsafe-return
  default: () => mockIdentityProviders(),
}));

// Mock useNotificationSenders
const mockNotificationSenders = vi.fn<() => {data: unknown[]; isLoading: boolean}>();
vi.mock('@/features/notification-senders/api/useNotificationSenders', () => ({
  // eslint-disable-next-line @typescript-eslint/no-unsafe-return
  default: () => mockNotificationSenders(),
}));

describe('ExecutionExtendedProperties', () => {
  const mockOnChange = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
    mockIdentityProviders.mockReturnValue({
      data: [],
      isLoading: false,
    });
    mockNotificationSenders.mockReturnValue({
      data: [],
      isLoading: false,
    });
  });

  describe('Google Federation Executor', () => {
    const googleResource = {
      id: 'google-executor-1',
      data: {
        action: {
          executor: {
            name: ExecutionTypes.GoogleFederation,
          },
        },
        properties: {
          idpId: '',
        },
      },
    } as unknown as Resource;

    it('should render connection selector for Google executor', async () => {
      mockIdentityProviders.mockReturnValue({
        data: [
          {id: 'google-idp-1', name: 'Google IDP', type: IdentityProviderTypes.GOOGLE},
        ],
        isLoading: false,
      });

      await render(<ExecutionExtendedProperties resource={googleResource} onChange={mockOnChange} />);

      await expect.element(page.getByText('Connection')).toBeInTheDocument();
      await expect.element(page.getByText('Select a connection from the following list to link it with the login flow.')).toBeInTheDocument();
    });

    it('should show available Google connections in dropdown', async () => {
            mockIdentityProviders.mockReturnValue({
        data: [
          {id: 'google-idp-1', name: 'My Google IDP', type: IdentityProviderTypes.GOOGLE},
          {id: 'google-idp-2', name: 'Another Google IDP', type: IdentityProviderTypes.GOOGLE},
        ],
        isLoading: false,
      });

      await render(<ExecutionExtendedProperties resource={googleResource} onChange={mockOnChange} />);

      const select = page.getByRole('combobox');
      await userEvent.click(select);

      await expect.element(page.getByText('My Google IDP')).toBeInTheDocument();
      await expect.element(page.getByText('Another Google IDP')).toBeInTheDocument();
    });

    it('should call onChange when connection is selected', async () => {
            mockIdentityProviders.mockReturnValue({
        data: [
          {id: 'google-idp-1', name: 'My Google IDP', type: IdentityProviderTypes.GOOGLE},
        ],
        isLoading: false,
      });

      await render(<ExecutionExtendedProperties resource={googleResource} onChange={mockOnChange} />);

      const select = page.getByRole('combobox');
      await userEvent.click(select);
      await userEvent.click(page.getByText('My Google IDP'));

      expect(mockOnChange).toHaveBeenCalledWith('data.properties.idpId', 'google-idp-1', googleResource);
    });

    it('should show error when connection is placeholder', async () => {
      mockIdentityProviders.mockReturnValue({
        data: [
          {id: 'google-idp-1', name: 'My Google IDP', type: IdentityProviderTypes.GOOGLE},
        ],
        isLoading: false,
      });

      const resourceWithPlaceholder = {
        ...googleResource,
        data: {
          ...(googleResource as unknown as {data: object}).data,
          properties: {idpId: '{{IDP_ID}}'},
        },
      } as unknown as Resource;

      await render(<ExecutionExtendedProperties resource={resourceWithPlaceholder} onChange={mockOnChange} />);

      await expect.element(page.getByText('Connection is required and must be selected.')).toBeInTheDocument();
    });

    it('should show validation error from notification', async () => {
      mockSelectedNotification.hasResourceFieldNotification.mockReturnValue(true);
      mockSelectedNotification.getResourceFieldNotification.mockReturnValue('Custom validation error');

      mockIdentityProviders.mockReturnValue({
        data: [
          {id: 'google-idp-1', name: 'My Google IDP', type: IdentityProviderTypes.GOOGLE},
        ],
        isLoading: false,
      });

      await render(<ExecutionExtendedProperties resource={googleResource} onChange={mockOnChange} />);

      await expect.element(page.getByText('Custom validation error')).toBeInTheDocument();
    });

    it('should show warning when no connections are available', async () => {
      mockIdentityProviders.mockReturnValue({
        data: [],
        isLoading: false,
      });

      await render(<ExecutionExtendedProperties resource={googleResource} onChange={mockOnChange} />);

      await expect.element(page.getByText('No connections available. Please create a connection to link with the login flow.')).toBeInTheDocument();
    });

    it('should disable dropdown while loading', async () => {
      mockIdentityProviders.mockReturnValue({
        data: [],
        isLoading: true,
      });

      await render(<ExecutionExtendedProperties resource={googleResource} onChange={mockOnChange} />);

      const select = page.getByRole('combobox');
      expect(select).toHaveAttribute('aria-disabled', 'true');
    });

    it('should show loading text in dropdown while loading', async () => {
            mockIdentityProviders.mockReturnValue({
        data: [],
        isLoading: true,
      });

      await render(<ExecutionExtendedProperties resource={googleResource} onChange={mockOnChange} />);

      const select = page.getByRole('combobox');
      await userEvent.click(select);

      await expect.element(page.getByText('Loading...')).toBeInTheDocument();
    });

    it('should show selected connection value', async () => {
      mockIdentityProviders.mockReturnValue({
        data: [
          {id: 'google-idp-1', name: 'My Google IDP', type: IdentityProviderTypes.GOOGLE},
        ],
        isLoading: false,
      });

      const resourceWithSelection = {
        ...googleResource,
        data: {
          ...(googleResource as unknown as {data: object}).data,
          properties: {idpId: 'google-idp-1'},
        },
      } as unknown as Resource;

      await render(<ExecutionExtendedProperties resource={resourceWithSelection} onChange={mockOnChange} />);

      await expect.element(page.getByRole('combobox')).toHaveTextContent('My Google IDP');
    });
  });

  describe('GitHub Federation Executor', () => {
    const githubResource = {
      id: 'github-executor-1',
      data: {
        action: {
          executor: {
            name: ExecutionTypes.GithubFederation,
          },
        },
        properties: {
          idpId: '',
        },
      },
    } as unknown as Resource;

    it('should render connection selector for GitHub executor', async () => {
      mockIdentityProviders.mockReturnValue({
        data: [
          {id: 'github-idp-1', name: 'GitHub IDP', type: IdentityProviderTypes.GITHUB},
        ],
        isLoading: false,
      });

      await render(<ExecutionExtendedProperties resource={githubResource} onChange={mockOnChange} />);

      await expect.element(page.getByText('Connection')).toBeInTheDocument();
    });

    it('should filter to only show GitHub connections', async () => {
            mockIdentityProviders.mockReturnValue({
        data: [
          {id: 'google-idp-1', name: 'Google IDP', type: IdentityProviderTypes.GOOGLE},
          {id: 'github-idp-1', name: 'GitHub IDP', type: IdentityProviderTypes.GITHUB},
        ],
        isLoading: false,
      });

      await render(<ExecutionExtendedProperties resource={githubResource} onChange={mockOnChange} />);

      const select = page.getByRole('combobox');
      await userEvent.click(select);

      await expect.element(page.getByText('GitHub IDP')).toBeInTheDocument();
      await expect.element(page.getByText('Google IDP')).not.toBeInTheDocument();
    });
  });

  describe('SMS OTP Executor', () => {
    const smsOtpResource = {
      id: 'sms-otp-executor-1',
      data: {
        action: {
          executor: {
            name: ExecutionTypes.SMSOTPAuth,
            mode: '',
          },
        },
        properties: {
          senderId: '',
        },
        display: {
          label: 'SMS OTP',
        },
      },
    } as unknown as Resource;

    it('should render SMS OTP configuration UI', async () => {
      mockNotificationSenders.mockReturnValue({
        data: [{id: 'sender-1', name: 'Twilio Sender'}],
        isLoading: false,
      });

      await render(<ExecutionExtendedProperties resource={smsOtpResource} onChange={mockOnChange} />);

      await expect.element(page.getByText('Configure SMS OTP settings')).toBeInTheDocument();
      await expect.element(page.getByText('Mode')).toBeInTheDocument();
      await expect.element(page.getByText('Sender')).toBeInTheDocument();
    });

    it('should show mode options', async () => {
            mockNotificationSenders.mockReturnValue({
        data: [{id: 'sender-1', name: 'Twilio Sender'}],
        isLoading: false,
      });

      await render(<ExecutionExtendedProperties resource={smsOtpResource} onChange={mockOnChange} />);

      const comboboxes = page.getByRole('combobox').all();
      const modeSelect = comboboxes[0];
      await userEvent.click(modeSelect);

      await expect.element(page.getByText('Send SMS OTP')).toBeInTheDocument();
      await expect.element(page.getByText('Verify SMS OTP')).toBeInTheDocument();
    });

    it('should call onChange with updated data when mode is selected', async () => {
            mockNotificationSenders.mockReturnValue({
        data: [{id: 'sender-1', name: 'Twilio Sender'}],
        isLoading: false,
      });

      await render(<ExecutionExtendedProperties resource={smsOtpResource} onChange={mockOnChange} />);

      const comboboxes = page.getByRole('combobox').all();
      const modeSelect = comboboxes[0];
      await userEvent.click(modeSelect);
      await userEvent.click(page.getByText('Send SMS OTP'));

      expect(mockOnChange).toHaveBeenCalledWith(
        'data',
        expect.objectContaining({
          action: expect.objectContaining({
            executor: expect.objectContaining({
              mode: 'send',
            }) as unknown,
          }) as unknown,
          display: expect.objectContaining({
            label: 'Send SMS OTP',
          }) as unknown,
        }),
        smsOtpResource,
      );
    });

    it('should show sender options', async () => {
            mockSelectedNotification.hasResourceFieldNotification.mockReturnValue(false);
      mockNotificationSenders.mockReturnValue({
        data: [
          {id: 'sender-1', name: 'Twilio Sender'},
          {id: 'sender-2', name: 'Vonage Sender'},
        ],
        isLoading: false,
      });

      await render(<ExecutionExtendedProperties resource={smsOtpResource} onChange={mockOnChange} />);

      const comboboxes = page.getByRole('combobox').all();
      const senderSelect = comboboxes[1]; // Second combobox is sender
      await userEvent.click(senderSelect);

      await expect.element(page.getByText('Twilio Sender')).toBeInTheDocument();
      await expect.element(page.getByText('Vonage Sender')).toBeInTheDocument();
    });

    it('should call onChange when sender is selected', async () => {
            mockSelectedNotification.hasResourceFieldNotification.mockReturnValue(false);
      mockNotificationSenders.mockReturnValue({
        data: [{id: 'sender-1', name: 'Twilio Sender'}],
        isLoading: false,
      });

      await render(<ExecutionExtendedProperties resource={smsOtpResource} onChange={mockOnChange} />);

      const comboboxes = page.getByRole('combobox').all();
      const senderSelect = comboboxes[1];
      await userEvent.click(senderSelect);
      await userEvent.click(page.getByText('Twilio Sender'));

      expect(mockOnChange).toHaveBeenCalledWith('data.properties.senderId', 'sender-1', smsOtpResource);
    });

    it('should show error when sender is placeholder', async () => {
      mockSelectedNotification.hasResourceFieldNotification.mockReturnValue(false);
      mockNotificationSenders.mockReturnValue({
        data: [{id: 'sender-1', name: 'Twilio Sender'}],
        isLoading: false,
      });

      const resourceWithPlaceholder = {
        ...smsOtpResource,
        data: {
          ...(smsOtpResource as unknown as {data: object}).data,
          properties: {senderId: '{{SENDER_ID}}'},
        },
      } as unknown as Resource;

      await render(<ExecutionExtendedProperties resource={resourceWithPlaceholder} onChange={mockOnChange} />);

      await expect.element(page.getByText('Sender is required')).toBeInTheDocument();
    });

    it('should show warning when no senders are configured', async () => {
      mockSelectedNotification.hasResourceFieldNotification.mockReturnValue(false);
      mockNotificationSenders.mockReturnValue({
        data: [],
        isLoading: false,
      });

      await render(<ExecutionExtendedProperties resource={smsOtpResource} onChange={mockOnChange} />);

      await expect.element(page.getByText('No SMS senders configured')).toBeInTheDocument();
    });

    it('should disable sender dropdown while loading', async () => {
      mockSelectedNotification.hasResourceFieldNotification.mockReturnValue(false);
      mockNotificationSenders.mockReturnValue({
        data: [],
        isLoading: true,
      });

      await render(<ExecutionExtendedProperties resource={smsOtpResource} onChange={mockOnChange} />);

      const comboboxes = page.getByRole('combobox').all();
      const senderSelect = comboboxes[1];
      expect(senderSelect).toHaveAttribute('aria-disabled', 'true');
    });

    it('should disable sender dropdown when no senders available', async () => {
      mockSelectedNotification.hasResourceFieldNotification.mockReturnValue(false);
      mockNotificationSenders.mockReturnValue({
        data: [],
        isLoading: false,
      });

      await render(<ExecutionExtendedProperties resource={smsOtpResource} onChange={mockOnChange} />);

      const comboboxes = page.getByRole('combobox').all();
      const senderSelect = comboboxes[1];
      expect(senderSelect).toHaveAttribute('aria-disabled', 'true');
    });

    it('should show selected sender value', async () => {
      mockSelectedNotification.hasResourceFieldNotification.mockReturnValue(false);
      mockNotificationSenders.mockReturnValue({
        data: [{id: 'sender-1', name: 'Twilio Sender'}],
        isLoading: false,
      });

      const resourceWithSender = {
        ...smsOtpResource,
        data: {
          ...(smsOtpResource as unknown as {data: object}).data,
          properties: {senderId: 'sender-1'},
        },
      } as unknown as Resource;

      await render(<ExecutionExtendedProperties resource={resourceWithSender} onChange={mockOnChange} />);

      const comboboxes = page.getByRole('combobox').all();
      const senderSelect = comboboxes[1];
      expect(senderSelect).toHaveTextContent('Twilio Sender');
    });

    it('should show selected mode value', async () => {
      mockSelectedNotification.hasResourceFieldNotification.mockReturnValue(false);
      mockNotificationSenders.mockReturnValue({
        data: [{id: 'sender-1', name: 'Twilio Sender'}],
        isLoading: false,
      });

      const resourceWithMode = {
        ...smsOtpResource,
        data: {
          ...(smsOtpResource as unknown as {data: object}).data,
          action: {
            executor: {
              name: ExecutionTypes.SMSOTPAuth,
              mode: 'verify',
            },
          },
        },
      } as unknown as Resource;

      await render(<ExecutionExtendedProperties resource={resourceWithMode} onChange={mockOnChange} />);

      const comboboxes = page.getByRole('combobox').all();
      const modeSelect = comboboxes[0];
      expect(modeSelect).toHaveTextContent('Verify SMS OTP');
    });

    it('should update display label when mode changes to verify', async () => {
            mockNotificationSenders.mockReturnValue({
        data: [{id: 'sender-1', name: 'Twilio Sender'}],
        isLoading: false,
      });

      await render(<ExecutionExtendedProperties resource={smsOtpResource} onChange={mockOnChange} />);

      const comboboxes = page.getByRole('combobox').all();
      const modeSelect = comboboxes[0];
      await userEvent.click(modeSelect);
      await userEvent.click(page.getByText('Verify SMS OTP'));

      expect(mockOnChange).toHaveBeenCalledWith(
        'data',
        expect.objectContaining({
          display: expect.objectContaining({
            label: 'Verify SMS OTP',
          }) as unknown,
        }),
        smsOtpResource,
      );
    });

    it('should preserve existing data properties when mode changes', async () => {
            mockNotificationSenders.mockReturnValue({
        data: [{id: 'sender-1', name: 'Twilio Sender'}],
        isLoading: false,
      });

      const resourceWithExistingData = {
        ...smsOtpResource,
        data: {
          ...(smsOtpResource as unknown as {data: object}).data,
          properties: {senderId: 'sender-1', someOtherProp: 'value'},
          display: {label: 'Old Label', icon: 'icon.png'},
        },
      } as unknown as Resource;

      await render(<ExecutionExtendedProperties resource={resourceWithExistingData} onChange={mockOnChange} />);

      const comboboxes = page.getByRole('combobox').all();
      const modeSelect = comboboxes[0];
      await userEvent.click(modeSelect);
      await userEvent.click(page.getByText('Send SMS OTP'));

      expect(mockOnChange).toHaveBeenCalledWith(
        'data',
        expect.objectContaining({
          properties: expect.objectContaining({
            senderId: 'sender-1',
            someOtherProp: 'value',
          }) as unknown,
          display: expect.objectContaining({
            label: 'Send SMS OTP',
            icon: 'icon.png',
          }) as unknown,
        }),
        resourceWithExistingData,
      );
    });

    it('should preserve display properties when mode changes', async () => {
            mockNotificationSenders.mockReturnValue({
        data: [{id: 'sender-1', name: 'Twilio Sender'}],
        isLoading: false,
      });

      const resourceWithDisplay = {
        ...smsOtpResource,
        data: {
          ...(smsOtpResource as unknown as {data: object}).data,
          display: {icon: 'sms-icon.png'},
        },
      } as unknown as Resource;

      await render(<ExecutionExtendedProperties resource={resourceWithDisplay} onChange={mockOnChange} />);

      const comboboxes = page.getByRole('combobox').all();
      const modeSelect = comboboxes[0];
      await userEvent.click(modeSelect);
      await userEvent.click(page.getByText('Send SMS OTP'));

      // Should preserve existing display properties while updating label
      expect(mockOnChange).toHaveBeenCalledWith(
        'data',
        expect.objectContaining({
          display: expect.objectContaining({
            label: 'Send SMS OTP',
            icon: 'sms-icon.png',
          }) as unknown,
        }),
        resourceWithDisplay,
      );
    });

    it('should show validation error for sender field', async () => {
      (mockSelectedNotification.hasResourceFieldNotification as unknown as ReturnType<typeof vi.fn>).mockImplementation((key: string) =>
        key === 'sms-otp-executor-1_data.properties.senderId'
      );
      (mockSelectedNotification.getResourceFieldNotification as unknown as ReturnType<typeof vi.fn>).mockImplementation((key: string) =>
        key === 'sms-otp-executor-1_data.properties.senderId' ? 'Sender ID is invalid' : ''
      );
      mockNotificationSenders.mockReturnValue({
        data: [{id: 'sender-1', name: 'Twilio Sender'}],
        isLoading: false,
      });

      await render(<ExecutionExtendedProperties resource={smsOtpResource} onChange={mockOnChange} />);

      await expect.element(page.getByText('Sender ID is invalid')).toBeInTheDocument();
    });

    it('should not show warning when senders are still loading', async () => {
      mockSelectedNotification.hasResourceFieldNotification.mockReturnValue(false);
      mockNotificationSenders.mockReturnValue({
        data: [],
        isLoading: true,
      });

      await render(<ExecutionExtendedProperties resource={smsOtpResource} onChange={mockOnChange} />);

      await expect.element(page.getByText('No SMS senders configured')).not.toBeInTheDocument();
    });
  });

  describe('Passkey Executor', () => {
    const passkeyResource = {
      id: 'passkey-executor-1',
      data: {
        action: {
          executor: {
            name: ExecutionTypes.PasskeyAuth,
            mode: '',
          },
        },
        display: {
          label: 'Passkey',
        },
      },
    } as unknown as Resource;

    it('should render Passkey configuration UI', async () => {
      await render(<ExecutionExtendedProperties resource={passkeyResource} onChange={mockOnChange} />);

      await expect.element(page.getByText('Configure Passkey settings')).toBeInTheDocument();
      await expect.element(page.getByText('Mode')).toBeInTheDocument();
    });

    it('should show mode options', async () => {
      
      await render(<ExecutionExtendedProperties resource={passkeyResource} onChange={mockOnChange} />);

      const modeSelect = page.getByRole('combobox');
      await userEvent.click(modeSelect);

      await expect.element(page.getByText('Passkey Challenge')).toBeInTheDocument();
      await expect.element(page.getByText('Passkey Verify')).toBeInTheDocument();
    });

    it('should call onChange with updated data when mode is selected', async () => {
      
      await render(<ExecutionExtendedProperties resource={passkeyResource} onChange={mockOnChange} />);

      const modeSelect = page.getByRole('combobox');
      await userEvent.click(modeSelect);
      await userEvent.click(page.getByText('Passkey Challenge'));

      expect(mockOnChange).toHaveBeenCalledWith(
        'data',
        expect.objectContaining({
          action: expect.objectContaining({
            executor: expect.objectContaining({
              mode: 'challenge',
            }) as unknown,
          }) as unknown,
          display: expect.objectContaining({
            label: 'Request Passkey',
          }) as unknown,
        }),
        passkeyResource,
      );
    });

    it('should show selected mode value', async () => {
      const resourceWithMode = {
        ...passkeyResource,
        data: {
          ...(passkeyResource as unknown as {data: object}).data,
          action: {
            executor: {
              name: ExecutionTypes.PasskeyAuth,
              mode: 'verify',
            },
          },
        },
      } as unknown as Resource;

      await render(<ExecutionExtendedProperties resource={resourceWithMode} onChange={mockOnChange} />);

      const modeSelect = page.getByRole('combobox');
      expect(modeSelect).toHaveTextContent('Passkey Verify');
    });

    it('should update display label when mode changes to verify', async () => {
      
      await render(<ExecutionExtendedProperties resource={passkeyResource} onChange={mockOnChange} />);

      const modeSelect = page.getByRole('combobox');
      await userEvent.click(modeSelect);
      await userEvent.click(page.getByText('Passkey Verify'));

      expect(mockOnChange).toHaveBeenCalledWith(
        'data',
        expect.objectContaining({
          display: expect.objectContaining({
            label: 'Verify Passkey',
          }) as unknown,
        }),
        passkeyResource,
      );
    });

    it('should preserve existing data properties when mode changes', async () => {
      
      const resourceWithExistingData = {
        ...passkeyResource,
        data: {
          ...(passkeyResource as unknown as {data: object}).data,
          properties: {relyingPartyId: 'localhost', relyingPartyName: 'Thunder'},
          display: {label: 'Old Label', icon: 'passkey-icon.png'},
        },
      } as unknown as Resource;

      await render(<ExecutionExtendedProperties resource={resourceWithExistingData} onChange={mockOnChange} />);

      const modeSelect = page.getByRole('combobox');
      await userEvent.click(modeSelect);
      await userEvent.click(page.getByText('Passkey Challenge'));

      expect(mockOnChange).toHaveBeenCalledWith(
        'data',
        expect.objectContaining({
          properties: expect.objectContaining({
            relyingPartyId: 'localhost',
            relyingPartyName: 'Thunder',
          }) as unknown,
          display: expect.objectContaining({
            label: 'Request Passkey',
            icon: 'passkey-icon.png',
          }) as unknown,
        }),
        resourceWithExistingData,
      );
    });
  });

  describe('Edge Cases', () => {
    it('should return null when executor name is not defined', async () => {
      const resourceWithoutExecutor = {
        id: 'resource-1',
        data: {},
      } as unknown as Resource;

      const {container} = await render(
        <ExecutionExtendedProperties resource={resourceWithoutExecutor} onChange={mockOnChange} />,
      );

      expect(container.firstChild).toBeNull();
    });

    it('should return null when executor type is not mapped', async () => {
      const resourceWithUnmappedExecutor = {
        id: 'resource-1',
        data: {
          action: {
            executor: {
              name: 'UnknownExecutor',
            },
          },
        },
      } as unknown as Resource;

      const {container} = await render(
        <ExecutionExtendedProperties resource={resourceWithUnmappedExecutor} onChange={mockOnChange} />,
      );

      expect(container.firstChild).toBeNull();
    });

    it('should handle undefined resource gracefully', async () => {
      const {container} = await render(
        <ExecutionExtendedProperties resource={undefined as unknown as Resource} onChange={mockOnChange} />,
      );

      expect(container.firstChild).toBeNull();
    });

    it('should handle null properties gracefully', async () => {
      const resourceWithNullProperties = {
        id: 'google-executor-1',
        data: {
          action: {
            executor: {
              name: ExecutionTypes.GoogleFederation,
            },
          },
          properties: null,
        },
      } as unknown as Resource;

      mockIdentityProviders.mockReturnValue({
        data: [
          {id: 'google-idp-1', name: 'Google IDP', type: IdentityProviderTypes.GOOGLE},
        ],
        isLoading: false,
      });

      await render(<ExecutionExtendedProperties resource={resourceWithNullProperties} onChange={mockOnChange} />);

      await expect.element(page.getByText('Connection')).toBeInTheDocument();
    });
  });
});
