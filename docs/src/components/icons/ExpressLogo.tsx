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

export default function ExpressLogo({size = 64}: {size?: number}) {
  return (
    <svg width={size} height={size} viewBox="0 0 20 20" style={{fill: 'white'}}>
      <path d="M6.504 7.181c1.47 0 1.812 1.29 1.812 2.108H4.5c.103-.906.683-2.108 2.004-2.108Z" />
      <path
        fillRule="evenodd"
        d="M10 20a10 10 0 1 0 0-20 10 10 0 0 0 0 20Zm-3.05-7.291c-1.321 0-2.438-.728-2.464-2.492l5.032.013c.04-.2.062-.405.058-.61 0-1.32-.621-3.37-2.955-3.37-2.109 0-3.385 1.737-3.385 3.875 0 2.137 1.328 3.625 3.535 3.625a5.738 5.738 0 0 0 2.39-.475l-.223-.938a4.65 4.65 0 0 1-1.988.372Zm5.833-4.78L11.759 6.4h-1.455l2.437 3.505-2.555 3.666h1.439l1.04-1.604a26.7 26.7 0 0 1 .261-.425c.171-.274.336-.538.494-.837h.031l.023.037c.245.413.479.807.75 1.225l1.067 1.604h1.471L14.238 9.86l2.45-3.46h-1.425l-.995 1.514c-.096.157-.194.312-.293.47-.146.231-.294.465-.435.704h-.03l-.165-.273c-.176-.291-.35-.58-.563-.887Z"
      />
    </svg>
  );
}
