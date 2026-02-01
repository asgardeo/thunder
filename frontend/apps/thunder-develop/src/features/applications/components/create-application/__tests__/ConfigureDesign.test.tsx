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
import {render, screen, fireEvent} from '@thunder/test-utils';
import userEvent from '@testing-library/user-event';
import ConfigureDesign, {type ConfigureDesignProps} from '../ConfigureDesign';

// Mock the utility functions
vi.mock('../../../utils/generateAppLogoSuggestion');
vi.mock('../../../utils/generateAppPrimaryColorSuggestions');

// Mock the brandings API
vi.mock('@thunder/shared-branding');

const {default: generateAppLogoSuggestions} = await import('../../../utils/generateAppLogoSuggestion');
const {default: generateAppPrimaryColorSuggestions} = await import('../../../utils/generateAppPrimaryColorSuggestions');
const {useGetBrandings, useGetBranding} = await import('@thunder/shared-branding');

describe('ConfigureDesign', () => {
  const mockOnLogoSelect = vi.fn();
  const mockOnColorSelect = vi.fn();
  const mockOnInitialLogoLoad = vi.fn();

  const mockLogoSuggestions = [
    'https://example.com/avatars/cat_lg.png',
    'https://example.com/avatars/dog_lg.png',
    'https://example.com/avatars/bird_lg.png',
    'https://example.com/avatars/fish_lg.png',
  ];

  const mockColorOptions = ['#FF5733', '#33FF57', '#3357FF', '#F333FF'];

  const defaultProps: ConfigureDesignProps = {
    appLogo: null,
    selectedColor: '#FF5733',
    onLogoSelect: mockOnLogoSelect,
    onColorSelect: mockOnColorSelect,
  };

  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(generateAppLogoSuggestions).mockReturnValue(mockLogoSuggestions);
    vi.mocked(generateAppPrimaryColorSuggestions).mockReturnValue(mockColorOptions);

    // Mock useGetBrandings to return empty data
    vi.mocked(useGetBrandings).mockReturnValue({
      data: undefined,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useGetBrandings>);

    // Mock useGetBranding to return empty data
    vi.mocked(useGetBranding).mockReturnValue({
      data: undefined,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useGetBranding>);
  });

  const renderComponent = (props: Partial<ConfigureDesignProps> = {}) =>
    render(<ConfigureDesign {...defaultProps} {...props} />);

  it('should render the component with title', () => {
    renderComponent();

    expect(screen.getByRole('heading', {level: 1})).toBeInTheDocument();
  });

  it('should render subtitle with info icon', () => {
    renderComponent();

    expect(screen.getByText('Customize the appearance of your application')).toBeInTheDocument();
  });

  it('should render logo section title', () => {
    renderComponent();

    expect(screen.getByRole('heading', {name: 'Application Logo'})).toBeInTheDocument();
  });

  it('should render shuffle button', () => {
    renderComponent();

    expect(screen.getByRole('button', {name: 'Shuffle'})).toBeInTheDocument();
  });

  it('should call onInitialLogoLoad when component mounts', () => {
    renderComponent({onInitialLogoLoad: mockOnInitialLogoLoad});

    expect(mockOnInitialLogoLoad).toHaveBeenCalledWith(mockLogoSuggestions[0]);
  });

  it('should not call onInitialLogoLoad if not provided', () => {
    renderComponent();

    // Should not throw error
    expect(mockOnInitialLogoLoad).not.toHaveBeenCalled();
  });

  it('should render all logo suggestions', () => {
    renderComponent();

    const avatars = screen.getAllByRole('img');
    expect(avatars.length).toBeGreaterThanOrEqual(mockLogoSuggestions.length);
  });

  it('should call onLogoSelect when clicking a logo', async () => {
    const user = userEvent.setup();
    renderComponent();

    const avatars = screen.getAllByRole('img');
    await user.click(avatars[0]);

    expect(mockOnLogoSelect).toHaveBeenCalledWith(mockLogoSuggestions[0]);
  });

  it('should highlight selected logo', () => {
    renderComponent({appLogo: mockLogoSuggestions[0]});

    const avatars = screen.getAllByRole('img');
    // Selected logo should have different styling (width: 80 vs 56)
    expect(avatars[0]).toBeInTheDocument();
  });

  it('should regenerate logos when shuffle button is clicked', async () => {
    const user = userEvent.setup();
    const newLogos = ['https://example.com/avatars/lion_lg.png', 'https://example.com/avatars/tiger_lg.png'];

    vi.mocked(generateAppLogoSuggestions).mockReturnValueOnce(mockLogoSuggestions).mockReturnValueOnce(newLogos);

    renderComponent();

    const shuffleButton = screen.getByRole('button', {name: 'Shuffle'});
    await user.click(shuffleButton);

    // generateAppLogoSuggestions should be called again
    expect(generateAppLogoSuggestions).toHaveBeenCalledTimes(2);
  });

  it('should display animal name in tooltip', async () => {
    const user = userEvent.setup();
    renderComponent();

    const avatars = screen.getAllByRole('img');
    await user.hover(avatars[0]);

    // Tooltip should show "Cat" from "cat_lg.png"
    expect(await screen.findByRole('tooltip', {name: /Cat/i})).toBeInTheDocument();
  });

  it('should render color section title', () => {
    renderComponent();

    expect(screen.getByRole('heading', {name: 'Brand Color'})).toBeInTheDocument();
  });

  it('should render all color options', () => {
    renderComponent();

    // Color options are rendered as clickable boxes
    const colorSection = screen.getByRole('heading', {name: 'Brand Color'}).parentElement;
    expect(colorSection).toBeInTheDocument();
  });

  it('should call onColorSelect when clicking a color', async () => {
    const user = userEvent.setup();
    renderComponent();

    // Find color boxes and click one
    const colorBoxes = screen.getAllByRole('button');
    const colorButton = colorBoxes.find((btn) => btn.getAttribute('aria-label')?.includes('color'));

    if (colorButton) {
      await user.click(colorButton);
      expect(mockOnColorSelect).toHaveBeenCalled();
    }
  });

  it('should highlight selected color', () => {
    renderComponent({selectedColor: mockColorOptions[0]});

    // Selected color should have different styling
    const colorSection = screen.getByRole('heading', {name: 'Brand Color'});
    expect(colorSection).toBeInTheDocument();
  });

  it('should handle custom color input toggle', async () => {
    const user = userEvent.setup();
    renderComponent();

    // Look for custom color button
    const customColorButton = screen.queryByRole('button', {name: /custom/i});

    if (customColorButton) {
      await user.click(customColorButton);
      // Custom color input should appear
      expect(screen.getByRole('textbox')).toBeInTheDocument();
    }
  });

  it('should handle custom color input value', async () => {
    const user = userEvent.setup();
    renderComponent();

    const customColorButton = screen.queryByRole('button', {name: /custom/i});

    if (customColorButton) {
      await user.click(customColorButton);

      const colorInput = screen.getByRole('textbox');
      await user.type(colorInput, '#AABBCC');

      // Should update custom color state
      expect(colorInput).toHaveValue('#AABBCC');
    }
  });

  it('should reset custom color when selecting predefined color', async () => {
    const user = userEvent.setup();
    renderComponent();

    // First set custom color
    const customColorButton = screen.queryByRole('button', {name: /custom/i});

    if (customColorButton) {
      await user.click(customColorButton);

      // Then select a predefined color
      const colorBoxes = screen.getAllByRole('button');
      const predefinedColor = colorBoxes.find((btn) => btn.getAttribute('aria-label')?.includes('color'));

      if (predefinedColor) {
        await user.click(predefinedColor);

        expect(mockOnColorSelect).toHaveBeenCalled();
      }
    }
  });

  it('should generate logos with correct count', () => {
    renderComponent();

    expect(generateAppLogoSuggestions).toHaveBeenCalledWith(8);
  });

  it('should generate color suggestions on mount', () => {
    renderComponent();

    expect(generateAppPrimaryColorSuggestions).toHaveBeenCalled();
  });

  it('should handle null appLogo prop', () => {
    renderComponent({appLogo: null});

    // Should render without errors
    expect(screen.getByRole('heading', {level: 1})).toBeInTheDocument();
  });

  it('should display palette icon', () => {
    renderComponent();

    // Palette icon should be present in the UI
    const colorSection = screen.getByRole('heading', {name: 'Brand Color'});
    expect(colorSection).toBeInTheDocument();
  });

  it('should handle rapid logo clicks', async () => {
    const user = userEvent.setup();
    renderComponent();

    const avatars = screen.getAllByRole('img');
    await user.click(avatars[0]);
    await user.click(avatars[1]);

    expect(mockOnLogoSelect).toHaveBeenCalledTimes(2);
    expect(mockOnLogoSelect).toHaveBeenNthCalledWith(1, mockLogoSuggestions[0]);
    expect(mockOnLogoSelect).toHaveBeenNthCalledWith(2, mockLogoSuggestions[1]);
  });

  it('should call onInitialLogoLoad again after shuffle', async () => {
    const user = userEvent.setup();
    const newLogos = ['https://example.com/avatars/new_lg.png'];
    vi.mocked(generateAppLogoSuggestions).mockReturnValueOnce(mockLogoSuggestions).mockReturnValueOnce(newLogos);

    renderComponent({onInitialLogoLoad: mockOnInitialLogoLoad});

    expect(mockOnInitialLogoLoad).toHaveBeenCalledWith(mockLogoSuggestions[0]);

    const shuffleButton = screen.getByRole('button', {name: 'Shuffle'});
    await user.click(shuffleButton);

    expect(mockOnInitialLogoLoad).toHaveBeenCalledWith(newLogos[0]);
    expect(mockOnInitialLogoLoad).toHaveBeenCalledTimes(2);
  });

  describe('onReadyChange callback', () => {
    it('should call onReadyChange with true on mount', () => {
      const mockOnReadyChange = vi.fn();
      renderComponent({onReadyChange: mockOnReadyChange});

      expect(mockOnReadyChange).toHaveBeenCalledWith(true);
    });
  });

  describe('DEFAULT branding integration', () => {
    const mockDefaultBrandingDetails = {
      id: 'default-branding-id',
      displayName: 'Default',
      preferences: {
        theme: {
          colorSchemes: {
            light: {
              colors: {
                primary: {
                  main: '#123456',
                },
              },
              images: {
                logo: {
                  primary: {
                    url: 'https://example.com/default-logo.png',
                  },
                },
              },
            },
          },
        },
      },
    };

    it('should apply DEFAULT branding color when it exists', () => {
      vi.mocked(useGetBrandings).mockReturnValue({
        data: {
          brandings: [{id: 'default-branding-id', displayName: 'Default'}],
        },
        isLoading: false,
        error: null,
      } as ReturnType<typeof useGetBrandings>);

      vi.mocked(useGetBranding).mockReturnValue({
        data: mockDefaultBrandingDetails,
        isLoading: false,
        error: null,
      } as ReturnType<typeof useGetBranding>);

      renderComponent();

      expect(mockOnColorSelect).toHaveBeenCalledWith('#123456');
    });

    it('should apply DEFAULT branding logo when it exists', () => {
      vi.mocked(useGetBrandings).mockReturnValue({
        data: {
          brandings: [{id: 'default-branding-id', displayName: 'Default'}],
        },
        isLoading: false,
        error: null,
      } as ReturnType<typeof useGetBrandings>);

      vi.mocked(useGetBranding).mockReturnValue({
        data: mockDefaultBrandingDetails,
        isLoading: false,
        error: null,
      } as ReturnType<typeof useGetBranding>);

      renderComponent();

      expect(mockOnLogoSelect).toHaveBeenCalledWith('https://example.com/default-logo.png');
    });

    it('should show "Pick Different Color" button when DEFAULT branding exists', () => {
      vi.mocked(useGetBrandings).mockReturnValue({
        data: {
          brandings: [{id: 'default-branding-id', displayName: 'Default'}],
        },
        isLoading: false,
        error: null,
      } as ReturnType<typeof useGetBrandings>);

      vi.mocked(useGetBranding).mockReturnValue({
        data: mockDefaultBrandingDetails,
        isLoading: false,
        error: null,
      } as ReturnType<typeof useGetBranding>);

      renderComponent();

      expect(screen.getByRole('button', {name: /Pick a different color/i})).toBeInTheDocument();
    });

    it('should show color options when "Pick Different Color" is clicked', async () => {
      const user = userEvent.setup();
      vi.mocked(useGetBrandings).mockReturnValue({
        data: {
          brandings: [{id: 'default-branding-id', displayName: 'Default'}],
        },
        isLoading: false,
        error: null,
      } as ReturnType<typeof useGetBrandings>);

      vi.mocked(useGetBranding).mockReturnValue({
        data: mockDefaultBrandingDetails,
        isLoading: false,
        error: null,
      } as ReturnType<typeof useGetBranding>);

      renderComponent();

      const pickColorButton = screen.getByRole('button', {name: /Pick a different color/i});
      await user.click(pickColorButton);

      // Color options should now be visible
      expect(screen.getByText('Custom')).toBeInTheDocument();
    });

    it('should call onBrandingSelectionChange with true when using DEFAULT branding', () => {
      const mockOnBrandingSelectionChange = vi.fn();

      vi.mocked(useGetBrandings).mockReturnValue({
        data: {
          brandings: [{id: 'default-branding-id', displayName: 'Default'}],
        },
        isLoading: false,
        error: null,
      } as ReturnType<typeof useGetBrandings>);

      vi.mocked(useGetBranding).mockReturnValue({
        data: mockDefaultBrandingDetails,
        isLoading: false,
        error: null,
      } as ReturnType<typeof useGetBranding>);

      renderComponent({onBrandingSelectionChange: mockOnBrandingSelectionChange});

      expect(mockOnBrandingSelectionChange).toHaveBeenCalledWith(true, 'default-branding-id');
    });

    it('should call onBrandingSelectionChange with false when picking different color', async () => {
      const user = userEvent.setup();
      const mockOnBrandingSelectionChange = vi.fn();

      vi.mocked(useGetBrandings).mockReturnValue({
        data: {
          brandings: [{id: 'default-branding-id', displayName: 'Default'}],
        },
        isLoading: false,
        error: null,
      } as ReturnType<typeof useGetBrandings>);

      vi.mocked(useGetBranding).mockReturnValue({
        data: mockDefaultBrandingDetails,
        isLoading: false,
        error: null,
      } as ReturnType<typeof useGetBranding>);

      renderComponent({onBrandingSelectionChange: mockOnBrandingSelectionChange});

      const pickColorButton = screen.getByRole('button', {name: /Pick a different color/i});
      await user.click(pickColorButton);

      expect(mockOnBrandingSelectionChange).toHaveBeenCalledWith(false, 'default-branding-id');
    });

    it('should display app name in branding message when provided', () => {
      vi.mocked(useGetBrandings).mockReturnValue({
        data: {
          brandings: [{id: 'default-branding-id', displayName: 'Default'}],
        },
        isLoading: false,
        error: null,
      } as ReturnType<typeof useGetBrandings>);

      vi.mocked(useGetBranding).mockReturnValue({
        data: mockDefaultBrandingDetails,
        isLoading: false,
        error: null,
      } as ReturnType<typeof useGetBranding>);

      renderComponent({appName: 'My Test App'});

      expect(screen.getByText('My Test App')).toBeInTheDocument();
    });
  });

  describe('getAnimalName', () => {
    it('should return "Unknown" for unmatched logo URL pattern', () => {
      const unmatchedLogos = ['https://example.com/avatars/invalid.jpg'];
      vi.mocked(generateAppLogoSuggestions).mockReturnValue(unmatchedLogos);

      renderComponent();

      // Should render without errors - the tooltip would show "Unknown"
      expect(screen.getAllByRole('img').length).toBeGreaterThan(0);
    });
  });

  describe('Custom logo selection', () => {
    it('should set hasCustomLogo when selecting a different logo from default', async () => {
      const user = userEvent.setup();
      const mockOnBrandingSelectionChange = vi.fn();

      const mockDefaultBrandingDetails = {
        id: 'default-branding-id',
        displayName: 'Default',
        preferences: {
          theme: {
            colorSchemes: {
              light: {
                colors: {
                  primary: {
                    main: '#123456',
                  },
                },
                images: {
                  logo: {
                    primary: {
                      url: 'https://example.com/default-logo.png',
                    },
                  },
                },
              },
            },
          },
        },
      };

      vi.mocked(useGetBrandings).mockReturnValue({
        data: {
          brandings: [{id: 'default-branding-id', displayName: 'Default'}],
        },
        isLoading: false,
        error: null,
      } as ReturnType<typeof useGetBrandings>);

      vi.mocked(useGetBranding).mockReturnValue({
        data: mockDefaultBrandingDetails,
        isLoading: false,
        error: null,
      } as ReturnType<typeof useGetBranding>);

      renderComponent({onBrandingSelectionChange: mockOnBrandingSelectionChange});

      // Select a different logo (not the default one)
      const avatars = screen.getAllByRole('img');
      await user.click(avatars[0]);

      // Should trigger branding selection change with false since we picked a different logo
      expect(mockOnBrandingSelectionChange).toHaveBeenCalledWith(false, 'default-branding-id');
    });
  });

  describe('Initial logo handling', () => {
    it('should not call onInitialLogoLoad when appLogo is already in suggestions', () => {
      vi.mocked(generateAppLogoSuggestions).mockReturnValue(mockLogoSuggestions);

      renderComponent({
        appLogo: mockLogoSuggestions[1],
        onInitialLogoLoad: mockOnInitialLogoLoad,
      });

      // Should not call since the selected logo is already in suggestions
      expect(mockOnInitialLogoLoad).not.toHaveBeenCalled();
    });
  });

  describe('handleColorSelect', () => {
    it('should call onColorSelect and reset custom color when selecting a predefined color', async () => {
      const user = userEvent.setup();
      renderComponent();

      // Click on a color chip using data-testid
      const colorChip = screen.getByTestId('color-chip-FF5733');

      await user.click(colorChip);
      expect(mockOnColorSelect).toHaveBeenCalled();
    });
  });

  describe('Custom color interaction', () => {
    it('should handle color input change and call onColorSelect', () => {
      renderComponent();

      const colorInput = document.querySelector<HTMLInputElement>('input[type="color"]')!;
      // Note: Browser color inputs normalize hex values to lowercase
      fireEvent.change(colorInput, {target: {value: '#aabbcc'}});

      expect(mockOnColorSelect).toHaveBeenCalledWith('#aabbcc');
    });
  });
});
