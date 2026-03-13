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

import {describe, it, expect, beforeEach, vi} from 'vitest';
import {render, screen} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import ConfigureOrganizationUnit, {type ConfigureOrganizationUnitProps} from '../ConfigureOrganizationUnit';

vi.mock('../../../../organization-units/components/OrganizationUnitTreePicker', () => ({
  default: ({value, onChange}: {value: string; onChange: (id: string) => void}) => (
    <div data-testid="ou-tree-picker">
      <span data-testid="ou-value">{value}</span>
      <button type="button" data-testid="select-ou" onClick={() => onChange('ou-123')}>
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

  beforeEach(() => {
    vi.clearAllMocks();
  });

  const renderComponent = (props: Partial<ConfigureOrganizationUnitProps> = {}) =>
    render(<ConfigureOrganizationUnit {...defaultProps} {...props} />);

  it('should render the component with test id', () => {
    renderComponent();

    expect(screen.getByTestId('configure-organization-unit')).toBeInTheDocument();
  });

  it('should render the title heading', () => {
    renderComponent();

    expect(screen.getByRole('heading', {level: 1})).toBeInTheDocument();
  });

  it('should render the subtitle', () => {
    renderComponent();

    expect(screen.getByText('Choose the organization unit this group will belong to.')).toBeInTheDocument();
  });

  it('should render the organization unit tree picker', () => {
    renderComponent();

    expect(screen.getByTestId('ou-tree-picker')).toBeInTheDocument();
  });

  it('should pass selectedOuId to the tree picker', () => {
    renderComponent({selectedOuId: 'ou-456'});

    expect(screen.getByTestId('ou-value')).toHaveTextContent('ou-456');
  });

  it('should call onOuIdChange when an OU is selected', async () => {
    const user = userEvent.setup();
    renderComponent();

    await user.click(screen.getByTestId('select-ou'));

    expect(mockOnOuIdChange).toHaveBeenCalledWith('ou-123');
  });

  it('should render required field indicator', () => {
    renderComponent();

    expect(screen.getByText('Organization Unit')).toBeInTheDocument();
  });

  describe('onReadyChange callback', () => {
    it('should call onReadyChange with true when selectedOuId is not empty', () => {
      const mockOnReadyChange = vi.fn();
      renderComponent({selectedOuId: 'ou-123', onReadyChange: mockOnReadyChange});

      expect(mockOnReadyChange).toHaveBeenCalledWith(true);
    });

    it('should call onReadyChange with false when selectedOuId is empty', () => {
      const mockOnReadyChange = vi.fn();
      renderComponent({selectedOuId: '', onReadyChange: mockOnReadyChange});

      expect(mockOnReadyChange).toHaveBeenCalledWith(false);
    });

    it('should not crash when onReadyChange is undefined', () => {
      expect(() => {
        renderComponent({selectedOuId: 'ou-123', onReadyChange: undefined});
      }).not.toThrow();
    });

    it('should call onReadyChange when selectedOuId transitions from empty to non-empty', () => {
      const mockOnReadyChange = vi.fn();
      const {rerender} = render(
        <ConfigureOrganizationUnit selectedOuId="" onOuIdChange={mockOnOuIdChange} onReadyChange={mockOnReadyChange} />,
      );

      expect(mockOnReadyChange).toHaveBeenCalledWith(false);
      mockOnReadyChange.mockClear();

      rerender(
        <ConfigureOrganizationUnit
          selectedOuId="ou-123"
          onOuIdChange={mockOnOuIdChange}
          onReadyChange={mockOnReadyChange}
        />,
      );

      expect(mockOnReadyChange).toHaveBeenCalledWith(true);
    });

    it('should call onReadyChange when selectedOuId transitions from non-empty to empty', () => {
      const mockOnReadyChange = vi.fn();
      const {rerender} = render(
        <ConfigureOrganizationUnit
          selectedOuId="ou-123"
          onOuIdChange={mockOnOuIdChange}
          onReadyChange={mockOnReadyChange}
        />,
      );

      expect(mockOnReadyChange).toHaveBeenCalledWith(true);
      mockOnReadyChange.mockClear();

      rerender(
        <ConfigureOrganizationUnit selectedOuId="" onOuIdChange={mockOnOuIdChange} onReadyChange={mockOnReadyChange} />,
      );

      expect(mockOnReadyChange).toHaveBeenCalledWith(false);
    });
  });
});
