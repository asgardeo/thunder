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
import {render, screen} from '@testing-library/react';
import TextAdapter from '../TextAdapter';
import type {FlowComponent} from '../../../../models/flow';

const mockUseDesign = vi.fn();

vi.mock('@wso2/oxygen-ui', () => ({
  Typography: ({children, variant, sx}: any) => (
    <p data-testid="typography" data-variant={variant} data-align={sx?.textAlign}>
      {children}
    </p>
  ),
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({t: (key: string) => key}),
}));

vi.mock('@thunder/shared-design', () => ({
  useDesign: () => mockUseDesign(),
  mapEmbeddedFlowTextVariant: (variant: string) => {
    if (variant === 'H1') return 'h1';
    if (variant === 'H2') return 'h2';
    return 'body1';
  },
}));

const baseComponent: FlowComponent = {
  id: 'text-1',
  type: 'TEXT',
  label: 'Hello World',
};

describe('TextAdapter', () => {
  beforeEach(() => {
    mockUseDesign.mockReturnValue({isDesignEnabled: false});
  });

  it('renders the resolved label', () => {
    render(<TextAdapter component={baseComponent} resolve={(s) => s} />);
    expect(screen.getByText('Hello World')).toBeInTheDocument();
  });

  it('passes resolved label through the resolve function', () => {
    const component = {...baseComponent, label: '{{t(greet)}}'};
    render(<TextAdapter component={component} resolve={() => 'Resolved Text'} />);
    expect(screen.getByText('Resolved Text')).toBeInTheDocument();
  });

  it('maps H1 variant to h1 via mapEmbeddedFlowTextVariant', () => {
    const component = {...baseComponent, variant: 'H1'};
    render(<TextAdapter component={component} resolve={(s) => s} />);
    expect(screen.getByTestId('typography')).toHaveAttribute('data-variant', 'h1');
  });

  it('maps H2 variant to h2 via mapEmbeddedFlowTextVariant', () => {
    const component = {...baseComponent, variant: 'H2'};
    render(<TextAdapter component={component} resolve={(s) => s} />);
    expect(screen.getByTestId('typography')).toHaveAttribute('data-variant', 'h2');
  });

  it('uses body1 for unknown variant', () => {
    render(<TextAdapter component={baseComponent} resolve={(s) => s} />);
    expect(screen.getByTestId('typography')).toHaveAttribute('data-variant', 'body1');
  });

  it('aligns text to center when isDesignEnabled is true', () => {
    mockUseDesign.mockReturnValue({isDesignEnabled: true});
    render(<TextAdapter component={baseComponent} resolve={(s) => s} />);
    expect(screen.getByTestId('typography')).toHaveAttribute('data-align', 'center');
  });

  it('aligns text to left when isDesignEnabled is false', () => {
    render(<TextAdapter component={baseComponent} resolve={(s) => s} />);
    expect(screen.getByTestId('typography')).toHaveAttribute('data-align', 'left');
  });
});
