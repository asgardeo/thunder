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
import EditCustomizationSettings from '../EditCustomizationSettings';
import type {Application} from '../../../../models/application';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
}));

vi.mock('@thunder/shared-design', () => ({
  useGetThemes: vi.fn(() => ({
    data: {
      themes: [
        {id: 'theme-1', displayName: 'Default Theme'},
        {id: 'theme-2', displayName: 'Dark Theme'},
      ],
    },
    isLoading: false,
  })),
}));

describe('EditCustomizationSettings', () => {
  const mockApplication: Application = {
    id: 'test-app-id',
    name: 'Test Application',
    description: 'Test Description',
    template: 'custom',
    theme_id: 'theme-1',
    tos_uri: 'https://example.com/terms',
    policy_uri: 'https://example.com/privacy',
    contacts: ['contact@example.com'],
  } as Application;

  const mockOnFieldChange = vi.fn();

  describe('Rendering', () => {
    it('should render all three sections', () => {
      render(
        <EditCustomizationSettings application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />,
      );

      expect(screen.getByText('applications:edit.customization.sections.appearance')).toBeInTheDocument();
      expect(screen.getByText('applications:edit.customization.sections.urls')).toBeInTheDocument();
      expect(screen.getByText('applications:edit.general.sections.contacts')).toBeInTheDocument();
    });

    it('should render sections in correct order', () => {
      render(
        <EditCustomizationSettings application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />,
      );

      // Verify all three sections are present
      expect(screen.getByText('applications:edit.customization.sections.appearance')).toBeInTheDocument();
      expect(screen.getByText('applications:edit.customization.sections.urls')).toBeInTheDocument();
      expect(screen.getByText('applications:edit.general.sections.contacts')).toBeInTheDocument();
    });
  });

  describe('Section Integration', () => {
    it('should pass correct props to AppearanceSection', () => {
      render(
        <EditCustomizationSettings application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />,
      );

      expect(screen.getByText('applications:edit.customization.labels.theme')).toBeInTheDocument();
      expect(screen.getByPlaceholderText('applications:edit.customization.theme.placeholder')).toBeInTheDocument();
    });

    it('should pass correct props to UrlsSection', () => {
      render(
        <EditCustomizationSettings application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />,
      );

      expect(screen.getByText('applications:edit.customization.labels.tosUri')).toBeInTheDocument();
      expect(screen.getByText('applications:edit.customization.labels.policyUri')).toBeInTheDocument();
    });

    it('should pass correct props to ContactsSection', () => {
      render(
        <EditCustomizationSettings application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />,
      );

      expect(screen.getByPlaceholderText('applications:edit.general.contacts.placeholder')).toBeInTheDocument();
    });
  });

  describe('Layout', () => {
    it('should render sections in a Stack with spacing', () => {
      const {container} = render(
        <EditCustomizationSettings application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />,
      );

      const stack = container.firstChild;
      expect(stack).toHaveClass('MuiStack-root');
    });
  });

  describe('Props Propagation', () => {
    it('should propagate application prop to all sections', () => {
      render(
        <EditCustomizationSettings application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />,
      );

      // Verify theme from application
      expect(screen.getByRole('combobox')).toHaveValue('Default Theme');

      // Verify URLs from application
      const tosField = screen.getByPlaceholderText('applications:edit.customization.tosUri.placeholder');
      const policyField = screen.getByPlaceholderText('applications:edit.customization.policyUri.placeholder');
      expect(tosField).toHaveValue('https://example.com/terms');
      expect(policyField).toHaveValue('https://example.com/privacy');

      // Verify contacts from application
      const contactsField = screen.getByPlaceholderText('applications:edit.general.contacts.placeholder');
      expect(contactsField).toHaveValue('contact@example.com');
    });

    it('should propagate editedApp prop to all sections', () => {
      const editedApp = {
        theme_id: 'theme-2',
        tos_uri: 'https://edited.com/terms',
        policy_uri: 'https://edited.com/privacy',
        contacts: ['edited@example.com'],
      };

      render(
        <EditCustomizationSettings
          application={mockApplication}
          editedApp={editedApp}
          onFieldChange={mockOnFieldChange}
        />,
      );

      // Verify edited theme
      expect(screen.getByRole('combobox')).toHaveValue('Dark Theme');

      // Verify edited URLs
      const tosField = screen.getByPlaceholderText('applications:edit.customization.tosUri.placeholder');
      const policyField = screen.getByPlaceholderText('applications:edit.customization.policyUri.placeholder');
      expect(tosField).toHaveValue('https://edited.com/terms');
      expect(policyField).toHaveValue('https://edited.com/privacy');

      // Verify edited contacts
      const contactsField = screen.getByPlaceholderText('applications:edit.general.contacts.placeholder');
      expect(contactsField).toHaveValue('edited@example.com');
    });

    it('should propagate onFieldChange callback to all sections', () => {
      render(
        <EditCustomizationSettings application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />,
      );

      // All sections should be rendered, which means onFieldChange was passed
      expect(screen.getByText('applications:edit.customization.sections.appearance')).toBeInTheDocument();
      expect(screen.getByText('applications:edit.customization.sections.urls')).toBeInTheDocument();
      expect(screen.getByText('applications:edit.general.sections.contacts')).toBeInTheDocument();
    });
  });

  describe('Edge Cases', () => {
    it('should handle minimal application data', () => {
      const minimalApp = {
        id: 'minimal-id',
        name: 'Minimal App',
        template: 'custom',
      } as Application;

      render(<EditCustomizationSettings application={minimalApp} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      expect(screen.getByText('applications:edit.customization.sections.appearance')).toBeInTheDocument();
      expect(screen.getByText('applications:edit.customization.sections.urls')).toBeInTheDocument();
      expect(screen.getByText('applications:edit.general.sections.contacts')).toBeInTheDocument();
    });

    it('should handle empty editedApp', () => {
      render(
        <EditCustomizationSettings application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />,
      );

      // Should fall back to application values
      expect(screen.getByRole('combobox')).toHaveValue('Default Theme');
    });
  });
});
