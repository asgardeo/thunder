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
import {renderHook} from '@thunder/test-utils/browser';
import useEdgeStyleSelector from '../useEdgeStyleSelector';

describe('useEdgeStyleSelector', () => {
  it('should initialize with null anchorEl', async () => {
    const {result} = await renderHook(() => useEdgeStyleSelector());

    expect(result.current.anchorEl).toBeNull();
  });

  it('should return handleClick function', async () => {
    const {result} = await renderHook(() => useEdgeStyleSelector());

    expect(typeof result.current.handleClick).toBe('function');
  });

  it('should return handleClose function', async () => {
    const {result} = await renderHook(() => useEdgeStyleSelector());

    expect(typeof result.current.handleClose).toBe('function');
  });

  it('should set anchorEl on handleClick', async () => {
    const {result} = await renderHook(() => useEdgeStyleSelector());

    const mockElement = document.createElement('button');
    const mockEvent = {
      currentTarget: mockElement,
    } as unknown as React.MouseEvent<HTMLElement>;

    result.current.handleClick(mockEvent);

    await vi.waitFor(() => {
      expect(result.current.anchorEl).toBe(mockElement);
    });
  });

  it('should reset anchorEl to null on handleClose', async () => {
    const {result} = await renderHook(() => useEdgeStyleSelector());

    const mockElement = document.createElement('button');
    const mockEvent = {
      currentTarget: mockElement,
    } as unknown as React.MouseEvent<HTMLElement>;

    // First open the menu
    result.current.handleClick(mockEvent);

    await vi.waitFor(() => {
      expect(result.current.anchorEl).toBe(mockElement);
    });

    // Then close it
    result.current.handleClose();

    await vi.waitFor(() => {
      expect(result.current.anchorEl).toBeNull();
    });
  });

  it('should handle multiple open/close cycles', async () => {
    const {result} = await renderHook(() => useEdgeStyleSelector());

    const mockElement1 = document.createElement('button');
    const mockElement2 = document.createElement('div');

    // First cycle
    result.current.handleClick({currentTarget: mockElement1} as unknown as React.MouseEvent<HTMLElement>);
    await vi.waitFor(() => {
      expect(result.current.anchorEl).toBe(mockElement1);
    });

    result.current.handleClose();
    await vi.waitFor(() => {
      expect(result.current.anchorEl).toBeNull();
    });

    // Second cycle with different element
    result.current.handleClick({currentTarget: mockElement2} as unknown as React.MouseEvent<HTMLElement>);
    await vi.waitFor(() => {
      expect(result.current.anchorEl).toBe(mockElement2);
    });

    result.current.handleClose();
    await vi.waitFor(() => {
      expect(result.current.anchorEl).toBeNull();
    });
  });

  it('should preserve function references between renders', async () => {
    const {result, rerender} = await renderHook(() => useEdgeStyleSelector());

    const initialHandleClick = result.current.handleClick;
    const initialHandleClose = result.current.handleClose;

    await rerender();

    // useCallback should maintain stable references
    expect(result.current.handleClick).toBe(initialHandleClick);
    expect(result.current.handleClose).toBe(initialHandleClose);
  });

  it('should return an object with correct shape', async () => {
    const {result} = await renderHook(() => useEdgeStyleSelector());

    expect(result.current).toHaveProperty('anchorEl');
    expect(result.current).toHaveProperty('handleClick');
    expect(result.current).toHaveProperty('handleClose');
    expect(Object.keys(result.current)).toHaveLength(3);
  });
});
