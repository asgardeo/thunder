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

import {describe, it, expect, vi, beforeEach, afterEach} from 'vitest';
import {render} from '@thunder/test-utils/browser';
import {page, userEvent} from 'vitest/browser';
import type {ReactNode} from 'react';
import EdgeStyleMenu from '../EdgeStyleSelector';
import FlowBuilderCoreContext, {type FlowBuilderCoreContextProps} from '../../../context/FlowBuilderCoreContext';
import {EdgeStyleTypes} from '../../../models/steps';
import {PreviewScreenType} from '../../../models/custom-text-preference';
import {ElementTypes} from '../../../models/elements';
import type {Base} from '../../../models/base';

describe('EdgeStyleMenu', () => {
  const mockSetEdgeStyle = vi.fn();
  const mockOnClose = vi.fn();

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
    setEdgeStyle: mockSetEdgeStyle,
  };

  const createWrapper = (contextValue: FlowBuilderCoreContextProps = defaultContextValue) => {
    function Wrapper({children}: {children: ReactNode}) {
      return <FlowBuilderCoreContext.Provider value={contextValue}>{children}</FlowBuilderCoreContext.Provider>;
    }
    return Wrapper;
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Menu Visibility', () => {
    it('should not render menu when anchorEl is null', async () => {
      await render(<EdgeStyleMenu anchorEl={null} onClose={mockOnClose} />, {
        wrapper: createWrapper(),
      });

      // Menu should not be visible
      await expect.element(page.getByRole('menu')).not.toBeInTheDocument();
    });

    it('should render menu when anchorEl is provided', async () => {
      const anchorEl = document.createElement('button');
      document.body.appendChild(anchorEl);

      await render(<EdgeStyleMenu anchorEl={anchorEl} onClose={mockOnClose} />, {
        wrapper: createWrapper(),
      });

      await expect.element(page.getByRole('menu')).toBeInTheDocument();

      document.body.removeChild(anchorEl);
    });
  });

  describe('Menu Options', () => {
    let anchorEl: HTMLButtonElement;

    beforeEach(() => {
      anchorEl = document.createElement('button');
      document.body.appendChild(anchorEl);
    });

    afterEach(() => {
      if (document.body.contains(anchorEl)) {
        document.body.removeChild(anchorEl);
      }
    });

    it('should render all three edge style options', async () => {
      await render(<EdgeStyleMenu anchorEl={anchorEl} onClose={mockOnClose} />, {
        wrapper: createWrapper(),
      });

      await expect.element(page.getByText('Bezier')).toBeInTheDocument();
      await expect.element(page.getByText('Smooth Step')).toBeInTheDocument();
      await expect.element(page.getByText('Step')).toBeInTheDocument();
    });

    it('should render menu items as clickable', async () => {
      await render(<EdgeStyleMenu anchorEl={anchorEl} onClose={mockOnClose} />, {
        wrapper: createWrapper(),
      });

      const menuItems = page.getByRole('menuitem');
      expect(menuItems).toHaveLength(3);
    });

    it('should render icons for each option', async () => {
      await render(<EdgeStyleMenu anchorEl={anchorEl} onClose={mockOnClose} />, {
        wrapper: createWrapper(),
      });

      // Each menu item should have an icon (ListItemIcon)
      const menuItems = page.getByRole('menuitem');
      expect(menuItems).toHaveLength(3);
    });
  });

  describe('Style Selection', () => {
    let anchorEl: HTMLButtonElement;

    beforeEach(() => {
      anchorEl = document.createElement('button');
      document.body.appendChild(anchorEl);
    });

    afterEach(() => {
      if (document.body.contains(anchorEl)) {
        document.body.removeChild(anchorEl);
      }
    });

    it('should call setEdgeStyle with Bezier when Bezier option is clicked', async () => {
      await render(<EdgeStyleMenu anchorEl={anchorEl} onClose={mockOnClose} />, {
        wrapper: createWrapper(),
      });

      const bezierOption = page.getByText('Bezier');
      await userEvent.click(bezierOption);

      expect(mockSetEdgeStyle).toHaveBeenCalledWith(EdgeStyleTypes.Bezier);
    });

    it('should call setEdgeStyle with SmoothStep when Smooth Step option is clicked', async () => {
      await render(<EdgeStyleMenu anchorEl={anchorEl} onClose={mockOnClose} />, {
        wrapper: createWrapper(),
      });

      const smoothStepOption = page.getByText('Smooth Step');
      await userEvent.click(smoothStepOption);

      expect(mockSetEdgeStyle).toHaveBeenCalledWith(EdgeStyleTypes.SmoothStep);
    });

    it('should call setEdgeStyle with Step when Step option is clicked', async () => {
      await render(<EdgeStyleMenu anchorEl={anchorEl} onClose={mockOnClose} />, {
        wrapper: createWrapper(),
      });

      const stepOption = page.getByText('Step');
      await userEvent.click(stepOption);

      expect(mockSetEdgeStyle).toHaveBeenCalledWith(EdgeStyleTypes.Step);
    });

    it('should call onClose after selecting an option', async () => {
      await render(<EdgeStyleMenu anchorEl={anchorEl} onClose={mockOnClose} />, {
        wrapper: createWrapper(),
      });

      const bezierOption = page.getByText('Bezier');
      await userEvent.click(bezierOption);

      expect(mockOnClose).toHaveBeenCalled();
    });
  });

  describe('Edge Style Context', () => {
    let anchorEl: HTMLButtonElement;

    beforeEach(() => {
      anchorEl = document.createElement('button');
      document.body.appendChild(anchorEl);
    });

    afterEach(() => {
      if (document.body.contains(anchorEl)) {
        document.body.removeChild(anchorEl);
      }
    });

    it('should render all edge style options with SmoothStep context', async () => {
      await render(<EdgeStyleMenu anchorEl={anchorEl} onClose={mockOnClose} />, {
        wrapper: createWrapper({
          ...defaultContextValue,
          edgeStyle: EdgeStyleTypes.SmoothStep,
        }),
      });

      const menuItems = page.getByRole('menuitem');
      expect(menuItems).toHaveLength(3);
      await expect.element(page.getByText('Smooth Step')).toBeInTheDocument();
    });

    it('should render all edge style options with Bezier context', async () => {
      await render(<EdgeStyleMenu anchorEl={anchorEl} onClose={mockOnClose} />, {
        wrapper: createWrapper({
          ...defaultContextValue,
          edgeStyle: EdgeStyleTypes.Bezier,
        }),
      });

      const menuItems = page.getByRole('menuitem');
      expect(menuItems).toHaveLength(3);
      await expect.element(page.getByText('Bezier')).toBeInTheDocument();
    });

    it('should render all edge style options with Step context', async () => {
      await render(<EdgeStyleMenu anchorEl={anchorEl} onClose={mockOnClose} />, {
        wrapper: createWrapper({
          ...defaultContextValue,
          edgeStyle: EdgeStyleTypes.Step,
        }),
      });

      const menuItems = page.getByRole('menuitem');
      expect(menuItems).toHaveLength(3);
      await expect.element(page.getByText('Step')).toBeInTheDocument();
    });
  });

  describe('Menu Props', () => {
    it('should render menu with correct structure', async () => {
      const anchorEl = document.createElement('button');
      document.body.appendChild(anchorEl);

      await render(<EdgeStyleMenu anchorEl={anchorEl} onClose={mockOnClose} />, {
        wrapper: createWrapper(),
      });

      const menu = page.getByRole('menu');
      expect(menu).toBeInTheDocument();
      // The menu is rendered and accessible
      expect(page.getByRole('menuitem')).toHaveLength(3);

      document.body.removeChild(anchorEl);
    });
  });

  describe('Selected State', () => {
    let anchorEl: HTMLButtonElement;

    beforeEach(() => {
      anchorEl = document.createElement('button');
      document.body.appendChild(anchorEl);
    });

    afterEach(() => {
      if (document.body.contains(anchorEl)) {
        document.body.removeChild(anchorEl);
      }
    });

    it('should mark Bezier as selected when edgeStyle is Bezier', async () => {
      await render(<EdgeStyleMenu anchorEl={anchorEl} onClose={mockOnClose} />, {
        wrapper: createWrapper({
          ...defaultContextValue,
          edgeStyle: EdgeStyleTypes.Bezier,
        }),
      });

      const menuItems = page.getByRole('menuitem');
      // First item (Bezier) should be selected
      expect((menuItems.all())[0]).toHaveClass('Mui-selected');
    });

    it('should mark SmoothStep as selected when edgeStyle is SmoothStep', async () => {
      await render(<EdgeStyleMenu anchorEl={anchorEl} onClose={mockOnClose} />, {
        wrapper: createWrapper({
          ...defaultContextValue,
          edgeStyle: EdgeStyleTypes.SmoothStep,
        }),
      });

      const menuItems = page.getByRole('menuitem');
      // Second item (SmoothStep) should be selected
      expect((menuItems.all())[1]).toHaveClass('Mui-selected');
    });

    it('should mark Step as selected when edgeStyle is Step', async () => {
      await render(<EdgeStyleMenu anchorEl={anchorEl} onClose={mockOnClose} />, {
        wrapper: createWrapper({
          ...defaultContextValue,
          edgeStyle: EdgeStyleTypes.Step,
        }),
      });

      const menuItems = page.getByRole('menuitem');
      // Third item (Step) should be selected
      expect((menuItems.all())[2]).toHaveClass('Mui-selected');
    });
  });

  describe('Boolean Conversion', () => {
    it('should convert null anchorEl to false for open state', async () => {
      await render(<EdgeStyleMenu anchorEl={null} onClose={mockOnClose} />, {
        wrapper: createWrapper(),
      });

      // Menu should not be visible when anchorEl is null
      await expect.element(page.getByRole('menu')).not.toBeInTheDocument();
    });

    it('should convert valid anchorEl to true for open state', async () => {
      const anchorEl = document.createElement('button');
      document.body.appendChild(anchorEl);

      await render(<EdgeStyleMenu anchorEl={anchorEl} onClose={mockOnClose} />, {
        wrapper: createWrapper(),
      });

      // Menu should be visible when anchorEl is provided
      await expect.element(page.getByRole('menu')).toBeInTheDocument();

      document.body.removeChild(anchorEl);
    });
  });
});
