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

import {beforeEach, describe, expect, it, vi} from 'vitest';
import {render, screen} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import ConfigureStack from '../ConfigureStack';
import ApplicationCreateContext, {
  type ApplicationCreateContextType,
} from '../../../contexts/ApplicationCreate/ApplicationCreateContext';
import {PlatformApplicationTemplate, TechnologyApplicationTemplate} from '../../../models/application-templates';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
}));

const renderWithContext = (
  props: Parameters<typeof ConfigureStack>[0],
  contextOverrides: Partial<ApplicationCreateContextType> = {},
) => {
  const baseContext: ApplicationCreateContextType = {
    currentStep: null as unknown as ApplicationCreateContextType['currentStep'],
    setCurrentStep: vi.fn(),
    appName: 'Test App',
    setAppName: vi.fn(),
    selectedColor: '#ffffff',
    setSelectedColor: vi.fn(),
    appLogo: null,
    setAppLogo: vi.fn(),
    integrations: {},
    setIntegrations: vi.fn(),
    toggleIntegration: vi.fn(),
    signInApproach: null as unknown as ApplicationCreateContextType['signInApproach'],
    setSignInApproach: vi.fn(),
    selectedTechnology: null,
    setSelectedTechnology: vi.fn(),
    selectedPlatform: null,
    setSelectedPlatform: vi.fn(),
    selectedTemplateConfig: null,
    setSelectedTemplateConfig: vi.fn(),
    hostingUrl: '',
    setHostingUrl: vi.fn(),
    callbackUrlFromConfig: '',
    setCallbackUrlFromConfig: vi.fn(),
    error: null,
    setError: vi.fn(),
    reset: vi.fn(),
    ...contextOverrides,
  };

  return render(
    <ApplicationCreateContext.Provider value={baseContext}>
      <ConfigureStack {...props} />
    </ApplicationCreateContext.Provider>,
  );
};

describe('ConfigureStack', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders technology and platform sections', () => {
    renderWithContext({oauthConfig: null, onOAuthConfigChange: vi.fn(), onReadyChange: vi.fn()});

    expect(screen.getByText('applications:onboarding.configure.stack.technology.title')).toBeInTheDocument();
    expect(screen.getByText('applications:onboarding.configure.stack.platform.title')).toBeInTheDocument();
  });

  it('calls setSelectedTechnology when a technology card is clicked', async () => {
    const user = userEvent.setup();
    const setSelectedTechnology = vi.fn();

    renderWithContext(
      {oauthConfig: null, onOAuthConfigChange: vi.fn(), onReadyChange: vi.fn()},
      {setSelectedTechnology},
    );

    await user.click(screen.getByText('applications:onboarding.configure.stack.technology.nextjs.title'));

    expect(setSelectedTechnology).toHaveBeenCalledWith(TechnologyApplicationTemplate.NEXTJS);
  });

  it('calls setSelectedPlatform when a platform card is clicked', async () => {
    const user = userEvent.setup();
    const setSelectedPlatform = vi.fn();

    renderWithContext({oauthConfig: null, onOAuthConfigChange: vi.fn(), onReadyChange: vi.fn()}, {setSelectedPlatform});

    await user.click(screen.getByText('applications:onboarding.configure.stack.platform.browser.title'));

    expect(setSelectedPlatform).toHaveBeenCalledWith(PlatformApplicationTemplate.BROWSER);
  });

  it('syncs the OAuth configuration on mount', () => {
    const setSelectedTemplateConfig = vi.fn();
    const mockOnOAuthConfigChange = vi.fn();

    renderWithContext(
      {oauthConfig: null, onOAuthConfigChange: mockOnOAuthConfigChange, onReadyChange: vi.fn()},
      {setSelectedTemplateConfig},
    );

    expect(setSelectedTemplateConfig).toHaveBeenCalled();
    expect(mockOnOAuthConfigChange).toHaveBeenCalledWith(
      expect.objectContaining({scopes: ['openid', 'profile', 'email']}),
    );
  });

  it('shows only technology section when stackTypes.technology is true', () => {
    renderWithContext({
      oauthConfig: null,
      onOAuthConfigChange: vi.fn(),
      onReadyChange: vi.fn(),
      stackTypes: {technology: true, platform: false},
    });

    expect(screen.getByText('applications:onboarding.configure.stack.technology.title')).toBeInTheDocument();
    expect(screen.queryByText('applications:onboarding.configure.stack.platform.title')).not.toBeInTheDocument();
  });

  it('shows only platform section when stackTypes.platform is true', () => {
    renderWithContext({
      oauthConfig: null,
      onOAuthConfigChange: vi.fn(),
      onReadyChange: vi.fn(),
      stackTypes: {technology: false, platform: true},
    });

    expect(screen.queryByText('applications:onboarding.configure.stack.technology.title')).not.toBeInTheDocument();
    expect(screen.getByText('applications:onboarding.configure.stack.platform.title')).toBeInTheDocument();
  });

  it('updates template config when technology selection changes', async () => {
    const user = userEvent.setup();
    const setSelectedTemplateConfig = vi.fn();
    const mockOnOAuthConfigChange = vi.fn();

    renderWithContext(
      {
        oauthConfig: null,
        onOAuthConfigChange: mockOnOAuthConfigChange,
        onReadyChange: vi.fn(),
      },
      {setSelectedTemplateConfig},
    );

    await user.click(screen.getByText('applications:onboarding.configure.stack.technology.react.title'));

    expect(setSelectedTemplateConfig).toHaveBeenCalledWith(
      expect.objectContaining({
        name: 'React Application',
      }),
    );
  });

  it('updates template config when platform selection changes', async () => {
    const user = userEvent.setup();
    const setSelectedTemplateConfig = vi.fn();
    const setSelectedTechnology = vi.fn();
    const setSelectedPlatform = vi.fn();
    const mockOnOAuthConfigChange = vi.fn();

    renderWithContext(
      {
        oauthConfig: null,
        onOAuthConfigChange: mockOnOAuthConfigChange,
        onReadyChange: vi.fn(),
      },
      {setSelectedTemplateConfig, setSelectedTechnology, setSelectedPlatform},
    );

    await user.click(screen.getByText('applications:onboarding.configure.stack.platform.mobile.title'));

    expect(setSelectedTechnology).toHaveBeenCalledWith(null);
    expect(setSelectedPlatform).toHaveBeenCalledWith(PlatformApplicationTemplate.MOBILE);
  });

  it('highlights selected technology card', () => {
    renderWithContext(
      {oauthConfig: null, onOAuthConfigChange: vi.fn(), onReadyChange: vi.fn()},
      {selectedTechnology: TechnologyApplicationTemplate.REACT},
    );

    const reactCard = screen.getByText('applications:onboarding.configure.stack.technology.react.title');
    expect(reactCard).toBeInTheDocument();
  });

  it('highlights selected platform card', () => {
    renderWithContext(
      {oauthConfig: null, onOAuthConfigChange: vi.fn(), onReadyChange: vi.fn()},
      {selectedPlatform: PlatformApplicationTemplate.BROWSER},
    );

    const browserCard = screen.getByText('applications:onboarding.configure.stack.platform.browser.title');
    expect(browserCard).toBeInTheDocument();
  });

  it('calls onReadyChange based on selection state', async () => {
    const onReadyChange = vi.fn();

    renderWithContext({
      oauthConfig: null,
      onOAuthConfigChange: vi.fn(),
      onReadyChange,
    });

    // Should be ready when no selection is required (both sections shown)
    expect(onReadyChange).toHaveBeenCalledWith(true);
  });

  it('calls onReadyChange false when technology is OTHER but platform not selected', () => {
    const onReadyChange = vi.fn();

    renderWithContext(
      {
        oauthConfig: null,
        onOAuthConfigChange: vi.fn(),
        onReadyChange,
      },
      {selectedTechnology: TechnologyApplicationTemplate.OTHER, selectedPlatform: null},
    );

    expect(onReadyChange).toHaveBeenCalledWith(false);
  });

  it('calls onReadyChange true when required technology is selected', () => {
    const onReadyChange = vi.fn();

    renderWithContext(
      {
        oauthConfig: null,
        onOAuthConfigChange: vi.fn(),
        onReadyChange,
        stackTypes: {technology: true, platform: false},
      },
      {selectedTechnology: TechnologyApplicationTemplate.REACT},
    );

    expect(onReadyChange).toHaveBeenCalledWith(true);
  });

  it('calls onReadyChange false when platform is required but not selected', () => {
    const onReadyChange = vi.fn();

    renderWithContext(
      {
        oauthConfig: null,
        onOAuthConfigChange: vi.fn(),
        onReadyChange,
        stackTypes: {technology: false, platform: true},
      },
      {selectedPlatform: null},
    );

    expect(onReadyChange).toHaveBeenCalledWith(false);
  });

  it('renders all technology options', () => {
    renderWithContext({
      oauthConfig: null,
      onOAuthConfigChange: vi.fn(),
      onReadyChange: vi.fn(),
    });

    expect(screen.getByText('applications:onboarding.configure.stack.technology.react.title')).toBeInTheDocument();
    expect(screen.getByText('applications:onboarding.configure.stack.technology.nextjs.title')).toBeInTheDocument();
  });

  it('renders all platform options', () => {
    renderWithContext({
      oauthConfig: null,
      onOAuthConfigChange: vi.fn(),
      onReadyChange: vi.fn(),
    });

    expect(screen.getByText('applications:onboarding.configure.stack.platform.browser.title')).toBeInTheDocument();
    expect(screen.getByText('applications:onboarding.configure.stack.platform.server.title')).toBeInTheDocument();
    expect(screen.getByText('applications:onboarding.configure.stack.platform.mobile.title')).toBeInTheDocument();
    expect(screen.getByText('applications:onboarding.configure.stack.platform.backend.title')).toBeInTheDocument();
  });

  it('shows divider when both technology and platform sections are visible', () => {
    renderWithContext({
      oauthConfig: null,
      onOAuthConfigChange: vi.fn(),
      onReadyChange: vi.fn(),
    });

    expect(screen.getByText('applications:onboarding.configure.stack.dividerLabel')).toBeInTheDocument();
  });

  it('does not show divider when only one section is visible', () => {
    renderWithContext({
      oauthConfig: null,
      onOAuthConfigChange: vi.fn(),
      onReadyChange: vi.fn(),
      stackTypes: {technology: true, platform: false},
    });

    expect(screen.queryByText('applications:onboarding.configure.stack.dividerLabel')).not.toBeInTheDocument();
  });
});
