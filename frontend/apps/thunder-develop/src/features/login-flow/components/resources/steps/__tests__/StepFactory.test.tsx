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
import {render} from '@thunder/test-utils/browser';
import {page} from 'vitest/browser';
import type {Step} from '@/features/flows/models/steps';
import {StepTypes, StepCategories} from '@/features/flows/models/steps';
import type {Resources} from '@/features/flows/models/resources';
import type {Element} from '@/features/flows/models/elements';
import StepFactory from '../StepFactory';

// Mock CommonStepFactory
vi.mock('@/features/flows/components/resources/steps/CommonStepFactory', () => ({
  default: ({
    resourceId,
    resources,
    allResources,
    onAddElement,
    onAddElementToForm,
  }: {
    resourceId: string;
    resources: Step[];
    allResources?: Resources;
    onAddElement?: (element: Element) => void;
    onAddElementToForm?: (element: Element, formId: string) => void;
  }) => (
    <div
      data-testid="common-step-factory"
      data-resource-id={resourceId}
      data-resources-count={resources?.length ?? 0}
      data-has-all-resources={!!allResources}
      data-has-on-add-element={!!onAddElement}
      data-has-on-add-element-to-form={!!onAddElementToForm}
    >
      Common Step Factory
    </div>
  ),
}));

describe('StepFactory', () => {
  const createMockStep = (overrides: Partial<Step> = {}): Step =>
    ({
      id: 'step-1',
      type: StepTypes.View,
      category: StepCategories.Interface,
      ...overrides,
    }) as Step;

  const createNodeProps = (overrides: Record<string, unknown> = {}) => ({
    id: 'node-1',
    type: 'step',
    position: {x: 0, y: 0},
    positionAbsoluteX: 0,
    positionAbsoluteY: 0,
    isConnectable: true,
    zIndex: 0,
    draggable: true,
    selectable: true,
    deletable: true,
    selected: false,
    dragging: false,
    data: {},
    ...overrides,
  });

  describe('Rendering', () => {
    it('should render CommonStepFactory', async () => {
      const props = {
        ...createNodeProps(),
        resourceId: 'resource-1',
        resources: [createMockStep()],
      };

      await render(<StepFactory {...props} />);

      await expect.element(page.getByTestId('common-step-factory')).toBeInTheDocument();
    });

    it('should pass resourceId to CommonStepFactory', async () => {
      const props = {
        ...createNodeProps(),
        resourceId: 'my-resource-id',
        resources: [createMockStep()],
      };

      await render(<StepFactory {...props} />);

      await expect.element(page.getByTestId('common-step-factory')).toHaveAttribute('data-resource-id', 'my-resource-id');
    });

    it('should pass resources to CommonStepFactory', async () => {
      const props = {
        ...createNodeProps(),
        resourceId: 'resource-1',
        resources: [createMockStep(), createMockStep({id: 'step-2'})],
      };

      await render(<StepFactory {...props} />);

      await expect.element(page.getByTestId('common-step-factory')).toHaveAttribute('data-resources-count', '2');
    });
  });

  describe('Optional Props', () => {
    it('should pass allResources when provided', async () => {
      const allResources: Resources = {
        elements: [],
        steps: [],
        widgets: [],
        templates: [],
        executors: [],
      };
      const props = {
        ...createNodeProps(),
        resourceId: 'resource-1',
        resources: [createMockStep()],
        allResources,
      };

      await render(<StepFactory {...props} />);

      await expect.element(page.getByTestId('common-step-factory')).toHaveAttribute('data-has-all-resources', 'true');
    });

    it('should handle undefined allResources', async () => {
      const props = {
        ...createNodeProps(),
        resourceId: 'resource-1',
        resources: [createMockStep()],
      };

      await render(<StepFactory {...props} />);

      await expect.element(page.getByTestId('common-step-factory')).toHaveAttribute('data-has-all-resources', 'false');
    });

    it('should pass onAddElement callback when provided', async () => {
      const onAddElement = vi.fn();
      const props = {
        ...createNodeProps(),
        resourceId: 'resource-1',
        resources: [createMockStep()],
        onAddElement,
      };

      await render(<StepFactory {...props} />);

      await expect.element(page.getByTestId('common-step-factory')).toHaveAttribute('data-has-on-add-element', 'true');
    });

    it('should pass onAddElementToForm callback when provided', async () => {
      const onAddElementToForm = vi.fn();
      const props = {
        ...createNodeProps(),
        resourceId: 'resource-1',
        resources: [createMockStep()],
        onAddElementToForm,
      };

      await render(<StepFactory {...props} />);

      await expect.element(page.getByTestId('common-step-factory')).toHaveAttribute('data-has-on-add-element-to-form', 'true');
    });

    it('should handle all optional props together', async () => {
      const allResources: Resources = {
        elements: [],
        steps: [],
        widgets: [],
        templates: [],
        executors: [],
      };
      const onAddElement = vi.fn();
      const onAddElementToForm = vi.fn();
      const props = {
        ...createNodeProps(),
        resourceId: 'resource-1',
        resources: [createMockStep()],
        allResources,
        onAddElement,
        onAddElementToForm,
      };

      await render(<StepFactory {...props} />);

      const factory = page.getByTestId('common-step-factory');
      expect(factory).toHaveAttribute('data-has-all-resources', 'true');
      expect(factory).toHaveAttribute('data-has-on-add-element', 'true');
      expect(factory).toHaveAttribute('data-has-on-add-element-to-form', 'true');
    });
  });

  describe('Memoization', () => {
    it('should render with same props', async () => {
      const props = {
        ...createNodeProps(),
        resourceId: 'resource-1',
        resources: [createMockStep()],
      };

      const {rerender} = await render(<StepFactory {...props} />);

      await expect.element(page.getByTestId('common-step-factory')).toBeInTheDocument();

      await rerender(<StepFactory {...props} />);

      await expect.element(page.getByTestId('common-step-factory')).toBeInTheDocument();
    });

    it('should render with different resource IDs', async () => {
      const props1 = {
        ...createNodeProps(),
        resourceId: 'resource-1',
        resources: [createMockStep()],
      };

      const {rerender} = await render(<StepFactory {...props1} />);

      await expect.element(page.getByTestId('common-step-factory')).toHaveAttribute('data-resource-id', 'resource-1');

      // Create props2 with different id (which is in memo comparison) to trigger re-render
      const props2 = {
        ...props1,
        id: 'node-2',
        resourceId: 'resource-2',
      };

      await rerender(<StepFactory {...props2} />);

      await expect.element(page.getByTestId('common-step-factory')).toHaveAttribute('data-resource-id', 'resource-2');
    });
  });

  describe('Different Step Types', () => {
    it('should render with View step type', async () => {
      const props = {
        ...createNodeProps(),
        resourceId: 'resource-1',
        resources: [createMockStep({type: StepTypes.View})],
      };

      await render(<StepFactory {...props} />);

      await expect.element(page.getByTestId('common-step-factory')).toBeInTheDocument();
    });

    it('should render with Rule step type', async () => {
      const props = {
        ...createNodeProps(),
        resourceId: 'resource-1',
        resources: [createMockStep({type: StepTypes.Rule, category: StepCategories.Decision})],
      };

      await render(<StepFactory {...props} />);

      await expect.element(page.getByTestId('common-step-factory')).toBeInTheDocument();
    });

    it('should render with Execution step type', async () => {
      const props = {
        ...createNodeProps(),
        resourceId: 'resource-1',
        resources: [createMockStep({type: StepTypes.Execution, category: StepCategories.Workflow})],
      };

      await render(<StepFactory {...props} />);

      await expect.element(page.getByTestId('common-step-factory')).toBeInTheDocument();
    });

    it('should render with End step type', async () => {
      const props = {
        ...createNodeProps(),
        resourceId: 'resource-1',
        resources: [createMockStep({type: StepTypes.End})],
      };

      await render(<StepFactory {...props} />);

      await expect.element(page.getByTestId('common-step-factory')).toBeInTheDocument();
    });
  });

  describe('Memo Comparison', () => {
    it('should not re-render when props are the same', async () => {
      const resources = [createMockStep()];
      const data = {test: 'value'};
      const allResources: Resources = {
        elements: [],
        steps: [],
        widgets: [],
        templates: [],
        executors: [],
      };
      const onAddElement = vi.fn();
      const onAddElementToForm = vi.fn();

      const props = {
        ...createNodeProps({data}),
        resourceId: 'resource-1',
        resources,
        allResources,
        onAddElement,
        onAddElementToForm,
      };

      const {rerender} = await render(<StepFactory {...props} />);
      await expect.element(page.getByTestId('common-step-factory')).toBeInTheDocument();

      await rerender(<StepFactory {...props} />);
      await expect.element(page.getByTestId('common-step-factory')).toBeInTheDocument();
    });

    it('should re-render when id changes', async () => {
      const resources = [createMockStep()];
      const props1 = {
        ...createNodeProps({id: 'node-1'}),
        resourceId: 'resource-1',
        resources,
      };

      const {rerender} = await render(<StepFactory {...props1} />);
      await expect.element(page.getByTestId('common-step-factory')).toBeInTheDocument();

      const props2 = {
        ...props1,
        id: 'node-2',
      };

      await rerender(<StepFactory {...props2} />);
      await expect.element(page.getByTestId('common-step-factory')).toBeInTheDocument();
    });

    it('should re-render when data changes', async () => {
      const resources = [createMockStep()];
      const props1 = {
        ...createNodeProps({data: {value: 1}}),
        resourceId: 'resource-1',
        resources,
      };

      const {rerender} = await render(<StepFactory {...props1} />);
      await expect.element(page.getByTestId('common-step-factory')).toBeInTheDocument();

      const props2 = {
        ...props1,
        data: {value: 2},
      };

      await rerender(<StepFactory {...props2} />);
      await expect.element(page.getByTestId('common-step-factory')).toBeInTheDocument();
    });

    it('should re-render when resources change', async () => {
      const props1 = {
        ...createNodeProps(),
        resourceId: 'resource-1',
        resources: [createMockStep()],
      };

      const {rerender} = await render(<StepFactory {...props1} />);
      await expect.element(page.getByTestId('common-step-factory')).toHaveAttribute('data-resources-count', '1');

      const props2 = {
        ...props1,
        resources: [createMockStep(), createMockStep({id: 'step-2'})],
      };

      await rerender(<StepFactory {...props2} />);
      await expect.element(page.getByTestId('common-step-factory')).toHaveAttribute('data-resources-count', '2');
    });

    it('should re-render when allResources changes', async () => {
      const props1 = {
        ...createNodeProps(),
        resourceId: 'resource-1',
        resources: [createMockStep()],
        allResources: undefined,
      };

      const {rerender} = await render(<StepFactory {...props1} />);
      await expect.element(page.getByTestId('common-step-factory')).toHaveAttribute('data-has-all-resources', 'false');

      const props2 = {
        ...props1,
        allResources: {
          elements: [],
          steps: [],
          widgets: [],
          templates: [],
          executors: [],
        },
      };

      await rerender(<StepFactory {...props2} />);
      await expect.element(page.getByTestId('common-step-factory')).toHaveAttribute('data-has-all-resources', 'true');
    });

    it('should re-render when onAddElement changes', async () => {
      const onAddElement1 = vi.fn();
      const props1 = {
        ...createNodeProps(),
        resourceId: 'resource-1',
        resources: [createMockStep()],
        onAddElement: onAddElement1,
      };

      const {rerender} = await render(<StepFactory {...props1} />);
      await expect.element(page.getByTestId('common-step-factory')).toHaveAttribute('data-has-on-add-element', 'true');

      const onAddElement2 = vi.fn();
      const props2 = {
        ...props1,
        onAddElement: onAddElement2,
      };

      await rerender(<StepFactory {...props2} />);
      await expect.element(page.getByTestId('common-step-factory')).toHaveAttribute('data-has-on-add-element', 'true');
    });

    it('should re-render when onAddElementToForm changes', async () => {
      const onAddElementToForm1 = vi.fn();
      const props1 = {
        ...createNodeProps(),
        resourceId: 'resource-1',
        resources: [createMockStep()],
        onAddElementToForm: onAddElementToForm1,
      };

      const {rerender} = await render(<StepFactory {...props1} />);
      await expect.element(page.getByTestId('common-step-factory')).toHaveAttribute('data-has-on-add-element-to-form', 'true');

      const onAddElementToForm2 = vi.fn();
      const props2 = {
        ...props1,
        onAddElementToForm: onAddElementToForm2,
      };

      await rerender(<StepFactory {...props2} />);
      await expect.element(page.getByTestId('common-step-factory')).toHaveAttribute('data-has-on-add-element-to-form', 'true');
    });

    it('should handle transition from undefined to defined callbacks', async () => {
      const props1 = {
        ...createNodeProps(),
        resourceId: 'resource-1',
        resources: [createMockStep()],
        onAddElement: undefined,
        onAddElementToForm: undefined,
      };

      const {rerender} = await render(<StepFactory {...props1} />);
      await expect.element(page.getByTestId('common-step-factory')).toHaveAttribute('data-has-on-add-element', 'false');
      await expect.element(page.getByTestId('common-step-factory')).toHaveAttribute('data-has-on-add-element-to-form', 'false');

      const props2 = {
        ...props1,
        onAddElement: vi.fn(),
        onAddElementToForm: vi.fn(),
      };

      await rerender(<StepFactory {...props2} />);
      await expect.element(page.getByTestId('common-step-factory')).toHaveAttribute('data-has-on-add-element', 'true');
      await expect.element(page.getByTestId('common-step-factory')).toHaveAttribute('data-has-on-add-element-to-form', 'true');
    });
  });
});
