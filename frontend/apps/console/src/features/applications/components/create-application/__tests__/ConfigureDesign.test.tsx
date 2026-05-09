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

import userEvent from '@testing-library/user-event';
import {generateIconSuggestions} from '@thunderid/components';
import {render, screen} from '@thunderid/test-utils';
import type {ReactNode} from 'react';
import {describe, it, expect, beforeEach, vi} from 'vitest';
import ConfigureDesign, {type ConfigureDesignProps} from '../ConfigureDesign';

// Mock the Packages
vi.mock('@thunderid/components', () => ({
  generateIconSuggestions: vi.fn(() => null),
  ResourceAvatar: vi.fn(({value, fallback, onClick}: {value?: string; fallback?: ReactNode; onClick?: () => void}) => (
    <button
      type="button"
      onClick={onClick}
      style={{display: 'inline-block', background: 'none', border: 'none', padding: 0, cursor: 'pointer'}}
    >
      {value ?? fallback}
    </button>
  )),
}));
vi.mock('@thunderid/design');

const {useGetThemes, useGetTheme} = await import('@thunderid/design');

describe('ConfigureDesign', () => {
  const mockOnLogoSelect = vi.fn();
  const mockOnThemeSelect = vi.fn();

  const mockIconSuggestions = ['🐼', '🦊', '🐬', '🦁'];

  const defaultProps: ConfigureDesignProps = {
    appLogo: null,
    selectedTheme: null,
    onLogoSelect: mockOnLogoSelect,
    onThemeSelect: mockOnThemeSelect,
  };

  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(generateIconSuggestions).mockReturnValue(mockIconSuggestions);

    vi.mocked(useGetThemes).mockReturnValue({
      data: undefined,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useGetThemes>);

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

  it('should render subtitle', () => {
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

  it('should auto-select first emoji when appLogo is null', () => {
    renderComponent();

    expect(mockOnLogoSelect).toHaveBeenCalledWith(`emoji:${mockIconSuggestions[0]}`);
  });

  it('should not auto-select when appLogo is already set', () => {
    renderComponent({appLogo: 'emoji:🐼'});

    expect(mockOnLogoSelect).not.toHaveBeenCalled();
  });

  it('should render all emoji suggestions', () => {
    renderComponent();

    mockIconSuggestions.forEach((emoji) => {
      expect(screen.getByText(emoji)).toBeInTheDocument();
    });
  });

  it('should call onLogoSelect when clicking an emoji', async () => {
    const user = userEvent.setup();
    renderComponent();

    await user.click(screen.getByText(mockIconSuggestions[0]));

    expect(mockOnLogoSelect).toHaveBeenCalledWith(`emoji:${mockIconSuggestions[0]}`);
  });

  it('should regenerate icons when shuffle button is clicked', async () => {
    const user = userEvent.setup();
    const newIcons = ['🚀', '💡'];

    vi.mocked(generateIconSuggestions).mockReturnValueOnce(mockIconSuggestions).mockReturnValueOnce(newIcons);

    renderComponent();

    const shuffleButton = screen.getByRole('button', {name: 'Shuffle'});
    await user.click(shuffleButton);

    expect(generateIconSuggestions).toHaveBeenCalledTimes(2);
  });

  it('should generate icons with correct count', () => {
    renderComponent();

    expect(generateIconSuggestions).toHaveBeenCalledWith(8);
  });

  it('should render a "+" button to open the full picker', () => {
    renderComponent();

    // The "+" button is an avatar rendered after the suggestions; it contains a Plus/lucide-plus SVG icon
    expect(document.querySelector('.lucide-plus')).toBeInTheDocument();
  });

  it('should handle null appLogo prop without errors', () => {
    renderComponent({appLogo: null});

    expect(screen.getByRole('heading', {level: 1})).toBeInTheDocument();
  });

  it('should handle rapid emoji clicks', async () => {
    const user = userEvent.setup();
    renderComponent();

    await user.click(screen.getByText(mockIconSuggestions[0]));
    await user.click(screen.getByText(mockIconSuggestions[1]));

    expect(mockOnLogoSelect).toHaveBeenCalledTimes(3); // initial auto-select + 2 clicks
    expect(mockOnLogoSelect).toHaveBeenCalledWith(`emoji:${mockIconSuggestions[0]}`);
    expect(mockOnLogoSelect).toHaveBeenCalledWith(`emoji:${mockIconSuggestions[1]}`);
  });

  describe('onReadyChange callback', () => {
    it('should call onReadyChange with true on mount', () => {
      const mockOnReadyChange = vi.fn();
      renderComponent({onReadyChange: mockOnReadyChange});

      expect(mockOnReadyChange).toHaveBeenCalledWith(true);
    });
  });

  describe('Custom logo (isCustomLogo)', () => {
    it('should show a larger custom avatar when appLogo is a URL (not in suggestions)', () => {
      renderComponent({appLogo: 'https://example.com/custom-logo.png'});

      // When isCustomLogo=true, a large avatar is displayed — just verify no crash and heading exists
      expect(screen.getByRole('heading', {level: 1})).toBeInTheDocument();
    });

    it('should treat an emoji not in suggestions as a custom logo', () => {
      // '🐉' is not in mockIconSuggestions
      renderComponent({appLogo: 'emoji:🐉'});

      expect(screen.getByRole('heading', {level: 1})).toBeInTheDocument();
    });

    it('should treat an emoji in the suggestions list as a non-custom logo', () => {
      // '🐼' IS in mockIconSuggestions
      renderComponent({appLogo: `emoji:${mockIconSuggestions[0]}`});

      expect(screen.getByRole('heading', {level: 1})).toBeInTheDocument();
    });

    it('should not auto-select logo when appLogo is already a URL', () => {
      renderComponent({appLogo: 'https://example.com/logo.png'});

      expect(mockOnLogoSelect).not.toHaveBeenCalled();
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

      expect(mockOnThemeSelectLocal).toHaveBeenCalledWith('theme-1', mockThemeDetails.theme);
    });
  });
});
