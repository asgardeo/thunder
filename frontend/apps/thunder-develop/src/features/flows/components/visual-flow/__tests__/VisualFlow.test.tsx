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

/* eslint-disable @typescript-eslint/no-explicit-any, @typescript-eslint/no-unsafe-return, @typescript-eslint/no-unsafe-assignment, react/button-has-type, react/require-default-props, jsx-a11y/click-events-have-key-events, jsx-a11y/no-static-element-interactions */

import {describe, it, expect, vi, beforeEach} from 'vitest';
import {render, screen, fireEvent} from '@testing-library/react';
import type {ReactNode} from 'react';
import VisualFlow from '../VisualFlow';
import FlowBuilderCoreContext, {type FlowBuilderCoreContextProps} from '../../../context/FlowBuilderCoreContext';
import {EdgeStyleTypes} from '../../../models/steps';
import {PreviewScreenType} from '../../../models/custom-text-preference';
import {ElementTypes} from '../../../models/elements';
import type {Base} from '../../../models/base';

// Mock @xyflow/react
vi.mock('@xyflow/react', () => ({
  ReactFlow: ({children, nodes, edges, colorMode}: any) => (
    <div
      data-testid="react-flow"
      data-nodes={JSON.stringify(nodes)}
      data-edges={JSON.stringify(edges)}
      data-color-mode={colorMode}
    >
      {children}
    </div>
  ),
  Background: ({gap}: any) => <div data-testid="react-flow-background" data-gap={gap} />,
  Controls: ({children, position, orientation}: any) => (
    <div data-testid="react-flow-controls" data-position={position} data-orientation={orientation}>
      {children}
    </div>
  ),
  ControlButton: ({children, onClick, 'aria-label': ariaLabel}: any) => (
    <button data-testid="control-button" onClick={onClick} aria-label={ariaLabel}>
      {children}
    </button>
  ),
}));

// Mock react-i18next
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => {
      const translations: Record<string, string> = {
        'flows:core.headerPanel.autoLayout': 'Auto Layout',
        'flows:core.headerPanel.edgeStyleTooltip': 'Edge Style',
        'flows:core.headerPanel.save': 'Save',
        'flows:core.headerPanel.edgeStyles.bezier': 'Bezier',
        'flows:core.headerPanel.edgeStyles.smoothStep': 'Smooth Step',
        'flows:core.headerPanel.edgeStyles.step': 'Step',
      };
      return translations[key] || key;
    },
  }),
}));

// Mock color scheme - allow modification for tests
let mockColorSchemeMode = 'light';
let mockColorSchemeSystemMode = 'light';

// Mock @wso2/oxygen-ui
vi.mock('@wso2/oxygen-ui', () => ({
  Button: ({children, onClick, startIcon, variant}: any) => (
    <button data-testid="save-button" onClick={onClick} data-variant={variant}>
      {startIcon}
      {children}
    </button>
  ),
  Card: ({children}: any) => <div data-testid="save-card">{children}</div>,
  Tooltip: ({children, title}: any) => (
    <div data-testid="tooltip" title={title}>
      {children}
    </div>
  ),
  useColorScheme: () => ({
    mode: mockColorSchemeMode,
    systemMode: mockColorSchemeSystemMode,
  }),
  Menu: ({children, open}: any) => (open ? <div data-testid="menu">{children}</div> : null),
  MenuItem: ({children, onClick, selected}: any) => (
    <div data-testid="menu-item" onClick={onClick} data-selected={selected}>
      {children}
    </div>
  ),
  ListItemIcon: ({children}: any) => <span>{children}</span>,
  ListItemText: ({children}: any) => <span>{children}</span>,
}));

// Mock @wso2/oxygen-ui-icons-react
vi.mock('@wso2/oxygen-ui-icons-react', () => ({
  LayoutGrid: ({size}: any) => <span data-testid="layout-grid-icon" data-size={size} />,
  Save: ({size}: any) => <span data-testid="save-icon" data-size={size} />,
}));

// Mock validation indicator
vi.mock('../../validation-panel/CanvasValidationIndicator', () => ({
  default: () => <div data-testid="canvas-validation-indicator" />,
}));

// Mock EdgeStyleMenu
vi.mock('../EdgeStyleSelector', () => ({
  default: ({anchorEl}: any) => (
    <div data-testid="edge-style-menu" data-open={Boolean(anchorEl)}>
      Edge Style Menu
    </div>
  ),
}));

// Mock getEdgeStyleIcon
vi.mock('../../../utils/getEdgeStyleIcon', () => ({
  default: (style: string) => <span data-testid="edge-style-icon" data-style={style} />,
}));

describe('VisualFlow', () => {
  const mockHandleAutoLayout = vi.fn();
  const mockOnSave = vi.fn();
  const mockOnNodesChange = vi.fn();
  const mockOnEdgesChange = vi.fn();
  const mockOnConnect = vi.fn();
  const mockOnNodesDelete = vi.fn();
  const mockOnEdgesDelete = vi.fn();
  const mockOnNodeDragStop = vi.fn();

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

  const defaultProps = {
    nodes: [],
    edges: [],
    onNodesChange: mockOnNodesChange,
    onEdgesChange: mockOnEdgesChange,
    onConnect: mockOnConnect,
    onNodesDelete: mockOnNodesDelete,
    onEdgesDelete: mockOnEdgesDelete,
    onNodeDragStop: mockOnNodeDragStop,
  };

  beforeEach(() => {
    vi.clearAllMocks();
    mockColorSchemeMode = 'light';
    mockColorSchemeSystemMode = 'light';
  });

  describe('Rendering', () => {
    it('should render ReactFlow component', () => {
      render(<VisualFlow {...defaultProps} />, {
        wrapper: createWrapper(),
      });

      expect(screen.getByTestId('react-flow')).toBeInTheDocument();
    });

    it('should render Background component', () => {
      render(<VisualFlow {...defaultProps} />, {
        wrapper: createWrapper(),
      });

      const background = screen.getByTestId('react-flow-background');
      expect(background).toBeInTheDocument();
      expect(background).toHaveAttribute('data-gap', '20');
    });

    it('should render Controls component', () => {
      render(<VisualFlow {...defaultProps} handleAutoLayout={mockHandleAutoLayout} />, {
        wrapper: createWrapper(),
      });

      const controls = screen.getByTestId('react-flow-controls');
      expect(controls).toBeInTheDocument();
      expect(controls).toHaveAttribute('data-position', 'top-center');
      expect(controls).toHaveAttribute('data-orientation', 'horizontal');
    });

    it('should render CanvasValidationIndicator', () => {
      render(<VisualFlow {...defaultProps} />, {
        wrapper: createWrapper(),
      });

      expect(screen.getByTestId('canvas-validation-indicator')).toBeInTheDocument();
    });

    it('should render EdgeStyleMenu', () => {
      render(<VisualFlow {...defaultProps} />, {
        wrapper: createWrapper(),
      });

      expect(screen.getByTestId('edge-style-menu')).toBeInTheDocument();
    });
  });

  describe('Auto Layout Button', () => {
    it('should render auto-layout button when handleAutoLayout is provided', () => {
      render(<VisualFlow {...defaultProps} handleAutoLayout={mockHandleAutoLayout} />, {
        wrapper: createWrapper(),
      });

      expect(screen.getByLabelText('Auto Layout')).toBeInTheDocument();
    });

    it('should not render auto-layout button when handleAutoLayout is not provided', () => {
      render(<VisualFlow {...defaultProps} />, {
        wrapper: createWrapper(),
      });

      expect(screen.queryByLabelText('Auto Layout')).not.toBeInTheDocument();
    });

    it('should call handleAutoLayout when auto-layout button is clicked', () => {
      render(<VisualFlow {...defaultProps} handleAutoLayout={mockHandleAutoLayout} />, {
        wrapper: createWrapper(),
      });

      const autoLayoutButton = screen.getByLabelText('Auto Layout');
      fireEvent.click(autoLayoutButton);

      expect(mockHandleAutoLayout).toHaveBeenCalledTimes(1);
    });

    it('should render LayoutGrid icon in auto-layout button', () => {
      render(<VisualFlow {...defaultProps} handleAutoLayout={mockHandleAutoLayout} />, {
        wrapper: createWrapper(),
      });

      const icon = screen.getByTestId('layout-grid-icon');
      expect(icon).toBeInTheDocument();
      expect(icon).toHaveAttribute('data-size', '20');
    });
  });

  describe('Edge Style Button', () => {
    it('should render edge style button', () => {
      render(<VisualFlow {...defaultProps} />, {
        wrapper: createWrapper(),
      });

      expect(screen.getByLabelText('Edge Style')).toBeInTheDocument();
    });

    it('should display current edge style icon', () => {
      render(<VisualFlow {...defaultProps} />, {
        wrapper: createWrapper(),
      });

      const icon = screen.getByTestId('edge-style-icon');
      expect(icon).toBeInTheDocument();
    });
  });

  describe('Save Button', () => {
    it('should render save button', () => {
      render(<VisualFlow {...defaultProps} onSave={mockOnSave} />, {
        wrapper: createWrapper(),
      });

      expect(screen.getByTestId('save-button')).toBeInTheDocument();
    });

    it('should render save card container', () => {
      render(<VisualFlow {...defaultProps} onSave={mockOnSave} />, {
        wrapper: createWrapper(),
      });

      expect(screen.getByTestId('save-card')).toBeInTheDocument();
    });

    it('should call onSave when save button is clicked', () => {
      render(<VisualFlow {...defaultProps} onSave={mockOnSave} />, {
        wrapper: createWrapper(),
      });

      const saveButton = screen.getByTestId('save-button');
      fireEvent.click(saveButton);

      expect(mockOnSave).toHaveBeenCalledTimes(1);
    });

    it('should render Save icon in save button', () => {
      render(<VisualFlow {...defaultProps} onSave={mockOnSave} />, {
        wrapper: createWrapper(),
      });

      const icon = screen.getByTestId('save-icon');
      expect(icon).toBeInTheDocument();
      expect(icon).toHaveAttribute('data-size', '18');
    });

    it('should render save button with contained variant', () => {
      render(<VisualFlow {...defaultProps} onSave={mockOnSave} />, {
        wrapper: createWrapper(),
      });

      const saveButton = screen.getByTestId('save-button');
      expect(saveButton).toHaveAttribute('data-variant', 'contained');
    });
  });

  describe('Nodes and Edges', () => {
    it('should pass nodes to ReactFlow', () => {
      const nodes = [
        {id: 'node-1', position: {x: 0, y: 0}, data: {label: 'Node 1'}},
        {id: 'node-2', position: {x: 100, y: 100}, data: {label: 'Node 2'}},
      ];

      render(<VisualFlow {...defaultProps} nodes={nodes} />, {
        wrapper: createWrapper(),
      });

      const reactFlow = screen.getByTestId('react-flow');
      expect(reactFlow).toHaveAttribute('data-nodes', JSON.stringify(nodes));
    });

    it('should pass edges to ReactFlow', () => {
      const edges = [{id: 'edge-1', source: 'node-1', target: 'node-2'}];

      render(<VisualFlow {...defaultProps} edges={edges} />, {
        wrapper: createWrapper(),
      });

      const reactFlow = screen.getByTestId('react-flow');
      expect(reactFlow).toHaveAttribute('data-edges', JSON.stringify(edges));
    });

    it('should handle empty nodes and edges', () => {
      render(<VisualFlow {...defaultProps} nodes={[]} edges={[]} />, {
        wrapper: createWrapper(),
      });

      const reactFlow = screen.getByTestId('react-flow');
      expect(reactFlow).toHaveAttribute('data-nodes', '[]');
      expect(reactFlow).toHaveAttribute('data-edges', '[]');
    });
  });

  describe('Color Mode', () => {
    it('should pass color mode to ReactFlow', () => {
      render(<VisualFlow {...defaultProps} />, {
        wrapper: createWrapper(),
      });

      const reactFlow = screen.getByTestId('react-flow');
      expect(reactFlow).toHaveAttribute('data-color-mode', 'light');
    });

    it('should use systemMode when mode is system', () => {
      mockColorSchemeMode = 'system';
      mockColorSchemeSystemMode = 'dark';

      render(<VisualFlow {...defaultProps} />, {
        wrapper: createWrapper(),
      });

      const reactFlow = screen.getByTestId('react-flow');
      expect(reactFlow).toHaveAttribute('data-color-mode', 'dark');
    });

    it('should use mode directly when mode is dark', () => {
      mockColorSchemeMode = 'dark';
      mockColorSchemeSystemMode = 'light';

      render(<VisualFlow {...defaultProps} />, {
        wrapper: createWrapper(),
      });

      const reactFlow = screen.getByTestId('react-flow');
      expect(reactFlow).toHaveAttribute('data-color-mode', 'dark');
    });
  });

  describe('Custom Node and Edge Types', () => {
    it('should accept custom nodeTypes', () => {
      const customNodeTypes = {
        customNode: () => <div>Custom Node</div>,
      };

      render(<VisualFlow {...defaultProps} nodeTypes={customNodeTypes} />, {
        wrapper: createWrapper(),
      });

      expect(screen.getByTestId('react-flow')).toBeInTheDocument();
    });

    it('should accept custom edgeTypes', () => {
      const customEdgeTypes = {
        customEdge: () => <div>Custom Edge</div>,
      };

      render(<VisualFlow {...defaultProps} edgeTypes={customEdgeTypes} />, {
        wrapper: createWrapper(),
      });

      expect(screen.getByTestId('react-flow')).toBeInTheDocument();
    });

    it('should default to empty objects for node and edge types', () => {
      render(<VisualFlow {...defaultProps} />, {
        wrapper: createWrapper(),
      });

      expect(screen.getByTestId('react-flow')).toBeInTheDocument();
    });
  });

  describe('Callback Stability', () => {
    it('should handle onSave being undefined', () => {
      render(<VisualFlow {...defaultProps} onSave={undefined} />, {
        wrapper: createWrapper(),
      });

      // Should not throw even if save button exists
      expect(screen.getByTestId('react-flow')).toBeInTheDocument();
    });

    it('should handle handleAutoLayout being undefined', () => {
      render(<VisualFlow {...defaultProps} handleAutoLayout={undefined} />, {
        wrapper: createWrapper(),
      });

      expect(screen.queryByLabelText('Auto Layout')).not.toBeInTheDocument();
    });
  });
});
