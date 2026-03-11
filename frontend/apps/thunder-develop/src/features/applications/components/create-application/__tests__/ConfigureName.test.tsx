/**
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
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
import {render} from '@thunder/test-utils/browser';
import ConfigureName, {type ConfigureNameProps} from '../ConfigureName';

// Mock the utility library
vi.mock('@thunder/utils');

const {generateRandomHumanReadableIdentifiers} = await import('@thunder/utils');

describe('ConfigureName', () => {
  const mockOnAppNameChange = vi.fn();
  const mockSuggestions = ['My Web App', 'Customer Portal', 'Mobile App', 'Internal Dashboard'];

  const defaultProps: ConfigureNameProps = {
    appName: '',
    onAppNameChange: mockOnAppNameChange,
  };

  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(generateRandomHumanReadableIdentifiers).mockReturnValue(mockSuggestions);
  });

  const renderComponent = async (props: Partial<ConfigureNameProps> = {}) =>
    render(<ConfigureName {...defaultProps} {...props} />);

  it('should render the component with title', async () => {
    await renderComponent();

    await expect.element(page.getByRole('heading', {level: 1})).toBeInTheDocument();
  });

  it('should render the text field with correct label', async () => {
    await renderComponent();

    await expect.element(page.getByText('Application Name')).toBeInTheDocument();
    await expect.element(page.getByRole('textbox')).toBeInTheDocument();
  });

  it('should display the current app name value', async () => {
    await renderComponent({appName: 'My Test App'});

    const input = page.getByRole('textbox');
    expect(input).toHaveValue('My Test App');
  });

  it('should call onAppNameChange when typing in the input', async () => {
    await renderComponent();

    const input = page.getByRole('textbox');
    await userEvent.fill(input, 'New App Name');

    // In browser mode, fill fires one change event with the full text value
    expect(mockOnAppNameChange).toHaveBeenCalled();
    expect(mockOnAppNameChange).toHaveBeenCalledWith('New App Name');
  });

  it('should render name suggestions', async () => {
    await renderComponent();

    // eslint-disable-next-line no-restricted-syntax, no-await-in-loop
    for (const suggestion of mockSuggestions) {
      // eslint-disable-next-line no-await-in-loop
      await expect.element(page.getByText(suggestion)).toBeInTheDocument();
    }
  });

  it('should display suggestions label with icon', async () => {
    await renderComponent();

    await expect.element(page.getByText('In a hurry? Pick a random name:')).toBeInTheDocument();
  });

  it('should call onAppNameChange when clicking a suggestion chip', async () => {
    await renderComponent();

    const suggestionChip = page.getByText('My Web App');
    await userEvent.click(suggestionChip);

    expect(mockOnAppNameChange).toHaveBeenCalledWith('My Web App');
  });

  it('should render all suggestion chips as clickable', async () => {
    await renderComponent();

    // eslint-disable-next-line no-restricted-syntax
    for (const suggestion of mockSuggestions) {
      const chip = page.getByText(suggestion);
      expect(chip.element().closest('div[role="button"]')).toBeInTheDocument();
    }
  });

  it('should generate suggestions only once on mount', async () => {
    const {rerender} = await renderComponent();

    expect(generateRandomHumanReadableIdentifiers).toHaveBeenCalledTimes(1);

    await rerender(<ConfigureName {...defaultProps} appName="Updated Name" />);

    // Should still be called only once due to useMemo
    expect(generateRandomHumanReadableIdentifiers).toHaveBeenCalledTimes(1);
  });

  it('should handle empty app name', async () => {
    await renderComponent({appName: ''});

    const input = page.getByRole('textbox');
    expect(input).toHaveValue('');
  });

  it('should display placeholder text', async () => {
    await renderComponent();

    const input = page.getByRole('textbox');
    expect(input).toHaveAttribute('placeholder');
  });

  it('should render required field indicator', async () => {
    await renderComponent();

    // FormControl with required prop should render asterisk or required indicator
    const label = page.getByText('Application Name');
    expect(label).toBeInTheDocument();
    // Check for the asterisk in the label's parent (which should be a <label> element)
    const labelElement = label.element().closest('label');
    expect(labelElement).toHaveClass('Mui-required');
  });

  it('should handle special characters in app name', async () => {
    await renderComponent();

    const input = page.getByRole('textbox');
    const specialName = 'App @#$ 123!';
    await userEvent.fill(input, specialName);

    // In browser mode, fill fires one change event with the full text value
    expect(mockOnAppNameChange).toHaveBeenCalled();
    expect(mockOnAppNameChange).toHaveBeenCalledWith(specialName);
  });

  it('should update input value when appName prop changes', async () => {
    const {rerender} = await renderComponent({appName: 'Initial Name'});

    let input = page.getByRole('textbox');
    expect(input).toHaveValue('Initial Name');

    await rerender(<ConfigureName appName="Updated Name" onAppNameChange={mockOnAppNameChange} />);

    input = page.getByRole('textbox');
    expect(input).toHaveValue('Updated Name');
  });

  it('should allow clearing the input', async () => {
    await renderComponent({appName: 'Some App'});

    const input = page.getByRole('textbox');
    await userEvent.clear(input);

    expect(mockOnAppNameChange).toHaveBeenCalledWith('');
  });

  it('should handle rapid suggestion clicks', async () => {
    await renderComponent();

    const firstSuggestion = page.getByText('My Web App');
    const secondSuggestion = page.getByText('Customer Portal');

    await userEvent.click(firstSuggestion);
    await userEvent.click(secondSuggestion);

    expect(mockOnAppNameChange).toHaveBeenCalledWith('My Web App');
    expect(mockOnAppNameChange).toHaveBeenCalledWith('Customer Portal');
    expect(mockOnAppNameChange).toHaveBeenCalledTimes(2);
  });

  it('should display lightbulb icon for suggestions', async () => {
    await renderComponent();

    // Check that the Lightbulb component is rendered (it's from lucide-react)
    const suggestionsSection = page.getByText('In a hurry? Pick a random name:').element().closest('div');
    expect(suggestionsSection).toBeInTheDocument();
  });

  it('should handle long app names', async () => {
    const longName = 'A'.repeat(100);
    await renderComponent();

    const input = page.getByRole('textbox');
    await userEvent.fill(input, longName);

    // In browser mode, fill fires one change event with the full text value
    expect(mockOnAppNameChange).toHaveBeenCalled();
    expect(mockOnAppNameChange).toHaveBeenCalledWith(longName);
  });

  describe('onReadyChange callback', () => {
    it('should call onReadyChange with true when appName is not empty', async () => {
      const mockOnReadyChange = vi.fn();
      await renderComponent({appName: 'My App', onReadyChange: mockOnReadyChange});

      expect(mockOnReadyChange).toHaveBeenCalledWith(true);
    });

    it('should call onReadyChange with false when appName is empty', async () => {
      const mockOnReadyChange = vi.fn();
      await renderComponent({appName: '', onReadyChange: mockOnReadyChange});

      expect(mockOnReadyChange).toHaveBeenCalledWith(false);
    });

    it('should call onReadyChange with false when appName contains only whitespace', async () => {
      const mockOnReadyChange = vi.fn();
      await renderComponent({appName: '   ', onReadyChange: mockOnReadyChange});

      expect(mockOnReadyChange).toHaveBeenCalledWith(false);
    });

    it('should not crash when onReadyChange is undefined', async () => {
      await expect(
        (async () => {
          await renderComponent({appName: 'Test App', onReadyChange: undefined});
        })(),
      ).resolves.not.toThrow();
    });

    it('should call onReadyChange when appName transitions from empty to non-empty', async () => {
      const mockOnReadyChange = vi.fn();
      const {rerender} = await render(
        <ConfigureName appName="" onAppNameChange={mockOnAppNameChange} onReadyChange={mockOnReadyChange} />,
      );

      expect(mockOnReadyChange).toHaveBeenCalledWith(false);
      mockOnReadyChange.mockClear();

      await rerender(
        <ConfigureName appName="New App" onAppNameChange={mockOnAppNameChange} onReadyChange={mockOnReadyChange} />,
      );

      expect(mockOnReadyChange).toHaveBeenCalledWith(true);
    });

    it('should call onReadyChange when appName transitions from non-empty to empty', async () => {
      const mockOnReadyChange = vi.fn();
      const {rerender} = await render(
        <ConfigureName appName="My App" onAppNameChange={mockOnAppNameChange} onReadyChange={mockOnReadyChange} />,
      );

      expect(mockOnReadyChange).toHaveBeenCalledWith(true);
      mockOnReadyChange.mockClear();

      await rerender(<ConfigureName appName="" onAppNameChange={mockOnAppNameChange} onReadyChange={mockOnReadyChange} />);

      expect(mockOnReadyChange).toHaveBeenCalledWith(false);
    });
  });
});
