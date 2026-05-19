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
import type {AgentTypeListParams, UpdateAgentTypeRequest} from '../requests';

describe('agent-types request types', () => {
  it('accepts UpdateAgentTypeRequest shape', () => {
    const req: UpdateAgentTypeRequest = {
      name: 'default',
      ouId: 'ou1',
      schema: {foo: {type: 'string'}},
    };
    expect(req.name).toBe('default');
  });

  it('accepts UpdateAgentTypeRequest with optional systemAttributes', () => {
    const req: UpdateAgentTypeRequest = {
      name: 'default',
      ouId: 'ou1',
      systemAttributes: {display: 'name'},
      schema: {},
    };
    expect(req.systemAttributes?.display).toBe('name');
  });

  it('accepts AgentTypeListParams pagination shape', () => {
    const params: AgentTypeListParams = {limit: 10, offset: 0};
    expect(params.limit).toBe(10);
    expect(params.offset).toBe(0);
  });

  it('accepts empty AgentTypeListParams', () => {
    const params: AgentTypeListParams = {};
    expect(params).toEqual({});
  });
});
