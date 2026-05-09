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
import {parseCommaSeparated} from '../utils';

describe('parseCommaSeparated', () => {
  it('should parse comma-separated values', () => {
    expect(parseCommaSeparated('a, b, c')).toEqual(['a', 'b', 'c']);
  });

  it('should trim whitespace from values', () => {
    expect(parseCommaSeparated('  foo ,  bar  , baz  ')).toEqual(['foo', 'bar', 'baz']);
  });

  it('should filter out empty strings', () => {
    expect(parseCommaSeparated('a,,b,,,c')).toEqual(['a', 'b', 'c']);
  });

  it('should return empty array for empty string', () => {
    expect(parseCommaSeparated('')).toEqual([]);
  });

  it('should return single item for no commas', () => {
    expect(parseCommaSeparated('single')).toEqual(['single']);
  });

  it('should handle trailing comma', () => {
    expect(parseCommaSeparated('a, b,')).toEqual(['a', 'b']);
  });

  it('should handle leading comma', () => {
    expect(parseCommaSeparated(',a, b')).toEqual(['a', 'b']);
  });

  it('should handle whitespace-only values as empty', () => {
    expect(parseCommaSeparated('a,   , b')).toEqual(['a', 'b']);
  });
});
