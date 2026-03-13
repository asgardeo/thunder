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
import ConfigureName, {type ConfigureNameProps} from '../ConfigureName';

vi.mock('@thunder/utils');

const {generateRandomHumanReadableIdentifiers} = await import('@thunder/utils');

describe('ConfigureName', () => {
  const mockOnNameChange = vi.fn();
  const mockSuggestions = ['Brave Tigers Squad', 'Crimson Hawks Team', 'Golden Wolves Pack', 'Silver Eagles Crew'];

  const defaultProps: ConfigureNameProps = {
    name: '',
    onNameChange: mockOnNameChange,
  };

  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(generateRandomHumanReadableIdentifiers).mockReturnValue(mockSuggestions);
  });

  const renderComponent = (props: Partial<ConfigureNameProps> = {}) =>
    render(<ConfigureName {...defaultProps} {...props} />);

  it('should render the component with test id', () => {
    renderComponent();

    expect(screen.getByTestId('configure-name')).toBeInTheDocument();
  });

  it('should render the title heading', () => {
    renderComponent();

    expect(screen.getByRole('heading', {level: 1})).toBeInTheDocument();
  });

  it('should render the text field with correct label', () => {
    renderComponent();

    expect(screen.getByText('Group Name')).toBeInTheDocument();
    expect(screen.getByRole('textbox')).toBeInTheDocument();
  });

  it('should display the current name value', () => {
    renderComponent({name: 'My Test Group'});

    const input = screen.getByRole('textbox');
    expect(input).toHaveValue('My Test Group');
  });

  it('should call onNameChange when typing in the input', async () => {
    const user = userEvent.setup();
    renderComponent();

    const input = screen.getByRole('textbox');
    await user.type(input, 'New Group');

    expect(mockOnNameChange).toHaveBeenCalledTimes(9); // Once per character
  });

  it('should render name suggestions', () => {
    renderComponent();

    mockSuggestions.forEach((suggestion) => {
      expect(screen.getByText(suggestion)).toBeInTheDocument();
    });
  });

  it('should display suggestions label', () => {
    renderComponent();

    expect(screen.getByText('In a hurry? Pick a random name:')).toBeInTheDocument();
  });

  it('should call onNameChange when clicking a suggestion chip', async () => {
    const user = userEvent.setup();
    renderComponent();

    const suggestionChip = screen.getByText('Brave Tigers Squad');
    await user.click(suggestionChip);

    expect(mockOnNameChange).toHaveBeenCalledWith('Brave Tigers Squad');
  });

  it('should render all suggestion chips as clickable', () => {
    renderComponent();

    mockSuggestions.forEach((suggestion) => {
      const chip = screen.getByText(suggestion);
      expect(chip.closest('div[role="button"]')).toBeInTheDocument();
    });
  });

  it('should generate suggestions only once on mount', () => {
    const {rerender} = renderComponent();

    expect(generateRandomHumanReadableIdentifiers).toHaveBeenCalledTimes(1);

    rerender(<ConfigureName {...defaultProps} name="Updated Name" />);

    expect(generateRandomHumanReadableIdentifiers).toHaveBeenCalledTimes(1);
  });

  it('should display placeholder text', () => {
    renderComponent();

    const input = screen.getByRole('textbox');
    expect(input).toHaveAttribute('placeholder');
  });

  it('should render required field indicator', () => {
    renderComponent();

    const label = screen.getByText('Group Name');
    const labelElement = label.closest('label');
    expect(labelElement).toHaveClass('Mui-required');
  });

  it('should allow clearing the input', async () => {
    const user = userEvent.setup();
    renderComponent({name: 'Some Group'});

    const input = screen.getByRole('textbox');
    await user.clear(input);

    expect(mockOnNameChange).toHaveBeenCalledWith('');
  });

  it('should handle rapid suggestion clicks', async () => {
    const user = userEvent.setup();
    renderComponent();

    await user.click(screen.getByText('Brave Tigers Squad'));
    await user.click(screen.getByText('Crimson Hawks Team'));

    expect(mockOnNameChange).toHaveBeenCalledWith('Brave Tigers Squad');
    expect(mockOnNameChange).toHaveBeenCalledWith('Crimson Hawks Team');
    expect(mockOnNameChange).toHaveBeenCalledTimes(2);
  });

  it('should update input value when name prop changes', () => {
    const {rerender} = renderComponent({name: 'Initial Name'});

    let input = screen.getByRole('textbox');
    expect(input).toHaveValue('Initial Name');

    rerender(<ConfigureName name="Updated Name" onNameChange={mockOnNameChange} />);

    input = screen.getByRole('textbox');
    expect(input).toHaveValue('Updated Name');
  });

  describe('onReadyChange callback', () => {
    it('should call onReadyChange with true when name is not empty', () => {
      const mockOnReadyChange = vi.fn();
      renderComponent({name: 'My Group', onReadyChange: mockOnReadyChange});

      expect(mockOnReadyChange).toHaveBeenCalledWith(true);
    });

    it('should call onReadyChange with false when name is empty', () => {
      const mockOnReadyChange = vi.fn();
      renderComponent({name: '', onReadyChange: mockOnReadyChange});

      expect(mockOnReadyChange).toHaveBeenCalledWith(false);
    });

    it('should call onReadyChange with false when name contains only whitespace', () => {
      const mockOnReadyChange = vi.fn();
      renderComponent({name: '   ', onReadyChange: mockOnReadyChange});

      expect(mockOnReadyChange).toHaveBeenCalledWith(false);
    });

    it('should not crash when onReadyChange is undefined', () => {
      expect(() => {
        renderComponent({name: 'Test Group', onReadyChange: undefined});
      }).not.toThrow();
    });

    it('should call onReadyChange when name transitions from empty to non-empty', () => {
      const mockOnReadyChange = vi.fn();
      const {rerender} = render(
        <ConfigureName name="" onNameChange={mockOnNameChange} onReadyChange={mockOnReadyChange} />,
      );

      expect(mockOnReadyChange).toHaveBeenCalledWith(false);
      mockOnReadyChange.mockClear();

      rerender(
        <ConfigureName name="New Group" onNameChange={mockOnNameChange} onReadyChange={mockOnReadyChange} />,
      );

      expect(mockOnReadyChange).toHaveBeenCalledWith(true);
    });

    it('should call onReadyChange when name transitions from non-empty to empty', () => {
      const mockOnReadyChange = vi.fn();
      const {rerender} = render(
        <ConfigureName name="My Group" onNameChange={mockOnNameChange} onReadyChange={mockOnReadyChange} />,
      );

      expect(mockOnReadyChange).toHaveBeenCalledWith(true);
      mockOnReadyChange.mockClear();

      rerender(<ConfigureName name="" onNameChange={mockOnNameChange} onReadyChange={mockOnReadyChange} />);

      expect(mockOnReadyChange).toHaveBeenCalledWith(false);
    });
  });
});
