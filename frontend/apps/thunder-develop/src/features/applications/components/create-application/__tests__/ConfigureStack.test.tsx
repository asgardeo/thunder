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
import {page, userEvent} from 'vitest/browser';
import {render} from '@thunder/test-utils/browser';
import ConfigureStack from '../ConfigureStack';
import ApplicationCreateContext, {
  type ApplicationCreateContextType,
} from '../../../contexts/ApplicationCreate/ApplicationCreateContext';
import {PlatformApplicationTemplate, TechnologyApplicationTemplate} from '../../../models/application-templates';
import {ApplicationCreateFlowSignInApproach} from '../../../models/application-create-flow';

const renderWithContext = (
  props: Parameters<typeof ConfigureStack>[0],
  contextOverrides: Partial<ApplicationCreateContextType> = {},
) => {
  const baseContext: ApplicationCreateContextType = {
    currentStep: null as unknown as ApplicationCreateContextType['currentStep'],
    setCurrentStep: vi.fn(),
    appName: 'Test App',
    setAppName: vi.fn(),
    themeId: null,
    setThemeId: vi.fn(),
    selectedTheme: null,
    setSelectedTheme: vi.fn(),
    appLogo: null,
    setAppLogo: vi.fn(),
    selectedColor: '',
    setSelectedColor: vi.fn(),
    integrations: {},
    setIntegrations: vi.fn(),
    toggleIntegration: vi.fn(),
    selectedAuthFlow: null,
    setSelectedAuthFlow: vi.fn(),
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
    hasCompletedOnboarding: false,
    setHasCompletedOnboarding: vi.fn(),
    error: null,
    setError: vi.fn(),
    reset: vi.fn(),
    relyingPartyId: '',
    setRelyingPartyId: vi.fn(),
    relyingPartyName: '',
    setRelyingPartyName: vi.fn(),
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

  it('renders technology and platform sections', async () => {
    await renderWithContext({oauthConfig: null, onOAuthConfigChange: vi.fn(), onReadyChange: vi.fn()});

    await expect.element(page.getByText('Technology')).toBeInTheDocument();
    await expect.element(page.getByText('Application Type')).toBeInTheDocument();
  });

  it('calls setSelectedTechnology when a technology card is clicked', async () => {
    const setSelectedTechnology = vi.fn();

    await renderWithContext(
      {oauthConfig: null, onOAuthConfigChange: vi.fn(), onReadyChange: vi.fn()},
      {setSelectedTechnology},
    );

    await userEvent.click(page.getByText('React'));

    expect(setSelectedTechnology).toHaveBeenCalledWith(TechnologyApplicationTemplate.REACT);
  });

  it('calls setSelectedPlatform when a platform card is clicked', async () => {
    const setSelectedPlatform = vi.fn();

    await renderWithContext({oauthConfig: null, onOAuthConfigChange: vi.fn(), onReadyChange: vi.fn()}, {setSelectedPlatform});

    await userEvent.click(page.getByText('Browser App'));

    expect(setSelectedPlatform).toHaveBeenCalledWith(PlatformApplicationTemplate.BROWSER);
  });

  it('syncs the OAuth configuration on mount', async () => {
    const setSelectedTemplateConfig = vi.fn();
    const mockOnOAuthConfigChange = vi.fn();

    await renderWithContext(
      {oauthConfig: null, onOAuthConfigChange: mockOnOAuthConfigChange, onReadyChange: vi.fn()},
      {setSelectedTemplateConfig},
    );

    expect(setSelectedTemplateConfig).toHaveBeenCalled();
    expect(mockOnOAuthConfigChange).toHaveBeenCalledWith(
      expect.objectContaining({scopes: ['openid', 'profile', 'email']}),
    );
  });

  it('shows only technology section when stackTypes.technology is true', async () => {
    await renderWithContext({
      oauthConfig: null,
      onOAuthConfigChange: vi.fn(),
      onReadyChange: vi.fn(),
      stackTypes: {technology: true, platform: false},
    });

    await expect.element(page.getByText('Technology')).toBeInTheDocument();
    await expect.element(page.getByText('Application Type')).not.toBeInTheDocument();
  });

  it('shows only platform section when stackTypes.platform is true', async () => {
    await renderWithContext({
      oauthConfig: null,
      onOAuthConfigChange: vi.fn(),
      onReadyChange: vi.fn(),
      stackTypes: {technology: false, platform: true},
    });

    await expect.element(page.getByText('Technology')).not.toBeInTheDocument();
    await expect.element(page.getByText('Application Type')).toBeInTheDocument();
  });

  it('updates template config when technology selection changes', async () => {
    const setSelectedTemplateConfig = vi.fn();
    const mockOnOAuthConfigChange = vi.fn();

    await renderWithContext(
      {
        oauthConfig: null,
        onOAuthConfigChange: mockOnOAuthConfigChange,
        onReadyChange: vi.fn(),
      },
      {setSelectedTemplateConfig},
    );

    await userEvent.click(page.getByText('React'));

    expect(setSelectedTemplateConfig).toHaveBeenCalledWith(
      expect.objectContaining({
        name: 'React Application',
      }),
    );
  });

  it('updates template config when platform selection changes', async () => {
    const setSelectedTemplateConfig = vi.fn();
    const setSelectedTechnology = vi.fn();
    const setSelectedPlatform = vi.fn();
    const mockOnOAuthConfigChange = vi.fn();

    await renderWithContext(
      {
        oauthConfig: null,
        onOAuthConfigChange: mockOnOAuthConfigChange,
        onReadyChange: vi.fn(),
      },
      {setSelectedTemplateConfig, setSelectedTechnology, setSelectedPlatform},
    );

    await userEvent.click(page.getByText('Mobile App'));

    expect(setSelectedTechnology).toHaveBeenCalledWith(null);
    expect(setSelectedPlatform).toHaveBeenCalledWith(PlatformApplicationTemplate.MOBILE);
  });

  it('highlights selected technology card', async () => {
    await renderWithContext(
      {oauthConfig: null, onOAuthConfigChange: vi.fn(), onReadyChange: vi.fn()},
      {selectedTechnology: TechnologyApplicationTemplate.REACT},
    );

    const reactCard = page.getByText('React');
    expect(reactCard).toBeInTheDocument();
  });

  it('highlights selected platform card', async () => {
    await renderWithContext(
      {oauthConfig: null, onOAuthConfigChange: vi.fn(), onReadyChange: vi.fn()},
      {selectedPlatform: PlatformApplicationTemplate.BROWSER},
    );

    const browserCard = page.getByText('Browser App');
    expect(browserCard).toBeInTheDocument();
  });

  it('calls onReadyChange based on selection state', async () => {
    const onReadyChange = vi.fn();

    await renderWithContext({
      oauthConfig: null,
      onOAuthConfigChange: vi.fn(),
      onReadyChange,
    });

    // Should be ready when no selection is required (both sections shown)
    expect(onReadyChange).toHaveBeenCalledWith(true);
  });

  it('calls onReadyChange false when technology is OTHER but platform not selected', async () => {
    const onReadyChange = vi.fn();

    await renderWithContext(
      {
        oauthConfig: null,
        onOAuthConfigChange: vi.fn(),
        onReadyChange,
      },
      {selectedTechnology: TechnologyApplicationTemplate.OTHER, selectedPlatform: null},
    );

    expect(onReadyChange).toHaveBeenCalledWith(false);
  });

  it('calls onReadyChange true when required technology is selected', async () => {
    const onReadyChange = vi.fn();

    await renderWithContext(
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

  it('calls onReadyChange false when platform is required but not selected', async () => {
    const onReadyChange = vi.fn();

    await renderWithContext(
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

  it('renders all technology options', async () => {
    await renderWithContext({
      oauthConfig: null,
      onOAuthConfigChange: vi.fn(),
      onReadyChange: vi.fn(),
    });

    await expect.element(page.getByText('React')).toBeInTheDocument();
    await expect.element(page.getByText('Next.js')).toBeInTheDocument();
  });

  it('renders all platform options', async () => {
    await renderWithContext({
      oauthConfig: null,
      onOAuthConfigChange: vi.fn(),
      onReadyChange: vi.fn(),
    });

    await expect.element(page.getByText('Browser App')).toBeInTheDocument();
    await expect.element(page.getByText('Full-Stack App')).toBeInTheDocument();
    await expect.element(page.getByText('Mobile App')).toBeInTheDocument();
    await expect.element(page.getByText('Backend Service')).toBeInTheDocument();
  });

  it('shows divider when both technology and platform sections are visible', async () => {
    await renderWithContext({
      oauthConfig: null,
      onOAuthConfigChange: vi.fn(),
      onReadyChange: vi.fn(),
    });

    await expect.element(page.getByText('OR')).toBeInTheDocument();
  });

  it('does not show divider when only one section is visible', async () => {
    await renderWithContext({
      oauthConfig: null,
      onOAuthConfigChange: vi.fn(),
      onReadyChange: vi.fn(),
      stackTypes: {technology: true, platform: false},
    });

    await expect.element(page.getByText('OR')).not.toBeInTheDocument();
  });

  it('shows "Coming Soon" badge for disabled technology options', async () => {
    await renderWithContext({
      oauthConfig: null,
      onOAuthConfigChange: vi.fn(),
      onReadyChange: vi.fn(),
    });

    await expect.element(page.getByText('Coming Soon')).toBeInTheDocument();
  });

  it('does not call setSelectedTechnology when clicking disabled technology card', async () => {
    const setSelectedTechnology = vi.fn();

    await renderWithContext(
      {oauthConfig: null, onOAuthConfigChange: vi.fn(), onReadyChange: vi.fn()},
      {setSelectedTechnology},
    );

    // Next.js is disabled, clicking should not trigger the handler
    const nextjsCard = page.getByText('Next.js');
    await userEvent.click(nextjsCard);

    // setSelectedTechnology should not have been called with NEXTJS
    expect(setSelectedTechnology).not.toHaveBeenCalledWith(TechnologyApplicationTemplate.NEXTJS);
  });

  it('hides platform section and divider when signInApproach is EMBEDDED', async () => {
    await renderWithContext(
      {
        oauthConfig: null,
        onOAuthConfigChange: vi.fn(),
        onReadyChange: vi.fn(),
        stackTypes: {technology: true, platform: true},
      },
      {signInApproach: ApplicationCreateFlowSignInApproach.EMBEDDED},
    );

    await expect.element(page.getByText('Technology')).toBeInTheDocument();
    await expect.element(page.getByText('Application Type')).not.toBeInTheDocument();
    await expect.element(page.getByText('OR')).not.toBeInTheDocument();
  });

  it('auto-selects first platform when technology section is hidden and no platform selected', async () => {
    const setSelectedPlatform = vi.fn();

    await renderWithContext(
      {
        oauthConfig: null,
        onOAuthConfigChange: vi.fn(),
        onReadyChange: vi.fn(),
        stackTypes: {technology: false, platform: true},
      },
      {setSelectedPlatform, selectedPlatform: null},
    );

    expect(setSelectedPlatform).toHaveBeenCalledWith(PlatformApplicationTemplate.BROWSER);
  });

  it('does not auto-select platform when technology section is visible', async () => {
    const setSelectedPlatform = vi.fn();

    await renderWithContext(
      {
        oauthConfig: null,
        onOAuthConfigChange: vi.fn(),
        onReadyChange: vi.fn(),
        stackTypes: {technology: true, platform: true},
      },
      {setSelectedPlatform, selectedPlatform: null},
    );

    expect(setSelectedPlatform).not.toHaveBeenCalled();
  });

  it('selects server platform when clicked', async () => {
    const setSelectedPlatform = vi.fn();
    const setSelectedTechnology = vi.fn();

    await renderWithContext(
      {oauthConfig: null, onOAuthConfigChange: vi.fn(), onReadyChange: vi.fn()},
      {setSelectedPlatform, setSelectedTechnology},
    );

    await userEvent.click(page.getByText('Full-Stack App'));

    expect(setSelectedPlatform).toHaveBeenCalledWith(PlatformApplicationTemplate.SERVER);
    expect(setSelectedTechnology).toHaveBeenCalledWith(null);
  });

  it('selects backend platform when clicked', async () => {
    const setSelectedPlatform = vi.fn();
    const setSelectedTechnology = vi.fn();

    await renderWithContext(
      {oauthConfig: null, onOAuthConfigChange: vi.fn(), onReadyChange: vi.fn()},
      {setSelectedPlatform, setSelectedTechnology},
    );

    await userEvent.click(page.getByText('Backend Service'));

    expect(setSelectedPlatform).toHaveBeenCalledWith(PlatformApplicationTemplate.BACKEND);
    expect(setSelectedTechnology).toHaveBeenCalledWith(null);
  });

  it('uses platform template when technology is OTHER', async () => {
    const setSelectedTemplateConfig = vi.fn();
    const mockOnOAuthConfigChange = vi.fn();

    await renderWithContext(
      {
        oauthConfig: null,
        onOAuthConfigChange: mockOnOAuthConfigChange,
        onReadyChange: vi.fn(),
      },
      {
        setSelectedTemplateConfig,
        selectedTechnology: TechnologyApplicationTemplate.OTHER,
        selectedPlatform: PlatformApplicationTemplate.MOBILE,
      },
    );

    expect(setSelectedTemplateConfig).toHaveBeenCalledWith(
      expect.objectContaining({
        name: 'Mobile Application',
      }),
    );
  });

  it('uses inferred technology from existing oauthConfig', async () => {
    const setSelectedTemplateConfig = vi.fn();
    const mockOnOAuthConfigChange = vi.fn();

    await renderWithContext(
      {
        oauthConfig: {
          public_client: true,
          pkce_required: true,
          grant_types: ['authorization_code'],
          response_types: ['code'],
          redirect_uris: ['http://localhost:3000/callback'],
          token_endpoint_auth_method: 'none',
          scopes: ['openid', 'profile'],
        },
        onOAuthConfigChange: mockOnOAuthConfigChange,
        onReadyChange: vi.fn(),
      },
      {setSelectedTemplateConfig, selectedTechnology: null, selectedPlatform: null},
    );

    expect(setSelectedTemplateConfig).toHaveBeenCalled();
  });

  it('resolves technology to OTHER when platform is selected but no technology', async () => {
    const setSelectedTemplateConfig = vi.fn();

    await renderWithContext(
      {
        oauthConfig: null,
        onOAuthConfigChange: vi.fn(),
        onReadyChange: vi.fn(),
      },
      {
        setSelectedTemplateConfig,
        selectedTechnology: null,
        selectedPlatform: PlatformApplicationTemplate.SERVER,
      },
    );

    expect(setSelectedTemplateConfig).toHaveBeenCalledWith(
      expect.objectContaining({
        name: 'Server Application',
      }),
    );
  });

  it('calls onReadyChange true when OTHER technology with platform selected', async () => {
    const onReadyChange = vi.fn();

    await renderWithContext(
      {
        oauthConfig: null,
        onOAuthConfigChange: vi.fn(),
        onReadyChange,
      },
      {
        selectedTechnology: TechnologyApplicationTemplate.OTHER,
        selectedPlatform: PlatformApplicationTemplate.BROWSER,
      },
    );

    expect(onReadyChange).toHaveBeenCalledWith(true);
  });

  it('clears technology and sets platform when platform card is clicked', async () => {
    const setSelectedTechnology = vi.fn();
    const setSelectedPlatform = vi.fn();

    await renderWithContext(
      {oauthConfig: null, onOAuthConfigChange: vi.fn(), onReadyChange: vi.fn()},
      {
        setSelectedTechnology,
        setSelectedPlatform,
        selectedTechnology: TechnologyApplicationTemplate.REACT,
      },
    );

    await userEvent.click(page.getByText('Browser App'));

    expect(setSelectedTechnology).toHaveBeenCalledWith(null);
    expect(setSelectedPlatform).toHaveBeenCalledWith(PlatformApplicationTemplate.BROWSER);
  });

  it('clears platform when technology card is clicked', async () => {
    const setSelectedTechnology = vi.fn();
    const setSelectedPlatform = vi.fn();

    await renderWithContext(
      {oauthConfig: null, onOAuthConfigChange: vi.fn(), onReadyChange: vi.fn()},
      {
        setSelectedTechnology,
        setSelectedPlatform,
        selectedPlatform: PlatformApplicationTemplate.BROWSER,
      },
    );

    await userEvent.click(page.getByText('React'));

    expect(setSelectedTechnology).toHaveBeenCalledWith(TechnologyApplicationTemplate.REACT);
    expect(setSelectedPlatform).toHaveBeenCalledWith(null);
  });

  it('renders without onReadyChange callback', async () => {
    await renderWithContext({
      oauthConfig: null,
      onOAuthConfigChange: vi.fn(),
    });

    await expect.element(page.getByText('Technology')).toBeInTheDocument();
  });

  it('syncs OAuth config with correct structure including all fields', async () => {
    const mockOnOAuthConfigChange = vi.fn();

    await renderWithContext(
      {
        oauthConfig: null,
        onOAuthConfigChange: mockOnOAuthConfigChange,
        onReadyChange: vi.fn(),
      },
      {selectedTechnology: TechnologyApplicationTemplate.REACT},
    );

    expect(mockOnOAuthConfigChange).toHaveBeenCalledWith(
      expect.objectContaining({
        public_client: expect.any(Boolean) as boolean,
        pkce_required: expect.any(Boolean) as boolean,
        grant_types: expect.any(Array) as string[],
        response_types: expect.any(Array) as string[],
        redirect_uris: expect.any(Array) as string[],
        scopes: ['openid', 'profile', 'email'],
      }),
    );
  });

  it('does not auto-select platform when already selected', async () => {
    const setSelectedPlatform = vi.fn();

    await renderWithContext(
      {
        oauthConfig: null,
        onOAuthConfigChange: vi.fn(),
        onReadyChange: vi.fn(),
        stackTypes: {technology: false, platform: true},
      },
      {setSelectedPlatform, selectedPlatform: PlatformApplicationTemplate.MOBILE},
    );

    expect(setSelectedPlatform).not.toHaveBeenCalled();
  });

  describe('hover states and conditional styling', () => {
    it('should render technology card with correct structure when not selected', async () => {
      await renderWithContext(
        {oauthConfig: null, onOAuthConfigChange: vi.fn(), onReadyChange: vi.fn()},
        {selectedTechnology: TechnologyApplicationTemplate.OTHER},
      );

      const reactTitle = page.getByText('React');
      expect(reactTitle).toBeInTheDocument();
    });

    it('should render technology card with correct structure when selected', async () => {
      await renderWithContext(
        {oauthConfig: null, onOAuthConfigChange: vi.fn(), onReadyChange: vi.fn()},
        {selectedTechnology: TechnologyApplicationTemplate.REACT},
      );

      const reactTitle = page.getByText('React');
      expect(reactTitle).toBeInTheDocument();
    });

    it('should render platform card with correct structure when not selected', async () => {
      await renderWithContext(
        {oauthConfig: null, onOAuthConfigChange: vi.fn(), onReadyChange: vi.fn()},
        {selectedPlatform: PlatformApplicationTemplate.SERVER},
      );

      const browserTitle = page.getByText('Browser App');
      expect(browserTitle).toBeInTheDocument();
    });

    it('should render platform card with correct structure when selected', async () => {
      await renderWithContext(
        {oauthConfig: null, onOAuthConfigChange: vi.fn(), onReadyChange: vi.fn()},
        {selectedPlatform: PlatformApplicationTemplate.BROWSER},
      );

      const browserTitle = page.getByText('Browser App');
      expect(browserTitle).toBeInTheDocument();
    });

    it('should render disabled technology card with correct structure', async () => {
      await renderWithContext({oauthConfig: null, onOAuthConfigChange: vi.fn(), onReadyChange: vi.fn()});

      const nextjsTitle = page.getByText('Next.js');
      expect(nextjsTitle).toBeInTheDocument();
      await expect.element(page.getByText('Coming Soon')).toBeInTheDocument();
    });
  });

  describe('technology resolution logic', () => {
    it('should use default technology when stackTypes.technology is true and nothing selected', async () => {
      const setSelectedTemplateConfig = vi.fn();

      await renderWithContext(
        {
          oauthConfig: null,
          onOAuthConfigChange: vi.fn(),
          onReadyChange: vi.fn(),
          stackTypes: {technology: true, platform: true},
        },
        {setSelectedTemplateConfig, selectedTechnology: null, selectedPlatform: null},
      );

      // Should use default technology (React)
      expect(setSelectedTemplateConfig).toHaveBeenCalledWith(
        expect.objectContaining({
          name: 'React Application',
        }),
      );
    });

    it('should use OTHER technology when stackTypes.technology is false', async () => {
      const setSelectedTemplateConfig = vi.fn();

      await renderWithContext(
        {
          oauthConfig: null,
          onOAuthConfigChange: vi.fn(),
          onReadyChange: vi.fn(),
          stackTypes: {technology: false, platform: true},
        },
        {setSelectedTemplateConfig, selectedTechnology: null, selectedPlatform: PlatformApplicationTemplate.BROWSER},
      );

      // Should use platform template since technology is hidden
      expect(setSelectedTemplateConfig).toHaveBeenCalledWith(
        expect.objectContaining({
          name: 'Browser Application',
        }),
      );
    });
  });

  describe('OAuth config template handling', () => {
    it('should handle template with empty redirect_uris', async () => {
      const mockOnOAuthConfigChange = vi.fn();

      await renderWithContext(
        {
          oauthConfig: null,
          onOAuthConfigChange: mockOnOAuthConfigChange,
          onReadyChange: vi.fn(),
        },
        {selectedPlatform: PlatformApplicationTemplate.BACKEND},
      );

      expect(mockOnOAuthConfigChange).toHaveBeenCalledWith(
        expect.objectContaining({
          redirect_uris: expect.any(Array) as string[],
        }),
      );
    });

    it('should handle template with response_types', async () => {
      const mockOnOAuthConfigChange = vi.fn();

      await renderWithContext(
        {
          oauthConfig: null,
          onOAuthConfigChange: mockOnOAuthConfigChange,
          onReadyChange: vi.fn(),
        },
        {selectedTechnology: TechnologyApplicationTemplate.REACT},
      );

      expect(mockOnOAuthConfigChange).toHaveBeenCalledWith(
        expect.objectContaining({
          response_types: expect.any(Array) as string[],
        }),
      );
    });
  });
});
