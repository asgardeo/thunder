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
import PasswordInputAdapter from '../PasswordInputAdapter';
import type {FlowFieldProps} from '../../../../models/flow';

vi.mock('@wso2/oxygen-ui', () => ({
  FormControl: ({children, required}: any) => <div data-required={required}>{children}</div>,
  FormLabel: ({children, htmlFor}: any) => <label htmlFor={htmlFor}>{children}</label>,
  TextField: ({onChange, value, type, id, name, error, helperText, disabled, placeholder, autoComplete, slotProps}: any) => (
    <div>
      <input
        data-testid="password-input"
        type={type}
        id={id}
        name={name}
        value={value}
        onChange={onChange}
        disabled={disabled}
        placeholder={placeholder}
        autoComplete={autoComplete}
        data-error={String(Boolean(error))}
      />
      {slotProps?.input?.endAdornment}
      {error && helperText && <span data-testid="helper-text">{helperText}</span>}
    </div>
  ),
  InputAdornment: ({children}: any) => <span>{children}</span>,
  IconButton: ({children, onClick, disabled, 'aria-label': ariaLabel}: any) => (
    <button
      type="button"
      data-testid="visibility-toggle"
      onClick={onClick}
      disabled={disabled}
      aria-label={ariaLabel}
    >
      {children}
    </button>
  ),
}));

vi.mock('@wso2/oxygen-ui-icons-react', () => ({
  Eye: () => <span data-testid="eye-icon" />,
  EyeClosed: () => <span data-testid="eye-closed-icon" />,
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({t: (key: string) => key}),
}));

const mockOnInputChange = vi.fn();

const baseProps: FlowFieldProps = {
  component: {id: 'pw-1', type: 'PASSWORD_INPUT', label: 'Password', ref: 'password'},
  values: {password: ''},
  isLoading: false,
  resolve: (s) => s,
  onInputChange: mockOnInputChange,
};

describe('PasswordInputAdapter', () => {
  beforeEach(() => {
    mockOnInputChange.mockClear();
  });

  it('returns null when ref is missing', () => {
    const props = {...baseProps, component: {...baseProps.component, ref: undefined}};
    const {container} = render(<PasswordInputAdapter {...props} />);
    expect(container.firstChild).toBeNull();
  });

  it('renders a password input by default', () => {
    render(<PasswordInputAdapter {...baseProps} />);
    expect(screen.getByTestId('password-input')).toHaveAttribute('type', 'password');
  });

  it('renders the EyeClosed icon when password is hidden', () => {
    render(<PasswordInputAdapter {...baseProps} />);
    expect(screen.getByTestId('eye-closed-icon')).toBeInTheDocument();
  });

  it('renders the Eye icon when password is visible', () => {
    render(<PasswordInputAdapter {...baseProps} />);
    fireEvent.click(screen.getByTestId('visibility-toggle'));
    expect(screen.getByTestId('eye-icon')).toBeInTheDocument();
  });

  it('toggles input type from password to text on toggle click', () => {
    render(<PasswordInputAdapter {...baseProps} />);
    const toggle = screen.getByTestId('visibility-toggle');
    fireEvent.click(toggle);
    expect(screen.getByTestId('password-input')).toHaveAttribute('type', 'text');
  });

  it('toggles input type back to password on second click', () => {
    render(<PasswordInputAdapter {...baseProps} />);
    const toggle = screen.getByTestId('visibility-toggle');
    fireEvent.click(toggle);
    fireEvent.click(toggle);
    expect(screen.getByTestId('password-input')).toHaveAttribute('type', 'password');
  });

  it('calls onInputChange when value changes', () => {
    render(<PasswordInputAdapter {...baseProps} />);
    fireEvent.change(screen.getByTestId('password-input'), {target: {value: 'secret123'}});
    expect(mockOnInputChange).toHaveBeenCalledWith('password', 'secret123');
  });

  it('shows error state when field is touched and has an error', () => {
    const props = {
      ...baseProps,
      touched: {password: true},
      fieldErrors: {password: 'Password is required'},
    };
    render(<PasswordInputAdapter {...props} />);
    expect(screen.getByTestId('password-input')).toHaveAttribute('data-error', 'true');
    expect(screen.getByTestId('helper-text')).toHaveTextContent('Password is required');
  });

  it('does not show error when field is not touched', () => {
    const props = {
      ...baseProps,
      touched: {password: false},
      fieldErrors: {password: 'Password is required'},
    };
    render(<PasswordInputAdapter {...props} />);
    expect(screen.getByTestId('password-input')).toHaveAttribute('data-error', 'false');
  });

  it('disables input and toggle when isLoading is true', () => {
    render(<PasswordInputAdapter {...baseProps} isLoading />);
    expect(screen.getByTestId('password-input')).toBeDisabled();
    expect(screen.getByTestId('visibility-toggle')).toBeDisabled();
  });

  it('uses current-password autocomplete by default', () => {
    render(<PasswordInputAdapter {...baseProps} />);
    expect(screen.getByTestId('password-input')).toHaveAttribute('autoComplete', 'current-password');
  });

  it('uses new-password autocomplete when passwordAutoComplete is new-password', () => {
    render(<PasswordInputAdapter {...baseProps} passwordAutoComplete="new-password" />);
    expect(screen.getByTestId('password-input')).toHaveAttribute('autoComplete', 'new-password');
  });

  it('renders the label', () => {
    render(<PasswordInputAdapter {...baseProps} />);
    expect(screen.getByText('Password')).toBeInTheDocument();
  });
});
