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

import type {ThemeListItem} from '@thunderid/design';
import {render} from '@thunderid/test-utils';
import {describe, it, expect} from 'vitest';
import ThemeThumbnail from '../ThemeThumbnail';

const baseTheme: ThemeListItem = {
  id: 'theme-1',
  handle: 'classic',
  displayName: 'Classic',
};

describe('ThemeThumbnail', () => {
  describe('Rendering without crashing', () => {
    it('renders a light theme with primaryColor', () => {
      const {container} = render(
        <ThemeThumbnail theme={{...baseTheme, defaultColorScheme: 'light', primaryColor: '#ff7300'}} />,
      );
      expect(container.firstChild).toBeInTheDocument();
    });

    it('renders a dark theme with primaryColor', () => {
      const {container} = render(
        <ThemeThumbnail theme={{...baseTheme, defaultColorScheme: 'dark', primaryColor: '#bb86fc'}} />,
      );
      expect(container.firstChild).toBeInTheDocument();
    });

    it('renders when primaryColor is absent (uses name-based fallback)', () => {
      const {container} = render(<ThemeThumbnail theme={{...baseTheme, defaultColorScheme: 'light'}} />);
      expect(container.firstChild).toBeInTheDocument();
    });

    it('renders when defaultColorScheme is absent (defaults to light palette)', () => {
      const {container} = render(<ThemeThumbnail theme={{...baseTheme, primaryColor: '#ff7300'}} />);
      expect(container.firstChild).toBeInTheDocument();
    });

    it('renders with only the required id, handle, and displayName props', () => {
      const {container} = render(<ThemeThumbnail theme={baseTheme} />);
      expect(container.firstChild).toBeInTheDocument();
    });

    it('renders without crashing for an empty displayName', () => {
      const {container} = render(<ThemeThumbnail theme={{...baseTheme, displayName: ''}} />);
      expect(container.firstChild).toBeInTheDocument();
    });
  });

  describe('Distinct rendering per color scheme', () => {
    it('renders different DOM for light vs dark', () => {
      const {container: light} = render(
        <ThemeThumbnail theme={{...baseTheme, defaultColorScheme: 'light', primaryColor: '#ff7300'}} />,
      );
      const {container: dark} = render(
        <ThemeThumbnail theme={{...baseTheme, defaultColorScheme: 'dark', primaryColor: '#ff7300'}} />,
      );
      expect(light.innerHTML).not.toBe(dark.innerHTML);
    });

    it('renders different DOM for different primary colors', () => {
      const {container: orange} = render(
        <ThemeThumbnail theme={{...baseTheme, defaultColorScheme: 'light', primaryColor: '#ff7300'}} />,
      );
      const {container: blue} = render(
        <ThemeThumbnail theme={{...baseTheme, defaultColorScheme: 'light', primaryColor: '#007bff'}} />,
      );
      expect(orange.innerHTML).not.toBe(blue.innerHTML);
    });
  });

  describe('Fallback color generation', () => {
    it('generates consistent output for the same displayName across renders', () => {
      const theme: ThemeListItem = {...baseTheme, displayName: 'Ocean Blue'};
      const {container: first} = render(<ThemeThumbnail theme={theme} />);
      const {container: second} = render(<ThemeThumbnail theme={theme} />);
      expect(first.innerHTML).toBe(second.innerHTML);
    });

    it('generates different output for different displayNames', () => {
      const {container: a} = render(<ThemeThumbnail theme={{...baseTheme, displayName: 'Ocean Blue'}} />);
      const {container: b} = render(<ThemeThumbnail theme={{...baseTheme, displayName: 'Forest Green'}} />);
      expect(a.innerHTML).not.toBe(b.innerHTML);
    });
  });

  describe('Visual structure', () => {
    it('renders child elements inside the root container', () => {
      const {container} = render(
        <ThemeThumbnail theme={{...baseTheme, defaultColorScheme: 'light', primaryColor: '#ff7300'}} />,
      );
      expect(container.firstElementChild?.childElementCount).toBeGreaterThan(0);
    });
  });
});
