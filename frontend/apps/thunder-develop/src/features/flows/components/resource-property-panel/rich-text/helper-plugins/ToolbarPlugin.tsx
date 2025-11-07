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

import {$isLinkNode, TOGGLE_LINK_COMMAND} from '@lexical/link';
import {useLexicalComposerContext} from '@lexical/react/LexicalComposerContext';
import {$createHeadingNode, $isHeadingNode, type HeadingTagType} from '@lexical/rich-text';
import {$setBlocksType} from '@lexical/selection';
import {mergeRegister} from '@lexical/utils';

import classNames from 'classnames';
import {
  $createParagraphNode,
  $getSelection,
  $isRangeSelection,
  CAN_REDO_COMMAND,
  CAN_UNDO_COMMAND,
  FORMAT_ELEMENT_COMMAND,
  FORMAT_TEXT_COMMAND,
  REDO_COMMAND,
  SELECTION_CHANGE_COMMAND,
  UNDO_COMMAND,
} from 'lexical';
import type {CommandListenerPriority, EditorState, ElementNode, BaseSelection, NodeKey, TextNode} from 'lexical';
import React, {type HTMLAttributes, type ReactElement, type SVGProps, useCallback, useEffect, useState} from 'react';
import {Button, Divider, IconButton, ListItemText, Menu, MenuItem, Paper, Stack} from '@wso2/oxygen-ui';
import {ChevronDownIcon, LinkIcon} from '@wso2/oxygen-ui-icons-react';
import getSelectedNode from '../utils/getSelectedNode';
import './ToolbarPlugin.scss';

const LowPriority: CommandListenerPriority = 1;

// Icon Components
function ArrowCounterClockwiseIcon({...rest}: SVGProps<SVGSVGElement>): ReactElement {
  return (
    <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" viewBox="0 0 16 16" {...rest}>
      <path fillRule="evenodd" d="M8 3a5 5 0 1 1-4.546 2.914.5.5 0 0 0-.908-.417A6 6 0 1 0 8 2v1z" />
      <path d="M8 4.466V.534a.25.25 0 0 0-.41-.192L5.23 2.308a.25.25 0 0 0 0 .384l2.36 1.966A.25.25 0 0 0 8 4.466z" />
    </svg>
  );
}

function ArrowClockwiseIcon({...rest}: SVGProps<SVGSVGElement>): ReactElement {
  return (
    <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" viewBox="0 0 16 16" {...rest}>
      <path fillRule="evenodd" d="M8 3a5 5 0 1 0 4.546 2.914.5.5 0 0 1 .908-.417A6 6 0 1 1 8 2v1z" />
      <path d="M8 4.466V.534a.25.25 0 0 1 .41-.192l2.36 1.966c.12.1.12.284 0 .384L8.41 4.658A.25.25 0 0 1 8 4.466z" />
    </svg>
  );
}

function BoldIcon({...rest}: SVGProps<SVGSVGElement>): ReactElement {
  return (
    <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" viewBox="0 0 16 16" {...rest}>
      {/* eslint-disable-next-line max-len */}
      <path d="M8.21 13c2.106 0 3.412-1.087 3.412-2.823 0-1.306-.984-2.283-2.324-2.386v-.055a2.176 2.176 0 0 0 1.852-2.14c0-1.51-1.162-2.46-3.014-2.46H3.843V13H8.21zM5.908 4.674h1.696c.963 0 1.517.451 1.517 1.244 0 .834-.629 1.32-1.73 1.32H5.908V4.673zm0 6.788V8.598h1.73c1.217 0 1.88.492 1.88 1.415 0 .943-.643 1.449-1.832 1.449H5.907z" />
    </svg>
  );
}

function ItalicIcon({...rest}: SVGProps<SVGSVGElement>): ReactElement {
  return (
    <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" viewBox="0 0 16 16" {...rest}>
      {/* eslint-disable-next-line max-len */}
      <path d="M7.991 11.674 9.53 4.455c.123-.595.246-.71 1.347-.807l.11-.52H7.211l-.11.52c1.06.096 1.128.212 1.005.807L6.57 11.674c-.123.595-.246.71-1.346.806l-.11.52h3.774l.11-.52c-1.06-.095-1.129-.211-1.006-.806z" />
    </svg>
  );
}

function UnderlineIcon({...rest}: SVGProps<SVGSVGElement>): ReactElement {
  return (
    <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" viewBox="0 0 16 16" {...rest}>
      {/* eslint-disable-next-line max-len */}
      <path d="M5.313 3.136h-1.23V9.54c0 2.105 1.47 3.623 3.917 3.623s3.917-1.518 3.917-3.623V3.136h-1.23v6.323c0 1.49-.978 2.57-2.687 2.57-1.709 0-2.687-1.08-2.687-2.57V3.136zM12.5 15h-9v-1h9v1z" />
    </svg>
  );
}

function LeftAlignIcon({...rest}: SVGProps<SVGSVGElement>): ReactElement {
  return (
    <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" viewBox="0 0 16 16" {...rest}>
      <path
        fillRule="evenodd"
        // eslint-disable-next-line max-len
        d="M2 12.5a.5.5 0 0 1 .5-.5h7a.5.5 0 0 1 0 1h-7a.5.5 0 0 1-.5-.5zm0-3a.5.5 0 0 1 .5-.5h11a.5.5 0 0 1 0 1h-11a.5.5 0 0 1-.5-.5zm0-3a.5.5 0 0 1 .5-.5h7a.5.5 0 0 1 0 1h-7a.5.5 0 0 1-.5-.5zm0-3a.5.5 0 0 1 .5-.5h11a.5.5 0 0 1 0 1h-11a.5.5 0 0 1-.5-.5z"
      />
    </svg>
  );
}

function RightAlignIcon({...rest}: SVGProps<SVGSVGElement>): ReactElement {
  return (
    <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" viewBox="0 0 16 16" {...rest}>
      <path
        fillRule="evenodd"
        // eslint-disable-next-line max-len
        d="M6 12.5a.5.5 0 0 1 .5-.5h7a.5.5 0 0 1 0 1h-7a.5.5 0 0 1-.5-.5zm-4-3a.5.5 0 0 1 .5-.5h11a.5.5 0 0 1 0 1h-11a.5.5 0 0 1-.5-.5zm4-3a.5.5 0 0 1 .5-.5h7a.5.5 0 0 1 0 1h-7a.5.5 0 0 1-.5-.5zm-4-3a.5.5 0 0 1 .5-.5h11a.5.5 0 0 1 0 1h-11a.5.5 0 0 1-.5-.5z"
      />
    </svg>
  );
}

function CenterAlignIcon({...rest}: SVGProps<SVGSVGElement>): ReactElement {
  return (
    <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" viewBox="0 0 16 16" {...rest}>
      <path
        fillRule="evenodd"
        // eslint-disable-next-line max-len
        d="M4 12.5a.5.5 0 0 1 .5-.5h7a.5.5 0 0 1 0 1h-7a.5.5 0 0 1-.5-.5zm-2-3a.5.5 0 0 1 .5-.5h11a.5.5 0 0 1 0 1h-11a.5.5 0 0 1-.5-.5zm2-3a.5.5 0 0 1 .5-.5h7a.5.5 0 0 1 0 1h-7a.5.5 0 0 1-.5-.5zm-2-3a.5.5 0 0 1 .5-.5h11a.5.5 0 0 1 0 1h-11a.5.5 0 0 1-.5-.5z"
      />
    </svg>
  );
}

function JustifyAlignIcon({...rest}: SVGProps<SVGSVGElement>): ReactElement {
  return (
    <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" viewBox="0 0 16 16" {...rest}>
      <path
        fillRule="evenodd"
        // eslint-disable-next-line max-len
        d="M2 12.5a.5.5 0 0 1 .5-.5h11a.5.5 0 0 1 0 1h-11a.5.5 0 0 1-.5-.5zm0-3a.5.5 0 0 1 .5-.5h11a.5.5 0 0 1 0 1h-11a.5.5 0 0 1-.5-.5zm0-3a.5.5 0 0 1 .5-.5h11a.5.5 0 0 1 0 1h-11a.5.5 0 0 1-.5-.5zm0-3a.5.5 0 0 1 .5-.5h11a.5.5 0 0 1 0 1h-11a.5.5 0 0 1-.5-.5z"
      />
    </svg>
  );
}

type BlockType = 'paragraph' | 'h1' | 'h2' | 'h3' | 'h4' | 'h5';

const blockTypeToBlockName: Record<BlockType, string> = {
  h1: 'Heading 1',
  h2: 'Heading 2',
  h3: 'Heading 3',
  h4: 'Heading 4',
  h5: 'Heading 5',
  paragraph: 'Paragraph',
};

/**
 * Props interface for the ToolbarPlugin component.
 *
 * The toolbar is designed with a responsive multi-row layout:
 * - Primary row: History controls, text formatting (bold, italic, underline), typography dropdown, and overflow menu
 * - Secondary row: Alignment controls (left, center, right, justify)
 * - Overflow menu: Less frequently used features like link insertion
 *
 * This layout ensures optimal usability in narrow side panels while maintaining access to all features.
 */
export interface ToolbarPluginProps extends HTMLAttributes<HTMLDivElement> {
  history?: boolean;
  bold?: boolean;
  italic?: boolean;
  underline?: boolean;
  alignment?: boolean;
  typography?: boolean;
  link?: boolean;
  disabled?: boolean;
}

/**
 * ToolbarPlugin component for rich text editor toolbar.
 */
function ToolbarPlugin({
  history = true,
  bold = true,
  italic = true,
  underline = true,
  alignment = true,
  typography = true,
  link = true,
  disabled = false,
  className,
}: ToolbarPluginProps): ReactElement {
  // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment, @typescript-eslint/no-unsafe-call
  const [editor] = useLexicalComposerContext();

  const [canUndo, setCanUndo] = useState(false);
  const [canRedo, setCanRedo] = useState(false);
  const [isBold, setIsBold] = useState(false);
  const [isItalic, setIsItalic] = useState(false);
  const [isUnderline, setIsUnderline] = useState(false);
  const [blockType, setBlockType] = useState<BlockType>('paragraph');
  const [isLink, setIsLink] = useState(false);
  const [typographyMenu, setTypographyMenu] = useState<null | HTMLElement>(null);
  const [selectedAlignment, setSelectedAlignment] = useState<number>(1);

  const openTypographyMenu = Boolean(typographyMenu);

  const handleTypographyMenuOpen = (event: React.MouseEvent<HTMLElement>): void => {
    setTypographyMenu(event.currentTarget);
  };

  const handleTypographyMenuClose = (): void => {
    setTypographyMenu(null);
  };

  /**
   * Formats the selected text to a paragraph block type.
   * If the current block type is not a paragraph, it changes it to a paragraph.
   * Closes the typography menu after formatting.
   */
  const formatParagraph = (): void => {
    editor.update(() => {
      const selection = $getSelection();

      if ($isRangeSelection(selection) && blockType !== 'paragraph') {
        $setBlocksType(selection, () => $createParagraphNode());
      }
    });
    handleTypographyMenuClose();
  };

  /**
   * Formats the selected text to a heading block type.
   * If the current block type is not the specified heading size, it changes it to the specified heading size.
   * Closes the typography menu after formatting.
   *
   * @param headingSize - The heading size to format the selected text to.
   */
  const formatHeading = (headingSize: 'h1' | 'h2' | 'h3' | 'h4' | 'h5'): void => {
    if (blockType !== headingSize) {
      editor.update(() => {
        const selection: BaseSelection | null = $getSelection();

        $setBlocksType(selection, () => $createHeadingNode(headingSize));
      });
    }
    handleTypographyMenuClose();
  };

  /**
   * Inserts or removes a link in the selected text.
   */
  const insertLink: () => void = useCallback(() => {
    if (!isLink) {
      editor.dispatchCommand(TOGGLE_LINK_COMMAND, 'https://');
    } else {
      editor.dispatchCommand(TOGGLE_LINK_COMMAND, null);
    }
  }, [editor, isLink]);

  /**
   * Updates the toolbar state based on the current selection.
   */
  const $updateToolbar: () => void = useCallback(() => {
    const selection: BaseSelection | null = $getSelection();

    if ($isRangeSelection(selection)) {
      // Update text format.
      setIsBold(selection.hasFormat('bold'));
      setIsItalic(selection.hasFormat('italic'));
      setIsUnderline(selection.hasFormat('underline'));

      // Update link.
      const node: TextNode | ElementNode = getSelectedNode(selection);
      const parent: ElementNode | null = node.getParent();

      if ($isLinkNode(parent) || $isLinkNode(node)) {
        setIsLink(true);
      } else {
        setIsLink(false);
      }

      const anchorNode: TextNode | ElementNode = selection.anchor.getNode();
      const element: TextNode | ElementNode =
        anchorNode.getKey() === 'root' ? anchorNode : anchorNode.getTopLevelElementOrThrow();

      // Update text alignment.
      setSelectedAlignment(element.getFormat());

      const elementKey: NodeKey = element.getKey();
      const elementDOM: HTMLElement | null = editor.getElementByKey(elementKey);

      // Update block type
      if (elementDOM !== null) {
        if ($isHeadingNode(element)) {
          const tag: HeadingTagType = element.getTag();

          setBlockType(tag as BlockType);
        } else {
          const type: string = element.getType();

          if (type in blockTypeToBlockName) {
            setBlockType(type as BlockType);
          } else {
            setBlockType('paragraph');
          }
        }
      }
    }
  }, [editor]);

  useEffect(
    () =>
      mergeRegister(
        editor.registerUpdateListener(({editorState}: {editorState: EditorState}) => {
          editorState.read(() => {
            $updateToolbar();
          });
        }),
        editor.registerCommand(
          SELECTION_CHANGE_COMMAND,
          () => {
            $updateToolbar();

            return false;
          },
          LowPriority,
        ),
        editor.registerCommand(
          CAN_UNDO_COMMAND,
          (payload: boolean) => {
            setCanUndo(payload);

            return false;
          },
          LowPriority,
        ),
        editor.registerCommand(
          CAN_REDO_COMMAND,
          (payload: boolean) => {
            setCanRedo(payload);

            return false;
          },
          LowPriority,
        ),
      ),
    [editor, $updateToolbar],
  );

  return (
    <Paper className={classNames('OxygenRichTextToolbar-root', className)} variant="outlined" sx={{boxShadow: 'none'}}>
      <Stack direction="row" alignItems="center" justifyContent="space-evenly" className="toolbar-row primary-row">
        {history && (
          <>
            <IconButton
              disabled={!canUndo || disabled}
              onClick={() => {
                editor.dispatchCommand(UNDO_COMMAND, undefined);
              }}
              aria-label="Undo"
            >
              <ArrowCounterClockwiseIcon height={16} width={16} />
            </IconButton>
            <IconButton
              disabled={!canRedo || disabled}
              onClick={() => {
                editor.dispatchCommand(REDO_COMMAND, undefined);
              }}
              aria-label="Redo"
            >
              <ArrowClockwiseIcon height={16} width={16} />
            </IconButton>
          </>
        )}
        <Divider orientation="vertical" flexItem />
        {typography && (
          <>
            <Button
              disableRipple
              variant="text"
              color="secondary"
              disabled={disabled}
              onClick={handleTypographyMenuOpen}
              endIcon={<ChevronDownIcon />}
            >
              {blockTypeToBlockName[blockType]}
            </Button>
            <Menu
              open={openTypographyMenu}
              anchorEl={typographyMenu}
              onClose={handleTypographyMenuClose}
              anchorOrigin={{horizontal: 'left', vertical: 'bottom'}}
              transformOrigin={{horizontal: 'left', vertical: 'top'}}
            >
              <MenuItem onClick={() => formatHeading('h1')}>
                <ListItemText primary={<h1>{blockTypeToBlockName.h1}</h1>} />
              </MenuItem>
              <MenuItem onClick={() => formatHeading('h2')}>
                <ListItemText primary={<h2>{blockTypeToBlockName.h2}</h2>} />
              </MenuItem>
              <MenuItem onClick={() => formatHeading('h3')}>
                <ListItemText primary={<h3>{blockTypeToBlockName.h3}</h3>} />
              </MenuItem>
              <MenuItem onClick={() => formatHeading('h4')}>
                <ListItemText primary={<h4>{blockTypeToBlockName.h4}</h4>} />
              </MenuItem>
              <MenuItem onClick={() => formatHeading('h5')}>
                <ListItemText primary={<h5>{blockTypeToBlockName.h5}</h5>} />
              </MenuItem>
              <MenuItem onClick={formatParagraph}>
                <ListItemText primary={blockTypeToBlockName.paragraph} />
              </MenuItem>
            </Menu>
          </>
        )}
      </Stack>
      <Stack direction="row" alignItems="center" justifyContent="space-evenly" className="toolbar-row secondary-row">
        {bold && (
          <IconButton
            disabled={disabled}
            onClick={() => {
              editor.dispatchCommand(FORMAT_TEXT_COMMAND, 'bold');
            }}
            className={classNames({active: isBold && !disabled})}
            aria-label="Format Bold"
          >
            <BoldIcon height={16} width={16} />
          </IconButton>
        )}
        {italic && (
          <IconButton
            disabled={disabled}
            onClick={() => {
              editor.dispatchCommand(FORMAT_TEXT_COMMAND, 'italic');
            }}
            className={classNames({active: isItalic && !disabled})}
            aria-label="Format Italics"
          >
            <ItalicIcon height={16} width={16} />
          </IconButton>
        )}
        {underline && (
          <IconButton
            disabled={disabled}
            onClick={() => {
              editor.dispatchCommand(FORMAT_TEXT_COMMAND, 'underline');
            }}
            className={classNames({active: isUnderline && !disabled})}
            aria-label="Format Underline"
          >
            <UnderlineIcon height={16} width={16} />
          </IconButton>
        )}
        {link && (
          <IconButton
            disabled={disabled}
            onClick={() => insertLink()}
            className={classNames({active: isLink && !disabled})}
            aria-label="Format Link"
          >
            <LinkIcon size={14} />
          </IconButton>
        )}
        {alignment && (
          <>
            <IconButton
              disabled={disabled}
              onClick={() => {
                editor.dispatchCommand(FORMAT_ELEMENT_COMMAND, 'left');
              }}
              className={classNames({active: selectedAlignment === 1 && !disabled})}
              aria-label="Left Align"
            >
              <LeftAlignIcon height={16} width={16} />
            </IconButton>
            <IconButton
              disabled={disabled}
              onClick={() => {
                editor.dispatchCommand(FORMAT_ELEMENT_COMMAND, 'center');
              }}
              className={classNames({active: selectedAlignment === 2 && !disabled})}
              aria-label="Center Align"
            >
              <CenterAlignIcon height={16} width={16} />
            </IconButton>
            <IconButton
              disabled={disabled}
              onClick={() => {
                editor.dispatchCommand(FORMAT_ELEMENT_COMMAND, 'right');
              }}
              className={classNames({active: selectedAlignment === 3 && !disabled})}
              aria-label="Right Align"
            >
              <RightAlignIcon height={16} width={16} />
            </IconButton>
            <IconButton
              disabled={disabled}
              onClick={() => {
                editor.dispatchCommand(FORMAT_ELEMENT_COMMAND, 'justify');
              }}
              className={classNames({active: selectedAlignment === 4 && !disabled})}
              aria-label="Justify Align"
            >
              <JustifyAlignIcon height={16} width={16} />
            </IconButton>
          </>
        )}
      </Stack>
    </Paper>
  );
}

export default ToolbarPlugin;
