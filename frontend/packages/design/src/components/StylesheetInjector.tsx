/**
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
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

import {useConfig} from '@thunderid/contexts';
import {createLogger} from '@thunderid/logger';
import {useEffect} from 'react';
import useDesign from '../contexts/Design/useDesign';
import type {Stylesheet} from '../models/layout';
import {sanitizeCss, isValidStylesheetUrl} from '../utils/cssSanitizer';

const logger = createLogger({component: 'StylesheetInjector'});

/**
 * Props for the StylesheetInjector component.
 */
export interface StylesheetInjectorProps {
  /** Optional stylesheet override; if omitted, reads from layout.head.stylesheets via useDesign() */
  stylesheets?: Stylesheet[];
}

/**
 * Component that injects stylesheets from the layout head configuration into the document head.
 *
 * Supports two stylesheet types:
 * - inline: Injects a style element with sanitized CSS content
 * - url: Injects a link rel="stylesheet" element (https only)
 *
 * Stylesheets are identified by their id field, prefixed with "<PRODUCT_NAME>-stylesheet-"
 * to avoid DOM ID collisions. Elements are cleaned up on unmount or when the stylesheet
 * list changes.
 */
export default function StylesheetInjector({stylesheets = undefined}: StylesheetInjectorProps): null {
  const {config} = useConfig();
  const {layout} = useDesign();
  const resolvedStylesheets = stylesheets ?? layout?.head?.stylesheets ?? [];
  const idPrefix = config.brand.product_name.toLowerCase().replace(/\s+/g, '-');
  const elementIdPrefix = `${idPrefix}-stylesheet-`;
  const dataAttr = `data-${idPrefix}-custom`;

  // Use a serialized key to avoid re-running the effect when the array reference
  // changes but the content is identical (common with JSON deserialization).
  const serialized = JSON.stringify(resolvedStylesheets);

  useEffect(() => {
    const parsed: Stylesheet[] = JSON.parse(serialized) as Stylesheet[];
    const injectedIds: string[] = [];

    parsed.forEach((stylesheet) => {
      const elementId = `${elementIdPrefix}${stylesheet.id}`;

      // Remove existing element with same ID to handle updates
      document.getElementById(elementId)?.remove();

      if (stylesheet.type === 'inline') {
        const style = document.createElement('style');
        style.id = elementId;
        style.setAttribute(dataAttr, 'true');
        style.textContent = sanitizeCss(stylesheet.content);
        document.head.appendChild(style);
        injectedIds.push(elementId);
      } else if (stylesheet.type === 'url') {
        if (isValidStylesheetUrl(stylesheet.href)) {
          const link = document.createElement('link');
          link.id = elementId;
          link.rel = 'stylesheet';
          link.href = stylesheet.href;
          link.setAttribute(dataAttr, 'true');
          document.head.appendChild(link);
          injectedIds.push(elementId);
        } else {
          logger.warn(`[StylesheetInjector] Skipping stylesheet "${stylesheet.id}": URL must use https protocol`);
        }
      }
    });

    return () => {
      injectedIds.forEach((id) => document.getElementById(id)?.remove());
    };
  }, [dataAttr, elementIdPrefix, serialized]);

  return null;
}
