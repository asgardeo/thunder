/**
 * Copyright (c) 2024, WSO2 LLC. (https://www.wso2.com).
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

export default function ReactRouterLogo({size = 64}: {size?: number}) {
  return (
    <svg width={size} height={size} viewBox="0 0 32 32">
      <path
        fill="#F44250"
        d="M25.6 15.568a3.196 3.196 0 0 1-3.2 3.19 3.193 3.193 0 0 0-3.2 3.189 3.196 3.196 0 0 1-3.2 3.19 3.193 3.193 0 0 1-2.263-.935 3.172 3.172 0 0 1-.937-2.255 3.193 3.193 0 0 1 3.2-3.19 3.196 3.196 0 0 0 3.2-3.189 3.194 3.194 0 0 0-3.2-3.19 3.193 3.193 0 0 1-3.2-3.189A3.193 3.193 0 0 1 16 6c1.767.001 3.2 1.429 3.2 3.19a3.193 3.193 0 0 0 3.2 3.19c1.768 0 3.2 1.427 3.2 3.19Z"
      />
      <path
        d="M12.8 15.568a3.195 3.195 0 0 0-3.2-3.19 3.195 3.195 0 0 0-3.2 3.19 3.195 3.195 0 0 0 3.2 3.19c1.767 0 3.2-1.428 3.2-3.19ZM6.4 21.947a3.195 3.195 0 0 0-3.2-3.19 3.195 3.195 0 0 0-3.2 3.19 3.195 3.195 0 0 0 3.2 3.19c1.767 0 3.2-1.429 3.2-3.19ZM32 21.947a3.195 3.195 0 0 0-3.2-3.19 3.195 3.195 0 0 0-3.2 3.19 3.195 3.195 0 0 0 3.2 3.19c1.767 0 3.2-1.429 3.2-3.19Z"
        style={{fill: '#fff'}}
      />
    </svg>
  );
}
