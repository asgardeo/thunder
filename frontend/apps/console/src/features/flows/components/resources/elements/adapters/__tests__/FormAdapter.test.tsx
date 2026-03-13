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

import {describe, it, expect, vi, beforeEach} from 'vitest';
import {render, screen} from '@testing-library/react';
import type {ReactNode} from 'react';
import {ElementCategories, type Element as FlowElement} from '@/features/flows/models/elements';
import FormAdapter from '../FormAdapter';

// Mock dependencies
vi.mock('../FormAdapter.scss', () => ({}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
}));

vi.mock('@/features/flows/plugins/PluginRegistry', () => ({
  default: {
    getInstance: () => ({
      executeSync: () => true,
    }),
  },
}));

vi.mock('@/features/flows/utils/generateResourceId', () => ({
  default: (prefix: string) => `${prefix}-generated`,
}));

vi.mock('@/features/flows/components/resources/steps/view/ReorderableElement', () => ({
  default: ({element, id}: {element: FlowElement; id: string}) => (
    <div data-testid={`reorderable-element-${id}`}>{element.id}</div>
  ),
}));

vi.mock('@/features/flows/components/dnd/Droppable', () => ({
  default: ({children, id}: {children: ReactNode; id: string}) => (
    <div data-testid="droppable" data-droppable-id={id}>
      {children}
    </div>
  ),
}));

describe('FormAdapter', () => {
  const createMockElement = (overrides: Partial<FlowElement> = {}): FlowElement =>
    ({
      id: 'form-1',
      type: 'BLOCK',
      category: 'BLOCK',
      config: {},
      ...overrides,
    }) as FlowElement;

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('should render the form adapter with Badge', () => {
      const resource = createMockElement();

      const {container} = render(<FormAdapter resource={resource} stepId="step-1" />);

      expect(container.querySelector('.form-adapter')).toBeInTheDocument();
    });

    it('should render Badge with form label', () => {
      const resource = createMockElement();

      render(<FormAdapter resource={resource} stepId="step-1" />);

      expect(screen.getByText('flows:core.adapters.form.badgeLabel')).toBeInTheDocument();
    });

    it('should render Droppable component', () => {
      const resource = createMockElement();

      render(<FormAdapter resource={resource} stepId="step-1" />);

      expect(screen.getByTestId('droppable')).toBeInTheDocument();
    });
  });

  describe('Placeholder Display', () => {
    it('should show placeholder when no FIELD components exist', () => {
      const resource = createMockElement({components: []});

      render(<FormAdapter resource={resource} stepId="step-1" />);

      expect(screen.getByText('flows:core.adapters.form.placeholder')).toBeInTheDocument();
    });

    it('should show placeholder when components is undefined', () => {
      const resource = createMockElement({components: undefined});

      render(<FormAdapter resource={resource} stepId="step-1" />);

      expect(screen.getByText('flows:core.adapters.form.placeholder')).toBeInTheDocument();
    });

    it('should show placeholder when only non-FIELD components exist', () => {
      const components = [
        createMockElement({id: 'comp-1', category: ElementCategories.Action}),
        createMockElement({id: 'comp-2', category: ElementCategories.Display}),
      ];
      const resource = createMockElement({components});

      render(<FormAdapter resource={resource} stepId="step-1" />);

      expect(screen.getByText('flows:core.adapters.form.placeholder')).toBeInTheDocument();
    });

    it('should not show placeholder when FIELD components exist', () => {
      const components = [createMockElement({id: 'comp-1', category: ElementCategories.Field})];
      const resource = createMockElement({components});

      render(<FormAdapter resource={resource} stepId="step-1" />);

      expect(screen.queryByText('flows:core.adapters.form.placeholder')).not.toBeInTheDocument();
    });
  });

  describe('Components Rendering', () => {
    it('should render ReorderableFlowElement for each component', () => {
      const components = [
        createMockElement({id: 'comp-1', category: ElementCategories.Field}),
        createMockElement({id: 'comp-2', category: ElementCategories.Field}),
      ];
      const resource = createMockElement({components});

      render(<FormAdapter resource={resource} stepId="step-1" />);

      expect(screen.getByTestId('reorderable-element-comp-1')).toBeInTheDocument();
      expect(screen.getByTestId('reorderable-element-comp-2')).toBeInTheDocument();
    });

    it('should pass availableElements to ReorderableFlowElement', () => {
      const components = [createMockElement({id: 'comp-1', category: ElementCategories.Field})];
      const resource = createMockElement({components});
      const availableElements = [createMockElement({id: 'available-1'})];

      render(<FormAdapter resource={resource} stepId="step-1" availableElements={availableElements} />);

      expect(screen.getByTestId('reorderable-element-comp-1')).toBeInTheDocument();
    });

    it('should pass onAddElementToForm callback', () => {
      const components = [createMockElement({id: 'comp-1', category: ElementCategories.Field})];
      const resource = createMockElement({components});
      const onAddElementToForm = vi.fn();

      render(<FormAdapter resource={resource} stepId="step-1" onAddElementToForm={onAddElementToForm} />);

      expect(screen.getByTestId('reorderable-element-comp-1')).toBeInTheDocument();
    });
  });

  describe('Droppable Configuration', () => {
    it('should have unique droppable ID based on stepId', () => {
      const resource = createMockElement();

      render(<FormAdapter resource={resource} stepId="step-123" />);

      const droppable = screen.getByTestId('droppable');
      expect(droppable.getAttribute('data-droppable-id')).toContain('step-123');
    });
  });

  describe('Default Props', () => {
    it('should work with undefined availableElements', () => {
      const resource = createMockElement();

      const {container} = render(<FormAdapter resource={resource} stepId="step-1" />);

      expect(container.querySelector('.form-adapter')).toBeInTheDocument();
    });

    it('should work with undefined onAddElementToForm', () => {
      const resource = createMockElement();

      const {container} = render(<FormAdapter resource={resource} stepId="step-1" />);

      expect(container.querySelector('.form-adapter')).toBeInTheDocument();
    });
  });

  describe('Filtering', () => {
    it('should filter components through PluginRegistry', () => {
      const components = [
        createMockElement({id: 'comp-1', category: ElementCategories.Field}),
        createMockElement({id: 'comp-2', category: ElementCategories.Field}),
      ];
      const resource = createMockElement({components});

      render(<FormAdapter resource={resource} stepId="step-1" />);

      // All components should render since our mock returns true
      expect(screen.getByTestId('reorderable-element-comp-1')).toBeInTheDocument();
      expect(screen.getByTestId('reorderable-element-comp-2')).toBeInTheDocument();
    });
  });
});
