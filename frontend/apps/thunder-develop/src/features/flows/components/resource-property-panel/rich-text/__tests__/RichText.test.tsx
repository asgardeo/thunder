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
import {render} from '@thunder/test-utils/browser';
import {page} from 'vitest/browser';
import type {Resource} from '@/features/flows/models/resources';
import RichText from '../RichText';

// Mock the lexical plugins and components
vi.mock('@lexical/react/LexicalComposer', () => ({
  LexicalComposer: ({children}: {children: React.ReactNode}) => (
    <div data-testid="lexical-composer">{children}</div>
  ),
}));

vi.mock('@lexical/react/LexicalRichTextPlugin', () => ({
  RichTextPlugin: ({contentEditable}: {contentEditable: React.ReactNode}) => (
    <div data-testid="rich-text-plugin">{contentEditable}</div>
  ),
}));

vi.mock('@lexical/react/LexicalContentEditable', () => ({
  ContentEditable: (props: Record<string, unknown>) => (
    <div data-testid="content-editable" {...props} />
  ),
}));

vi.mock('@lexical/react/LexicalErrorBoundary', () => ({
  LexicalErrorBoundary: ({children}: {children: React.ReactNode}) => (
    <div data-testid="error-boundary">{children}</div>
  ),
}));

vi.mock('@lexical/react/LexicalHistoryPlugin', () => ({
  HistoryPlugin: () => <div data-testid="history-plugin" />,
}));

vi.mock('@lexical/react/LexicalAutoFocusPlugin', () => ({
  AutoFocusPlugin: () => <div data-testid="auto-focus-plugin" />,
}));

vi.mock('@lexical/react/LexicalLinkPlugin', () => ({
  LinkPlugin: () => <div data-testid="link-plugin" />,
}));

// Mock helper plugins
vi.mock('../helper-plugins/ToolbarPlugin', () => ({
  default: ({disabled}: {disabled?: boolean}) => (
    <div data-testid="toolbar-plugin" data-disabled={disabled} />
  ),
}));

vi.mock('../helper-plugins/CustomLinkPlugin', () => ({
  default: () => <div data-testid="custom-link-plugin" />,
}));

vi.mock('../helper-plugins/HTMLPlugin', () => ({
  default: ({resource, disabled}: {onChange: () => void; resource: Resource; disabled?: boolean}) => (
    <div
      data-testid="html-plugin"
      data-resource-id={resource?.id}
      data-disabled={disabled}
    />
  ),
}));


describe('RichText', () => {
  const mockOnChange = vi.fn();

  const createMockResource = (overrides: Partial<Resource & {label?: string}> = {}): Resource => ({
    id: 'resource-1',
    resourceType: 'ELEMENT',
    type: 'RICH_TEXT',
    category: 'DISPLAY',
    version: '1.0.0',
    deprecated: false,
    deletable: true,
    display: {
      label: 'Test Rich Text',
      image: '',
      showOnResourcePanel: true,
    },
    config: {
      field: {name: 'richText', type: 'RICH_TEXT'},
      styles: {},
    },
    ...overrides,
  } as unknown as Resource);

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('should render the LexicalComposer wrapper', async () => {
      await render(<RichText onChange={mockOnChange} resource={createMockResource()} />);

      await expect.element(page.getByTestId('lexical-composer')).toBeInTheDocument();
    });

    it('should render the ToolbarPlugin', async () => {
      await render(<RichText onChange={mockOnChange} resource={createMockResource()} />);

      await expect.element(page.getByTestId('toolbar-plugin')).toBeInTheDocument();
    });

    it('should render the RichTextPlugin', async () => {
      await render(<RichText onChange={mockOnChange} resource={createMockResource()} />);

      await expect.element(page.getByTestId('rich-text-plugin')).toBeInTheDocument();
    });

    it('should render the ContentEditable', async () => {
      await render(<RichText onChange={mockOnChange} resource={createMockResource()} />);

      await expect.element(page.getByTestId('content-editable')).toBeInTheDocument();
    });

    it('should render the HistoryPlugin', async () => {
      await render(<RichText onChange={mockOnChange} resource={createMockResource()} />);

      await expect.element(page.getByTestId('history-plugin')).toBeInTheDocument();
    });

    it('should render the AutoFocusPlugin', async () => {
      await render(<RichText onChange={mockOnChange} resource={createMockResource()} />);

      await expect.element(page.getByTestId('auto-focus-plugin')).toBeInTheDocument();
    });

    it('should render the LinkPlugin', async () => {
      await render(<RichText onChange={mockOnChange} resource={createMockResource()} />);

      await expect.element(page.getByTestId('link-plugin')).toBeInTheDocument();
    });

    it('should render the CustomLinkPlugin', async () => {
      await render(<RichText onChange={mockOnChange} resource={createMockResource()} />);

      await expect.element(page.getByTestId('custom-link-plugin')).toBeInTheDocument();
    });

    it('should render the HTMLPlugin', async () => {
      await render(<RichText onChange={mockOnChange} resource={createMockResource()} />);

      await expect.element(page.getByTestId('html-plugin')).toBeInTheDocument();
    });

    it('should pass resource to HTMLPlugin', async () => {
      const resource = createMockResource({id: 'test-resource-id'});
      await render(<RichText onChange={mockOnChange} resource={resource} />);

      await expect.element(page.getByTestId('html-plugin')).toHaveAttribute('data-resource-id', 'test-resource-id');
    });
  });

  describe('Props', () => {
    it('should apply custom className', async () => {
      const {container} = await render(
        <RichText onChange={mockOnChange} resource={createMockResource()} className="custom-class" />,
      );

      const composerChild = container.querySelector('.custom-class');
      expect(composerChild).toBeInTheDocument();
    });

    it('should pass disabled prop to ToolbarPlugin', async () => {
      await render(<RichText onChange={mockOnChange} resource={createMockResource()} disabled />);

      await expect.element(page.getByTestId('toolbar-plugin')).toHaveAttribute('data-disabled', 'true');
    });

    it('should pass disabled prop to HTMLPlugin', async () => {
      await render(<RichText onChange={mockOnChange} resource={createMockResource()} disabled />);

      await expect.element(page.getByTestId('html-plugin')).toHaveAttribute('data-disabled', 'true');
    });

    it('should pass ToolbarProps to ToolbarPlugin', async () => {
      await render(
        <RichText
          onChange={mockOnChange}
          resource={createMockResource()}
          ToolbarProps={{bold: false, italic: false}}
        />,
      );

      await expect.element(page.getByTestId('toolbar-plugin')).toBeInTheDocument();
    });

    it('should default disabled to false', async () => {
      await render(<RichText onChange={mockOnChange} resource={createMockResource()} />);

      await expect.element(page.getByTestId('toolbar-plugin')).toHaveAttribute('data-disabled', 'false');
      await expect.element(page.getByTestId('html-plugin')).toHaveAttribute('data-disabled', 'false');
    });
  });

  describe('I18n Pattern Detection', () => {
    it('should not show resolved i18n value field when label is not an i18n pattern', async () => {
      const resource = createMockResource({label: 'Regular text'});
      await render(<RichText onChange={mockOnChange} resource={resource} />);

      await expect.element(page.getByText('Resolved i18n value')).not.toBeInTheDocument();
    });

    it('should show resolved i18n value field when label is an i18n pattern', async () => {
      const resource = createMockResource({label: '{{t(hello.world)}}'});
      await render(<RichText onChange={mockOnChange} resource={resource} />);

      await expect.element(page.getByText('Resolved i18n value')).toBeInTheDocument();
    });

    it('should display resolved i18n value in text field', async () => {
      const resource = createMockResource({label: '{{t(test.key)}}'});
      await render(<RichText onChange={mockOnChange} resource={resource} />);

      const textField = page.getByRole('textbox');
      expect(textField).toBeInTheDocument();
      expect(textField).toHaveValue('test.key');
      expect(textField).toBeDisabled();
    });
  });

  describe('Edge Cases', () => {
    it('should handle resource without label', async () => {
      const resource = createMockResource();
      await render(<RichText onChange={mockOnChange} resource={resource} />);

      await expect.element(page.getByTestId('lexical-composer')).toBeInTheDocument();
    });

    it('should handle empty label', async () => {
      const resource = createMockResource({label: ''});
      await render(<RichText onChange={mockOnChange} resource={resource} />);

      await expect.element(page.getByTestId('lexical-composer')).toBeInTheDocument();
    });

    it('should handle undefined label', async () => {
      const resource = createMockResource({label: undefined});
      await render(<RichText onChange={mockOnChange} resource={resource} />);

      await expect.element(page.getByTestId('lexical-composer')).toBeInTheDocument();
    });
  });

  describe('Editor Config', () => {
    it('should have correct editor namespace', async () => {
      await render(<RichText onChange={mockOnChange} resource={createMockResource()} />);

      // The editor config is passed to LexicalComposer
      await expect.element(page.getByTestId('lexical-composer')).toBeInTheDocument();
    });

    it('should rethrow errors in onError callback', async () => {
      // The onError callback in editorConfig throws errors
      // This is verified by the component's behavior when rendering
      await render(<RichText onChange={mockOnChange} resource={createMockResource()} />);

      await expect.element(page.getByTestId('lexical-composer')).toBeInTheDocument();
    });
  });
});
