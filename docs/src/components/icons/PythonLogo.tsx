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

export default function PythonLogo({size = 64}: {size?: number}) {
  return (
    <svg width={size} height={size * 1.018} viewBox="0 0 112 114" fill="none">
      <path
        fill="url(#python-gradient-1)"
        d="M55.333 0c-4.618.022-9.029.416-12.91 1.103-11.43 2.02-13.507 6.247-13.507 14.043V25.44h27.015v3.432H18.778c-7.851 0-14.726 4.72-16.876 13.697-2.48 10.29-2.59 16.71 0 27.455 1.92 7.998 6.506 13.696 14.357 13.696h9.288V71.38c0-8.917 7.715-16.782 16.877-16.782h26.983c7.511 0 13.508-6.184 13.508-13.728V15.146c0-7.322-6.177-12.821-13.508-14.043C64.767.33 59.951-.021 55.333 0Zm-14.61 8.282c2.791 0 5.07 2.316 5.07 5.163 0 2.838-2.279 5.133-5.07 5.133-2.8 0-5.069-2.295-5.069-5.133 0-2.847 2.27-5.163 5.07-5.163Z"
      />
      <path
        fill="url(#python-gradient-2)"
        d="M86.284 28.873V40.87c0 9.3-7.885 17.128-16.876 17.128H42.424c-7.39 0-13.507 6.326-13.507 13.728v25.723c0 7.322 6.366 11.628 13.507 13.728 8.552 2.515 16.752 2.969 26.984 0 6.8-1.969 13.507-5.932 13.507-13.728V87.154H55.932v-3.432h40.49c7.852 0 10.777-5.477 13.508-13.697 2.82-8.462 2.7-16.6 0-27.455-1.941-7.816-5.646-13.697-13.508-13.697H86.284ZM71.108 94.018c2.8 0 5.07 2.294 5.07 5.132 0 2.847-2.27 5.163-5.07 5.163-2.79 0-5.07-2.316-5.07-5.163 0-2.838 2.28-5.133 5.07-5.133Z"
      />
      <defs>
        <linearGradient id="python-gradient-1" x1="0" x2="62.293" y1="0" y2="53.082" gradientUnits="userSpaceOnUse">
          <stop stopColor="#5A9FD4" />
          <stop offset="1" stopColor="#306998" />
        </linearGradient>
        <linearGradient
          id="python-gradient-2"
          x1="70.459"
          x2="48.11"
          y1="98.284"
          y2="66.967"
          gradientUnits="userSpaceOnUse"
        >
          <stop stopColor="#FFD43B" />
          <stop offset="1" stopColor="#FFE873" />
        </linearGradient>
      </defs>
    </svg>
  );
}
