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

import {describe, it, expect, vi} from 'vitest';
import {render, screen} from '@testing-library/react';

// Mock i18next top-level await before importing withI18n
vi.mock('i18next', () => ({
  default: {
    use: vi.fn().mockReturnThis(),
    init: vi.fn().mockResolvedValue(undefined),
  },
}));

vi.mock('react-i18next', () => ({
  initReactI18next: {},
  useTranslation: () => ({t: (key: string) => key, i18n: {language: 'en-US', changeLanguage: vi.fn()}}),
}));

vi.mock('@thunder/i18n/locales/en-US', () => ({
  default: {common: {}, navigation: {}},
}));

vi.mock('../../i18n/I18nProvider', () => ({
  default: ({children}: {children: React.ReactNode}) => <div data-testid="i18n-provider">{children}</div>,
}));

describe('withI18n (console)', () => {
  it('renders without crashing', async () => {
    const {default: withI18n} = await import('../withI18n');
    function MockChild() {
      return <div data-testid="mock-child">Child</div>;
    }
    const WithI18nComponent = withI18n(MockChild);

    const {container} = render(<WithI18nComponent />);
    expect(container).toBeInTheDocument();
  });

  it('renders the wrapped component', async () => {
    const {default: withI18n} = await import('../withI18n');
    function MockChild() {
      return <div data-testid="mock-child">Child</div>;
    }
    const WithI18nComponent = withI18n(MockChild);

    render(<WithI18nComponent />);
    expect(screen.getByTestId('mock-child')).toBeInTheDocument();
  });

  it('wraps with I18nProvider', async () => {
    const {default: withI18n} = await import('../withI18n');
    function MockChild() {
      return <div data-testid="mock-child">Child</div>;
    }
    const WithI18nComponent = withI18n(MockChild);

    render(<WithI18nComponent />);
    expect(screen.getByTestId('i18n-provider')).toBeInTheDocument();
  });

  it('places the wrapped component inside I18nProvider', async () => {
    const {default: withI18n} = await import('../withI18n');
    function MockChild() {
      return <div data-testid="mock-child">Child</div>;
    }
    const WithI18nComponent = withI18n(MockChild);

    render(<WithI18nComponent />);
    const provider = screen.getByTestId('i18n-provider');
    const child = screen.getByTestId('mock-child');
    expect(provider).toContainElement(child);
  });

  it('wraps different components correctly', async () => {
    const {default: withI18n} = await import('../withI18n');
    function AnotherChild() {
      return <div data-testid="another-child">Another</div>;
    }
    const AnotherWrapped = withI18n(AnotherChild);

    render(<AnotherWrapped />);
    expect(screen.getByTestId('another-child')).toBeInTheDocument();
    expect(screen.getByTestId('i18n-provider')).toBeInTheDocument();
  });
});
