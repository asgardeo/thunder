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

import {type PropsWithChildren, type ReactElement} from 'react';
import {useTranslation} from 'react-i18next';
import {isI18nPattern as checkIsI18nPattern, isMetaPattern, resolveI18nValue} from '@/features/flows/utils/i18nPatternUtils';
import DynamicValueSyntax from '../../../resource-property-panel/DynamicValueSyntax';
import './PlaceholderComponent.scss';

/**
 * Props interface of {@link PlaceholderComponent}
 */
export interface PlaceholderComponentProps {
  value: string;
}

/**
 * Placeholder component for displaying a placeholder text.
 * If the value matches the i18n pattern {{t(key)}}, it resolves and displays the translated value.
 * Otherwise, it displays the raw text content.
 *
 * @param props - Props injected to the component.
 * @returns The PlaceholderComponent component.
 */
function PlaceholderComponent({value, children = null}: PropsWithChildren<PlaceholderComponentProps>): ReactElement {
  // Use bindI18n to ensure component re-renders when translations are added/changed
  const {t} = useTranslation('flowI18n', {bindI18n: 'languageChanged loaded added'});

  /**
   * Check if the value matches the i18n pattern and resolve it if so.
   * Computed on every render to ensure translations are always up-to-date.
   */
  const isI18nPattern = checkIsI18nPattern(value);
  const displayValue = isI18nPattern ? resolveI18nValue(value, t) : value;

  if (isI18nPattern) {
    return <span>{displayValue}</span>;
  }

  if (isMetaPattern(value)) {
    return <DynamicValueSyntax value={value} />;
  }

  if (children) {
    return children as ReactElement;
  }

  return <span>{value}</span>;
}

export default PlaceholderComponent;
