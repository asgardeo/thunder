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
import userEvent from '@testing-library/user-event';
import {ExecutionTypes} from '@/features/flows/models/steps';
import type {Resource} from '@/features/flows/models/resources';
import {IdentityProviderTypes} from '@/features/integrations/models/identity-provider';
import ExecutionExtendedProperties from '../ExecutionExtendedProperties';

// Mock react-i18next
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => {
      const translations: Record<string, string> = {
        'common:loading': 'Loading...',
        'flows:core.executions.smsOtp.description': 'Configure SMS OTP settings',
        'flows:core.executions.smsOtp.mode.label': 'Mode',
        'flows:core.executions.smsOtp.mode.placeholder': 'Select mode',
        'flows:core.executions.smsOtp.mode.send': 'Send SMS OTP',
        'flows:core.executions.smsOtp.mode.verify': 'Verify SMS OTP',
        'flows:core.executions.smsOtp.sender.label': 'Sender',
        'flows:core.executions.smsOtp.sender.placeholder': 'Select sender',
        'flows:core.executions.smsOtp.sender.required': 'Sender is required',
        'flows:core.executions.smsOtp.sender.noSenders': 'No SMS senders configured',
        'flows:core.executions.passkey.description': 'Configure Passkey settings',
        'flows:core.executions.passkey.mode.label': 'Mode',
        'flows:core.executions.passkey.mode.placeholder': 'Select mode',
        'flows:core.executions.passkey.mode.challenge': 'Passkey Challenge',
        'flows:core.executions.passkey.mode.verify': 'Passkey Verify',
      };
      return translations[key] || key;
    },
  }),
}));

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

    it('should render connection selector for Google executor', () => {
      mockIdentityProviders.mockReturnValue({
        data: [
          {id: 'google-idp-1', name: 'Google IDP', type: IdentityProviderTypes.GOOGLE},
        ],
        isLoading: false,
      });

      render(<ExecutionExtendedProperties resource={googleResource} onChange={mockOnChange} />);

      expect(screen.getByText('Connection')).toBeInTheDocument();
      expect(screen.getByText('Select a connection from the following list to link it with the login flow.')).toBeInTheDocument();
    });

    it('should show available Google connections in dropdown', async () => {
      const user = userEvent.setup();
      mockIdentityProviders.mockReturnValue({
        data: [
          {id: 'google-idp-1', name: 'My Google IDP', type: IdentityProviderTypes.GOOGLE},
          {id: 'google-idp-2', name: 'Another Google IDP', type: IdentityProviderTypes.GOOGLE},
        ],
        isLoading: false,
      });

      render(<ExecutionExtendedProperties resource={googleResource} onChange={mockOnChange} />);

      const select = screen.getByRole('combobox');
      await user.click(select);

      expect(screen.getByText('My Google IDP')).toBeInTheDocument();
      expect(screen.getByText('Another Google IDP')).toBeInTheDocument();
    });

    it('should call onChange when connection is selected', async () => {
      const user = userEvent.setup();
      mockIdentityProviders.mockReturnValue({
        data: [
          {id: 'google-idp-1', name: 'My Google IDP', type: IdentityProviderTypes.GOOGLE},
        ],
        isLoading: false,
      });

      render(<ExecutionExtendedProperties resource={googleResource} onChange={mockOnChange} />);

      const select = screen.getByRole('combobox');
      await user.click(select);
      await user.click(screen.getByText('My Google IDP'));

      expect(mockOnChange).toHaveBeenCalledWith('data.properties.idpId', 'google-idp-1', googleResource);
    });

    it('should show error when connection is placeholder', () => {
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

      render(<ExecutionExtendedProperties resource={resourceWithPlaceholder} onChange={mockOnChange} />);

      expect(screen.getByText('Connection is required and must be selected.')).toBeInTheDocument();
    });

    it('should show validation error from notification', () => {
      mockSelectedNotification.hasResourceFieldNotification.mockReturnValue(true);
      mockSelectedNotification.getResourceFieldNotification.mockReturnValue('Custom validation error');

      mockIdentityProviders.mockReturnValue({
        data: [
          {id: 'google-idp-1', name: 'My Google IDP', type: IdentityProviderTypes.GOOGLE},
        ],
        isLoading: false,
      });

      render(<ExecutionExtendedProperties resource={googleResource} onChange={mockOnChange} />);

      expect(screen.getByText('Custom validation error')).toBeInTheDocument();
    });

    it('should show warning when no connections are available', () => {
      mockIdentityProviders.mockReturnValue({
        data: [],
        isLoading: false,
      });

      render(<ExecutionExtendedProperties resource={googleResource} onChange={mockOnChange} />);

      expect(screen.getByText('No connections available. Please create a connection to link with the login flow.')).toBeInTheDocument();
    });

    it('should disable dropdown while loading', () => {
      mockIdentityProviders.mockReturnValue({
        data: [],
        isLoading: true,
      });

      render(<ExecutionExtendedProperties resource={googleResource} onChange={mockOnChange} />);

      const select = screen.getByRole('combobox');
      expect(select).toHaveAttribute('aria-disabled', 'true');
    });

    it('should show loading text in dropdown while loading', async () => {
      const user = userEvent.setup();
      mockIdentityProviders.mockReturnValue({
        data: [],
        isLoading: true,
      });

      render(<ExecutionExtendedProperties resource={googleResource} onChange={mockOnChange} />);

      const select = screen.getByRole('combobox');
      await user.click(select);

      expect(screen.getByText('Loading...')).toBeInTheDocument();
    });

    it('should show selected connection value', () => {
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

      render(<ExecutionExtendedProperties resource={resourceWithSelection} onChange={mockOnChange} />);

      expect(screen.getByRole('combobox')).toHaveTextContent('My Google IDP');
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

    it('should render connection selector for GitHub executor', () => {
      mockIdentityProviders.mockReturnValue({
        data: [
          {id: 'github-idp-1', name: 'GitHub IDP', type: IdentityProviderTypes.GITHUB},
        ],
        isLoading: false,
      });

      render(<ExecutionExtendedProperties resource={githubResource} onChange={mockOnChange} />);

      expect(screen.getByText('Connection')).toBeInTheDocument();
    });

    it('should filter to only show GitHub connections', async () => {
      const user = userEvent.setup();
      mockIdentityProviders.mockReturnValue({
        data: [
          {id: 'google-idp-1', name: 'Google IDP', type: IdentityProviderTypes.GOOGLE},
          {id: 'github-idp-1', name: 'GitHub IDP', type: IdentityProviderTypes.GITHUB},
        ],
        isLoading: false,
      });

      render(<ExecutionExtendedProperties resource={githubResource} onChange={mockOnChange} />);

      const select = screen.getByRole('combobox');
      await user.click(select);

      expect(screen.getByText('GitHub IDP')).toBeInTheDocument();
      expect(screen.queryByText('Google IDP')).not.toBeInTheDocument();
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

    it('should render SMS OTP configuration UI', () => {
      mockNotificationSenders.mockReturnValue({
        data: [{id: 'sender-1', name: 'Twilio Sender'}],
        isLoading: false,
      });

      render(<ExecutionExtendedProperties resource={smsOtpResource} onChange={mockOnChange} />);

      expect(screen.getByText('Configure SMS OTP settings')).toBeInTheDocument();
      expect(screen.getByText('Mode')).toBeInTheDocument();
      expect(screen.getByText('Sender')).toBeInTheDocument();
    });

    it('should show mode options', async () => {
      const user = userEvent.setup();
      mockNotificationSenders.mockReturnValue({
        data: [{id: 'sender-1', name: 'Twilio Sender'}],
        isLoading: false,
      });

      render(<ExecutionExtendedProperties resource={smsOtpResource} onChange={mockOnChange} />);

      const comboboxes = screen.getAllByRole('combobox');
      const modeSelect = comboboxes[0];
      await user.click(modeSelect);

      expect(screen.getByText('Send SMS OTP')).toBeInTheDocument();
      expect(screen.getByText('Verify SMS OTP')).toBeInTheDocument();
    });

    it('should call onChange with updated data when mode is selected', async () => {
      const user = userEvent.setup();
      mockNotificationSenders.mockReturnValue({
        data: [{id: 'sender-1', name: 'Twilio Sender'}],
        isLoading: false,
      });

      render(<ExecutionExtendedProperties resource={smsOtpResource} onChange={mockOnChange} />);

      const comboboxes = screen.getAllByRole('combobox');
      const modeSelect = comboboxes[0];
      await user.click(modeSelect);
      await user.click(screen.getByText('Send SMS OTP'));

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
      const user = userEvent.setup();
      mockSelectedNotification.hasResourceFieldNotification.mockReturnValue(false);
      mockNotificationSenders.mockReturnValue({
        data: [
          {id: 'sender-1', name: 'Twilio Sender'},
          {id: 'sender-2', name: 'Vonage Sender'},
        ],
        isLoading: false,
      });

      render(<ExecutionExtendedProperties resource={smsOtpResource} onChange={mockOnChange} />);

      const comboboxes = screen.getAllByRole('combobox');
      const senderSelect = comboboxes[1]; // Second combobox is sender
      await user.click(senderSelect);

      expect(screen.getByText('Twilio Sender')).toBeInTheDocument();
      expect(screen.getByText('Vonage Sender')).toBeInTheDocument();
    });

    it('should call onChange when sender is selected', async () => {
      const user = userEvent.setup();
      mockSelectedNotification.hasResourceFieldNotification.mockReturnValue(false);
      mockNotificationSenders.mockReturnValue({
        data: [{id: 'sender-1', name: 'Twilio Sender'}],
        isLoading: false,
      });

      render(<ExecutionExtendedProperties resource={smsOtpResource} onChange={mockOnChange} />);

      const comboboxes = screen.getAllByRole('combobox');
      const senderSelect = comboboxes[1];
      await user.click(senderSelect);
      await user.click(screen.getByText('Twilio Sender'));

      expect(mockOnChange).toHaveBeenCalledWith('data.properties.senderId', 'sender-1', smsOtpResource);
    });

    it('should show error when sender is placeholder', () => {
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

      render(<ExecutionExtendedProperties resource={resourceWithPlaceholder} onChange={mockOnChange} />);

      expect(screen.getByText('Sender is required')).toBeInTheDocument();
    });

    it('should show warning when no senders are configured', () => {
      mockSelectedNotification.hasResourceFieldNotification.mockReturnValue(false);
      mockNotificationSenders.mockReturnValue({
        data: [],
        isLoading: false,
      });

      render(<ExecutionExtendedProperties resource={smsOtpResource} onChange={mockOnChange} />);

      expect(screen.getByText('No SMS senders configured')).toBeInTheDocument();
    });

    it('should disable sender dropdown while loading', () => {
      mockSelectedNotification.hasResourceFieldNotification.mockReturnValue(false);
      mockNotificationSenders.mockReturnValue({
        data: [],
        isLoading: true,
      });

      render(<ExecutionExtendedProperties resource={smsOtpResource} onChange={mockOnChange} />);

      const comboboxes = screen.getAllByRole('combobox');
      const senderSelect = comboboxes[1];
      expect(senderSelect).toHaveAttribute('aria-disabled', 'true');
    });

    it('should disable sender dropdown when no senders available', () => {
      mockSelectedNotification.hasResourceFieldNotification.mockReturnValue(false);
      mockNotificationSenders.mockReturnValue({
        data: [],
        isLoading: false,
      });

      render(<ExecutionExtendedProperties resource={smsOtpResource} onChange={mockOnChange} />);

      const comboboxes = screen.getAllByRole('combobox');
      const senderSelect = comboboxes[1];
      expect(senderSelect).toHaveAttribute('aria-disabled', 'true');
    });

    it('should show selected sender value', () => {
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

      render(<ExecutionExtendedProperties resource={resourceWithSender} onChange={mockOnChange} />);

      const comboboxes = screen.getAllByRole('combobox');
      const senderSelect = comboboxes[1];
      expect(senderSelect).toHaveTextContent('Twilio Sender');
    });

    it('should show selected mode value', () => {
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

      render(<ExecutionExtendedProperties resource={resourceWithMode} onChange={mockOnChange} />);

      const comboboxes = screen.getAllByRole('combobox');
      const modeSelect = comboboxes[0];
      expect(modeSelect).toHaveTextContent('Verify SMS OTP');
    });

    it('should update display label when mode changes to verify', async () => {
      const user = userEvent.setup();
      mockNotificationSenders.mockReturnValue({
        data: [{id: 'sender-1', name: 'Twilio Sender'}],
        isLoading: false,
      });

      render(<ExecutionExtendedProperties resource={smsOtpResource} onChange={mockOnChange} />);

      const comboboxes = screen.getAllByRole('combobox');
      const modeSelect = comboboxes[0];
      await user.click(modeSelect);
      await user.click(screen.getByText('Verify SMS OTP'));

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
      const user = userEvent.setup();
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

      render(<ExecutionExtendedProperties resource={resourceWithExistingData} onChange={mockOnChange} />);

      const comboboxes = screen.getAllByRole('combobox');
      const modeSelect = comboboxes[0];
      await user.click(modeSelect);
      await user.click(screen.getByText('Send SMS OTP'));

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
      const user = userEvent.setup();
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

      render(<ExecutionExtendedProperties resource={resourceWithDisplay} onChange={mockOnChange} />);

      const comboboxes = screen.getAllByRole('combobox');
      const modeSelect = comboboxes[0];
      await user.click(modeSelect);
      await user.click(screen.getByText('Send SMS OTP'));

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

    it('should show validation error for sender field', () => {
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

      render(<ExecutionExtendedProperties resource={smsOtpResource} onChange={mockOnChange} />);

      expect(screen.getByText('Sender ID is invalid')).toBeInTheDocument();
    });

    it('should not show warning when senders are still loading', () => {
      mockSelectedNotification.hasResourceFieldNotification.mockReturnValue(false);
      mockNotificationSenders.mockReturnValue({
        data: [],
        isLoading: true,
      });

      render(<ExecutionExtendedProperties resource={smsOtpResource} onChange={mockOnChange} />);

      expect(screen.queryByText('No SMS senders configured')).not.toBeInTheDocument();
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

    it('should render Passkey configuration UI', () => {
      render(<ExecutionExtendedProperties resource={passkeyResource} onChange={mockOnChange} />);

      expect(screen.getByText('Configure Passkey settings')).toBeInTheDocument();
      expect(screen.getByText('Mode')).toBeInTheDocument();
    });

    it('should show mode options', async () => {
      const user = userEvent.setup();

      render(<ExecutionExtendedProperties resource={passkeyResource} onChange={mockOnChange} />);

      const modeSelect = screen.getByRole('combobox');
      await user.click(modeSelect);

      expect(screen.getByText('Passkey Challenge')).toBeInTheDocument();
      expect(screen.getByText('Passkey Verify')).toBeInTheDocument();
    });

    it('should call onChange with updated data when mode is selected', async () => {
      const user = userEvent.setup();

      render(<ExecutionExtendedProperties resource={passkeyResource} onChange={mockOnChange} />);

      const modeSelect = screen.getByRole('combobox');
      await user.click(modeSelect);
      await user.click(screen.getByText('Passkey Challenge'));

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

    it('should show selected mode value', () => {
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

      render(<ExecutionExtendedProperties resource={resourceWithMode} onChange={mockOnChange} />);

      const modeSelect = screen.getByRole('combobox');
      expect(modeSelect).toHaveTextContent('Passkey Verify');
    });

    it('should update display label when mode changes to verify', async () => {
      const user = userEvent.setup();

      render(<ExecutionExtendedProperties resource={passkeyResource} onChange={mockOnChange} />);

      const modeSelect = screen.getByRole('combobox');
      await user.click(modeSelect);
      await user.click(screen.getByText('Passkey Verify'));

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
      const user = userEvent.setup();

      const resourceWithExistingData = {
        ...passkeyResource,
        data: {
          ...(passkeyResource as unknown as {data: object}).data,
          properties: {relyingPartyId: 'localhost', relyingPartyName: 'Thunder'},
          display: {label: 'Old Label', icon: 'passkey-icon.png'},
        },
      } as unknown as Resource;

      render(<ExecutionExtendedProperties resource={resourceWithExistingData} onChange={mockOnChange} />);

      const modeSelect = screen.getByRole('combobox');
      await user.click(modeSelect);
      await user.click(screen.getByText('Passkey Challenge'));

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
    it('should return null when executor name is not defined', () => {
      const resourceWithoutExecutor = {
        id: 'resource-1',
        data: {},
      } as unknown as Resource;

      const {container} = render(
        <ExecutionExtendedProperties resource={resourceWithoutExecutor} onChange={mockOnChange} />,
      );

      expect(container.firstChild).toBeNull();
    });

    it('should return null when executor type is not mapped', () => {
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

      const {container} = render(
        <ExecutionExtendedProperties resource={resourceWithUnmappedExecutor} onChange={mockOnChange} />,
      );

      expect(container.firstChild).toBeNull();
    });

    it('should handle undefined resource gracefully', () => {
      const {container} = render(
        <ExecutionExtendedProperties resource={undefined as unknown as Resource} onChange={mockOnChange} />,
      );

      expect(container.firstChild).toBeNull();
    });

    it('should handle null properties gracefully', () => {
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

      render(<ExecutionExtendedProperties resource={resourceWithNullProperties} onChange={mockOnChange} />);

      expect(screen.getByText('Connection')).toBeInTheDocument();
    });
  });
});
