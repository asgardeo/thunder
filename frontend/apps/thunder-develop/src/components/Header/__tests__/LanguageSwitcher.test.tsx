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
import {render, screen, fireEvent, waitFor} from '@thunder/test-utils';
import LanguageSwitcher from '../LanguageSwitcher';

// Use vi.hoisted to create the mock function before vi.mock is hoisted
const {mockUseLanguage} = vi.hoisted(() => ({
  mockUseLanguage: vi.fn(() => ({
    currentLanguage: 'en-US' as const,
    availableLanguages: [
      {code: 'en-US' as const, name: 'English (United States)', nativeName: 'English (US)', direction: 'ltr' as const},
    ],
    setLanguage: vi.fn(),
  })),
}));

// Mock @thunder/i18n
vi.mock('@thunder/i18n', () => ({
  useLanguage: mockUseLanguage,
}));

describe('LanguageSwitcher', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should render the language switcher button', () => {
    render(<LanguageSwitcher />);

    const button = screen.getByRole('button', {name: /change language/i});
    expect(button).toBeInTheDocument();
  });

  it('should open language menu when button is clicked', async () => {
    render(<LanguageSwitcher />);

    const button = screen.getByRole('button', {name: /change language/i});
    fireEvent.click(button);

    await waitFor(() => {
      expect(screen.getByText('English (US)')).toBeInTheDocument();
    });
  });

  it('should call setLanguage when selecting a language', async () => {
    const mockSetLanguage = vi.fn().mockResolvedValue(undefined);
    mockUseLanguage.mockReturnValue({
      currentLanguage: 'en-US' as const,
      availableLanguages: [
        {code: 'en-US' as const, name: 'English (United States)', nativeName: 'English (US)', direction: 'ltr' as const},
      ],
      setLanguage: mockSetLanguage,
    });

    render(<LanguageSwitcher />);

    const button = screen.getByRole('button', {name: /change language/i});
    fireEvent.click(button);

    await waitFor(() => {
      const englishOption = screen.getByText('English (US)');
      fireEvent.click(englishOption);
    });

    expect(mockSetLanguage).toHaveBeenCalledWith('en-US');
  });

  it('should close the menu after selecting a language', async () => {
    const mockSetLanguage = vi.fn().mockResolvedValue(undefined);
    mockUseLanguage.mockReturnValue({
      currentLanguage: 'en-US' as const,
      availableLanguages: [
        {code: 'en-US' as const, name: 'English (United States)', nativeName: 'English (US)', direction: 'ltr' as const},
      ],
      setLanguage: mockSetLanguage,
    });

    render(<LanguageSwitcher />);

    const button = screen.getByRole('button', {name: /change language/i});
    fireEvent.click(button);

    await waitFor(() => {
      const englishOption = screen.getByText('English (US)');
      fireEvent.click(englishOption);
    });

    await waitFor(() => {
      expect(screen.queryByRole('menu')).not.toBeInTheDocument();
    });
  });

  it('should handle multiple languages', async () => {
    const mockSetLanguage = vi.fn().mockResolvedValue(undefined);
    mockUseLanguage.mockReturnValue({
      currentLanguage: 'en-US' as const,
      availableLanguages: [
        {code: 'en-US' as const, name: 'English (United States)', nativeName: 'English (US)', direction: 'ltr' as const},
        {code: 'es-ES' as unknown as 'en-US', name: 'Spanish (Spain)', nativeName: 'Español (España)', direction: 'ltr' as const},
        {code: 'fr-FR' as unknown as 'en-US', name: 'French (France)', nativeName: 'Français (France)', direction: 'ltr' as const},
      ],
      setLanguage: mockSetLanguage,
    });

    render(<LanguageSwitcher />);

    const button = screen.getByRole('button', {name: /change language/i});
    fireEvent.click(button);

    await waitFor(() => {
      expect(screen.getByText('English (US)')).toBeInTheDocument();
      expect(screen.getByText('Español (España)')).toBeInTheDocument();
      expect(screen.getByText('Français (France)')).toBeInTheDocument();
    });
  });

  it('should show secondary name when different from native name', async () => {
    const mockSetLanguage = vi.fn().mockResolvedValue(undefined);
    mockUseLanguage.mockReturnValue({
      currentLanguage: 'es-ES' as unknown as 'en-US',
      availableLanguages: [
        {code: 'es-ES' as unknown as 'en-US', name: 'Spanish (Spain)', nativeName: 'Español (España)', direction: 'ltr' as const},
      ],
      setLanguage: mockSetLanguage,
    });

    render(<LanguageSwitcher />);

    const button = screen.getByRole('button', {name: /change language/i});
    fireEvent.click(button);

    await waitFor(() => {
      expect(screen.getByText('Español (España)')).toBeInTheDocument();
      expect(screen.getByText('Spanish (Spain)')).toBeInTheDocument();
    });
  });

  it('should not show secondary name when same as native name', async () => {
    const mockSetLanguage = vi.fn().mockResolvedValue(undefined);
    mockUseLanguage.mockReturnValue({
      currentLanguage: 'en-US' as const,
      availableLanguages: [
        {code: 'en-US' as const, name: 'English (US)', nativeName: 'English (US)', direction: 'ltr' as const},
      ],
      setLanguage: mockSetLanguage,
    });

    render(<LanguageSwitcher />);

    const button = screen.getByRole('button', {name: /change language/i});
    fireEvent.click(button);

    await waitFor(() => {
      const menuItems = screen.getAllByRole('menuitem');
      expect(menuItems[0]).toHaveTextContent('English (US)');
      // Secondary text should not be present since name equals nativeName
      const secondaryText = menuItems[0].querySelector('.MuiListItemText-secondary');
      expect(secondaryText).not.toBeInTheDocument();
    });
  });

  it('should handle language change error gracefully', async () => {
    const mockSetLanguage = vi.fn().mockRejectedValue(new Error('Language change failed'));
    mockUseLanguage.mockReturnValue({
      currentLanguage: 'en-US' as const,
      availableLanguages: [
        {code: 'en-US' as const, name: 'English (United States)', nativeName: 'English (US)', direction: 'ltr' as const},
        {code: 'es-ES' as unknown as 'en-US', name: 'Spanish (Spain)', nativeName: 'Español (España)', direction: 'ltr' as const},
      ],
      setLanguage: mockSetLanguage,
    });

    render(<LanguageSwitcher />);

    const button = screen.getByRole('button', {name: /change language/i});
    fireEvent.click(button);

    await waitFor(() => {
      const spanishOption = screen.getByText('Español (España)');
      fireEvent.click(spanishOption);
    });

    // Should still close the menu even if language change fails
    await waitFor(() => {
      expect(screen.queryByRole('menu')).not.toBeInTheDocument();
    });

    expect(mockSetLanguage).toHaveBeenCalledWith('es-ES');
  });

  it('should mark current language as selected', async () => {
    const mockSetLanguage = vi.fn().mockResolvedValue(undefined);
    mockUseLanguage.mockReturnValue({
      currentLanguage: 'es-ES' as unknown as 'en-US',
      availableLanguages: [
        {code: 'en-US' as const, name: 'English (United States)', nativeName: 'English (US)', direction: 'ltr' as const},
        {code: 'es-ES' as unknown as 'en-US', name: 'Spanish (Spain)', nativeName: 'Español (España)', direction: 'ltr' as const},
      ],
      setLanguage: mockSetLanguage,
    });

    render(<LanguageSwitcher />);

    const button = screen.getByRole('button', {name: /change language/i});
    fireEvent.click(button);

    await waitFor(() => {
      const menuItems = screen.getAllByRole('menuitem');
      const spanishOption = menuItems.find((item) => item.textContent?.includes('Español'));
      expect(spanishOption).toHaveClass('Mui-selected');
    });
  });

  it('should have proper accessibility attributes when menu is closed', () => {
    render(<LanguageSwitcher />);

    const button = screen.getByRole('button', {name: /change language/i});
    expect(button).not.toHaveAttribute('aria-controls');
    expect(button).toHaveAttribute('aria-haspopup', 'true');
    // aria-expanded may not be set when menu is closed
    expect(button).not.toHaveAttribute('aria-expanded');
  });

  it('should have proper accessibility attributes when menu is open', async () => {
    render(<LanguageSwitcher />);

    const button = screen.getByRole('button', {name: /change language/i});
    fireEvent.click(button);

    await waitFor(() => {
      expect(button).toHaveAttribute('aria-controls', 'language-menu');
      expect(button).toHaveAttribute('aria-haspopup', 'true');
      expect(button).toHaveAttribute('aria-expanded', 'true');
    });
  });

  it('should not mark non-current language as selected', async () => {
    const mockSetLanguage = vi.fn().mockResolvedValue(undefined);
    mockUseLanguage.mockReturnValue({
      currentLanguage: 'en-US' as const,
      availableLanguages: [
        {code: 'en-US' as const, name: 'English (United States)', nativeName: 'English (US)', direction: 'ltr' as const},
        {code: 'es-ES' as unknown as 'en-US', name: 'Spanish (Spain)', nativeName: 'Español (España)', direction: 'ltr' as const},
      ],
      setLanguage: mockSetLanguage,
    });

    render(<LanguageSwitcher />);

    const button = screen.getByRole('button', {name: /change language/i});
    fireEvent.click(button);

    await waitFor(() => {
      const menuItems = screen.getAllByRole('menuitem');
      const englishOption = menuItems.find((item) => item.textContent?.includes('English'));
      const spanishOption = menuItems.find((item) => item.textContent?.includes('Español'));
      expect(englishOption).toHaveClass('Mui-selected');
      expect(spanishOption).not.toHaveClass('Mui-selected');
    });
  });

  it('should show secondary text for language with different name and nativeName in multi-language list', async () => {
    const mockSetLanguage = vi.fn().mockResolvedValue(undefined);
    mockUseLanguage.mockReturnValue({
      currentLanguage: 'en-US' as const,
      availableLanguages: [
        {code: 'en-US' as const, name: 'English (US)', nativeName: 'English (US)', direction: 'ltr' as const},
        {code: 'es-ES' as unknown as 'en-US', name: 'Spanish (Spain)', nativeName: 'Español (España)', direction: 'ltr' as const},
      ],
      setLanguage: mockSetLanguage,
    });

    render(<LanguageSwitcher />);

    const button = screen.getByRole('button', {name: /change language/i});
    fireEvent.click(button);

    await waitFor(() => {
      const menuItems = screen.getAllByRole('menuitem');
      // English (US) should NOT have secondary text since name === nativeName
      const englishSecondary = menuItems[0].querySelector('.MuiListItemText-secondary');
      expect(englishSecondary).not.toBeInTheDocument();

      // Spanish should have secondary text since name !== nativeName
      const spanishSecondary = menuItems[1].querySelector('.MuiListItemText-secondary');
      expect(spanishSecondary).toBeInTheDocument();
      expect(spanishSecondary).toHaveTextContent('Spanish (Spain)');
    });
  });

  it('should select a different language from the menu', async () => {
    const mockSetLanguage = vi.fn().mockResolvedValue(undefined);
    mockUseLanguage.mockReturnValue({
      currentLanguage: 'en-US' as const,
      availableLanguages: [
        {code: 'en-US' as const, name: 'English (United States)', nativeName: 'English (US)', direction: 'ltr' as const},
        {code: 'fr-FR' as unknown as 'en-US', name: 'French (France)', nativeName: 'Français (France)', direction: 'ltr' as const},
      ],
      setLanguage: mockSetLanguage,
    });

    render(<LanguageSwitcher />);

    const button = screen.getByRole('button', {name: /change language/i});
    fireEvent.click(button);

    await waitFor(() => {
      const frenchOption = screen.getByText('Français (France)');
      fireEvent.click(frenchOption);
    });

    expect(mockSetLanguage).toHaveBeenCalledWith('fr-FR');
  });

  it('should render Languages icon inside the button', () => {
    const {container} = render(<LanguageSwitcher />);

    const languagesIcon = container.querySelector('svg.lucide-languages');
    expect(languagesIcon).toBeInTheDocument();
  });

  it('should render menu with correct id when open', async () => {
    render(<LanguageSwitcher />);

    const button = screen.getByRole('button', {name: /change language/i});
    fireEvent.click(button);

    await waitFor(() => {
      const menu = document.getElementById('language-menu');
      expect(menu).toBeInTheDocument();
    });
  });

  it('should render menu items with ListItemText', async () => {
    const mockSetLanguage = vi.fn().mockResolvedValue(undefined);
    mockUseLanguage.mockReturnValue({
      currentLanguage: 'en-US' as const,
      availableLanguages: [
        {code: 'en-US' as const, name: 'English (United States)', nativeName: 'English (US)', direction: 'ltr' as const},
        {code: 'es-ES' as unknown as 'en-US', name: 'Spanish (Spain)', nativeName: 'Español (España)', direction: 'ltr' as const},
      ],
      setLanguage: mockSetLanguage,
    });

    render(<LanguageSwitcher />);

    const button = screen.getByRole('button', {name: /change language/i});
    fireEvent.click(button);

    await waitFor(() => {
      const menuItems = screen.getAllByRole('menuitem');
      expect(menuItems).toHaveLength(2);

      // Each menu item should have a ListItemText with primary text
      const primaryTexts = document.querySelectorAll('.MuiListItemText-primary');
      expect(primaryTexts.length).toBe(2);
    });
  });

  it('should render with empty available languages', async () => {
    mockUseLanguage.mockReturnValue({
      currentLanguage: 'en-US' as const,
      availableLanguages: [],
      setLanguage: vi.fn().mockResolvedValue(undefined),
    });

    render(<LanguageSwitcher />);

    const button = screen.getByRole('button', {name: /change language/i});
    fireEvent.click(button);

    await waitFor(() => {
      const menuItems = screen.queryAllByRole('menuitem');
      expect(menuItems).toHaveLength(0);
    });
  });

  it('should preserve button after re-render', () => {
    const {rerender} = render(<LanguageSwitcher />);

    rerender(<LanguageSwitcher />);

    const button = screen.getByRole('button', {name: /change language/i});
    expect(button).toBeInTheDocument();
  });
});
