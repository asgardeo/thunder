/**
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
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
import {render, screen, waitFor} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {MemoryRouter} from 'react-router';
import RegistrationFlowSection from '../RegistrationFlowSection';
import useGetFlows from '../../../../../flows/api/useGetFlows';
import type {Application} from '../../../../models/application';

// Mock the useGetFlows hook
vi.mock('../../../../../flows/api/useGetFlows');

type MockedUseGetFlows = ReturnType<typeof useGetFlows>;

// Mock the SettingsCard component
vi.mock('../../../../../../components/SettingsCard', () => ({
  default: ({
    title,
    description,
    enabled,
    onToggle,
    children,
  }: {
    title: string;
    description: string;
    enabled?: boolean;
    onToggle?: (enabled: boolean) => void;
    children: React.ReactNode;
  }) => (
    <div data-testid="settings-card">
      <div data-testid="card-title">{title}</div>
      <div data-testid="card-description">{description}</div>
      {onToggle && (
        <button type="button" data-testid="toggle-button" onClick={() => onToggle(!enabled)}>
          Toggle: {enabled ? 'ON' : 'OFF'}
        </button>
      )}
      {children}
    </div>
  ),
}));

describe('RegistrationFlowSection', () => {
  const mockOnFieldChange = vi.fn();
  const mockApplication: Application = {
    id: 'app-123',
    name: 'Test App',
    registration_flow_id: 'reg-flow-1',
    is_registration_flow_enabled: true,
  } as Application;

  const mockRegFlows = [
    {id: 'reg-flow-1', name: 'Default Registration Flow', handle: 'default-reg'},
    {id: 'reg-flow-2', name: 'Custom Registration Flow', handle: 'custom-reg'},
    {id: 'reg-flow-3', name: 'SSO Registration Flow', handle: 'sso-reg'},
  ];

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('should render the settings card with title and description', () => {
      vi.mocked(useGetFlows).mockReturnValue({
        data: {flows: []},
        isLoading: false,
      } as unknown as MockedUseGetFlows);

      render(
        <MemoryRouter>
          <RegistrationFlowSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />
        </MemoryRouter>,
      );

      expect(screen.getByTestId('card-title')).toHaveTextContent('Registration Flow');
      expect(screen.getByTestId('card-description')).toHaveTextContent(
        'Choose the flow that handles user sign-up and account creation.',
      );
    });

    it('should render autocomplete field', () => {
      vi.mocked(useGetFlows).mockReturnValue({
        data: {flows: mockRegFlows},
        isLoading: false,
      } as MockedUseGetFlows);

      render(
        <MemoryRouter>
          <RegistrationFlowSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />
        </MemoryRouter>,
      );

      expect(screen.getByPlaceholderText('Select a registration flow')).toBeInTheDocument();
      expect(
        screen.getByText('Select the flow that handles user registration for this application.'),
      ).toBeInTheDocument();
    });

    it('should render toggle button', () => {
      vi.mocked(useGetFlows).mockReturnValue({
        data: {flows: mockRegFlows},
        isLoading: false,
      } as MockedUseGetFlows);

      render(
        <MemoryRouter>
          <RegistrationFlowSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />
        </MemoryRouter>,
      );

      expect(screen.getByTestId('toggle-button')).toBeInTheDocument();
    });

    it('should display alert when registration flow is selected', () => {
      vi.mocked(useGetFlows).mockReturnValue({
        data: {flows: mockRegFlows},
        isLoading: false,
      } as MockedUseGetFlows);

      render(
        <MemoryRouter>
          <RegistrationFlowSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />
        </MemoryRouter>,
      );

      expect(screen.getByRole('alert')).toBeInTheDocument();
    });

    it('should not display alert when no registration flow is selected', () => {
      vi.mocked(useGetFlows).mockReturnValue({
        data: {flows: mockRegFlows},
        isLoading: false,
      } as MockedUseGetFlows);

      const appWithoutFlow = {...mockApplication, registration_flow_id: undefined};

      render(
        <MemoryRouter>
          <RegistrationFlowSection application={appWithoutFlow} editedApp={{}} onFieldChange={mockOnFieldChange} />
        </MemoryRouter>,
      );

      expect(screen.queryByRole('alert')).not.toBeInTheDocument();
    });

    it('should display alert when registration flow is in editedApp', () => {
      vi.mocked(useGetFlows).mockReturnValue({
        data: {flows: mockRegFlows},
        isLoading: false,
      } as MockedUseGetFlows);

      const appWithoutFlow = {...mockApplication, registration_flow_id: undefined};

      render(
        <MemoryRouter>
          <RegistrationFlowSection
            application={appWithoutFlow}
            editedApp={{registration_flow_id: 'reg-flow-2'}}
            onFieldChange={mockOnFieldChange}
          />
        </MemoryRouter>,
      );

      expect(screen.getByRole('alert')).toBeInTheDocument();
    });
  });

  describe('Loading State', () => {
    it('should show loading indicator while fetching flows', () => {
      vi.mocked(useGetFlows).mockReturnValue({
        data: undefined,
        isLoading: true,
      } as MockedUseGetFlows);

      render(
        <MemoryRouter>
          <RegistrationFlowSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />
        </MemoryRouter>,
      );

      expect(screen.getByRole('progressbar')).toBeInTheDocument();
    });

    it('should not show loading indicator when flows are loaded', () => {
      vi.mocked(useGetFlows).mockReturnValue({
        data: {flows: mockRegFlows},
        isLoading: false,
      } as MockedUseGetFlows);

      render(
        <MemoryRouter>
          <RegistrationFlowSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />
        </MemoryRouter>,
      );

      expect(screen.queryByRole('progressbar')).not.toBeInTheDocument();
    });
  });

  describe('Enable/Disable Toggle', () => {
    it('should pass enabled state from application to SettingsCard', () => {
      vi.mocked(useGetFlows).mockReturnValue({
        data: {flows: mockRegFlows},
        isLoading: false,
      } as MockedUseGetFlows);

      render(
        <MemoryRouter>
          <RegistrationFlowSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />
        </MemoryRouter>,
      );

      expect(screen.getByTestId('toggle-button')).toHaveTextContent('Toggle: ON');
    });

    it('should pass enabled state from editedApp to SettingsCard', () => {
      vi.mocked(useGetFlows).mockReturnValue({
        data: {flows: mockRegFlows},
        isLoading: false,
      } as MockedUseGetFlows);

      render(
        <MemoryRouter>
          <RegistrationFlowSection
            application={mockApplication}
            editedApp={{is_registration_flow_enabled: false}}
            onFieldChange={mockOnFieldChange}
          />
        </MemoryRouter>,
      );

      expect(screen.getByTestId('toggle-button')).toHaveTextContent('Toggle: OFF');
    });

    it('should default to false when is_registration_flow_enabled is undefined', () => {
      vi.mocked(useGetFlows).mockReturnValue({
        data: {flows: mockRegFlows},
        isLoading: false,
      } as MockedUseGetFlows);

      const appWithoutEnabled = {...mockApplication, is_registration_flow_enabled: undefined};

      render(
        <MemoryRouter>
          <RegistrationFlowSection application={appWithoutEnabled} editedApp={{}} onFieldChange={mockOnFieldChange} />
        </MemoryRouter>,
      );

      expect(screen.getByTestId('toggle-button')).toHaveTextContent('Toggle: OFF');
    });

    it('should call onFieldChange when toggle is clicked', async () => {
      const user = userEvent.setup();
      vi.mocked(useGetFlows).mockReturnValue({
        data: {flows: mockRegFlows},
        isLoading: false,
      } as MockedUseGetFlows);

      render(
        <MemoryRouter>
          <RegistrationFlowSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />
        </MemoryRouter>,
      );

      await user.click(screen.getByTestId('toggle-button'));

      expect(mockOnFieldChange).toHaveBeenCalledWith('is_registration_flow_enabled', false);
    });
  });

  describe('Flow Selection', () => {
    it('should display selected flow from application', () => {
      vi.mocked(useGetFlows).mockReturnValue({
        data: {flows: mockRegFlows},
        isLoading: false,
      } as MockedUseGetFlows);

      render(
        <MemoryRouter>
          <RegistrationFlowSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />
        </MemoryRouter>,
      );

      const input = screen.getByPlaceholderText('Select a registration flow');
      expect(input).toHaveValue('Default Registration Flow');
    });

    it('should display selected flow from editedApp over application', () => {
      vi.mocked(useGetFlows).mockReturnValue({
        data: {flows: mockRegFlows},
        isLoading: false,
      } as MockedUseGetFlows);

      render(
        <MemoryRouter>
          <RegistrationFlowSection
            application={mockApplication}
            editedApp={{registration_flow_id: 'reg-flow-2'}}
            onFieldChange={mockOnFieldChange}
          />
        </MemoryRouter>,
      );

      const input = screen.getByPlaceholderText('Select a registration flow');
      expect(input).toHaveValue('Custom Registration Flow');
    });

    it('should handle flow selection', async () => {
      const user = userEvent.setup();
      vi.mocked(useGetFlows).mockReturnValue({
        data: {flows: mockRegFlows},
        isLoading: false,
      } as MockedUseGetFlows);

      render(
        <MemoryRouter>
          <RegistrationFlowSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />
        </MemoryRouter>,
      );

      const input = screen.getByPlaceholderText('Select a registration flow');
      await user.click(input);

      await waitFor(() => {
        expect(screen.getByText('SSO Registration Flow')).toBeInTheDocument();
      });

      await user.click(screen.getByText('SSO Registration Flow'));

      expect(mockOnFieldChange).toHaveBeenCalledWith('registration_flow_id', 'reg-flow-3');
    });

    it('should handle clearing selection', async () => {
      const user = userEvent.setup();
      vi.mocked(useGetFlows).mockReturnValue({
        data: {flows: mockRegFlows},
        isLoading: false,
      } as MockedUseGetFlows);

      render(
        <MemoryRouter>
          <RegistrationFlowSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />
        </MemoryRouter>,
      );

      const clearButton = screen.getByTitle('Clear');
      await user.click(clearButton);

      expect(mockOnFieldChange).toHaveBeenCalledWith('registration_flow_id', '');
    });
  });

  describe('Flow Options Display', () => {
    it('should display flow name and handle in options', async () => {
      const user = userEvent.setup();
      vi.mocked(useGetFlows).mockReturnValue({
        data: {flows: mockRegFlows},
        isLoading: false,
      } as MockedUseGetFlows);

      render(
        <MemoryRouter>
          <RegistrationFlowSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />
        </MemoryRouter>,
      );

      const input = screen.getByPlaceholderText('Select a registration flow');
      await user.click(input);

      await waitFor(() => {
        expect(screen.getByText('Custom Registration Flow')).toBeInTheDocument();
        expect(screen.getByText('custom-reg')).toBeInTheDocument();
      });
    });

    it('should display all available flows in dropdown', async () => {
      const user = userEvent.setup();
      vi.mocked(useGetFlows).mockReturnValue({
        data: {flows: mockRegFlows},
        isLoading: false,
      } as MockedUseGetFlows);

      render(
        <MemoryRouter>
          <RegistrationFlowSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />
        </MemoryRouter>,
      );

      const input = screen.getByPlaceholderText('Select a registration flow');
      await user.click(input);

      await waitFor(() => {
        expect(screen.getByText('Default Registration Flow')).toBeInTheDocument();
        expect(screen.getByText('Custom Registration Flow')).toBeInTheDocument();
        expect(screen.getByText('SSO Registration Flow')).toBeInTheDocument();
      });
    });
  });

  describe('Empty State', () => {
    it('should handle empty flows array', () => {
      vi.mocked(useGetFlows).mockReturnValue({
        data: {flows: []},
        isLoading: false,
      } as unknown as MockedUseGetFlows);

      render(
        <MemoryRouter>
          <RegistrationFlowSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />
        </MemoryRouter>,
      );

      expect(screen.getByPlaceholderText('Select a registration flow')).toBeInTheDocument();
    });

    it('should handle undefined flows data', () => {
      vi.mocked(useGetFlows).mockReturnValue({
        data: undefined,
        isLoading: false,
      } as MockedUseGetFlows);

      render(
        <MemoryRouter>
          <RegistrationFlowSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />
        </MemoryRouter>,
      );

      expect(screen.getByPlaceholderText('Select a registration flow')).toBeInTheDocument();
    });
  });

  describe('Alert Links', () => {
    it('should display edit link with correct flow ID from application', () => {
      vi.mocked(useGetFlows).mockReturnValue({
        data: {flows: mockRegFlows},
        isLoading: false,
      } as MockedUseGetFlows);

      const {container} = render(
        <MemoryRouter>
          <RegistrationFlowSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />
        </MemoryRouter>,
      );

      const links = container.querySelectorAll('a');
      const editLink = Array.from(links).find((link) => link.getAttribute('href')?.includes('/flows/signup/'));
      expect(editLink).toHaveAttribute('href', '/flows/signup/reg-flow-1');
    });

    it('should display edit link with correct flow ID from editedApp', () => {
      vi.mocked(useGetFlows).mockReturnValue({
        data: {flows: mockRegFlows},
        isLoading: false,
      } as MockedUseGetFlows);

      const {container} = render(
        <MemoryRouter>
          <RegistrationFlowSection
            application={mockApplication}
            editedApp={{registration_flow_id: 'reg-flow-2'}}
            onFieldChange={mockOnFieldChange}
          />
        </MemoryRouter>,
      );

      const links = container.querySelectorAll('a');
      const editLink = Array.from(links).find((link) => link.getAttribute('href')?.includes('/flows/signup/'));
      expect(editLink).toHaveAttribute('href', '/flows/signup/reg-flow-2');
    });

    it('should display create link', () => {
      vi.mocked(useGetFlows).mockReturnValue({
        data: {flows: mockRegFlows},
        isLoading: false,
      } as MockedUseGetFlows);

      const {container} = render(
        <MemoryRouter>
          <RegistrationFlowSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />
        </MemoryRouter>,
      );

      const links = container.querySelectorAll('a');
      const createLink = Array.from(links).find((link) => link.getAttribute('href') === '/flows');
      expect(createLink).toHaveAttribute('href', '/flows');
    });
  });
});
