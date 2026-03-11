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
import {render} from '@thunder/test-utils/browser';
import {page, userEvent} from 'vitest/browser';

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
    it('should render the settings card with title and description', async () => {
      vi.mocked(useGetFlows).mockReturnValue({
        data: {flows: []},
        isLoading: false,
      } as unknown as MockedUseGetFlows);

      await render(
          <AuthenticationFlowSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />
      );

      await expect.element(page.getByTestId('card-title')).toHaveTextContent('Authentication Flow');
      await expect.element(page.getByTestId('card-description')).toHaveTextContent(
        'Choose the flow that handles user login and authentication.',
      );
    });

    it('should render autocomplete field', async () => {
      vi.mocked(useGetFlows).mockReturnValue({
        data: {flows: mockAuthFlows},
        isLoading: false,
      } as MockedUseGetFlows);

      await render(
          <AuthenticationFlowSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />
      );

      await expect.element(page.getByPlaceholder('Select an authentication flow')).toBeInTheDocument();
      await expect.element(page.getByText('Select the flow that handles user sign-in for this application.')).toBeInTheDocument();
    });

    it('should display alert when auth flow is selected', async () => {
      vi.mocked(useGetFlows).mockReturnValue({
        data: {flows: mockAuthFlows},
        isLoading: false,
      } as MockedUseGetFlows);

      await render(
          <AuthenticationFlowSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />
      );

      await expect.element(page.getByRole('alert')).toBeInTheDocument();
    });

    it('should not display alert when no auth flow is selected', async () => {
      vi.mocked(useGetFlows).mockReturnValue({
        data: {flows: mockAuthFlows},
        isLoading: false,
      } as MockedUseGetFlows);

      const appWithoutFlow = {...mockApplication, auth_flow_id: undefined};

      await render(
          <AuthenticationFlowSection application={appWithoutFlow} editedApp={{}} onFieldChange={mockOnFieldChange} />
      );

      await expect.element(page.getByRole('alert')).not.toBeInTheDocument();
    });

    it('should display alert when auth flow is in editedApp', async () => {
      vi.mocked(useGetFlows).mockReturnValue({
        data: {flows: mockAuthFlows},
        isLoading: false,
      } as MockedUseGetFlows);

      const appWithoutFlow = {...mockApplication, auth_flow_id: undefined};

      await render(
          <AuthenticationFlowSection
            application={appWithoutFlow}
            editedApp={{auth_flow_id: 'auth-flow-2'}}
            onFieldChange={mockOnFieldChange}
          />
      );

      await expect.element(page.getByRole('alert')).toBeInTheDocument();
    });
  });

  describe('Loading State', () => {
    it('should show loading indicator while fetching flows', async () => {
      vi.mocked(useGetFlows).mockReturnValue({
        data: undefined,
        isLoading: true,
      } as MockedUseGetFlows);

      await render(
          <AuthenticationFlowSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />
      );

      await expect.element(page.getByRole('progressbar')).toBeInTheDocument();
    });

    it('should not show loading indicator when flows are loaded', async () => {
      vi.mocked(useGetFlows).mockReturnValue({
        data: {flows: mockAuthFlows},
        isLoading: false,
      } as MockedUseGetFlows);

      await render(
          <AuthenticationFlowSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />
      );

      await expect.element(page.getByRole('progressbar')).not.toBeInTheDocument();
    });
  });

  describe('Flow Selection', () => {
    it('should display selected flow from application', async () => {
      vi.mocked(useGetFlows).mockReturnValue({
        data: {flows: mockAuthFlows},
        isLoading: false,
      } as MockedUseGetFlows);

      await render(
          <AuthenticationFlowSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />
      );

      const input = page.getByPlaceholder('Select an authentication flow');
      expect(input).toHaveValue('Default Auth Flow');
    });

    it('should display selected flow from editedApp over application', async () => {
      vi.mocked(useGetFlows).mockReturnValue({
        data: {flows: mockAuthFlows},
        isLoading: false,
      } as MockedUseGetFlows);

      await render(
          <AuthenticationFlowSection
            application={mockApplication}
            editedApp={{auth_flow_id: 'auth-flow-2'}}
            onFieldChange={mockOnFieldChange}
          />
      );

      const input = page.getByPlaceholder('Select an authentication flow');
      expect(input).toHaveValue('Custom Auth Flow');
    });

    it('should handle flow selection', async () => {
            vi.mocked(useGetFlows).mockReturnValue({
        data: {flows: mockAuthFlows},
        isLoading: false,
      } as MockedUseGetFlows);

      await render(
          <AuthenticationFlowSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />
      );

      const input = page.getByPlaceholder('Select an authentication flow');
      await userEvent.click(input);

      await vi.waitFor(async () => {
        await expect.element(page.getByText('MFA Auth Flow')).toBeInTheDocument();
      });

      await userEvent.click(page.getByText('MFA Auth Flow'));

      expect(mockOnFieldChange).toHaveBeenCalledWith('auth_flow_id', 'auth-flow-3');
    });

    it('should handle clearing selection', async () => {
      vi.mocked(useGetFlows).mockReturnValue({
        data: {flows: mockAuthFlows},
        isLoading: false,
      } as MockedUseGetFlows);

      await render(
        <AuthenticationFlowSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />,
      );

      // Hover over the autocomplete to reveal the clear button
      const input = page.getByPlaceholder('Select an authentication flow');
      await userEvent.hover(input);

      const clearButton = page.getByRole('button', {name: 'Clear'});
      await userEvent.click(clearButton);

      expect(mockOnFieldChange).toHaveBeenCalledWith('auth_flow_id', '');
    });
  });

  describe('Flow Options Display', () => {
    it('should display flow name and handle in options', async () => {
            vi.mocked(useGetFlows).mockReturnValue({
        data: {flows: mockAuthFlows},
        isLoading: false,
      } as MockedUseGetFlows);

      await render(
          <AuthenticationFlowSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />
      );

      const input = page.getByPlaceholder('Select an authentication flow');
      await userEvent.click(input);

      await vi.waitFor(async () => {
        await expect.element(page.getByText('Custom Auth Flow')).toBeInTheDocument();
        await expect.element(page.getByText('custom-auth')).toBeInTheDocument();
      });
    });

    it('should display all available flows in dropdown', async () => {
            vi.mocked(useGetFlows).mockReturnValue({
        data: {flows: mockAuthFlows},
        isLoading: false,
      } as MockedUseGetFlows);

      await render(
          <AuthenticationFlowSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />
      );

      const input = page.getByPlaceholder('Select an authentication flow');
      await userEvent.click(input);

      await vi.waitFor(async () => {
        await expect.element(page.getByText('Default Auth Flow')).toBeInTheDocument();
        await expect.element(page.getByText('Custom Auth Flow')).toBeInTheDocument();
        await expect.element(page.getByText('MFA Auth Flow')).toBeInTheDocument();
      });
    });
  });

  describe('Empty State', () => {
    it('should handle empty flows array', async () => {
      vi.mocked(useGetFlows).mockReturnValue({
        data: {flows: []},
        isLoading: false,
      } as unknown as MockedUseGetFlows);

      await render(
          <AuthenticationFlowSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />
      );

      await expect.element(page.getByPlaceholder('Select an authentication flow')).toBeInTheDocument();
    });

    it('should handle undefined flows data', async () => {
      vi.mocked(useGetFlows).mockReturnValue({
        data: undefined,
        isLoading: false,
      } as MockedUseGetFlows);

      await render(
          <AuthenticationFlowSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />
      );

      await expect.element(page.getByPlaceholder('Select an authentication flow')).toBeInTheDocument();
    });
  });

  describe('Alert Links', () => {
    it('should display edit link with correct flow ID from application', async () => {
      vi.mocked(useGetFlows).mockReturnValue({
        data: {flows: mockAuthFlows},
        isLoading: false,
      } as MockedUseGetFlows);

      const {container} = await render(
          <AuthenticationFlowSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />
      );

      const links = container.querySelectorAll('a');
      const editLink = Array.from(links).find((link: Element) => link.getAttribute('href')?.includes('/flows/signin/'));
      expect(editLink).toHaveAttribute('href', '/flows/signin/auth-flow-1');
    });

    it('should display edit link with correct flow ID from editedApp', async () => {
      vi.mocked(useGetFlows).mockReturnValue({
        data: {flows: mockAuthFlows},
        isLoading: false,
      } as MockedUseGetFlows);

      const {container} = await render(
          <AuthenticationFlowSection
            application={mockApplication}
            editedApp={{auth_flow_id: 'auth-flow-2'}}
            onFieldChange={mockOnFieldChange}
          />
      );

      const links = container.querySelectorAll('a');
      const editLink = Array.from(links).find((link: Element) => link.getAttribute('href')?.includes('/flows/signin/'));
      expect(editLink).toHaveAttribute('href', '/flows/signin/auth-flow-2');
    });

    it('should display create link', async () => {
      vi.mocked(useGetFlows).mockReturnValue({
        data: {flows: mockAuthFlows},
        isLoading: false,
      } as MockedUseGetFlows);

      const {container} = await render(
          <AuthenticationFlowSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />
      );

      const links = container.querySelectorAll('a');
      const createLink = Array.from(links).find((link: Element) => link.getAttribute('href') === '/flows');
      expect(createLink).toHaveAttribute('href', '/flows');
    });
  });
});
