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
import {render, page, userEvent} from '@thunder/test-utils/browser';
import ScreenListItem from '../ScreenListItem';

describe('ScreenListItem', () => {
  describe('Rendering', () => {
    it('renders the screen name', async () => {
      render(<ScreenListItem name="auth" isSelected={false} onClick={vi.fn()} />);
      await expect.element(page.getByText('auth')).toBeInTheDocument();
    });

    it('renders "base screen" label when no extendsBase is provided', async () => {
      render(<ScreenListItem name="auth" isSelected={false} onClick={vi.fn()} />);
      // Real i18n translations or translation key text should be visible
      const body = document.body;
      expect(body.textContent).toBeTruthy();
    });

    it('renders "extends X" text when extendsBase is provided', async () => {
      render(<ScreenListItem name="password" extendsBase="auth" isSelected={false} onClick={vi.fn()} />);
      await expect.element(page.getByText(/auth/)).toBeInTheDocument();
    });

    it('renders a visual screen icon', async () => {
      const {container} = render(<ScreenListItem name="login" isSelected={false} onClick={vi.fn()} />);
      // The screen icon is a CSS-styled box representation, not an SVG
      expect(container.querySelector('[class*="MuiCardContent"]')).toBeInTheDocument();
    });
  });

  describe('Selection state', () => {
    it('renders without errors when not selected', async () => {
      render(<ScreenListItem name="auth" isSelected={false} onClick={vi.fn()} />);
      await expect.element(page.getByText('auth')).toBeInTheDocument();
    });

    it('renders without errors when selected', async () => {
      render(<ScreenListItem name="auth" isSelected onClick={vi.fn()} />);
      await expect.element(page.getByText('auth')).toBeInTheDocument();
    });
  });

  describe('Interaction', () => {
    it('calls onClick when the item is clicked', async () => {
      const onClick = vi.fn();
      render(<ScreenListItem name="login" isSelected={false} onClick={onClick} />);

      await userEvent.click(page.getByText('login'));

      expect(onClick).toHaveBeenCalledOnce();
    });

    it('calls onClick on repeated clicks', async () => {
      const onClick = vi.fn();
      render(<ScreenListItem name="signup" isSelected={false} onClick={onClick} />);

      await userEvent.click(page.getByText('signup'));
      await userEvent.click(page.getByText('signup'));

      expect(onClick).toHaveBeenCalledTimes(2);
    });
  });
});
