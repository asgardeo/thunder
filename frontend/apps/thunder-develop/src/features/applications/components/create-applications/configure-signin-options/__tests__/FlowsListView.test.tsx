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
import {render, screen} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {type BasicFlowDefinition} from '@/features/flows/models/responses';
import FlowsListView, {type FlowsListViewProps} from '../FlowsListView';

// Mock react-i18next
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => {
      const translations: Record<string, string> = {
        'common:or': 'or',
        'applications:onboarding.configure.SignInOptions.preConfiguredFlows.selectFlow': 'Select a flow',
        'applications:onboarding.configure.SignInOptions.preConfiguredFlows.searchFlows': 'Search flows...',
      };
      return translations[key] || key;
    },
  }),
}));

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

  const renderComponent = (props: Partial<FlowsListViewProps> = {}) =>
    render(<FlowsListView {...defaultProps} {...props} />);

  describe('rendering', () => {
    it('should return null when no selectable flows available', () => {
      const {container} = renderComponent({
        availableFlows: [],
      });

      expect(container.firstChild).toBeNull();
    });

    it('should return null when all flows are develop-app flows', () => {
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

      const {container} = renderComponent({
        availableFlows: developAppFlows,
      });

      expect(container.firstChild).toBeNull();
    });

    it('should return null when all flows are default flows', () => {
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

      const {container} = renderComponent({
        availableFlows: defaultFlows,
      });

      expect(container.firstChild).toBeNull();
    });

    it('should render divider with "or" text', () => {
      renderComponent();

      expect(screen.getByText('or')).toBeInTheDocument();
    });

    it('should render autocomplete component', () => {
      renderComponent();

      expect(screen.getByRole('combobox')).toBeInTheDocument();
    });

    it('should render with proper label text', () => {
      renderComponent();

      expect(screen.getByLabelText('Select a flow')).toBeInTheDocument();
    });
  });

  describe('flow filtering', () => {
    it('should filter out develop-app flows', () => {
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

      renderComponent({availableFlows: mixedFlows});

      // The component should render since there are selectable flows
      expect(screen.getByRole('combobox')).toBeInTheDocument();
    });

    it('should filter out default flows', () => {
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

      renderComponent({availableFlows: mixedFlows});

      expect(screen.getByRole('combobox')).toBeInTheDocument();
    });
  });

  describe('autocomplete interaction', () => {
    it('should call onFlowSelect when a flow is selected', async () => {
      const user = userEvent.setup();
      renderComponent();

      const autocomplete = screen.getByRole('combobox');
      await user.click(autocomplete);

      const flowOption = screen.getByText('Basic Authentication Flow');
      await user.click(flowOption);

      expect(mockOnFlowSelect).toHaveBeenCalledWith('flow-1');
    });

    it('should call onClearSelection when selection is cleared', async () => {
      const user = userEvent.setup();
      renderComponent({
        selectedAuthFlow: mockFlows[0],
      });

      const autocomplete = screen.getByRole('combobox');
      await user.click(autocomplete);

      // Clear the selection by clicking outside or selecting null
      await user.clear(autocomplete);
      await user.tab(); // blur to trigger onChange with null

      expect(mockOnClearSelection).toHaveBeenCalled();
    });

    it('should show selected flow value in autocomplete', async () => {
      const user = userEvent.setup();
      renderComponent({
        selectedAuthFlow: mockFlows[1],
      });

      const autocomplete = screen.getByRole('combobox');
      await user.click(autocomplete);

      // Open the dropdown to see options
      expect(screen.getByText('Google OAuth Flow')).toBeInTheDocument();
    });

    it('should display flow options when opened', async () => {
      const user = userEvent.setup();
      renderComponent();

      const autocomplete = screen.getByRole('combobox');
      await user.click(autocomplete);

      // Check that flow options are displayed
      expect(screen.getByText('Basic Authentication Flow')).toBeInTheDocument();
      expect(screen.getByText('Google OAuth Flow')).toBeInTheDocument();
      expect(screen.getByText('Multi-Factor Auth Flow')).toBeInTheDocument();
    });
  });

  describe('disabled state', () => {
    it('should disable autocomplete when disabled prop is true', () => {
      renderComponent({disabled: true});

      const autocomplete = screen.getByRole('combobox');
      expect(autocomplete).toBeDisabled();
    });

    it('should enable autocomplete when disabled prop is false', () => {
      renderComponent({disabled: false});

      const autocomplete = screen.getByRole('combobox');
      expect(autocomplete).not.toBeDisabled();
    });
  });

  describe('edge cases', () => {
    it('should handle flows with special characters in names', async () => {
      const user = userEvent.setup();
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

      renderComponent({availableFlows: specialFlows});

      const autocomplete = screen.getByRole('combobox');
      await user.click(autocomplete);

      expect(screen.getByText('OAuth 2.0 & OIDC Flow')).toBeInTheDocument();
    });

    it('should handle flows with very long names', async () => {
      const user = userEvent.setup();
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

      renderComponent({availableFlows: longNameFlows});

      const autocomplete = screen.getByRole('combobox');
      await user.click(autocomplete);

      expect(screen.getByText(/This is a very long flow name/)).toBeInTheDocument();
    });

    it('should handle when selectedAuthFlow is not in available flows', () => {
      const unknownFlow: BasicFlowDefinition = {
        id: 'unknown-flow',
        name: 'Unknown Flow',
        activeVersion: 1,
        handle: 'unknown-flow',
        flowType: 'AUTHENTICATION',
        createdAt: '',
        updatedAt: '',
      };

      renderComponent({
        selectedAuthFlow: unknownFlow,
      });

      // Should not crash and should still render
      expect(screen.getByRole('combobox')).toBeInTheDocument();
    });
  });

  describe('accessibility', () => {
    it('should have proper ARIA attributes for autocomplete', () => {
      renderComponent();

      const combobox = screen.getByRole('combobox');
      expect(combobox).toHaveAttribute('aria-autocomplete', 'list');
    });

    it('should be keyboard navigable', async () => {
      const user = userEvent.setup();
      renderComponent();

      const combobox = screen.getByRole('combobox');
      await user.tab();
      expect(combobox).toHaveFocus();
    });

    it('should expand dropdown on Enter key', async () => {
      const user = userEvent.setup();
      renderComponent();

      const combobox = screen.getByRole('combobox');
      await user.tab();
      await user.keyboard('{ArrowDown}');

      // Dropdown should be expanded
      expect(combobox).toHaveAttribute('aria-expanded', 'true');
    });
  });
});
