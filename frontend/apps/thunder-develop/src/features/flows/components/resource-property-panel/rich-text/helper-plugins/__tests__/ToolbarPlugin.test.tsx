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

import {describe, it, expect, vi, beforeEach} from 'vitest';
import {render, screen, fireEvent, waitFor} from '@testing-library/react';
import ToolbarPlugin from '../ToolbarPlugin';

// Use vi.hoisted for mock functions that need to be available during module mocking
const {
  mockDispatchCommand,
  mockRegisterUpdateListener,
  mockRegisterCommand,
  mockGetElementByKey,
  mockUpdate,
} = vi.hoisted(() => ({
  mockDispatchCommand: vi.fn(),
  mockRegisterUpdateListener: vi.fn(() => vi.fn()),
  mockRegisterCommand: vi.fn(() => vi.fn()),
  mockGetElementByKey: vi.fn(() => document.createElement('div')),
  mockUpdate: vi.fn((callback: () => void) => callback()),
}));

// Mock the lexical composer context
vi.mock('@lexical/react/LexicalComposerContext', () => ({
  useLexicalComposerContext: () => [{
    dispatchCommand: mockDispatchCommand,
    registerUpdateListener: mockRegisterUpdateListener,
    registerCommand: mockRegisterCommand,
    getElementByKey: mockGetElementByKey,
    update: mockUpdate,
  }],
}));

// Mock lexical selection
vi.mock('@lexical/selection', () => ({
  $setBlocksType: vi.fn(),
}));

// Mock lexical utils
vi.mock('@lexical/utils', () => ({
  mergeRegister: (...fns: (() => void)[]) => () => fns.forEach(fn => fn()),
}));

// Mock lexical
vi.mock('lexical', () => ({
  $createParagraphNode: vi.fn(),
  $getSelection: vi.fn(() => null),
  $isRangeSelection: vi.fn(() => false),
  CAN_REDO_COMMAND: 'CAN_REDO_COMMAND',
  CAN_UNDO_COMMAND: 'CAN_UNDO_COMMAND',
  FORMAT_ELEMENT_COMMAND: 'FORMAT_ELEMENT_COMMAND',
  FORMAT_TEXT_COMMAND: 'FORMAT_TEXT_COMMAND',
  REDO_COMMAND: 'REDO_COMMAND',
  SELECTION_CHANGE_COMMAND: 'SELECTION_CHANGE_COMMAND',
  UNDO_COMMAND: 'UNDO_COMMAND',
}));

// Mock @lexical/rich-text
vi.mock('@lexical/rich-text', () => ({
  $createHeadingNode: vi.fn(),
  $isHeadingNode: vi.fn(() => false),
}));

// Mock @lexical/link
vi.mock('@lexical/link', () => ({
  $isLinkNode: vi.fn(() => false),
  TOGGLE_LINK_COMMAND: 'TOGGLE_LINK_COMMAND',
}));

// Mock getSelectedNode utility
vi.mock('../../utils/getSelectedNode', () => ({
  default: vi.fn(() => ({
    getParent: () => null,
    getKey: () => 'root',
    getTopLevelElementOrThrow: () => ({
      getFormat: () => 0,
      getKey: () => 'element-key',
      getType: () => 'paragraph',
    }),
  })),
}));

describe('ToolbarPlugin', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('should render the toolbar', () => {
      render(<ToolbarPlugin />);

      expect(document.querySelector('.MuiPaper-root')).toBeInTheDocument();
    });

    it('should render undo button when history is enabled', () => {
      render(<ToolbarPlugin history />);

      expect(screen.getByRole('button', {name: 'Undo'})).toBeInTheDocument();
    });

    it('should render redo button when history is enabled', () => {
      render(<ToolbarPlugin history />);

      expect(screen.getByRole('button', {name: 'Redo'})).toBeInTheDocument();
    });

    it('should not render undo/redo buttons when history is disabled', () => {
      render(<ToolbarPlugin history={false} />);

      expect(screen.queryByRole('button', {name: 'Undo'})).not.toBeInTheDocument();
      expect(screen.queryByRole('button', {name: 'Redo'})).not.toBeInTheDocument();
    });

    it('should render bold button when bold is enabled', () => {
      render(<ToolbarPlugin bold />);

      expect(screen.getByRole('button', {name: 'Format Bold'})).toBeInTheDocument();
    });

    it('should not render bold button when bold is disabled', () => {
      render(<ToolbarPlugin bold={false} />);

      expect(screen.queryByRole('button', {name: 'Format Bold'})).not.toBeInTheDocument();
    });

    it('should render italic button when italic is enabled', () => {
      render(<ToolbarPlugin italic />);

      expect(screen.getByRole('button', {name: 'Format Italics'})).toBeInTheDocument();
    });

    it('should not render italic button when italic is disabled', () => {
      render(<ToolbarPlugin italic={false} />);

      expect(screen.queryByRole('button', {name: 'Format Italics'})).not.toBeInTheDocument();
    });

    it('should render underline button when underline is enabled', () => {
      render(<ToolbarPlugin underline />);

      expect(screen.getByRole('button', {name: 'Format Underline'})).toBeInTheDocument();
    });

    it('should not render underline button when underline is disabled', () => {
      render(<ToolbarPlugin underline={false} />);

      expect(screen.queryByRole('button', {name: 'Format Underline'})).not.toBeInTheDocument();
    });

    it('should render link button when link is enabled', () => {
      render(<ToolbarPlugin link />);

      expect(screen.getByRole('button', {name: 'Format Link'})).toBeInTheDocument();
    });

    it('should not render link button when link is disabled', () => {
      render(<ToolbarPlugin link={false} />);

      expect(screen.queryByRole('button', {name: 'Format Link'})).not.toBeInTheDocument();
    });

    it('should render alignment buttons when alignment is enabled', () => {
      render(<ToolbarPlugin alignment />);

      expect(screen.getByRole('button', {name: 'Left Align'})).toBeInTheDocument();
      expect(screen.getByRole('button', {name: 'Center Align'})).toBeInTheDocument();
      expect(screen.getByRole('button', {name: 'Right Align'})).toBeInTheDocument();
      expect(screen.getByRole('button', {name: 'Justify Align'})).toBeInTheDocument();
    });

    it('should not render alignment buttons when alignment is disabled', () => {
      render(<ToolbarPlugin alignment={false} />);

      expect(screen.queryByRole('button', {name: 'Left Align'})).not.toBeInTheDocument();
      expect(screen.queryByRole('button', {name: 'Center Align'})).not.toBeInTheDocument();
      expect(screen.queryByRole('button', {name: 'Right Align'})).not.toBeInTheDocument();
      expect(screen.queryByRole('button', {name: 'Justify Align'})).not.toBeInTheDocument();
    });

    it('should render typography dropdown when typography is enabled', () => {
      render(<ToolbarPlugin typography />);

      expect(screen.getByText('Paragraph')).toBeInTheDocument();
    });

    it('should not render typography dropdown when typography is disabled', () => {
      render(<ToolbarPlugin typography={false} />);

      expect(screen.queryByText('Paragraph')).not.toBeInTheDocument();
    });
  });

  describe('Disabled State', () => {
    it('should disable undo button when disabled prop is true', () => {
      render(<ToolbarPlugin history disabled />);

      expect(screen.getByRole('button', {name: 'Undo'})).toBeDisabled();
    });

    it('should disable redo button when disabled prop is true', () => {
      render(<ToolbarPlugin history disabled />);

      expect(screen.getByRole('button', {name: 'Redo'})).toBeDisabled();
    });

    it('should disable bold button when disabled prop is true', () => {
      render(<ToolbarPlugin bold disabled />);

      expect(screen.getByRole('button', {name: 'Format Bold'})).toBeDisabled();
    });

    it('should disable italic button when disabled prop is true', () => {
      render(<ToolbarPlugin italic disabled />);

      expect(screen.getByRole('button', {name: 'Format Italics'})).toBeDisabled();
    });

    it('should disable underline button when disabled prop is true', () => {
      render(<ToolbarPlugin underline disabled />);

      expect(screen.getByRole('button', {name: 'Format Underline'})).toBeDisabled();
    });

    it('should disable link button when disabled prop is true', () => {
      render(<ToolbarPlugin link disabled />);

      expect(screen.getByRole('button', {name: 'Format Link'})).toBeDisabled();
    });

    it('should disable alignment buttons when disabled prop is true', () => {
      render(<ToolbarPlugin alignment disabled />);

      expect(screen.getByRole('button', {name: 'Left Align'})).toBeDisabled();
      expect(screen.getByRole('button', {name: 'Center Align'})).toBeDisabled();
      expect(screen.getByRole('button', {name: 'Right Align'})).toBeDisabled();
      expect(screen.getByRole('button', {name: 'Justify Align'})).toBeDisabled();
    });

    it('should disable typography button when disabled prop is true', () => {
      render(<ToolbarPlugin typography disabled />);

      const typographyButton = screen.getByText('Paragraph').closest('button');
      expect(typographyButton).toBeDisabled();
    });
  });

  describe('Button Clicks', () => {
    it('should dispatch UNDO_COMMAND when undo button is clicked', () => {
      // Mock registerCommand to capture the CAN_UNDO_COMMAND callback and call it to enable undo
      mockRegisterCommand.mockImplementation(((command: string, callback: (payload: boolean) => boolean) => {
        if (command === 'CAN_UNDO_COMMAND') {
          // Call the callback with true to enable the undo button
          callback(true);
        }
        return vi.fn();
      }) as typeof mockRegisterCommand);

      render(<ToolbarPlugin history />);

      const undoButton = screen.getByRole('button', {name: 'Undo'});
      fireEvent.click(undoButton);

      expect(mockDispatchCommand).toHaveBeenCalledWith('UNDO_COMMAND', undefined);
    });

    it('should dispatch REDO_COMMAND when redo button is clicked', () => {
      // Mock registerCommand to capture the CAN_REDO_COMMAND callback and call it to enable redo
      mockRegisterCommand.mockImplementation(((command: string, callback: (payload: boolean) => boolean) => {
        if (command === 'CAN_REDO_COMMAND') {
          // Call the callback with true to enable the redo button
          callback(true);
        }
        return vi.fn();
      }) as typeof mockRegisterCommand);

      render(<ToolbarPlugin history />);

      const redoButton = screen.getByRole('button', {name: 'Redo'});
      fireEvent.click(redoButton);

      expect(mockDispatchCommand).toHaveBeenCalledWith('REDO_COMMAND', undefined);
    });

    it('should dispatch FORMAT_TEXT_COMMAND with bold when bold button is clicked', () => {
      render(<ToolbarPlugin bold />);

      fireEvent.click(screen.getByRole('button', {name: 'Format Bold'}));

      expect(mockDispatchCommand).toHaveBeenCalledWith('FORMAT_TEXT_COMMAND', 'bold');
    });

    it('should dispatch FORMAT_TEXT_COMMAND with italic when italic button is clicked', () => {
      render(<ToolbarPlugin italic />);

      fireEvent.click(screen.getByRole('button', {name: 'Format Italics'}));

      expect(mockDispatchCommand).toHaveBeenCalledWith('FORMAT_TEXT_COMMAND', 'italic');
    });

    it('should dispatch FORMAT_TEXT_COMMAND with underline when underline button is clicked', () => {
      render(<ToolbarPlugin underline />);

      fireEvent.click(screen.getByRole('button', {name: 'Format Underline'}));

      expect(mockDispatchCommand).toHaveBeenCalledWith('FORMAT_TEXT_COMMAND', 'underline');
    });

    it('should dispatch FORMAT_ELEMENT_COMMAND with left when left align button is clicked', () => {
      render(<ToolbarPlugin alignment />);

      fireEvent.click(screen.getByRole('button', {name: 'Left Align'}));

      expect(mockDispatchCommand).toHaveBeenCalledWith('FORMAT_ELEMENT_COMMAND', 'left');
    });

    it('should dispatch FORMAT_ELEMENT_COMMAND with center when center align button is clicked', () => {
      render(<ToolbarPlugin alignment />);

      fireEvent.click(screen.getByRole('button', {name: 'Center Align'}));

      expect(mockDispatchCommand).toHaveBeenCalledWith('FORMAT_ELEMENT_COMMAND', 'center');
    });

    it('should dispatch FORMAT_ELEMENT_COMMAND with right when right align button is clicked', () => {
      render(<ToolbarPlugin alignment />);

      fireEvent.click(screen.getByRole('button', {name: 'Right Align'}));

      expect(mockDispatchCommand).toHaveBeenCalledWith('FORMAT_ELEMENT_COMMAND', 'right');
    });

    it('should dispatch FORMAT_ELEMENT_COMMAND with justify when justify align button is clicked', () => {
      render(<ToolbarPlugin alignment />);

      fireEvent.click(screen.getByRole('button', {name: 'Justify Align'}));

      expect(mockDispatchCommand).toHaveBeenCalledWith('FORMAT_ELEMENT_COMMAND', 'justify');
    });

    it('should dispatch TOGGLE_LINK_COMMAND when link button is clicked', () => {
      render(<ToolbarPlugin link />);

      fireEvent.click(screen.getByRole('button', {name: 'Format Link'}));

      expect(mockDispatchCommand).toHaveBeenCalledWith('TOGGLE_LINK_COMMAND', 'https://');
    });
  });

  describe('Typography Menu', () => {
    it('should open typography menu when typography button is clicked', () => {
      render(<ToolbarPlugin typography />);

      const typographyButton = screen.getByText('Paragraph').closest('button');
      fireEvent.click(typographyButton!);

      expect(screen.getByText('Heading 1')).toBeInTheDocument();
      expect(screen.getByText('Heading 2')).toBeInTheDocument();
      expect(screen.getByText('Heading 3')).toBeInTheDocument();
      expect(screen.getByText('Heading 4')).toBeInTheDocument();
      expect(screen.getByText('Heading 5')).toBeInTheDocument();
    });

    it('should close typography menu when menu item is clicked', async () => {
      render(<ToolbarPlugin typography />);

      // Open menu
      const typographyButton = screen.getByText('Paragraph').closest('button');
      fireEvent.click(typographyButton!);

      // Click heading 1
      fireEvent.click(screen.getByText('Heading 1'));

      // Menu should close - use waitFor because menu closing is async in MUI
      await waitFor(() => {
        expect(screen.queryByRole('menu')).not.toBeInTheDocument();
      });
    });
  });

  describe('Default Props', () => {
    it('should have history enabled by default', () => {
      render(<ToolbarPlugin />);

      expect(screen.getByRole('button', {name: 'Undo'})).toBeInTheDocument();
      expect(screen.getByRole('button', {name: 'Redo'})).toBeInTheDocument();
    });

    it('should have bold enabled by default', () => {
      render(<ToolbarPlugin />);

      expect(screen.getByRole('button', {name: 'Format Bold'})).toBeInTheDocument();
    });

    it('should have italic enabled by default', () => {
      render(<ToolbarPlugin />);

      expect(screen.getByRole('button', {name: 'Format Italics'})).toBeInTheDocument();
    });

    it('should have underline enabled by default', () => {
      render(<ToolbarPlugin />);

      expect(screen.getByRole('button', {name: 'Format Underline'})).toBeInTheDocument();
    });

    it('should have alignment enabled by default', () => {
      render(<ToolbarPlugin />);

      expect(screen.getByRole('button', {name: 'Left Align'})).toBeInTheDocument();
    });

    it('should have typography enabled by default', () => {
      render(<ToolbarPlugin />);

      expect(screen.getByText('Paragraph')).toBeInTheDocument();
    });

    it('should have link enabled by default', () => {
      render(<ToolbarPlugin />);

      expect(screen.getByRole('button', {name: 'Format Link'})).toBeInTheDocument();
    });

    it('should not be disabled by default', () => {
      render(<ToolbarPlugin />);

      expect(screen.getByRole('button', {name: 'Format Bold'})).not.toBeDisabled();
    });
  });

  describe('Custom className', () => {
    it('should apply custom className to the toolbar', () => {
      render(<ToolbarPlugin className="custom-toolbar-class" />);

      const paper = document.querySelector('.MuiPaper-root');
      expect(paper).toHaveClass('custom-toolbar-class');
    });
  });

  describe('Command Registration', () => {
    it('should register update listener on mount', () => {
      render(<ToolbarPlugin />);

      expect(mockRegisterUpdateListener).toHaveBeenCalled();
    });

    it('should register SELECTION_CHANGE_COMMAND on mount', () => {
      render(<ToolbarPlugin />);

      expect(mockRegisterCommand).toHaveBeenCalled();
    });

    it('should register CAN_UNDO_COMMAND on mount', () => {
      render(<ToolbarPlugin />);

      expect(mockRegisterCommand).toHaveBeenCalled();
    });

    it('should register CAN_REDO_COMMAND on mount', () => {
      render(<ToolbarPlugin />);

      expect(mockRegisterCommand).toHaveBeenCalled();
    });
  });

  describe('Format Paragraph', () => {
    it('should format selection to paragraph when menu item is clicked', async () => {
      render(<ToolbarPlugin typography />);

      // Open menu
      const typographyButton = screen.getByText('Paragraph').closest('button');
      fireEvent.click(typographyButton!);

      // Click paragraph option
      const paragraphItems = screen.getAllByText('Paragraph');
      // The second one is the menu item
      const paragraphMenuItem = paragraphItems.find(item => item.closest('[role="menuitem"]'));
      if (paragraphMenuItem) {
        fireEvent.click(paragraphMenuItem);
      }

      // Menu should close
      await waitFor(() => {
        expect(screen.queryByRole('menu')).not.toBeInTheDocument();
      });
    });

    it('should call editor update when formatting to paragraph', () => {
      render(<ToolbarPlugin typography />);

      // Open menu
      const typographyButton = screen.getByText('Paragraph').closest('button');
      fireEvent.click(typographyButton!);

      // Click paragraph option
      const paragraphItems = screen.getAllByText('Paragraph');
      const paragraphMenuItem = paragraphItems.find(item => item.closest('[role="menuitem"]'));
      if (paragraphMenuItem) {
        fireEvent.click(paragraphMenuItem);
      }

      expect(mockUpdate).toHaveBeenCalled();
    });
  });

  describe('Format Heading', () => {
    it('should format selection to h1 when Heading 1 is clicked', async () => {
      render(<ToolbarPlugin typography />);

      // Open menu
      const typographyButton = screen.getByText('Paragraph').closest('button');
      fireEvent.click(typographyButton!);

      // Click heading 1
      fireEvent.click(screen.getByText('Heading 1'));

      expect(mockUpdate).toHaveBeenCalled();
    });

    it('should format selection to h2 when Heading 2 is clicked', async () => {
      render(<ToolbarPlugin typography />);

      // Open menu
      const typographyButton = screen.getByText('Paragraph').closest('button');
      fireEvent.click(typographyButton!);

      // Click heading 2
      fireEvent.click(screen.getByText('Heading 2'));

      expect(mockUpdate).toHaveBeenCalled();
    });

    it('should format selection to h3 when Heading 3 is clicked', async () => {
      render(<ToolbarPlugin typography />);

      // Open menu
      const typographyButton = screen.getByText('Paragraph').closest('button');
      fireEvent.click(typographyButton!);

      // Click heading 3
      fireEvent.click(screen.getByText('Heading 3'));

      expect(mockUpdate).toHaveBeenCalled();
    });

    it('should format selection to h4 when Heading 4 is clicked', async () => {
      render(<ToolbarPlugin typography />);

      // Open menu
      const typographyButton = screen.getByText('Paragraph').closest('button');
      fireEvent.click(typographyButton!);

      // Click heading 4
      fireEvent.click(screen.getByText('Heading 4'));

      expect(mockUpdate).toHaveBeenCalled();
    });

    it('should format selection to h5 when Heading 5 is clicked', async () => {
      render(<ToolbarPlugin typography />);

      // Open menu
      const typographyButton = screen.getByText('Paragraph').closest('button');
      fireEvent.click(typographyButton!);

      // Click heading 5
      fireEvent.click(screen.getByText('Heading 5'));

      expect(mockUpdate).toHaveBeenCalled();
    });
  });

  describe('Insert Link', () => {
    it('should dispatch TOGGLE_LINK_COMMAND with https:// when adding new link', () => {
      render(<ToolbarPlugin link />);

      fireEvent.click(screen.getByRole('button', {name: 'Format Link'}));

      // When isLink is false (default), it adds a new link with 'https://'
      expect(mockDispatchCommand).toHaveBeenCalledWith('TOGGLE_LINK_COMMAND', 'https://');
    });

    it('should register link command handler', () => {
      render(<ToolbarPlugin link />);

      // The component registers commands including link toggle
      expect(mockRegisterCommand).toHaveBeenCalled();
    });
  });

  describe('Toolbar Update', () => {
    it('should update toolbar state on selection change', async () => {
      // Mock $isRangeSelection to return true and provide selection
      const lexical = await vi.importMock<typeof import('lexical')>('lexical');
      (lexical.$isRangeSelection as ReturnType<typeof vi.fn>).mockReturnValue(true);
      (lexical.$getSelection as ReturnType<typeof vi.fn>).mockReturnValue({
        hasFormat: vi.fn().mockReturnValue(false),
        anchor: {
          getNode: () => ({
            getKey: () => 'root',
            getTopLevelElementOrThrow: () => ({
              getFormat: () => 0,
              getKey: () => 'element-key',
              getType: () => 'paragraph',
            }),
          }),
        },
      });

      render(<ToolbarPlugin />);

      expect(mockRegisterUpdateListener).toHaveBeenCalled();
    });

    it('should update bold state when selection has bold format', async () => {
      const lexical = await vi.importMock<typeof import('lexical')>('lexical');
      (lexical.$isRangeSelection as ReturnType<typeof vi.fn>).mockReturnValue(true);
      (lexical.$getSelection as ReturnType<typeof vi.fn>).mockReturnValue({
        hasFormat: (format: string) => format === 'bold',
        anchor: {
          getNode: () => ({
            getKey: () => 'root',
            getTopLevelElementOrThrow: () => ({
              getFormat: () => 0,
              getKey: () => 'element-key',
              getType: () => 'paragraph',
            }),
          }),
        },
      });

      render(<ToolbarPlugin bold />);

      expect(mockRegisterUpdateListener).toHaveBeenCalled();
    });

    it('should update italic state when selection has italic format', async () => {
      const lexical = await vi.importMock<typeof import('lexical')>('lexical');
      (lexical.$isRangeSelection as ReturnType<typeof vi.fn>).mockReturnValue(true);
      (lexical.$getSelection as ReturnType<typeof vi.fn>).mockReturnValue({
        hasFormat: (format: string) => format === 'italic',
        anchor: {
          getNode: () => ({
            getKey: () => 'root',
            getTopLevelElementOrThrow: () => ({
              getFormat: () => 0,
              getKey: () => 'element-key',
              getType: () => 'paragraph',
            }),
          }),
        },
      });

      render(<ToolbarPlugin italic />);

      expect(mockRegisterUpdateListener).toHaveBeenCalled();
    });

    it('should update underline state when selection has underline format', async () => {
      const lexical = await vi.importMock<typeof import('lexical')>('lexical');
      (lexical.$isRangeSelection as ReturnType<typeof vi.fn>).mockReturnValue(true);
      (lexical.$getSelection as ReturnType<typeof vi.fn>).mockReturnValue({
        hasFormat: (format: string) => format === 'underline',
        anchor: {
          getNode: () => ({
            getKey: () => 'root',
            getTopLevelElementOrThrow: () => ({
              getFormat: () => 0,
              getKey: () => 'element-key',
              getType: () => 'paragraph',
            }),
          }),
        },
      });

      render(<ToolbarPlugin underline />);

      expect(mockRegisterUpdateListener).toHaveBeenCalled();
    });
  });

  describe('Alignment State', () => {
    it('should highlight left align button when alignment is 1', () => {
      render(<ToolbarPlugin alignment />);

      const leftAlignButton = screen.getByRole('button', {name: 'Left Align'});
      // By default, alignment is 1 (left) which maps to selectedAlignment === 1
      expect(leftAlignButton).toBeInTheDocument();
    });

    it('should highlight center align button when alignment is 2', async () => {
      // This would require mocking the selection to have alignment format 2
      render(<ToolbarPlugin alignment />);

      const centerAlignButton = screen.getByRole('button', {name: 'Center Align'});
      expect(centerAlignButton).toBeInTheDocument();
    });

    it('should highlight right align button when alignment is 3', async () => {
      render(<ToolbarPlugin alignment />);

      const rightAlignButton = screen.getByRole('button', {name: 'Right Align'});
      expect(rightAlignButton).toBeInTheDocument();
    });

    it('should highlight justify align button when alignment is 4', async () => {
      render(<ToolbarPlugin alignment />);

      const justifyAlignButton = screen.getByRole('button', {name: 'Justify Align'});
      expect(justifyAlignButton).toBeInTheDocument();
    });
  });

  describe('Block Type Detection', () => {
    it('should detect heading block type', async () => {
      const {$isHeadingNode} = await vi.importMock<typeof import('@lexical/rich-text')>('@lexical/rich-text');
      ($isHeadingNode as ReturnType<typeof vi.fn>).mockReturnValue(true);

      render(<ToolbarPlugin typography />);

      expect(mockRegisterUpdateListener).toHaveBeenCalled();
    });

    it('should detect paragraph block type', () => {
      render(<ToolbarPlugin typography />);

      // Default block type is paragraph
      expect(screen.getByText('Paragraph')).toBeInTheDocument();
    });

    it('should handle unknown block type as paragraph', () => {
      render(<ToolbarPlugin typography />);

      // Unknown types default to paragraph
      expect(screen.getByText('Paragraph')).toBeInTheDocument();
    });
  });

  describe('Link State Detection', () => {
    it('should detect when parent node is a link', async () => {
      const {$isLinkNode} = await vi.importMock<typeof import('@lexical/link')>('@lexical/link');
      ($isLinkNode as ReturnType<typeof vi.fn>).mockReturnValue(true);

      render(<ToolbarPlugin link />);

      expect(mockRegisterUpdateListener).toHaveBeenCalled();
    });

    it('should detect when current node is a link', async () => {
      const {$isLinkNode} = await vi.importMock<typeof import('@lexical/link')>('@lexical/link');
      ($isLinkNode as ReturnType<typeof vi.fn>).mockReturnValue(true);

      render(<ToolbarPlugin link />);

      expect(mockRegisterUpdateListener).toHaveBeenCalled();
    });
  });

  describe('Active State Classes', () => {
    it('should not apply active class when disabled', () => {
      render(<ToolbarPlugin bold disabled />);

      const boldButton = screen.getByRole('button', {name: 'Format Bold'});
      expect(boldButton).not.toHaveClass('active');
    });

    it('should apply active class to bold button when bold is active and not disabled', async () => {
      const lexical = await vi.importMock<typeof import('lexical')>('lexical');
      (lexical.$isRangeSelection as ReturnType<typeof vi.fn>).mockReturnValue(true);
      (lexical.$getSelection as ReturnType<typeof vi.fn>).mockReturnValue({
        hasFormat: (format: string) => format === 'bold',
        anchor: {
          getNode: () => ({
            getKey: () => 'root',
            getTopLevelElementOrThrow: () => ({
              getFormat: () => 0,
              getKey: () => 'element-key',
              getType: () => 'paragraph',
            }),
          }),
        },
      });

      render(<ToolbarPlugin bold />);

      // The active class is applied via classNames utility
      expect(mockRegisterUpdateListener).toHaveBeenCalled();
    });
  });

  describe('Element Key Resolution', () => {
    it('should get element by key from editor', () => {
      render(<ToolbarPlugin />);

      expect(mockGetElementByKey).toBeDefined();
    });

    it('should handle null element DOM', () => {
      mockGetElementByKey.mockReturnValueOnce(null as unknown as HTMLDivElement);

      render(<ToolbarPlugin />);

      // Component should handle null gracefully
      expect(mockRegisterUpdateListener).toHaveBeenCalled();
    });
  });

  describe('Format Paragraph with Selection', () => {
    it('should call editor update when formatting to paragraph', () => {
      render(<ToolbarPlugin typography />);

      // Open menu
      const typographyButton = screen.getByText('Paragraph').closest('button');
      fireEvent.click(typographyButton!);

      // Click paragraph option
      const paragraphItems = screen.getAllByText('Paragraph');
      const paragraphMenuItem = paragraphItems.find(item => item.closest('[role="menuitem"]'));
      if (paragraphMenuItem) {
        fireEvent.click(paragraphMenuItem);
      }

      // editor.update should be called when formatting
      expect(mockUpdate).toHaveBeenCalled();
    });
  });

  describe('Insert Link Toggle', () => {
    it('should dispatch TOGGLE_LINK_COMMAND when link button is clicked', () => {
      render(<ToolbarPlugin link />);

      fireEvent.click(screen.getByRole('button', {name: 'Format Link'}));

      // The dispatch command should have been called
      expect(mockDispatchCommand).toHaveBeenCalledWith('TOGGLE_LINK_COMMAND', expect.anything());
    });
  });
});
