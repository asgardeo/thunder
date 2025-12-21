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
import {useMemo, type ReactElement} from 'react';
import {Trans, useTranslation} from 'react-i18next';
import DOMPurify from 'dompurify';
import parse from 'html-react-parser';
import type {RequiredFieldInterface} from '@/features/flows/hooks/useRequiredFields';
import useRequiredFields from '@/features/flows/hooks/useRequiredFields';
import type {Element as FlowElement} from '@/features/flows/models/elements';
import PlaceholderComponent from './PlaceholderComponent';
import './RichTextAdapter.scss';

// Register DOMPurify hook once at module level to handle anchor tags.
// eslint-disable-next-line @typescript-eslint/no-unsafe-call, @typescript-eslint/no-unsafe-member-access -- DOMPurify types issue
(DOMPurify as unknown as {addHook: (name: string, fn: (node: globalThis.Element) => void) => void}).addHook(
  'afterSanitizeAttributes',
  (node: globalThis.Element) => {
    if (node.hasAttribute('target')) {
      const target: string | null = node.getAttribute('target');

      if (target === '_blank') {
        node.setAttribute('rel', 'noopener noreferrer');
      }
    }
  },
);

const RICHTEXT_VALIDATION_FIELD_NAMES = {
  label: 'label',
} as const;

/**
 * RichText element type.
 */
export type RichTextElement = FlowElement & {
  label?: string;
};

/**
 * Props interface of {@link RichTextAdapter}
 */
export interface RichTextAdapterPropsInterface {
  /**
   * The rich text element properties.
   */
  resource: FlowElement;
}

/**
 * Adapter for the Rich Text component.
 *
 * @param props - Props injected to the component.
 * @returns The RichTextAdapter component.
 */
function RichTextAdapter({resource}: RichTextAdapterPropsInterface): ReactElement {
  const {t} = useTranslation();

  const generalMessage: ReactElement = useMemo(
    () => (
      <Trans i18nKey="flows:core.validation.fields.richText.general" values={{id: resource.id}}>
        Required fields are not properly configured for the rich text with ID <code>{resource.id}</code>.
      </Trans>
    ),
    [resource.id],
  );

  const validationFields: RequiredFieldInterface[] = useMemo(
    () => [
      {
        errorMessage: t('flows:core.validation.fields.richText.label'),
        name: RICHTEXT_VALIDATION_FIELD_NAMES.label,
      },
    ],
    [t],
  );

  useRequiredFields(resource, generalMessage, validationFields);

  const richTextElement = resource as RichTextElement;
  const textContent = richTextElement?.label ?? '';

  const sanitizedHtml: string = useMemo(
    () =>
      DOMPurify.sanitize(textContent, {
        ADD_ATTR: ['target'],
        RETURN_TRUSTED_TYPE: false,
      }),
    [textContent],
  );

  return (
    <div className="rich-text-content">
      {/* eslint-disable-next-line @typescript-eslint/no-unsafe-call -- html-react-parser types issue */}
      <PlaceholderComponent value={textContent}>{parse(sanitizedHtml)}</PlaceholderComponent>
    </div>
  );
}

export default RichTextAdapter;
