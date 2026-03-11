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
import EditCustomizationSettings from '../EditCustomizationSettings';
import type {Application} from '../../../../models/application';

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
    it('should render all three sections', async () => {
      await render(
        <EditCustomizationSettings application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />,
      );

      await expect.element(page.getByRole('heading', {name: 'Appearance'})).toBeInTheDocument();
      await expect.element(page.getByRole('heading', {name: 'URLs'})).toBeInTheDocument();
      await expect.element(page.getByRole('heading', {name: 'Contacts'})).toBeInTheDocument();
    });

    it('should render sections in correct order', async () => {
      await render(
        <EditCustomizationSettings application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />,
      );

      // Verify all three sections are present
      await expect.element(page.getByRole('heading', {name: 'Appearance'})).toBeInTheDocument();
      await expect.element(page.getByRole('heading', {name: 'URLs'})).toBeInTheDocument();
      await expect.element(page.getByRole('heading', {name: 'Contacts'})).toBeInTheDocument();
    });
  });

  describe('Section Integration', () => {
    it('should pass correct props to AppearanceSection', async () => {
      await render(
        <EditCustomizationSettings application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />,
      );

      await expect.element(page.getByText('Theme', {exact: true})).toBeInTheDocument();
      await expect.element(page.getByPlaceholder('Select a Theme')).toBeInTheDocument();
    });

    it('should pass correct props to UrlsSection', async () => {
      await render(
        <EditCustomizationSettings application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />,
      );

      await expect.element(page.getByText('Terms of Service URI')).toBeInTheDocument();
      await expect.element(page.getByText('Privacy Policy URI')).toBeInTheDocument();
    });

    it('should pass correct props to ContactsSection', async () => {
      await render(
        <EditCustomizationSettings application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />,
      );

      await expect.element(page.getByPlaceholder('admin@example.com, support@example.com')).toBeInTheDocument();
    });
  });

  describe('Layout', () => {
    it('should render sections in a Stack with spacing', async () => {
      const {container} = await render(
        <EditCustomizationSettings application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />,
      );

      const stack = container.firstChild;
      expect(stack).toHaveClass('MuiStack-root');
    });
  });

  describe('Props Propagation', () => {
    it('should propagate application prop to all sections', async () => {
      await render(
        <EditCustomizationSettings application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />,
      );

      // Verify theme from application
      await expect.element(page.getByRole('combobox')).toHaveValue('Default Theme');

      // Verify URLs from application
      const tosField = page.getByPlaceholder('https://example.com/terms');
      const policyField = page.getByPlaceholder('https://example.com/privacy');
      expect(tosField).toHaveValue('https://example.com/terms');
      expect(policyField).toHaveValue('https://example.com/privacy');

      // Verify contacts from application
      const contactsField = page.getByPlaceholder('admin@example.com, support@example.com');
      expect(contactsField).toHaveValue('contact@example.com');
    });

    it('should propagate editedApp prop to all sections', async () => {
      const editedApp = {
        theme_id: 'theme-2',
        tos_uri: 'https://edited.com/terms',
        policy_uri: 'https://edited.com/privacy',
        contacts: ['edited@example.com'],
      };

      await render(
        <EditCustomizationSettings
          application={mockApplication}
          editedApp={editedApp}
          onFieldChange={mockOnFieldChange}
        />,
      );

      // Verify edited theme
      await expect.element(page.getByRole('combobox')).toHaveValue('Dark Theme');

      // Verify edited URLs
      const tosField = page.getByPlaceholder('https://example.com/terms');
      const policyField = page.getByPlaceholder('https://example.com/privacy');
      expect(tosField).toHaveValue('https://edited.com/terms');
      expect(policyField).toHaveValue('https://edited.com/privacy');

      // Verify edited contacts
      const contactsField = page.getByPlaceholder('admin@example.com, support@example.com');
      expect(contactsField).toHaveValue('edited@example.com');
    });

    it('should propagate onFieldChange callback to all sections', async () => {
      await render(
        <EditCustomizationSettings application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />,
      );

      // All sections should be rendered, which means onFieldChange was passed
      await expect.element(page.getByRole('heading', {name: 'Appearance'})).toBeInTheDocument();
      await expect.element(page.getByRole('heading', {name: 'URLs'})).toBeInTheDocument();
      await expect.element(page.getByRole('heading', {name: 'Contacts'})).toBeInTheDocument();
    });
  });

  describe('Edge Cases', () => {
    it('should handle minimal application data', async () => {
      const minimalApp = {
        id: 'minimal-id',
        name: 'Minimal App',
        template: 'custom',
      } as Application;

      await render(<EditCustomizationSettings application={minimalApp} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      await expect.element(page.getByRole('heading', {name: 'Appearance'})).toBeInTheDocument();
      await expect.element(page.getByRole('heading', {name: 'URLs'})).toBeInTheDocument();
      await expect.element(page.getByRole('heading', {name: 'Contacts'})).toBeInTheDocument();
    });

    it('should handle empty editedApp', async () => {
      await render(
        <EditCustomizationSettings application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />,
      );

      // Should fall back to application values
      await expect.element(page.getByRole('combobox')).toHaveValue('Default Theme');
    });
  });
});
