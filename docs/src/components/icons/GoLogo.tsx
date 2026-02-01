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

export default function GoLogo({size = 64}: {size?: number}) {
  return (
    <svg width={size} height={size} viewBox="0 0 32 32" fill="none">
      <path
        fill="#384E54"
        d="M24.93 7.577c-.972-1.721-.27-3.958 1.567-4.995 1.837-1.038 4.115-.484 5.087 1.238.972 1.721.27 3.958-1.567 4.995-1.838 1.038-4.115.483-5.087-1.238ZM7.164 7.577c.972-1.721.27-3.958-1.567-4.995C3.759 1.544 1.482 2.098.51 3.82-.462 5.54.24 7.778 2.077 8.815c1.838 1.038 4.115.483 5.087-1.238Z"
      />
      <path
        fill="#384E54"
        d="M16.092 1.002c-1.106.01-2.21.049-3.316.09C6.842 1.312 2 6.032 2 12v20h28V12c0-5.968-4.667-10.491-10.59-10.908a42.248 42.248 0 0 0-3.318-.09Z"
      />
      <path fill="#76E1FE" d="M16 3C7 3 4 6.715 4 12v20h24c-.049-7.356 0-18 0-20 0-5.285-3-9-12-9Z" />
    </svg>
  );
}
