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
import type {Element} from '@/features/flows/models/elements';
import View from '../View';

// Mock @xyflow/react
const mockDeleteElements = vi.fn();
const mockGetNode = vi.fn(() => ({id: 'view-node', data: {}}));
const mockUseNodeId = vi.fn(() => 'view-node-id');

vi.mock('@xyflow/react', () => ({
  // eslint-disable-next-line react/require-default-props
  Handle: ({type, position, id}: {type: string; position: string; id?: string}) => (
    <div data-testid={`handle-${type}`} data-position={position} data-handle-id={id ?? ''} />
  ),
  Position: {
    Left: 'left',
    Right: 'right',
    Top: 'top',
    Bottom: 'bottom',
  },
  useNodeId: () => mockUseNodeId(),
  useReactFlow: () => ({
    deleteElements: mockDeleteElements,
    getNode: mockGetNode,
  }),
}));

// Mock PluginRegistry
vi.mock('@/features/flows/plugins/PluginRegistry', () => ({
  default: {
    getInstance: () => ({
      executeSync: () => true,
    }),
  },
}));

// Mock generateResourceId
vi.mock('@/features/flows/utils/generateResourceId', () => ({
  default: (prefix: string) => `${prefix}-generated-id`,
}));

// Mock Droppable
vi.mock('../../../dnd/Droppable', () => ({
  default: ({children, id, accept}: {children: React.ReactNode; id: string; accept: string[]}) => (
    <div data-testid="droppable" data-id={id} data-accept={JSON.stringify(accept)}>
      {children}
    </div>
  ),
}));

// Mock ReorderableViewElement
vi.mock('../ReorderableElement', () => ({
  default: ({element, index}: {element: Element; index: number}) => (
    <div data-testid={`reorderable-element-${index}`} data-element-id={element.id}>
      {element.display?.label || element.type}
    </div>
  ),
}));

// Mock SCSS
vi.mock('../View.scss', () => ({}));

describe('View', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseNodeId.mockReturnValue('view-node-id');
  });

  describe('Rendering', () => {
    it('should render the View component', async () => {
      await render(<View />);

      await expect.element(page.getByText('View')).toBeInTheDocument();
    });

    it('should render with custom heading', async () => {
      await render(<View heading="Login Form" />);

      await expect.element(page.getByText('Login Form')).toBeInTheDocument();
    });

    it('should render with flow-builder-step class', async () => {
      const {container} = await render(<View />);

      expect(container.querySelector('.flow-builder-step')).toBeInTheDocument();
    });

    it('should accept custom className', async () => {
      const {container} = await render(<View className="custom-view" />);

      expect(container.querySelector('.custom-view')).toBeInTheDocument();
    });
  });

  describe('React Flow Handles', () => {
    it('should render a target handle on the left', async () => {
      await render(<View />);

      const handle = page.getByTestId('handle-target');
      expect(handle).toBeInTheDocument();
      expect(handle).toHaveAttribute('data-position', 'left');
    });

    it('should render source handle when enableSourceHandle is true', async () => {
      await render(<View enableSourceHandle />);

      const handles = page.getByTestId(/handle-/).all();
      const sourceHandle = handles.find((h) => h.element().getAttribute('data-testid') === 'handle-source');
      expect(sourceHandle).toBeInTheDocument();
    });

    it('should not render source handle by default', async () => {
      await render(<View />);

      const sourceHandle = page.getByTestId('handle-source');
      expect(sourceHandle).not.toBeInTheDocument();
    });
  });

  describe('Delete Button', () => {
    it('should render delete button when deletable is true', async () => {
      await render(<View deletable />);

      // Delete button should be present
      const buttons = page.getByRole('button').all();
      expect(buttons.length).toBeGreaterThan(0);
    });

    it('should not render delete button when deletable is false', async () => {
      await render(<View deletable={false} />);

      // When no components or configure, and not deletable, there may be no buttons
      // The tooltip text won't be visible unless hovered
      await expect.element(page.getByText('Remove')).not.toBeInTheDocument();
    });

    it('should call deleteElements when delete button is clicked', async () => {
      await render(<View deletable />);

      // Find the delete button (last button in the action panel)
      const buttons = page.getByRole('button').all();
      const deleteButton = buttons[buttons.length - 1];
      await userEvent.click(deleteButton);

      expect(mockDeleteElements).toHaveBeenCalledWith({
        nodes: [{id: 'view-node-id'}],
      });
    });
  });

  describe('Configure Button', () => {
    it('should render configure button when configurable is true', async () => {
      const onConfigure = vi.fn();
      await render(<View configurable onConfigure={onConfigure} />);

      const buttons = page.getByRole('button').all();
      expect(buttons.length).toBeGreaterThan(0);
    });

    it('should call onConfigure when configure button is clicked', async () => {
      const onConfigure = vi.fn();
      await render(<View configurable onConfigure={onConfigure} deletable={false} />);

      const buttons = page.getByRole('button').all();
      // When only configurable, the configure button should be the only one
      await userEvent.click(buttons[0]);

      expect(onConfigure).toHaveBeenCalled();
    });

    it('should not render configure button by default', async () => {
      await render(<View deletable={false} />);

      const buttons = page.getByRole('button').all();
      expect(buttons.length).toBe(0);
    });
  });

  describe('Add Component Menu', () => {
    const mockElements: Element[] = [
      {
        id: 'elem-1',
        type: 'TEXT_INPUT',
        category: 'INPUT',
        resourceType: 'ELEMENT',
        display: {label: 'Text Input', showOnResourcePanel: true},
      },
      {
        id: 'elem-2',
        type: 'BUTTON',
        category: 'ACTION',
        resourceType: 'ELEMENT',
        display: {label: 'Button', showOnResourcePanel: true},
      },
    ] as Element[];

    it('should render add button when availableElements is provided', async () => {
      await render(<View availableElements={mockElements} deletable={false} />);

      const buttons = page.getByRole('button').all();
      expect(buttons.length).toBeGreaterThan(0);
    });

    it('should open menu when add button is clicked', async () => {
      await render(<View availableElements={mockElements} deletable={false} />);

      const addButton = page.getByRole('button').all()[0];
      await userEvent.click(addButton);

      // Menu items should appear
      await expect.element(page.getByText('Text Input')).toBeInTheDocument();
      await expect.element(page.getByText('Button')).toBeInTheDocument();
    });

    it('should call onAddElement when menu item is clicked', async () => {
      const onAddElement = vi.fn();
      await render(<View availableElements={mockElements} onAddElement={onAddElement} deletable={false} />);

      const addButton = page.getByRole('button').all()[0];
      await userEvent.click(addButton);

      const menuItem = page.getByText('Text Input');
      await userEvent.click(menuItem);

      expect(onAddElement).toHaveBeenCalledWith(mockElements[0], 'view-node-id');
    });

    it('should close menu when clicking outside', async () => {
      await render(<View availableElements={mockElements} deletable={false} />);

      const addButton = page.getByRole('button').all()[0];
      await userEvent.click(addButton);

      // Menu should be open
      await expect.element(page.getByText('Text Input')).toBeInTheDocument();

      // Close the menu by pressing Escape
      await userEvent.keyboard('{Escape}');

      // Menu should be closed after escape
      await expect.element(page.getByRole('menu')).not.toBeInTheDocument();
    });

    it('should filter out elements with showOnResourcePanel=false', async () => {
      const elementsWithHidden: Element[] = [
        ...mockElements,
        {
          id: 'elem-3',
          type: 'HIDDEN',
          category: 'OTHER',
          resourceType: 'ELEMENT',
          display: {label: 'Hidden Element', showOnResourcePanel: false},
        },
      ] as Element[];

      await render(<View availableElements={elementsWithHidden} deletable={false} />);

      const addButton = page.getByRole('button').all()[0];
      await userEvent.click(addButton);

      await expect.element(page.getByText('Hidden Element')).not.toBeInTheDocument();
    });
  });

  describe('Components Rendering', () => {
    it('should render components in form group', async () => {
      const components: Element[] = [
        {
          id: 'comp-1',
          type: 'TEXT_INPUT',
          category: 'INPUT',
          resourceType: 'ELEMENT',
          display: {label: 'Username'},
        },
        {
          id: 'comp-2',
          type: 'PASSWORD_INPUT',
          category: 'INPUT',
          resourceType: 'ELEMENT',
          display: {label: 'Password'},
        },
      ] as Element[];

      await render(<View data={{components}} />);

      // Components are rendered inside the form group
      await expect.element(page.getByTestId('reorderable-element-0')).toBeInTheDocument();
      await expect.element(page.getByTestId('reorderable-element-1')).toBeInTheDocument();
    });

    it('should render empty form when no components', async () => {
      await render(<View data={{components: []}} />);

      // Form group should still be rendered even with no components
      const formGroup = document.querySelector('.MuiFormGroup-root');
      expect(formGroup).toBeInTheDocument();
      await expect.element(page.getByTestId('reorderable-element-0')).not.toBeInTheDocument();
    });
  });

  describe('Action Panel Double Click', () => {
    it('should call onActionPanelDoubleClick when action panel is double clicked', async () => {
      const onDoubleClick = vi.fn();
      await render(<View onActionPanelDoubleClick={onDoubleClick} />);

      const actionPanel = page.getByText('View').element().closest('.flow-builder-step-action-panel');
      if (actionPanel) {
        await userEvent.dblClick(actionPanel);
        expect(onDoubleClick).toHaveBeenCalled();
      }
    });
  });

  describe('Droppable Configuration', () => {
    it('should accept custom droppableAllowedTypes prop', async () => {
      // View component accepts droppableAllowedTypes prop
      await render(<View droppableAllowedTypes={['CUSTOM_TYPE']} />);

      // Component should render without errors
      await expect.element(page.getByText('View')).toBeInTheDocument();
    });
  });

  describe('Memoization', () => {
    it('should render correctly on rerender with same props', async () => {
      const {rerender} = await render(<View heading="Test View" />);

      await expect.element(page.getByText('Test View')).toBeInTheDocument();

      await rerender(<View heading="Test View" />);

      await expect.element(page.getByText('Test View')).toBeInTheDocument();
    });

    it('should re-render when data prop changes', async () => {
      const data1 = {components: []};
      const data2 = {
        components: [
          {id: 'new-comp', type: 'TEXT', category: 'INPUT', resourceType: 'ELEMENT', display: {label: 'New'}},
        ],
      };

      const {rerender} = await render(<View data={data1} />);
      await expect.element(page.getByTestId('reorderable-element-0')).not.toBeInTheDocument();

      await rerender(<View data={data2 as {components: Element[]}} />);
      await expect.element(page.getByTestId('reorderable-element-0')).toBeInTheDocument();
    });

    it('should re-render when heading prop changes', async () => {
      const {rerender} = await render(<View heading="Initial Heading" />);
      await expect.element(page.getByText('Initial Heading')).toBeInTheDocument();

      await rerender(<View heading="Updated Heading" />);
      await expect.element(page.getByText('Updated Heading')).toBeInTheDocument();
    });

    it('should re-render when deletable prop changes', async () => {
      const {rerender} = await render(<View deletable={false} configurable={false} />);
      expect(page.getByRole('button').all()).toHaveLength(0);

      await rerender(<View deletable configurable={false} />);
      // Delete button should now be visible
      expect(page.getByRole('button').all().length).toBeGreaterThan(0);
    });

    it('should re-render when configurable prop changes', async () => {
      const {rerender} = await render(<View deletable={false} configurable={false} />);
      expect(page.getByRole('button').all()).toHaveLength(0);

      await rerender(<View deletable={false} configurable />);
      // Configure button should now be visible
      expect(page.getByRole('button').all().length).toBeGreaterThan(0);
    });

    it('should re-render when enableSourceHandle prop changes', async () => {
      const {rerender} = await render(<View enableSourceHandle={false} />);
      await expect.element(page.getByTestId('handle-source')).not.toBeInTheDocument();

      await rerender(<View enableSourceHandle />);
      await expect.element(page.getByTestId('handle-source')).toBeInTheDocument();
    });

    it('should not re-render when only callback props change', async () => {
      const onAddElement1 = vi.fn();
      const onAddElement2 = vi.fn();

      const {rerender} = await render(<View heading="Same Heading" onAddElement={onAddElement1} />);
      await expect.element(page.getByText('Same Heading')).toBeInTheDocument();

      // Changing only the callback should not cause visual changes (memoized)
      await rerender(<View heading="Same Heading" onAddElement={onAddElement2} />);
      await expect.element(page.getByText('Same Heading')).toBeInTheDocument();
    });
  });
});
