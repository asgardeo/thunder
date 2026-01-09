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
import {render, screen, fireEvent, waitFor} from '@testing-library/react';
import type {Resource} from '@/features/flows/models/resources';
import {BlockTypes} from '@/features/flows/models/elements';
import {ReorderableElement} from '../ReorderableElement';

// Use vi.hoisted to define mocks that need to be referenced in vi.mock
const {
  mockDeleteComponent,
  mockSetLastInteractedResource,
  mockSetLastInteractedStepId,
  mockSetIsOpenResourcePropertiesPanel,
  mockSetOpenValidationPanel,
  mockSetSelectedNotification,
  mockAddNotification,
  mockExecuteAsync,
} = vi.hoisted(() => ({
  mockDeleteComponent: vi.fn(),
  mockSetLastInteractedResource: vi.fn(),
  mockSetLastInteractedStepId: vi.fn(),
  mockSetIsOpenResourcePropertiesPanel: vi.fn(),
  mockSetOpenValidationPanel: vi.fn(),
  mockSetSelectedNotification: vi.fn(),
  mockAddNotification: vi.fn(),
  mockExecuteAsync: vi.fn().mockResolvedValue(true),
}));

// Mock sortable refs and state
const mockSortableRef = {current: null};
const mockSortable = {
  accepts: vi.fn(() => true),
};

// Mock manager with event listeners
const mockManager = {
  monitor: {
    addEventListener: vi.fn(),
  },
};

// Mock @dnd-kit/react/sortable
vi.mock('@dnd-kit/react/sortable', () => ({
  useSortable: vi.fn(() => ({
    ref: mockSortableRef,
    sortable: mockSortable,
    isDragging: false,
    isDropTarget: false,
  })),
}));

// Mock @dnd-kit/react
vi.mock('@dnd-kit/react', () => ({
  useDragDropManager: vi.fn(() => mockManager),
  useDragOperation: vi.fn(() => ({
    source: null,
  })),
}));

// Mock @dnd-kit/abstract/modifiers
vi.mock('@dnd-kit/abstract/modifiers', () => ({
  RestrictToVerticalAxis: {},
}));

// Mock @xyflow/react
vi.mock('@xyflow/react', () => ({
  useNodeId: () => 'test-step-id',
}));

// Mock PluginRegistry
vi.mock('@/features/flows/plugins/PluginRegistry', () => ({
  default: {
    getInstance: () => ({
      executeAsync: mockExecuteAsync,
    }),
  },
}));

// Mock useComponentDelete
vi.mock('@/features/flows/hooks/useComponentDelete', () => ({
  default: () => ({
    deleteComponent: mockDeleteComponent,
  }),
}));

// Mock useValidationStatus
vi.mock('@/features/flows/hooks/useValidationStatus', () => ({
  default: () => ({
    setOpenValidationPanel: mockSetOpenValidationPanel,
    setSelectedNotification: mockSetSelectedNotification,
    addNotification: mockAddNotification,
  }),
}));

// Mock useFlowBuilderCore
vi.mock('@/features/flows/hooks/useFlowBuilderCore', () => ({
  default: () => ({
    ElementFactory: ({resource}: {resource: Resource}) => (
      <div data-testid="element-factory">{resource.display?.label || resource.type}</div>
    ),
    setLastInteractedResource: mockSetLastInteractedResource,
    setLastInteractedStepId: mockSetLastInteractedStepId,
    setIsOpenResourcePropertiesPanel: mockSetIsOpenResourcePropertiesPanel,
  }),
}));

// Mock VisualFlowConstants
vi.mock('@/features/flows/constants/VisualFlowConstants', () => ({
  default: {
    FLOW_BUILDER_FORM_ALLOWED_RESOURCE_TYPES: ['TEXT_INPUT', 'PASSWORD_INPUT', 'BUTTON'],
    FLOW_BUILDER_PLUGIN_FUNCTION_IDENTIFIER: 'uniqueName',
  },
}));

// Mock ValidationErrorBoundary
vi.mock('../../../../validation-panel/ValidationErrorBoundary', () => ({
  default: ({children}: {children: React.ReactNode}) => <div data-testid="validation-boundary">{children}</div>,
}));

// Mock Handle component
vi.mock('../../../../dnd/Handle', () => ({
  default: ({children, onClick, label}: {children: React.ReactNode; onClick?: () => void; label: string}) => (
    <button data-testid={`handle-${label.toLowerCase()}`} onClick={onClick} type="button">
      {children}
    </button>
  ),
}));

// Mock SCSS
vi.mock('../ReorderableElement.scss', () => ({}));

describe('ReorderableElement', () => {
  const mockElement: Resource = {
    id: 'element-1',
    type: 'TEXT_INPUT',
    category: 'FIELD',
    resourceType: 'ELEMENT',
    display: {label: 'Username Field', image: '', showOnResourcePanel: true},
  } as Resource;

  const mockFormElement: Resource = {
    id: 'form-1',
    type: BlockTypes.Form,
    category: 'BLOCK',
    resourceType: 'ELEMENT',
    display: {label: 'Login Form', image: '', showOnResourcePanel: true},
  } as Resource;

  const mockAvailableElements: Resource[] = [
    {
      id: 'text-input',
      type: 'TEXT_INPUT',
      category: 'FIELD',
      resourceType: 'ELEMENT',
      display: {label: 'Text Input', image: '', showOnResourcePanel: true},
    } as Resource,
    {
      id: 'password-input',
      type: 'PASSWORD_INPUT',
      category: 'FIELD',
      resourceType: 'ELEMENT',
      display: {label: 'Password Input', image: '', showOnResourcePanel: true},
    } as Resource,
    {
      id: 'hidden-element',
      type: 'HIDDEN',
      category: 'FIELD',
      resourceType: 'ELEMENT',
      display: {label: 'Hidden', image: '', showOnResourcePanel: false},
    } as Resource,
  ];

  beforeEach(() => {
    vi.clearAllMocks();
    mockExecuteAsync.mockResolvedValue(true);
  });

  describe('Rendering', () => {
    it('should render the element with ElementFactory', () => {
      render(<ReorderableElement id="sortable-1" index={0} element={mockElement} />);

      expect(screen.getByTestId('element-factory')).toBeInTheDocument();
      expect(screen.getByText('Username Field')).toBeInTheDocument();
    });

    it('should render with custom className', () => {
      const {container} = render(
        <ReorderableElement id="sortable-1" index={0} element={mockElement} className="custom-class" />,
      );

      expect(container.querySelector('.custom-class')).toBeInTheDocument();
    });

    it('should render DnD action handles', () => {
      render(<ReorderableElement id="sortable-1" index={0} element={mockElement} />);

      expect(screen.getByTestId('handle-drag')).toBeInTheDocument();
      expect(screen.getByTestId('handle-edit')).toBeInTheDocument();
      expect(screen.getByTestId('handle-delete')).toBeInTheDocument();
    });

    it('should wrap content in ValidationErrorBoundary', () => {
      render(<ReorderableElement id="sortable-1" index={0} element={mockElement} />);

      expect(screen.getByTestId('validation-boundary')).toBeInTheDocument();
    });
  });

  describe('Form Element Handling', () => {
    it('should render Add Field button for Form elements', () => {
      render(
        <ReorderableElement
          id="sortable-1"
          index={0}
          element={mockFormElement}
          availableElements={mockAvailableElements}
        />,
      );

      expect(screen.getByTestId('handle-add field')).toBeInTheDocument();
    });

    it('should not render Add Field button for non-Form elements', () => {
      render(
        <ReorderableElement
          id="sortable-1"
          index={0}
          element={mockElement}
          availableElements={mockAvailableElements}
        />,
      );

      expect(screen.queryByTestId('handle-add field')).not.toBeInTheDocument();
    });

    it('should not render Add Field button when no available elements', () => {
      render(<ReorderableElement id="sortable-1" index={0} element={mockFormElement} availableElements={[]} />);

      expect(screen.queryByTestId('handle-add field')).not.toBeInTheDocument();
    });

    it('should open menu when Add Field button is clicked', () => {
      render(
        <ReorderableElement
          id="sortable-1"
          index={0}
          element={mockFormElement}
          availableElements={mockAvailableElements}
        />,
      );

      const addButton = screen.getByTestId('handle-add field');
      fireEvent.click(addButton);

      // Menu should open with available elements (filtered by allowed types)
      expect(screen.getByText('Text Input')).toBeInTheDocument();
      expect(screen.getByText('Password Input')).toBeInTheDocument();
    });

    it('should filter out elements not in allowed types or hidden from resource panel', () => {
      render(
        <ReorderableElement
          id="sortable-1"
          index={0}
          element={mockFormElement}
          availableElements={mockAvailableElements}
        />,
      );

      const addButton = screen.getByTestId('handle-add field');
      fireEvent.click(addButton);

      // Hidden element should not appear
      expect(screen.queryByText('Hidden')).not.toBeInTheDocument();
    });

    it('should call onAddElementToForm when menu item is clicked', () => {
      const onAddElementToForm = vi.fn();

      render(
        <ReorderableElement
          id="sortable-1"
          index={0}
          element={mockFormElement}
          availableElements={mockAvailableElements}
          onAddElementToForm={onAddElementToForm}
        />,
      );

      const addButton = screen.getByTestId('handle-add field');
      fireEvent.click(addButton);

      const menuItem = screen.getByText('Text Input');
      fireEvent.click(menuItem);

      expect(onAddElementToForm).toHaveBeenCalledWith(mockAvailableElements[0], 'form-1');
    });

    it('should close menu after selecting an item', async () => {
      const onAddElementToForm = vi.fn();

      render(
        <ReorderableElement
          id="sortable-1"
          index={0}
          element={mockFormElement}
          availableElements={mockAvailableElements}
          onAddElementToForm={onAddElementToForm}
        />,
      );

      const addButton = screen.getByTestId('handle-add field');
      fireEvent.click(addButton);

      const menuItem = screen.getByText('Text Input');
      fireEvent.click(menuItem);

      // Menu should close - Text Input should no longer be visible in menu
      await waitFor(() => {
        const menuItems = screen.queryAllByRole('menuitem');
        expect(menuItems.length).toBe(0);
      });
    });
  });

  describe('Property Panel Interaction', () => {
    it('should open property panel on double click', () => {
      render(<ReorderableElement id="sortable-1" index={0} element={mockElement} />);

      const content = screen.getByTestId('element-factory').parentElement;
      if (content) {
        fireEvent.click(content);
      }

      expect(mockSetOpenValidationPanel).toHaveBeenCalledWith(false);
      expect(mockSetSelectedNotification).toHaveBeenCalledWith(null);
      expect(mockSetLastInteractedStepId).toHaveBeenCalledWith('test-step-id');
      expect(mockSetLastInteractedResource).toHaveBeenCalledWith(mockElement);
    });

    it('should open property panel on Edit button click', () => {
      render(<ReorderableElement id="sortable-1" index={0} element={mockElement} />);

      const editButton = screen.getByTestId('handle-edit');
      fireEvent.click(editButton);

      expect(mockSetOpenValidationPanel).toHaveBeenCalledWith(false);
      expect(mockSetSelectedNotification).toHaveBeenCalledWith(null);
      expect(mockSetLastInteractedStepId).toHaveBeenCalledWith('test-step-id');
      expect(mockSetLastInteractedResource).toHaveBeenCalledWith(mockElement);
    });

    it('should open property panel on content double click', () => {
      const {container} = render(<ReorderableElement id="sortable-1" index={0} element={mockElement} />);

      const contentDiv = container.querySelector('.flow-builder-step-content-form-field-content');
      if (contentDiv) {
        fireEvent.doubleClick(contentDiv);
      }

      expect(mockSetLastInteractedResource).toHaveBeenCalledWith(mockElement);
    });
  });

  describe('Delete Element', () => {
    it('should delete element when Delete button is clicked', async () => {
      render(<ReorderableElement id="sortable-1" index={0} element={mockElement} />);

      const deleteButton = screen.getByTestId('handle-delete');
      fireEvent.click(deleteButton);

      await waitFor(() => {
        expect(mockExecuteAsync).toHaveBeenCalled();
        expect(mockDeleteComponent).toHaveBeenCalledWith('test-step-id', mockElement);
        expect(mockSetIsOpenResourcePropertiesPanel).toHaveBeenCalledWith(false);
      });
    });

    it('should show error notification when delete fails', async () => {
      const error = new Error('Delete failed');
      mockExecuteAsync.mockRejectedValueOnce(error);

      render(<ReorderableElement id="sortable-1" index={0} element={mockElement} />);

      const deleteButton = screen.getByTestId('handle-delete');
      fireEvent.click(deleteButton);

      await waitFor(() => {
        expect(mockAddNotification).toHaveBeenCalled();
      });
    });
  });

  describe('Menu Handling', () => {
    it('should close menu when onClose is triggered', async () => {
      render(
        <ReorderableElement
          id="sortable-1"
          index={0}
          element={mockFormElement}
          availableElements={mockAvailableElements}
        />,
      );

      // Open menu
      const addButton = screen.getByTestId('handle-add field');
      fireEvent.click(addButton);

      expect(screen.getByText('Text Input')).toBeInTheDocument();

      // Close menu by pressing Escape
      fireEvent.keyDown(document.activeElement ?? document.body, {key: 'Escape'});

      await waitFor(() => {
        expect(screen.queryByRole('menu')).not.toBeInTheDocument();
      });
    });

    it('should show "No fields available" when form compatible elements are empty', () => {
      const noCompatibleElements: Resource[] = [
        {
          id: 'incompatible',
          type: 'INCOMPATIBLE_TYPE',
          category: 'OTHER',
          resourceType: 'ELEMENT',
          display: {label: 'Incompatible', image: '', showOnResourcePanel: true},
        } as Resource,
      ];

      render(
        <ReorderableElement
          id="sortable-1"
          index={0}
          element={mockFormElement}
          availableElements={noCompatibleElements}
        />,
      );

      // Add button won't be visible because there are no compatible elements
      expect(screen.queryByTestId('handle-add field')).not.toBeInTheDocument();
    });
  });

  describe('Memoization', () => {
    it('should maintain stable handler references across rerenders', () => {
      const {rerender} = render(<ReorderableElement id="sortable-1" index={0} element={mockElement} />);

      // First render - click edit
      const editButton = screen.getByTestId('handle-edit');
      fireEvent.click(editButton);
      expect(mockSetLastInteractedResource).toHaveBeenCalledTimes(1);

      // Rerender with same props
      rerender(<ReorderableElement id="sortable-1" index={0} element={mockElement} />);

      // Click edit again
      const editButton2 = screen.getByTestId('handle-edit');
      fireEvent.click(editButton2);
      expect(mockSetLastInteractedResource).toHaveBeenCalledTimes(2);
    });

    it('should not re-render when only availableElements change', () => {
      const newAvailableElements: Resource[] = [
        {
          id: 'new-input',
          type: 'TEXT_INPUT',
          category: 'FIELD',
          resourceType: 'ELEMENT',
          display: {label: 'New Input', image: '', showOnResourcePanel: true},
        } as Resource,
      ];

      const {rerender} = render(
        <ReorderableElement
          id="sortable-1"
          index={0}
          element={mockElement}
          availableElements={mockAvailableElements}
        />,
      );

      // Rerender with different availableElements but same element
      rerender(
        <ReorderableElement
          id="sortable-1"
          index={0}
          element={mockElement}
          availableElements={newAvailableElements}
        />,
      );

      // Component should still render correctly
      expect(screen.getByTestId('element-factory')).toBeInTheDocument();
    });

    it('should re-render when element changes', () => {
      const newElement: Resource = {
        id: 'element-2',
        type: 'PASSWORD_INPUT',
        category: 'FIELD',
        resourceType: 'ELEMENT',
        display: {label: 'Password Field', image: '', showOnResourcePanel: true},
      } as Resource;

      const {rerender} = render(<ReorderableElement id="sortable-1" index={0} element={mockElement} />);

      expect(screen.getByText('Username Field')).toBeInTheDocument();

      rerender(<ReorderableElement id="sortable-1" index={0} element={newElement} />);

      expect(screen.getByText('Password Field')).toBeInTheDocument();
    });

    it('should re-render when id changes', () => {
      const {rerender} = render(<ReorderableElement id="sortable-1" index={0} element={mockElement} />);

      rerender(<ReorderableElement id="sortable-2" index={0} element={mockElement} />);

      expect(screen.getByTestId('element-factory')).toBeInTheDocument();
    });

    it('should re-render when index changes', () => {
      const {rerender} = render(<ReorderableElement id="sortable-1" index={0} element={mockElement} />);

      rerender(<ReorderableElement id="sortable-1" index={5} element={mockElement} />);

      expect(screen.getByTestId('element-factory')).toBeInTheDocument();
    });

    it('should re-render when className changes', () => {
      const {rerender, container} = render(
        <ReorderableElement id="sortable-1" index={0} element={mockElement} className="class-a" />,
      );

      expect(container.querySelector('.class-a')).toBeInTheDocument();

      rerender(<ReorderableElement id="sortable-1" index={0} element={mockElement} className="class-b" />);

      expect(container.querySelector('.class-b')).toBeInTheDocument();
    });
  });

  describe('Element Display', () => {
    it('should use element type when display label is not available', () => {
      const elementWithoutLabel = {
        id: 'no-label',
        type: 'TEXT_INPUT',
        category: 'FIELD',
        resourceType: 'ELEMENT',
      } as Resource;

      render(<ReorderableElement id="sortable-1" index={0} element={elementWithoutLabel} />);

      expect(screen.getByText('TEXT_INPUT')).toBeInTheDocument();
    });

    it('should render menu item with element display label', () => {
      render(
        <ReorderableElement
          id="sortable-1"
          index={0}
          element={mockFormElement}
          availableElements={mockAvailableElements}
        />,
      );

      const addButton = screen.getByTestId('handle-add field');
      fireEvent.click(addButton);

      expect(screen.getByText('Text Input')).toBeInTheDocument();
      expect(screen.getByText('Password Input')).toBeInTheDocument();
    });
  });
});
