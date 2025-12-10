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
import {memo, useMemo, type ReactElement} from 'react';
import {Trans} from 'react-i18next';
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

// PERFORMANCE: Define fields outside component to prevent recreation on every render
const RICHTEXT_VALIDATION_FIELDS: RequiredFieldInterface[] = [
  {
    errorMessage: 'Text is required',
    name: 'text',
  },
];

/**
 * Configuration interface for RichText element.
 */
interface RichTextConfig {
  text?: string;
}

/**
 * RichText element type.
 */
export type RichTextElement = FlowElement<RichTextConfig>;

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
 * PERFORMANCE: This component has been optimized to:
 * 1. Use static validation fields defined outside the component
 * 2. Remove useTranslation hook to avoid re-renders
 * 3. Memoize the general message and sanitized HTML
 *
 * @param props - Props injected to the component.
 * @returns The RichTextAdapter component.
 */
function RichTextAdapter({resource}: RichTextAdapterPropsInterface): ReactElement {
  // PERFORMANCE: Memoize general message - only depends on resource.id
  const generalMessage: ReactElement = useMemo(
    () => (
      <Trans i18nKey="flows:core.validation.fields.richText.general" values={{id: resource.id}}>
        Required fields are not properly configured for the rich text with ID <code>{resource.id}</code>.
      </Trans>
    ),
    [resource.id],
  );

  // PERFORMANCE: Use static fields array defined outside component
  useRequiredFields(resource, generalMessage, RICHTEXT_VALIDATION_FIELDS);

  // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment -- Config type is validated at runtime
  const richTextConfig = resource.config as RichTextConfig | undefined;

  // PERFORMANCE: Memoize sanitized HTML to avoid re-computation
  const sanitizedHtml: string = useMemo(
    () =>
      DOMPurify.sanitize(richTextConfig?.text ?? '', {
        ADD_ATTR: ['target'],
        RETURN_TRUSTED_TYPE: false,
      }),
    [richTextConfig?.text],
  );

  return (
    <div className="rich-text-content">
      {/* eslint-disable-next-line @typescript-eslint/no-unsafe-call -- html-react-parser types issue */}
      <PlaceholderComponent value={richTextConfig?.text ?? ''}>{parse(sanitizedHtml)}</PlaceholderComponent>
    </div>
  );
}

// PERFORMANCE: Memoize to prevent re-renders during drag operations
export default memo(RichTextAdapter, (prevProps, nextProps) =>
  prevProps.resource === nextProps.resource
);
