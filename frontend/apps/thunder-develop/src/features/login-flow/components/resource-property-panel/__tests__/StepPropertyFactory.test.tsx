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
import {StepTypes} from '@/features/flows/models/steps';
import type {Resource} from '@/features/flows/models/resources';
import StepPropertyFactory from '../StepPropertyFactory';

// Mock CommonStepPropertyFactory
vi.mock('@/features/flows/components/resource-property-panel/CommonStepPropertyFactory', () => ({
  default: ({resource, propertyKey}: {resource: Resource; propertyKey: string}) => (
    <div data-testid="common-step-property-factory" data-resource-id={resource.id} data-property-key={propertyKey}>
      Common Step Property Factory
    </div>
  ),
}));

// Mock RulesProperties
vi.mock('../nodes/RulesProperties', () => ({
  default: () => <div data-testid="rules-properties">Rules Properties</div>,
}));

describe('StepPropertyFactory', () => {
  const mockOnChange = vi.fn();

  const createMockResource = (type: string, overrides: Partial<Resource> = {}): Resource =>
    ({
      id: 'step-1',
      type,
      category: 'DECISION',
      resourceType: 'STEP',
      ...overrides,
    }) as Resource;

  describe('Rule Step Type', () => {
    it('should render RulesProperties for Rule step type', () => {
      const resource = createMockResource(StepTypes.Rule);

      render(
        <StepPropertyFactory
          resource={resource}
          propertyKey="condition"
          propertyValue="test"
          onChange={mockOnChange}
        />,
      );

      expect(screen.getByTestId('rules-properties')).toBeInTheDocument();
    });

    it('should not render CommonStepPropertyFactory for Rule step type', () => {
      const resource = createMockResource(StepTypes.Rule);

      render(
        <StepPropertyFactory
          resource={resource}
          propertyKey="condition"
          propertyValue="test"
          onChange={mockOnChange}
        />,
      );

      expect(screen.queryByTestId('common-step-property-factory')).not.toBeInTheDocument();
    });
  });

  describe('Other Step Types', () => {
    it('should render CommonStepPropertyFactory for View step type', () => {
      const resource = createMockResource(StepTypes.View);

      render(
        <StepPropertyFactory
          resource={resource}
          propertyKey="name"
          propertyValue="Step Name"
          onChange={mockOnChange}
        />,
      );

      expect(screen.getByTestId('common-step-property-factory')).toBeInTheDocument();
    });

    it('should render CommonStepPropertyFactory for Execution step type', () => {
      const resource = createMockResource(StepTypes.Execution);

      render(
        <StepPropertyFactory
          resource={resource}
          propertyKey="executor"
          propertyValue="executor-1"
          onChange={mockOnChange}
        />,
      );

      expect(screen.getByTestId('common-step-property-factory')).toBeInTheDocument();
    });

    it('should pass props to CommonStepPropertyFactory', () => {
      const resource = createMockResource(StepTypes.View, {id: 'view-step-123'});

      render(
        <StepPropertyFactory
          resource={resource}
          propertyKey="description"
          propertyValue="Step description"
          onChange={mockOnChange}
        />,
      );

      const factory = screen.getByTestId('common-step-property-factory');
      expect(factory).toHaveAttribute('data-resource-id', 'view-step-123');
      expect(factory).toHaveAttribute('data-property-key', 'description');
    });
  });
});
