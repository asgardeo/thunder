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
import {describe, it, expect, beforeEach, vi} from 'vitest';
import ConfigureName, {type ConfigureNameProps} from '../ConfigureName';

// Mock the utility library
vi.mock('@thunderid/utils');

const {generateRandomHumanReadableIdentifiers} = await import('@thunderid/utils');

describe('ConfigureName', () => {
  const mockOnAgentNameChange = vi.fn();
  const mockSuggestions = ['Billing Service', 'Customer Sync', 'Inventory Bot', 'Analytics Worker'];

  const defaultProps: ConfigureNameProps = {
    agentName: '',
    onAgentNameChange: mockOnAgentNameChange,
  };

  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(generateRandomHumanReadableIdentifiers).mockReturnValue(mockSuggestions);
  });

  const renderComponent = (props: Partial<ConfigureNameProps> = {}) =>
    render(<ConfigureName {...defaultProps} {...props} />);

  it('should render the component with title', () => {
    renderComponent();

    expect(screen.getByRole('heading', {level: 1})).toBeInTheDocument();
  });

  it('should render the text field with the correct label', () => {
    renderComponent();

    expect(screen.getByText('Agent name')).toBeInTheDocument();
    expect(screen.getByRole('textbox')).toBeInTheDocument();
  });

  it('should display the current agent name value', () => {
    renderComponent({agentName: 'My Test Agent'});

    const input = screen.getByRole('textbox');
    expect(input).toHaveValue('My Test Agent');
  });

  it('should call onAgentNameChange when typing in the input', async () => {
    const user = userEvent.setup();
    renderComponent();

    const input = screen.getByRole('textbox');
    await user.type(input, 'Hello');

    expect(mockOnAgentNameChange).toHaveBeenCalledTimes(5);
    expect(mockOnAgentNameChange).toHaveBeenLastCalledWith('o');
  });

  it('should render name suggestions', () => {
    renderComponent();

    mockSuggestions.forEach((suggestion) => {
      expect(screen.getByText(suggestion)).toBeInTheDocument();
    });
  });

  it('should call onAgentNameChange when clicking a suggestion chip', async () => {
    const user = userEvent.setup();
    renderComponent();

    const chip = screen.getByText('Billing Service');
    await user.click(chip);

    expect(mockOnAgentNameChange).toHaveBeenCalledWith('Billing Service');
  });

  it('should generate suggestions only once on mount', () => {
    const {rerender} = renderComponent();

    expect(generateRandomHumanReadableIdentifiers).toHaveBeenCalledTimes(1);

    rerender(<ConfigureName {...defaultProps} agentName="Updated Name" />);

    expect(generateRandomHumanReadableIdentifiers).toHaveBeenCalledTimes(1);
  });

  it('should display placeholder text', () => {
    renderComponent();

    const input = screen.getByRole('textbox');
    expect(input).toHaveAttribute('placeholder');
  });

  it('should allow clearing the input', async () => {
    const user = userEvent.setup();
    renderComponent({agentName: 'Some Agent'});

    const input = screen.getByRole('textbox');
    await user.clear(input);

    expect(mockOnAgentNameChange).toHaveBeenCalledWith('');
  });

  describe('onReadyChange callback', () => {
    it('should call onReadyChange with true when agentName is not empty', () => {
      const mockOnReadyChange = vi.fn();
      renderComponent({agentName: 'My Agent', onReadyChange: mockOnReadyChange});

      expect(mockOnReadyChange).toHaveBeenCalledWith(true);
    });

    it('should call onReadyChange with false when agentName is empty', () => {
      const mockOnReadyChange = vi.fn();
      renderComponent({agentName: '', onReadyChange: mockOnReadyChange});

      expect(mockOnReadyChange).toHaveBeenCalledWith(false);
    });

    it('should call onReadyChange with false when agentName contains only whitespace', () => {
      const mockOnReadyChange = vi.fn();
      renderComponent({agentName: '   ', onReadyChange: mockOnReadyChange});

      expect(mockOnReadyChange).toHaveBeenCalledWith(false);
    });

    it('should not crash when onReadyChange is undefined', () => {
      expect(() => {
        renderComponent({agentName: 'Test Agent', onReadyChange: undefined});
      }).not.toThrow();
    });

    it('should call onReadyChange when agentName transitions from empty to non-empty', () => {
      const mockOnReadyChange = vi.fn();
      const {rerender} = render(
        <ConfigureName agentName="" onAgentNameChange={mockOnAgentNameChange} onReadyChange={mockOnReadyChange} />,
      );

      expect(mockOnReadyChange).toHaveBeenCalledWith(false);
      mockOnReadyChange.mockClear();

      rerender(
        <ConfigureName
          agentName="New Agent"
          onAgentNameChange={mockOnAgentNameChange}
          onReadyChange={mockOnReadyChange}
        />,
      );

      expect(mockOnReadyChange).toHaveBeenCalledWith(true);
    });
  });
});
