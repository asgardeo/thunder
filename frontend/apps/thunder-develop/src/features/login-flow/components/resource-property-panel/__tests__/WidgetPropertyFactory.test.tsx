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
import {render, screen} from '@testing-library/react';
import type {Resource} from '@/features/flows/models/resources';
import WidgetPropertyFactory from '../WidgetPropertyFactory';

// Mock CommonWidgetPropertyFactory
vi.mock('@/features/flows/components/resource-property-panel/CommonWidgetPropertyFactory', () => ({
  default: ({resource, propertyKey, propertyValue}: {resource: Resource; propertyKey: string; propertyValue: unknown}) => (
    <div data-testid="common-widget-property-factory" data-resource-id={resource.id} data-property-key={propertyKey} data-property-value={String(propertyValue)}>
      Common Widget Property Factory
    </div>
  ),
}));

describe('WidgetPropertyFactory', () => {
  const mockOnChange = vi.fn();

  const createMockResource = (overrides: Partial<Resource> = {}): Resource =>
    ({
      id: 'widget-1',
      type: 'HEADER',
      category: 'WIDGET',
      resourceType: 'WIDGET',
      ...overrides,
    }) as Resource;

  describe('Rendering', () => {
    it('should render CommonWidgetPropertyFactory', () => {
      const resource = createMockResource();

      render(
        <WidgetPropertyFactory
          resource={resource}
          propertyKey="title"
          propertyValue="Widget Title"
          onChange={mockOnChange}
        />,
      );

      expect(screen.getByTestId('common-widget-property-factory')).toBeInTheDocument();
    });

    it('should pass resource to CommonWidgetPropertyFactory', () => {
      const resource = createMockResource({id: 'custom-widget'});

      render(
        <WidgetPropertyFactory
          resource={resource}
          propertyKey="title"
          propertyValue="Title"
          onChange={mockOnChange}
        />,
      );

      expect(screen.getByTestId('common-widget-property-factory')).toHaveAttribute('data-resource-id', 'custom-widget');
    });

    it('should pass propertyKey to CommonWidgetPropertyFactory', () => {
      const resource = createMockResource();

      render(
        <WidgetPropertyFactory
          resource={resource}
          propertyKey="description"
          propertyValue="Description"
          onChange={mockOnChange}
        />,
      );

      expect(screen.getByTestId('common-widget-property-factory')).toHaveAttribute('data-property-key', 'description');
    });

    it('should pass propertyValue to CommonWidgetPropertyFactory', () => {
      const resource = createMockResource();

      render(
        <WidgetPropertyFactory
          resource={resource}
          propertyKey="title"
          propertyValue="My Widget"
          onChange={mockOnChange}
        />,
      );

      expect(screen.getByTestId('common-widget-property-factory')).toHaveAttribute('data-property-value', 'My Widget');
    });

    it('should handle different widget types', () => {
      const resource = createMockResource({type: 'FOOTER'});

      render(
        <WidgetPropertyFactory
          resource={resource}
          propertyKey="content"
          propertyValue="Footer content"
          onChange={mockOnChange}
        />,
      );

      expect(screen.getByTestId('common-widget-property-factory')).toBeInTheDocument();
    });
  });
});
