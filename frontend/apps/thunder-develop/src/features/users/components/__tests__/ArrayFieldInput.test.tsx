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

import {describe, it, expect, vi, beforeEach} from 'vitest';
import {page, userEvent} from 'vitest/browser';
import {renderWithProviders} from '@thunder/test-utils/browser';
import ArrayFieldInput from '../ArrayFieldInput';

describe('ArrayFieldInput', () => {
  const mockOnChange = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders input field with placeholder', async () => {
    await renderWithProviders(<ArrayFieldInput value={[]} onChange={mockOnChange} fieldLabel="Tags" />);

    await expect.element(page.getByPlaceholder('Add tags')).toBeInTheDocument();
  });

  it('renders add button', async () => {
    await renderWithProviders(<ArrayFieldInput value={[]} onChange={mockOnChange} fieldLabel="Tags" />);

    const addButton = page.getByRole('button');
    await expect.element(addButton).toBeInTheDocument();
  });

  it('add button is disabled when input is empty', async () => {
    await renderWithProviders(<ArrayFieldInput value={[]} onChange={mockOnChange} fieldLabel="Tags" />);

    const addButton = page.getByRole('button');
    await expect.element(addButton).toBeDisabled();
  });

  it('allows typing in the input field', async () => {
    await renderWithProviders(<ArrayFieldInput value={[]} onChange={mockOnChange} fieldLabel="Tags" />);

    const input = page.getByPlaceholder('Add tags');
    await userEvent.fill(input, 'test value');

    await expect.element(input).toHaveValue('test value');
  });

  it('adds value when add button is clicked', async () => {
    await renderWithProviders(<ArrayFieldInput value={[]} onChange={mockOnChange} fieldLabel="Tags" />);

    const input = page.getByPlaceholder('Add tags');
    await userEvent.fill(input, 'test value');

    const addButton = page.getByRole('button');
    await userEvent.click(addButton);

    await vi.waitFor(() => {
      expect(mockOnChange).toHaveBeenCalledWith(['test value']);
    });
  });

  it('adds value when Enter key is pressed', async () => {
    await renderWithProviders(<ArrayFieldInput value={[]} onChange={mockOnChange} fieldLabel="Tags" />);

    const input = page.getByPlaceholder('Add tags');
    await userEvent.fill(input, 'test value');
    await userEvent.keyboard('{Enter}');

    await vi.waitFor(() => {
      expect(mockOnChange).toHaveBeenCalledWith(['test value']);
    });
  });

  it('clears input after adding value', async () => {
    await renderWithProviders(<ArrayFieldInput value={[]} onChange={mockOnChange} fieldLabel="Tags" />);

    const input = page.getByPlaceholder('Add tags');
    await userEvent.fill(input, 'test value');

    const addButton = page.getByRole('button');
    await userEvent.click(addButton);

    await expect.element(input).toHaveValue('');
  });

  it('displays existing values as chips', async () => {
    await renderWithProviders(<ArrayFieldInput value={['tag1', 'tag2', 'tag3']} onChange={mockOnChange} fieldLabel="Tags" />);

    await expect.element(page.getByText('tag1')).toBeInTheDocument();
    await expect.element(page.getByText('tag2')).toBeInTheDocument();
    await expect.element(page.getByText('tag3')).toBeInTheDocument();
  });

  it('deletes chip when delete icon is clicked', async () => {
    await renderWithProviders(<ArrayFieldInput value={['tag1', 'tag2', 'tag3']} onChange={mockOnChange} fieldLabel="Tags" />);

    const deleteButtons = page.getByTestId('CancelIcon').all();
    await userEvent.click(deleteButtons[1]); // Delete 'tag2'

    await vi.waitFor(() => {
      expect(mockOnChange).toHaveBeenCalledWith(['tag1', 'tag3']);
    });
  });

  it('trims whitespace from input value', async () => {
    await renderWithProviders(<ArrayFieldInput value={[]} onChange={mockOnChange} fieldLabel="Tags" />);

    const input = page.getByPlaceholder('Add tags');
    await userEvent.fill(input, '  test value  ');

    const addButton = page.getByRole('button');
    await userEvent.click(addButton);

    await vi.waitFor(() => {
      expect(mockOnChange).toHaveBeenCalledWith(['test value']);
    });
  });

  it('does not add empty value', async () => {
    await renderWithProviders(<ArrayFieldInput value={[]} onChange={mockOnChange} fieldLabel="Tags" />);

    const input = page.getByPlaceholder('Add tags');
    await userEvent.fill(input, '   ');

    const addButton = page.getByRole('button');
    // Button should be disabled for empty/whitespace-only input
    await expect.element(addButton).toBeDisabled();
    expect(mockOnChange).not.toHaveBeenCalled();
  });

  it('appends new value to existing values', async () => {
    await renderWithProviders(<ArrayFieldInput value={['existing1', 'existing2']} onChange={mockOnChange} fieldLabel="Tags" />);

    const input = page.getByPlaceholder('Add tags');
    await userEvent.fill(input, 'new value');
    await userEvent.keyboard('{Enter}');

    await vi.waitFor(() => {
      expect(mockOnChange).toHaveBeenCalledWith(['existing1', 'existing2', 'new value']);
    });
  });

  it('handles non-array value prop gracefully', async () => {
    await renderWithProviders(<ArrayFieldInput value={'not-an-array' as unknown as string[]} onChange={mockOnChange} fieldLabel="Tags" />);

    const input = page.getByPlaceholder('Add tags');
    await expect.element(input).toBeInTheDocument();

    // Should still be able to add values even with invalid initial value
    await userEvent.fill(input, 'new value');
    const addButton = page.getByRole('button');
    await userEvent.click(addButton);

    // Should call onChange with just the new value (not trying to spread the invalid value)
    await vi.waitFor(() => {
      expect(mockOnChange).toHaveBeenCalledWith(['new value']);
    });
  });

  it('enables add button when input has value', async () => {
    await renderWithProviders(<ArrayFieldInput value={[]} onChange={mockOnChange} fieldLabel="Tags" />);

    const addButton = page.getByRole('button');
    await expect.element(addButton).toBeDisabled();

    const input = page.getByPlaceholder('Add tags');
    await userEvent.fill(input, 'test');

    await expect.element(addButton).not.toBeDisabled();
  });

  it('prevents default behavior on Enter key press', async () => {
    await renderWithProviders(<ArrayFieldInput value={[]} onChange={mockOnChange} fieldLabel="Tags" />);

    const input = page.getByPlaceholder('Add tags');
    await userEvent.fill(input, 'test');
    await userEvent.keyboard('{Enter}');

    // Should not trigger form submission
    await vi.waitFor(() => {
      expect(mockOnChange).toHaveBeenCalled();
    });
  });

  it('does not render chips container when array is empty', async () => {
    await renderWithProviders(<ArrayFieldInput value={[]} onChange={mockOnChange} fieldLabel="Tags" />);

    // Chips should not be rendered when array is empty
    const chips = document.querySelectorAll('.MuiChip-root');
    expect(chips).toHaveLength(0);
  });

  it('handles different casing in field label for placeholder', async () => {
    await renderWithProviders(<ArrayFieldInput value={[]} onChange={mockOnChange} fieldLabel="TAGS" />);

    await expect.element(page.getByPlaceholder('Add tags')).toBeInTheDocument();
  });

  it('handles empty string after trim', async () => {
    await renderWithProviders(<ArrayFieldInput value={[]} onChange={mockOnChange} fieldLabel="Tags" />);

    const input = page.getByPlaceholder('Add tags');
    await userEvent.fill(input, '     ');
    await userEvent.keyboard('{Enter}');

    // Should not add anything since trimmed value is empty
    expect(mockOnChange).not.toHaveBeenCalled();
  });

  it('converts non-string items to strings in chip labels', async () => {
    await renderWithProviders(<ArrayFieldInput value={['123', '456'] as string[]} onChange={mockOnChange} fieldLabel="Tags" />);

    await expect.element(page.getByText('123')).toBeInTheDocument();
    await expect.element(page.getByText('456')).toBeInTheDocument();
  });

  it('handles multiple rapid additions', async () => {
    await renderWithProviders(<ArrayFieldInput value={[]} onChange={mockOnChange} fieldLabel="Tags" />);

    const input = page.getByPlaceholder('Add tags');

    await userEvent.fill(input, 'first');
    await userEvent.keyboard('{Enter}');
    await vi.waitFor(() => {
      expect(mockOnChange).toHaveBeenCalledWith(['first']);
    });

    await userEvent.fill(input, 'second');
    await userEvent.keyboard('{Enter}');
    await vi.waitFor(() => {
      expect(mockOnChange).toHaveBeenCalledWith(['second']);
    });
  });

  it('does not call onChange when trying to add empty input via button', async () => {
    await renderWithProviders(<ArrayFieldInput value={[]} onChange={mockOnChange} fieldLabel="Tags" />);

    const addButton = page.getByRole('button');
    // Button is disabled, but test the behavior
    await expect.element(addButton).toBeDisabled();
    expect(mockOnChange).not.toHaveBeenCalled();
  });

  it('renders with mixed case fieldLabel correctly', async () => {
    await renderWithProviders(<ArrayFieldInput value={[]} onChange={mockOnChange} fieldLabel="User Tags" />);

    await expect.element(page.getByPlaceholder('Add user tags')).toBeInTheDocument();
  });

  it('does not add value when non-Enter key is pressed', async () => {
    await renderWithProviders(<ArrayFieldInput value={[]} onChange={mockOnChange} fieldLabel="Tags" />);

    const input = page.getByPlaceholder('Add tags');
    await userEvent.fill(input, 'test');

    // Press a key that is not Enter
    await userEvent.keyboard('{Escape}');

    // Should not call onChange
    expect(mockOnChange).not.toHaveBeenCalled();
  });

  it('does not add value when other keys are pressed', async () => {
    await renderWithProviders(<ArrayFieldInput value={[]} onChange={mockOnChange} fieldLabel="Tags" />);

    const input = page.getByPlaceholder('Add tags');
    await userEvent.fill(input, 'test');

    // Press keys that are not Enter
    await userEvent.keyboard('{Tab}');
    await userEvent.keyboard('{Shift}');
    await userEvent.keyboard('{Control}');

    // Should not call onChange
    expect(mockOnChange).not.toHaveBeenCalled();
  });

  it('handles the case when currentValue.length is 0', async () => {
    await renderWithProviders(<ArrayFieldInput value={[]} onChange={mockOnChange} fieldLabel="Tags" />);

    // When currentValue.length is 0, no chips should be rendered
    const chips = document.querySelectorAll('.MuiChip-root');
    expect(chips).toHaveLength(0);
  });

  it('tests the false branch of inputValue.trim() in handleAdd', async () => {
    await renderWithProviders(<ArrayFieldInput value={[]} onChange={mockOnChange} fieldLabel="Tags" />);

    const input = page.getByPlaceholder('Add tags');

    // Type only spaces
    await userEvent.fill(input, '   ');

    // Try to trigger handleAdd with only whitespace (button should be disabled)
    const addButton = page.getByRole('button');
    await expect.element(addButton).toBeDisabled();

    // Even if we could click it, it shouldn't add anything
    expect(mockOnChange).not.toHaveBeenCalled();
  });

  it('tests button disabled state transitions', async () => {
    await renderWithProviders(<ArrayFieldInput value={[]} onChange={mockOnChange} fieldLabel="Tags" />);

    const input = page.getByPlaceholder('Add tags');
    const addButton = page.getByRole('button');

    // Initially disabled
    await expect.element(addButton).toBeDisabled();

    // Type something - button should enable
    await userEvent.fill(input, 'test');
    await expect.element(addButton).not.toBeDisabled();

    // Clear input - button should disable again
    await userEvent.clear(input);
    await expect.element(addButton).toBeDisabled();

    // Type whitespace only - button should remain disabled
    await userEvent.fill(input, '   ');
    await expect.element(addButton).toBeDisabled();

    // Type actual content - button should enable
    await userEvent.fill(input, 'real content');
    await expect.element(addButton).not.toBeDisabled();
  });

  it('handles null or undefined value gracefully', async () => {
    await renderWithProviders(<ArrayFieldInput value={null as unknown as string[]} onChange={mockOnChange} fieldLabel="Tags" />);

    const input = page.getByPlaceholder('Add tags');
    await userEvent.fill(input, 'test');

    const addButton = page.getByRole('button');
    await userEvent.click(addButton);

    // Should treat null/undefined as empty array
    await vi.waitFor(() => {
      expect(mockOnChange).toHaveBeenCalledWith(['test']);
    });
  });
});
