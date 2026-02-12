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

export default function TanStackLogo({size = 64}: {size?: number}) {
  return (
    <svg width={size} height={size} viewBox="0 0 256 256" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path
        d="M141.022 38.339L198.36 153.663L198.467 153.877C202.456 162.11 199.267 172.165 191.034 176.154C188.14 177.602 184.928 178.363 181.665 178.363H68.335C58.945 178.363 51.335 170.753 51.335 161.363C51.335 158.1 52.096 154.888 53.544 151.994L53.545 151.992L110.978 38.339C114.967 30.106 125.022 26.917 133.255 30.906C136.149 32.354 138.574 34.445 140.359 37.014C140.893 37.747 141.393 38.535 141.022 38.339Z"
        fill="url(#paint0_linear_953_11347)"
      />
      <path
        fillRule="evenodd"
        clipRule="evenodd"
        d="M125 64.9641L67.7891 177.791H182.211L125 64.9641ZM113.022 38.339L55.6846 151.665L55.5777 151.879C51.5887 160.112 54.7777 170.167 63.0107 174.156C65.9047 175.604 69.1167 176.365 72.3797 176.365H185.71C195.1 176.365 202.71 168.755 202.71 159.365C202.71 156.102 201.949 152.89 200.501 149.996L200.5 149.994L143.067 36.341C139.078 28.108 129.023 24.919 120.79 28.908C117.896 30.356 115.471 32.447 113.686 35.016C113.152 35.749 112.652 36.537 113.022 38.339Z"
        fill="url(#paint1_linear_953_11347)"
      />
      <path d="M143 106.964L184.642 181.363H101.358L143 106.964Z" fill="url(#paint2_linear_953_11347)" />
      <defs>
        <linearGradient
          id="paint0_linear_953_11347"
          x1="125"
          y1="28.0817"
          x2="125"
          y2="178.363"
          gradientUnits="userSpaceOnUse"
        >
          <stop stopColor="#00C7B7" />
          <stop offset="1" stopColor="#4299E1" />
        </linearGradient>
        <linearGradient
          id="paint1_linear_953_11347"
          x1="129.355"
          y1="24.2578"
          x2="129.355"
          y2="176.365"
          gradientUnits="userSpaceOnUse"
        >
          <stop stopColor="#00C7B7" />
          <stop offset="1" stopColor="#4299E1" />
        </linearGradient>
        <linearGradient
          id="paint2_linear_953_11347"
          x1="143"
          y1="106.964"
          x2="143"
          y2="181.363"
          gradientUnits="userSpaceOnUse"
        >
          <stop stopColor="#FACC15" />
          <stop offset="1" stopColor="#F97316" />
        </linearGradient>
      </defs>
    </svg>
  );
}
