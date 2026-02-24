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
import type {Theme} from '@thunder/shared-design';
import type {Application} from '../../models/application';
import ApplicationCreatePage from '../ApplicationCreatePage';
import ApplicationCreateProvider from '../../contexts/ApplicationCreate/ApplicationCreateProvider';

// Mock functions
const mockCreateApplication = vi.fn();
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

// Mock design hooks
vi.mock('@thunder/shared-design', () => ({
  useGetThemes: () => ({
    data: {themes: [{id: 'theme-1', displayName: 'Default Theme', theme: {}}]},
    isLoading: false,
  }),
  useGetTheme: () => ({
    data: null,
    isLoading: false,
  }),
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
const {mockCreateFlow, mockGenerateFlowGraph} = vi.hoisted(() => ({
  mockCreateFlow: vi.fn(),
  mockGenerateFlowGraph: vi.fn(),
}));

vi.mock('../../../flows/api/useCreateFlow', () => ({
  default: () => ({
    mutate: mockCreateFlow,
    isPending: false,
  }),
}));

vi.mock('../../../flows/utils/generateFlowGraph', () => ({
  default: mockGenerateFlowGraph,
}));

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
    onLogoSelect,
    onThemeSelect,
  }: {
    appLogo: string | null;
    selectedTheme: Theme | null;
    onLogoSelect: (logo: string) => void;
    onInitialLogoLoad: (logo: string) => void;
    onReadyChange: (ready: boolean) => void;
    onThemeSelect?: (themeId: string, themeConfig: Theme) => void;
  }) => (
    <div data-testid="configure-design">
      <button type="button" data-testid="logo-select-btn" onClick={() => onLogoSelect('test-logo.png')}>
        Select Logo
      </button>
      <button
        type="button"
        data-testid="select-theme-btn"
        onClick={() => onThemeSelect?.('theme-1', {} as Theme)}
      >
        Select Theme
      </button>
    </div>
  ),
}));

vi.mock('../../components/create-application/configure-signin-options/ConfigureSignInOptions', async () => {
  const useApplicationCreateContextModule = await import('../../hooks/useApplicationCreateContext');

  return {
    default: vi.fn(({
      integrations,
      onIntegrationToggle,
      onReadyChange,
    }: {
      integrations: Record<string, boolean>;
      onIntegrationToggle: (id: string) => void;
      onReadyChange: (ready: boolean) => void;
    }) => {
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
    }),
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

vi.mock('../../components/create-application/ShowClientSecret', () => ({
  default: ({
    appName,
    clientSecret,
    onContinue,
  }: {
    appName: string;
    clientSecret: string;
    onCopySecret: () => void;
    onContinue: () => void;
  }) => (
    <div data-testid="show-client-secret">
      <div data-testid="client-secret-app-name">{appName}</div>
      <div data-testid="client-secret-value">{clientSecret}</div>
      <button type="button" data-testid="client-secret-continue" onClick={onContinue}>
        Continue
      </button>
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
        expect(mockCreateApplication).toHaveBeenCalled();
      });

      // Verify OAuth config was included
      const createAppCall = mockCreateApplication.mock.calls[0][0] as Application;
      expect(createAppCall.inbound_auth_config).toBeDefined();
      expect(createAppCall.inbound_auth_config?.[0]).toBeDefined();
      expect(createAppCall.inbound_auth_config?.[0]?.type).toBe('oauth2');
    });

    it('should navigate to application details page after creation', async () => {
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
        expect(mockCreateApplication).toHaveBeenCalled();
      });

      // Verify OAuth config was NOT included
      const createAppCall = mockCreateApplication.mock.calls[0][0] as Application;
      expect(createAppCall.inbound_auth_config).toBeUndefined();
    });

    it('should skip configure step for embedded approach', async () => {
      const getConfigurationTypeFromTemplate = await import('../../utils/getConfigurationTypeFromTemplate');
      vi.mocked(getConfigurationTypeFromTemplate.default).mockReturnValue('NONE');

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
    it('should show error when application creation fails', async () => {
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

      const closeButton = screen.getByLabelText(/close/i);
      await user.click(closeButton);

      await waitFor(() => {
        expect(screen.queryByText(/failed to create application/i)).not.toBeInTheDocument();
      });
    });
  });

  describe('Theme Selection', () => {
    it('should allow selecting a theme', async () => {
      mockCreateApplication.mockImplementation((_data, {onSuccess}: {onSuccess: (app: Application) => void}) => {
        onSuccess({id: 'app-123', name: 'My App'} as Application);
      });

      renderWithProviders();

      await user.type(screen.getByTestId('app-name-input'), 'My App');
      await user.click(screen.getByRole('button', {name: /continue/i}));

      // Select a theme
      const selectThemeBtn = screen.getByTestId('select-theme-btn');
      await user.click(selectThemeBtn);

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
      });

      // Verify theme_id was included in the application creation
      const createAppCall = mockCreateApplication.mock.calls[0][0] as Application;
      expect(createAppCall.theme_id).toBe('theme-1');
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

  describe('Client Secret Display (COMPLETE Step)', () => {
    it('should show COMPLETE step when application is created with client_secret', async () => {
      mockCreateApplication.mockImplementation((_data, {onSuccess}: {onSuccess: (app: Application) => void}) => {
        onSuccess({
          id: 'app-123',
          name: 'My App',
          inbound_auth_config: [
            {
              type: 'oauth2',
              config: {
                client_id: 'test-client-id',
                client_secret: 'test_secret_12345',
                redirect_uris: ['https://example.com/callback'],
              },
            },
          ],
        } as Application);
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

      // Should show COMPLETE step with client secret
      await waitFor(() => {
        expect(screen.getByTestId('show-client-secret')).toBeInTheDocument();
      });

      expect(screen.getByTestId('client-secret-app-name')).toHaveTextContent('My App');
      expect(screen.getByTestId('client-secret-value')).toHaveTextContent('test_secret_12345');
    });

    it('should not show COMPLETE step when application is created without client_secret', async () => {
      mockCreateApplication.mockImplementation((_data, {onSuccess}: {onSuccess: (app: Application) => void}) => {
        onSuccess({
          id: 'app-123',
          name: 'My App',
          inbound_auth_config: [],
        } as Application);
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

      // Should navigate directly to application details page
      await waitFor(() => {
        expect(mockNavigate).toHaveBeenCalledWith('/applications/app-123');
      });

      // Should not show COMPLETE step
      expect(screen.queryByTestId('show-client-secret')).not.toBeInTheDocument();
    });

    it('should navigate to application details when continue is clicked on COMPLETE step', async () => {
      mockCreateApplication.mockImplementation((_data, {onSuccess}: {onSuccess: (app: Application) => void}) => {
        onSuccess({
          id: 'app-456',
          name: 'My App',
          inbound_auth_config: [
            {
              type: 'oauth2',
              config: {
                client_id: 'test-client-id',
                client_secret: 'test_secret_12345',
                redirect_uris: ['https://example.com/callback'],
              },
            },
          ],
        } as Application);
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

      // Should show COMPLETE step
      await waitFor(() => {
        expect(screen.getByTestId('show-client-secret')).toBeInTheDocument();
      });

      // Click continue on COMPLETE step
      const continueButton = screen.getByTestId('client-secret-continue');
      await user.click(continueButton);

      // Should navigate to application details page
      await waitFor(() => {
        expect(mockNavigate).toHaveBeenCalledWith('/applications/app-456');
      });
    });

    it('should not show back button on COMPLETE step', async () => {
      mockCreateApplication.mockImplementation((_data, {onSuccess}: {onSuccess: (app: Application) => void}) => {
        onSuccess({
          id: 'app-123',
          name: 'My App',
          inbound_auth_config: [
            {
              type: 'oauth2',
              config: {
                client_id: 'test-client-id',
                client_secret: 'test_secret_12345',
                redirect_uris: ['https://example.com/callback'],
              },
            },
          ],
        } as Application);
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

      // Should show COMPLETE step
      await waitFor(() => {
        expect(screen.getByTestId('show-client-secret')).toBeInTheDocument();
      });

      // Back button should not be present
      expect(screen.queryByRole('button', {name: /back/i})).not.toBeInTheDocument();
    });

    it('should not show preview panel on COMPLETE step', async () => {
      mockCreateApplication.mockImplementation((_data, {onSuccess}: {onSuccess: (app: Application) => void}) => {
        onSuccess({
          id: 'app-123',
          name: 'My App',
          inbound_auth_config: [
            {
              type: 'oauth2',
              config: {
                client_id: 'test-client-id',
                client_secret: 'test_secret_12345',
                redirect_uris: ['https://example.com/callback'],
              },
            },
          ],
        } as Application);
      });

      renderWithProviders();

      await user.type(screen.getByTestId('app-name-input'), 'My App');
      await user.click(screen.getByRole('button', {name: /continue/i}));

      // Preview should be visible on DESIGN step
      await waitFor(() => {
        expect(screen.getByTestId('preview')).toBeInTheDocument();
      });

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

      // Should show COMPLETE step
      await waitFor(() => {
        expect(screen.getByTestId('show-client-secret')).toBeInTheDocument();
      });

      // Preview should not be visible on COMPLETE step
      expect(screen.queryByTestId('preview')).not.toBeInTheDocument();
    });
  });

  describe('Flow Generation', () => {
    it('should generate flow and create application when integrations are selected but no flow matches', async () => {
      // Mock createFlow to return success
      mockCreateFlow.mockImplementation((_data, {onSuccess}: {onSuccess: (flow: unknown) => void}) => {
        onSuccess({
          id: 'generated-flow-id',
          name: 'Generated Flow',
          handle: 'generated-flow',
        });
      });

      // Mock createApplication to success
      mockCreateApplication.mockImplementation((_data, {onSuccess}: {onSuccess: (app: Application) => void}) => {
        onSuccess({id: 'app-with-generated-flow', name: 'My App'} as Application);
      });

      // Override MockConfigureSignInOptions to simulate selection without setting a flow
      const ConfigureSignInOptionsModule = await import('../../components/create-application/configure-signin-options/ConfigureSignInOptions');
      const useApplicationCreateContextModule = await import('../../hooks/useApplicationCreateContext');

      vi.mocked(ConfigureSignInOptionsModule.default).mockImplementation(
        ({onReadyChange}: {onReadyChange?: (ready: boolean) => void}) => {
          // eslint-disable-next-line @typescript-eslint/no-unsafe-call, @typescript-eslint/no-unsafe-assignment
          const {setSelectedAuthFlow, setIntegrations} = useApplicationCreateContextModule.default();

          const handleSetup = () => {
             // Explicitly set flow to null to trigger generation logic
            setSelectedAuthFlow(null);
            // Explicitly set integrations
            setIntegrations({'basic_auth': true});
            onReadyChange?.(true);
          };

          return (
            <div data-testid="configure-sign-in">
              <button type="button" data-testid="setup-flow-generation" onClick={handleSetup}>Setup Flow Generation</button>
            </div>
          );
        }
      );

      renderWithProviders();

      // Navigate to options step
      await user.type(screen.getByTestId('app-name-input'), 'My App');
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await user.click(screen.getByRole('button', {name: /continue/i}));

      // At Options step
      await waitFor(() => {
        expect(screen.getByTestId('configure-sign-in')).toBeInTheDocument();
      });

      // Trigger setup
      await user.click(screen.getByTestId('setup-flow-generation'));

      await user.click(screen.getByRole('button', {name: /continue/i}));

      // Experience -> Stack -> Configure
      await user.click(screen.getByRole('button', {name: /continue/i})); // Experience
      await user.click(screen.getByRole('button', {name: /continue/i})); // Stack
      await user.click(screen.getByRole('button', {name: /continue/i})); // Configure Details

      // Verify generateFlowGraph called
      await waitFor(() => {
        expect(mockGenerateFlowGraph).toHaveBeenCalled();
        expect(mockCreateFlow).toHaveBeenCalled();
        expect(mockCreateApplication).toHaveBeenCalled();
      });

      // Verify createApplication called with generated flow ID
      const createAppCall = mockCreateApplication.mock.calls[0][0] as Application;
      expect(createAppCall.auth_flow_id).toBe('generated-flow-id');
    });

    it('should show error when flow generation fails', async () => {
       // Mock createFlow to fail
       mockCreateFlow.mockImplementation((_data, {onError}: {onError: (error: Error) => void}) => {
        onError(new Error('Flow generation failed'));
      });

       // Override MockConfigureSignInOptions to simulate selection without setting a flow
       const ConfigureSignInOptionsModule = await import('../../components/create-application/configure-signin-options/ConfigureSignInOptions');
       const useApplicationCreateContextModule = await import('../../hooks/useApplicationCreateContext');

       vi.mocked(ConfigureSignInOptionsModule.default).mockImplementation(
        ({onReadyChange}: {onReadyChange?: (ready: boolean) => void}) => {
          // eslint-disable-next-line @typescript-eslint/no-unsafe-call, @typescript-eslint/no-unsafe-assignment
          const {setSelectedAuthFlow, setIntegrations} = useApplicationCreateContextModule.default();

          const handleSetup = () => {
             setSelectedAuthFlow(null);
             setIntegrations({'basic_auth': true});
             onReadyChange?.(true);
          };

          return (
            <div data-testid="configure-sign-in">
               <button type="button" data-testid="setup-flow-generation-error" onClick={handleSetup}>Setup Flow Generation Error</button>
            </div>
          );
        }
      );

      renderWithProviders();

      // Navigate to trigger point
      await user.type(screen.getByTestId('app-name-input'), 'My App');
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await user.click(screen.getByRole('button', {name: /continue/i})); // Design

      // Options step
      await waitFor(() => {
        expect(screen.getByTestId('configure-sign-in')).toBeInTheDocument();
      });

      // Trigger setup
      await user.click(screen.getByTestId('setup-flow-generation-error'));

      await user.click(screen.getByRole('button', {name: /continue/i})); // Options -> Experience
      await user.click(screen.getByRole('button', {name: /continue/i})); // Experience -> Stack
      await user.click(screen.getByRole('button', {name: /continue/i})); // Stack -> Configure
      await user.click(screen.getByRole('button', {name: /continue/i})); // Configure -> Create

      await waitFor(() => {
         expect(screen.getByText('Flow generation failed')).toBeInTheDocument();
      });
    });
  });
});
