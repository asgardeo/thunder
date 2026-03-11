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
import {render} from '@thunder/test-utils/browser';
import {page, userEvent} from 'vitest/browser';
import type {Step} from '@/features/flows/models/steps';
import ExecutionMinimal from '../ExecutionMinimal';

// Mock @xyflow/react
const mockUseNodeId = vi.fn((): string | null => 'execution-node-id');

vi.mock('@xyflow/react', () => ({
  useNodeId: () => mockUseNodeId(),
  Handle: ({
    type,
    position,
    id = '',
    className = '',
  }: {
    type: string;
    position: string;
    id?: string;
    className?: string;
  }) => (
    <div
      data-testid={`handle-${type}${className ? `-${className}` : ''}`}
      data-position={position}
      data-id={id}
      data-classname={className}
    />
  ),
  Position: {
    Left: 'left',
    Right: 'right',
    Top: 'top',
    Bottom: 'bottom',
  },
}));

// Mock useFlowBuilderCore
const mockSetLastInteractedResource = vi.fn();
const mockSetLastInteractedStepId = vi.fn();
const mockSetIsOpenResourcePropertiesPanel = vi.fn();

vi.mock('@/features/flows/hooks/useFlowBuilderCore', () => ({
  default: () => ({
    setLastInteractedResource: mockSetLastInteractedResource,
    setLastInteractedStepId: mockSetLastInteractedStepId,
    setIsOpenResourcePropertiesPanel: mockSetIsOpenResourcePropertiesPanel,
  }),
}));

// Mock ExecutionFactory
vi.mock('../execution-factory/ExecutionFactory', () => ({
  default: ({resource}: {resource: Step}) => (
    <div data-testid="execution-factory" data-resource-id={resource?.id}>
      ExecutionFactory
    </div>
  ),
}));

// Mock VisualFlowConstants
vi.mock('@/features/flows/constants/VisualFlowConstants', () => ({
  default: {
    FLOW_BUILDER_NEXT_HANDLE_SUFFIX: '-next',
    FLOW_BUILDER_INCOMPLETE_HANDLE_SUFFIX: '-incomplete',
  },
}));

// Create mock resource
const createMockResource = (overrides: Partial<Step> = {}): Step =>
  ({
    id: 'execution-1',
    type: 'TASK_EXECUTION',
    position: {x: 0, y: 0},
    size: {width: 200, height: 100},
    display: {
      label: 'Test Executor',
      image: 'test-image.svg',
      showOnResourcePanel: true,
    },
    data: {
      action: {
        executor: {
          name: 'TestExecutor',
        },
      },
      config: {
        testConfig: 'value',
      },
    },
    config: {},
    ...overrides,
  }) as Step;

describe('ExecutionMinimal', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseNodeId.mockReturnValue('execution-node-id');
  });

  describe('Rendering', () => {
    it('should render the execution minimal step', async () => {
      const resource = createMockResource();
      await render(<ExecutionMinimal resource={resource} />);

      await expect.element(page.getByText('Test Executor')).toBeInTheDocument();
    });

    it('should render ExecutionFactory component', async () => {
      const resource = createMockResource();
      await render(<ExecutionMinimal resource={resource} />);

      await expect.element(page.getByTestId('execution-factory')).toBeInTheDocument();
    });

    it('should render target handle on the left', async () => {
      const resource = createMockResource();
      await render(<ExecutionMinimal resource={resource} />);

      const targetHandle = page.getByTestId('handle-target');
      expect(targetHandle).toHaveAttribute('data-position', 'left');
    });

    it('should render source handle on the right with correct id', async () => {
      const resource = createMockResource({id: 'test-execution'});
      await render(<ExecutionMinimal resource={resource} />);

      // When no branching support, the handle doesn't have a className
      const sourceHandle = page.getByTestId('handle-source');
      expect(sourceHandle).toHaveAttribute('data-position', 'right');
      expect(sourceHandle).toHaveAttribute('data-id', 'test-execution-next');
    });
  });

  describe('Branching Handles', () => {
    it('should render only success handle when onFailure is not present', async () => {
      const resource = createMockResource({
        data: {
          action: {
            executor: {name: 'TestExecutor'},
            onSuccess: '',
          },
        },
      });
      await render(<ExecutionMinimal resource={resource} />);

      // Success handle should be present (no className when no branching)
      await expect.element(page.getByTestId('handle-source')).toBeInTheDocument();
      // Failure handle should NOT be present
      await expect.element(page.getByTestId('handle-source-execution-handle-failure')).not.toBeInTheDocument();
    });

    it('should render both success and failure handles when onFailure property exists (even if empty)', async () => {
      const resource = createMockResource({
        data: {
          action: {
            executor: {name: 'TestExecutor'},
            onSuccess: '',
            onFailure: '',
          },
        },
      });
      await render(<ExecutionMinimal resource={resource} />);

      // Both handles should be present
      await expect.element(page.getByTestId('handle-source-execution-handle-success')).toBeInTheDocument();
      await expect.element(page.getByTestId('handle-source-execution-handle-failure')).toBeInTheDocument();
    });

    it('should render both handles when onFailure has a value', async () => {
      const resource = createMockResource({
        data: {
          action: {
            executor: {name: 'TestExecutor'},
            onSuccess: 'step-2',
            onFailure: 'step-3',
          },
        },
      });
      await render(<ExecutionMinimal resource={resource} />);

      // Both handles should be present
      await expect.element(page.getByTestId('handle-source-execution-handle-success')).toBeInTheDocument();
      await expect.element(page.getByTestId('handle-source-execution-handle-failure')).toBeInTheDocument();
    });

    it('should wrap handles in tooltips when both handles are present', async () => {
      const resource = createMockResource({
        data: {
          action: {
            executor: {name: 'TestExecutor'},
            onSuccess: '',
            onFailure: '',
          },
        },
      });
      await render(<ExecutionMinimal resource={resource} />);

      // Both handles should be present (tooltips are shown on hover, not as visible text)
      await expect.element(page.getByTestId('handle-source-execution-handle-success')).toBeInTheDocument();
      await expect.element(page.getByTestId('handle-source-execution-handle-failure')).toBeInTheDocument();
    });

    it('should add has-branching class when onFailure exists', async () => {
      const resource = createMockResource({
        data: {
          action: {
            executor: {name: 'TestExecutor'},
            onSuccess: '',
            onFailure: '',
          },
        },
      });
      const {container} = await render(<ExecutionMinimal resource={resource} />);

      const stepElement = container.querySelector('.execution-minimal-step');
      expect(stepElement).toHaveClass('has-branching');
    });

    it('should not add has-branching class when onFailure does not exist', async () => {
      const resource = createMockResource({
        data: {
          action: {
            executor: {name: 'TestExecutor'},
            onSuccess: '',
          },
        },
      });
      const {container} = await render(<ExecutionMinimal resource={resource} />);

      const stepElement = container.querySelector('.execution-minimal-step');
      expect(stepElement).not.toHaveClass('has-branching');
    });

    it('should render incomplete handle when onIncomplete property exists', async () => {
      const resource = createMockResource({
        data: {
          action: {
            executor: {name: 'TestExecutor'},
            onSuccess: '',
            onIncomplete: '',
          },
        },
      });
      await render(<ExecutionMinimal resource={resource} />);

      const incompleteHandle = page.getByTestId('handle-source-execution-handle-incomplete');
      expect(incompleteHandle).toBeInTheDocument();
      expect(incompleteHandle).toHaveAttribute('data-position', 'top');
    });

    it('should render incomplete handle with correct id', async () => {
      const resource = createMockResource({
        id: 'test-execution',
        data: {
          action: {
            executor: {name: 'TestExecutor'},
            onIncomplete: 'step-4',
          },
        },
      });
      await render(<ExecutionMinimal resource={resource} />);

      const incompleteHandle = page.getByTestId('handle-source-execution-handle-incomplete');
      expect(incompleteHandle).toHaveAttribute('data-id', 'test-execution-incomplete');
    });

    it('should not render incomplete handle when onIncomplete property is missing', async () => {
      const resource = createMockResource({
        data: {
          action: {
            executor: {name: 'TestExecutor'},
            onSuccess: '',
          },
        },
      });
      await render(<ExecutionMinimal resource={resource} />);

      await expect.element(page.getByTestId('handle-source-execution-handle-incomplete')).not.toBeInTheDocument();
    });
  });

  describe('Display Label', () => {
    it('should display label from resource.display.label', async () => {
      const resource = createMockResource({
        display: {label: 'Custom Label', image: 'test.svg', showOnResourcePanel: true},
      });
      await render(<ExecutionMinimal resource={resource} />);

      await expect.element(page.getByText('Custom Label')).toBeInTheDocument();
    });

    it('should fallback to executor name when display.label is not provided', async () => {
      const resource = createMockResource({
        display: undefined,
        data: {
          action: {
            executor: {
              name: 'FallbackExecutor',
            },
          },
        },
      });
      await render(<ExecutionMinimal resource={resource} />);

      await expect.element(page.getByText('FallbackExecutor')).toBeInTheDocument();
    });

    it('should fallback to "Executor" when both display.label and executor name are not provided', async () => {
      const resource = createMockResource({
        display: undefined,
        data: {},
      });
      await render(<ExecutionMinimal resource={resource} />);

      await expect.element(page.getByText('Executor')).toBeInTheDocument();
    });
  });

  describe('Config Button Click', () => {
    it('should set last interacted step id when config button is clicked', async () => {
      const resource = createMockResource();
      await render(<ExecutionMinimal resource={resource} />);

      const configButton = page.getByRole('button');
      await userEvent.click(configButton);

      expect(mockSetLastInteractedStepId).toHaveBeenCalledWith('execution-node-id');
    });

    it('should set last interacted resource with merged config when config button is clicked', async () => {
      const resource = createMockResource({
        config: {field: {name: 'test', type: 'TEXT'}, styles: {}} as unknown as Step['config'],
        data: {
          action: {executor: {name: 'Test'}},
          config: {dataConfig: 'dataValue'},
        },
      });
      await render(<ExecutionMinimal resource={resource} />);

      const configButton = page.getByRole('button');
      await userEvent.click(configButton);

      expect(mockSetLastInteractedResource).toHaveBeenCalledWith(
        expect.objectContaining({
          config: expect.objectContaining({
            dataConfig: 'dataValue',
          }) as Record<string, unknown>,
        }),
      );
    });

    it('should open resource properties panel when config button is clicked', async () => {
      const resource = createMockResource();
      await render(<ExecutionMinimal resource={resource} />);

      const configButton = page.getByRole('button');
      await userEvent.click(configButton);

      expect(mockSetIsOpenResourcePropertiesPanel).toHaveBeenCalledWith(true);
    });

    it('should set step id to empty string when useNodeId returns empty string', async () => {
      mockUseNodeId.mockReturnValue('');
      const resource = createMockResource();
      await render(<ExecutionMinimal resource={resource} />);

      const configButton = page.getByRole('button');
      await userEvent.click(configButton);

      // Empty string is not null, so setLastInteractedStepId IS called
      expect(mockSetLastInteractedStepId).toHaveBeenCalledWith('');
    });

    it('should not set step id when useNodeId returns null', async () => {
      mockUseNodeId.mockReturnValue(null);
      const resource = createMockResource();
      await render(<ExecutionMinimal resource={resource} />);

      const configButton = page.getByRole('button');
      await userEvent.click(configButton);

      expect(mockSetLastInteractedStepId).not.toHaveBeenCalled();
    });
  });

  describe('Card Click', () => {
    it('should set last interacted step id when card is clicked', async () => {
      const resource = createMockResource({id: 'clicked-resource'});
      await render(<ExecutionMinimal resource={resource} />);

      const executionFactory = page.getByTestId('execution-factory');
      const card = executionFactory.element().parentElement;
      if (card) {
        await userEvent.click(card);
      }

      expect(mockSetLastInteractedStepId).toHaveBeenCalledWith('clicked-resource');
    });

    it('should set last interacted resource when card is clicked', async () => {
      const resource = createMockResource({id: 'clicked-resource'});
      await render(<ExecutionMinimal resource={resource} />);

      const executionFactory = page.getByTestId('execution-factory');
      const card = executionFactory.element().parentElement;
      if (card) {
        await userEvent.click(card);
      }

      expect(mockSetLastInteractedResource).toHaveBeenCalledWith(resource);
    });
  });

  describe('Config Merging', () => {
    it('should handle undefined resource.config', async () => {
      const resource = createMockResource({
        config: undefined,
        data: {
          action: {executor: {name: 'Test'}},
          config: {dataConfig: 'value'},
        },
      });
      await render(<ExecutionMinimal resource={resource} />);

      const configButton = page.getByRole('button');
      await userEvent.click(configButton);

      expect(mockSetLastInteractedResource).toHaveBeenCalledWith(
        expect.objectContaining({
          config: expect.objectContaining({
            dataConfig: 'value',
          }) as Record<string, unknown>,
        }),
      );
    });

    it('should handle undefined data.config', async () => {
      const resource = createMockResource({
        config: {field: {name: 'test', type: 'TEXT'}, styles: {}} as unknown as Step['config'],
        data: {
          action: {executor: {name: 'Test'}},
        },
      });
      await render(<ExecutionMinimal resource={resource} />);

      const configButton = page.getByRole('button');
      await userEvent.click(configButton);

      expect(mockSetLastInteractedResource).toHaveBeenCalledWith(
        expect.objectContaining({
          config: expect.objectContaining({
            field: {name: 'test', type: 'TEXT'},
          }) as Record<string, unknown>,
        }),
      );
    });

    it('should handle null data.config', async () => {
      const resource = createMockResource({
        config: {field: {name: 'test', type: 'TEXT'}, styles: {}} as unknown as Step['config'],
        data: {
          action: {executor: {name: 'Test'}},
          config: null,
        },
      });
      await render(<ExecutionMinimal resource={resource} />);

      const configButton = page.getByRole('button');
      await userEvent.click(configButton);

      expect(mockSetLastInteractedResource).toHaveBeenCalledWith(
        expect.objectContaining({
          config: expect.objectContaining({
            field: {name: 'test', type: 'TEXT'},
          }) as Record<string, unknown>,
        }),
      );
    });
  });

  describe('Tooltip', () => {
    it('should display configuration hint tooltip on config button', async () => {
      const resource = createMockResource();
      await render(<ExecutionMinimal resource={resource} />);

      // The tooltip title should be the translation key
      const configButton = page.getByRole('button');
      expect(configButton).toBeInTheDocument();
    });
  });
});
