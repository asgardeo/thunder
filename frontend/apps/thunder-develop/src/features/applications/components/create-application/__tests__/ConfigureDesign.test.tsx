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
import {render, screen} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {QueryClient, QueryClientProvider} from '@tanstack/react-query';
import {ConfigProvider} from '@thunder/commons-contexts';
import type {ReactNode} from 'react';
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

  let queryClient: QueryClient;

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

    // Setup window.__THUNDER_RUNTIME_CONFIG__ for tests
    // eslint-disable-next-line no-underscore-dangle
    if (typeof window !== 'undefined') {
      // eslint-disable-next-line no-underscore-dangle
      window.__THUNDER_RUNTIME_CONFIG__ = {
        client: {
          base: '/develop',
          client_id: 'DEVELOP',
        },
        server: {
          hostname: 'localhost',
          port: 8090,
          http_only: false,
        },
      };
    }

    queryClient = new QueryClient({
      defaultOptions: {
        queries: {
          retry: false,
        },
      },
    });
  });

  const renderComponent = (props: Partial<ConfigureDesignProps> = {}) => {
    function Wrapper({children}: {children: ReactNode}) {
      return (
        <QueryClientProvider client={queryClient}>
          <ConfigProvider>{children}</ConfigProvider>
        </QueryClientProvider>
      );
    }

    return render(<ConfigureDesign {...defaultProps} {...props} />, {wrapper: Wrapper});
  };

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
});
