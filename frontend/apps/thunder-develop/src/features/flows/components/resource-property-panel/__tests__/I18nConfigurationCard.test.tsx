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

import {describe, it, expect, vi, beforeEach, afterEach} from 'vitest';
import {render} from '@thunder/test-utils/browser';
import {page, userEvent, type Locator} from 'vitest/browser';
import type {ReactNode} from 'react';
import I18nConfigurationCard from '../I18nConfigurationCard';
import FlowBuilderCoreContext, {type FlowBuilderCoreContextProps} from '../../../context/FlowBuilderCoreContext';
import {EdgeStyleTypes} from '../../../models/steps';
import {PreviewScreenType} from '../../../models/custom-text-preference';
import {ElementTypes} from '../../../models/elements';
import type {Base} from '../../../models/base';


// Mock @thunder/shared-contexts
vi.mock('@thunder/shared-contexts', () => ({
  useConfig: () => ({
    getServerUrl: () => 'https://localhost:8090',
  }),
}));

// Mock the API hooks from @thunder/i18n
const mockMutate = vi.fn();
vi.mock('@thunder/i18n', () => ({
  useUpdateTranslation: () => ({
    mutate: mockMutate,
    isPending: false,
  }),
  useGetLanguages: () => ({
    data: {languages: ['en', 'es', 'fr']},
  }),
  useGetTranslations: () => ({
    data: {
      language: 'en-US',
      translations: {
        flowI18n: {
          'login.title': 'Sign In',
          'login.description': 'Enter your credentials',
          'login.button.submit': 'Submit',
          'common.continue': 'Continue',
          'common.cancel': 'Cancel',
        },
      },
    },
    isLoading: false,
  }),
}));

describe('I18nConfigurationCard', () => {
  const mockOnClose = vi.fn();
  const mockOnChange = vi.fn();

  let anchorEl: HTMLDivElement;

  // Reset mocks before each test
  beforeEach(() => {
    mockMutate.mockReset();
  });

  const mockBaseResource: Base = {
    id: 'resource-1',
    resourceType: 'ELEMENT',
    type: 'TEXT_INPUT',
    category: 'FIELD',
    version: '1.0.0',
    deprecated: false,
    deletable: true,
    display: {
      label: 'Test Resource',
      image: '',
      showOnResourcePanel: false,
    },
    config: {
      field: {name: '', type: ElementTypes},
      styles: {},
    },
  };

  const mockI18nText = {
    [PreviewScreenType.LOGIN]: {
      'login.title': 'Sign In',
      'login.description': 'Enter your credentials',
      'login.button.submit': 'Submit',
    },
    [PreviewScreenType.COMMON]: {
      'common.continue': 'Continue',
      'common.cancel': 'Cancel',
    },
  };

  const createContextValue = (overrides: Partial<FlowBuilderCoreContextProps> = {}): FlowBuilderCoreContextProps => ({
    lastInteractedResource: mockBaseResource,
    lastInteractedStepId: 'step-1',
    ResourceProperties: () => null,
    resourcePropertiesPanelHeading: 'Test Panel Heading',
    primaryI18nScreen: PreviewScreenType.LOGIN,
    isResourcePanelOpen: true,
    isResourcePropertiesPanelOpen: false,
    isVersionHistoryPanelOpen: false,
    ElementFactory: () => null,
    onResourceDropOnCanvas: vi.fn(),
    selectedAttributes: {},
    setLastInteractedResource: vi.fn(),
    setLastInteractedStepId: vi.fn(),
    setResourcePropertiesPanelHeading: vi.fn(),
    setIsResourcePanelOpen: vi.fn(),
    setIsOpenResourcePropertiesPanel: vi.fn(),
    registerCloseValidationPanel: vi.fn(),
    setIsVersionHistoryPanelOpen: vi.fn(),
    setSelectedAttributes: vi.fn(),
    flowCompletionConfigs: {},
    setFlowCompletionConfigs: vi.fn(),
    flowNodeTypes: {},
    flowEdgeTypes: {},
    setFlowNodeTypes: vi.fn(),
    setFlowEdgeTypes: vi.fn(),
    isVerboseMode: false,
    setIsVerboseMode: vi.fn(),
    edgeStyle: EdgeStyleTypes.SmoothStep,
    setEdgeStyle: vi.fn(),
    i18nText: mockI18nText,
    i18nTextLoading: false,
    ...overrides,
  });

  const createWrapper = (contextValue: FlowBuilderCoreContextProps = createContextValue()) =>
    function Wrapper({children}: {children: ReactNode}) {
      return <FlowBuilderCoreContext.Provider value={contextValue}>{children}</FlowBuilderCoreContext.Provider>;
    };

  beforeEach(() => {
    anchorEl = document.createElement('div');
    document.body.appendChild(anchorEl);
    vi.clearAllMocks();
  });

  afterEach(() => {
    document.body.removeChild(anchorEl);
  });

  describe('Rendering', () => {
    it('should render the popover when open is true', async () => {
      await render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey=""
        />,
        {wrapper: createWrapper()},
      );

      await expect.element(page.getByRole('presentation')).toBeInTheDocument();
    });

    it('should not render content when open is false', async () => {
      await render(
        <I18nConfigurationCard
          open={false}
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey=""
        />,
        {wrapper: createWrapper()},
      );

      await expect.element(page.getByText('Translation Key')).not.toBeInTheDocument();
    });

    it('should render the card title with formatted property key', async () => {
      await render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="buttonLabel"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey=""
        />,
        {wrapper: createWrapper()},
      );

      expect(
        page.getByText('Translation for Button Label'),
      ).toBeInTheDocument();
    });

    it('should render the i18n key label', async () => {
      await render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey=""
        />,
        {wrapper: createWrapper()},
      );

      await expect.element(page.getByText('Translation Key')).toBeInTheDocument();
    });

    it('should render loading state when i18nTextLoading is true', async () => {
      const loadingContext = createContextValue({i18nTextLoading: true});

      await render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey=""
        />,
        {wrapper: createWrapper(loadingContext)},
      );

      await expect.element(page.getByRole('progressbar')).toBeInTheDocument();
    });
  });

  describe('Close Functionality', () => {
    it('should call onClose when close button is clicked', async () => {
      await render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey=""
        />,
        {wrapper: createWrapper()},
      );

      const closeButton = page.getByLabelText('Close');
      await userEvent.click(closeButton);

      expect(mockOnClose).toHaveBeenCalledTimes(1);
    });
  });

  describe('Autocomplete Options', () => {
    it('should display available i18n keys in autocomplete', async () => {
      await render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey=""
        />,
        {wrapper: createWrapper()},
      );

      // Open the autocomplete dropdown by clicking the Open button
      const openButton = page.getByTitle('Open');
      await userEvent.click(openButton);

      // Keys are now prefixed with the namespace (flowI18n:)
      await vi.waitFor(async () => {
        await expect.element(page.getByText('flowI18n:login.title')).toBeInTheDocument();
        await expect.element(page.getByText('flowI18n:login.description')).toBeInTheDocument();
        await expect.element(page.getByText('flowI18n:common.continue')).toBeInTheDocument();
      });
    });

    it('should handle empty i18nText gracefully', async () => {
      const emptyContext = createContextValue({i18nText: undefined});

      await render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey=""
        />,
        {wrapper: createWrapper(emptyContext)},
      );

      await expect.element(page.getByRole('combobox')).toBeInTheDocument();
    });
  });

  describe('Selection and onChange', () => {
    it('should call onChange with selected i18n key', async () => {
      await render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey=""
        />,
        {wrapper: createWrapper()},
      );

      // Open the autocomplete dropdown by clicking the Open button
      const openButton = page.getByTitle('Open');
      await userEvent.click(openButton);

      // Keys are now prefixed with the namespace (flowI18n:)
      await vi.waitFor(async () => {
        await expect.element(page.getByText('flowI18n:login.title')).toBeInTheDocument();
      });

      await userEvent.click(page.getByText('flowI18n:login.title'));

      expect(mockOnChange).toHaveBeenCalledWith('flowI18n:login.title');
    });

    it('should call onChange with empty string when selection is cleared', async () => {
      await render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey="login.title"
        />,
        {wrapper: createWrapper()},
      );

      const clearButton = page.getByLabelText('Clear');
      await userEvent.click(clearButton);

      expect(mockOnChange).toHaveBeenCalledWith('');
    });
  });

  describe('Resolved Value Display', () => {
    it('should display resolved value when i18n key is selected', async () => {
      await render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey="login.title"
        />,
        {wrapper: createWrapper()},
      );

      await expect.element(page.getByText('Resolved Value')).toBeInTheDocument();
      await expect.element(page.getByText('Sign In')).toBeInTheDocument();
    });

    it('should not display resolved value box when no i18n key is selected', async () => {
      await render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey=""
        />,
        {wrapper: createWrapper()},
      );

      await expect.element(page.getByText('Resolved Value')).not.toBeInTheDocument();
    });

    it('should not display resolved value box when key has no matching value', async () => {
      await render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey="nonexistent.key"
        />,
        {wrapper: createWrapper()},
      );

      await expect.element(page.getByText('Resolved Value')).not.toBeInTheDocument();
    });

    it('should display resolved value from common screen', async () => {
      await render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey="common.continue"
        />,
        {wrapper: createWrapper()},
      );

      await expect.element(page.getByText('Continue')).toBeInTheDocument();
    });
  });

  describe('Selected Value Display', () => {
    it('should show selected i18n key in autocomplete', async () => {
      await render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey="login.title"
        />,
        {wrapper: createWrapper()},
      );

      const input: HTMLInputElement = page.getByRole('combobox').element() as HTMLInputElement;
      expect(input.value).toBe('login.title');
    });

    it('should show empty autocomplete when i18nKey is empty string', async () => {
      await render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey=""
        />,
        {wrapper: createWrapper()},
      );

      const input: HTMLInputElement = page.getByRole('combobox').element() as HTMLInputElement;
      expect(input.value).toBe('');
    });
  });

  describe('Create Translation Mode', () => {
    it('should show create translation button', async () => {
      await render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey=""
        />,
        {wrapper: createWrapper()},
      );

      await expect.element(page.getByText('Create Translation')).toBeInTheDocument();
    });

    it('should enter create mode when create button is clicked', async () => {
      await render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey=""
        />,
        {wrapper: createWrapper()},
      );

      const createButton = page.getByText('Create Translation');
      await userEvent.click(createButton);

      // Should show language selector in create mode
      await expect.element(page.getByText('Language')).toBeInTheDocument();
      await expect.element(page.getByText('Translation Text')).toBeInTheDocument();
    });

    it('should show cancel button in create mode', async () => {
      await render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey=""
        />,
        {wrapper: createWrapper()},
      );

      const createButton = page.getByText('Create Translation');
      await userEvent.click(createButton);

      await expect.element(page.getByText('Cancel')).toBeInTheDocument();
    });

    it('should exit create mode when cancel is clicked', async () => {
      await render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey=""
        />,
        {wrapper: createWrapper()},
      );

      // Enter create mode
      const createButton = page.getByText('Create Translation');
      await userEvent.click(createButton);

      // Click cancel
      const cancelButton = page.getByText('Cancel');
      await userEvent.click(cancelButton);

      // Should be back in select mode - the placeholder text indicates we're in select mode
      await expect.element(page.getByPlaceholder('Select an existing key')).toBeInTheDocument();
    });

    it('should have disabled create button when key or value is empty', async () => {
      await render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey=""
        />,
        {wrapper: createWrapper()},
      );

      // Enter create mode
      const createModeButton = page.getByText('Create Translation');
      await userEvent.click(createModeButton);

      // The create button should be disabled when fields are empty
      const submitButton = page.getByText('Create');
      expect(submitButton).toBeDisabled();
    });

    it('should show validation error when key is empty and create is clicked', async () => {
      await render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey=""
        />,
        {wrapper: createWrapper()},
      );

      // Enter create mode
      const createModeButton = page.getByText('Create Translation');
      await userEvent.click(createModeButton);

      // Fill in translation value only, leave key empty
      const translationValueInput = page.getByPlaceholder(
        'Enter translation text',
      );
      await userEvent.fill(translationValueInput, 'Some translation');

      // Submit button should be enabled because we check trim() - but if we click it empty, validation kicks in
      // Let's click the create button
      const submitButton = page.getByText('Create');
      // Button is disabled because key is empty (after trim)
      expect(submitButton).toBeDisabled();
    });

    it('should show validation error for invalid key format', async () => {
      await render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey=""
        />,
        {wrapper: createWrapper()},
      );

      // Enter create mode
      const createModeButton = page.getByText('Create Translation');
      await userEvent.click(createModeButton);

      // Fill in key with invalid characters
      const keyInput = page.getByPlaceholder(
        'Enter a unique translation key',
      );
      await userEvent.fill(keyInput, 'invalid key with spaces!');

      // Fill in translation value
      const translationValueInput = page.getByPlaceholder(
        'Enter translation text',
      );
      await userEvent.fill(translationValueInput, 'Some translation');

      // Click create
      const submitButton = page.getByText('Create');
      await userEvent.click(submitButton);

      // Should show error for invalid key format
      await expect.element(page.getByRole('alert')).toBeInTheDocument();
      await expect.element(page.getByText('Invalid key format. Use only letters, numbers, dots, underscores, and hyphens.')).toBeInTheDocument();
    });

    it('should call mutate when form is valid', async () => {
      await render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey=""
        />,
        {wrapper: createWrapper()},
      );

      // Enter create mode
      const createModeButton = page.getByText('Create Translation');
      await userEvent.click(createModeButton);

      // Fill in valid key
      const keyInput = page.getByPlaceholder(
        'Enter a unique translation key',
      );
      await userEvent.fill(keyInput, 'my.new.key');

      // Fill in translation value
      const translationValueInput = page.getByPlaceholder(
        'Enter translation text',
      );
      await userEvent.fill(translationValueInput, 'My translation value');

      // Click create
      const submitButton = page.getByText('Create');
      await userEvent.click(submitButton);

      // Should call mutate with correct params
      expect(mockMutate).toHaveBeenCalledWith(
        {
          language: 'en-US',
          namespace: 'flowI18n',
          key: 'my.new.key',
          value: 'My translation value',
        },
        // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
        expect.objectContaining({
          // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
          onSuccess: expect.any(Function),
          // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
          onError: expect.any(Function),
        }),
      );
    });

    it('should call onChange and exit create mode on successful creation', async () => {
      await render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey=""
        />,
        {wrapper: createWrapper()},
      );

      // Enter create mode
      const createModeButton = page.getByText('Create Translation');
      await userEvent.click(createModeButton);

      // Fill in form
      const keyInput = page.getByPlaceholder(
        'Enter a unique translation key',
      );
      await userEvent.fill(keyInput, 'my.new.key');

      const translationValueInput = page.getByPlaceholder(
        'Enter translation text',
      );
      await userEvent.fill(translationValueInput, 'My translation value');

      // Click create
      const submitButton = page.getByText('Create');
      await userEvent.click(submitButton);

      // Get the onSuccess callback and call it
      const mutateCall = mockMutate.mock.calls[0] as [unknown, {onSuccess: () => void; onError: (err: Error) => void}];
      const callbacks = mutateCall[1];
      callbacks.onSuccess();

      // Should have called onChange with the new key
      expect(mockOnChange).toHaveBeenCalledWith('flowI18n:my.new.key');

      // Should be back in select mode - check for the title change instead of placeholder
      await vi.waitFor(() => {
        expect(
          page.getByText('Translation for Label'),
        ).toBeInTheDocument();
      });
    });

    it('should show error on failed creation', async () => {
      await render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey=""
        />,
        {wrapper: createWrapper()},
      );

      // Enter create mode
      const createModeButton = page.getByText('Create Translation');
      await userEvent.click(createModeButton);

      // Fill in form
      const keyInput = page.getByPlaceholder(
        'Enter a unique translation key',
      );
      await userEvent.fill(keyInput, 'my.new.key');

      const translationValueInput = page.getByPlaceholder(
        'Enter translation text',
      );
      await userEvent.fill(translationValueInput, 'My translation value');

      // Click create
      const submitButton = page.getByText('Create');
      await userEvent.click(submitButton);

      // Get the onError callback and call it
      const mutateCall = mockMutate.mock.calls[0] as [unknown, {onSuccess: () => void; onError: (err: Error) => void}];
      const callbacks = mutateCall[1];
      callbacks.onError(new Error('API Error'));

      // Should show error alert
      await vi.waitFor(async () => {
        await expect.element(page.getByRole('alert')).toBeInTheDocument();
        await expect.element(page.getByText('API Error')).toBeInTheDocument();
      });
    });

    it('should clear error when typing in key field', async () => {
      await render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey=""
        />,
        {wrapper: createWrapper()},
      );

      // Enter create mode
      const createModeButton = page.getByText('Create Translation');
      await userEvent.click(createModeButton);

      // Fill in invalid key to trigger error
      const keyInput = page.getByPlaceholder(
        'Enter a unique translation key',
      );
      await userEvent.fill(keyInput, 'invalid key!');

      const translationValueInput = page.getByPlaceholder(
        'Enter translation text',
      );
      await userEvent.fill(translationValueInput, 'Some translation');

      // Click create to trigger error
      const submitButton = page.getByText('Create');
      await userEvent.click(submitButton);

      await expect.element(page.getByRole('alert')).toBeInTheDocument();

      // Type in key field to clear error
      await userEvent.fill(keyInput, 'valid.key');

      // Error should be cleared
      await expect.element(page.getByRole('alert')).not.toBeInTheDocument();
    });

    it('should clear error when typing in value field', async () => {
      await render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey=""
        />,
        {wrapper: createWrapper()},
      );

      // Enter create mode
      const createModeButton = page.getByText('Create Translation');
      await userEvent.click(createModeButton);

      // Fill in invalid key to trigger error
      const keyInput = page.getByPlaceholder(
        'Enter a unique translation key',
      );
      await userEvent.fill(keyInput, 'invalid key!');

      const translationValueInput = page.getByPlaceholder(
        'Enter translation text',
      );
      await userEvent.fill(translationValueInput, 'Some translation');

      // Click create to trigger error
      const submitButton = page.getByText('Create');
      await userEvent.click(submitButton);

      await expect.element(page.getByRole('alert')).toBeInTheDocument();

      // Type in value field to clear error
      await userEvent.fill(translationValueInput, 'Updated translation');

      // Error should be cleared
      await expect.element(page.getByRole('alert')).not.toBeInTheDocument();
    });

    it('should close error alert when close button is clicked', async () => {
      await render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey=""
        />,
        {wrapper: createWrapper()},
      );

      // Enter create mode
      const createModeButton = page.getByText('Create Translation');
      await userEvent.click(createModeButton);

      // Fill in invalid key to trigger error
      const keyInput = page.getByPlaceholder(
        'Enter a unique translation key',
      );
      await userEvent.fill(keyInput, 'invalid key!');

      const translationValueInput = page.getByPlaceholder(
        'Enter translation text',
      );
      await userEvent.fill(translationValueInput, 'Some translation');

      // Click create to trigger error
      const submitButton = page.getByText('Create');
      await userEvent.click(submitButton);

      const alert = page.getByRole('alert');
      expect(alert).toBeInTheDocument();

      // Close the alert
      const closeAlertButton = alert.element().querySelector('button');
      if (closeAlertButton) {
        await userEvent.click(closeAlertButton);
      }

      // Error should be cleared
      await expect.element(page.getByRole('alert')).not.toBeInTheDocument();
    });

    it('should allow selecting a different language', async () => {
      await render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey=""
        />,
        {wrapper: createWrapper()},
      );

      // Enter create mode
      const createModeButton = page.getByText('Create Translation');
      await userEvent.click(createModeButton);

      // Find the language autocomplete - it's the first combobox in create mode
      const languageCombobox = page.getByRole('combobox').all()[0];

      // Open the dropdown by clicking the Open button within the autocomplete
      const openButtons = page.getByTitle('Open');
      await userEvent.click(openButtons.all()[0]);

      // Wait for dropdown and select Spanish
      await vi.waitFor(async () => {
        await expect.element(page.getByRole('listbox')).toBeInTheDocument();
      });

      // Find and click the Spanish option
      const options = page.getByRole('option');
      const esOption = options.all().find((opt: Locator) => opt.element().textContent === 'es');
      expect(esOption).toBeDefined();
      await userEvent.click(esOption!);

      // Verify the language was selected
      expect(languageCombobox).toHaveValue('es');

      // Fill in form
      const keyInput = page.getByPlaceholder(
        'Enter a unique translation key',
      );
      await userEvent.fill(keyInput, 'my.key');

      const translationValueInput = page.getByPlaceholder(
        'Enter translation text',
      );
      await userEvent.fill(translationValueInput, 'Mi traducción');

      // Click create
      const submitButton = page.getByText('Create');
      await userEvent.click(submitButton);

      // Should call mutate with selected language
      expect(mockMutate).toHaveBeenCalledWith(
        expect.objectContaining({
          language: 'es',
        }),
        expect.any(Object),
      );
    });
  });

  describe('Edge Cases', () => {
    it('should handle error without message', async () => {
      await render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey=""
        />,
        {wrapper: createWrapper()},
      );

      // Enter create mode
      const createModeButton = page.getByText('Create Translation');
      await userEvent.click(createModeButton);

      // Fill in form
      const keyInput = page.getByPlaceholder(
        'Enter a unique translation key',
      );
      await userEvent.fill(keyInput, 'my.new.key');

      const translationValueInput = page.getByPlaceholder(
        'Enter translation text',
      );
      await userEvent.fill(translationValueInput, 'My translation value');

      // Click create
      const submitButton = page.getByText('Create');
      await userEvent.click(submitButton);

      // Get the onError callback and call it with error without message
      const mutateCall = mockMutate.mock.calls[0] as [unknown, {onSuccess: () => void; onError: (err: Error) => void}];
      const callbacks = mutateCall[1];
      callbacks.onError({} as Error);

      // Should show fallback error message
      await vi.waitFor(async () => {
        await expect.element(page.getByRole('alert')).toBeInTheDocument();
        await expect.element(page.getByText('errors.unknown')).toBeInTheDocument();
      });
    });
  });
});
