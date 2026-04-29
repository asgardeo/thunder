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

/* eslint-disable @typescript-eslint/no-unsafe-assignment */
import {screen} from '@testing-library/react';
import {TEST_CN_PREFIX} from '@thunder/test-utils';
import {describe, it, expect, vi} from 'vitest';
import type {FlowFieldProps} from '../../../../models/flow';
import renderWithProviders from '../../../../test/renderWithProviders';
import TextInputAdapter from '../TextInputAdapter';

const baseProps: FlowFieldProps = {
  component: {id: 'username', type: 'TEXT_INPUT', ref: 'username', label: 'Username', required: true},
  values: {},
  isLoading: false,
  resolve: (s) => s,
  onInputChange: vi.fn(),
};

describe('TextInputAdapter', () => {
  it('renders the label', () => {
    renderWithProviders(<TextInputAdapter {...baseProps} />);
    expect(screen.getByText('Username')).toBeTruthy();
  });

  it('renders a text input', () => {
    renderWithProviders(<TextInputAdapter {...baseProps} />);
    const input = document.querySelector('input[name="username"]');
    expect(input).toBeTruthy();
    expect(input?.getAttribute('type')).toBe('text');
  });

  it('renders email type for EMAIL_INPUT', () => {
    const emailProps = {
      ...baseProps,
      component: {...baseProps.component, type: 'EMAIL_INPUT', ref: 'email', label: 'Email'},
    };
    renderWithProviders(<TextInputAdapter {...emailProps} />);
    const input = document.querySelector('input[name="email"]');
    expect(input?.getAttribute('type')).toBe('email');
  });

  it('applies product prefix CSS class names', () => {
    renderWithProviders(<TextInputAdapter {...baseProps} />);
    const formControl = document.querySelector(`.${TEST_CN_PREFIX}Flow--textInput`);
    expect(formControl).toBeTruthy();
    expect(formControl?.classList.contains(`${TEST_CN_PREFIX}FormControl--root`)).toBe(true);
  });

  it('returns null when ref is missing', () => {
    const noRefProps = {...baseProps, component: {...baseProps.component, ref: undefined}};
    const {container} = renderWithProviders(<TextInputAdapter {...noRefProps} />);
    expect(container.innerHTML).toBe('');
  });
});
