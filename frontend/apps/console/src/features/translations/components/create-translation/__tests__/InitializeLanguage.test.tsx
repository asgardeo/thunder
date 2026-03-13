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

import {describe, expect, it, vi, beforeEach} from 'vitest';
import {render, screen, fireEvent} from '@thunder/test-utils';
import userEvent from '@testing-library/user-event';
import InitializeLanguage from '../InitializeLanguage';

vi.mock('react-i18next', async () => {
  const actual = await vi.importActual<typeof import('react-i18next')>('react-i18next');
  return {
    ...actual,
    useTranslation: () => ({t: (key: string) => key}),
  };
});

const defaultProps = {
  populateFromEnglish: true,
  onPopulateChange: vi.fn(),
  isCreating: false,
  progress: 0,
};

describe('InitializeLanguage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('renders the step title and subtitle', () => {
      render(<InitializeLanguage {...defaultProps} />);

      expect(screen.getByText('language.create.initialize.title')).toBeInTheDocument();
      expect(screen.getByText('language.create.initialize.subtitle')).toBeInTheDocument();
    });

    it('renders both strategy card labels', () => {
      render(<InitializeLanguage {...defaultProps} />);

      expect(screen.getByText('language.create.initialize.copyFromEnglish.label')).toBeInTheDocument();
      expect(screen.getByText('language.create.initialize.startEmpty.label')).toBeInTheDocument();
    });

    it('renders both strategy card descriptions', () => {
      render(<InitializeLanguage {...defaultProps} />);

      expect(screen.getByText('language.create.initialize.copyFromEnglish.description')).toBeInTheDocument();
      expect(screen.getByText('language.create.initialize.startEmpty.description')).toBeInTheDocument();
    });
  });

  describe('Progress indicator', () => {
    it('does not show the progress bar or spinner when not creating', () => {
      render(<InitializeLanguage {...defaultProps} isCreating={false} />);

      expect(screen.queryByRole('progressbar')).not.toBeInTheDocument();
    });

    it('shows a spinner and progress bar while creating', () => {
      render(<InitializeLanguage {...defaultProps} isCreating progress={42} />);

      // LinearProgress and CircularProgress both use role="progressbar"
      expect(screen.getAllByRole('progressbar')).toHaveLength(2);
    });

    it('displays the current progress percentage while creating', () => {
      render(<InitializeLanguage {...defaultProps} isCreating progress={65} />);

      expect(screen.getByText(/65%/)).toBeInTheDocument();
    });
  });

  describe('Strategy selection', () => {
    it('calls onPopulateChange(true) when the Copy from English card is clicked', async () => {
      const onPopulateChange = vi.fn();
      const user = userEvent.setup();

      render(<InitializeLanguage {...defaultProps} populateFromEnglish={false} onPopulateChange={onPopulateChange} />);

      await user.click(screen.getByText('language.create.initialize.copyFromEnglish.label'));

      expect(onPopulateChange).toHaveBeenCalledWith(true);
    });

    it('calls onPopulateChange(false) when the Start Empty card is clicked', async () => {
      const onPopulateChange = vi.fn();
      const user = userEvent.setup();

      render(<InitializeLanguage {...defaultProps} populateFromEnglish onPopulateChange={onPopulateChange} />);

      await user.click(screen.getByText('language.create.initialize.startEmpty.label'));

      expect(onPopulateChange).toHaveBeenCalledWith(false);
    });

    it('does not call onPopulateChange when a card is clicked while creating', () => {
      const onPopulateChange = vi.fn();

      render(<InitializeLanguage {...defaultProps} isCreating onPopulateChange={onPopulateChange} />);

      // CardActionArea has pointer-events:none when disabled, so use fireEvent to
      // dispatch a native click that still bubbles to the Card's onClick handler.
      fireEvent.click(screen.getByText('language.create.initialize.startEmpty.label'));

      expect(onPopulateChange).not.toHaveBeenCalled();
    });
  });
});
