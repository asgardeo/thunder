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

import {ParagraphNode, TextNode, type EditorThemeClasses} from 'lexical';
import {type InitialConfigType, LexicalComposer} from '@lexical/react/LexicalComposer';
import {HeadingNode} from '@lexical/rich-text';
import {AutoLinkNode, LinkNode} from '@lexical/link';
import type {Resource} from '@/features/flows/models/resources';
import {useMemo, type ReactElement} from 'react';
import {useTranslation} from 'react-i18next';
import {Paper} from '@wso2/oxygen-ui';
import classNames from 'classnames';
import {HistoryPlugin} from '@lexical/react/LexicalHistoryPlugin';
import {LinkPlugin} from '@lexical/react/LexicalLinkPlugin';
import {RichTextPlugin} from '@lexical/react/LexicalRichTextPlugin';
import {AutoFocusPlugin} from '@lexical/react/LexicalAutoFocusPlugin';
import {ContentEditable} from '@lexical/react/LexicalContentEditable';
import {LexicalErrorBoundary} from '@lexical/react/LexicalErrorBoundary';
import ToolbarPlugin from './helper-plugins/ToolbarPlugin';
import type {ToolbarPluginProps} from './helper-plugins/ToolbarPlugin';
import CustomLinkPlugin from './helper-plugins/CustomLinkPlugin';
import './RichText.scss';
import HTMLPlugin from './helper-plugins/HTMLPlugin';

/**
 * Theme classes for the rich text editor.
 */
const ThemeClasses: EditorThemeClasses = {
  heading: {
    h1: 'rich-text-heading-h1',
    h2: 'rich-text-heading-h2',
    h3: 'rich-text-heading-h3',
    h4: 'rich-text-heading-h4',
    h5: 'rich-text-heading-h5',
    h6: 'rich-text-heading-h6',
  },
  link: 'rich-text-link',
  paragraph: 'rich-text-paragraph',
  text: {
    bold: 'rich-text-bold',
    italic: 'rich-text-italic',
    underline: 'rich-text-underline',
  },
};

/**
 * Configs for the rich text editor.
 */
const editorConfig: InitialConfigType = {
  namespace: 'Rich Text',
  nodes: [ParagraphNode, TextNode, HeadingNode, LinkNode, AutoLinkNode],
  onError(error: Error) {
    throw error;
  },
  theme: ThemeClasses,
};

/**
 * Props interface for the RichText component.
 */
export interface RichTextProps {
  /**
   * Options to customize the rich text editor toolbar.
   */
  ToolbarProps?: ToolbarPluginProps;
  /**
   * Listener for changes in the rich text editor content.
   *
   * @param value - The HTML string representation of the rich text editor content.
   */
  onChange: (value: string) => void;
  /**
   * Additional CSS class names to apply to the rich text editor container.
   */
  className?: string;
  /**
   * The resource associated with the rich text editor.
   */
  resource: Resource;
  /**
   * Whether the rich text editor is disabled. If true, the editor will not be editable.
   */
  disabled?: boolean;
  /**
   * Whether the rich text editor has an error state.
   */
  hasError?: boolean;
}

/**
 * Rich text editor component.
 */
function RichText({
  ToolbarProps = {},
  className = '',
  onChange,
  resource,
  disabled = false,
  hasError = false,
}: RichTextProps): ReactElement {
  const {t} = useTranslation();

  const labelValue: string | undefined = useMemo(
    () => (resource as Resource & {label?: string})?.label,
    [resource],
  );

  /**
   * Check if the resource label matches the i18n pattern.
   */
  const isI18nPattern: boolean = useMemo(() => {
    if (!labelValue) return false;

    const i18nPattern = /^\{\{[^}]+\}\}$/;

    return i18nPattern.test(labelValue.trim());
  }, [labelValue]);

  return (
    <LexicalComposer initialConfig={editorConfig}>
      <div className={classNames('OxygenRichText-root', className)}>
        <ToolbarPlugin {...ToolbarProps} disabled={isI18nPattern || disabled} />
        <Paper className={classNames('OxygenRichText-editor-root', {error: hasError})} elevation={0} variant="outlined">
          {isI18nPattern ? (
            <div className="OxygenRichText-i18n-placeholder">
              <div className="OxygenRichText-i18n-placeholder-key">{labelValue}</div>
            </div>
          ) : (
            <>
              <RichTextPlugin
                contentEditable={
                  <ContentEditable
                    className="OxygenRichText-editor-input"
                    aria-placeholder="Enter some rich text..."
                    placeholder={
                      <div className="OxygenRichText-editor-input-placeholder">
                        {t('flows:core.elements.richText.placeholder')}
                      </div>
                    }
                  />
                }
                ErrorBoundary={LexicalErrorBoundary}
              />
              <HistoryPlugin />
              <AutoFocusPlugin />
              <LinkPlugin />
              <CustomLinkPlugin />
              <HTMLPlugin resource={resource} onChange={onChange} disabled={disabled} />
            </>
          )}
        </Paper>
      </div>
    </LexicalComposer>
  );
}

export default RichText;
