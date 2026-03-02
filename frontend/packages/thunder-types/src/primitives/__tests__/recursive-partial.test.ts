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

import {expectTypeOf} from 'vitest';
import type {RecursivePartial} from '../recursive-partial';

describe('RecursivePartial', () => {
  it('should keep primitive types unchanged', () => {
    expectTypeOf<RecursivePartial<string>>().toEqualTypeOf<string>();
    expectTypeOf<RecursivePartial<number>>().toEqualTypeOf<number>();
    expectTypeOf<RecursivePartial<boolean>>().toEqualTypeOf<boolean>();
    expectTypeOf<RecursivePartial<null>>().toEqualTypeOf<null>();
    expectTypeOf<RecursivePartial<undefined>>().toEqualTypeOf<undefined>();
  });

  it('should keep function types unchanged', () => {
    type Fn = (x: number) => string;
    expectTypeOf<RecursivePartial<Fn>>().toEqualTypeOf<Fn>();

    type VoidFn = () => void;
    expectTypeOf<RecursivePartial<VoidFn>>().toEqualTypeOf<VoidFn>();
  });

  it('should make all object properties optional', () => {
    interface Input {
      a: string;
      b: number;
    }
    interface Expected {
      a?: string;
      b?: number;
    }
    expectTypeOf<RecursivePartial<Input>>().toEqualTypeOf<Expected>();
  });

  it('should recursively make nested object properties optional', () => {
    interface Input {
      a: {b: {c: string}};
    }
    interface Expected {
      a?: {b?: {c?: string}};
    }
    expectTypeOf<RecursivePartial<Input>>().toEqualTypeOf<Expected>();
  });

  it('should keep function-valued properties as-is while making the key optional', () => {
    type Fn = () => void;
    interface Input {
      fn: Fn;
      name: string;
    }
    interface Expected {
      fn?: Fn;
      name?: string;
    }
    expectTypeOf<RecursivePartial<Input>>().toEqualTypeOf<Expected>();
  });

  it('should handle objects with mixed primitive and nested object properties', () => {
    interface Input {
      id: number;
      meta: {
        label: string;
        active: boolean;
      };
    }
    interface Expected {
      id?: number;
      meta?: {
        label?: string;
        active?: boolean;
      };
    }
    expectTypeOf<RecursivePartial<Input>>().toEqualTypeOf<Expected>();
  });
});
