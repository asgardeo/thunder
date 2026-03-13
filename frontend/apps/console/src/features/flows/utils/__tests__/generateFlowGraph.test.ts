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
import {FlowNodeType} from '../../models/flows';
import generateFlowGraph from '../generateFlowGraph';

describe('generateFlowGraph', () => {
  it('should generate a Basic Auth flow', () => {
    const request = generateFlowGraph({
      hasBasicAuth: true,
      hasPasskey: false,
      hasSmsOtp: false,
    });

    expect(request.handle).toBe('generated-basic-flow');
    expect(request.nodes).toHaveLength(5); // START, PROMPT, BASIC_EXEC, AUTH_ASSERT, END
    // START, PROMPT, BASIC_EXEC, AUTH_ASSERT, END = 5 nodes

    const promptNode = request.nodes.find((n) => n.type === FlowNodeType.PROMPT);
    expect(promptNode).toBeDefined();
    expect(promptNode?.meta?.components).toBeDefined();

    // Check for username/password inputs
    const components = promptNode?.meta?.components as {id: string}[];
    const basicBlock = components.find((c) => c.id === 'block_basic');
    expect(basicBlock).toBeDefined();
  });

  it('should generate a Passkey flow', () => {
    const request = generateFlowGraph({
      hasBasicAuth: false,
      hasPasskey: true,
      hasSmsOtp: false,
    });

    expect(request.handle).toBe('generated-passkey-flow');

    const promptNode = request.nodes.find((n) => n.type === FlowNodeType.PROMPT);
    const components = promptNode?.meta?.components as {id: string}[];

    // Should have passkey block
    const passkeyBlock = components.find((c) => c.id === 'block_passkey');
    expect(passkeyBlock).toBeDefined();

    // Should NOT have basic block
    const basicBlock = components.find((c) => c.id === 'block_basic');
    expect(basicBlock).toBeUndefined();

    // Executors
    const executors = request.nodes.filter((n) => n.type === FlowNodeType.TASK_EXECUTION);
    const passkeyExecutors = executors.filter((n) => n.executor?.name === 'PasskeyAuthExecutor');
    expect(passkeyExecutors).toHaveLength(2); // Challenge and Verify
  });

  it('should generate a Combined flow (Basic + Passkey + Google)', () => {
    const request = generateFlowGraph({
      hasBasicAuth: true,
      hasPasskey: true,
      googleIdpId: 'google-p-id',
      hasSmsOtp: false,
    });

    expect(request.handle).toBe('generated-basic-google-passkey-flow');

    const promptNode = request.nodes.find((n) => n.type === FlowNodeType.PROMPT);
    const components = promptNode?.meta?.components as {id: string}[];

    expect(components.find((c) => c.id === 'block_basic')).toBeDefined();
    expect(components.find((c) => c.id === 'block_passkey')).toBeDefined();
    expect(components.find((c) => c.id === 'block_social')).toBeDefined();

    // Executors
    const executors = request.nodes.filter((n) => n.type === FlowNodeType.TASK_EXECUTION);
    expect(executors.find((n) => n.executor?.name === 'BasicAuthExecutor')).toBeDefined();
    expect(executors.find((n) => n.executor?.name === 'PasskeyAuthExecutor')).toBeDefined();
    expect(executors.find((n) => n.executor?.name === 'GoogleOIDCAuthExecutor')).toBeDefined();
    expect(executors.find((n) => n.executor?.name === 'ProvisioningExecutor')).toBeDefined();
  });

  it('should generate a Combined flow (Basic + Github)', () => {
    const request = generateFlowGraph({
      hasBasicAuth: true,
      hasPasskey: false,
      githubIdpId: 'github-id',
      hasSmsOtp: false,
    });

    expect(request.handle).toBe('generated-basic-github-flow');

    // Executors
    const executors = request.nodes.filter((n) => n.type === FlowNodeType.TASK_EXECUTION);
    expect(executors.find((n) => n.executor?.name === 'BasicAuthExecutor')).toBeDefined();
    expect(executors.find((n) => n.executor?.name === 'GithubOAuthExecutor')).toBeDefined();
    expect(executors.find((n) => n.executor?.name === 'ProvisioningExecutor')).toBeDefined();
  });

  it('should use provided relying party options for Passkey flow', () => {
    const request = generateFlowGraph({
      hasBasicAuth: false,
      hasPasskey: true,
      hasSmsOtp: false,
      relyingPartyId: 'my-app.com',
      relyingPartyName: 'My App',
    });

    const challengeNode = request.nodes.find((n) => n.id === 'passkey_challenge');
    expect(challengeNode).toBeDefined();
    expect(challengeNode?.properties?.relyingPartyId).toBe('my-app.com');
    expect(challengeNode?.properties?.relyingPartyName).toBe('My App');
  });
});
