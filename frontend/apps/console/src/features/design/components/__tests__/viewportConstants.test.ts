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
import {VIEWPORT_WIDTHS, VIEWPORT_HEIGHTS} from '../viewportConstants';

describe('VIEWPORT_WIDTHS', () => {
  it('returns "85%" for desktop', () => {
    expect(VIEWPORT_WIDTHS.desktop).toBe('85%');
  });

  it('returns "60%" for tablet', () => {
    expect(VIEWPORT_WIDTHS.tablet).toBe('60%');
  });

  it('returns "40%" for mobile', () => {
    expect(VIEWPORT_WIDTHS.mobile).toBe('40%');
  });

  it('has entries for all three viewports', () => {
    expect(Object.keys(VIEWPORT_WIDTHS)).toEqual(expect.arrayContaining(['desktop', 'tablet', 'mobile']));
  });
});

describe('VIEWPORT_HEIGHTS', () => {
  it('returns "85%" for desktop', () => {
    expect(VIEWPORT_HEIGHTS.desktop).toBe('85%');
  });

  it('returns "90%" for tablet', () => {
    expect(VIEWPORT_HEIGHTS.tablet).toBe('90%');
  });

  it('returns "80%" for mobile', () => {
    expect(VIEWPORT_HEIGHTS.mobile).toBe('80%');
  });

  it('has entries for all three viewports', () => {
    expect(Object.keys(VIEWPORT_HEIGHTS)).toEqual(expect.arrayContaining(['desktop', 'tablet', 'mobile']));
  });
});
