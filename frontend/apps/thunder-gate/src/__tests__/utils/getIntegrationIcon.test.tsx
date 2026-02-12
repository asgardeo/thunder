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

import {describe, it, expect, vi} from 'vitest';
import {render, screen} from '@testing-library/react';
import getIntegrationIcon from '../../utils/getIntegrationIcon';

// Mock the icons
vi.mock('@wso2/oxygen-ui-icons-react', () => ({
  Google: () => <span data-testid="google-icon">Google Icon</span>,
  GitHub: () => <span data-testid="github-icon">GitHub Icon</span>,
}));

describe('getIntegrationIcon', () => {
  it('returns Google icon when label contains google', () => {
    const icon = getIntegrationIcon('Continue with google', '');
    expect(icon).not.toBeNull();

    render(<div>{icon}</div>);
    expect(screen.getByTestId('google-icon')).toBeInTheDocument();
  });

  it('returns Google icon when image contains google', () => {
    const icon = getIntegrationIcon('', 'assets/images/google.svg');
    expect(icon).not.toBeNull();

    render(<div>{icon}</div>);
    expect(screen.getByTestId('google-icon')).toBeInTheDocument();
  });

  it('returns GitHub icon when label contains github', () => {
    const icon = getIntegrationIcon('Sign in with github', '');
    expect(icon).not.toBeNull();

    render(<div>{icon}</div>);
    expect(screen.getByTestId('github-icon')).toBeInTheDocument();
  });

  it('returns GitHub icon when image contains github', () => {
    const icon = getIntegrationIcon('', 'icons/github-icon.png');
    expect(icon).not.toBeNull();

    render(<div>{icon}</div>);
    expect(screen.getByTestId('github-icon')).toBeInTheDocument();
  });

  it('returns null for unknown provider', () => {
    const icon = getIntegrationIcon('Sign in with Facebook', 'facebook.svg');
    expect(icon).toBeNull();
  });

  it('returns null for empty strings', () => {
    const icon = getIntegrationIcon('', '');
    expect(icon).toBeNull();
  });

  it('uses case-sensitive matching for labels', () => {
    const icon = getIntegrationIcon('GOOGLE', '');
    // The function uses case-sensitive includes() check
    // So "GOOGLE" won't match "google"
    expect(icon).toBeNull();
  });

  it('returns Google icon when label matches google even if image contains github', () => {
    // Due to short-circuit evaluation in OR logic (label.includes() || image.includes()),
    // when label contains 'google', the image is never checked
    const icon = getIntegrationIcon('google login', 'github.svg');
    expect(icon).not.toBeNull();

    render(<div>{icon}</div>);
    expect(screen.getByTestId('google-icon')).toBeInTheDocument();
  });

  it('checks Google before GitHub regardless of source', () => {
    // The function checks for 'google' first (in both label and image)
    // before checking for 'github', so image containing 'google' wins
    const icon = getIntegrationIcon('github login', 'google.svg');
    expect(icon).not.toBeNull();

    render(<div>{icon}</div>);
    // Google check happens first, and image contains 'google'
    expect(screen.getByTestId('google-icon')).toBeInTheDocument();
  });

  it('falls back to image when label does not match', () => {
    const icon = getIntegrationIcon('Sign in', 'github-logo.png');
    expect(icon).not.toBeNull();

    render(<div>{icon}</div>);
    expect(screen.getByTestId('github-icon')).toBeInTheDocument();
  });
});
