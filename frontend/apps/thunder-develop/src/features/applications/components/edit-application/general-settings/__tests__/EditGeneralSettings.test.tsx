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
import {render, screen} from '@testing-library/react';
import EditGeneralSettings from '../EditGeneralSettings';
import type {Application} from '../../../../models/application';
import type {OAuth2Config} from '../../../../models/oauth';

// Mock the child components
vi.mock('../QuickCopySection', () => ({
  default: ({
    application,
    oauth2Config,
    copiedField,
  }: {
    application: Application;
    oauth2Config?: OAuth2Config;
    copiedField: string | null;
  }) => (
    <div data-testid="quick-copy-section">
      QuickCopySection - App: {application.id}, OAuth: {oauth2Config?.client_id ?? 'None'}, Copied:{' '}
      {copiedField ?? 'None'}
    </div>
  ),
}));

vi.mock('../AccessSection', () => ({
  default: ({
    application,
    editedApp,
    oauth2Config,
  }: {
    application: Application;
    editedApp: Partial<Application>;
    oauth2Config?: OAuth2Config;
  }) => (
    <div data-testid="access-section">
      AccessSection - App: {application.id}, Edited URL: {editedApp.url ?? 'None'}, OAuth:{' '}
      {oauth2Config?.client_id ?? 'None'}
    </div>
  ),
}));

describe('EditGeneralSettings', () => {
  const mockOnFieldChange = vi.fn();
  const mockOnCopyToClipboard = vi.fn();
  const mockApplication: Application = {
    id: 'app-123',
    name: 'Test App',
    url: 'https://example.com',
  } as Application;

  const mockOAuth2Config: OAuth2Config = {
    client_id: 'client-123',
    client_secret: 'secret-456',
  } as OAuth2Config;

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('should render both QuickCopySection and AccessSection', () => {
      render(
        <EditGeneralSettings
          application={mockApplication}
          editedApp={{}}
          onFieldChange={mockOnFieldChange}
          copiedField={null}
          onCopyToClipboard={mockOnCopyToClipboard}
        />,
      );

      expect(screen.getByTestId('quick-copy-section')).toBeInTheDocument();
      expect(screen.getByTestId('access-section')).toBeInTheDocument();
    });

    it('should pass application to child components', () => {
      render(
        <EditGeneralSettings
          application={mockApplication}
          editedApp={{}}
          onFieldChange={mockOnFieldChange}
          copiedField={null}
          onCopyToClipboard={mockOnCopyToClipboard}
        />,
      );

      expect(screen.getByTestId('quick-copy-section')).toHaveTextContent('App: app-123');
      expect(screen.getByTestId('access-section')).toHaveTextContent('App: app-123');
    });

    it('should pass editedApp to AccessSection', () => {
      const editedApp = {url: 'https://edited.com'};

      render(
        <EditGeneralSettings
          application={mockApplication}
          editedApp={editedApp}
          onFieldChange={mockOnFieldChange}
          copiedField={null}
          onCopyToClipboard={mockOnCopyToClipboard}
        />,
      );

      expect(screen.getByTestId('access-section')).toHaveTextContent('Edited URL: https://edited.com');
    });

    it('should pass oauth2Config to child components when provided', () => {
      render(
        <EditGeneralSettings
          application={mockApplication}
          editedApp={{}}
          onFieldChange={mockOnFieldChange}
          oauth2Config={mockOAuth2Config}
          copiedField={null}
          onCopyToClipboard={mockOnCopyToClipboard}
        />,
      );

      expect(screen.getByTestId('quick-copy-section')).toHaveTextContent('OAuth: client-123');
      expect(screen.getByTestId('access-section')).toHaveTextContent('OAuth: client-123');
    });

    it('should handle missing oauth2Config', () => {
      render(
        <EditGeneralSettings
          application={mockApplication}
          editedApp={{}}
          onFieldChange={mockOnFieldChange}
          copiedField={null}
          onCopyToClipboard={mockOnCopyToClipboard}
        />,
      );

      expect(screen.getByTestId('quick-copy-section')).toHaveTextContent('OAuth: None');
      expect(screen.getByTestId('access-section')).toHaveTextContent('OAuth: None');
    });

    it('should pass copiedField to QuickCopySection', () => {
      render(
        <EditGeneralSettings
          application={mockApplication}
          editedApp={{}}
          onFieldChange={mockOnFieldChange}
          copiedField="app_id"
          onCopyToClipboard={mockOnCopyToClipboard}
        />,
      );

      expect(screen.getByTestId('quick-copy-section')).toHaveTextContent('Copied: app_id');
    });

    it('should handle null copiedField', () => {
      render(
        <EditGeneralSettings
          application={mockApplication}
          editedApp={{}}
          onFieldChange={mockOnFieldChange}
          copiedField={null}
          onCopyToClipboard={mockOnCopyToClipboard}
        />,
      );

      expect(screen.getByTestId('quick-copy-section')).toHaveTextContent('Copied: None');
    });
  });

  describe('Props Propagation', () => {
    it('should pass onFieldChange to AccessSection', () => {
      const {container} = render(
        <EditGeneralSettings
          application={mockApplication}
          editedApp={{}}
          onFieldChange={mockOnFieldChange}
          copiedField={null}
          onCopyToClipboard={mockOnCopyToClipboard}
        />,
      );

      expect(container.querySelector('[data-testid="access-section"]')).toBeInTheDocument();
    });

    it('should pass onCopyToClipboard to QuickCopySection', () => {
      const {container} = render(
        <EditGeneralSettings
          application={mockApplication}
          editedApp={{}}
          onFieldChange={mockOnFieldChange}
          copiedField={null}
          onCopyToClipboard={mockOnCopyToClipboard}
        />,
      );

      expect(container.querySelector('[data-testid="quick-copy-section"]')).toBeInTheDocument();
    });

    it('should pass all required props to both child components', () => {
      const editedApp = {url: 'https://new.com'};

      render(
        <EditGeneralSettings
          application={mockApplication}
          editedApp={editedApp}
          onFieldChange={mockOnFieldChange}
          oauth2Config={mockOAuth2Config}
          copiedField="client_id"
          onCopyToClipboard={mockOnCopyToClipboard}
        />,
      );

      expect(screen.getByTestId('quick-copy-section')).toBeInTheDocument();
      expect(screen.getByTestId('access-section')).toBeInTheDocument();
    });
  });

  describe('Layout', () => {
    it('should render sections in correct order', () => {
      const {container} = render(
        <EditGeneralSettings
          application={mockApplication}
          editedApp={{}}
          onFieldChange={mockOnFieldChange}
          copiedField={null}
          onCopyToClipboard={mockOnCopyToClipboard}
        />,
      );

      const sections = container.querySelectorAll('[data-testid]');
      expect(sections[0]).toHaveAttribute('data-testid', 'quick-copy-section');
      expect(sections[1]).toHaveAttribute('data-testid', 'access-section');
    });
  });

  describe('Optional Props', () => {
    it('should render correctly when oauth2Config is explicitly undefined', () => {
      render(
        <EditGeneralSettings
          application={mockApplication}
          editedApp={{}}
          onFieldChange={mockOnFieldChange}
          oauth2Config={undefined}
          copiedField={null}
          onCopyToClipboard={mockOnCopyToClipboard}
        />,
      );

      expect(screen.getByTestId('quick-copy-section')).toHaveTextContent('OAuth: None');
      expect(screen.getByTestId('access-section')).toHaveTextContent('OAuth: None');
    });
  });
});
