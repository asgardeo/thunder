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

import {render, screen, cleanup} from '@testing-library/react';
import {useContext} from 'react';
import {describe, it, expect, vi, beforeEach, afterEach} from 'vitest';
import type {DesignResolveResponse} from '../../../models/responses';
import DesignContext from '../DesignContext';
import DesignProvider from '../DesignProvider';

// Mock @thunder/contexts
const mockGetClientUuid = vi.fn();
vi.mock('@thunder/contexts', () => ({
  useConfig: () => ({
    getClientUuid: mockGetClientUuid,
  }),
}));

// Mock useGetDesignResolve
const mockUseGetDesignResolve = vi.fn();
vi.mock('../../../api/useGetDesignResolve', () => ({
  default: (...args: unknown[]) => mockUseGetDesignResolve(...args),
}));

function ContextConsumer() {
  const ctx = useContext(DesignContext);
  if (!ctx) return <div data-testid="no-context" />;
  return (
    <div
      data-testid="context-values"
      data-is-design-enabled={String(ctx.isDesignEnabled)}
      data-is-loading={String(ctx.isLoading)}
      data-has-error={String(Boolean(ctx.error))}
    />
  );
}

afterEach(() => {
  cleanup();
});

describe('DesignProvider', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockGetClientUuid.mockReturnValue('test-client-uuid');
    mockUseGetDesignResolve.mockReturnValue({
      data: undefined,
      isLoading: false,
      error: null,
    });
  });

  it('renders children', () => {
    render(
      <DesignProvider>
        <span>child content</span>
      </DesignProvider>,
    );
    expect(screen.getByText('child content')).toBeTruthy();
  });

  it('provides design context to children', () => {
    render(
      <DesignProvider>
        <ContextConsumer />
      </DesignProvider>,
    );
    expect(screen.getByTestId('context-values')).toBeTruthy();
  });

  describe('isLoading', () => {
    it('reflects the hook loading state when resolving internally', () => {
      mockUseGetDesignResolve.mockReturnValue({
        data: undefined,
        isLoading: true,
        error: null,
      });
      render(
        <DesignProvider>
          <ContextConsumer />
        </DesignProvider>,
      );
      expect(screen.getByTestId('context-values').getAttribute('data-is-loading')).toBe('true');
    });

    it('is false when externalDesign is provided', () => {
      const externalDesign: DesignResolveResponse = {
        theme: {colorSchemes: {light: {}}} as DesignResolveResponse['theme'],
        layout: {} as DesignResolveResponse['layout'],
      };
      render(
        <DesignProvider design={externalDesign}>
          <ContextConsumer />
        </DesignProvider>,
      );
      expect(screen.getByTestId('context-values').getAttribute('data-is-loading')).toBe('false');
    });

    it('is true when shouldResolveDesignInternally is false and no external design is provided', () => {
      render(
        <DesignProvider shouldResolveDesignInternally={false}>
          <ContextConsumer />
        </DesignProvider>,
      );
      expect(screen.getByTestId('context-values').getAttribute('data-is-loading')).toBe('true');
    });
  });

  describe('isDesignEnabled', () => {
    it('is false when no design is resolved', () => {
      render(
        <DesignProvider>
          <ContextConsumer />
        </DesignProvider>,
      );
      expect(screen.getByTestId('context-values').getAttribute('data-is-design-enabled')).toBe('false');
    });

    it('is true when external design has a non-empty theme', () => {
      const externalDesign: DesignResolveResponse = {
        theme: {colorSchemes: {light: {}}} as DesignResolveResponse['theme'],
        layout: {} as DesignResolveResponse['layout'],
      };
      render(
        <DesignProvider design={externalDesign}>
          <ContextConsumer />
        </DesignProvider>,
      );
      expect(screen.getByTestId('context-values').getAttribute('data-is-design-enabled')).toBe('true');
    });

    it('is true when resolved design has a non-empty theme', () => {
      mockUseGetDesignResolve.mockReturnValue({
        data: {
          theme: {colorSchemes: {light: {}}} as DesignResolveResponse['theme'],
          layout: {} as DesignResolveResponse['layout'],
        },
        isLoading: false,
        error: null,
      });
      render(
        <DesignProvider>
          <ContextConsumer />
        </DesignProvider>,
      );
      expect(screen.getByTestId('context-values').getAttribute('data-is-design-enabled')).toBe('true');
    });
  });

  describe('API call behavior', () => {
    it('enables the resolve query when clientUuid is available and shouldResolveDesignInternally is true', () => {
      mockGetClientUuid.mockReturnValue('valid-uuid');
      render(
        <DesignProvider>
          <span />
        </DesignProvider>,
      );
      expect(mockUseGetDesignResolve).toHaveBeenCalledWith(
        expect.objectContaining({id: 'valid-uuid'}),
        expect.objectContaining({enabled: true}),
      );
    });

    it('disables the resolve query when clientUuid is empty', () => {
      mockGetClientUuid.mockReturnValue('');
      render(
        <DesignProvider>
          <span />
        </DesignProvider>,
      );
      expect(mockUseGetDesignResolve).toHaveBeenCalledWith(
        expect.anything(),
        expect.objectContaining({enabled: false}),
      );
    });

    it('disables the resolve query when externalDesign is provided', () => {
      const externalDesign: DesignResolveResponse = {
        theme: {} as DesignResolveResponse['theme'],
        layout: {} as DesignResolveResponse['layout'],
      };
      render(
        <DesignProvider design={externalDesign}>
          <span />
        </DesignProvider>,
      );
      expect(mockUseGetDesignResolve).toHaveBeenCalledWith(
        expect.anything(),
        expect.objectContaining({enabled: false}),
      );
    });

    it('disables the resolve query when shouldResolveDesignInternally is false', () => {
      render(
        <DesignProvider shouldResolveDesignInternally={false}>
          <span />
        </DesignProvider>,
      );
      expect(mockUseGetDesignResolve).toHaveBeenCalledWith(
        expect.anything(),
        expect.objectContaining({enabled: false}),
      );
    });
  });
});
