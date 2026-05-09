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

import userEvent from '@testing-library/user-event';
import {render, screen} from '@thunderid/test-utils';
import {describe, it, expect, vi, afterEach} from 'vitest';
import ConfigureOrganizationUnit from '../ConfigureOrganizationUnit';
import type {ConfigureOrganizationUnitProps} from '../ConfigureOrganizationUnit';

// Mock OrganizationUnitTreePicker
vi.mock('@thunderid/configure-organization-units', () => ({
  OrganizationUnitTreePicker: ({value, onChange}: {value: string; onChange: (ouId: string) => void}) => (
    <div data-testid="ou-tree-picker">
      <span data-testid="ou-picker-value">{value}</span>
      <button type="button" data-testid="ou-select-btn" onClick={() => onChange('selected-ou-id')}>
        Select OU
      </button>
    </div>
  ),
}));

describe('ConfigureOrganizationUnit', () => {
  const mockOnOuIdChange = vi.fn();

  const defaultProps: ConfigureOrganizationUnitProps = {
    selectedOuId: '',
    onOuIdChange: mockOnOuIdChange,
  };

  const renderComponent = (props = defaultProps) => render(<ConfigureOrganizationUnit {...props} />);

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('should render title and subtitle', () => {
    renderComponent();

    expect(screen.getByRole('heading')).toBeInTheDocument();
  });

  it('should render OrganizationUnitTreePicker', () => {
    renderComponent();

    expect(screen.getByTestId('ou-tree-picker')).toBeInTheDocument();
  });

  it('should pass selectedOuId to picker', () => {
    renderComponent({...defaultProps, selectedOuId: 'ou-123'});

    expect(screen.getByTestId('ou-picker-value')).toHaveTextContent('ou-123');
  });

  it('should call onOuIdChange when OU is selected', async () => {
    const user = userEvent.setup();
    renderComponent();

    const selectBtn = screen.getByTestId('ou-select-btn');
    await user.click(selectBtn);

    expect(mockOnOuIdChange).toHaveBeenCalledWith('selected-ou-id');
  });

  describe('onReadyChange callback', () => {
    it('should call onReadyChange with true when selectedOuId is not empty', () => {
      const mockOnReadyChange = vi.fn();
      renderComponent({...defaultProps, selectedOuId: 'ou-1', onReadyChange: mockOnReadyChange});

      expect(mockOnReadyChange).toHaveBeenCalledWith(true);
    });

    it('should call onReadyChange with false when selectedOuId is empty string', () => {
      const mockOnReadyChange = vi.fn();
      renderComponent({...defaultProps, selectedOuId: '', onReadyChange: mockOnReadyChange});

      expect(mockOnReadyChange).toHaveBeenCalledWith(false);
    });

    it('should not crash when onReadyChange is undefined', () => {
      expect(() => {
        renderComponent({...defaultProps, onReadyChange: undefined});
      }).not.toThrow();
    });
  });
});
