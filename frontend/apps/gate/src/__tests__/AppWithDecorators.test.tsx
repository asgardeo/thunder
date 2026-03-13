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
import {render, screen} from '@testing-library/react';
import AppWithDecorators from '../AppWithDecorators';

// Mock HOCs as pass-through so tests focus on composition
vi.mock('../hocs/withConfig', () => ({
  default: (Component: React.ComponentType) => Component,
}));
vi.mock('../hocs/withI18n', () => ({
  default: (Component: React.ComponentType) => Component,
}));
vi.mock('../hocs/withTheme', () => ({
  default: (Component: React.ComponentType) => Component,
}));

// Mock App
vi.mock('../App', () => ({
  default: () => <div data-testid="app">App</div>,
}));

// Mock i18next (used at module level in withI18n even though HOC is mocked)
vi.mock('i18next', () => ({
  default: {
    use: vi.fn().mockReturnThis(),
    init: vi.fn().mockResolvedValue(undefined),
  },
}));
vi.mock('react-i18next', () => ({
  initReactI18next: {},
}));
vi.mock('@thunder/i18n/locales/en-US', () => ({
  default: {common: {}, navigation: {}},
}));

describe('AppWithDecorators (gate)', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders without crashing', () => {
    const {container} = render(<AppWithDecorators />);
    expect(container).toBeInTheDocument();
  });

  it('renders the App component', () => {
    render(<AppWithDecorators />);
    expect(screen.getByTestId('app')).toBeInTheDocument();
  });
});
