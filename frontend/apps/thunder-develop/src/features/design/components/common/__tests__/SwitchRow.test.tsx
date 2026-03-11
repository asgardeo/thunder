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
import SwitchRow from '../SwitchRow';

describe('SwitchRow', () => {
  describe('Rendering', () => {
    it('renders the label', async () => {
      render(<SwitchRow label="Show Logo" value={false} onChange={vi.fn()} />);
      await expect.element(page.getByText('Show Logo')).toBeInTheDocument();
    });

    it('renders a switch input', async () => {
      render(<SwitchRow label="Enable" value={false} onChange={vi.fn()} />);
      await expect.element(page.getByRole('switch')).toBeInTheDocument();
    });

    it('switch is checked when value is true', async () => {
      render(<SwitchRow label="Active" value onChange={vi.fn()} />);
      await expect.element(page.getByRole('switch')).toBeChecked();
    });

    it('switch is unchecked when value is false', async () => {
      render(<SwitchRow label="Active" value={false} onChange={vi.fn()} />);
      await expect.element(page.getByRole('switch')).not.toBeChecked();
    });
  });

  describe('Interaction', () => {
    it('calls onChange with true when toggled on', async () => {
      const onChange = vi.fn();
      render(<SwitchRow label="Enabled" value={false} onChange={onChange} />);

      await userEvent.click(page.getByRole('switch'));

      expect(onChange).toHaveBeenCalledWith(true);
    });

    it('calls onChange with false when toggled off', async () => {
      const onChange = vi.fn();
      render(<SwitchRow label="Enabled" value onChange={onChange} />);

      await userEvent.click(page.getByRole('switch'));

      expect(onChange).toHaveBeenCalledWith(false);
    });

    it('calls onChange exactly once per click', async () => {
      const onChange = vi.fn();
      render(<SwitchRow label="Enabled" value={false} onChange={onChange} />);

      await userEvent.click(page.getByRole('switch'));

      expect(onChange).toHaveBeenCalledOnce();
    });
  });
});
