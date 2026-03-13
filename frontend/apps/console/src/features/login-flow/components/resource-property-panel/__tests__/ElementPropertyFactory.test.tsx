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
import ElementPropertyFactory from '../ElementPropertyFactory';

// Mock CommonElementPropertyFactory
vi.mock('@/features/flows/components/resource-property-panel/CommonElementPropertyFactory', () => ({
  default: ({resource, propertyKey, propertyValue}: {resource: Resource; propertyKey: string; propertyValue: unknown}) => (
    <div data-testid="common-element-property-factory" data-resource-id={resource.id} data-property-key={propertyKey} data-property-value={String(propertyValue)}>
      Common Element Property Factory
    </div>
  ),
}));

describe('ElementPropertyFactory', () => {
  const createMockResource = (overrides: Partial<Resource> = {}): Resource =>
    ({
      id: 'element-1',
      type: 'TEXT_INPUT',
      category: 'FIELD',
      resourceType: 'ELEMENT',
      ...overrides,
    }) as Resource;

  const mockOnChange = vi.fn();

  describe('Rendering', () => {
    it('should render CommonElementPropertyFactory', () => {
      const resource = createMockResource();

      render(
        <ElementPropertyFactory
          resource={resource}
          propertyKey="label"
          propertyValue="Test Label"
          onChange={mockOnChange}
        />,
      );

      expect(screen.getByTestId('common-element-property-factory')).toBeInTheDocument();
    });

    it('should pass resource to CommonElementPropertyFactory', () => {
      const resource = createMockResource({id: 'custom-element'});

      render(
        <ElementPropertyFactory
          resource={resource}
          propertyKey="label"
          propertyValue="Test Label"
          onChange={mockOnChange}
        />,
      );

      expect(screen.getByTestId('common-element-property-factory')).toHaveAttribute('data-resource-id', 'custom-element');
    });

    it('should pass propertyKey to CommonElementPropertyFactory', () => {
      const resource = createMockResource();

      render(
        <ElementPropertyFactory
          resource={resource}
          propertyKey="placeholder"
          propertyValue="Enter text"
          onChange={mockOnChange}
        />,
      );

      expect(screen.getByTestId('common-element-property-factory')).toHaveAttribute('data-property-key', 'placeholder');
    });

    it('should pass propertyValue to CommonElementPropertyFactory', () => {
      const resource = createMockResource();

      render(
        <ElementPropertyFactory
          resource={resource}
          propertyKey="label"
          propertyValue="My Label"
          onChange={mockOnChange}
        />,
      );

      expect(screen.getByTestId('common-element-property-factory')).toHaveAttribute('data-property-value', 'My Label');
    });
  });
});
