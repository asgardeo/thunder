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
import {render, screen} from '@thunder/test-utils';
import IntegrationsPage from '../IntegrationsPage';

describe('IntegrationsPage', () => {
  it('renders page title', () => {
    render(<IntegrationsPage />);

    expect(screen.getByText('Integrations')).toBeInTheDocument();
  });

  it('renders page description', () => {
    render(<IntegrationsPage />);

    expect(screen.getByText('Manage your integrations and connections')).toBeInTheDocument();
  });

  it('displays coming soon message', () => {
    render(<IntegrationsPage />);

    expect(screen.getByText('Coming Soon')).toBeInTheDocument();
  });

  it('shows functionality message', () => {
    render(<IntegrationsPage />);

    expect(screen.getByText('Integrations management functionality will be available soon.')).toBeInTheDocument();
  });

  it('renders with correct title hierarchy', () => {
    render(<IntegrationsPage />);

    const title = screen.getByText('Integrations');
    expect(title.tagName).toBe('H1');
  });

  it('renders coming soon with correct heading level', () => {
    render(<IntegrationsPage />);

    const comingSoon = screen.getByText('Coming Soon');
    expect(comingSoon.tagName).toBe('H4');
  });
});
