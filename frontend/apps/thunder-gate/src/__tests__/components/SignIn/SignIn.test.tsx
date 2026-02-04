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
import SignIn from '../../../components/SignIn/SignIn';

// Mock child components
vi.mock('../../../components/SignIn/SignInBox', () => ({
  default: () => <div data-testid="signin-box">SignInBox</div>,
}));

vi.mock('../../../components/SignIn/SignInSlogan', () => ({
  default: () => <div data-testid="signin-slogan">SignInSlogan</div>,
}));

// Mock useBranding hook
const mockUseBranding = vi.fn();
vi.mock('@thunder/shared-branding', () => ({
  // eslint-disable-next-line @typescript-eslint/no-unsafe-return
  useBranding: () => mockUseBranding(),
  LayoutType: {
    LEFT_ALIGNED: 'LEFT_ALIGNED',
    CENTERED: 'CENTERED',
  },
}));

describe('SignIn', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseBranding.mockReturnValue({
      isBrandingEnabled: false,
      layout: null,
    });
  });

  it('renders SignInBox component', async () => {
    await render(<SignIn />);
    await expect.element(page.getByTestId('signin-box')).toBeInTheDocument();
  });

  it('shows SignInSlogan when branding is not enabled', async () => {
    mockUseBranding.mockReturnValue({
      isBrandingEnabled: false,
      layout: null,
    });
    await render(<SignIn />);
    await expect.element(page.getByTestId('signin-slogan')).toBeInTheDocument();
  });

  it('shows SignInSlogan when branding is enabled with LEFT_ALIGNED layout', async () => {
    mockUseBranding.mockReturnValue({
      isBrandingEnabled: true,
      layout: {type: 'LEFT_ALIGNED'},
    });
    await render(<SignIn />);
    await expect.element(page.getByTestId('signin-slogan')).toBeInTheDocument();
  });

  it('hides SignInSlogan when branding is enabled with non-LEFT_ALIGNED layout', async () => {
    mockUseBranding.mockReturnValue({
      isBrandingEnabled: true,
      layout: {type: 'CENTERED'},
    });
    await render(<SignIn />);
    await expect.element(page.getByTestId('signin-slogan')).not.toBeInTheDocument();
  });
});
