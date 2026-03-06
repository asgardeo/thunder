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
import SignIn from '../SignIn';

// Mock child components
vi.mock('../SignInBox', () => ({
  default: () => <div data-testid="signin-box">SignInBox</div>,
}));

vi.mock('../SignInSlogan', () => ({
  default: () => <div data-testid="signin-slogan">SignInSlogan</div>,
}));

// Mock useDesign hook
const mockUseDesign = vi.fn();
vi.mock('@thunder/shared-design', () => ({
  // eslint-disable-next-line @typescript-eslint/no-unsafe-return
  useDesign: () => mockUseDesign(),
}));

describe('SignIn', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseDesign.mockReturnValue({
      isDesignEnabled: false,
    });
  });

  it('renders without crashing', () => {
    const {container} = render(<SignIn />);
    expect(container).toBeInTheDocument();
  });

  it('renders SignInBox component', () => {
    render(<SignIn />);
    expect(screen.getByTestId('signin-box')).toBeInTheDocument();
  });

  it('shows SignInSlogan when design is not enabled', () => {
    mockUseDesign.mockReturnValue({
      isDesignEnabled: false,
    });
    render(<SignIn />);
    expect(screen.getByTestId('signin-slogan')).toBeInTheDocument();
  });

  it('hides SignInSlogan when design is enabled', () => {
    mockUseDesign.mockReturnValue({
      isDesignEnabled: true,
    });
    render(<SignIn />);
    expect(screen.queryByTestId('signin-slogan')).not.toBeInTheDocument();
  });
});
