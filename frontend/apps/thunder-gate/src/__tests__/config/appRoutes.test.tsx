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

import {describe, it, expect} from 'vitest';
import appRoutes, {type AppRoute} from '../../config/appRoutes';
import ROUTES from '../../constants/routes';

describe('appRoutes', () => {
  it('exports an array of routes', () => {
    expect(Array.isArray(appRoutes)).toBe(true);
  });

  it('has a root route', () => {
    const rootRoute = appRoutes.find((route) => route.path === ROUTES.ROOT);
    expect(rootRoute).toBeDefined();
  });

  it('root route has DefaultLayout as element', () => {
    const rootRoute = appRoutes.find((route) => route.path === ROUTES.ROOT);
    expect(rootRoute?.element).toBeDefined();
  });

  it('root route has children routes', () => {
    const rootRoute = appRoutes.find((route) => route.path === ROUTES.ROOT);
    expect(rootRoute?.children).toBeDefined();
    expect(Array.isArray(rootRoute?.children)).toBe(true);
  });

  it('has sign-in route in children', () => {
    const rootRoute = appRoutes.find((route) => route.path === ROUTES.ROOT);
    const signInRoute = rootRoute?.children?.find((route) => route.path === ROUTES.AUTH.SIGN_IN);
    expect(signInRoute).toBeDefined();
  });

  it('has sign-up route in children', () => {
    const rootRoute = appRoutes.find((route) => route.path === ROUTES.ROOT);
    const signUpRoute = rootRoute?.children?.find((route) => route.path === ROUTES.AUTH.SIGN_UP);
    expect(signUpRoute).toBeDefined();
  });

  it('has invite route in children', () => {
    const rootRoute = appRoutes.find((route) => route.path === ROUTES.ROOT);
    const inviteRoute = rootRoute?.children?.find((route) => route.path === ROUTES.AUTH.INVITE);
    expect(inviteRoute).toBeDefined();
  });

  it('has redirect from root to sign-in', () => {
    const rootRoute = appRoutes.find((route) => route.path === ROUTES.ROOT);
    const redirectRoute = rootRoute?.children?.find((route) => route.path === ROUTES.ROOT);
    expect(redirectRoute).toBeDefined();
    expect(redirectRoute?.element).toBeDefined();
  });

  it('AppRoute interface allows children of same type', () => {
    const testRoute: AppRoute = {
      path: '/test',
      element: <div>Test</div>,
      children: [
        {
          path: '/test/child',
          element: <div>Child</div>,
        },
      ],
    };
    expect(testRoute.children).toHaveLength(1);
  });
});
