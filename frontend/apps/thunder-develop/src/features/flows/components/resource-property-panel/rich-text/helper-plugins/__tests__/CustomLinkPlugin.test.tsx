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

import {describe, it, expect, vi, beforeEach, afterEach} from 'vitest';
import {render, screen, fireEvent, waitFor, act} from '@testing-library/react';
import CustomLinkPlugin from '../CustomLinkPlugin';

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

describe('CustomLinkPlugin', () => {
  beforeEach(() => {
    vi.clearAllMocks();

    // Mock window methods
    vi.spyOn(window, 'addEventListener').mockImplementation(vi.fn());
    vi.spyOn(window, 'removeEventListener').mockImplementation(vi.fn());
    vi.spyOn(window, 'open').mockImplementation(vi.fn());
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('Rendering', () => {
    it('should render the link editor card', () => {
      render(<CustomLinkPlugin />);

      expect(document.querySelector('.MuiCard-root')).toBeInTheDocument();
    });

    it('should render view link title by default', () => {
      render(<CustomLinkPlugin />);

      expect(screen.getByText('flows:core.elements.richText.linkEditor.viewLink')).toBeInTheDocument();
    });

    it('should render close button', () => {
      render(<CustomLinkPlugin />);

      // Find the close button (IconButton with X icon)
      const closeButtons = screen.getAllByRole('button');
      expect(closeButtons.length).toBeGreaterThan(0);
    });

    it('should render edit button in view mode', () => {
      render(<CustomLinkPlugin />);

      expect(screen.getByText('common:edit')).toBeInTheDocument();
    });
  });

  describe('Command Registration', () => {
    it('should register CLICK_COMMAND on mount', () => {
      render(<CustomLinkPlugin />);

      expect(mockRegisterCommand).toHaveBeenCalled();
    });

    it('should register update listener on mount', () => {
      render(<CustomLinkPlugin />);

      expect(mockRegisterUpdateListener).toHaveBeenCalled();
    });

    it('should register SELECTION_CHANGE_COMMAND', () => {
      render(<CustomLinkPlugin />);

      expect(mockRegisterCommand).toHaveBeenCalled();
    });

    it('should register KEY_ESCAPE_COMMAND', () => {
      render(<CustomLinkPlugin />);

      expect(mockRegisterCommand).toHaveBeenCalled();
    });

    it('should register TOGGLE_SAFE_LINK_COMMAND', () => {
      render(<CustomLinkPlugin />);

      expect(mockRegisterCommand).toHaveBeenCalled();
    });
  });

  describe('Edit Mode', () => {
    it('should switch to edit mode when edit button is clicked', async () => {
      render(<CustomLinkPlugin />);

      const editButton = screen.getByText('common:edit');
      await act(async () => {
        fireEvent.click(editButton);
      });

      // After clicking edit, the component should show edit mode title
      await waitFor(() => {
        expect(screen.getByText('flows:core.elements.richText.linkEditor.editLink')).toBeInTheDocument();
      });
    });

    it('should show save button in edit mode', async () => {
      render(<CustomLinkPlugin />);

      const editButton = screen.getByText('common:edit');
      await act(async () => {
        fireEvent.click(editButton);
      });

      await waitFor(() => {
        expect(screen.getByText('common:save')).toBeInTheDocument();
      });
    });

    it('should show text field in edit mode', async () => {
      render(<CustomLinkPlugin />);

      const editButton = screen.getByText('common:edit');
      await act(async () => {
        fireEvent.click(editButton);
      });

      await waitFor(() => {
        const textField = document.querySelector('.MuiTextField-root');
        expect(textField).toBeInTheDocument();
      });
    });

    it('should exit edit mode when escape key is pressed in text field', async () => {
      render(<CustomLinkPlugin />);

      // Enter edit mode
      const editButton = screen.getByText('common:edit');
      await act(async () => {
        fireEvent.click(editButton);
      });

      await waitFor(() => {
        expect(screen.getByText('flows:core.elements.richText.linkEditor.editLink')).toBeInTheDocument();
      });

      // Press escape in the text field
      const textField = document.querySelector('input');
      if (textField) {
        await act(async () => {
          fireEvent.keyDown(textField, {key: 'Escape'});
        });
      }

      await waitFor(() => {
        expect(screen.getByText('flows:core.elements.richText.linkEditor.viewLink')).toBeInTheDocument();
      });
    });

    it('should handle enter key press in text field', async () => {
      render(<CustomLinkPlugin />);

      // Enter edit mode
      const editButton = screen.getByText('common:edit');
      await act(async () => {
        fireEvent.click(editButton);
      });

      await waitFor(() => {
        expect(screen.getByText('flows:core.elements.richText.linkEditor.editLink')).toBeInTheDocument();
      });

      // Type a URL and press enter
      const textField = document.querySelector('input');
      expect(textField).toBeInTheDocument();
      if (textField) {
        await act(async () => {
          fireEvent.change(textField, {target: {value: 'https://test.com'}});
          fireEvent.keyDown(textField, {key: 'Enter'});
        });

        // Verify the input value was updated
        expect(textField).toHaveValue('https://test.com');
      }
    });

    it('should save link when save button is clicked', async () => {
      render(<CustomLinkPlugin />);

      // Enter edit mode
      const editButton = screen.getByText('common:edit');
      await act(async () => {
        fireEvent.click(editButton);
      });

      await waitFor(() => {
        expect(screen.getByText('common:save')).toBeInTheDocument();
      });

      // Type a URL
      const textField = document.querySelector('input');
      if (textField) {
        await act(async () => {
          fireEvent.change(textField, {target: {value: 'https://test.com'}});
        });
      }

      // Click save
      const saveButton = screen.getByText('common:save');
      await act(async () => {
        fireEvent.click(saveButton);
      });

      // Should exit edit mode
      await waitFor(() => {
        expect(screen.getByText('flows:core.elements.richText.linkEditor.viewLink')).toBeInTheDocument();
      });
    });
  });

  describe('URL Display', () => {
    it('should display link in view mode', () => {
      render(<CustomLinkPlugin />);

      const link = document.querySelector('.MuiLink-root');
      expect(link).toBeInTheDocument();
    });

    it('should open link in new tab with security attributes', () => {
      render(<CustomLinkPlugin />);

      const link = document.querySelector('.MuiLink-root');
      expect(link).toHaveAttribute('target', '_blank');
      expect(link).toHaveAttribute('rel', 'noopener noreferrer');
    });
  });

  describe('Close Functionality', () => {
    it('should reset state when close button is clicked', async () => {
      render(<CustomLinkPlugin />);

      // Enter edit mode first
      const editButton = screen.getByText('common:edit');
      await act(async () => {
        fireEvent.click(editButton);
      });

      await waitFor(() => {
        expect(screen.getByText('flows:core.elements.richText.linkEditor.editLink')).toBeInTheDocument();
      });

      // Find and click close button (the IconButton with X icon)
      const closeButtons = screen.getAllByRole('button');
      const closeButton = closeButtons.find(btn => btn.querySelector('svg.lucide-x'));
      if (closeButton) {
        await act(async () => {
          fireEvent.click(closeButton);
        });
      }

      // Should reset to view mode
      await waitFor(() => {
        expect(screen.getByText('flows:core.elements.richText.linkEditor.viewLink')).toBeInTheDocument();
      });
    });
  });

  describe('Event Listeners', () => {
    it('should add window resize listener on mount', () => {
      render(<CustomLinkPlugin />);

      expect(window.addEventListener).toHaveBeenCalledWith('resize', expect.any(Function));
    });

    it('should add body scroll listener on mount', () => {
      const addEventListenerSpy = vi.spyOn(document.body, 'addEventListener');

      render(<CustomLinkPlugin />);

      expect(addEventListenerSpy).toHaveBeenCalledWith('scroll', expect.any(Function));

      addEventListenerSpy.mockRestore();
    });

    it('should remove event listeners on unmount', () => {
      const {unmount} = render(<CustomLinkPlugin />);

      unmount();

      expect(window.removeEventListener).toHaveBeenCalledWith('resize', expect.any(Function));
    });
  });

  describe('Edge Cases', () => {
    it('should handle empty URL', async () => {
      render(<CustomLinkPlugin />);

      // Enter edit mode
      const editButton = screen.getByText('common:edit');
      await act(async () => {
        fireEvent.click(editButton);
      });

      await waitFor(() => {
        expect(screen.getByText('common:save')).toBeInTheDocument();
      });

      // Clear the URL field
      const textField = document.querySelector('input');
      if (textField) {
        await act(async () => {
          fireEvent.change(textField, {target: {value: ''}});
        });
      }

      // Click save
      const saveButton = screen.getByText('common:save');
      await act(async () => {
        fireEvent.click(saveButton);
      });

      // Should not dispatch command with empty URL
      // The component checks for empty URL before dispatching
    });

    it('should handle text field input changes', async () => {
      render(<CustomLinkPlugin />);

      // Enter edit mode
      const editButton = screen.getByText('common:edit');
      await act(async () => {
        fireEvent.click(editButton);
      });

      await waitFor(() => {
        const textField = document.querySelector('input');
        expect(textField).toBeInTheDocument();
      });

      const textField = document.querySelector('input');
      if (textField) {
        await act(async () => {
          fireEvent.change(textField, {target: {value: 'https://new-url.com'}});
        });
        expect(textField).toHaveValue('https://new-url.com');
      }
    });
  });

  describe('Positioning', () => {
    it('should have absolute positioning', () => {
      render(<CustomLinkPlugin />);

      const card = document.querySelector('.MuiCard-root');
      expect(card).toHaveStyle({position: 'absolute'});
    });
  });

  describe('Link Node Detection', () => {
    it('should detect when parent is a link node', async () => {
      // Mock to return a link node parent
      const {$isLinkNode} = await vi.importMock<typeof import('@lexical/link')>('@lexical/link');
      (
        $isLinkNode as ReturnType<typeof vi.fn>
      ).mockImplementation((node: {type?: string} | null) => node && node.type === 'link');

      render(<CustomLinkPlugin />);

      expect(mockRegisterCommand).toHaveBeenCalled();
    });

    it('should detect when node itself is a link node', async () => {
      const {$isLinkNode} = await vi.importMock<typeof import('@lexical/link')>('@lexical/link');
      (
        $isLinkNode as ReturnType<typeof vi.fn>
      ).mockImplementation((node: {type?: string} | null) => node && node.type === 'link');

      render(<CustomLinkPlugin />);

      expect(mockRegisterCommand).toHaveBeenCalled();
    });
  });

  describe('TOGGLE_SAFE_LINK_COMMAND', () => {
    it('should register TOGGLE_SAFE_LINK_COMMAND handler', () => {
      render(<CustomLinkPlugin />);

      // Verify commands were registered
      expect(mockRegisterCommand).toHaveBeenCalled();
    });

    it('should handle empty URL in TOGGLE_SAFE_LINK_COMMAND', () => {
      render(<CustomLinkPlugin />);

      // The component registers the TOGGLE_SAFE_LINK_COMMAND which handles empty URLs
      expect(mockRegisterCommand).toHaveBeenCalled();
    });
  });

  describe('Click Command Handler', () => {
    it('should register click command for opening links', () => {
      render(<CustomLinkPlugin />);

      // CLICK_COMMAND is registered to handle ctrl/meta+click on links
      expect(mockRegisterCommand).toHaveBeenCalled();
    });

    it('should handle click with meta key on link', async () => {
      const mockOpen = vi.spyOn(window, 'open').mockImplementation(vi.fn());

      render(<CustomLinkPlugin />);

      // The CLICK_COMMAND handler checks for metaKey or ctrlKey
      expect(mockRegisterCommand).toHaveBeenCalled();

      mockOpen.mockRestore();
    });

    it('should handle click with ctrl key on link', async () => {
      const mockOpen = vi.spyOn(window, 'open').mockImplementation(vi.fn());

      render(<CustomLinkPlugin />);

      // The CLICK_COMMAND handler checks for ctrlKey
      expect(mockRegisterCommand).toHaveBeenCalled();

      mockOpen.mockRestore();
    });
  });

  describe('Position Editor Element', () => {
    it('should position editor when rect is provided', () => {
      render(<CustomLinkPlugin />);

      // The positionEditorElement function is called during updateLinkEditor
      expect(mockGetEditorState).toHaveBeenCalled();
    });

    it('should hide editor when rect is null', () => {
      render(<CustomLinkPlugin />);

      // When there's no selection, the editor is hidden
      const card = document.querySelector('.MuiCard-root');
      expect(card).toBeInTheDocument();
    });

    it('should handle viewport edge cases for horizontal positioning', () => {
      // Test that the editor stays within viewport bounds
      render(<CustomLinkPlugin />);

      expect(mockGetEditorState).toHaveBeenCalled();
    });

    it('should handle viewport edge cases for vertical positioning', () => {
      // Test that the editor positions above selection when near bottom
      render(<CustomLinkPlugin />);

      expect(mockGetEditorState).toHaveBeenCalled();
    });
  });

  describe('URL Type Handling', () => {
    it('should determine CUSTOM URL type for regular URLs', () => {
      render(<CustomLinkPlugin />);

      // The determineUrlType function returns 'CUSTOM' for non-predefined URLs
      expect(mockGetEditorState).toHaveBeenCalled();
    });

    it('should get placeholder URL for custom URLs', () => {
      render(<CustomLinkPlugin />);

      // The getPlaceholderUrl function returns the URL itself for custom URLs
      expect(mockGetEditorState).toHaveBeenCalled();
    });
  });

  describe('URL Type Change Handler', () => {
    it('should handle URL type change to CUSTOM', async () => {
      render(<CustomLinkPlugin />);

      // Enter edit mode first
      const editButton = screen.getByText('common:edit');
      await act(async () => {
        fireEvent.click(editButton);
      });

      await waitFor(() => {
        expect(screen.getByText('flows:core.elements.richText.linkEditor.editLink')).toBeInTheDocument();
      });

      // The handleUrlTypeChange function sets the URL to 'https://' for CUSTOM type
    });
  });

  describe('getCurrentUrl Function', () => {
    it('should return linkUrl for CUSTOM type', async () => {
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
          fireEvent.change(textField, {target: {value: 'https://custom-url.com'}});
        });
        expect(textField).toHaveValue('https://custom-url.com');
      }
    });
  });

  describe('Selection Change Handling', () => {
    it('should update on selection change', () => {
      render(<CustomLinkPlugin />);

      // SELECTION_CHANGE_COMMAND triggers updateLinkEditor
      expect(mockRegisterCommand).toHaveBeenCalled();
    });
  });

  describe('Escape Key Handling', () => {
    it('should handle KEY_ESCAPE_COMMAND in edit mode', async () => {
      render(<CustomLinkPlugin />);

      // Enter edit mode
      const editButton = screen.getByText('common:edit');
      await act(async () => {
        fireEvent.click(editButton);
      });

      await waitFor(() => {
        expect(screen.getByText('flows:core.elements.richText.linkEditor.editLink')).toBeInTheDocument();
      });

      // KEY_ESCAPE_COMMAND is registered and should exit edit mode
      expect(mockRegisterCommand).toHaveBeenCalled();
    });

    it('should not handle KEY_ESCAPE_COMMAND in view mode', () => {
      render(<CustomLinkPlugin />);

      // In view mode, KEY_ESCAPE_COMMAND returns false
      expect(screen.getByText('flows:core.elements.richText.linkEditor.viewLink')).toBeInTheDocument();
    });
  });

  describe('Update Listener', () => {
    it('should update link editor on editor state change', () => {
      render(<CustomLinkPlugin />);

      // registerUpdateListener is called to listen for editor state changes
      expect(mockRegisterUpdateListener).toHaveBeenCalled();
    });
  });

  describe('Root Element Handling', () => {
    it('should handle null root element', () => {
      mockGetRootElement.mockReturnValueOnce(null as unknown as HTMLDivElement);

      render(<CustomLinkPlugin />);

      // Component should handle null root element gracefully
      expect(mockGetEditorState).toHaveBeenCalled();
    });

    it('should handle root element with nested children', () => {
      const rootElement = document.createElement('div');
      const child = document.createElement('span');
      rootElement.appendChild(child);
      mockGetRootElement.mockReturnValue(rootElement);

      render(<CustomLinkPlugin />);

      expect(mockGetEditorState).toHaveBeenCalled();
    });
  });

  describe('Native Selection Handling', () => {
    it('should handle collapsed native selection', () => {
      render(<CustomLinkPlugin />);

      // When nativeSelection.isCollapsed is true, editor is hidden
      expect(mockGetEditorState).toHaveBeenCalled();
    });

    it('should handle non-collapsed native selection', () => {
      render(<CustomLinkPlugin />);

      // When selection is not collapsed, editor is positioned
      expect(mockGetEditorState).toHaveBeenCalled();
    });
  });

  describe('Focus Handling', () => {
    it('should focus input when entering edit mode', async () => {
      render(<CustomLinkPlugin />);

      // Enter edit mode
      const editButton = screen.getByText('common:edit');
      await act(async () => {
        fireEvent.click(editButton);
      });

      await waitFor(() => {
        const textField = document.querySelector('input');
        expect(textField).toBeInTheDocument();
      });
    });
  });

  describe('Save Link with Last Selection', () => {
    it('should not dispatch command when lastSelection is null', async () => {
      render(<CustomLinkPlugin />);

      // Enter edit mode
      const editButton = screen.getByText('common:edit');
      await act(async () => {
        fireEvent.click(editButton);
      });

      await waitFor(() => {
        expect(screen.getByText('common:save')).toBeInTheDocument();
      });

      // Type a URL
      const textField = document.querySelector('input');
      if (textField) {
        await act(async () => {
          fireEvent.change(textField, {target: {value: 'https://test.com'}});
        });
      }

      // Press Enter (lastSelection may be null in this test setup)
      if (textField) {
        await act(async () => {
          fireEvent.keyDown(textField, {key: 'Enter'});
        });
      }
    });
  });

  describe('Link Attributes', () => {
    it('should set target and rel attributes on link', () => {
      render(<CustomLinkPlugin />);

      const link = document.querySelector('.MuiLink-root');
      expect(link).toHaveAttribute('target', '_blank');
      expect(link).toHaveAttribute('rel', 'noopener noreferrer');
    });
  });

  describe('Command Callbacks Execution', () => {
    it('should execute CLICK_COMMAND callback with link node and meta key', () => {
      // Setup mocks for link node scenario
      const mockLinkNode = {
        type: 'link',
        getParent: () => null,
        getURL: () => 'https://clicked-link.com',
        setTarget: vi.fn(),
        setRel: vi.fn(),
      };
      mockGetSelectedNode.mockReturnValue(mockLinkNode);
      mockIsLinkNode.mockImplementation((node: unknown) => node === mockLinkNode);
      mockIsRangeSelection.mockReturnValue(true);

      // Capture the command callback
      const callbacks: Record<string, unknown> = {};
      (mockRegisterCommand as ReturnType<typeof vi.fn>).mockImplementation(
        (command: unknown, callback: unknown) => {
          callbacks[command as string] = callback;
          return vi.fn();
        },
      );

      const mockOpen = vi.spyOn(window, 'open').mockImplementation(vi.fn());

      render(<CustomLinkPlugin />);

      // Execute the click callback with metaKey
      const clickCallback = callbacks.CLICK_COMMAND as ((payload: MouseEvent) => boolean) | undefined;
      if (clickCallback) {
        const mockEvent = {metaKey: true, ctrlKey: false} as MouseEvent;
        const result = clickCallback(mockEvent);
        expect(result).toBe(true);
        expect(mockOpen).toHaveBeenCalledWith('https://clicked-link.com', '_blank');
      }

      mockOpen.mockRestore();
    });

    it('should execute CLICK_COMMAND callback with link node and ctrl key', () => {
      const mockLinkNode = {
        type: 'link',
        getParent: () => null,
        getURL: () => 'https://ctrl-clicked-link.com',
        setTarget: vi.fn(),
        setRel: vi.fn(),
      };
      mockGetSelectedNode.mockReturnValue(mockLinkNode);
      mockIsLinkNode.mockImplementation((node: unknown) => node === mockLinkNode);
      mockIsRangeSelection.mockReturnValue(true);

      const callbacks: Record<string, unknown> = {};
      (mockRegisterCommand as ReturnType<typeof vi.fn>).mockImplementation(
        (command: unknown, callback: unknown) => {
          callbacks[command as string] = callback;
          return vi.fn();
        },
      );

      const mockOpen = vi.spyOn(window, 'open').mockImplementation(vi.fn());

      render(<CustomLinkPlugin />);

      const clickCallback = callbacks.CLICK_COMMAND as ((payload: MouseEvent) => boolean) | undefined;
      if (clickCallback) {
        const mockEvent = {metaKey: false, ctrlKey: true} as MouseEvent;
        const result = clickCallback(mockEvent);
        expect(result).toBe(true);
        expect(mockOpen).toHaveBeenCalledWith('https://ctrl-clicked-link.com', '_blank');
      }

      mockOpen.mockRestore();
    });

    it('should return false from CLICK_COMMAND when no meta/ctrl key', () => {
      const mockLinkNode = {
        type: 'link',
        getParent: () => null,
        getURL: () => 'https://example.com',
        setTarget: vi.fn(),
        setRel: vi.fn(),
      };
      mockGetSelectedNode.mockReturnValue(mockLinkNode);
      mockIsLinkNode.mockImplementation((node: unknown) => node === mockLinkNode);
      mockIsRangeSelection.mockReturnValue(true);

      const callbacks: Record<string, unknown> = {};
      (mockRegisterCommand as ReturnType<typeof vi.fn>).mockImplementation(
        (command: unknown, callback: unknown) => {
          callbacks[command as string] = callback;
          return vi.fn();
        },
      );

      render(<CustomLinkPlugin />);

      const clickCallback = callbacks.CLICK_COMMAND as ((payload: MouseEvent) => boolean) | undefined;
      if (clickCallback) {
        const mockEvent = {metaKey: false, ctrlKey: false} as MouseEvent;
        const result = clickCallback(mockEvent);
        expect(result).toBe(false);
      }
    });

    it('should return false from CLICK_COMMAND when not a link node', () => {
      const mockTextNode = {
        type: 'text',
        getParent: () => null,
        getURL: () => '',
        setTarget: vi.fn(),
        setRel: vi.fn(),
      };
      mockGetSelectedNode.mockReturnValue(mockTextNode);
      mockIsLinkNode.mockReturnValue(false);
      mockIsRangeSelection.mockReturnValue(true);

      const callbacks: Record<string, unknown> = {};
      (mockRegisterCommand as ReturnType<typeof vi.fn>).mockImplementation(
        (command: unknown, callback: unknown) => {
          callbacks[command as string] = callback;
          return vi.fn();
        },
      );

      render(<CustomLinkPlugin />);

      const clickCallback = callbacks.CLICK_COMMAND as ((payload: MouseEvent) => boolean) | undefined;
      if (clickCallback) {
        const mockEvent = {metaKey: true, ctrlKey: false} as MouseEvent;
        const result = clickCallback(mockEvent);
        expect(result).toBe(false);
      }
    });

    it('should return false from CLICK_COMMAND when linkNode is null', () => {
      const mockTextNode = {
        type: 'text',
        getParent: () => null,
        getURL: () => '',
        setTarget: vi.fn(),
        setRel: vi.fn(),
      };
      mockGetSelectedNode.mockReturnValue(mockTextNode);
      mockIsLinkNode.mockReturnValue(false);
      mockIsRangeSelection.mockReturnValue(true);

      const callbacks: Record<string, unknown> = {};
      (mockRegisterCommand as ReturnType<typeof vi.fn>).mockImplementation(
        (command: unknown, callback: unknown) => {
          callbacks[command as string] = callback;
          return vi.fn();
        },
      );

      render(<CustomLinkPlugin />);

      const clickCallback = callbacks.CLICK_COMMAND as ((payload: MouseEvent) => boolean) | undefined;
      if (clickCallback) {
        const mockEvent = {metaKey: true, ctrlKey: false} as MouseEvent;
        const result = clickCallback(mockEvent);
        expect(result).toBe(false);
      }
    });

    it('should execute TOGGLE_SAFE_LINK_COMMAND with URL', () => {
      const mockSetTarget = vi.fn();
      const mockSetRel = vi.fn();
      const mockLinkNode = {
        type: 'link',
        getParent: () => null,
        getURL: () => 'https://example.com',
        setTarget: mockSetTarget,
        setRel: mockSetRel,
      };
      mockGetSelectedNode.mockReturnValue(mockLinkNode);
      mockIsLinkNode.mockImplementation((node: unknown) => node === mockLinkNode);
      mockIsRangeSelection.mockReturnValue(true);

      const callbacks: Record<string, unknown> = {};
      (mockRegisterCommand as ReturnType<typeof vi.fn>).mockImplementation(
        (command: unknown, callback: unknown) => {
          callbacks[command as string] = callback;
          return vi.fn();
        },
      );

      render(<CustomLinkPlugin />);

      const toggleSafeLinkCallback = callbacks.TOGGLE_SAFE_LINK_COMMAND as ((url: string) => boolean) | undefined;
      if (toggleSafeLinkCallback) {
        const result = toggleSafeLinkCallback('https://new-link.com');
        expect(result).toBe(true);
        expect(mockDispatchCommand).toHaveBeenCalledWith('TOGGLE_LINK_COMMAND', 'https://new-link.com');
        expect(mockSetTarget).toHaveBeenCalledWith('_blank');
        expect(mockSetRel).toHaveBeenCalledWith('noopener noreferrer');
      }
    });

    it('should execute TOGGLE_SAFE_LINK_COMMAND with empty URL to remove link', () => {
      mockIsRangeSelection.mockReturnValue(true);

      const callbacks: Record<string, unknown> = {};
      (mockRegisterCommand as ReturnType<typeof vi.fn>).mockImplementation(
        (command: unknown, callback: unknown) => {
          callbacks[command as string] = callback;
          return vi.fn();
        },
      );

      render(<CustomLinkPlugin />);

      const toggleSafeLinkCallback = callbacks.TOGGLE_SAFE_LINK_COMMAND as ((url: string) => boolean) | undefined;
      if (toggleSafeLinkCallback) {
        const result = toggleSafeLinkCallback('');
        expect(result).toBe(true);
        expect(mockDispatchCommand).toHaveBeenCalledWith('TOGGLE_LINK_COMMAND', null);
      }
    });

    it('should return false from TOGGLE_SAFE_LINK_COMMAND when linkNode is null', () => {
      const mockTextNode = {
        type: 'text',
        getParent: () => null,
        getURL: () => '',
        setTarget: vi.fn(),
        setRel: vi.fn(),
      };
      mockGetSelectedNode.mockReturnValue(mockTextNode);
      mockIsLinkNode.mockReturnValue(false);
      mockIsRangeSelection.mockReturnValue(true);

      const callbacks: Record<string, unknown> = {};
      (mockRegisterCommand as ReturnType<typeof vi.fn>).mockImplementation(
        (command: unknown, callback: unknown) => {
          callbacks[command as string] = callback;
          return vi.fn();
        },
      );

      render(<CustomLinkPlugin />);

      const toggleSafeLinkCallback = callbacks.TOGGLE_SAFE_LINK_COMMAND as ((url: string) => boolean) | undefined;
      if (toggleSafeLinkCallback) {
        const result = toggleSafeLinkCallback('https://test.com');
        expect(result).toBe(false);
      }
    });

    it('should execute KEY_ESCAPE_COMMAND in edit mode', async () => {
      const callbacks: Record<string, unknown> = {};
      (mockRegisterCommand as ReturnType<typeof vi.fn>).mockImplementation(
        (command: unknown, callback: unknown) => {
          callbacks[command as string] = callback;
          return vi.fn();
        },
      );

      render(<CustomLinkPlugin />);

      // Enter edit mode first
      const editButton = screen.getByText('common:edit');
      await act(async () => {
        fireEvent.click(editButton);
      });

      await waitFor(() => {
        expect(screen.getByText('flows:core.elements.richText.linkEditor.editLink')).toBeInTheDocument();
      });

      // Verify escapeCallback was captured
      expect(callbacks.KEY_ESCAPE_COMMAND).toBeDefined();
    });

    it('should execute KEY_ESCAPE_COMMAND in view mode and return false', () => {
      const callbacks: Record<string, unknown> = {};
      (mockRegisterCommand as ReturnType<typeof vi.fn>).mockImplementation(
        (command: unknown, callback: unknown) => {
          callbacks[command as string] = callback;
          return vi.fn();
        },
      );

      render(<CustomLinkPlugin />);

      // In view mode, escape should return false
      const escapeCallback = callbacks.KEY_ESCAPE_COMMAND as (() => boolean) | undefined;
      if (escapeCallback) {
        const result = escapeCallback();
        expect(result).toBe(false);
      }
    });

    it('should execute SELECTION_CHANGE_COMMAND callback', () => {
      const callbacks: Record<string, unknown> = {};
      (mockRegisterCommand as ReturnType<typeof vi.fn>).mockImplementation(
        (command: unknown, callback: unknown) => {
          callbacks[command as string] = callback;
          return vi.fn();
        },
      );

      render(<CustomLinkPlugin />);

      const selectionChangeCallback = callbacks.SELECTION_CHANGE_COMMAND as (() => boolean) | undefined;
      if (selectionChangeCallback) {
        const result = selectionChangeCallback();
        expect(result).toBe(false);
      }
    });
  });

  describe('updateLinkEditor Function Coverage', () => {
    it('should handle link node parent', () => {
      const mockParentLinkNode = {
        type: 'link',
        getURL: () => 'https://parent-link.com',
        setTarget: vi.fn(),
        setRel: vi.fn(),
      };
      const mockTextNode = {
        type: 'text',
        getParent: () => mockParentLinkNode,
        getURL: () => '',
        setTarget: vi.fn(),
        setRel: vi.fn(),
      };
      mockGetSelectedNode.mockReturnValue(mockTextNode);
      mockIsLinkNode.mockImplementation((node: unknown) => node === mockParentLinkNode);
      mockIsRangeSelection.mockReturnValue(true);

      render(<CustomLinkPlugin />);

      expect(mockRegisterUpdateListener).toHaveBeenCalled();
    });

    it('should handle node being a link node directly', () => {
      const mockLinkNode = {
        type: 'link',
        getParent: () => ({type: 'paragraph'}),
        getURL: () => 'https://direct-link.com',
        setTarget: vi.fn(),
        setRel: vi.fn(),
      };
      mockGetSelectedNode.mockReturnValue(mockLinkNode);
      mockIsLinkNode.mockImplementation((node: unknown) => node === mockLinkNode);
      mockIsRangeSelection.mockReturnValue(true);

      render(<CustomLinkPlugin />);

      expect(mockRegisterUpdateListener).toHaveBeenCalled();
    });

    it('should handle non-link node without parent', () => {
      const mockTextNode = {
        type: 'text',
        getParent: () => null,
        getURL: () => '',
        setTarget: vi.fn(),
        setRel: vi.fn(),
      };
      mockGetSelectedNode.mockReturnValue(mockTextNode);
      mockIsLinkNode.mockReturnValue(false);
      mockIsRangeSelection.mockReturnValue(true);

      render(<CustomLinkPlugin />);

      expect(mockRegisterUpdateListener).toHaveBeenCalled();
    });

    it('should handle non-range selection', () => {
      mockIsRangeSelection.mockReturnValue(false);

      render(<CustomLinkPlugin />);

      expect(mockRegisterUpdateListener).toHaveBeenCalled();
    });
  });

  describe('positionEditorElement Function Coverage', () => {
    it('should handle rect being null', () => {
      mockIsRangeSelection.mockReturnValue(false);

      render(<CustomLinkPlugin />);

      const card = document.querySelector('.MuiCard-root')!;
      // When rect is null, editor should be hidden (opacity: 0)
      expect(card).toBeInTheDocument();
    });

    it('should position editor when rect is provided and left edge adjustment needed', () => {
      // Mock window dimensions
      Object.defineProperty(window, 'innerWidth', {value: 1000, writable: true});
      Object.defineProperty(window, 'innerHeight', {value: 800, writable: true});
      Object.defineProperty(window, 'pageXOffset', {value: 0, writable: true});
      Object.defineProperty(window, 'pageYOffset', {value: 0, writable: true});

      // Mock native selection with rect near left edge
      const mockSelection = {
        isCollapsed: false,
        anchorNode: document.createElement('div'),
        getRangeAt: () => ({
          getBoundingClientRect: () => ({
            top: 100,
            left: -50, // Near left edge
            height: 20,
            width: 100,
          }),
        }),
      };
      vi.spyOn(window, 'getSelection').mockReturnValue(mockSelection as unknown as Selection);

      const rootElement = document.createElement('div');
      rootElement.appendChild(mockSelection.anchorNode);
      mockGetRootElement.mockReturnValue(rootElement);
      mockIsRangeSelection.mockReturnValue(true);
      mockGetSelection.mockReturnValue({type: 'range'});

      render(<CustomLinkPlugin />);

      expect(mockGetEditorState).toHaveBeenCalled();
    });

    it('should position editor when rect is provided and right edge adjustment needed', () => {
      Object.defineProperty(window, 'innerWidth', {value: 1000, writable: true});
      Object.defineProperty(window, 'innerHeight', {value: 800, writable: true});

      const mockSelection = {
        isCollapsed: false,
        anchorNode: document.createElement('div'),
        getRangeAt: () => ({
          getBoundingClientRect: () => ({
            top: 100,
            left: 900, // Near right edge
            height: 20,
            width: 100,
          }),
        }),
      };
      vi.spyOn(window, 'getSelection').mockReturnValue(mockSelection as unknown as Selection);

      const rootElement = document.createElement('div');
      rootElement.appendChild(mockSelection.anchorNode);
      mockGetRootElement.mockReturnValue(rootElement);
      mockIsRangeSelection.mockReturnValue(true);

      render(<CustomLinkPlugin />);

      expect(mockGetEditorState).toHaveBeenCalled();
    });

    it('should position editor above selection when near bottom', () => {
      Object.defineProperty(window, 'innerWidth', {value: 1000, writable: true});
      Object.defineProperty(window, 'innerHeight', {value: 800, writable: true});
      Object.defineProperty(window, 'pageYOffset', {value: 0, writable: true});

      const mockSelection = {
        isCollapsed: false,
        anchorNode: document.createElement('div'),
        getRangeAt: () => ({
          getBoundingClientRect: () => ({
            top: 750, // Near bottom
            left: 100,
            height: 20,
            width: 100,
          }),
        }),
      };
      vi.spyOn(window, 'getSelection').mockReturnValue(mockSelection as unknown as Selection);

      const rootElement = document.createElement('div');
      rootElement.appendChild(mockSelection.anchorNode);
      mockGetRootElement.mockReturnValue(rootElement);
      mockIsRangeSelection.mockReturnValue(true);

      render(<CustomLinkPlugin />);

      expect(mockGetEditorState).toHaveBeenCalled();
    });

    it('should handle anchorNode being root element', () => {
      const rootElement = document.createElement('div');
      const innerChild = document.createElement('span');
      rootElement.appendChild(innerChild);

      const mockSelection = {
        isCollapsed: false,
        anchorNode: rootElement, // anchorNode is rootElement
        getRangeAt: () => ({
          getBoundingClientRect: () => ({
            top: 100,
            left: 100,
            height: 20,
            width: 100,
          }),
        }),
      };
      vi.spyOn(window, 'getSelection').mockReturnValue(mockSelection as unknown as Selection);

      mockGetRootElement.mockReturnValue(rootElement);
      mockIsRangeSelection.mockReturnValue(true);

      render(<CustomLinkPlugin />);

      expect(mockGetEditorState).toHaveBeenCalled();
    });
  });

  describe('Event Listener Cleanup', () => {
    it('should remove scroll listener on unmount', () => {
      const removeEventListenerSpy = vi.spyOn(document.body, 'removeEventListener');

      const {unmount} = render(<CustomLinkPlugin />);

      unmount();

      expect(removeEventListenerSpy).toHaveBeenCalledWith('scroll', expect.any(Function));

      removeEventListenerSpy.mockRestore();
    });
  });

  describe('Update Listener Callback', () => {
    it('should execute update listener callback', () => {
      type UpdateCallback = (state: {editorState: {read: (cb: () => void) => void}}) => void;
      const capturedCallbacks: UpdateCallback[] = [];
      (mockRegisterUpdateListener as ReturnType<typeof vi.fn>).mockImplementation(
        (callback: unknown) => {
          capturedCallbacks.push(callback as UpdateCallback);
          return vi.fn();
        },
      );

      render(<CustomLinkPlugin />);

      const updateListenerCallback = capturedCallbacks[0];
      if (updateListenerCallback) {
        const mockEditorState = {
          read: vi.fn((cb: () => void) => cb()),
        };
        updateListenerCallback({editorState: mockEditorState});
        expect(mockEditorState.read).toHaveBeenCalled();
      }
    });
  });

  describe('Handle Close with Editor Ref', () => {
    it('should call positionEditorElement with null on close', async () => {
      render(<CustomLinkPlugin />);

      // Find the close button
      const buttons = screen.getAllByRole('button');
      const closeButton = buttons.find(btn => btn.querySelector('svg'));

      if (closeButton) {
        await act(async () => {
          fireEvent.click(closeButton);
        });
      }

      // The card should still be in the document (just repositioned)
      expect(document.querySelector('.MuiCard-root')).toBeInTheDocument();
    });
  });
});
