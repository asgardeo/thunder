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

import {beforeEach, describe, expect, it, vi} from 'vitest';
import {render, screen} from '@testing-library/react';
import JwtPreview from '../JwtPreview';

const deltaDecorationsSpy = vi.fn().mockReturnValue([]);
const disposeSpy = vi.fn();
const onDidChangeModelContentSpy = vi.fn().mockReturnValue({dispose: disposeSpy});

// Mock Monaco editor to call onMount synchronously with stub editor/monaco objects so
// that handleMount, applyDecorations, and the content-listener setup are exercised.
vi.mock('@monaco-editor/react', () => ({
  default: ({value, onMount}: {value: string; onMount?: (editor: unknown, monaco: unknown) => void}) => {
    if (onMount) {
      const mockModel = {getValue: () => value};
      const mockEditor = {
        getModel: () => mockModel,
        deltaDecorations: deltaDecorationsSpy,
        onDidChangeModelContent: onDidChangeModelContentSpy,
      };
      const mockMonaco = {
        // eslint-disable-next-line @typescript-eslint/no-extraneous-class
        Range: class Range {
          startLine: number;

          startCol: number;

          endLine: number;

          endCol: number;

          constructor(startLine: number, startCol: number, endLine: number, endCol: number) {
            this.startLine = startLine;
            this.startCol = startCol;
            this.endLine = endLine;
            this.endCol = endCol;
          }
        },
      };
      onMount(mockEditor, mockMonaco);
    }
    return <pre data-testid="monaco-editor">{value}</pre>;
  },
}));

describe('JwtPreview', () => {
  beforeEach(() => {
    deltaDecorationsSpy.mockClear();
    disposeSpy.mockClear();
    onDidChangeModelContentSpy.mockClear();
  });

  it('renders the title text', () => {
    render(<JwtPreview title="Access Token" payload={{sub: 'user-123'}} />);

    expect(screen.getByText('Access Token')).toBeInTheDocument();
  });

  it('renders the payload as JSON in the editor', () => {
    const payload = {sub: 'user-123', iss: 'https://example.com'};
    render(<JwtPreview title="Access Token" payload={payload} />);

    const editor = screen.getByTestId('monaco-editor');
    const content = editor.textContent ?? '';

    expect(content).toContain('"sub"');
    expect(content).toContain('"user-123"');
    expect(content).toContain('"iss"');
    expect(content).toContain('"https://example.com"');
  });

  it('renders without errors when defaultClaims prop is provided', () => {
    expect(() =>
      render(
        <JwtPreview
          title="ID Token"
          payload={{sub: 'user-123', iss: 'https://example.com'}}
          defaultClaims={['sub', 'iss']}
        />,
      ),
    ).not.toThrow();
  });

  it('renders the JWT logo SVG element', () => {
    const {container} = render(<JwtPreview title="Access Token" payload={{sub: 'user-123'}} />);

    expect(container.querySelector('svg')).toBeInTheDocument();
  });

  it('renders an empty JSON object when payload is empty', () => {
    render(<JwtPreview title="Access Token" payload={{}} />);

    const editor = screen.getByTestId('monaco-editor');

    expect(editor.textContent).toContain('{}');
  });

  it('calls deltaDecorations on mount', () => {
    render(<JwtPreview title="Access Token" payload={{sub: 'user-123'}} defaultClaims={[]} />);

    expect(screen.getByTestId('monaco-editor')).toBeInTheDocument();
    expect(deltaDecorationsSpy).toHaveBeenCalled();
  });

  it('calls deltaDecorations again when defaultClaims prop changes', () => {
    const {rerender} = render(
      <JwtPreview title="Access Token" payload={{sub: 'user-123', iss: 'https://example.com'}} defaultClaims={[]} />,
    );

    const callsAfterInitialRender = deltaDecorationsSpy.mock.calls.length;

    rerender(
      <JwtPreview
        title="Access Token"
        payload={{sub: 'user-123', iss: 'https://example.com'}}
        defaultClaims={['sub', 'iss']}
      />,
    );

    expect(screen.getByTestId('monaco-editor')).toBeInTheDocument();
    expect(deltaDecorationsSpy.mock.calls.length).toBeGreaterThan(callsAfterInitialRender);
  });

  it('disposes content listener on unmount', () => {
    const {unmount} = render(<JwtPreview title="Access Token" payload={{sub: 'user-123'}} defaultClaims={['sub']} />);

    unmount();

    expect(disposeSpy).toHaveBeenCalled();
  });
});
