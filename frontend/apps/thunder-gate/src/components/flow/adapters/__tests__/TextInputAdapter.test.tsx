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

import {describe, it, expect, vi, beforeEach} from 'vitest';
import {render, screen, fireEvent} from '@testing-library/react';
import TextInputAdapter from '../TextInputAdapter';
import type {FlowFieldProps} from '../../../../models/flow';

vi.mock('@wso2/oxygen-ui', () => ({
  FormControl: ({children, required}: any) => <div data-required={required}>{children}</div>,
  FormLabel: ({children, htmlFor}: any) => <label htmlFor={htmlFor}>{children}</label>,
  TextField: ({
    onChange,
    value,
    type,
    id,
    name,
    error,
    helperText,
    disabled,
    placeholder,
    autoComplete,
    autoFocus,
  }: any) => (
    <div>
      <input
        data-testid="text-field-input"
        type={type}
        id={id}
        name={name}
        value={value}
        onChange={onChange}
        disabled={disabled}
        placeholder={placeholder}
        autoComplete={autoComplete}
        // eslint-disable-next-line jsx-a11y/no-autofocus
        autoFocus={autoFocus}
        data-autofocus={autoFocus ? 'true' : undefined}
        data-error={String(Boolean(error))}
        aria-describedby={helperText ? 'helper' : undefined}
      />
      {error && helperText && <span data-testid="helper-text">{helperText}</span>}
    </div>
  ),
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({t: (key: string) => key}),
}));

const mockOnInputChange = vi.fn();

const baseProps: FlowFieldProps = {
  component: {id: 'u1', type: 'TEXT_INPUT', label: 'Username', ref: 'username'},
  values: {username: ''},
  isLoading: false,
  resolve: (s) => s,
  onInputChange: mockOnInputChange,
};

describe('TextInputAdapter', () => {
  beforeEach(() => {
    mockOnInputChange.mockClear();
  });

  it('returns null when ref is missing', () => {
    const props = {...baseProps, component: {...baseProps.component, ref: undefined}};
    const {container} = render(<TextInputAdapter {...props} />);
    expect(container.firstChild).toBeNull();
  });

  it('renders a text input for TEXT_INPUT type', () => {
    render(<TextInputAdapter {...baseProps} />);
    expect(screen.getByTestId('text-field-input')).toHaveAttribute('type', 'text');
  });

  it('renders an email input for EMAIL_INPUT type', () => {
    const props = {...baseProps, component: {...baseProps.component, type: 'EMAIL_INPUT', ref: 'email'}};
    render(<TextInputAdapter {...props} />);
    expect(screen.getByTestId('text-field-input')).toHaveAttribute('type', 'email');
  });

  it('renders a tel input for PHONE_INPUT type', () => {
    const props = {...baseProps, component: {...baseProps.component, type: 'PHONE_INPUT', ref: 'phone'}};
    render(<TextInputAdapter {...props} />);
    expect(screen.getByTestId('text-field-input')).toHaveAttribute('type', 'tel');
  });

  it('calls onInputChange when value changes', () => {
    render(<TextInputAdapter {...baseProps} />);
    fireEvent.change(screen.getByTestId('text-field-input'), {target: {value: 'john'}});
    expect(mockOnInputChange).toHaveBeenCalledWith('username', 'john');
  });

  it('displays the current value', () => {
    const props = {...baseProps, values: {username: 'alice'}};
    render(<TextInputAdapter {...props} />);
    expect(screen.getByTestId('text-field-input')).toHaveValue('alice');
  });

  it('shows an error state when field is touched and has an error', () => {
    const props = {
      ...baseProps,
      touched: {username: true},
      fieldErrors: {username: 'Required field'},
    };
    render(<TextInputAdapter {...props} />);
    expect(screen.getByTestId('text-field-input')).toHaveAttribute('data-error', 'true');
    expect(screen.getByTestId('helper-text')).toHaveTextContent('Required field');
  });

  it('does not show error when field is not touched', () => {
    const props = {
      ...baseProps,
      touched: {username: false},
      fieldErrors: {username: 'Required field'},
    };
    render(<TextInputAdapter {...props} />);
    expect(screen.getByTestId('text-field-input')).toHaveAttribute('data-error', 'false');
  });

  it('disables the input when isLoading is true', () => {
    render(<TextInputAdapter {...baseProps} isLoading />);
    expect(screen.getByTestId('text-field-input')).toBeDisabled();
  });

  it('sets autoFocus for username ref', () => {
    render(<TextInputAdapter {...baseProps} />);
    expect(screen.getByTestId('text-field-input')).toHaveAttribute('data-autofocus', 'true');
  });

  it('does not set autoFocus for non-username ref', () => {
    const props = {...baseProps, component: {...baseProps.component, ref: 'email'}};
    render(<TextInputAdapter {...props} />);
    expect(screen.getByTestId('text-field-input')).not.toHaveAttribute('data-autofocus', 'true');
  });

  it('sets autocomplete to username for username ref', () => {
    render(<TextInputAdapter {...baseProps} />);
    expect(screen.getByTestId('text-field-input')).toHaveAttribute('autoComplete', 'username');
  });

  it('sets autocomplete to email for EMAIL_INPUT type', () => {
    const props = {...baseProps, component: {...baseProps.component, type: 'EMAIL_INPUT', ref: 'email'}};
    render(<TextInputAdapter {...props} />);
    expect(screen.getByTestId('text-field-input')).toHaveAttribute('autoComplete', 'email');
  });

  it('sets autocomplete to tel for PHONE_INPUT type', () => {
    const props = {...baseProps, component: {...baseProps.component, type: 'PHONE_INPUT', ref: 'phone'}};
    render(<TextInputAdapter {...props} />);
    expect(screen.getByTestId('text-field-input')).toHaveAttribute('autoComplete', 'tel');
  });

  it('renders the label', () => {
    render(<TextInputAdapter {...baseProps} />);
    expect(screen.getByText('Username')).toBeInTheDocument();
  });
});
