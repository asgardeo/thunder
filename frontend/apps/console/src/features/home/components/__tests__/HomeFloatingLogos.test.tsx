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
import {render} from '@thunder/test-utils';
import HomeFloatingLogos from '../HomeFloatingLogos';

let mockMode: 'light' | 'dark' | 'system' = 'light';

vi.mock('@wso2/oxygen-ui', async () => {
  const actual = await vi.importActual<typeof import('@wso2/oxygen-ui')>('@wso2/oxygen-ui');
  return {
    ...actual,
    useColorScheme: () => ({mode: mockMode}),
  };
});

describe('HomeFloatingLogos', () => {
  beforeEach(() => {
    mockMode = 'light';
  });

  it('renders without crashing in light mode', () => {
    const {container} = render(<HomeFloatingLogos />);
    expect(container.firstChild).not.toBeNull();
  });

  it('renders without crashing in dark mode', () => {
    mockMode = 'dark';
    const {container} = render(<HomeFloatingLogos />);
    expect(container.firstChild).not.toBeNull();
  });

  it('renders SVG logo elements', () => {
    const {container} = render(<HomeFloatingLogos />);
    const svgs = container.querySelectorAll('svg');
    expect(svgs.length).toBeGreaterThan(0);
  });

  it('renders in system mode without crashing', () => {
    mockMode = 'system';
    const {container} = render(<HomeFloatingLogos />);
    expect(container.firstChild).not.toBeNull();
  });
});
