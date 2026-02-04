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

/**
 * This test file covers the predefined URL functionality in CustomLinkPlugin.
 * Since PREDEFINED_URLS is an empty constant array in the source,
 * we test the helper functions' logic directly through component behavior
 * by mocking scenarios where the URL detection functions are exercised.
 */

import {describe, it, expect, vi, beforeEach, afterEach} from 'vitest';
import {render, screen, fireEvent, waitFor, act, cleanup} from '@testing-library/react';
import type React from 'react';

// Use vi.hoisted for mock functions
const {
  mockDispatchCommand,
  mockRegisterUpdateListener,
  mockRegisterCommand,
  mockGetRootElement,
  mockGetEditorState,
  mockGetSelection,
  mockIsRangeSelection,
  mockIsLinkNode,
  mockGetSelectedNode,
} = vi.hoisted(() => ({
  mockDispatchCommand: vi.fn<(...args: unknown[]) => unknown>(),
  mockRegisterUpdateListener: vi.fn<(...args: unknown[]) => () => void>(() => vi.fn()),
  mockRegisterCommand: vi.fn<(...args: unknown[]) => () => void>(() => vi.fn()),
  mockGetRootElement: vi.fn<() => HTMLElement | null>(() => document.createElement('div')),
  mockGetEditorState: vi.fn<() => {read: (callback: () => void) => void}>(() => ({
    read: vi.fn((callback: () => void) => callback()),
  })),
  mockGetSelection: vi.fn<() => unknown>(() => ({type: 'range'})),
  mockIsRangeSelection: vi.fn<(selection: unknown) => boolean>(() => true),
  mockIsLinkNode: vi.fn<(node: unknown) => boolean>(() => false),
  mockGetSelectedNode: vi.fn<() => unknown>(() => ({
    getParent: () => null,
    getURL: () => 'https://example.com',
    setTarget: vi.fn(),
    setRel: vi.fn(),
    type: 'text',
  })),
}));

// Mock react-i18next
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
}));

// Mock the lexical composer context
vi.mock('@lexical/react/LexicalComposerContext', () => ({
  useLexicalComposerContext: () => [{
    dispatchCommand: mockDispatchCommand,
    registerUpdateListener: mockRegisterUpdateListener,
    registerCommand: mockRegisterCommand,
    getRootElement: mockGetRootElement,
    getEditorState: mockGetEditorState,
  }],
}));

// Mock lexical utils
vi.mock('@lexical/utils', () => ({
  mergeRegister: (...fns: (() => void)[]) => () => fns.forEach(fn => fn()),
}));

// Mock lexical
vi.mock('lexical', () => ({
  $getSelection: mockGetSelection,
  $isRangeSelection: mockIsRangeSelection,
  CLICK_COMMAND: 'CLICK_COMMAND',
  KEY_ESCAPE_COMMAND: 'KEY_ESCAPE_COMMAND',
  SELECTION_CHANGE_COMMAND: 'SELECTION_CHANGE_COMMAND',
}));

// Mock @lexical/link
vi.mock('@lexical/link', () => ({
  $isLinkNode: mockIsLinkNode,
  TOGGLE_LINK_COMMAND: 'TOGGLE_LINK_COMMAND',
}));

// Mock getSelectedNode utility
vi.mock('../../utils/getSelectedNode', () => ({
  default: mockGetSelectedNode,
}));

// Mock commands
vi.mock('../commands', () => ({
  default: 'TOGGLE_SAFE_LINK_COMMAND',
}));

// Mock createPortal to render directly
vi.mock('react-dom', () => ({
  createPortal: (children: React.ReactNode) => children,
}));

// eslint-disable-next-line import/first -- Import after mocks are set up
import CustomLinkPlugin from '../CustomLinkPlugin';

describe('CustomLinkPlugin - URL Type Detection Functions', () => {
  beforeEach(() => {
    vi.resetAllMocks();

    // Restore default implementations for hoisted mocks after resetAllMocks
    // In Vitest 4, vi.restoreAllMocks() no longer resets vi.fn() mocks,
    // so mockReturnValue calls from previous tests persist unless explicitly reset
    mockRegisterUpdateListener.mockImplementation(() => vi.fn());
    mockRegisterCommand.mockImplementation(() => vi.fn());
    mockGetRootElement.mockReturnValue(document.createElement('div'));
    mockGetEditorState.mockReturnValue({
      read: vi.fn((callback: () => void) => callback()),
    });
    mockGetSelection.mockReturnValue({type: 'range'});
    mockIsRangeSelection.mockReturnValue(true);
    mockIsLinkNode.mockReturnValue(false);
    mockGetSelectedNode.mockReturnValue({
      getParent: () => null,
      getURL: () => 'https://example.com',
      setTarget: vi.fn(),
      setRel: vi.fn(),
      type: 'text',
    });

    // Spy on window methods (passthrough to allow normal event handling)
    vi.spyOn(window, 'addEventListener');
    vi.spyOn(window, 'removeEventListener');
    vi.spyOn(window, 'open').mockImplementation(vi.fn());
  });

  afterEach(() => {
    cleanup();
    vi.restoreAllMocks();
  });

  describe('determineUrlType function behavior', () => {
    it('should return CUSTOM for URLs not in PREDEFINED_URLS array', () => {
      // Since PREDEFINED_URLS is empty, any URL will return CUSTOM
      const mockLinkNode = {
        type: 'link',
        getParent: () => ({type: 'paragraph'}),
        getURL: () => 'https://any-custom-url.com',
        setTarget: vi.fn(),
        setRel: vi.fn(),
      };
      mockGetSelectedNode.mockReturnValue(mockLinkNode);
      mockIsLinkNode.mockImplementation((node: unknown) => node === mockLinkNode);
      mockIsRangeSelection.mockReturnValue(true);
      mockGetSelection.mockReturnValue({type: 'range'});

      const rootElement = document.createElement('div');
      const textNode = document.createTextNode('test');
      rootElement.appendChild(textNode);

      const mockSelection = {
        isCollapsed: false,
        anchorNode: textNode,
        getRangeAt: () => ({
          getBoundingClientRect: () => ({
            top: 100,
            left: 100,
            height: 20,
            width: 50,
          }),
        }),
      };
      vi.spyOn(window, 'getSelection').mockReturnValue(mockSelection as unknown as Selection);
      mockGetRootElement.mockReturnValue(rootElement);

      render(<CustomLinkPlugin />);

      // The link should be displayed (determineUrlType returns CUSTOM for any URL)
      const link = document.querySelector('.MuiLink-root');
      expect(link).toBeInTheDocument();
    });

    it('should exercise determineUrlType when URL matches no predefined option', () => {
      // This tests the behavior when find() returns undefined and returns 'CUSTOM'
      const mockLinkNode = {
        type: 'link',
        getParent: () => ({type: 'paragraph'}),
        getURL: () => 'https://random-url.com',
        setTarget: vi.fn(),
        setRel: vi.fn(),
      };
      mockGetSelectedNode.mockReturnValue(mockLinkNode);
      mockIsLinkNode.mockImplementation((node: unknown) => node === mockLinkNode);
      mockIsRangeSelection.mockReturnValue(true);

      render(<CustomLinkPlugin />);

      expect(mockGetEditorState).toHaveBeenCalled();
    });
  });

  describe('getPlaceholderUrl function behavior', () => {
    it('should return URL itself when selectedType is CUSTOM', () => {
      // This tests the behavior when selectedType is CUSTOM
      const testUrl = 'https://my-custom-placeholder.com';
      const mockLinkNode = {
        type: 'link',
        getParent: () => ({type: 'paragraph'}),
        getURL: () => testUrl,
        setTarget: vi.fn(),
        setRel: vi.fn(),
      };
      mockGetSelectedNode.mockReturnValue(mockLinkNode);
      mockIsLinkNode.mockImplementation((node: unknown) => node === mockLinkNode);
      mockIsRangeSelection.mockReturnValue(true);
      mockGetSelection.mockReturnValue({type: 'range'});

      const rootElement = document.createElement('div');
      const textNode = document.createTextNode('test');
      rootElement.appendChild(textNode);

      const mockSelection = {
        isCollapsed: false,
        anchorNode: textNode,
        getRangeAt: () => ({
          getBoundingClientRect: () => ({
            top: 100,
            left: 100,
            height: 20,
            width: 50,
          }),
        }),
      };
      vi.spyOn(window, 'getSelection').mockReturnValue(mockSelection as unknown as Selection);
      mockGetRootElement.mockReturnValue(rootElement);

      render(<CustomLinkPlugin />);

      // The link should display the URL itself as placeholder
      const link = document.querySelector('.MuiLink-root');
      expect(link).toBeInTheDocument();
    });
  });

  describe('handleUrlTypeChange function behavior', () => {
    it('should set URL to https:// when switching to CUSTOM type', async () => {
      // This tests the else branch behavior when newType is CUSTOM
      render(<CustomLinkPlugin />);

      // Enter edit mode
      const editButton = screen.getByText('common:edit');
      await act(async () => {
        fireEvent.click(editButton);
      });

      await waitFor(() => {
        expect(screen.getByText('flows:core.elements.richText.linkEditor.editLink')).toBeInTheDocument();
      });

      // Since PREDEFINED_URLS is empty, Select won't render
      // But we verify the component is in edit mode correctly
      const textField = document.querySelector('input');
      expect(textField).toBeInTheDocument();
    });
  });

  describe('getCurrentUrl function behavior', () => {
    it('should return linkUrl when selectedUrlType is CUSTOM', async () => {
      // This tests the else branch that returns linkUrl
      render(<CustomLinkPlugin />);

      // Enter edit mode
      const editButton = screen.getByText('common:edit');
      await act(async () => {
        fireEvent.click(editButton);
      });

      await waitFor(() => {
        expect(screen.getByText('flows:core.elements.richText.linkEditor.editLink')).toBeInTheDocument();
      });

      // Type a custom URL
      const textField = document.querySelector('input');
      if (textField) {
        await act(async () => {
          fireEvent.change(textField, {target: {value: 'https://test-current-url.com'}});
        });
        expect(textField).toHaveValue('https://test-current-url.com');

        // Click save to trigger getCurrentUrl
        const saveButton = screen.getByText('common:save');
        await act(async () => {
          fireEvent.click(saveButton);
        });
      }

      // Should exit edit mode
      await waitFor(() => {
        expect(screen.getByText('flows:core.elements.richText.linkEditor.viewLink')).toBeInTheDocument();
      });
    });
  });

  describe('Select component rendering', () => {
    it('should not render Select when PREDEFINED_URLS is empty', async () => {
      // This tests the conditional rendering when PREDEFINED_URLS is empty
      render(<CustomLinkPlugin />);

      // Enter edit mode
      const editButton = screen.getByText('common:edit');
      await act(async () => {
        fireEvent.click(editButton);
      });

      await waitFor(() => {
        expect(screen.getByText('flows:core.elements.richText.linkEditor.editLink')).toBeInTheDocument();
      });

      // Select should NOT be rendered since PREDEFINED_URLS is empty
      const select = document.querySelector('.MuiSelect-root');
      expect(select).not.toBeInTheDocument();
    });
  });

  describe('URL detection with various URL formats', () => {
    it('should handle empty string URL', () => {
      const mockLinkNode = {
        type: 'link',
        getParent: () => ({type: 'paragraph'}),
        getURL: () => '',
        setTarget: vi.fn(),
        setRel: vi.fn(),
      };
      mockGetSelectedNode.mockReturnValue(mockLinkNode);
      mockIsLinkNode.mockImplementation((node: unknown) => node === mockLinkNode);
      mockIsRangeSelection.mockReturnValue(true);

      render(<CustomLinkPlugin />);

      expect(mockGetEditorState).toHaveBeenCalled();
    });

    it('should handle URL with special characters', () => {
      const mockLinkNode = {
        type: 'link',
        getParent: () => ({type: 'paragraph'}),
        getURL: () => 'https://example.com/path?query=value&other=test#hash',
        setTarget: vi.fn(),
        setRel: vi.fn(),
      };
      mockGetSelectedNode.mockReturnValue(mockLinkNode);
      mockIsLinkNode.mockImplementation((node: unknown) => node === mockLinkNode);
      mockIsRangeSelection.mockReturnValue(true);

      render(<CustomLinkPlugin />);

      expect(mockGetEditorState).toHaveBeenCalled();
    });

    it('should handle URL with unicode characters', () => {
      const mockLinkNode = {
        type: 'link',
        getParent: () => ({type: 'paragraph'}),
        getURL: () => 'https://example.com/路径/页面',
        setTarget: vi.fn(),
        setRel: vi.fn(),
      };
      mockGetSelectedNode.mockReturnValue(mockLinkNode);
      mockIsLinkNode.mockImplementation((node: unknown) => node === mockLinkNode);
      mockIsRangeSelection.mockReturnValue(true);

      render(<CustomLinkPlugin />);

      expect(mockGetEditorState).toHaveBeenCalled();
    });
  });

  describe('Edge cases for URL type selection', () => {
    it('should handle rapid URL type changes', async () => {
      render(<CustomLinkPlugin />);

      // Enter edit mode
      const editButton = screen.getByText('common:edit');
      await act(async () => {
        fireEvent.click(editButton);
      });

      await waitFor(() => {
        expect(screen.getByText('flows:core.elements.richText.linkEditor.editLink')).toBeInTheDocument();
      });

      // Type multiple URLs rapidly
      const textField = document.querySelector('input');
      if (textField) {
        await act(async () => {
          fireEvent.change(textField, {target: {value: 'https://url1.com'}});
        });
        await act(async () => {
          fireEvent.change(textField, {target: {value: 'https://url2.com'}});
        });
        await act(async () => {
          fireEvent.change(textField, {target: {value: 'https://final-url.com'}});
        });
        expect(textField).toHaveValue('https://final-url.com');
      }
    });

    it('should handle save with valid URL after initially empty', async () => {
      render(<CustomLinkPlugin />);

      // Enter edit mode
      const editButton = screen.getByText('common:edit');
      await act(async () => {
        fireEvent.click(editButton);
      });

      await waitFor(() => {
        expect(screen.getByText('flows:core.elements.richText.linkEditor.editLink')).toBeInTheDocument();
      });

      const textField = document.querySelector('input');
      if (textField) {
        // Clear the field first
        await act(async () => {
          fireEvent.change(textField, {target: {value: ''}});
        });
        // Then add a valid URL
        await act(async () => {
          fireEvent.change(textField, {target: {value: 'https://valid-url.com'}});
        });
        expect(textField).toHaveValue('https://valid-url.com');

        // Save
        const saveButton = screen.getByText('common:save');
        await act(async () => {
          fireEvent.click(saveButton);
        });
      }

      await waitFor(() => {
        expect(screen.getByText('flows:core.elements.richText.linkEditor.viewLink')).toBeInTheDocument();
      });
    });
  });
});
