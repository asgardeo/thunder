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

import {describe, it, expect, vi} from 'vitest';
import {render} from '@thunder/test-utils/browser';
import {page} from 'vitest/browser';
import {Hint} from '../hint';

// Mock the PlaceholderComponent
vi.mock('../adapters/PlaceholderComponent', () => ({
  default: ({children, value}: {children: React.ReactNode; value: string}) => (
    <div data-testid="placeholder-component" data-value={value}>
      {children}
    </div>
  ),
}));

// Mock the icons - use importOriginal to preserve all exports while mocking specific ones
vi.mock('@wso2/oxygen-ui-icons-react', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@wso2/oxygen-ui-icons-react')>();
  return {
    ...actual,
    InfoIcon: () => <svg data-testid="info-icon" />,
  };
});

describe('Hint', () => {
  it('should render hint text', async () => {
    await render(<Hint hint="This is a helpful hint" />);

    await expect.element(page.getByText('This is a helpful hint')).toBeInTheDocument();
  });

  it('should render the InfoIcon', async () => {
    await render(<Hint hint="Test hint" />);

    await expect.element(page.getByTestId('info-icon')).toBeInTheDocument();
  });

  it('should render the PlaceholderComponent with the hint value', async () => {
    await render(<Hint hint="Placeholder hint" />);

    const placeholder = page.getByTestId('placeholder-component');
    expect(placeholder).toBeInTheDocument();
    expect(placeholder).toHaveAttribute('data-value', 'Placeholder hint');
  });

  it('should render empty hint when hint prop is empty string', async () => {
    await render(<Hint hint="" />);

    const placeholder = page.getByTestId('placeholder-component');
    expect(placeholder).toHaveAttribute('data-value', '');
  });

  it('should render hint with special characters', async () => {
    const specialCharsHint = 'Use special characters: &, "test"';
    await render(<Hint hint={specialCharsHint} />);

    await expect.element(page.getByText(specialCharsHint)).toBeInTheDocument();
  });

  it('should render hint with long text', async () => {
    const longHint =
      'This is a very long hint text that provides detailed information about the input field and its expected format for the user to understand.';
    await render(<Hint hint={longHint} />);

    await expect.element(page.getByText(longHint)).toBeInTheDocument();
  });
});
