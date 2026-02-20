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

export default function MicrosoftLogo({size = 64}: {size?: number}) {
  return (
    <svg xmlns="http://www.w3.org/2000/svg" width={size} height={size} viewBox="0 0 24 24" fill="currentColor">
      <rect x="1" y="1" width="10" height="10" opacity="0.7" />
      <rect x="13" y="1" width="10" height="10" opacity="0.6" />
      <rect x="1" y="13" width="10" height="10" opacity="0.6" />
      <rect x="13" y="13" width="10" height="10" opacity="0.5" />
    </svg>
  );
}
