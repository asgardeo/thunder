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

import {describe, expect, it, vi, beforeEach} from 'vitest';
import {page, userEvent, renderWithProviders} from '@thunder/test-utils/browser';
import NamespaceSelector from '../NamespaceSelector';

const defaultProps = {
  namespaces: ['commonNamespace', 'loginFlow', 'userProfile'],
  value: 'commonNamespace',
  loading: false,
  onChange: vi.fn(),
};

describe('NamespaceSelector', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('renders the namespace label', async () => {
      await renderWithProviders(<NamespaceSelector {...defaultProps} />);

      await expect.element(page.getByText('Namespace')).toBeInTheDocument();
    });

    it('renders the helper text', async () => {
      await renderWithProviders(<NamespaceSelector {...defaultProps} />);

      await expect.element(
        page.getByText(
          'A namespace typically represents a page or a section within a page. It helps group and organize related translation keys for better structure and maintainability.',
        ),
      ).toBeInTheDocument();
    });

    it('renders with the current value displayed in the input', async () => {
      await renderWithProviders(<NamespaceSelector {...defaultProps} value="loginFlow" />);

      await expect.element(page.getByRole('combobox')).toHaveValue('Login Flow');
    });

    it('renders with empty string when value is null', async () => {
      await renderWithProviders(<NamespaceSelector {...defaultProps} value={null} />);

      await expect.element(page.getByRole('combobox')).toHaveValue('');
    });
  });

  describe('Option label formatting', () => {
    it('formats camelCase namespace keys into human-readable labels', async () => {
      await renderWithProviders(<NamespaceSelector {...defaultProps} />);

      await userEvent.click(page.getByRole('combobox'));

      await expect.element(page.getByText('Common Namespace')).toBeInTheDocument();
      await expect.element(page.getByText('Login Flow')).toBeInTheDocument();
      await expect.element(page.getByText('User Profile')).toBeInTheDocument();
    });
  });

  describe('Interaction', () => {
    it('calls onChange when a namespace option is selected', async () => {
      const onChange = vi.fn();

      await renderWithProviders(<NamespaceSelector {...defaultProps} onChange={onChange} />);

      await userEvent.click(page.getByRole('combobox'));
      await userEvent.click(page.getByText('Login Flow'));

      expect(onChange).toHaveBeenCalledWith('loginFlow');
    });
  });

  describe('Loading state', () => {
    it('shows no namespace options while loading', async () => {
      await renderWithProviders(<NamespaceSelector {...defaultProps} loading namespaces={[]} value={null} />);

      await userEvent.click(page.getByRole('combobox'));

      expect(page.getByRole('option').all()).toHaveLength(0);
    });

    it('does not show loading indicator when loading is false', async () => {
      await renderWithProviders(<NamespaceSelector {...defaultProps} loading={false} />);

      await expect.element(page.getByRole('progressbar')).not.toBeInTheDocument();
    });
  });
});
