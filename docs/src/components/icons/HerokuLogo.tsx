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

export default function HerokuLogo({size = 64}: {size?: number}) {
  return (
    <svg xmlns="http://www.w3.org/2000/svg" width={size * 0.7} height={size} viewBox="0 0 18 24" fill="currentColor">
      <path
        d="M15.5 0h-13C1.12 0 0 1.12 0 2.5v19C0 22.88 1.12 24 2.5 24h13c1.38 0 2.5-1.12 2.5-2.5v-19C18 1.12 16.88 0 15.5 0zM16 21.5c0 .28-.22.5-.5.5h-13c-.28 0-.5-.22-.5-.5v-19c0-.28.22-.5.5-.5h13c.28 0 .5.22.5.5v19zM6 19l3-3-3-3v6zm3-11.5c1.1 0 2.13-.37 3-1V10h2V4h-2c-.87.63-1.9 1-3 1s-2.13-.37-3-1H4v3h2V5.5c.87.63 1.9 1 3 1zM4 14h2V7H4v7z"
        opacity="0.8"
      />
    </svg>
  );
}
