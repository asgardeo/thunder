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
import {describe, expect, it, vi} from 'vitest';
import DynamicInputPlaceholderAdapter from '../DynamicInputPlaceholderAdapter';
import {ElementCategories, ElementTypes, type Element} from '@/features/flows/models/elements';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => {
      const translations: Record<string, string> = {
        'flows:core.placeholders.dynamicInputPlaceholder.title': 'Dynamic Input',
        'flows:core.placeholders.dynamicInputPlaceholder.hint': 'Resolves input fields passed from runtime.',
      };

      return translations[key] ?? key;
    },
  }),
}));

describe('DynamicInputPlaceholderAdapter', () => {
  const createResource = (overrides: Partial<Element> = {}): Element =>
    ({
      id: 'dynamic-inputs',
      type: ElementTypes.DynamicInputPlaceholder,
      category: ElementCategories.Display,
      resourceType: 'ELEMENT',
      config: {},
      ...overrides,
    }) as Element;

  it('should render fallback translation copy when placeholder and hint are absent', () => {
    render(<DynamicInputPlaceholderAdapter resource={createResource()} />);

    expect(screen.getByText('Dynamic Input')).toBeInTheDocument();
    expect(screen.getByText('Resolves input fields passed from runtime.')).toBeInTheDocument();
  });

  it('should render custom placeholder and hint when present on the resource', () => {
    render(
      <DynamicInputPlaceholderAdapter
        resource={createResource({
          hint: 'Custom hint',
          placeholder: 'Custom placeholder',
        } as Partial<Element>)}
      />,
    );

    expect(screen.getByText('Custom placeholder')).toBeInTheDocument();
    expect(screen.getByText('Custom hint')).toBeInTheDocument();
  });
});
