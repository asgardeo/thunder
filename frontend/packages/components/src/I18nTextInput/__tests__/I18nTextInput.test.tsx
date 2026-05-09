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

import {render, screen} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {describe, it, expect, vi, beforeEach} from 'vitest';
import I18nTextInput from '../I18nTextInput';

const {mockResolve, mockMutate, mockAddResourceBundle, mockOnTranslationCreated, mockLanguages, mockTranslations} =
  vi.hoisted(() => ({
    mockResolve: vi.fn<(value: string) => string | null>(),
    mockMutate: vi.fn(),
    mockAddResourceBundle: vi.fn(),
    mockOnTranslationCreated: vi.fn(),
    mockLanguages: {value: {languages: ['en-US', 'fr-FR']} as {languages: string[]} | undefined},
    mockTranslations: {
      value: {
        translations: {
          custom: {
            display_name: 'Display Name',
            first_name: 'First Name',
          },
        },
      },
    },
  }));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
    i18n: {
      addResourceBundle: mockAddResourceBundle,
    },
  }),
}));

vi.mock('@thunderid/hooks', () => ({
  useTemplateLiteralResolver: () => ({
    resolve: mockResolve,
  }),
}));

vi.mock('@thunderid/i18n', () => ({
  NamespaceConstants: {
    CUSTOM_NAMESPACE: 'custom',
  },
  I18nDefaultConstants: {
    FALLBACK_LANGUAGE: 'en-US',
  },
  useGetLanguages: () => ({
    data: mockLanguages.value,
  }),
  useGetTranslations: () => ({
    data: mockTranslations.value,
    isLoading: false,
  }),
  useUpdateTranslation: ({onMutationSuccess}: {onMutationSuccess?: () => void} = {}) => ({
    mutate: (payload: unknown, callbacks?: {onSuccess?: () => void; onError?: (err: Error) => void}) => {
      mockMutate(payload, callbacks);
      onMutationSuccess?.();
    },
    isPending: false,
  }),
}));

describe('I18nTextInput', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockResolve.mockReturnValue('Resolved Display');
    mockLanguages.value = {languages: ['en-US', 'fr-FR']};
    mockTranslations.value = {
      translations: {
        custom: {
          display_name: 'Display Name',
          first_name: 'First Name',
        },
      },
    };
  });

  it('renders the label and input', () => {
    render(<I18nTextInput label="Display Name" value="" onChange={vi.fn()} placeholder="Type a value" />);

    expect(screen.getByText('Display Name')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('Type a value')).toBeInTheDocument();
  });

  it('shows the resolved value when input has an i18n template', () => {
    render(<I18nTextInput label="Display Name" value="{{t(custom:display_name)}}" onChange={vi.fn()} />);

    expect(screen.getByText('Resolved value')).toBeInTheDocument();
    expect(screen.getByText('Resolved Display')).toBeInTheDocument();
  });

  it('uses provided labels when given', () => {
    render(
      <I18nTextInput
        label="Display Name"
        value="{{t(custom:display_name)}}"
        onChange={vi.fn()}
        labels={{resolvedValueLabel: 'Translated value'}}
      />,
    );

    expect(screen.getByText('Translated value')).toBeInTheDocument();
  });

  it('opens popover and switches to create mode', async () => {
    const user = userEvent.setup();
    const {container} = render(
      <I18nTextInput label="Display Name" value="" onChange={vi.fn()} defaultNewKey="Display Name" />,
    );

    const configureButton = container.querySelector('button');
    expect(configureButton).not.toBeNull();
    await user.click(configureButton!);

    await user.click(screen.getByText('Create New Translation'));

    expect(screen.getByDisplayValue('display_name')).toBeInTheDocument();
    expect(screen.getByText('Create')).toBeInTheDocument();
    expect(screen.getByText('Cancel')).toBeInTheDocument();
  });

  it('creates a translation, calls onTranslationCreated, and emits the template value', async () => {
    const onChange = vi.fn();
    const user = userEvent.setup();

    mockLanguages.value = undefined;
    mockMutate.mockImplementation((_payload: unknown, callbacks?: {onSuccess?: () => void}) => {
      callbacks?.onSuccess?.();
    });

    const {container} = render(
      <I18nTextInput
        label="Display Name"
        value=""
        onChange={onChange}
        defaultNewKey="Display Name"
        onTranslationCreated={mockOnTranslationCreated}
      />,
    );

    const configureButton = container.querySelector('button');
    await user.click(configureButton!);
    await user.click(screen.getByText('Create New Translation'));

    await user.type(screen.getByPlaceholderText('e.g., First Name'), 'Shown Display Name');
    await user.click(screen.getByText('Create'));

    expect(mockMutate).toHaveBeenCalled();
    expect(mockMutate.mock.calls[0][0]).toEqual({
      language: 'en-US',
      namespace: 'custom',
      key: 'display_name',
      value: 'Shown Display Name',
    });

    expect(mockAddResourceBundle).toHaveBeenCalledWith(
      'en-US',
      'custom',
      {display_name: 'Shown Display Name'},
      true,
      true,
    );
    expect(onChange).toHaveBeenCalledWith('{{t(custom:display_name)}}');
    expect(mockOnTranslationCreated).toHaveBeenCalled();
  });
});
