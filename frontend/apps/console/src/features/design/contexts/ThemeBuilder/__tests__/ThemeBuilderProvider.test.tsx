/**
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
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
import {render, screen, waitFor} from '@thunderid/test-utils';
import {describe, it, expect, vi, beforeEach} from 'vitest';
import ThemeBuilderProvider from '../ThemeBuilderProvider';
import useThemeBuilder from '../useThemeBuilder';

vi.mock('react-router', async () => {
  const actual = await vi.importActual<typeof import('react-router')>('react-router');
  return {
    ...actual,
    useParams: () => ({themeId: 'theme-123'}),
  };
});

const mockTheme = {
  defaultColorScheme: 'light' as const,
  colorSchemes: {
    light: {
      palette: {
        primary: {main: '#1a73e8'},
        secondary: {main: '#9c27b0'},
        background: {body: '#ffffff', surface: '#f5f5f5'},
        text: {primary: '#000000'},
      },
    },
    dark: {
      palette: {
        primary: {main: '#90caf9'},
        secondary: {main: '#ce93d8'},
        background: {body: '#121212', surface: '#1e1e1e'},
        text: {primary: '#ffffff'},
      },
    },
  },
  typography: {fontFamily: 'Inter'},
  shape: {borderRadius: 8},
  direction: 'ltr' as const,
};

const mockUseGetTheme = vi.fn();

vi.mock('@thunderid/design', () => ({
  useGetTheme: (...args: unknown[]) => mockUseGetTheme(...args) as unknown,
}));

/**
 * Helper consumer component that exposes context values as readable elements
 */
function TestConsumer() {
  const ctx = useThemeBuilder();
  return (
    <div>
      <span data-testid="themeId">{ctx.themeId}</span>
      <span data-testid="displayName">{ctx.displayName ?? 'null'}</span>
      <span data-testid="isDirty">{String(ctx.isDirty)}</span>
      <span data-testid="previewColorScheme">{ctx.previewColorScheme}</span>
      <span data-testid="activeSection">{ctx.activeSection}</span>
      <span data-testid="viewport">{ctx.viewport}</span>
      <span data-testid="draft-primary">{ctx.draftTheme?.colorSchemes?.light?.palette?.primary?.main ?? 'null'}</span>
      <button
        type="button"
        onClick={() => ctx.updateDraftTheme(['colorSchemes', 'light', 'palette', 'primary', 'main'], '#ff0000')}
      >
        UpdatePrimary
      </button>
      <button type="button" onClick={ctx.resetDraft}>
        Reset
      </button>
    </div>
  );
}

describe('ThemeBuilderProvider', () => {
  beforeEach(() => {
    mockUseGetTheme.mockReset();
  });

  describe('Loading state', () => {
    it('renders null while loading', () => {
      mockUseGetTheme.mockReturnValue({data: undefined, isLoading: true});
      const {container} = render(
        <ThemeBuilderProvider>
          <TestConsumer />
        </ThemeBuilderProvider>,
      );

      // When isLoading, provider returns null — nothing should be rendered
      expect(container).toBeEmptyDOMElement();
    });

    it('does not render children while loading', () => {
      mockUseGetTheme.mockReturnValue({data: undefined, isLoading: true});
      render(
        <ThemeBuilderProvider>
          <span data-testid="child">Child</span>
        </ThemeBuilderProvider>,
      );

      expect(screen.queryByTestId('child')).not.toBeInTheDocument();
    });
  });

  describe('Loaded state', () => {
    beforeEach(() => {
      mockUseGetTheme.mockReturnValue({
        data: {id: 'theme-123', displayName: 'Ocean Blue', theme: mockTheme},
        isLoading: false,
      });
    });

    it('renders children when not loading', () => {
      render(
        <ThemeBuilderProvider>
          <TestConsumer />
        </ThemeBuilderProvider>,
      );

      expect(screen.getByTestId('themeId')).toBeInTheDocument();
    });

    it('provides the themeId from route params', () => {
      render(
        <ThemeBuilderProvider>
          <TestConsumer />
        </ThemeBuilderProvider>,
      );

      expect(screen.getByTestId('themeId')).toHaveTextContent('theme-123');
    });

    it('provides the displayName from fetched data', () => {
      render(
        <ThemeBuilderProvider>
          <TestConsumer />
        </ThemeBuilderProvider>,
      );

      expect(screen.getByTestId('displayName')).toHaveTextContent('Ocean Blue');
    });

    it('initialises draftTheme from fetched theme data', () => {
      render(
        <ThemeBuilderProvider>
          <TestConsumer />
        </ThemeBuilderProvider>,
      );

      expect(screen.getByTestId('draft-primary')).toHaveTextContent('#1a73e8');
    });

    it('starts with isDirty=false', () => {
      render(
        <ThemeBuilderProvider>
          <TestConsumer />
        </ThemeBuilderProvider>,
      );

      expect(screen.getByTestId('isDirty')).toHaveTextContent('false');
    });

    it('starts with activeSection="colors"', () => {
      render(
        <ThemeBuilderProvider>
          <TestConsumer />
        </ThemeBuilderProvider>,
      );

      expect(screen.getByTestId('activeSection')).toHaveTextContent('colors');
    });

    it('starts with viewport="desktop"', () => {
      render(
        <ThemeBuilderProvider>
          <TestConsumer />
        </ThemeBuilderProvider>,
      );

      expect(screen.getByTestId('viewport')).toHaveTextContent('desktop');
    });
  });

  describe('previewColorScheme initialisation', () => {
    it('sets previewColorScheme to "light" when theme defaultColorScheme is "light"', () => {
      mockUseGetTheme.mockReturnValue({
        data: {id: 'theme-123', displayName: 'Light Theme', theme: {...mockTheme, defaultColorScheme: 'light'}},
        isLoading: false,
      });

      render(
        <ThemeBuilderProvider>
          <TestConsumer />
        </ThemeBuilderProvider>,
      );

      expect(screen.getByTestId('previewColorScheme')).toHaveTextContent('light');
    });

    it('sets previewColorScheme to "dark" when theme defaultColorScheme is "dark"', async () => {
      mockUseGetTheme.mockReturnValue({
        data: {
          id: 'theme-123',
          displayName: 'Dark Theme',
          theme: {...mockTheme, defaultColorScheme: 'dark' as const},
        },
        isLoading: false,
      });

      render(
        <ThemeBuilderProvider>
          <TestConsumer />
        </ThemeBuilderProvider>,
      );

      await waitFor(() => {
        expect(screen.getByTestId('previewColorScheme')).toHaveTextContent('dark');
      });
    });
  });

  describe('updateDraftTheme', () => {
    beforeEach(() => {
      mockUseGetTheme.mockReturnValue({
        data: {id: 'theme-123', displayName: 'Ocean Blue', theme: mockTheme},
        isLoading: false,
      });
    });

    it('updates the nested path in draftTheme', async () => {
      const user = userEvent.setup();
      render(
        <ThemeBuilderProvider>
          <TestConsumer />
        </ThemeBuilderProvider>,
      );

      await user.click(screen.getByText('UpdatePrimary'));

      await waitFor(() => {
        expect(screen.getByTestId('draft-primary')).toHaveTextContent('#ff0000');
      });
    });

    it('marks isDirty as true after an update', async () => {
      const user = userEvent.setup();
      render(
        <ThemeBuilderProvider>
          <TestConsumer />
        </ThemeBuilderProvider>,
      );

      await user.click(screen.getByText('UpdatePrimary'));

      await waitFor(() => {
        expect(screen.getByTestId('isDirty')).toHaveTextContent('true');
      });
    });
  });

  describe('resetDraft', () => {
    beforeEach(() => {
      mockUseGetTheme.mockReturnValue({
        data: {id: 'theme-123', displayName: 'Ocean Blue', theme: mockTheme},
        isLoading: false,
      });
    });

    it('reverts draftTheme to original on reset', async () => {
      const user = userEvent.setup();
      render(
        <ThemeBuilderProvider>
          <TestConsumer />
        </ThemeBuilderProvider>,
      );

      // First update
      await user.click(screen.getByText('UpdatePrimary'));
      await waitFor(() => expect(screen.getByTestId('draft-primary')).toHaveTextContent('#ff0000'));

      // Then reset
      await user.click(screen.getByText('Reset'));
      await waitFor(() => {
        expect(screen.getByTestId('draft-primary')).toHaveTextContent('#1a73e8');
      });
    });

    it('sets isDirty back to false after reset', async () => {
      const user = userEvent.setup();
      render(
        <ThemeBuilderProvider>
          <TestConsumer />
        </ThemeBuilderProvider>,
      );

      await user.click(screen.getByText('UpdatePrimary'));
      await waitFor(() => expect(screen.getByTestId('isDirty')).toHaveTextContent('true'));

      await user.click(screen.getByText('Reset'));
      await waitFor(() => {
        expect(screen.getByTestId('isDirty')).toHaveTextContent('false');
      });
    });
  });
});
