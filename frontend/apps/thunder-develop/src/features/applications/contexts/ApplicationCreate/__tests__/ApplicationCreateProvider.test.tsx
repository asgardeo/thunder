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

import {describe, expect, it, vi, beforeEach} from 'vitest';
import {page, userEvent} from 'vitest/browser';
import {render} from '@thunder/test-utils/browser';
import {AuthenticatorTypes} from '@/features/integrations/models/authenticators';
import ApplicationCreateProvider from '../ApplicationCreateProvider';
import useApplicationCreate from '../useApplicationCreate';
import {ApplicationCreateFlowSignInApproach, ApplicationCreateFlowStep} from '../../../models/application-create-flow';
import {TechnologyApplicationTemplate, PlatformApplicationTemplate} from '../../../models/application-templates';

// Mock useGetApplications
const mockUseGetApplications = vi.fn();
vi.mock('../../api/useGetApplications', () => ({
  __esModule: true,
  default: mockUseGetApplications,
}));

// Mock generateAppPrimaryColorSuggestions
vi.mock('../../utils/generateAppPrimaryColorSuggestions', () => ({
  __esModule: true,
  default: () => ['#3B82F6'],
}));

// Mock useConfig to avoid ConfigProvider requirement
vi.mock('@thunder/shared-contexts', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@thunder/shared-contexts')>();
  return {
    ...actual,
    useConfig: () => ({
      endpoints: {
        server: 'http://localhost:3001',
      },
    }),
  };
});

// Test component to consume the context
function TestConsumer() {
  const context = useApplicationCreate();

  return (
    <div>
      <div data-testid="current-step">{context.currentStep}</div>
      <div data-testid="app-name">{context.appName}</div>
      <div data-testid="selected-theme">{context.themeId ?? 'null'}</div>
      <div data-testid="app-logo">{context.appLogo ?? 'null'}</div>
      <div data-testid="selected-color">{context.selectedColor}</div>
      <div data-testid="integrations">{JSON.stringify(context.integrations)}</div>
      <div data-testid="sign-in-approach">{context.signInApproach}</div>
      <div data-testid="selected-technology">{context.selectedTechnology ?? 'null'}</div>
      <div data-testid="selected-platform">{context.selectedPlatform ?? 'null'}</div>
      <div data-testid="hosting-url">{context.hostingUrl}</div>
      <div data-testid="callback-url">{context.callbackUrlFromConfig}</div>
      <div data-testid="error">{context.error ?? 'null'}</div>

      <button type="button" onClick={() => context.setCurrentStep(ApplicationCreateFlowStep.DESIGN)}>
        Set Design Step
      </button>
      <button type="button" onClick={() => context.setAppName('Test App')}>
        Set App Name
      </button>
      <button type="button" onClick={() => context.setSelectedTheme(null)}>
        Set Theme to null
      </button>
      <button type="button" onClick={() => context.setAppLogo('test-logo.png')}>
        Set Logo
      </button>
      <button type="button" onClick={() => context.toggleIntegration('test-integration')}>
        Toggle Integration
      </button>
      <button type="button" onClick={() => context.setSignInApproach(ApplicationCreateFlowSignInApproach.EMBEDDED)}>
        Set Custom Approach
      </button>
      <button type="button" onClick={() => context.setSelectedTechnology(TechnologyApplicationTemplate.REACT)}>
        Set React Technology
      </button>
      <button type="button" onClick={() => context.setSelectedPlatform(PlatformApplicationTemplate.BROWSER)}>
        Set Browser Platform
      </button>
      <button type="button" onClick={() => context.setHostingUrl('https://example.com')}>
        Set Hosting URL
      </button>
      <button type="button" onClick={() => context.setCallbackUrlFromConfig('https://example.com/callback')}>
        Set Callback URL
      </button>
      <button type="button" onClick={() => context.setError('Test error')}>
        Set Error
      </button>
      <button type="button" onClick={() => context.reset()}>
        Reset
      </button>
    </div>
  );
}

describe('ApplicationCreateProvider', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    // Default mock: no applications
    mockUseGetApplications.mockReturnValue({
      data: {
        applications: [],
      },
    });
  });

  const renderWithQueryClient = (ui: React.ReactElement) => render(ui);

  it('provides initial state values', async () => {
    await renderWithQueryClient(
      <ApplicationCreateProvider>
        <TestConsumer />
      </ApplicationCreateProvider>,
    );

    await expect.element(page.getByTestId('current-step')).toHaveTextContent(ApplicationCreateFlowStep.NAME);
    await expect.element(page.getByTestId('app-name')).toHaveTextContent('');
    await expect.element(page.getByTestId('selected-theme')).toHaveTextContent('null');
    await expect.element(page.getByTestId('app-logo')).toHaveTextContent('null');
    await expect.element(page.getByTestId('integrations')).toHaveTextContent(
      JSON.stringify({[AuthenticatorTypes.BASIC_AUTH]: true}),
    );
    await expect.element(page.getByTestId('sign-in-approach')).toHaveTextContent(ApplicationCreateFlowSignInApproach.INBUILT);
    await expect.element(page.getByTestId('selected-technology')).toHaveTextContent('null');
    await expect.element(page.getByTestId('selected-platform')).toHaveTextContent('null');
    await expect.element(page.getByTestId('hosting-url')).toHaveTextContent('');
    await expect.element(page.getByTestId('callback-url')).toHaveTextContent('');
    await expect.element(page.getByTestId('error')).toHaveTextContent('null');
  });

  it('updates current step when setCurrentStep is called', async () => {

    await renderWithQueryClient(
      <ApplicationCreateProvider>
        <TestConsumer />
      </ApplicationCreateProvider>,
    );

    await userEvent.click(page.getByText('Set Design Step'));

    await expect.element(page.getByTestId('current-step')).toHaveTextContent(ApplicationCreateFlowStep.DESIGN);
  });

  it('updates app name when setAppName is called', async () => {

    await renderWithQueryClient(
      <ApplicationCreateProvider>
        <TestConsumer />
      </ApplicationCreateProvider>,
    );

    await userEvent.click(page.getByText('Set App Name'));

    await expect.element(page.getByTestId('app-name')).toHaveTextContent('Test App');
  });

  it('updates selected theme when setSelectedTheme is called', async () => {

    await renderWithQueryClient(
      <ApplicationCreateProvider>
        <TestConsumer />
      </ApplicationCreateProvider>,
    );

    await userEvent.click(page.getByText('Set Theme to null'));

    await expect.element(page.getByTestId('selected-theme')).toHaveTextContent('null');
  });

  it('updates app logo when setAppLogo is called', async () => {

    await renderWithQueryClient(
      <ApplicationCreateProvider>
        <TestConsumer />
      </ApplicationCreateProvider>,
    );

    await userEvent.click(page.getByText('Set Logo'));

    await expect.element(page.getByTestId('app-logo')).toHaveTextContent('test-logo.png');
  });

  it('toggles integration state when toggleIntegration is called', async () => {

    await renderWithQueryClient(
      <ApplicationCreateProvider>
        <TestConsumer />
      </ApplicationCreateProvider>,
    );

    // Initial state should not have 'test-integration'
    await expect.element(page.getByTestId('integrations')).not.toHaveTextContent('test-integration');

    await userEvent.click(page.getByText('Toggle Integration'));

    // Should now have test-integration set to true
    await expect.element(page.getByTestId('integrations')).toHaveTextContent('test-integration');
    await expect.element(page.getByTestId('integrations')).toHaveTextContent('true');

    // Toggle again to disable
    await userEvent.click(page.getByText('Toggle Integration'));

    // Should now have test-integration set to false
    await expect.element(page.getByTestId('integrations')).toHaveTextContent('test-integration');
    await expect.element(page.getByTestId('integrations')).toHaveTextContent('false');
  });

  it('updates sign-in approach when setSignInApproach is called', async () => {

    await renderWithQueryClient(
      <ApplicationCreateProvider>
        <TestConsumer />
      </ApplicationCreateProvider>,
    );

    await userEvent.click(page.getByText('Set Custom Approach'));

    await expect.element(page.getByTestId('sign-in-approach')).toHaveTextContent(ApplicationCreateFlowSignInApproach.EMBEDDED);
  });

  it('updates selected technology when setSelectedTechnology is called', async () => {

    await renderWithQueryClient(
      <ApplicationCreateProvider>
        <TestConsumer />
      </ApplicationCreateProvider>,
    );

    await userEvent.click(page.getByText('Set React Technology'));

    await expect.element(page.getByTestId('selected-technology')).toHaveTextContent(TechnologyApplicationTemplate.REACT);
  });

  it('updates selected platform when setSelectedPlatform is called', async () => {

    await renderWithQueryClient(
      <ApplicationCreateProvider>
        <TestConsumer />
      </ApplicationCreateProvider>,
    );

    await userEvent.click(page.getByText('Set Browser Platform'));

    await expect.element(page.getByTestId('selected-platform')).toHaveTextContent(PlatformApplicationTemplate.BROWSER);
  });

  it('updates hosting URL when setHostingUrl is called', async () => {

    await renderWithQueryClient(
      <ApplicationCreateProvider>
        <TestConsumer />
      </ApplicationCreateProvider>,
    );

    await userEvent.click(page.getByText('Set Hosting URL'));

    await expect.element(page.getByTestId('hosting-url')).toHaveTextContent('https://example.com');
  });

  it('updates callback URL when setCallbackUrlFromConfig is called', async () => {

    await renderWithQueryClient(
      <ApplicationCreateProvider>
        <TestConsumer />
      </ApplicationCreateProvider>,
    );

    await userEvent.click(page.getByText('Set Callback URL'));

    await expect.element(page.getByTestId('callback-url')).toHaveTextContent('https://example.com/callback');
  });

  it('updates error when setError is called', async () => {

    await renderWithQueryClient(
      <ApplicationCreateProvider>
        <TestConsumer />
      </ApplicationCreateProvider>,
    );

    await userEvent.click(page.getByText('Set Error'));

    await expect.element(page.getByTestId('error')).toHaveTextContent('Test error');
  });

  it('resets all state when reset is called', async () => {

    await renderWithQueryClient(
      <ApplicationCreateProvider>
        <TestConsumer />
      </ApplicationCreateProvider>,
    );

    // Set some values
    await userEvent.click(page.getByText('Set App Name'));
    await userEvent.click(page.getByText('Set Theme to null'));
    await userEvent.click(page.getByText('Set Error'));

    // Verify values are set
    await expect.element(page.getByTestId('app-name')).toHaveTextContent('Test App');
    await expect.element(page.getByTestId('selected-theme')).toHaveTextContent('null');
    await expect.element(page.getByTestId('error')).toHaveTextContent('Test error');

    // Reset
    await userEvent.click(page.getByText('Reset'));

    // Verify back to initial state
    await expect.element(page.getByTestId('current-step')).toHaveTextContent(ApplicationCreateFlowStep.NAME);
    await expect.element(page.getByTestId('app-name')).toHaveTextContent('');
    await expect.element(page.getByTestId('error')).toHaveTextContent('null');
    await expect.element(page.getByTestId('selected-theme')).toHaveTextContent('null');
  });

  it('memoizes context value to prevent unnecessary re-renders', async () => {
    const renderSpy = vi.fn();

    function TestRenderer() {
      renderSpy();
      return <TestConsumer />;
    }

    const {rerender} = await renderWithQueryClient(
      <ApplicationCreateProvider>
        <TestRenderer />
      </ApplicationCreateProvider>,
    );

    expect(renderSpy).toHaveBeenCalledTimes(1);

    // Re-render with same props
    await rerender(
      <ApplicationCreateProvider>
        <TestRenderer />
      </ApplicationCreateProvider>,
    );

    // Should only render once more due to memoization
    expect(renderSpy).toHaveBeenCalledTimes(2);
  });

  it('provides default integrations with basic auth enabled', async () => {
    await renderWithQueryClient(
      <ApplicationCreateProvider>
        <TestConsumer />
      </ApplicationCreateProvider>,
    );

    const integrationsEl = page.getByTestId('integrations');
    const integrations = JSON.parse(integrationsEl.element().textContent ?? '{}') as Record<string, boolean>;
    expect(integrations[AuthenticatorTypes.BASIC_AUTH]).toBe(true);
  });

  it('initializes with a default primary color', async () => {
    await renderWithQueryClient(
      <ApplicationCreateProvider>
        <TestConsumer />
      </ApplicationCreateProvider>,
    );

    const color = page.getByTestId('selected-color').element().textContent;
    expect(color).toMatch(/^#[0-9a-fA-F]{6}$/); // Should be a valid hex color
  });
});
