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

import {describe, it, expect, vi, beforeEach} from 'vitest';
import {render, screen} from '@thunder/test-utils';
import SignInSlogan from '../../../components/SignIn/SignInSlogan';

// Mock useBranding
const mockUseBranding = vi.fn();
vi.mock('@thunder/shared-branding', () => ({
  // eslint-disable-next-line @typescript-eslint/no-unsafe-return
  useBranding: () => mockUseBranding(),
}));

describe('SignInSlogan', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseBranding.mockReturnValue({
      images: null,
    });
  });

  it('renders without crashing', () => {
    const {container} = render(<SignInSlogan />);
    expect(container).toBeInTheDocument();
  });

  it('renders all slogan items', () => {
    render(<SignInSlogan />);
    expect(screen.getByText('Flexible Identity Platform')).toBeInTheDocument();
    expect(screen.getByText('Zero-trust Security')).toBeInTheDocument();
    expect(screen.getByText('Developer-first Experience')).toBeInTheDocument();
    expect(screen.getByText('Extensible & Enterprise-ready')).toBeInTheDocument();
  });

  it('renders item descriptions', () => {
    render(<SignInSlogan />);
    expect(
      screen.getByText(/Centralizes identity management/),
    ).toBeInTheDocument();
    expect(
      screen.getByText(/Leverage adaptive authentication/),
    ).toBeInTheDocument();
    expect(
      screen.getByText(/Configure auth flows and manage organizations/),
    ).toBeInTheDocument();
    expect(
      screen.getByText(/Built for scale/),
    ).toBeInTheDocument();
  });

  it('uses branded logo when available', () => {
    mockUseBranding.mockReturnValue({
      images: {
        logo: {
          primary: {
            url: 'https://example.com/branded-logo.png',
          },
        },
      },
    });
    render(<SignInSlogan />);
    // Component should render with branded logo
    expect(screen.getByText('Flexible Identity Platform')).toBeInTheDocument();
  });

  it('uses default logo when no branded logo', () => {
    mockUseBranding.mockReturnValue({
      images: null,
    });
    render(<SignInSlogan />);
    // Component should render with default logo
    expect(screen.getByText('Flexible Identity Platform')).toBeInTheDocument();
  });

  it('uses default logo when images object exists but no logo', () => {
    mockUseBranding.mockReturnValue({
      images: {},
    });
    render(<SignInSlogan />);
    expect(screen.getByText('Flexible Identity Platform')).toBeInTheDocument();
  });

  it('uses default logo when logo object exists but no primary', () => {
    mockUseBranding.mockReturnValue({
      images: {
        logo: {},
      },
    });
    render(<SignInSlogan />);
    expect(screen.getByText('Flexible Identity Platform')).toBeInTheDocument();
  });

  it('uses default logo when primary object exists but no url', () => {
    mockUseBranding.mockReturnValue({
      images: {
        logo: {
          primary: {},
        },
      },
    });
    render(<SignInSlogan />);
    expect(screen.getByText('Flexible Identity Platform')).toBeInTheDocument();
  });
});
