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
import OtpInputAdapter from '../OtpInputAdapter';
import type {FlowFieldProps} from '../../../../models/flow';

vi.mock('@wso2/oxygen-ui', () => ({
  Box: ({children, sx}: any) => <div style={sx}>{children}</div>,
  FormControl: ({children}: any) => <div>{children}</div>,
  FormLabel: ({children, htmlFor}: any) => <label htmlFor={htmlFor}>{children}</label>,
  TextField: ({onChange, value, error, disabled, slotProps, onKeyDown, onPaste}: any) => (
    <input
      type="text"
      data-testid="otp-digit"
      value={value}
      onChange={onChange}
      onKeyDown={onKeyDown}
      onPaste={onPaste}
      disabled={disabled}
      aria-label={slotProps?.htmlInput?.['aria-label']}
      maxLength={slotProps?.htmlInput?.maxLength}
      data-error={String(Boolean(error))}
    />
  ),
  Typography: ({children, variant, color}: any) => (
    <p data-testid="otp-error" data-variant={variant} data-color={color}>
      {children}
    </p>
  ),
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({t: (key: string) => key}),
}));

const mockOnInputChange = vi.fn();

const baseProps: FlowFieldProps = {
  component: {id: 'otp-1', type: 'OTP_INPUT', label: 'Enter OTP', ref: 'otp'},
  values: {otp: ''},
  isLoading: false,
  resolve: (s) => s,
  onInputChange: mockOnInputChange,
};

describe('OtpInputAdapter', () => {
  beforeEach(() => {
    mockOnInputChange.mockClear();
  });

  it('returns null when ref is missing', () => {
    const props = {...baseProps, component: {...baseProps.component, ref: undefined}};
    const {container} = render(<OtpInputAdapter {...props} />);
    expect(container.firstChild).toBeNull();
  });

  it('renders 6 digit input fields', () => {
    render(<OtpInputAdapter {...baseProps} />);
    expect(screen.getAllByTestId('otp-digit')).toHaveLength(6);
  });

  it('labels each digit with its position for accessibility', () => {
    render(<OtpInputAdapter {...baseProps} />);
    const inputs = screen.getAllByTestId('otp-digit');
    inputs.forEach((input, idx) => {
      expect(input).toHaveAttribute('aria-label', `OTP digit ${idx + 1}`);
    });
  });

  it('renders the label', () => {
    render(<OtpInputAdapter {...baseProps} />);
    expect(screen.getByText('Enter OTP')).toBeInTheDocument();
  });

  it('calls onInputChange when a digit is typed', () => {
    render(<OtpInputAdapter {...baseProps} />);
    const inputs = screen.getAllByTestId('otp-digit');
    fireEvent.change(inputs[0], {target: {value: '5'}});
    expect(mockOnInputChange).toHaveBeenCalledWith('otp', '5     ');
  });

  it('populates digit values from current OTP value', () => {
    const props = {...baseProps, values: {otp: '123456'}};
    render(<OtpInputAdapter {...props} />);
    const inputs = screen.getAllByTestId('otp-digit');
    expect(inputs[0]).toHaveValue('1');
    expect(inputs[1]).toHaveValue('2');
    expect(inputs[2]).toHaveValue('3');
  });

  it('ignores non-digit input', () => {
    render(<OtpInputAdapter {...baseProps} />);
    const inputs = screen.getAllByTestId('otp-digit');
    fireEvent.change(inputs[0], {target: {value: 'a'}});
    expect(mockOnInputChange).not.toHaveBeenCalled();
  });

  it('handles paste by stripping non-digits and trimming to 6 chars', () => {
    render(<OtpInputAdapter {...baseProps} />);
    const inputs = screen.getAllByTestId('otp-digit');
    fireEvent.paste(inputs[0], {
      clipboardData: {getData: () => '123abc456'},
    });
    expect(mockOnInputChange).toHaveBeenCalledWith('otp', '123456');
  });

  it('handles paste truncating to 6 digits', () => {
    render(<OtpInputAdapter {...baseProps} />);
    const inputs = screen.getAllByTestId('otp-digit');
    fireEvent.paste(inputs[0], {
      clipboardData: {getData: () => '1234567890'},
    });
    expect(mockOnInputChange).toHaveBeenCalledWith('otp', '123456');
  });

  it('shows error state when field is touched and has an error', () => {
    const props = {
      ...baseProps,
      touched: {otp: true},
      fieldErrors: {otp: 'Invalid OTP'},
    };
    render(<OtpInputAdapter {...props} />);
    const inputs = screen.getAllByTestId('otp-digit');
    inputs.forEach((input) => {
      expect(input).toHaveAttribute('data-error', 'true');
    });
    expect(screen.getByTestId('otp-error')).toHaveTextContent('Invalid OTP');
  });

  it('does not show error text when field is not touched', () => {
    const props = {
      ...baseProps,
      touched: {otp: false},
      fieldErrors: {otp: 'Invalid OTP'},
    };
    render(<OtpInputAdapter {...props} />);
    expect(screen.queryByTestId('otp-error')).toBeNull();
  });

  it('disables all inputs when isLoading is true', () => {
    render(<OtpInputAdapter {...baseProps} isLoading />);
    screen.getAllByTestId('otp-digit').forEach((input) => {
      expect(input).toBeDisabled();
    });
  });
});
