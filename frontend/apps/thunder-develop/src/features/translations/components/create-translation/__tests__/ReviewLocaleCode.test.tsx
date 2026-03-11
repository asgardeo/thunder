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
import {page, userEvent, renderWithProviders} from '@thunder/test-utils/browser';
import ReviewLocaleCode from '../ReviewLocaleCode';

vi.mock('@thunder/i18n', () => ({
  getDisplayNameForCode: (code: string) => (code ? `Name(${code})` : null),
  toFlagEmoji: (code: string) => (code ? `Flag(${code})` : ''),
}));

const derivedLocale = {code: 'fr-FR', displayName: 'French (France)', flag: '🇫🇷'};

const defaultProps = {
  derivedLocale,
  localeCode: '',
  onLocaleCodeChange: vi.fn(),
  onReadyChange: vi.fn(),
};

describe('ReviewLocaleCode', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('renders the step title and subtitle', async () => {
      await renderWithProviders(<ReviewLocaleCode {...defaultProps} />);

      await expect.element(page.getByText('Review Locale Code')).toBeInTheDocument();
      await expect.element(
        page.getByText(
          'The locale code was derived from your selection. Override it here if you need a different tag.',
        ),
      ).toBeInTheDocument();
    });

    it('renders the locale code input with the derived locale as placeholder', async () => {
      await renderWithProviders(<ReviewLocaleCode {...defaultProps} />);

      await expect.element(page.getByPlaceholder('fr-FR')).toBeInTheDocument();
    });

    it('renders the BCP 47 helper tip', async () => {
      await renderWithProviders(<ReviewLocaleCode {...defaultProps} />);

      await expect.element(
        page.getByText(
          'If you are manually modifying the generated code, use BCP 47 format (e.g. fr-FR for French, de-DE for German, etc.).',
        ),
      ).toBeInTheDocument();
    });
  });

  describe('Preview', () => {
    it('shows the derived locale code in the chip when localeCode is empty', async () => {
      await renderWithProviders(<ReviewLocaleCode {...defaultProps} localeCode="" />);

      await expect.element(page.getByText('fr-FR')).toBeInTheDocument();
    });

    it('shows the override code in the chip when localeCode is set', async () => {
      await renderWithProviders(<ReviewLocaleCode {...defaultProps} localeCode="fr-CA" />);

      await expect.element(page.getByText('fr-CA')).toBeInTheDocument();
    });

    it('shows the resolved display name from the effective code', async () => {
      await renderWithProviders(<ReviewLocaleCode {...defaultProps} localeCode="" />);

      // effectiveCode = 'fr-FR' → getDisplayNameForCode('fr-FR') = 'Name(fr-FR)'
      await expect.element(page.getByText('Name(fr-FR)')).toBeInTheDocument();
    });

    it('shows the display name for the override code when set', async () => {
      await renderWithProviders(<ReviewLocaleCode {...defaultProps} localeCode="fr-CA" />);

      await expect.element(page.getByText('Name(fr-CA)')).toBeInTheDocument();
    });
  });

  describe('onReadyChange', () => {
    it('calls onReadyChange(true) on mount when derivedLocale.code is non-empty', async () => {
      const onReadyChange = vi.fn();

      await renderWithProviders(<ReviewLocaleCode {...defaultProps} onReadyChange={onReadyChange} localeCode="" />);

      // effectiveCode falls back to derivedLocale.code = 'fr-FR', so ready
      expect(onReadyChange).toHaveBeenCalledWith(true);
    });

    it('calls onReadyChange(true) when an override code is provided', async () => {
      const onReadyChange = vi.fn();

      await renderWithProviders(
        <ReviewLocaleCode {...defaultProps} onReadyChange={onReadyChange} localeCode="en-AU" />,
      );

      expect(onReadyChange).toHaveBeenCalledWith(true);
    });
  });

  describe('Interaction', () => {
    it('calls onLocaleCodeChange when the user types in the input', async () => {
      const onLocaleCodeChange = vi.fn();

      await renderWithProviders(
        <ReviewLocaleCode {...defaultProps} onLocaleCodeChange={onLocaleCodeChange} localeCode="" />,
      );

      // The input is controlled (value={localeCode}) so onChange fires once per
      // keystroke with just that character. Type a single character to keep the
      // assertion simple and deterministic.
      await userEvent.type(page.getByPlaceholder('fr-FR'), 'f');

      expect(onLocaleCodeChange).toHaveBeenCalledWith('f');
    });
  });
});
