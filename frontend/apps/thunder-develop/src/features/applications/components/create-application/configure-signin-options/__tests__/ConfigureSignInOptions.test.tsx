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
import type {BasicFlowDefinition} from '@/features/flows/models/responses';
import {IdentityProviderTypes, type IdentityProvider} from '@/features/integrations/models/identity-provider';
import {AuthenticatorTypes} from '@/features/integrations/models/authenticators';
import findMatchingFlowForIntegrations from '@/features/flows/utils/findMatchingFlowForIntegrations';
import ConfigureSignInOptions, {type ConfigureSignInOptionsProps} from '../ConfigureSignInOptions';

// Mock useIdentityProviders
interface MockIdentityProviderResponse {
  data: IdentityProvider[] | null;
  isLoading: boolean;
  error: Error | null;
}
const mockUseIdentityProviders = vi.fn<() => MockIdentityProviderResponse>();
vi.mock('@/features/integrations/api/useIdentityProviders', () => ({
  default: () => mockUseIdentityProviders(),
}));

// Mock useGetFlows
interface MockFlowsResponse {
  data: {flows: BasicFlowDefinition[]} | null;
  isLoading: boolean;
  error: Error | null;
}
const mockUseGetFlows = vi.fn<() => MockFlowsResponse>();
vi.mock('@/features/flows/api/useGetFlows', () => ({
  default: () => mockUseGetFlows(),
}));

// Mock useApplicationCreateContext - need to mock the correct path used by the component
const mockSetSelectedAuthFlow = vi.fn();
const mockSelectedAuthFlow: BasicFlowDefinition | null = null;
vi.mock('@/features/applications/hooks/useApplicationCreateContext', () => ({
  default: () => ({
    selectedAuthFlow: mockSelectedAuthFlow,
    setSelectedAuthFlow: mockSetSelectedAuthFlow,
  }),
}));

// Mock findMatchingFlowForIntegrations
vi.mock('@/features/flows/utils/findMatchingFlowForIntegrations', () => ({
  default: vi.fn(() => null),
}));

// Mock child components
vi.mock('../FlowsListView', () => ({
  default: ({onFlowSelect, onClearSelection}: {onFlowSelect: (id: string) => void; onClearSelection: () => void}) => (
    <div data-testid="flows-list-view">
      <button type="button" data-testid="select-flow-btn" onClick={() => onFlowSelect('flow-1')}>
        Select Flow
      </button>
      <button type="button" data-testid="clear-flow-btn" onClick={() => onClearSelection()}>
        Clear
      </button>
    </div>
  ),
}));

vi.mock('../IndividualMethodsToggleView', () => ({
  default: ({onIntegrationToggle}: {onIntegrationToggle: (id: string) => void}) => (
    <div data-testid="individual-methods-view">
      <button
        type="button"
        data-testid="toggle-basic-auth"
        onClick={() => onIntegrationToggle(AuthenticatorTypes.BASIC_AUTH)}
      >
        Toggle Basic Auth
      </button>
      <button type="button" data-testid="toggle-google" onClick={() => onIntegrationToggle('google-idp')}>
        Toggle Google
      </button>
    </div>
  ),
}));

describe('ConfigureSignInOptions', () => {
  const mockOnIntegrationToggle = vi.fn();
  const mockOnReadyChange = vi.fn();

  const mockIdentityProviders: IdentityProvider[] = [
    {
      id: 'google-idp',
      name: 'Google',
      type: IdentityProviderTypes.GOOGLE,
      description: 'Sign in with Google',
    },
    {
      id: 'github-idp',
      name: 'GitHub',
      type: IdentityProviderTypes.GITHUB,
      description: 'Sign in with GitHub',
    },
  ];

  const mockFlows: BasicFlowDefinition[] = [
    {
      id: 'flow-1',
      name: 'Basic Flow',
    } as BasicFlowDefinition,
  ];

  const defaultProps: ConfigureSignInOptionsProps = {
    integrations: {
      [AuthenticatorTypes.BASIC_AUTH]: false,
    },
    onIntegrationToggle: mockOnIntegrationToggle,
    onReadyChange: mockOnReadyChange,
  };

  beforeEach(() => {
    vi.clearAllMocks();
    mockUseIdentityProviders.mockReturnValue({
      data: mockIdentityProviders,
      isLoading: false,
      error: null,
    });
    mockUseGetFlows.mockReturnValue({
      data: {flows: mockFlows},
      isLoading: false,
      error: null,
    });
  });

  const renderComponent = async (props: Partial<ConfigureSignInOptionsProps> = {}) =>
    render(<ConfigureSignInOptions {...defaultProps} {...props} />);

  describe('rendering', () => {
    it('should render title and subtitle', async () => {
      await renderComponent();

      await expect.element(page.getByText('Sign In Options')).toBeInTheDocument();
      await expect.element(page.getByText('Choose how users will sign-in to your application')).toBeInTheDocument();
    });

    it('should render IndividualMethodsToggleView', async () => {
      await renderComponent();

      await expect.element(page.getByTestId('individual-methods-view')).toBeInTheDocument();
    });

    it('should render FlowsListView', async () => {
      await renderComponent();

      await expect.element(page.getByTestId('flows-list-view')).toBeInTheDocument();
    });

    it('should render hint text', async () => {
      await renderComponent();

      await expect.element(page.getByText('You can always change these settings later in the application settings.')).toBeInTheDocument();
    });
  });

  describe('loading state', () => {
    it('should show loading spinner when identity providers are loading', async () => {
      mockUseIdentityProviders.mockReturnValue({
        data: null,
        isLoading: true,
        error: null,
      });

      await renderComponent();

      await expect.element(page.getByRole('progressbar')).toBeInTheDocument();
    });

    it('should show loading spinner when flows are loading', async () => {
      mockUseGetFlows.mockReturnValue({
        data: null,
        isLoading: true,
        error: null,
      });

      await renderComponent();

      await expect.element(page.getByRole('progressbar')).toBeInTheDocument();
    });
  });

  describe('error state', () => {
    it('should show error alert when identity providers fetch fails', async () => {
      mockUseIdentityProviders.mockReturnValue({
        data: null,
        isLoading: false,
        error: new Error('Failed to fetch providers'),
      });

      await renderComponent();

      await expect.element(page.getByRole('alert')).toBeInTheDocument();
      await expect.element(page.getByText('Failed to load authentication methods: Failed to fetch providers')).toBeInTheDocument();
    });

    it('should show error alert when flows fetch fails', async () => {
      mockUseGetFlows.mockReturnValue({
        data: null,
        isLoading: false,
        error: new Error('Failed to fetch flows'),
      });

      await renderComponent();

      await expect.element(page.getByRole('alert')).toBeInTheDocument();
    });
  });

  describe('validation warning', () => {
    it('should show warning when no options are selected', async () => {
      await renderComponent({
        integrations: {
          [AuthenticatorTypes.BASIC_AUTH]: false,
        },
      });

      await expect.element(
        page.getByText('At least one sign-in option is required. Please select at least one authentication method.'),
      ).toBeInTheDocument();
    });

    it('should not show warning when at least one option is selected', async () => {
      await renderComponent({
        integrations: {
          [AuthenticatorTypes.BASIC_AUTH]: true,
        },
      });

      await expect.element(
        page.getByText('At least one sign-in option is required. Please select at least one authentication method.'),
      ).not.toBeInTheDocument();
    });
  });

  describe('integration toggle', () => {
    it('should call onIntegrationToggle when toggling an integration', async () => {
      await renderComponent();

      await userEvent.click(page.getByTestId('toggle-basic-auth'));

      expect(mockOnIntegrationToggle).toHaveBeenCalledWith(AuthenticatorTypes.BASIC_AUTH);
    });

    it('should select matching flow when integration toggle matches a flow', async () => {
      const mockedFindMatchingFlow = vi.mocked(findMatchingFlowForIntegrations);
      mockedFindMatchingFlow.mockReturnValue(mockFlows[0]);

      await renderComponent();

      await userEvent.click(page.getByTestId('toggle-basic-auth'));

      expect(mockSetSelectedAuthFlow).toHaveBeenCalledWith(mockFlows[0]);
    });
  });

  describe('flow selection', () => {
    it('should call setSelectedAuthFlow when selecting a flow', async () => {
      await renderComponent();

      await userEvent.click(page.getByTestId('select-flow-btn'));

      expect(mockSetSelectedAuthFlow).toHaveBeenCalled();
    });

    it('should call setSelectedAuthFlow with null when clearing selection', async () => {
      await renderComponent();

      await userEvent.click(page.getByTestId('clear-flow-btn'));

      expect(mockSetSelectedAuthFlow).toHaveBeenCalledWith(null);
    });
  });

  describe('onReadyChange callback', () => {
    it('should call onReadyChange with false when no options are selected', async () => {
      await renderComponent({
        integrations: {
          [AuthenticatorTypes.BASIC_AUTH]: false,
        },
      });

      expect(mockOnReadyChange).toHaveBeenCalledWith(false);
    });

    it('should call onReadyChange with true when at least one option is selected', async () => {
      await renderComponent({
        integrations: {
          [AuthenticatorTypes.BASIC_AUTH]: true,
        },
      });

      expect(mockOnReadyChange).toHaveBeenCalledWith(true);
    });

    it('should not throw when onReadyChange is not provided', async () => {
      await render(
        <ConfigureSignInOptions
          integrations={{[AuthenticatorTypes.BASIC_AUTH]: false}}
          onIntegrationToggle={mockOnIntegrationToggle}
        />,
      );
      // If we get here without throwing, the test passes
    });
  });

  describe('empty data handling', () => {
    it('should handle empty identity providers list', async () => {
      mockUseIdentityProviders.mockReturnValue({
        data: [],
        isLoading: false,
        error: null,
      });

      await renderComponent();

      await expect.element(page.getByTestId('individual-methods-view')).toBeInTheDocument();
    });

    it('should handle empty flows list', async () => {
      mockUseGetFlows.mockReturnValue({
        data: {flows: []},
        isLoading: false,
        error: null,
      });

      await renderComponent();

      await expect.element(page.getByTestId('flows-list-view')).toBeInTheDocument();
    });

    it('should handle null data from useIdentityProviders', async () => {
      mockUseIdentityProviders.mockReturnValue({
        data: null,
        isLoading: false,
        error: null,
      });

      await renderComponent();

      await expect.element(page.getByTestId('individual-methods-view')).toBeInTheDocument();
    });

    it('should handle null flows from useGetFlows', async () => {
      mockUseGetFlows.mockReturnValue({
        data: null,
        isLoading: false,
        error: null,
      });

      await renderComponent();

      await expect.element(page.getByTestId('flows-list-view')).toBeInTheDocument();
    });
  });
});
