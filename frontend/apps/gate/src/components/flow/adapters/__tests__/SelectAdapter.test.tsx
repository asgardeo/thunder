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
import SelectAdapter from '../SelectAdapter';
import type {FlowFieldProps} from '../../../../models/flow';

vi.mock('@wso2/oxygen-ui', () => ({
  FormControl: ({children}: any) => <div>{children}</div>,
  FormLabel: ({children, htmlFor}: any) => <label htmlFor={htmlFor}>{children}</label>,
  Select: ({children, value, onChange, id, name, disabled, error}: any) => (
    <select
      data-testid="select-input"
      id={id}
      name={name}
      value={value}
      onChange={onChange}
      disabled={disabled}
      data-error={String(Boolean(error))}
    >
      {children}
    </select>
  ),
  MenuItem: ({children, value, disabled}: any) => (
    <option value={value} disabled={disabled}>
      {children}
    </option>
  ),
  Typography: ({children, variant, color}: any) => (
    <p data-testid="typography" data-variant={variant} data-color={color}>
      {children}
    </p>
  ),
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({t: (key: string) => key}),
}));

const mockOnInputChange = vi.fn();

const baseProps: FlowFieldProps = {
  component: {
    id: 'sel-1',
    type: 'SELECT',
    label: 'Country',
    ref: 'country',
    options: ['Canada', 'Mexico', 'USA'],
  },
  values: {country: ''},
  isLoading: false,
  resolve: (s) => s,
  onInputChange: mockOnInputChange,
};

describe('SelectAdapter', () => {
  beforeEach(() => {
    mockOnInputChange.mockClear();
  });

  it('returns null when ref is missing', () => {
    const props = {...baseProps, component: {...baseProps.component, ref: undefined, options: ['A']}};
    const {container} = render(<SelectAdapter {...props} />);
    expect(container.firstChild).toBeNull();
  });

  it('returns null when options are missing', () => {
    const props = {...baseProps, component: {...baseProps.component, options: undefined}};
    const {container} = render(<SelectAdapter {...props} />);
    expect(container.firstChild).toBeNull();
  });

  it('renders a select element', () => {
    render(<SelectAdapter {...baseProps} />);
    expect(screen.getByTestId('select-input')).toBeInTheDocument();
  });

  it('renders the label', () => {
    render(<SelectAdapter {...baseProps} />);
    expect(screen.getByText('Country')).toBeInTheDocument();
  });

  it('renders string options as option elements', () => {
    render(<SelectAdapter {...baseProps} />);
    expect(screen.getByRole('option', {name: 'Canada'})).toBeInTheDocument();
    expect(screen.getByRole('option', {name: 'Mexico'})).toBeInTheDocument();
    expect(screen.getByRole('option', {name: 'USA'})).toBeInTheDocument();
  });

  it('renders object options with value and label', () => {
    const props = {
      ...baseProps,
      component: {
        ...baseProps.component,
        options: [
          {value: 'ca', label: 'Canada'},
          {value: 'us', label: 'United States'},
        ],
      },
    };
    render(<SelectAdapter {...props} />);
    expect(screen.getByRole('option', {name: 'Canada'})).toBeInTheDocument();
    expect(screen.getByRole('option', {name: 'United States'})).toBeInTheDocument();
  });

  it('calls onInputChange when an option is selected', () => {
    render(<SelectAdapter {...baseProps} />);
    fireEvent.change(screen.getByTestId('select-input'), {target: {value: 'USA'}});
    expect(mockOnInputChange).toHaveBeenCalledWith('country', 'USA');
  });

  it('displays the current value', () => {
    const props = {...baseProps, values: {country: 'Canada'}};
    render(<SelectAdapter {...props} />);
    expect(screen.getByTestId('select-input')).toHaveValue('Canada');
  });

  it('shows error state when field is touched and has an error', () => {
    const props = {
      ...baseProps,
      touched: {country: true},
      fieldErrors: {country: 'Please select a country'},
    };
    render(<SelectAdapter {...props} />);
    expect(screen.getByTestId('select-input')).toHaveAttribute('data-error', 'true');
    expect(screen.getByText('Please select a country')).toBeInTheDocument();
  });

  it('does not show error when field is not touched', () => {
    const props = {
      ...baseProps,
      touched: {country: false},
      fieldErrors: {country: 'Please select a country'},
    };
    render(<SelectAdapter {...props} />);
    expect(screen.getByTestId('select-input')).toHaveAttribute('data-error', 'false');
  });

  it('disables the select when isLoading is true', () => {
    render(<SelectAdapter {...baseProps} isLoading />);
    expect(screen.getByTestId('select-input')).toBeDisabled();
  });

  it('renders a hint when provided', () => {
    const props = {
      ...baseProps,
      component: {...baseProps.component, hint: 'Select your country of residence'},
    };
    render(<SelectAdapter {...props} />);
    expect(screen.getByText('Select your country of residence')).toBeInTheDocument();
  });
});
