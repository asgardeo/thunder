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

import {describe, it, expect, vi} from 'vitest';

// Mock lexical's createCommand
const mockCreateCommand = vi.fn((name: string) => ({type: name}));
vi.mock('lexical', () => ({
  createCommand: mockCreateCommand,
}));

describe('commands', () => {
  it('should export TOGGLE_SAFE_LINK_COMMAND created with createCommand', async () => {
    // Import the module which will trigger createCommand
    const {default: TOGGLE_SAFE_LINK_COMMAND} = await import('../commands');

    // Verify createCommand was called with the correct command name
    expect(mockCreateCommand).toHaveBeenCalledWith('TOGGLE_SAFE_LINK_COMMAND');

    // Verify the exported command has the expected structure
    expect(TOGGLE_SAFE_LINK_COMMAND).toBeDefined();
    expect(TOGGLE_SAFE_LINK_COMMAND).toEqual({type: 'TOGGLE_SAFE_LINK_COMMAND'});
  });

  it('should be a LexicalCommand type for string payloads', async () => {
    const {default: TOGGLE_SAFE_LINK_COMMAND} = await import('../commands');

    // The command should be defined and usable
    expect(TOGGLE_SAFE_LINK_COMMAND).toBeDefined();
  });
});
