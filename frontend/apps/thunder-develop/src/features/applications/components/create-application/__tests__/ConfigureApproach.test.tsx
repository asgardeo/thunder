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

import {beforeEach, describe, expect, it, vi} from 'vitest';
import {render, screen} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import ConfigureApproach from '../ConfigureApproach';
import {ApplicationCreateFlowSignInApproach} from '../../../models/application-create-flow';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
}));

describe('ConfigureApproach', () => {
  const mockOnApproachChange = vi.fn();
  const mockOnReadyChange = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('marks the step ready on mount', () => {
    render(
      <ConfigureApproach
        selectedApproach={ApplicationCreateFlowSignInApproach.INBUILT}
        onApproachChange={mockOnApproachChange}
        onReadyChange={mockOnReadyChange}
      />,
    );

    expect(mockOnReadyChange).toHaveBeenCalledWith(true);
  });

  it('switches to the custom approach when the native card is clicked', async () => {
    const user = userEvent.setup();

    render(
      <ConfigureApproach
        selectedApproach={ApplicationCreateFlowSignInApproach.INBUILT}
        onApproachChange={mockOnApproachChange}
        onReadyChange={mockOnReadyChange}
      />,
    );

    await user.click(screen.getByText('applications:onboarding.configure.approach.native.title'));

    expect(mockOnApproachChange).toHaveBeenCalledWith(ApplicationCreateFlowSignInApproach.EMBEDDED);
  });

  it('switches to the inbuilt approach when the inbuilt card is clicked', async () => {
    const user = userEvent.setup();

    render(
      <ConfigureApproach
        selectedApproach={ApplicationCreateFlowSignInApproach.EMBEDDED}
        onApproachChange={mockOnApproachChange}
        onReadyChange={mockOnReadyChange}
      />,
    );

    await user.click(screen.getByText('applications:onboarding.configure.approach.inbuilt.title'));

    expect(mockOnApproachChange).toHaveBeenCalledWith(ApplicationCreateFlowSignInApproach.INBUILT);
  });

  it('displays the correct titles and descriptions', () => {
    render(
      <ConfigureApproach
        selectedApproach={ApplicationCreateFlowSignInApproach.INBUILT}
        onApproachChange={mockOnApproachChange}
        onReadyChange={mockOnReadyChange}
      />,
    );

    expect(screen.getByText('applications:onboarding.configure.approach.title')).toBeInTheDocument();
    expect(screen.getByText('applications:onboarding.configure.approach.subtitle')).toBeInTheDocument();
    expect(screen.getByText('applications:onboarding.configure.approach.inbuilt.title')).toBeInTheDocument();
    expect(screen.getByText('applications:onboarding.configure.approach.inbuilt.description')).toBeInTheDocument();
    expect(screen.getByText('applications:onboarding.configure.approach.native.title')).toBeInTheDocument();
    expect(screen.getByText('applications:onboarding.configure.approach.native.description')).toBeInTheDocument();
  });

  it('shows the correct approach as selected via radio button', () => {
    render(
      <ConfigureApproach
        selectedApproach={ApplicationCreateFlowSignInApproach.INBUILT}
        onApproachChange={mockOnApproachChange}
        onReadyChange={mockOnReadyChange}
      />,
    );

    const radios = screen.getAllByRole('radio');
    const inbuiltRadio = radios.find((radio) => radio.getAttribute('value') === 'INBUILT');
    const customRadio = radios.find((radio) => radio.getAttribute('value') === 'CUSTOM');

    expect(inbuiltRadio).toBeChecked();
    expect(customRadio).not.toBeChecked();
  });

  it('shows custom approach as selected when provided', () => {
    render(
      <ConfigureApproach
        selectedApproach={ApplicationCreateFlowSignInApproach.EMBEDDED}
        onApproachChange={mockOnApproachChange}
        onReadyChange={mockOnReadyChange}
      />,
    );

    const radios = screen.getAllByRole('radio');
    const inbuiltRadio = radios.find((radio) => radio.getAttribute('value') === 'INBUILT');
    const customRadio = radios.find((radio) => radio.getAttribute('value') === 'CUSTOM');

    expect(inbuiltRadio).not.toBeChecked();
    expect(customRadio).toBeChecked();
  });

  it('handles radio button change events', async () => {
    const user = userEvent.setup();

    render(
      <ConfigureApproach
        selectedApproach={ApplicationCreateFlowSignInApproach.INBUILT}
        onApproachChange={mockOnApproachChange}
        onReadyChange={mockOnReadyChange}
      />,
    );

    const customRadio = screen.getAllByRole('radio')[1];
    await user.click(customRadio);

    expect(mockOnApproachChange).toHaveBeenCalledWith(ApplicationCreateFlowSignInApproach.EMBEDDED);
  });

  it('handles card click events for inbuilt approach', async () => {
    const user = userEvent.setup();

    render(
      <ConfigureApproach
        selectedApproach={ApplicationCreateFlowSignInApproach.EMBEDDED}
        onApproachChange={mockOnApproachChange}
        onReadyChange={mockOnReadyChange}
      />,
    );

    // Click the radio button directly
    const radios = screen.getAllByRole('radio');
    const inbuiltRadio = radios.find((radio) => radio.getAttribute('value') === 'INBUILT');
    await user.click(inbuiltRadio!);

    expect(mockOnApproachChange).toHaveBeenCalledWith(ApplicationCreateFlowSignInApproach.INBUILT);
  });

  it('handles card click events for custom approach', async () => {
    const user = userEvent.setup();

    render(
      <ConfigureApproach
        selectedApproach={ApplicationCreateFlowSignInApproach.INBUILT}
        onApproachChange={mockOnApproachChange}
        onReadyChange={mockOnReadyChange}
      />,
    );

    // Click the radio button directly
    const radios = screen.getAllByRole('radio');
    const customRadio = radios.find((radio) => radio.getAttribute('value') === 'CUSTOM');
    await user.click(customRadio!);

    expect(mockOnApproachChange).toHaveBeenCalledWith(ApplicationCreateFlowSignInApproach.EMBEDDED);
  });

  it('does not call onReadyChange when not provided', () => {
    render(
      <ConfigureApproach
        selectedApproach={ApplicationCreateFlowSignInApproach.INBUILT}
        onApproachChange={mockOnApproachChange}
      />,
    );

    // Should not throw error when onReadyChange is undefined
    expect(() =>
      render(
        <ConfigureApproach
          selectedApproach={ApplicationCreateFlowSignInApproach.INBUILT}
          onApproachChange={mockOnApproachChange}
        />,
      ),
    ).not.toThrow();
  });

  it('displays icons for each approach option', () => {
    render(
      <ConfigureApproach
        selectedApproach={ApplicationCreateFlowSignInApproach.INBUILT}
        onApproachChange={mockOnApproachChange}
        onReadyChange={mockOnReadyChange}
      />,
    );

    // Check that icons are rendered (SVG elements should be present)
    const icons = document.querySelectorAll('svg');
    expect(icons.length).toBeGreaterThanOrEqual(2); // At least ExternalLink and Code icons
  });
});
