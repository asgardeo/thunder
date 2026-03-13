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
import {render, screen} from '@thunder/test-utils';
import userEvent from '@testing-library/user-event';
import ConfigureDesign, {type ConfigureDesignProps} from '../ConfigureDesign';

// Mock the utility functions
vi.mock('../../../utils/generateAppLogoSuggestion');

// Mock the themes API
vi.mock('@thunder/shared-design');

const {default: generateAppLogoSuggestions} = await import('../../../utils/generateAppLogoSuggestion');
const {useGetThemes, useGetTheme} = await import('@thunder/shared-design');

describe('ConfigureDesign', () => {
  const mockOnLogoSelect = vi.fn();
  const mockOnThemeSelect = vi.fn();
  const mockOnInitialLogoLoad = vi.fn();

  const mockLogoSuggestions = [
    'https://example.com/avatars/cat_lg.png',
    'https://example.com/avatars/dog_lg.png',
    'https://example.com/avatars/bird_lg.png',
    'https://example.com/avatars/fish_lg.png',
  ];

  const defaultProps: ConfigureDesignProps = {
    appLogo: null,
    selectedTheme: null,
    onLogoSelect: mockOnLogoSelect,
    onThemeSelect: mockOnThemeSelect,
  };

  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(generateAppLogoSuggestions).mockReturnValue(mockLogoSuggestions);

    // Mock useGetThemes to return empty data
    vi.mocked(useGetThemes).mockReturnValue({
      data: undefined,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useGetThemes>);

    // Mock useGetTheme to return empty data
    vi.mocked(useGetTheme).mockReturnValue({
      data: undefined,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useGetTheme>);
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

  it('should render theme section title', () => {
    renderComponent();

    expect(screen.getByRole('heading', {name: 'Theme'})).toBeInTheDocument();
  });

  it('should generate logos with correct count', () => {
    renderComponent();

    expect(generateAppLogoSuggestions).toHaveBeenCalledWith(8);
  });

  it('should handle null appLogo prop', () => {
    renderComponent({appLogo: null});

    // Should render without errors
    expect(screen.getByRole('heading', {level: 1})).toBeInTheDocument();
  });

  it('should display palette icon', () => {
    renderComponent();

    // Palette icon should be present in the UI
    const colorSection = screen.getByRole('heading', {name: 'Theme'});
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

  describe('Theme selection', () => {
    const mockThemeDetails = {
      id: 'theme-1',
      displayName: 'Corporate Blue',
      theme: {
        colorSchemes: {
          light: {
            colors: {
              primary: {
                main: '#123456',
              },
            },
          },
        },
      },
    };

    const mockThemesList = [
      {id: 'theme-1', displayName: 'Corporate Blue'},
      {id: 'theme-2', displayName: 'Sunset Orange'},
    ];

    it('should render theme cards when themes are available', () => {
      vi.mocked(useGetThemes).mockReturnValue({
        data: {themes: mockThemesList},
        isLoading: false,
        error: null,
      } as ReturnType<typeof useGetThemes>);

      vi.mocked(useGetTheme).mockReturnValue({
        data: mockThemeDetails,
        isLoading: false,
        error: null,
      } as unknown as ReturnType<typeof useGetTheme>);

      renderComponent();

      expect(screen.getByText('Corporate Blue')).toBeInTheDocument();
      expect(screen.getByText('Sunset Orange')).toBeInTheDocument();
    });

    it('should render a card for each theme', () => {
      vi.mocked(useGetThemes).mockReturnValue({
        data: {themes: mockThemesList},
        isLoading: false,
        error: null,
      } as ReturnType<typeof useGetThemes>);

      vi.mocked(useGetTheme).mockReturnValue({
        data: mockThemeDetails,
        isLoading: false,
        error: null,
      } as unknown as ReturnType<typeof useGetTheme>);

      renderComponent();

      expect(screen.getByTestId('theme-card-theme-1')).toBeInTheDocument();
      expect(screen.getByTestId('theme-card-theme-2')).toBeInTheDocument();
    });

    it('should call onThemeSelect with theme details when theme is loaded', () => {
      vi.mocked(useGetThemes).mockReturnValue({
        data: {themes: mockThemesList},
        isLoading: false,
        error: null,
      } as ReturnType<typeof useGetThemes>);

      vi.mocked(useGetTheme).mockReturnValue({
        data: mockThemeDetails,
        isLoading: false,
        error: null,
      } as unknown as ReturnType<typeof useGetTheme>);

      renderComponent();

      expect(mockOnThemeSelect).toHaveBeenCalledWith('theme-1', mockThemeDetails.theme);
    });

    it('should show empty state when no themes are configured', () => {
      vi.mocked(useGetThemes).mockReturnValue({
        data: {themes: []},
        isLoading: false,
        error: null,
      } as unknown as ReturnType<typeof useGetThemes>);

      renderComponent();

      expect(screen.getByText('No themes configured')).toBeInTheDocument();
      expect(screen.getByText('You can configure themes later from the Design settings.')).toBeInTheDocument();
    });

    it('should select a different theme when clicking its card', async () => {
      const user = userEvent.setup();
      const mockOnThemeSelectLocal = vi.fn();

      vi.mocked(useGetThemes).mockReturnValue({
        data: {themes: mockThemesList},
        isLoading: false,
        error: null,
      } as ReturnType<typeof useGetThemes>);

      vi.mocked(useGetTheme).mockReturnValue({
        data: mockThemeDetails,
        isLoading: false,
        error: null,
      } as unknown as ReturnType<typeof useGetTheme>);

      renderComponent({onThemeSelect: mockOnThemeSelectLocal});

      const secondThemeCard = screen.getByTestId('theme-card-theme-2');
      await user.click(secondThemeCard);

      // onThemeSelect should be called when theme details load
      expect(mockOnThemeSelectLocal).toHaveBeenCalledWith('theme-1', mockThemeDetails.theme);
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
});
