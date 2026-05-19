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

export default function JavaScriptLogo({size = 64}: {size?: number}) {
  return (
    <svg width={size} height={size} viewBox="0 0 64 64" fill="none" xmlns="http://www.w3.org/2000/svg" role="img" aria-label="JavaScript logo">
      <rect x="6" y="6" width="52" height="52" rx="10" fill="#F7DF1E" />
      <path
        d="M29.4 42.3C30.2 43.6 31.2 44.5 33 44.5C34.5 44.5 35.4 43.8 35.4 42.8C35.4 41.6 34.6 41.2 33.2 40.6L32.4 40.3C30 39.3 28.4 38 28.4 35.3C28.4 32.9 30.3 31.1 33.2 31.1C35.2 31.1 36.6 31.8 37.7 33.7L35.2 35.3C34.7 34.3 34 33.9 33.2 33.9C32.4 33.9 31.8 34.4 31.8 35.1C31.8 36.1 32.4 36.5 33.8 37.1L34.6 37.4C37.4 38.6 39 39.9 39 42.8C39 46 36.4 47.6 33.1 47.6C29.8 47.6 27.7 46.1 26.8 44.2L29.4 42.3Z"
        fill="#1B1F23"
      />
      <path
        d="M23.3 42.4C23.9 43.4 24.5 44.2 25.8 44.2C27.1 44.2 27.9 43.7 27.9 41.9V31.3H31.2V42.1C31.2 45.4 29.3 47 26.3 47C23.6 47 22 45.6 21.2 43.9L23.3 42.4Z"
        fill="#1B1F23"
      />
    </svg>
  );
}
