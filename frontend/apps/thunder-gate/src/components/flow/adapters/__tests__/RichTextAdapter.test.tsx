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

/* eslint-disable @typescript-eslint/no-explicit-any */
/* eslint-disable @typescript-eslint/no-unsafe-assignment */
/* eslint-disable @typescript-eslint/no-unsafe-member-access */
/* eslint-disable @typescript-eslint/no-unsafe-return */

import {describe, it, expect, vi, beforeEach} from 'vitest';
import {render} from '@testing-library/react';
import RichTextAdapter from '../RichTextAdapter';
import type {FlowComponent} from '../../../../models/flow';

const mockUseDesign = vi.fn();

vi.mock('@wso2/oxygen-ui', () => ({
  Box: ({sx, dangerouslySetInnerHTML}: any) => (
    <div
      data-testid="rich-text-box"
      data-align={sx?.textAlign}
      // eslint-disable-next-line react/no-danger
      dangerouslySetInnerHTML={dangerouslySetInnerHTML}
    />
  ),
}));

vi.mock('@thunder/shared-design', () => ({
  useDesign: () => mockUseDesign(),
}));

const baseComponent: FlowComponent = {
  id: 'rich-1',
  type: 'RICH_TEXT',
  label: '<p>Hello <strong>World</strong></p>',
};

describe('RichTextAdapter', () => {
  beforeEach(() => {
    mockUseDesign.mockReturnValue({isDesignEnabled: false});
  });

  it('renders resolved HTML content', () => {
    const {getByTestId} = render(<RichTextAdapter component={baseComponent} resolve={(s) => s} />);
    expect(getByTestId('rich-text-box').innerHTML).toBe('<p>Hello <strong>World</strong></p>');
  });

  it('uses resolved label from resolve function', () => {
    const {getByTestId} = render(
      <RichTextAdapter component={baseComponent} resolve={() => '<em>Resolved</em>'} />,
    );
    expect(getByTestId('rich-text-box').innerHTML).toBe('<em>Resolved</em>');
  });

  it('falls back to component.label when resolve returns undefined', () => {
    const {getByTestId} = render(<RichTextAdapter component={baseComponent} resolve={() => undefined} />);
    expect(getByTestId('rich-text-box').innerHTML).toBe('<p>Hello <strong>World</strong></p>');
  });

  it('renders empty string when resolve returns undefined and label is not a string', () => {
    const component = {...baseComponent, label: undefined};
    const {getByTestId} = render(<RichTextAdapter component={component} resolve={() => undefined} />);
    expect(getByTestId('rich-text-box').innerHTML).toBe('');
  });

  it('aligns text to center when isDesignEnabled is true', () => {
    mockUseDesign.mockReturnValue({isDesignEnabled: true});
    const {getByTestId} = render(<RichTextAdapter component={baseComponent} resolve={(s) => s} />);
    expect(getByTestId('rich-text-box')).toHaveAttribute('data-align', 'center');
  });

  it('aligns text to left when isDesignEnabled is false', () => {
    const {getByTestId} = render(<RichTextAdapter component={baseComponent} resolve={(s) => s} />);
    expect(getByTestId('rich-text-box')).toHaveAttribute('data-align', 'left');
  });
});
