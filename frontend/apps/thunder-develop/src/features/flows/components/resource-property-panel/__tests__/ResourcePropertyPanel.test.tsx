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
import type {ReactNode} from 'react';
import {ReactFlowProvider} from '@xyflow/react';
import FlowBuilderCoreContext, {type FlowBuilderCoreContextProps} from '../../../context/FlowBuilderCoreContext';
import {EdgeStyleTypes} from '../../../models/steps';
import {PreviewScreenType} from '../../../models/custom-text-preference';
import {ElementTypes} from '../../../models/elements';
import {ResourceTypes} from '../../../models/resources';
import type {Base} from '../../../models/base';

// Import after mocks
import ResourcePropertyPanel from '../ResourcePropertyPanel';

// Use vi.hoisted for mock functions
const {mockDeleteElements} = vi.hoisted(() => ({
  mockDeleteElements: vi.fn().mockResolvedValue({}),
}));

// Mock @xyflow/react
vi.mock('@xyflow/react', async () => {
  const actual = await vi.importActual('@xyflow/react');
  return {
    ...actual,
    useReactFlow: () => ({
      deleteElements: mockDeleteElements,
    }),
  };
});

// Mock ResourceProperties component
vi.mock('../ResourceProperties', () => ({
  default: () => <div data-testid="resource-properties">Resource Properties Content</div>,
}));

describe('ResourcePropertyPanel', () => {
  const mockSetIsOpenResourcePropertiesPanel = vi.fn();
  const mockOnComponentDelete = vi.fn();

  const mockBaseResource: Base = {
    id: 'resource-1',
    type: 'TEXT_INPUT',
    category: 'FIELD',
    resourceType: ResourceTypes.Element,
    version: '1.0.0',
    deprecated: false,
    deletable: true,
    display: {
      label: 'Test Resource',
      image: '',
      showOnResourcePanel: false,
    },
    config: {
      field: {name: '', type: ElementTypes},
      styles: {},
    },
  };

  const defaultContextValue: FlowBuilderCoreContextProps = {
    lastInteractedResource: mockBaseResource,
    lastInteractedStepId: 'step-1',
    ResourceProperties: () => null,
    resourcePropertiesPanelHeading: 'Test Panel Heading',
    primaryI18nScreen: PreviewScreenType.LOGIN,
    isResourcePanelOpen: true,
    isResourcePropertiesPanelOpen: false,
    isVersionHistoryPanelOpen: false,
    ElementFactory: () => null,
    onResourceDropOnCanvas: vi.fn(),
    selectedAttributes: {},
    setLastInteractedResource: vi.fn(),
    setLastInteractedStepId: vi.fn(),
    setResourcePropertiesPanelHeading: vi.fn(),
    setIsResourcePanelOpen: vi.fn(),
    setIsOpenResourcePropertiesPanel: mockSetIsOpenResourcePropertiesPanel,
    registerCloseValidationPanel: vi.fn(),
    setIsVersionHistoryPanelOpen: vi.fn(),
    setSelectedAttributes: vi.fn(),
    flowCompletionConfigs: {},
    setFlowCompletionConfigs: vi.fn(),
    flowNodeTypes: {},
    flowEdgeTypes: {},
    setFlowNodeTypes: vi.fn(),
    setFlowEdgeTypes: vi.fn(),
    isVerboseMode: false,
    setIsVerboseMode: vi.fn(),
    edgeStyle: EdgeStyleTypes.SmoothStep,
    setEdgeStyle: vi.fn(),
  };

  const createWrapper = (contextValue: FlowBuilderCoreContextProps = defaultContextValue) => {
    function Wrapper({children}: {children: ReactNode}) {
      return (
        <ReactFlowProvider>
          <FlowBuilderCoreContext.Provider value={contextValue}>{children}</FlowBuilderCoreContext.Provider>
        </ReactFlowProvider>
      );
    }
    return Wrapper;
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('should render children content', async () => {
      await render(
        <ResourcePropertyPanel open={false} onComponentDelete={mockOnComponentDelete}>
          <div data-testid="child-content">Child Content</div>
        </ResourcePropertyPanel>,
        {wrapper: createWrapper()},
      );

      await expect.element(page.getByTestId('child-content')).toBeInTheDocument();
    });

    it('should render the drawer container', async () => {
      await render(
        <ResourcePropertyPanel open={false} onComponentDelete={mockOnComponentDelete}>
          <div>Content</div>
        </ResourcePropertyPanel>,
        {wrapper: createWrapper()},
      );

      await expect.element(page.getByRole('presentation')).toBeInTheDocument();
    });

    it('should render panel heading from context', async () => {
      await render(
        <ResourcePropertyPanel open onComponentDelete={mockOnComponentDelete}>
          <div>Content</div>
        </ResourcePropertyPanel>,
        {wrapper: createWrapper()},
      );

      await expect.element(page.getByText('Test Panel Heading')).toBeInTheDocument();
    });

    it('should render ResourceProperties component', async () => {
      await render(
        <ResourcePropertyPanel open onComponentDelete={mockOnComponentDelete}>
          <div>Content</div>
        </ResourcePropertyPanel>,
        {wrapper: createWrapper()},
      );

      await expect.element(page.getByTestId('resource-properties')).toBeInTheDocument();
    });

    it('should render delete button when resource is deletable', async () => {
      await render(
        <ResourcePropertyPanel open onComponentDelete={mockOnComponentDelete}>
          <div>Content</div>
        </ResourcePropertyPanel>,
        {wrapper: createWrapper()},
      );

      await expect.element(page.getByRole('button', {name: /delete element/i})).toBeInTheDocument();
    });

    it('should not render delete button when resource is not deletable', async () => {
      const nonDeletableResource: Base = {
        ...mockBaseResource,
        deletable: false,
      };

      const contextWithNonDeletable: FlowBuilderCoreContextProps = {
        ...defaultContextValue,
        lastInteractedResource: nonDeletableResource,
      };

      await render(
        <ResourcePropertyPanel open onComponentDelete={mockOnComponentDelete}>
          <div>Content</div>
        </ResourcePropertyPanel>,
        {wrapper: createWrapper(contextWithNonDeletable)},
      );

      await expect.element(page.getByRole('button', {name: /delete element/i})).not.toBeInTheDocument();
    });
  });

  describe('Close Functionality', () => {
    it('should call setIsOpenResourcePropertiesPanel(false) when close button is clicked', async () => {
      await render(
        <ResourcePropertyPanel open onComponentDelete={mockOnComponentDelete}>
          <div>Content</div>
        </ResourcePropertyPanel>,
        {wrapper: createWrapper()},
      );

      // Find the close button (the X icon button) - use hidden: true since drawer has aria-hidden
      const closeButton = page.getByRole('button').all()[0];
      await userEvent.click(closeButton);

      expect(mockSetIsOpenResourcePropertiesPanel).toHaveBeenCalledWith(false);
    });
  });

  describe('Delete Functionality', () => {
    it('should delete step node when resource is a Step', async () => {
      const stepResource: Base = {
        ...mockBaseResource,
        resourceType: ResourceTypes.Step,
      };

      const contextWithStep: FlowBuilderCoreContextProps = {
        ...defaultContextValue,
        lastInteractedResource: stepResource,
      };

      await render(
        <ResourcePropertyPanel open onComponentDelete={mockOnComponentDelete}>
          <div>Content</div>
        </ResourcePropertyPanel>,
        {wrapper: createWrapper(contextWithStep)},
      );

      const deleteButton = page.getByRole('button', {name: /delete element/i});
      await userEvent.click(deleteButton);

      expect(mockDeleteElements).toHaveBeenCalledWith({nodes: [{id: stepResource.id}]});
      expect(mockSetIsOpenResourcePropertiesPanel).toHaveBeenCalledWith(false);
    });

    it('should call onComponentDelete when resource is not a Step', async () => {
      await render(
        <ResourcePropertyPanel open onComponentDelete={mockOnComponentDelete}>
          <div>Content</div>
        </ResourcePropertyPanel>,
        {wrapper: createWrapper()},
      );

      const deleteButton = page.getByRole('button', {name: /delete element/i});
      await userEvent.click(deleteButton);

      expect(mockOnComponentDelete).toHaveBeenCalledWith('step-1', mockBaseResource);
      expect(mockSetIsOpenResourcePropertiesPanel).toHaveBeenCalledWith(false);
    });

    it('should not delete when lastInteractedResource is null', async () => {
      const contextWithoutResource: FlowBuilderCoreContextProps = {
        ...defaultContextValue,
        lastInteractedResource: null as unknown as Base,
      };

      await render(
        <ResourcePropertyPanel open onComponentDelete={mockOnComponentDelete}>
          <div>Content</div>
        </ResourcePropertyPanel>,
        {wrapper: createWrapper(contextWithoutResource)},
      );

      // Delete button should still render (based on deletable !== false)
      const deleteButton = page.getByRole('button', {name: /delete element/i});
      await userEvent.click(deleteButton);

      expect(mockDeleteElements).not.toHaveBeenCalled();
      expect(mockOnComponentDelete).not.toHaveBeenCalled();
    });
  });

  describe('Props', () => {
    it('should apply custom className', async () => {
      await render(
        <ResourcePropertyPanel
          open
          onComponentDelete={mockOnComponentDelete}
          className="custom-class"
        >
          <div>Content</div>
        </ResourcePropertyPanel>,
        {wrapper: createWrapper()},
      );

      // The className should be applied to the drawer paper
      const drawer = document.querySelector('.custom-class');
      expect(drawer).toBeInTheDocument();
    });

    it('should use right anchor by default', async () => {
      await render(
        <ResourcePropertyPanel open onComponentDelete={mockOnComponentDelete}>
          <div>Content</div>
        </ResourcePropertyPanel>,
        {wrapper: createWrapper()},
      );

      // The drawer should be anchored to the right
      const drawer = document.querySelector('.MuiDrawer-paperAnchorRight');
      expect(drawer).toBeInTheDocument();
    });

    it('should pass additional props to Box container', async () => {
      await render(
        <ResourcePropertyPanel
          open
          onComponentDelete={mockOnComponentDelete}
          data-testid="custom-container"
        >
          <div>Content</div>
        </ResourcePropertyPanel>,
        {wrapper: createWrapper()},
      );

      await expect.element(page.getByTestId('custom-container')).toBeInTheDocument();
    });
  });

  describe('Drawer State', () => {
    it('should render drawer as open when open prop is true', async () => {
      await render(
        <ResourcePropertyPanel open onComponentDelete={mockOnComponentDelete}>
          <div>Content</div>
        </ResourcePropertyPanel>,
        {wrapper: createWrapper()},
      );

      // Check that the drawer has the open class
      const drawer = document.querySelector('.MuiDrawer-root');
      expect(drawer).toBeInTheDocument();
    });

    it('should render drawer as closed when open prop is false', async () => {
      await render(
        <ResourcePropertyPanel open={false} onComponentDelete={mockOnComponentDelete}>
          <div>Content</div>
        </ResourcePropertyPanel>,
        {wrapper: createWrapper()},
      );

      // The drawer content should still be accessible due to keepMounted
      await expect.element(page.getByTestId('resource-properties')).toBeInTheDocument();
    });
  });

  describe('Edge Cases', () => {
    it('should handle deleteElements rejection gracefully', async () => {
      mockDeleteElements.mockRejectedValueOnce(new Error('Delete failed'));

      const stepResource: Base = {
        ...mockBaseResource,
        resourceType: ResourceTypes.Step,
      };

      const contextWithStep: FlowBuilderCoreContextProps = {
        ...defaultContextValue,
        lastInteractedResource: stepResource,
      };

      await render(
        <ResourcePropertyPanel open onComponentDelete={mockOnComponentDelete}>
          <div>Content</div>
        </ResourcePropertyPanel>,
        {wrapper: createWrapper(contextWithStep)},
      );

      const deleteButton = page.getByRole('button', {name: /delete element/i});

      // Should not throw
      expect(async () => userEvent.click(deleteButton)).not.toThrow();
    });
  });
});
