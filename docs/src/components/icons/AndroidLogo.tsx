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

export default function AndroidLogo({size = 64}: {size?: number}) {
  return (
    <svg width={size} height={size * 0.586} viewBox="0 0 430 252" fill="none">
      <defs>
        <linearGradient id="android-gradient-1" x1="68.188%" x2="27.823%" y1="17.487%" y2="89.755%">
          <stop offset="0%" stopColor="#41873F" />
          <stop offset="32.88%" stopColor="#418B3D" />
          <stop offset="63.52%" stopColor="#419637" />
          <stop offset="93.19%" stopColor="#3FA92D" />
          <stop offset="100%" stopColor="#3FAE2A" />
        </linearGradient>
        <linearGradient id="android-gradient-2" x1="43.277%" x2="159.245%" y1="55.169%" y2="-18.306%">
          <stop offset="13.76%" stopColor="#41873F" />
          <stop offset="40.32%" stopColor="#54A044" />
          <stop offset="71.36%" stopColor="#66B848" />
          <stop offset="90.81%" stopColor="#6CC04A" />
        </linearGradient>
      </defs>
      <path
        fill="url(#android-gradient-1)"
        d="M425.097 223.146a215 215 0 0 0-8.475-26.84 215 215 0 0 0-9.733-21.445 218 218 0 0 0-15.693-25.584 216 216 0 0 0-23.698-27.898 213 213 0 0 0-11.283-10.445 216.6 216.6 0 0 0-26.871-19.78c.083-.135.157-.282.241-.418 4.337-7.49 8.684-14.97 13.022-22.461l12.728-21.938c3.049-5.248 6.098-10.508 9.136-15.756a20 20 0 0 0 1.718-3.855 19.78 19.78 0 0 0-.608-13.745 20 20 0 0 0-1.246-2.473 19.6 19.6 0 0 0-6.852-6.82 19.9 19.9 0 0 0-8.412-2.703 20 20 0 0 0-3.72.032 19.735 19.735 0 0 0-15.065 9.733c-3.048 5.248-6.097 10.507-9.135 15.756l-12.729 21.937c-4.337 7.49-8.684 14.97-13.022 22.461-.471.817-.953 1.634-1.424 2.462-.66-.262-1.31-.524-1.97-.775-23.917-9.115-49.867-14.101-76.99-14.101-.744 0-1.477 0-2.221.01-24.117.241-47.29 4.432-68.903 11.954a197 197 0 0 0-7.428 2.755c-.44-.765-.89-1.53-1.33-2.294-4.337-7.491-8.685-14.971-13.022-22.462l-12.729-21.937c-3.048-5.249-6.097-10.508-9.135-15.756a20 20 0 0 0-2.494-3.405 19.74 19.74 0 0 0-12.571-6.328 20 20 0 0 0-3.72-.031 19.64 19.64 0 0 0-8.412 2.703 19.7 19.7 0 0 0-6.851 6.82 20 20 0 0 0-1.247 2.472 19 19 0 0 0-.89 2.609 19.8 19.8 0 0 0 .282 11.136c.43 1.32.996 2.609 1.719 3.855 3.048 5.249 6.097 10.508 9.135 15.757l12.729 21.937c4.337 7.49 8.685 14.97 13.022 22.461.031.063.073.126.104.189a217 217 0 0 0-24.933 18.05A218 218 0 0 0 62.533 121.4a217.7 217.7 0 0 0-23.697 27.898 214.5 214.5 0 0 0-15.694 25.583 215.2 215.2 0 0 0-18.208 48.286A216 216 0 0 0 0 251.233h430c-1.027-9.434-2.639-18.845-4.903-28.087"
      />
      <path
        fill="url(#android-gradient-2)"
        d="M128.142.55l119.26 68.846v137.691l-119.26 68.846L8.884 207.087V69.396L128.142.55z"
        transform="translate(86 40)"
      />
      <circle cx="123" cy="159" r="7" fill="#fff" />
      <circle cx="307" cy="159" r="7" fill="#fff" />
    </svg>
  );
}
