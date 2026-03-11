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

import {describe, it, expect, vi, beforeEach, afterEach} from 'vitest';
import {page, userEvent} from 'vitest/browser';
import {renderWithProviders, getByDisplayValue} from '@thunder/test-utils/browser';
import type {JSX} from 'react';
import {EmbeddedFlowComponentType, EmbeddedFlowEventType} from '@asgardeo/react';
import type {InviteUserRenderProps, EmbeddedFlowComponent} from '@asgardeo/react';
import InviteUserDialog from '../InviteUserDialog';

const {mockLoggerError} = vi.hoisted(() => ({
  mockLoggerError: vi.fn(),
}));

vi.mock('@thunder/logger/react', () => ({
  useLogger: () => ({
    info: vi.fn(),
    warn: vi.fn(),
    error: mockLoggerError,
    debug: vi.fn(),
  }),
}));

// Mock InviteUser component
const mockHandleInputChange = vi.fn();
const mockHandleInputBlur = vi.fn();
const mockHandleSubmit = vi.fn();
const mockCopyInviteLink = vi.fn().mockResolvedValue(undefined);
const mockResetFlow = vi.fn();

const mockInviteUserRenderProps: InviteUserRenderProps = {
  values: {},
  fieldErrors: {},
  touched: {},
  error: null,
  isLoading: false,
  components: [],
  handleInputChange: mockHandleInputChange,
  handleInputBlur: mockHandleInputBlur,
  handleSubmit: mockHandleSubmit,
  isInviteGenerated: false,
  inviteLink: undefined,
  copyInviteLink: mockCopyInviteLink,
  inviteLinkCopied: false,
  resetFlow: mockResetFlow,
  isValid: false,
  meta: null,
};

// Track whether to simulate an error in the InviteUser mock
let simulateInviteUserError = false;
const mockInviteUserError = new Error('User onboarding failed');

vi.mock('@asgardeo/react', async () => {
  const actual = await vi.importActual<typeof import('@asgardeo/react')>('@asgardeo/react');
  return {
    ...actual,
    InviteUser: ({
      children,
      onInviteLinkGenerated,
      onError,
    }: {
      children: (props: InviteUserRenderProps) => JSX.Element;
      onInviteLinkGenerated?: (link: string) => void;
      onError?: (error: Error) => void;
    }) => {
      // Call onInviteLinkGenerated if invite is generated
      if (
        mockInviteUserRenderProps.isInviteGenerated &&
        mockInviteUserRenderProps.inviteLink &&
        onInviteLinkGenerated
      ) {
        // Use setTimeout to simulate async behavior
        setTimeout(() => {
          onInviteLinkGenerated(mockInviteUserRenderProps.inviteLink!);
        }, 0);
      }
      // Call onError if simulating an error
      if (simulateInviteUserError && onError) {
        setTimeout(() => {
          onError(mockInviteUserError);
        }, 0);
      }
      return children(mockInviteUserRenderProps);
    },
  };
});

// Mock useTemplateLiteralResolver
vi.mock('@thunder/shared-hooks', () => ({
  useTemplateLiteralResolver: () => ({
    resolve: (key: string) => key,
  }),
}));

describe('InviteUserDialog', () => {
  const mockOnClose = vi.fn();
  const mockOnSuccess = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
    // Reset error simulation flag
    simulateInviteUserError = false;
    // Reset mock props
    Object.assign(mockInviteUserRenderProps, {
      values: {},
      fieldErrors: {},
      touched: {},
      error: null,
      isLoading: false,
      components: [],
      handleInputChange: mockHandleInputChange,
      handleInputBlur: mockHandleInputBlur,
      handleSubmit: mockHandleSubmit,
      isInviteGenerated: false,
      inviteLink: undefined,
      copyInviteLink: mockCopyInviteLink,
      inviteLinkCopied: false,
      resetFlow: mockResetFlow,
      isValid: false,
    });
  });

  afterEach(() => {
    vi.clearAllTimers();
  });

  it('renders dialog when open', async () => {
    await renderWithProviders(<InviteUserDialog open onClose={mockOnClose} />);

    await expect.element(page.getByText('Invite User')).toBeInTheDocument();
    await expect.element(page.getByText(/Send an invite link to a new user/i)).toBeInTheDocument();
  });

  it('does not render when closed', async () => {
    await renderWithProviders(<InviteUserDialog open={false} onClose={mockOnClose} />);

    await expect.element(page.getByText('Invite User')).not.toBeInTheDocument();
  });

  it('closes dialog when close button is clicked', async () => {
    await renderWithProviders(<InviteUserDialog open onClose={mockOnClose} />);

    const closeButton = page.getByLabelText('close');
    await userEvent.click(closeButton);

    expect(mockOnClose).toHaveBeenCalled();
  });

  it('displays loading state when initializing', async () => {
    Object.assign(mockInviteUserRenderProps, {
      isLoading: true,
      components: [],
      isInviteGenerated: false,
    });

    await renderWithProviders(<InviteUserDialog open onClose={mockOnClose} />);

    await expect.element(page.getByRole('progressbar')).toBeInTheDocument();
  });

  it('displays form fields when components are available', async () => {
    const emailComponent: EmbeddedFlowComponent = {
      id: 'email_input',
      ref: 'email',
      type: 'EMAIL_INPUT' as EmbeddedFlowComponent['type'],
      label: 'Email',
      placeholder: 'Enter email',
      required: true,
    };

    const submitAction: EmbeddedFlowComponent = {
      id: 'submit_action',
      type: EmbeddedFlowComponentType.Action,
      label: 'Next',
      variant: 'PRIMARY',
      eventType: EmbeddedFlowEventType.Submit,
    };

    const blockComponent: EmbeddedFlowComponent = {
      id: 'block',
      type: EmbeddedFlowComponentType.Block,
      components: [emailComponent, submitAction],
    };

    Object.assign(mockInviteUserRenderProps, {
      components: [blockComponent],
      values: {email: ''},
      isValid: false,
      isLoading: false,
    });

    await renderWithProviders(<InviteUserDialog open onClose={mockOnClose} />);

    // Wait for form to render and check for email input by placeholder
    await expect.element(page.getByPlaceholder('Enter email')).toBeInTheDocument();
    // Check for label - it might be "Email" or a translation key
    const labels = page.getByText('Email').all();
    expect(labels.length).toBeGreaterThan(0);
  });

  it('disables submit button when form is invalid', async () => {
    const emailComponent: EmbeddedFlowComponent = {
      id: 'email_input',
      ref: 'email',
      type: 'EMAIL_INPUT' as EmbeddedFlowComponent['type'],
      label: 'Email',
      placeholder: 'Enter email',
      required: true,
    };

    const submitAction: EmbeddedFlowComponent = {
      id: 'submit_action',
      type: EmbeddedFlowComponentType.Action,
      label: 'Next',
      variant: 'PRIMARY',
      eventType: EmbeddedFlowEventType.Submit,
    };

    const blockComponent: EmbeddedFlowComponent = {
      id: 'block',
      type: EmbeddedFlowComponentType.Block,
      components: [emailComponent, submitAction],
    };

    Object.assign(mockInviteUserRenderProps, {
      components: [blockComponent],
      values: {email: ''},
      isValid: false,
    });

    await renderWithProviders(<InviteUserDialog open onClose={mockOnClose} />);

    const submitButton = page.getByRole('button', {name: /next/i});
    await expect.element(submitButton).toBeDisabled();
  });

  it('enables submit button when form is valid', async () => {
    const emailComponent: EmbeddedFlowComponent = {
      id: 'email_input',
      ref: 'email',
      type: 'EMAIL_INPUT' as EmbeddedFlowComponent['type'],
      label: 'Email',
      placeholder: 'Enter email',
      required: true,
    };

    const submitAction: EmbeddedFlowComponent = {
      id: 'submit_action',
      type: EmbeddedFlowComponentType.Action,
      label: 'Next',
      variant: 'PRIMARY',
      eventType: EmbeddedFlowEventType.Submit,
    };

    const blockComponent: EmbeddedFlowComponent = {
      id: 'block',
      type: EmbeddedFlowComponentType.Block,
      components: [emailComponent, submitAction],
    };

    Object.assign(mockInviteUserRenderProps, {
      components: [blockComponent],
      values: {email: 'test@example.com'},
      isValid: true,
    });

    await renderWithProviders(<InviteUserDialog open onClose={mockOnClose} />);

    const submitButton = page.getByRole('button', {name: /next/i});
    await expect.element(submitButton).not.toBeDisabled();
  });

  it('calls handleInputChange when input value changes', async () => {
    const emailComponent: EmbeddedFlowComponent = {
      id: 'email_input',
      ref: 'email',
      type: 'EMAIL_INPUT' as EmbeddedFlowComponent['type'],
      label: 'Email',
      placeholder: 'Enter email',
      required: true,
    };

    const submitAction: EmbeddedFlowComponent = {
      id: 'submit_action',
      type: EmbeddedFlowComponentType.Action,
      label: 'Next',
      variant: 'PRIMARY',
      eventType: EmbeddedFlowEventType.Submit,
    };

    const blockComponent: EmbeddedFlowComponent = {
      id: 'block',
      type: EmbeddedFlowComponentType.Block,
      components: [emailComponent, submitAction],
    };

    Object.assign(mockInviteUserRenderProps, {
      components: [blockComponent],
      values: {email: ''},
      isValid: false,
    });

    await renderWithProviders(<InviteUserDialog open onClose={mockOnClose} />);

    const emailInput = page.getByPlaceholder('Enter email');
    await userEvent.fill(emailInput, 'test@example.com');

    await vi.waitFor(() => {
      expect(mockHandleInputChange).toHaveBeenCalledWith('email', 'test@example.com');
    });
  });

  it('submits form when submit button is clicked', async () => {
    const emailComponent: EmbeddedFlowComponent = {
      id: 'email_input',
      ref: 'email',
      type: 'EMAIL_INPUT' as EmbeddedFlowComponent['type'],
      label: 'Email',
      placeholder: 'Enter email',
      required: true,
    };

    const submitAction: EmbeddedFlowComponent = {
      id: 'submit_action',
      type: EmbeddedFlowComponentType.Action,
      label: 'Next',
      variant: 'PRIMARY',
      eventType: EmbeddedFlowEventType.Submit,
    };

    const blockComponent: EmbeddedFlowComponent = {
      id: 'block',
      type: EmbeddedFlowComponentType.Block,
      components: [emailComponent, submitAction],
    };

    Object.assign(mockInviteUserRenderProps, {
      components: [blockComponent],
      values: {email: 'test@example.com'},
      isValid: true,
    });

    mockHandleSubmit.mockResolvedValue({});

    await renderWithProviders(<InviteUserDialog open onClose={mockOnClose} />);

    const submitButton = page.getByRole('button', {name: /next/i});
    await userEvent.click(submitButton);

    await vi.waitFor(() => {
      expect(mockHandleSubmit).toHaveBeenCalledWith(submitAction, {email: 'test@example.com'});
    });
  });

  it('displays error message when error occurs', async () => {
    Object.assign(mockInviteUserRenderProps, {
      error: new Error('Failed to invite user'),
      components: [],
    });

    await renderWithProviders(<InviteUserDialog open onClose={mockOnClose} />);

    await expect.element(page.getByText(/Error/i)).toBeInTheDocument();
    await expect.element(page.getByText('Failed to invite user')).toBeInTheDocument();
  });

  it('displays invite link when invite is generated', async () => {
    const inviteLink = 'https://example.com/invite?token=abc123';
    Object.assign(mockInviteUserRenderProps, {
      isInviteGenerated: true,
      inviteLink,
    });

    await renderWithProviders(<InviteUserDialog open onClose={mockOnClose} onSuccess={mockOnSuccess} />);

    await expect.element(page.getByText(/Invite Link Generated!/i)).toBeInTheDocument();
    await expect.element(getByDisplayValue(inviteLink)).toBeInTheDocument();
  });

  it('calls onSuccess when invite link is generated', async () => {
    const inviteLink = 'https://example.com/invite?token=abc123';
    Object.assign(mockInviteUserRenderProps, {
      isInviteGenerated: true,
      inviteLink,
    });

    await renderWithProviders(<InviteUserDialog open onClose={mockOnClose} onSuccess={mockOnSuccess} />);

    await vi.waitFor(() => {
      expect(mockOnSuccess).toHaveBeenCalledWith(inviteLink);
    });
  });

  it('copies invite link when copy button is clicked', async () => {
    const inviteLink = 'https://example.com/invite?token=abc123';
    Object.assign(mockInviteUserRenderProps, {
      isInviteGenerated: true,
      inviteLink,
      inviteLinkCopied: false,
    });

    await renderWithProviders(<InviteUserDialog open onClose={mockOnClose} />);

    const copyButton = page.getByRole('button', {name: /copy/i});
    await userEvent.click(copyButton);

    expect(mockCopyInviteLink).toHaveBeenCalled();
  });

  it('resets flow when "Invite Another User" is clicked', async () => {
    const inviteLink = 'https://example.com/invite?token=abc123';
    Object.assign(mockInviteUserRenderProps, {
      isInviteGenerated: true,
      inviteLink,
    });

    await renderWithProviders(<InviteUserDialog open onClose={mockOnClose} />);

    const inviteAnotherButton = page.getByRole('button', {name: /Invite Another User/i});
    await userEvent.click(inviteAnotherButton);

    expect(mockResetFlow).toHaveBeenCalled();
  });

  it('displays stepper with correct steps', async () => {
    await renderWithProviders(<InviteUserDialog open onClose={mockOnClose} />);

    await expect.element(page.getByText('User Details')).toBeInTheDocument();
    await expect.element(page.getByText('Invite Link')).toBeInTheDocument();
  });

  it('renders SELECT field with placeholder when no value selected', async () => {
    const selectComponent: EmbeddedFlowComponent = {
      id: 'user_type_select',
      ref: 'userType',
      type: 'SELECT' as EmbeddedFlowComponent['type'],
      label: 'User Type',
      placeholder: 'Select user type',
      required: true,
      options: [
        {value: 'admin', label: 'Admin'},
        {value: 'user', label: 'User'},
      ],
    };

    const submitAction: EmbeddedFlowComponent = {
      id: 'submit_action',
      type: EmbeddedFlowComponentType.Action,
      label: 'Continue',
      variant: 'PRIMARY',
      eventType: EmbeddedFlowEventType.Submit,
    };

    const blockComponent: EmbeddedFlowComponent = {
      id: 'block',
      type: EmbeddedFlowComponentType.Block,
      components: [selectComponent, submitAction],
    };

    Object.assign(mockInviteUserRenderProps, {
      components: [blockComponent],
      values: {userType: ''},
      isValid: false,
    });

    await renderWithProviders(<InviteUserDialog open onClose={mockOnClose} />);

    const select = page.getByRole('combobox');
    await expect.element(select).toHaveTextContent('Select user type');
  });

  it('displays loading spinner on submit button when loading', async () => {
    const emailComponent: EmbeddedFlowComponent = {
      id: 'email_input',
      ref: 'email',
      type: 'EMAIL_INPUT' as EmbeddedFlowComponent['type'],
      label: 'Email',
      placeholder: 'Enter email',
      required: true,
    };

    const submitAction: EmbeddedFlowComponent = {
      id: 'submit_action',
      type: EmbeddedFlowComponentType.Action,
      label: 'Next',
      variant: 'PRIMARY',
      eventType: EmbeddedFlowEventType.Submit,
    };

    const blockComponent: EmbeddedFlowComponent = {
      id: 'block',
      type: EmbeddedFlowComponentType.Block,
      components: [emailComponent, submitAction],
    };

    Object.assign(mockInviteUserRenderProps, {
      components: [blockComponent],
      values: {email: 'test@example.com'},
      isValid: true,
      isLoading: true,
    });

    await renderWithProviders(<InviteUserDialog open onClose={mockOnClose} />);

    // When loading, the submit button shows a spinner instead of text
    const spinner = page.getByRole('progressbar');
    await expect.element(spinner).toBeInTheDocument();
    // The spinner's parent button should be disabled
    const spinnerEl = spinner.element() as HTMLElement;
    const submitButton = spinnerEl.closest('button');
    expect(submitButton).toBeDisabled();
  });

  it('disables cancel button when loading', async () => {
    const emailComponent: EmbeddedFlowComponent = {
      id: 'email_input',
      ref: 'email',
      type: 'EMAIL_INPUT' as EmbeddedFlowComponent['type'],
      label: 'Email',
      placeholder: 'Enter email',
      required: true,
    };

    const submitAction: EmbeddedFlowComponent = {
      id: 'submit_action',
      type: EmbeddedFlowComponentType.Action,
      label: 'Next',
      variant: 'PRIMARY',
      eventType: EmbeddedFlowEventType.Submit,
    };

    const blockComponent: EmbeddedFlowComponent = {
      id: 'block',
      type: EmbeddedFlowComponentType.Block,
      components: [emailComponent, submitAction],
    };

    Object.assign(mockInviteUserRenderProps, {
      isLoading: true,
      components: [blockComponent],
      values: {email: ''},
      isValid: false,
    });

    await renderWithProviders(<InviteUserDialog open onClose={mockOnClose} />);

    const cancelButton = page.getByRole('button', {name: /cancel/i});
    await expect.element(cancelButton).toBeDisabled();
  });

  it('validates email format', async () => {
    const emailComponent: EmbeddedFlowComponent = {
      id: 'email_input',
      ref: 'email',
      type: 'EMAIL_INPUT' as EmbeddedFlowComponent['type'],
      label: 'Email',
      placeholder: 'Enter email',
      required: true,
    };

    const submitAction: EmbeddedFlowComponent = {
      id: 'submit_action',
      type: EmbeddedFlowComponentType.Action,
      label: 'Next',
      variant: 'PRIMARY',
      eventType: EmbeddedFlowEventType.Submit,
    };

    const blockComponent: EmbeddedFlowComponent = {
      id: 'block',
      type: EmbeddedFlowComponentType.Block,
      components: [emailComponent, submitAction],
    };

    Object.assign(mockInviteUserRenderProps, {
      components: [blockComponent],
      values: {email: 'invalid-email'},
      isValid: false,
    });

    await renderWithProviders(<InviteUserDialog open onClose={mockOnClose} />);

    const emailInput = page.getByPlaceholder('Enter email');
    await userEvent.fill(emailInput, 'invalid-email');

    // Form validation should prevent submission
    const submitButton = page.getByRole('button', {name: /next/i});
    await expect.element(submitButton).toBeDisabled();
  });

  it('renders TEXT_INPUT field correctly', async () => {
    const textInputComponent: EmbeddedFlowComponent = {
      id: 'first_name_input',
      ref: 'firstName',
      type: EmbeddedFlowComponentType.TextInput,
      label: 'First Name',
      placeholder: 'Enter first name',
      required: true,
    };

    const submitAction: EmbeddedFlowComponent = {
      id: 'submit_action',
      type: EmbeddedFlowComponentType.Action,
      label: 'Next',
      variant: 'PRIMARY',
      eventType: EmbeddedFlowEventType.Submit,
    };

    const blockComponent: EmbeddedFlowComponent = {
      id: 'block',
      type: EmbeddedFlowComponentType.Block,
      components: [textInputComponent, submitAction],
    };

    Object.assign(mockInviteUserRenderProps, {
      components: [blockComponent],
      values: {firstName: ''},
      isValid: false,
    });

    await renderWithProviders(<InviteUserDialog open onClose={mockOnClose} />);

    const textInput = page.getByPlaceholder('Enter first name');
    await expect.element(textInput).toBeInTheDocument();

    await userEvent.fill(textInput, 'John');

    await vi.waitFor(() => {
      expect(mockHandleInputChange).toHaveBeenCalledWith('firstName', 'John');
    });
  });

  it('renders TEXT component with body variant', async () => {
    const textComponent: EmbeddedFlowComponent = {
      id: 'description_text',
      type: EmbeddedFlowComponentType.Text,
      label: 'Please fill in the user details below',
      variant: 'BODY_1',
    };

    const emailComponent: EmbeddedFlowComponent = {
      id: 'email_input',
      ref: 'email',
      type: 'EMAIL_INPUT' as EmbeddedFlowComponent['type'],
      label: 'Email',
      placeholder: 'Enter email',
      required: true,
    };

    const submitAction: EmbeddedFlowComponent = {
      id: 'submit_action',
      type: EmbeddedFlowComponentType.Action,
      label: 'Next',
      variant: 'PRIMARY',
      eventType: EmbeddedFlowEventType.Submit,
    };

    const blockComponent: EmbeddedFlowComponent = {
      id: 'block',
      type: EmbeddedFlowComponentType.Block,
      components: [emailComponent, submitAction],
    };

    Object.assign(mockInviteUserRenderProps, {
      components: [textComponent, blockComponent],
      values: {email: ''},
      isValid: false,
    });

    await renderWithProviders(<InviteUserDialog open onClose={mockOnClose} />);

    await expect.element(page.getByText('Please fill in the user details below')).toBeInTheDocument();
  });

  it('skips TEXT component with HEADING_1 variant', async () => {
    const headingComponent: EmbeddedFlowComponent = {
      id: 'heading_text',
      type: EmbeddedFlowComponentType.Text,
      label: 'Main Heading',
      variant: 'HEADING_1',
    };

    const emailComponent: EmbeddedFlowComponent = {
      id: 'email_input',
      ref: 'email',
      type: 'EMAIL_INPUT' as EmbeddedFlowComponent['type'],
      label: 'Email',
      placeholder: 'Enter email',
      required: true,
    };

    const submitAction: EmbeddedFlowComponent = {
      id: 'submit_action',
      type: EmbeddedFlowComponentType.Action,
      label: 'Next',
      variant: 'PRIMARY',
      eventType: EmbeddedFlowEventType.Submit,
    };

    const blockComponent: EmbeddedFlowComponent = {
      id: 'block',
      type: EmbeddedFlowComponentType.Block,
      components: [emailComponent, submitAction],
    };

    Object.assign(mockInviteUserRenderProps, {
      components: [headingComponent, blockComponent],
      values: {email: ''},
      isValid: false,
    });

    await renderWithProviders(<InviteUserDialog open onClose={mockOnClose} />);

    // HEADING_1 should not be rendered
    await expect.element(page.getByText('Main Heading')).not.toBeInTheDocument();
  });

  it('displays error message when error occurs with components present', async () => {
    const emailComponent: EmbeddedFlowComponent = {
      id: 'email_input',
      ref: 'email',
      type: 'EMAIL_INPUT' as EmbeddedFlowComponent['type'],
      label: 'Email',
      placeholder: 'Enter email',
      required: true,
    };

    const submitAction: EmbeddedFlowComponent = {
      id: 'submit_action',
      type: EmbeddedFlowComponentType.Action,
      label: 'Next',
      variant: 'PRIMARY',
      eventType: EmbeddedFlowEventType.Submit,
    };

    const blockComponent: EmbeddedFlowComponent = {
      id: 'block',
      type: EmbeddedFlowComponentType.Block,
      components: [emailComponent, submitAction],
    };

    Object.assign(mockInviteUserRenderProps, {
      error: new Error('Email already exists'),
      components: [blockComponent],
      values: {email: 'existing@example.com'},
      isValid: true,
    });

    await renderWithProviders(<InviteUserDialog open onClose={mockOnClose} />);

    // Error should be displayed above the form
    await expect.element(page.getByText('Email already exists')).toBeInTheDocument();
    // Form should still be visible
    await expect.element(page.getByPlaceholder('Enter email')).toBeInTheDocument();
  });

  it('closes dialog when cancel button is clicked', async () => {
    const emailComponent: EmbeddedFlowComponent = {
      id: 'email_input',
      ref: 'email',
      type: 'EMAIL_INPUT' as EmbeddedFlowComponent['type'],
      label: 'Email',
      placeholder: 'Enter email',
      required: true,
    };

    const submitAction: EmbeddedFlowComponent = {
      id: 'submit_action',
      type: EmbeddedFlowComponentType.Action,
      label: 'Next',
      variant: 'PRIMARY',
      eventType: EmbeddedFlowEventType.Submit,
    };

    const blockComponent: EmbeddedFlowComponent = {
      id: 'block',
      type: EmbeddedFlowComponentType.Block,
      components: [emailComponent, submitAction],
    };

    Object.assign(mockInviteUserRenderProps, {
      components: [blockComponent],
      values: {email: ''},
      isValid: false,
    });

    await renderWithProviders(<InviteUserDialog open onClose={mockOnClose} />);

    const cancelButton = page.getByRole('button', {name: /cancel/i});
    await userEvent.click(cancelButton);

    expect(mockOnClose).toHaveBeenCalled();
  });

  it('returns null for block without submit action', async () => {
    const emailComponent: EmbeddedFlowComponent = {
      id: 'email_input',
      ref: 'email',
      type: 'EMAIL_INPUT' as EmbeddedFlowComponent['type'],
      label: 'Email',
      placeholder: 'Enter email',
      required: true,
    };

    // Block without submit action
    const blockComponent: EmbeddedFlowComponent = {
      id: 'block',
      type: EmbeddedFlowComponentType.Block,
      components: [emailComponent], // No submit action
    };

    Object.assign(mockInviteUserRenderProps, {
      components: [blockComponent],
      values: {email: ''},
      isValid: false,
    });

    await renderWithProviders(<InviteUserDialog open onClose={mockOnClose} />);

    // Block should not render without submit action
    await expect.element(page.getByPlaceholder('Enter email')).not.toBeInTheDocument();
  });

  it('handles SELECT field value changes', async () => {
    const selectComponent: EmbeddedFlowComponent = {
      id: 'role_select',
      ref: 'role',
      type: 'SELECT' as EmbeddedFlowComponent['type'],
      label: 'Role',
      placeholder: 'Select a role',
      required: true,
      options: [
        {value: 'admin', label: 'Administrator'},
        {value: 'user', label: 'Standard User'},
      ],
    };

    const submitAction: EmbeddedFlowComponent = {
      id: 'submit_action',
      type: EmbeddedFlowComponentType.Action,
      label: 'Next',
      variant: 'PRIMARY',
      eventType: EmbeddedFlowEventType.Submit,
    };

    const blockComponent: EmbeddedFlowComponent = {
      id: 'block',
      type: EmbeddedFlowComponentType.Block,
      components: [selectComponent, submitAction],
    };

    Object.assign(mockInviteUserRenderProps, {
      components: [blockComponent],
      values: {role: ''},
      isValid: false,
    });

    await renderWithProviders(<InviteUserDialog open onClose={mockOnClose} />);

    // Open the select dropdown
    const select = page.getByRole('combobox');
    await userEvent.click(select);

    // Select an option
    const adminOption = page.getByRole('option', {name: 'Administrator'});
    await userEvent.click(adminOption);

    await vi.waitFor(() => {
      expect(mockHandleInputChange).toHaveBeenCalledWith('role', 'admin');
    });
  });

  it('displays check icon when invite link is already copied', async () => {
    const inviteLink = 'https://example.com/invite?token=abc123';
    Object.assign(mockInviteUserRenderProps, {
      isInviteGenerated: true,
      inviteLink,
      inviteLinkCopied: true, // Already copied via SDK
    });

    await renderWithProviders(<InviteUserDialog open onClose={mockOnClose} />);

    // The copy button should show check icon when inviteLinkCopied is true
    const copyButton = page.getByRole('button', {name: /copy/i});
    await expect.element(copyButton).toBeInTheDocument();
  });

  it('closes dialog from invite generated screen', async () => {
    const inviteLink = 'https://example.com/invite?token=abc123';
    Object.assign(mockInviteUserRenderProps, {
      isInviteGenerated: true,
      inviteLink,
    });

    await renderWithProviders(<InviteUserDialog open onClose={mockOnClose} />);

    // Find the "Close" button (not the X icon button which has aria-label="close")
    const closeButtons = page.getByRole('button', {name: /close/i}).all();
    // The "Close" text button should be in the generated invite screen
    const closeTextButton = closeButtons.find((btn) => btn.element().textContent === 'Close');
    expect(closeTextButton).toBeDefined();
    await userEvent.click(closeTextButton!);

    expect(mockOnClose).toHaveBeenCalled();
  });

  it('handles getOptionValue with string option', async () => {
    const selectComponent: EmbeddedFlowComponent = {
      id: 'country_select',
      ref: 'country',
      type: 'SELECT' as EmbeddedFlowComponent['type'],
      label: 'Country',
      placeholder: 'Select country',
      required: true,
      options: ['USA', 'Canada', 'UK'], // String options
    };

    const submitAction: EmbeddedFlowComponent = {
      id: 'submit_action',
      type: EmbeddedFlowComponentType.Action,
      label: 'Next',
      variant: 'PRIMARY',
      eventType: EmbeddedFlowEventType.Submit,
    };

    const blockComponent: EmbeddedFlowComponent = {
      id: 'block',
      type: EmbeddedFlowComponentType.Block,
      components: [selectComponent, submitAction],
    };

    Object.assign(mockInviteUserRenderProps, {
      components: [blockComponent],
      values: {country: ''},
      isValid: false,
    });

    await renderWithProviders(<InviteUserDialog open onClose={mockOnClose} />);

    const select = page.getByRole('combobox');
    await userEvent.click(select);

    const usaOption = page.getByRole('option', {name: 'USA'});
    await userEvent.click(usaOption);

    await vi.waitFor(() => {
      expect(mockHandleInputChange).toHaveBeenCalledWith('country', 'USA');
    });
  });

  it('renders multiple form fields in a block', async () => {
    const firstNameComponent: EmbeddedFlowComponent = {
      id: 'first_name_input',
      ref: 'firstName',
      type: EmbeddedFlowComponentType.TextInput,
      label: 'First Name',
      placeholder: 'Enter first name',
      required: true,
    };

    const lastNameComponent: EmbeddedFlowComponent = {
      id: 'last_name_input',
      ref: 'lastName',
      type: EmbeddedFlowComponentType.TextInput,
      label: 'Last Name',
      placeholder: 'Enter last name',
      required: false,
    };

    const emailComponent: EmbeddedFlowComponent = {
      id: 'email_input',
      ref: 'email',
      type: 'EMAIL_INPUT' as EmbeddedFlowComponent['type'],
      label: 'Email',
      placeholder: 'Enter email',
      required: true,
    };

    const submitAction: EmbeddedFlowComponent = {
      id: 'submit_action',
      type: EmbeddedFlowComponentType.Action,
      label: 'Submit',
      variant: 'PRIMARY',
      eventType: EmbeddedFlowEventType.Submit,
    };

    const blockComponent: EmbeddedFlowComponent = {
      id: 'block',
      type: EmbeddedFlowComponentType.Block,
      components: [firstNameComponent, lastNameComponent, emailComponent, submitAction],
    };

    Object.assign(mockInviteUserRenderProps, {
      components: [blockComponent],
      values: {firstName: '', lastName: '', email: ''},
      isValid: false,
    });

    await renderWithProviders(<InviteUserDialog open onClose={mockOnClose} />);

    await expect.element(page.getByPlaceholder('Enter first name')).toBeInTheDocument();
    await expect.element(page.getByPlaceholder('Enter last name')).toBeInTheDocument();
    await expect.element(page.getByPlaceholder('Enter email')).toBeInTheDocument();
    await expect.element(page.getByRole('button', {name: /submit/i})).toBeInTheDocument();
  });

  it('renders submit button with outlined variant when not PRIMARY', async () => {
    const emailComponent: EmbeddedFlowComponent = {
      id: 'email_input',
      ref: 'email',
      type: 'EMAIL_INPUT' as EmbeddedFlowComponent['type'],
      label: 'Email',
      placeholder: 'Enter email',
      required: true,
    };

    const submitAction: EmbeddedFlowComponent = {
      id: 'submit_action',
      type: EmbeddedFlowComponentType.Action,
      label: 'Submit',
      variant: 'SECONDARY', // Non-primary variant
      eventType: EmbeddedFlowEventType.Submit,
    };

    const blockComponent: EmbeddedFlowComponent = {
      id: 'block',
      type: EmbeddedFlowComponentType.Block,
      components: [emailComponent, submitAction],
    };

    Object.assign(mockInviteUserRenderProps, {
      components: [blockComponent],
      values: {email: 'test@example.com'},
      isValid: true,
    });

    await renderWithProviders(<InviteUserDialog open onClose={mockOnClose} />);

    const submitButton = page.getByRole('button', {name: /submit/i});
    await expect.element(submitButton).toHaveClass('MuiButton-outlined');
  });

  it('does not render component without ref in renderFormField', async () => {
    // Component without ref should not render a form field
    const componentWithoutRef: EmbeddedFlowComponent = {
      id: 'no_ref_input',
      type: EmbeddedFlowComponentType.TextInput,
      label: 'No Ref Field',
      placeholder: 'This should not render',
      // No ref property
    };

    const submitAction: EmbeddedFlowComponent = {
      id: 'submit_action',
      type: EmbeddedFlowComponentType.Action,
      label: 'Next',
      variant: 'PRIMARY',
      eventType: EmbeddedFlowEventType.Submit,
    };

    const blockComponent: EmbeddedFlowComponent = {
      id: 'block',
      type: EmbeddedFlowComponentType.Block,
      components: [componentWithoutRef, submitAction],
    };

    Object.assign(mockInviteUserRenderProps, {
      components: [blockComponent],
      values: {},
      isValid: false,
    });

    await renderWithProviders(<InviteUserDialog open onClose={mockOnClose} />);

    // Field without ref should not be rendered
    await expect.element(page.getByPlaceholder('This should not render')).not.toBeInTheDocument();
  });

  it('displays close button when error without components', async () => {
    Object.assign(mockInviteUserRenderProps, {
      error: new Error('Authentication failed'),
      components: [],
    });

    await renderWithProviders(<InviteUserDialog open onClose={mockOnClose} />);

    await expect.element(page.getByText('Authentication failed')).toBeInTheDocument();
    // Find the "Close" text button (not the X icon button)
    const closeButtons = page.getByRole('button', {name: /close/i}).all();
    const closeTextButton = closeButtons.find((btn) => btn.element().textContent === 'Close');
    expect(closeTextButton).toBeDefined();
  });

  it('handles form submission error gracefully', async () => {
    const emailComponent: EmbeddedFlowComponent = {
      id: 'email_input',
      ref: 'email',
      type: 'EMAIL_INPUT' as EmbeddedFlowComponent['type'],
      label: 'Email',
      placeholder: 'Enter email',
      required: true,
    };

    const submitAction: EmbeddedFlowComponent = {
      id: 'submit_action',
      type: EmbeddedFlowComponentType.Action,
      label: 'Next',
      variant: 'PRIMARY',
      eventType: EmbeddedFlowEventType.Submit,
    };

    const blockComponent: EmbeddedFlowComponent = {
      id: 'block',
      type: EmbeddedFlowComponentType.Block,
      components: [emailComponent, submitAction],
    };

    Object.assign(mockInviteUserRenderProps, {
      components: [blockComponent],
      values: {email: 'test@example.com'},
      isValid: true,
    });

    // Make handleSubmit reject
    mockHandleSubmit.mockRejectedValue(new Error('Network error'));

    await renderWithProviders(<InviteUserDialog open onClose={mockOnClose} />);

    const submitButton = page.getByRole('button', {name: /next/i});
    await userEvent.click(submitButton);

    // Should not throw, error is caught
    await vi.waitFor(() => {
      expect(mockHandleSubmit).toHaveBeenCalled();
    });
  });

  it('does not submit form when button is disabled', async () => {
    const emailComponent: EmbeddedFlowComponent = {
      id: 'email_input',
      ref: 'email',
      type: 'EMAIL_INPUT' as EmbeddedFlowComponent['type'],
      label: 'Email',
      placeholder: 'Enter email',
      required: true,
    };

    const submitAction: EmbeddedFlowComponent = {
      id: 'submit_action',
      type: EmbeddedFlowComponentType.Action,
      label: 'Next',
      variant: 'PRIMARY',
      eventType: EmbeddedFlowEventType.Submit,
    };

    const blockComponent: EmbeddedFlowComponent = {
      id: 'block',
      type: EmbeddedFlowComponentType.Block,
      components: [emailComponent, submitAction],
    };

    Object.assign(mockInviteUserRenderProps, {
      components: [blockComponent],
      values: {email: ''},
      isValid: false, // Form is invalid
    });

    await renderWithProviders(<InviteUserDialog open onClose={mockOnClose} />);

    // Try to submit the form programmatically (simulate form submission)
    const submitButtonEl = page.getByRole('button', {name: /next/i}).element() as HTMLElement;
    const form = submitButtonEl.closest('form');
    if (form) {
      const submitEvent = new Event('submit', {bubbles: true, cancelable: true});
      form.dispatchEvent(submitEvent);
    }

    // handleSubmit should not be called because form is invalid
    expect(mockHandleSubmit).not.toHaveBeenCalled();
  });

  it('renders loading state when components are empty but not loading', async () => {
    Object.assign(mockInviteUserRenderProps, {
      isLoading: false,
      components: [],
      isInviteGenerated: false,
      error: null,
    });

    await renderWithProviders(<InviteUserDialog open onClose={mockOnClose} />);

    // Should show loading spinner when waiting for components
    await expect.element(page.getByRole('progressbar')).toBeInTheDocument();
  });

  it('handles SELECT with hint text', async () => {
    const selectComponent = {
      id: 'tier_select',
      ref: 'tier',
      type: 'SELECT',
      label: 'Subscription Tier',
      placeholder: 'Select tier',
      required: false,
      options: [
        {value: 'free', label: 'Free'},
        {value: 'pro', label: 'Pro'},
      ],
      hint: 'Choose your subscription level',
    } as EmbeddedFlowComponent;

    const submitAction: EmbeddedFlowComponent = {
      id: 'submit_action',
      type: EmbeddedFlowComponentType.Action,
      label: 'Next',
      variant: 'PRIMARY',
      eventType: EmbeddedFlowEventType.Submit,
    };

    const blockComponent: EmbeddedFlowComponent = {
      id: 'block',
      type: EmbeddedFlowComponentType.Block,
      components: [selectComponent, submitAction],
    };

    Object.assign(mockInviteUserRenderProps, {
      components: [blockComponent],
      values: {tier: ''},
      isValid: false,
    });

    await renderWithProviders(<InviteUserDialog open onClose={mockOnClose} />);

    await expect.element(page.getByText('Choose your subscription level')).toBeInTheDocument();
  });

  it('uses fallback form validation when propsIsValid is undefined', async () => {
    const emailComponent: EmbeddedFlowComponent = {
      id: 'email_input',
      ref: 'email',
      type: 'EMAIL_INPUT' as EmbeddedFlowComponent['type'],
      label: 'Email',
      placeholder: 'Enter email',
      required: true,
    };

    const submitAction: EmbeddedFlowComponent = {
      id: 'submit_action',
      type: EmbeddedFlowComponentType.Action,
      label: 'Next',
      variant: 'PRIMARY',
      eventType: EmbeddedFlowEventType.Submit,
    };

    const blockComponent: EmbeddedFlowComponent = {
      id: 'block',
      type: EmbeddedFlowComponentType.Block,
      components: [emailComponent, submitAction],
    };

    // Set isValid to undefined to trigger fallback
    Object.assign(mockInviteUserRenderProps, {
      components: [blockComponent],
      values: {email: ''},
      isValid: undefined,
    });

    await renderWithProviders(<InviteUserDialog open onClose={mockOnClose} />);

    // Button should be disabled based on form validation
    const submitButton = page.getByRole('button', {name: /next/i});
    await expect.element(submitButton).toBeDisabled();
  });

  it('returns null for unsupported component types in renderFormField', async () => {
    // Component with unsupported type should return null
    const unsupportedComponent: EmbeddedFlowComponent = {
      id: 'unsupported_component',
      ref: 'unsupported',
      type: 'UNSUPPORTED_TYPE' as EmbeddedFlowComponent['type'],
      label: 'Unsupported',
    };

    const submitAction: EmbeddedFlowComponent = {
      id: 'submit_action',
      type: EmbeddedFlowComponentType.Action,
      label: 'Next',
      variant: 'PRIMARY',
      eventType: EmbeddedFlowEventType.Submit,
    };

    const blockComponent: EmbeddedFlowComponent = {
      id: 'block',
      type: EmbeddedFlowComponentType.Block,
      components: [unsupportedComponent, submitAction],
    };

    Object.assign(mockInviteUserRenderProps, {
      components: [blockComponent],
      values: {},
      isValid: false,
    });

    await renderWithProviders(<InviteUserDialog open onClose={mockOnClose} />);

    // Unsupported component should not render any input
    await expect.element(page.getByLabelText('Unsupported')).not.toBeInTheDocument();
  });

  it('handles getOptionValue with non-string value in object option', async () => {
    const selectComponent: EmbeddedFlowComponent = {
      id: 'complex_select',
      ref: 'complexOption',
      type: 'SELECT' as EmbeddedFlowComponent['type'],
      label: 'Complex Option',
      placeholder: 'Select option',
      required: true,
      options: [
        {value: {id: 1, name: 'Option 1'} as unknown as string, label: 'Complex Option 1'},
        {value: {id: 2, name: 'Option 2'} as unknown as string, label: 'Complex Option 2'},
      ],
    };

    const submitAction: EmbeddedFlowComponent = {
      id: 'submit_action',
      type: EmbeddedFlowComponentType.Action,
      label: 'Next',
      variant: 'PRIMARY',
      eventType: EmbeddedFlowEventType.Submit,
    };

    const blockComponent: EmbeddedFlowComponent = {
      id: 'block',
      type: EmbeddedFlowComponentType.Block,
      components: [selectComponent, submitAction],
    };

    Object.assign(mockInviteUserRenderProps, {
      components: [blockComponent],
      values: {complexOption: ''},
      isValid: false,
    });

    await renderWithProviders(<InviteUserDialog open onClose={mockOnClose} />);

    const select = page.getByRole('combobox');
    await userEvent.click(select);

    // The option should render with the serialized value
    const option = page.getByRole('option', {name: 'Complex Option 1'});
    await expect.element(option).toBeInTheDocument();
  });

  it('handles getOptionLabel with non-string label in object option', async () => {
    const selectComponent: EmbeddedFlowComponent = {
      id: 'label_select',
      ref: 'labelOption',
      type: 'SELECT' as EmbeddedFlowComponent['type'],
      label: 'Label Option',
      placeholder: 'Select option',
      required: true,
      options: [
        {value: 'opt1', label: {text: 'Label 1'} as unknown as string}, // Non-string label
      ],
    };

    const submitAction: EmbeddedFlowComponent = {
      id: 'submit_action',
      type: EmbeddedFlowComponentType.Action,
      label: 'Next',
      variant: 'PRIMARY',
      eventType: EmbeddedFlowEventType.Submit,
    };

    const blockComponent: EmbeddedFlowComponent = {
      id: 'block',
      type: EmbeddedFlowComponentType.Block,
      components: [selectComponent, submitAction],
    };

    Object.assign(mockInviteUserRenderProps, {
      components: [blockComponent],
      values: {labelOption: ''},
      isValid: false,
    });

    await renderWithProviders(<InviteUserDialog open onClose={mockOnClose} />);

    // Should render without crashing
    await expect.element(page.getByRole('combobox')).toBeInTheDocument();
  });

  it('handles option without label or value properties', async () => {
    const selectComponent: EmbeddedFlowComponent = {
      id: 'bare_select',
      ref: 'bareOption',
      type: 'SELECT' as EmbeddedFlowComponent['type'],
      label: 'Bare Option',
      placeholder: 'Select option',
      required: true,
      options: [
        {key: 'option1'} as unknown as {value: string; label: string}, // Object without value or label
      ],
    };

    const submitAction: EmbeddedFlowComponent = {
      id: 'submit_action',
      type: EmbeddedFlowComponentType.Action,
      label: 'Next',
      variant: 'PRIMARY',
      eventType: EmbeddedFlowEventType.Submit,
    };

    const blockComponent: EmbeddedFlowComponent = {
      id: 'block',
      type: EmbeddedFlowComponentType.Block,
      components: [selectComponent, submitAction],
    };

    Object.assign(mockInviteUserRenderProps, {
      components: [blockComponent],
      values: {bareOption: ''},
      isValid: false,
    });

    await renderWithProviders(<InviteUserDialog open onClose={mockOnClose} />);

    const select = page.getByRole('combobox');
    await userEvent.click(select);

    // Should render the serialized option
    const option = page.getByRole('option', {name: /option1/i});
    await expect.element(option).toBeInTheDocument();
  });

  it('handles copyInviteLink returning undefined', async () => {
    const inviteLink = 'https://example.com/invite?token=abc123';

    // Set copyInviteLink to undefined
    Object.assign(mockInviteUserRenderProps, {
      isInviteGenerated: true,
      inviteLink,
      inviteLinkCopied: false,
      copyInviteLink: undefined,
    });

    await renderWithProviders(<InviteUserDialog open onClose={mockOnClose} />);

    const copyButton = page.getByRole('button', {name: /copy/i});
    await userEvent.click(copyButton);

    // Should not crash when copyInviteLink is undefined
    await expect.element(copyButton).toBeInTheDocument();
  });

  it('renders SELECT with selected value showing label', async () => {
    const selectComponent: EmbeddedFlowComponent = {
      id: 'preselected_select',
      ref: 'preselectedOption',
      type: 'SELECT' as EmbeddedFlowComponent['type'],
      label: 'Preselected',
      placeholder: 'Select option',
      required: true,
      options: [
        {value: 'admin', label: 'Administrator'},
        {value: 'user', label: 'Standard User'},
      ],
    };

    const submitAction: EmbeddedFlowComponent = {
      id: 'submit_action',
      type: EmbeddedFlowComponentType.Action,
      label: 'Next',
      variant: 'PRIMARY',
      eventType: EmbeddedFlowEventType.Submit,
    };

    const blockComponent: EmbeddedFlowComponent = {
      id: 'block',
      type: EmbeddedFlowComponentType.Block,
      components: [selectComponent, submitAction],
    };

    Object.assign(mockInviteUserRenderProps, {
      components: [blockComponent],
      values: {preselectedOption: 'admin'},
      isValid: true,
    });

    await renderWithProviders(<InviteUserDialog open onClose={mockOnClose} />);

    // Should show the label for the selected value
    await expect.element(page.getByText('Administrator')).toBeInTheDocument();
  });

  it('renders SELECT with unknown selected value', async () => {
    const selectComponent: EmbeddedFlowComponent = {
      id: 'unknown_select',
      ref: 'unknownOption',
      type: 'SELECT' as EmbeddedFlowComponent['type'],
      label: 'Unknown',
      placeholder: 'Select option',
      required: true,
      options: [{value: 'admin', label: 'Administrator'}],
    };

    const submitAction: EmbeddedFlowComponent = {
      id: 'submit_action',
      type: EmbeddedFlowComponentType.Action,
      label: 'Next',
      variant: 'PRIMARY',
      eventType: EmbeddedFlowEventType.Submit,
    };

    const blockComponent: EmbeddedFlowComponent = {
      id: 'block',
      type: EmbeddedFlowComponentType.Block,
      components: [selectComponent, submitAction],
    };

    Object.assign(mockInviteUserRenderProps, {
      components: [blockComponent],
      values: {unknownOption: 'unknown_value'},
      isValid: true,
    });

    await renderWithProviders(<InviteUserDialog open onClose={mockOnClose} />);

    // Should show the raw value when option not found
    await expect.element(page.getByText('unknown_value')).toBeInTheDocument();
  });

  it('returns null for non-Block component type at top level', async () => {
    // Test when a top-level component is not a Block type
    const nonBlockComponent: EmbeddedFlowComponent = {
      id: 'non_block',
      ref: 'nonBlock',
      type: 'TEXT_INPUT' as EmbeddedFlowComponent['type'],
      label: 'Non Block',
    };

    Object.assign(mockInviteUserRenderProps, {
      components: [nonBlockComponent],
      values: {},
      isValid: false,
    });

    await renderWithProviders(<InviteUserDialog open onClose={mockOnClose} />);

    // Non-Block component at top level should return null
    await expect.element(page.getByLabelText('Non Block')).not.toBeInTheDocument();
  });

  it('logs error when onError callback is triggered', async () => {
    // Enable error simulation
    simulateInviteUserError = true;

    await renderWithProviders(<InviteUserDialog open onClose={mockOnClose} />);

    // Wait for the onError callback to be triggered
    await vi.waitFor(() => {
      expect(mockLoggerError).toHaveBeenCalledWith('User onboarding error', {error: mockInviteUserError});
    });
  });
});
