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
import {AgentCreateFlowStep} from '../agent-create-flow';

describe('AgentCreateFlowStep', () => {
  it('should have NAME step', () => {
    expect(AgentCreateFlowStep.NAME).toBe('NAME');
  });

  it('should have ORGANIZATION_UNIT step', () => {
    expect(AgentCreateFlowStep.ORGANIZATION_UNIT).toBe('ORGANIZATION_UNIT');
  });

  it('should have PROFILE step', () => {
    expect(AgentCreateFlowStep.PROFILE).toBe('PROFILE');
  });

  it('should have OWNER step', () => {
    expect(AgentCreateFlowStep.OWNER).toBe('OWNER');
  });

  it('should have exactly 4 steps', () => {
    expect(Object.keys(AgentCreateFlowStep)).toHaveLength(4);
  });
});
