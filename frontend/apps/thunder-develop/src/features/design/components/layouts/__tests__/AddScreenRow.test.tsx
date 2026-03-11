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
import AddScreenRow from '../AddScreenRow';

const baseScreens = ['auth', 'login'];

describe('AddScreenRow', () => {
  describe('Initial state', () => {
    it('renders the "Add screen" trigger button', async () => {
      render(<AddScreenRow baseScreens={baseScreens} onAdd={vi.fn()} />);
      // The add trigger is visible (i18n key returned as-is)
      await expect.element(page.getByRole('button')).toBeInTheDocument();
    });

    it('does not show the text field before activation', async () => {
      render(<AddScreenRow baseScreens={baseScreens} onAdd={vi.fn()} />);
      await expect.element(page.getByRole('textbox')).not.toBeInTheDocument();
    });
  });

  describe('Expanded state', () => {
    it('shows the text input after clicking the add button', async () => {
      render(<AddScreenRow baseScreens={baseScreens} onAdd={vi.fn()} />);

      await userEvent.click(page.getByRole('button'));

      await expect.element(page.getByRole('textbox')).toBeInTheDocument();
    });

    it('shows a cancel button after expansion', async () => {
      render(<AddScreenRow baseScreens={baseScreens} onAdd={vi.fn()} />);

      await userEvent.click(page.getByRole('button'));

      // Expanded form should have at least 2 buttons (add + cancel)
      expect(page.getByRole('button').all().length).toBeGreaterThanOrEqual(2);
    });
  });

  describe('Adding a screen', () => {
    it('calls onAdd with the typed name and first baseScreen when confirmed', async () => {
      const onAdd = vi.fn();
      render(<AddScreenRow baseScreens={baseScreens} onAdd={onAdd} />);

      await userEvent.click(page.getByRole('button'));
      await userEvent.type(page.getByRole('textbox'), 'my-screen');

      // Click the confirm/add button (find by looking for non-cancel buttons)
      const buttons = page.getByRole('button').all();
      // The add/confirm button triggers onAdd — find the first non-cancel button
      const confirmBtn = buttons.find(async (btn) => {
        const text = (await btn.element()).textContent ?? '';
        return !text.includes('cancel');
      });
      if (confirmBtn) {
        await userEvent.click(confirmBtn);
      }

      expect(onAdd).toHaveBeenCalledWith('my-screen', baseScreens[0]);
    });

    it('calls onAdd on Enter key press', async () => {
      const onAdd = vi.fn();
      render(<AddScreenRow baseScreens={baseScreens} onAdd={onAdd} />);

      await userEvent.click(page.getByRole('button'));
      await userEvent.type(page.getByRole('textbox'), 'custom-screen{Enter}');

      expect(onAdd).toHaveBeenCalledWith('custom-screen', baseScreens[0]);
    });

    it('does NOT call onAdd when name is empty', async () => {
      const onAdd = vi.fn();
      render(<AddScreenRow baseScreens={baseScreens} onAdd={onAdd} />);

      await userEvent.click(page.getByRole('button'));
      await userEvent.keyboard('{Enter}');

      expect(onAdd).not.toHaveBeenCalled();
    });
  });

  describe('Cancellation', () => {
    it('hides the input after pressing Escape', async () => {
      render(<AddScreenRow baseScreens={baseScreens} onAdd={vi.fn()} />);

      await userEvent.click(page.getByRole('button'));
      await userEvent.keyboard('{Escape}');

      await expect.element(page.getByRole('textbox')).not.toBeInTheDocument();
    });
  });
});
