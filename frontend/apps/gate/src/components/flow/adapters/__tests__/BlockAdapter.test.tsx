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

/* eslint-disable @typescript-eslint/no-explicit-any */
/* eslint-disable @typescript-eslint/no-unsafe-assignment */
/* eslint-disable @typescript-eslint/no-unsafe-member-access */

import {describe, it, expect, vi, beforeEach} from 'vitest';
import {render, screen, fireEvent} from '@testing-library/react';
import BlockAdapter from '../BlockAdapter';

vi.mock('@wso2/oxygen-ui', () => ({
  Box: ({children, component: Comp, onSubmit, noValidate, sx}: any) => {
    if (Comp === 'form') {
      return (
        <form data-testid="block-form" onSubmit={onSubmit} noValidate={noValidate} style={sx}>
          {children}
        </form>
      );
    }
    return (
      <div data-testid="trigger-block" style={sx}>
        {children}
      </div>
    );
  },
  Button: ({children, onClick, disabled, variant, type, startIcon}: any) => (
    <button
      data-testid={type === 'submit' ? 'submit-button' : 'trigger-button'}
      type={type === 'submit' ? 'submit' : 'button'}
      onClick={onClick}
      disabled={disabled}
      data-variant={variant}
    >
      {startIcon}
      {children}
    </button>
  ),
}));

vi.mock('@asgardeo/react', () => ({
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

vi.mock('react-i18next', () => ({
  useTranslation: () => ({t: (key: string) => key}),
}));

vi.mock('../../../utils/getIntegrationIcon', () => ({
  default: () => null,
}));

// Mock sub-adapters so their own dependencies don't need to be resolved
vi.mock('../TextInputAdapter', () => ({
  default: ({component}: any) => <div data-testid="text-input-adapter" data-ref={component.ref} />,
}));
vi.mock('../PasswordInputAdapter', () => ({
  default: ({component}: any) => <div data-testid="password-input-adapter" data-ref={component.ref} />,
}));
vi.mock('../OtpInputAdapter', () => ({
  default: ({component}: any) => <div data-testid="otp-input-adapter" data-ref={component.ref} />,
}));
vi.mock('../SelectAdapter', () => ({
  default: ({component}: any) => <div data-testid="select-adapter" data-ref={component.ref} />,
}));

const mockOnSubmit = vi.fn();
const mockOnInputChange = vi.fn();
const mockOnValidate = vi.fn(() => true);

const submitAction = {id: 'action-submit', type: 'ACTION', eventType: 'SUBMIT', label: 'Sign In'};
const primaryConsentAction = {
  id: 'action-allow',
  type: 'ACTION',
  eventType: 'SUBMIT',
  label: 'Allow',
  variant: 'PRIMARY',
};
const secondaryConsentAction = {
  id: 'action-deny',
  type: 'ACTION',
  eventType: 'SUBMIT',
  label: 'Deny',
  variant: 'SECONDARY',
};
const triggerAction = {id: 'action-trigger', type: 'ACTION', eventType: 'TRIGGER', label: 'Sign in with Google'};
const textInput = {id: 'inp-1', type: 'TEXT_INPUT', label: 'Username', ref: 'username'};
const passwordInput = {id: 'inp-2', type: 'PASSWORD_INPUT', label: 'Password', ref: 'password'};

const baseProps = {
  index: 0,
  values: {username: 'alice', password: 'secret'},
  isLoading: false,
  resolve: (s: string | undefined) => s,
  onInputChange: mockOnInputChange,
  onSubmit: mockOnSubmit,
};

describe('BlockAdapter', () => {
  beforeEach(() => {
    mockOnSubmit.mockClear();
    mockOnInputChange.mockClear();
    mockOnValidate.mockClear().mockReturnValue(true);
  });

  describe('Form Block (has submit action)', () => {
    const formBlock = {
      id: 'block-1',
      type: 'BLOCK',
      components: [textInput, passwordInput, submitAction],
    };

    it('renders a form element', () => {
      render(<BlockAdapter {...baseProps} component={formBlock} />);
      expect(screen.getByTestId('block-form')).toBeInTheDocument();
    });

    it('renders TextInputAdapter for TEXT_INPUT sub-components', () => {
      render(<BlockAdapter {...baseProps} component={formBlock} />);
      expect(screen.getByTestId('text-input-adapter')).toBeInTheDocument();
    });

    it('renders PasswordInputAdapter for PASSWORD_INPUT sub-components', () => {
      render(<BlockAdapter {...baseProps} component={formBlock} />);
      expect(screen.getByTestId('password-input-adapter')).toBeInTheDocument();
    });

    it('renders a submit button for the ACTION/SUBMIT sub-component', () => {
      render(<BlockAdapter {...baseProps} component={formBlock} />);
      expect(screen.getByTestId('submit-button')).toBeInTheDocument();
    });

    it('calls onSubmit with the submit action and values when form is submitted', () => {
      render(<BlockAdapter {...baseProps} component={formBlock} />);
      fireEvent.submit(screen.getByTestId('block-form'));
      expect(mockOnSubmit).toHaveBeenCalledWith(submitAction, {username: 'alice', password: 'secret'});
    });

    it('calls onValidate before submitting and aborts if it returns false', () => {
      mockOnValidate.mockReturnValue(false);
      render(<BlockAdapter {...baseProps} component={formBlock} onValidate={mockOnValidate} />);
      fireEvent.submit(screen.getByTestId('block-form'));
      expect(mockOnSubmit).not.toHaveBeenCalled();
    });

    it('proceeds with submit when onValidate returns true', () => {
      render(<BlockAdapter {...baseProps} component={formBlock} onValidate={mockOnValidate} />);
      fireEvent.submit(screen.getByTestId('block-form'));
      expect(mockOnSubmit).toHaveBeenCalled();
    });

    it('disables the submit button when isLoading is true', () => {
      render(<BlockAdapter {...baseProps} component={formBlock} isLoading />);
      expect(screen.getByTestId('submit-button')).toBeDisabled();
    });
  });

  describe('Trigger Block (has trigger actions only)', () => {
    const triggerBlock = {
      id: 'block-2',
      type: 'BLOCK',
      components: [triggerAction],
    };

    it('renders a div container (not a form)', () => {
      render(<BlockAdapter {...baseProps} component={triggerBlock} />);
      expect(screen.getByTestId('trigger-block')).toBeInTheDocument();
      expect(screen.queryByTestId('block-form')).toBeNull();
    });

    it('renders a trigger button for each ACTION/TRIGGER sub-component', () => {
      render(<BlockAdapter {...baseProps} component={triggerBlock} />);
      expect(screen.getByTestId('trigger-button')).toBeInTheDocument();
    });

    it('calls onSubmit with the trigger action and values when trigger button is clicked', () => {
      render(<BlockAdapter {...baseProps} component={triggerBlock} />);
      fireEvent.click(screen.getByTestId('trigger-button'));
      expect(mockOnSubmit).toHaveBeenCalledWith(triggerAction, {username: 'alice', password: 'secret'});
    });

    it('disables trigger buttons when isLoading is true', () => {
      render(<BlockAdapter {...baseProps} component={triggerBlock} isLoading />);
      expect(screen.getByTestId('trigger-button')).toBeDisabled();
    });
  });

  describe('Empty Block', () => {
    it('returns null when block has no submit or trigger actions', () => {
      const emptyBlock = {
        id: 'block-3',
        type: 'BLOCK',
        components: [textInput],
      };
      const {container} = render(<BlockAdapter {...baseProps} component={emptyBlock} />);
      expect(container.firstChild).toBeNull();
    });

    it('returns null when block has no components', () => {
      const emptyBlock = {id: 'block-4', type: 'BLOCK', components: []};
      const {container} = render(<BlockAdapter {...baseProps} component={emptyBlock} />);
      expect(container.firstChild).toBeNull();
    });
  });

  describe('RESEND sub-component', () => {
    it('renders a resend button inside a form block when RESEND/SUBMIT is present', () => {
      const resendAction = {id: 'resend-1', type: 'RESEND', eventType: 'SUBMIT', label: 'Resend Code'};
      const resendBlock = {
        id: 'block-5',
        type: 'BLOCK',
        components: [resendAction],
      };
      render(<BlockAdapter {...baseProps} component={resendBlock} />);
      expect(screen.getByTestId('block-form')).toBeInTheDocument();
      expect(screen.getByTestId('submit-button')).toBeInTheDocument();
    });
  });

  describe('Multiple submit actions (consent-style)', () => {
    const multiSubmitBlock = {
      id: 'block-consent-submit',
      type: 'BLOCK',
      components: [textInput, primaryConsentAction, secondaryConsentAction],
    };

    it('keeps PRIMARY action as type=submit and SECONDARY as type=button', () => {
      render(<BlockAdapter {...baseProps} component={multiSubmitBlock} />);

      expect(screen.getByRole('button', {name: 'Allow'})).toHaveAttribute('type', 'submit');
      expect(screen.getByRole('button', {name: 'Deny'})).toHaveAttribute('type', 'button');
    });

    it('submits the PRIMARY action via form submit (Enter-key/default submit path)', () => {
      render(<BlockAdapter {...baseProps} component={multiSubmitBlock} onValidate={mockOnValidate} />);

      fireEvent.submit(screen.getByTestId('block-form'));

      expect(mockOnValidate).toHaveBeenCalled();
      expect(mockOnSubmit).toHaveBeenCalledWith(primaryConsentAction, {username: 'alice', password: 'secret'});
    });

    it('submits SECONDARY action through its own click handler', () => {
      render(<BlockAdapter {...baseProps} component={multiSubmitBlock} onValidate={mockOnValidate} />);

      fireEvent.click(screen.getByRole('button', {name: 'Deny'}));

      expect(mockOnValidate).toHaveBeenCalled();
      expect(mockOnSubmit).toHaveBeenCalledWith(secondaryConsentAction, {username: 'alice', password: 'secret'});
    });

    it('blocks both form-submit and secondary click when validation fails', () => {
      mockOnValidate.mockReturnValue(false);
      render(<BlockAdapter {...baseProps} component={multiSubmitBlock} onValidate={mockOnValidate} />);

      fireEvent.submit(screen.getByTestId('block-form'));
      fireEvent.click(screen.getByRole('button', {name: 'Deny'}));

      expect(mockOnSubmit).not.toHaveBeenCalled();
    });
  });
});
