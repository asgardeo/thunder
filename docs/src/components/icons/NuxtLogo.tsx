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

export default function NuxtLogo({size = 64}: {size?: number}) {
  return (
    <svg width={size} height={size} viewBox="0 0 300 300">
      <path
        fill="#00DC82"
        d="M168 250h111c3.542 0 6.932-1.244 10-3 3.068-1.756 6.23-3.959 8-7 1.77-3.041 3.002-6.49 3-10.001-.002-3.511-1.227-6.959-3-9.998L222 91c-1.77-3.04-3.933-5.245-7-7s-7.458-3-11-3-6.933 1.245-10 3-5.23 3.96-7 7l-19 33-38-64.002c-1.772-3.04-3.932-6.243-7-7.998s-6.458-2-10-2-6.932.245-10 2c-3.068 1.755-6.228 4.958-8 7.998L2 220c-1.773 3.039-1.998 6.487-2 9.998-.002 3.511.23 6.96 2 10.001 1.77 3.04 4.932 5.244 8 7 3.068 1.756 6.458 3 10 3h70c27.737 0 47.925-12.442 62-36l34-59 18-31 55 94h-73l-18 32Zm-79-32H40l73-126 37 63-24.509 42.725C116.144 213.01 105.488 218 89 218Z"
      />
    </svg>
  );
}
