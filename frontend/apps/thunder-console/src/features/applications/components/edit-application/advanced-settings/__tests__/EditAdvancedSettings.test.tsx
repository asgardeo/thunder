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
import userEvent from '@testing-library/user-event';
import {render, screen} from '@testing-library/react';
import EditAdvancedSettings from '../EditAdvancedSettings';
import type {Application} from '../../../../models/application';
import type {OAuth2Config} from '../../../../models/oauth';
import type {InboundAuthConfig} from '../../../../models/inbound-auth';
import CertificateTypes from '../../../../constants/certificate-types';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
}));

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
    createdAt: '2025-01-01T00:00:00Z',
    updatedAt: '2025-01-15T00:00:00Z',
  } as Application;

  const mockOAuth2Config: OAuth2Config = {
    grantTypes: ['authorization_code', 'refresh_token'],
    responseTypes: ['code'],
    pkceRequired: true,
    publicClient: false,
  };

  const mockOnFieldChange = vi.fn();

  describe('Rendering', () => {
    it('should render all three sections', () => {
      render(
        <EditAdvancedSettings
          application={mockApplication}
          editedApp={{}}
          oauth2Config={mockOAuth2Config}
          onFieldChange={mockOnFieldChange}
        />,
      );

      expect(screen.getByText('applications:edit.advanced.labels.oauth2Config')).toBeInTheDocument();
      expect(screen.getByText('applications:edit.advanced.labels.certificate')).toBeInTheDocument();
      expect(screen.getByText('applications:edit.advanced.labels.metadata')).toBeInTheDocument();
    });

    it('should render without OAuth2 config when not provided', () => {
      render(<EditAdvancedSettings application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      expect(screen.queryByText('applications:edit.advanced.labels.oauth2Config')).not.toBeInTheDocument();
      expect(screen.getByText('applications:edit.advanced.labels.certificate')).toBeInTheDocument();
      expect(screen.getByText('applications:edit.advanced.labels.metadata')).toBeInTheDocument();
    });

    it('should render without metadata section when timestamps are missing', () => {
      const appWithoutMetadata = {...mockApplication};
      delete (appWithoutMetadata as Partial<Application>).createdAt;
      delete (appWithoutMetadata as Partial<Application>).updatedAt;

      render(
        <EditAdvancedSettings
          application={appWithoutMetadata}
          editedApp={{}}
          oauth2Config={mockOAuth2Config}
          onFieldChange={mockOnFieldChange}
        />,
      );

      expect(screen.getByText('applications:edit.advanced.labels.oauth2Config')).toBeInTheDocument();
      expect(screen.getByText('applications:edit.advanced.labels.certificate')).toBeInTheDocument();
      expect(screen.queryByText('applications:edit.advanced.labels.metadata')).not.toBeInTheDocument();
    });
  });

  describe('Section Integration', () => {
    it('should pass correct props to OAuth2ConfigSection', () => {
      render(
        <EditAdvancedSettings
          application={mockApplication}
          editedApp={{}}
          oauth2Config={mockOAuth2Config}
          onFieldChange={mockOnFieldChange}
        />,
      );

      expect(screen.getByText('authorization_code')).toBeInTheDocument();
      expect(screen.getByText('refresh_token')).toBeInTheDocument();
      expect(screen.getByText('code')).toBeInTheDocument();
    });

    it('should pass correct props to CertificateSection', () => {
      render(
        <EditAdvancedSettings
          application={mockApplication}
          editedApp={{}}
          oauth2Config={mockOAuth2Config}
          onFieldChange={mockOnFieldChange}
        />,
      );

      expect(screen.getByLabelText('applications:edit.advanced.labels.certificateType')).toBeInTheDocument();
    });

    it('should pass correct props to MetadataSection', () => {
      render(
        <EditAdvancedSettings
          application={mockApplication}
          editedApp={{}}
          oauth2Config={mockOAuth2Config}
          onFieldChange={mockOnFieldChange}
        />,
      );

      expect(screen.getByText('applications:edit.advanced.labels.createdAt')).toBeInTheDocument();
      expect(screen.getByText('applications:edit.advanced.labels.updatedAt')).toBeInTheDocument();
    });
  });

  describe('Layout', () => {
    it('should render sections in a Stack with spacing', () => {
      const {container} = render(
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
    it('should handle undefined oauth2Config', () => {
      render(
        <EditAdvancedSettings
          application={mockApplication}
          editedApp={{}}
          oauth2Config={undefined}
          onFieldChange={mockOnFieldChange}
        />,
      );

      expect(screen.queryByText('applications:edit.advanced.labels.oauth2Config')).not.toBeInTheDocument();
    });

    it('should handle empty editedApp', () => {
      render(
        <EditAdvancedSettings
          application={mockApplication}
          editedApp={{}}
          oauth2Config={mockOAuth2Config}
          onFieldChange={mockOnFieldChange}
        />,
      );

      expect(screen.getByText('applications:edit.advanced.labels.certificate')).toBeInTheDocument();
    });

    it('should render with minimal application data', () => {
      const minimalApp = {
        id: 'minimal-id',
        name: 'Minimal App',
        template: 'custom',
      } as Application;

      render(<EditAdvancedSettings application={minimalApp} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      expect(screen.getByText('applications:edit.advanced.labels.certificate')).toBeInTheDocument();
    });
  });

  describe('OAuth2 Config Change Handler', () => {
    it('should call onFieldChange with updated inboundAuthConfig when a toggle is changed', async () => {
      const user = userEvent.setup();
      const onFieldChange = vi.fn();

      render(
        <EditAdvancedSettings
          application={mockApplication}
          editedApp={{}}
          oauth2Config={mockOAuth2Config}
          onFieldChange={onFieldChange}
        />,
      );

      // Change the PKCE switch to trigger onOAuth2ConfigChange → handleOAuth2ConfigChange
      const pkceSwitch = screen.getByRole('switch', {
        name: 'applications:edit.advanced.labels.pkceRequired',
      });
      await user.click(pkceSwitch);

      expect(onFieldChange).toHaveBeenCalledWith('inboundAuthConfig', []);
    });

    it('should merge OAuth2 config updates into existing inboundAuthConfig', async () => {
      const user = userEvent.setup();
      const onFieldChange = vi.fn();
      const appWithInbound: Application = {
        ...mockApplication,
        inboundAuthConfig: [
          {type: 'oauth2', config: {grantTypes: ['authorization_code']} as OAuth2Config} as InboundAuthConfig,
        ],
      };

      render(
        <EditAdvancedSettings
          application={appWithInbound}
          editedApp={{}}
          oauth2Config={mockOAuth2Config}
          onFieldChange={onFieldChange}
        />,
      );

      const pkceSwitch = screen.getByRole('switch', {
        name: 'applications:edit.advanced.labels.pkceRequired',
      });
      await user.click(pkceSwitch);

      expect(onFieldChange).toHaveBeenCalledWith('inboundAuthConfig', [
        {type: 'oauth2', config: {grantTypes: ['authorization_code'], pkceRequired: false}},
      ]);
    });

    it('should prefer editedApp.inboundAuthConfig over application.inboundAuthConfig', async () => {
      const user = userEvent.setup();
      const onFieldChange = vi.fn();
      const editedInbound: InboundAuthConfig[] = [
        {type: 'oauth2', config: {grantTypes: ['implicit']} as OAuth2Config} as InboundAuthConfig,
      ];

      render(
        <EditAdvancedSettings
          application={mockApplication}
          editedApp={{inboundAuthConfig: editedInbound}}
          oauth2Config={mockOAuth2Config}
          onFieldChange={onFieldChange}
        />,
      );

      const pkceSwitch = screen.getByRole('switch', {
        name: 'applications:edit.advanced.labels.pkceRequired',
      });
      await user.click(pkceSwitch);

      expect(onFieldChange).toHaveBeenCalledWith('inboundAuthConfig', [
        {type: 'oauth2', config: {grantTypes: ['implicit'], pkceRequired: false}},
      ]);
    });
  });
});
