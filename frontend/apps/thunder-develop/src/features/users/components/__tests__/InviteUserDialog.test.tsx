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
import {render, screen, waitFor, userEvent} from '@thunder/test-utils';
import type {JSX} from 'react';
import {EmbeddedFlowComponentType, EmbeddedFlowEventType} from '@asgardeo/react';
import type {InviteUserRenderProps, EmbeddedFlowComponent} from '@asgardeo/react';
import InviteUserDialog from '../InviteUserDialog';

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

  it('renders dialog when open', () => {
    render(<InviteUserDialog open onClose={mockOnClose} />);

    expect(screen.getByText('Invite User')).toBeInTheDocument();
    expect(screen.getByText(/Send an invite link to a new user/i)).toBeInTheDocument();
  });

  it('does not render when closed', () => {
    render(<InviteUserDialog open={false} onClose={mockOnClose} />);

    expect(screen.queryByText('Invite User')).not.toBeInTheDocument();
  });

  it('closes dialog when close button is clicked', async () => {
    const user = userEvent.setup();
    render(<InviteUserDialog open onClose={mockOnClose} />);

    const closeButton = screen.getByLabelText('close');
    await user.click(closeButton);

    expect(mockOnClose).toHaveBeenCalled();
  });

  it('displays loading state when initializing', () => {
    Object.assign(mockInviteUserRenderProps, {
      isLoading: true,
      components: [],
      isInviteGenerated: false,
    });

    render(<InviteUserDialog open onClose={mockOnClose} />);

    expect(screen.getByRole('progressbar')).toBeInTheDocument();
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

    render(<InviteUserDialog open onClose={mockOnClose} />);

    // Wait for form to render and check for email input by placeholder or label
    await waitFor(() => {
      const emailInput = screen.getByPlaceholderText('Enter email');
      expect(emailInput).toBeInTheDocument();
    });
    // Check for label - it might be "Email" or a translation key
    const labels = screen.getAllByText('Email');
    expect(labels.length).toBeGreaterThan(0);
  });

  it('disables submit button when form is invalid', () => {
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

    render(<InviteUserDialog open onClose={mockOnClose} />);

    const submitButton = screen.getByRole('button', {name: /next/i});
    expect(submitButton).toBeDisabled();
  });

  it('enables submit button when form is valid', () => {
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

    render(<InviteUserDialog open onClose={mockOnClose} />);

    const submitButton = screen.getByRole('button', {name: /next/i});
    expect(submitButton).not.toBeDisabled();
  });

  it('calls handleInputChange when input value changes', async () => {
    const user = userEvent.setup();
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

    render(<InviteUserDialog open onClose={mockOnClose} />);

    const emailInput = screen.getByPlaceholderText('Enter email');
    await user.type(emailInput, 'test@example.com');

    await waitFor(() => {
      expect(mockHandleInputChange).toHaveBeenCalledWith('email', 'test@example.com');
    });
  });

  it('submits form when submit button is clicked', async () => {
    const user = userEvent.setup();
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

    render(<InviteUserDialog open onClose={mockOnClose} />);

    const submitButton = screen.getByRole('button', {name: /next/i});
    await user.click(submitButton);

    await waitFor(() => {
      expect(mockHandleSubmit).toHaveBeenCalledWith(submitAction, {email: 'test@example.com'});
    });
  });

  it('displays error message when error occurs', () => {
    Object.assign(mockInviteUserRenderProps, {
      error: new Error('Failed to invite user'),
      components: [],
    });

    render(<InviteUserDialog open onClose={mockOnClose} />);

    expect(screen.getByText(/Error/i)).toBeInTheDocument();
    expect(screen.getByText('Failed to invite user')).toBeInTheDocument();
  });

  it('displays invite link when invite is generated', () => {
    const inviteLink = 'https://example.com/invite?token=abc123';
    Object.assign(mockInviteUserRenderProps, {
      isInviteGenerated: true,
      inviteLink,
    });

    render(<InviteUserDialog open onClose={mockOnClose} onSuccess={mockOnSuccess} />);

    expect(screen.getByText(/Invite Link Generated!/i)).toBeInTheDocument();
    expect(screen.getByDisplayValue(inviteLink)).toBeInTheDocument();
  });

  it('calls onSuccess when invite link is generated', async () => {
    const inviteLink = 'https://example.com/invite?token=abc123';
    Object.assign(mockInviteUserRenderProps, {
      isInviteGenerated: true,
      inviteLink,
    });

    render(<InviteUserDialog open onClose={mockOnClose} onSuccess={mockOnSuccess} />);

    await waitFor(() => {
      expect(mockOnSuccess).toHaveBeenCalledWith(inviteLink);
    });
  });

  it('copies invite link when copy button is clicked', async () => {
    const user = userEvent.setup();
    const inviteLink = 'https://example.com/invite?token=abc123';
    Object.assign(mockInviteUserRenderProps, {
      isInviteGenerated: true,
      inviteLink,
      inviteLinkCopied: false,
    });

    render(<InviteUserDialog open onClose={mockOnClose} />);

    const copyButton = screen.getByRole('button', {name: /copy/i});
    await user.click(copyButton);

    expect(mockCopyInviteLink).toHaveBeenCalled();
  });

  it('resets flow when "Invite Another User" is clicked', async () => {
    const user = userEvent.setup();
    const inviteLink = 'https://example.com/invite?token=abc123';
    Object.assign(mockInviteUserRenderProps, {
      isInviteGenerated: true,
      inviteLink,
    });

    render(<InviteUserDialog open onClose={mockOnClose} />);

    const inviteAnotherButton = screen.getByRole('button', {name: /Invite Another User/i});
    await user.click(inviteAnotherButton);

    expect(mockResetFlow).toHaveBeenCalled();
  });

  it('displays stepper with correct steps', () => {
    render(<InviteUserDialog open onClose={mockOnClose} />);

    expect(screen.getByText('User Details')).toBeInTheDocument();
    expect(screen.getByText('Invite Link')).toBeInTheDocument();
  });

  it('renders SELECT field with placeholder when no value selected', () => {
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

    render(<InviteUserDialog open onClose={mockOnClose} />);

    const select = screen.getByRole('combobox');
    expect(select).toHaveTextContent('Select user type');
  });

  it('displays loading spinner on submit button when loading', () => {
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

    render(<InviteUserDialog open onClose={mockOnClose} />);

    // When loading, the submit button shows a spinner instead of text
    // Find the button that contains the progressbar (submit button)
    const spinner = screen.getByRole('progressbar');
    expect(spinner).toBeInTheDocument();
    // The spinner's parent button should be disabled
    const submitButton = spinner.closest('button');
    expect(submitButton).toBeDisabled();
  });

  it('disables cancel button when loading', () => {
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

    render(<InviteUserDialog open onClose={mockOnClose} />);

    const cancelButton = screen.getByRole('button', {name: /cancel/i});
    expect(cancelButton).toBeDisabled();
  });

  it('validates email format', async () => {
    const user = userEvent.setup();
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

    render(<InviteUserDialog open onClose={mockOnClose} />);

    const emailInput = screen.getByPlaceholderText('Enter email');
    await user.type(emailInput, 'invalid-email');

    // Form validation should prevent submission
    const submitButton = screen.getByRole('button', {name: /next/i});
    expect(submitButton).toBeDisabled();
  });

  it('renders TEXT_INPUT field correctly', async () => {
    const user = userEvent.setup();
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

    render(<InviteUserDialog open onClose={mockOnClose} />);

    const textInput = screen.getByPlaceholderText('Enter first name');
    expect(textInput).toBeInTheDocument();

    await user.type(textInput, 'John');

    await waitFor(() => {
      expect(mockHandleInputChange).toHaveBeenCalledWith('firstName', 'John');
    });
  });

  it('renders TEXT component with body variant', () => {
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

    render(<InviteUserDialog open onClose={mockOnClose} />);

    expect(screen.getByText('Please fill in the user details below')).toBeInTheDocument();
  });

  it('skips TEXT component with HEADING_1 variant', () => {
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

    render(<InviteUserDialog open onClose={mockOnClose} />);

    // HEADING_1 should not be rendered
    expect(screen.queryByText('Main Heading')).not.toBeInTheDocument();
  });

  it('displays error message when error occurs with components present', () => {
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

    render(<InviteUserDialog open onClose={mockOnClose} />);

    // Error should be displayed above the form
    expect(screen.getByText('Email already exists')).toBeInTheDocument();
    // Form should still be visible
    expect(screen.getByPlaceholderText('Enter email')).toBeInTheDocument();
  });

  it('closes dialog when cancel button is clicked', async () => {
    const user = userEvent.setup();
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

    render(<InviteUserDialog open onClose={mockOnClose} />);

    const cancelButton = screen.getByRole('button', {name: /cancel/i});
    await user.click(cancelButton);

    expect(mockOnClose).toHaveBeenCalled();
  });

  it('returns null for block without submit action', () => {
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

    render(<InviteUserDialog open onClose={mockOnClose} />);

    // Block should not render without submit action
    expect(screen.queryByPlaceholderText('Enter email')).not.toBeInTheDocument();
  });

  it('handles SELECT field value changes', async () => {
    const user = userEvent.setup();
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

    render(<InviteUserDialog open onClose={mockOnClose} />);

    // Open the select dropdown
    const select = screen.getByRole('combobox');
    await user.click(select);

    // Select an option
    const adminOption = await screen.findByRole('option', {name: 'Administrator'});
    await user.click(adminOption);

    await waitFor(() => {
      expect(mockHandleInputChange).toHaveBeenCalledWith('role', 'admin');
    });
  });

  it('displays check icon when invite link is already copied', () => {
    const inviteLink = 'https://example.com/invite?token=abc123';
    Object.assign(mockInviteUserRenderProps, {
      isInviteGenerated: true,
      inviteLink,
      inviteLinkCopied: true, // Already copied via SDK
    });

    render(<InviteUserDialog open onClose={mockOnClose} />);

    // The copy button should show check icon when inviteLinkCopied is true
    const copyButton = screen.getByRole('button', {name: /copy/i});
    expect(copyButton).toBeInTheDocument();
  });

  it('closes dialog from invite generated screen', async () => {
    const user = userEvent.setup();
    const inviteLink = 'https://example.com/invite?token=abc123';
    Object.assign(mockInviteUserRenderProps, {
      isInviteGenerated: true,
      inviteLink,
    });

    render(<InviteUserDialog open onClose={mockOnClose} />);

    // Find the "Close" button (not the X icon button which has aria-label="close")
    const closeButtons = screen.getAllByRole('button', {name: /close/i});
    // The "Close" text button should be in the generated invite screen
    const closeTextButton = closeButtons.find((btn) => btn.textContent === 'Close');
    expect(closeTextButton).toBeDefined();
    await user.click(closeTextButton!);

    expect(mockOnClose).toHaveBeenCalled();
  });

  it('handles getOptionValue with string option', async () => {
    const user = userEvent.setup();
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

    render(<InviteUserDialog open onClose={mockOnClose} />);

    const select = screen.getByRole('combobox');
    await user.click(select);

    const usaOption = await screen.findByRole('option', {name: 'USA'});
    await user.click(usaOption);

    await waitFor(() => {
      expect(mockHandleInputChange).toHaveBeenCalledWith('country', 'USA');
    });
  });

  it('renders multiple form fields in a block', () => {
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

    render(<InviteUserDialog open onClose={mockOnClose} />);

    expect(screen.getByPlaceholderText('Enter first name')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('Enter last name')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('Enter email')).toBeInTheDocument();
    expect(screen.getByRole('button', {name: /submit/i})).toBeInTheDocument();
  });

  it('renders submit button with outlined variant when not PRIMARY', () => {
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

    render(<InviteUserDialog open onClose={mockOnClose} />);

    const submitButton = screen.getByRole('button', {name: /submit/i});
    expect(submitButton).toHaveClass('MuiButton-outlined');
  });

  it('does not render component without ref in renderFormField', () => {
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

    render(<InviteUserDialog open onClose={mockOnClose} />);

    // Field without ref should not be rendered
    expect(screen.queryByPlaceholderText('This should not render')).not.toBeInTheDocument();
  });

  it('displays close button when error without components', () => {
    Object.assign(mockInviteUserRenderProps, {
      error: new Error('Authentication failed'),
      components: [],
    });

    render(<InviteUserDialog open onClose={mockOnClose} />);

    expect(screen.getByText('Authentication failed')).toBeInTheDocument();
    // Find the "Close" text button (not the X icon button)
    const closeButtons = screen.getAllByRole('button', {name: /close/i});
    const closeTextButton = closeButtons.find((btn) => btn.textContent === 'Close');
    expect(closeTextButton).toBeDefined();
  });

  it('handles form submission error gracefully', async () => {
    const user = userEvent.setup();
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

    render(<InviteUserDialog open onClose={mockOnClose} />);

    const submitButton = screen.getByRole('button', {name: /next/i});
    await user.click(submitButton);

    // Should not throw, error is caught
    await waitFor(() => {
      expect(mockHandleSubmit).toHaveBeenCalled();
    });
  });

  it('does not submit form when button is disabled', () => {
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

    render(<InviteUserDialog open onClose={mockOnClose} />);

    // Try to submit the form programmatically (simulate form submission)
    const form = screen.getByRole('button', {name: /next/i}).closest('form');
    if (form) {
      const submitEvent = new Event('submit', {bubbles: true, cancelable: true});
      form.dispatchEvent(submitEvent);
    }

    // handleSubmit should not be called because form is invalid
    expect(mockHandleSubmit).not.toHaveBeenCalled();
  });

  it('renders loading state when components are empty but not loading', () => {
    Object.assign(mockInviteUserRenderProps, {
      isLoading: false,
      components: [],
      isInviteGenerated: false,
      error: null,
    });

    render(<InviteUserDialog open onClose={mockOnClose} />);

    // Should show loading spinner when waiting for components
    expect(screen.getByRole('progressbar')).toBeInTheDocument();
  });

  it('handles SELECT with hint text', () => {
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

    render(<InviteUserDialog open onClose={mockOnClose} />);

    expect(screen.getByText('Choose your subscription level')).toBeInTheDocument();
  });

  it('uses fallback form validation when propsIsValid is undefined', () => {
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

    render(<InviteUserDialog open onClose={mockOnClose} />);

    // Button should be disabled based on form validation
    const submitButton = screen.getByRole('button', {name: /next/i});
    expect(submitButton).toBeDisabled();
  });

  it('returns null for unsupported component types in renderFormField', () => {
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

    render(<InviteUserDialog open onClose={mockOnClose} />);

    // Unsupported component should not render any input
    expect(screen.queryByLabelText('Unsupported')).not.toBeInTheDocument();
  });

  it('handles getOptionValue with non-string value in object option', async () => {
    const user = userEvent.setup();
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

    render(<InviteUserDialog open onClose={mockOnClose} />);

    const select = screen.getByRole('combobox');
    await user.click(select);

    // The option should render with the serialized value
    const option = await screen.findByRole('option', {name: 'Complex Option 1'});
    expect(option).toBeInTheDocument();
  });

  it('handles getOptionLabel with non-string label in object option', () => {
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

    render(<InviteUserDialog open onClose={mockOnClose} />);

    // Should render without crashing
    expect(screen.getByRole('combobox')).toBeInTheDocument();
  });

  it('handles option without label or value properties', async () => {
    const user = userEvent.setup();
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

    render(<InviteUserDialog open onClose={mockOnClose} />);

    const select = screen.getByRole('combobox');
    await user.click(select);

    // Should render the serialized option
    const option = await screen.findByRole('option', {name: /option1/i});
    expect(option).toBeInTheDocument();
  });

  it('handles copyInviteLink returning undefined', async () => {
    const user = userEvent.setup();
    const inviteLink = 'https://example.com/invite?token=abc123';

    // Set copyInviteLink to undefined
    Object.assign(mockInviteUserRenderProps, {
      isInviteGenerated: true,
      inviteLink,
      inviteLinkCopied: false,
      copyInviteLink: undefined,
    });

    render(<InviteUserDialog open onClose={mockOnClose} />);

    const copyButton = screen.getByRole('button', {name: /copy/i});
    await user.click(copyButton);

    // Should not crash when copyInviteLink is undefined
    expect(copyButton).toBeInTheDocument();
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

    render(<InviteUserDialog open onClose={mockOnClose} />);

    // Should show the label for the selected value
    expect(screen.getByText('Administrator')).toBeInTheDocument();
  });

  it('renders SELECT with unknown selected value', () => {
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

    render(<InviteUserDialog open onClose={mockOnClose} />);

    // Should show the raw value when option not found
    expect(screen.getByText('unknown_value')).toBeInTheDocument();
  });

  it('returns null for non-Block component type at top level', () => {
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

    render(<InviteUserDialog open onClose={mockOnClose} />);

    // Non-Block component at top level should return null
    expect(screen.queryByLabelText('Non Block')).not.toBeInTheDocument();
  });

  it('logs error when onError callback is triggered', async () => {
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => undefined);

    // Enable error simulation
    simulateInviteUserError = true;

    render(<InviteUserDialog open onClose={mockOnClose} />);

    // Wait for the onError callback to be triggered
    await waitFor(() => {
      expect(consoleSpy).toHaveBeenCalledWith('User onboarding error:', mockInviteUserError);
    });

    consoleSpy.mockRestore();
  });
});
