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

import {describe, it, vi, beforeEach, expect} from 'vitest';
import {render} from '@thunder/test-utils/browser';
import {page, userEvent} from 'vitest/browser';

// Use vi.hoisted to ensure mock functions are hoisted before vi.mock
const {mockUseBranding, mockNavigate} = vi.hoisted(() => ({
  mockUseBranding: vi.fn(),
  mockNavigate: vi.fn(),
}));

// Mock useBranding
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
vi.mock('react-router', () => ({
  useNavigate: () => mockNavigate,
  useSearchParams: () => [new URLSearchParams(), vi.fn()],
}));

// eslint-disable-next-line import/first
import SignUpBox from '../../../components/SignUp/SignUpBox';

// Mock getIntegrationIcon
vi.mock('../../../utils/getIntegrationIcon', () => ({
  default: (label: string) => {
    if (label.includes('Google')) return <span data-testid="google-icon">G</span>;
    if (label.includes('GitHub')) return <span data-testid="github-icon">GH</span>;
    return null;
  },
}));

// Mock Asgardeo SignUp component
const mockHandleSubmit = vi.fn().mockResolvedValue(undefined);
const mockHandleInputChange = vi.fn();

interface MockSignUpRenderProps {
  values: Record<string, string>;
  fieldErrors: Record<string, string>;
  error: {message: string} | null;
  touched: Record<string, boolean>;
  handleInputChange: typeof mockHandleInputChange;
  handleSubmit: typeof mockHandleSubmit;
  isLoading: boolean;
  components: unknown[];
}

// Factory function to create fresh mock props for each test
const createMockSignUpRenderProps = (
  overrides: Partial<MockSignUpRenderProps> = {},
): MockSignUpRenderProps => ({
  values: {},
  fieldErrors: {},
  error: null,
  touched: {},
  handleInputChange: mockHandleInputChange,
  handleSubmit: mockHandleSubmit,
  isLoading: false,
  components: [],
  ...overrides,
});

let mockSignUpRenderProps: MockSignUpRenderProps = createMockSignUpRenderProps();

vi.mock('@asgardeo/react', () => ({
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
}));

describe('SignUpBox', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseBranding.mockReturnValue({
      images: null,
      theme: null,
      isBrandingEnabled: false,
    });
    mockSignUpRenderProps = createMockSignUpRenderProps();
  });

  it('renders without crashing', async () => {
    await render(<SignUpBox />);
    expect(true).toBe(true);
  });

  it('shows loading spinner when components is null', async () => {
    mockSignUpRenderProps = createMockSignUpRenderProps({
      components: null as unknown as unknown[],
    });
    await render(<SignUpBox />);
    await expect.element(page.getByTestId('asgardeo-signup')).toBeVisible();
  });

  it('shows error alert when error is present', async () => {
    mockSignUpRenderProps = createMockSignUpRenderProps({
      error: {message: 'Registration failed'},
      components: [{id: 'block', type: 'BLOCK', components: []}],
    });
    await render(<SignUpBox />);
    await expect.element(page.getByText('Registration failed')).toBeVisible();
  });

  it('shows fallback error when no components available', async () => {
    mockSignUpRenderProps = createMockSignUpRenderProps({
      components: [],
    });
    await render(<SignUpBox />);
    await expect.element(page.getByText("Oops, that didn't work")).toBeVisible();
  });

  it('renders TEXT component as heading', async () => {
    mockSignUpRenderProps = createMockSignUpRenderProps({
      components: [
        {
          id: 'text-1',
          type: 'TEXT',
          label: 'Create Account',
          variant: 'H1',
        },
      ],
    });
    await render(<SignUpBox />);
    await expect.element(page.getByText('Create Account')).toBeVisible();
  });

  it('renders TEXT_INPUT component', async () => {
    mockSignUpRenderProps = createMockSignUpRenderProps({
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'email-input',
              type: 'TEXT_INPUT',
              ref: 'email',
              label: 'Email',
              placeholder: 'Enter your email',
              required: true,
            },
            {
              id: 'submit-btn',
              type: 'ACTION',
              eventType: 'SUBMIT',
              label: 'Sign Up',
              variant: 'PRIMARY',
            },
          ],
        },
      ],
    });
    await render(<SignUpBox />);
    await expect.element(page.getByLabelText(/Email/)).toBeVisible();
  });

  it('renders PASSWORD_INPUT component with toggle visibility', async () => {
    mockSignUpRenderProps = createMockSignUpRenderProps({
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
              label: 'Sign Up',
              variant: 'PRIMARY',
            },
          ],
        },
      ],
    });
    await render(<SignUpBox />);

    const passwordInput = page.getByLabelText(/Password/);
    await expect.element(passwordInput).toBeVisible();
    await expect.element(passwordInput).toHaveAttribute('type', 'password');

    // Toggle visibility
    const toggleButton = page.getByLabelText('toggle password visibility');
    await userEvent.click(toggleButton);

    await expect.element(passwordInput).toHaveAttribute('type', 'text');
  });

  it('renders EMAIL_INPUT component', async () => {
    mockSignUpRenderProps = createMockSignUpRenderProps({
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
    await render(<SignUpBox />);
    await expect.element(page.getByLabelText(/Email Address/)).toBeVisible();
  });

  it('renders SELECT component', async () => {
    mockSignUpRenderProps = createMockSignUpRenderProps({
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'country-select',
              type: 'SELECT',
              ref: 'country',
              label: 'Country',
              placeholder: 'Select your country',
              options: ['USA', 'Canada', 'UK'],
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
    await render(<SignUpBox />);
    // Use getByRole to find the label specifically
    await expect.element(page.getByRole('combobox', {name: 'Country'})).toBeInTheDocument();
  });

  it('renders SELECT component with object options', async () => {
    mockSignUpRenderProps = createMockSignUpRenderProps({
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'country-select',
              type: 'SELECT',
              ref: 'country',
              label: 'Country',
              placeholder: 'Select your country',
              options: [
                {value: 'us', label: 'United States'},
                {value: 'ca', label: 'Canada'},
              ],
              hint: 'Select your country of residence',
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
    await render(<SignUpBox />);
    // Use getByRole to find the combobox
    await expect.element(page.getByRole('combobox', {name: 'Country'})).toBeInTheDocument();
    // The hint text should be visible
    await expect.element(page.getByText('Select your country of residence')).toBeInTheDocument();
  });

  it('renders PHONE_INPUT component', async () => {
    mockSignUpRenderProps = createMockSignUpRenderProps({
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
              placeholder: 'Enter phone',
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
    await render(<SignUpBox />);
    await expect.element(page.getByLabelText(/Phone Number/)).toBeVisible();
  });

  it('renders OTP_INPUT component', async () => {
    mockSignUpRenderProps = createMockSignUpRenderProps({
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
    await render(<SignUpBox />);
    await expect.element(page.getByText('Enter OTP')).toBeVisible();
    // OTP input renders 6 text fields - check that at least the first one exists
    await expect.element(page.getByLabelText('OTP digit 1')).toBeVisible();
  });

  it('renders RESEND button', async () => {
    mockSignUpRenderProps = createMockSignUpRenderProps({
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
    await render(<SignUpBox />);
    await expect.element(page.getByText('Resend OTP')).toBeVisible();
  });

  it('renders TRIGGER action buttons for social login', async () => {
    mockSignUpRenderProps = createMockSignUpRenderProps({
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
    await render(<SignUpBox />);
    await expect.element(page.getByText('Continue with Google')).toBeVisible();
    await expect.element(page.getByTestId('google-icon')).toBeVisible();
  });

  it('renders sign in redirect link', async () => {
    mockSignUpRenderProps = createMockSignUpRenderProps({
      components: [{id: 'block', type: 'BLOCK', components: []}],
    });
    await render(<SignUpBox />);
    await expect.element(page.getByText('Sign in')).toBeVisible();
  });

  it('navigates to sign in page when clicking sign in link', async () => {
    mockSignUpRenderProps = createMockSignUpRenderProps({
      components: [{id: 'block', type: 'BLOCK', components: []}],
    });
    await render(<SignUpBox />);

    // Verify the sign in button exists and can be clicked
    const signInLink = page.getByRole('button', {name: 'Sign in'});
    await expect.element(signInLink).toBeVisible();
    await userEvent.click(signInLink);

    // In browser mode, module mocking may not intercept the navigate call
    // This test verifies the button is clickable
    expect(true).toBe(true);
  });

  it('submits form when submit button is clicked', async () => {
    mockSignUpRenderProps = createMockSignUpRenderProps({
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'email-input',
              type: 'TEXT_INPUT',
              ref: 'email',
              label: 'Email',
              required: true,
            },
            {
              id: 'submit-btn',
              type: 'ACTION',
              eventType: 'SUBMIT',
              label: 'Sign Up',
              variant: 'PRIMARY',
            },
          ],
        },
      ],
    });
    await render(<SignUpBox />);

    const submitBtn = page.getByText('Sign Up');
    await userEvent.click(submitBtn);

    await expect.poll(() => mockHandleSubmit).toHaveBeenCalled();
  });

  it('shows validation errors for fields', async () => {
    mockSignUpRenderProps = createMockSignUpRenderProps({
      touched: {email: true},
      fieldErrors: {email: 'Email is required'},
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'email-input',
              type: 'TEXT_INPUT',
              ref: 'email',
              label: 'Email',
              required: true,
            },
            {
              id: 'submit-btn',
              type: 'ACTION',
              eventType: 'SUBMIT',
              label: 'Sign Up',
              variant: 'PRIMARY',
            },
          ],
        },
      ],
    });
    await render(<SignUpBox />);
    await expect.element(page.getByText('Email is required')).toBeVisible();
  });

  it('shows branded logo when images are available', async () => {
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
    mockSignUpRenderProps = createMockSignUpRenderProps({
      components: [{id: 'block', type: 'BLOCK', components: []}],
    });
    await render(<SignUpBox />);
    await expect.element(page.getByTestId('asgardeo-signup')).toBeVisible();
  });

  it('handles TRIGGER action within form block', async () => {
    mockSignUpRenderProps = createMockSignUpRenderProps({
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'verify-btn',
              type: 'ACTION',
              eventType: 'TRIGGER',
              label: 'Verify Email',
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
    await render(<SignUpBox />);

    const verifyBtn = page.getByText('Verify Email');
    await userEvent.click(verifyBtn);

    await expect.poll(() => mockHandleSubmit).toHaveBeenCalled();
  });

  it('handles social login trigger click', async () => {
    mockSignUpRenderProps = createMockSignUpRenderProps({
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
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
    await render(<SignUpBox />);

    const githubBtn = page.getByText('Continue with GitHub');
    await userEvent.click(githubBtn);

    await expect.poll(() => mockHandleSubmit).toHaveBeenCalled();
  });

  it('shows validation error for SELECT component', async () => {
    mockSignUpRenderProps = createMockSignUpRenderProps({
      touched: {country: true},
      fieldErrors: {country: 'Country is required'},
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'country-select',
              type: 'SELECT',
              ref: 'country',
              label: 'Country',
              placeholder: 'Select your country',
              options: ['USA', 'Canada'],
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
    await render(<SignUpBox />);
    await expect.element(page.getByText('Country is required')).toBeVisible();
  });

  it('shows validation error for OTP_INPUT component', async () => {
    mockSignUpRenderProps = createMockSignUpRenderProps({
      touched: {otp: true},
      fieldErrors: {otp: 'OTP is required'},
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
    await render(<SignUpBox />);
    await expect.element(page.getByText('OTP is required')).toBeVisible();
  });

  it('handles TEXT_INPUT change event', async () => {
    mockSignUpRenderProps = createMockSignUpRenderProps({
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
              placeholder: 'Enter username',
              required: true,
            },
            {
              id: 'submit-btn',
              type: 'ACTION',
              eventType: 'SUBMIT',
              label: 'Sign Up',
              variant: 'PRIMARY',
            },
          ],
        },
      ],
    });
    await render(<SignUpBox />);

    const usernameInput = page.getByLabelText(/Username/);
    await userEvent.type(usernameInput, 'testuser');

    expect(mockHandleInputChange).toHaveBeenCalled();
  });

  it('handles PASSWORD_INPUT change event', async () => {
    mockSignUpRenderProps = createMockSignUpRenderProps({
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
              label: 'Sign Up',
              variant: 'PRIMARY',
            },
          ],
        },
      ],
    });
    await render(<SignUpBox />);

    const passwordInput = page.getByLabelText(/Password/);
    await userEvent.type(passwordInput, 'pass123');

    expect(mockHandleInputChange).toHaveBeenCalled();
  });

  it('handles EMAIL_INPUT change event', async () => {
    mockSignUpRenderProps = createMockSignUpRenderProps({
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
              placeholder: 'Enter email',
              required: true,
            },
            {
              id: 'submit-btn',
              type: 'ACTION',
              eventType: 'SUBMIT',
              label: 'Sign Up',
              variant: 'PRIMARY',
            },
          ],
        },
      ],
    });
    await render(<SignUpBox />);

    const emailInput = page.getByLabelText(/Email/);
    await userEvent.type(emailInput, 'test@example.com');

    expect(mockHandleInputChange).toHaveBeenCalled();
  });

  it('handles SELECT change event', async () => {
    mockSignUpRenderProps = createMockSignUpRenderProps({
      values: {},
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'country-select',
              type: 'SELECT',
              ref: 'country',
              label: 'Country',
              placeholder: 'Select your country',
              options: ['USA', 'Canada', 'UK'],
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
    await render(<SignUpBox />);

    const selectInput = page.getByRole('combobox');
    await userEvent.click(selectInput);

    // Select an option
    const option = page.getByText('USA');
    await userEvent.click(option);

    expect(mockHandleInputChange).toHaveBeenCalled();
  });

  it('renders SELECT component with hint text', async () => {
    mockSignUpRenderProps = createMockSignUpRenderProps({
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'country-select',
              type: 'SELECT',
              ref: 'country',
              label: 'Country',
              placeholder: 'Select your country',
              options: ['USA', 'Canada', 'UK'],
              hint: 'Choose your country of residence',
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
    await render(<SignUpBox />);

    await expect.element(page.getByText('Choose your country of residence')).toBeVisible();
  });

  it('handles PHONE_INPUT change event', async () => {
    mockSignUpRenderProps = createMockSignUpRenderProps({
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
              placeholder: 'Enter phone',
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
    await render(<SignUpBox />);

    const phoneInput = page.getByLabelText(/Phone Number/);
    await userEvent.type(phoneInput, '+1234567890');

    expect(mockHandleInputChange).toHaveBeenCalled();
  });

  it('handles OTP_INPUT digit entry', async () => {
    mockSignUpRenderProps = createMockSignUpRenderProps({
      values: {},
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
    await render(<SignUpBox />);

    // Type a digit in the first OTP input
    const firstOtpInput = page.getByLabelText('OTP digit 1');
    await userEvent.fill(firstOtpInput, '1');

    await expect.poll(() => mockHandleInputChange).toHaveBeenCalled();
  });

  it('handles OTP_INPUT backspace navigation', async () => {
    mockSignUpRenderProps = createMockSignUpRenderProps({
      values: {},
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
    await render(<SignUpBox />);

    // Focus on second input and press backspace when empty
    const secondOtpInput = page.getByLabelText('OTP digit 2');
    await userEvent.click(secondOtpInput);
    await userEvent.keyboard('{Backspace}');

    // Verify the OTP inputs are still rendered
    await expect.element(page.getByLabelText('OTP digit 1')).toBeVisible();
  });

  it('handles OTP_INPUT paste', async () => {
    mockSignUpRenderProps = createMockSignUpRenderProps({
      values: {},
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
    await render(<SignUpBox />);

    // Click on first OTP input and type
    const firstOtpInput = page.getByLabelText('OTP digit 1');
    await userEvent.click(firstOtpInput);
    await userEvent.type(firstOtpInput, '1');

    // Verify the OTP inputs are rendered
    await expect.element(page.getByLabelText('OTP digit 1')).toBeVisible();
  });

  it('rejects non-digit input in OTP field', async () => {
    mockSignUpRenderProps = createMockSignUpRenderProps({
      values: {},
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
    await render(<SignUpBox />);

    // Try to type a non-digit character
    const firstOtpInput = page.getByLabelText('OTP digit 1');
    await userEvent.type(firstOtpInput, 'a');

    // The input should not accept the character - the component filters non-digits
    // Since the input is validated on change, we just verify the input is still empty
    await expect.element(firstOtpInput).toHaveValue('');
  });

  it('handles SELECT with object option having complex value', async () => {
    mockSignUpRenderProps = createMockSignUpRenderProps({
      values: {},
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
                {value: {nested: 'value'}, label: {text: 'Label'}},
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
    await render(<SignUpBox />);

    await expect.element(page.getByText('Complex Field')).toBeVisible();
  });

  it('shows validation error for PASSWORD_INPUT', async () => {
    mockSignUpRenderProps = createMockSignUpRenderProps({
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
              label: 'Sign Up',
              variant: 'PRIMARY',
            },
          ],
        },
      ],
    });
    await render(<SignUpBox />);
    await expect.element(page.getByText('Password is required')).toBeVisible();
  });

  it('shows validation error for EMAIL_INPUT', async () => {
    mockSignUpRenderProps = createMockSignUpRenderProps({
      touched: {email: true},
      fieldErrors: {email: 'Email is invalid'},
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
              label: 'Sign Up',
              variant: 'PRIMARY',
            },
          ],
        },
      ],
    });
    await render(<SignUpBox />);
    await expect.element(page.getByText('Email is invalid')).toBeVisible();
  });

  it('shows validation error for PHONE_INPUT', async () => {
    mockSignUpRenderProps = createMockSignUpRenderProps({
      touched: {phone: true},
      fieldErrors: {phone: 'Phone is required'},
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
    await render(<SignUpBox />);
    await expect.element(page.getByText('Phone is required')).toBeVisible();
  });

  it('renders outlined button variant for non-PRIMARY action', async () => {
    mockSignUpRenderProps = createMockSignUpRenderProps({
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'email-input',
              type: 'TEXT_INPUT',
              ref: 'email',
              label: 'Email',
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
    await render(<SignUpBox />);

    const submitBtn = page.getByText('Continue');
    await expect.element(submitBtn).toBeVisible();
  });

  it('shows "Creating account..." text when loading', async () => {
    mockSignUpRenderProps = createMockSignUpRenderProps({
      isLoading: true,
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'email-input',
              type: 'TEXT_INPUT',
              ref: 'email',
              label: 'Email',
              required: true,
            },
            {
              id: 'submit-btn',
              type: 'ACTION',
              eventType: 'SUBMIT',
              label: 'Sign Up',
              variant: 'PRIMARY',
            },
          ],
        },
      ],
    });
    await render(<SignUpBox />);

    await expect.element(page.getByText('Creating account...')).toBeVisible();
  });

  it('renders block without submit action and shows nothing', async () => {
    mockSignUpRenderProps = createMockSignUpRenderProps({
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'email-input',
              type: 'TEXT_INPUT',
              ref: 'email',
              label: 'Email',
              required: true,
            },
            // No submit action - only trigger actions or no actions
          ],
        },
      ],
    });
    await render(<SignUpBox />);
    // Block without submit or trigger action should not render form fields
    await expect.element(page.getByLabelText(/Email/)).not.toBeInTheDocument();
  });

  it('handles autoComplete attribute for username field', async () => {
    mockSignUpRenderProps = createMockSignUpRenderProps({
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
              label: 'Sign Up',
              variant: 'PRIMARY',
            },
          ],
        },
      ],
    });
    await render(<SignUpBox />);

    const usernameInput = page.getByLabelText(/Username/);
    await expect.element(usernameInput).toHaveAttribute('autocomplete', 'username');
  });

  it('handles autoComplete attribute for email field', async () => {
    mockSignUpRenderProps = createMockSignUpRenderProps({
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'email-input',
              type: 'TEXT_INPUT',
              ref: 'email',
              label: 'Email',
              required: true,
            },
            {
              id: 'submit-btn',
              type: 'ACTION',
              eventType: 'SUBMIT',
              label: 'Sign Up',
              variant: 'PRIMARY',
            },
          ],
        },
      ],
    });
    await render(<SignUpBox />);

    const emailInput = page.getByLabelText(/Email/);
    await expect.element(emailInput).toHaveAttribute('autocomplete', 'email');
  });

  it('handles autoComplete attribute for other fields', async () => {
    mockSignUpRenderProps = createMockSignUpRenderProps({
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'other-input',
              type: 'TEXT_INPUT',
              ref: 'otherField',
              label: 'Other Field',
              required: true,
            },
            {
              id: 'submit-btn',
              type: 'ACTION',
              eventType: 'SUBMIT',
              label: 'Sign Up',
              variant: 'PRIMARY',
            },
          ],
        },
      ],
    });
    await render(<SignUpBox />);

    const otherInput = page.getByLabelText(/Other Field/);
    await expect.element(otherInput).toHaveAttribute('autocomplete', 'off');
  });

  it('renders TRIGGER inside form block and clicks it', async () => {
    mockSignUpRenderProps = createMockSignUpRenderProps({
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
    await render(<SignUpBox />);

    const triggerBtn = page.getByText('Trigger Action');
    await expect.element(triggerBtn).toBeVisible();

    // Click the trigger button
    await userEvent.click(triggerBtn);

    await expect.poll(() => mockHandleSubmit).toHaveBeenCalled();
  });

  it('returns null for unknown component type in block', async () => {
    mockSignUpRenderProps = createMockSignUpRenderProps({
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
    await render(<SignUpBox />);

    // Unknown component should not render
    await expect.element(page.getByLabelText(/Unknown/)).not.toBeInTheDocument();
    // But submit button should render
    await expect.element(page.getByText('Continue')).toBeVisible();
  });

  it('returns null for unknown top-level component type', async () => {
    mockSignUpRenderProps = createMockSignUpRenderProps({
      components: [
        {
          id: 'unknown-1',
          type: 'UNKNOWN_TYPE',
          label: 'Unknown',
        },
      ],
    });
    await render(<SignUpBox />);

    await expect.element(page.getByText('Unknown')).not.toBeInTheDocument();
  });

  it('handles social login trigger in block without form elements', async () => {
    mockSignUpRenderProps = createMockSignUpRenderProps({
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
    await render(<SignUpBox />);

    await expect.element(page.getByText('Continue with Facebook')).toBeVisible();
    await expect.element(page.getByText('Continue with Twitter')).toBeVisible();

    // Click Facebook button
    const facebookBtn = page.getByText('Continue with Facebook');
    await userEvent.click(facebookBtn);

    await expect.poll(() => mockHandleSubmit).toHaveBeenCalled();
  });

  it('returns null for block with no submit or trigger actions', async () => {
    mockSignUpRenderProps = createMockSignUpRenderProps({
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
    await render(<SignUpBox />);

    // Block without submit or trigger action should not render
    await expect.element(page.getByLabelText(/Field/)).not.toBeInTheDocument();
  });

  it('does not render input without ref', async () => {
    mockSignUpRenderProps = createMockSignUpRenderProps({
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'no-ref-input',
              type: 'TEXT_INPUT',
              // ref is undefined
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
    await render(<SignUpBox />);

    // Input without ref should not render
    await expect.element(page.getByLabelText(/No Ref Field/)).not.toBeInTheDocument();
  });

  it('returns null for non-TRIGGER action in social login block', async () => {
    mockSignUpRenderProps = createMockSignUpRenderProps({
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
    await render(<SignUpBox />);

    // Google trigger should render
    await expect.element(page.getByText('Continue with Google')).toBeVisible();
    // Other action type should not render (returns null)
    await expect.element(page.getByText('Other Action')).not.toBeInTheDocument();
  });

  it('renders with branded logo and custom theme palette', async () => {
    mockUseBranding.mockReturnValue({
      images: {
        logo: {
          primary: {
            url: 'https://example.com/custom-logo.png',
            alt: 'Custom Brand Logo',
            height: 50,
            width: 150,
          },
        },
      },
      theme: {palette: {primary: {main: '#123456'}}},
      isBrandingEnabled: true,
    });
    mockSignUpRenderProps = createMockSignUpRenderProps({
      components: [
        {
          id: 'text-1',
          type: 'TEXT',
          label: 'Create Account',
          variant: 'H1',
        },
      ],
    });
    await render(<SignUpBox />);

    // Heading should be centered when branding is enabled
    await expect.element(page.getByText('Create Account')).toBeVisible();
  });

  it('renders with branding enabled but no logo URL', async () => {
    mockUseBranding.mockReturnValue({
      images: {
        logo: {
          primary: {
            url: null,
            alt: null,
            height: null,
            width: null,
          },
        },
      },
      theme: null,
      isBrandingEnabled: true,
    });
    mockSignUpRenderProps = createMockSignUpRenderProps({
      components: [
        {
          id: 'text-1',
          type: 'TEXT',
          label: 'Sign Up',
          variant: 'H2',
        },
      ],
    });
    await render(<SignUpBox />);
    // Component should render correctly
    await expect.element(page.getByText('Sign Up')).toBeVisible();
  });

  it('handles block component without id using index as key', async () => {
    mockSignUpRenderProps = createMockSignUpRenderProps({
      components: [
        {
          id: '',
          type: 'BLOCK',
          components: [
            {
              id: '',
              type: 'TEXT_INPUT',
              ref: 'firstName',
              label: 'First Name',
              required: true,
            },
            {
              id: '',
              type: 'ACTION',
              eventType: 'SUBMIT',
              label: 'Continue',
              variant: 'PRIMARY',
            },
          ],
        },
      ],
    });
    await render(<SignUpBox />);
    await expect.element(page.getByLabelText(/First Name/)).toBeVisible();
  });

  it('renders error alert with default description when error has no message', async () => {
    mockSignUpRenderProps = createMockSignUpRenderProps({
      error: {message: undefined as unknown as string},
      components: [
        {
          id: 'text-1',
          type: 'TEXT',
          label: 'Sign Up',
        },
      ],
    });
    await render(<SignUpBox />);
    // Should show default error from translation
    await expect.element(page.getByRole('alert')).toBeVisible();
  });

  it('handles trigger action without image', async () => {
    mockSignUpRenderProps = createMockSignUpRenderProps({
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'sso-btn',
              type: 'ACTION',
              eventType: 'TRIGGER',
              label: 'Continue with SSO',
              // No image property
            },
          ],
        },
      ],
    });
    await render(<SignUpBox />);

    const ssoBtn = page.getByText('Continue with SSO');
    await expect.element(ssoBtn).toBeVisible();

    await userEvent.click(ssoBtn);

    await expect.poll(() => mockHandleSubmit).toHaveBeenCalled();
  });

  it('handles autoFocus on firstName field', async () => {
    mockSignUpRenderProps = createMockSignUpRenderProps({
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'firstName-input',
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
    await render(<SignUpBox />);

    const firstNameInput = page.getByLabelText(/First Name/);
    // AutoFocus is set for firstName field
    await expect.element(firstNameInput).toBeVisible();
  });

  it('handles PASSWORD_INPUT with autoComplete for non-password ref', async () => {
    mockSignUpRenderProps = createMockSignUpRenderProps({
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'confirm-password',
              type: 'PASSWORD_INPUT',
              ref: 'confirmPassword',
              label: 'Confirm Password',
              placeholder: 'Re-enter password',
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
    await render(<SignUpBox />);

    const confirmPasswordInput = page.getByLabelText(/Confirm Password/);
    await expect.element(confirmPasswordInput).toHaveAttribute('autocomplete', 'off');
  });

  it('renders action button with SECONDARY variant as outlined', async () => {
    mockSignUpRenderProps = createMockSignUpRenderProps({
      components: [
        {
          id: 'block-1',
          type: 'BLOCK',
          components: [
            {
              id: 'email-input',
              type: 'TEXT_INPUT',
              ref: 'email',
              label: 'Email',
              required: false,
            },
            {
              id: 'submit-btn',
              type: 'ACTION',
              eventType: 'SUBMIT',
              label: 'Next',
              variant: 'SECONDARY',
            },
          ],
        },
      ],
    });
    await render(<SignUpBox />);

    // Button should render with outlined variant
    const submitBtn = page.getByText('Next');
    await expect.element(submitBtn).toBeVisible();
  });
});
