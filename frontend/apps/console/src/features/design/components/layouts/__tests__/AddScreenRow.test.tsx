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
import {render, screen} from '@thunder/test-utils';
import userEvent from '@testing-library/user-event';
import AddScreenRow from '../AddScreenRow';

vi.mock('react-i18next', async () => {
  const actual = await vi.importActual<typeof import('react-i18next')>('react-i18next');
  return {
    ...actual,
    useTranslation: () => ({t: (key: string) => key}),
  };
});

const baseScreens = ['auth', 'login'];

describe('AddScreenRow', () => {
  describe('Initial state', () => {
    it('renders the "Add screen" trigger button', () => {
      render(<AddScreenRow baseScreens={baseScreens} onAdd={vi.fn()} />);
      // The add trigger is visible (i18n key returned as-is)
      expect(screen.getByRole('button')).toBeInTheDocument();
    });

    it('does not show the text field before activation', () => {
      render(<AddScreenRow baseScreens={baseScreens} onAdd={vi.fn()} />);
      expect(screen.queryByRole('textbox')).not.toBeInTheDocument();
    });
  });

  describe('Expanded state', () => {
    it('shows the text input after clicking the add button', async () => {
      const user = userEvent.setup();
      render(<AddScreenRow baseScreens={baseScreens} onAdd={vi.fn()} />);

      await user.click(screen.getByRole('button'));

      expect(screen.getByRole('textbox')).toBeInTheDocument();
    });

    it('shows a cancel button after expansion', async () => {
      const user = userEvent.setup();
      render(<AddScreenRow baseScreens={baseScreens} onAdd={vi.fn()} />);

      await user.click(screen.getByRole('button'));

      // Expanded form should have at least 2 buttons (add + cancel)
      expect(screen.getAllByRole('button').length).toBeGreaterThanOrEqual(2);
    });
  });

  describe('Adding a screen', () => {
    it('calls onAdd with the typed name and first baseScreen when confirmed', async () => {
      const onAdd = vi.fn();
      const user = userEvent.setup();
      render(<AddScreenRow baseScreens={baseScreens} onAdd={onAdd} />);

      await user.click(screen.getByRole('button'));
      await user.type(screen.getByRole('textbox'), 'my-screen');

      // Click the confirm/add button (find by looking for non-cancel buttons)
      const buttons = screen.getAllByRole('button');
      // The add/confirm button triggers onAdd — find the first non-cancel button
      const confirmBtn = buttons.find((btn) => btn.textContent && !btn.textContent.includes('cancel'));
      if (confirmBtn) {
        await user.click(confirmBtn);
      }

      expect(onAdd).toHaveBeenCalledWith('my-screen', baseScreens[0]);
    });

    it('calls onAdd on Enter key press', async () => {
      const onAdd = vi.fn();
      const user = userEvent.setup();
      render(<AddScreenRow baseScreens={baseScreens} onAdd={onAdd} />);

      await user.click(screen.getByRole('button'));
      await user.type(screen.getByRole('textbox'), 'custom-screen{Enter}');

      expect(onAdd).toHaveBeenCalledWith('custom-screen', baseScreens[0]);
    });

    it('does NOT call onAdd when name is empty', async () => {
      const onAdd = vi.fn();
      const user = userEvent.setup();
      render(<AddScreenRow baseScreens={baseScreens} onAdd={onAdd} />);

      await user.click(screen.getByRole('button'));
      await user.keyboard('{Enter}');

      expect(onAdd).not.toHaveBeenCalled();
    });
  });

  describe('Cancellation', () => {
    it('hides the input after pressing Escape', async () => {
      const user = userEvent.setup();
      render(<AddScreenRow baseScreens={baseScreens} onAdd={vi.fn()} />);

      await user.click(screen.getByRole('button'));
      await user.keyboard('{Escape}');

      expect(screen.queryByRole('textbox')).not.toBeInTheDocument();
    });
  });
});
