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

import {describe, it, expect, vi, beforeEach} from 'vitest';
import {render, page, userEvent} from '@thunder/test-utils/browser';
import LayoutBuilderProvider from '../LayoutBuilderProvider';
import useLayoutBuilder from '../useLayoutBuilder';

vi.mock('react-router', async () => {
  const actual = await vi.importActual<typeof import('react-router')>('react-router');
  return {
    ...actual,
    useParams: () => ({layoutId: 'layout-123'}),
  };
});

const mockLayout = {
  screens: {
    auth: {
      background: {type: 'solid', color: '#ffffff'},
      slots: {},
    },
    login: {
      extends: 'auth',
      slots: {},
    },
  },
};

const mockUseGetLayout = vi.fn();

vi.mock('@thunder/shared-design', () => ({
  useGetLayout: (...args: unknown[]): unknown => mockUseGetLayout(...args),
}));

/**
 * Helper consumer component that exposes context values as readable elements
 */
function TestConsumer() {
  const ctx = useLayoutBuilder();
  const allScreens = ctx.getAllScreens();
  const baseNames = ctx.getBaseScreenNames();

  return (
    <div>
      <span data-testid="layoutId">{ctx.layoutId}</span>
      <span data-testid="displayName">{ctx.displayName ?? 'null'}</span>
      <span data-testid="isDirty">{String(ctx.isDirty)}</span>
      <span data-testid="selectedScreen">{ctx.selectedScreen ?? 'null'}</span>
      <span data-testid="allScreens">{Object.keys(allScreens).join(',')}</span>
      <span data-testid="baseScreenNames">{baseNames.join(',')}</span>
      <button type="button" onClick={() => ctx.addScreen('recovery', 'auth')}>
        AddScreen
      </button>
      <button
        type="button"
        onClick={() => ctx.updateDraftLayout(['screens', 'auth', 'background', 'color'], '#000000')}
      >
        UpdateBg
      </button>
      <button type="button" onClick={ctx.resetDraft}>
        Reset
      </button>
    </div>
  );
}

describe('LayoutBuilderProvider', () => {
  beforeEach(() => {
    mockUseGetLayout.mockReset();
  });

  describe('Loading state', () => {
    it('renders null while loading', async () => {
      mockUseGetLayout.mockReturnValue({data: undefined, isLoading: true});
      const {container} = render(
        <LayoutBuilderProvider>
          <TestConsumer />
        </LayoutBuilderProvider>,
      );

      expect(container).toBeEmptyDOMElement();
    });

    it('does not render children while loading', async () => {
      mockUseGetLayout.mockReturnValue({data: undefined, isLoading: true});
      render(
        <LayoutBuilderProvider>
          <span data-testid="child">Child</span>
        </LayoutBuilderProvider>,
      );

      await expect.element(page.getByTestId('child')).not.toBeInTheDocument();
    });
  });

  describe('Loaded state', () => {
    beforeEach(() => {
      mockUseGetLayout.mockReturnValue({
        data: {id: 'layout-123', displayName: 'Default Layout', layout: mockLayout},
        isLoading: false,
      });
    });

    it('renders children when not loading', async () => {
      render(
        <LayoutBuilderProvider>
          <TestConsumer />
        </LayoutBuilderProvider>,
      );

      await expect.element(page.getByTestId('layoutId')).toBeInTheDocument();
    });

    it('provides the layoutId from route params', async () => {
      render(
        <LayoutBuilderProvider>
          <TestConsumer />
        </LayoutBuilderProvider>,
      );

      await expect.element(page.getByTestId('layoutId')).toHaveTextContent('layout-123');
    });

    it('provides the displayName from fetched data', async () => {
      render(
        <LayoutBuilderProvider>
          <TestConsumer />
        </LayoutBuilderProvider>,
      );

      await expect.element(page.getByTestId('displayName')).toHaveTextContent('Default Layout');
    });

    it('starts with isDirty=false', async () => {
      render(
        <LayoutBuilderProvider>
          <TestConsumer />
        </LayoutBuilderProvider>,
      );

      await expect.element(page.getByTestId('isDirty')).toHaveTextContent('false');
    });

    it('auto-selects the first screen', async () => {
      render(
        <LayoutBuilderProvider>
          <TestConsumer />
        </LayoutBuilderProvider>,
      );

      await vi.waitFor(async () => {
        // First screen key in mockLayout.screens is 'auth'
        await expect.element(page.getByTestId('selectedScreen')).toHaveTextContent('auth');
      });
    });
  });

  describe('getAllScreens', () => {
    beforeEach(() => {
      mockUseGetLayout.mockReturnValue({
        data: {id: 'layout-123', displayName: 'Default Layout', layout: mockLayout},
        isLoading: false,
      });
    });

    it('returns all screens from the draft layout', async () => {
      render(
        <LayoutBuilderProvider>
          <TestConsumer />
        </LayoutBuilderProvider>,
      );

      await vi.waitFor(async () => {
        const el = await page.getByTestId('allScreens').element();
        const text = el.textContent ?? '';
        expect(text).toContain('auth');
        expect(text).toContain('login');
      });
    });
  });

  describe('getBaseScreenNames', () => {
    beforeEach(() => {
      mockUseGetLayout.mockReturnValue({
        data: {id: 'layout-123', displayName: 'Default Layout', layout: mockLayout},
        isLoading: false,
      });
    });

    it('returns only screens that do not have an extends property', async () => {
      render(
        <LayoutBuilderProvider>
          <TestConsumer />
        </LayoutBuilderProvider>,
      );

      await vi.waitFor(async () => {
        const el = await page.getByTestId('baseScreenNames').element();
        const text = el.textContent ?? '';
        expect(text).toContain('auth');
        expect(text).not.toContain('login');
      });
    });
  });

  describe('addScreen', () => {
    beforeEach(() => {
      mockUseGetLayout.mockReturnValue({
        data: {id: 'layout-123', displayName: 'Default Layout', layout: mockLayout},
        isLoading: false,
      });
    });

    it('adds a new screen to getAllScreens', async () => {
      render(
        <LayoutBuilderProvider>
          <TestConsumer />
        </LayoutBuilderProvider>,
      );

      await userEvent.click(page.getByText('AddScreen'));

      await vi.waitFor(async () => {
        const el = await page.getByTestId('allScreens').element();
        expect(el.textContent).toContain('recovery');
      });
    });

    it('selects the newly added screen', async () => {
      render(
        <LayoutBuilderProvider>
          <TestConsumer />
        </LayoutBuilderProvider>,
      );

      await userEvent.click(page.getByText('AddScreen'));

      await vi.waitFor(async () => {
        await expect.element(page.getByTestId('selectedScreen')).toHaveTextContent('recovery');
      });
    });

    it('marks isDirty after adding a screen', async () => {
      render(
        <LayoutBuilderProvider>
          <TestConsumer />
        </LayoutBuilderProvider>,
      );

      await userEvent.click(page.getByText('AddScreen'));

      await vi.waitFor(async () => {
        await expect.element(page.getByTestId('isDirty')).toHaveTextContent('true');
      });
    });

    it('new screen does not appear in base screen names (has extends)', async () => {
      render(
        <LayoutBuilderProvider>
          <TestConsumer />
        </LayoutBuilderProvider>,
      );

      await userEvent.click(page.getByText('AddScreen'));

      await vi.waitFor(async () => {
        const el = await page.getByTestId('baseScreenNames').element();
        expect(el.textContent).not.toContain('recovery');
      });
    });
  });

  describe('updateDraftLayout', () => {
    beforeEach(() => {
      mockUseGetLayout.mockReturnValue({
        data: {id: 'layout-123', displayName: 'Default Layout', layout: mockLayout},
        isLoading: false,
      });
    });

    it('marks isDirty as true after an update', async () => {
      render(
        <LayoutBuilderProvider>
          <TestConsumer />
        </LayoutBuilderProvider>,
      );

      await userEvent.click(page.getByText('UpdateBg'));

      await vi.waitFor(async () => {
        await expect.element(page.getByTestId('isDirty')).toHaveTextContent('true');
      });
    });
  });

  describe('resetDraft', () => {
    beforeEach(() => {
      mockUseGetLayout.mockReturnValue({
        data: {id: 'layout-123', displayName: 'Default Layout', layout: mockLayout},
        isLoading: false,
      });
    });

    it('clears isDirty after reset', async () => {
      render(
        <LayoutBuilderProvider>
          <TestConsumer />
        </LayoutBuilderProvider>,
      );

      await userEvent.click(page.getByText('UpdateBg'));
      await vi.waitFor(async () => expect.element(page.getByTestId('isDirty')).toHaveTextContent('true'));

      await userEvent.click(page.getByText('Reset'));
      await vi.waitFor(async () => {
        await expect.element(page.getByTestId('isDirty')).toHaveTextContent('false');
      });
    });

    it('clears extraScreens on reset', async () => {
      render(
        <LayoutBuilderProvider>
          <TestConsumer />
        </LayoutBuilderProvider>,
      );

      // Add a screen, then reset
      await userEvent.click(page.getByText('AddScreen'));
      await vi.waitFor(async () => {
        const el = await page.getByTestId('allScreens').element();
        expect(el.textContent).toContain('recovery');
      });

      await userEvent.click(page.getByText('Reset'));
      await vi.waitFor(async () => {
        const el = await page.getByTestId('allScreens').element();
        expect(el.textContent).not.toContain('recovery');
      });
    });
  });
});
