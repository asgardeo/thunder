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

import {describe, it, expect} from 'vitest';
import {renderHook} from '@thunder/test-utils/browser';
import useFlowNaming from '../useFlowNaming';

describe('useFlowNaming', () => {
  describe('Hook Interface', () => {
    it('should return flowName', async () => {
      const {result} = await renderHook(() => useFlowNaming());
      expect(result.current.flowName).toBeDefined();
      expect(typeof result.current.flowName).toBe('string');
    });

    it('should return flowHandle', async () => {
      const {result} = await renderHook(() => useFlowNaming());
      expect(result.current.flowHandle).toBeDefined();
      expect(typeof result.current.flowHandle).toBe('string');
    });

    it('should return needsAutoLayout', async () => {
      const {result} = await renderHook(() => useFlowNaming());
      expect(typeof result.current.needsAutoLayout).toBe('boolean');
    });

    it('should return setNeedsAutoLayout function', async () => {
      const {result} = await renderHook(() => useFlowNaming());
      expect(typeof result.current.setNeedsAutoLayout).toBe('function');
    });

    it('should return handleFlowNameChange function', async () => {
      const {result} = await renderHook(() => useFlowNaming());
      expect(typeof result.current.handleFlowNameChange).toBe('function');
    });
  });

  describe('Default values', () => {
    it('should use default name when no props provided', async () => {
      const {result} = await renderHook(() => useFlowNaming());
      expect(result.current.flowName).toBe('Login Flow');
    });

    it('should use default handle when no props provided', async () => {
      const {result} = await renderHook(() => useFlowNaming());
      expect(result.current.flowHandle).toBe('login-flow');
    });

    it('should use custom default name when provided', async () => {
      const {result} = await renderHook(() => useFlowNaming({defaultName: 'Custom Flow'}));
      expect(result.current.flowName).toBe('Custom Flow');
    });

    it('should use custom default handle when provided', async () => {
      const {result} = await renderHook(() => useFlowNaming({defaultHandle: 'custom-handle'}));
      expect(result.current.flowHandle).toBe('custom-handle');
    });

    it('should initialize needsAutoLayout to false', async () => {
      const {result} = await renderHook(() => useFlowNaming());
      expect(result.current.needsAutoLayout).toBe(false);
    });
  });

  describe('handleFlowNameChange', () => {
    it('should update flowName when called', async () => {
      const {result} = await renderHook(() => useFlowNaming());

      result.current.handleFlowNameChange('New Flow Name');

      expect(result.current.flowName).toBe('New Flow Name');
    });

    it('should generate handle from name', async () => {
      const {result} = await renderHook(() => useFlowNaming());

      result.current.handleFlowNameChange('My Custom Flow');

      expect(result.current.flowHandle).toBe('my-custom-flow');
    });

    it('should convert name to lowercase for handle', async () => {
      const {result} = await renderHook(() => useFlowNaming());

      result.current.handleFlowNameChange('UPPERCASE NAME');

      expect(result.current.flowHandle).toBe('uppercase-name');
    });

    it('should replace spaces with hyphens in handle', async () => {
      const {result} = await renderHook(() => useFlowNaming());

      result.current.handleFlowNameChange('Multiple   Spaces   Here');

      expect(result.current.flowHandle).toBe('multiple-spaces-here');
    });

    it('should remove special characters from handle', async () => {
      const {result} = await renderHook(() => useFlowNaming());

      result.current.handleFlowNameChange('Flow@Name#With$Special!Characters');

      expect(result.current.flowHandle).toBe('flownamewithspecialcharacters');
    });

    it('should remove leading and trailing hyphens', async () => {
      const {result} = await renderHook(() => useFlowNaming());

      result.current.handleFlowNameChange(' - Trimmed Name - ');

      expect(result.current.flowHandle).toBe('trimmed-name');
    });

    it('should collapse multiple hyphens into one', async () => {
      const {result} = await renderHook(() => useFlowNaming());

      result.current.handleFlowNameChange('Name---With---Multiple---Hyphens');

      expect(result.current.flowHandle).toBe('name-with-multiple-hyphens');
    });
  });

  describe('setNeedsAutoLayout', () => {
    it('should update needsAutoLayout when called', async () => {
      const {result} = await renderHook(() => useFlowNaming());

      expect(result.current.needsAutoLayout).toBe(false);

      result.current.setNeedsAutoLayout(true);

      expect(result.current.needsAutoLayout).toBe(true);
    });

    it('should allow toggling needsAutoLayout', async () => {
      const {result} = await renderHook(() => useFlowNaming());

      result.current.setNeedsAutoLayout(true);
      expect(result.current.needsAutoLayout).toBe(true);

      result.current.setNeedsAutoLayout(false);
      expect(result.current.needsAutoLayout).toBe(false);
    });
  });

  describe('existingFlowData synchronization', () => {
    it('should sync flowName from existingFlowData', async () => {
      const {result} = await renderHook(() =>
        useFlowNaming({
          existingFlowData: {
            name: 'Existing Flow Name',
            handle: 'existing-handle',
          },
        }),
      );

      expect(result.current.flowName).toBe('Existing Flow Name');
    });

    it('should sync flowHandle from existingFlowData', async () => {
      const {result} = await renderHook(() =>
        useFlowNaming({
          existingFlowData: {
            name: 'Existing Flow Name',
            handle: 'existing-handle',
          },
        }),
      );

      expect(result.current.flowHandle).toBe('existing-handle');
    });

    it('should generate handle from name when existingFlowData has name but no handle', async () => {
      const {result} = await renderHook(() =>
        useFlowNaming({
          existingFlowData: {
            name: 'Flow Without Handle',
            // No handle provided
          },
        }),
      );

      // Should generate handle from name
      expect(result.current.flowHandle).toBe('flow-without-handle');
    });

    it('should handle existingFlowData with only handle', async () => {
      const {result} = await renderHook(() =>
        useFlowNaming({
          existingFlowData: {
            handle: 'only-handle',
            // No name provided
          },
        }),
      );

      // Should use the provided handle
      expect(result.current.flowHandle).toBe('only-handle');
      // Name should remain as default
      expect(result.current.flowName).toBe('Login Flow');
    });

    it('should handle empty existingFlowData', async () => {
      const {result} = await renderHook(() =>
        useFlowNaming({
          existingFlowData: {},
        }),
      );

      expect(result.current.flowName).toBe('Login Flow');
      expect(result.current.flowHandle).toBe('login-flow');
    });

    it('should update when existingFlowData changes', async () => {
      const {result, rerender} = await renderHook(
        (props?) => useFlowNaming({existingFlowData: props!.existingFlowData}),
        {
          initialProps: {
            existingFlowData: {
              name: 'Initial Name',
              handle: 'initial-handle',
            },
          },
        },
      );

      expect(result.current.flowName).toBe('Initial Name');
      expect(result.current.flowHandle).toBe('initial-handle');

      await rerender({
        existingFlowData: {
          name: 'Updated Name',
          handle: 'updated-handle',
        },
      });

      expect(result.current.flowName).toBe('Updated Name');
      expect(result.current.flowHandle).toBe('updated-handle');
    });

    it('should generate handle from complex name when no handle provided', async () => {
      const {result} = await renderHook(() =>
        useFlowNaming({
          existingFlowData: {
            name: 'Complex Flow Name With UPPERCASE and Special@Chars!',
            // No handle - should be generated
          },
        }),
      );

      // Handle should be generated from the name
      expect(result.current.flowHandle).toBe('complex-flow-name-with-uppercase-and-specialchars');
    });

    it('should prefer explicit handle over generated one', async () => {
      const {result} = await renderHook(() =>
        useFlowNaming({
          existingFlowData: {
            name: 'Different Name',
            handle: 'explicit-handle',
          },
        }),
      );

      // Should use the explicit handle, not generate from name
      expect(result.current.flowHandle).toBe('explicit-handle');
    });
  });

  describe('generateHandleFromName edge cases', () => {
    it('should handle empty string', async () => {
      const {result} = await renderHook(() => useFlowNaming());

      result.current.handleFlowNameChange('');

      expect(result.current.flowHandle).toBe('');
    });

    it('should handle string with only special characters', async () => {
      const {result} = await renderHook(() => useFlowNaming());

      result.current.handleFlowNameChange('@#$%^&*()');

      expect(result.current.flowHandle).toBe('');
    });

    it('should handle string with only spaces', async () => {
      const {result} = await renderHook(() => useFlowNaming());

      result.current.handleFlowNameChange('     ');

      expect(result.current.flowHandle).toBe('');
    });

    it('should handle numbers in name', async () => {
      const {result} = await renderHook(() => useFlowNaming());

      result.current.handleFlowNameChange('Flow 123 Test');

      expect(result.current.flowHandle).toBe('flow-123-test');
    });

    it('should trim whitespace from name before generating handle', async () => {
      const {result} = await renderHook(() => useFlowNaming());

      result.current.handleFlowNameChange('  Trimmed Flow  ');

      expect(result.current.flowHandle).toBe('trimmed-flow');
    });
  });

  describe('undefined props handling', () => {
    it('should handle undefined props gracefully', async () => {
      const {result} = await renderHook(() => useFlowNaming(undefined));

      expect(result.current.flowName).toBe('Login Flow');
      expect(result.current.flowHandle).toBe('login-flow');
    });
  });
});
