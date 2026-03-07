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
import StandaloneTriggerAdapter from '../StandaloneTriggerAdapter';
import type {FlowComponent} from '../../../../models/flow';

vi.mock('@wso2/oxygen-ui', () => ({
  Box: ({component: Comp, src, alt, sx}: any) => {
    if (Comp === 'img') return <img data-testid="trigger-icon-img" src={src} alt={alt} style={sx} />;
    return <div style={sx} />;
  },
  Button: ({children, onClick, disabled, variant, startIcon}: any) => (
    <button type="button" data-testid="trigger-button" onClick={onClick} disabled={disabled} data-variant={variant}>
      {startIcon}
      {children}
    </button>
  ),
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({t: (key: string) => key}),
}));

vi.mock('../../../../utils/getIntegrationIcon', () => ({
  default: (label: string, image: string) => (
    <span data-testid="integration-icon" data-label={label} data-image={image} />
  ),
}));

const mockOnSubmit = vi.fn();

const baseComponent: FlowComponent = {
  id: 'trigger-1',
  type: 'ACTION',
  eventType: 'TRIGGER',
  label: 'Sign in with Google',
  variant: 'OUTLINED',
};

const baseProps = {
  component: baseComponent,
  index: 0,
  isLoading: false,
  resolve: (s: string | undefined) => s,
  onSubmit: mockOnSubmit,
  values: {username: 'alice'},
};

describe('StandaloneTriggerAdapter', () => {
  beforeEach(() => {
    mockOnSubmit.mockClear();
  });

  it('renders a button', () => {
    render(<StandaloneTriggerAdapter {...baseProps} />);
    expect(screen.getByTestId('trigger-button')).toBeInTheDocument();
  });

  it('renders the resolved label', () => {
    render(<StandaloneTriggerAdapter {...baseProps} />);
    expect(screen.getByTestId('trigger-button')).toHaveTextContent('Sign in with Google');
  });

  it('calls onSubmit with the component and values when clicked', () => {
    render(<StandaloneTriggerAdapter {...baseProps} />);
    fireEvent.click(screen.getByTestId('trigger-button'));
    expect(mockOnSubmit).toHaveBeenCalledWith(baseComponent, {username: 'alice'});
  });

  it('disables the button when isLoading is true', () => {
    render(<StandaloneTriggerAdapter {...baseProps} isLoading />);
    expect(screen.getByTestId('trigger-button')).toBeDisabled();
  });

  it('renders with outlined variant when component.variant is OUTLINED', () => {
    render(<StandaloneTriggerAdapter {...baseProps} />);
    expect(screen.getByTestId('trigger-button')).toHaveAttribute('data-variant', 'outlined');
  });

  it('renders with contained variant when component.variant is not OUTLINED', () => {
    const props = {...baseProps, component: {...baseComponent, variant: 'PRIMARY'}};
    render(<StandaloneTriggerAdapter {...props} />);
    expect(screen.getByTestId('trigger-button')).toHaveAttribute('data-variant', 'contained');
  });

  it('renders an img icon when startIcon is a URL', () => {
    const props = {
      ...baseProps,
      component: {...baseComponent, startIcon: 'https://example.com/icon.svg'},
      resolve: () => 'https://example.com/icon.svg',
    };
    render(<StandaloneTriggerAdapter {...props} />);
    expect(screen.getByTestId('trigger-icon-img')).toBeInTheDocument();
    expect(screen.getByTestId('trigger-icon-img')).toHaveAttribute('src', 'https://example.com/icon.svg');
  });

  it('renders integration icon when startIcon is not a URL', () => {
    const props = {
      ...baseProps,
      component: {...baseComponent, startIcon: 'google'},
      resolve: (s: string | undefined) => s,
    };
    render(<StandaloneTriggerAdapter {...props} />);
    expect(screen.getByTestId('integration-icon')).toBeInTheDocument();
  });

  it('falls back to image field for icon resolution when startIcon is absent', () => {
    const props = {
      ...baseProps,
      component: {...baseComponent, startIcon: undefined, image: 'https://example.com/img.png'},
      resolve: () => 'https://example.com/img.png',
    };
    render(<StandaloneTriggerAdapter {...props} />);
    expect(screen.getByTestId('trigger-icon-img')).toBeInTheDocument();
  });
});
