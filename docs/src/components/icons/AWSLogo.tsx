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

export default function AWSLogo({size = 64}: {size?: number}) {
  return (
    <svg xmlns="http://www.w3.org/2000/svg" width={size * 1.6} height={size} viewBox="0 0 80 48" fill="currentColor">
      <path
        d="M22.8 28.6c-2.4 1.8-5.9 2.7-8.9 2.7-4.2 0-8-1.6-10.9-4.2-.2-.2 0-.5.3-.3 3.1 1.8 6.9 2.9 10.9 2.9 2.7 0 5.6-.6 8.3-1.7.4-.2.7.3.3.6z"
        opacity="0.8"
      />
      <path d="M24 26.9c-.3-.4-2.1-.2-2.9-.1-.2 0-.3-.2-.1-.3 1.4-1 3.8-.7 4-.4.3.4-.1 2.8-1.4 4-.2.2-.4.1-.3-.1.3-.7.9-2.3.7-3.1z" opacity="0.8" />
      <path
        d="M21.2 17.1v-1.4c0-.2.2-.4.3-.4l6.1-.0c.2 0 .3.1.3.3v1.2c0 .2-.2.5-.5.8l-3.2 4.5c1.2 0 2.4.1 3.5.8.2.1.3.4.3.6v1.5c0 .2-.2.5-.5.3-1.9-1-4.5-1.1-6.6 0-.2.1-.5-.1-.5-.3v-1.4c0-.2 0-.6.3-1l3.7-5.3-3.2 0c-.2 0-.3-.2-.4-.3zm-10.1 8h-1.9c-.2 0-.3-.2-.3-.3v-9.4c0-.2.2-.3.4-.3h1.7c.2 0 .3.2.3.3v1.2h0c.4-1.2 1.2-1.7 2.2-1.7 1.1 0 1.7.6 2.2 1.7.4-1.2 1.3-1.7 2.2-1.7.7 0 1.4.3 1.8.9.5.7.4 1.7.4 2.6v6.4c0 .2-.2.3-.4.3h-1.9c-.2 0-.3-.2-.3-.3v-5.4c0-.3 0-1.2-.1-1.5-.1-.5-.4-.7-.8-.7-.3 0-.7.2-.8.6-.2.4-.2 1-.2 1.5v5.4c0 .2-.2.3-.4.3h-1.9c-.2 0-.3-.2-.3-.3v-5.4c0-.9 0-2.2-.9-2.2s-.9 1.3-.9 2.2v5.4c0 .2-.2.3-.4.3zm29.6-10c2.8 0 4.4 2.4 4.4 5.5 0 3-1.7 5.3-4.4 5.3-2.8 0-4.3-2.4-4.3-5.4 0-3 1.6-5.4 4.3-5.4zm0 2c-1.4 0-1.5 2-1.5 3.2 0 1.2 0 3.5 1.5 3.5 1.5 0 1.5-1.9 1.5-3.1 0-.8 0-1.7-.2-2.4-.2-.6-.6-1.2-1.3-1.2zm8.1 8h-1.9c-.2 0-.3-.2-.3-.3v-9.5c0-.2.2-.3.4-.3h1.7c.2 0 .3.1.3.3v1.4h0c.5-1.3 1.1-1.9 2.3-1.9.7 0 1.5.3 1.9 1 .4.6.4 1.7.4 2.5v6.4c0 .2-.2.3-.4.3h-1.9c-.2 0-.3-.2-.3-.3v-5.5c0-.9.1-2.1-.9-2.1-.4 0-.7.2-.9.6-.2.5-.3 1-.3 1.6v5.4c0 .2-.2.3-.4.3z"
        opacity="0.8"
      />
    </svg>
  );
}
