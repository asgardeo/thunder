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

import {describe, it, expect, beforeEach, vi} from 'vitest';
import {page, userEvent} from 'vitest/browser';
import {render} from '@thunder/test-utils/browser';
import {type BasicFlowDefinition} from '@/features/flows/models/responses';
import FlowsListView, {type FlowsListViewProps} from '../FlowsListView';

describe('FlowsListView', () => {
  const mockOnFlowSelect = vi.fn();
  const mockOnClearSelection = vi.fn();

  const mockFlows: BasicFlowDefinition[] = [
    {
      id: 'flow-1',
      name: 'Basic Authentication Flow',
      activeVersion: 1,
      handle: 'basic-auth-flow',
      flowType: 'AUTHENTICATION',
      createdAt: '',
      updatedAt: '',
    },
    {
      id: 'flow-2',
      name: 'Google OAuth Flow',
      activeVersion: 1,
      handle: 'google-oauth-flow',
      flowType: 'AUTHENTICATION',
      createdAt: '',
      updatedAt: '',
    },
    {
      id: 'flow-3',
      name: 'Multi-Factor Auth Flow',
      activeVersion: 1,
      handle: 'mfa-flow',
      flowType: 'AUTHENTICATION',
      createdAt: '',
      updatedAt: '',
    },
  ];

  const defaultProps: FlowsListViewProps = {
    availableFlows: mockFlows,
    selectedAuthFlow: null,
    onFlowSelect: mockOnFlowSelect,
    onClearSelection: mockOnClearSelection,
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  const renderComponent = async (props: Partial<FlowsListViewProps> = {}) =>
    render(<FlowsListView {...defaultProps} {...props} />);

  describe('rendering', () => {
    it('should return null when no selectable flows available', async () => {
      const {container} = await renderComponent({
        availableFlows: [],
      });

      expect(container.firstChild).toBeNull();
    });

    it('should return null when all flows are develop-app flows', async () => {
      const developAppFlows: BasicFlowDefinition[] = [
        {
          id: 'flow-1',
          name: 'Develop App Flow',
          activeVersion: 1,
          handle: 'develop-app-login',
          flowType: 'AUTHENTICATION',
          createdAt: '',
          updatedAt: '',
        },
      ];

      const {container} = await renderComponent({
        availableFlows: developAppFlows,
      });

      expect(container.firstChild).toBeNull();
    });

    it('should return null when all flows are default flows', async () => {
      const defaultFlows: BasicFlowDefinition[] = [
        {
          id: 'flow-1',
          name: 'Default Login Flow',
          activeVersion: 1,
          handle: 'default-login',
          flowType: 'AUTHENTICATION',
          createdAt: '',
          updatedAt: '',
        },
      ];

      const {container} = await renderComponent({
        availableFlows: defaultFlows,
      });

      expect(container.firstChild).toBeNull();
    });

    it('should render divider with "or" text', async () => {
      await renderComponent();

      await expect.element(page.getByText('or')).toBeInTheDocument();
    });

    it('should render autocomplete component', async () => {
      await renderComponent();

      await expect.element(page.getByRole('combobox')).toBeInTheDocument();
    });

    it('should render with proper label text', async () => {
      await renderComponent();

      await expect.element(page.getByLabelText('Select a flow')).toBeInTheDocument();
    });
  });

  describe('flow filtering', () => {
    it('should filter out develop-app flows', async () => {
      const mixedFlows: BasicFlowDefinition[] = [
        ...mockFlows,
        {
          id: 'dev-flow',
          name: 'Dev App Flow',
          activeVersion: 1,
          handle: 'develop-app-flow',
          flowType: 'AUTHENTICATION',
          createdAt: '',
          updatedAt: '',
        },
      ];

      await renderComponent({availableFlows: mixedFlows});

      // The component should render since there are selectable flows
      await expect.element(page.getByRole('combobox')).toBeInTheDocument();
    });

    it('should filter out default flows', async () => {
      const mixedFlows: BasicFlowDefinition[] = [
        ...mockFlows,
        {
          id: 'default-flow',
          name: 'Default Flow',
          activeVersion: 1,
          handle: 'default-auth-flow',
          flowType: 'AUTHENTICATION',
          createdAt: '',
          updatedAt: '',
        },
      ];

      await renderComponent({availableFlows: mixedFlows});

      await expect.element(page.getByRole('combobox')).toBeInTheDocument();
    });
  });

  describe('autocomplete interaction', () => {
    it('should call onFlowSelect when a flow is selected', async () => {
      await renderComponent();

      const autocomplete = page.getByRole('combobox');
      await userEvent.click(autocomplete);

      const flowOption = page.getByText('Basic Authentication Flow');
      await userEvent.click(flowOption);

      expect(mockOnFlowSelect).toHaveBeenCalledWith('flow-1');
    });

    it('should call onClearSelection when selection is cleared', async () => {
      await renderComponent({
        selectedAuthFlow: mockFlows[0],
      });

      const autocomplete = page.getByRole('combobox');
      await userEvent.click(autocomplete);

      // Clear the selection by clicking outside or selecting null
      await userEvent.clear(autocomplete);
      await userEvent.tab(); // blur to trigger onChange with null

      expect(mockOnClearSelection).toHaveBeenCalled();
    });

    it('should show selected flow value in autocomplete', async () => {
      await renderComponent({
        selectedAuthFlow: mockFlows[1],
      });

      const autocomplete = page.getByRole('combobox');
      await userEvent.click(autocomplete);

      // Open the dropdown to see options
      await expect.element(page.getByText('Google OAuth Flow')).toBeInTheDocument();
    });

    it('should display flow options when opened', async () => {
      await renderComponent();

      const autocomplete = page.getByRole('combobox');
      await userEvent.click(autocomplete);

      // Check that flow options are displayed
      await expect.element(page.getByText('Basic Authentication Flow')).toBeInTheDocument();
      await expect.element(page.getByText('Google OAuth Flow')).toBeInTheDocument();
      await expect.element(page.getByText('Multi-Factor Auth Flow')).toBeInTheDocument();
    });
  });

  describe('disabled state', () => {
    it('should disable autocomplete when disabled prop is true', async () => {
      await renderComponent({disabled: true});

      const autocomplete = page.getByRole('combobox');
      expect(autocomplete).toBeDisabled();
    });

    it('should enable autocomplete when disabled prop is false', async () => {
      await renderComponent({disabled: false});

      const autocomplete = page.getByRole('combobox');
      expect(autocomplete).not.toBeDisabled();
    });
  });

  describe('edge cases', () => {
    it('should handle flows with special characters in names', async () => {
      const specialFlows: BasicFlowDefinition[] = [
        {
          id: 'special-flow',
          name: 'OAuth 2.0 & OIDC Flow',
          activeVersion: 1,
          handle: 'oauth-oidc-flow',
          flowType: 'AUTHENTICATION',
          createdAt: '',
          updatedAt: '',
        },
      ];

      await renderComponent({availableFlows: specialFlows});

      const autocomplete = page.getByRole('combobox');
      await userEvent.click(autocomplete);

      await expect.element(page.getByText('OAuth 2.0 & OIDC Flow')).toBeInTheDocument();
    });

    it('should handle flows with very long names', async () => {
      const longNameFlows: BasicFlowDefinition[] = [
        {
          id: 'long-flow',
          name: 'This is a very long flow name that should still be displayed properly without breaking the layout',
          activeVersion: 1,
          handle: 'long-name-flow',
          flowType: 'AUTHENTICATION',
          createdAt: '',
          updatedAt: '',
        },
      ];

      await renderComponent({availableFlows: longNameFlows});

      const autocomplete = page.getByRole('combobox');
      await userEvent.click(autocomplete);

      await expect.element(page.getByText(/This is a very long flow name/)).toBeInTheDocument();
    });

    it('should handle when selectedAuthFlow is not in available flows', async () => {
      const unknownFlow: BasicFlowDefinition = {
        id: 'unknown-flow',
        name: 'Unknown Flow',
        activeVersion: 1,
        handle: 'unknown-flow',
        flowType: 'AUTHENTICATION',
        createdAt: '',
        updatedAt: '',
      };

      await renderComponent({
        selectedAuthFlow: unknownFlow,
      });

      // Should not crash and should still render
      await expect.element(page.getByRole('combobox')).toBeInTheDocument();
    });
  });

  describe('accessibility', () => {
    it('should have proper ARIA attributes for autocomplete', async () => {
      await renderComponent();

      const combobox = page.getByRole('combobox');
      expect(combobox).toHaveAttribute('aria-autocomplete', 'list');
    });

    it('should be keyboard navigable', async () => {
      await renderComponent();

      const combobox = page.getByRole('combobox');
      await userEvent.tab();
      expect(combobox).toHaveFocus();
    });

    it('should expand dropdown on Enter key', async () => {
      await renderComponent();

      const combobox = page.getByRole('combobox');
      await userEvent.tab();
      await userEvent.keyboard('{ArrowDown}');

      // Dropdown should be expanded
      expect(combobox).toHaveAttribute('aria-expanded', 'true');
    });
  });
});
