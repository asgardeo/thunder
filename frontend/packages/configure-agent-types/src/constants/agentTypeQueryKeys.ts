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

/**
 * Query key constants for agent types feature cache management.
 */
const AgentTypeQueryKeys = {
  /**
   * Base key for all agent type list queries
   */
  AGENT_TYPES: 'agent-types',
  /**
   * Key for a single agent type query
   */
  AGENT_TYPE: 'agent-type',
} as const;

export default AgentTypeQueryKeys;
