/**
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
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

import {describe, it, expect} from 'vitest';
import {screen} from '@testing-library/react';
import render from '@/test/test-utils';
import ApplicationsPage from './ApplicationsPage';

describe('ApplicationsPage', () => {
  it('renders page title', () => {
    render(<ApplicationsPage />);

    expect(screen.getByText('Applications')).toBeInTheDocument();
  });

  it('renders page description', () => {
    render(<ApplicationsPage />);

    expect(screen.getByText('Manage your applications and services')).toBeInTheDocument();
  });

  it('displays coming soon message', () => {
    render(<ApplicationsPage />);

    expect(screen.getByText('Coming Soon')).toBeInTheDocument();
  });

  it('shows functionality message', () => {
    render(<ApplicationsPage />);

    expect(screen.getByText('Applications management functionality will be available soon.')).toBeInTheDocument();
  });

  it('renders with correct title hierarchy', () => {
    render(<ApplicationsPage />);

    const title = screen.getByText('Applications');
    expect(title.tagName).toBe('H1');
  });

  it('renders coming soon with correct heading level', () => {
    render(<ApplicationsPage />);

    const comingSoon = screen.getByText('Coming Soon');
    expect(comingSoon.tagName).toBe('H2');
  });
});
