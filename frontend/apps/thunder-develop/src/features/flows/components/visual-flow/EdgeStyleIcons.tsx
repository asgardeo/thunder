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

import {type ReactElement} from 'react';

/**
 * Icon component for Bezier edge style
 */
export function BezierEdgeIcon(): ReactElement {
  return (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
      <path d="M4 12 C 8 4, 16 20, 20 12" strokeLinecap="round" />
    </svg>
  );
}

/**
 * Icon component for Smooth Step edge style
 */
export function SmoothStepEdgeIcon(): ReactElement {
  return (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
      <path d="M4 6 H 10 Q 12 6, 12 8 V 16 Q 12 18, 14 18 H 20" strokeLinecap="round" />
    </svg>
  );
}

/**
 * Icon component for Step edge style
 */
export function StepEdgeIcon(): ReactElement {
  return (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
      <path d="M4 6 H 12 V 18 H 20" strokeLinecap="round" strokeLinejoin="miter" />
    </svg>
  );
}
