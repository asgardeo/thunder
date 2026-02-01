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

import {render, screen, waitFor} from '@thunder/test-utils';
import {describe, it, expect, vi, beforeEach} from 'vitest';
import userEvent from '@testing-library/user-event';
import type {Branding} from '@thunder/shared-branding';
import type {Application} from '../../models/application';
import ApplicationCreatePage from '../ApplicationCreatePage';
import ApplicationCreateProvider from '../../contexts/ApplicationCreate/ApplicationCreateProvider';

// Mock functions
const mockCreateApplication = vi.fn();
const mockCreateBranding = vi.fn();
const mockNavigate = vi.fn();

// Mock logger
vi.mock('@thunder/logger/react', () => ({
  useLogger: () => ({
    info: vi.fn(),
    warn: vi.fn(),
    error: vi.fn(),
    debug: vi.fn(),
    withComponent: vi.fn().mockReturnThis(),
  }),
}));

// Mock react-router
vi.mock('react-router', async () => {
  const actual = await vi.importActual('react-router');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

// Mock branding hooks
vi.mock('@thunder/shared-branding', () => ({
  useCreateBranding: () => ({
    mutate: mockCreateBranding,
    isPending: false,
  }),
  useGetBrandings: () => ({
    data: {brandings: []},
    isLoading: false,
  }),
  useGetBranding: () => ({
    data: null,
    isLoading: false,
  }),
  LayoutType: {
    CENTERED: 'centered',
  },
}));

// Mock application API
vi.mock('../../api/useCreateApplication', () => ({
  default: () => ({
    mutate: mockCreateApplication,
    isPending: false,
  }),
}));

// Mock user types API
vi.mock('../../../user-types/api/useGetUserTypes', () => ({
  default: () => ({
    data: {
      schemas: [
        {name: 'customer', displayName: 'Customer'},
        {name: 'employee', displayName: 'Employee'},
      ],
    },
    isLoading: false,
    error: null,
  }),
}));

// Mock integrations API
vi.mock('../../../integrations/api/useIdentityProviders', () => ({
  default: () => ({
    data: [
      {id: 'google', name: 'Google', type: 'social'},
      {id: 'github', name: 'GitHub', type: 'social'},
    ],
    isLoading: false,
    error: null,
  }),
}));

// Mock flows API
vi.mock('../../../flows/api/useGetFlows', () => ({
  default: () => ({
    data: {
      flows: [
        {id: 'flow1', name: 'Basic Auth Flow', handle: 'basic-auth'},
        {id: 'flow2', name: 'Google Flow', handle: 'google-flow'},
      ],
    },
    isLoading: false,
    error: null,
  }),
}));

// Mock configuration type utility
vi.mock('../../utils/getConfigurationTypeFromTemplate', () => ({
  default: vi.fn(() => 'URL'),
}));

// Mock child components
vi.mock('../../components/create-application/ConfigureName', () => ({
  default: ({
    appName,
    onAppNameChange,
    onReadyChange,
  }: {
    appName: string;
    onAppNameChange: (name: string) => void;
    onReadyChange: (ready: boolean) => void;
  }) => (
    <div data-testid="configure-name">
      <input
        data-testid="app-name-input"
        value={appName}
        onChange={(e) => {
          onAppNameChange(e.target.value);
          onReadyChange(e.target.value.length > 0);
        }}
        placeholder="Enter app name"
      />
    </div>
  ),
}));

vi.mock('../../components/create-application/ConfigureDesign', () => ({
  default: ({
    selectedColor,
    onColorSelect,
    onLogoSelect,
    onBrandingSelectionChange,
  }: {
    appLogo: string | null;
    appName: string;
    selectedColor: string;
    onColorSelect: (color: string) => void;
    onLogoSelect: (logo: string) => void;
    onInitialLogoLoad: (logo: string) => void;
    onReadyChange: (ready: boolean) => void;
    onBrandingSelectionChange?: (useDefault: boolean, brandingId?: string) => void;
  }) => (
    <div data-testid="configure-design">
      <input
        data-testid="color-picker"
        type="color"
        value={selectedColor}
        onChange={(e) => onColorSelect(e.target.value)}
      />
      <button type="button" data-testid="logo-select-btn" onClick={() => onLogoSelect('test-logo.png')}>
        Select Logo
      </button>
      <button
        type="button"
        data-testid="use-default-branding-btn"
        onClick={() => onBrandingSelectionChange?.(true, 'default-branding-id')}
      >
        Use Default Branding
      </button>
      <button type="button" data-testid="use-custom-branding-btn" onClick={() => onBrandingSelectionChange?.(false)}>
        Use Custom Branding
      </button>
    </div>
  ),
}));

vi.mock('../../components/create-application/configure-signin-options/ConfigureSignInOptions', async () => {
  const useApplicationCreateContextModule = await import('../../hooks/useApplicationCreateContext');

  return {
    default: function MockConfigureSignInOptions({
      integrations,
      onIntegrationToggle,
      onReadyChange,
    }: {
      integrations: Record<string, boolean>;
      onIntegrationToggle: (id: string) => void;
      onReadyChange: (ready: boolean) => void;
    }) {
      const {setSelectedAuthFlow} = useApplicationCreateContextModule.default();

      setTimeout(() => {
        setSelectedAuthFlow({
          id: 'test-flow-id',
          name: 'Test Flow',
          flowType: 'AUTHENTICATION',
          handle: 'test-flow',
          activeVersion: 1,
          createdAt: '2024-01-01T00:00:00Z',
          updatedAt: '2024-01-01T00:00:00Z',
        });
        const hasSelection = Object.values(integrations).some((enabled: boolean) => enabled);
        onReadyChange(hasSelection);
      }, 0);

      return (
        <div data-testid="configure-sign-in">
          <button type="button" data-testid="toggle-integration" onClick={() => onIntegrationToggle('basic_auth')}>
            Toggle Integration
          </button>
        </div>
      );
    },
  };
});

vi.mock('../../components/create-application/ConfigureExperience', () => ({
  default: ({
    onReadyChange,
    onApproachChange,
    selectedApproach,
  }: {
    onReadyChange: (ready: boolean) => void;
    onApproachChange: (approach: string) => void;
    selectedApproach: string;
    userTypes: {name: string}[];
    selectedUserTypes: string[];
    onUserTypesChange: (types: string[]) => void;
  }) => {
    setTimeout(() => onReadyChange(true), 0);
    return (
      <div data-testid="configure-experience">
        <span data-testid="current-approach">{selectedApproach}</span>
        <button type="button" data-testid="select-embedded-approach" onClick={() => onApproachChange('EMBEDDED')}>
          Select Embedded
        </button>
        <button type="button" data-testid="select-inbuilt-approach" onClick={() => onApproachChange('INBUILT')}>
          Select Inbuilt
        </button>
      </div>
    );
  },
}));

vi.mock('../../components/create-application/ConfigureStack', () => ({
  default: ({onReadyChange}: {onReadyChange: (ready: boolean) => void}) => {
    setTimeout(() => onReadyChange(true), 0);
    return <div data-testid="configure-stack">Configure Stack</div>;
  },
}));

vi.mock('../../components/create-application/ConfigureDetails', () => ({
  default: ({
    onReadyChange,
    onCallbackUrlChange,
  }: {
    onReadyChange: (ready: boolean) => void;
    onCallbackUrlChange: (url: string) => void;
    technology?: string;
    platform?: string;
    onHostingUrlChange: (url: string) => void;
  }) => {
    setTimeout(() => onReadyChange(true), 0);
    return (
      <div data-testid="configure-details">
        <input
          data-testid="callback-url-input"
          onChange={(e) => onCallbackUrlChange(e.target.value)}
          placeholder="Callback URL"
        />
      </div>
    );
  },
}));

vi.mock('../../components/create-application/Preview', () => ({
  default: ({
    appLogo,
    selectedColor,
  }: {
    appLogo: string | null;
    selectedColor: string;
    integrations: Record<string, boolean>;
  }) => (
    <div data-testid="preview">
      <div data-testid="preview-logo">{appLogo}</div>
      <div data-testid="preview-color">{selectedColor}</div>
    </div>
  ),
}));

describe('ApplicationCreatePage', () => {
  let user: ReturnType<typeof userEvent.setup>;

  const renderWithProviders = () =>
    render(
      <ApplicationCreateProvider>
        <ApplicationCreatePage />
      </ApplicationCreateProvider>,
    );

  beforeEach(async () => {
    user = userEvent.setup();

    window.history.replaceState({}, '', '/');

    vi.clearAllMocks();
    mockNavigate.mockResolvedValue(undefined);

    const getConfigurationTypeFromTemplate = await import('../../utils/getConfigurationTypeFromTemplate');
    vi.mocked(getConfigurationTypeFromTemplate.default).mockReturnValue('URL');
  });

  describe('Initial Rendering', () => {
    it('should render the name step by default', () => {
      renderWithProviders();

      expect(screen.getByTestId('configure-name')).toBeInTheDocument();
      expect(screen.queryByTestId('configure-design')).not.toBeInTheDocument();
    });

    it('should not show preview on first step', () => {
      renderWithProviders();

      expect(screen.queryByTestId('preview')).not.toBeInTheDocument();
    });

    it('should render close button', () => {
      const {container} = renderWithProviders();

      const buttons = container.querySelectorAll('button');
      expect(buttons.length).toBeGreaterThan(0);
    });

    it('should show breadcrumb with current step', () => {
      renderWithProviders();

      expect(screen.getByText('Create an Application')).toBeInTheDocument();
    });
  });

  describe('Step Navigation', () => {
    it('should disable Continue button when name is empty', () => {
      renderWithProviders();

      const continueButton = screen.getByRole('button', {name: /continue/i});
      expect(continueButton).toBeDisabled();
    });

    it('should enable Continue button when name is entered', async () => {
      renderWithProviders();

      const nameInput = screen.getByTestId('app-name-input');
      await user.type(nameInput, 'My App');

      const continueButton = screen.getByRole('button', {name: /continue/i});
      expect(continueButton).toBeEnabled();
    });

    it('should navigate to design step from name step', async () => {
      renderWithProviders();

      const nameInput = screen.getByTestId('app-name-input');
      await user.type(nameInput, 'My App');

      const continueButton = screen.getByRole('button', {name: /continue/i});
      await user.click(continueButton);

      expect(screen.getByTestId('configure-design')).toBeInTheDocument();
      expect(screen.queryByTestId('configure-name')).not.toBeInTheDocument();
    });

    it('should show preview from design step onwards', async () => {
      renderWithProviders();

      const nameInput = screen.getByTestId('app-name-input');
      await user.type(nameInput, 'My App');
      await user.click(screen.getByRole('button', {name: /continue/i}));

      expect(screen.getByTestId('preview')).toBeInTheDocument();
    });

    it('should navigate through all steps', async () => {
      renderWithProviders();

      // Step 1: Name
      await user.type(screen.getByTestId('app-name-input'), 'My App');
      await user.click(screen.getByRole('button', {name: /continue/i}));

      // Step 2: Design
      expect(screen.getByTestId('configure-design')).toBeInTheDocument();
      await user.click(screen.getByRole('button', {name: /continue/i}));

      // Step 3: Sign In Options
      await waitFor(() => {
        expect(screen.getByTestId('configure-sign-in')).toBeInTheDocument();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      // Step 4: Experience
      await waitFor(() => {
        expect(screen.getByTestId('configure-experience')).toBeInTheDocument();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      // Step 5: Stack
      await waitFor(() => {
        expect(screen.getByTestId('configure-stack')).toBeInTheDocument();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      // Step 6: Configure Details
      await waitFor(() => {
        expect(screen.getByTestId('configure-details')).toBeInTheDocument();
      });
    });

    it('should show Back button from design step onwards', async () => {
      renderWithProviders();

      await user.type(screen.getByTestId('app-name-input'), 'My App');
      await user.click(screen.getByRole('button', {name: /continue/i}));

      expect(screen.getByRole('button', {name: /back/i})).toBeInTheDocument();
    });

    it('should navigate back to previous step', async () => {
      renderWithProviders();

      await user.type(screen.getByTestId('app-name-input'), 'My App');
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await user.click(screen.getByRole('button', {name: /back/i}));

      expect(screen.getByTestId('configure-name')).toBeInTheDocument();
      expect(screen.queryByTestId('configure-design')).not.toBeInTheDocument();
    });
  });

  describe('Breadcrumb Navigation', () => {
    it('should update breadcrumb as user progresses', async () => {
      renderWithProviders();

      expect(screen.getByText('Create an Application')).toBeInTheDocument();

      await user.type(screen.getByTestId('app-name-input'), 'My App');
      await user.click(screen.getByRole('button', {name: /continue/i}));

      expect(screen.getByText('Design')).toBeInTheDocument();

      await user.click(screen.getByRole('button', {name: /continue/i}));

      expect(screen.getByText('Sign In Options')).toBeInTheDocument();
    });

    it('should allow clicking on previous breadcrumb steps', async () => {
      renderWithProviders();

      await user.type(screen.getByTestId('app-name-input'), 'My App');
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await user.click(screen.getByRole('button', {name: /continue/i}));

      const firstBreadcrumb = screen.getByText('Create an Application');
      await user.click(firstBreadcrumb);

      expect(screen.getByTestId('configure-name')).toBeInTheDocument();
    });
  });

  describe('Close Functionality', () => {
    it('should navigate to applications list when close button is clicked', async () => {
      const {container} = renderWithProviders();

      const closeButton = container.querySelector('button');
      expect(closeButton).toBeInTheDocument();

      if (closeButton) {
        await user.click(closeButton);

        await waitFor(() => {
          expect(mockNavigate).toHaveBeenCalledWith('/applications');
        });
      }
    });
  });

  describe('Form State Management', () => {
    it('should update app name state', async () => {
      renderWithProviders();

      const nameInput = screen.getByTestId('app-name-input');
      await user.type(nameInput, 'Test App');

      expect(nameInput).toHaveValue('Test App');
    });

    it('should preserve app name when navigating between steps', async () => {
      renderWithProviders();

      const nameInput = screen.getByTestId('app-name-input');
      await user.type(nameInput, 'My App');

      await user.click(screen.getByRole('button', {name: /continue/i}));
      await user.click(screen.getByRole('button', {name: /back/i}));

      expect(screen.getByTestId('app-name-input')).toHaveValue('My App');
    });

    it('should update logo in state', async () => {
      renderWithProviders();

      await user.type(screen.getByTestId('app-name-input'), 'My App');
      await user.click(screen.getByRole('button', {name: /continue/i}));

      const logoButton = screen.getByTestId('logo-select-btn');
      await user.click(logoButton);

      expect(screen.getByTestId('preview-logo')).toHaveTextContent('test-logo.png');
    });
  });

  describe('Application Creation - Inbuilt Approach', () => {
    it('should create application with OAuth config for inbuilt approach', async () => {
      mockCreateBranding.mockImplementation((_data, {onSuccess}: {onSuccess: (branding: Branding) => void}) => {
        onSuccess({id: 'branding-123', displayName: 'Test Branding', preferences: {}} as Branding);
      });

      mockCreateApplication.mockImplementation((_data, {onSuccess}: {onSuccess: (app: Application) => void}) => {
        onSuccess({id: 'app-123', name: 'My App'} as Application);
      });

      renderWithProviders();

      // Navigate through all steps
      await user.type(screen.getByTestId('app-name-input'), 'My App');
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await user.click(screen.getByRole('button', {name: /continue/i}));

      await waitFor(() => {
        expect(screen.getByTestId('configure-sign-in')).toBeInTheDocument();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      await waitFor(() => {
        expect(screen.getByTestId('configure-experience')).toBeInTheDocument();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      await waitFor(() => {
        expect(screen.getByTestId('configure-stack')).toBeInTheDocument();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      await waitFor(() => {
        expect(screen.getByTestId('configure-details')).toBeInTheDocument();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      await waitFor(() => {
        expect(mockCreateBranding).toHaveBeenCalled();
        expect(mockCreateApplication).toHaveBeenCalled();
      });

      // Verify OAuth config was included
      const createAppCall = mockCreateApplication.mock.calls[0][0] as Application;
      expect(createAppCall.inbound_auth_config).toBeDefined();
      expect(createAppCall.inbound_auth_config?.[0]).toBeDefined();
      expect(createAppCall.inbound_auth_config?.[0]?.type).toBe('oauth2');
    });

    it('should navigate to application details page after creation', async () => {
      mockCreateBranding.mockImplementation((_data, {onSuccess}: {onSuccess: (branding: Branding) => void}) => {
        onSuccess({id: 'branding-123', displayName: 'Test Branding', preferences: {}} as Branding);
      });

      mockCreateApplication.mockImplementation((_data, {onSuccess}: {onSuccess: (app: Application) => void}) => {
        onSuccess({id: 'app-123', name: 'My App'} as Application);
      });

      renderWithProviders();

      // Navigate through all steps
      await user.type(screen.getByTestId('app-name-input'), 'My App');
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await user.click(screen.getByRole('button', {name: /continue/i}));

      await waitFor(() => {
        expect(screen.getByTestId('configure-sign-in')).toBeInTheDocument();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      await waitFor(() => {
        expect(screen.getByTestId('configure-experience')).toBeInTheDocument();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      await waitFor(() => {
        expect(screen.getByTestId('configure-stack')).toBeInTheDocument();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      await waitFor(() => {
        expect(screen.getByTestId('configure-details')).toBeInTheDocument();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      await waitFor(() => {
        expect(mockNavigate).toHaveBeenCalledWith('/applications/app-123');
      });
    });
  });

  describe('Application Creation - Embedded Approach', () => {
    it('should create application without OAuth config for embedded approach', async () => {
      const getConfigurationTypeFromTemplate = await import('../../utils/getConfigurationTypeFromTemplate');
      vi.mocked(getConfigurationTypeFromTemplate.default).mockReturnValue('NONE');

      mockCreateBranding.mockImplementation((_data, {onSuccess}: {onSuccess: (branding: Branding) => void}) => {
        onSuccess({id: 'branding-123', displayName: 'Test Branding', preferences: {}} as Branding);
      });

      mockCreateApplication.mockImplementation((_data, {onSuccess}: {onSuccess: (app: Application) => void}) => {
        onSuccess({id: 'app-123', name: 'My App'} as Application);
      });

      renderWithProviders();

      // Navigate to experience step
      await user.type(screen.getByTestId('app-name-input'), 'My App');
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await user.click(screen.getByRole('button', {name: /continue/i}));

      await waitFor(() => {
        expect(screen.getByTestId('configure-sign-in')).toBeInTheDocument();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      // Select embedded approach
      await waitFor(() => {
        expect(screen.getByTestId('configure-experience')).toBeInTheDocument();
      });
      const selectEmbeddedBtn = screen.getByTestId('select-embedded-approach');
      await user.click(selectEmbeddedBtn);
      await user.click(screen.getByRole('button', {name: /continue/i}));

      // Continue from stack - should create app immediately
      await waitFor(() => {
        expect(screen.getByTestId('configure-stack')).toBeInTheDocument();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      await waitFor(() => {
        expect(mockCreateBranding).toHaveBeenCalled();
        expect(mockCreateApplication).toHaveBeenCalled();
      });

      // Verify OAuth config was NOT included
      const createAppCall = mockCreateApplication.mock.calls[0][0] as Application;
      expect(createAppCall.inbound_auth_config).toBeUndefined();
    });

    it('should skip configure step for embedded approach', async () => {
      const getConfigurationTypeFromTemplate = await import('../../utils/getConfigurationTypeFromTemplate');
      vi.mocked(getConfigurationTypeFromTemplate.default).mockReturnValue('NONE');

      mockCreateBranding.mockImplementation((_data, {onSuccess}: {onSuccess: (branding: Branding) => void}) => {
        onSuccess({id: 'branding-123', displayName: 'Test Branding', preferences: {}} as Branding);
      });

      mockCreateApplication.mockImplementation((_data, {onSuccess}: {onSuccess: (app: Application) => void}) => {
        onSuccess({id: 'app-123', name: 'My App'} as Application);
      });

      renderWithProviders();

      await user.type(screen.getByTestId('app-name-input'), 'My App');
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await user.click(screen.getByRole('button', {name: /continue/i}));

      await waitFor(() => {
        expect(screen.getByTestId('configure-sign-in')).toBeInTheDocument();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      await waitFor(() => {
        expect(screen.getByTestId('configure-experience')).toBeInTheDocument();
      });
      await user.click(screen.getByTestId('select-embedded-approach'));
      await user.click(screen.getByRole('button', {name: /continue/i}));

      await waitFor(() => {
        expect(screen.getByTestId('configure-stack')).toBeInTheDocument();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      // Should NOT show configure details step
      await waitFor(() => {
        expect(screen.queryByTestId('configure-details')).not.toBeInTheDocument();
        expect(mockCreateApplication).toHaveBeenCalled();
      });
    });
  });

  describe('Error Handling', () => {
    it('should show error when branding creation fails', async () => {
      mockCreateBranding.mockImplementation((_data, {onError}: {onError: (error: Error) => void}) => {
        onError(new Error('Failed to create branding'));
      });

      renderWithProviders();

      await user.type(screen.getByTestId('app-name-input'), 'My App');
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await user.click(screen.getByRole('button', {name: /continue/i}));

      await waitFor(() => {
        expect(screen.getByTestId('configure-sign-in')).toBeInTheDocument();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      await waitFor(() => {
        expect(screen.getByTestId('configure-experience')).toBeInTheDocument();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      await waitFor(() => {
        expect(screen.getByTestId('configure-stack')).toBeInTheDocument();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      await waitFor(() => {
        expect(screen.getByTestId('configure-details')).toBeInTheDocument();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      await waitFor(() => {
        expect(screen.getByText(/failed to create branding/i)).toBeInTheDocument();
      });
    });

    it('should show error when application creation fails', async () => {
      mockCreateBranding.mockImplementation((_data, {onSuccess}: {onSuccess: (branding: Branding) => void}) => {
        onSuccess({id: 'branding-123', displayName: 'Test Branding', preferences: {}} as Branding);
      });

      mockCreateApplication.mockImplementation((_data, {onError}: {onError: (error: Error) => void}) => {
        onError(new Error('Failed to create application'));
      });

      renderWithProviders();

      await user.type(screen.getByTestId('app-name-input'), 'My App');
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await user.click(screen.getByRole('button', {name: /continue/i}));

      await waitFor(() => {
        expect(screen.getByTestId('configure-sign-in')).toBeInTheDocument();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      await waitFor(() => {
        expect(screen.getByTestId('configure-experience')).toBeInTheDocument();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      await waitFor(() => {
        expect(screen.getByTestId('configure-stack')).toBeInTheDocument();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      await waitFor(() => {
        expect(screen.getByTestId('configure-details')).toBeInTheDocument();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      await waitFor(
        () => {
          expect(screen.getByText(/failed to create application/i)).toBeInTheDocument();
        },
        {timeout: 10000},
      );
    });

    it('should allow dismissing error message', async () => {
      mockCreateBranding.mockImplementation((_data, {onError}: {onError: (error: Error) => void}) => {
        onError(new Error('Failed to create branding'));
      });

      renderWithProviders();

      await user.type(screen.getByTestId('app-name-input'), 'My App');
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await user.click(screen.getByRole('button', {name: /continue/i}));

      await waitFor(() => {
        expect(screen.getByTestId('configure-sign-in')).toBeInTheDocument();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      await waitFor(() => {
        expect(screen.getByTestId('configure-experience')).toBeInTheDocument();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      await waitFor(() => {
        expect(screen.getByTestId('configure-stack')).toBeInTheDocument();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      await waitFor(() => {
        expect(screen.getByTestId('configure-details')).toBeInTheDocument();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      await waitFor(
        () => {
          expect(screen.getByText(/failed to create branding/i)).toBeInTheDocument();
        },
        {timeout: 10000},
      );

      const closeButton = screen.getByLabelText(/close/i);
      await user.click(closeButton);

      await waitFor(() => {
        expect(screen.queryByText(/failed to create branding/i)).not.toBeInTheDocument();
      });
    });
  });

  describe('Branding Selection', () => {
    it('should use custom branding by default', async () => {
      mockCreateBranding.mockImplementation((_data, {onSuccess}: {onSuccess: (branding: Branding) => void}) => {
        onSuccess({id: 'branding-123', displayName: 'Test Branding', preferences: {}} as Branding);
      });

      mockCreateApplication.mockImplementation((_data, {onSuccess}: {onSuccess: (app: Application) => void}) => {
        onSuccess({id: 'app-123', name: 'My App'} as Application);
      });

      renderWithProviders();

      await user.type(screen.getByTestId('app-name-input'), 'My App');
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await user.click(screen.getByRole('button', {name: /continue/i}));

      await waitFor(() => {
        expect(screen.getByTestId('configure-sign-in')).toBeInTheDocument();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      await waitFor(() => {
        expect(screen.getByTestId('configure-experience')).toBeInTheDocument();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      await waitFor(() => {
        expect(screen.getByTestId('configure-stack')).toBeInTheDocument();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      await waitFor(() => {
        expect(screen.getByTestId('configure-details')).toBeInTheDocument();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      await waitFor(() => {
        expect(mockCreateBranding).toHaveBeenCalled();
      });
    });

    it('should use default branding when selected', async () => {
      mockCreateApplication.mockImplementation((_data, {onSuccess}: {onSuccess: (app: Application) => void}) => {
        onSuccess({id: 'app-123', name: 'My App'} as Application);
      });

      renderWithProviders();

      await user.type(screen.getByTestId('app-name-input'), 'My App');
      await user.click(screen.getByRole('button', {name: /continue/i}));

      // Select default branding
      const useDefaultBtn = screen.getByTestId('use-default-branding-btn');
      await user.click(useDefaultBtn);

      await user.click(screen.getByRole('button', {name: /continue/i}));

      await waitFor(() => {
        expect(screen.getByTestId('configure-sign-in')).toBeInTheDocument();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      await waitFor(() => {
        expect(screen.getByTestId('configure-experience')).toBeInTheDocument();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      await waitFor(() => {
        expect(screen.getByTestId('configure-stack')).toBeInTheDocument();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      await waitFor(() => {
        expect(screen.getByTestId('configure-details')).toBeInTheDocument();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      await waitFor(() => {
        expect(mockCreateApplication).toHaveBeenCalled();
        // Should NOT create new branding
        expect(mockCreateBranding).not.toHaveBeenCalled();
      });
    });
  });

  describe('Integration Toggle', () => {
    it('should allow toggling integrations', async () => {
      renderWithProviders();

      await user.type(screen.getByTestId('app-name-input'), 'My App');
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await user.click(screen.getByRole('button', {name: /continue/i}));

      await waitFor(() => {
        expect(screen.getByTestId('configure-sign-in')).toBeInTheDocument();
      });

      const toggleButton = screen.getByTestId('toggle-integration');
      await user.click(toggleButton);

      expect(screen.getByTestId('configure-sign-in')).toBeInTheDocument();
    });
  });

  describe('Callback URL Configuration', () => {
    it('should update OAuth config when callback URL changes', async () => {
      renderWithProviders();

      await user.type(screen.getByTestId('app-name-input'), 'My App');
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await user.click(screen.getByRole('button', {name: /continue/i}));

      await waitFor(() => {
        expect(screen.getByTestId('configure-sign-in')).toBeInTheDocument();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      await waitFor(() => {
        expect(screen.getByTestId('configure-experience')).toBeInTheDocument();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      await waitFor(() => {
        expect(screen.getByTestId('configure-stack')).toBeInTheDocument();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      await waitFor(() => {
        expect(screen.getByTestId('configure-details')).toBeInTheDocument();
      });

      const callbackInput = screen.getByTestId('callback-url-input');
      await user.type(callbackInput, 'https://example.com/callback');

      expect(callbackInput).toHaveValue('https://example.com/callback');
    });
  });
});
