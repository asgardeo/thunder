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

import {describe, it, expect, vi} from 'vitest';
import {screen} from '@testing-library/react';
import {renderWithProviders} from '../../../../../test/test-utils';
import EditGeneralSettings from '../general-settings/EditGeneralSettings';
import type {OrganizationUnit} from '../../../types/organization-units';

// Mock translations
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => {
      const translations: Record<string, string> = {
        'organizationUnits:view.general.title': 'General Information',
        'organizationUnits:view.general.subtitle': 'Basic details about this organization unit',
        'organizationUnits:form.handle': 'Handle',
        'organizationUnits:view.general.id': 'Organization Unit ID',
      };
      return translations[key] ?? key;
    },
  }),
}));

describe('EditGeneralSettings', () => {
  const mockOrganizationUnit: OrganizationUnit = {
    id: 'ou-123',
    handle: 'test-handle',
    name: 'Test Organization Unit',
    description: 'A test description',
    parent: null,
  };

  it('should render title and subtitle', () => {
    renderWithProviders(<EditGeneralSettings organizationUnit={mockOrganizationUnit} />);

    expect(screen.getByText('General Information')).toBeInTheDocument();
    expect(screen.getByText('Basic details about this organization unit')).toBeInTheDocument();
  });

  it('should display organization unit handle', () => {
    renderWithProviders(<EditGeneralSettings organizationUnit={mockOrganizationUnit} />);

    expect(screen.getByText('Handle')).toBeInTheDocument();
    expect(screen.getByText('test-handle')).toBeInTheDocument();
  });

  it('should display organization unit ID', () => {
    renderWithProviders(<EditGeneralSettings organizationUnit={mockOrganizationUnit} />);

    expect(screen.getByText('Organization Unit ID')).toBeInTheDocument();
    expect(screen.getByText('ou-123')).toBeInTheDocument();
  });

  it('should render with different organization unit data', () => {
    const differentOU: OrganizationUnit = {
      id: 'ou-456',
      handle: 'another-handle',
      name: 'Another OU',
      description: null,
      parent: 'parent-ou',
    };

    renderWithProviders(<EditGeneralSettings organizationUnit={differentOU} />);

    expect(screen.getByText('another-handle')).toBeInTheDocument();
    expect(screen.getByText('ou-456')).toBeInTheDocument();
  });
});
