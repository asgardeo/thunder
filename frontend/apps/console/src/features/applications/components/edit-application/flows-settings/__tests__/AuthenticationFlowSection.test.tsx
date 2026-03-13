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
import AuthenticationFlowSection from '../AuthenticationFlowSection';
import useGetFlows from '../../../../../flows/api/useGetFlows';
import type {Application} from '../../../../models/application';

// Mock the useGetFlows hook
vi.mock('../../../../../flows/api/useGetFlows');

type MockedUseGetFlows = ReturnType<typeof useGetFlows>;

// Mock the SettingsCard component
vi.mock('../../../../../../components/SettingsCard', () => ({
  default: ({title, description, children}: {title: string; description: string; children: React.ReactNode}) => (
    <div data-testid="settings-card">
      <div data-testid="card-title">{title}</div>
      <div data-testid="card-description">{description}</div>
      {children}
    </div>
  ),
}));

describe('AuthenticationFlowSection', () => {
  const mockOnFieldChange = vi.fn();
  const mockApplication: Application = {
    id: 'app-123',
    name: 'Test App',
    auth_flow_id: 'auth-flow-1',
  } as Application;

  const mockAuthFlows = [
    {id: 'auth-flow-1', name: 'Default Auth Flow', handle: 'default-auth'},
    {id: 'auth-flow-2', name: 'Custom Auth Flow', handle: 'custom-auth'},
    {id: 'auth-flow-3', name: 'MFA Auth Flow', handle: 'mfa-auth'},
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
          <AuthenticationFlowSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />
        </MemoryRouter>,
      );

      expect(screen.getByTestId('card-title')).toHaveTextContent('Authentication Flow');
      expect(screen.getByTestId('card-description')).toHaveTextContent(
        'Choose the flow that handles user login and authentication.',
      );
    });

    it('should render autocomplete field', () => {
      vi.mocked(useGetFlows).mockReturnValue({
        data: {flows: mockAuthFlows},
        isLoading: false,
      } as MockedUseGetFlows);

      render(
        <MemoryRouter>
          <AuthenticationFlowSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />
        </MemoryRouter>,
      );

      expect(screen.getByPlaceholderText('Select an authentication flow')).toBeInTheDocument();
      expect(screen.getByText('Select the flow that handles user sign-in for this application.')).toBeInTheDocument();
    });

    it('should display alert when auth flow is selected', () => {
      vi.mocked(useGetFlows).mockReturnValue({
        data: {flows: mockAuthFlows},
        isLoading: false,
      } as MockedUseGetFlows);

      render(
        <MemoryRouter>
          <AuthenticationFlowSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />
        </MemoryRouter>,
      );

      expect(screen.getByRole('alert')).toBeInTheDocument();
    });

    it('should not display alert when no auth flow is selected', () => {
      vi.mocked(useGetFlows).mockReturnValue({
        data: {flows: mockAuthFlows},
        isLoading: false,
      } as MockedUseGetFlows);

      const appWithoutFlow = {...mockApplication, auth_flow_id: undefined};

      render(
        <MemoryRouter>
          <AuthenticationFlowSection application={appWithoutFlow} editedApp={{}} onFieldChange={mockOnFieldChange} />
        </MemoryRouter>,
      );

      expect(screen.queryByRole('alert')).not.toBeInTheDocument();
    });

    it('should display alert when auth flow is in editedApp', () => {
      vi.mocked(useGetFlows).mockReturnValue({
        data: {flows: mockAuthFlows},
        isLoading: false,
      } as MockedUseGetFlows);

      const appWithoutFlow = {...mockApplication, auth_flow_id: undefined};

      render(
        <MemoryRouter>
          <AuthenticationFlowSection
            application={appWithoutFlow}
            editedApp={{auth_flow_id: 'auth-flow-2'}}
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
          <AuthenticationFlowSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />
        </MemoryRouter>,
      );

      expect(screen.getByRole('progressbar')).toBeInTheDocument();
    });

    it('should not show loading indicator when flows are loaded', () => {
      vi.mocked(useGetFlows).mockReturnValue({
        data: {flows: mockAuthFlows},
        isLoading: false,
      } as MockedUseGetFlows);

      render(
        <MemoryRouter>
          <AuthenticationFlowSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />
        </MemoryRouter>,
      );

      expect(screen.queryByRole('progressbar')).not.toBeInTheDocument();
    });
  });

  describe('Flow Selection', () => {
    it('should display selected flow from application', () => {
      vi.mocked(useGetFlows).mockReturnValue({
        data: {flows: mockAuthFlows},
        isLoading: false,
      } as MockedUseGetFlows);

      render(
        <MemoryRouter>
          <AuthenticationFlowSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />
        </MemoryRouter>,
      );

      const input = screen.getByPlaceholderText('Select an authentication flow');
      expect(input).toHaveValue('Default Auth Flow');
    });

    it('should display selected flow from editedApp over application', () => {
      vi.mocked(useGetFlows).mockReturnValue({
        data: {flows: mockAuthFlows},
        isLoading: false,
      } as MockedUseGetFlows);

      render(
        <MemoryRouter>
          <AuthenticationFlowSection
            application={mockApplication}
            editedApp={{auth_flow_id: 'auth-flow-2'}}
            onFieldChange={mockOnFieldChange}
          />
        </MemoryRouter>,
      );

      const input = screen.getByPlaceholderText('Select an authentication flow');
      expect(input).toHaveValue('Custom Auth Flow');
    });

    it('should handle flow selection', async () => {
      const user = userEvent.setup();
      vi.mocked(useGetFlows).mockReturnValue({
        data: {flows: mockAuthFlows},
        isLoading: false,
      } as MockedUseGetFlows);

      render(
        <MemoryRouter>
          <AuthenticationFlowSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />
        </MemoryRouter>,
      );

      const input = screen.getByPlaceholderText('Select an authentication flow');
      await user.click(input);

      await waitFor(() => {
        expect(screen.getByText('MFA Auth Flow')).toBeInTheDocument();
      });

      await user.click(screen.getByText('MFA Auth Flow'));

      expect(mockOnFieldChange).toHaveBeenCalledWith('auth_flow_id', 'auth-flow-3');
    });

    it('should handle clearing selection', async () => {
      const user = userEvent.setup();
      vi.mocked(useGetFlows).mockReturnValue({
        data: {flows: mockAuthFlows},
        isLoading: false,
      } as MockedUseGetFlows);

      render(
        <MemoryRouter>
          <AuthenticationFlowSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />
        </MemoryRouter>,
      );

      const clearButton = screen.getByTitle('Clear');
      await user.click(clearButton);

      expect(mockOnFieldChange).toHaveBeenCalledWith('auth_flow_id', '');
    });
  });

  describe('Flow Options Display', () => {
    it('should display flow name and handle in options', async () => {
      const user = userEvent.setup();
      vi.mocked(useGetFlows).mockReturnValue({
        data: {flows: mockAuthFlows},
        isLoading: false,
      } as MockedUseGetFlows);

      render(
        <MemoryRouter>
          <AuthenticationFlowSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />
        </MemoryRouter>,
      );

      const input = screen.getByPlaceholderText('Select an authentication flow');
      await user.click(input);

      await waitFor(() => {
        expect(screen.getByText('Custom Auth Flow')).toBeInTheDocument();
        expect(screen.getByText('custom-auth')).toBeInTheDocument();
      });
    });

    it('should display all available flows in dropdown', async () => {
      const user = userEvent.setup();
      vi.mocked(useGetFlows).mockReturnValue({
        data: {flows: mockAuthFlows},
        isLoading: false,
      } as MockedUseGetFlows);

      render(
        <MemoryRouter>
          <AuthenticationFlowSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />
        </MemoryRouter>,
      );

      const input = screen.getByPlaceholderText('Select an authentication flow');
      await user.click(input);

      await waitFor(() => {
        expect(screen.getByText('Default Auth Flow')).toBeInTheDocument();
        expect(screen.getByText('Custom Auth Flow')).toBeInTheDocument();
        expect(screen.getByText('MFA Auth Flow')).toBeInTheDocument();
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
          <AuthenticationFlowSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />
        </MemoryRouter>,
      );

      expect(screen.getByPlaceholderText('Select an authentication flow')).toBeInTheDocument();
    });

    it('should handle undefined flows data', () => {
      vi.mocked(useGetFlows).mockReturnValue({
        data: undefined,
        isLoading: false,
      } as MockedUseGetFlows);

      render(
        <MemoryRouter>
          <AuthenticationFlowSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />
        </MemoryRouter>,
      );

      expect(screen.getByPlaceholderText('Select an authentication flow')).toBeInTheDocument();
    });
  });

  describe('Alert Links', () => {
    it('should display edit link with correct flow ID from application', () => {
      vi.mocked(useGetFlows).mockReturnValue({
        data: {flows: mockAuthFlows},
        isLoading: false,
      } as MockedUseGetFlows);

      const {container} = render(
        <MemoryRouter>
          <AuthenticationFlowSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />
        </MemoryRouter>,
      );

      const links = container.querySelectorAll('a');
      const editLink = Array.from(links).find((link) => link.getAttribute('href')?.includes('/flows/signin/'));
      expect(editLink).toHaveAttribute('href', '/flows/signin/auth-flow-1');
    });

    it('should display edit link with correct flow ID from editedApp', () => {
      vi.mocked(useGetFlows).mockReturnValue({
        data: {flows: mockAuthFlows},
        isLoading: false,
      } as MockedUseGetFlows);

      const {container} = render(
        <MemoryRouter>
          <AuthenticationFlowSection
            application={mockApplication}
            editedApp={{auth_flow_id: 'auth-flow-2'}}
            onFieldChange={mockOnFieldChange}
          />
        </MemoryRouter>,
      );

      const links = container.querySelectorAll('a');
      const editLink = Array.from(links).find((link) => link.getAttribute('href')?.includes('/flows/signin/'));
      expect(editLink).toHaveAttribute('href', '/flows/signin/auth-flow-2');
    });

    it('should display create link', () => {
      vi.mocked(useGetFlows).mockReturnValue({
        data: {flows: mockAuthFlows},
        isLoading: false,
      } as MockedUseGetFlows);

      const {container} = render(
        <MemoryRouter>
          <AuthenticationFlowSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />
        </MemoryRouter>,
      );

      const links = container.querySelectorAll('a');
      const createLink = Array.from(links).find((link) => link.getAttribute('href') === '/flows');
      expect(createLink).toHaveAttribute('href', '/flows');
    });
  });
});
