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
import {render, screen} from '@testing-library/react';
import EditAdvancedSettings from '../EditAdvancedSettings';
import type {Application} from '../../../../models/application';
import type {OAuth2Config} from '../../../../models/oauth';
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
      delete (appWithoutMetadata as Partial<Application>).created_at;
      delete (appWithoutMetadata as Partial<Application>).updated_at;

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
});
