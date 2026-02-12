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

export default function BrowserLogo({size = 64}: {size?: number}) {
  return (
    <svg width={size} height={size} viewBox="0 0 63 63" fill="none">
      <path
        fill="#fff"
        d="M31.499 47.247c8.698 0 15.75-7.052 15.75-15.75s-7.052-15.75-15.75-15.75-15.75 7.051-15.75 15.75 7.051 15.75 15.75 15.75z"
      />
      <path
        fill="url(#browser-gradient-1)"
        d="M17.86 39.375 4.22 15.754a31.492 31.492 0 0 0 27.28 47.242l13.638-23.62v-.005a15.746 15.746 0 0 1-27.28.004z"
      />
      <path
        fill="url(#browser-gradient-2)"
        d="M45.138 39.374 31.5 62.995A31.491 31.491 0 0 0 58.773 15.75H31.498l-.003.002a15.747 15.747 0 0 1 13.643 23.623z"
      />
      <path
        fill="#1a73e8"
        d="M31.499 43.969c6.886 0 12.469-5.583 12.469-12.469s-5.583-12.469-12.469-12.469S19.03 24.614 19.03 31.5 24.613 43.969 31.5 43.969z"
      />
      <path
        fill="url(#browser-gradient-3)"
        d="M31.5 15.75h27.274a31.493 31.493 0 0 0-54.553.004L17.86 39.375l.004.002A15.747 15.747 0 0 1 31.499 15.75z"
      />
      <defs>
        <linearGradient
          id="browser-gradient-1"
          x1="34.909"
          x2="7.632"
          y1="61.029"
          y2="13.785"
          gradientUnits="userSpaceOnUse"
        >
          <stop offset="0" stopColor="#1e8e3e" />
          <stop offset="1" stopColor="#34a853" />
        </linearGradient>
        <linearGradient
          id="browser-gradient-2"
          x1="26.904"
          x2="54.181"
          y1="63.079"
          y2="15.835"
          gradientUnits="userSpaceOnUse"
        >
          <stop offset="0" stopColor="#fcc934" />
          <stop offset="1" stopColor="#fbbc04" />
        </linearGradient>
        <linearGradient
          id="browser-gradient-3"
          x1="4.221"
          x2="58.775"
          y1="19.688"
          y2="19.688"
          gradientUnits="userSpaceOnUse"
        >
          <stop offset="0" stopColor="#d93025" />
          <stop offset="1" stopColor="#ea4335" />
        </linearGradient>
      </defs>
    </svg>
  );
}
