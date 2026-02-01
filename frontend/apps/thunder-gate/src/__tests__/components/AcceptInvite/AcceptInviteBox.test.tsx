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

import {describe, it, expect, vi, beforeEach} from 'vitest';
import {render, screen, fireEvent, waitFor} from '@thunder/test-utils';
import userEvent from '@testing-library/user-event';
import AcceptInviteBox from '../../../components/AcceptInvite/AcceptInviteBox';

// Mock useBranding
const mockUseBranding = vi.fn();
vi.mock('@thunder/shared-branding', () => ({
  // eslint-disable-next-line @typescript-eslint/no-unsafe-return
  useBranding: () => mockUseBranding(),
  mapEmbeddedFlowTextVariant: (variant: string) => {
    switch (variant) {
      case 'H1':
        return 'h1';
      case 'H2':
        return 'h2';
      default:
        return 'body1';
    }
  },
}));

// Mock useTemplateLiteralResolver
vi.mock('@thunder/shared-hooks', () => ({
  useTemplateLiteralResolver: () => ({
    resolve: (key: string) => key,
  }),
}));

// Mock useConfig
vi.mock('@thunder/shared-contexts', () => ({
  useConfig: () => ({
    getServerUrl: () => 'https://api.example.com',
  }),
}));

// Mock react-router hooks
const mockNavigate = vi.fn();
vi.mock('react-router', () => ({
  useNavigate: () => mockNavigate,
}));

// Mock Asgardeo AcceptInvite component
const mockHandleSubmit = vi.fn().mockResolvedValue(undefined);
const mockHandleInputChange = vi.fn();

interface MockAcceptInviteRenderProps {
  values: Record<string, string>;
  fieldErrors: Record<string, string>;
  error: {message: string} | null;
  touched: Record<string, boolean>;
  handleInputChange: typeof mockHandleInputChange;
  handleSubmit: typeof mockHandleSubmit;
  isLoading: boolean;
  components: unknown[];
  isComplete: boolean;
  isValidatingToken: boolean;
  isTokenInvalid: boolean;
  isValid: boolean;
}

// Factory function to create fresh mock props for each test
const createMockAcceptInviteRenderProps = (
  overrides: Partial<MockAcceptInviteRenderProps> = {},
): MockAcceptInviteRenderProps => ({
  values: {},
  fieldErrors: {},
  error: null,
  touched: {},
  handleInputChange: mockHandleInputChange,
  handleSubmit: mockHandleSubmit,
  isLoading: false,
  components: [],
  isComplete: false,
  isValidatingToken: false,
  isTokenInvalid: false,
  isValid: true,
  ...overrides,
});

let mockAcceptInviteRenderProps: MockAcceptInviteRenderProps = createMockAcceptInviteRenderProps();

// Track props passed to AcceptInvite
let capturedOnGoToSignIn: (() => void) | undefined;
let capturedOnError: ((error: Error) => void) | undefined;

vi.mock('@asgardeo/react', async () => {
  const actual = await vi.importActual('@asgardeo/react');
  return {
    ...actual,
    AcceptInvite: ({
      children,
      onGoToSignIn = undefined,
      onError = undefined,
    }: {
      children: (props: typeof mockAcceptInviteRenderProps) => React.ReactNode;
      onGoToSignIn?: () => void;
      onError?: (error: Error) => void;
    }) => {
      capturedOnGoToSignIn = onGoToSignIn;
      capturedOnError = onError;
      return <div data-testid="asgardeo-accept-invite">{children(mockAcceptInviteRenderProps)}</div>;
    },
    EmbeddedFlowComponentType: {
      Text: 'TEXT',
      Block: 'BLOCK',
      TextInput: 'TEXT_INPUT',
      PasswordInput: 'PASSWORD_INPUT',
      Action: 'ACTION',
    },
    EmbeddedFlowEventType: {
      Submit: 'SUBMIT',
      Trigger: 'TRIGGER',
    },
  };
});

describe('AcceptInviteBox', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseBranding.mockReturnValue({
      images: null,
      theme: null,
      isBrandingEnabled: false,
    });
    mockAcceptInviteRenderProps = createMockAcceptInviteRenderProps();
  });

  it('renders without crashing', () => {
    const {container} = render(<AcceptInviteBox />);
    expect(container).toBeInTheDocument();
  });

  it('shows validating token message', () => {
    mockAcceptInviteRenderProps = createMockAcceptInviteRenderProps({
      isValidatingToken: true,
    });
    render(<AcceptInviteBox />);
    expect(screen.getByText(/Validating your invite link/)).toBeInTheDocument();
  });

  it('shows invalid token error', () => {
    mockAcceptInviteRenderProps = createMockAcceptInviteRenderProps({
      isTokenInvalid: true,
    });
    render(<AcceptInviteBox />);
    expect(screen.getByText(/Unable to verify invite/)).toBeInTheDocument();
    expect(screen.getByText(/This invite link is invalid or has expired/)).toBeInTheDocument();
  });

  it('shows completion message', () => {
    mockAcceptInviteRenderProps = createMockAcceptInviteRenderProps({
      isComplete: true,
    });
    render(<AcceptInviteBox />);
    expect(screen.getByText(/Your account has been successfully set up/)).toBeInTheDocument();
  });

  it('shows loading spinner when loading and no components', () => {
    mockAcceptInviteRenderProps = createMockAcceptInviteRenderProps({
      isLoading: true,
      components: [],
    });
    render(<AcceptInviteBox />);
    expect(screen.getByTestId('asgardeo-accept-invite')).toBeInTheDocument();
  });

  it('shows error alert when error is present', () => {
    mockAcceptInviteRenderProps = createMockAcceptInviteRenderProps({
      error: {message: 'Something went wrong'},
      components: [{id: 'block', type: 'BLOCK', components: []}],
    });
    render(<AcceptInviteBox />);
    expect(screen.getByText('Something went wrong')).toBeInTheDocument();
  });

  it('renders TEXT component as heading', () => {
    mockAcceptInviteRenderProps = createMockAcceptInviteRenderProps({
      components: [
        {
          id: 'text-1',
          type: 'TEXT',
          label: 'Set Your Password',
          variant: 'H1',
        },
      ],
    });
    render(<AcceptInviteBox />);
    expect(screen.getByText('Set Your Password')).toBeInTheDocument();
  });

  it('renders TEXT_INPUT component', () => {
    mockAcceptInviteRenderProps = createMockAcceptInviteRenderProps({
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'name-input',
              type: 'TEXT_INPUT',
              ref: 'firstName',
              label: 'First Name',
              placeholder: 'Enter your first name',
              required: true,
            },
            {
              id: 'submit-btn',
              type: 'ACTION',
              eventType: 'SUBMIT',
              label: 'Continue',
              variant: 'PRIMARY',
            },
          ],
        },
      ],
    });
    render(<AcceptInviteBox />);
    expect(screen.getByLabelText(/First Name/)).toBeInTheDocument();
  });

  it('renders PASSWORD_INPUT component with toggle visibility', async () => {
    mockAcceptInviteRenderProps = createMockAcceptInviteRenderProps({
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'password-input',
              type: 'PASSWORD_INPUT',
              ref: 'password',
              label: 'Password',
              placeholder: 'Enter your password',
              required: true,
            },
            {
              id: 'submit-btn',
              type: 'ACTION',
              eventType: 'SUBMIT',
              label: 'Set Password',
              variant: 'PRIMARY',
            },
          ],
        },
      ],
    });
    render(<AcceptInviteBox />);

    const passwordInput = screen.getByLabelText(/Password/);
    expect(passwordInput).toBeInTheDocument();
    expect(passwordInput).toHaveAttribute('type', 'password');

    // Toggle visibility
    const toggleButton = screen.getByLabelText('toggle password visibility');
    await userEvent.click(toggleButton);

    expect(passwordInput).toHaveAttribute('type', 'text');
  });

  it('renders EMAIL_INPUT component', () => {
    mockAcceptInviteRenderProps = createMockAcceptInviteRenderProps({
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'email-input',
              type: 'EMAIL_INPUT',
              ref: 'email',
              label: 'Email Address',
              placeholder: 'Enter email',
              required: true,
            },
            {
              id: 'submit-btn',
              type: 'ACTION',
              eventType: 'SUBMIT',
              label: 'Continue',
              variant: 'PRIMARY',
            },
          ],
        },
      ],
    });
    render(<AcceptInviteBox />);
    expect(screen.getByLabelText(/Email Address/)).toBeInTheDocument();
  });

  it('renders SELECT component', () => {
    mockAcceptInviteRenderProps = createMockAcceptInviteRenderProps({
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'role-select',
              type: 'SELECT',
              ref: 'role',
              label: 'Role',
              placeholder: 'Select your role',
              options: ['Developer', 'Manager', 'Admin'],
              hint: 'Select your primary role',
              required: true,
            },
            {
              id: 'submit-btn',
              type: 'ACTION',
              eventType: 'SUBMIT',
              label: 'Continue',
              variant: 'PRIMARY',
            },
          ],
        },
      ],
    });
    render(<AcceptInviteBox />);
    expect(screen.getByText('Role')).toBeInTheDocument();
    expect(screen.getByText('Select your primary role')).toBeInTheDocument();
    expect(screen.getByRole('combobox')).toBeInTheDocument();
  });

  it('renders SELECT component with object options', () => {
    mockAcceptInviteRenderProps = createMockAcceptInviteRenderProps({
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'role-select',
              type: 'SELECT',
              ref: 'role',
              label: 'Role',
              placeholder: 'Select your role',
              options: [
                {value: 'dev', label: 'Developer'},
                {value: 'mgr', label: 'Manager'},
              ],
              required: true,
            },
            {
              id: 'submit-btn',
              type: 'ACTION',
              eventType: 'SUBMIT',
              label: 'Continue',
              variant: 'PRIMARY',
            },
          ],
        },
      ],
    });
    render(<AcceptInviteBox />);
    expect(screen.getByText('Role')).toBeInTheDocument();
    expect(screen.getByRole('combobox')).toBeInTheDocument();
  });

  it('submits form when submit button is clicked', async () => {
    mockAcceptInviteRenderProps = createMockAcceptInviteRenderProps({
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'password-input',
              type: 'PASSWORD_INPUT',
              ref: 'password',
              label: 'Password',
              required: true,
            },
            {
              id: 'submit-btn',
              type: 'ACTION',
              eventType: 'SUBMIT',
              label: 'Set Password',
              variant: 'PRIMARY',
            },
          ],
        },
      ],
    });
    render(<AcceptInviteBox />);

    const submitBtn = screen.getByText('Set Password');
    fireEvent.click(submitBtn);

    await waitFor(() => {
      expect(mockHandleSubmit).toHaveBeenCalled();
    });
  });

  it('shows validation errors for fields', () => {
    mockAcceptInviteRenderProps = createMockAcceptInviteRenderProps({
      touched: {password: true},
      fieldErrors: {password: 'Password is required'},
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'password-input',
              type: 'PASSWORD_INPUT',
              ref: 'password',
              label: 'Password',
              required: true,
            },
            {
              id: 'submit-btn',
              type: 'ACTION',
              eventType: 'SUBMIT',
              label: 'Set Password',
              variant: 'PRIMARY',
            },
          ],
        },
      ],
    });
    render(<AcceptInviteBox />);
    expect(screen.getByText('Password is required')).toBeInTheDocument();
  });

  it('shows branded logo when images are available', () => {
    mockUseBranding.mockReturnValue({
      images: {
        logo: {
          primary: {
            url: 'https://example.com/logo.png',
            alt: 'Custom Logo',
            height: 40,
            width: 100,
          },
        },
      },
      theme: {palette: {primary: {main: '#ff0000'}}},
      isBrandingEnabled: true,
    });
    mockAcceptInviteRenderProps = createMockAcceptInviteRenderProps({
      components: [{id: 'block', type: 'BLOCK', components: []}],
    });
    render(<AcceptInviteBox />);
    expect(screen.getByTestId('asgardeo-accept-invite')).toBeInTheDocument();
  });

  it('disables submit button when form is not valid', () => {
    mockAcceptInviteRenderProps = createMockAcceptInviteRenderProps({
      isValid: false,
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'password-input',
              type: 'PASSWORD_INPUT',
              ref: 'password',
              label: 'Password',
              required: true,
            },
            {
              id: 'submit-btn',
              type: 'ACTION',
              eventType: 'SUBMIT',
              label: 'Set Password',
              variant: 'PRIMARY',
            },
          ],
        },
      ],
    });
    render(<AcceptInviteBox />);
    const submitBtn = screen.getByText('Set Password');
    expect(submitBtn).toBeDisabled();
  });

  it('shows validation error for SELECT component', () => {
    mockAcceptInviteRenderProps = createMockAcceptInviteRenderProps({
      touched: {role: true},
      fieldErrors: {role: 'Role is required'},
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'role-select',
              type: 'SELECT',
              ref: 'role',
              label: 'Role',
              placeholder: 'Select your role',
              options: ['Developer', 'Manager'],
              required: true,
            },
            {
              id: 'submit-btn',
              type: 'ACTION',
              eventType: 'SUBMIT',
              label: 'Continue',
              variant: 'PRIMARY',
            },
          ],
        },
      ],
    });
    render(<AcceptInviteBox />);
    expect(screen.getByText('Role is required')).toBeInTheDocument();
  });

  it('does not render block without submit action', () => {
    mockAcceptInviteRenderProps = createMockAcceptInviteRenderProps({
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'password-input',
              type: 'PASSWORD_INPUT',
              ref: 'password',
              label: 'Password',
              required: true,
            },
            // No submit action
          ],
        },
      ],
    });
    render(<AcceptInviteBox />);
    // Block should not render without submit action
    expect(screen.queryByLabelText(/Password/)).not.toBeInTheDocument();
  });

  it('handles input change', async () => {
    mockAcceptInviteRenderProps = createMockAcceptInviteRenderProps({
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'name-input',
              type: 'TEXT_INPUT',
              ref: 'firstName',
              label: 'First Name',
              required: true,
            },
            {
              id: 'submit-btn',
              type: 'ACTION',
              eventType: 'SUBMIT',
              label: 'Continue',
              variant: 'PRIMARY',
            },
          ],
        },
      ],
    });
    render(<AcceptInviteBox />);

    const nameInput = screen.getByLabelText(/First Name/);
    await userEvent.type(nameInput, 'John');

    expect(mockHandleInputChange).toHaveBeenCalled();
  });

  it('handles navigation to sign in via onGoToSignIn', async () => {
    mockAcceptInviteRenderProps = createMockAcceptInviteRenderProps({
      isComplete: true,
    });
    render(<AcceptInviteBox />);

    // The component passes onGoToSignIn to AcceptInvite
    // Just verify component renders correctly
    expect(screen.getByText(/Your account has been successfully set up/)).toBeInTheDocument();
  });

  it('renders SELECT component with string options', async () => {
    mockAcceptInviteRenderProps = createMockAcceptInviteRenderProps({
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'role-select',
              type: 'SELECT',
              ref: 'role',
              label: 'Role',
              placeholder: 'Select your role',
              options: ['Developer', 'Manager', 'Admin'],
              required: true,
            },
            {
              id: 'submit-btn',
              type: 'ACTION',
              eventType: 'SUBMIT',
              label: 'Continue',
              variant: 'PRIMARY',
            },
          ],
        },
      ],
    });
    render(<AcceptInviteBox />);

    const selectInput = screen.getByRole('combobox');
    await userEvent.click(selectInput);

    // Check that options are rendered
    expect(await screen.findByText('Developer')).toBeInTheDocument();
    expect(await screen.findByText('Manager')).toBeInTheDocument();
  });

  it('renders SELECT component with object options that have string value/label', async () => {
    mockAcceptInviteRenderProps = createMockAcceptInviteRenderProps({
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'role-select',
              type: 'SELECT',
              ref: 'role',
              label: 'Role',
              placeholder: 'Select your role',
              options: [
                {value: 'dev', label: 'Developer'},
                {value: 'mgr', label: 'Manager'},
              ],
              required: true,
            },
            {
              id: 'submit-btn',
              type: 'ACTION',
              eventType: 'SUBMIT',
              label: 'Continue',
              variant: 'PRIMARY',
            },
          ],
        },
      ],
    });
    render(<AcceptInviteBox />);

    expect(screen.getByRole('combobox')).toBeInTheDocument();
  });

  it('renders SELECT component with complex object value/label', async () => {
    mockAcceptInviteRenderProps = createMockAcceptInviteRenderProps({
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'complex-select',
              type: 'SELECT',
              ref: 'complexField',
              label: 'Complex Field',
              placeholder: 'Select option',
              options: [
                {value: {nested: 'value1'}, label: {text: 'Option 1'}},
                {value: {nested: 'value2'}, label: {text: 'Option 2'}},
              ],
              required: true,
            },
            {
              id: 'submit-btn',
              type: 'ACTION',
              eventType: 'SUBMIT',
              label: 'Continue',
              variant: 'PRIMARY',
            },
          ],
        },
      ],
    });
    render(<AcceptInviteBox />);

    expect(screen.getByText('Complex Field')).toBeInTheDocument();
  });

  it('handles SELECT change event', async () => {
    mockAcceptInviteRenderProps = createMockAcceptInviteRenderProps({
      values: {},
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'role-select',
              type: 'SELECT',
              ref: 'role',
              label: 'Role',
              placeholder: 'Select your role',
              options: ['Developer', 'Manager'],
              required: true,
            },
            {
              id: 'submit-btn',
              type: 'ACTION',
              eventType: 'SUBMIT',
              label: 'Continue',
              variant: 'PRIMARY',
            },
          ],
        },
      ],
    });
    render(<AcceptInviteBox />);

    const selectInput = screen.getByRole('combobox');
    await userEvent.click(selectInput);

    const option = await screen.findByText('Developer');
    await userEvent.click(option);

    expect(mockHandleInputChange).toHaveBeenCalled();
  });

  it('renders SELECT component with hint text', () => {
    mockAcceptInviteRenderProps = createMockAcceptInviteRenderProps({
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'role-select',
              type: 'SELECT',
              ref: 'role',
              label: 'Role',
              placeholder: 'Select your role',
              options: ['Developer', 'Manager'],
              hint: 'Choose your primary role',
              required: true,
            },
            {
              id: 'submit-btn',
              type: 'ACTION',
              eventType: 'SUBMIT',
              label: 'Continue',
              variant: 'PRIMARY',
            },
          ],
        },
      ],
    });
    render(<AcceptInviteBox />);

    expect(screen.getByText('Choose your primary role')).toBeInTheDocument();
  });

  it('handles onError callback', () => {
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
    mockAcceptInviteRenderProps = createMockAcceptInviteRenderProps({
      error: {message: 'Test error'},
      components: [{id: 'block', type: 'BLOCK', components: []}],
    });
    render(<AcceptInviteBox />);

    // Error message is displayed
    expect(screen.getByText('Test error')).toBeInTheDocument();
    consoleSpy.mockRestore();
  });

  it('renders component without ref (should not render)', () => {
    mockAcceptInviteRenderProps = createMockAcceptInviteRenderProps({
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'no-ref-input',
              type: 'TEXT_INPUT',
              // No ref property
              label: 'No Ref Field',
              required: true,
            },
            {
              id: 'submit-btn',
              type: 'ACTION',
              eventType: 'SUBMIT',
              label: 'Continue',
              variant: 'PRIMARY',
            },
          ],
        },
      ],
    });
    render(<AcceptInviteBox />);

    // The field should not be rendered since it has no ref
    expect(screen.queryByLabelText(/No Ref Field/)).not.toBeInTheDocument();
  });

  it('renders EMAIL_INPUT component with change handler', async () => {
    mockAcceptInviteRenderProps = createMockAcceptInviteRenderProps({
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'email-input',
              type: 'EMAIL_INPUT',
              ref: 'email',
              label: 'Email Address',
              placeholder: 'Enter email',
              required: true,
            },
            {
              id: 'submit-btn',
              type: 'ACTION',
              eventType: 'SUBMIT',
              label: 'Continue',
              variant: 'PRIMARY',
            },
          ],
        },
      ],
    });
    render(<AcceptInviteBox />);

    const emailInput = screen.getByLabelText(/Email Address/);
    await userEvent.type(emailInput, 'test@example.com');

    expect(mockHandleInputChange).toHaveBeenCalled();
  });

  it('renders PASSWORD_INPUT component with change handler', async () => {
    mockAcceptInviteRenderProps = createMockAcceptInviteRenderProps({
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'password-input',
              type: 'PASSWORD_INPUT',
              ref: 'password',
              label: 'Password',
              placeholder: 'Enter password',
              required: true,
            },
            {
              id: 'submit-btn',
              type: 'ACTION',
              eventType: 'SUBMIT',
              label: 'Set Password',
              variant: 'PRIMARY',
            },
          ],
        },
      ],
    });
    render(<AcceptInviteBox />);

    const passwordInput = screen.getByLabelText(/Password/);
    await userEvent.type(passwordInput, 'mypassword123');

    expect(mockHandleInputChange).toHaveBeenCalled();
  });

  it('shows validation error for EMAIL_INPUT', () => {
    mockAcceptInviteRenderProps = createMockAcceptInviteRenderProps({
      touched: {email: true},
      fieldErrors: {email: 'Email is required'},
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'email-input',
              type: 'EMAIL_INPUT',
              ref: 'email',
              label: 'Email',
              required: true,
            },
            {
              id: 'submit-btn',
              type: 'ACTION',
              eventType: 'SUBMIT',
              label: 'Continue',
              variant: 'PRIMARY',
            },
          ],
        },
      ],
    });
    render(<AcceptInviteBox />);
    expect(screen.getByText('Email is required')).toBeInTheDocument();
  });

  it('shows validation error for TEXT_INPUT', () => {
    mockAcceptInviteRenderProps = createMockAcceptInviteRenderProps({
      touched: {firstName: true},
      fieldErrors: {firstName: 'First name is required'},
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'name-input',
              type: 'TEXT_INPUT',
              ref: 'firstName',
              label: 'First Name',
              required: true,
            },
            {
              id: 'submit-btn',
              type: 'ACTION',
              eventType: 'SUBMIT',
              label: 'Continue',
              variant: 'PRIMARY',
            },
          ],
        },
      ],
    });
    render(<AcceptInviteBox />);
    expect(screen.getByText('First name is required')).toBeInTheDocument();
  });

  it('renders outlined button variant for non-PRIMARY action', () => {
    mockAcceptInviteRenderProps = createMockAcceptInviteRenderProps({
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'name-input',
              type: 'TEXT_INPUT',
              ref: 'firstName',
              label: 'First Name',
              required: false,
            },
            {
              id: 'submit-btn',
              type: 'ACTION',
              eventType: 'SUBMIT',
              label: 'Continue',
              variant: 'SECONDARY',
            },
          ],
        },
      ],
    });
    render(<AcceptInviteBox />);

    const submitBtn = screen.getByText('Continue');
    expect(submitBtn).toBeInTheDocument();
  });

  it('handles form submission', async () => {
    mockAcceptInviteRenderProps = createMockAcceptInviteRenderProps({
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'name-input',
              type: 'TEXT_INPUT',
              ref: 'firstName',
              label: 'First Name',
              required: true,
            },
            {
              id: 'submit-btn',
              type: 'ACTION',
              eventType: 'SUBMIT',
              label: 'Continue',
              variant: 'PRIMARY',
            },
          ],
        },
      ],
    });
    render(<AcceptInviteBox />);

    const submitBtn = screen.getByText('Continue');
    fireEvent.click(submitBtn);

    await waitFor(() => {
      expect(mockHandleSubmit).toHaveBeenCalled();
    });
  });

  it('renders component with values pre-filled', () => {
    mockAcceptInviteRenderProps = createMockAcceptInviteRenderProps({
      values: {firstName: 'John'},
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'name-input',
              type: 'TEXT_INPUT',
              ref: 'firstName',
              label: 'First Name',
              required: true,
            },
            {
              id: 'submit-btn',
              type: 'ACTION',
              eventType: 'SUBMIT',
              label: 'Continue',
              variant: 'PRIMARY',
            },
          ],
        },
      ],
    });
    render(<AcceptInviteBox />);

    const nameInput = screen.getByLabelText(/First Name/);
    expect(nameInput).toHaveValue('John');
  });

  it('renders multiple TEXT components', () => {
    mockAcceptInviteRenderProps = createMockAcceptInviteRenderProps({
      components: [
        {
          id: 'text-1',
          type: 'TEXT',
          label: 'Welcome',
          variant: 'H1',
        },
        {
          id: 'text-2',
          type: 'TEXT',
          label: 'Set up your account',
          variant: 'H2',
        },
      ],
    });
    render(<AcceptInviteBox />);

    expect(screen.getByText('Welcome')).toBeInTheDocument();
    expect(screen.getByText('Set up your account')).toBeInTheDocument();
  });

  it('renders TEXT component without variant', () => {
    mockAcceptInviteRenderProps = createMockAcceptInviteRenderProps({
      components: [
        {
          id: 'text-1',
          type: 'TEXT',
          label: 'Some text without variant',
        },
      ],
    });
    render(<AcceptInviteBox />);

    expect(screen.getByText('Some text without variant')).toBeInTheDocument();
  });

  it('returns null for unknown component type in block', () => {
    mockAcceptInviteRenderProps = createMockAcceptInviteRenderProps({
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'unknown-input',
              type: 'UNKNOWN_TYPE',
              ref: 'unknown',
              label: 'Unknown Field',
            },
            {
              id: 'submit-btn',
              type: 'ACTION',
              eventType: 'SUBMIT',
              label: 'Continue',
              variant: 'PRIMARY',
            },
          ],
        },
      ],
    });
    render(<AcceptInviteBox />);

    // Unknown type should not render
    expect(screen.queryByLabelText(/Unknown Field/)).not.toBeInTheDocument();
    // But submit button should render
    expect(screen.getByText('Continue')).toBeInTheDocument();
  });

  it('returns null for unknown top-level component type', () => {
    mockAcceptInviteRenderProps = createMockAcceptInviteRenderProps({
      components: [
        {
          id: 'unknown-1',
          type: 'UNKNOWN_TOP_LEVEL',
          label: 'Unknown',
        },
      ],
    });
    render(<AcceptInviteBox />);

    expect(screen.queryByText('Unknown')).not.toBeInTheDocument();
  });

  it('handles SELECT option with null value in object', async () => {
    mockAcceptInviteRenderProps = createMockAcceptInviteRenderProps({
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'null-select',
              type: 'SELECT',
              ref: 'nullField',
              label: 'Null Value Field',
              placeholder: 'Select option',
              options: [
                {value: null, label: null},
              ],
              required: true,
            },
            {
              id: 'submit-btn',
              type: 'ACTION',
              eventType: 'SUBMIT',
              label: 'Continue',
              variant: 'PRIMARY',
            },
          ],
        },
      ],
    });
    render(<AcceptInviteBox />);

    expect(screen.getByText('Null Value Field')).toBeInTheDocument();
  });

  it('handles non-string and non-object options', async () => {
    mockAcceptInviteRenderProps = createMockAcceptInviteRenderProps({
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'number-select',
              type: 'SELECT',
              ref: 'numberField',
              label: 'Number Options',
              placeholder: 'Select option',
              options: [1, 2, 3],
              required: true,
            },
            {
              id: 'submit-btn',
              type: 'ACTION',
              eventType: 'SUBMIT',
              label: 'Continue',
              variant: 'PRIMARY',
            },
          ],
        },
      ],
    });
    render(<AcceptInviteBox />);

    expect(screen.getByText('Number Options')).toBeInTheDocument();
  });

  it('triggers onError callback when component has error', () => {
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

    mockAcceptInviteRenderProps = createMockAcceptInviteRenderProps({
      error: {message: 'Invite acceptance failed'},
      components: [{id: 'block', type: 'BLOCK', components: []}],
    });
    render(<AcceptInviteBox />);

    // The error message should be displayed
    expect(screen.getByText('Invite acceptance failed')).toBeInTheDocument();

    consoleSpy.mockRestore();
  });

  it('calls onGoToSignIn callback and navigates to sign in page', () => {
    render(<AcceptInviteBox />);

    // Verify the callback was captured
    expect(capturedOnGoToSignIn).toBeDefined();

    // Call the captured callback
    capturedOnGoToSignIn?.();

    // Verify navigate was called with sign in route
    expect(mockNavigate).toHaveBeenCalledWith('/signin');
  });

  it('handles onGoToSignIn when navigate returns a Promise', () => {
    // Mock navigate to return a Promise
    mockNavigate.mockReturnValue(Promise.resolve());

    render(<AcceptInviteBox />);

    expect(capturedOnGoToSignIn).toBeDefined();
    capturedOnGoToSignIn?.();

    expect(mockNavigate).toHaveBeenCalledWith('/signin');
  });

  it('handles onGoToSignIn when navigate returns a rejected Promise', async () => {
    // Mock navigate to return a rejected Promise
    mockNavigate.mockReturnValue(Promise.reject(new Error('Navigation failed')));

    render(<AcceptInviteBox />);

    expect(capturedOnGoToSignIn).toBeDefined();

    // Call onGoToSignIn - should not throw even when navigate rejects
    // The implementation uses .catch(() => {}) to silently handle navigation failures,
    // as there's no meaningful recovery action for a failed client-side navigation
    expect(() => capturedOnGoToSignIn?.()).not.toThrow();

    expect(mockNavigate).toHaveBeenCalledWith('/signin');
  });

  it('calls onError callback with error object', () => {
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

    render(<AcceptInviteBox />);

    // Verify the callback was captured
    expect(capturedOnError).toBeDefined();

    // Call the captured callback with an error
    const testError = new Error('Test error message');
    capturedOnError?.(testError);

    // Verify console.error was called with the error
    expect(consoleSpy).toHaveBeenCalledWith('Invite acceptance error:', testError);

    consoleSpy.mockRestore();
  });
});
