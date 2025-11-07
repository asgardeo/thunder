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

import React, {type CSSProperties, type HTMLAttributes, type ReactElement, type Ref} from 'react';

/**
 * Props interface for the Action component.
 */
export interface ActionProps extends HTMLAttributes<HTMLButtonElement> {
  /**
   * Variant of the action button.
   */
  variant?: 'light' | 'dark' | 'destructive';
  /**
   * Cursor style for the action button.
   */
  cursor?: CSSProperties['cursor'];
}

/**
 * Action component.
 *
 * @param props - Props injected to the component.
 * @returns The Action component.
 */
function Action(
  {className, cursor = 'pointer', style, variant = 'light', ...rest}: ActionProps,
  ref: Ref<HTMLButtonElement>,
): ReactElement {
  const classes = ['flow-builder-dnd-action', variant && `flow-builder-dnd-action--${variant}`, className]
    .filter(Boolean)
    .join(' ');

  return (
    <button
      ref={ref}
      type="button"
      className={classes}
      style={
        {
          ...style,
          cursor,
          backgroundColor: 'transparent',
          border: 'none',
          transition: 'background-color 0.2s ease',
          height: '100%',
          width: '50px',
        } as CSSProperties
      }
      onMouseEnter={(e) => {
        e.currentTarget.style.backgroundColor = 'rgba(0, 0, 0, 0.15)';
      }}
      onMouseLeave={(e) => {
        e.currentTarget.style.backgroundColor = 'transparent';
      }}
      {...rest}
    />
  );
}

export default React.forwardRef(Action);
