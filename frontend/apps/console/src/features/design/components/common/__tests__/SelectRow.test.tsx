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
import SelectRow from '../SelectRow';

const options = [
  {label: 'Light', value: 'light'},
  {label: 'Dark', value: 'dark'},
  {label: 'System', value: 'system'},
];

describe('SelectRow', () => {
  describe('Rendering', () => {
    it('renders the label', () => {
      render(<SelectRow label="Color Scheme" value="light" options={options} onChange={vi.fn()} />);
      expect(screen.getByText('Color Scheme')).toBeInTheDocument();
    });

    it('renders the currently selected value', () => {
      render(<SelectRow label="Color Scheme" value="dark" options={options} onChange={vi.fn()} />);
      expect(screen.getByText('Dark')).toBeInTheDocument();
    });

    it('renders options in the dropdown when opened', async () => {
      const user = userEvent.setup();
      render(<SelectRow label="Color Scheme" value="light" options={options} onChange={vi.fn()} />);

      await user.click(screen.getByRole('combobox'));

      expect(screen.getByRole('option', {name: 'Dark'})).toBeInTheDocument();
      expect(screen.getByRole('option', {name: 'System'})).toBeInTheDocument();
    });

    it('renders a select element', () => {
      render(<SelectRow label="Direction" value="ltr" options={[{label: 'LTR', value: 'ltr'}]} onChange={vi.fn()} />);
      expect(screen.getByRole('combobox')).toBeInTheDocument();
    });
  });

  describe('Interaction', () => {
    it('calls onChange with selected option value', async () => {
      const onChange = vi.fn();
      const user = userEvent.setup();
      render(<SelectRow label="Color Scheme" value="light" options={options} onChange={onChange} />);

      await user.click(screen.getByRole('combobox'));
      await user.click(screen.getByRole('option', {name: 'Dark'}));

      expect(onChange).toHaveBeenCalledWith('dark');
    });

    it('calls onChange exactly once per selection', async () => {
      const onChange = vi.fn();
      const user = userEvent.setup();
      render(<SelectRow label="Color Scheme" value="light" options={options} onChange={onChange} />);

      await user.click(screen.getByRole('combobox'));
      await user.click(screen.getByRole('option', {name: 'System'}));

      expect(onChange).toHaveBeenCalledOnce();
    });
  });
});
