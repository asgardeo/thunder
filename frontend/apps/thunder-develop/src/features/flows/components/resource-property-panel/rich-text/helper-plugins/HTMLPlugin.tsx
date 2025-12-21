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

import {$generateHtmlFromNodes, $generateNodesFromDOM} from '@lexical/html';
import {useLexicalComposerContext} from '@lexical/react/LexicalComposerContext';
import {$getRoot, $insertNodes, type EditorState, type LexicalNode, RootNode} from 'lexical';
import {type ReactElement, useCallback, useEffect, useRef} from 'react';
import {UPDATE_TYPES, type UpdateType} from '@/features/flows/models/rich-text';
import type {Resource} from '../../../../models/resources';

/**
 * Props interface for the HTML plugin.
 */
interface HTMLPluginProps {
  /**
   * Listener for changes in the editor state.
   */
  onChange: (value: string) => void;
  /**
   * The resource associated with the rich text editor.
   */
  resource: Resource;
  /**
   * Whether the editor is disabled.
   */
  disabled?: boolean;
}

const PRE_WRAP_STYLE_WITH_CLASS = '" style="white-space: pre-wrap;"';
const PRE_WRAP_STYLE = 'style="white-space: pre-wrap;"';
const TEXT_ALIGN_TYPES: string[] = ['left', 'right', 'center', 'justify'];
const TEXT_ALIGN_PLACEHOLDER = '{{textAlign}}';
const TEXT_ALIGN_STYLE_WITH_CLASS = `" style="text-align: ${TEXT_ALIGN_PLACEHOLDER};"`;
const TEXT_ALIGN_STYLE = `style="text-align: ${TEXT_ALIGN_PLACEHOLDER};"`;
const DIR_LTR_CLASS = '" dir="ltr"';
const DIR_LTR = 'dir="ltr"';
const CLASS_NAME_PLACEHOLDER = '{{className}}';
const ADDITIONAL_CLASSES = `class="${CLASS_NAME_PLACEHOLDER}"`;
const EMPTY_CONTENT = '<p class="rich-text-paragraph"><br></p>';

/**
 * Convert nodes tree to HTML string.
 */
function HTMLPlugin({onChange, resource, disabled = false}: HTMLPluginProps): ReactElement | null {
  // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment, @typescript-eslint/no-unsafe-call
  const [editor] = useLexicalComposerContext();
  const updateType = useRef<UpdateType>(UPDATE_TYPES.NONE);

  /**
   * Pre-process the HTML string to add additional classes and styles.
   *
   * @param html - The HTML string to pre-process.
   * @returns The pre-processed HTML string.
   */
  const preProcessHTML = useCallback((html: string): string => {
    let processedHtml = html.replaceAll(DIR_LTR_CLASS, '"');
    processedHtml = processedHtml.replaceAll(DIR_LTR, '');

    processedHtml = processedHtml.replaceAll(PRE_WRAP_STYLE_WITH_CLASS, ' rich-text-pre-wrap"');
    processedHtml = processedHtml.replaceAll(
      PRE_WRAP_STYLE,
      ADDITIONAL_CLASSES.replace(CLASS_NAME_PLACEHOLDER, 'rich-text-pre-wrap'),
    );

    TEXT_ALIGN_TYPES.forEach((textAlign) => {
      processedHtml = processedHtml.replaceAll(
        TEXT_ALIGN_STYLE_WITH_CLASS.replace(TEXT_ALIGN_PLACEHOLDER, textAlign),
        ` rich-text-align-${textAlign}"`,
      );
      processedHtml = processedHtml.replaceAll(
        TEXT_ALIGN_STYLE.replace(TEXT_ALIGN_PLACEHOLDER, textAlign),
        ADDITIONAL_CLASSES.replace(CLASS_NAME_PLACEHOLDER, `rich-text-align-${textAlign}`),
      );
    });

    return processedHtml;
  }, []);

  /**
   * Post-process the HTML string to reverse the transformations done by preProcessHTML.
   * This method converts processed HTML back to its original format.
   *
   * @param html - The processed HTML string to reverse.
   * @returns The original HTML string.
   */
  const postProcessHTML = (html: string): string => {
    let processedHtml = html;
    // Reverse text alignment class replacements.
    TEXT_ALIGN_TYPES.forEach((textAlign) => {
      processedHtml = processedHtml.replaceAll(
        ` rich-text-align-${textAlign}"`,
        TEXT_ALIGN_STYLE_WITH_CLASS.replace(TEXT_ALIGN_PLACEHOLDER, textAlign),
      );
      processedHtml = processedHtml.replaceAll(
        ADDITIONAL_CLASSES.replace(CLASS_NAME_PLACEHOLDER, `rich-text-align-${textAlign}`),
        TEXT_ALIGN_STYLE.replace(TEXT_ALIGN_PLACEHOLDER, textAlign),
      );
    });

    // Reverse pre-wrap style replacements.
    processedHtml = processedHtml.replaceAll(' rich-text-pre-wrap"', PRE_WRAP_STYLE_WITH_CLASS);
    processedHtml = processedHtml.replaceAll(
      ADDITIONAL_CLASSES.replace(CLASS_NAME_PLACEHOLDER, 'rich-text-pre-wrap'),
      PRE_WRAP_STYLE,
    );

    return processedHtml;
  };

  useEffect(() => {
    if (!editor || !resource) {
      return;
    }

    if (updateType.current === UPDATE_TYPES.INTERNAL) {
      updateType.current = UPDATE_TYPES.NONE;

      return;
    }

    const parser: DOMParser = new DOMParser();
    const labelValue = (resource as Resource & {label?: string})?.label ?? '';
    const dom: Document = parser.parseFromString(postProcessHTML(labelValue), 'text/html');

    editor.update(() => {
      updateType.current = UPDATE_TYPES.EXTERNAL;

      const root: RootNode = $getRoot();

      root.clear(); // clear existing content if needed.

      const nodes: LexicalNode[] = $generateNodesFromDOM(editor, dom);

      $insertNodes(nodes); // insert new nodes into the editor.
    });
  }, [editor, resource]);

  /**
   * Register the update listener to process the editor state changes.
   */
  useEffect(() => {
    if (!editor || !onChange) {
      return undefined;
    }

    return editor.registerUpdateListener(({editorState}: {editorState: EditorState}) => {
      if (updateType.current === UPDATE_TYPES.EXTERNAL) {
        updateType.current = UPDATE_TYPES.NONE;

        return;
      }

      editorState.read(() => {
        updateType.current = UPDATE_TYPES.INTERNAL;

        const htmlString: string = $generateHtmlFromNodes(editor);

        const processedHTML: string = preProcessHTML(htmlString);

        onChange(processedHTML === EMPTY_CONTENT ? '' : processedHTML);
      });
    });
  }, [editor, onChange, preProcessHTML]);

  /**
   * Handle the editor's disabled state.
   */
  useEffect(() => {
    if (disabled) {
      editor.setEditable(false);
    } else if (!editor.isEditable()) {
      editor.setEditable(true);
    }
  }, [disabled, editor]);

  return null;
}

export default HTMLPlugin;
