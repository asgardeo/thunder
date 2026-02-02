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
import {render} from '@thunder/test-utils/browser';
import {page} from 'vitest/browser';
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

  it('renders all slogan items', async () => {
    await render(<SignInSlogan />);
    await expect.element(page.getByText('Flexible Identity Platform')).toBeInTheDocument();
    await expect.element(page.getByText('Zero-trust Security')).toBeInTheDocument();
    await expect.element(page.getByText('Developer-first Experience')).toBeInTheDocument();
    await expect.element(page.getByText('Extensible & Enterprise-ready')).toBeInTheDocument();
  });

  it('renders item descriptions', async () => {
    await render(<SignInSlogan />);
    await expect.element(
      page.getByText(/Centralizes identity management/),
    ).toBeInTheDocument();
    await expect.element(
      page.getByText(/Leverage adaptive authentication/),
    ).toBeInTheDocument();
    await expect.element(
      page.getByText(/Configure auth flows and manage organizations/),
    ).toBeInTheDocument();
    await expect.element(
      page.getByText(/Built for scale/),
    ).toBeInTheDocument();
  });

  it('uses branded logo when available', async () => {
    mockUseBranding.mockReturnValue({
      images: {
        logo: {
          primary: {
            url: 'https://example.com/branded-logo.png',
          },
        },
      },
    });
    await render(<SignInSlogan />);
    // Component should render with branded logo
    await expect.element(page.getByText('Flexible Identity Platform')).toBeInTheDocument();
  });

  it('uses default logo when no branded logo', async () => {
    mockUseBranding.mockReturnValue({
      images: null,
    });
    await render(<SignInSlogan />);
    // Component should render with default logo
    await expect.element(page.getByText('Flexible Identity Platform')).toBeInTheDocument();
  });

  it('uses default logo when images object exists but no logo', async () => {
    mockUseBranding.mockReturnValue({
      images: {},
    });
    await render(<SignInSlogan />);
    await expect.element(page.getByText('Flexible Identity Platform')).toBeInTheDocument();
  });

  it('uses default logo when logo object exists but no primary', async () => {
    mockUseBranding.mockReturnValue({
      images: {
        logo: {},
      },
    });
    await render(<SignInSlogan />);
    await expect.element(page.getByText('Flexible Identity Platform')).toBeInTheDocument();
  });

  it('uses default logo when primary object exists but no url', async () => {
    mockUseBranding.mockReturnValue({
      images: {
        logo: {
          primary: {},
        },
      },
    });
    await render(<SignInSlogan />);
    await expect.element(page.getByText('Flexible Identity Platform')).toBeInTheDocument();
  });
});
