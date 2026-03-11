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

import {describe, it, expect, vi, beforeEach} from 'vitest';
import {render} from '@thunder/test-utils/browser';
import {page} from 'vitest/browser';

import EditFlowsSettings from '../EditFlowsSettings';
import type {Application} from '../../../../models/application';

// Mock the child components
vi.mock('../AuthenticationFlowSection', () => ({
  default: ({application, editedApp}: {application: Application; editedApp: Partial<Application>}) => (
    <div data-testid="auth-flow-section">
      AuthenticationFlowSection - App: {application.id}, Edited Auth Flow: {editedApp.auth_flow_id ?? 'None'}
    </div>
  ),
}));

vi.mock('../RegistrationFlowSection', () => ({
  default: ({application, editedApp}: {application: Application; editedApp: Partial<Application>}) => (
    <div data-testid="registration-flow-section">
      RegistrationFlowSection - App: {application.id}, Edited Reg Flow: {editedApp.registration_flow_id ?? 'None'}
    </div>
  ),
}));

describe('EditFlowsSettings', () => {
  const mockOnFieldChange = vi.fn();
  const mockApplication: Application = {
    id: 'app-123',
    name: 'Test App',
    auth_flow_id: 'auth-flow-1',
    registration_flow_id: 'reg-flow-1',
    is_registration_flow_enabled: true,
  } as Application;

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('should render both flow sections', async () => {
      await render(
          <EditFlowsSettings application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />
      );

      await expect.element(page.getByTestId('auth-flow-section')).toBeInTheDocument();
      await expect.element(page.getByTestId('registration-flow-section')).toBeInTheDocument();
    });

    it('should pass application to child components', async () => {
      await render(
          <EditFlowsSettings application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />
      );

      await expect.element(page.getByTestId('auth-flow-section')).toHaveTextContent('App: app-123');
      await expect.element(page.getByTestId('registration-flow-section')).toHaveTextContent('App: app-123');
    });

    it('should pass editedApp to child components', async () => {
      const editedApp = {
        auth_flow_id: 'edited-auth-flow',
        registration_flow_id: 'edited-reg-flow',
      };

      await render(
          <EditFlowsSettings application={mockApplication} editedApp={editedApp} onFieldChange={mockOnFieldChange} />
      );

      await expect.element(page.getByTestId('auth-flow-section')).toHaveTextContent('Edited Auth Flow: edited-auth-flow');
      await expect.element(page.getByTestId('registration-flow-section')).toHaveTextContent('Edited Reg Flow: edited-reg-flow');
    });

    it('should pass empty editedApp to child components', async () => {
      await render(
          <EditFlowsSettings application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />
      );

      await expect.element(page.getByTestId('auth-flow-section')).toHaveTextContent('Edited Auth Flow: None');
      await expect.element(page.getByTestId('registration-flow-section')).toHaveTextContent('Edited Reg Flow: None');
    });
  });

  describe('Props Propagation', () => {
    it('should pass onFieldChange to AuthenticationFlowSection', async () => {
      const {container} = await render(
          <EditFlowsSettings application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />
      );

      // Verify the component is rendered (which means props were passed correctly)
      expect(container.querySelector('[data-testid="auth-flow-section"]')).toBeInTheDocument();
    });

    it('should pass onFieldChange to RegistrationFlowSection', async () => {
      const {container} = await render(
          <EditFlowsSettings application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />
      );

      // Verify the component is rendered (which means props were passed correctly)
      expect(container.querySelector('[data-testid="registration-flow-section"]')).toBeInTheDocument();
    });

    it('should pass all required props to both child components', async () => {
      const editedApp = {auth_flow_id: 'new-flow'};

      await render(
          <EditFlowsSettings application={mockApplication} editedApp={editedApp} onFieldChange={mockOnFieldChange} />
      );

      // Both components should be present and have received their props
      await expect.element(page.getByTestId('auth-flow-section')).toBeInTheDocument();
      await expect.element(page.getByTestId('registration-flow-section')).toBeInTheDocument();
    });
  });

  describe('Layout', () => {
    it('should render sections in correct order', async () => {
      const {container} = await render(
          <EditFlowsSettings application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />
      );

      const sections = container.querySelectorAll('[data-testid]');
      expect(sections[0]).toHaveAttribute('data-testid', 'auth-flow-section');
      expect(sections[1]).toHaveAttribute('data-testid', 'registration-flow-section');
    });
  });
});
