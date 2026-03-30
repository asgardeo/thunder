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
import {render, screen} from '@thunder/test-utils';
import {describe, it, expect, vi, beforeEach, afterEach} from 'vitest';
import ConfigureBasicInfo from '../ConfigureBasicInfo';
import type {ConfigureBasicInfoProps} from '../ConfigureBasicInfo';

vi.mock('@thunder/utils');

const mockSuggestions = ['Alpha Manager', 'Beta Editor', 'Gamma Viewer'];
const {generateRandomHumanReadableIdentifiers} = await import('@thunder/utils');

describe('ConfigureBasicInfo', () => {
  const mockOnNameChange = vi.fn();

  const defaultProps: ConfigureBasicInfoProps = {
    name: '',
    onNameChange: mockOnNameChange,
  };

  const renderComponent = (props = defaultProps) => render(<ConfigureBasicInfo {...props} />);

  beforeEach(() => {
    vi.mocked(generateRandomHumanReadableIdentifiers).mockReturnValue(mockSuggestions);
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('should render the component with title', () => {
    renderComponent();

    expect(screen.getByRole('heading')).toBeInTheDocument();
  });

  it('should render name text field', () => {
    renderComponent();

    expect(screen.getByLabelText(/name/i)).toBeInTheDocument();
  });

  it('should display current name value', () => {
    renderComponent({...defaultProps, name: 'Test Role'});

    const nameInput = screen.getByLabelText(/name/i);
    expect(nameInput).toHaveValue('Test Role');
  });

  it('should call onNameChange when typing in name input', async () => {
    const user = userEvent.setup();
    renderComponent();

    const nameInput = screen.getByLabelText(/name/i);
    await user.type(nameInput, 'A');

    expect(mockOnNameChange).toHaveBeenCalledWith('A');
  });

  it('should render name suggestions', () => {
    renderComponent();

    mockSuggestions.forEach((suggestion) => {
      expect(screen.getByText(suggestion)).toBeInTheDocument();
    });
  });

  it('should call onNameChange when clicking a suggestion chip', async () => {
    const user = userEvent.setup();
    renderComponent();

    const suggestionChip = screen.getByText('Alpha Manager');
    await user.click(suggestionChip);

    expect(mockOnNameChange).toHaveBeenCalledWith('Alpha Manager');
  });

  it('should generate suggestions only once on mount', () => {
    const {rerender} = renderComponent();

    expect(generateRandomHumanReadableIdentifiers).toHaveBeenCalledTimes(1);

    rerender(<ConfigureBasicInfo {...defaultProps} name="Updated" />);

    expect(generateRandomHumanReadableIdentifiers).toHaveBeenCalledTimes(1);
  });

  it('should handle special characters in name', async () => {
    const user = userEvent.setup();
    renderComponent();

    const nameInput = screen.getByLabelText(/name/i);
    await user.type(nameInput, '@');

    expect(mockOnNameChange).toHaveBeenCalledWith('@');
  });

  it('should update input values when props change', () => {
    const {rerender} = renderComponent({...defaultProps, name: 'Initial Name'});

    let nameInput = screen.getByLabelText(/name/i);
    expect(nameInput).toHaveValue('Initial Name');

    rerender(<ConfigureBasicInfo {...defaultProps} name="Updated Name" />);

    nameInput = screen.getByLabelText(/name/i);
    expect(nameInput).toHaveValue('Updated Name');
  });

  describe('onReadyChange callback', () => {
    it('should call onReadyChange with true when name is not empty', () => {
      const mockOnReadyChange = vi.fn();
      renderComponent({...defaultProps, name: 'My Role', onReadyChange: mockOnReadyChange});

      expect(mockOnReadyChange).toHaveBeenCalledWith(true);
    });

    it('should call onReadyChange with false when name is empty', () => {
      const mockOnReadyChange = vi.fn();
      renderComponent({...defaultProps, name: '', onReadyChange: mockOnReadyChange});

      expect(mockOnReadyChange).toHaveBeenCalledWith(false);
    });

    it('should call onReadyChange with false when name contains only whitespace', () => {
      const mockOnReadyChange = vi.fn();
      renderComponent({...defaultProps, name: '   ', onReadyChange: mockOnReadyChange});

      expect(mockOnReadyChange).toHaveBeenCalledWith(false);
    });

    it('should not crash when onReadyChange is undefined', () => {
      expect(() => {
        renderComponent({...defaultProps, onReadyChange: undefined});
      }).not.toThrow();
    });
  });
});
