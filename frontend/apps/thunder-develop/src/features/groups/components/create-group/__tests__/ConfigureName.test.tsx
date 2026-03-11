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
import {page, userEvent} from 'vitest/browser';
import {renderWithProviders} from '@thunder/test-utils/browser';
import ConfigureName, {type ConfigureNameProps} from '../ConfigureName';

const mockGenerateRandomHumanReadableIdentifiers = vi.hoisted(() => vi.fn());

vi.mock('@thunder/utils', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@thunder/utils')>();
  return {
    ...actual,
    generateRandomHumanReadableIdentifiers: mockGenerateRandomHumanReadableIdentifiers,
  };
});

describe('ConfigureName', () => {
  const mockOnNameChange = vi.fn();
  const mockSuggestions = ['Brave Tigers Squad', 'Crimson Hawks Team', 'Golden Wolves Pack', 'Silver Eagles Crew'];

  const defaultProps: ConfigureNameProps = {
    name: '',
    onNameChange: mockOnNameChange,
  };

  beforeEach(() => {
    vi.clearAllMocks();
    mockGenerateRandomHumanReadableIdentifiers.mockReturnValue(mockSuggestions);
  });

  it('should render the component with test id', async () => {
    await renderWithProviders(<ConfigureName {...defaultProps} />);

    await expect.element(page.getByTestId('configure-name')).toBeInTheDocument();
  });

  it('should render the title heading', async () => {
    await renderWithProviders(<ConfigureName {...defaultProps} />);

    await expect.element(page.getByRole('heading', {level: 1})).toBeInTheDocument();
  });

  it('should render the text field with correct label', async () => {
    await renderWithProviders(<ConfigureName {...defaultProps} />);

    await expect.element(page.getByText('Group Name')).toBeInTheDocument();
    await expect.element(page.getByRole('textbox')).toBeInTheDocument();
  });

  it('should display the current name value', async () => {
    await renderWithProviders(<ConfigureName {...defaultProps} name="My Test Group" />);

    await expect.element(page.getByRole('textbox')).toHaveValue('My Test Group');
  });

  it('should call onNameChange when typing in the input', async () => {
    await renderWithProviders(<ConfigureName {...defaultProps} />);

    const input = page.getByRole('textbox');
    await userEvent.type(input, 'New Group');

    expect(mockOnNameChange).toHaveBeenCalledTimes(9); // Once per character
  });

  it('should render name suggestions', async () => {
    await renderWithProviders(<ConfigureName {...defaultProps} />);

    for (const suggestion of mockSuggestions) {
      await expect.element(page.getByText(suggestion)).toBeInTheDocument();
    }
  });

  it('should display suggestions label', async () => {
    await renderWithProviders(<ConfigureName {...defaultProps} />);

    await expect.element(page.getByText('In a hurry? Pick a random name:')).toBeInTheDocument();
  });

  it('should call onNameChange when clicking a suggestion chip', async () => {
    await renderWithProviders(<ConfigureName {...defaultProps} />);

    await userEvent.click(page.getByText('Brave Tigers Squad'));

    expect(mockOnNameChange).toHaveBeenCalledWith('Brave Tigers Squad');
  });

  it('should render all suggestion chips as clickable', async () => {
    await renderWithProviders(<ConfigureName {...defaultProps} />);

    for (const suggestion of mockSuggestions) {
      await expect.element(page.getByText(suggestion)).toBeInTheDocument();
    }
  });

  it('should generate suggestions only once on mount', async () => {
    const {rerender} = await renderWithProviders(<ConfigureName {...defaultProps} />);

    expect(mockGenerateRandomHumanReadableIdentifiers).toHaveBeenCalledTimes(1);

    await rerender(<ConfigureName {...defaultProps} name="Updated Name" />);

    expect(mockGenerateRandomHumanReadableIdentifiers).toHaveBeenCalledTimes(1);
  });

  it('should display placeholder text', async () => {
    await renderWithProviders(<ConfigureName {...defaultProps} />);

    await expect.element(page.getByRole('textbox')).toHaveAttribute('placeholder');
  });

  it('should allow clearing the input', async () => {
    await renderWithProviders(<ConfigureName {...defaultProps} name="Some Group" />);

    const input = page.getByRole('textbox');
    await userEvent.clear(input);

    expect(mockOnNameChange).toHaveBeenCalledWith('');
  });

  it('should handle rapid suggestion clicks', async () => {
    await renderWithProviders(<ConfigureName {...defaultProps} />);

    await userEvent.click(page.getByText('Brave Tigers Squad'));
    await userEvent.click(page.getByText('Crimson Hawks Team'));

    expect(mockOnNameChange).toHaveBeenCalledWith('Brave Tigers Squad');
    expect(mockOnNameChange).toHaveBeenCalledWith('Crimson Hawks Team');
    expect(mockOnNameChange).toHaveBeenCalledTimes(2);
  });

  it('should update input value when name prop changes', async () => {
    const {rerender} = await renderWithProviders(<ConfigureName {...defaultProps} name="Initial Name" />);

    await expect.element(page.getByRole('textbox')).toHaveValue('Initial Name');

    await rerender(<ConfigureName name="Updated Name" onNameChange={mockOnNameChange} />);

    await expect.element(page.getByRole('textbox')).toHaveValue('Updated Name');
  });

  describe('onReadyChange callback', () => {
    it('should call onReadyChange with true when name is not empty', async () => {
      const mockOnReadyChange = vi.fn();
      await renderWithProviders(<ConfigureName {...defaultProps} name="My Group" onReadyChange={mockOnReadyChange} />);

      expect(mockOnReadyChange).toHaveBeenCalledWith(true);
    });

    it('should call onReadyChange with false when name is empty', async () => {
      const mockOnReadyChange = vi.fn();
      await renderWithProviders(<ConfigureName {...defaultProps} name="" onReadyChange={mockOnReadyChange} />);

      expect(mockOnReadyChange).toHaveBeenCalledWith(false);
    });

    it('should call onReadyChange with false when name contains only whitespace', async () => {
      const mockOnReadyChange = vi.fn();
      await renderWithProviders(<ConfigureName {...defaultProps} name="   " onReadyChange={mockOnReadyChange} />);

      expect(mockOnReadyChange).toHaveBeenCalledWith(false);
    });

    it('should not crash when onReadyChange is undefined', async () => {
      await expect(
        renderWithProviders(<ConfigureName {...defaultProps} name="Test Group" onReadyChange={undefined} />),
      ).resolves.not.toThrow();
    });

    it('should call onReadyChange when name transitions from empty to non-empty', async () => {
      const mockOnReadyChange = vi.fn();
      const {rerender} = await renderWithProviders(
        <ConfigureName name="" onNameChange={mockOnNameChange} onReadyChange={mockOnReadyChange} />,
      );

      expect(mockOnReadyChange).toHaveBeenCalledWith(false);
      mockOnReadyChange.mockClear();

      await rerender(<ConfigureName name="New Group" onNameChange={mockOnNameChange} onReadyChange={mockOnReadyChange} />);

      expect(mockOnReadyChange).toHaveBeenCalledWith(true);
    });

    it('should call onReadyChange when name transitions from non-empty to empty', async () => {
      const mockOnReadyChange = vi.fn();
      const {rerender} = await renderWithProviders(
        <ConfigureName name="My Group" onNameChange={mockOnNameChange} onReadyChange={mockOnReadyChange} />,
      );

      expect(mockOnReadyChange).toHaveBeenCalledWith(true);
      mockOnReadyChange.mockClear();

      await rerender(<ConfigureName name="" onNameChange={mockOnNameChange} onReadyChange={mockOnReadyChange} />);

      expect(mockOnReadyChange).toHaveBeenCalledWith(false);
    });
  });
});
