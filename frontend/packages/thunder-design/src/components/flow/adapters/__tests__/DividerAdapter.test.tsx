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
import {describe, it, expect} from 'vitest';
import type {FlowComponent} from '../../../../models/flow';
import renderWithProviders from '../../../../test/renderWithProviders';
import DividerAdapter from '../DividerAdapter';

describe('DividerAdapter', () => {
  it('renders a divider element', () => {
    const component: FlowComponent = {id: 'div-1', type: 'DIVIDER'};
    renderWithProviders(<DividerAdapter component={component} resolve={(s) => s} />);
    const divider = document.querySelector(`.${TEST_CN_PREFIX}Flow--divider`);
    expect(divider).toBeTruthy();
  });

  it('renders with label text', () => {
    const component: FlowComponent = {id: 'div-1', type: 'DIVIDER', label: 'OR'};
    renderWithProviders(<DividerAdapter component={component} resolve={(s) => s} />);
    expect(screen.getByText('OR')).toBeTruthy();
  });

  it('applies product prefix CSS class names', () => {
    const component: FlowComponent = {id: 'div-1', type: 'DIVIDER'};
    renderWithProviders(<DividerAdapter component={component} resolve={(s) => s} />);
    const divider = document.querySelector(`.${TEST_CN_PREFIX}Divider--root`);
    expect(divider).toBeTruthy();
  });
});
