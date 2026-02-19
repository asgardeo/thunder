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

export default function CiscoLogo({size = 64}: {size?: number}) {
  return (
    <svg xmlns="http://www.w3.org/2000/svg" width={size * 1.8} height={size} viewBox="0 0 90 40" fill="currentColor">
      <rect x="8" y="14" width="3" height="16" rx="1.5" opacity="0.8" />
      <rect x="15" y="8" width="3" height="22" rx="1.5" opacity="0.8" />
      <rect x="22" y="14" width="3" height="16" rx="1.5" opacity="0.8" />
      <rect x="36" y="14" width="3" height="16" rx="1.5" opacity="0.8" />
      <rect x="43" y="8" width="3" height="22" rx="1.5" opacity="0.8" />
      <rect x="50" y="14" width="3" height="16" rx="1.5" opacity="0.8" />
      <rect x="64" y="14" width="3" height="16" rx="1.5" opacity="0.8" />
      <rect x="71" y="8" width="3" height="22" rx="1.5" opacity="0.8" />
      <rect x="78" y="14" width="3" height="16" rx="1.5" opacity="0.8" />
    </svg>
  );
}
