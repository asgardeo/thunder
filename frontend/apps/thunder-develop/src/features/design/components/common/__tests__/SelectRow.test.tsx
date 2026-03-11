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
import SelectRow from '../SelectRow';

const options = [
  {label: 'Light', value: 'light'},
  {label: 'Dark', value: 'dark'},
  {label: 'System', value: 'system'},
];

describe('SelectRow', () => {
  describe('Rendering', () => {
    it('renders the label', async () => {
      render(<SelectRow label="Color Scheme" value="light" options={options} onChange={vi.fn()} />);
      await expect.element(page.getByText('Color Scheme')).toBeInTheDocument();
    });

    it('renders the currently selected value', async () => {
      render(<SelectRow label="Color Scheme" value="dark" options={options} onChange={vi.fn()} />);
      await expect.element(page.getByText('Dark')).toBeInTheDocument();
    });

    it('renders options in the dropdown when opened', async () => {
      render(<SelectRow label="Color Scheme" value="light" options={options} onChange={vi.fn()} />);

      await userEvent.click(page.getByRole('combobox'));

      await expect.element(page.getByRole('option', {name: 'Dark'})).toBeInTheDocument();
      await expect.element(page.getByRole('option', {name: 'System'})).toBeInTheDocument();
    });

    it('renders a select element', async () => {
      render(<SelectRow label="Direction" value="ltr" options={[{label: 'LTR', value: 'ltr'}]} onChange={vi.fn()} />);
      await expect.element(page.getByRole('combobox')).toBeInTheDocument();
    });
  });

  describe('Interaction', () => {
    it('calls onChange with selected option value', async () => {
      const onChange = vi.fn();
      render(<SelectRow label="Color Scheme" value="light" options={options} onChange={onChange} />);

      await userEvent.click(page.getByRole('combobox'));
      await userEvent.click(page.getByRole('option', {name: 'Dark'}));

      expect(onChange).toHaveBeenCalledWith('dark');
    });

    it('calls onChange exactly once per selection', async () => {
      const onChange = vi.fn();
      render(<SelectRow label="Color Scheme" value="light" options={options} onChange={onChange} />);

      await userEvent.click(page.getByRole('combobox'));
      await userEvent.click(page.getByRole('option', {name: 'System'}));

      expect(onChange).toHaveBeenCalledOnce();
    });
  });
});
