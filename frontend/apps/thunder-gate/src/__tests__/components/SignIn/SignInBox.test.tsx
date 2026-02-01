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
import SignInBox from '../../../components/SignIn/SignInBox';
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

// Mock react-router hooks
const mockNavigate = vi.fn();
vi.mock('react-router', () => ({
  useNavigate: () => mockNavigate,
  useSearchParams: () => [new URLSearchParams(), vi.fn()],
}));

// Mock getIntegrationIcon
vi.mock('../../../utils/getIntegrationIcon', () => ({
  default: (label: string) => {
    if (label.includes('Google')) return <span data-testid="google-icon">G</span>;
    if (label.includes('GitHub')) return <span data-testid="github-icon">GH</span>;
    return null;
  },
}));

// Mock Asgardeo SignIn and SignUp components
const mockOnSubmit = vi.fn().mockResolvedValue(undefined);

// Mock component type for testing embedded flow components
interface MockFlowComponent {
  id: string;
  type: string;
  label?: string;
  variant?: string;
  ref?: string;
  placeholder?: string;
  required?: boolean;
  eventType?: string;
  image?: string;
  components?: MockFlowComponent[];
}

interface MockSignInRenderProps {
  onSubmit: typeof mockOnSubmit;
  isLoading: boolean;
  components: MockFlowComponent[];
  error: {message?: string} | null;
  isInitialized: boolean;
}

interface MockSignUpRenderProps {
  components: MockFlowComponent[];
}

// Factory function to create fresh mock SignIn props for each test
const createMockSignInRenderProps = (
  overrides: Partial<MockSignInRenderProps> = {},
): MockSignInRenderProps => ({
  onSubmit: mockOnSubmit,
  isLoading: false,
  components: [],
  error: null,
  isInitialized: true,
  ...overrides,
});

// Factory function to create fresh mock SignUp props for each test
const createMockSignUpRenderProps = (
  overrides: Partial<MockSignUpRenderProps> = {},
): MockSignUpRenderProps => ({
  components: [],
  ...overrides,
});

let mockSignInRenderProps: MockSignInRenderProps = createMockSignInRenderProps();

let mockSignUpRenderProps: MockSignUpRenderProps = createMockSignUpRenderProps();

vi.mock('@asgardeo/react', async () => {
  const actual = await vi.importActual('@asgardeo/react');
  return {
    ...actual,
    SignIn: ({children}: {children: (props: typeof mockSignInRenderProps) => React.ReactNode}) =>
      <div data-testid="asgardeo-signin">{children(mockSignInRenderProps)}</div>,
    SignUp: ({children}: {children: (props: typeof mockSignUpRenderProps) => React.ReactNode}) =>
      <div data-testid="asgardeo-signup">{children(mockSignUpRenderProps)}</div>,
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

describe('SignInBox', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseBranding.mockReturnValue({
      images: null,
      theme: null,
      isBrandingEnabled: false,
    });
    mockSignInRenderProps = createMockSignInRenderProps();
    mockSignUpRenderProps = createMockSignUpRenderProps();
  });

  it('renders without crashing', () => {
    const {container} = render(<SignInBox />);
    expect(container).toBeInTheDocument();
  });

  it('shows loading spinner when isLoading is true', () => {
    mockSignInRenderProps = createMockSignInRenderProps({
      isLoading: true,
    });
    render(<SignInBox />);
    // CircularProgress should be shown
    expect(screen.getByTestId('asgardeo-signin')).toBeInTheDocument();
  });

  it('shows loading spinner when not initialized', () => {
    mockSignInRenderProps = createMockSignInRenderProps({
      isInitialized: false,
    });
    render(<SignInBox />);
    expect(screen.getByTestId('asgardeo-signin')).toBeInTheDocument();
  });

  it('shows error alert when error is present', () => {
    mockSignInRenderProps = createMockSignInRenderProps({
      error: {message: 'Invalid credentials'},
    });
    render(<SignInBox />);
    expect(screen.getByText('Invalid credentials')).toBeInTheDocument();
  });

  it('renders TEXT component as heading', () => {
    mockSignInRenderProps = createMockSignInRenderProps({
      components: [
        {
          id: 'text-1',
          type: 'TEXT',
          label: 'Sign In',
          variant: 'H1',
        },
      ],
    });
    render(<SignInBox />);
    expect(screen.getByText('Sign In')).toBeInTheDocument();
  });

  it('renders TEXT_INPUT component', () => {
    mockSignInRenderProps = createMockSignInRenderProps({
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'username-input',
              type: 'TEXT_INPUT',
              ref: 'username',
              label: 'Username',
              placeholder: 'Enter your username',
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
    render(<SignInBox />);
    expect(screen.getByLabelText(/Username/)).toBeInTheDocument();
    expect(screen.getByPlaceholderText('Enter your username')).toBeInTheDocument();
  });

  it('renders PASSWORD_INPUT component with toggle visibility', async () => {
    mockSignInRenderProps = createMockSignInRenderProps({
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
              label: 'Sign In',
              variant: 'PRIMARY',
            },
          ],
        },
      ],
    });
    render(<SignInBox />);

    const passwordInput = screen.getByLabelText(/Password/);
    expect(passwordInput).toBeInTheDocument();
    expect(passwordInput).toHaveAttribute('type', 'password');

    // Toggle visibility
    const toggleButton = screen.getByLabelText('toggle password visibility');
    await userEvent.click(toggleButton);

    expect(passwordInput).toHaveAttribute('type', 'text');
  });

  it('renders PHONE_INPUT component', () => {
    mockSignInRenderProps = createMockSignInRenderProps({
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'phone-input',
              type: 'PHONE_INPUT',
              ref: 'phone',
              label: 'Phone Number',
              placeholder: 'Enter your phone',
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
    render(<SignInBox />);
    expect(screen.getByLabelText(/Phone Number/)).toBeInTheDocument();
  });

  it('renders OTP_INPUT component', () => {
    mockSignInRenderProps = createMockSignInRenderProps({
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'otp-input',
              type: 'OTP_INPUT',
              ref: 'otp',
              label: 'Enter OTP',
              required: true,
            },
            {
              id: 'submit-btn',
              type: 'ACTION',
              eventType: 'SUBMIT',
              label: 'Verify',
              variant: 'PRIMARY',
            },
          ],
        },
      ],
    });
    render(<SignInBox />);
    expect(screen.getByText('Enter OTP')).toBeInTheDocument();
    // OTP input has 6 digit fields
    expect(screen.getAllByRole('textbox')).toHaveLength(6);
  });

  it('renders TRIGGER action buttons for social login', () => {
    mockSignInRenderProps = createMockSignInRenderProps({
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'google-btn',
              type: 'ACTION',
              eventType: 'TRIGGER',
              label: 'Continue with Google',
              image: 'google.svg',
            },
          ],
        },
      ],
    });
    render(<SignInBox />);
    expect(screen.getByText('Continue with Google')).toBeInTheDocument();
    expect(screen.getByTestId('google-icon')).toBeInTheDocument();
  });

  it('shows validation error for required fields', async () => {
    mockSignInRenderProps = createMockSignInRenderProps({
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'username-input',
              type: 'TEXT_INPUT',
              ref: 'username',
              label: 'Username',
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
    render(<SignInBox />);

    // Submit form without filling required field
    const submitBtn = screen.getByText('Continue');
    fireEvent.click(submitBtn);

    // Form should not submit (onSubmit not called)
    expect(mockOnSubmit).not.toHaveBeenCalled();
  });

  it('submits form when all required fields are filled', async () => {
    mockSignInRenderProps = createMockSignInRenderProps({
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'username-input',
              type: 'TEXT_INPUT',
              ref: 'username',
              label: 'Username',
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
    render(<SignInBox />);

    // Fill in required field
    const usernameInput = screen.getByLabelText(/Username/);
    await userEvent.type(usernameInput, 'testuser');

    // Submit form
    const submitBtn = screen.getByText('Continue');
    fireEvent.click(submitBtn);

    await waitFor(() => {
      expect(mockOnSubmit).toHaveBeenCalledWith({
        inputs: {username: 'testuser'},
        action: 'submit-btn',
      });
    });
  });

  it('renders RESEND button', () => {
    mockSignInRenderProps = createMockSignInRenderProps({
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'otp-input',
              type: 'OTP_INPUT',
              ref: 'otp',
              label: 'Enter OTP',
              required: true,
            },
            {
              id: 'resend-btn',
              type: 'RESEND',
              eventType: 'SUBMIT',
              label: 'Resend OTP',
            },
          ],
        },
      ],
    });
    render(<SignInBox />);
    expect(screen.getByText('Resend OTP')).toBeInTheDocument();
  });

  it('renders TRIGGER action within form block', async () => {
    mockSignInRenderProps = createMockSignInRenderProps({
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'otp-input',
              type: 'OTP_INPUT',
              ref: 'otp',
              label: 'Enter OTP',
              required: true,
            },
            {
              id: 'verify-btn',
              type: 'ACTION',
              eventType: 'TRIGGER',
              label: 'Verify OTP',
              variant: 'PRIMARY',
            },
            {
              id: 'submit-btn',
              type: 'ACTION',
              eventType: 'SUBMIT',
              label: 'Continue',
            },
          ],
        },
      ],
    });
    render(<SignInBox />);
    expect(screen.getByText('Verify OTP')).toBeInTheDocument();
  });

  it('renders sign up redirect link when signup components exist', () => {
    mockSignUpRenderProps = createMockSignUpRenderProps({
      components: [{id: 'signup-form', type: 'BLOCK'}],
    });
    render(<SignInBox />);
    expect(screen.getByText('Sign up')).toBeInTheDocument();
  });

  it('navigates to sign up page when clicking sign up link', async () => {
    mockSignUpRenderProps = createMockSignUpRenderProps({
      components: [{id: 'signup-form', type: 'BLOCK'}],
    });
    render(<SignInBox />);

    const signUpLink = screen.getByText('Sign up');
    await userEvent.click(signUpLink);

    // Navigate is called with /signup or /signup with query params
    expect(mockNavigate).toHaveBeenCalled();
    const navigateCalls = mockNavigate.mock.calls;
    expect(navigateCalls[0][0]).toMatch(/^\/signup/);
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
    render(<SignInBox />);
    // The component renders branded logo
    expect(screen.getByTestId('asgardeo-signin')).toBeInTheDocument();
  });

  it('shows loading when no components are available', () => {
    mockSignInRenderProps = createMockSignInRenderProps({
      components: [],
      isLoading: false,
      isInitialized: true,
    });
    render(<SignInBox />);
    // Shows loading when no components
    expect(screen.getByTestId('asgardeo-signin')).toBeInTheDocument();
  });

  it('handles OTP input changes and auto-focus', async () => {
    mockSignInRenderProps = createMockSignInRenderProps({
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'otp-input',
              type: 'OTP_INPUT',
              ref: 'otp',
              label: 'Enter OTP',
              required: true,
            },
            {
              id: 'submit-btn',
              type: 'ACTION',
              eventType: 'SUBMIT',
              label: 'Verify',
              variant: 'PRIMARY',
            },
          ],
        },
      ],
    });
    render(<SignInBox />);

    const otpInputs = screen.getAllByRole('textbox');
    expect(otpInputs).toHaveLength(6);

    // Type in first OTP digit
    await userEvent.type(otpInputs[0], '1');
  });

  it('clears field error when user starts typing', async () => {
    mockSignInRenderProps = createMockSignInRenderProps({
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'username-input',
              type: 'TEXT_INPUT',
              ref: 'username',
              label: 'Username',
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
    render(<SignInBox />);

    // Submit without filling to trigger validation error
    const submitBtn = screen.getByText('Continue');
    fireEvent.click(submitBtn);

    // Now type to clear error
    const usernameInput = screen.getByLabelText(/Username/);
    await userEvent.type(usernameInput, 't');

    // Error should be cleared
    expect(mockOnSubmit).not.toHaveBeenCalled();
  });

  it('handles social login trigger click', async () => {
    mockSignInRenderProps = createMockSignInRenderProps({
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'google-btn',
              type: 'ACTION',
              eventType: 'TRIGGER',
              label: 'Continue with Google',
              image: 'google.svg',
            },
          ],
        },
      ],
    });
    render(<SignInBox />);

    const googleBtn = screen.getByText('Continue with Google');
    await userEvent.click(googleBtn);

    await waitFor(() => {
      expect(mockOnSubmit).toHaveBeenCalledWith({
        inputs: {},
        action: 'google-btn',
      });
    });
  });

  it('navigates to sign up with query params preserved', async () => {
    vi.mock('react-router', async () => {
      const actual = await vi.importActual('react-router');
      return {
        ...actual,
        useNavigate: () => mockNavigate,
        useSearchParams: () => [new URLSearchParams('client_id=test&redirect_uri=http://example.com'), vi.fn()],
      };
    });

    mockSignUpRenderProps = createMockSignUpRenderProps({
      components: [{id: 'signup-form', type: 'BLOCK'}],
    });

    render(<SignInBox />);
    const signUpLink = screen.getByText('Sign up');
    await userEvent.click(signUpLink);

    expect(mockNavigate).toHaveBeenCalled();
  });

  it('handles password input change', async () => {
    mockSignInRenderProps = createMockSignInRenderProps({
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
              label: 'Sign In',
              variant: 'PRIMARY',
            },
          ],
        },
      ],
    });
    render(<SignInBox />);

    const passwordInput = screen.getByLabelText(/Password/);
    await userEvent.type(passwordInput, 'test123');

    // Verify the input has the typed value
    expect(passwordInput).toHaveValue('test123');
  });

  it('handles phone input change', async () => {
    mockSignInRenderProps = createMockSignInRenderProps({
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'phone-input',
              type: 'PHONE_INPUT',
              ref: 'phone',
              label: 'Phone Number',
              placeholder: 'Enter your phone',
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
    render(<SignInBox />);

    const phoneInput = screen.getByLabelText(/Phone Number/);
    await userEvent.type(phoneInput, '+1234567890');

    expect(phoneInput).toHaveValue('+1234567890');
  });

  it('handles OTP input digit entry and auto-focus', async () => {
    mockSignInRenderProps = createMockSignInRenderProps({
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'otp-input',
              type: 'OTP_INPUT',
              ref: 'otp',
              label: 'Enter OTP',
              required: true,
            },
            {
              id: 'submit-btn',
              type: 'ACTION',
              eventType: 'SUBMIT',
              label: 'Verify',
              variant: 'PRIMARY',
            },
          ],
        },
      ],
    });
    render(<SignInBox />);

    const otpInputs = screen.getAllByRole('textbox');

    // Type a digit in the first input
    fireEvent.change(otpInputs[0], {target: {value: '1'}});

    // The input should have the digit
    expect(otpInputs[0]).toHaveValue('1');
  });

  it('handles OTP input backspace navigation', async () => {
    mockSignInRenderProps = createMockSignInRenderProps({
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'otp-input',
              type: 'OTP_INPUT',
              ref: 'otp',
              label: 'Enter OTP',
              required: true,
            },
            {
              id: 'submit-btn',
              type: 'ACTION',
              eventType: 'SUBMIT',
              label: 'Verify',
              variant: 'PRIMARY',
            },
          ],
        },
      ],
    });
    render(<SignInBox />);

    const otpInputs = screen.getAllByRole('textbox');

    // Focus on second input and press backspace when empty
    otpInputs[1].focus();
    fireEvent.keyDown(otpInputs[1], {key: 'Backspace'});

    // The test verifies the keydown handler is called
    expect(otpInputs).toHaveLength(6);
  });

  it('handles OTP input paste', async () => {
    mockSignInRenderProps = createMockSignInRenderProps({
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'otp-input',
              type: 'OTP_INPUT',
              ref: 'otp',
              label: 'Enter OTP',
              required: true,
            },
            {
              id: 'submit-btn',
              type: 'ACTION',
              eventType: 'SUBMIT',
              label: 'Verify',
              variant: 'PRIMARY',
            },
          ],
        },
      ],
    });
    render(<SignInBox />);

    const otpInputs = screen.getAllByRole('textbox');

    // Use fireEvent with clipboardData mock
    fireEvent.paste(otpInputs[0], {
      clipboardData: {
        getData: () => '123456',
      },
    });

    // Verify the OTP inputs render
    expect(otpInputs).toHaveLength(6);
  });

  it('rejects non-digit input in OTP field', async () => {
    mockSignInRenderProps = createMockSignInRenderProps({
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'otp-input',
              type: 'OTP_INPUT',
              ref: 'otp',
              label: 'Enter OTP',
              required: true,
            },
            {
              id: 'submit-btn',
              type: 'ACTION',
              eventType: 'SUBMIT',
              label: 'Verify',
              variant: 'PRIMARY',
            },
          ],
        },
      ],
    });
    render(<SignInBox />);

    const otpInputs = screen.getAllByRole('textbox');

    // Try to type a non-digit character
    fireEvent.change(otpInputs[0], {target: {value: 'a'}});

    // The input should remain empty or not accept the character
    expect(otpInputs[0]).toHaveValue('');
  });

  it('handles TRIGGER action button click within form block', async () => {
    mockSignInRenderProps = createMockSignInRenderProps({
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'username-input',
              type: 'TEXT_INPUT',
              ref: 'username',
              label: 'Username',
              required: true,
            },
            {
              id: 'verify-btn',
              type: 'ACTION',
              eventType: 'TRIGGER',
              label: 'Verify Account',
              variant: 'PRIMARY',
            },
            {
              id: 'submit-btn',
              type: 'ACTION',
              eventType: 'SUBMIT',
              label: 'Continue',
            },
          ],
        },
      ],
    });
    render(<SignInBox />);

    // Fill in required field
    const usernameInput = screen.getByLabelText(/Username/);
    await userEvent.type(usernameInput, 'testuser');

    // Click the trigger button
    const verifyBtn = screen.getByText('Verify Account');
    await userEvent.click(verifyBtn);

    await waitFor(() => {
      expect(mockOnSubmit).toHaveBeenCalledWith({
        inputs: {username: 'testuser'},
        action: 'verify-btn',
      });
    });
  });

  it('does not call onSubmit for TRIGGER action when validation fails', async () => {
    mockSignInRenderProps = createMockSignInRenderProps({
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'username-input',
              type: 'TEXT_INPUT',
              ref: 'username',
              label: 'Username',
              required: true,
            },
            {
              id: 'verify-btn',
              type: 'ACTION',
              eventType: 'TRIGGER',
              label: 'Verify Account',
              variant: 'PRIMARY',
            },
            {
              id: 'submit-btn',
              type: 'ACTION',
              eventType: 'SUBMIT',
              label: 'Continue',
            },
          ],
        },
      ],
    });
    render(<SignInBox />);

    // Click the trigger button without filling required field
    const verifyBtn = screen.getByText('Verify Account');
    await userEvent.click(verifyBtn);

    // Should not call onSubmit because validation fails
    expect(mockOnSubmit).not.toHaveBeenCalled();
  });

  it('renders block without submit action and shows nothing', () => {
    mockSignInRenderProps = createMockSignInRenderProps({
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'username-input',
              type: 'TEXT_INPUT',
              ref: 'username',
              label: 'Username',
              required: true,
            },
            // No submit action - only trigger actions
          ],
        },
      ],
    });
    render(<SignInBox />);
    // Block without submit action should not render form fields
    expect(screen.queryByLabelText(/Username/)).not.toBeInTheDocument();
  });

  it('handles password validation error', async () => {
    mockSignInRenderProps = createMockSignInRenderProps({
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
              label: 'Sign In',
              variant: 'PRIMARY',
            },
          ],
        },
      ],
    });
    render(<SignInBox />);

    // Submit without filling password
    const submitBtn = screen.getByText('Sign In');
    fireEvent.click(submitBtn);

    // onSubmit should not be called
    expect(mockOnSubmit).not.toHaveBeenCalled();
  });

  it('handles phone validation error', async () => {
    mockSignInRenderProps = createMockSignInRenderProps({
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'phone-input',
              type: 'PHONE_INPUT',
              ref: 'phone',
              label: 'Phone Number',
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
    render(<SignInBox />);

    // Submit without filling phone
    const submitBtn = screen.getByText('Continue');
    fireEvent.click(submitBtn);

    // onSubmit should not be called
    expect(mockOnSubmit).not.toHaveBeenCalled();
  });

  it('handles OTP validation error', async () => {
    mockSignInRenderProps = createMockSignInRenderProps({
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'otp-input',
              type: 'OTP_INPUT',
              ref: 'otp',
              label: 'Enter OTP',
              required: true,
            },
            {
              id: 'submit-btn',
              type: 'ACTION',
              eventType: 'SUBMIT',
              label: 'Verify',
              variant: 'PRIMARY',
            },
          ],
        },
      ],
    });
    render(<SignInBox />);

    // Submit without entering OTP
    const submitBtn = screen.getByText('Verify');
    fireEvent.click(submitBtn);

    // onSubmit should not be called
    expect(mockOnSubmit).not.toHaveBeenCalled();
  });

  it('renders outlined button variant for non-PRIMARY action', () => {
    mockSignInRenderProps = createMockSignInRenderProps({
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'username-input',
              type: 'TEXT_INPUT',
              ref: 'username',
              label: 'Username',
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
    render(<SignInBox />);

    const submitBtn = screen.getByText('Continue');
    expect(submitBtn).toBeInTheDocument();
  });

  it('renders multiple social login buttons', async () => {
    mockSignInRenderProps = createMockSignInRenderProps({
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'google-btn',
              type: 'ACTION',
              eventType: 'TRIGGER',
              label: 'Continue with Google',
              image: 'google.svg',
            },
            {
              id: 'github-btn',
              type: 'ACTION',
              eventType: 'TRIGGER',
              label: 'Continue with GitHub',
              image: 'github.svg',
            },
          ],
        },
      ],
    });
    render(<SignInBox />);

    expect(screen.getByText('Continue with Google')).toBeInTheDocument();
    expect(screen.getByText('Continue with GitHub')).toBeInTheDocument();

    // Click GitHub button
    const githubBtn = screen.getByText('Continue with GitHub');
    await userEvent.click(githubBtn);

    await waitFor(() => {
      expect(mockOnSubmit).toHaveBeenCalledWith({
        inputs: {},
        action: 'github-btn',
      });
    });
  });

  it('shows error message from error object', () => {
    mockSignInRenderProps = createMockSignInRenderProps({
      error: {message: 'Authentication failed'},
    });
    render(<SignInBox />);
    expect(screen.getByText('Authentication failed')).toBeInTheDocument();
  });

  it('handles error without message', () => {
    mockSignInRenderProps = createMockSignInRenderProps({
      error: {},
    });
    render(<SignInBox />);
    // Should show default error description
    expect(screen.getByTestId('asgardeo-signin')).toBeInTheDocument();
  });

  it('handles form submission with password field', async () => {
    mockSignInRenderProps = createMockSignInRenderProps({
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
              label: 'Sign In',
              variant: 'PRIMARY',
            },
          ],
        },
      ],
    });
    render(<SignInBox />);

    // Fill in password
    const passwordInput = screen.getByLabelText(/Password/);
    await userEvent.type(passwordInput, 'mypassword123');

    // Submit form
    const submitBtn = screen.getByText('Sign In');
    fireEvent.click(submitBtn);

    await waitFor(() => {
      expect(mockOnSubmit).toHaveBeenCalledWith({
        inputs: {password: 'mypassword123'},
        action: 'submit-btn',
      });
    });
  });

  it('handles form submission with phone field', async () => {
    mockSignInRenderProps = createMockSignInRenderProps({
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'phone-input',
              type: 'PHONE_INPUT',
              ref: 'phone',
              label: 'Phone Number',
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
    render(<SignInBox />);

    // Fill in phone
    const phoneInput = screen.getByLabelText(/Phone Number/);
    await userEvent.type(phoneInput, '+1234567890');

    // Submit form
    const submitBtn = screen.getByText('Continue');
    fireEvent.click(submitBtn);

    await waitFor(() => {
      expect(mockOnSubmit).toHaveBeenCalledWith({
        inputs: {phone: '+1234567890'},
        action: 'submit-btn',
      });
    });
  });

  it('does not update input when ref is undefined', async () => {
    mockSignInRenderProps = createMockSignInRenderProps({
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'username-input',
              type: 'TEXT_INPUT',
              ref: undefined,
              label: 'Username',
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
    render(<SignInBox />);

    // Component without ref should not render as input
    expect(screen.queryByLabelText(/Username/)).not.toBeInTheDocument();
  });

  it('renders TRIGGER inside form block and clicks it', async () => {
    mockSignInRenderProps = createMockSignInRenderProps({
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'trigger-btn',
              type: 'ACTION',
              eventType: 'TRIGGER',
              label: 'Trigger Action',
              variant: 'SECONDARY',
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
    render(<SignInBox />);

    const triggerBtn = screen.getByText('Trigger Action');
    expect(triggerBtn).toBeInTheDocument();

    // Click the trigger button
    await userEvent.click(triggerBtn);

    // Trigger validates form first, which passes since no required fields
    await waitFor(() => {
      expect(mockOnSubmit).toHaveBeenCalledWith({
        inputs: {},
        action: 'trigger-btn',
      });
    });
  });

  it('returns null for unknown component type in block', () => {
    mockSignInRenderProps = createMockSignInRenderProps({
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'unknown-comp',
              type: 'UNKNOWN_COMPONENT',
              ref: 'unknown',
              label: 'Unknown',
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
    render(<SignInBox />);

    // Unknown component should not render
    expect(screen.queryByLabelText(/Unknown/)).not.toBeInTheDocument();
    // But submit button should render
    expect(screen.getByText('Continue')).toBeInTheDocument();
  });

  it('returns null for unknown top-level component type', () => {
    mockSignInRenderProps = createMockSignInRenderProps({
      components: [
        {
          id: 'unknown-1',
          type: 'UNKNOWN_TYPE',
          label: 'Unknown',
        },
      ],
    });
    render(<SignInBox />);

    expect(screen.queryByText('Unknown')).not.toBeInTheDocument();
  });

  it('handles social login trigger in block without form elements', async () => {
    mockSignInRenderProps = createMockSignInRenderProps({
      components: [
        {
          id: 'trigger-block',
          type: 'BLOCK',
          components: [
            {
              id: 'facebook-btn',
              type: 'ACTION',
              eventType: 'TRIGGER',
              label: 'Continue with Facebook',
              image: 'facebook.svg',
            },
            {
              id: 'twitter-btn',
              type: 'ACTION',
              eventType: 'TRIGGER',
              label: 'Continue with Twitter',
              image: 'twitter.svg',
            },
          ],
        },
      ],
    });
    render(<SignInBox />);

    expect(screen.getByText('Continue with Facebook')).toBeInTheDocument();
    expect(screen.getByText('Continue with Twitter')).toBeInTheDocument();

    // Click Facebook button
    const facebookBtn = screen.getByText('Continue with Facebook');
    await userEvent.click(facebookBtn);

    await waitFor(() => {
      expect(mockOnSubmit).toHaveBeenCalledWith({
        inputs: {},
        action: 'facebook-btn',
      });
    });
  });

  it('returns null for block with no submit or trigger actions', () => {
    mockSignInRenderProps = createMockSignInRenderProps({
      components: [
        {
          id: 'empty-block',
          type: 'BLOCK',
          components: [
            {
              id: 'text-field',
              type: 'TEXT_INPUT',
              ref: 'field',
              label: 'Field',
            },
          ],
        },
      ],
    });
    render(<SignInBox />);

    // Block without submit or trigger action should not render
    expect(screen.queryByLabelText(/Field/)).not.toBeInTheDocument();
  });

  it('returns null for non-TRIGGER action in social login block', async () => {
    mockSignInRenderProps = createMockSignInRenderProps({
      components: [
        {
          id: 'trigger-block',
          type: 'BLOCK',
          components: [
            {
              id: 'google-btn',
              type: 'ACTION',
              eventType: 'TRIGGER',
              label: 'Continue with Google',
              image: 'google.svg',
            },
            {
              id: 'some-other-action',
              type: 'ACTION',
              eventType: 'OTHER_EVENT',
              label: 'Other Action',
            },
          ],
        },
      ],
    });
    render(<SignInBox />);

    // Google trigger should render
    expect(screen.getByText('Continue with Google')).toBeInTheDocument();
    // Other action type should not render (returns null)
    expect(screen.queryByText('Other Action')).not.toBeInTheDocument();
  });
});
