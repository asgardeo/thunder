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
import type {ReactNode} from 'react';
import {ReactFlowProvider} from '@xyflow/react';
import CommonStepFactory from '../CommonStepFactory';
import {StepTypes, StaticStepTypes, type Step} from '../../../../models/steps';
import type {Resources} from '../../../../models/resources';
import type {Element} from '../../../../models/elements';

// Mock step components
vi.mock('../view/View', () => ({
  default: ({resources, data}: {resources: Step[]; data: unknown}) => (
    <div data-testid="view-step" data-resource-count={resources.length} data-has-data={!!data}>
      View Step
    </div>
  ),
}));

vi.mock('../rule/Rule', () => ({
  default: ({resources, data}: {resources: Step[]; data: unknown}) => (
    <div data-testid="rule-step" data-resource-count={resources.length} data-has-data={!!data}>
      Rule Step
    </div>
  ),
}));

vi.mock('../execution/Execution', () => ({
  default: ({resources, data}: {resources: Step[]; data: unknown}) => (
    <div data-testid="execution-step" data-resource-count={resources.length} data-has-data={!!data}>
      Execution Step
    </div>
  ),
}));

vi.mock('../end/End', () => ({
  default: ({resources, data}: {resources: Step[]; data: unknown}) => (
    <div data-testid="end-step" data-resource-count={resources.length} data-has-data={!!data}>
      End Step
    </div>
  ),
}));

describe('CommonStepFactory', () => {
  const createWrapper = () => {
    function Wrapper({children}: {children: ReactNode}) {
      return <ReactFlowProvider>{children}</ReactFlowProvider>;
    }
    return Wrapper;
  };

  const createMockStep = (overrides: Partial<Step> = {}): Step =>
    ({
      id: 'step-1',
      type: StepTypes.View,
      category: 'STEP',
      config: {},
      ...overrides,
    }) as Step;

  const defaultNodeProps = {
    id: 'node-1',
    type: 'custom',
    data: {components: []},
    positionAbsoluteX: 0,
    positionAbsoluteY: 0,
    zIndex: 0,
    isConnectable: true,
    xPos: 0,
    yPos: 0,
    dragging: false,
    selected: false,
    dragHandle: undefined,
    sourcePosition: undefined,
    targetPosition: undefined,
    draggable: true,
    selectable: true,
    deletable: true,
  };

  describe('View Step', () => {
    it('should render View component for VIEW step type', async () => {
      const viewStep = createMockStep({type: StepTypes.View});

      await render(
        <CommonStepFactory {...defaultNodeProps} resourceId="resource-1" resources={[viewStep]} />,
        {wrapper: createWrapper()},
      );

      await expect.element(page.getByTestId('view-step')).toBeInTheDocument();
    });

    it('should pass resources array to View component', async () => {
      const viewStep1 = createMockStep({id: 'step-1', type: StepTypes.View});
      const viewStep2 = createMockStep({id: 'step-2', type: StepTypes.View});

      await render(
        <CommonStepFactory {...defaultNodeProps} resourceId="resource-1" resources={[viewStep1, viewStep2]} />,
        {wrapper: createWrapper()},
      );

      await expect.element(page.getByTestId('view-step')).toHaveAttribute('data-resource-count', '2');
    });

    it('should pass data to View component', async () => {
      const viewStep = createMockStep({type: StepTypes.View});

      await render(
        <CommonStepFactory
          {...defaultNodeProps}
          resourceId="resource-1"
          resources={[viewStep]}
          data={{components: [{id: 'comp-1'}]}}
        />,
        {wrapper: createWrapper()},
      );

      await expect.element(page.getByTestId('view-step')).toHaveAttribute('data-has-data', 'true');
    });

    it('should pass availableElements from allResources', async () => {
      const viewStep = createMockStep({type: StepTypes.View});
      const allResources: Resources = {
        steps: [],
        elements: [{id: 'element-1'} as Element],
        widgets: [],
        templates: [],
        executors: [],
      };

      await render(
        <CommonStepFactory
          {...defaultNodeProps}
          resourceId="resource-1"
          resources={[viewStep]}
          allResources={allResources}
        />,
        {wrapper: createWrapper()},
      );

      await expect.element(page.getByTestId('view-step')).toBeInTheDocument();
    });

    it('should pass onAddElement callback', async () => {
      const viewStep = createMockStep({type: StepTypes.View});
      const onAddElement = vi.fn();

      await render(
        <CommonStepFactory
          {...defaultNodeProps}
          resourceId="resource-1"
          resources={[viewStep]}
          onAddElement={onAddElement}
        />,
        {wrapper: createWrapper()},
      );

      await expect.element(page.getByTestId('view-step')).toBeInTheDocument();
    });

    it('should pass onAddElementToForm callback', async () => {
      const viewStep = createMockStep({type: StepTypes.View});
      const onAddElementToForm = vi.fn();

      await render(
        <CommonStepFactory
          {...defaultNodeProps}
          resourceId="resource-1"
          resources={[viewStep]}
          onAddElementToForm={onAddElementToForm}
        />,
        {wrapper: createWrapper()},
      );

      await expect.element(page.getByTestId('view-step')).toBeInTheDocument();
    });
  });

  describe('Rule Step', () => {
    it('should render Rule component for RULE step type', async () => {
      const ruleStep = createMockStep({type: StepTypes.Rule});

      await render(
        <CommonStepFactory {...defaultNodeProps} resourceId="resource-1" resources={[ruleStep]} />,
        {wrapper: createWrapper()},
      );

      await expect.element(page.getByTestId('rule-step')).toBeInTheDocument();
    });

    it('should pass resources to Rule component', async () => {
      const ruleStep = createMockStep({type: StepTypes.Rule});

      await render(
        <CommonStepFactory {...defaultNodeProps} resourceId="resource-1" resources={[ruleStep]} />,
        {wrapper: createWrapper()},
      );

      await expect.element(page.getByTestId('rule-step')).toHaveAttribute('data-resource-count', '1');
    });

    it('should pass data to Rule component', async () => {
      const ruleStep = createMockStep({type: StepTypes.Rule});

      await render(
        <CommonStepFactory
          {...defaultNodeProps}
          resourceId="resource-1"
          resources={[ruleStep]}
          data={{condition: 'true'}}
        />,
        {wrapper: createWrapper()},
      );

      await expect.element(page.getByTestId('rule-step')).toHaveAttribute('data-has-data', 'true');
    });
  });

  describe('Execution Step', () => {
    it('should render Execution component for EXECUTION step type', async () => {
      const executionStep = createMockStep({type: StepTypes.Execution});

      await render(
        <CommonStepFactory {...defaultNodeProps} resourceId="resource-1" resources={[executionStep]} />,
        {wrapper: createWrapper()},
      );

      await expect.element(page.getByTestId('execution-step')).toBeInTheDocument();
    });

    it('should pass resources to Execution component', async () => {
      const executionStep = createMockStep({type: StepTypes.Execution});

      await render(
        <CommonStepFactory {...defaultNodeProps} resourceId="resource-1" resources={[executionStep]} />,
        {wrapper: createWrapper()},
      );

      await expect.element(page.getByTestId('execution-step')).toHaveAttribute('data-resource-count', '1');
    });

    it('should pass data to Execution component', async () => {
      const executionStep = createMockStep({type: StepTypes.Execution});

      await render(
        <CommonStepFactory
          {...defaultNodeProps}
          resourceId="resource-1"
          resources={[executionStep]}
          data={{executor: 'test'}}
        />,
        {wrapper: createWrapper()},
      );

      await expect.element(page.getByTestId('execution-step')).toHaveAttribute('data-has-data', 'true');
    });
  });

  describe('End Step', () => {
    it('should render End component for END step type', async () => {
      const endStep = createMockStep({type: StepTypes.End});

      await render(
        <CommonStepFactory {...defaultNodeProps} resourceId="resource-1" resources={[endStep]} />,
        {wrapper: createWrapper()},
      );

      await expect.element(page.getByTestId('end-step')).toBeInTheDocument();
    });

    it('should pass resources to End component', async () => {
      const endStep = createMockStep({type: StepTypes.End});

      await render(
        <CommonStepFactory {...defaultNodeProps} resourceId="resource-1" resources={[endStep]} />,
        {wrapper: createWrapper()},
      );

      await expect.element(page.getByTestId('end-step')).toHaveAttribute('data-resource-count', '1');
    });

    it('should pass data to End component', async () => {
      const endStep = createMockStep({type: StepTypes.End});

      await render(
        <CommonStepFactory
          {...defaultNodeProps}
          resourceId="resource-1"
          resources={[endStep]}
          data={{status: 'complete'}}
        />,
        {wrapper: createWrapper()},
      );

      await expect.element(page.getByTestId('end-step')).toHaveAttribute('data-has-data', 'true');
    });
  });

  describe('Unknown Step Type', () => {
    it('should return null for unknown step type', async () => {
      const unknownStep = createMockStep({type: 'UNKNOWN' as StepTypes});

      const {container} = await render(
        <CommonStepFactory {...defaultNodeProps} resourceId="resource-1" resources={[unknownStep]} />,
        {wrapper: createWrapper()},
      );

      expect(container.firstChild).toBeNull();
    });
  });

  describe('Start Step', () => {
    it('should return null for START step type (not handled in factory)', async () => {
      const startStep = createMockStep({type: StaticStepTypes.Start as unknown as StepTypes});

      const {container} = await render(
        <CommonStepFactory {...defaultNodeProps} resourceId="resource-1" resources={[startStep]} />,
        {wrapper: createWrapper()},
      );

      expect(container.firstChild).toBeNull();
    });
  });

  describe('Default Props', () => {
    it('should work with undefined allResources', async () => {
      const viewStep = createMockStep({type: StepTypes.View});

      await render(
        <CommonStepFactory {...defaultNodeProps} resourceId="resource-1" resources={[viewStep]} />,
        {wrapper: createWrapper()},
      );

      await expect.element(page.getByTestId('view-step')).toBeInTheDocument();
    });

    it('should work with undefined onAddElement', async () => {
      const viewStep = createMockStep({type: StepTypes.View});

      await render(
        <CommonStepFactory {...defaultNodeProps} resourceId="resource-1" resources={[viewStep]} />,
        {wrapper: createWrapper()},
      );

      await expect.element(page.getByTestId('view-step')).toBeInTheDocument();
    });

    it('should work with undefined onAddElementToForm', async () => {
      const viewStep = createMockStep({type: StepTypes.View});

      await render(
        <CommonStepFactory {...defaultNodeProps} resourceId="resource-1" resources={[viewStep]} />,
        {wrapper: createWrapper()},
      );

      await expect.element(page.getByTestId('view-step')).toBeInTheDocument();
    });
  });

  describe('Rest Props Spreading', () => {
    it('should spread additional props to step components', async () => {
      const viewStep = createMockStep({type: StepTypes.View});

      await render(
        <CommonStepFactory
          {...defaultNodeProps}
          resourceId="resource-1"
          resources={[viewStep]}
          data-custom="value"
        />,
        {wrapper: createWrapper()},
      );

      await expect.element(page.getByTestId('view-step')).toBeInTheDocument();
    });
  });
});
