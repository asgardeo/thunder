/**
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
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

import type {JsonValue} from '../models/json';
import generateResourceId from './generateResourceId';

const replaceIds = (obj: JsonValue, matcher: string): JsonValue => {
  if (Array.isArray(obj)) {
    return obj.map((item) => replaceIds(item, matcher));
  }
  if (typeof obj === 'object' && obj !== null) {
    const objAsRecord = obj;

    return Object.fromEntries(
      Object.entries(objAsRecord).map(([key, value]) => {
        if (key === 'id' && value === `{{${matcher}}}`) {
          const typeValue = objAsRecord.type;
          const type = typeof typeValue === 'string' ? typeValue.toLowerCase() : 'component';

          return [key, generateResourceId(type)];
        }

        return [key, replaceIds(value, matcher)];
      }),
    );
  }

  return obj;
};

const generateIdsForResources = <T = unknown>(resources: T, matcher = 'ID'): T =>
  replaceIds(resources as JsonValue, matcher) as T;

export default generateIdsForResources;
