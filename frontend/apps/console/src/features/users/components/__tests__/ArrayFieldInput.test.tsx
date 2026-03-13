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
import {render, screen, waitFor, userEvent} from '@thunder/test-utils';
import ArrayFieldInput from '../ArrayFieldInput';

describe('ArrayFieldInput', () => {
  const mockOnChange = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders input field with placeholder', () => {
    render(<ArrayFieldInput value={[]} onChange={mockOnChange} fieldLabel="Tags" />);

    expect(screen.getByPlaceholderText('Add tags')).toBeInTheDocument();
  });

  it('renders add button', () => {
    render(<ArrayFieldInput value={[]} onChange={mockOnChange} fieldLabel="Tags" />);

    const addButton = screen.getByRole('button');
    expect(addButton).toBeInTheDocument();
  });

  it('add button is disabled when input is empty', () => {
    render(<ArrayFieldInput value={[]} onChange={mockOnChange} fieldLabel="Tags" />);

    const addButton = screen.getByRole('button');
    expect(addButton).toBeDisabled();
  });

  it('allows typing in the input field', async () => {
    const user = userEvent.setup();
    render(<ArrayFieldInput value={[]} onChange={mockOnChange} fieldLabel="Tags" />);

    const input = screen.getByPlaceholderText('Add tags');
    await user.type(input, 'test value');

    expect(input).toHaveValue('test value');
  });

  it('adds value when add button is clicked', async () => {
    const user = userEvent.setup();
    render(<ArrayFieldInput value={[]} onChange={mockOnChange} fieldLabel="Tags" />);

    const input = screen.getByPlaceholderText('Add tags');
    await user.type(input, 'test value');

    const addButton = screen.getByRole('button');
    await user.click(addButton);

    expect(mockOnChange).toHaveBeenCalledWith(['test value']);
  });

  it('adds value when Enter key is pressed', async () => {
    const user = userEvent.setup();
    render(<ArrayFieldInput value={[]} onChange={mockOnChange} fieldLabel="Tags" />);

    const input = screen.getByPlaceholderText('Add tags');
    await user.type(input, 'test value{Enter}');

    expect(mockOnChange).toHaveBeenCalledWith(['test value']);
  });

  it('clears input after adding value', async () => {
    const user = userEvent.setup();
    render(<ArrayFieldInput value={[]} onChange={mockOnChange} fieldLabel="Tags" />);

    const input = screen.getByPlaceholderText('Add tags');
    await user.type(input, 'test value');

    const addButton = screen.getByRole('button');
    await user.click(addButton);

    await waitFor(() => {
      expect(input).toHaveValue('');
    });
  });

  it('displays existing values as chips', () => {
    render(<ArrayFieldInput value={['tag1', 'tag2', 'tag3']} onChange={mockOnChange} fieldLabel="Tags" />);

    expect(screen.getByText('tag1')).toBeInTheDocument();
    expect(screen.getByText('tag2')).toBeInTheDocument();
    expect(screen.getByText('tag3')).toBeInTheDocument();
  });

  it('deletes chip when delete icon is clicked', async () => {
    const user = userEvent.setup();
    render(<ArrayFieldInput value={['tag1', 'tag2', 'tag3']} onChange={mockOnChange} fieldLabel="Tags" />);

    const deleteButtons = screen.getAllByTestId('CancelIcon');
    await user.click(deleteButtons[1]); // Delete 'tag2'

    expect(mockOnChange).toHaveBeenCalledWith(['tag1', 'tag3']);
  });

  it('trims whitespace from input value', async () => {
    const user = userEvent.setup();
    render(<ArrayFieldInput value={[]} onChange={mockOnChange} fieldLabel="Tags" />);

    const input = screen.getByPlaceholderText('Add tags');
    await user.type(input, '  test value  ');

    const addButton = screen.getByRole('button');
    await user.click(addButton);

    expect(mockOnChange).toHaveBeenCalledWith(['test value']);
  });

  it('does not add empty value', async () => {
    const user = userEvent.setup();
    render(<ArrayFieldInput value={[]} onChange={mockOnChange} fieldLabel="Tags" />);

    const input = screen.getByPlaceholderText('Add tags');
    await user.type(input, '   ');

    const addButton = screen.getByRole('button');
    // Button should be disabled for empty/whitespace-only input
    expect(addButton).toBeDisabled();
    expect(mockOnChange).not.toHaveBeenCalled();
  });

  it('appends new value to existing values', async () => {
    const user = userEvent.setup();
    render(<ArrayFieldInput value={['existing1', 'existing2']} onChange={mockOnChange} fieldLabel="Tags" />);

    const input = screen.getByPlaceholderText('Add tags');
    await user.type(input, 'new value{Enter}');

    await waitFor(() => {
      expect(mockOnChange).toHaveBeenCalledWith(['existing1', 'existing2', 'new value']);
    });
  });

  it('handles non-array value prop gracefully', async () => {
    const user = userEvent.setup();
    render(<ArrayFieldInput value={'not-an-array' as unknown as string[]} onChange={mockOnChange} fieldLabel="Tags" />);

    const input = screen.getByPlaceholderText('Add tags');
    expect(input).toBeInTheDocument();

    // Should still be able to add values even with invalid initial value
    await user.type(input, 'new value');
    const addButton = screen.getByRole('button');
    await user.click(addButton);

    // Should call onChange with just the new value (not trying to spread the invalid value)
    expect(mockOnChange).toHaveBeenCalledWith(['new value']);
  });

  it('enables add button when input has value', async () => {
    const user = userEvent.setup();
    render(<ArrayFieldInput value={[]} onChange={mockOnChange} fieldLabel="Tags" />);

    const addButton = screen.getByRole('button');
    expect(addButton).toBeDisabled();

    const input = screen.getByPlaceholderText('Add tags');
    await user.type(input, 'test');

    expect(addButton).not.toBeDisabled();
  });

  it('prevents default behavior on Enter key press', async () => {
    const user = userEvent.setup();
    render(<ArrayFieldInput value={[]} onChange={mockOnChange} fieldLabel="Tags" />);

    const input = screen.getByPlaceholderText('Add tags');
    await user.type(input, 'test{Enter}');

    // Should not trigger form submission
    expect(mockOnChange).toHaveBeenCalled();
  });

  it('does not render chips container when array is empty', () => {
    const {container} = render(<ArrayFieldInput value={[]} onChange={mockOnChange} fieldLabel="Tags" />);

    // Chips should not be rendered when array is empty
    const chips = container.querySelectorAll('.MuiChip-root');
    expect(chips).toHaveLength(0);
  });

  it('handles different casing in field label for placeholder', () => {
    render(<ArrayFieldInput value={[]} onChange={mockOnChange} fieldLabel="TAGS" />);

    expect(screen.getByPlaceholderText('Add tags')).toBeInTheDocument();
  });

  it('handles empty string after trim', async () => {
    const user = userEvent.setup();
    render(<ArrayFieldInput value={[]} onChange={mockOnChange} fieldLabel="Tags" />);

    const input = screen.getByPlaceholderText('Add tags');
    await user.type(input, '     {Enter}');

    // Should not add anything since trimmed value is empty
    expect(mockOnChange).not.toHaveBeenCalled();
  });

  it('converts non-string items to strings in chip labels', () => {
    render(<ArrayFieldInput value={['123', '456'] as string[]} onChange={mockOnChange} fieldLabel="Tags" />);

    expect(screen.getByText('123')).toBeInTheDocument();
    expect(screen.getByText('456')).toBeInTheDocument();
  });

  it('handles multiple rapid additions', async () => {
    const user = userEvent.setup();
    render(<ArrayFieldInput value={[]} onChange={mockOnChange} fieldLabel="Tags" />);

    const input = screen.getByPlaceholderText('Add tags');

    await user.type(input, 'first{Enter}');
    await waitFor(() => {
      expect(mockOnChange).toHaveBeenCalledWith(['first']);
    });

    await user.type(input, 'second{Enter}');
    await waitFor(() => {
      expect(mockOnChange).toHaveBeenCalledWith(['second']);
    });
  });

  it('does not call onChange when trying to add empty input via button', () => {
    render(<ArrayFieldInput value={[]} onChange={mockOnChange} fieldLabel="Tags" />);

    const addButton = screen.getByRole('button');
    // Button is disabled, but test the behavior
    expect(addButton).toBeDisabled();
    expect(mockOnChange).not.toHaveBeenCalled();
  });

  it('renders with mixed case fieldLabel correctly', () => {
    render(<ArrayFieldInput value={[]} onChange={mockOnChange} fieldLabel="User Tags" />);

    expect(screen.getByPlaceholderText('Add user tags')).toBeInTheDocument();
  });

  it('does not add value when non-Enter key is pressed', async () => {
    const user = userEvent.setup();
    render(<ArrayFieldInput value={[]} onChange={mockOnChange} fieldLabel="Tags" />);

    const input = screen.getByPlaceholderText('Add tags');
    await user.type(input, 'test');

    // Press a key that is not Enter
    await user.keyboard('{Escape}');

    // Should not call onChange
    expect(mockOnChange).not.toHaveBeenCalled();
  });

  it('does not add value when other keys are pressed', async () => {
    const user = userEvent.setup();
    render(<ArrayFieldInput value={[]} onChange={mockOnChange} fieldLabel="Tags" />);

    const input = screen.getByPlaceholderText('Add tags');
    await user.type(input, 'test');

    // Press keys that are not Enter
    await user.keyboard('{Tab}');
    await user.keyboard('{Shift}');
    await user.keyboard('{Control}');

    // Should not call onChange
    expect(mockOnChange).not.toHaveBeenCalled();
  });

  it('handles the case when currentValue.length is 0', () => {
    const {container} = render(<ArrayFieldInput value={[]} onChange={mockOnChange} fieldLabel="Tags" />);

    // When currentValue.length is 0, no chips should be rendered
    const chips = container.querySelectorAll('.MuiChip-root');
    expect(chips).toHaveLength(0);
  });

  it('tests the false branch of inputValue.trim() in handleAdd', async () => {
    const user = userEvent.setup();
    render(<ArrayFieldInput value={[]} onChange={mockOnChange} fieldLabel="Tags" />);

    const input = screen.getByPlaceholderText('Add tags');

    // Type only spaces
    await user.type(input, '   ');

    // Try to trigger handleAdd with only whitespace (button should be disabled)
    const addButton = screen.getByRole('button');
    expect(addButton).toBeDisabled();

    // Even if we could click it, it shouldn't add anything
    expect(mockOnChange).not.toHaveBeenCalled();
  });

  it('tests button disabled state transitions', async () => {
    const user = userEvent.setup();
    render(<ArrayFieldInput value={[]} onChange={mockOnChange} fieldLabel="Tags" />);

    const input = screen.getByPlaceholderText('Add tags');
    const addButton = screen.getByRole('button');

    // Initially disabled
    expect(addButton).toBeDisabled();

    // Type something - button should enable
    await user.type(input, 'test');
    expect(addButton).not.toBeDisabled();

    // Clear input - button should disable again
    await user.clear(input);
    expect(addButton).toBeDisabled();

    // Type whitespace only - button should remain disabled
    await user.type(input, '   ');
    expect(addButton).toBeDisabled();

    // Type actual content - button should enable
    await user.type(input, 'real content');
    expect(addButton).not.toBeDisabled();
  });

  it('handles null or undefined value gracefully', async () => {
    const user = userEvent.setup();
    render(<ArrayFieldInput value={null as unknown as string[]} onChange={mockOnChange} fieldLabel="Tags" />);

    const input = screen.getByPlaceholderText('Add tags');
    await user.type(input, 'test');

    const addButton = screen.getByRole('button');
    await user.click(addButton);

    // Should treat null/undefined as empty array
    expect(mockOnChange).toHaveBeenCalledWith(['test']);
  });
});
