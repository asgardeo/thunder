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

import {describe, expect, it, beforeEach} from 'vitest';
import cn, {setCnPrefix, getCnPrefix} from '../cn';

describe('cn', () => {
  beforeEach(() => {
    setCnPrefix('AwesomeProduct');
  });

  it('should prefix a single class name with the default prefix', () => {
    expect(cn('SignIn--root')).toBe('AwesomeProductSignIn--root');
  });

  it('should join multiple class names with the prefix', () => {
    expect(cn('SignInBox--root', 'SignInBox--paper')).toBe(
      'AwesomeProductSignInBox--root AwesomeProductSignInBox--paper',
    );
  });

  it('should filter out falsy values', () => {
    // eslint-disable-next-line no-constant-binary-expression
    expect(cn('SignIn--root', false && 'SignIn--primary')).toBe('AwesomeProductSignIn--root');
    expect(cn('SignIn--root', null)).toBe('AwesomeProductSignIn--root');
    expect(cn('SignIn--root', undefined)).toBe('AwesomeProductSignIn--root');
    expect(cn('SignIn--root', 0)).toBe('AwesomeProductSignIn--root');
  });

  it('should include truthy conditional classes', () => {
    // eslint-disable-next-line no-constant-binary-expression
    expect(cn('SignIn--root', true && 'SignIn--primary')).toBe(
      'AwesomeProductSignIn--root AwesomeProductSignIn--primary',
    );
  });

  it('should return an empty string when no truthy classes are provided', () => {
    expect(cn(false, null, undefined, 0)).toBe('');
  });

  it('should return an empty string when called with no arguments', () => {
    expect(cn()).toBe('');
  });
});

describe('setCnPrefix / getCnPrefix', () => {
  beforeEach(() => {
    setCnPrefix('AwesomeProduct');
  });

  it('should change the prefix used by cn()', () => {
    setCnPrefix('JoyUI');
    expect(cn('SignIn--root')).toBe('JoyUISignIn--root');
  });

  it('should return the current prefix via getCnPrefix()', () => {
    expect(getCnPrefix()).toBe('AwesomeProduct');
    setCnPrefix('Custom');
    expect(getCnPrefix()).toBe('Custom');
  });
});
