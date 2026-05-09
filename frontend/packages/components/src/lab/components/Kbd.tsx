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

import type {PropsWithChildren} from 'react';

/**
 * Props for the {@link Kbd} component.
 */
type KbdProps = PropsWithChildren;

/**
 * Renders keyboard key labels styled as a `<kbd>` element.
 *
 * @param props - Component props
 * @returns A styled keyboard key element
 */
export default function Kbd({children}: KbdProps) {
  return (
    <kbd
      style={{
        display: 'inline-block',
        padding: '1px 5px',
        fontSize: '0.7rem',
        fontFamily: 'inherit',
        lineHeight: '1.4',
        color: 'inherit',
        backgroundColor: 'rgba(0, 0, 0, 0.06)',
        border: '1px solid rgba(0, 0, 0, 0.18)',
        borderRadius: '4px',
        boxShadow: 'inset 0 -1px 0 rgba(0, 0, 0, 0.12)',
        whiteSpace: 'nowrap',
      }}
    >
      {children}
    </kbd>
  );
}
