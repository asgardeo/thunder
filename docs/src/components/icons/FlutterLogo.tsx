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

import React from 'react';

export default function FlutterLogo({size = 64}: {size?: number}) {
  return (
    <svg width={size} height={size * 1.238} viewBox="0 0 300 371.43">
      <defs>
        <linearGradient
          id="flutter-gradient"
          x1="6254.1"
          y1="5576.56"
          x2="6424.34"
          y2="5406.31"
          gradientTransform="translate(-1404 -1054.53) scale(0.25)"
          gradientUnits="userSpaceOnUse"
        >
          <stop offset="0" stopColor="#1a237e" stopOpacity="0.4" />
          <stop offset="1" stopColor="#1a237e" stopOpacity="0" />
        </linearGradient>
      </defs>
      <path
        fill="#54c5f8"
        d="M300 171.43l-100 100 100 100H185.72l-42.86-42.86h0L85.71 271.42l100-100ZM185.72 0 0 185.72l57.15 57.15L300 0Z"
      />
      <path fill="#01579b" d="M142.85 328.57 185.72 371.44H300l-100-100z" />
      <path fill="url(#flutter-gradient)" d="M142.85 328.57 227.61 299.24 200.01 271.44z" />
      <path fill="#29b6f6" d="M102.45 231.03h80.81l-40.41 40.4z" />
    </svg>
  );
}
