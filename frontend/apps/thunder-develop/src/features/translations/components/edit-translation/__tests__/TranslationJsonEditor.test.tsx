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

import {describe, expect, it, vi, beforeEach, afterEach} from 'vitest';
import {page, renderWithProviders} from '@thunder/test-utils/browser';
import TranslationJsonEditor from '../TranslationJsonEditor';

// Monaco Editor is not available in browser tests; replace it with a plain textarea
// that mirrors the same value/onChange contract.
vi.mock('@monaco-editor/react', () => ({
  default: ({value, onChange}: {value: string; onChange?: (v: string | undefined) => void}) => (
    <textarea data-testid="monaco-editor" value={value} onChange={(e) => onChange?.(e.target.value)} />
  ),
}));

const sampleValues = {'actions.save': 'Save', 'actions.cancel': 'Cancel'};
const sampleServerKeys = Object.keys(sampleValues);

describe('TranslationJsonEditor', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.useFakeTimers({shouldAdvanceTime: true});
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  describe('Rendering', () => {
    it('renders the Monaco editor with the initial JSON value', async () => {
      await renderWithProviders(
        <TranslationJsonEditor
          values={sampleValues}
          serverKeys={sampleServerKeys}
          isCustomNamespace={false}
          colorMode="light"
          onChange={vi.fn()}
        />,
      );

      const editor = document.querySelector<HTMLTextAreaElement>('[data-testid="monaco-editor"]');
      expect(editor).toBeTruthy();
      const parsed = JSON.parse(editor!.value) as Record<string, string>;
      expect(parsed).toEqual(sampleValues);
    });

    it('does not show the invalid-JSON warning on initial render', async () => {
      await renderWithProviders(
        <TranslationJsonEditor
          values={sampleValues}
          serverKeys={sampleServerKeys}
          isCustomNamespace={false}
          colorMode="light"
          onChange={vi.fn()}
        />,
      );

      await expect.element(page.getByText('Invalid JSON — fix errors before saving.')).not.toBeInTheDocument();
    });
  });

  describe('Valid JSON changes', () => {
    it('calls onChange with the parsed record after the debounce fires', async () => {
      const onChange = vi.fn();

      await renderWithProviders(
        <TranslationJsonEditor
          values={sampleValues}
          serverKeys={sampleServerKeys}
          isCustomNamespace={false}
          colorMode="light"
          onChange={onChange}
        />,
      );

      const editor = document.querySelector<HTMLTextAreaElement>('[data-testid="monaco-editor"]');
      expect(editor).toBeTruthy();

      // Dispatch a change event on the textarea
      editor!.value = JSON.stringify({'actions.save': 'Enregistrer'});
      editor!.dispatchEvent(new Event('change', {bubbles: true}));

      // Advance the 400ms debounce
      await vi.advanceTimersByTimeAsync(400);

      expect(onChange).toHaveBeenCalledWith({'actions.save': 'Enregistrer'});
    });

    it('does not show the invalid-JSON warning for valid JSON', async () => {
      await renderWithProviders(
        <TranslationJsonEditor
          values={sampleValues}
          serverKeys={sampleServerKeys}
          isCustomNamespace={false}
          colorMode="light"
          onChange={vi.fn()}
        />,
      );

      const editor = document.querySelector<HTMLTextAreaElement>('[data-testid="monaco-editor"]');
      editor!.value = '{"actions.save": "value"}';
      editor!.dispatchEvent(new Event('change', {bubbles: true}));

      await vi.advanceTimersByTimeAsync(400);

      await expect.element(page.getByText('Invalid JSON — fix errors before saving.')).not.toBeInTheDocument();
    });
  });

  describe('Invalid JSON handling', () => {
    it('shows a warning alert when the editor contains invalid JSON', async () => {
      await renderWithProviders(
        <TranslationJsonEditor
          values={sampleValues}
          serverKeys={sampleServerKeys}
          isCustomNamespace={false}
          colorMode="light"
          onChange={vi.fn()}
        />,
      );

      const editor = document.querySelector<HTMLTextAreaElement>('[data-testid="monaco-editor"]');
      editor!.value = '{not valid json';
      editor!.dispatchEvent(new Event('change', {bubbles: true}));

      await vi.advanceTimersByTimeAsync(400);

      await expect.element(page.getByText('Invalid JSON — fix errors before saving.')).toBeInTheDocument();
    });

    it('does not call onChange while JSON is invalid', async () => {
      const onChange = vi.fn();

      await renderWithProviders(
        <TranslationJsonEditor
          values={sampleValues}
          serverKeys={sampleServerKeys}
          isCustomNamespace={false}
          colorMode="light"
          onChange={onChange}
        />,
      );

      const editor = document.querySelector<HTMLTextAreaElement>('[data-testid="monaco-editor"]');
      editor!.value = '{invalid';
      editor!.dispatchEvent(new Event('change', {bubbles: true}));

      await vi.advanceTimersByTimeAsync(400);

      expect(onChange).not.toHaveBeenCalled();
    });

    it('does not show the warning alert when the editor is empty', async () => {
      await renderWithProviders(
        <TranslationJsonEditor
          values={sampleValues}
          serverKeys={sampleServerKeys}
          isCustomNamespace={false}
          colorMode="light"
          onChange={vi.fn()}
        />,
      );

      const editor = document.querySelector<HTMLTextAreaElement>('[data-testid="monaco-editor"]');
      editor!.value = '';
      editor!.dispatchEvent(new Event('change', {bubbles: true}));

      await vi.advanceTimersByTimeAsync(400);

      await expect.element(page.getByText('Invalid JSON — fix errors before saving.')).not.toBeInTheDocument();
    });
  });

  describe('External value updates', () => {
    it('syncs the editor when values prop changes to a new object reference', async () => {
      const {rerender} = await renderWithProviders(
        <TranslationJsonEditor
          values={sampleValues}
          serverKeys={sampleServerKeys}
          isCustomNamespace={false}
          colorMode="light"
          onChange={vi.fn()}
        />,
      );

      const newValues = {'page.title': 'My Page'};
      await rerender(
        <TranslationJsonEditor
          values={newValues}
          serverKeys={Object.keys(newValues)}
          isCustomNamespace={false}
          colorMode="light"
          onChange={vi.fn()}
        />,
      );

      const editor = document.querySelector<HTMLTextAreaElement>('[data-testid="monaco-editor"]');
      expect(editor).toBeTruthy();
      const parsed = JSON.parse(editor!.value) as Record<string, string>;
      expect(parsed).toEqual(newValues);
    });
  });
});
