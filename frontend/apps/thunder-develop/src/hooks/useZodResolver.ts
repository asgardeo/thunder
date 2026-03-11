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

import type {FieldValues, Resolver} from 'react-hook-form';
import {zodResolver as baseZodResolver} from '@hookform/resolvers/zod';
import type {ZodType} from 'zod';

/**
 * Type-safe wrapper around `zodResolver` from `@hookform/resolvers/zod`.
 *
 * Works around a type incompatibility between `@hookform/resolvers@5.2.2`
 * and `zod@4.3.x` where the internal `_zod.version.minor` changed from
 * `2` to `3`, causing TypeScript to reject the schema argument.
 *
 * @see https://github.com/colinhacks/zod/issues/4992
 */
// eslint-disable-next-line import/prefer-default-export
export const zodResolver = <TFieldValues extends FieldValues>(schema: ZodType<TFieldValues>): Resolver<TFieldValues> =>
  // eslint-disable-next-line @typescript-eslint/no-unsafe-argument, @typescript-eslint/no-explicit-any
  baseZodResolver(schema as any) as Resolver<TFieldValues>;
