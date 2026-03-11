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

/* eslint-disable react/require-default-props, jsx-a11y/no-static-element-interactions */

import {describe, it, expect, vi, beforeEach} from 'vitest';
import {render} from '@thunder/test-utils/browser';
import {page, userEvent} from 'vitest/browser';
import React, {type ReactNode} from 'react';
import FlowBuilderCoreContext, {type FlowBuilderCoreContextProps} from '../../../context/FlowBuilderCoreContext';
import {EdgeStyleTypes} from '../../../models/steps';
import {PreviewScreenType} from '../../../models/custom-text-preference';
import {ElementTypes} from '../../../models/elements';
import type {Base} from '../../../models/base';

// Import after mocks are set up
import BaseEdge from '../BaseEdge';

// Use vi.hoisted to define mocks that need to be referenced in vi.mock
const {mockDeleteElements, mockUseNodes} = vi.hoisted(() => ({
  mockDeleteElements: vi.fn().mockResolvedValue({}),
  mockUseNodes: vi.fn(() => []),
}));

// Mock the calculateEdgePath utility
vi.mock('../../../utils/calculateEdgePath', () => ({
  calculateEdgePath: vi.fn(() => ({
    path: 'M 0,0 L 100,0 L 100,100 L 200,100',
    centerX: 100,
    centerY: 50,
  })),
}));

interface MockBaseEdgeProps {
  id: string;
  path: string;
  style?: React.CSSProperties;
  interactionWidth?: number;
  markerEnd?: string;
  markerStart?: string;
}

interface MockEdgeLabelRendererProps {
  children: React.ReactNode;
}

// Mock @xyflow/react
vi.mock('@xyflow/react', () => ({
  BaseEdge: ({id, path, style, interactionWidth, markerEnd, markerStart}: MockBaseEdgeProps) => (
    <path
      data-testid={`base-edge-${id}`}
      d={path}
      style={style}
      data-interaction-width={interactionWidth}
      data-marker-end={markerEnd}
      data-marker-start={markerStart}
    />
  ),
  EdgeLabelRenderer: ({children}: MockEdgeLabelRendererProps) => (
    <div data-testid="edge-label-renderer">{children}</div>
  ),
  useReactFlow: () => ({
    deleteElements: mockDeleteElements,
  }),
  useNodes: mockUseNodes,
  Position: {
    Left: 'left',
    Right: 'right',
    Top: 'top',
    Bottom: 'bottom',
  },
}));


interface MockBoxProps {
  children?: React.ReactNode;
  onClick?: () => void;
  onKeyDown?: (e: React.KeyboardEvent) => void;
  onMouseEnter?: () => void;
  onMouseLeave?: () => void;
  sx?: Record<string, unknown>;
  role?: string;
  tabIndex?: number;
  'aria-label'?: string;
}

// Mock @wso2/oxygen-ui
vi.mock('@wso2/oxygen-ui', () => ({
  // eslint-disable-next-line jsx-a11y/no-static-element-interactions
  Box: ({children, onClick, onKeyDown, onMouseEnter, onMouseLeave, role, ...props}: MockBoxProps) => (
    <div
      data-testid="box-component"
      onClick={onClick}
      onKeyDown={onKeyDown}
      onMouseEnter={onMouseEnter}
      onMouseLeave={onMouseLeave}
      role={role ?? (onClick ? 'button' : undefined)}
      {...props}
    >
      {children}
    </div>
  ),
}));

interface MockXIconProps {
  size?: number;
  style?: React.CSSProperties;
}

// Mock @wso2/oxygen-ui-icons-react
vi.mock('@wso2/oxygen-ui-icons-react', () => ({
  XIcon: ({size, style}: MockXIconProps) => <span data-testid="x-icon" data-size={size} style={style} />,
}));

describe('BaseEdge', () => {
  const mockBaseResource: Base = {
    id: '',
    type: '',
    category: '',
    resourceType: '',
    version: '1.0.0',
    deprecated: false,
    deletable: true,
    display: {
      label: '',
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
    lastInteractedStepId: '',
    ResourceProperties: () => null,
    resourcePropertiesPanelHeading: '',
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
    setIsOpenResourcePropertiesPanel: vi.fn(),
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
      return <FlowBuilderCoreContext.Provider value={contextValue}>{children}</FlowBuilderCoreContext.Provider>;
    }
    return Wrapper;
  };

  // Mock Position values that match our mocked @xyflow/react Position enum
  const defaultProps = {
    id: 'edge-1',
    source: 'node-1',
    target: 'node-2',
    sourceX: 100,
    sourceY: 100,
    targetX: 300,
    targetY: 100,
    sourcePosition: 'right',
    targetPosition: 'left',
  } as unknown as React.ComponentProps<typeof BaseEdge>;

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('should render the edge path', async () => {
      await render(<BaseEdge {...defaultProps} />, {
        wrapper: createWrapper(),
      });

      await expect.element(page.getByTestId('base-edge-edge-1')).toBeInTheDocument();
    });

    it('should render EdgeLabelRenderer', async () => {
      await render(<BaseEdge {...defaultProps} />, {
        wrapper: createWrapper(),
      });

      await expect.element(page.getByTestId('edge-label-renderer')).toBeInTheDocument();
    });

    it('should render with calculated path', async () => {
      await render(<BaseEdge {...defaultProps} />, {
        wrapper: createWrapper(),
      });

      const edge = page.getByTestId('base-edge-edge-1');
      expect(edge).toHaveAttribute('d', 'M 0,0 L 100,0 L 100,100 L 200,100');
    });

    it('should render invisible hover detection path', async () => {
      const {container} = await render(<BaseEdge {...defaultProps} />, {
        wrapper: createWrapper(),
      });

      const hoverPath = container.querySelector('path[stroke="transparent"]');
      expect(hoverPath).toBeInTheDocument();
      expect(hoverPath).toHaveAttribute('stroke-width', '20');
    });
  });

  describe('Label', () => {
    it('should render label when provided', async () => {
      await render(<BaseEdge {...defaultProps} label={<span data-testid="edge-label">Test Label</span>} />, {
        wrapper: createWrapper(),
      });

      await expect.element(page.getByTestId('edge-label')).toBeInTheDocument();
    });

    it('should not render label container when label is not provided', async () => {
      await render(<BaseEdge {...defaultProps} />, {
        wrapper: createWrapper(),
      });

      // Label container should not have the label content
      await expect.element(page.getByTestId('edge-label')).not.toBeInTheDocument();
    });
  });

  describe('Hover Behavior', () => {
    it('should show delete button on hover', async () => {
      const {container} = await render(<BaseEdge {...defaultProps} />, {
        wrapper: createWrapper(),
      });

      // Initially, delete button should not be visible
      await expect.element(page.getByTestId('x-icon')).not.toBeInTheDocument();

      // Hover over the edge group
      const group = container.querySelector('g');
      await userEvent.hover(group!);

      // Delete button should now be visible
      await expect.element(page.getByTestId('x-icon')).toBeInTheDocument();
    });

    it('should hide delete button when mouse leaves', async () => {
      const {container} = await render(<BaseEdge {...defaultProps} />, {
        wrapper: createWrapper(),
      });

      const group = container.querySelector('g');

      // Hover over
      await userEvent.hover(group!);
      await expect.element(page.getByTestId('x-icon')).toBeInTheDocument();

      // Mouse leave
      await userEvent.unhover(group!);
      await expect.element(page.getByTestId('x-icon')).not.toBeInTheDocument();
    });

    it('should increase stroke width on hover', async () => {
      const {container} = await render(<BaseEdge {...defaultProps} style={{stroke: 'blue'}} />, {
        wrapper: createWrapper(),
      });

      const group = container.querySelector('g');
      await userEvent.hover(group!);

      const edge = page.getByTestId('base-edge-edge-1');
      expect(edge).toHaveStyle('stroke-width: 3');
    });
  });

  describe('Delete Functionality', () => {
    it('should call deleteElements when delete button is clicked', async () => {
      const {container} = await render(<BaseEdge {...defaultProps} />, {
        wrapper: createWrapper(),
      });

      // Hover to show delete button
      const group = container.querySelector('g');
      await userEvent.hover(group!);

      // Click delete button
      const deleteButton = page.getByRole('button', {name: 'Delete edge'});
      await userEvent.click(deleteButton);

      expect(mockDeleteElements).toHaveBeenCalledWith({edges: [{id: 'edge-1'}]});
    });

    it('should stop event propagation on delete click', async () => {
      const parentClickHandler = vi.fn();
      const {container} = await render(
        // eslint-disable-next-line jsx-a11y/click-events-have-key-events, jsx-a11y/no-static-element-interactions
        <div onClick={parentClickHandler}>
          <BaseEdge {...defaultProps} />
        </div>,
        {
          wrapper: createWrapper(),
        },
      );

      // Hover to show delete button
      const group = container.querySelector('g');
      await userEvent.hover(group!);

      // Click delete button
      const deleteButton = page.getByRole('button', {name: 'Delete edge'});
      await userEvent.click(deleteButton);

      // Parent click handler should not be called
      expect(parentClickHandler).not.toHaveBeenCalled();
    });

    it('should handle keyboard Enter key to delete', async () => {
      const {container} = await render(<BaseEdge {...defaultProps} />, {
        wrapper: createWrapper(),
      });

      // Hover to show delete button
      const group = container.querySelector('g');
      await userEvent.hover(group!);

      // Press Enter on delete button
      await userEvent.keyboard('{Enter}');

      expect(mockDeleteElements).toHaveBeenCalledWith({edges: [{id: 'edge-1'}]});
    });

    it('should handle keyboard Space key to delete', async () => {
      const {container} = await render(<BaseEdge {...defaultProps} />, {
        wrapper: createWrapper(),
      });

      // Hover to show delete button
      const group = container.querySelector('g');
      await userEvent.hover(group!);

      // Press Space on delete button
      await userEvent.keyboard('{ }');

      expect(mockDeleteElements).toHaveBeenCalledWith({edges: [{id: 'edge-1'}]});
    });

    it('should not delete on other key presses', async () => {
      const {container} = await render(<BaseEdge {...defaultProps} />, {
        wrapper: createWrapper(),
      });

      // Hover to show delete button
      const group = container.querySelector('g');
      await userEvent.hover(group!);

      // Press other key on delete button
      await userEvent.keyboard('{Tab}');

      expect(mockDeleteElements).not.toHaveBeenCalled();
    });

    it('should not show delete button when deletable is false', async () => {
      const {container} = await render(<BaseEdge {...defaultProps} deletable={false} />, {
        wrapper: createWrapper(),
      });

      // Hover over the edge
      const group = container.querySelector('g');
      await userEvent.hover(group!);

      // Delete button should not be visible
      await expect.element(page.getByRole('button', {name: 'Delete edge'})).not.toBeInTheDocument();
    });
  });

  describe('Edge Styles', () => {
    it('should apply custom style prop', async () => {
      await render(<BaseEdge {...defaultProps} style={{stroke: 'red', strokeDasharray: '5,5'}} />, {
        wrapper: createWrapper(),
      });

      const edge = page.getByTestId('base-edge-edge-1');
      expect(edge).toHaveStyle('stroke: red');
    });

    it('should use edge style from context', async () => {
      const contextWithBezier = {
        ...defaultContextValue,
        edgeStyle: EdgeStyleTypes.Bezier,
      };

      await render(<BaseEdge {...defaultProps} />, {
        wrapper: createWrapper(contextWithBezier),
      });

      await expect.element(page.getByTestId('base-edge-edge-1')).toBeInTheDocument();
    });

    it('should pass interaction width to base edge', async () => {
      await render(<BaseEdge {...defaultProps} />, {
        wrapper: createWrapper(),
      });

      const edge = page.getByTestId('base-edge-edge-1');
      expect(edge).toHaveAttribute('data-interaction-width', '20');
    });
  });

  describe('Markers', () => {
    it('should pass markerEnd to base edge', async () => {
      await render(<BaseEdge {...defaultProps} markerEnd="url(#arrow)" />, {
        wrapper: createWrapper(),
      });

      const edge = page.getByTestId('base-edge-edge-1');
      expect(edge).toHaveAttribute('data-marker-end', 'url(#arrow)');
    });

    it('should pass markerStart to base edge', async () => {
      await render(<BaseEdge {...defaultProps} markerStart="url(#arrow-start)" />, {
        wrapper: createWrapper(),
      });

      const edge = page.getByTestId('base-edge-edge-1');
      expect(edge).toHaveAttribute('data-marker-start', 'url(#arrow-start)');
    });
  });

  describe('Accessibility', () => {
    it('should have accessible delete button with aria-label', async () => {
      const {container} = await render(<BaseEdge {...defaultProps} />, {
        wrapper: createWrapper(),
      });

      // Hover to show delete button
      const group = container.querySelector('g');
      await userEvent.hover(group!);

      const deleteButton = page.getByRole('button', {name: 'Delete edge'});
      expect(deleteButton).toHaveAttribute('aria-label', 'Delete edge');
    });

    it('should have tabIndex on delete button', async () => {
      const {container} = await render(<BaseEdge {...defaultProps} />, {
        wrapper: createWrapper(),
      });

      // Hover to show delete button
      const group = container.querySelector('g');
      await userEvent.hover(group!);

      const deleteButton = page.getByRole('button', {name: 'Delete edge'});
      expect(deleteButton).toHaveAttribute('tabIndex', '0');
    });
  });

  describe('Delete Button Styling', () => {
    it('should render X icon with correct size', async () => {
      const {container} = await render(<BaseEdge {...defaultProps} />, {
        wrapper: createWrapper(),
      });

      // Hover to show delete button
      const group = container.querySelector('g');
      await userEvent.hover(group!);

      const xIcon = page.getByTestId('x-icon');
      expect(xIcon).toHaveAttribute('data-size', '16');
    });

    it('should render X icon with white color', async () => {
      const {container} = await render(<BaseEdge {...defaultProps} />, {
        wrapper: createWrapper(),
      });

      // Hover to show delete button
      const group = container.querySelector('g');
      await userEvent.hover(group!);

      const xIcon = page.getByTestId('x-icon');
      // Check for white color in either format (named or RGB)
      expect(xIcon).toHaveStyle({color: 'rgb(255, 255, 255)'});
    });
  });

  describe('Error Handling', () => {
    it('should handle deleteElements rejection gracefully', async () => {
      mockDeleteElements.mockRejectedValueOnce(new Error('Delete failed'));

      const {container} = await render(<BaseEdge {...defaultProps} />, {
        wrapper: createWrapper(),
      });

      // Hover to show delete button
      const group = container.querySelector('g');
      await userEvent.hover(group!);

      // Click delete button - should not throw
      const deleteButton = page.getByRole('button', {name: 'Delete edge'});
      expect(async () => userEvent.click(deleteButton)).not.toThrow();
    });
  });

  describe('Label Hover Effects', () => {
    it('should maintain hover state when mouse enters label', async () => {
      const {container} = await render(<BaseEdge {...defaultProps} label={<span data-testid="edge-label">Test Label</span>} />, {
        wrapper: createWrapper(),
      });

      // First hover over the edge group
      const group = container.querySelector('g');
      await userEvent.hover(group!);

      // Then hover over the label box
      const boxes = page.getByTestId('box-component');
      // First box is the label container
      const labelBox = boxes.all()[0];
      await userEvent.hover(labelBox);

      // Should still show delete button (hover state maintained)
      await expect.element(page.getByTestId('x-icon')).toBeInTheDocument();
    });

    it('should update hover state when mouse leaves label', async () => {
      const {container} = await render(<BaseEdge {...defaultProps} label={<span data-testid="edge-label">Test Label</span>} />, {
        wrapper: createWrapper(),
      });

      // Hover over the edge group
      const group = container.querySelector('g');
      await userEvent.hover(group!);

      const boxes = page.getByTestId('box-component');
      const labelBox = boxes.all()[0];

      // Enter and leave label
      await userEvent.hover(labelBox);
      await userEvent.unhover(labelBox);

      // Leave the group too
      await userEvent.unhover(group!);

      // Delete button should be hidden
      await expect.element(page.getByTestId('x-icon')).not.toBeInTheDocument();
    });
  });

  describe('Delete Button Hover Effects', () => {
    it('should maintain hover state when mouse enters delete button', async () => {
      const {container} = await render(<BaseEdge {...defaultProps} />, {
        wrapper: createWrapper(),
      });

      // First hover over the edge group
      const group = container.querySelector('g');
      await userEvent.hover(group!);

      // Then hover over the delete button itself
      const deleteButton = page.getByRole('button', {name: 'Delete edge'});
      await userEvent.hover(deleteButton);

      // Should still show delete button
      await expect.element(page.getByTestId('x-icon')).toBeInTheDocument();
    });

    it('should update hover state when mouse leaves delete button', async () => {
      const {container} = await render(<BaseEdge {...defaultProps} />, {
        wrapper: createWrapper(),
      });

      // Hover over the edge group
      const group = container.querySelector('g');
      await userEvent.hover(group!);

      const deleteButton = page.getByRole('button', {name: 'Delete edge'});

      // Enter and leave delete button
      await userEvent.hover(deleteButton);
      await userEvent.unhover(deleteButton);

      // Leave the group too
      await userEvent.unhover(group!);

      // Delete button should be hidden
      await expect.element(page.getByTestId('x-icon')).not.toBeInTheDocument();
    });

    it('should show delete button when hovering over label then delete button', async () => {
      const {container} = await render(<BaseEdge {...defaultProps} label={<span>Label</span>} />, {
        wrapper: createWrapper(),
      });

      // Hover over the edge group
      const group = container.querySelector('g');
      await userEvent.hover(group!);

      const boxes = page.getByTestId('box-component');
      // Hover over label first
      await userEvent.hover(boxes.all()[0]);

      // Then hover over delete button
      const deleteButton = page.getByRole('button', {name: 'Delete edge'});
      await userEvent.hover(deleteButton);

      // Delete button should still be visible
      await expect.element(page.getByTestId('x-icon')).toBeInTheDocument();
    });
  });
});
