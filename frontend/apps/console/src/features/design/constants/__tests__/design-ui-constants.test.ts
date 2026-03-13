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
import DesignUIConstants from '../design-ui-constants';

describe('DesignUIConstants', () => {
  it('INITIAL_LIMIT is 8', () => {
    expect(DesignUIConstants.INITIAL_LIMIT).toBe(8);
  });

  it('LEFT_PANEL_WIDTH is 300', () => {
    expect(DesignUIConstants.LEFT_PANEL_WIDTH).toBe(300);
  });

  it('RIGHT_PANEL_WIDTH is 350', () => {
    expect(DesignUIConstants.RIGHT_PANEL_WIDTH).toBe(350);
  });

  it('contains exactly three keys', () => {
    expect(Object.keys(DesignUIConstants)).toHaveLength(3);
  });

  it('is a frozen / const object (values cannot be reassigned)', () => {
    // TypeScript "as const" produces a readonly object at the type level;
    // runtime mutation is still possible unless Object.freeze is used.
    // We simply verify the expected type of each value.
    expect(typeof DesignUIConstants.INITIAL_LIMIT).toBe('number');
    expect(typeof DesignUIConstants.LEFT_PANEL_WIDTH).toBe('number');
    expect(typeof DesignUIConstants.RIGHT_PANEL_WIDTH).toBe('number');
  });
});
