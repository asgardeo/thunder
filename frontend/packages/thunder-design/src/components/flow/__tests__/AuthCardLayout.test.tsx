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

import {screen, cleanup} from '@testing-library/react';
import {TEST_CN_PREFIX} from '@thunder/test-utils';
import {describe, it, expect, afterEach} from 'vitest';
import renderWithProviders from '../../../test/renderWithProviders';
import AuthCardLayout from '../AuthCardLayout';

afterEach(() => {
  cleanup();
});

describe('AuthCardLayout', () => {
  it('renders children inside a paper container', () => {
    renderWithProviders(
      <AuthCardLayout>
        <span>Child content</span>
      </AuthCardLayout>,
    );
    expect(screen.getByText('Child content')).toBeTruthy();
  });

  describe('variant prop', () => {
    it('applies product prefix CSS root class when variant is provided', () => {
      const {container} = renderWithProviders(
        <AuthCardLayout variant="SignInBox">
          <span>Content</span>
        </AuthCardLayout>,
      );
      expect(container.querySelector(`.${TEST_CN_PREFIX}SignInBox--root`)).toBeTruthy();
    });

    it('applies product prefix CSS paper class when variant is provided', () => {
      const {container} = renderWithProviders(
        <AuthCardLayout variant="SignInBox">
          <span>Content</span>
        </AuthCardLayout>,
      );
      expect(container.querySelector(`.${TEST_CN_PREFIX}SignInBox--paper`)).toBeTruthy();
    });

    it('does not apply product prefix CSS classes when variant is not provided', () => {
      const {container} = renderWithProviders(
        <AuthCardLayout>
          <span>Content</span>
        </AuthCardLayout>,
      );
      const allElements = container.querySelectorAll(`[class*="${TEST_CN_PREFIX}"]`);
      expect(allElements.length).toBe(0);
    });
  });

  describe('logo prop', () => {
    it('renders the logo when logo prop is provided and showLogo is true (default)', () => {
      const logo = {
        src: {light: '/logo-light.svg', dark: '/logo-dark.svg'},
      };
      const {container} = renderWithProviders(
        <AuthCardLayout variant="SignInBox" logo={logo}>
          <span>Content</span>
        </AuthCardLayout>,
      );
      const img = container.querySelector('img');
      expect(img).toBeTruthy();
    });

    it('applies product prefix CSS logo class when variant and logo are provided', () => {
      const logo = {
        src: {light: '/logo-light.svg', dark: '/logo-dark.svg'},
      };
      const {container} = renderWithProviders(
        <AuthCardLayout variant="SignInBox" logo={logo}>
          <span>Content</span>
        </AuthCardLayout>,
      );
      expect(container.querySelector(`.${TEST_CN_PREFIX}SignInBox--logo`)).toBeTruthy();
    });

    it('uses custom alt text when provided in logo', () => {
      const logo = {
        src: {light: '/logo-light.svg', dark: '/logo-dark.svg'},
        alt: {light: 'Custom Light', dark: 'Custom Dark'},
      };
      renderWithProviders(
        <AuthCardLayout logo={logo}>
          <span>Content</span>
        </AuthCardLayout>,
      );
      const img = screen.getByAltText('Custom Light');
      expect(img).toBeTruthy();
    });

    it('uses default alt text when alt is not provided in logo', () => {
      const logo = {
        src: {light: '/logo-light.svg', dark: '/logo-dark.svg'},
      };
      renderWithProviders(
        <AuthCardLayout logo={logo}>
          <span>Content</span>
        </AuthCardLayout>,
      );
      const img = screen.getByAltText('Logo (Light)');
      expect(img).toBeTruthy();
    });

    it('does not render logo when showLogo is false', () => {
      const logo = {
        src: {light: '/logo-light.svg', dark: '/logo-dark.svg'},
      };
      const {container} = renderWithProviders(
        <AuthCardLayout logo={logo} showLogo={false}>
          <span>Content</span>
        </AuthCardLayout>,
      );
      const img = container.querySelector('img');
      expect(img).toBeNull();
    });

    it('does not render logo when logo prop is not provided', () => {
      const {container} = renderWithProviders(
        <AuthCardLayout showLogo>
          <span>Content</span>
        </AuthCardLayout>,
      );
      const img = container.querySelector('img');
      expect(img).toBeNull();
    });
  });

  describe('logoDisplay prop', () => {
    it('renders logo with default logoDisplay (mobile-only) when not specified', () => {
      const logo = {
        src: {light: '/logo-light.svg', dark: '/logo-dark.svg'},
      };
      const {container} = renderWithProviders(
        <AuthCardLayout logo={logo}>
          <span>Content</span>
        </AuthCardLayout>,
      );
      // Logo should still render (display is a CSS concern, not DOM)
      const img = container.querySelector('img');
      expect(img).toBeTruthy();
    });

    it('renders logo with custom logoDisplay', () => {
      const logo = {
        src: {light: '/logo-light.svg', dark: '/logo-dark.svg'},
      };
      const {container} = renderWithProviders(
        <AuthCardLayout logo={logo} logoDisplay={{xs: 'flex', md: 'flex'}}>
          <span>Content</span>
        </AuthCardLayout>,
      );
      const img = container.querySelector('img');
      expect(img).toBeTruthy();
    });
  });

  describe('variant-less rendering', () => {
    it('does not add logo class when variant is not provided but logo is', () => {
      const logo = {
        src: {light: '/logo-light.svg', dark: '/logo-dark.svg'},
      };
      const {container} = renderWithProviders(
        <AuthCardLayout logo={logo}>
          <span>Content</span>
        </AuthCardLayout>,
      );
      const thunderElements = container.querySelectorAll(`[class*="${TEST_CN_PREFIX}"]`);
      expect(thunderElements.length).toBe(0);
    });
  });
});
