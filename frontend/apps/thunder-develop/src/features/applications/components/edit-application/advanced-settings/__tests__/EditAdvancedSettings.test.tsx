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

import {describe, it, expect, vi} from 'vitest';
import {render} from '@thunder/test-utils/browser';
import {page} from 'vitest/browser';
import EditAdvancedSettings from '../EditAdvancedSettings';
import type {Application} from '../../../../models/application';
import type {OAuth2Config} from '../../../../models/oauth';
import CertificateTypes from '../../../../constants/certificate-types';

describe('EditAdvancedSettings', () => {
  const mockApplication: Application = {
    id: 'test-app-id',
    name: 'Test Application',
    description: 'Test Description',
    template: 'custom',
    certificate: {
      type: CertificateTypes.NONE,
      value: '',
    },
    created_at: '2025-01-01T00:00:00Z',
    updated_at: '2025-01-15T00:00:00Z',
  } as Application;

  const mockOAuth2Config: OAuth2Config = {
    grant_types: ['authorization_code', 'refresh_token'],
    response_types: ['code'],
    pkce_required: true,
    public_client: false,
  };

  const mockOnFieldChange = vi.fn();

  describe('Rendering', () => {
    it('should render all three sections', async () => {
      await render(
        <EditAdvancedSettings
          application={mockApplication}
          editedApp={{}}
          oauth2Config={mockOAuth2Config}
          onFieldChange={mockOnFieldChange}
        />,
      );

      await expect.element(page.getByText('OAuth2 Configuration')).toBeInTheDocument();
      await expect.element(page.getByRole('heading', {name: 'Certificate'})).toBeInTheDocument();
      await expect.element(page.getByRole('heading', {name: 'Metadata'})).toBeInTheDocument();
    });

    it('should render without OAuth2 config when not provided', async () => {
      await render(<EditAdvancedSettings application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      await expect.element(page.getByText('OAuth2 Configuration')).not.toBeInTheDocument();
      await expect.element(page.getByRole('heading', {name: 'Certificate'})).toBeInTheDocument();
      await expect.element(page.getByRole('heading', {name: 'Metadata'})).toBeInTheDocument();
    });

    it('should render without metadata section when timestamps are missing', async () => {
      const appWithoutMetadata = {...mockApplication};
      delete (appWithoutMetadata as Partial<Application>).created_at;
      delete (appWithoutMetadata as Partial<Application>).updated_at;

      await render(
        <EditAdvancedSettings
          application={appWithoutMetadata}
          editedApp={{}}
          oauth2Config={mockOAuth2Config}
          onFieldChange={mockOnFieldChange}
        />,
      );

      await expect.element(page.getByText('OAuth2 Configuration')).toBeInTheDocument();
      await expect.element(page.getByRole('heading', {name: 'Certificate'})).toBeInTheDocument();
      await expect.element(page.getByRole('heading', {name: 'Metadata'})).not.toBeInTheDocument();
    });
  });

  describe('Section Integration', () => {
    it('should pass correct props to OAuth2ConfigSection', async () => {
      await render(
        <EditAdvancedSettings
          application={mockApplication}
          editedApp={{}}
          oauth2Config={mockOAuth2Config}
          onFieldChange={mockOnFieldChange}
        />,
      );

      await expect.element(page.getByText('authorization_code', {exact: true})).toBeInTheDocument();
      await expect.element(page.getByText('refresh_token', {exact: true})).toBeInTheDocument();
      await expect.element(page.getByText('code', {exact: true})).toBeInTheDocument();
    });

    it('should pass correct props to CertificateSection', async () => {
      await render(
        <EditAdvancedSettings
          application={mockApplication}
          editedApp={{}}
          oauth2Config={mockOAuth2Config}
          onFieldChange={mockOnFieldChange}
        />,
      );

      await expect.element(page.getByLabelText('Certificate Type')).toBeInTheDocument();
    });

    it('should pass correct props to MetadataSection', async () => {
      await render(
        <EditAdvancedSettings
          application={mockApplication}
          editedApp={{}}
          oauth2Config={mockOAuth2Config}
          onFieldChange={mockOnFieldChange}
        />,
      );

      await expect.element(page.getByText('Created At')).toBeInTheDocument();
      await expect.element(page.getByText('Updated At')).toBeInTheDocument();
    });
  });

  describe('Layout', () => {
    it('should render sections in a Stack with spacing', async () => {
      const {container} = await render(
        <EditAdvancedSettings
          application={mockApplication}
          editedApp={{}}
          oauth2Config={mockOAuth2Config}
          onFieldChange={mockOnFieldChange}
        />,
      );

      const stack = container.firstChild;
      expect(stack).toHaveClass('MuiStack-root');
    });
  });

  describe('Edge Cases', () => {
    it('should handle undefined oauth2Config', async () => {
      await render(
        <EditAdvancedSettings
          application={mockApplication}
          editedApp={{}}
          oauth2Config={undefined}
          onFieldChange={mockOnFieldChange}
        />,
      );

      await expect.element(page.getByText('OAuth2 Configuration')).not.toBeInTheDocument();
    });

    it('should handle empty editedApp', async () => {
      await render(
        <EditAdvancedSettings
          application={mockApplication}
          editedApp={{}}
          oauth2Config={mockOAuth2Config}
          onFieldChange={mockOnFieldChange}
        />,
      );

      await expect.element(page.getByRole('heading', {name: 'Certificate'})).toBeInTheDocument();
    });

    it('should render with minimal application data', async () => {
      const minimalApp = {
        id: 'minimal-id',
        name: 'Minimal App',
        template: 'custom',
      } as Application;

      await render(<EditAdvancedSettings application={minimalApp} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      await expect.element(page.getByRole('heading', {name: 'Certificate'})).toBeInTheDocument();
    });
  });
});
