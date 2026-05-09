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

import {render} from '@testing-library/react';
import {describe, it, expect, vi, beforeEach} from 'vitest';
import type {FlowComponent} from '../../../../models/flow';
import RichTextAdapter from '../RichTextAdapter';

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

vi.mock('../../../../contexts/Design/useDesign', () => ({
  default: () => mockUseDesign(),
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
    const {getByTestId} = render(<RichTextAdapter component={baseComponent} resolve={(s: string | undefined) => s} />);
    expect(getByTestId('rich-text-box').innerHTML).toBe('<p>Hello <strong>World</strong></p>');
  });

  it('uses resolved label from resolve function', () => {
    const {getByTestId} = render(<RichTextAdapter component={baseComponent} resolve={() => '<em>Resolved</em>'} />);
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
    const {getByTestId} = render(<RichTextAdapter component={baseComponent} resolve={(s: string | undefined) => s} />);
    expect(getByTestId('rich-text-box')).toHaveAttribute('data-align', 'center');
  });

  it('aligns text to left when isDesignEnabled is false', () => {
    const {getByTestId} = render(<RichTextAdapter component={baseComponent} resolve={(s: string | undefined) => s} />);
    expect(getByTestId('rich-text-box')).toHaveAttribute('data-align', 'left');
  });

  describe('sign-up URL handling', () => {
    const signUpLabel = '<p>Don\'t have an account? <a href="{{meta(application.sign_up_url)}}">Sign up</a></p>';
    const signUpComponent: FlowComponent = {
      id: 'signup-richtext',
      type: 'RICH_TEXT',
      label: signUpLabel,
    };

    it('returns null when registration is disabled', () => {
      const resolve = (template: string | undefined) =>
        template?.includes('isRegistrationFlowEnabled') ? 'false' : template;

      const {queryByTestId} = render(
        <RichTextAdapter component={signUpComponent} resolve={resolve} signUpFallbackUrl="/signup" />,
      );
      expect(queryByTestId('rich-text-box')).not.toBeInTheDocument();
    });

    it('renders the sign-up link when registration is enabled and the server resolves the URL', () => {
      const resolve = (template: string | undefined) => {
        if (template?.includes('isRegistrationFlowEnabled')) return 'true';
        return template?.replace('{{meta(application.sign_up_url)}}', '/custom/signup');
      };

      const {getByTestId} = render(<RichTextAdapter component={signUpComponent} resolve={resolve} />);
      const box = getByTestId('rich-text-box');
      expect(box).toBeInTheDocument();
      expect(box.innerHTML).toContain('/custom/signup');
    });

    it('uses signUpFallbackUrl when the server does not resolve the sign-up URL template', () => {
      const resolve = (template: string | undefined) =>
        template?.includes('isRegistrationFlowEnabled') ? 'true' : template;

      const {getByTestId} = render(
        <RichTextAdapter component={signUpComponent} resolve={resolve} signUpFallbackUrl="/signup?client_id=abc" />,
      );
      expect(getByTestId('rich-text-box').innerHTML).toContain('/signup?client_id=abc');
    });

    it('renders sign-up content without href substitution when signUpFallbackUrl is not provided', () => {
      const resolve = (template: string | undefined) =>
        template?.includes('isRegistrationFlowEnabled') ? 'true' : template;

      const {getByTestId} = render(<RichTextAdapter component={signUpComponent} resolve={resolve} />);
      // Component renders (registration enabled) but no fallback URL is substituted
      expect(getByTestId('rich-text-box')).toBeInTheDocument();
      expect(getByTestId('rich-text-box').innerHTML).not.toContain('/signup?');
    });
  });
});
